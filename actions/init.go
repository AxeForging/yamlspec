package actions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"
)

// InitAction handles the init command
type InitAction struct{}

// NewInitAction creates a new InitAction
func NewInitAction() *InitAction {
	return &InitAction{}
}

// Execute scaffolds a new test spec
func (a *InitAction) Execute(c *cli.Context) error {
	name := c.Args().First()
	if name == "" {
		name = "example"
	}

	testDir := c.String("test-dir")
	specDir := filepath.Join(testDir, name)

	if _, err := os.Stat(specDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", specDir)
	}

	manifestsDir := filepath.Join(specDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write spec.yaml
	specContent := fmt.Sprintf(`name: "%s"
tags: ["example"]

# pre_run:
#   - helm template my-release ../../chart -f values.yaml > manifests/rendered.yaml

describe:
  - name: "Resource validation"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "have correct replica count"
        expect: spec.replicas
        toEqual: 3

      - should: "use correct image"
        expect: spec.template.spec.containers[0].image
        toStartWith: "nginx:"

      - should: "have resource limits"
        expect: spec.template.spec.containers[0].resources.limits
        toExist: true
`, name)

	if err := os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(specContent), 0644); err != nil {
		return fmt.Errorf("write spec.yaml: %w", err)
	}

	// Write example manifest
	manifestContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: app
          image: "nginx:1.25.3"
          resources:
            limits:
              cpu: "500m"
              memory: "256Mi"
`

	if err := os.WriteFile(filepath.Join(manifestsDir, "deployment.yaml"), []byte(manifestContent), 0644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	fmt.Printf("Created test spec at %s/\n", specDir)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. Edit %s/spec.yaml to define your assertions\n", specDir)
	fmt.Printf("  2. Add manifests to %s/manifests/\n", specDir)
	fmt.Printf("  3. Run: yamlspec validate --test-dir %s\n", testDir)

	return nil
}
