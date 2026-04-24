package main

import (
	"time"

	"github.com/urfave/cli"
)

var (
	verboseFlag = cli.BoolFlag{
		Name:  "verbose",
		Usage: "Enable verbose output",
	}
	quietFlag = cli.BoolFlag{
		Name:  "quiet, q",
		Usage: "Only show summary",
	}
	testDirFlag = cli.StringFlag{
		Name:  "test-dir, d",
		Usage: "Directory containing test specs",
		Value: "tests",
	}
	tagFlag = cli.StringSliceFlag{
		Name:  "tag, t",
		Usage: "Filter specs by tag (repeatable)",
	}
	workersFlag = cli.IntFlag{
		Name:  "workers, w",
		Usage: "Number of parallel workers",
		Value: 1,
	}
	failFastFlag = cli.BoolFlag{
		Name:  "fail-fast",
		Usage: "Stop on first failure (not compatible with --workers > 1)",
	}
	preRunTimeoutFlag = cli.DurationFlag{
		Name:  "pre-run-timeout",
		Usage: "Max duration for each pre_run command (e.g. 30s, 2m)",
		Value: 60 * time.Second,
	}
	jsonOutputFlag = cli.StringFlag{
		Name:  "json-output",
		Usage: "Write JSON results to file",
	}
	yamlOutputFlag = cli.StringFlag{
		Name:  "yaml-output",
		Usage: "Write YAML results to file",
	}
	markdownOutputFlag = cli.StringFlag{
		Name:  "markdown-output, md",
		Usage: "Write Markdown results to file",
	}
	emdOutputFlag = cli.StringFlag{
		Name:  "emd-output",
		Usage: "Write enriched Markdown results to file",
	}
	junitOutputFlag = cli.StringFlag{
		Name:  "junit-output",
		Usage: "Write JUnit XML results to file",
	}
)
