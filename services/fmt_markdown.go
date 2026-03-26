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
		b.WriteString(fmt.Sprintf("## %s %s\n\n", statusEmoji, spec.Name))

		if len(spec.Tags) > 0 {
			tags := make([]string, len(spec.Tags))
			for i, t := range spec.Tags {
				tags[i] = "`" + t + "`"
			}
			b.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(tags, ", ")))
		}

		if spec.Error != "" {
			b.WriteString(fmt.Sprintf("**Error:** %s\n\n", spec.Error))
			continue
		}

		for _, desc := range spec.Describes {
			b.WriteString(fmt.Sprintf("### %s\n\n", desc.Name))
			b.WriteString("| # | Assertion | Status | Details |\n")
			b.WriteString("|---|-----------|--------|---------|\n")

			for i, ar := range desc.Assertions {
				status := strings.ToUpper(string(ar.Status))
				detail := ""
				if ar.Error != "" {
					detail = ar.Error
				}
				b.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", i+1, ar.Should, status, detail))
			}
			b.WriteString("\n")
		}
	}

	// Summary
	s := result.Summary
	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- **Specs:** %d total, %d passed, %d failed\n", s.TotalSpecs, s.PassedSpecs, s.FailedSpecs))
	b.WriteString(fmt.Sprintf("- **Assertions:** %d total, %d passed, %d failed\n", s.TotalAssertions, s.PassedAssertions, s.FailedAssertions))
	b.WriteString(fmt.Sprintf("- **Duration:** %.2fs\n", s.Duration.Seconds()))

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
