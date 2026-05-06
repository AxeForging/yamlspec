# Examples

Five runnable examples covering the main use cases. Each lives in its own
directory and is a complete, self-contained spec — clone the repo and run them
locally to see what passing/failing output looks like.

## Running an example

From the repo root, after `make build-local` (binary lands at `./yamlspec`):

```bash
./yamlspec validate --test-dir examples/<example-name>
```

Or run all examples at once:

```bash
./yamlspec validate --test-dir examples
```

Filter by tag:

```bash
./yamlspec validate --test-dir examples --tag security
```

## What each example shows

### [`plain-manifests/`](plain-manifests)

The simplest case — no `pre_run`, no rendering. Just YAML files in
`manifests/` and a `spec.yaml` next to them.

Demonstrates: deployment replica count, image-tag policy (no `:latest`,
semver-only), resource limits/requests, health-probe presence, ConfigMap
key/value checks, set-membership (`toBeOneOf` for log levels).

```bash
./yamlspec validate --test-dir examples/plain-manifests
```

### [`helm-chart/`](helm-chart)

Full Helm pipeline — `pre_run` runs `helm template` against a real chart with
a values override, writes the output to `manifests/rendered.yaml`, then
yamlspec asserts on the rendered result.

Demonstrates: Helm rendering integration, asserting that `values.yaml`
overrides actually take effect, multi-resource validation in one spec.

Requires `helm` on `$PATH`.

```bash
./yamlspec validate --test-dir examples/helm-chart
```

### [`kustomize-overlay/`](kustomize-overlay)

Kustomize equivalent — `pre_run` runs `kustomize build` against a base, asserts
on the built output.

Demonstrates: kustomize integration, namespace/label patches, replica patches.

Requires `kustomize` on `$PATH`.

```bash
./yamlspec validate --test-dir examples/kustomize-overlay
```

### [`multi-resource/`](multi-resource)

Several resources of different kinds (Deployment + Service + HPA + PDB) tested
in a single spec, each isolated via JQ selector.

Demonstrates: `select: 'select(.kind == "X")'` filtering, asserting
relationships between resources (HPA targets the Deployment, PDB selector
matches the pods), array-length checks (`toHaveLength`, `toHaveMinLength`).

```bash
./yamlspec validate --test-dir examples/multi-resource
```

### [`security-checks/`](security-checks)

A reusable security baseline — runAsNonRoot, readOnlyRootFilesystem, dropped
capabilities, no `:latest` tags, mandatory probes, mandatory resource limits.

Demonstrates: how to write a security-policy spec you can drop into any chart
or kustomize repo. No `pre_run` — point it at any rendered manifest with
`./yamlspec validate --test-dir examples/security-checks`.

This example is also a good template for organization-wide policy checks.

## Output formats

Try the different formatters on any example:

```bash
./yamlspec validate --test-dir examples/multi-resource \
  --json-output results.json \
  --emd-output results.md \
  --junit-output results.xml
```

See [docs/spec-format.md](../docs/spec-format.md) for the full assertion
reference and [docs/recipes.md](../docs/recipes.md) for more patterns.
