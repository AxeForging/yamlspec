# yamlspec vs other tools

There's no shortage of tools that touch YAML/Kubernetes/Helm validation, and
they overlap in confusing ways. Here's an honest read on where yamlspec fits
and when something else is the better pick.

## TL;DR

| Tool | Primary purpose | Pick it over yamlspec when |
|------|-----------------|----------------------------|
| **yamlspec** | Per-environment behavioral assertions on rendered YAML | You want RSpec-style "this deployment, in this overlay, should have these properties" |
| [helm unittest](https://github.com/helm-unittest/helm-unittest) | Unit testing Helm templates | You're a chart author and want to test the templating logic itself |
| [chart-testing (`ct`)](https://github.com/helm/chart-testing) | Lint + install + smoke-test Helm charts | You want a real cluster to actually `helm install` your chart and confirm it comes up |
| [conftest](https://www.conftest.dev/) / [OPA](https://www.openpolicyagent.org/) | Policy-as-code with Rego | You already use Rego elsewhere or need policies enforced across many tool boundaries (admission controllers, Terraform, Dockerfiles) |
| [kyverno](https://kyverno.io/) | Cluster admission policy | You want runtime enforcement, not pre-merge testing |
| [datree](https://www.datree.io/) | Built-in best-practice rules + custom policies | You want a curated rule library out of the box |
| [kubeconform](https://github.com/yannh/kubeconform) | Schema validation against Kubernetes API | You only need "is this valid Kubernetes YAML" — not behavioral assertions |

## The detailed reads

### vs `helm unittest`

`helm unittest` tests **template logic**. yamlspec tests **rendered output**.

```yaml
# helm unittest — asserts what the template did
asserts:
  - equal:
      path: spec.replicas
      value: 3
    template: deployment.yaml
    set:
      replicaCount: 3
```

```yaml
# yamlspec — asserts what the rendered manifest looks like in a real values context
pre_run:
  - helm template app ../../chart -f values-prod.yaml > manifests/rendered.yaml
describe:
  - name: "Deployment"
    select: 'select(.kind == "Deployment")'
    it:
      - should: "have 3 replicas in production"
        expect: spec.replicas
        toEqual: 3
```

**Pick `helm unittest` when:** you're a chart author and you want fast,
isolated tests of the templating itself — "if I set this value, the template
emits that output."

**Pick yamlspec when:** you maintain a chart-consumer repo (or your own chart
+ values) and want to assert "given my actual production values, the result
has the properties production needs." It also works on plain manifests and
Kustomize, which `helm unittest` does not.

You can use both. They answer different questions.

### vs `chart-testing` (`ct`)

`ct` lints chart structure (Chart.yaml, README, version bumps), then optionally
spins up a kind cluster and `helm install`s the chart to confirm pods come up.
That's a smoke test of the chart's runtime behavior.

yamlspec is a static check against rendered YAML. No cluster required, runs in
seconds, and checks **what's in the manifest** rather than **whether it
deploys**.

**Pick `ct` when:** you want CI to catch "this chart no longer installs" or
"this chart needs a version bump."

**Pick yamlspec when:** you want to catch "production values dropped resource
limits" or "the staging overlay forgot the env label" before the manifest ever
reaches a cluster.

These are complementary — many repos run both.

### vs `conftest` / OPA / Rego

conftest runs Rego policies against any structured input. It's strictly more
powerful than yamlspec — Rego can do things yamlspec can't.

The trade-off is the language. Rego has a steep learning curve, error messages
are notoriously cryptic, and a simple "deployment must have 3 replicas" check
takes more lines and more cognitive overhead than the yamlspec equivalent.

```rego
# conftest
deny[msg] {
  input.kind == "Deployment"
  input.spec.replicas != 3
  msg := sprintf("expected 3 replicas, got %d", [input.spec.replicas])
}
```

```yaml
# yamlspec
- should: "have 3 replicas"
  expect: spec.replicas
  toEqual: 3
```

**Pick conftest when:** you're already running Rego policies elsewhere
(Terraform, admission control, Dockerfile linting), or you need logic that
genuinely requires a real policy language — multi-resource correlations, set
algebra, complex string parsing.

**Pick yamlspec when:** the team isn't going to invest in learning Rego, and
your assertions are mostly "this field should equal/match/contain this value."
Spec files are readable by anyone who's written a unit test in any language.

### vs Kyverno

Kyverno is a cluster admission controller. It rejects bad manifests at the
apiserver, generates resources, mutates incoming requests. It's a runtime
control plane, not a CI tool.

You can run Kyverno policies against YAML files via the `kyverno` CLI, but the
policy syntax is shaped by its admission-control origins — the YAML is more
verbose than yamlspec for equivalent checks.

**Pick Kyverno when:** you want runtime enforcement (the cluster rejects
non-conformant resources). yamlspec doesn't replace Kyverno; they operate at
different stages.

**Pick yamlspec when:** you want pre-merge feedback in a PR. You can use
yamlspec's checks as a pre-image of your Kyverno policies — catch violations
before they hit the cluster, then let Kyverno enforce at runtime.

### vs Datree

Datree ships a curated library of best-practice rules ("don't use latest tag",
"set resource limits", etc.) plus the ability to write custom policies in
JSONSchema or Rego.

If your goal is "I want a sane default rule set without writing anything
myself," Datree is the fast path. yamlspec doesn't ship policies — you author
them yourself, though [examples/security-checks](../examples/security-checks)
is a usable starting baseline.

**Pick Datree when:** you want managed best-practice rules out of the box and
don't want to hand-write policies.

**Pick yamlspec when:** you have specific, environment-dependent assertions
that don't map to generic best-practice rules. "Production has 3 replicas,
staging has 1, both are in the right namespace" is a yamlspec test, not a
Datree rule.

### vs `kubeconform` / `kubeval`

These check "is this valid Kubernetes YAML according to the OpenAPI schema?"
That's structural validation — type/required-field checking — not behavioral.

**Pick kubeconform when:** you want to catch "I typoed `containrs` instead of
`containers`" or "this field doesn't exist in 1.29." It's fast and cheap and
should run in every CI pipeline that produces manifests.

**Pick yamlspec when:** the YAML is valid but the *content* might be wrong.
Schema-valid manifests can still set `replicas: 1` in production or use
`:latest` images.

These are layered: run kubeconform for structure, yamlspec for behavior.

## When yamlspec is *not* the right answer

Be honest with yourself:

- **You need to enforce policies at admission time** — use Kyverno or OPA
  Gatekeeper.
- **You need to test that a chart actually deploys** — use `chart-testing` or
  hit a kind cluster.
- **You need a curated, maintained rule set with zero authoring effort** —
  use Datree.
- **You want to test templating logic in isolation as a chart author** — use
  `helm unittest`.
- **You need cross-tool policy consistency** (Terraform + K8s +
  Dockerfiles + …) — use Rego/conftest.

yamlspec's sweet spot is the case those tools handle awkwardly: **per-environment
behavioral assertions on rendered YAML, written in a syntax developers can read
in 30 seconds**.
