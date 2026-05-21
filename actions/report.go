package actions

import (
	"fmt"

	"github.com/AxeForging/yamlspec/domain"
	"github.com/urfave/cli"
)

// ReportAction generates the interactive HTML report.
type ReportAction struct {
	validate *ValidateAction
}

// NewReportAction creates a new ReportAction.
func NewReportAction() *ReportAction {
	return &ReportAction{
		validate: NewValidateAction(),
	}
}

// Execute runs specs and writes a self-contained HTML report.
func (a *ReportAction) Execute(c *cli.Context) error {
	output := c.String("output")
	if output == "" {
		return fmt.Errorf("--output cannot be empty")
	}

	config := &domain.Config{
		TestDir:       c.String("test-dir"),
		Tags:          c.StringSlice("tag"),
		Workers:       c.Int("workers"),
		FailFast:      c.Bool("fail-fast"),
		PreRunTimeout: c.Duration("pre-run-timeout"),
		Verbose:       c.GlobalBool("verbose"),
		Quiet:         c.Bool("quiet"),
		HTMLOutput:    output,
	}

	return a.validate.Run(config)
}
