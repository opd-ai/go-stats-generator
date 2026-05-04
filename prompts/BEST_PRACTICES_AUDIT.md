# TASK: Perform a best practices audit of a Go project, evaluating adherence to Go community conventions, standard library idioms, effective patterns, and language-specific quality standards while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the best practices audit report
2. **`GAPS.md`** — gaps between current practices and Go community standards

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Conventions
1. Read the project README to understand its purpose, target audience, and maturity level.
2. Examine `go.mod` for module path, Go version, and dependency choices.
3. List packages (`go list ./...`) and understand the project structure.
4. Identify the project's **existing conventions** — these are the baseline:
   - Error handling pattern (wrapping style, sentinel errors, custom types)
   - Naming conventions (abbreviations, acronyms, export decisions)
   - Package organization (flat vs nested, internal packages, cmd structure)
   - Testing approach (table-driven, subtests, mocks, integration tests)
   - Documentation style (GoDoc format, README structure, example functions)
   - Build and CI practices (Makefile, GitHub Actions, linting tools)
5. Note which conventions the project follows consistently — these are its own standards.
6. Identify Go version-specific features available to the project (e.g., generics if Go ≥1.18, `slices` package if Go ≥1.21).

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub — read contributor guidelines, code review comments, and discussions that reveal the project's quality standards.
2. Reference authoritative Go best practices sources: Effective Go, Go Code Review Comments, Go Proverbs, and the Go standard library as exemplars.
3. Check if the project uses linters (golangci-lint, staticcheck) and what rules are enabled.

Keep research brief (≤10 minutes). Record only findings that calibrate expectations for this specific project.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,packages,patterns,interfaces,structs,duplication > tmp/practices-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee tmp/practices-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Best Practices Audit

#### 3a. Package Design
- [ ] Each package has a clear, singular purpose described in its package-level doc comment.
- [ ] Package names are short, lowercase, singular nouns — not `util`, `common`, `misc`, `helpers`, or `base`.
- [ ] `internal/` packages are used to prevent external consumers from depending on unstable implementation details.
- [ ] Circular dependencies between packages do not exist.
- [ ] The `cmd/` directory structure follows Go conventions: `cmd/<binary>/main.go` with minimal code delegating to library packages.
- [ ] Package-level variables are minimized — prefer dependency injection over global state.
- [ ] `init()` functions are used sparingly and only for truly necessary initialization (registering drivers, codecs).

#### 3b. Naming and Style
- [ ] Exported identifiers have GoDoc comments starting with the identifier name: `// FooBar does...`
- [ ] Acronyms are consistently cased: `URL`, `HTTP`, `ID` (not `Url`, `Http`, `Id`) per Go convention.
- [ ] Interface names follow Go conventions: single-method interfaces named with `-er` suffix (`Reader`, `Writer`, `Closer`).
- [ ] Error variable names use `Err` prefix: `var ErrNotFound = errors.New("not found")`.
- [ ] Error type names use `Error` suffix: `type NotFoundError struct{}`.
- [ ] Getter methods do not use `Get` prefix: `obj.Name()` not `obj.GetName()` (Go convention).
- [ ] Boolean variables and fields use positive names: `enabled` not `disabled`, `visible` not `hidden`.
- [ ] Constants use `PascalCase` (exported) or `camelCase` (unexported), not `SCREAMING_SNAKE_CASE`.
- [ ] File names use lowercase with underscores: `my_file.go`, `my_file_test.go`.

#### 3c. Error Handling Idioms
- [ ] Error values are returned, not panicked, for expected failure cases.
- [ ] Error wrapping uses `fmt.Errorf("context: %w", err)` with lowercase, punctuation-free messages.
- [ ] `errors.Is` and `errors.As` are used instead of `==` and type assertions for error matching.
- [ ] Errors are handled exactly once — not logged AND returned (which causes duplicate reporting).
- [ ] `if err != nil { return err }` does not discard context — each level adds information about what it was doing.
- [ ] `defer` is used for cleanup with `Close()` immediately after resource acquisition.
- [ ] `panic` is reserved for programmer errors (impossible states), not for operational errors.

#### 3d. Concurrency Idioms
- [ ] Goroutines have clear ownership and lifecycle management — every goroutine has a way to stop.
- [ ] Channels are used for communication, mutexes for state protection — not the reverse.
- [ ] `context.Context` is the first parameter of functions that do I/O, blocking, or long-running work.
- [ ] `sync.WaitGroup.Add` is called before `go func()`, not inside the goroutine.
- [ ] `select` with `context.Done()` prevents goroutines from blocking forever.
- [ ] Shared mutable state is documented as such, with the protecting mechanism identified.

#### 3e. Interface Design
- [ ] Interfaces are defined in the package that uses them (consumer-side), not the package that implements them — unless the interface is a widely-used contract.
- [ ] Interfaces are small: prefer 1-3 methods. Large interfaces should be composed from smaller ones.
- [ ] Functions accept interfaces, return concrete types — this maximizes flexibility and minimizes indirection.
- [ ] `interface{}` / `any` is avoided where type-specific functions or generics provide type safety.
- [ ] Interface satisfaction is verified at compile time with `var _ Interface = (*Type)(nil)` where appropriate.

#### 3f. Testing Practices
- [ ] Tests use table-driven patterns with `t.Run` subtests for clear, independent test cases.
- [ ] Test helpers call `t.Helper()` so failure locations are reported correctly.
- [ ] Tests avoid `time.Sleep` — use channels, `sync.WaitGroup`, or `testing.T.Deadline()` for synchronization.
- [ ] `testdata/` directories are used for test fixtures, not inline string literals for complex data.
- [ ] Golden files use `testdata/` and are updated with a `-update` flag pattern.
- [ ] `_test.go` files are in the same package for white-box tests or `package_test` for black-box tests, chosen intentionally.
- [ ] Test coverage is meaningful — it tests behavior and edge cases, not just line coverage.

#### 3g. API Design
- [ ] Exported functions return concrete errors, not `bool` for success/failure (errors carry context).
- [ ] Constructor functions (`New*`) return the concrete type (or interface if multiple implementations exist), not `interface{}`.
- [ ] Options pattern (`functional options` or `Option struct`) is used for functions with many optional parameters.
- [ ] `context.Context` does not carry values that should be explicit function parameters (user ID, request ID are acceptable; business logic inputs are not).
- [ ] Zero-value types are useful — `sync.Mutex{}`, `bytes.Buffer{}` are usable without construction.

#### 3h. Module and Dependency Management
- [ ] `go.mod` uses the correct module path matching the repository URL.
- [ ] `go.sum` is committed and not in `.gitignore`.
- [ ] Dependencies are current and not pinned to known-vulnerable versions.
- [ ] `replace` directives in `go.mod` are temporary and documented (not committed for development convenience).
- [ ] Indirect dependencies are not unnecessarily promoted to direct dependencies.

#### 3i. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Respect the project's own conventions**: If the project consistently uses a pattern that differs from "standard" Go, it is the project's convention — do not report every instance. Report it once as a project-wide observation and assess whether it causes real problems.
2. **Check Go version constraints**: Do not recommend generics for a Go 1.17 project or `slices.Contains` for Go <1.21.
3. **Verify it matters**: A naming convention violation in an internal, unexported helper is LOW at most. Focus on exported API surface and critical paths.
4. **Read existing linter configuration**: If the project has a `.golangci.yml` that disables a rule, the team has intentionally opted out. Note it but do not report it as a finding.
5. **Assess consistency over perfection**: A codebase that consistently follows its own pattern is more maintainable than one that inconsistently follows "best" practices. Prioritize consistency findings.
6. **Check for domain-specific exceptions**: Some domains require patterns that look wrong in general Go (e.g., generated code, protocol buffers, CGo wrappers). Do not apply general rules to generated or framework-constrained code.

**Rule**: If a practice deviation does not cause bugs, confuse readers, or hinder maintainability, it is at most LOW severity. Best practices are guidelines, not laws.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# BEST PRACTICES AUDIT — [date]

## Project Conventions Summary
[Summary of the project's own established conventions — this is the primary standard against which deviations are measured]

## Go Version and Feature Availability
[Go version from go.mod, available language features, linter configuration]

## Practices Scorecard
| Category | Adherence | Notes |
|----------|-----------|-------|
| Package Design | ✅/⚠️/❌ | [summary] |
| Naming & Style | ✅/⚠️/❌ | [summary] |
| Error Handling | ✅/⚠️/❌ | [summary] |
| Concurrency | ✅/⚠️/❌ | [summary] |
| Interface Design | ✅/⚠️/❌ | [summary] |
| Testing | ✅/⚠️/❌ | [summary] |
| API Design | ✅/⚠️/❌ | [summary] |
| Module Management | ✅/⚠️/❌ | [summary] |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [practice violation] — [concrete impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is an acceptable deviation in this project] |
```

Generate **`GAPS.md`**:
```markdown
# Best Practices Gaps — [date]

## [Gap Title]
- **Best Practice**: [the Go community standard or convention]
- **Current Practice**: [what the project does instead]
- **Impact**: [how this deviation affects readability, maintainability, or correctness]
- **Recommendation**: [specific changes to align with best practices]
- **Scope**: [how many instances exist and whether it is a project-wide pattern]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Practice violation that causes or will cause bugs: e.g., missing error check on data path, goroutine leak from missing context, or `math/rand` used for security-critical randomness |
| HIGH | Practice violation that significantly hinders maintainability or causes confusion: e.g., exported API without documentation, inconsistent error handling causing lost context, or package design that creates circular dependencies |
| MEDIUM | Practice deviation from Go conventions that affects readability: e.g., non-standard naming in exported API, large interfaces where small ones suffice, or missing table-driven tests |
| LOW | Minor style deviations, documentation formatting issues, or deviations from Go conventions that are consistent within the project |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific change**: State exactly what to rename, restructure, or rewrite. Do not recommend "follow Go conventions."
2. **Respect the project**: If changing a pattern requires touching 100 files, note the scope. If the project's pattern is consistent and functional, a LOW-priority gradual migration is appropriate.
3. **Verifiable**: Include a validation approach (e.g., `golangci-lint run`, `go vet ./...`, `go doc` output).
4. **Prioritized**: Distinguish quick wins (rename a variable) from large refactors (restructure packages).

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence (especially documentation coverage, function complexity, and duplication).
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate the code against **Go community standards AND the project's own conventions** — internal consistency is more important than external conformity.
- Apply the Phase 3i false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: bug-causing practice violations → maintainability-harming patterns → readability issues → style inconsistencies → cosmetic preferences. Within a level, prioritize by impact on the exported API surface (public-facing deviations matter more than internal ones).
