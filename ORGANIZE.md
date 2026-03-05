# TASK: Reorganize codebase file structure using package cohesion/coupling metrics — move code between files without modifying logic.

## Execution Mode
**Autonomous action** — reorganize one sub-package at a time, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections packages,structs,interfaces,functions
```

### Phase 2: Identify and Reorganize
1. From baseline JSON, identify packages with poor organization:
   - Cohesion <0.3: CRITICAL — package serves too many unrelated purposes
   - Cohesion <0.5: HIGH — package would benefit from splitting
   - Cohesion <0.7: MEDIUM — minor organizational improvements possible
   - Coupling >0.7: package has too many external dependencies
2. Select the worst-scoring package (lowest cohesion first).
3. Analyze the package structure:
   - What types/functions are defined in each file?
   - Which symbols are related (share types, call each other)?
   - Are there files with mixed concerns?
4. Apply reorganization moves:
   - **Group by concern**: move related functions/types into the same file.
   - **Split large files**: break files with >500 lines or >20 functions into focused files.
   - **Consolidate tiny files**: merge files with <50 lines into related files.
   - **Fix naming**: rename files to match their primary concern (e.g., `helpers.go` → `validation.go`).
5. Rules for moving code:
   - Only move entire top-level declarations (functions, types, constants).
   - Do NOT modify function bodies, signatures, or logic.
   - Update imports as needed after moves.
   - Maintain `doc.go` as the package documentation file.
6. Run `go build ./...` and `go test -race ./...` after each package reorganization.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections packages,structs,interfaces,functions
go-stats-generator diff baseline.json post.json
```
Confirm: cohesion improved, zero test regressions, no logic changes.

## Thresholds
| Metric | Critical | Warning | Target |
|--------|----------|---------|--------|
| Package cohesion | <0.3 | <0.5 | >=0.5 |
| Package coupling | >0.7 | >0.5 | <0.5 |
| File size | >500 lines | >300 lines | <300 lines |
| Functions per file | >20 | >15 | <=15 |

## Move Rules
- Only move code — never modify logic, add features, or fix bugs.
- Each move must improve at least one metric (cohesion, coupling, file size).
- Preserve all existing public API signatures and import paths.
- After reorganization, the package must compile and all tests must pass.
- Document each move: `[symbol] [from_file] -> [to_file] — [reason]`

## Output Format
```
Package: [name] (cohesion: [before] -> [after])
Moves:
  [Type/Func] [old_file] -> [new_file]
  ...
Tests: PASS
```

## Tiebreaker
Reorganize the package with the lowest cohesion score first. If tied, highest coupling. If still tied, most files.
## File Organization Patterns
| Pattern | When to Apply |
|---------|---------------|
| One type per file | When a file has multiple unrelated types |
| Group by lifecycle | When init/New/Close are scattered across files |
| Separate interface from implementation | When interface + all impls are in one large file |
| Consolidate constants | When constants are scattered across many files |

## Validation Checklist
- [ ] Package cohesion improved (metrics confirmed via diff)
- [ ] All tests pass with -race flag
- [ ] No logic changes — only code movement
- [ ] All imports updated correctly
- [ ] `go build ./...` succeeds
