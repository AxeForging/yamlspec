package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AxeForging/yamlspec/domain"
	"github.com/AxeForging/yamlspec/helpers"
	"gopkg.in/yaml.v3"
)

// DiscoveredSpec holds a parsed spec with its filesystem context
type DiscoveredSpec struct {
	Spec      domain.Spec
	Dir       string   // absolute path to the spec directory
	Manifests []string // absolute paths to manifest files
}

// SpecInfo is a lightweight listing entry
type SpecInfo struct {
	Name string
	Dir  string
	Tags []string
}

// DiscoveryService finds and loads spec files
type DiscoveryService struct{}

// NewDiscoveryService creates a new DiscoveryService
func NewDiscoveryService() *DiscoveryService {
	return &DiscoveryService{}
}

// DiscoverSpecs finds all spec.yaml files under testDir, parses them, and filters by tags
func (ds *DiscoveryService) DiscoverSpecs(testDir string, tags []string) ([]DiscoveredSpec, error) {
	absDir, err := filepath.Abs(testDir)
	if err != nil {
		return nil, helpers.WrapError("discovery", "resolve path", err)
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, helpers.WrapError("discovery", fmt.Sprintf("read directory '%s'", testDir), err)
	}

	var specs []DiscoveredSpec

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specDir := filepath.Join(absDir, entry.Name())
		specFile := filepath.Join(specDir, "spec.yaml")

		if _, err := os.Stat(specFile); os.IsNotExist(err) {
			continue
		}

		spec, err := ds.parseSpec(specFile)
		if err != nil {
			return nil, helpers.WrapError("discovery", fmt.Sprintf("parse '%s'", specFile), err)
		}

		// Filter by tags
		if len(tags) > 0 && !spec.HasTag(tags...) {
			continue
		}

		manifests, err := ds.discoverManifests(specDir)
		if err != nil {
			return nil, helpers.WrapError("discovery", fmt.Sprintf("discover manifests in '%s'", specDir), err)
		}

		specs = append(specs, DiscoveredSpec{
			Spec:      *spec,
			Dir:       specDir,
			Manifests: manifests,
		})
	}

	// Sort by directory name for deterministic ordering
	sort.Slice(specs, func(i, j int) bool {
		return filepath.Base(specs[i].Dir) < filepath.Base(specs[j].Dir)
	})

	return specs, nil
}

// ListSpecs returns lightweight info about all discovered specs
func (ds *DiscoveryService) ListSpecs(testDir string) ([]SpecInfo, error) {
	absDir, err := filepath.Abs(testDir)
	if err != nil {
		return nil, helpers.WrapError("discovery", "resolve path", err)
	}

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, helpers.WrapError("discovery", fmt.Sprintf("read directory '%s'", testDir), err)
	}

	var infos []SpecInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		specFile := filepath.Join(absDir, entry.Name(), "spec.yaml")
		if _, err := os.Stat(specFile); os.IsNotExist(err) {
			continue
		}

		spec, err := ds.parseSpec(specFile)
		if err != nil {
			continue
		}

		infos = append(infos, SpecInfo{
			Name: spec.Name,
			Dir:  entry.Name(),
			Tags: spec.Tags,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Dir < infos[j].Dir
	})

	return infos, nil
}

func (ds *DiscoveryService) parseSpec(path string) (*domain.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var spec domain.Spec
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&spec); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("spec file is empty")
		}
		return nil, err
	}

	if spec.Name == "" {
		spec.Name = filepath.Base(filepath.Dir(path))
	}

	spec.SourceFile = path
	if err := annotateSourceLines(&spec, data); err != nil {
		// Line annotations are best-effort — failure here shouldn't block validation.
		helpers.Log.Debug().Err(err).Str("path", path).Msg("could not annotate source lines")
	}

	if err := validateSpec(&spec); err != nil {
		return nil, err
	}

	return &spec, nil
}

// annotateSourceLines re-decodes the YAML as a Node tree and stamps each
// assertion's SourceLine. Used for emitting GitHub Actions annotations that
// link back to the right line in spec.yaml.
func annotateSourceLines(spec *domain.Spec, data []byte) error {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}
	if len(root.Content) == 0 {
		return nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return nil
	}

	describeNode := findMapValue(doc, "describe")
	if describeNode == nil || describeNode.Kind != yaml.SequenceNode {
		return nil
	}

	for di, descNode := range describeNode.Content {
		if di >= len(spec.Describe) || descNode.Kind != yaml.MappingNode {
			continue
		}
		itNode := findMapValue(descNode, "it")
		if itNode == nil || itNode.Kind != yaml.SequenceNode {
			continue
		}
		for ai, assertNode := range itNode.Content {
			if ai >= len(spec.Describe[di].It) || assertNode.Kind != yaml.MappingNode {
				continue
			}
			// Prefer the line of `should:` itself; fall back to the mapping's start.
			line := assertNode.Line
			if shouldNode := findMapKey(assertNode, "should"); shouldNode != nil {
				line = shouldNode.Line
			}
			spec.Describe[di].It[ai].SourceLine = line
		}
	}
	return nil
}

// findMapValue returns the value node for the given key in a MappingNode, or nil.
func findMapValue(m *yaml.Node, key string) *yaml.Node {
	if m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// findMapKey returns the key node for the given key in a MappingNode, or nil.
func findMapKey(m *yaml.Node, key string) *yaml.Node {
	if m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i]
		}
	}
	return nil
}

// validateSpec ensures the spec is structurally sound.
// It catches empty describes, missing fields, and assertions with no operators
// set — all of which would otherwise silently pass.
func validateSpec(spec *domain.Spec) error {
	if len(spec.Describe) == 0 {
		return fmt.Errorf("spec must contain at least one describe block")
	}

	for di, desc := range spec.Describe {
		if desc.Name == "" {
			return fmt.Errorf("describe[%d].name is required", di)
		}
		if len(desc.It) == 0 {
			return fmt.Errorf("describe[%d] (%q).it must contain at least one assertion", di, desc.Name)
		}

		for ai, assertion := range desc.It {
			if assertion.Should == "" {
				return fmt.Errorf("describe[%d] (%q).it[%d].should is required", di, desc.Name, ai)
			}
			if assertion.Expect == "" {
				return fmt.Errorf("describe[%d] (%q).it[%d] (%q).expect is required",
					di, desc.Name, ai, assertion.Should)
			}
			if !assertion.HasAnyOperator() {
				return fmt.Errorf("describe[%d] (%q).it[%d] (%q) has no assertion operator "+
					"(add one of toEqual, toExist, toContain, etc.)",
					di, desc.Name, ai, assertion.Should)
			}
		}
	}

	return nil
}

// discoverManifests finds YAML files to test in a spec directory
func (ds *DiscoveryService) discoverManifests(specDir string) ([]string, error) {
	// First check for manifests/ subdirectory
	manifestsDir := filepath.Join(specDir, "manifests")
	if info, err := os.Stat(manifestsDir); err == nil && info.IsDir() {
		return ds.findYAMLFiles(manifestsDir)
	}

	// Fall back to any YAML files in the spec directory (excluding spec.yaml and values.yaml)
	return ds.findYAMLFiles(specDir, "spec.yaml", "values.yaml")
}

func (ds *DiscoveryService) findYAMLFiles(dir string, excludes ...string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	excludeSet := make(map[string]bool)
	for _, e := range excludes {
		excludeSet[e] = true
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if (ext == ".yaml" || ext == ".yml") && !excludeSet[name] {
			files = append(files, filepath.Join(dir, name))
		}
	}

	sort.Strings(files)
	return files, nil
}
