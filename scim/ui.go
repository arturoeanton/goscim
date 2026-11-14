package scim

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/gin-gonic/gin"
)

//go:embed ui/assets/* ui/templates/*.html
var uiEmbeddedFS embed.FS

const uiCookieName = "goscim_ui_token"

// UIOptions configures the embedded HTMX management and inspector console.
type UIOptions struct {
	Version string
	UIToken string
}

// UIEnabledFromEnv reports whether the embedded HTMX Web UI should be mounted
// at /ui (`SCIM_UI=true|1|yes|on`).
func UIEnabledFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_UI")))
	return v == "true" || v == "1" || v == "yes" || v == "on"
}

// MountUI registers the embedded HTMX Web Console at /ui and enables the
// bounded live SCIM request/response inspector buffer.
func MountUI(r *gin.Engine, opts UIOptions) error {
	tmpl, err := template.ParseFS(uiEmbeddedFS, "ui/templates/*.html")
	if err != nil {
		return fmt.Errorf("parsing embedded UI templates: %w", err)
	}
	DefaultInspector.SetEnabled(true)

	version := strings.TrimSpace(opts.Version)
	if version == "" {
		if v, ok := DefaultMetrics.buildVersion.Load().(string); ok && v != "" {
			version = v
		} else {
			version = "dev"
		}
	}
	uiToken := opts.UIToken
	if uiToken == "" {
		uiToken = strings.TrimSpace(os.Getenv("SCIM_UI_TOKEN"))
	}

	h := &uiController{
		engine:  r,
		tmpl:    tmpl,
		version: version,
		uiToken: uiToken,
	}

	r.GET("/ui/assets/:file", h.serveAsset)
	r.GET("/ui/login", h.loginPage)
	r.POST("/ui/login", h.loginSubmit)
	r.POST("/ui/logout", h.logoutSubmit)

	uiGroup := r.Group("/ui", h.requireUIAuth())
	uiGroup.GET("", h.overviewPage)
	uiGroup.GET("/", h.overviewPage)
	uiGroup.GET("/partials/stats", h.statsPartial)
	uiGroup.POST("/actions/seed", h.seedAction)

	uiGroup.GET("/resources", h.resourcesPage)
	uiGroup.GET("/partials/resources-table", h.resourcesTablePartial)
	uiGroup.GET("/resources/detail", h.resourceDetailPartial)
	uiGroup.POST("/resources/users", h.createUserAction)
	uiGroup.POST("/resources/groups", h.createGroupAction)
	uiGroup.POST("/resources/toggle-active", h.toggleActiveAction)
	uiGroup.DELETE("/resources/item", h.deleteResourceAction)

	uiGroup.GET("/playground", h.playgroundPage)
	uiGroup.POST("/playground/eval", h.playgroundEvalPartial)

	uiGroup.GET("/inspector", h.inspectorPage)
	uiGroup.GET("/partials/inspector-rows", h.inspectorRowsPartial)
	uiGroup.GET("/inspector/detail", h.inspectorDetailPartial)
	uiGroup.POST("/inspector/clear", h.inspectorClearAction)

	uiGroup.GET("/conformance", h.conformancePage)
	uiGroup.POST("/conformance/run", h.conformanceRunPartial)

	return nil
}

type uiController struct {
	engine  *gin.Engine
	tmpl    *template.Template
	version string
	uiToken string
}

type uiPageBase struct {
	Title          string
	Version        string
	ActiveNav      string
	CurrentPath    string
	Tenant         string
	TenantQuery    string
	StoreBackend   string
	AuthMode       string
	PasswordMode   string
	SCIMBasePath   string
	SCIMBaseURL    string
	MetricsEnabled bool
	HasUIToken     bool
}

type uiOverviewData struct {
	uiPageBase
	UserCount       int
	GroupCount      int
	Summary         MetricsSummary
	StoreHealthy    bool
	WebhookCount    int
	SeedNotice      string
	SampleTokenHint string
}

type uiResourceRow struct {
	ID           string
	Primary      string
	Secondary    string
	Tertiary     string
	ExternalID   string
	HasActive    bool
	Active       bool
	BadgeText    string
	LastModified string
}

type uiResourcesData struct {
	uiPageBase
	ResourceType  string
	ResourceLabel string
	Filter        string
	Total         int
	Rows          []uiResourceRow
	FlashMessage  string
	FlashError    bool
}

type uiGroupBadge struct {
	Display string
	Type    string
}

type uiResourceDetailData struct {
	ID          string
	ResourceURL string
	PrettyJSON  string
	Groups      []uiGroupBadge
}

type uiPlaygroundData struct {
	uiPageBase
	ResourceType   string
	Filter         string
	N1QLStatement  string
	N1QLParamsJSON string
	TotalResults   int
	ResponseJSON   string
	Error          string
}

type uiInspectorData struct {
	uiPageBase
	Exchanges []InspectedExchange
}

type uiInspectorDetailData struct {
	Seq             uint64
	Timestamp       string
	RequestID       string
	Tenant          string
	Subject         string
	Method          string
	Path            string
	Action          string
	ResourceType    string
	ResourceID      string
	Status          int
	ScimType        string
	DurationMs      float64
	RequestPreview  string
	ResponsePreview string
}

type uiConformanceData struct {
	uiPageBase
	WebhookURLs []string
	Deliveries  []WebhookDelivery
	Report      ConformanceReport
}

func (h *uiController) currentToken() string {
	if h.uiToken != "" {
		return h.uiToken
	}
	return strings.TrimSpace(os.Getenv("SCIM_UI_TOKEN"))
}

func (h *uiController) requireUIAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := h.currentToken()
		if expected == "" {
			c.Next()
			return
		}
		if bearerHeaderMatches(c.GetHeader("Authorization"), expected) {
			c.Next()
			return
		}
		if cookieVal, err := c.Cookie(uiCookieName); err == nil && constantTimeEqual(cookieVal, expected) {
			c.Next()
			return
		}
		c.Redirect(http.StatusSeeOther, "/ui/login")
		c.Abort()
	}
}

func (h *uiController) serveAsset(c *gin.Context) {
	name := c.Param("file")
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		c.Status(http.StatusNotFound)
		return
	}
	raw, err := uiEmbeddedFS.ReadFile("ui/assets/" + name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	contentType := "application/octet-stream"
	switch {
	case strings.HasSuffix(name, ".css"):
		contentType = "text/css; charset=utf-8"
	case strings.HasSuffix(name, ".js"):
		contentType = "application/javascript; charset=utf-8"
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.Data(http.StatusOK, contentType, raw)
}

func (h *uiController) loginPage(c *gin.Context) {
	if h.currentToken() == "" {
		c.Redirect(http.StatusSeeOther, "/ui")
		return
	}
	h.render(c, http.StatusOK, "login.html", map[string]interface{}{
		"Title": "Sign In",
		"Error": "",
	})
}

func (h *uiController) loginSubmit(c *gin.Context) {
	expected := h.currentToken()
	if expected == "" {
		c.Redirect(http.StatusSeeOther, "/ui")
		return
	}
	presented := strings.TrimSpace(c.PostForm("token"))
	if !constantTimeEqual(presented, expected) {
		h.render(c, http.StatusUnauthorized, "login.html", map[string]interface{}{
			"Title": "Sign In",
			"Error": "Invalid console access token.",
		})
		return
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(uiCookieName, expected, 86400, "/ui", "", c.Request.TLS != nil, true)
	c.Redirect(http.StatusSeeOther, "/ui")
}

func (h *uiController) logoutSubmit(c *gin.Context) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(uiCookieName, "", -1, "/ui", "", c.Request.TLS != nil, true)
	c.Redirect(http.StatusSeeOther, "/ui/login")
}

func (h *uiController) resolveTenant(c *gin.Context) string {
	raw := strings.TrimSpace(c.Query("tenant"))
	if raw == "" {
		raw = strings.TrimSpace(c.PostForm("tenant"))
	}
	if !IsValidTenantID(raw) {
		return ""
	}
	return raw
}

func (h *uiController) pageBase(c *gin.Context, title, nav string) uiPageBase {
	tenant := h.resolveTenant(c)
	tq := ""
	basePath := PREFIX
	if tenant != "" {
		tq = "?tenant=" + url.QueryEscape(tenant)
		basePath = "/tenants/" + tenant + PREFIX
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := safeHost(c.Request.Host)
	if host == "" {
		host = "127.0.0.1:8080"
	}
	baseURL := scheme + "://" + host + basePath
	if pub := strings.TrimRight(strings.TrimSpace(os.Getenv("SCIM_PUBLIC_URL")), "/"); pub != "" {
		baseURL = pub + basePath
	}

	authMode := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_AUTH")))
	if authMode == "" {
		authMode = "configured"
	}
	pwMode, err := PasswordModeFromEnv()
	pwLabel := string(pwMode)
	if err != nil {
		pwLabel = "argon2id"
	}

	return uiPageBase{
		Title:          title,
		Version:        h.version,
		ActiveNav:      nav,
		CurrentPath:    c.Request.URL.Path,
		Tenant:         tenant,
		TenantQuery:    tq,
		StoreBackend:   ActiveStoreBackend(),
		AuthMode:       authMode,
		PasswordMode:   pwLabel,
		SCIMBasePath:   basePath,
		SCIMBaseURL:    baseURL,
		MetricsEnabled: MetricsEnabledFromEnv(),
		HasUIToken:     h.currentToken() != "",
	}
}

func (h *uiController) buildOverviewData(c *gin.Context, seedNotice string) uiOverviewData {
	base := h.pageBase(c, "Overview", "overview")
	var userCount, groupCount int
	storeHealthy := false
	if DB != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		storeHealthy = DB.Ping(ctx) == nil
		cancel()
		if uRT, ok := Resources["/Users"]; ok {
			userCount, _, _ = DB.Search(SearchQuery{Bucket: ScopedBucket(base.Tenant, uRT.Name), Limit: 0})
		}
		if gRT, ok := Resources["/Groups"]; ok {
			groupCount, _, _ = DB.Search(SearchQuery{Bucket: ScopedBucket(base.Tenant, gRT.Name), Limit: 0})
		}
	}
	webhookCount := 0
	if d := ActiveWebhookDispatcher(); d != nil {
		webhookCount = len(d.URLs())
	}
	tokenHint := os.Getenv("SCIM_BEARER_TOKEN")
	if tokenHint == "" {
		tokenHint = "<YOUR_SCIM_TOKEN>"
	}
	return uiOverviewData{
		uiPageBase:      base,
		UserCount:       userCount,
		GroupCount:      groupCount,
		Summary:         DefaultMetrics.Summary(),
		StoreHealthy:    storeHealthy,
		WebhookCount:    webhookCount,
		SeedNotice:      seedNotice,
		SampleTokenHint: tokenHint,
	}
}

func (h *uiController) overviewPage(c *gin.Context) {
	notice := ""
	if seeded := c.Query("seeded"); seeded != "" {
		notice = fmt.Sprintf("Seeded %s sample SCIM resource(s) into the selected namespace.", seeded)
	}
	h.render(c, http.StatusOK, "overview.html", h.buildOverviewData(c, notice))
}

func (h *uiController) statsPartial(c *gin.Context) {
	h.render(c, http.StatusOK, "stats_strip", h.buildOverviewData(c, ""))
}

func (h *uiController) seedAction(c *gin.Context) {
	tenant := h.resolveTenant(c)
	count, err := SeedDemoResources(tenant, true)
	target := "/ui?seeded=" + strconv.Itoa(count)
	if tenant != "" {
		target += "&tenant=" + url.QueryEscape(tenant)
	}
	if err != nil {
		target = "/ui"
	}
	c.Redirect(http.StatusSeeOther, target)
}

func (h *uiController) resourcesPage(c *gin.Context) {
	resType := normalizeUIResourceTab(c.Query("type"))
	filter := strings.TrimSpace(c.Query("filter"))
	data := h.buildResourcesData(c, resType, filter, "", false)
	h.render(c, http.StatusOK, "resources.html", data)
}

func (h *uiController) resourcesTablePartial(c *gin.Context) {
	resType := normalizeUIResourceTab(c.Query("type"))
	filter := strings.TrimSpace(c.Query("filter"))
	data := h.buildResourcesData(c, resType, filter, "", false)
	h.render(c, http.StatusOK, "resources_table", data)
}

func normalizeUIResourceTab(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "groups", "group":
		return "Groups"
	case "elements", "element":
		return "Elements"
	default:
		return "Users"
	}
}

func (h *uiController) buildResourcesData(c *gin.Context, resType, filter, flash string, flashErr bool) uiResourcesData {
	base := h.pageBase(c, "Directory Resources", "resources")
	endpoint := "/" + resType
	rt, ok := Resources[endpoint]
	out := uiResourcesData{
		uiPageBase:    base,
		ResourceType:  resType,
		ResourceLabel: resType,
		Filter:        filter,
		FlashMessage:  flash,
		FlashError:    flashErr,
	}
	if !ok || DB == nil {
		return out
	}

	total, docs, err := DB.Search(SearchQuery{
		Bucket: ScopedBucket(base.Tenant, rt.Name),
		Filter: filter,
		ResolveName: func(p string) string {
			return CanonicalAttributePath(rt, p)
		},
		IsCaseExact: func(p string) bool {
			return AttributeIsCaseExact(rt, p)
		},
		IsMultiValued: func(p string) bool {
			return AttributeIsMultiValuedComplex(rt, p)
		},
		Limit: 100,
	})
	if err != nil {
		out.FlashMessage = err.Error()
		out.FlashError = true
		return out
	}
	out.Total = total
	rows := make([]uiResourceRow, 0, len(docs))
	for _, doc := range docs {
		rows = append(rows, summarizeDocRow(resType, doc))
	}
	out.Rows = rows
	return out
}

func summarizeDocRow(resType string, doc map[string]interface{}) uiResourceRow {
	id, _ := doc["id"].(string)
	extID, _ := doc["externalId"].(string)
	lastMod := ""
	if meta, ok := doc["meta"].(map[string]interface{}); ok {
		lastMod, _ = meta["lastModified"].(string)
	}
	row := uiResourceRow{
		ID:           id,
		ExternalID:   extID,
		LastModified: lastMod,
	}

	switch resType {
	case "Users":
		row.Primary, _ = doc["userName"].(string)
		row.Secondary, _ = doc["displayName"].(string)
		if row.Secondary == "" {
			if nm, ok := doc["name"].(map[string]interface{}); ok {
				gn, _ := nm["givenName"].(string)
				fn, _ := nm["familyName"].(string)
				row.Secondary = strings.TrimSpace(gn + " " + fn)
			}
		}
		if title, _ := doc["title"].(string); title != "" {
			row.Tertiary = title
		}
		if ext, ok := doc[enterpriseUserSchemaURI].(map[string]interface{}); ok {
			if dept, _ := ext["department"].(string); dept != "" {
				if row.Tertiary != "" {
					row.Tertiary += " · " + dept
				} else {
					row.Tertiary = dept
				}
			}
		}
		row.HasActive = true
		if act, ok := doc["active"].(bool); ok {
			row.Active = act
		} else {
			row.Active = true
		}
		if groups, ok := doc["groups"].([]interface{}); ok && len(groups) > 0 {
			row.BadgeText = fmt.Sprintf("%d group(s)", len(groups))
		}

	case "Groups":
		row.Primary, _ = doc["displayName"].(string)
		members, _ := doc["members"].([]interface{})
		row.Secondary = fmt.Sprintf("%d member(s)", len(members))
		row.BadgeText = fmt.Sprintf("%d member(s)", len(members))

	default:
		row.Primary, _ = doc["name"].(string)
		row.Secondary, _ = doc["description"].(string)
	}
	return row
}

func (h *uiController) dispatchInternalSCIM(ctx context.Context, method, path string, body interface{}) (int, []byte) {
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	reqCtx := withInternalUIPrincipal(ctx)
	req := httptest.NewRequest(method, path, bytes.NewReader(raw)).WithContext(reqCtx)
	if body != nil {
		req.Header.Set("Content-Type", scimContentType)
	}
	req.Header.Set("Accept", scimContentType)
	rec := httptest.NewRecorder()
	h.engine.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

func extractSCIMErrorDetail(raw []byte, code int) string {
	var parsed map[string]interface{}
	if json.Unmarshal(raw, &parsed) == nil {
		if d, _ := parsed["detail"].(string); d != "" {
			return fmt.Sprintf("HTTP %d: %s", code, d)
		}
	}
	return fmt.Sprintf("HTTP %d", code)
}

func (h *uiController) createUserAction(c *gin.Context) {
	base := h.pageBase(c, "Directory Resources", "resources")
	uName := strings.TrimSpace(c.PostForm("userName"))
	payload := map[string]interface{}{
		"schemas":  []string{coreUserSchemaURI},
		"userName": uName,
		"active":   true,
	}
	if dn := strings.TrimSpace(c.PostForm("displayName")); dn != "" {
		payload["displayName"] = dn
	}
	gn := strings.TrimSpace(c.PostForm("givenName"))
	fn := strings.TrimSpace(c.PostForm("familyName"))
	if gn != "" || fn != "" {
		payload["name"] = map[string]interface{}{
			"givenName":  gn,
			"familyName": fn,
			"formatted":  strings.TrimSpace(gn + " " + fn),
		}
	}
	if email := strings.TrimSpace(c.PostForm("email")); email != "" {
		payload["emails"] = []interface{}{
			map[string]interface{}{"value": email, "type": "work", "primary": true},
		}
	}
	if extID := strings.TrimSpace(c.PostForm("externalId")); extID != "" {
		payload["externalId"] = extID
	}
	if pw := c.PostForm("password"); pw != "" {
		payload["password"] = pw
	}
	if dept := strings.TrimSpace(c.PostForm("department")); dept != "" {
		payload["schemas"] = []string{coreUserSchemaURI, enterpriseUserSchemaURI}
		payload[enterpriseUserSchemaURI] = map[string]interface{}{"department": dept}
	}

	code, respBytes := h.dispatchInternalSCIM(c.Request.Context(), http.MethodPost, base.SCIMBasePath+"/Users", payload)
	if code != http.StatusCreated {
		data := h.buildResourcesData(c, "Users", "", extractSCIMErrorDetail(respBytes, code), true)
		h.render(c, http.StatusOK, "resources_table", data)
		return
	}
	data := h.buildResourcesData(c, "Users", "", "Created user "+uName, false)
	h.render(c, http.StatusOK, "resources_table", data)
}

func (h *uiController) createGroupAction(c *gin.Context) {
	base := h.pageBase(c, "Directory Resources", "resources")
	dName := strings.TrimSpace(c.PostForm("displayName"))
	payload := map[string]interface{}{
		"schemas":     []string{coreGroupSchemaURI},
		"displayName": dName,
	}
	if extID := strings.TrimSpace(c.PostForm("externalId")); extID != "" {
		payload["externalId"] = extID
	}
	if rawMembers := strings.TrimSpace(c.PostForm("memberIDs")); rawMembers != "" {
		var members []interface{}
		for _, part := range splitAndTrim(rawMembers) {
			members = append(members, map[string]interface{}{"value": part})
		}
		if len(members) > 0 {
			payload["members"] = members
		}
	}

	code, respBytes := h.dispatchInternalSCIM(c.Request.Context(), http.MethodPost, base.SCIMBasePath+"/Groups", payload)
	if code != http.StatusCreated {
		data := h.buildResourcesData(c, "Groups", "", extractSCIMErrorDetail(respBytes, code), true)
		h.render(c, http.StatusOK, "resources_table", data)
		return
	}
	data := h.buildResourcesData(c, "Groups", "", "Created group "+dName, false)
	h.render(c, http.StatusOK, "resources_table", data)
}

func (h *uiController) toggleActiveAction(c *gin.Context) {
	base := h.pageBase(c, "Directory Resources", "resources")
	resType := normalizeUIResourceTab(c.PostForm("type"))
	id := strings.TrimSpace(c.PostForm("id"))
	nextActive := strings.EqualFold(strings.TrimSpace(c.PostForm("active")), "true")

	patch := map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []interface{}{
			map[string]interface{}{
				"op":    "replace",
				"path":  "active",
				"value": nextActive,
			},
		},
	}
	code, respBytes := h.dispatchInternalSCIM(c.Request.Context(), http.MethodPatch, base.SCIMBasePath+"/"+resType+"/"+url.PathEscape(id), patch)
	if code != http.StatusOK {
		data := h.buildResourcesData(c, resType, "", extractSCIMErrorDetail(respBytes, code), true)
		h.render(c, http.StatusOK, "resources_table", data)
		return
	}
	data := h.buildResourcesData(c, resType, "", fmt.Sprintf("Updated active=%v on %s", nextActive, id), false)
	h.render(c, http.StatusOK, "resources_table", data)
}

func (h *uiController) deleteResourceAction(c *gin.Context) {
	base := h.pageBase(c, "Directory Resources", "resources")
	resType := normalizeUIResourceTab(c.Query("type"))
	id := strings.TrimSpace(c.Query("id"))

	code, respBytes := h.dispatchInternalSCIM(c.Request.Context(), http.MethodDelete, base.SCIMBasePath+"/"+resType+"/"+url.PathEscape(id), nil)
	if code != http.StatusNoContent {
		data := h.buildResourcesData(c, resType, "", extractSCIMErrorDetail(respBytes, code), true)
		h.render(c, http.StatusOK, "resources_table", data)
		return
	}
	data := h.buildResourcesData(c, resType, "", "Deleted resource "+id, false)
	h.render(c, http.StatusOK, "resources_table", data)
}

func (h *uiController) resourceDetailPartial(c *gin.Context) {
	base := h.pageBase(c, "Directory Resources", "resources")
	resType := normalizeUIResourceTab(c.Query("type"))
	id := strings.TrimSpace(c.Query("id"))
	rt, ok := Resources["/"+resType]
	if !ok || DB == nil {
		c.Status(http.StatusNotFound)
		return
	}
	doc, err := DB.Get(ScopedBucket(base.Tenant, rt.Name), id)
	if err != nil {
		c.String(http.StatusNotFound, "Resource not found")
		return
	}
	formatted := FormatResource([]string{"admin", "superadmin", "user", "role1", "role2"}, rt, doc, Projection{})
	pretty, _ := json.MarshalIndent(formatted, "", "  ")

	var badges []uiGroupBadge
	if rawGroups, ok := formatted["groups"].([]interface{}); ok {
		for _, item := range rawGroups {
			if gm, ok := item.(map[string]interface{}); ok {
				disp, _ := gm["display"].(string)
				if disp == "" {
					disp, _ = gm["value"].(string)
				}
				tp, _ := gm["type"].(string)
				badges = append(badges, uiGroupBadge{Display: disp, Type: tp})
			}
		}
	}

	h.render(c, http.StatusOK, "resource_detail", uiResourceDetailData{
		ID:          id,
		ResourceURL: base.SCIMBasePath + "/" + resType + "/" + url.PathEscape(id),
		PrettyJSON:  string(pretty),
		Groups:      badges,
	})
}

func (h *uiController) playgroundPage(c *gin.Context) {
	data := h.evalPlayground(c, "Users", "active eq true", "", "", "ascending")
	h.render(c, http.StatusOK, "playground.html", data)
}

func (h *uiController) playgroundEvalPartial(c *gin.Context) {
	resType := normalizeUIResourceTab(c.PostForm("resourceType"))
	filter := strings.TrimSpace(c.PostForm("filter"))
	attrs := strings.TrimSpace(c.PostForm("attributes"))
	sortBy := strings.TrimSpace(c.PostForm("sortBy"))
	sortOrder := strings.TrimSpace(c.PostForm("sortOrder"))
	data := h.evalPlayground(c, resType, filter, attrs, sortBy, sortOrder)
	h.render(c, http.StatusOK, "playground_result", data)
}

func (h *uiController) evalPlayground(c *gin.Context, resType, filter, attrs, sortBy, sortOrder string) uiPlaygroundData {
	base := h.pageBase(c, "Filter & N1QL Lab", "playground")
	out := uiPlaygroundData{
		uiPageBase:   base,
		ResourceType: resType,
		Filter:       filter,
	}
	rt, ok := Resources["/"+resType]
	if !ok {
		out.Error = "Unknown resource type " + resType
		return out
	}

	query, err := parser.FilterToN1QLWithOptions(rt.Name, filter, parser.FilterOptions{
		Resolve: func(p string) string {
			return CanonicalAttributePath(rt, p)
		},
		IsCaseExact: func(p string) bool {
			return AttributeIsCaseExact(rt, p)
		},
		IsMultiValued: func(p string) bool {
			return AttributeIsMultiValuedComplex(rt, p)
		},
	})
	if err != nil {
		out.Error = err.Error()
		return out
	}
	applyN1QLTenantFilter(&query, strings.TrimSpace(filter) != "", base.Tenant)

	normSort, err := NormalizeSortBy(rt, sortBy)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	orderField := "id"
	if normSort != "" {
		parts := make([]string, 0)
		for _, s := range strings.Split(normSort, ",") {
			parts = append(parts, parser.AddQuote(s))
		}
		orderField = strings.Join(parts, ",")
	}
	dir := "ASC"
	if strings.EqualFold(sortOrder, "descending") {
		dir = "DESC"
	}
	out.N1QLStatement = query.Page + "\nORDER BY " + orderField + " " + dir + "\nLIMIT 50 OFFSET 0"
	if len(query.Params) > 0 {
		paramMap := make(map[string]interface{}, len(query.Params))
		for i, val := range query.Params {
			paramMap[fmt.Sprintf("$%d", i+1)] = val
		}
		rawParams, _ := json.MarshalIndent(paramMap, "", "  ")
		out.N1QLParamsJSON = string(rawParams)
	}

	proj, err := ParseProjectionParams(splitAndTrim(attrs), nil)
	if err != nil {
		out.Error = err.Error()
		return out
	}

	if DB != nil {
		total, docs, err := DB.Search(SearchQuery{
			Bucket: ScopedBucket(base.Tenant, rt.Name),
			Filter: filter,
			ResolveName: func(p string) string {
				return CanonicalAttributePath(rt, p)
			},
			IsCaseExact: func(p string) bool {
				return AttributeIsCaseExact(rt, p)
			},
			IsMultiValued: func(p string) bool {
				return AttributeIsMultiValuedComplex(rt, p)
			},
			SortBy:         normSort,
			SortDescending: strings.EqualFold(sortOrder, "descending"),
			Limit:          50,
		})
		if err != nil {
			out.Error = err.Error()
			return out
		}
		out.TotalResults = total
		roles := []string{"admin", "superadmin", "user", "role1", "role2"}
		items := make([]interface{}, 0, len(docs))
		for _, doc := range docs {
			items = append(items, FormatResource(roles, rt, doc, proj))
		}
		respObj := ListResponse{
			Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
			TotalResults: total,
			ItemsPerPage: len(items),
			StartIndex:   1,
			Resources:    items,
		}
		prettyResp, _ := json.MarshalIndent(respObj, "", "  ")
		out.ResponseJSON = string(prettyResp)
	}
	return out
}

func (h *uiController) inspectorPage(c *gin.Context) {
	base := h.pageBase(c, "Live Inspector", "inspector")
	h.render(c, http.StatusOK, "inspector.html", uiInspectorData{
		uiPageBase: base,
		Exchanges:  DefaultInspector.Recent(100),
	})
}

func (h *uiController) inspectorRowsPartial(c *gin.Context) {
	base := h.pageBase(c, "Live Inspector", "inspector")
	h.render(c, http.StatusOK, "inspector_rows", uiInspectorData{
		uiPageBase: base,
		Exchanges:  DefaultInspector.Recent(100),
	})
}

func (h *uiController) inspectorDetailPartial(c *gin.Context) {
	seq, err := strconv.ParseUint(strings.TrimSpace(c.Query("seq")), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	ex, ok := DefaultInspector.Lookup(seq)
	if !ok {
		c.String(http.StatusNotFound, "Exchange is no longer in the ring buffer")
		return
	}
	h.render(c, http.StatusOK, "inspector_detail", uiInspectorDetailData{
		Seq:             ex.Seq,
		Timestamp:       ex.Timestamp.Format(time.RFC3339Nano),
		RequestID:       ex.RequestID,
		Tenant:          ex.Tenant,
		Subject:         ex.Subject,
		Method:          ex.Method,
		Path:            ex.Path,
		Action:          ex.Action,
		ResourceType:    ex.ResourceType,
		ResourceID:      ex.ResourceID,
		Status:          ex.Status,
		ScimType:        ex.ScimType,
		DurationMs:      ex.DurationMs,
		RequestPreview:  ex.RequestPreview,
		ResponsePreview: ex.ResponsePreview,
	})
}

func (h *uiController) inspectorClearAction(c *gin.Context) {
	DefaultInspector.Clear()
	h.inspectorRowsPartial(c)
}

func (h *uiController) conformancePage(c *gin.Context) {
	base := h.pageBase(c, "Webhooks & Conformance", "conformance")
	var urls []string
	var deliveries []WebhookDelivery
	if d := ActiveWebhookDispatcher(); d != nil {
		urls = d.URLs()
		deliveries = d.RecentDeliveries(50)
	}
	h.render(c, http.StatusOK, "conformance.html", uiConformanceData{
		uiPageBase:  base,
		WebhookURLs: urls,
		Deliveries:  deliveries,
	})
}

type inProcessRoundTripper struct {
	handler http.Handler
}

func (rt inProcessRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	internalReq := req.WithContext(withInternalUIPrincipal(req.Context()))
	rec := httptest.NewRecorder()
	rt.handler.ServeHTTP(rec, internalReq)
	res := rec.Result()
	if res.Body == nil {
		res.Body = io.NopCloser(bytes.NewReader(nil))
	}
	return res, nil
}

func (h *uiController) conformanceRunPartial(c *gin.Context) {
	base := h.pageBase(c, "Webhooks & Conformance", "conformance")
	profile := strings.TrimSpace(c.PostForm("profile"))
	if profile == "" {
		profile = "all"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	report := RunConformance(ctx, ConformanceOptions{
		BaseURL: "http://goscim.internal" + base.SCIMBasePath,
		Profile: profile,
		HTTPClient: &http.Client{
			Transport: inProcessRoundTripper{handler: h.engine},
			Timeout:   10 * time.Second,
		},
	})
	// Display the actual external SCIM Base URL in the UI report header.
	report.TargetURL = base.SCIMBaseURL

	h.render(c, http.StatusOK, "conformance_report", uiConformanceData{
		uiPageBase: base,
		Report:     report,
	})
}

func (h *uiController) render(c *gin.Context, status int, templateName string, data interface{}) {
	var buf bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		c.String(http.StatusInternalServerError, "UI template error: %v", err)
		return
	}
	c.Data(status, "text/html; charset=utf-8", buf.Bytes())
}
