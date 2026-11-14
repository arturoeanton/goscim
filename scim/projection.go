package scim

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Projection holds parsed RFC 7644 3.4.2.5 / 3.9 attribute selection
// parameters. At most one of Attributes and ExcludedAttributes may be non-empty.
type Projection struct {
	Attributes         []string
	ExcludedAttributes []string
}

// ParseProjectionParams normalises raw attribute lists (splitting comma-
// delimited entries) and enforces mutual exclusivity between attributes and
// excludedAttributes.
func ParseProjectionParams(rawAttrs, rawExcluded []string) (Projection, error) {
	attrs := splitParamList(rawAttrs)
	excluded := splitParamList(rawExcluded)
	if len(attrs) > 0 && len(excluded) > 0 {
		return Projection{}, errors.New("attributes and excludedAttributes are mutually exclusive")
	}
	return Projection{
		Attributes:         attrs,
		ExcludedAttributes: excluded,
	}, nil
}

func splitParamList(raw []string) []string {
	var out []string
	for _, entry := range raw {
		for _, part := range strings.Split(entry, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				out = append(out, trimmed)
			}
		}
	}
	return out
}

// parseProjectionFromQuery extracts ?attributes= and ?excludedAttributes= from
// the request URL. On error it writes a 400 SCIM error and returns false.
func parseProjectionFromQuery(c *gin.Context) (Projection, bool) {
	proj, err := ParseProjectionParams(
		c.QueryArray("attributes"),
		c.QueryArray("excludedAttributes"),
	)
	if err != nil {
		MakeTypedError(c, http.StatusBadRequest, "invalidValue", err.Error())
		return Projection{}, false
	}
	return proj, true
}

// FormatResource applies $reader / writeOnly / returned:never filtering first
// and then applies returned:always / default / request rules alongside proj.
func FormatResource(roles []string, resourceType ResourceType, element map[string]interface{}, proj Projection) map[string]interface{} {
	visible := ValidateReadRole(roles, resourceType, element)
	if visible == nil {
		return nil
	}
	return applyProjection(resourceType, visible, proj)
}

// selectionTree represents parsed attribute paths scoped to one schema (core or
// a specific extension).
type selectionTree struct {
	wholeSchema bool
	attrs       map[string]*attrSelection // lowercase canonical attr name -> selection
}

type attrSelection struct {
	whole bool
	subs  map[string]bool // lowercase canonical sub-attribute name -> true
}

func newSelectionTree() *selectionTree {
	return &selectionTree{attrs: make(map[string]*attrSelection)}
}

func (t *selectionTree) addAttr(attrName, subName string) {
	key := strings.ToLower(attrName)
	sel, ok := t.attrs[key]
	if !ok {
		sel = &attrSelection{subs: make(map[string]bool)}
		t.attrs[key] = sel
	}
	if subName == "" {
		sel.whole = true
		return
	}
	sel.subs[strings.ToLower(subName)] = true
}

type parsedSelection struct {
	core       *selectionTree
	extensions map[string]*selectionTree // canonical extension URN -> tree
	externalID bool
	metaWhole  bool
	metaSubs   map[string]bool // lowercase meta field name -> true
}

func parseSelection(resourceType ResourceType, paths []string) *parsedSelection {
	ps := &parsedSelection{
		core:       newSelectionTree(),
		extensions: make(map[string]*selectionTree),
		metaSubs:   make(map[string]bool),
	}
	for _, ext := range resourceType.SchemaExtensions {
		ps.extensions[ext.Schema] = newSelectionTree()
	}

	coreSchema, _ := LookupSchema(resourceType.Schema)

	for _, raw := range paths {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}

		// 1. Check if raw names an entire extension URN or has an extension URN prefix.
		matchedExt := false
		for _, ext := range resourceType.SchemaExtensions {
			if strings.EqualFold(trimmed, ext.Schema) {
				ps.extensions[ext.Schema].wholeSchema = true
				matchedExt = true
				break
			}
			if rest, ok := trimSchemaPrefix(trimmed, ext.Schema); ok {
				extSchema, _ := LookupSchema(ext.Schema)
				recordSchemaPath(ps.extensions[ext.Schema], extSchema.Attributes, rest)
				matchedExt = true
				break
			}
		}
		if matchedExt {
			continue
		}

		// 2. Check if raw carries the core schema URN prefix.
		if strings.EqualFold(trimmed, resourceType.Schema) {
			ps.core.wholeSchema = true
			continue
		}
		if rest, ok := trimSchemaPrefix(trimmed, resourceType.Schema); ok {
			trimmed = rest
		} else if strings.HasPrefix(strings.ToLower(trimmed), "urn:") {
			// Belongs to another resource type's schema (e.g. on root /.search); ignore here.
			continue
		}

		// 3. Common attributes: id, externalId, schemas, meta(.sub).
		lower := strings.ToLower(trimmed)
		if lower == "id" || lower == "schemas" {
			continue
		}
		if lower == "externalid" {
			ps.externalID = true
			continue
		}
		if lower == "meta" {
			ps.metaWhole = true
			continue
		}
		if strings.HasPrefix(lower, "meta.") {
			sub := trimmed[len("meta."):]
			if sub != "" {
				ps.metaSubs[strings.ToLower(sub)] = true
			}
			continue
		}

		// 4. Core schema attribute path.
		recordSchemaPath(ps.core, coreSchema.Attributes, trimmed)
	}
	return ps
}

// trimSchemaPrefix strips schemaURN followed by ':' or '.' (RFC 7644 3.10 uses
// ':', while GoSCIM filters also accept '.').
func trimSchemaPrefix(raw, schemaURN string) (string, bool) {
	if len(raw) <= len(schemaURN) {
		return "", false
	}
	if !strings.EqualFold(raw[:len(schemaURN)], schemaURN) {
		return "", false
	}
	sep := raw[len(schemaURN)]
	if sep != ':' && sep != '.' {
		return "", false
	}
	return raw[len(schemaURN)+1:], true
}

func recordSchemaPath(tree *selectionTree, attributes []Attribute, attrPath string) {
	parts := strings.SplitN(attrPath, ".", 2)
	topAttr, found := FindAttribute(attributes, parts[0])
	if !found {
		return
	}
	if len(parts) == 1 {
		tree.addAttr(topAttr.Name, "")
		return
	}
	subAttr, subFound := FindAttribute(topAttr.SubAttributes, parts[1])
	if !subFound {
		return
	}
	tree.addAttr(topAttr.Name, subAttr.Name)
}

func applyProjection(resourceType ResourceType, visible map[string]interface{}, proj Projection) map[string]interface{} {
	coreSchema, _ := LookupSchema(resourceType.Schema)

	switch {
	case len(proj.Attributes) > 0:
		sel := parseSelection(resourceType, proj.Attributes)
		out := make(map[string]interface{})
		if id, ok := visible["id"]; ok {
			out["id"] = id
		}
		if sel.externalID || sel.core.wholeSchema {
			if extID, ok := visible["externalId"]; ok {
				out["externalId"] = extID
			}
		}
		if metaVal, ok := projectMetaInclude(visible["meta"], sel); ok {
			out["meta"] = metaVal
		}
		projectIncludeInto(out, coreSchema.Attributes, visible, sel.core)

		keptExtensions := make(map[string]bool)
		for _, ext := range resourceType.SchemaExtensions {
			rawExt, present := visible[ext.Schema]
			if !present {
				continue
			}
			extMap, ok := rawExt.(map[string]interface{})
			if !ok {
				continue
			}
			extSchema, _ := LookupSchema(ext.Schema)
			projectedExt := make(map[string]interface{})
			projectIncludeInto(projectedExt, extSchema.Attributes, extMap, sel.extensions[ext.Schema])
			if len(projectedExt) > 0 {
				out[ext.Schema] = projectedExt
				keptExtensions[strings.ToLower(ext.Schema)] = true
			}
		}
		out["schemas"] = projectedSchemasList(resourceType, visible["schemas"], keptExtensions)
		return out

	case len(proj.ExcludedAttributes) > 0:
		sel := parseSelection(resourceType, proj.ExcludedAttributes)
		out := make(map[string]interface{})
		if id, ok := visible["id"]; ok {
			out["id"] = id
		}
		if !sel.externalID {
			if extID, ok := visible["externalId"]; ok {
				out["externalId"] = extID
			}
		}
		if metaVal, ok := projectMetaExclude(visible["meta"], sel); ok {
			out["meta"] = metaVal
		}
		projectExcludeInto(out, coreSchema.Attributes, visible, sel.core)

		keptExtensions := make(map[string]bool)
		for _, ext := range resourceType.SchemaExtensions {
			rawExt, present := visible[ext.Schema]
			if !present {
				continue
			}
			extMap, ok := rawExt.(map[string]interface{})
			if !ok {
				continue
			}
			extSchema, _ := LookupSchema(ext.Schema)
			projectedExt := make(map[string]interface{})
			projectExcludeInto(projectedExt, extSchema.Attributes, extMap, sel.extensions[ext.Schema])
			if len(projectedExt) > 0 || (len(extMap) == 0 && !sel.extensions[ext.Schema].wholeSchema) {
				out[ext.Schema] = projectedExt
				keptExtensions[strings.ToLower(ext.Schema)] = true
			}
		}
		out["schemas"] = projectedSchemasList(resourceType, visible["schemas"], keptExtensions)
		return out

	default:
		// Default mode: keep id, schemas, meta and all default/always attributes;
		// strip attributes declared with returned: "request".
		out := make(map[string]interface{}, len(visible))
		for k, v := range visible {
			out[k] = v
		}
		stripRequestOnlyAttributes(out, coreSchema.Attributes)
		for _, ext := range resourceType.SchemaExtensions {
			if extMap, ok := out[ext.Schema].(map[string]interface{}); ok {
				extSchema, _ := LookupSchema(ext.Schema)
				stripRequestOnlyAttributes(extMap, extSchema.Attributes)
			}
		}
		return out
	}
}

func projectIncludeInto(dst map[string]interface{}, attributes []Attribute, src map[string]interface{}, tree *selectionTree) {
	for _, attr := range attributes {
		val, present := src[attr.Name]
		if !present {
			continue
		}
		always := isReturnedAlways(attr)
		reqOnly := isReturnedRequest(attr)
		var sel *attrSelection
		if tree != nil {
			sel = tree.attrs[strings.ToLower(attr.Name)]
		}
		wholeRequested := (tree != nil && tree.wholeSchema && !reqOnly) || (sel != nil && sel.whole)
		subsRequested := sel != nil && len(sel.subs) > 0

		if !always && !wholeRequested && !subsRequested {
			continue
		}

		if len(attr.SubAttributes) == 0 {
			dst[attr.Name] = val
			continue
		}

		projected, keep := projectComplexInclude(attr.SubAttributes, val, always || wholeRequested, sel)
		if keep {
			dst[attr.Name] = projected
		}
	}
}

func projectComplexInclude(subAttrs []Attribute, val interface{}, includeDefaultSubs bool, sel *attrSelection) (interface{}, bool) {
	switch typed := val.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{})
		for _, sub := range subAttrs {
			sv, ok := typed[sub.Name]
			if !ok {
				continue
			}
			subExplicit := sel != nil && sel.subs[strings.ToLower(sub.Name)]
			if isReturnedAlways(sub) || subExplicit || (includeDefaultSubs && !isReturnedRequest(sub)) {
				m[sub.Name] = sv
			}
		}
		return m, len(m) > 0

	case []interface{}:
		items := make([]interface{}, 0, len(typed))
		for _, elem := range typed {
			em, ok := elem.(map[string]interface{})
			if !ok {
				if includeDefaultSubs {
					items = append(items, elem)
				}
				continue
			}
			projectedElem, keep := projectComplexInclude(subAttrs, em, includeDefaultSubs, sel)
			if keep {
				items = append(items, projectedElem)
			}
		}
		if len(typed) == 0 && includeDefaultSubs {
			return items, true
		}
		return items, len(items) > 0

	default:
		return val, includeDefaultSubs
	}
}

func projectExcludeInto(dst map[string]interface{}, attributes []Attribute, src map[string]interface{}, tree *selectionTree) {
	for _, attr := range attributes {
		val, present := src[attr.Name]
		if !present {
			continue
		}
		if isReturnedRequest(attr) {
			continue
		}
		always := isReturnedAlways(attr)
		var sel *attrSelection
		if tree != nil {
			sel = tree.attrs[strings.ToLower(attr.Name)]
		}
		wholeExcluded := (tree != nil && tree.wholeSchema) || (sel != nil && sel.whole)

		if len(attr.SubAttributes) == 0 {
			if wholeExcluded && !always {
				continue
			}
			dst[attr.Name] = val
			continue
		}

		if wholeExcluded && !always && !hasAlwaysSubAttribute(attr.SubAttributes) {
			continue
		}
		projected, keep := projectComplexExclude(attr.SubAttributes, val, wholeExcluded && !always, sel)
		if keep {
			dst[attr.Name] = projected
		}
	}
}

func projectComplexExclude(subAttrs []Attribute, val interface{}, excludeNonAlways bool, sel *attrSelection) (interface{}, bool) {
	switch typed := val.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{})
		for _, sub := range subAttrs {
			sv, ok := typed[sub.Name]
			if !ok || isReturnedRequest(sub) {
				continue
			}
			if isReturnedAlways(sub) {
				m[sub.Name] = sv
				continue
			}
			if excludeNonAlways {
				continue
			}
			if sel != nil && sel.subs[strings.ToLower(sub.Name)] {
				continue
			}
			m[sub.Name] = sv
		}
		return m, len(m) > 0

	case []interface{}:
		items := make([]interface{}, 0, len(typed))
		for _, elem := range typed {
			em, ok := elem.(map[string]interface{})
			if !ok {
				if !excludeNonAlways {
					items = append(items, elem)
				}
				continue
			}
			projectedElem, keep := projectComplexExclude(subAttrs, em, excludeNonAlways, sel)
			if keep {
				items = append(items, projectedElem)
			}
		}
		if len(typed) == 0 && !excludeNonAlways {
			return items, true
		}
		return items, len(items) > 0

	default:
		return val, !excludeNonAlways
	}
}

func stripRequestOnlyAttributes(obj map[string]interface{}, attributes []Attribute) {
	for _, attr := range attributes {
		val, present := obj[attr.Name]
		if !present {
			continue
		}
		if isReturnedRequest(attr) {
			delete(obj, attr.Name)
			continue
		}
		if len(attr.SubAttributes) == 0 {
			continue
		}
		switch typed := val.(type) {
		case map[string]interface{}:
			cp := make(map[string]interface{}, len(typed))
			for k, v := range typed {
				cp[k] = v
			}
			stripRequestOnlyAttributes(cp, attr.SubAttributes)
			obj[attr.Name] = cp
		case []interface{}:
			items := make([]interface{}, 0, len(typed))
			for _, elem := range typed {
				if em, ok := elem.(map[string]interface{}); ok {
					cp := make(map[string]interface{}, len(em))
					for k, v := range em {
						cp[k] = v
					}
					stripRequestOnlyAttributes(cp, attr.SubAttributes)
					items = append(items, cp)
				} else {
					items = append(items, elem)
				}
			}
			obj[attr.Name] = items
		}
	}
}

func projectMetaInclude(rawMeta interface{}, sel *parsedSelection) (interface{}, bool) {
	if rawMeta == nil {
		return nil, false
	}
	if sel.metaWhole || sel.core.wholeSchema {
		return rawMeta, true
	}
	if len(sel.metaSubs) == 0 {
		return nil, false
	}
	metaMap, ok := toMetaMap(rawMeta)
	if !ok {
		return nil, false
	}
	filtered := make(map[string]interface{})
	for k, v := range metaMap {
		if sel.metaSubs[strings.ToLower(k)] {
			filtered[k] = v
		}
	}
	return filtered, len(filtered) > 0
}

func projectMetaExclude(rawMeta interface{}, sel *parsedSelection) (interface{}, bool) {
	if rawMeta == nil || sel.metaWhole {
		return nil, false
	}
	if len(sel.metaSubs) == 0 {
		return rawMeta, true
	}
	metaMap, ok := toMetaMap(rawMeta)
	if !ok {
		return rawMeta, true
	}
	filtered := make(map[string]interface{}, len(metaMap))
	for k, v := range metaMap {
		if !sel.metaSubs[strings.ToLower(k)] {
			filtered[k] = v
		}
	}
	return filtered, len(filtered) > 0
}

func toMetaMap(rawMeta interface{}) (map[string]interface{}, bool) {
	if m, ok := rawMeta.(map[string]interface{}); ok {
		return m, true
	}
	if m, ok := rawMeta.(Meta); ok {
		out := map[string]interface{}{
			"resourceType": m.ResourceType,
			"created":      m.Created,
			"lastModified": m.LastModified,
			"version":      m.Version,
			"location":     m.Location,
		}
		return out, true
	}
	return nil, false
}

func projectedSchemasList(resourceType ResourceType, rawSchemas interface{}, keptExtensions map[string]bool) []string {
	out := []string{resourceType.Schema}
	seen := map[string]bool{strings.ToLower(resourceType.Schema): true}

	addExt := func(urn string) {
		lower := strings.ToLower(urn)
		if keptExtensions[lower] && !seen[lower] {
			seen[lower] = true
			out = append(out, canonicalSchemaID(urn))
		}
	}

	switch typed := rawSchemas.(type) {
	case []string:
		for _, s := range typed {
			addExt(s)
		}
	case []interface{}:
		for _, item := range typed {
			if s, ok := item.(string); ok {
				addExt(s)
			}
		}
	}
	for _, ext := range resourceType.SchemaExtensions {
		addExt(ext.Schema)
	}
	return out
}

func isReturnedAlways(attr Attribute) bool {
	return strings.EqualFold(attr.Returned, "always")
}

func isReturnedRequest(attr Attribute) bool {
	return strings.EqualFold(attr.Returned, "request")
}

func hasAlwaysSubAttribute(subs []Attribute) bool {
	for _, s := range subs {
		if isReturnedAlways(s) {
			return true
		}
	}
	return false
}
