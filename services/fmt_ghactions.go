package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AxeForging/yamlspec/domain"
)

// GitHubAnnotationsFormatter emits GitHub Actions workflow commands
// (::error file=...,line=...,title=...::message) for failing assertions and
// errored specs. When written to stdout in a workflow, GitHub renders these
// inline in the PR "Files changed" view.
//
// Spec format: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-error-message
type GitHubAnnotationsFormatter struct{}

// NewGitHubAnnotationsFormatter creates a new GitHubAnnotationsFormatter.
func NewGitHubAnnotationsFormatter() *GitHubAnnotationsFormatter {
	return &GitHubAnnotationsFormatter{}
}

func (f *GitHubAnnotationsFormatter) Format(result *domain.SuiteResult) ([]byte, error) {
	cwd, _ := os.Getwd()

	var b strings.Builder
	for _, spec := range result.Specs {
		path := relPath(cwd, spec.SourceFile)

		if spec.Error != "" {
			fmt.Fprintf(&b, "::error file=%s,title=yamlspec: %s::%s\n",
				escapePath(path), escapeProp(spec.Name), escapeMsg(spec.Error))
			continue
		}

		for _, desc := range spec.Describes {
			for _, ar := range desc.Assertions {
				if ar.Status != domain.StatusFailed && ar.Status != domain.StatusError {
					continue
				}
				title := fmt.Sprintf("yamlspec: %s — %s", spec.Name, desc.Name)
				msg := fmt.Sprintf("%s: %s", ar.Should, ar.Error)
				if ar.Error == "" {
					msg = ar.Should
				}

				if ar.SourceLine > 0 && path != "" {
					fmt.Fprintf(&b, "::error file=%s,line=%d,title=%s::%s\n",
						escapePath(path), ar.SourceLine, escapeProp(title), escapeMsg(msg))
				} else if path != "" {
					fmt.Fprintf(&b, "::error file=%s,title=%s::%s\n",
						escapePath(path), escapeProp(title), escapeMsg(msg))
				} else {
					fmt.Fprintf(&b, "::error title=%s::%s\n", escapeProp(title), escapeMsg(msg))
				}
			}
		}
	}
	return []byte(b.String()), nil
}

// relPath returns specPath relative to cwd. Falls back to specPath unchanged
// if Rel fails or specPath is empty.
func relPath(cwd, specPath string) string {
	if specPath == "" {
		return ""
	}
	if cwd == "" {
		return specPath
	}
	rel, err := filepath.Rel(cwd, specPath)
	if err != nil {
		return specPath
	}
	return rel
}

// escapeProp escapes workflow-command property values.
// See https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#example-using-properties
func escapeProp(s string) string {
	r := strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
		":", "%3A",
		",", "%2C",
	)
	return r.Replace(s)
}

// escapeMsg escapes workflow-command message bodies.
func escapeMsg(s string) string {
	r := strings.NewReplacer(
		"%", "%25",
		"\r", "%0D",
		"\n", "%0A",
	)
	return r.Replace(s)
}

// escapePath is a path-safe variant of escapeProp — keeps slashes intact.
func escapePath(s string) string {
	return escapeProp(s)
}
