package services

import (
	"fmt"
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
		fieldValue, exists, err := ae.extractFieldValue(fieldPath, resource)
		if err != nil {
			result.Status = domain.StatusError
			result.Error = fmt.Sprintf("failed to access '%s': %v", fieldPath, err)
			return result
		}

		result.Actual = fieldValue

		// Existence checks
		if assertion.ToExist != nil {
			if *assertion.ToExist && !exists {
				result.Status = domain.StatusFailed
				result.Error = fmt.Sprintf("expected '%s' to exist", fieldPath)
				return result
			}
			if !*assertion.ToExist && exists {
				result.Status = domain.StatusFailed
				result.Error = fmt.Sprintf("expected '%s' to not exist, got: %v", fieldPath, fieldValue)
				return result
			}
		}

		// Null checks
		if assertion.ToBeNull != nil {
			isNull := fieldValue == nil
			if *assertion.ToBeNull && !isNull {
				result.Status = domain.StatusFailed
				result.Error = fmt.Sprintf("expected '%s' to be null, got: %v", fieldPath, fieldValue)
				return result
			}
			if !*assertion.ToBeNull && isNull {
				result.Status = domain.StatusFailed
				result.Error = fmt.Sprintf("expected '%s' to not be null", fieldPath)
				return result
			}
		}

		// If field doesn't exist and we have value operators, fail
		if !exists && assertion.HasValueOperators() {
			result.Status = domain.StatusFailed
			result.Error = fmt.Sprintf("'%s' does not exist (cannot evaluate assertion)", fieldPath)
			return result
		}

		if err := ae.evaluateOperators(assertion, fieldPath, fieldValue); err != nil {
			result.Status = domain.StatusFailed
			result.Error = err.Error()
			return result
		}
	}

	return result
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

// extractFieldValue extracts a value from a resource using a field path
func (ae *AssertionEngine) extractFieldValue(fieldPath string, resource interface{}) (interface{}, bool, error) {
	jqPath := ae.fieldPathToJQ(fieldPath)

	query, err := gojq.Parse(jqPath)
	if err != nil {
		return nil, false, fmt.Errorf("invalid field path: %w", err)
	}

	iter := query.Run(resource)
	v, ok := iter.Next()
	if !ok {
		return nil, false, nil
	}

	if _, ok := v.(error); ok {
		return nil, false, nil
	}

	return v, v != nil, nil
}

// fieldPathToJQ converts a field path to JQ syntax
func (ae *AssertionEngine) fieldPathToJQ(fieldPath string) string {
	if strings.HasPrefix(fieldPath, ".") {
		return fieldPath
	}
	path := strings.ReplaceAll(fieldPath, "[*]", "[]")
	return "." + path
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

// ParseManifests parses multi-document YAML into a slice of interfaces
func ParseManifests(yamlContent string) ([]interface{}, error) {
	var manifests []interface{}

	documents := strings.Split(yamlContent, "---")
	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" || doc == "null" {
			continue
		}

		var manifest interface{}
		if err := yaml.Unmarshal([]byte(doc), &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		if manifest != nil {
			manifests = append(manifests, manifest)
		}
	}

	return manifests, nil
}
