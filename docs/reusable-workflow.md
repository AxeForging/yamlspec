# Reusable Workflow

[← back to docs index](index.md)


yamlspec provides a reusable GitHub Actions workflow that any repo can call to run tests and post results as PR comments.

## Quick Start

### Plain YAML manifests

```yaml
# .github/workflows/yamlspec.yml
name: yamlspec
on: [pull_request, push]

jobs:
  test:
    uses: AxeForging/yamlspec/.github/workflows/reusable.yml@main
    with:
      test-dir: tests
```

### Helm chart repo

```yaml
name: yamlspec
on: [pull_request, push]

jobs:
  test:
    uses: AxeForging/yamlspec/.github/workflows/reusable.yml@main
    with:
      test-dir: tests
      install-helm: true
```

### Kustomize repo

```yaml
name: yamlspec
on: [pull_request, push]

jobs:
  test:
    uses: AxeForging/yamlspec/.github/workflows/reusable.yml@main
    with:
      test-dir: tests
      install-kustomize: true
```

### Helm + Kustomize

```yaml
name: yamlspec
on: [pull_request, push]

jobs:
  test:
    uses: AxeForging/yamlspec/.github/workflows/reusable.yml@main
    with:
      test-dir: tests
      install-helm: true
      install-kustomize: true
      workers: 4
```

## All Inputs

| Input | Default | Description |
|-------|---------|-------------|
| `test-dir` | `tests` | Directory containing test specs |
| `tag` | `""` | Filter specs by tag |
| `yamlspec-version` | `latest` | Version to install (`latest` or `v1.0.0`) |
| `fail-fast` | `false` | Stop on first failure |
| `workers` | `1` | Parallel workers |
| `install-helm` | `false` | Install Helm for `pre_run` commands |
| `helm-version` | `latest` | Helm version |
| `install-kustomize` | `false` | Install Kustomize for `pre_run` commands |
| `kustomize-version` | `v5.6.0` | Kustomize version |
| `comment-on-pr` | `true` | Post enriched markdown results as PR comment |
| `extra-args` | `""` | Additional yamlspec CLI arguments |

## What it does

1. Installs yamlspec (and optionally Helm/Kustomize)
2. Runs `yamlspec validate` with your configuration
3. Generates JSON, EMD, HTML, and JUnit XML results
4. Uploads results as workflow artifacts
5. Posts enriched markdown as a PR comment (updates existing comment on re-push)
6. Fails the job if any tests fail

The HTML artifact is a self-contained standards report with embedded run data,
styles, filters, original specs, and rendered manifests. It can be downloaded
from workflow artifacts and opened directly in a browser, or served locally
with `yamlspec serve --file yamlspec-results.html`.

## PR Comment

On pull requests, the workflow posts a collapsible results comment:

- Passing specs show as collapsed `<details>` blocks
- Failing specs are expanded with error details
- The comment is updated (not duplicated) on subsequent pushes
- Comment includes total counts and duration

## Example: Helm chart repo setup

```
my-helm-chart/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   └── service.yaml
├── tests/
│   ├── default/
│   │   └── spec.yaml
│   ├── production/
│   │   ├── values.yaml
│   │   └── spec.yaml
│   └── security/
│       ├── values.yaml
│       └── spec.yaml
└── .github/workflows/
    └── yamlspec.yml
```

**tests/default/spec.yaml:**
```yaml
name: "Default values"
tags: ["default"]

pre_run:
  - helm template test-release ../../ > manifests/rendered.yaml

describe:
  - name: "Deployment"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "have 1 replica by default"
        expect: spec.replicas
        toEqual: 1
```

**tests/production/spec.yaml:**
```yaml
name: "Production overrides"
tags: ["production"]

pre_run:
  - helm template prod-release ../../ -f values.yaml > manifests/rendered.yaml

describe:
  - name: "Deployment"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "have 3+ replicas"
        expect: spec.replicas
        toBeGreaterOrEqual: 3

      - should: "have resource limits"
        expect: spec.template.spec.containers[0].resources.limits
        toExist: true
```

## Example: Kustomize repo setup

```
my-kustomize-app/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml
│   └── service.yaml
├── overlays/
│   ├── staging/
│   │   └── kustomization.yaml
│   └── production/
│       └── kustomization.yaml
├── tests/
│   ├── staging/
│   │   └── spec.yaml
│   └── production/
│       └── spec.yaml
└── .github/workflows/
    └── yamlspec.yml
```

**tests/staging/spec.yaml:**
```yaml
name: "Staging overlay"
tags: ["staging"]

pre_run:
  - kustomize build ../../overlays/staging > manifests/rendered.yaml

describe:
  - name: "Deployment"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "be in staging namespace"
        expect: metadata.namespace
        toEqual: "staging"

      - should: "have 2 replicas"
        expect: spec.replicas
        toEqual: 2
```
