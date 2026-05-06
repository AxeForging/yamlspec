# Contributing

Thanks for considering a contribution. yamlspec aims to stay small, fast, and
opinionated — pull requests that keep it that way are very welcome.

## Dev setup

Requirements:

- Go 1.25+ (the minimum tracks the upstream stdlib security floor)
- `golangci-lint` for `make lint`
- [`lefthook`](https://github.com/evilmartians/lefthook) for pre-commit/pre-push hooks (optional but recommended)
- `helm` and/or `kustomize` if you want to run those examples locally

```bash
git clone git@github.com:AxeForging/yamlspec.git
cd yamlspec
go mod download
make build-local       # produces ./yamlspec
make test              # runs unit + integration tests
```

Install hooks (optional):

```bash
lefthook install
```

This wires up `gofmt`, `go vet`, `golangci-lint`, `structlint`, and
conventional-commits message validation on `pre-commit`, plus `govulncheck` on
`pre-push`.

## Project layout

```
yamlspec/
├── main.go              # CLI entry point (urfave/cli v1)
├── flags.go             # CLI flag definitions
├── actions/             # Command handlers (validate, list, init, version, ai-help)
├── domain/              # Models (Spec, Results, Config) — no business logic
├── services/            # Business logic
│   ├── assertion.go     # Assertion engine (operator evaluation, field paths)
│   ├── discovery.go     # Spec file discovery + strict YAML decode
│   ├── runner.go        # Test execution (sequential + parallel)
│   ├── runner_unix.go   # Unix process-group handling for pre_run timeouts
│   ├── runner_windows.go
│   ├── formatter.go     # Formatter interface
│   └── fmt_*.go         # Output formatters
├── helpers/             # Utilities (errors, logger, terminal detection)
├── integration/         # E2E tests with testdata fixtures
├── examples/            # Worked examples
└── docs/                # Public-facing documentation
```

Keep the layers honest: `actions/` orchestrates, `services/` does the work,
`domain/` is data only. New cross-cutting code generally belongs in
`services/`.

## Common contribution patterns

### Adding a new assertion operator

1. Add the field to `domain.Assertion` in `domain/models.go`. Pointer types
   (`*float64`, `*int`, `*bool`) for "not set" / "set to zero" disambiguation.
2. If it's value-based, add it to `Assertion.HasValueOperators()`. If it can
   stand alone (like `toExist` / `toBeNull`), wire it into `HasAnyOperator()`.
3. Add the evaluation logic to `services/assertion.go` `evaluateOperators()` —
   each operator gets its own short helper.
4. Add unit tests in `services/assertion_test.go`. Cover: passing case,
   failing case, type mismatch, missing field.
5. Update [docs/spec-format.md](docs/spec-format.md) and the README operator
   table.

### Adding a new output format

1. Create `services/fmt_<name>.go` implementing the `Formatter` interface
   (single method: `Format(*SuiteResult) ([]byte, error)`).
2. Add a constructor `NewXFormatter()`.
3. Wire it into the `formatters` map in `actions/validate.go`.
4. Add the CLI flag in `flags.go` and reference it from `main.go`.
5. Add an integration test in `integration/`.

### Adding a new CLI command

1. Create `actions/<command>.go` with a struct + `Execute(c *cli.Context)`
   method.
2. Add a constructor `NewXAction()`.
3. Wire it into the `app.Commands` slice in `main.go`.

## Testing

```bash
make test              # everything (unit + integration)
make test-unit         # services/ only — fast
make test-e2e          # integration/ — builds the binary first
make test-coverage     # writes coverage.html
```

Integration tests live in `integration/` and use `testdata/` fixtures. They
build the binary from source and run it as a subprocess against real spec
files, so they catch regressions in the actual CLI surface, not just the
library code.

## Linting and formatting

```bash
make lint              # golangci-lint
gofmt -l -w .          # auto-format (or let lefthook do it)
```

Pre-commit hooks (via `lefthook install`) catch these automatically.

## Commit conventions

Conventional Commits are enforced on the commit message. Format:

```
<type>[optional scope][!]: <description>
```

Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`,
`build`, `ci`, `perf`, `revert`.

Examples:

```
feat: add toContainItem operator
fix(assertion): handle nil values in toEqual
docs: add troubleshooting guide
test: add integration tests for tag filtering
```

Use `!` after the type/scope to mark a breaking change:

```
feat!: rename --test-dir to --suite-dir
```

## Pull requests

- Keep changes focused. One feature or fix per PR.
- Update tests. New code without tests will not be merged.
- Update docs. If you change user-facing behavior, update the README, the
  relevant `docs/*.md` page, and `CHANGELOG.md` under `## [Unreleased]`.
- The CI workflow (`.github/workflows/pr.yml`) runs lint, tests, and a
  govulncheck. All must pass.

## Releasing

Releases are cut via the workflow_dispatch `release` workflow:

1. Land all changes on `main`. Make sure `CHANGELOG.md` has a populated
   `## [Unreleased]` section.
2. Run the **Release** workflow from the GitHub Actions tab. Optional `tag`
   input — if omitted, the patch version auto-bumps.
3. The workflow runs tests, creates the tag, and runs GoReleaser, which
   publishes platform binaries (`linux/darwin/windows × amd64/arm64`) and a
   checksums file.
4. After the release lands, move the `## [Unreleased]` entries under a new
   `## [vX.Y.Z]` heading in `CHANGELOG.md` and commit.

GoReleaser config lives in `.goreleaser.yml`.

## Code of conduct

Be respectful, give constructive feedback, assume good intent. Standard stuff.

## Questions?

Open an issue, or start a draft PR with a question — early feedback beats a
finished PR going the wrong direction.
