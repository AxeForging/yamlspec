# yamlspec

[![CI](https://github.com/AxeForging/yamlspec/actions/workflows/pr.yml/badge.svg)](https://github.com/AxeForging/yamlspec/actions/workflows/pr.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/AxeForging/yamlspec.svg)](https://pkg.go.dev/github.com/AxeForging/yamlspec)
[![Go Report Card](https://goreportcard.com/badge/github.com/AxeForging/yamlspec)](https://goreportcard.com/report/github.com/AxeForging/yamlspec)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

YAML test framework with RSpec-like assertions. Validate any YAML manifests — Kubernetes, Helm, Kustomize, or plain files — with a clean, readable syntax.

![yamlspec demo](docs/demo.gif)

**Documentation:** [docs index](docs/index.md) · [spec.yaml format](docs/spec-format.md) · [recipes](docs/recipes.md) · [troubleshooting](docs/troubleshooting.md) · [vs other tools](docs/comparison.md) · [CI workflow](docs/reusable-workflow.md)

## Why yamlspec?

- **RSpec-like syntax** — `describe`/`it`/`should` vocabulary developers already know
- **20+ assertion operators** — equality, comparisons, strings, regex, existence, arrays, sets
- **Engine agnostic** — test plain YAML, Helm charts, Kustomize overlays, or anything that outputs YAML
- **6 output formats** — console, JSON, YAML, Markdown, enriched Markdown (GitHub PRs), JUnit XML
- **Reusable CI workflow** — one-line GitHub Actions setup with PR commenting
- **Zero YAML dependencies** — no Helm or Kustomize libraries; rendering happens via `pre_run` shell commands
- **5 direct Go dependencies** — minimal, fast, easy to build

## Install

No tagged release yet — install from source:

```bash
# Latest commit on main
go install github.com/AxeForging/yamlspec@main

# Or clone and build
git clone https://github.com/AxeForging/yamlspec.git
cd yamlspec
make install        # builds and installs to /usr/local/bin
```

Once v0.1.0 is published you'll also be able to use:

```bash
# Direct download (post-release)
curl -sSL https://github.com/AxeForging/yamlspec/releases/latest/download/yamlspec-linux-amd64.tar.gz | tar xz
sudo mv yamlspec /usr/local/bin/
```

See [CHANGELOG.md](CHANGELOG.md) for release status.

## Quick Start

```bash
# Scaffold a test
yamlspec init my-feature --test-dir tests

# Run tests
yamlspec validate --test-dir tests

# Filter by tag
yamlspec validate --tag deployment

# Multiple output formats
yamlspec validate --json-output results.json --junit-output results.xml

# Parallel execution
yamlspec validate --workers 4
```

## Sample output

```
  ✓ Production deployment
    Deployment configuration
      [select: select(.kind == "Deployment")]
      ✓ have 3 replicas
      ✓ use a pinned image tag
      ✓ have resource limits
      ✓ be in production namespace

  ✗ Staging overlay
    Deployment
      [select: select(.kind == "Deployment")]
      ✓ be in staging namespace
      ✗ have 2 replicas

Failures:

  1) Staging overlay > Deployment > have 2 replicas
     expected 2, got 1

26 assertions, 25 passed, 1 failed
Finished in 0.18s
```

In CI, the same run also produces a JUnit XML, an enriched-Markdown summary
(rendered as a GitHub PR comment and on the workflow run page), and inline
`::error file=...,line=...::` annotations against `spec.yaml` lines.

## Usage with Helm Charts

Use `pre_run` to render templates before validation:

```yaml
# tests/production/spec.yaml
name: "Production values"
tags: ["production", "helm"]

pre_run:
  - helm template my-app ../../ -f values.yaml > manifests/rendered.yaml

describe:
  - name: "Deployment"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "have 3+ replicas for HA"
        expect: spec.replicas
        toBeGreaterOrEqual: 3

      - should: "not use latest tag"
        expect: spec.template.spec.containers[0].image
        toNotEndWith: ":latest"

      - should: "have resource limits"
        expect: spec.template.spec.containers[0].resources.limits
        toExist: true
```

## Usage with Kustomize

```yaml
# tests/staging/spec.yaml
name: "Staging overlay"
tags: ["staging", "kustomize"]

pre_run:
  - kustomize build ../../overlays/staging > manifests/rendered.yaml

describe:
  - name: "Deployment"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "be in staging namespace"
        expect: metadata.namespace
        toEqual: "staging"

      - should: "have env label"
        expect: metadata.labels.env
        toEqual: "staging"
```

## Usage with Plain Manifests

No `pre_run` needed — just put YAML files in a `manifests/` directory:

```
tests/my-feature/
  spec.yaml
  manifests/
    deployment.yaml
    service.yaml
```

## spec.yaml

Full schema reference: [docs/spec-format.md](docs/spec-format.md).

Tests are defined in `spec.yaml` files with an RSpec-like vocabulary:

```yaml
name: "Production deployment"
tags: ["deployment", "production"]

pre_run:
  - helm template my-app ../../chart -f values.yaml > manifests/rendered.yaml

describe:
  - name: "Deployment configuration"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "have 3 replicas"
        expect: spec.replicas
        toEqual: 3

      - should: "use a pinned image tag"
        expect: spec.template.spec.containers[0].image
        toNotEndWith: ":latest"

      - should: "have resource limits"
        expect: spec.template.spec.containers[0].resources.limits
        toExist: true

      - should: "be in production namespace"
        expect: metadata.namespace
        toBeOneOf: ["production", "prod"]
```

## Directory Structure

```
tests/
  my-feature/
    spec.yaml              # Test specification
    manifests/             # YAML files to test
      deployment.yaml
      service.yaml
    values.yaml            # Optional: Helm values for pre_run
```

## Field paths

The `expect:` field accepts these syntaxes — mix freely:

```yaml
expect: spec.replicas                              # dotted path
expect: spec.template.spec.containers[0].image     # array index
expect: spec.template.spec.containers[*].image     # wildcard — assertion runs against every element
expect: metadata.labels["app.kubernetes.io/name"]  # bracket notation for keys with dots/special chars
expect: .spec.replicas                             # leading dot (JQ-style) also works
```

Wildcards (`[*]`) iterate every element — useful for "every container must have resource limits" style checks. See [docs/spec-format.md](docs/spec-format.md#field-path-syntax) for the full reference.

## Assertion Operators

| Operator | Example | Description |
|----------|---------|-------------|
| `toEqual` | `toEqual: 3` | Exact match |
| `toNotEqual` | `toNotEqual: "default"` | Must not match |
| `toBeGreaterThan` | `toBeGreaterThan: 2` | Numeric > |
| `toBeLessThan` | `toBeLessThan: 10` | Numeric < |
| `toBeGreaterOrEqual` | `toBeGreaterOrEqual: 1` | Numeric >= |
| `toBeLessOrEqual` | `toBeLessOrEqual: 5` | Numeric <= |
| `toContain` | `toContain: "nginx"` | Substring match |
| `toNotContain` | `toNotContain: "debug"` | Must not contain |
| `toStartWith` | `toStartWith: "nginx:"` | Prefix |
| `toEndWith` | `toEndWith: "-alpine"` | Suffix |
| `toNotStartWith` | `toNotStartWith: "alpine:"` | Must not start with |
| `toNotEndWith` | `toNotEndWith: ":latest"` | Must not end with |
| `toMatch` | `toMatch: "^v\\d+"` | Regex |
| `toExist` | `toExist: true` | Field exists |
| `toBeNull` | `toBeNull: false` | Null check |
| `toHaveKey` | `toHaveKey: "app"` | Object has key |
| `toBeOneOf` | `toBeOneOf: ["a", "b"]` | Set membership |
| `toNotBeOneOf` | `toNotBeOneOf: ["x"]` | Not in set |
| `toContainItem` | `toContainItem: 80` | Array contains |
| `toHaveLength` | `toHaveLength: 3` | Exact length |
| `toHaveMinLength` | `toHaveMinLength: 1` | Min length |
| `toHaveMaxLength` | `toHaveMaxLength: 5` | Max length |

Operators can be combined — all must pass:

```yaml
- should: "be a valid image"
  expect: spec.template.spec.containers[0].image
  toStartWith: "nginx:"
  toNotEndWith: ":latest"
  toMatch: ":\\d+\\.\\d+\\.\\d+$"
```

## Output Formats

| Flag | Format | Use case |
|------|--------|----------|
| (default) | Console | RSpec-like colored tree |
| `--json-output` | JSON | Programmatic consumption |
| `--yaml-output` | YAML | YAML-native workflows |
| `--markdown-output` | Markdown | Documentation |
| `--emd-output` | Enriched MD | GitHub PR comments |
| `--junit-output` | JUnit XML | CI/CD integration |

## CLI Reference

```
COMMANDS:
  validate, test, run   Run test specs against YAML manifests
  list, ls              List discovered specs and tags
  init                  Scaffold a new test spec
  version               Show version information

VALIDATE FLAGS:
  --test-dir, -d        Test directory (default: "tests")
  --tag, -t             Filter by tag (repeatable)
  --workers, -w         Parallel workers (default: 1)
  --fail-fast           Stop on first failure
  --pre-run-timeout     Max duration for each pre_run command (default: 60s)
  --quiet, -q           Summary only
  --json-output         JSON output file
  --yaml-output         YAML output file
  --markdown-output     Markdown output file
  --emd-output          Enriched markdown output file
  --junit-output        JUnit XML output file
  --github-annotations  Emit ::error file=...,line=... for failing assertions
                        (auto-enabled when GITHUB_ACTIONS=true)
```

## CI/CD — Reusable Workflow

Add yamlspec to any repo with one workflow file:

```yaml
# .github/workflows/yamlspec.yml
name: yamlspec
on: [pull_request, push]

jobs:
  test:
    uses: AxeForging/yamlspec/.github/workflows/reusable.yml@main
    with:
      test-dir: tests
      install-helm: true        # if your specs use helm template
      install-kustomize: true   # if your specs use kustomize build
      comment-on-pr: true       # post results as PR comment
```

This installs yamlspec, runs your specs, generates JSON/EMD/JUnit artifacts, and posts a collapsible results comment on PRs. See [docs/reusable-workflow.md](docs/reusable-workflow.md) for all options.

## Examples

Five runnable examples live under [`examples/`](examples) — see the [examples README](examples/README.md) for what each demonstrates and the exact command to run it.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full guide. Quick reference:

```bash
make build-local    # Build binary
make test           # Run all tests
make test-unit      # Unit tests only
make test-e2e       # Integration tests
make lint           # Linter
make install        # Install to /usr/local/bin
```

Release history is in [CHANGELOG.md](CHANGELOG.md).

## License

MIT
