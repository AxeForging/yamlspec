package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSpec(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	specDir := filepath.Join(dir, "mytest")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.yaml"), []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "manifest.yaml"), []byte("kind: Pod\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestParseSpec_UnknownOperatorFails(t *testing.T) {
	// Common typo: "toEquals" instead of "toEqual". Must not silently pass.
	dir := writeSpec(t, `
name: typo
describe:
  - name: broken
    it:
      - should: match replicas
        expect: spec.replicas
        toEquals: 3
`)
	ds := NewDiscoveryService()
	_, err := ds.DiscoverSpecs(dir, nil)
	if err == nil {
		t.Fatal("expected error for unknown operator 'toEquals', got nil")
	}
	if !strings.Contains(err.Error(), "toEquals") {
		t.Errorf("expected error to mention 'toEquals', got: %v", err)
	}
}

func TestParseSpec_AssertionWithoutOperatorFails(t *testing.T) {
	dir := writeSpec(t, `
name: no-op
describe:
  - name: empty
    it:
      - should: do something
        expect: spec.replicas
`)
	ds := NewDiscoveryService()
	_, err := ds.DiscoverSpecs(dir, nil)
	if err == nil {
		t.Fatal("expected error for missing operator, got nil")
	}
	if !strings.Contains(err.Error(), "no assertion operator") {
		t.Errorf("expected error to mention missing operator, got: %v", err)
	}
}

func TestParseSpec_MissingShouldFails(t *testing.T) {
	dir := writeSpec(t, `
name: no-should
describe:
  - name: block
    it:
      - expect: spec.replicas
        toEqual: 3
`)
	ds := NewDiscoveryService()
	_, err := ds.DiscoverSpecs(dir, nil)
	if err == nil || !strings.Contains(err.Error(), "should is required") {
		t.Fatalf("expected 'should is required' error, got: %v", err)
	}
}

func TestParseSpec_MissingExpectFails(t *testing.T) {
	dir := writeSpec(t, `
name: no-expect
describe:
  - name: block
    it:
      - should: check something
        toEqual: 3
`)
	ds := NewDiscoveryService()
	_, err := ds.DiscoverSpecs(dir, nil)
	if err == nil || !strings.Contains(err.Error(), "expect is required") {
		t.Fatalf("expected 'expect is required' error, got: %v", err)
	}
}

func TestParseSpec_EmptyDescribeFails(t *testing.T) {
	dir := writeSpec(t, `
name: empty
describe: []
`)
	ds := NewDiscoveryService()
	_, err := ds.DiscoverSpecs(dir, nil)
	if err == nil || !strings.Contains(err.Error(), "at least one describe") {
		t.Fatalf("expected describe-required error, got: %v", err)
	}
}

func TestParseSpec_ToExistCountsAsOperator(t *testing.T) {
	// Regression: HasValueOperators excludes ToExist/ToBeNull, so validation
	// must use HasAnyOperator to accept toExist-only assertions.
	dir := writeSpec(t, `
name: exist-only
describe:
  - name: block
    it:
      - should: resource limits are set
        expect: spec.containers
        toExist: true
`)
	ds := NewDiscoveryService()
	if _, err := ds.DiscoverSpecs(dir, nil); err != nil {
		t.Errorf("toExist should be a valid sole operator, got: %v", err)
	}
}

func TestParseSpec_ToBeNullCountsAsOperator(t *testing.T) {
	dir := writeSpec(t, `
name: null-only
describe:
  - name: block
    it:
      - should: field is null
        expect: spec.foo
        toBeNull: true
`)
	ds := NewDiscoveryService()
	if _, err := ds.DiscoverSpecs(dir, nil); err != nil {
		t.Errorf("toBeNull should be a valid sole operator, got: %v", err)
	}
}

func TestParseSpec_ValidSpecPasses(t *testing.T) {
	dir := writeSpec(t, `
name: good
tags: ["smoke"]
describe:
  - name: sanity
    it:
      - should: have name
        expect: metadata.name
        toExist: true
`)
	ds := NewDiscoveryService()
	specs, err := ds.DiscoverSpecs(dir, nil)
	if err != nil {
		t.Fatalf("valid spec failed to parse: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if specs[0].Spec.Name != "good" {
		t.Errorf("expected name 'good', got %q", specs[0].Spec.Name)
	}
}
