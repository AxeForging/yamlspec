# Changelog

All notable changes to yamlspec are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `docs/index.md` documentation landing page with cross-links between docs.
- `docs/recipes.md` cookbook of common assertion patterns
  (multi-environment, security baseline, image pinning, label conventions,
  HPA/PDB invariants, NetworkPolicy, etc.).
- `docs/troubleshooting.md` with the common authoring/runtime errors and how
  to read them.
- `examples/README.md` describing each worked example and the exact command
  to run it.
- `CONTRIBUTING.md` covering dev setup, layer conventions, how to add an
  operator/formatter/command, commit conventions, and the release flow.
- README field-path quick reference (dotted, `[index]`, `[*]` wildcard,
  bracket notation, leading-dot JQ form) and cross-links to the docs.
- GitHub Actions reusable workflow now writes the EMD output to
  `$GITHUB_STEP_SUMMARY` so results show on the workflow run page, not only
  as a PR comment.
- `--github-annotations` output mode (auto-enabled when `GITHUB_ACTIONS=true`)
  emits `::error file=...,line=...::` lines so failing assertions appear
  inline in the GitHub "Files changed" diff view.

### Changed
- README install instructions now reflect that no GitHub release exists yet —
  use `go install` from source until v0.1.0 is cut.

## [0.1.0] — Initial release

The first usable yamlspec build, including a correctness-hardening pass.

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
- Reusable GitHub Actions workflow with one-line opt-in via
  `uses: AxeForging/yamlspec/.github/workflows/reusable.yml@main`. Posts
  enriched-markdown PR comments and updates them in place on subsequent
  pushes.
- `init` command to scaffold new specs.
- `list` / `list --tags` for spec/tag discovery (deterministic ordering).
- `ai-help` command emitting a comprehensive reference for AI assistants.
- Built on Go 1.25.8 (covers the recent stdlib CVE).

### Validation
- 118 tests passing across 6 packages.
- Examples for plain manifests, Helm charts, Kustomize overlays,
  multi-resource specs, and a security baseline.

[Unreleased]: https://github.com/AxeForging/yamlspec/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/AxeForging/yamlspec/releases/tag/v0.1.0
