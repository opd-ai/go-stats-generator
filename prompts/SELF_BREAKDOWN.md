# TASK: Use `go-stats-generator` to analyze and refactor the target codebase (dogfooding) — reduce complexity of functions below thresholds while verifying the build still works.

## Execution Mode
**Autonomous action** — refactor and fix discovered bugs, validate with tests, rebuild, and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Codebase
1. Read the project README to understand purpose, domain, and build process.
2. Examine `go.mod` and discover the build/install command (e.g., `go build`, `go install`, `make build`).
3. Identify the project's coding patterns, naming conventions, and error handling style.
4. Note whether the project is a tool/binary (requires rebuild verification) or a library (build verification sufficient).

### Phase 1: Self-Analysis Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator analyze . --skip-tests --max-complexity 10 --max-function-length 30
```

### Phase 2: Refactor
1. Identify functions exceeding thresholds from the baseline.
2. For each violating function (sorted by overall complexity descending):
   - **Understand its role** before refactoring — read callers and context.
   - Apply extract-method refactoring matching the project's idioms:
     - Extract cohesive blocks into named helpers (<20 lines, cyclomatic <8).
     - Preserve all public API signatures.
   - If a bug is discovered during review, fix it as part of this pass.
3. Run `go test -race ./...` after each refactoring.
4. Run `go vet ./...` to confirm no issues.
5. If the project is a buildable tool, rebuild and verify it still works:
   ```bash
   go build ./... && echo "BUILD PASS"
   ```

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator diff baseline.json post.json
```
Confirm: all target functions below thresholds, zero regressions, project still builds and runs.

## Default Thresholds (calibrate to project)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 10.0 |
| Cyclomatic complexity | 10 |
| Function length | 30 lines |
| Nesting depth | 3 |
| Extracted function length | 20 |
| Extracted function cyclomatic | 8 |

## Dogfooding Rules
- After refactoring, rebuild the project and verify it still produces correct output.
- If the project's own output changes (beyond metric improvements), investigate whether a bug was introduced.
- May fix bugs discovered during analysis (unlike BREAKDOWN.md which is refactor-only).

## Output Format
```
Analysis: [N] functions above thresholds
Refactored:
  [function] [file]: [old] -> [new] ([reduction]%)
Bugs fixed: [count] (or "none")
Build verification: PASS
Tests: PASS
```

## Tiebreaker
Refactor the longest function first when complexity scores are tied.
