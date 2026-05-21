package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/AxeForging/yamlspec/domain"
	"github.com/AxeForging/yamlspec/helpers"
)

// RunnerService executes test specs
type RunnerService struct {
	engine *AssertionEngine
}

// NewRunnerService creates a new RunnerService
func NewRunnerService() *RunnerService {
	return &RunnerService{
		engine: NewAssertionEngine(),
	}
}

// RunAll executes all specs with the given configuration
func (rs *RunnerService) RunAll(ctx context.Context, specs []DiscoveredSpec, config *domain.Config) *domain.SuiteResult {
	start := time.Now()
	result := &domain.SuiteResult{}

	if config.Workers <= 1 {
		for _, spec := range specs {
			sr := rs.RunOne(ctx, spec, config)
			result.Specs = append(result.Specs, *sr)

			if config.FailFast && sr.Status == domain.StatusFailed {
				break
			}
		}
	} else {
		result.Specs = rs.runParallel(ctx, specs, config)
	}

	result.Summary = rs.buildSummary(result.Specs, time.Since(start))
	return result
}

// RunOne executes a single spec
func (rs *RunnerService) RunOne(ctx context.Context, spec DiscoveredSpec, config *domain.Config) *domain.SpecResult {
	start := time.Now()
	result := &domain.SpecResult{
		Name:          spec.Spec.Name,
		Tags:          spec.Spec.Tags,
		Status:        domain.StatusPassed,
		SourceFile:    spec.Spec.SourceFile,
		SourceContent: readFileBestEffort(spec.Spec.SourceFile),
	}

	// Execute pre_run commands
	for _, cmd := range spec.Spec.PreRun {
		if err := rs.execPreRun(ctx, cmd, spec.Dir, config.PreRunTimeout); err != nil {
			result.Status = domain.StatusError
			result.Error = fmt.Sprintf("pre_run failed: %v", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Re-discover manifests after pre_run (they may have been generated)
	manifests, manifestResults, err := rs.loadManifests(spec)
	if err != nil {
		result.Status = domain.StatusError
		result.Error = fmt.Sprintf("failed to load manifests: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	result.Manifests = manifestResults

	if len(manifests) == 0 {
		result.Status = domain.StatusError
		result.Error = "no manifests found"
		result.Duration = time.Since(start)
		return result
	}

	// Run assertions
	result.Describes = rs.engine.Evaluate(spec.Spec.Describe, manifests)

	// Determine overall status
	for _, dr := range result.Describes {
		if dr.Status == domain.StatusFailed || dr.Status == domain.StatusError {
			result.Status = domain.StatusFailed
			break
		}
	}

	result.Duration = time.Since(start)
	return result
}

func (rs *RunnerService) runParallel(ctx context.Context, specs []DiscoveredSpec, config *domain.Config) []domain.SpecResult {
	results := make([]domain.SpecResult, len(specs))
	sem := make(chan struct{}, config.Workers)
	var wg sync.WaitGroup

	for i, spec := range specs {
		wg.Add(1)
		go func(idx int, s DiscoveredSpec) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			sr := rs.RunOne(ctx, s, config)
			results[idx] = *sr
		}(i, spec)
	}

	wg.Wait()
	return results
}

func (rs *RunnerService) execPreRun(ctx context.Context, command, dir string, timeout time.Duration) error {
	helpers.Log.Debug().Str("cmd", command).Str("dir", dir).Dur("timeout", timeout).Msg("executing pre_run")

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Put the shell in its own process group so we can SIGKILL the whole
	// tree on cancel — otherwise grandchildren like `sleep` orphan and may
	// hold inherited fds, making cmd.Wait() block past the deadline.
	setProcessGroup(cmd)
	cmd.Cancel = func() error {
		return killProcessGroup(cmd)
	}
	// After the cancel function fires, don't wait more than 2s for cleanup
	// before force-closing our end and returning from Wait.
	cmd.WaitDelay = 2 * time.Second

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command '%s' timed out after %s", command, timeout)
		}
		return fmt.Errorf("command '%s' failed: %w", command, err)
	}
	return nil
}

func (rs *RunnerService) loadManifests(spec DiscoveredSpec) ([]interface{}, []domain.ManifestResult, error) {
	// Re-discover manifests (pre_run may have generated new files)
	ds := NewDiscoveryService()
	manifestFiles, err := ds.discoverManifests(spec.Dir)
	if err != nil {
		return nil, nil, err
	}

	// Also use any originally discovered manifests not in the refreshed list
	fileSet := make(map[string]bool)
	for _, f := range manifestFiles {
		fileSet[f] = true
	}
	for _, f := range spec.Manifests {
		if !fileSet[f] {
			manifestFiles = append(manifestFiles, f)
		}
	}

	var allManifests []interface{}
	var manifestResults []domain.ManifestResult
	for _, file := range manifestFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, nil, fmt.Errorf("read '%s': %w", file, err)
		}

		content := strings.TrimSpace(string(data))
		if content == "" {
			continue
		}

		parsed, err := ParseManifests(content)
		if err != nil {
			return nil, nil, fmt.Errorf("parse '%s': %w", file, err)
		}

		manifestResults = append(manifestResults, domain.ManifestResult{
			Path:      file,
			Content:   string(data),
			Documents: len(parsed),
		})
		allManifests = append(allManifests, parsed...)
	}

	return allManifests, manifestResults, nil
}

func readFileBestEffort(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (rs *RunnerService) buildSummary(specs []domain.SpecResult, duration time.Duration) domain.Summary {
	s := domain.Summary{
		Duration:  duration,
		Timestamp: time.Now(),
	}

	for _, spec := range specs {
		s.TotalSpecs++
		switch spec.Status {
		case domain.StatusPassed:
			s.PassedSpecs++
		case domain.StatusFailed:
			s.FailedSpecs++
		case domain.StatusSkipped:
			s.SkippedSpecs++
		case domain.StatusError:
			s.ErrorSpecs++
		}

		for _, dr := range spec.Describes {
			for _, ar := range dr.Assertions {
				s.TotalAssertions++
				if ar.Status == domain.StatusPassed {
					s.PassedAssertions++
				} else {
					s.FailedAssertions++
				}
			}
		}
	}

	s.Success = s.FailedSpecs == 0 && s.ErrorSpecs == 0
	return s
}
