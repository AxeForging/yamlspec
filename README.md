# yamlspec

YAML test framework with RSpec-like assertions. Validate any YAML manifests — Kubernetes, Helm, Kustomize, or plain files — with a clean, readable syntax.

## Install

```bash
# From source
go install github.com/AxeForging/yamlspec@latest

# Or download a release
curl -sSL https://github.com/AxeForging/yamlspec/releases/latest/download/yamlspec-linux-amd64.tar.gz | tar xz
sudo mv yamlspec /usr/local/bin/
```

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
```

## spec.yaml

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
  --quiet, -q           Summary only
  --json-output         JSON output file
  --yaml-output         YAML output file
  --markdown-output     Markdown output file
  --emd-output          Enriched markdown output file
  --junit-output        JUnit XML output file
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

## Development

```bash
make build-local    # Build binary
make test           # Run all tests
make test-unit      # Unit tests only
make test-e2e       # Integration tests
make lint           # Linter
make install        # Install to /usr/local/bin
```

## License

MIT
