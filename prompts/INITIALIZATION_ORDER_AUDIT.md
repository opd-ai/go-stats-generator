# TASK: Perform a focused audit of initialization order and use-before-ready logic errors in Go code, identifying use-before-initialization, incorrect operation ordering, nil map/slice write panics, zero-value assumption violations, circular initialization, and time-of-check to time-of-use (TOCTOU) logic bugs while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the initialization order audit report
2. **`GAPS.md`** — gaps in initialization correctness relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Initialization and Lifecycle Model
1. Read the project README to understand its purpose, startup sequence, configuration loading, and the lifecycle of its main data structures.
2. Examine `go.mod` for module path, Go version, and dependencies that require specific initialization order (database drivers, codec registration, plugin loading).
3. List packages (`go list ./...`) and identify which packages perform global initialization, register handlers, or build data structures that other packages depend on.
4. Build an **initialization inventory** by scanning for:
   - `var` declarations at package level, especially those that reference other package-level variables
   - `init()` functions — each package may have multiple; their order within a package is the source order, but cross-package order depends on import graph
   - `sync.Once` usage — what is initialized lazily, and whether the initialization result is checked
   - Constructor functions (`New*`) that accept pre-requisites as arguments and what happens if those arguments are nil
   - Struct literal initialization: fields not explicitly set receive Go zero values (`0`, `false`, `nil`, `""`) — verify zero values are valid initial states
   - Nil map and nil slice: `var m map[K]V` produces a nil map (read-safe, write-panics); `var s []T` produces a nil slice (safe to append, safe to range)
   - Pointer fields on structs: unset pointer fields are nil — methods called on nil receivers may panic
   - `flag.Parse()` and `viper.ReadInConfig()` — configuration values are not available before these calls
   - Database schema migrations or `db.AutoMigrate()` — tables/columns may not exist before migration runs
5. Identify the project's initialization conventions — does it use constructors? Does it validate struct fields after construction? Does it use option patterns?
6. Map the initialization dependencies: if package A initializes using a value from package B, and package B initializes using a value from package A, there is a circular initialization dependency.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "nil pointer", "panic on startup", "not initialized", "called before setup", "wrong order", "uninitialized", or "zero value" to understand known initialization bugs.
2. Research key dependencies from `go.mod` for required initialization sequences (e.g., must call `db.Open` before any query, must register drivers before `sql.Open`, must call `flag.Parse` before accessing `flag.Args()`).
3. Look up Go initialization order rules (package-level variable initialization order, `init()` execution order) for any patterns relevant to the project.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's initialization model.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > /tmp/init-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -count=1 ./... 2>&1 | tee /tmp/init-test-results.txt
go vet ./... 2>&1 | tee /tmp/init-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Initialization Order Audit

#### 3a. Use Before Initialization
For every value that requires explicit initialization before use, verify the initialization occurs before the first use:

- [ ] Package-level variables initialized with expressions that call functions — Go evaluates these in dependency order within a package, but the order between packages follows the import graph. Verify no cross-package variable initialization creates an order dependency that cannot be satisfied by the import graph (e.g., package A's `var` initialization reads a value from package B, but A does not import B — this fails to compile; or A imports B but B's `var` is not yet fully initialized when A's `init` reads it, which is a runtime ordering issue).
- [ ] `init()` functions that call functions from other packages whose `init()` has not yet run — the Go specification guarantees packages are initialized before the packages that import them, but within a package the order is source order. Verify cross-`init()` dependencies are reflected in import relationships.
- [ ] `sync.Once.Do` usage: the `once.Do(f)` call initializes lazily — any code that uses the initialized value without going through `once.Do` first may see the zero value.
- [ ] Functions that must be called before others are actually called in that order throughout the codebase: e.g., a `Setup()` function must be called before `Process()`; verify there is no code path where `Process()` is called without a preceding `Setup()`.
- [ ] Configuration values read before `flag.Parse()` or `viper.ReadInConfig()` — these return zero/empty values before parsing, and code that uses them before initialization silently uses incorrect configuration.
- [ ] Global registry patterns: `Register(handler)` must be called before `Dispatch(event)`. If `Register` is called in an `init()` function and `Dispatch` is called from `main()`, the ordering is guaranteed. If both can be called from arbitrary code, verify the ordering invariant.

#### 3b. Nil Map and Nil Slice Write Panics
For every map and slice variable, verify it is initialized before use as a write target:

- [ ] `var m map[K]V` followed by `m[key] = value` without an intervening `m = make(map[K]V)` or map literal assignment — writing to a nil map panics.
- [ ] Struct fields of map type not initialized in the constructor or before first write: `type Foo struct { Index map[string]int }; f := Foo{}; f.Index["k"] = 1` panics because `f.Index` is nil.
- [ ] Map fields initialized conditionally: `if condition { m = make(map[K]V) }; m[key] = value` — the write panics when `condition` is false.
- [ ] Nil map returned from a function and used as a write target by the caller without checking for nil.
- [ ] `append` to a nil slice is safe in Go and returns a new slice — but if the caller expects the original variable to be updated, it must use `s = append(s, item)`, not just `append(s, item)` (discarded return).
- [ ] Struct fields of pointer type not initialized before method calls that dereference the pointer: `type Foo struct { Helper *Helper }; f := Foo{}; f.Helper.DoSomething()` panics when `f.Helper` is nil.

#### 3c. Incorrect Operation Ordering
For every sequence of operations that have dependencies, verify the order is correct:

- [ ] Validate-then-use patterns: the validation occurs before the use, not after. A common error is to validate input, transform it, then validate the transformed result — but then use the original unvalidated input.
- [ ] Open-then-write patterns: a file, database, or connection is opened before it is written to, and closed after — not written to before opening or written to after closing.
- [ ] Compute-then-store vs. store-then-compute: a result should be computed first, verified correct, then stored. Code that stores a placeholder and then overwrites it with the correct value may leave the wrong value if the computation fails.
- [ ] Dependency-ordering in multi-step pipelines: step N+1 depends on the output of step N being complete. If these steps run concurrently, verify there is a synchronization point between them; if they run sequentially, verify the function calls are in the correct order.
- [ ] Accumulation before reporting: metrics are fully accumulated before any reporting output is generated. Code that begins writing a report while still populating the data structure it reads from may produce partial or inconsistent output.
- [ ] Sort-then-dedup vs. dedup-then-sort: if deduplication relies on the order produced by sorting, sorting must come first. If sorting is done after dedup, the output is still sorted but the dedup may operate on unsorted data and miss duplicates.

#### 3d. Zero Value Assumption Violations
For every struct type and every function that accepts or returns a struct, verify that the zero value is a valid state:

- [ ] Struct fields that must be non-zero to be useful but have no enforcement at construction time: a `Threshold int` field with zero value of `0` that, if left unset, causes all inputs to pass a threshold check unconditionally.
- [ ] Counter fields initialized to zero that are used in division before being incremented: `average = total / count` where `count` starts at zero.
- [ ] Boolean flags with zero value `false` where the safe default should be `true`: an `Enabled bool` field that defaults to false may silently disable a feature that the user expected to be on.
- [ ] Pointer fields with zero value `nil` that are dereferenced without nil checks in methods: `func (t *T) Process() { t.helper.Do() }` panics when `t.helper` is nil.
- [ ] `time.Time` zero value (`0001-01-01 00:00:00 UTC`) used as a "not set" sentinel — code that checks `t == time.Time{}` or `t.IsZero()` is correct, but code that directly uses the zero `time.Time` in date arithmetic produces nonsensical results.
- [ ] Interface fields with zero value `nil` — calling a method on a nil interface panics. Unlike nil pointer receivers (which can be handled in the method), nil interface values always panic.

#### 3e. Circular and Conflicting Initialization
For every package with `init()` functions and package-level variable initializations, verify there are no circular dependencies:

- [ ] Package A's `init()` calls a function in package B, and package B's `init()` calls a function in package A — this is a circular import and will fail to compile, but indirect circular dependencies through function calls at runtime may not be caught at compile time.
- [ ] Package-level variables that call `init`-like functions: `var cache = NewCache(defaultSize)` where `NewCache` calls `globalRegistry.Register()`, and `globalRegistry` is initialized by another package's `var` that hasn't run yet — the initialization order may not be as expected.
- [ ] Mutual registration patterns: package A registers a handler that calls into package B, and package B registers a handler that calls into package A — if both registrations happen in `init()`, the cross-calls at dispatch time are safe, but the registration order may matter if handlers override each other.
- [ ] `sync.Once` initialization that calls a function which also uses `sync.Once` for the same `Once` variable — this is a deadlock, not a circular import. `sync.Once` is not reentrant; a goroutine calling `once.Do(f)` where `f` calls `once.Do(g)` will deadlock.
- [ ] Database migrations or schema setup called inside `init()` — if the database connection is not yet established at `init()` time, the migration will fail silently or panic.

#### 3f. Time-of-Check to Time-of-Use (TOCTOU) Logic Errors
For every check on a condition that is assumed to remain valid until a subsequent use, verify the assumption holds:

- [ ] `if _, err := os.Stat(path); err == nil { f, err := os.Open(path) }` — the file may be deleted between `Stat` and `Open`; `Open` may fail even after a successful `Stat`. The `Open` error must still be checked.
- [ ] `if len(s) > 0 { v := s[0] }` where `s` can be modified by another goroutine between the check and the access — this is a race condition, but in sequential code it is always safe. Flag only in concurrent contexts.
- [ ] Cache hit checks: `if val, ok := cache[key]; ok { return val }` — between the cache check and the use, another goroutine may invalidate the cache entry. In sequential code this is safe; in concurrent code it requires synchronization.
- [ ] Configuration loaded once at startup and assumed to be valid throughout — if the configuration source is mutated after loading (e.g., environment variables changed, config file rewritten), the loaded values become stale. Verify the project either reloads configuration or documents that it is snapshot-at-startup.
- [ ] `if db != nil { db.Query(...) }` — the nil check does not prevent a concurrent goroutine from setting `db = nil` between the check and the query. In sequential code this is safe; flag only in concurrent contexts.
- [ ] Computed index or position stored in a variable and later used as an index after the underlying collection may have been modified: `idx := findIndex(items, target); items = append(items, newItem); use(items[idx])` — if `append` reallocates, `idx` is still valid (it's a position, not a pointer), but if `items` is otherwise reordered between `findIndex` and `use`, `idx` is stale.

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the uninitialized path is reachable**: Confirm the code path where the nil map write, zero value use, or use-before-init actually executes. Trace the initialization sequence from `main()` or from test setup.
2. **Check for lazy initialization patterns**: A nil map or pointer that is initialized on first use (checked before each write with `if m == nil { m = make(...) }`) is not a bug — it is lazy initialization. Verify whether this pattern is used consistently.
3. **Verify the zero value is actually invalid**: Go's zero values are intentionally useful. A `false` boolean, `0` integer, or `nil` slice is often a valid initial state. Verify that the specific zero value actually causes incorrect behavior, not just that it looks uninitialized.
4. **Check for constructor enforcement**: If the only way to create a type is through a `New*` function that initializes all fields, zero value use is not a bug (users cannot create the struct directly). Verify the `New*` function is the only constructor.
5. **Read surrounding comments**: If a comment explicitly acknowledges an initialization decision (e.g., `// initialized lazily on first use`, `// zero value is valid: means "unlimited"`, `//nolint:`, or a TODO), treat it as an acknowledged pattern — do not report it as a new finding.
6. **Assess sequentiality**: TOCTOU logic errors require the ability for state to change between the check and the use. In purely sequential, single-goroutine code, the state cannot change. Flag TOCTOU only when the gap is exposed to concurrent modification or external mutation.

**Rule**: If you cannot trace a concrete execution path where the initialization error causes a panic, wrong result, or incorrect behavior, do NOT report it. Initialization findings require a specific call sequence or state that triggers the bug.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# INITIALIZATION ORDER AUDIT — [date]

## Project Initialization Profile
[Summary: initialization sequence (package init, constructor functions, configuration loading, database setup), lifecycle model, zero value safety, lazy initialization patterns]

## Initialization Inventory
| Package | init() Functions | Package-Level Vars | Nil Map Writes | Zero-Value Structs | sync.Once | Constructor Enforcement |
|---------|-----------------|-------------------|---------------|-------------------|-----------|------------------------|
| [pkg]   | N               | N                 | N             | ✅/❌              | N         | ✅/❌                   |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: execution path where uninitialized state is reached] — [impact: panic or wrong result] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why the initialization is actually safe or the zero value is valid in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Initialization Order Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about correctness, startup reliability, or configuration handling]
- **Current State**: [what initialization logic exists, and where it may be incomplete]
- **Risk**: [what execution sequence causes a panic or wrong result: nil write, use-before-init, wrong order]
- **Closing the Gap**: [specific constructor change, guard clause, init order fix, or zero-value handling needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Nil map write panic on a reachable code path; use of uninitialized state causing a panic; operation ordering that corrupts data on every execution (e.g., reporting before accumulation is complete) |
| HIGH | Zero value assumption violation causing systematically wrong results (e.g., division by zero counter, wrong boolean default disabling a feature); configuration used before parsing producing empty/wrong values |
| MEDIUM | Initialization that works in practice but relies on undocumented ordering assumptions; lazy initialization pattern that is inconsistent (sometimes initialized, sometimes not); TOCTOU logic error in concurrent code |
| LOW | Zero value that is technically valid but fragile (easy to forget to set); undocumented initialization requirements; `sync.Once` used without error return when initialization can fail |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly where to add an initialization call, nil check, constructor enforcement, or operation reordering. Do not recommend "ensure initialization before use" — show exactly what code to add and where.
2. **Include the triggering sequence**: Identify the specific call sequence or state transition that triggers the bug.
3. **Respect project idioms**: If the project uses constructors consistently, recommend fixing the constructor. If it uses `sync.Once`, recommend `sync.Once`. Do not introduce new initialization patterns.
4. **Verifiable**: Include a test case that exercises the uninitialized path and demonstrates the bug before the fix and correct behavior after.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for struct complexity and constructor usage patterns.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete execution path demonstrating the initialization error — no speculative findings.
- Evaluate the code against its **own initialization conventions** and lifecycle model, not arbitrary external standards.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: panics from nil map writes or nil pointer dereferences on common code paths → wrong results from zero-value fields used in computation before being set → initialization ordering bugs that manifest only in specific startup sequences → TOCTOU logic errors in concurrent code → fragile initialization that works currently but breaks under refactoring. Within a level, prioritize by how early in the program lifecycle the bug manifests and how visible the failure is.
