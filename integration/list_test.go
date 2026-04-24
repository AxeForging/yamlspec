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

func TestList_TagsSorted(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "list", "--test-dir", "integration/testdata", "--tags")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	// Extract tag column (first token of each data line)
	var tags []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "TAG") || strings.HasPrefix(line, "-") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		tags = append(tags, fields[0])
	}

	if len(tags) < 2 {
		t.Fatalf("expected multiple tags, got: %v", tags)
	}
	for i := 1; i < len(tags); i++ {
		if tags[i] < tags[i-1] {
			t.Errorf("tags not sorted: %v (out of order: %q before %q)", tags, tags[i-1], tags[i])
			break
		}
	}
}
