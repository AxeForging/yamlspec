package main

import (
	"fmt"
	"os"

	"github.com/AxeForging/yamlspec/actions"
	"github.com/AxeForging/yamlspec/helpers"
	"github.com/urfave/cli"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	helpers.SetupLogger("info")

	validateAction := actions.NewValidateAction()
	listAction := actions.NewListAction()
	initAction := actions.NewInitAction()
	aiHelpAction := actions.NewAIHelpAction()
	versionAction := actions.NewVersionAction(Version, BuildTime, GitCommit)
	reportAction := actions.NewReportAction()
	serveAction := actions.NewServeAction()

	app := cli.NewApp()
	app.Name = "yamlspec"
	app.Usage = "YAML test framework with RSpec-like assertions"
	app.Version = Version

	app.Flags = []cli.Flag{
		verboseFlag,
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("verbose") {
			helpers.SetupLogger("debug")
		}
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:    "validate",
			Aliases: []string{"test", "run"},
			Usage:   "Run test specs against YAML manifests",
			Flags: []cli.Flag{
				testDirFlag,
				tagFlag,
				workersFlag,
				failFastFlag,
				preRunTimeoutFlag,
				quietFlag,
				jsonOutputFlag,
				yamlOutputFlag,
				markdownOutputFlag,
				emdOutputFlag,
				htmlOutputFlag,
				junitOutputFlag,
				githubAnnotationsFlag,
			},
			Action: validateAction.Execute,
		},
		{
			Name:  "report",
			Usage: "Generate an interactive single-file HTML report",
			Flags: []cli.Flag{
				testDirFlag,
				tagFlag,
				workersFlag,
				failFastFlag,
				preRunTimeoutFlag,
				quietFlag,
				reportOutputFlag,
			},
			Action: reportAction.Execute,
		},
		{
			Name:  "serve",
			Usage: "Serve a generated HTML report locally",
			Flags: []cli.Flag{
				serveFileFlag,
				serveAddrFlag,
			},
			Action: serveAction.Execute,
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List discovered specs and tags",
			Flags: []cli.Flag{
				testDirFlag,
				cli.BoolFlag{
					Name:  "tags",
					Usage: "List all unique tags",
				},
			},
			Action: listAction.Execute,
		},
		{
			Name:  "init",
			Usage: "Scaffold a new test spec",
			Flags: []cli.Flag{
				testDirFlag,
			},
			Action: initAction.Execute,
		},
		{
			Name:   "ai-help",
			Usage:  "Print comprehensive reference guide for AI assistants",
			Action: aiHelpAction.Execute,
		},
		{
			Name:   "version",
			Usage:  "Show version information",
			Action: versionAction.Execute,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
