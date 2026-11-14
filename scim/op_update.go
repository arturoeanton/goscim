package scim

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/gin-gonic/gin"
)

// patchError is a client error in a PATCH request, carrying the RFC 7644 3.12
// scimType keyword that tells the client which condition it hit.
type patchError struct {
	scimType string
	message  string
}

func (e *patchError) Error() string { return e.message }

func patchFailure(scimType, message string) error {
	return &patchError{scimType: scimType, message: message}
}

type patchSegment struct {
	Attr   string
	Filter string
}

// Update is	PATCH https://example.com/{v}/{resource}/{id}
func Update(resource string) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		resourceType := Resources[resource]
		setAuditTarget(c, "patch", resourceType.Name, id)
		raw, ok := readBody(c)
		if !ok {
			return
		}
		var patchRequest Patch
		if err := json.Unmarshal(raw, &patchRequest); err != nil {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", err.Error())
			return
		}
		if len(patchRequest.Operations) == 0 {
			MakeTypedError(c, http.StatusBadRequest, "invalidSyntax", "Operations is required and must not be empty")
			return
		}
		element, err := getElementByID(c, id, resourceType)
		if err != nil {
			return
		}
		patchedPaths := make(map[string]bool)
		for _, op := range patchRequest.Operations {
			if op.Path != "" && !strings.ContainsAny(op.Path, "[]") {
				op.Path = canonicalPatchPath(resourceType, op.Path)
			}
			if !enforcePatchOperationWriteAccess(c, resourceType, op) {
				return
			}
			recordPatchedWriteOnlyPaths(resourceType, op, patchedPaths)
			element, err = applyPatchOperationForResource(resourceType, op, element)
			if err != nil {
				var failure *patchError
				if errors.As(err, &failure) {
					MakeTypedError(c, http.StatusBadRequest, failure.scimType, failure.message)
					return
				}
				MakeError(c, http.StatusBadRequest, err.Error())
				return
			}
		}

		replaceWithMode(c, resourceType, id, element, true, patchedPaths)
	}
}

// applyPatchOperation preserves the resourceType-agnostic helper for direct
// unit-test callers.
func applyPatchOperation(op Operation, element map[string]interface{}) (map[string]interface{}, error) {
	return applyPatchOperationForResource(ResourceType{}, op, element)
}

// applyPatchOperationForResource applies one RFC 7644 3.5.2 operation to
// element, supporting plain paths, URN-prefixed paths (both "." and ":"
// delimiters), and multi-valued complex value filters such as
// emails[type eq "work"].value.
func applyPatchOperationForResource(resourceType ResourceType, op Operation, element map[string]interface{}) (map[string]interface{}, error) {
	operation := strings.ToLower(strings.TrimSpace(op.Op))
	switch operation {
	case "add", "replace":
		if op.Path == "" {
			// RFC 7644 3.5.2.1/3.5.2.3: with no path the value is an object
			// whose members are applied to the resource itself.
			values, ok := op.Value.(map[string]interface{})
			if !ok {
				return nil, patchFailure("invalidSyntax",
					`an "`+operation+`" operation without a path must carry an object value`)
			}
			if dup, hasDup := hasDuplicateCaseKeys(values); hasDup {
				return nil, patchFailure("invalidSyntax", "duplicate attribute "+strconv.Quote(dup))
			}
			for key, value := range values {
				if extURN := matchingExtensionURN(resourceType, key); extURN != "" {
					if value == nil {
						deleteKeyFold(element, extURN)
						continue
					}
					if extPatch, ok := value.(map[string]interface{}); ok {
						existingExt, _ := lookupValueFold(element, extURN)
						targetExt, ok := existingExt.(map[string]interface{})
						if !ok || targetExt == nil {
							targetExt = make(map[string]interface{})
						}
						deleteKeyFold(element, extURN)
						for subKey, subVal := range extPatch {
							if prevKey, found := lookupKey(targetExt, subKey); found && prevKey != subKey {
								delete(targetExt, prevKey)
							}
							targetExt[subKey] = subVal
						}
						element[extURN] = targetExt
						ensureSchemaListed(element, extURN)
						continue
					}
				}
				if strings.ContainsAny(key, ".:[") {
					subPath := key
					if !strings.ContainsAny(subPath, "[]") {
						subPath = canonicalPatchPath(resourceType, subPath)
					}
					var err error
					element, err = applyPatchOperationForResource(resourceType, Operation{
						Op:    operation,
						Path:  subPath,
						Value: value,
					}, element)
					if err != nil {
						return nil, err
					}
					continue
				}
				if existingKey, found := lookupKey(element, key); found && existingKey != key {
					delete(element, existingKey)
				}
				element[key] = value
			}
			return element, nil
		}
	case "remove":
		if op.Path == "" {
			return nil, patchFailure("noTarget", `a "remove" operation requires a path`)
		}
	default:
		return nil, patchFailure("invalidSyntax", "unknown patch operation "+strconv.Quote(op.Op))
	}

	extURN, attrPath := splitPatchURN(resourceType, op.Path)
	if extURN != "" && attrPath == "" {
		// Target is the whole extension object itself.
		switch operation {
		case "remove":
			deleteKeyFold(element, extURN)
		case "add", "replace":
			extObj, ok := op.Value.(map[string]interface{})
			if !ok {
				return nil, patchFailure("invalidValue", "extension value must be a JSON object")
			}
			deleteKeyFold(element, extURN)
			element[extURN] = extObj
			ensureSchemaListed(element, extURN)
		}
		return element, nil
	}

	segments, err := parsePatchSegments(attrPath)
	if err != nil {
		return nil, err
	}

	targetRoot := element
	activeSchemaID := resourceType.Schema
	createdExt := false
	if extURN != "" {
		activeSchemaID = extURN
		rawExt, hasExt := lookupValueFold(element, extURN)
		extMap, isMap := rawExt.(map[string]interface{})
		if !hasExt || !isMap || extMap == nil {
			if operation == "remove" {
				return nil, patchFailure("noTarget", "path "+strconv.Quote(op.Path)+" does not resolve")
			}
			extMap = make(map[string]interface{})
			deleteKeyFold(element, extURN)
			element[extURN] = extMap
			createdExt = true
		}
		targetRoot = extMap
	}

	var activeAttrs []Attribute
	if schema, ok := LookupSchema(activeSchemaID); ok {
		activeAttrs = schema.Attributes
	}

	if err := applySegments(operation, op.Path, segments, activeAttrs, targetRoot, op.Value); err != nil {
		return nil, err
	}
	if createdExt {
		ensureSchemaListed(element, extURN)
	}
	return element, nil
}

func applySegments(operation, fullPath string, segments []patchSegment, schemaAttrs []Attribute, current map[string]interface{}, value interface{}) error {
	if len(segments) == 0 {
		return patchFailure("invalidPath", "empty path")
	}

	seg := segments[0]
	var segAttr Attribute
	var hasSegAttr bool
	if len(schemaAttrs) > 0 {
		segAttr, hasSegAttr = FindAttribute(schemaAttrs, seg.Attr)
	}

	field := seg.Attr
	if existingKey, found := lookupKey(current, seg.Attr); found {
		field = existingKey
	} else if hasSegAttr {
		field = segAttr.Name
	}

	// Plain segment without a bracketed value filter.
	if seg.Filter == "" {
		if len(segments) == 1 {
			switch operation {
			case "add":
				current[field] = addPatchValue(current[field], value)
			case "replace":
				current[field] = value
			case "remove":
				if value != nil {
					if existingArr, isArr := current[field].([]interface{}); isArr {
						remaining := removeMatchingPatchElements(existingArr, value, segAttr)
						if len(remaining) == 0 {
							delete(current, field)
						} else {
							current[field] = remaining
						}
						return nil
					}
					if current[field] == nil && hasSegAttr && segAttr.MultiValued {
						return nil
					}
				}
				delete(current, field)
			}
			return nil
		}

		next, ok := current[field].(map[string]interface{})
		if !ok {
			// Auto-create an omitted complex object on add/replace only when
			// the schema declares it as a single-valued complex attribute.
			if (operation == "add" || operation == "replace") && current[field] == nil &&
				hasSegAttr && strings.EqualFold(segAttr.Type, "complex") && !segAttr.MultiValued {
				next = make(map[string]interface{})
				current[field] = next
			} else {
				return patchFailure("noTarget",
					"path "+strconv.Quote(fullPath)+" does not resolve: "+strconv.Quote(field)+" is not an object")
			}
		}
		return applySegments(operation, fullPath, segments[1:], segAttr.SubAttributes, next, value)
	}

	// Bracketed value-selection segment: seg.Attr[seg.Filter]
	if hasSegAttr && !segAttr.MultiValued {
		return patchFailure("invalidPath",
			"value filter cannot be applied to single-valued attribute "+strconv.Quote(segAttr.Name))
	}

	matcher, err := parser.CompileMatcher(seg.Filter, func(subPath string) bool {
		if !hasSegAttr {
			return false
		}
		sub := GetAttribute(segAttr.SubAttributes, subPath)
		return sub.CaseExact
	})
	if err != nil {
		return patchFailure("invalidPath", err.Error())
	}

	var arr []interface{}
	if rawExisting, exists := current[field]; exists && rawExisting != nil {
		typed, ok := rawExisting.([]interface{})
		if !ok {
			return patchFailure("invalidPath",
				"path "+strconv.Quote(fullPath)+" does not resolve: "+strconv.Quote(field)+" is not multi-valued")
		}
		arr = typed
	}

	matched := make([]int, 0)
	for idx, item := range arr {
		if elemMap, ok := item.(map[string]interface{}); ok && matcher(elemMap) {
			matched = append(matched, idx)
		}
	}

	tail := segments[1:]
	if len(tail) == 0 {
		return applyTerminalFilteredSegment(operation, fullPath, current, field, arr, matched, seg.Filter, segAttr.SubAttributes, value)
	}
	return applySubPathFilteredSegment(operation, fullPath, current, field, arr, matched, seg.Filter, segAttr.SubAttributes, tail, value)
}

func applyTerminalFilteredSegment(
	operation, fullPath string,
	parent map[string]interface{},
	field string,
	arr []interface{},
	matched []int,
	filter string,
	subAttrs []Attribute,
	value interface{},
) error {
	switch operation {
	case "remove":
		if len(matched) == 0 {
			return patchFailure("noTarget", "no element matched "+strconv.Quote(fullPath))
		}
		dropSet := make(map[int]struct{}, len(matched))
		for _, idx := range matched {
			dropSet[idx] = struct{}{}
		}
		remaining := make([]interface{}, 0, len(arr)-len(matched))
		for idx, item := range arr {
			if _, drop := dropSet[idx]; !drop {
				remaining = append(remaining, item)
			}
		}
		if len(remaining) == 0 {
			delete(parent, field)
		} else {
			parent[field] = remaining
		}
		return nil

	case "replace":
		if len(matched) == 0 {
			return patchFailure("noTarget", "no element matched "+strconv.Quote(fullPath))
		}
		for _, idx := range matched {
			if valMap, ok := value.(map[string]interface{}); ok {
				arr[idx] = copyShallowMap(valMap)
			} else {
				arr[idx] = value
			}
		}
		parent[field] = arr
		return nil

	case "add":
		valMap, ok := value.(map[string]interface{})
		if !ok {
			return patchFailure("invalidValue", "value for filtered complex attribute must be an object")
		}
		if len(matched) > 0 {
			for _, idx := range matched {
				elemMap := arr[idx].(map[string]interface{})
				mergeIntoObject(elemMap, valMap, subAttrs)
			}
			parent[field] = arr
			return nil
		}
		newElem, err := buildSeededElement(filter, subAttrs)
		if err != nil {
			return patchFailure("invalidPath", err.Error())
		}
		mergeIntoObject(newElem, valMap, subAttrs)
		parent[field] = append(arr, newElem)
		return nil
	}
	return nil
}

func applySubPathFilteredSegment(
	operation, fullPath string,
	parent map[string]interface{},
	field string,
	arr []interface{},
	matched []int,
	filter string,
	subAttrs []Attribute,
	tail []patchSegment,
	value interface{},
) error {
	switch operation {
	case "remove", "replace":
		if len(matched) == 0 {
			return patchFailure("noTarget", "no element matched "+strconv.Quote(fullPath))
		}
		for _, idx := range matched {
			elemMap := arr[idx].(map[string]interface{})
			if err := applySegments(operation, fullPath, tail, subAttrs, elemMap, value); err != nil {
				return err
			}
		}
		parent[field] = arr
		return nil

	case "add":
		if len(matched) > 0 {
			for _, idx := range matched {
				elemMap := arr[idx].(map[string]interface{})
				if err := applySegments("add", fullPath, tail, subAttrs, elemMap, value); err != nil {
					return err
				}
			}
			parent[field] = arr
			return nil
		}
		newElem, err := buildSeededElement(filter, subAttrs)
		if err != nil {
			return patchFailure("invalidPath", err.Error())
		}
		if err := applySegments("add", fullPath, tail, subAttrs, newElem, value); err != nil {
			return err
		}
		parent[field] = append(arr, newElem)
		return nil
	}
	return nil
}

func buildSeededElement(filter string, subAttrs []Attribute) (map[string]interface{}, error) {
	seeds, err := parser.EqualitySeeds(filter)
	if err != nil {
		return nil, err
	}
	elem := make(map[string]interface{}, len(seeds)+1)
	for k, v := range seeds {
		if sub, found := FindAttribute(subAttrs, k); found {
			elem[sub.Name] = v
		} else {
			elem[k] = v
		}
	}
	return elem, nil
}

func mergeIntoObject(dst, src map[string]interface{}, subAttrs []Attribute) {
	for k, v := range src {
		targetKey := k
		if existing, found := lookupKey(dst, k); found {
			targetKey = existing
		} else if sub, found := FindAttribute(subAttrs, k); found {
			targetKey = sub.Name
		}
		dst[targetKey] = v
	}
}

func copyShallowMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// splitPatchURN separates a schema URN prefix (using either ":" or "." as the
// delimiter) from the attribute path inside it. Core schema URNs strip down to
// "", attrPath because core attributes live at the top level of the resource.
func splitPatchURN(resourceType ResourceType, rawPath string) (extURN string, attrPath string) {
	trimmed := strings.TrimSpace(rawPath)
	if resourceType.Schema != "" {
		if rest, ok := trimKnownURNPrefix(trimmed, resourceType.Schema); ok {
			return "", rest
		}
	}
	for _, ext := range resourceType.SchemaExtensions {
		if strings.EqualFold(trimmed, ext.Schema) {
			return canonicalSchemaID(ext.Schema), ""
		}
		if rest, ok := trimKnownURNPrefix(trimmed, ext.Schema); ok {
			return canonicalSchemaID(ext.Schema), rest
		}
	}
	urn, rest := parser.SplitURNPath(trimmed)
	if urn != "" {
		if resourceType.Schema != "" && strings.EqualFold(urn, resourceType.Schema) {
			return "", rest
		}
		return canonicalSchemaID(urn), rest
	}
	return "", trimmed
}

func trimKnownURNPrefix(rawPath, schemaURN string) (string, bool) {
	if len(rawPath) <= len(schemaURN)+1 {
		return "", false
	}
	if !strings.EqualFold(rawPath[:len(schemaURN)], schemaURN) {
		return "", false
	}
	sep := rawPath[len(schemaURN)]
	if sep != ':' && sep != '.' {
		return "", false
	}
	return rawPath[len(schemaURN)+1:], true
}

func canonicalPatchPath(resourceType ResourceType, rawPath string) string {
	extURN, attrPath := splitPatchURN(resourceType, rawPath)
	if extURN != "" && attrPath == "" {
		return extURN
	}
	if extURN != "" {
		return CanonicalAttributePath(resourceType, extURN+"."+attrPath)
	}
	return CanonicalAttributePath(resourceType, attrPath)
}

// parsePatchSegments splits an attribute path on dots outside [...] and "...",
// validating any value-selection filter expressions.
func parsePatchSegments(path string) ([]patchSegment, error) {
	if path == "" {
		return nil, patchFailure("invalidPath", "empty path")
	}
	var segments []patchSegment
	segStart := 0
	bracketDepth := 0
	inQuote := false

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '"':
			inQuote = !inQuote
		case '[':
			if !inQuote {
				bracketDepth++
			}
		case ']':
			if !inQuote {
				bracketDepth--
				if bracketDepth < 0 {
					return nil, patchFailure("invalidPath", "unbalanced brackets in path "+strconv.Quote(path))
				}
			}
		case '.':
			if !inQuote && bracketDepth == 0 {
				seg, err := parseOnePatchSegment(path[segStart:i], path)
				if err != nil {
					return nil, err
				}
				segments = append(segments, seg)
				segStart = i + 1
			}
		}
	}
	if inQuote || bracketDepth != 0 {
		return nil, patchFailure("invalidPath", "malformed path "+strconv.Quote(path))
	}
	lastSeg, err := parseOnePatchSegment(path[segStart:], path)
	if err != nil {
		return nil, err
	}
	segments = append(segments, lastSeg)
	return segments, nil
}

func parseOnePatchSegment(rawSeg, fullPath string) (patchSegment, error) {
	if rawSeg == "" {
		return patchSegment{}, patchFailure("invalidPath", "empty segment in path "+strconv.Quote(fullPath))
	}
	openIdx := strings.IndexByte(rawSeg, '[')
	if openIdx < 0 {
		if strings.ContainsRune(rawSeg, ']') {
			return patchSegment{}, patchFailure("invalidPath", "unbalanced brackets in path "+strconv.Quote(fullPath))
		}
		return patchSegment{Attr: rawSeg}, nil
	}
	if openIdx == 0 || rawSeg[len(rawSeg)-1] != ']' {
		return patchSegment{}, patchFailure("invalidPath", "malformed value filter in path "+strconv.Quote(fullPath))
	}
	attr := rawSeg[:openIdx]
	filter := strings.TrimSpace(rawSeg[openIdx+1 : len(rawSeg)-1])
	if filter == "" {
		return patchSegment{}, patchFailure("invalidPath", "empty value filter in path "+strconv.Quote(fullPath))
	}
	if err := parser.Validate(filter); err != nil {
		return patchSegment{}, patchFailure("invalidPath", err.Error())
	}
	return patchSegment{Attr: attr, Filter: filter}, nil
}

// addPatchValue implements the "add" semantics: appending to an existing array,
// and otherwise setting the value.
func addPatchValue(current interface{}, value interface{}) interface{} {
	existing, ok := current.([]interface{})
	if !ok {
		return value
	}
	if values, ok := value.([]interface{}); ok {
		return append(existing, values...)
	}
	return append(existing, value)
}

// removeMatchingPatchElements removes from existing every element matching one
// of the target items in value (RFC 7644 3.5.2.2 / Microsoft Entra ID group
// member removal with path:"members" and value:[{"value":"<id>"}]).
func removeMatchingPatchElements(existing []interface{}, value interface{}, attr Attribute) []interface{} {
	var targets []interface{}
	if arr, ok := value.([]interface{}); ok {
		targets = arr
	} else {
		targets = []interface{}{value}
	}
	remaining := make([]interface{}, 0, len(existing))
	for _, elem := range existing {
		matched := false
		for _, target := range targets {
			if patchElementMatchesTarget(elem, target, attr) {
				matched = true
				break
			}
		}
		if !matched {
			remaining = append(remaining, elem)
		}
	}
	return remaining
}

func patchElementMatchesTarget(elem, target interface{}, attr Attribute) bool {
	targetMap, targetIsMap := target.(map[string]interface{})
	elemMap, elemIsMap := elem.(map[string]interface{})
	if targetIsMap && elemIsMap {
		if len(targetMap) == 0 {
			return false
		}
		for k, tv := range targetMap {
			ev, found := lookupValueFold(elemMap, k)
			if !found {
				return false
			}
			sub, hasSub := FindAttribute(attr.SubAttributes, k)
			if !patchScalarEqual(ev, tv, hasSub && sub.CaseExact) {
				return false
			}
		}
		return true
	}
	return patchScalarEqual(elem, target, attr.CaseExact)
}

func patchScalarEqual(a, b interface{}, caseExact bool) bool {
	sa, aStr := a.(string)
	sb, bStr := b.(string)
	if aStr && bStr {
		if caseExact {
			return sa == sb
		}
		return strings.EqualFold(sa, sb)
	}
	return valuesEqual(a, b)
}

// opPathTopathArray splits a patch path into its segments, keeping a schema URN
// prefix whole as the first one.
func opPathTopathArray(value string) []string {
	urn, path := parser.SplitURNPath(value)
	pathArray := make([]string, 0)
	if urn != "" {
		pathArray = append(pathArray, urn)
	}
	return append(pathArray, strings.Split(path, ".")...)
}
