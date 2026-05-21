package domain

import "time"

// Status represents the outcome of a test
type Status string

const (
	StatusPassed  Status = "passed"
	StatusFailed  Status = "failed"
	StatusSkipped Status = "skipped"
	StatusError   Status = "error"
)

// SuiteResult is the top-level result of a full test run
type SuiteResult struct {
	Specs   []SpecResult `json:"specs" yaml:"specs"`
	Summary Summary      `json:"summary" yaml:"summary"`
}

// SpecResult is the result of one spec.yaml file
type SpecResult struct {
	Name          string           `json:"name" yaml:"name"`
	Tags          []string         `json:"tags,omitempty" yaml:"tags,omitempty"`
	Status        Status           `json:"status" yaml:"status"`
	Duration      time.Duration    `json:"duration" yaml:"duration"`
	Describes     []DescribeResult `json:"describes,omitempty" yaml:"describes,omitempty"`
	Error         string           `json:"error,omitempty" yaml:"error,omitempty"`
	SourceFile    string           `json:"source_file,omitempty" yaml:"source_file,omitempty"`
	SourceContent string           `json:"source_content,omitempty" yaml:"source_content,omitempty"`
	Manifests     []ManifestResult `json:"manifests,omitempty" yaml:"manifests,omitempty"`
}

// ManifestResult captures a manifest file as tested after pre_run rendering.
type ManifestResult struct {
	Path      string `json:"path" yaml:"path"`
	Content   string `json:"content" yaml:"content"`
	Documents int    `json:"documents" yaml:"documents"`
}

// DescribeResult is the result of one describe block
type DescribeResult struct {
	Name       string            `json:"name" yaml:"name"`
	Select     string            `json:"select,omitempty" yaml:"select,omitempty"`
	Status     Status            `json:"status" yaml:"status"`
	Assertions []AssertionResult `json:"assertions" yaml:"assertions"`
}

// AssertionResult is the result of a single assertion
type AssertionResult struct {
	Should     string      `json:"should" yaml:"should"`
	Status     Status      `json:"status" yaml:"status"`
	Expected   interface{} `json:"expected,omitempty" yaml:"expected,omitempty"`
	Actual     interface{} `json:"actual,omitempty" yaml:"actual,omitempty"`
	Error      string      `json:"error,omitempty" yaml:"error,omitempty"`
	SourceLine int         `json:"source_line,omitempty" yaml:"source_line,omitempty"`
}

// Summary aggregates counts across the run
type Summary struct {
	TotalSpecs       int           `json:"total_specs" yaml:"total_specs"`
	PassedSpecs      int           `json:"passed_specs" yaml:"passed_specs"`
	FailedSpecs      int           `json:"failed_specs" yaml:"failed_specs"`
	SkippedSpecs     int           `json:"skipped_specs" yaml:"skipped_specs"`
	ErrorSpecs       int           `json:"error_specs" yaml:"error_specs"`
	TotalAssertions  int           `json:"total_assertions" yaml:"total_assertions"`
	PassedAssertions int           `json:"passed_assertions" yaml:"passed_assertions"`
	FailedAssertions int           `json:"failed_assertions" yaml:"failed_assertions"`
	Duration         time.Duration `json:"duration" yaml:"duration"`
	Success          bool          `json:"success" yaml:"success"`
	Timestamp        time.Time     `json:"timestamp" yaml:"timestamp"`
}
