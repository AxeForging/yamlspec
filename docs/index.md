# yamlspec documentation

A YAML test framework with RSpec-like assertions. Validate plain YAML, Helm
charts, Kustomize overlays, or anything else that produces YAML.

If you're new, start with the project [README](../README.md) for install and a
60-second tour.

## Reference

| Page | What's in it |
|------|--------------|
| [spec-format.md](spec-format.md) | Full `spec.yaml` schema — fields, field-path syntax, all 22 assertion operators, selector examples |
| [recipes.md](recipes.md) | Cookbook of common patterns: multi-environment Helm values, security policy checks, label conventions, ConfigMap data, image pinning |
| [troubleshooting.md](troubleshooting.md) | Common failure modes and how to read the error messages |
| [reusable-workflow.md](reusable-workflow.md) | GitHub Actions reusable workflow — one-line CI integration with PR comments |

## Project meta

| Page | What's in it |
|------|--------------|
| [../CONTRIBUTING.md](../CONTRIBUTING.md) | Dev setup, how to add an operator or output format, release process |
| [../CHANGELOG.md](../CHANGELOG.md) | Release history (Keep a Changelog format) |

## Examples

Five worked examples live under [`../examples/`](../examples) — see the
[examples README](../examples/README.md) for what each one demonstrates and the
exact command to run it.
