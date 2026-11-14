package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/arturoeanton/goscim/scim/parser"
)

func TestArrayValueFiltersTranslateToN1QL(t *testing.T) {
	resolve := func(path string) string {
		canonical := map[string]string{
			"emails":         "emails",
			"emails.type":    "emails.type",
			"emails.value":   "emails.value",
			"emails.primary": "emails.primary",
			"userName":       "userName",
		}
		for k, v := range canonical {
			if strings.EqualFold(k, path) {
				return v
			}
		}
		return path
	}

	cases := []struct {
		name       string
		filter     string
		wantWhere  string
		wantParams []interface{}
	}{
		{
			name:       "simple equality inside brackets",
			filter:     `emails[type eq "work"]`,
			wantWhere:  "WHERE (ANY `_v0` IN `emails` SATISFIES `_v0`.`type` = $1 END)",
			wantParams: []interface{}{"work"},
		},
		{
			name:       "compound boolean and substring inside brackets with wildcard escaping",
			filter:     `emails[type eq "work" and (primary eq true or value ew "%_corp.com")]`,
			wantWhere:  `WHERE (ANY ` + "`_v0`" + ` IN ` + "`emails`" + ` SATISFIES ` + "`_v0`.`type`" + ` = $1 and (` + "`_v0`.`primary`" + ` = $2 or ` + "`_v0`.`value`" + ` LIKE $3 ESCAPE "\\") END)`,
			wantParams: []interface{}{"work", true, `%\%\_corp.com`},
		},
		{
			name:       "RFC 7644 3.4.2.2 trailing sub-attribute comparison",
			filter:     `emails[type eq "work"].value co "@corp.com"`,
			wantWhere:  `WHERE (ANY ` + "`_v0`" + ` IN ` + "`emails`" + ` SATISFIES (` + "`_v0`.`type`" + ` = $1) and ` + "`_v0`.`value`" + ` LIKE $2 ESCAPE "\\" END)`,
			wantParams: []interface{}{"work", "%@corp.com%"},
		},
		{
			name:       "RFC 7644 3.4.2.2 trailing sub-attribute presence",
			filter:     `emails[type eq "work"].value pr`,
			wantWhere:  "WHERE (ANY `_v0` IN `emails` SATISFIES (`_v0`.`type` = $1) and `_v0`.`value` IS NOT NULL END)",
			wantParams: []interface{}{"work"},
		},
		{
			name:       "case-insensitive NameResolver inside brackets",
			filter:     `EMAILS[TYPE eq "work"].VALUE eq "a@b.com"`,
			wantWhere:  "WHERE (ANY `_v0` IN `emails` SATISFIES (`_v0`.`type` = $1) and `_v0`.`value` = $2 END)",
			wantParams: []interface{}{"work", "a@b.com"},
		},
		{
			name:       "negated bracket expression",
			filter:     `not emails[type eq "home"]`,
			wantWhere:  "WHERE not (ANY `_v0` IN `emails` SATISFIES `_v0`.`type` = $1 END)",
			wantParams: []interface{}{"home"},
		},
		{
			name:       "quoted text resembling bracket suffix is untouched",
			filter:     `userName eq "a[type eq x].value co y"`,
			wantWhere:  "WHERE `userName` = $1",
			wantParams: []interface{}{"a[type eq x].value co y"},
		},
		{
			name:       "schema URN prefixed multi-valued attribute",
			filter:     `urn:ietf:params:scim:schemas:core:2.0:User.emails[type eq "work"]`,
			wantWhere:  "WHERE (ANY `_v0` IN `urn:ietf:params:scim:schemas:core:2.0:User`.`emails` SATISFIES `_v0`.`type` = $1 END)",
			wantParams: []interface{}{"work"},
		},
		{
			name:       "nested bracket scopes bind distinct iterator variables",
			filter:     `groups[type eq "direct" and roles[value eq "admin"]]`,
			wantWhere:  "WHERE (ANY `_v0` IN `groups` SATISFIES `_v0`.`type` = $1 and (ANY `_v1` IN `_v0`.`roles` SATISFIES `_v1`.`value` = $2 END) END)",
			wantParams: []interface{}{"direct", "admin"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := parser.FilterToN1QL("User", tc.filter, resolve)
			if err != nil {
				t.Fatalf("FilterToN1QL(%q) failed: %v", tc.filter, err)
			}
			wantPage := "SELECT * FROM `User` " + tc.wantWhere
			if q.Page != wantPage {
				t.Errorf("Page =\n  %s\nwant:\n  %s", q.Page, wantPage)
			}
			if !reflect.DeepEqual(q.Params, tc.wantParams) {
				t.Errorf("Params = %#v, want %#v", q.Params, tc.wantParams)
			}
		})
	}
}

func TestReferencedAttributesScopesBracketedSubAttributes(t *testing.T) {
	attrs, err := parser.ReferencedAttributes(`emails[type eq "work" and primary eq true].value co "@corp.com"`)
	if err != nil {
		t.Fatalf("ReferencedAttributes failed: %v", err)
	}
	want := []string{"emails", "emails.type", "emails.primary", "emails.value"}
	if !reflect.DeepEqual(attrs, want) {
		t.Errorf("ReferencedAttributes = %v, want %v", attrs, want)
	}
}

func TestInMemoryMatchAndEqualitySeeds(t *testing.T) {
	elem := map[string]interface{}{
		"type":    "Work",
		"value":   "jane@corp.example.com",
		"primary": true,
		"rank":    float64(3),
	}

	isCaseExact := func(path string) bool {
		return path == "value"
	}

	matchCases := []struct {
		filter string
		want   bool
	}{
		{`type eq "work"`, true},
		{`type eq "home"`, false},
		{`value eq "jane@corp.example.com"`, true},
		{`value eq "JANE@CORP.EXAMPLE.COM"`, false}, // caseExact for value
		{`value ew "example.com" and primary eq true`, true},
		{`rank ge 3 and rank lt 5`, true},
		{`display pr`, false},
		{`value pr`, true},
		{`not (type eq "home")`, true},
		{`missing ne "anything"`, true},
	}

	for _, tc := range matchCases {
		got, err := parser.Match(tc.filter, elem, isCaseExact)
		if err != nil {
			t.Fatalf("Match(%q) error: %v", tc.filter, err)
		}
		if got != tc.want {
			t.Errorf("Match(%q) = %v, want %v", tc.filter, got, tc.want)
		}
	}

	seeds, err := parser.EqualitySeeds(`type eq "work" and (primary eq true and rank eq 2)`)
	if err != nil {
		t.Fatalf("EqualitySeeds error: %v", err)
	}
	wantSeeds := map[string]interface{}{
		"type":    "work",
		"primary": true,
		"rank":    float64(2),
	}
	if !reflect.DeepEqual(seeds, wantSeeds) {
		t.Errorf("EqualitySeeds = %#v, want %#v", seeds, wantSeeds)
	}
}

func TestFilterToN1QLWithOptionsIdPPatterns(t *testing.T) {
	opts := parser.FilterOptions{
		Resolve: func(path string) string {
			switch strings.ToLower(path) {
			case "username", "urn:ietf:params:scim:schemas:core:2.0:user:username":
				return "userName"
			case "externalid":
				return "externalId"
			case "emails", "urn:ietf:params:scim:schemas:core:2.0:user.emails":
				return "emails"
			case "emails.value":
				return "emails.value"
			case "emails.type":
				return "emails.type"
			case "members":
				return "members"
			case "members.value":
				return "members.value"
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:employeenumber":
				return "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User.employeeNumber"
			case "urn:ietf:params:scim:schemas:extension:enterprise:2.0:user:manager.value":
				return "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User.manager.value"
			default:
				return path
			}
		},
		IsCaseExact: func(path string) bool {
			return strings.EqualFold(path, "externalId") || strings.EqualFold(path, "id")
		},
		IsMultiValued: func(path string) bool {
			switch strings.ToLower(path) {
			case "emails", "members", "urn:ietf:params:scim:schemas:core:2.0:user:emails":
				return true
			default:
				return false
			}
		},
	}

	cases := []struct {
		name       string
		bucket     string
		filter     string
		wantWhere  string
		wantParams []interface{}
	}{
		{
			name:       "case-insensitive userName equality",
			bucket:     "User",
			filter:     `userName eq "AdeleV@contoso.com"`,
			wantWhere:  "WHERE LOWER(`userName`) = LOWER($1)",
			wantParams: []interface{}{"AdeleV@contoso.com"},
		},
		{
			name:       "case-exact externalId equality",
			bucket:     "User",
			filter:     `externalId eq "entra-oid-123"`,
			wantWhere:  "WHERE `externalId` = $1",
			wantParams: []interface{}{"entra-oid-123"},
		},
		{
			name:       "colon-delimited EnterpriseUser extension path",
			bucket:     "User",
			filter:     `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber eq "EMP-9001"`,
			wantWhere:  "WHERE LOWER(`urn:ietf:params:scim:schemas:extension:enterprise:2.0:User`.`employeeNumber`) = LOWER($1)",
			wantParams: []interface{}{"EMP-9001"},
		},
		{
			name:       "colon-delimited EnterpriseUser nested sub-attribute path",
			bucket:     "User",
			filter:     `urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value eq "mgr-1"`,
			wantWhere:  "WHERE LOWER(`urn:ietf:params:scim:schemas:extension:enterprise:2.0:User`.`manager`.`value`) = LOWER($1)",
			wantParams: []interface{}{"mgr-1"},
		},
		{
			name:       "Okta/Keycloak dotted multi-valued emails.value filter",
			bucket:     "User",
			filter:     `emails.value eq "jane@example.com"`,
			wantWhere:  "WHERE (ANY `_v0` IN `emails` SATISFIES LOWER(`_v0`.`value`) = LOWER($1) END)",
			wantParams: []interface{}{"jane@example.com"},
		},
		{
			name:       "Keycloak dotted multi-valued members.value filter",
			bucket:     "Group",
			filter:     `members.value eq "user-42"`,
			wantWhere:  "WHERE (ANY `_v0` IN `members` SATISFIES LOWER(`_v0`.`value`) = LOWER($1) END)",
			wantParams: []interface{}{"user-42"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := parser.FilterToN1QLWithOptions(tc.bucket, tc.filter, opts)
			if err != nil {
				t.Fatalf("FilterToN1QLWithOptions(%q) failed: %v", tc.filter, err)
			}
			wantPage := "SELECT * FROM `" + tc.bucket + "` " + tc.wantWhere
			if q.Page != wantPage {
				t.Errorf("Page =\n  %s\nwant:\n  %s", q.Page, wantPage)
			}
			if !reflect.DeepEqual(q.Params, tc.wantParams) {
				t.Errorf("Params = %#v, want %#v", q.Params, tc.wantParams)
			}
		})
	}
}
