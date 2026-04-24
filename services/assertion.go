package services

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/AxeForging/yamlspec/domain"
	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v3"
)

// AssertionEngine evaluates spec assertions against YAML manifests
type AssertionEngine struct{}

// NewAssertionEngine creates a new assertion engine
func NewAssertionEngine() *AssertionEngine {
	return &AssertionEngine{}
}

// Evaluate runs all describe blocks against the given manifests
func (ae *AssertionEngine) Evaluate(describes []domain.DescBlock, manifests []interface{}) []domain.DescribeResult {
	var results []domain.DescribeResult

	for _, desc := range describes {
		result := ae.evaluateDescribe(&desc, manifests)
		results = append(results, result)
	}

	return results
}

func (ae *AssertionEngine) evaluateDescribe(desc *domain.DescBlock, manifests []interface{}) domain.DescribeResult {
	result := domain.DescribeResult{
		Name:   desc.Name,
		Select: desc.Select,
		Status: domain.StatusPassed,
	}

	selected, err := ae.applySelector(desc.Select, manifests)
	if err != nil {
		result.Status = domain.StatusError
		result.Assertions = []domain.AssertionResult{{
			Should: "apply selector",
			Status: domain.StatusError,
			Error:  fmt.Sprintf("invalid selector '%s': %v", desc.Select, err),
		}}
		return result
	}

	if len(selected) == 0 {
		result.Status = domain.StatusFailed
		result.Assertions = []domain.AssertionResult{{
			Should: "match resources",
			Status: domain.StatusFailed,
			Error:  fmt.Sprintf("selector '%s' matched no resources", desc.Select),
		}}
		return result
	}

	for _, assertion := range desc.It {
		ar := ae.evaluateAssertion(&assertion, selected)
		result.Assertions = append(result.Assertions, ar)
		if ar.Status == domain.StatusFailed || ar.Status == domain.StatusError {
			result.Status = domain.StatusFailed
		}
	}

	return result
}

func (ae *AssertionEngine) evaluateAssertion(assertion *domain.Assertion, resources []interface{}) domain.AssertionResult {
	result := domain.AssertionResult{
		Should: assertion.Should,
		Status: domain.StatusPassed,
	}

	fieldPath := assertion.Expect

	for _, resource := range resources {
		values, err := ae.extractFieldValues(fieldPath, resource)
		if err != nil {
			result.Status = domain.StatusError
			result.Error = fmt.Sprintf("failed to access '%s': %v", fieldPath, err)
			return result
		}

		// No values produced: the path yielded nothing (missing field, or wildcard
		// over an empty/missing array). Treat as "does not exist" for a single
		// virtual value, so toExist / toBeNull semantics still work.
		if len(values) == 0 {
			if failed, err := ae.checkSingle(assertion, fieldPath, nil, false); failed {
				result.Status = domain.StatusFailed
				result.Error = err
				result.Actual = nil
				return result
			}
			continue
		}

		for _, fv := range values {
			result.Actual = fv.value
			if failed, err := ae.checkSingle(assertion, fieldPath, fv.value, fv.exists); failed {
				result.Status = domain.StatusFailed
				result.Error = err
				return result
			}
		}
	}

	return result
}

// checkSingle evaluates all operators against a single (value, exists) pair.
// Returns (true, errorMessage) if the assertion fails, (false, "") if it passes.
func (ae *AssertionEngine) checkSingle(assertion *domain.Assertion, fieldPath string, value interface{}, exists bool) (bool, string) {
	if assertion.ToExist != nil {
		if *assertion.ToExist && !exists {
			return true, fmt.Sprintf("expected '%s' to exist", fieldPath)
		}
		if !*assertion.ToExist && exists {
			return true, fmt.Sprintf("expected '%s' to not exist, got: %v", fieldPath, value)
		}
	}

	if assertion.ToBeNull != nil {
		isNull := exists && value == nil
		if *assertion.ToBeNull && !isNull {
			if !exists {
				return true, fmt.Sprintf("expected '%s' to be null, but field does not exist", fieldPath)
			}
			return true, fmt.Sprintf("expected '%s' to be null, got: %v", fieldPath, value)
		}
		if !*assertion.ToBeNull && isNull {
			return true, fmt.Sprintf("expected '%s' to not be null", fieldPath)
		}
	}

	if !exists && assertion.HasValueOperators() {
		return true, fmt.Sprintf("'%s' does not exist (cannot evaluate assertion)", fieldPath)
	}

	if err := ae.evaluateOperators(assertion, fieldPath, value); err != nil {
		return true, err.Error()
	}
	return false, ""
}

// evaluateOperators checks all operators specified on an assertion
func (ae *AssertionEngine) evaluateOperators(a *domain.Assertion, path string, value interface{}) error {
	if a.ToEqual != nil {
		if !ae.valuesEqual(value, a.ToEqual) {
			return fmt.Errorf("expected '%s' to equal %v, got %v", path, a.ToEqual, value)
		}
	}

	if a.ToNotEqual != nil {
		if ae.valuesEqual(value, a.ToNotEqual) {
			return fmt.Errorf("expected '%s' to not equal %v", path, a.ToNotEqual)
		}
	}

	if a.ToBeGreaterThan != nil {
		n, err := ae.toFloat64(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a number (%T), cannot use toBeGreaterThan", path, value)
		}
		if n <= *a.ToBeGreaterThan {
			return fmt.Errorf("expected '%s' > %v, got %v", path, *a.ToBeGreaterThan, n)
		}
	}

	if a.ToBeLessThan != nil {
		n, err := ae.toFloat64(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a number (%T), cannot use toBeLessThan", path, value)
		}
		if n >= *a.ToBeLessThan {
			return fmt.Errorf("expected '%s' < %v, got %v", path, *a.ToBeLessThan, n)
		}
	}

	if a.ToBeGreaterOrEqual != nil {
		n, err := ae.toFloat64(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a number (%T), cannot use toBeGreaterOrEqual", path, value)
		}
		if n < *a.ToBeGreaterOrEqual {
			return fmt.Errorf("expected '%s' >= %v, got %v", path, *a.ToBeGreaterOrEqual, n)
		}
	}

	if a.ToBeLessOrEqual != nil {
		n, err := ae.toFloat64(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a number (%T), cannot use toBeLessOrEqual", path, value)
		}
		if n > *a.ToBeLessOrEqual {
			return fmt.Errorf("expected '%s' <= %v, got %v", path, *a.ToBeLessOrEqual, n)
		}
	}

	if a.ToContain != "" {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("'%s' is not a string (%T), cannot use toContain", path, value)
		}
		if !strings.Contains(s, a.ToContain) {
			return fmt.Errorf("expected '%s' to contain '%s', got '%s'", path, a.ToContain, s)
		}
	}

	if a.ToNotContain != "" {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("'%s' is not a string (%T), cannot use toNotContain", path, value)
		}
		if strings.Contains(s, a.ToNotContain) {
			return fmt.Errorf("expected '%s' to not contain '%s', got '%s'", path, a.ToNotContain, s)
		}
	}

	if a.ToStartWith != "" {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("'%s' is not a string (%T), cannot use toStartWith", path, value)
		}
		if !strings.HasPrefix(s, a.ToStartWith) {
			return fmt.Errorf("expected '%s' to start with '%s', got '%s'", path, a.ToStartWith, s)
		}
	}

	if a.ToEndWith != "" {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("'%s' is not a string (%T), cannot use toEndWith", path, value)
		}
		if !strings.HasSuffix(s, a.ToEndWith) {
			return fmt.Errorf("expected '%s' to end with '%s', got '%s'", path, a.ToEndWith, s)
		}
	}

	if a.ToNotStartWith != "" {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("'%s' is not a string (%T), cannot use toNotStartWith", path, value)
		}
		if strings.HasPrefix(s, a.ToNotStartWith) {
			return fmt.Errorf("expected '%s' to not start with '%s', got '%s'", path, a.ToNotStartWith, s)
		}
	}

	if a.ToNotEndWith != "" {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("'%s' is not a string (%T), cannot use toNotEndWith", path, value)
		}
		if strings.HasSuffix(s, a.ToNotEndWith) {
			return fmt.Errorf("expected '%s' to not end with '%s', got '%s'", path, a.ToNotEndWith, s)
		}
	}

	if a.ToMatch != "" {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("'%s' is not a string (%T), cannot use toMatch", path, value)
		}
		matched, err := regexp.MatchString(a.ToMatch, s)
		if err != nil {
			return fmt.Errorf("invalid regex '%s': %v", a.ToMatch, err)
		}
		if !matched {
			return fmt.Errorf("expected '%s' to match '%s', got '%s'", path, a.ToMatch, s)
		}
	}

	if a.ToHaveKey != "" {
		m, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("'%s' is not an object (%T), cannot use toHaveKey", path, value)
		}
		if _, exists := m[a.ToHaveKey]; !exists {
			return fmt.Errorf("expected '%s' to have key '%s'", path, a.ToHaveKey)
		}
	}

	if len(a.ToBeOneOf) > 0 {
		found := false
		for _, allowed := range a.ToBeOneOf {
			if ae.valuesEqual(value, allowed) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected '%s' to be one of %v, got %v", path, a.ToBeOneOf, value)
		}
	}

	if len(a.ToNotBeOneOf) > 0 {
		for _, disallowed := range a.ToNotBeOneOf {
			if ae.valuesEqual(value, disallowed) {
				return fmt.Errorf("expected '%s' to not be one of %v, got %v", path, a.ToNotBeOneOf, value)
			}
		}
	}

	if a.ToContainItem != nil {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("'%s' is not an array (%T), cannot use toContainItem", path, value)
		}
		found := false
		for _, item := range arr {
			if ae.valuesEqual(item, a.ToContainItem) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected '%s' to contain %v", path, a.ToContainItem)
		}
	}

	if a.ToHaveLength != nil {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("'%s' is not an array (%T), cannot use toHaveLength", path, value)
		}
		if len(arr) != *a.ToHaveLength {
			return fmt.Errorf("expected '%s' length %d, got %d", path, *a.ToHaveLength, len(arr))
		}
	}

	if a.ToHaveMinLength != nil {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("'%s' is not an array (%T), cannot use toHaveMinLength", path, value)
		}
		if len(arr) < *a.ToHaveMinLength {
			return fmt.Errorf("expected '%s' length >= %d, got %d", path, *a.ToHaveMinLength, len(arr))
		}
	}

	if a.ToHaveMaxLength != nil {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("'%s' is not an array (%T), cannot use toHaveMaxLength", path, value)
		}
		if len(arr) > *a.ToHaveMaxLength {
			return fmt.Errorf("expected '%s' length <= %d, got %d", path, *a.ToHaveMaxLength, len(arr))
		}
	}

	return nil
}

// applySelector filters manifests using a JQ expression
func (ae *AssertionEngine) applySelector(selector string, manifests []interface{}) ([]interface{}, error) {
	if selector == "" {
		return manifests, nil
	}

	query, err := gojq.Parse(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid JQ selector: %w", err)
	}

	var results []interface{}
	for _, manifest := range manifests {
		iter := query.Run(manifest)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return nil, fmt.Errorf("JQ error: %w", err)
			}
			if v != nil {
				results = append(results, v)
			}
		}
	}

	return results, nil
}

// fieldValue is one result of extracting a path from a resource.
// For wildcard paths (e.g. containers[*].image) extractFieldValues returns
// one entry per matching element; for simple paths it returns exactly one
// entry when the path produces a result, or nothing when it doesn't.
type fieldValue struct {
	value  interface{}
	exists bool
}

// extractFieldValues pulls all values at the given path from a resource.
// It distinguishes "missing key" from "present but null" by probing the
// parent object with has() in a separate gojq query.
func (ae *AssertionEngine) extractFieldValues(fieldPath string, resource interface{}) ([]fieldValue, error) {
	jqPath := ae.fieldPathToJQ(fieldPath)

	query, err := gojq.Parse(jqPath)
	if err != nil {
		return nil, fmt.Errorf("invalid field path: %w", err)
	}

	var raw []interface{}
	iter := query.Run(resource)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if e, isErr := v.(error); isErr {
			// gojq path errors (e.g. indexing a number) mean the path cannot
			// be reached — treat as producing no results, let existence handle it.
			_ = e
			continue
		}
		raw = append(raw, v)
	}

	// Wildcard path: each produced element is considered to exist.
	if isWildcardPath(jqPath) {
		results := make([]fieldValue, 0, len(raw))
		for _, v := range raw {
			results = append(results, fieldValue{value: v, exists: true})
		}
		return results, nil
	}

	// Simple path: there should be at most one result. Probe parent for
	// real existence so present-null is distinguished from missing.
	if len(raw) == 0 {
		return nil, nil
	}
	exists, err := ae.probeExistence(jqPath, resource)
	if err != nil {
		return nil, err
	}
	return []fieldValue{{value: raw[0], exists: exists}}, nil
}

// probeExistence runs a companion query against the parent of the path
// that returns true iff the final key is actually present on the parent.
// Falls back to (value != nil) if the path can't be split.
func (ae *AssertionEngine) probeExistence(jqPath string, resource interface{}) (bool, error) {
	parent, probe, ok := splitLastSegment(jqPath)
	if !ok {
		// Fallback: evaluate the original path and treat any non-nil as existing.
		q, err := gojq.Parse(jqPath)
		if err != nil {
			return false, nil
		}
		iter := q.Run(resource)
		if v, ok := iter.Next(); ok {
			if _, isErr := v.(error); isErr {
				return false, nil
			}
			return v != nil, nil
		}
		return false, nil
	}

	probeExpr := parent + " | " + probe
	q, err := gojq.Parse(probeExpr)
	if err != nil {
		return false, nil
	}
	iter := q.Run(resource)
	v, ok := iter.Next()
	if !ok {
		return false, nil
	}
	if _, isErr := v.(error); isErr {
		return false, nil
	}
	b, _ := v.(bool)
	return b, nil
}

// fieldPathToJQ converts a field path to JQ syntax. Normalizes [*] (a
// yamlspec-user convenience) to [] (jq's array iteration) regardless of
// whether the path already has a leading dot.
func (ae *AssertionEngine) fieldPathToJQ(fieldPath string) string {
	path := strings.ReplaceAll(fieldPath, "[*]", "[]")
	if strings.HasPrefix(path, ".") {
		return path
	}
	return "." + path
}

// isWildcardPath returns true if a jq path contains an unconstrained array
// iteration that can produce multiple results.
func isWildcardPath(jqPath string) bool {
	inQuote := byte(0)
	for i := 0; i < len(jqPath); i++ {
		c := jqPath[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			inQuote = c
			continue
		}
		if c == '[' && i+1 < len(jqPath) && jqPath[i+1] == ']' {
			return true
		}
	}
	return false
}

// splitLastSegment splits a simple jq path into (parentExpr, probeExpr).
// probeExpr is a jq fragment that, when piped to parentExpr, returns true
// iff the final segment is present on that parent.
// Returns ok=false when the path is a wildcard path, the root (.), or
// can't be parsed — callers must fall back to value-based existence.
func splitLastSegment(jqPath string) (string, string, bool) {
	if jqPath == "" || jqPath == "." {
		return "", "", false
	}
	if isWildcardPath(jqPath) {
		return "", "", false
	}

	p := jqPath
	if p[0] == '.' {
		p = p[1:]
	}
	if p == "" {
		return "", "", false
	}

	// Walk to find the last top-level boundary ('.' or '[').
	inQuote := byte(0)
	lastBoundary := -1
	for i := 0; i < len(p); i++ {
		c := p[i]
		if inQuote != 0 {
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			inQuote = c
			continue
		}
		if c == '.' || c == '[' {
			lastBoundary = i
		}
	}

	var parent, last string
	if lastBoundary == -1 {
		parent = "."
		last = "." + p
	} else {
		if lastBoundary == 0 {
			parent = "."
		} else {
			parent = "." + p[:lastBoundary]
		}
		last = p[lastBoundary:]
	}

	if strings.HasPrefix(last, ".") {
		key := strings.TrimPrefix(last, ".")
		if key == "" {
			return "", "", false
		}
		return parent, fmt.Sprintf(`(type == "object" and has(%q))`, key), true
	}

	if strings.HasPrefix(last, "[") && strings.HasSuffix(last, "]") {
		inner := last[1 : len(last)-1]
		if len(inner) >= 2 && (inner[0] == '"' || inner[0] == '\'') && inner[0] == inner[len(inner)-1] {
			key := inner[1 : len(inner)-1]
			return parent, fmt.Sprintf(`(type == "object" and has(%q))`, key), true
		}
		if idx, err := strconv.Atoi(inner); err == nil {
			return parent, fmt.Sprintf(`(type == "array" and length > %d)`, idx), true
		}
	}
	return "", "", false
}

// valuesEqual compares two values for equality with type coercion
func (ae *AssertionEngine) valuesEqual(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}

	// Try numeric comparison (YAML may parse ints, expected may be float or vice versa)
	aNum, aErr := ae.toFloat64(a)
	bNum, bErr := ae.toFloat64(b)
	if aErr == nil && bErr == nil {
		return aNum == bNum
	}

	// Try string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return aStr == bStr
}

// toFloat64 converts a value to float64
func (ae *AssertionEngine) toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", v)
	}
}

// ParseManifests parses multi-document YAML into a slice of interfaces.
// Uses a real YAML decoder rather than splitting on "---" so that "---"
// appearing inside literal blocks, comments, or strings is not mistaken
// for a document separator.
func ParseManifests(yamlContent string) ([]interface{}, error) {
	decoder := yaml.NewDecoder(strings.NewReader(yamlContent))

	var manifests []interface{}
	for {
		var manifest interface{}
		err := decoder.Decode(&manifest)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		if manifest != nil {
			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}
