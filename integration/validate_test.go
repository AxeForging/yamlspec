package integration

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Fatalf("git rev-parse: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func buildBinary(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	bin := filepath.Join(t.TempDir(), "yamlspec")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func runAt(t *testing.T, bin, dir string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run failed: %v", err)
		}
	}
	return string(out), exitCode
}

func run(t *testing.T, bin string, args ...string) (string, int) {
	t.Helper()
	return runAt(t, bin, repoRoot(t), args...)
}

// --- Basic discovery and execution ---

func TestE2E_DiscoverAndRunAllSpecs(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata")

	// Must exit 1 because "failing" fixture has intentional failures
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d\noutput:\n%s", exitCode, output)
	}

	// Every fixture should appear in output
	for _, name := range []string{
		"Basic deployment test",
		"Failing deployment test",
		"Tagged test",
		"Pre-run test",
		"All operators test",
		"Multi-document YAML test",
		"Helm chart rendering",
		"Kustomize overlay rendering",
	} {
		if !strings.Contains(output, name) {
			t.Errorf("expected '%s' in output", name)
		}
	}

	// Summary line should exist
	if !strings.Contains(output, "assertions") && !strings.Contains(output, "passed") {
		t.Error("expected summary line with assertion counts")
	}
}

// --- Tag filtering ---

func TestE2E_TagFilter_SingleTag(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "staging")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\n%s", exitCode, output)
	}

	if !strings.Contains(output, "Tagged test") {
		t.Error("expected Tagged test")
	}
	if strings.Contains(output, "Basic deployment test") {
		t.Error("basic spec should be filtered out")
	}
	if strings.Contains(output, "Failing deployment test") {
		t.Error("failing spec should be filtered out")
	}
}

func TestE2E_TagFilter_OnlyPassingSpecs(t *testing.T) {
	bin := buildBinary(t)

	_, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "service")

	if exitCode != 0 {
		t.Errorf("expected exit 0 for passing-only filter, got %d", exitCode)
	}
}

func TestE2E_TagFilter_OnlyFailingSpecs(t *testing.T) {
	bin := buildBinary(t)

	_, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "failing")

	if exitCode != 1 {
		t.Errorf("expected exit 1 for failing-only filter, got %d", exitCode)
	}
}

func TestE2E_TagFilter_NoMatch(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "nonexistent-tag")

	if exitCode != 0 {
		t.Errorf("expected exit 0 for no matches, got %d", exitCode)
	}
	if !strings.Contains(output, "No specs found") {
		t.Error("expected 'No specs found' message")
	}
}

// --- Pre-run execution ---

func TestE2E_PreRun_GeneratesManifests(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "prerun")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\n%s", exitCode, output)
	}

	if !strings.Contains(output, "Pre-run test") {
		t.Error("expected Pre-run test in output")
	}
	if !strings.Contains(output, "have 2 replicas") {
		t.Error("expected assertion about replicas")
	}

	// Clean up generated manifests
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-prerun/manifests"))
}

func TestE2E_PreRun_FailingCommand(t *testing.T) {
	bin := buildBinary(t)
	tmpDir := t.TempDir()
	specDir := filepath.Join(tmpDir, "bad-prerun")
	manifestsDir := filepath.Join(specDir, "manifests")
	os.MkdirAll(manifestsDir, 0755)

	spec := `name: "Bad pre-run"
tags: ["bad-prerun"]
pre_run:
  - exit 1
describe:
  - name: "Should not reach"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "never run"
        expect: metadata.name
        toExist: true
`
	os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(spec), 0644)

	output, exitCode := run(t, bin, "validate", "--test-dir", tmpDir)

	if exitCode == 0 {
		t.Error("expected non-zero exit for failed pre_run")
	}
	if !strings.Contains(output, "pre_run failed") {
		t.Errorf("expected 'pre_run failed' in output, got:\n%s", output)
	}
}

// --- Helm pre_run ---

func TestE2E_Helm_TemplateRendering(t *testing.T) {
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm not installed, skipping")
	}

	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "helm")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\n%s", exitCode, output)
	}

	if !strings.Contains(output, "Helm chart rendering") {
		t.Error("expected spec name in output")
	}

	// Verify Helm-specific assertions ran
	for _, check := range []string{
		"have release name",
		"have 3 replicas from override values",
		"be in production namespace",
		"use nginx image",
		"have resource limits",
		"be ClusterIP",
		"target port 8080",
	} {
		if !strings.Contains(output, check) {
			t.Errorf("expected assertion '%s' in output", check)
		}
	}

	// Clean up generated manifests
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-helm/manifests"))
}

func TestE2E_Helm_JSONOutputHasCorrectValues(t *testing.T) {
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm not installed, skipping")
	}

	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "helm-results.json")

	run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "helm", "--json-output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var result struct {
		Specs []struct {
			Name      string `json:"name"`
			Status    string `json:"status"`
			Describes []struct {
				Name       string `json:"name"`
				Assertions []struct {
					Should string `json:"should"`
					Status string `json:"status"`
				} `json:"assertions"`
			} `json:"describes"`
		} `json:"specs"`
		Summary struct {
			TotalAssertions  int  `json:"total_assertions"`
			PassedAssertions int  `json:"passed_assertions"`
			Success          bool `json:"success"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(result.Specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(result.Specs))
	}
	if result.Specs[0].Status != "passed" {
		t.Errorf("expected passed, got %s", result.Specs[0].Status)
	}
	if !result.Summary.Success {
		t.Error("expected summary.success=true")
	}
	if result.Summary.TotalAssertions != 7 {
		t.Errorf("expected 7 assertions, got %d", result.Summary.TotalAssertions)
	}

	// Clean up
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-helm/manifests"))
}

// --- Kustomize pre_run ---

func TestE2E_Kustomize_BuildRendering(t *testing.T) {
	if _, err := exec.LookPath("kustomize"); err != nil {
		t.Skip("kustomize not installed, skipping")
	}

	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "kustomize")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\n%s", exitCode, output)
	}

	if !strings.Contains(output, "Kustomize overlay rendering") {
		t.Error("expected spec name in output")
	}

	// Verify kustomize-specific assertions
	for _, check := range []string{
		"have 2 replicas from patch",
		"be in staging namespace",
		"have env label from commonLabels",
		"use nginx image",
		"have env label",
	} {
		if !strings.Contains(output, check) {
			t.Errorf("expected assertion '%s' in output", check)
		}
	}

	// Clean up
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-kustomize/manifests"))
}

func TestE2E_Kustomize_JSONOutputHasCorrectValues(t *testing.T) {
	if _, err := exec.LookPath("kustomize"); err != nil {
		t.Skip("kustomize not installed, skipping")
	}

	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "kustomize-results.json")

	run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "kustomize", "--json-output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var result struct {
		Specs []struct {
			Name      string `json:"name"`
			Status    string `json:"status"`
			Describes []struct {
				Name string `json:"name"`
			} `json:"describes"`
		} `json:"specs"`
		Summary struct {
			TotalAssertions  int  `json:"total_assertions"`
			PassedAssertions int  `json:"passed_assertions"`
			Success          bool `json:"success"`
		} `json:"summary"`
	}
	json.Unmarshal(data, &result)

	if len(result.Specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(result.Specs))
	}
	if result.Specs[0].Status != "passed" {
		t.Errorf("expected passed, got %s", result.Specs[0].Status)
	}
	if result.Summary.TotalAssertions != 6 {
		t.Errorf("expected 6 assertions, got %d", result.Summary.TotalAssertions)
	}

	// Verify both describe blocks present
	descNames := make(map[string]bool)
	for _, d := range result.Specs[0].Describes {
		descNames[d.Name] = true
	}
	if !descNames["Deployment from Kustomize"] {
		t.Error("expected 'Deployment from Kustomize' describe block")
	}
	if !descNames["Service from Kustomize"] {
		t.Error("expected 'Service from Kustomize' describe block")
	}

	// Clean up
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-kustomize/manifests"))
}

// --- All three engines in one run ---

func TestE2E_AllEngines_RunTogether(t *testing.T) {
	helmAvailable := true
	kustomizeAvailable := true
	if _, err := exec.LookPath("helm"); err != nil {
		helmAvailable = false
	}
	if _, err := exec.LookPath("kustomize"); err != nil {
		kustomizeAvailable = false
	}

	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "all-engines.json")

	// Run all specs (will include plain manifests, helm, kustomize, prerun, etc.)
	output, _ := run(t, bin, "validate", "--test-dir", "integration/testdata", "--json-output", outFile)

	// Plain manifests always work
	if !strings.Contains(output, "Basic deployment test") {
		t.Error("expected plain manifests spec")
	}

	// Verify helm ran if available
	if helmAvailable {
		if !strings.Contains(output, "Helm chart rendering") {
			t.Error("expected helm spec to run")
		}
	}

	// Verify kustomize ran if available
	if kustomizeAvailable {
		if !strings.Contains(output, "Kustomize overlay rendering") {
			t.Error("expected kustomize spec to run")
		}
	}

	// Parse JSON and verify multi-engine results
	data, _ := os.ReadFile(outFile)
	var result struct {
		Summary struct {
			TotalSpecs int  `json:"total_specs"`
			Success    bool `json:"success"`
		} `json:"summary"`
	}
	json.Unmarshal(data, &result)

	// Should have at least 6 specs (basic, failing, tagged, prerun, operators, multi-doc)
	// Plus helm and kustomize if available
	minSpecs := 6
	if helmAvailable {
		minSpecs++
	}
	if kustomizeAvailable {
		minSpecs++
	}
	if result.Summary.TotalSpecs < minSpecs {
		t.Errorf("expected at least %d specs, got %d", minSpecs, result.Summary.TotalSpecs)
	}

	// Clean up generated manifests
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-helm/manifests"))
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-kustomize/manifests"))
	os.RemoveAll(filepath.Join(repoRoot(t), "integration/testdata/with-prerun/manifests"))
}

// --- All operators end-to-end ---

func TestE2E_AllOperators_Pass(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "operators")

	if exitCode != 0 {
		t.Errorf("expected all operators to pass, got exit %d\n%s", exitCode, output)
	}

	// Verify each describe block appeared
	for _, block := range []string{
		"Equality operators",
		"Numeric comparison operators",
		"String operators",
		"Existence operators",
		"Object operators",
		"Set membership operators",
		"Array operators",
		"Bracket notation for special keys",
	} {
		if !strings.Contains(output, block) {
			t.Errorf("expected describe block '%s' in output", block)
		}
	}
}

// --- Multi-document YAML ---

func TestE2E_MultiDocument_SelectsDifferentKinds(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "multi-doc")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\n%s", exitCode, output)
	}

	for _, expected := range []string{"Deployment", "Service", "ConfigMap"} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected '%s' describe block in output", expected)
		}
	}
}

// --- Fail-fast mode ---

func TestE2E_FailFast_StopsAfterFirstFailure(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "validate", "--test-dir", "integration/testdata", "--fail-fast")

	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}

	// Count how many spec headers appear — fail-fast should stop early
	specCount := 0
	for _, name := range []string{
		"Basic deployment test",
		"Failing deployment test",
		"Tagged test",
		"Pre-run test",
		"All operators test",
		"Multi-document YAML test",
	} {
		if strings.Contains(output, name) {
			specCount++
		}
	}

	// With fail-fast, we should see fewer specs than the total 6
	// (basic passes, failing stops execution)
	if specCount >= 6 {
		t.Errorf("fail-fast should have stopped early, but saw all %d specs", specCount)
	}

	_ = output // suppress unused
}

// --- Quiet mode ---

func TestE2E_QuietMode_OnlySummary(t *testing.T) {
	bin := buildBinary(t)

	output, _ := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "service", "--quiet")

	// Quiet mode should NOT show individual assertions
	if strings.Contains(output, "have 3 replicas") {
		t.Error("quiet mode should not show individual assertions")
	}

	// But should still show summary
	if !strings.Contains(output, "assertions") {
		t.Error("quiet mode should still show summary line")
	}
}

// --- No specs ---

func TestE2E_EmptyTestDir(t *testing.T) {
	bin := buildBinary(t)
	emptyDir := t.TempDir()

	output, exitCode := run(t, bin, "validate", "--test-dir", emptyDir)

	if exitCode != 0 {
		t.Errorf("expected exit 0 for empty dir, got %d", exitCode)
	}
	if !strings.Contains(output, "No specs found") {
		t.Error("expected 'No specs found'")
	}
}

func TestE2E_NonexistentTestDir(t *testing.T) {
	bin := buildBinary(t)

	_, exitCode := run(t, bin, "validate", "--test-dir", "/tmp/yamlspec-nonexistent-dir-12345")

	if exitCode == 0 {
		t.Error("expected non-zero exit for nonexistent dir")
	}
}

// --- JSON output deep validation ---

func TestE2E_JSONOutput_Structure(t *testing.T) {
	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "results.json")

	run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "operators", "--json-output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var result struct {
		Specs []struct {
			Name      string `json:"name"`
			Tags      []string `json:"tags"`
			Status    string `json:"status"`
			Describes []struct {
				Name       string `json:"name"`
				Select     string `json:"select"`
				Status     string `json:"status"`
				Assertions []struct {
					Should string `json:"should"`
					Status string `json:"status"`
					Error  string `json:"error"`
				} `json:"assertions"`
			} `json:"describes"`
		} `json:"specs"`
		Summary struct {
			TotalSpecs       int  `json:"total_specs"`
			PassedSpecs      int  `json:"passed_specs"`
			TotalAssertions  int  `json:"total_assertions"`
			PassedAssertions int  `json:"passed_assertions"`
			Success          bool `json:"success"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, string(data))
	}

	if len(result.Specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(result.Specs))
	}

	spec := result.Specs[0]
	if spec.Name != "All operators test" {
		t.Errorf("expected spec name 'All operators test', got '%s'", spec.Name)
	}
	if spec.Status != "passed" {
		t.Errorf("expected status 'passed', got '%s'", spec.Status)
	}
	if len(spec.Tags) == 0 {
		t.Error("expected tags in JSON output")
	}
	if len(spec.Describes) < 8 {
		t.Errorf("expected at least 8 describe blocks, got %d", len(spec.Describes))
	}

	// Verify all assertions passed
	for _, desc := range spec.Describes {
		for _, ar := range desc.Assertions {
			if ar.Status != "passed" {
				t.Errorf("assertion '%s' in '%s' should be passed, got '%s': %s",
					ar.Should, desc.Name, ar.Status, ar.Error)
			}
		}
	}

	// Verify summary
	if !result.Summary.Success {
		t.Error("expected summary.success to be true")
	}
	if result.Summary.TotalSpecs != 1 {
		t.Errorf("expected 1 total spec, got %d", result.Summary.TotalSpecs)
	}
	if result.Summary.PassedAssertions != result.Summary.TotalAssertions {
		t.Errorf("expected all assertions passed: %d/%d",
			result.Summary.PassedAssertions, result.Summary.TotalAssertions)
	}
}

func TestE2E_JSONOutput_FailuresIncludeErrors(t *testing.T) {
	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "results.json")

	run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "failing", "--json-output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}

	var result struct {
		Specs []struct {
			Status    string `json:"status"`
			Describes []struct {
				Assertions []struct {
					Should string `json:"should"`
					Status string `json:"status"`
					Error  string `json:"error"`
				} `json:"assertions"`
			} `json:"describes"`
		} `json:"specs"`
		Summary struct {
			FailedAssertions int  `json:"failed_assertions"`
			Success          bool `json:"success"`
		} `json:"summary"`
	}
	json.Unmarshal(data, &result)

	if result.Summary.Success {
		t.Error("expected summary.success to be false")
	}
	if result.Summary.FailedAssertions == 0 {
		t.Error("expected non-zero failed assertions")
	}

	// At least one assertion should have an error message
	foundError := false
	for _, spec := range result.Specs {
		for _, desc := range spec.Describes {
			for _, ar := range desc.Assertions {
				if ar.Error != "" {
					foundError = true
				}
			}
		}
	}
	if !foundError {
		t.Error("expected at least one assertion with error message")
	}
}

// --- JUnit XML deep validation ---

func TestE2E_JUnitOutput_Structure(t *testing.T) {
	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "results.xml")

	run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "operators", "--junit-output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read junit: %v", err)
	}

	var suites struct {
		XMLName xml.Name `xml:"testsuites"`
		Suites  []struct {
			Name     string `xml:"name,attr"`
			Tests    int    `xml:"tests,attr"`
			Failures int    `xml:"failures,attr"`
			Cases    []struct {
				Name      string `xml:"name,attr"`
				ClassName string `xml:"classname,attr"`
			} `xml:"testcase"`
		} `xml:"testsuite"`
	}
	if err := xml.Unmarshal(data, &suites); err != nil {
		t.Fatalf("invalid JUnit XML: %v", err)
	}

	if len(suites.Suites) != 1 {
		t.Fatalf("expected 1 test suite, got %d", len(suites.Suites))
	}

	suite := suites.Suites[0]
	if suite.Name != "All operators test" {
		t.Errorf("expected suite name 'All operators test', got '%s'", suite.Name)
	}
	if suite.Tests == 0 {
		t.Error("expected non-zero test count")
	}
	if suite.Failures != 0 {
		t.Errorf("expected 0 failures, got %d", suite.Failures)
	}
	if len(suite.Cases) == 0 {
		t.Error("expected test cases in JUnit output")
	}

	// Each case should have a classname like "SpecName.DescribeName"
	for _, tc := range suite.Cases {
		if !strings.Contains(tc.ClassName, ".") {
			t.Errorf("expected classname with dot separator, got '%s'", tc.ClassName)
		}
	}
}

// --- Markdown + EMD output validation ---

func TestE2E_MarkdownOutput_Content(t *testing.T) {
	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "results.md")

	run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "multi-doc", "--markdown-output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "# Test Results") {
		t.Error("expected markdown header")
	}
	if !strings.Contains(content, "Multi-document YAML test") {
		t.Error("expected spec name in markdown")
	}
	if !strings.Contains(content, "Deployment") {
		t.Error("expected describe block in markdown")
	}
	if !strings.Contains(content, "PASS") {
		t.Error("expected PASS status")
	}
	if !strings.Contains(content, "## Summary") {
		t.Error("expected summary section")
	}
}

func TestE2E_EMDOutput_CollapsibleSections(t *testing.T) {
	bin := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "results.emd.md")

	run(t, bin, "validate", "--test-dir", "integration/testdata", "--json-output", filepath.Join(t.TempDir(), "r.json"), "--emd-output", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read emd: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "<details") {
		t.Error("expected <details> tag in EMD output")
	}
	if !strings.Contains(content, "</details>") {
		t.Error("expected </details> tag in EMD output")
	}
	if !strings.Contains(content, "<summary>") {
		t.Error("expected <summary> tag in EMD output")
	}

	// Failed specs should have "open" attribute
	if strings.Contains(content, "Failing deployment test") {
		if !strings.Contains(content, "<details open>") {
			t.Error("expected failed spec to have <details open>")
		}
	}
}

// --- Multiple simultaneous output formats ---

func TestE2E_AllFormatsSimultaneously(t *testing.T) {
	bin := buildBinary(t)
	tmp := t.TempDir()

	jsonFile := filepath.Join(tmp, "results.json")
	yamlFile := filepath.Join(tmp, "results.yaml")
	mdFile := filepath.Join(tmp, "results.md")
	emdFile := filepath.Join(tmp, "results.emd.md")
	junitFile := filepath.Join(tmp, "results.xml")

	output, exitCode := run(t, bin, "validate",
		"--test-dir", "integration/testdata",
		"--tag", "service",
		"--json-output", jsonFile,
		"--yaml-output", yamlFile,
		"--markdown-output", mdFile,
		"--emd-output", emdFile,
		"--junit-output", junitFile,
	)

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\n%s", exitCode, output)
	}

	// All files should exist and be non-empty
	for _, f := range []string{jsonFile, yamlFile, mdFile, emdFile, junitFile} {
		info, err := os.Stat(f)
		if err != nil {
			t.Errorf("expected %s to exist: %v", filepath.Base(f), err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("expected %s to be non-empty", filepath.Base(f))
		}
	}

	// Verify JSON is valid
	jsonData, _ := os.ReadFile(jsonFile)
	var jsonResult map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonResult); err != nil {
		t.Errorf("JSON output is invalid: %v", err)
	}

	// Verify XML is valid
	xmlData, _ := os.ReadFile(junitFile)
	var xmlResult struct {
		XMLName xml.Name `xml:"testsuites"`
	}
	if err := xml.Unmarshal(xmlData, &xmlResult); err != nil {
		t.Errorf("JUnit XML output is invalid: %v", err)
	}

	// Verify YAML contains expected content
	yamlData, _ := os.ReadFile(yamlFile)
	if !strings.Contains(string(yamlData), "specs:") {
		t.Error("YAML output should contain 'specs:'")
	}
}

// --- Console output structure ---

func TestE2E_ConsoleOutput_FailureDetails(t *testing.T) {
	bin := buildBinary(t)

	output, _ := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "failing")

	// Should show failures section
	if !strings.Contains(output, "Failures:") {
		t.Error("expected 'Failures:' section in output")
	}

	// Should show breadcrumb path in failures
	if !strings.Contains(output, "Failing deployment test") {
		t.Error("expected spec name in failure details")
	}

	// Should include the actual error message
	if !strings.Contains(output, "to equal 3, got 1") {
		t.Error("expected detailed error message in failure output")
	}
}

func TestE2E_ConsoleOutput_PassingShowsNoFailures(t *testing.T) {
	bin := buildBinary(t)

	output, _ := run(t, bin, "validate", "--test-dir", "integration/testdata", "--tag", "service")

	if strings.Contains(output, "Failures:") {
		t.Error("passing run should not show Failures section")
	}
}

// --- Version command ---

func TestE2E_Version(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "version")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(output, "yamlspec") {
		t.Error("expected 'yamlspec' in version output")
	}
}

// --- Help ---

func TestE2E_Help(t *testing.T) {
	bin := buildBinary(t)

	output, exitCode := run(t, bin, "--help")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
	if !strings.Contains(output, "validate") {
		t.Error("expected 'validate' command in help")
	}
	if !strings.Contains(output, "list") {
		t.Error("expected 'list' command in help")
	}
	if !strings.Contains(output, "init") {
		t.Error("expected 'init' command in help")
	}
}
