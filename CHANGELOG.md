# Changelog

All notable changes to yamlspec are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — 2026-05-06

The first usable yamlspec build, including a correctness-hardening pass and
the full documentation set.

### Added
- RSpec-like `describe`/`it`/`should` test syntax in `spec.yaml`.
- 22 assertion operators: equality (`toEqual`, `toNotEqual`), numeric
  comparison (`toBeGreaterThan`/`Less`/`OrEqual`), string
  (`toContain`/`Start`/`End`/`Match` and negations), existence (`toExist`,
  `toBeNull`), object (`toHaveKey`), set membership
  (`toBeOneOf`/`toNotBeOneOf`), array (`toContainItem`, `toHaveLength`,
  `toHaveMinLength`, `toHaveMaxLength`).
- Wildcard array iteration in field paths
  (`spec.containers[*].image`) — assertion runs against every element.
- Distinction between "field missing" and "field is null" for
  `toExist`/`toBeNull` semantics.
- Six output formats: console (colored RSpec-style tree), JSON, YAML,
  Markdown, enriched Markdown (collapsible `<details>` for PR comments),
  JUnit XML.
- `--github-annotations` output mode (auto-enabled when `GITHUB_ACTIONS=true`)
  emits `::error file=...,line=...::` lines so failing assertions appear
  inline in the GitHub "Files changed" diff view.
- Parallel execution via `--workers N`.
- `--fail-fast` for fast feedback on sequential runs.
- Configurable `--pre-run-timeout` (default 60s) for `pre_run` shell
  commands.
- Tag filtering via repeatable `--tag` flag.
- Strict YAML decoding for `spec.yaml` — typos in field names are rejected
  rather than silently ignored.
- Comprehensive spec validation: empty `describe`, missing `should`/`expect`,
  assertions with no operator are caught at parse time.
- `pre_run` execution kills the entire process group on timeout — orphaned
  `sleep`-style children no longer keep the command alive past its deadline.
- Cross-platform process management (Unix `setpgid` + `kill -PGID`; Windows
  fallback).
- Reusable GitHub Actions workflow with one-line opt-in. Posts
  enriched-markdown PR comments (updated in place on subsequent pushes) and
  writes the same EMD to `$GITHUB_STEP_SUMMARY` so results show on the
  workflow run page.
- `init` command to scaffold new specs.
- `list` / `list --tags` for spec/tag discovery (deterministic ordering).
- `ai-help` command emitting a comprehensive reference for AI assistants.
- Built on Go 1.25.8 (covers the recent stdlib CVE).
- Documentation: `docs/index.md`, `docs/spec-format.md`, `docs/recipes.md`,
  `docs/troubleshooting.md`, `docs/reusable-workflow.md`,
  `examples/README.md`, `CONTRIBUTING.md`, `CHANGELOG.md`.
- README field-path quick reference, badges, sample output, and demo GIF.

### Validation
- 120 tests passing across 6 packages.
- Examples for plain manifests, Helm charts, Kustomize overlays,
  multi-resource specs, and a security baseline.

[Unreleased]: https://github.com/AxeForging/yamlspec/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/AxeForging/yamlspec/releases/tag/v0.1.0
