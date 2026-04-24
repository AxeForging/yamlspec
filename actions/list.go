package actions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AxeForging/yamlspec/services"
	"github.com/urfave/cli"
)

// ListAction handles the list command
type ListAction struct {
	discovery *services.DiscoveryService
}

// NewListAction creates a new ListAction
func NewListAction() *ListAction {
	return &ListAction{
		discovery: services.NewDiscoveryService(),
	}
}

// Execute lists discovered specs
func (a *ListAction) Execute(c *cli.Context) error {
	testDir := c.String("test-dir")

	if c.Bool("tags") {
		return a.listTags(testDir)
	}

	infos, err := a.discovery.ListSpecs(testDir)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	if len(infos) == 0 {
		fmt.Println("No specs found.")
		return nil
	}

	fmt.Printf("%-30s %-20s %s\n", "DIRECTORY", "NAME", "TAGS")
	fmt.Printf("%-30s %-20s %s\n", strings.Repeat("-", 28), strings.Repeat("-", 18), strings.Repeat("-", 20))

	for _, info := range infos {
		tags := ""
		if len(info.Tags) > 0 {
			tags = strings.Join(info.Tags, ", ")
		}
		fmt.Printf("%-30s %-20s %s\n", info.Dir, info.Name, tags)
	}

	fmt.Printf("\n%d specs found.\n", len(infos))
	return nil
}

func (a *ListAction) listTags(testDir string) error {
	infos, err := a.discovery.ListSpecs(testDir)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}

	tagSet := make(map[string]int)
	for _, info := range infos {
		for _, tag := range info.Tags {
			tagSet[tag]++
		}
	}

	if len(tagSet) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	fmt.Printf("%-30s %s\n", "TAG", "SPECS")
	fmt.Printf("%-30s %s\n", strings.Repeat("-", 28), strings.Repeat("-", 8))
	for _, tag := range tags {
		fmt.Printf("%-30s %d\n", tag, tagSet[tag])
	}

	return nil
}
