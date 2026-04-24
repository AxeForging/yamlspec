package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/AxeForging/yamlspec/domain"
	"github.com/AxeForging/yamlspec/services"
	"github.com/urfave/cli"
)

// ValidateAction handles the validate command
type ValidateAction struct {
	discovery *services.DiscoveryService
	runner    *services.RunnerService
}

// NewValidateAction creates a new ValidateAction
func NewValidateAction() *ValidateAction {
	return &ValidateAction{
		discovery: services.NewDiscoveryService(),
		runner:    services.NewRunnerService(),
	}
}

// Execute runs test specs
func (a *ValidateAction) Execute(c *cli.Context) error {
	config := &domain.Config{
		TestDir:        c.String("test-dir"),
		Tags:           c.StringSlice("tag"),
		Workers:        c.Int("workers"),
		FailFast:       c.Bool("fail-fast"),
		PreRunTimeout:  c.Duration("pre-run-timeout"),
		Verbose:        c.GlobalBool("verbose"),
		Quiet:          c.Bool("quiet"),
		JSONOutput:     c.String("json-output"),
		YAMLOutput:     c.String("yaml-output"),
		MarkdownOutput: c.String("markdown-output"),
		EMDOutput:      c.String("emd-output"),
		JUnitOutput:    c.String("junit-output"),
	}

	if config.FailFast && config.Workers > 1 {
		return fmt.Errorf("--fail-fast cannot be used with --workers > 1 " +
			"(parallel workers can't honor ordered short-circuit; pick one)")
	}

	specs, err := a.discovery.DiscoverSpecs(config.TestDir, config.Tags)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if len(specs) == 0 {
		fmt.Println("No specs found.")
		return nil
	}

	ctx := context.Background()
	result := a.runner.RunAll(ctx, specs, config)

	// Console output (always)
	console := services.NewConsoleFormatter(config.Verbose, config.Quiet)
	output, _ := console.Format(result)
	fmt.Print(string(output))

	// Write additional output formats
	formatters := map[string]services.Formatter{
		config.JSONOutput:     services.NewJSONFormatter(),
		config.YAMLOutput:     services.NewYAMLFormatter(),
		config.MarkdownOutput: services.NewMarkdownFormatter(),
		config.EMDOutput:      services.NewEMDFormatter(),
		config.JUnitOutput:    services.NewJUnitFormatter(),
	}

	for path, formatter := range formatters {
		if path == "" {
			continue
		}
		data, err := formatter.Format(result)
		if err != nil {
			return fmt.Errorf("format error: %w", err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return fmt.Errorf("write '%s': %w", path, err)
		}
	}

	if !result.Summary.Success {
		return cli.NewExitError("", 1)
	}

	return nil
}
