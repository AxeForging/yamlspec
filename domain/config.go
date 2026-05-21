package domain

import "time"

// Config holds all CLI configuration
type Config struct {
	TestDir       string
	Tags          []string
	Workers       int
	FailFast      bool
	Verbose       bool
	Quiet         bool
	PreRunTimeout time.Duration

	// Output format flags (file paths; empty means skip)
	JSONOutput     string
	YAMLOutput     string
	MarkdownOutput string
	EMDOutput      string
	HTMLOutput     string
	JUnitOutput    string

	// GitHubAnnotations emits ::error file=...,line=...:: lines on stdout for
	// failing assertions, so they appear inline in GitHub PR diffs. Auto-enabled
	// when GITHUB_ACTIONS=true.
	GitHubAnnotations bool
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		TestDir:       "tests",
		Workers:       1,
		PreRunTimeout: 60 * time.Second,
	}
}
