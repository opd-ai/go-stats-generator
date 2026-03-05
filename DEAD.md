# TASK: Identify and remove deprecated API usage and dead code (unreachable functions, unused types, orphaned constants/variables).

## Execution Mode
**Autonomous action** — remove dead code, replace deprecated APIs, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Codebase
1. Read the project README to learn its domain, supported Go version, and compatibility guarantees.
2. Examine `go.mod` for the minimum Go version and all direct/indirect dependencies.
3. Identify public API surface — exported functions, types, and interfaces that external consumers may depend on. These cannot be removed even if internally unused.
4. Note whether the project uses build tags, code generation, or reflection that may reference symbols not visible to static analysis.
5. Check for `//go:linkname`, `_ = pkg.Symbol` force-imports, or plugin/RPC patterns that create invisible references.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output pre-dead.json --sections functions,documentation
go vet ./...
```

### Phase 2: Remove Dead Code and Deprecated APIs

**Step 1 — Detect deprecated API usage:**
- Search for `// Deprecated:` comments in the project's own code to find self-deprecated symbols.
- Run `go vet ./...` and `staticcheck ./...` (if available) to surface deprecated stdlib and dependency calls.
- Grep for known deprecated patterns: `ioutil.*` (replaced in Go 1.16), `io/ioutil` imports, `strings.Title` (Go 1.18), `golang.org/x/net/context` (use `context` since Go 1.7), `sort.Sort` on slices (prefer `slices.Sort` in Go 1.21+).
- For each deprecated call, identify the replacement API and the Go version it requires.
- Only replace if the project's `go.mod` minimum version supports the replacement.

**Step 2 — Detect dead code:**
- Identify unexported functions/methods with zero callers (exclude `init()`, test helpers, interface satisfiers).
- Find unexported types, interfaces, and structs with zero references.
- Find unexported constants and variables with zero references.
- Detect unreachable code: branches after unconditional `return`/`panic`/`os.Exit`, always-false conditions.
- Flag unused function parameters (but only remove if safe — interface conformance may require them).

**Step 3 — Classify and prioritize:**
| Priority | Category | Action |
|----------|----------|--------|
| CRITICAL | Deprecated APIs with security implications | Replace immediately |
| HIGH | Deprecated stdlib calls (`ioutil`, etc.) | Replace with modern equivalents |
| HIGH | Entire unexported functions with zero callers | Remove |
| MEDIUM | Unused types/interfaces/constants | Remove |
| LOW | Unused parameters (interface-safe) | Replace with `_` |
| SKIP | Exported symbols (public API) | Never remove — external consumers may depend on them |

**Step 4 — Execute removals:**
- Replace deprecated API calls with their modern equivalents, one package at a time.
- Remove dead functions/types/constants, starting with leaf symbols (no dependents).
- After removing a symbol, re-check its callers — removal may cascade (a helper used only by dead code is now also dead).
- Remove orphaned imports after each file edit.
- Run `go build ./...` after each file to catch compile errors immediately.
- Run `go test -race ./...` after each package is complete.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post-dead.json --sections functions,documentation
go-stats-generator diff pre-dead.json post-dead.json
go vet ./...
go test -race ./...
```
Confirm: all tests pass, no new vet warnings, function count decreased, doc coverage did not drop.

## Default Thresholds (calibrate to project)
| Metric | Target |
|--------|--------|
| Dead unexported functions | 0 |
| Deprecated API calls | 0 |
| Unused unexported types | 0 |
| Orphaned constants/vars | 0 |

## Safety Rules
- **Never remove exported symbols** — they are public API regardless of internal usage.
- **Never remove `init()` functions** — they have implicit callers.
- **Never remove interface-satisfying methods** — even if not called directly, they fulfill a contract.
- **Preserve build-tagged files** — a symbol may appear unused but is active under a different build tag.
- **Preserve `//go:generate` targets** — generated code may reference seemingly-dead symbols.
- **Check reflection usage** — `reflect.TypeOf`, `reflect.ValueOf`, and struct tags can create invisible references.
- **One package at a time** — compile and test after each package before moving to the next.

## Output Format
```
Deprecated APIs replaced: [N]
  [pkg.OldFunc -> pkg.NewFunc] x[count] ([file list])
Dead functions removed: [N] ([total lines removed])
Dead types removed: [N]
Dead constants/variables removed: [N]
Cascaded removals: [N] (symbols that became dead after initial removal)
Tests: PASS
Functions: [before] -> [after]
```

## Tiebreaker
When multiple dead symbols exist at the same priority, remove the one with the most lines first.
