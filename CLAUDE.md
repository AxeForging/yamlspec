# AI Assistant Documentation for yamlspec

## Project Overview

yamlspec is a YAML test framework with RSpec-like assertions. It validates any YAML manifests (Kubernetes, Helm, Kustomize, or plain files) using a clean `describe`/`it`/`should` syntax.

## Architecture

```
yamlspec/
├── main.go              # CLI entry point (urfave/cli v1)
├── flags.go             # CLI flag definitions
├── actions/             # Command handlers (validate, list, init, version)
├── domain/              # Models (Spec, Results, Config)
├── services/            # Business logic
│   ├── assertion.go     # JQ-based assertion engine (20+ operators)
│   ├── discovery.go     # Spec file discovery
│   ├── runner.go        # Test execution with parallel support
│   ├── formatter.go     # Formatter interface
│   └── fmt_*.go         # Output formatters (console, json, yaml, md, emd, junit)
├── helpers/             # Utilities (errors, logger, terminal detection)
├── integration/         # E2E tests with testdata fixtures
├── examples/            # Example test suites
└── docs/                # Documentation
```

## Key Design Decisions

1. **No rendering libraries** — Helm/Kustomize rendering via `pre_run` shell commands only
2. **RSpec vocabulary** — `describe`/`it`/`should` in spec.yaml
3. **Structured assertions only** — no raw JQ `assert:` field
4. **urfave/cli v1** — consistent with AxeForging org conventions
5. **5 direct dependencies** — gojq, zerolog, yaml.v3, urfave/cli, go-isatty

## Common Tasks

### Adding a new assertion operator

1. Add field to `domain.Assertion` struct in `domain/models.go`
2. Add to `HasValueOperators()` if it's value-based
3. Add evaluation logic in `services/assertion.go` `evaluateOperators()`
4. Add tests in `services/assertion_test.go`

### Adding a new output format

1. Create `services/fmt_<name>.go` implementing the `Formatter` interface
2. Add constructor `NewXFormatter()`
3. Wire into `actions/validate.go` formatters map
4. Add CLI flag in `flags.go`
5. Add to validate command flags in `main.go`

### Adding a new CLI command

1. Create `actions/<command>.go` with action struct and `Execute` method
2. Wire into `main.go` commands list

## Testing

```bash
make test          # All tests (unit + integration)
make test-unit     # Unit tests only (services/)
make test-e2e      # Integration tests (build binary, run against fixtures)
```

## Dependencies

- `github.com/urfave/cli` — CLI framework
- `github.com/itchyny/gojq` — JQ expression evaluation
- `github.com/rs/zerolog` — Structured logging
- `gopkg.in/yaml.v3` — YAML parsing
- `github.com/mattn/go-isatty` — Terminal detection
