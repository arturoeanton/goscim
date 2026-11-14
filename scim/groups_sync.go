package scim

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"sort"
	"strings"
)

const (
	coreUserSchemaURI  = "urn:ietf:params:scim:schemas:core:2.0:User"
	coreGroupSchemaURI = "urn:ietf:params:scim:schemas:core:2.0:Group"
	groupScanPageSize  = 500
)

func findUserAndGroupResourceTypes() (userRT ResourceType, groupRT ResourceType, ok bool) {
	var hasUser, hasGroup bool
	for _, rt := range Resources {
		if strings.EqualFold(rt.Schema, coreUserSchemaURI) {
			userRT = rt
			hasUser = true
		} else if strings.EqualFold(rt.Schema, coreGroupSchemaURI) {
			groupRT = rt
			hasGroup = true
		}
	}
	if !hasUser || !hasGroup {
		return ResourceType{}, ResourceType{}, false
	}
	userSchema, okU := LookupSchema(userRT.Schema)
	groupSchema, okG := LookupSchema(groupRT.Schema)
	if !okU || !okG {
		return ResourceType{}, ResourceType{}, false
	}
	if _, hasGroupsAttr := FindAttribute(userSchema.Attributes, "groups"); !hasGroupsAttr {
		return ResourceType{}, ResourceType{}, false
	}
	if _, hasMembersAttr := FindAttribute(groupSchema.Attributes, "members"); !hasMembersAttr {
		return ResourceType{}, ResourceType{}, false
	}
	return userRT, groupRT, true
}

func isGroupResourceType(rt ResourceType) bool {
	return strings.EqualFold(rt.Schema, coreGroupSchemaURI)
}

func isUserResourceType(rt ResourceType) bool {
	return strings.EqualFold(rt.Schema, coreUserSchemaURI)
}

// syncAfterCreate propagates Group.members to User.groups when a Group with
// initial members is created.
func syncAfterCreate(rt ResourceType, id string, created map[string]interface{}) {
	syncAfterCreateInTenant("", PREFIX, rt, id, created)
}

func syncAfterCreateInTenant(tenant, prefix string, rt ResourceType, id string, created map[string]interface{}) {
	if !isGroupResourceType(rt) {
		return
	}
	members, _ := created["members"].([]interface{})
	if len(members) == 0 {
		return
	}
	userRT, groupRT, ok := findUserAndGroupResourceTypes()
	if !ok {
		return
	}
	groupBucket := ScopedBucket(tenant, groupRT.Name)
	userBucket := ScopedBucket(tenant, userRT.Name)

	allGroups, err := loadAllGroups(groupBucket)
	if err != nil {
		log.Printf("scim group sync (create %s): %v", id, err)
		return
	}
	groupsBefore := copyGroupMap(allGroups)
	delete(groupsBefore, id)

	groupsAfter := copyGroupMap(allGroups)
	if cloned, err := roundTripMap(created); err == nil {
		groupsAfter[id] = cloned
	}

	if err := applyUserGroupDiffs(
		userBucket,
		computeUserGroupsMapWithPrefix(prefix, groupsBefore, groupRT),
		computeUserGroupsMapWithPrefix(prefix, groupsAfter, groupRT),
	); err != nil {
		log.Printf("scim group sync (create %s): %v", id, err)
	}
}

// syncAfterReplace propagates changes in Group.members or Group.displayName to
// affected User.groups entries on PUT and PATCH.
func syncAfterReplace(rt ResourceType, id string, previous, updated map[string]interface{}) {
	syncAfterReplaceInTenant("", PREFIX, rt, id, previous, updated)
}

func syncAfterReplaceInTenant(tenant, prefix string, rt ResourceType, id string, previous, updated map[string]interface{}) {
	if !isGroupResourceType(rt) {
		return
	}
	if previous["displayName"] == updated["displayName"] && reflect.DeepEqual(previous["members"], updated["members"]) {
		return
	}
	userRT, groupRT, ok := findUserAndGroupResourceTypes()
	if !ok {
		return
	}
	groupBucket := ScopedBucket(tenant, groupRT.Name)
	userBucket := ScopedBucket(tenant, userRT.Name)

	allGroups, err := loadAllGroups(groupBucket)
	if err != nil {
		log.Printf("scim group sync (replace %s): %v", id, err)
		return
	}

	groupsBefore := copyGroupMap(allGroups)
	if clonedPrev, err := roundTripMap(previous); err == nil {
		groupsBefore[id] = clonedPrev
	}

	groupsAfter := copyGroupMap(allGroups)
	if clonedNext, err := roundTripMap(updated); err == nil {
		groupsAfter[id] = clonedNext
	}

	if err := applyUserGroupDiffs(
		userBucket,
		computeUserGroupsMapWithPrefix(prefix, groupsBefore, groupRT),
		computeUserGroupsMapWithPrefix(prefix, groupsAfter, groupRT),
	); err != nil {
		log.Printf("scim group sync (replace %s): %v", id, err)
	}
}

// syncAfterDelete cleans up dangling Group.members references when a User or
// Group is deleted and updates affected User.groups entries.
func syncAfterDelete(rt ResourceType, deletedID string, deletedDoc map[string]interface{}) {
	syncAfterDeleteInTenant("", PREFIX, rt, deletedID, deletedDoc)
}

func syncAfterDeleteInTenant(tenant, prefix string, rt ResourceType, deletedID string, deletedDoc map[string]interface{}) {
	userRT, groupRT, ok := findUserAndGroupResourceTypes()
	if !ok {
		return
	}
	groupBucket := ScopedBucket(tenant, groupRT.Name)
	userBucket := ScopedBucket(tenant, userRT.Name)

	switch {
	case isUserResourceType(rt):
		allGroups, err := loadAllGroups(groupBucket)
		if err != nil {
			log.Printf("scim group sync (delete user %s): %v", deletedID, err)
			return
		}
		for gID, gDoc := range allGroups {
			if removeMemberFromGroupDoc(prefix, gDoc, deletedID, groupRT) {
				if err := DB.Replace(groupBucket, gID, gDoc); err != nil && !errors.Is(err, ErrNotFound) {
					log.Printf("scim group sync (update group %s): %v", gID, err)
				}
			}
		}

	case isGroupResourceType(rt):
		remainingGroups, err := loadAllGroups(groupBucket)
		if err != nil {
			log.Printf("scim group sync (delete group %s): %v", deletedID, err)
			return
		}
		delete(remainingGroups, deletedID)

		groupsBefore := copyGroupMap(remainingGroups)
		if clonedDeleted, err := roundTripMap(deletedDoc); err == nil {
			groupsBefore[deletedID] = clonedDeleted
		}

		groupsAfter := copyGroupMap(remainingGroups)
		for gID, gDoc := range groupsAfter {
			if removeMemberFromGroupDoc(prefix, gDoc, deletedID, groupRT) {
				if err := DB.Replace(groupBucket, gID, gDoc); err != nil && !errors.Is(err, ErrNotFound) {
					log.Printf("scim group sync (unlink parent group %s): %v", gID, err)
				}
			}
		}

		if err := applyUserGroupDiffs(
			userBucket,
			computeUserGroupsMapWithPrefix(prefix, groupsBefore, groupRT),
			computeUserGroupsMapWithPrefix(prefix, groupsAfter, groupRT),
		); err != nil {
			log.Printf("scim group sync (delete group %s): %v", deletedID, err)
		}
	}
}

func removeMemberFromGroupDoc(prefix string, groupDoc map[string]interface{}, targetID string, groupRT ResourceType) bool {
	rawMembers, ok := groupDoc["members"].([]interface{})
	if !ok || len(rawMembers) == 0 {
		return false
	}
	kept := make([]interface{}, 0, len(rawMembers))
	removed := false
	for _, raw := range rawMembers {
		m, ok := raw.(map[string]interface{})
		if !ok {
			kept = append(kept, raw)
			continue
		}
		val, _ := m["value"].(string)
		if val == targetID {
			removed = true
			continue
		}
		kept = append(kept, raw)
	}
	if !removed {
		return false
	}
	if len(kept) == 0 {
		delete(groupDoc, "members")
	} else {
		groupDoc["members"] = kept
	}
	prevMeta, _ := groupDoc["meta"].(map[string]interface{})
	meta := updateMetaForPrefix(prefix, prevMeta, groupDoc, groupRT)
	if metaMap, err := structToMap(meta); err == nil {
		groupDoc["meta"] = metaMap
	}
	return true
}

func loadAllGroups(groupBucket string) (map[string]map[string]interface{}, error) {
	byID := make(map[string]map[string]interface{})
	offset := 0
	for {
		total, docs, err := DB.Search(SearchQuery{
			Bucket: groupBucket,
			Offset: offset,
			Limit:  groupScanPageSize,
		})
		if err != nil {
			return nil, err
		}
		for _, doc := range docs {
			if id, _ := doc["id"].(string); id != "" {
				byID[id] = doc
			}
		}
		offset += len(docs)
		if len(docs) == 0 || offset >= total {
			break
		}
	}
	return byID, nil
}

// computeUserGroupsMap derives the RFC 7643 4.1.2 readOnly User.groups array
// for every user referenced directly or transitively across groupsByID.
func computeUserGroupsMap(groupsByID map[string]map[string]interface{}, groupRT ResourceType) map[string][]interface{} {
	return computeUserGroupsMapWithPrefix(PREFIX, groupsByID, groupRT)
}

func computeUserGroupsMapWithPrefix(prefix string, groupsByID map[string]map[string]interface{}, groupRT ResourceType) map[string][]interface{} {
	directUsers := make(map[string][]string, len(groupsByID))
	parentGroups := make(map[string][]string)

	groupIDs := make([]string, 0, len(groupsByID))
	for gID := range groupsByID {
		groupIDs = append(groupIDs, gID)
	}
	sort.Strings(groupIDs)

	for _, gID := range groupIDs {
		gDoc := groupsByID[gID]
		rawMembers, _ := gDoc["members"].([]interface{})
		for _, raw := range rawMembers {
			m, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			memberID, _ := m["value"].(string)
			if memberID == "" {
				continue
			}
			if isMemberGroupRef(m, memberID, groupsByID, groupRT) {
				parentGroups[memberID] = append(parentGroups[memberID], gID)
			} else {
				directUsers[gID] = append(directUsers[gID], memberID)
			}
		}
	}

	// userID -> groupID -> "direct" | "indirect"
	memberships := make(map[string]map[string]string)
	ensureUser := func(uID string) map[string]string {
		m, ok := memberships[uID]
		if !ok {
			m = make(map[string]string)
			memberships[uID] = m
		}
		return m
	}

	for _, gID := range groupIDs {
		users := directUsers[gID]
		if len(users) == 0 {
			continue
		}
		// Find all ancestor groups containing gID (cycle-safe BFS).
		ancestors := collectAncestorGroups(gID, parentGroups, groupsByID)
		for _, uID := range users {
			um := ensureUser(uID)
			um[gID] = "direct"
			for _, ancID := range ancestors {
				if um[ancID] != "direct" {
					um[ancID] = "indirect"
				}
			}
		}
	}

	result := make(map[string][]interface{}, len(memberships))
	for uID, groupMap := range memberships {
		ids := make([]string, 0, len(groupMap))
		for gID := range groupMap {
			ids = append(ids, gID)
		}
		sort.Strings(ids)

		entries := make([]interface{}, 0, len(ids))
		for _, gID := range ids {
			gDoc, exists := groupsByID[gID]
			if !exists {
				continue
			}
			entry := map[string]interface{}{
				"value": gID,
				"$ref":  resourcePathWithPrefix(prefix, groupRT, gID),
				"type":  groupMap[gID],
			}
			if dn, ok := gDoc["displayName"].(string); ok && dn != "" {
				entry["display"] = dn
			}
			entries = append(entries, entry)
		}
		if len(entries) > 0 {
			result[uID] = entries
		}
	}
	return result
}

func isMemberGroupRef(member map[string]interface{}, memberID string, groupsByID map[string]map[string]interface{}, groupRT ResourceType) bool {
	mType, _ := member["type"].(string)
	switch {
	case strings.EqualFold(mType, "Group"):
		return true
	case strings.EqualFold(mType, "User"):
		return false
	}
	if mRef, _ := member["$ref"].(string); mRef != "" && strings.Contains(mRef, groupRT.Endpoint+"/") {
		return true
	}
	_, existsInGroups := groupsByID[memberID]
	return existsInGroups
}

func collectAncestorGroups(startGroupID string, parentGroups map[string][]string, groupsByID map[string]map[string]interface{}) []string {
	visited := map[string]bool{startGroupID: true}
	queue := append([]string(nil), parentGroups[startGroupID]...)
	var ancestors []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if visited[curr] {
			continue
		}
		visited[curr] = true
		if _, exists := groupsByID[curr]; !exists {
			continue
		}
		ancestors = append(ancestors, curr)
		queue = append(queue, parentGroups[curr]...)
	}
	sort.Strings(ancestors)
	return ancestors
}

func applyUserGroupDiffs(userBucket string, before, after map[string][]interface{}) error {
	affected := make(map[string]struct{}, len(before)+len(after))
	for uID := range before {
		affected[uID] = struct{}{}
	}
	for uID := range after {
		affected[uID] = struct{}{}
	}

	userIDs := make([]string, 0, len(affected))
	for uID := range affected {
		userIDs = append(userIDs, uID)
	}
	sort.Strings(userIDs)

	var firstErr error
	for _, uID := range userIDs {
		oldList := before[uID]
		newList := after[uID]
		if reflect.DeepEqual(oldList, newList) {
			continue
		}
		userDoc, err := DB.Get(userBucket, uID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				continue
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if len(newList) == 0 {
			deleteKeyFold(userDoc, "groups")
		} else {
			userDoc["groups"] = newList
		}
		if err := DB.Replace(userBucket, uID, userDoc); err != nil && !errors.Is(err, ErrNotFound) {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func copyGroupMap(src map[string]map[string]interface{}) map[string]map[string]interface{} {
	out := make(map[string]map[string]interface{}, len(src))
	for k, v := range src {
		if cloned, err := roundTripMap(v); err == nil {
			out[k] = cloned
		} else {
			out[k] = v
		}
	}
	return out
}

func roundTripMap(doc map[string]interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func structToMap(v interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}
