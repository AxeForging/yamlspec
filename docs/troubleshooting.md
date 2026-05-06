# Troubleshooting

Common failure modes and how to read what yamlspec is telling you.

## "spec must contain at least one describe block"

```
Error: discovery failed: parse '.../tests/foo/spec.yaml': spec must contain at least one describe block
```

Your `spec.yaml` parsed cleanly but has no `describe:` entries (or the field
is missing entirely). Add at least one `describe` block with at least one
`it` assertion.

## "describe[0] (\"X\").it[0] (\"Y\") has no assertion operator"

```
describe[0] ("Deployment").it[2] ("be in production namespace") has no assertion operator
(add one of toEqual, toExist, toContain, etc.)
```

The assertion has a `should:` and `expect:` but no operator like `toEqual:`.
This is the most common authoring mistake — easy to forget the operator line.
Add one (or more):

```yaml
- should: "be in production namespace"
  expect: metadata.namespace
  toEqual: "production"   # ← was missing
```

## "field X not found in type domain.Spec"

```
Error: yaml: unmarshal errors:
  line 7: field describes not found in type domain.Spec
```

yamlspec uses **strict YAML decoding** — typos in field names are rejected,
not silently ignored. Fix the typo (`describes` → `describe`, `pre-run` →
`pre_run`, etc.). The valid top-level fields are `name`, `tags`, `pre_run`,
`describe`.

## "selector 'X' matched no resources"

```
Status: FAILED
Error: selector 'select(.kind == "Deployment")' matched no resources
```

The JQ selector ran fine but matched zero documents. Possible causes:

1. **No manifests were discovered** — yamlspec looks for YAML in
   `<spec-dir>/manifests/` first, falling back to YAML files alongside
   `spec.yaml` (excluding `spec.yaml` and `values.yaml` themselves). Confirm
   the files exist and have a `.yaml` or `.yml` extension.

2. **`pre_run` ran but didn't write where you expected** — make sure your
   redirect targets `manifests/`:

   ```yaml
   pre_run:
     - mkdir -p manifests
     - helm template app ../../chart > manifests/rendered.yaml
   ```

3. **The selector itself is wrong** — typo in the kind, namespace mismatch,
   etc. Try a broader selector first (`select: ""` matches everything) and
   narrow from there.

## "pre_run failed: command 'X' timed out after Ys"

```
Status: ERROR
Error: pre_run failed: command 'helm template ...' timed out after 1m0s
```

The default per-command timeout is 60 seconds. Bump it for slow renders:

```bash
yamlspec validate --pre-run-timeout 5m
```

If a command was hanging on a child process (e.g. `sleep` left over from a
test), the timeout now kills the entire process group — no orphaned children.

## "no manifests found"

```
Status: ERROR
Error: no manifests found
```

The spec ran its `pre_run` (if any) but yamlspec found zero YAML files to
test. Check:

- Files have `.yaml` / `.yml` extension (not `.txt`, not extension-less)
- If using `manifests/` subdirectory, the directory exists and contains files
- If your `pre_run` writes to a different path, point yamlspec at the right
  directory by writing into `manifests/`

## "field path not found" vs "field is null"

These are different. yamlspec distinguishes:

| Situation | `toExist: true` | `toExist: false` | `toBeNull: true` | `toBeNull: false` |
|-----------|-----------------|------------------|------------------|-------------------|
| Field absent | FAIL | PASS | FAIL | FAIL |
| Field present, value `null` | FAIL | PASS | PASS | FAIL |
| Field present, non-null value | PASS | FAIL | FAIL | PASS |

If you're checking that a field both exists *and* is non-null, `toExist: true`
is what you want. `toBeNull: true` only passes when the field is explicitly
present but null.

## "--fail-fast cannot be used with --workers > 1"

```
Error: --fail-fast cannot be used with --workers > 1
(parallel workers can't honor ordered short-circuit; pick one)
```

These are mutually exclusive — parallel execution can't honor a
deterministic short-circuit. Pick one:

- `--fail-fast` (sequential, stops on first failure — fast feedback)
- `--workers N` (parallel, runs everything — fast wall time)

## Empty or malformed spec.yaml

```
Error: discovery failed: parse '.../tests/foo/spec.yaml': spec file is empty
```

The file was readable but had no content (or only whitespace). Re-scaffold
with `yamlspec init <name>` if you want a starter template.

## YAML parsing errors in manifests

```
parse '.../manifests/rendered.yaml': yaml: line 42: did not find expected node content
```

Your manifests YAML is malformed. If this came from a `pre_run` rendering
step, run that command manually and inspect the output — `helm template`
sometimes emits empty documents or comment-only blocks that confuse strict
parsers. yamlspec uses `gopkg.in/yaml.v3`, which is stricter than `yaml.v2`.

## Wildcard matched nothing

A wildcard path like `spec.containers[*].image` over an empty array yields
zero values. yamlspec treats this as "no values to assert against" — your
assertion will pass vacuously for value-based operators (`toEqual`,
`toMatch`, etc.) but `toExist: true` will still fail because the path
produced no values.

If you want to assert "there is at least one container," add a separate
length check:

```yaml
- should: "has at least one container"
  expect: spec.containers
  toHaveMinLength: 1
```

## Verbose mode

When in doubt, run with `--verbose` to see debug logs (pre_run commands,
discovered specs, etc.):

```bash
yamlspec --verbose validate --test-dir tests
```

## Still stuck?

Open an issue with:

- Your `spec.yaml`
- A minimal manifest that reproduces the problem
- The exact command you ran
- The full error output

The smaller the repro, the faster the fix.
