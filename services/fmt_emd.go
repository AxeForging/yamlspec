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
		fmt.Fprintf(&b, "## ✅ yamlspec: All %d assertions passed\n\n", s.TotalAssertions)
	} else {
		fmt.Fprintf(&b, "## ❌ yamlspec: %d of %d assertions failed\n\n", s.FailedAssertions, s.TotalAssertions)
	}

	for _, spec := range result.Specs {
		open := ""
		icon := "✅"
		if spec.Status != domain.StatusPassed {
			open = " open"
			icon = "❌"
		}

		fmt.Fprintf(&b, "<details%s>\n", open)
		fmt.Fprintf(&b, "<summary>%s <strong>%s</strong>", icon, spec.Name)
		if len(spec.Tags) > 0 {
			tags := make([]string, len(spec.Tags))
			for i, t := range spec.Tags {
				tags[i] = "<code>" + t + "</code>"
			}
			fmt.Fprintf(&b, " — %s", strings.Join(tags, " "))
		}
		b.WriteString("</summary>\n\n")

		if spec.Error != "" {
			fmt.Fprintf(&b, "**Error:** %s\n\n", spec.Error)
		} else {
			for _, desc := range spec.Describes {
				fmt.Fprintf(&b, "#### %s\n\n", desc.Name)
				if desc.Select != "" {
					fmt.Fprintf(&b, "`select: %s`\n\n", desc.Select)
				}

				for _, ar := range desc.Assertions {
					arIcon := "✅"
					if ar.Status != domain.StatusPassed {
						arIcon = "❌"
					}
					fmt.Fprintf(&b, "- %s %s", arIcon, ar.Should)
					if ar.Error != "" {
						fmt.Fprintf(&b, "\n  > %s", ar.Error)
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
	fmt.Fprintf(&b, "**%d** specs, **%d** assertions, **%d** passed, **%d** failed — %.2fs\n",
		s.TotalSpecs, s.TotalAssertions, s.PassedAssertions, s.FailedAssertions, s.Duration.Seconds())

	return []byte(b.String()), nil
}
