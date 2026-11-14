package scim

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestEntraIDSecretBearerAuthentication exercises Microsoft Entra ID's "Secret
// Token" provisioning credential mode (SCIM_AUTH=bearer).
func TestEntraIDSecretBearerAuthentication(t *testing.T) {
	t.Setenv("SCIM_AUTH", "bearer")
	t.Setenv("SCIM_BEARER_TOKEN", "entra-secret-provisioning-token-2026")
	t.Setenv("SCIM_BEARER_SUBJECT", "entra-provisioning")
	t.Setenv("SCIM_BEARER_ROLES", "admin,role1,role2")

	auth, err := NewAuthenticatorFromEnv()
	if err != nil {
		t.Fatalf("NewAuthenticatorFromEnv(bearer): %v", err)
	}
	r, _ := newTestServerAs(t, auth)

	// Discovery stays unauthenticated and advertises oauthbearertoken.
	w := do(t, r, http.MethodGet, "/scim/v2/ServiceProviderConfig", "")
	requireStatus(t, w, http.StatusOK)
	spc := decode(t, w)
	schemes, _ := spc["authenticationSchemes"].([]interface{})
	if len(schemes) != 1 || schemes[0].(map[string]interface{})["type"] != "oauthbearertoken" {
		t.Fatalf("authenticationSchemes = %v, want oauthbearertoken", schemes)
	}

	// Missing or invalid bearer token returns 401 with WWW-Authenticate.
	w = do(t, r, http.MethodGet, usersPath, "")
	requireSCIMError(t, w, http.StatusUnauthorized)
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate challenge on 401")
	}

	badReq := newRequest(http.MethodGet, usersPath, "")
	badReq.Header.Set("Authorization", "Bearer wrong-token")
	requireSCIMError(t, serve(r, badReq), http.StatusUnauthorized)

	// Valid Secret Token grants access and passes roles to RBAC filters.
	okReq := newRequest(http.MethodPost, elementsPath, validElement("EntraElem", 1))
	okReq.Header.Set("Authorization", "Bearer entra-secret-provisioning-token-2026")
	w = serve(r, okReq)
	requireStatus(t, w, http.StatusCreated)
	if decode(t, w)["description"] != "description of EntraElem" {
		t.Errorf("expected role2 from SCIM_BEARER_ROLES to expose description")
	}
}

// TestEntraIDv2OAuthJWTAuthentication exercises Microsoft Entra ID v2.0 OAuth2
// access tokens carrying "scp" (delegated scopes), "roles" (app roles), and
// "oid"/"azp" principal identifiers.
func TestEntraIDv2OAuthJWTAuthentication(t *testing.T) {
	issuer := newTestIssuer(t)
	tenantIssuer := "https://login.microsoftonline.com/11111111-2222-3333-4444-555555555555/v2.0"
	auth, err := NewJWTAuthenticator(issuer.server.URL, tenantIssuer, "api://goscim-prod", "")
	if err != nil {
		t.Fatalf("NewJWTAuthenticator: %v", err)
	}
	r, _ := newTestServerAs(t, auth)

	claims := jwt.MapClaims{
		"iss":   tenantIssuer,
		"aud":   "api://goscim-prod",
		"oid":   "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		"azp":   "99999999-8888-7777-6666-555555555555",
		"scp":   "role1 role2",
		"roles": []interface{}{"admin"},
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	req := newRequest(http.MethodPost, elementsPath, validElement("EntraJWTElem", 1))
	req.Header.Set("Authorization", "Bearer "+issuer.sign(t, claims))
	w := serve(r, req)
	requireStatus(t, w, http.StatusCreated)
	if decode(t, w)["description"] != "description of EntraJWTElem" {
		t.Errorf("expected Entra scp claim to grant role2 read access")
	}
}

// TestEntraIDFullProvisioningLifecycle reproduces the exact sequence of HTTP
// calls issued by the Microsoft Entra ID Provisioning Service and the Microsoft
// SCIM Validator (covering both ?aadOptscim062020 and legacy Entra quirks).
func TestEntraIDFullProvisioningLifecycle(t *testing.T) {
	auth, err := NewBearerAuthenticator([]BearerTokenEntry{{
		Subject: "entra-sync-engine",
		Token:   "entra-token",
		Roles:   testRoles,
	}})
	if err != nil {
		t.Fatal(err)
	}
	r, _ := newTestServerAs(t, auth)

	entraDo := func(method, target, body string) map[string]interface{} {
		t.Helper()
		req := newRequest(method, target, body)
		req.Header.Set("Authorization", "Bearer entra-token")
		w := serve(r, req)
		if method == http.MethodPost {
			requireStatus(t, w, http.StatusCreated)
		} else if method == http.MethodDelete {
			requireStatus(t, w, http.StatusNoContent)
			return nil
		} else {
			requireStatus(t, w, http.StatusOK)
		}
		return decode(t, w)
	}

	// 1. Create two managers first so EnterpriseUser.manager can reference them.
	mgr1 := entraDo(http.MethodPost, usersPath, `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
		"externalId": "oid-mgr-1",
		"userName": "MeganB@contoso.com",
		"active": true
	}`)
	mgr1ID := mgr1["id"].(string)

	mgr2 := entraDo(http.MethodPost, usersPath, `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
		"externalId": "oid-mgr-2",
		"userName": "AlexW@contoso.com",
		"active": true
	}`)
	mgr2ID := mgr2["id"].(string)

	// 2. Entra probes existence by userName before creating Adele.
	probe := entraDo(http.MethodGet, usersPath+"?filter="+url.QueryEscape(`userName eq "AdeleV@contoso.com"`), "")
	if probe["totalResults"] != float64(0) {
		t.Fatalf("expected 0 results before create, got %v", probe["totalResults"])
	}

	// 3. Entra creates Adele with externalId, addresses[].primary, and EnterpriseUser.
	adele := entraDo(http.MethodPost, usersPath, `{
		"schemas": [
			"urn:ietf:params:scim:schemas:core:2.0:User",
			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
		],
		"externalId": "entra-oid-adele-001",
		"userName": "AdeleV@contoso.com",
		"active": true,
		"displayName": "Adele Vance",
		"title": "Product Manager",
		"preferredLanguage": "en-US",
		"name": {
			"formatted": "Adele Vance",
			"givenName": "Adele",
			"familyName": "Vance"
		},
		"emails": [
			{"primary": true, "type": "work", "value": "adelev@contoso.com"}
		],
		"addresses": [
			{
				"type": "work",
				"formatted": "One Microsoft Way",
				"streetAddress": "One Microsoft Way",
				"locality": "Redmond",
				"region": "WA",
				"postalCode": "98052",
				"country": "USA",
				"primary": true
			}
		],
		"phoneNumbers": [
			{"primary": true, "type": "work", "value": "+1-425-555-0100"}
		],
		"roles": [
			{"primary": false, "type": "WindowsAzureActiveDirectoryRole", "value": "User"}
		],
		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
			"employeeNumber": "EMP-1001",
			"department": "Retail",
			"costCenter": "CC-100",
			"organization": "Contoso",
			"division": "North America",
			"manager": {"value": "`+mgr1ID+`"}
		}
	}`)
	adeleID := adele["id"].(string)
	if adele["externalId"] != "entra-oid-adele-001" {
		t.Fatalf("externalId = %v, want entra-oid-adele-001", adele["externalId"])
	}

	// 4. Case-insensitive userName filter and case-exact externalId filter.
	byLowerUser := entraDo(http.MethodGet, usersPath+"?filter="+url.QueryEscape(`userName eq "adelev@contoso.com"`), "")
	if byLowerUser["totalResults"] != float64(1) {
		t.Fatalf("case-insensitive userName filter returned %v, want 1", byLowerUser["totalResults"])
	}
	byExactExt := entraDo(http.MethodGet, usersPath+"?filter="+url.QueryEscape(`externalId eq "entra-oid-adele-001"`), "")
	if byExactExt["totalResults"] != float64(1) {
		t.Fatalf("exact externalId filter returned %v, want 1", byExactExt["totalResults"])
	}
	byWrongCaseExt := entraDo(http.MethodGet, usersPath+"?filter="+url.QueryEscape(`externalId eq "ENTRA-OID-ADELE-001"`), "")
	if byWrongCaseExt["totalResults"] != float64(0) {
		t.Fatalf("externalId must be caseExact:true per RFC 7643 3.1, got %v", byWrongCaseExt["totalResults"])
	}
	byEntEmp := entraDo(http.MethodGet, usersPath+"?filter="+url.QueryEscape(
		`urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber eq "emp-1001"`), "")
	if byEntEmp["totalResults"] != float64(1) {
		t.Fatalf("colon-delimited EnterpriseUser filter returned %v, want 1", byEntEmp["totalResults"])
	}

	// 5. Entra multi-op PATCH: bracketed value filters, pathless colon-URN keys,
	// legacy capitalized Ops ("Replace"/"Add"), legacy stringified boolean "False",
	// and legacy stringified manager ID.
	patched := entraDo(http.MethodPatch, usersPath+"/"+adeleID, `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations": [
			{"op": "Replace", "path": "emails[type eq \"work\"].value", "value": "adele.vance@contoso.com"},
			{"op": "Add", "path": "phoneNumbers[type eq \"mobile\"].value", "value": "+1-425-555-0199"},
			{"op": "Replace", "path": "addresses[type eq \"work\"].streetAddress", "value": "500 Fifth Ave"},
			{"op": "Add", "value": {
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department": "Cloud Engineering",
				"name.givenName": "Adele M."
			}},
			{"op": "Replace", "path": "active", "value": "False"},
			{"op": "Replace", "path": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager", "value": "`+mgr2ID+`"}
		]
	}`)
	if patched["active"] != false {
		t.Fatalf("active = %#v (%T), want bool false", patched["active"], patched["active"])
	}
	entMap := patched[schemaUserEnt].(map[string]interface{})
	if entMap["department"] != "Cloud Engineering" {
		t.Errorf("department = %v, want Cloud Engineering", entMap["department"])
	}
	mgrObj, ok := entMap["manager"].(map[string]interface{})
	if !ok || mgrObj["value"] != mgr2ID {
		t.Errorf("manager = %#v, want {value:%s}", entMap["manager"], mgr2ID)
	}
	if patched["name"].(map[string]interface{})["givenName"] != "Adele M." {
		t.Errorf("name.givenName = %v, want Adele M.", patched["name"])
	}

	// Clearing manager via empty string (legacy Entra behaviour).
	clearedMgr := entraDo(http.MethodPatch, usersPath+"/"+adeleID, `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations": [
			{"op": "Replace", "path": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager", "value": ""},
			{"op": "Replace", "path": "active", "value": "True"}
		]
	}`)
	if clearedMgr["active"] != true {
		t.Errorf("active = %#v, want true", clearedMgr["active"])
	}
	if _, stillHasMgr := clearedMgr[schemaUserEnt].(map[string]interface{})["manager"]; stillHasMgr {
		t.Errorf("expected manager to be cleared, got %v", clearedMgr[schemaUserEnt])
	}

	// 6. Entra Group lifecycle with ?excludedAttributes=members, PATCH add,
	// legacy PATCH remove with value array, and ?aadOptscim062020 bracket remove.
	grp := entraDo(http.MethodPost, "/scim/v2/Groups?excludedAttributes=members", `{
		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
		"externalId": "entra-grp-eng",
		"displayName": "Cloud Engineering"
	}`)
	grpID := grp["id"].(string)
	if grp["externalId"] != "entra-grp-eng" {
		t.Errorf("group externalId = %v, want entra-grp-eng", grp["externalId"])
	}

	// Add both Adele and Megan to the group.
	entraDo(http.MethodPatch, "/scim/v2/Groups/"+grpID+"?excludedAttributes=members", `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations": [{
			"op": "Add",
			"path": "members",
			"value": [{"value": "`+adeleID+`"}, {"value": "`+mgr1ID+`"}]
		}]
	}`)

	// Remove ONLY Adele using legacy Entra payload (path:"members", value:[{value:adeleID}]).
	entraDo(http.MethodPatch, "/scim/v2/Groups/"+grpID, `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations": [{
			"op": "Remove",
			"path": "members",
			"value": [{"value": "`+adeleID+`"}]
		}]
	}`)

	// Verify Megan is STILL a member and her User.groups still contains the group,
	// whereas Adele was removed from both Group.members and Adele's User.groups!
	grpState := entraDo(http.MethodGet, "/scim/v2/Groups/"+grpID, "")
	members, _ := grpState["members"].([]interface{})
	if len(members) != 1 || members[0].(map[string]interface{})["value"] != mgr1ID {
		t.Fatalf("legacy Entra member remove wiped non-targeted members: got %v", members)
	}
	adeleAfter := entraDo(http.MethodGet, usersPath+"/"+adeleID, "")
	if g, has := adeleAfter["groups"].([]interface{}); has && len(g) > 0 {
		t.Errorf("Adele User.groups should be empty after removal, got %v", g)
	}
	meganAfter := entraDo(http.MethodGet, usersPath+"/"+mgr1ID, "")
	meganGroups, _ := meganAfter["groups"].([]interface{})
	if len(meganGroups) != 1 || meganGroups[0].(map[string]interface{})["value"] != grpID {
		t.Errorf("Megan User.groups = %v, want group %s", meganGroups, grpID)
	}

	// Remove Megan using modern ?aadOptscim062020 bracket filter path.
	entraDo(http.MethodPatch, "/scim/v2/Groups/"+grpID, `{
		"schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
		"Operations": [{
			"op": "Remove",
			"path": "members[value eq \"`+mgr1ID+`\"]"
		}]
	}`)
	meganFinal := entraDo(http.MethodGet, usersPath+"/"+mgr1ID, "")
	if g, has := meganFinal["groups"].([]interface{}); has && len(g) > 0 {
		t.Errorf("Megan User.groups should be empty after bracket removal, got %v", g)
	}
}
