package services

import (
	"fmt"
	"strings"

	"github.com/AxeForging/yamlspec/domain"
)

// EMDFormatter outputs enriched markdown with collapsible sections for GitHub PR comments
type EMDFormatter struct{}

// NewEMDFormatter creates a new EMDFormatter
func NewEMDFormatter() *EMDFormatter {
	return &EMDFormatter{}
}

func (f *EMDFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	var b strings.Builder

	s := result.Summary
	if s.Success {
		b.WriteString(fmt.Sprintf("## ✅ yamlspec: All %d assertions passed\n\n", s.TotalAssertions))
	} else {
		b.WriteString(fmt.Sprintf("## ❌ yamlspec: %d of %d assertions failed\n\n", s.FailedAssertions, s.TotalAssertions))
	}

	for _, spec := range result.Specs {
		open := ""
		icon := "✅"
		if spec.Status != domain.StatusPassed {
			open = " open"
			icon = "❌"
		}

		b.WriteString(fmt.Sprintf("<details%s>\n", open))
		b.WriteString(fmt.Sprintf("<summary>%s <strong>%s</strong>", icon, spec.Name))
		if len(spec.Tags) > 0 {
			tags := make([]string, len(spec.Tags))
			for i, t := range spec.Tags {
				tags[i] = "<code>" + t + "</code>"
			}
			b.WriteString(fmt.Sprintf(" — %s", strings.Join(tags, " ")))
		}
		b.WriteString("</summary>\n\n")

		if spec.Error != "" {
			b.WriteString(fmt.Sprintf("**Error:** %s\n\n", spec.Error))
		} else {
			for _, desc := range spec.Describes {
				b.WriteString(fmt.Sprintf("#### %s\n\n", desc.Name))
				if desc.Select != "" {
					b.WriteString(fmt.Sprintf("`select: %s`\n\n", desc.Select))
				}

				for _, ar := range desc.Assertions {
					arIcon := "✅"
					if ar.Status != domain.StatusPassed {
						arIcon = "❌"
					}
					b.WriteString(fmt.Sprintf("- %s %s", arIcon, ar.Should))
					if ar.Error != "" {
						b.WriteString(fmt.Sprintf("\n  > %s", ar.Error))
					}
					b.WriteString("\n")
				}
				b.WriteString("\n")
			}
		}

		b.WriteString("</details>\n\n")
	}

	// Summary footer
	b.WriteString("---\n\n")
	b.WriteString(fmt.Sprintf("**%d** specs, **%d** assertions, **%d** passed, **%d** failed — %.2fs\n",
		s.TotalSpecs, s.TotalAssertions, s.PassedAssertions, s.FailedAssertions, s.Duration.Seconds()))

	return []byte(b.String()), nil
}
