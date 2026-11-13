package scim

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the correlation header accepted on incoming requests and
// echoed on every response.
const RequestIDHeader = "X-Request-Id"

const (
	requestIDContextKey = "scim.request_id"
	auditInfoContextKey = "scim.audit_info"
	maxRequestIDLen     = 128
)

// AuditMode selects the output format for SCIM_AUDIT_LOG.
type AuditMode string

const (
	AuditModeJSON AuditMode = "json"
	AuditModeText AuditMode = "text"
	AuditModeOff  AuditMode = "off"
)

type auditInfo struct {
	action       string
	resourceType string
	resourceID   string
	scimType     string
}

var activeAuditLogger atomic.Pointer[slog.Logger]

func init() {
	activeAuditLogger.Store(newAuditLogger(AuditModeJSON, os.Stdout))
}

// SetAuditLogger installs logger as the active audit sink. Passing nil disables
// audit output until another logger is installed.
func SetAuditLogger(logger *slog.Logger) {
	if logger == nil {
		logger = newAuditLogger(AuditModeOff, io.Discard)
	}
	activeAuditLogger.Store(logger)
}

// AuditLogger returns the active structured audit logger.
func AuditLogger() *slog.Logger {
	if l := activeAuditLogger.Load(); l != nil {
		return l
	}
	return newAuditLogger(AuditModeOff, io.Discard)
}

// AuditModeFromEnv parses and validates SCIM_AUDIT_LOG, defaulting to json.
func AuditModeFromEnv() (AuditMode, error) {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_AUDIT_LOG")))
	switch AuditMode(raw) {
	case "", AuditModeJSON:
		return AuditModeJSON, nil
	case AuditModeText:
		return AuditModeText, nil
	case AuditModeOff, "none", "false":
		return AuditModeOff, nil
	default:
		return "", fmt.Errorf("unknown SCIM_AUDIT_LOG=%q (expected json, text, or off)", raw)
	}
}

// ConfigureAuditLoggerFromEnv builds and installs the audit logger selected by
// SCIM_AUDIT_LOG, writing to out (os.Stdout when nil).
func ConfigureAuditLoggerFromEnv(out io.Writer) (*slog.Logger, error) {
	mode, err := AuditModeFromEnv()
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = os.Stdout
	}
	logger := newAuditLogger(mode, out)
	SetAuditLogger(logger)
	return logger, nil
}

func newAuditLogger(mode AuditMode, out io.Writer) *slog.Logger {
	switch mode {
	case AuditModeText:
		return slog.New(slog.NewTextHandler(out, nil))
	case AuditModeOff:
		return slog.New(slog.NewJSONHandler(io.Discard, nil))
	default:
		return slog.New(slog.NewJSONHandler(out, nil))
	}
}

// RequestID attaches a validated correlation identifier to the request context
// and writes it back on the X-Request-Id response header. Untrusted values
// carrying whitespace, control characters, or unexpected symbols are replaced
// with a fresh server-generated UUID so callers cannot inject log delimiters.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader(RequestIDHeader))
		if !isSafeRequestID(id) {
			id = uuid.NewString()
		}
		c.Set(requestIDContextKey, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// RequestIDOf returns the correlation ID assigned to c, or "" if RequestID
// middleware did not run.
func RequestIDOf(c *gin.Context) string {
	if c == nil {
		return ""
	}
	val, ok := c.Get(requestIDContextKey)
	if !ok {
		return ""
	}
	id, _ := val.(string)
	return id
}

func isSafeRequestID(id string) bool {
	if len(id) == 0 || len(id) > maxRequestIDLen {
		return false
	}
	for i := 0; i < len(id); i++ {
		b := id[i]
		isAlphaNum := (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
		isAllowedPunct := b == '-' || b == '_' || b == '.' || b == ':'
		if !isAlphaNum && !isAllowedPunct {
			return false
		}
	}
	return true
}

func ensureAuditInfo(c *gin.Context) *auditInfo {
	if val, ok := c.Get(auditInfoContextKey); ok {
		if info, ok := val.(*auditInfo); ok {
			return info
		}
	}
	info := &auditInfo{}
	c.Set(auditInfoContextKey, info)
	return info
}

func setAuditTarget(c *gin.Context, action, resourceType, resourceID string) {
	if c == nil {
		return
	}
	info := ensureAuditInfo(c)
	if action != "" {
		info.action = action
	}
	if resourceType != "" {
		info.resourceType = resourceType
	}
	if resourceID != "" {
		info.resourceID = resourceID
	}
}

func setAuditScimType(c *gin.Context, scimType string) {
	if c == nil || scimType == "" {
		return
	}
	ensureAuditInfo(c).scimType = scimType
}

// AuditMiddleware emits one structured audit record per SCIM API request once
// the handler chain completes. Query parameters, headers, and request/response
// bodies are never logged so credentials, passwords, and filter values cannot
// leak into log aggregation pipelines.
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ensureAuditInfo(c)

		c.Next()

		status := c.Writer.Status()
		durationMs := float64(time.Since(start).Microseconds()) / 1000.0
		info := ensureAuditInfo(c)

		action := info.action
		resourceType := info.resourceType
		resourceID := info.resourceID
		if action == "" {
			action, resourceType, resourceID = inferAuditTarget(c.Request.Method, c.Request.URL.Path)
		}

		attrs := make([]slog.Attr, 0, 14)
		if reqID := RequestIDOf(c); reqID != "" {
			attrs = append(attrs, slog.String("request_id", reqID))
		}
		attrs = append(attrs, slog.String("action", action))
		if principal := PrincipalOf(c); principal != nil {
			attrs = append(attrs, slog.String("subject", principal.Subject))
			if len(principal.Roles) > 0 {
				attrs = append(attrs, slog.Any("roles", principal.Roles))
			}
		}
		if resourceType != "" {
			attrs = append(attrs, slog.String("resource_type", resourceType))
		}
		if resourceID != "" {
			attrs = append(attrs, slog.String("resource_id", resourceID))
		}
		attrs = append(attrs,
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.String("outcome", classifyOutcome(status)),
		)
		if info.scimType != "" {
			attrs = append(attrs, slog.String("scim_type", info.scimType))
		}
		if ip := c.ClientIP(); ip != "" {
			attrs = append(attrs, slog.String("client_ip", ip))
		}
		attrs = append(attrs, slog.Float64("duration_ms", durationMs))

		level := slog.LevelInfo
		if status == http.StatusUnauthorized || status == http.StatusForbidden || status >= http.StatusInternalServerError {
			level = slog.LevelWarn
		}
		AuditLogger().LogAttrs(c.Request.Context(), level, "scim.audit", attrs...)
	}
}

func classifyOutcome(status int) string {
	switch {
	case status < http.StatusBadRequest:
		return "success"
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return "denied"
	case status == http.StatusConflict || status == http.StatusPreconditionFailed:
		return "conflict"
	case status < http.StatusInternalServerError:
		return "client_error"
	default:
		return "server_error"
	}
}

// inferAuditTarget fills in action/resource metadata when middleware such as
// Authenticate or LimitBody aborts the request before a route handler runs.
func inferAuditTarget(method, urlPath string) (action, resourceType, resourceID string) {
	trimmed := strings.TrimPrefix(urlPath, PREFIX)
	if trimmed == "/.search" && method == http.MethodPost {
		return "search", "*", ""
	}
	for endpoint, rt := range Resources {
		if trimmed == endpoint+"/.search" && method == http.MethodPost {
			return "search", rt.Name, ""
		}
		if trimmed == endpoint {
			switch method {
			case http.MethodPost:
				return "create", rt.Name, ""
			case http.MethodGet:
				return "search", rt.Name, ""
			}
		}
		if strings.HasPrefix(trimmed, endpoint+"/") {
			id := strings.TrimPrefix(trimmed, endpoint+"/")
			if id != "" && !strings.Contains(id, "/") {
				switch method {
				case http.MethodGet:
					return "read", rt.Name, id
				case http.MethodPut:
					return "replace", rt.Name, id
				case http.MethodPatch:
					return "patch", rt.Name, id
				case http.MethodDelete:
					return "delete", rt.Name, id
				}
			}
		}
	}
	return strings.ToLower(method), "", ""
}
