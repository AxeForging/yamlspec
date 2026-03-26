package services

import (
	"fmt"
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
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}

	// Default name to directory name
	if spec.Name == "" {
		spec.Name = filepath.Base(filepath.Dir(path))
	}

	return &spec, nil
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
