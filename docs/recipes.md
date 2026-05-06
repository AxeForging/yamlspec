# Recipes

Patterns that come up over and over. Copy, adapt, run.

For the full schema and operator list see [spec-format.md](spec-format.md).

## Multi-environment Helm values

One spec per environment, each rendering against its own values file.
Filesystem layout:

```
tests/
  default/
    spec.yaml
  staging/
    spec.yaml
    values.yaml
  production/
    spec.yaml
    values.yaml
```

```yaml
# tests/production/spec.yaml
name: "Production overrides"
tags: ["production"]

pre_run:
  - helm template prod ../../chart -f values.yaml > manifests/rendered.yaml

describe:
  - name: "Deployment"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "scale up for prod traffic"
        expect: spec.replicas
        toBeGreaterOrEqual: 3

      - should: "land in the prod namespace"
        expect: metadata.namespace
        toEqual: "production"
```

Run all environments at once with `yamlspec validate --test-dir tests`, or
filter with `--tag production`.

## Every container has resource limits (wildcard iteration)

`[*]` runs the assertion against every element. Failure on any one fails the
assertion.

```yaml
- should: "every container has CPU limits"
  expect: spec.template.spec.containers[*].resources.limits.cpu
  toExist: true

- should: "every container has memory limits"
  expect: spec.template.spec.containers[*].resources.limits.memory
  toExist: true

- should: "no container uses :latest"
  expect: spec.template.spec.containers[*].image
  toNotEndWith: ":latest"
```

This is one of the most useful patterns — multi-container pods (sidecars,
init containers) silently regressing is a common production incident.

## Image pinning with semver

```yaml
- should: "image is pinned to a semver tag"
  expect: spec.template.spec.containers[0].image
  toMatch: ':v?\d+\.\d+\.\d+(-[a-z0-9.-]+)?$'
  toNotEndWith: ":latest"
```

For SHA-pinned images:

```yaml
- should: "image is pinned to a digest"
  expect: spec.template.spec.containers[0].image
  toMatch: '@sha256:[a-f0-9]{64}$'
```

## Required labels and annotations (the standard k8s set)

```yaml
- name: "Standard recommended labels"
  select: 'select(.kind == "Deployment")'
  it:
    - should: "have app.kubernetes.io/name"
      expect: metadata.labels["app.kubernetes.io/name"]
      toExist: true

    - should: "have app.kubernetes.io/instance"
      expect: metadata.labels["app.kubernetes.io/instance"]
      toExist: true

    - should: "have app.kubernetes.io/version"
      expect: metadata.labels["app.kubernetes.io/version"]
      toMatch: '^v?\d+\.\d+\.\d+'

    - should: "have app.kubernetes.io/managed-by"
      expect: metadata.labels["app.kubernetes.io/managed-by"]
      toBeOneOf: ["Helm", "kustomize", "argocd"]
```

Bracket notation is required for keys containing dots or slashes.

## ConfigMap data validation

```yaml
- name: "App configuration"
  select: 'select(.kind == "ConfigMap" and .metadata.name == "app-config")'
  it:
    - should: "set the right environment"
      expect: data.APP_ENV
      toEqual: "production"

    - should: "use a valid log level"
      expect: data.LOG_LEVEL
      toBeOneOf: ["debug", "info", "warn", "error"]

    - should: "not enable debug in prod"
      expect: data.DEBUG
      toNotEqual: "true"

    - should: "have a sane request timeout"
      expect: data.REQUEST_TIMEOUT_MS
      toMatch: '^\d+$'
```

## Pod security baseline

```yaml
- name: "Pod security context"
  select: 'select(.kind == "Deployment")'
  it:
    - should: "run as non-root"
      expect: spec.template.spec.securityContext.runAsNonRoot
      toEqual: true

    - should: "set a non-root UID"
      expect: spec.template.spec.securityContext.runAsUser
      toBeGreaterThan: 0

    - should: "disable service account token automount"
      expect: spec.template.spec.automountServiceAccountToken
      toEqual: false

- name: "Container security context"
  select: 'select(.kind == "Deployment")'
  it:
    - should: "every container has a read-only root filesystem"
      expect: spec.template.spec.containers[*].securityContext.readOnlyRootFilesystem
      toEqual: true

    - should: "every container drops all capabilities"
      expect: spec.template.spec.containers[*].securityContext.capabilities.drop
      toContainItem: "ALL"

    - should: "no container allows privilege escalation"
      expect: spec.template.spec.containers[*].securityContext.allowPrivilegeEscalation
      toEqual: false
```

## HPA / PDB invariants

```yaml
- name: "HorizontalPodAutoscaler"
  select: 'select(.kind == "HorizontalPodAutoscaler")'
  it:
    - should: "have min replicas in [3, 10]"
      expect: spec.minReplicas
      toBeGreaterOrEqual: 3
      toBeLessOrEqual: 10

    - should: "have max replicas <= 50"
      expect: spec.maxReplicas
      toBeLessOrEqual: 50

- name: "PodDisruptionBudget"
  select: 'select(.kind == "PodDisruptionBudget")'
  it:
    - should: "guarantee at least 2 pods stay available"
      expect: spec.minAvailable
      toBeGreaterOrEqual: 2
```

## NetworkPolicy egress rules

```yaml
- name: "Default-deny + allow-list"
  select: 'select(.kind == "NetworkPolicy")'
  it:
    - should: "include Egress in policy types"
      expect: spec.policyTypes
      toContainItem: "Egress"

    - should: "include Ingress in policy types"
      expect: spec.policyTypes
      toContainItem: "Ingress"

    - should: "have at least one egress rule (not pure deny-all)"
      expect: spec.egress
      toHaveMinLength: 1
```

## Service / Ingress sanity

```yaml
- name: "Service"
  select: 'select(.kind == "Service" and .metadata.name == "web")'
  it:
    - should: "be ClusterIP (not LoadBalancer in this env)"
      expect: spec.type
      toEqual: "ClusterIP"

    - should: "expose exactly one port"
      expect: spec.ports
      toHaveLength: 1

    - should: "select pods by app label"
      expect: spec.selector.app
      toEqual: "web"

- name: "Ingress"
  select: 'select(.kind == "Ingress")'
  it:
    - should: "use TLS"
      expect: spec.tls
      toHaveMinLength: 1

    - should: "use the prod ingress class"
      expect: spec.ingressClassName
      toEqual: "nginx-prod"
```

## Cross-overlay assertion (Kustomize)

```yaml
# tests/staging/spec.yaml
pre_run:
  - kustomize build ../../overlays/staging > manifests/rendered.yaml

describe:
  - name: "Staging-specific overrides"
    select: 'select(.kind == "Deployment" and .metadata.name == "api")'
    it:
      - should: "use the staging image registry"
        expect: spec.template.spec.containers[0].image
        toStartWith: "registry.staging.example.com/"

      - should: "have the env label patched in"
        expect: metadata.labels.env
        toEqual: "staging"
```

## Multi-document manifests

yamlspec parses every `---`-separated document in every YAML file under
`manifests/`. No special config needed:

```yaml
# manifests/rendered.yaml (output of helm template or kustomize build)
apiVersion: v1
kind: Service
...
---
apiVersion: apps/v1
kind: Deployment
...
---
apiVersion: v1
kind: ConfigMap
...
```

Each `describe` block uses its `select:` to pick which documents the
assertions apply to.

## Combining multiple operators

Operators on a single assertion are AND'd — all must pass:

```yaml
- should: "image is on our registry, semver-tagged, and not :latest"
  expect: spec.template.spec.containers[0].image
  toStartWith: "registry.example.com/"
  toMatch: ':v\d+\.\d+\.\d+$'
  toNotEndWith: ":latest"
```

For OR semantics, write multiple assertions or use `toBeOneOf`.

## Reusable security baseline

The [`examples/security-checks`](../examples/security-checks) spec is a
ready-to-run baseline you can drop into any chart or kustomize repo as
`tests/security/spec.yaml`.

```bash
yamlspec validate --test-dir tests --tag security
```
