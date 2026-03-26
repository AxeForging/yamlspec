package services

import (
	"fmt"
	"strings"

	"github.com/AxeForging/yamlspec/domain"
)

// MarkdownFormatter outputs results as Markdown tables
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new MarkdownFormatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

func (f *MarkdownFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	var b strings.Builder

	b.WriteString("# Test Results\n\n")

	for _, spec := range result.Specs {
		statusEmoji := statusToEmoji(spec.Status)
		fmt.Fprintf(&b, "## %s %s\n\n", statusEmoji, spec.Name)

		if len(spec.Tags) > 0 {
			tags := make([]string, len(spec.Tags))
			for i, t := range spec.Tags {
				tags[i] = "`" + t + "`"
			}
			fmt.Fprintf(&b, "**Tags:** %s\n\n", strings.Join(tags, ", "))
		}

		if spec.Error != "" {
			fmt.Fprintf(&b, "**Error:** %s\n\n", spec.Error)
			continue
		}

		for _, desc := range spec.Describes {
			fmt.Fprintf(&b, "### %s\n\n", desc.Name)
			b.WriteString("| # | Assertion | Status | Details |\n")
			b.WriteString("|---|-----------|--------|---------|\n")

			for i, ar := range desc.Assertions {
				status := strings.ToUpper(string(ar.Status))
				detail := ""
				if ar.Error != "" {
					detail = ar.Error
				}
				fmt.Fprintf(&b, "| %d | %s | %s | %s |\n", i+1, ar.Should, status, detail)
			}
			b.WriteString("\n")
		}
	}

	// Summary
	s := result.Summary
	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Specs:** %d total, %d passed, %d failed\n", s.TotalSpecs, s.PassedSpecs, s.FailedSpecs)
	fmt.Fprintf(&b, "- **Assertions:** %d total, %d passed, %d failed\n", s.TotalAssertions, s.PassedAssertions, s.FailedAssertions)
	fmt.Fprintf(&b, "- **Duration:** %.2fs\n", s.Duration.Seconds())

	return []byte(b.String()), nil
}

func statusToEmoji(status domain.Status) string {
	switch status {
	case domain.StatusPassed:
		return "PASS"
	case domain.StatusFailed:
		return "FAIL"
	case domain.StatusError:
		return "ERROR"
	default:
		return "SKIP"
	}
}
