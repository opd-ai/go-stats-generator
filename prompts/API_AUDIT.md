# TASK: Perform a focused API design and contract audit of a Go project, evaluating exported API surface quality, contract clarity, backward compatibility, documentation completeness, and consumer ergonomics while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the API audit report
2. **`GAPS.md`** — gaps in API design, documentation, and contract clarity

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Map the Exported API Surface
1. Read the project README to understand its purpose, target consumers (library users, CLI users, API clients), and stated API guarantees.
2. Examine `go.mod` for module path, Go version, and dependency profile.
3. List packages (`go list ./...`) and classify:
   - **Public API packages**: Intended for external consumers (not under `internal/`)
   - **Internal packages**: Implementation details not part of the API contract
   - **Command packages**: CLI entry points under `cmd/`
4. For each public API package, catalog:
   - Exported types (structs, interfaces, type aliases, constants, variables)
   - Exported functions and methods with their full signatures
   - Exported error types and sentinel errors
   - Configuration types and option patterns
5. Identify the project's API evolution stage: v0 (unstable), v1+ (stability expected), or library vs application.
6. Check for API documentation: GoDoc comments, README examples, API reference docs, example functions in `_test.go` files.

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub — read issues about API confusion, breaking changes, or usage questions that indicate API design problems.
2. Check if the project follows semantic versioning and whether past releases introduced breaking changes.
3. Look at how consumers actually use the API (search GitHub for import paths).
4. Review Go API design guidance from the standard library, `golang.org/x` packages, and established community patterns.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to API quality.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,interfaces,structs,packages > tmp/api-audit-metrics.json
go-stats-generator analyze . --skip-tests
go doc ./... 2>&1 | head -500 | tee tmp/api-doc-output.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: API Design Audit

#### 3a. API Surface Minimality
- [ ] The exported API surface is minimal — only types, functions, and methods that consumers need are exported.
- [ ] Implementation details are not leaked through exported types (internal state, helper types, intermediate representations).
- [ ] `internal/` packages are used to enforce encapsulation of non-API code.
- [ ] Unexported fields in exported structs do not create confusion (consumers cannot construct the struct with all fields).
- [ ] Exported constants and variables are intentional API, not accidental exports of implementation details.

#### 3b. Function and Method Signatures
- [ ] Function parameters are in a logical order: `context.Context` first, then required inputs, then optional configuration.
- [ ] Return types follow Go conventions: `(result, error)` for fallible operations, concrete types over interfaces for returns.
- [ ] Functions with more than 5 parameters use an options struct or functional options pattern.
- [ ] `context.Context` is accepted (not stored in structs) by functions that perform I/O, blocking, or long-running work.
- [ ] Variadic parameters are used appropriately and documented clearly.
- [ ] Method receivers are consistent within a type: all pointer receivers or all value receivers (with documented exceptions).

#### 3c. Type Design
- [ ] Struct zero-values are useful — a newly declared `var x MyType` behaves sensibly without calling a constructor.
- [ ] Constructor functions (`New*`) are provided when zero-values are not useful and document required initialization.
- [ ] Interface types are small (1-3 methods) and defined by the consumer, not the implementer, unless they represent a widely-used contract.
- [ ] Type aliases (`type X = Y`) are not confused with type definitions (`type X Y`) — aliases share identity, definitions create new types.
- [ ] Enums use `iota` with a named type and a `String()` method for debuggability.
- [ ] Sentinel errors are exported as `var Err* = errors.New(...)` and documented as part of the API contract.

#### 3d. Documentation Quality
- [ ] Every exported type, function, method, and constant has a GoDoc comment starting with the identifier name.
- [ ] Package-level comments (`// Package foo ...`) exist for all public API packages.
- [ ] Example functions (`func ExampleFoo()`) exist for non-trivial API entry points.
- [ ] Error conditions are documented: which errors can a function return and under what circumstances?
- [ ] Concurrency safety is documented: is the type safe for concurrent use? If so, what synchronization guarantees does it provide?
- [ ] Parameter constraints are documented: valid ranges, nil behavior, empty string behavior.
- [ ] Deprecation notices use `// Deprecated:` comment convention with migration guidance.

#### 3e. Backward Compatibility
- [ ] Exported function signatures have not changed in breaking ways between tagged releases (removed parameters, changed return types).
- [ ] Exported struct fields have not been removed or retyped.
- [ ] Exported interfaces have not added methods (which breaks all existing implementations).
- [ ] Sentinel errors have not been renamed or removed.
- [ ] Configuration options have not changed default behavior silently.
- [ ] If breaking changes were necessary, they are documented in CHANGELOG and use a major version bump.

#### 3f. Error Contract
- [ ] Functions that return errors document which error types or values the caller should check for.
- [ ] Error wrapping preserves the ability to use `errors.Is` and `errors.As` for documented error types.
- [ ] Error messages provide sufficient context for the caller to diagnose and handle the error.
- [ ] Nil errors are never returned alongside invalid results — the success path is unambiguous.
- [ ] Error types that consumers need to inspect are exported; implementation-detail errors are wrapped opaquely.

#### 3g. Consumer Ergonomics
- [ ] Common use cases require minimal code — the API makes the simple case simple and the complex case possible.
- [ ] Default behavior is safe and useful — consumers should not need to configure 10 options for basic usage.
- [ ] Related functionality is grouped in a single package, not scattered across multiple packages that must be imported together.
- [ ] Type conversion between the project's types and standard library types is straightforward.
- [ ] The API does not force consumers to handle internal concerns (e.g., manual resource management that could be automatic).

#### 3h. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the API is intended for external consumption**: Code under `internal/` or in unexported functions is not part of the API contract. Do not apply API design standards to internal code.
2. **Check the project's maturity**: A v0 project is expected to have an evolving API. Do not report API instability as a finding for pre-v1 projects — note it as context.
3. **Respect the project's design choices**: If a design decision is documented and intentional (e.g., large interfaces for a specific pattern like `io.ReadWriteCloser`), it is not a finding.
4. **Check for generated code**: Generated API surfaces (protobuf, OpenAPI) follow their generator's conventions, not hand-written Go conventions.
5. **Assess consumer impact**: An undocumented function used only internally is LOW priority. An undocumented function in the main consumer-facing package is HIGH.

**Rule**: Evaluate the API against its own stated contract and the expectations of its target consumers, not against a platonic ideal of API design.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# API AUDIT — [date]

## API Surface Summary
[Public packages, exported symbol count, API maturity level, target consumers]

## API Quality Scorecard
| Category | Rating | Notes |
|----------|--------|-------|
| Surface Minimality | ✅/⚠️/❌ | [summary] |
| Signature Design | ✅/⚠️/❌ | [summary] |
| Type Design | ✅/⚠️/❌ | [summary] |
| Documentation | ✅/⚠️/❌ | [N% documented] |
| Backward Compatibility | ✅/⚠️/❌ | [summary] |
| Error Contract | ✅/⚠️/❌ | [summary] |
| Consumer Ergonomics | ✅/⚠️/❌ | [summary] |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [API issue] — [consumer impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not an API design issue in this context] |
```

Generate **`GAPS.md`**:
```markdown
# API Design Gaps — [date]

## [Gap Title]
- **API Element**: [function/type/package affected]
- **Issue**: [what is missing or problematic]
- **Consumer Impact**: [how this affects API users]
- **Recommendation**: [specific API change or documentation addition]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Exported API returns incorrect results, has undocumented breaking behavior, or exposes internal state that creates security/correctness risks |
| HIGH | Missing documentation on a primary API entry point, breaking change without version bump, or error contract that prevents proper error handling |
| MEDIUM | Inconsistent naming in exported API, missing example functions, or parameter validation that returns confusing errors |
| LOW | Minor documentation gaps on rarely-used exports, style preferences, or internal-only API surface issues |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific change**: State exactly what to rename, document, add, or remove in the API. Do not recommend "improve the API."
2. **Backward compatible**: Wherever possible, recommend additions (new functions, aliases) over modifications to existing API.
3. **Verifiable**: Include `go doc` output or test case that demonstrates the improvement.
4. **Consumer-focused**: Frame remediation in terms of how it improves the consumer experience.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` documentation and interface metrics as primary evidence.
- Every finding must reference a specific exported symbol and file location.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate the API against its **own stated contract and target consumers**, not arbitrary design standards.
- Apply the Phase 3h false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: correctness issues in exported API → missing documentation on primary entry points → breaking changes → ergonomic issues → naming conventions → internal API issues. Within a level, prioritize by consumer usage frequency.
