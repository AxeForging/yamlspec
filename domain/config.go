package domain

// Config holds all CLI configuration
type Config struct {
	TestDir  string
	Tags     []string
	Workers  int
	FailFast bool
	Verbose  bool
	Quiet    bool

	// Output format flags (file paths; empty means skip)
	JSONOutput     string
	YAMLOutput     string
	MarkdownOutput string
	EMDOutput      string
	JUnitOutput    string
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		TestDir: "tests",
		Workers: 1,
	}
}
