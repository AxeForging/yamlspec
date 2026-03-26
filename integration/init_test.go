package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_CreatesScaffold(t *testing.T) {
	bin := buildBinary(t)
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "tests")

	output, exitCode := run(t, bin, "init", "my-test", "--test-dir", testDir)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d\n%s", exitCode, output)
	}

	// Check spec.yaml exists
	specFile := filepath.Join(testDir, "my-test", "spec.yaml")
	if _, err := os.Stat(specFile); os.IsNotExist(err) {
		t.Error("expected spec.yaml to be created")
	}

	// Check manifests dir exists
	manifestsDir := filepath.Join(testDir, "my-test", "manifests")
	if _, err := os.Stat(manifestsDir); os.IsNotExist(err) {
		t.Error("expected manifests/ directory to be created")
	}

	// Check manifest file exists
	manifestFile := filepath.Join(manifestsDir, "deployment.yaml")
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		t.Error("expected deployment.yaml to be created")
	}

	if !strings.Contains(output, "Created test spec") {
		t.Error("expected success message")
	}
}

func TestInit_ExistingDirFails(t *testing.T) {
	bin := buildBinary(t)
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "tests")

	// Create it first
	run(t, bin, "init", "existing", "--test-dir", testDir)

	// Try again — should fail
	_, exitCode := run(t, bin, "init", "existing", "--test-dir", testDir)

	if exitCode == 0 {
		t.Error("expected non-zero exit code for existing directory")
	}
}

func TestInit_RunsSuccessfully(t *testing.T) {
	bin := buildBinary(t)
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "tests")

	// Init
	run(t, bin, "init", "runnable", "--test-dir", testDir)

	// Validate the scaffolded spec
	output, exitCode := run(t, bin, "validate", "--test-dir", testDir)

	if exitCode != 0 {
		t.Errorf("expected scaffolded spec to pass, got exit code %d\n%s", exitCode, output)
	}
}
