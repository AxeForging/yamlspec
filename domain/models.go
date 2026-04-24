package domain

// Spec represents a test specification file (spec.yaml)
type Spec struct {
	Name     string      `yaml:"name"`
	Tags     []string    `yaml:"tags,omitempty"`
	PreRun   []string    `yaml:"pre_run,omitempty"`
	Describe []DescBlock `yaml:"describe"`
}

// HasTag returns true if the spec has at least one of the given tags
func (s *Spec) HasTag(tags ...string) bool {
	for _, t := range tags {
		for _, st := range s.Tags {
			if st == t {
				return true
			}
		}
	}
	return false
}

// DescBlock groups related assertions under a selector
type DescBlock struct {
	Name   string      `yaml:"name"`
	Select string      `yaml:"select"`
	It     []Assertion `yaml:"it"`
}

// Assertion is a single RSpec-like check
type Assertion struct {
	Should string `yaml:"should"`
	Expect string `yaml:"expect"`

	// Equality
	ToEqual    interface{} `yaml:"toEqual,omitempty"`
	ToNotEqual interface{} `yaml:"toNotEqual,omitempty"`

	// Numeric comparison
	ToBeGreaterThan    *float64 `yaml:"toBeGreaterThan,omitempty"`
	ToBeLessThan       *float64 `yaml:"toBeLessThan,omitempty"`
	ToBeGreaterOrEqual *float64 `yaml:"toBeGreaterOrEqual,omitempty"`
	ToBeLessOrEqual    *float64 `yaml:"toBeLessOrEqual,omitempty"`

	// String
	ToContain      string `yaml:"toContain,omitempty"`
	ToNotContain   string `yaml:"toNotContain,omitempty"`
	ToStartWith    string `yaml:"toStartWith,omitempty"`
	ToEndWith      string `yaml:"toEndWith,omitempty"`
	ToNotStartWith string `yaml:"toNotStartWith,omitempty"`
	ToNotEndWith   string `yaml:"toNotEndWith,omitempty"`
	ToMatch        string `yaml:"toMatch,omitempty"`

	// Existence
	ToExist  *bool `yaml:"toExist,omitempty"`
	ToBeNull *bool `yaml:"toBeNull,omitempty"`

	// Object
	ToHaveKey string `yaml:"toHaveKey,omitempty"`

	// Set membership
	ToBeOneOf    []interface{} `yaml:"toBeOneOf,omitempty"`
	ToNotBeOneOf []interface{} `yaml:"toNotBeOneOf,omitempty"`

	// Array
	ToContainItem   interface{} `yaml:"toContainItem,omitempty"`
	ToHaveLength    *int        `yaml:"toHaveLength,omitempty"`
	ToHaveMinLength *int        `yaml:"toHaveMinLength,omitempty"`
	ToHaveMaxLength *int        `yaml:"toHaveMaxLength,omitempty"`
}

// HasAnyOperator returns true if the assertion has any operator set,
// including existence/null checks.
func (a *Assertion) HasAnyOperator() bool {
	return a.ToExist != nil || a.ToBeNull != nil || a.HasValueOperators()
}

// HasValueOperators returns true if the assertion has any value-based operators set
func (a *Assertion) HasValueOperators() bool {
	return a.ToEqual != nil ||
		a.ToNotEqual != nil ||
		a.ToBeGreaterThan != nil ||
		a.ToBeLessThan != nil ||
		a.ToBeGreaterOrEqual != nil ||
		a.ToBeLessOrEqual != nil ||
		a.ToContain != "" ||
		a.ToNotContain != "" ||
		a.ToStartWith != "" ||
		a.ToEndWith != "" ||
		a.ToNotStartWith != "" ||
		a.ToNotEndWith != "" ||
		a.ToMatch != "" ||
		a.ToHaveKey != "" ||
		len(a.ToBeOneOf) > 0 ||
		len(a.ToNotBeOneOf) > 0 ||
		a.ToContainItem != nil ||
		a.ToHaveLength != nil ||
		a.ToHaveMinLength != nil ||
		a.ToHaveMaxLength != nil
}
