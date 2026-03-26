package actions

import (
	"fmt"

	"github.com/urfave/cli"
)

// AIHelpAction handles the ai-help command
type AIHelpAction struct{}

// NewAIHelpAction creates a new AIHelpAction
func NewAIHelpAction() *AIHelpAction {
	return &AIHelpAction{}
}

// Execute prints a comprehensive reference for AI assistants
func (a *AIHelpAction) Execute(c *cli.Context) error {
	fmt.Print(aiHelpText)
	return nil
}

const aiHelpText = `# yamlspec — AI Reference Guide

yamlspec validates YAML manifests using RSpec-like assertions.
It works with plain YAML files, Helm charts, and Kustomize overlays.

## Project Structure

Tests live in a directory (default: tests/), one subdirectory per test spec:

    tests/
      my-feature/
        spec.yaml              # Required: test definition
        manifests/             # Option A: static YAML files to test
          deployment.yaml
          service.yaml
        values.yaml            # Optional: used by pre_run commands

## spec.yaml Format

    name: "Human-readable test name"
    tags: ["tag1", "tag2"]              # optional, for filtering with --tag

    pre_run:                            # optional, shell commands run before testing
      - helm template my-release ../../chart -f values.yaml > manifests/rendered.yaml

    describe:
      - name: "Describe block name"
        select: 'select(.kind == "Deployment")'    # JQ expression to pick resources
        it:
          - should: "what this checks"
            expect: field.path
            <operator>: <value>

## Selectors

Selectors use JQ select() to pick which YAML documents to test:

    select: 'select(.kind == "Deployment")'
    select: 'select(.kind == "Service")'
    select: 'select(.metadata.name == "my-app")'
    select: 'select(.kind == "Deployment" and .metadata.namespace == "production")'
    select: ''                                      # empty = test all documents

## Field Paths

    expect: spec.replicas                                       # dotted path
    expect: spec.template.spec.containers[0].image              # array index
    expect: metadata.labels["app.kubernetes.io/name"]           # bracket notation
    expect: .spec.replicas                                      # leading dot (JQ-style)
    expect: spec.template.spec.containers[*].image              # array iteration

## Assertion Operators — Complete Reference

### Equality
    - should: "have 3 replicas"
      expect: spec.replicas
      toEqual: 3

    - should: "not be in default namespace"
      expect: metadata.namespace
      toNotEqual: "default"

### Numeric Comparison
    - should: "have more than 2 replicas"
      expect: spec.replicas
      toBeGreaterThan: 2

    - should: "have fewer than 10 replicas"
      expect: spec.replicas
      toBeLessThan: 10

    - should: "have at least 2 replicas"
      expect: spec.replicas
      toBeGreaterOrEqual: 2

    - should: "have at most 5 replicas"
      expect: spec.replicas
      toBeLessOrEqual: 5

### String
    - should: "contain nginx"
      expect: spec.template.spec.containers[0].image
      toContain: "nginx"

    - should: "not contain debug"
      expect: spec.template.spec.containers[0].image
      toNotContain: "debug"

    - should: "start with our registry"
      expect: spec.template.spec.containers[0].image
      toStartWith: "myregistry.io/"

    - should: "not use latest tag"
      expect: spec.template.spec.containers[0].image
      toNotEndWith: ":latest"

    - should: "end with a semver tag"
      expect: spec.template.spec.containers[0].image
      toEndWith: ":v1.0.0"

    - should: "not start with docker.io"
      expect: spec.template.spec.containers[0].image
      toNotStartWith: "docker.io/"

    - should: "match semver pattern"
      expect: spec.template.spec.containers[0].image
      toMatch: ':\d+\.\d+\.\d+$'

### Existence
    - should: "have resource limits"
      expect: spec.template.spec.containers[0].resources.limits
      toExist: true

    - should: "not have a nodeSelector"
      expect: spec.template.spec.nodeSelector
      toExist: false

### Null
    - should: "be null"
      expect: spec.template.spec.containers[0].securityContext
      toBeNull: true

    - should: "not be null"
      expect: metadata.name
      toBeNull: false

### Object Keys
    - should: "have app label"
      expect: metadata.labels
      toHaveKey: "app"

### Set Membership
    - should: "be a valid namespace"
      expect: metadata.namespace
      toBeOneOf: ["production", "staging"]

    - should: "not be a system namespace"
      expect: metadata.namespace
      toNotBeOneOf: ["kube-system", "kube-public", "default"]

### Array
    - should: "have exactly 2 ports"
      expect: spec.template.spec.containers[0].ports
      toHaveLength: 2

    - should: "have at least 1 container"
      expect: spec.template.spec.containers
      toHaveMinLength: 1

    - should: "have at most 3 containers"
      expect: spec.template.spec.containers
      toHaveMaxLength: 3

    - should: "contain port 80"
      expect: spec.ports
      toContainItem: 80

### Combining Multiple Operators (all must pass)
    - should: "have replicas in valid range"
      expect: spec.replicas
      toBeGreaterOrEqual: 2
      toBeLessOrEqual: 10

    - should: "use correct image"
      expect: spec.template.spec.containers[0].image
      toStartWith: "nginx:"
      toNotEndWith: ":latest"
      toMatch: ':\d+\.\d+\.\d+$'

## Complete Examples

### Example 1: Plain Manifests

    # tests/basic/manifests/deployment.yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: web-app
      namespace: production
    spec:
      replicas: 3
      template:
        spec:
          containers:
            - name: app
              image: "nginx:1.25.3"
              resources:
                limits:
                  cpu: "500m"
                  memory: "256Mi"

    # tests/basic/spec.yaml
    name: "Web app deployment"
    tags: ["deployment", "production"]

    describe:
      - name: "Deployment"
        select: 'select(.kind == "Deployment")'
        it:
          - should: "have 3 replicas"
            expect: spec.replicas
            toEqual: 3
          - should: "be in production"
            expect: metadata.namespace
            toEqual: "production"
          - should: "have resource limits"
            expect: spec.template.spec.containers[0].resources.limits
            toExist: true
          - should: "not use latest tag"
            expect: spec.template.spec.containers[0].image
            toNotEndWith: ":latest"

### Example 2: Helm Chart

    # tests/helm-prod/values.yaml
    replicas: 3
    image:
      tag: "v2.0.0"

    # tests/helm-prod/spec.yaml
    name: "Production Helm values"
    tags: ["helm", "production"]

    pre_run:
      - mkdir -p manifests
      - helm template my-release ../../chart -f values.yaml > manifests/rendered.yaml

    describe:
      - name: "Deployment"
        select: 'select(.kind == "Deployment")'
        it:
          - should: "have 3 replicas"
            expect: spec.replicas
            toEqual: 3
          - should: "use correct image tag"
            expect: spec.template.spec.containers[0].image
            toEndWith: ":v2.0.0"

### Example 3: Kustomize Overlay

    # tests/staging-overlay/spec.yaml
    name: "Staging overlay"
    tags: ["kustomize", "staging"]

    pre_run:
      - mkdir -p manifests
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

### Example 4: Security Checks

    # tests/security/spec.yaml
    name: "Security best practices"
    tags: ["security"]

    describe:
      - name: "Pod security"
        select: 'select(.kind == "Deployment")'
        it:
          - should: "run as non-root"
            expect: spec.template.spec.securityContext.runAsNonRoot
            toEqual: true
          - should: "disable privilege escalation"
            expect: spec.template.spec.containers[0].securityContext.allowPrivilegeEscalation
            toEqual: false
          - should: "drop all capabilities"
            expect: spec.template.spec.containers[0].securityContext.capabilities.drop
            toContainItem: "ALL"
          - should: "use read-only root filesystem"
            expect: spec.template.spec.containers[0].securityContext.readOnlyRootFilesystem
            toEqual: true

### Example 5: Multi-Resource Validation

    # tests/full-stack/spec.yaml
    name: "Full stack validation"
    tags: ["deployment", "service", "hpa"]

    describe:
      - name: "Deployment"
        select: 'select(.kind == "Deployment")'
        it:
          - should: "have 3+ replicas"
            expect: spec.replicas
            toBeGreaterOrEqual: 3
          - should: "have health checks"
            expect: spec.template.spec.containers[0].livenessProbe
            toExist: true

      - name: "Service"
        select: 'select(.kind == "Service")'
        it:
          - should: "be ClusterIP"
            expect: spec.type
            toEqual: "ClusterIP"

      - name: "HPA"
        select: 'select(.kind == "HorizontalPodAutoscaler")'
        it:
          - should: "scale between 3 and 10"
            expect: spec.minReplicas
            toBeGreaterOrEqual: 3
          - should: "cap at 10 max"
            expect: spec.maxReplicas
            toBeLessOrEqual: 10

## CLI Usage

    yamlspec validate                                  # run all tests in tests/
    yamlspec validate --test-dir my-tests              # custom test directory
    yamlspec validate --tag production                  # filter by tag
    yamlspec validate --tag helm --tag kustomize        # multiple tags (OR)
    yamlspec validate --workers 4                       # parallel execution
    yamlspec validate --fail-fast                       # stop on first failure
    yamlspec validate --quiet                           # summary only
    yamlspec validate --json-output results.json        # JSON output
    yamlspec validate --junit-output results.xml        # JUnit XML for CI
    yamlspec validate --emd-output results.md           # GitHub PR comment format
    yamlspec validate --markdown-output results.md      # plain markdown
    yamlspec validate --yaml-output results.yaml        # YAML output
    yamlspec init my-test                               # scaffold new test
    yamlspec list                                       # list all specs
    yamlspec list --tags                                # list all tags

## CI/CD — Reusable GitHub Actions Workflow

    # .github/workflows/yamlspec.yml
    name: yamlspec
    on: [pull_request]

    jobs:
      test:
        uses: AxeForging/yamlspec/.github/workflows/reusable.yml@main
        with:
          test-dir: tests
          install-helm: true
          install-kustomize: true
          comment-on-pr: true
`
