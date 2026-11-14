package scim

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/arturoeanton/goscim/scim/parser"
	"github.com/couchbase/gocb/v2"
)

var nonIdentChars = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

// AutoIndexEnabledFromEnv reports whether secondary GSI N1QL indexes should be
// automatically created from JSON schema metadata (default true unless
// SCIM_AUTO_INDEX=false or off).
func AutoIndexEnabledFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_AUTO_INDEX")))
	return v != "false" && v != "off" && v != "0"
}

// SecondaryIndexSpec describes a schema-derived Couchbase GSI index.
type SecondaryIndexSpec struct {
	Name string
	DDL  string
}

// BuildSchemaIndexSpecs derives secondary N1QL GSI index definitions for
// resourceType from its core and extension schemas:
//   - common `externalId` (exact) and `_scim_tenant` (multi-tenant filter)
//   - `uniqueness: server | global` scalar attributes (`LOWER(...)` when
//     `caseExact: false`, exact otherwise)
//   - top-level `multiValued: true` complex attributes carrying a `value`
//     sub-attribute (quantified `DISTINCT ARRAY` index for `ANY ... SATISFIES`)
func BuildSchemaIndexSpecs(rt ResourceType) []SecondaryIndexSpec {
	bucket := rt.Name
	quotedBucket := "`" + strings.ReplaceAll(bucket, "`", "``") + "`"
	prefix := sanitizeIndexToken(bucket)

	var specs []SecondaryIndexSpec
	seen := make(map[string]bool)
	addSpec := func(suffix, keyExpr string) {
		idxName := "idx_scim_" + prefix + "_" + sanitizeIndexToken(suffix)
		if len(idxName) > 60 {
			idxName = idxName[:60]
		}
		if seen[idxName] {
			return
		}
		seen[idxName] = true
		ddl := fmt.Sprintf("CREATE INDEX `%s` IF NOT EXISTS ON %s(%s)", idxName, quotedBucket, keyExpr)
		specs = append(specs, SecondaryIndexSpec{Name: idxName, DDL: ddl})
	}

	// Common RFC 7643 3.1 externalId + multi-tenant discriminator.
	addSpec("externalId", "`externalId`")
	addSpec("tenant", "`_scim_tenant`")

	if core, ok := LookupSchema(rt.Schema); ok {
		collectAttributeIndexSpecs(core.Attributes, "", addSpec)
	}
	for _, ext := range rt.SchemaExtensions {
		if extSchema, ok := LookupSchema(ext.Schema); ok {
			collectAttributeIndexSpecs(extSchema.Attributes, extSchema.ID+".", addSpec)
		}
	}
	return specs
}

func collectAttributeIndexSpecs(attrs []Attribute, pathPrefix string, add func(suffix, expr string)) {
	for _, attr := range attrs {
		fullPath := pathPrefix + attr.Name
		quoted := parser.AddQuote(fullPath)

		if isUniqueAttribute(attr) {
			if strings.EqualFold(attr.Type, "string") && !attr.CaseExact {
				add("uniq_"+fullPath, "LOWER("+quoted+")")
			} else {
				add("uniq_"+fullPath, quoted)
			}
		}

		if attr.MultiValued && strings.EqualFold(attr.Type, "complex") && pathPrefix == "" {
			if sub, hasVal := FindAttribute(attr.SubAttributes, "value"); hasVal {
				if strings.EqualFold(sub.Type, "string") && !sub.CaseExact {
					add("arr_"+attr.Name+"_lower",
						fmt.Sprintf("DISTINCT ARRAY LOWER(`_v`.`value`) FOR `_v` IN %s END", quoted))
				}
				add("arr_"+attr.Name+"_exact",
					fmt.Sprintf("DISTINCT ARRAY `_v`.`value` FOR `_v` IN %s END", quoted))
			}
		}

		if !attr.MultiValued && len(attr.SubAttributes) > 0 {
			collectAttributeIndexSpecs(attr.SubAttributes, fullPath+".", add)
		}
	}
}

func sanitizeIndexToken(s string) string {
	cleaned := strings.Trim(nonIdentChars.ReplaceAllString(s, "_"), "_")
	if cleaned == "" {
		return "attr"
	}
	return cleaned
}

// EnsureResourceIndexes provisions secondary N1QL GSI indexes derived from rt's
// schema when s is active and SCIM_AUTO_INDEX is enabled.
func (s *CouchbaseStore) EnsureResourceIndexes(rt ResourceType) error {
	if s == nil || s.cluster == nil || !AutoIndexEnabledFromEnv() {
		return nil
	}
	specs := BuildSchemaIndexSpecs(rt)
	for _, spec := range specs {
		if err := s.execDDLWithRetry(spec.Name, spec.DDL); err != nil {
			return err
		}
	}
	return nil
}

func (s *CouchbaseStore) execDDLWithRetry(name, ddl string) error {
	deadline := time.Now().Add(indexReadyTimeout)
	for attempt := 1; ; attempt++ {
		rows, err := s.cluster.Query(ddl, nil)
		if rows != nil {
			_ = rows.Close()
		}
		if err == nil || strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("creating secondary index %q after %d attempts: %w", name, attempt, err)
		}
		time.Sleep(indexRetryInterval)
	}
}

// ListBucketIndexNames returns the N1QL index names currently present on bucket
// (used by integration tests and operational diagnostics).
func (s *CouchbaseStore) ListBucketIndexNames(bucket string) ([]string, error) {
	if s == nil || s.cluster == nil {
		return nil, nil
	}
	_, phys := SplitScopedBucket(bucket)
	indexes, err := s.cluster.QueryIndexes().GetAllIndexes(phys, &gocb.GetAllQueryIndexesOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		out = append(out, idx.Name)
	}
	log.Printf("scim indexes on %s: %d active", phys, len(out))
	return out, nil
}
