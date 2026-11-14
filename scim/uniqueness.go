package scim

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractUniqueClaims collects every unique attribute value present on element
// for resourceType, returning them in deterministic LockID order so multi-key
// acquisitions never deadlock.
func ExtractUniqueClaims(resourceType ResourceType, element map[string]interface{}) []UniqueClaim {
	if element == nil {
		return nil
	}
	var claims []UniqueClaim
	if coreSchema, ok := LookupSchema(resourceType.Schema); ok {
		claims = collectObjectUniqueClaims(claims, coreSchema.Attributes, element, "")
	}
	for _, extension := range resourceType.SchemaExtensions {
		rawExt, ok := lookupValueFold(element, extension.Schema)
		if !ok {
			continue
		}
		values, ok := rawExt.(map[string]interface{})
		if !ok {
			continue
		}
		if extSchema, ok := LookupSchema(extension.Schema); ok {
			claims = collectObjectUniqueClaims(claims, extSchema.Attributes, values, extSchema.ID+".")
		}
	}
	sort.Slice(claims, func(i, j int) bool {
		return claims[i].LockID < claims[j].LockID
	})
	return claims
}

func collectObjectUniqueClaims(out []UniqueClaim, attributes []Attribute, element map[string]interface{}, prefix string) []UniqueClaim {
	for _, attribute := range attributes {
		val, present := lookupValueFold(element, attribute.Name)
		if !present || val == nil {
			continue
		}
		path := prefix + attribute.Name
		if isUniqueAttribute(attribute) {
			key, ok := canonicalUniqueKey(attribute, val)
			if ok {
				out = append(out, UniqueClaim{
					Attribute: path,
					LockID:    uniqueLockID(path, key),
					Key:       key,
					RawValue:  val,
					CaseExact: attribute.CaseExact,
					Global:    strings.EqualFold(attribute.Uniqueness, "global"),
				})
			}
		}
		if !attribute.MultiValued && len(attribute.SubAttributes) > 0 {
			if subMap, ok := val.(map[string]interface{}); ok {
				out = collectObjectUniqueClaims(out, attribute.SubAttributes, subMap, path+".")
			}
		}
	}
	return out
}

func canonicalUniqueKey(attribute Attribute, value interface{}) (string, bool) {
	if s, ok := value.(string); ok {
		if !attribute.CaseExact {
			s = strings.ToLower(s)
		}
		raw, err := json.Marshal(s)
		if err != nil {
			return "", false
		}
		return string(raw), true
	}
	norm, err := normalizeValue(value)
	if err != nil {
		return "", false
	}
	raw, err := json.Marshal(norm)
	if err != nil {
		return "", false
	}
	return string(raw), true
}

func uniqueLockID(attributePath, canonicalKey string) string {
	sum := sha256.Sum256([]byte(attributePath + "\x00" + canonicalKey))
	return "u:" + hex.EncodeToString(sum[:])
}

// writeStoreUniqueError translates a UniqueConflictError to RFC 7644 409
// uniqueness and any other error to internalError.
func writeStoreUniqueError(c *gin.Context, resourceType ResourceType, err error) {
	var conflict *UniqueConflictError
	if errors.As(err, &conflict) {
		MakeTypedError(c, http.StatusConflict, "uniqueness",
			"a "+resourceType.Name+" with this "+conflict.Attribute+" already exists")
		return
	}
	internalError(c, err)
}

// EnforceUniqueness checks uniqueness against the Store for callers invoking
// validation outside CreateUnique/ReplaceUnique.
func EnforceUniqueness(c *gin.Context, resourceType ResourceType, element map[string]interface{}, currentID string) bool {
	bucket := storeBucketForContext(c, resourceType.Name)
	for _, claim := range ExtractUniqueClaims(resourceType, element) {
		existingID, err := DB.FindIDByAttribute(bucket, claim.Attribute, claim.RawValue)
		if err != nil {
			internalError(c, err)
			return false
		}
		if existingID != "" && existingID != currentID {
			MakeTypedError(c, http.StatusConflict, "uniqueness",
				"a "+resourceType.Name+" with this "+claim.Attribute+" already exists")
			return false
		}
	}
	return true
}

// isUniqueAttribute reports whether the schema asks for this attribute to be
// unique. Multi-valued attributes are skipped: uniqueness across the elements
// of an array is a different question, and nothing in the shipped schemas asks
// for it.
func isUniqueAttribute(attribute Attribute) bool {
	if attribute.MultiValued {
		return false
	}
	switch strings.ToLower(attribute.Uniqueness) {
	case "server", "global":
		return true
	default:
		return false
	}
}
