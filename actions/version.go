package actions

import (
	"fmt"

	"github.com/urfave/cli"
)

// VersionAction handles the version command
type VersionAction struct {
	Version   string
	BuildTime string
	GitCommit string
}

// NewVersionAction creates a new VersionAction
func NewVersionAction(version, buildTime, gitCommit string) *VersionAction {
	return &VersionAction{
		Version:   version,
		BuildTime: buildTime,
		GitCommit: gitCommit,
	}
}

// Execute prints version information
func (a *VersionAction) Execute(c *cli.Context) error {
	fmt.Printf("yamlspec %s\n", a.Version)
	fmt.Printf("  build:  %s\n", a.BuildTime)
	fmt.Printf("  commit: %s\n", a.GitCommit)
	return nil
}
