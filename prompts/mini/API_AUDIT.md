# TASK: Perform a focused API design and contract audit of a Go project, evaluating exported API surface quality, contract clarity, backward compatibility, documentation completeness, and consumer ergonomics while rigorously preventing false positives.

## Output
Write two files in repo root: `AUDIT.md` (API audit report) and `GAPS.md` (API design gaps).

## Workflow

### Phase 0: Map Exported API Surface
1. Read README for purpose, target consumers, API guarantees.
2. Check `go.mod` for module path, version, dependencies.
3. Classify packages: public API (not `internal/`), internal, `cmd/`.
4. Catalog: exported types, functions/methods, error types, config types.
5. Identify API maturity: v0 (unstable), v1+ (stable), library vs app.
6. Check docs: GoDoc comments, README examples, `_test.go` examples.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,interfaces,structs,packages > tmp/api-metrics.json
go doc ./...
```

### Phase 2: Audit Checklist
- [ ] **Surface minimality**: Only consumer-needed types exported; implementation details in `internal/`; no accidental exports
- [ ] **Signatures**: `context.Context` first, required inputs, then options; returns `(T, error)` not `(error, T)`; minimal params (use config structs)
- [ ] **Type design**: Interfaces for behavior not data; opaque types for internals; error types implement `error`; options pattern for config
- [ ] **Documentation**: All exports have GoDoc; examples for primary entry points; error values documented
- [ ] **Backward compat**: No breaking changes without version bump; deprecation comments; aliases for renamed symbols
- [ ] **Error contract**: Sentinel errors for expected failures; wrap with `%w`; `errors.Is`/`As` works; no unexported error types in returns
- [ ] **Ergonomics**: Sensible defaults; builder/options for complex config; `io.Reader`/`Writer` for streams; clear naming

### Phase 3: Report
**AUDIT.md**: API surface summary, quality scorecard, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, consumer impact, remediation.

**GAPS.md**: Per gap — API Element, Issue, Consumer Impact, Recommendation.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete API issue, actionable fix
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Evaluate against project's own contract and consumers
- Prefer backward-compatible additions over modifications