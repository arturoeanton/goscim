package parser

import (
	"strconv"
	"strings"
)

// CompileMatcher parses filter once and returns a predicate that evaluates it
// against decoded JSON objects in memory. It is used by PATCH value-selection
// paths such as emails[type eq "work"].value.
func CompileMatcher(filter string, isCaseExact func(attrPath string) bool) (func(map[string]interface{}) bool, error) {
	tree, _, err := parseFilter(filter)
	if err != nil {
		return nil, err
	}
	root := tree.AllExpression()[0]
	return func(doc map[string]interface{}) bool {
		return evalExpr(root, doc, "", isCaseExact)
	}, nil
}

// Match evaluates filter against doc in memory.
func Match(filter string, doc map[string]interface{}, isCaseExact func(attrPath string) bool) (bool, error) {
	matcher, err := CompileMatcher(filter, isCaseExact)
	if err != nil {
		return false, err
	}
	return matcher(doc), nil
}

// EqualitySeeds extracts conjunctive top-level equality predicates (such as
// `type eq "work"` or `type eq "work" and primary eq true`) so a PATCH "add"
// targeting `emails[type eq "work"].value` on a resource with no matching
// element can synthesize the new array entry expected by clients like Entra ID.
func EqualitySeeds(filter string) (map[string]interface{}, error) {
	tree, _, err := parseFilter(filter)
	if err != nil {
		return nil, err
	}
	seeds := make(map[string]interface{})
	collectEqualitySeeds(tree.AllExpression()[0], seeds)
	return seeds, nil
}

func collectEqualitySeeds(expr IExpressionContext, out map[string]interface{}) {
	switch ctx := expr.(type) {
	case *LPAREN_EXPR_RPARENContext:
		collectEqualitySeeds(ctx.Expression(), out)
	case *EXPR_AND_EXPRContext:
		collectEqualitySeeds(ctx.Expression(0), out)
		collectEqualitySeeds(ctx.Expression(1), out)
	case *ATTR_OPER_CRITERIAContext:
		if strings.EqualFold(ctx.Operator().GetText(), "eq") {
			attr := ctx.ATTRNAME().GetText()
			if !strings.Contains(attr, ".") {
				out[attr] = unquoteCriteria(ctx.Criteria().GetText())
			}
		}
	case *ATTR_OPER_VALUEContext:
		if strings.EqualFold(ctx.Operator().GetText(), "eq") {
			attr := ctx.ATTRNAME().GetText()
			if !strings.Contains(attr, ".") {
				if val, ok := parseCriteriaValue(ctx.CriteriaValue().GetText()); ok {
					out[attr] = val
				}
			}
		}
	}
}

func evalExpr(expr IExpressionContext, doc map[string]interface{}, scopePrefix string, isCaseExact func(string) bool) bool {
	if expr == nil || doc == nil {
		return false
	}
	switch ctx := expr.(type) {
	case *LPAREN_EXPR_RPARENContext:
		return evalExpr(ctx.Expression(), doc, scopePrefix, isCaseExact)
	case *NOT_EXPRContext:
		return !evalExpr(ctx.Expression(), doc, scopePrefix, isCaseExact)
	case *EXPR_AND_EXPRContext:
		return evalExpr(ctx.Expression(0), doc, scopePrefix, isCaseExact) &&
			evalExpr(ctx.Expression(1), doc, scopePrefix, isCaseExact)
	case *EXPR_OR_EXPRContext:
		return evalExpr(ctx.Expression(0), doc, scopePrefix, isCaseExact) ||
			evalExpr(ctx.Expression(1), doc, scopePrefix, isCaseExact)
	case *LBRAC_EXPR_RBRACContext:
		attrName := ctx.ATTRNAME().GetText()
		rawVal, ok := lookupEvalPath(doc, attrName)
		if !ok {
			return false
		}
		nextScope := attrName
		if scopePrefix != "" {
			nextScope = scopePrefix + "." + attrName
		}
		items, ok := rawVal.([]interface{})
		if !ok {
			if m, ok := rawVal.(map[string]interface{}); ok {
				return evalExpr(ctx.Expression(), m, nextScope, isCaseExact)
			}
			return false
		}
		for _, item := range items {
			if elemMap, ok := item.(map[string]interface{}); ok {
				if evalExpr(ctx.Expression(), elemMap, nextScope, isCaseExact) {
					return true
				}
			}
		}
		return false
	case *ATTR_PRContext:
		attrName := ctx.ATTRNAME().GetText()
		val, ok := lookupEvalPath(doc, attrName)
		if !ok || val == nil {
			return false
		}
		switch typed := val.(type) {
		case string:
			return typed != ""
		case []interface{}:
			return len(typed) > 0
		case map[string]interface{}:
			return len(typed) > 0
		default:
			return true
		}
	case *ATTR_OPER_CRITERIAContext:
		attrName := ctx.ATTRNAME().GetText()
		op := strings.ToLower(ctx.Operator().GetText())
		rhs := unquoteCriteria(ctx.Criteria().GetText())
		fullPath := attrName
		if scopePrefix != "" {
			fullPath = scopePrefix + "." + attrName
		}
		exact := isCaseExact != nil && isCaseExact(fullPath)
		lhs, present := lookupEvalPath(doc, attrName)
		return compareStringPredicate(lhs, present, op, rhs, exact)
	case *ATTR_OPER_VALUEContext:
		attrName := ctx.ATTRNAME().GetText()
		op := strings.ToLower(ctx.Operator().GetText())
		rhs, ok := parseCriteriaValue(ctx.CriteriaValue().GetText())
		if !ok {
			return false
		}
		lhs, present := lookupEvalPath(doc, attrName)
		return compareScalarPredicate(lhs, present, op, rhs)
	}
	return false
}

func unquoteCriteria(raw string) string {
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		return raw[1 : len(raw)-1]
	}
	return raw
}

func parseCriteriaValue(raw string) (interface{}, bool) {
	switch raw {
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, false
		}
		return n, true
	}
}

func lookupEvalPath(doc map[string]interface{}, attrPath string) (interface{}, bool) {
	urn, rest := SplitURNPath(attrPath)
	segments := make([]string, 0, 3)
	if urn != "" {
		segments = append(segments, urn)
	}
	segments = append(segments, strings.Split(rest, ".")...)

	var current interface{} = doc
	for _, seg := range segments {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, found := lookupMapFold(m, seg)
		if !found {
			return nil, false
		}
		current = val
	}
	return current, true
}

func lookupMapFold(m map[string]interface{}, key string) (interface{}, bool) {
	if val, ok := m[key]; ok {
		return val, true
	}
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

func compareStringPredicate(lhs interface{}, present bool, op, rhs string, caseExact bool) bool {
	if !present || lhs == nil {
		return op == "ne"
	}
	lhsStr, ok := lhs.(string)
	if !ok {
		return op == "ne"
	}
	a, b := lhsStr, rhs
	if !caseExact {
		a = strings.ToLower(a)
		b = strings.ToLower(b)
	}
	switch op {
	case "eq":
		return a == b
	case "ne":
		return a != b
	case "co":
		return strings.Contains(a, b)
	case "sw":
		return strings.HasPrefix(a, b)
	case "ew":
		return strings.HasSuffix(a, b)
	case "gt":
		return a > b
	case "ge":
		return a >= b
	case "lt":
		return a < b
	case "le":
		return a <= b
	default:
		return false
	}
}

func compareScalarPredicate(lhs interface{}, present bool, op string, rhs interface{}) bool {
	if !present || lhs == nil {
		return op == "ne"
	}
	switch rhsVal := rhs.(type) {
	case bool:
		lhsBool, ok := lhs.(bool)
		if !ok {
			return op == "ne"
		}
		switch op {
		case "eq":
			return lhsBool == rhsVal
		case "ne":
			return lhsBool != rhsVal
		default:
			return false
		}
	case float64:
		lhsNum, ok := toFloat64(lhs)
		if !ok {
			return op == "ne"
		}
		switch op {
		case "eq":
			return lhsNum == rhsVal
		case "ne":
			return lhsNum != rhsVal
		case "gt":
			return lhsNum > rhsVal
		case "ge":
			return lhsNum >= rhsVal
		case "lt":
			return lhsNum < rhsVal
		case "le":
			return lhsNum <= rhsVal
		default:
			return false
		}
	default:
		return false
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}
