package integration

import (
	"strings"
	"testing"
)

func TestList_ShowsAllSpecs(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "list", "--test-dir", "integration/testdata")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	for _, name := range []string{"basic", "failing", "tagged", "all-operators", "multi-doc", "with-prerun", "with-helm", "with-kustomize"} {
		if !strings.Contains(output, name) {
			t.Errorf("expected '%s' in list output", name)
		}
	}
	if !strings.Contains(output, "8 specs found") {
		t.Errorf("expected '8 specs found', got:\n%s", output)
	}
}

func TestList_Tags(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "list", "--test-dir", "integration/testdata", "--tags")

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "deployment") {
		t.Error("expected 'deployment' tag in output")
	}
	if !strings.Contains(output, "staging") {
		t.Error("expected 'staging' tag in output")
	}
}
