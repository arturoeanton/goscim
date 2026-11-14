package scim

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/gin-gonic/gin"
)

// EnforceWriteAccess applies the write-side rules to an incoming resource:
// SCIM's own mutability keyword (RFC 7643 7) and this project's $writer role
// lists, the mirror of $reader.
//
// existing is the stored resource on an update and nil on a create. It reports
// whether the caller may proceed; when it returns false the response has
// already been written.
func EnforceWriteAccess(c *gin.Context, resourceType ResourceType, incoming, existing map[string]interface{}) bool {
	return enforceWriteAccess(c, resourceType, incoming, existing, false)
}

func enforceWriteAccess(c *gin.Context, resourceType ResourceType, incoming, existing map[string]interface{}, isPatch bool) bool {
	if dup, ok := hasDuplicateCaseKeys(incoming); ok {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax",
			"duplicate attribute "+strconv.Quote(dup))
		return false
	}
	canonicaliseKey(incoming, "externalId")
	roles := currentRoles(c)

	coreSchema, _ := LookupSchema(resourceType.Schema)
	if !enforceAttributeWrites(c, roles, coreSchema.Attributes, incoming, existing, isPatch) {
		return false
	}
	for _, extension := range resourceType.SchemaExtensions {
		rawIncoming, hasIncoming := canonicaliseKey(incoming, extension.Schema)
		rawExisting, hasExisting := lookupValueFold(existing, extension.Schema)
		storedValues, _ := rawExisting.(map[string]interface{})
		extensionSchema, _ := LookupSchema(extension.Schema)

		if hasIncoming {
			values, ok := rawIncoming.(map[string]interface{})
			if !ok {
				if storedValues != nil {
					// Replacing an existing extension object with a non-object
					// still has to pass mutability and $writer checks first.
					empty := make(map[string]interface{})
					if !enforceAttributeWrites(c, roles, extensionSchema.Attributes, empty, storedValues, isPatch) {
						return false
					}
				}
				continue
			}
			if !enforceAttributeWrites(c, roles, extensionSchema.Attributes, values, storedValues, isPatch) {
				return false
			}
			continue
		}

		if hasExisting && storedValues != nil {
			restored := make(map[string]interface{})
			if !enforceAttributeWrites(c, roles, extensionSchema.Attributes, restored, storedValues, isPatch) {
				return false
			}
			if len(restored) > 0 {
				incoming[extension.Schema] = restored
				ensureSchemaListed(incoming, extension.Schema)
			}
		}
	}
	return true
}

func ensureSchemaListed(element map[string]interface{}, schemaID string) {
	declared, ok := element["schemas"].([]interface{})
	if !ok {
		return
	}
	for _, item := range declared {
		if name, ok := item.(string); ok && strings.EqualFold(name, schemaID) {
			return
		}
	}
	element["schemas"] = append(declared, schemaID)
}

func enforceAttributeWrites(c *gin.Context, roles []string, attributes []Attribute, incoming, existing map[string]interface{}, isPatch bool) bool {
	if dup, ok := hasDuplicateCaseKeys(incoming); ok {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax",
			"duplicate attribute "+strconv.Quote(dup))
		return false
	}
	for _, attribute := range attributes {
		value, present := canonicaliseKey(incoming, attribute.Name)
		stored, storedPresent := lookupValueFold(existing, attribute.Name)

		if isReadOnly(attribute) {
			// A read-only attribute is server-owned. The client's value is
			// dropped rather than refused, because read-modify-write clients
			// echo the whole resource back on a PUT and rejecting that would
			// make an ordinary update impossible.
			deleteKeyFold(incoming, attribute.Name)
			if storedPresent {
				incoming[attribute.Name] = stored
			}
			continue
		}

		if existing != nil && strings.EqualFold(attribute.Mutability, "immutable") && storedPresent {
			if !present || !valuesEqual(stored, value) {
				MakeTypedError(c, http.StatusBadRequest, "mutability",
					attribute.Name+" is immutable and cannot be changed")
				return false
			}
		}

		if !canWrite(roles, attribute) {
			switch {
			case existing == nil:
				if present {
					return forbidWrite(c, attribute.Name)
				}
			case isPatch:
				if present != storedPresent || (present && !valuesEqual(stored, value)) {
					return forbidWrite(c, attribute.Name)
				}
			default: // PUT
				if present {
					// Only a caller allowed to read the attribute may echo its
					// stored value back unchanged; otherwise a 200-vs-403
					// difference would let a blind caller guess hidden values.
					if !canRead(roles, attribute) || !storedPresent || !valuesEqual(stored, value) {
						return forbidWrite(c, attribute.Name)
					}
				} else if storedPresent {
					if canRead(roles, attribute) {
						return forbidWrite(c, attribute.Name)
					}
					// Hidden from the caller on GET: keep the stored value so a
					// read-modify-write PUT cannot silently wipe it.
					incoming[attribute.Name] = stored
					continue
				}
			}
		}

		if !present {
			continue
		}
		if !enforceNestedWrites(c, roles, attribute, value, stored, isPatch) {
			return false
		}
	}
	return true
}

// enforceNestedWrites applies the same rules to the sub-attributes of a complex
// value, and to every element of a multi-valued one.
func enforceNestedWrites(c *gin.Context, roles []string, attribute Attribute, value, existing interface{}, isPatch bool) bool {
	if len(attribute.SubAttributes) == 0 {
		return true
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		storedValues, _ := existing.(map[string]interface{})
		return enforceAttributeWrites(c, roles, attribute.SubAttributes, typed, storedValues, isPatch)
	case []interface{}:
		for _, item := range typed {
			values, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			// Elements of a multi-valued attribute have no stable identity
			// here, so there is nothing to compare an immutable sub-attribute
			// against; the role and read-only rules still apply.
			if !enforceAttributeWrites(c, roles, attribute.SubAttributes, values, nil, false) {
				return false
			}
		}
	}
	return true
}

// enforcePatchOperationWriteAccess gates a single PATCH operation before it
// mutates the loaded document, so touching or removing a $writer-restricted
// attribute is refused even when the resulting value happens to match what was
// already stored.
func enforcePatchOperationWriteAccess(c *gin.Context, resourceType ResourceType, op Operation) bool {
	roles := currentRoles(c)
	if op.Path != "" {
		return checkPatchPathWritable(c, roles, resourceType, op.Path)
	}
	values, ok := op.Value.(map[string]interface{})
	if !ok {
		return true // syntax error handled by applyPatchOperation
	}
	if dup, ok := hasDuplicateCaseKeys(values); ok {
		MakeTypedError(c, http.StatusBadRequest, "invalidSyntax",
			"duplicate attribute "+strconv.Quote(dup))
		return false
	}
	coreSchema, _ := LookupSchema(resourceType.Schema)
	for key, val := range values {
		if attr, found := FindAttribute(coreSchema.Attributes, key); found {
			if !checkValueWritable(c, roles, attr, val) {
				return false
			}
			continue
		}
		if containsSchemaFold(resourceType.SchemaExtensions, key) {
			extSchema, ok := LookupSchema(key)
			if !ok {
				continue
			}
			extMap, ok := val.(map[string]interface{})
			if !ok {
				continue
			}
			for extKey, extVal := range extMap {
				if attr, found := FindAttribute(extSchema.Attributes, extKey); found {
					if !checkValueWritable(c, roles, attr, extVal) {
						return false
					}
				}
			}
			continue
		}
		if strings.ContainsAny(key, ".:[") {
			if !checkPatchPathWritable(c, roles, resourceType, key) {
				return false
			}
		}
	}
	return true
}

func checkPatchPathWritable(c *gin.Context, roles []string, resourceType ResourceType, rawPath string) bool {
	extURN, attrPath := splitPatchURN(resourceType, rawPath)
	schemaID := resourceType.Schema
	if extURN != "" {
		if !containsSchemaFold(resourceType.SchemaExtensions, extURN) {
			return true
		}
		schemaID = extURN
	}
	schema, ok := LookupSchema(schemaID)
	if !ok {
		return true
	}
	if extURN != "" && attrPath == "" {
		for _, attr := range schema.Attributes {
			if !isReadOnly(attr) && !canWrite(roles, attr) {
				return forbidWrite(c, attr.Name)
			}
		}
		return true
	}

	segments, err := parsePatchSegments(attrPath)
	if err != nil {
		return true // syntax error handled by applyPatchOperationForResource
	}

	attributes := schema.Attributes
	for _, seg := range segments {
		attr, found := FindAttribute(attributes, seg.Attr)
		if !found {
			return true
		}
		if isReadOnly(attr) {
			return true
		}
		if !canWrite(roles, attr) {
			return forbidWrite(c, attr.Name)
		}
		if seg.Filter != "" {
			refPaths, err := parser.ReferencedAttributes(seg.Filter)
			if err != nil {
				return true
			}
			for _, ref := range refPaths {
				for _, subAttr := range resolveSubAttributeChain(attr.SubAttributes, ref) {
					if !isReturnable(subAttr) {
						MakeTypedError(c, http.StatusBadRequest, "invalidFilter",
							"attribute "+strconv.Quote(subAttr.Name)+" cannot be filtered")
						return false
					}
					if !canRead(roles, subAttr) {
						MakeTypedError(c, http.StatusBadRequest, "invalidFilter",
							"the caller is not allowed to filter by "+strconv.Quote(subAttr.Name))
						return false
					}
				}
			}
		}
		attributes = attr.SubAttributes
	}
	return true
}

func resolveSubAttributeChain(attributes []Attribute, dottedPath string) []Attribute {
	segments := strings.Split(dottedPath, ".")
	chain := make([]Attribute, 0, len(segments))
	for _, segment := range segments {
		attr, found := FindAttribute(attributes, segment)
		if !found {
			break
		}
		chain = append(chain, attr)
		attributes = attr.SubAttributes
	}
	return chain
}

func checkValueWritable(c *gin.Context, roles []string, attribute Attribute, value interface{}) bool {
	if isReadOnly(attribute) {
		return true
	}
	if !canWrite(roles, attribute) {
		return forbidWrite(c, attribute.Name)
	}
	if len(attribute.SubAttributes) == 0 {
		return true
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		for k, v := range typed {
			if sub, found := FindAttribute(attribute.SubAttributes, k); found {
				if !checkValueWritable(c, roles, sub, v) {
					return false
				}
			}
		}
	case []interface{}:
		for _, item := range typed {
			if m, ok := item.(map[string]interface{}); ok {
				for k, v := range m {
					if sub, found := FindAttribute(attribute.SubAttributes, k); found {
						if !checkValueWritable(c, roles, sub, v) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

// EnforceDeleteAccess checks that the caller holds $writer permission for the
// resource being removed: any required schema attribute and any writable
// attribute present on the stored document must be writable by the caller.
func EnforceDeleteAccess(c *gin.Context, resourceType ResourceType, stored map[string]interface{}) bool {
	roles := currentRoles(c)
	coreSchema, _ := LookupSchema(resourceType.Schema)
	if !canDeleteAttributes(c, roles, coreSchema.Attributes, stored) {
		return false
	}
	for _, extension := range resourceType.SchemaExtensions {
		rawExt, present := lookupValueFold(stored, extension.Schema)
		extMap, _ := rawExt.(map[string]interface{})
		if !present && !extension.Required {
			continue
		}
		extSchema, _ := LookupSchema(extension.Schema)
		if !canDeleteAttributes(c, roles, extSchema.Attributes, extMap) {
			return false
		}
	}
	return true
}

func canDeleteAttributes(c *gin.Context, roles []string, attributes []Attribute, stored map[string]interface{}) bool {
	for _, attribute := range attributes {
		if isReadOnly(attribute) {
			continue
		}
		val, present := lookupValueFold(stored, attribute.Name)
		if !present && !attribute.Required {
			continue
		}
		if !canWrite(roles, attribute) {
			MakeError(c, http.StatusForbidden,
				"the caller is not allowed to delete a resource carrying "+attribute.Name)
			return false
		}
		if present && len(attribute.SubAttributes) > 0 {
			switch typed := val.(type) {
			case map[string]interface{}:
				if !canDeleteAttributes(c, roles, attribute.SubAttributes, typed) {
					return false
				}
			case []interface{}:
				for _, item := range typed {
					if m, ok := item.(map[string]interface{}); ok {
						if !canDeleteAttributes(c, roles, attribute.SubAttributes, m) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}

func forbidWrite(c *gin.Context, attributeName string) bool {
	MakeError(c, http.StatusForbidden,
		"the caller is not allowed to write "+attributeName)
	return false
}

// valuesEqual compares two JSON-like values, normalising numeric types so an
// int64 and a float64 representing the same number compare equal.
func valuesEqual(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	rawA, errA := json.Marshal(a)
	rawB, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	var normA, normB interface{}
	if json.Unmarshal(rawA, &normA) != nil || json.Unmarshal(rawB, &normB) != nil {
		return false
	}
	return reflect.DeepEqual(normA, normB)
}

// canWrite reports whether any of the caller's roles appears in the attribute's
// $writer list, with "*" meaning any role.
func canWrite(roles []string, attribute Attribute) bool {
	if attribute.Writer == nil {
		return true
	}
	for _, allowed := range *attribute.Writer {
		if allowed == "*" {
			return true
		}
		for _, role := range roles {
			if allowed == role {
				return true
			}
		}
	}
	return false
}

func isReadOnly(attribute Attribute) bool {
	return strings.EqualFold(attribute.Mutability, "readOnly")
}

// isReturnable reports whether an attribute may appear in a response at all,
// before any role check. RFC 7643 7 defines "returned": "never" for values the
// service provider must not disclose, and a writeOnly attribute -- a password
// being the archetype -- is never returned either.
func isReturnable(attribute Attribute) bool {
	if strings.EqualFold(attribute.Returned, "never") {
		return false
	}
	return !strings.EqualFold(attribute.Mutability, "writeOnly")
}
