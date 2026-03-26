# spec.yaml Format Reference

## Structure

```yaml
name: "Human-readable test name"
tags: ["tag1", "tag2"]

pre_run:
  - shell command 1
  - shell command 2

describe:
  - name: "Describe block name"
    select: 'JQ selector expression'
    it:
      - should: "assertion description"
        expect: field.path
        <operator>: <value>
```

## Fields

### Top-level

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Test name (defaults to directory name) |
| `tags` | string[] | No | Tags for filtering (`--tag`) |
| `pre_run` | string[] | No | Shell commands to run before testing |
| `describe` | DescBlock[] | Yes | Assertion groups |

### DescBlock

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Block name (shown in output) |
| `select` | string | No | JQ expression to filter manifests |
| `it` | Assertion[] | Yes | Individual assertions |

### Assertion

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `should` | string | Yes | Human-readable description |
| `expect` | string | Yes | Field path to inspect |
| `<operator>` | varies | Yes | One or more assertion operators |

## Field Path Syntax

```yaml
# Simple dotted path
expect: spec.replicas

# Array index
expect: spec.containers[0].image

# Bracket notation (for keys with dots/special chars)
expect: metadata.labels["app.kubernetes.io/name"]

# Leading dot (JQ-style, also works)
expect: .spec.replicas

# Array iteration
expect: spec.containers[*].image
```

## Assertion Operators

### Equality
```yaml
toEqual: 3                    # exact match
toNotEqual: "default"         # must not match
```

### Numeric Comparison
```yaml
toBeGreaterThan: 2            # >
toBeLessThan: 10              # <
toBeGreaterOrEqual: 1         # >=
toBeLessOrEqual: 5            # <=
```

### String
```yaml
toContain: "nginx"            # substring match
toNotContain: "debug"         # must not contain
toStartWith: "nginx:"         # prefix match
toEndWith: ":latest"          # suffix match
toNotStartWith: "alpine:"     # must not start with
toNotEndWith: ":latest"       # must not end with
toMatch: "^nginx:\\d+\\.\\d+" # regex match
```

### Existence
```yaml
toExist: true                 # field must exist and be non-null
toExist: false                # field must not exist
toBeNull: true                # field must be null
toBeNull: false               # field must not be null
```

### Object
```yaml
toHaveKey: "app"              # object must have this key
```

### Set Membership
```yaml
toBeOneOf: ["production", "staging"]     # must be in list
toNotBeOneOf: ["default", "kube-system"] # must not be in list
```

### Array
```yaml
toContainItem: 80             # array must contain value
toHaveLength: 3               # exact array length
toHaveMinLength: 1            # minimum array length
toHaveMaxLength: 5            # maximum array length
```

## Multiple Operators

You can combine operators on a single assertion. All must pass:

```yaml
- should: "be in valid range"
  expect: spec.replicas
  toBeGreaterOrEqual: 1
  toBeLessOrEqual: 10

- should: "be a proper image reference"
  expect: spec.template.spec.containers[0].image
  toStartWith: "nginx:"
  toNotEndWith: ":latest"
  toMatch: ":\\d+\\.\\d+\\.\\d+$"
```

## Selectors

Selectors use JQ `select()` syntax:

```yaml
# By kind
select: 'select(.kind == "Deployment")'

# By name
select: 'select(.metadata.name == "my-app")'

# By label
select: 'select(.metadata.labels.app == "web")'

# Combined
select: 'select(.kind == "Deployment" and .metadata.namespace == "production")'

# Empty selector matches all manifests
select: ""
```

## Pre-run Commands

Commands run with `sh -c` in the spec's directory:

```yaml
pre_run:
  # Helm template rendering
  - helm template my-release ../../chart -f values.yaml > manifests/rendered.yaml

  # Kustomize build
  - kustomize build ../../overlays/production > manifests/rendered.yaml

  # Any shell command
  - cat base.yaml overrides.yaml > manifests/merged.yaml
```
