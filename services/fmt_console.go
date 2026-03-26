package services

import (
	"fmt"
	"strings"

	"github.com/AxeForging/yamlspec/domain"
	"github.com/AxeForging/yamlspec/helpers"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
	colorBold   = "\033[1m"
)

// ConsoleFormatter produces RSpec-like colored terminal output
type ConsoleFormatter struct {
	verbose bool
	quiet   bool
	color   bool
}

// NewConsoleFormatter creates a new ConsoleFormatter
func NewConsoleFormatter(verbose, quiet bool) *ConsoleFormatter {
	return &ConsoleFormatter{
		verbose: verbose,
		quiet:   quiet,
		color:   helpers.IsTerminal(),
	}
}

func (f *ConsoleFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	var b strings.Builder

	if !f.quiet {
		b.WriteString("\n")
		for _, spec := range result.Specs {
			f.writeSpec(&b, &spec)
		}

		// Failures section
		failures := f.collectFailures(result)
		if len(failures) > 0 {
			b.WriteString("\n")
			b.WriteString(f.styled("Failures:", colorRed, colorBold))
			b.WriteString("\n\n")
			for i, fail := range failures {
				fmt.Fprintf(&b, "  %s%d) %s%s\n", colorRed, i+1, fail.path, colorReset)
				fmt.Fprintf(&b, "     %s%s%s\n", colorDim, fail.error, colorReset)
				b.WriteString("\n")
			}
		}
	}

	// Summary line
	b.WriteString(f.summaryLine(&result.Summary))
	b.WriteString("\n")

	return []byte(b.String()), nil
}

func (f *ConsoleFormatter) writeSpec(b *strings.Builder, spec *domain.SpecResult) {
	icon := f.statusIcon(spec.Status)
	fmt.Fprintf(b, "  %s %s\n", icon, f.styled(spec.Name, colorBold))

	if spec.Error != "" {
		fmt.Fprintf(b, "    %s%s%s\n", colorRed, spec.Error, colorReset)
		return
	}

	for _, desc := range spec.Describes {
		f.writeDescribe(b, &desc, spec.Name)
	}

	b.WriteString("\n")
}

func (f *ConsoleFormatter) writeDescribe(b *strings.Builder, desc *domain.DescribeResult, specName string) {
	fmt.Fprintf(b, "    %s\n", desc.Name)

	if desc.Select != "" {
		fmt.Fprintf(b, "      %s[select: %s]%s\n", colorCyan, desc.Select, colorReset)
	}

	for _, ar := range desc.Assertions {
		icon := f.statusIcon(ar.Status)
		fmt.Fprintf(b, "      %s %s\n", icon, ar.Should)

		if ar.Status == domain.StatusFailed && ar.Error != "" {
			fmt.Fprintf(b, "        %s%s%s\n", colorDim+colorRed, ar.Error, colorReset)
		}
	}
}

type failure struct {
	path  string
	error string
}

func (f *ConsoleFormatter) collectFailures(result *domain.SuiteResult) []failure {
	var failures []failure
	for _, spec := range result.Specs {
		if spec.Error != "" {
			failures = append(failures, failure{
				path:  spec.Name,
				error: spec.Error,
			})
			continue
		}
		for _, desc := range spec.Describes {
			for _, ar := range desc.Assertions {
				if ar.Status == domain.StatusFailed || ar.Status == domain.StatusError {
					failures = append(failures, failure{
						path:  fmt.Sprintf("%s > %s > %s", spec.Name, desc.Name, ar.Should),
						error: ar.Error,
					})
				}
			}
		}
	}
	return failures
}

func (f *ConsoleFormatter) summaryLine(s *domain.Summary) string {
	parts := []string{
		fmt.Sprintf("%d assertions", s.TotalAssertions),
		f.styled(fmt.Sprintf("%d passed", s.PassedAssertions), colorGreen),
	}
	if s.FailedAssertions > 0 {
		parts = append(parts, f.styled(fmt.Sprintf("%d failed", s.FailedAssertions), colorRed))
	}
	if s.ErrorSpecs > 0 {
		parts = append(parts, f.styled(fmt.Sprintf("%d errors", s.ErrorSpecs), colorYellow))
	}

	return fmt.Sprintf("\n%s\nFinished in %.2fs\n", strings.Join(parts, ", "), s.Duration.Seconds())
}

func (f *ConsoleFormatter) statusIcon(status domain.Status) string {
	if !f.color {
		switch status {
		case domain.StatusPassed:
			return "+"
		case domain.StatusFailed:
			return "x"
		case domain.StatusError:
			return "!"
		default:
			return "-"
		}
	}

	switch status {
	case domain.StatusPassed:
		return colorGreen + "✓" + colorReset
	case domain.StatusFailed:
		return colorRed + "✗" + colorReset
	case domain.StatusError:
		return colorYellow + "!" + colorReset
	default:
		return colorDim + "-" + colorReset
	}
}

func (f *ConsoleFormatter) styled(text string, codes ...string) string {
	if !f.color {
		return text
	}
	return strings.Join(codes, "") + text + colorReset
}
