package scim

import (
	"encoding/json"
	"net/http"
	"testing"
)

const (
	usersEndpoint  = "/scim/v2/Users"
	groupsEndpoint = "/scim/v2/Groups"
	groupSchemaURI = "urn:ietf:params:scim:schemas:core:2.0:Group"
)

func postUserForSync(t *testing.T, r http.Handler, userName string) string {
	t.Helper()
	payload, _ := json.Marshal(map[string]interface{}{
		"schemas":  []string{coreUserSchemaURI},
		"userName": userName,
	})
	w := do(t, r, http.MethodPost, usersEndpoint, string(payload))
	requireStatus(t, w, http.StatusCreated)
	return decode(t, w)["id"].(string)
}

func postGroupForSync(t *testing.T, r http.Handler, displayName string, members []map[string]interface{}) string {
	t.Helper()
	body := map[string]interface{}{
		"schemas":     []string{groupSchemaURI},
		"displayName": displayName,
	}
	if members != nil {
		body["members"] = members
	}
	payload, _ := json.Marshal(body)
	w := do(t, r, http.MethodPost, groupsEndpoint, string(payload))
	requireStatus(t, w, http.StatusCreated)
	return decode(t, w)["id"].(string)
}

func readUserGroupsMap(t *testing.T, r http.Handler, userID string) map[string]map[string]interface{} {
	t.Helper()
	w := do(t, r, http.MethodGet, usersEndpoint+"/"+userID, "")
	requireStatus(t, w, http.StatusOK)
	user := decode(t, w)
	out := make(map[string]map[string]interface{})
	rawGroups, _ := user["groups"].([]interface{})
	for _, raw := range rawGroups {
		g, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if val, _ := g["value"].(string); val != "" {
			out[val] = g
		}
	}
	return out
}

func TestGroupMembershipSyncDirectLifecycle(t *testing.T) {
	r, _ := newTestServer(t)

	u1 := postUserForSync(t, r, "ada@example.com")
	u2 := postUserForSync(t, r, "grace@example.com")
	u3 := postUserForSync(t, r, "linus@example.com")

	gID := postGroupForSync(t, r, "Platform", []map[string]interface{}{
		{"value": u1, "type": "User", "display": "Ada Lovelace"},
		{"value": u2, "$ref": PREFIX + "/Users/" + u2, "display": "Grace Hopper"},
	})

	u1Groups := readUserGroupsMap(t, r, u1)
	if len(u1Groups) != 1 {
		t.Fatalf("u1 groups = %v, want 1 group", u1Groups)
	}
	entry := u1Groups[gID]
	if entry["display"] != "Platform" || entry["type"] != "direct" || entry["$ref"] != PREFIX+"/Groups/"+gID {
		t.Fatalf("unexpected u1 group entry: %#v", entry)
	}
	if len(readUserGroupsMap(t, r, u3)) != 0 {
		t.Fatalf("u3 should have no groups")
	}

	// PATCH Group: rename displayName, add u3, remove u1 via value filter.
	patchBody, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{"op": "replace", "path": "displayName", "value": "Core Platform"},
			{"op": "add", "path": "members", "value": []map[string]interface{}{{"value": u3, "type": "User"}}},
			{"op": "remove", "path": `members[value eq "` + u1 + `"]`},
		},
	})
	w := do(t, r, http.MethodPatch, groupsEndpoint+"/"+gID, string(patchBody))
	requireStatus(t, w, http.StatusOK)

	if len(readUserGroupsMap(t, r, u1)) != 0 {
		t.Fatalf("u1 should have been removed from Platform")
	}
	if got := readUserGroupsMap(t, r, u2)[gID]; got["display"] != "Core Platform" {
		t.Fatalf("u2 group display did not update: %#v", got)
	}
	if got := readUserGroupsMap(t, r, u3)[gID]; got["display"] != "Core Platform" || got["type"] != "direct" {
		t.Fatalf("u3 did not receive Core Platform membership: %#v", got)
	}

	// PUT on User u2 cannot wipe or forge server-owned readOnly groups.
	putUserBody, _ := json.Marshal(map[string]interface{}{
		"schemas":     []string{coreUserSchemaURI},
		"userName":    "grace@example.com",
		"displayName": "Rear Admiral Grace Hopper",
		"groups":      []map[string]interface{}{{"value": "forged"}},
	})
	w = do(t, r, http.MethodPut, usersEndpoint+"/"+u2, string(putUserBody))
	requireStatus(t, w, http.StatusOK)
	u2AfterPut := readUserGroupsMap(t, r, u2)
	if len(u2AfterPut) != 1 || u2AfterPut[gID]["display"] != "Core Platform" {
		t.Fatalf("User.groups corrupted after PUT /Users: %#v", u2AfterPut)
	}

	// PUT on Group replacing members with only u1 updates u1, u2, and u3.
	putGroupBody, _ := json.Marshal(map[string]interface{}{
		"schemas":     []string{groupSchemaURI},
		"displayName": "Core Platform",
		"members":     []map[string]interface{}{{"value": u1}},
	})
	w = do(t, r, http.MethodPut, groupsEndpoint+"/"+gID, string(putGroupBody))
	requireStatus(t, w, http.StatusOK)

	if len(readUserGroupsMap(t, r, u1)) != 1 {
		t.Fatalf("u1 should be back in group after PUT /Groups")
	}
	if len(readUserGroupsMap(t, r, u2)) != 0 || len(readUserGroupsMap(t, r, u3)) != 0 {
		t.Fatalf("u2 and u3 should have no groups after PUT /Groups")
	}
}

func TestNestedGroupsIndirectAndCycleSafety(t *testing.T) {
	r, _ := newTestServer(t)

	u1 := postUserForSync(t, r, "nested@example.com")

	gBackend := postGroupForSync(t, r, "Backend", []map[string]interface{}{
		{"value": u1, "type": "User"},
	})
	gEng := postGroupForSync(t, r, "Engineering", []map[string]interface{}{
		{"value": gBackend, "type": "Group"},
	})
	gAll := postGroupForSync(t, r, "AllStaff", []map[string]interface{}{
		{"value": gEng, "type": "Group"},
	})

	u1Groups := readUserGroupsMap(t, r, u1)
	if len(u1Groups) != 3 {
		t.Fatalf("expected 3 memberships (1 direct, 2 indirect), got %#v", u1Groups)
	}
	if u1Groups[gBackend]["type"] != "direct" {
		t.Errorf("Backend type = %v, want direct", u1Groups[gBackend]["type"])
	}
	if u1Groups[gEng]["type"] != "indirect" {
		t.Errorf("Engineering type = %v, want indirect", u1Groups[gEng]["type"])
	}
	if u1Groups[gAll]["type"] != "indirect" {
		t.Errorf("AllStaff type = %v, want indirect", u1Groups[gAll]["type"])
	}

	// Adding u1 directly to Engineering promotes Engineering from indirect to direct
	// without duplicating the entry, and introducing a cycle Backend -> AllStaff
	// must terminate cleanly without infinite recursion.
	patchEng, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{"op": "add", "path": "members", "value": []map[string]interface{}{{"value": u1, "type": "User"}}},
		},
	})
	requireStatus(t, do(t, r, http.MethodPatch, groupsEndpoint+"/"+gEng, string(patchEng)), http.StatusOK)

	patchCycle, _ := json.Marshal(map[string]interface{}{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:PatchOp"},
		"Operations": []map[string]interface{}{
			{"op": "add", "path": "members", "value": []map[string]interface{}{{"value": gAll, "type": "Group"}}},
		},
	})
	requireStatus(t, do(t, r, http.MethodPatch, groupsEndpoint+"/"+gBackend, string(patchCycle)), http.StatusOK)

	u1Groups = readUserGroupsMap(t, r, u1)
	if len(u1Groups) != 3 {
		t.Fatalf("expected 3 deduplicated groups even with cycle, got %#v", u1Groups)
	}
	if u1Groups[gEng]["type"] != "direct" {
		t.Errorf("Engineering should be promoted to direct, got %v", u1Groups[gEng]["type"])
	}
	if u1Groups[gAll]["type"] != "indirect" {
		t.Errorf("AllStaff should remain indirect, got %v", u1Groups[gAll]["type"])
	}

	// Deleting Engineering unlinks it from AllStaff and breaks the transitive chain
	// from Backend to AllStaff, leaving u1 only in Backend.
	requireStatus(t, do(t, r, http.MethodDelete, groupsEndpoint+"/"+gEng, ""), http.StatusNoContent)

	u1Groups = readUserGroupsMap(t, r, u1)
	if len(u1Groups) != 1 || u1Groups[gBackend]["type"] != "direct" {
		t.Fatalf("after deleting middle group Engineering, u1 groups = %#v, want only Backend", u1Groups)
	}

	wAll := do(t, r, http.MethodGet, groupsEndpoint+"/"+gAll, "")
	requireStatus(t, wAll, http.StatusOK)
	if _, hasMembers := decode(t, wAll)["members"]; hasMembers {
		t.Errorf("AllStaff should have unlinked deleted Engineering member: %s", wAll.Body.String())
	}
}

func TestDeleteUserCleansUpGroupMembers(t *testing.T) {
	r, _ := newTestServer(t)

	u1 := postUserForSync(t, r, "temp1@example.com")
	u2 := postUserForSync(t, r, "temp2@example.com")

	gID := postGroupForSync(t, r, "Ops", []map[string]interface{}{
		{"value": u1, "type": "User"},
		{"value": u2, "type": "User"},
	})

	requireStatus(t, do(t, r, http.MethodDelete, usersEndpoint+"/"+u1, ""), http.StatusNoContent)

	w := do(t, r, http.MethodGet, groupsEndpoint+"/"+gID, "")
	requireStatus(t, w, http.StatusOK)
	members, _ := decode(t, w)["members"].([]interface{})
	if len(members) != 1 || members[0].(map[string]interface{})["value"] != u2 {
		t.Fatalf("expected only u2 remaining in Ops.members, got %#v", members)
	}

	// Deleting the last member removes the empty members array altogether.
	requireStatus(t, do(t, r, http.MethodDelete, usersEndpoint+"/"+u2, ""), http.StatusNoContent)
	w = do(t, r, http.MethodGet, groupsEndpoint+"/"+gID, "")
	requireStatus(t, w, http.StatusOK)
	if _, present := decode(t, w)["members"]; present {
		t.Fatalf("expected empty members array to be removed: %s", w.Body.String())
	}
}
