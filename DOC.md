# TASK: Audit documentation accuracy across all markdown files and fix discrepancies between documented behavior and actual code.

## Execution Mode
**Autonomous action** — apply safe fixes directly, flag ambiguous issues for review.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections documentation,naming
```

### Phase 2: Audit and Fix
1. Extract documentation metrics from baseline JSON:
   - Overall doc coverage percentage
   - Per-package doc coverage
   - Undocumented exported symbols list
2. For each markdown file in the repository:
   - Verify code examples still compile and produce described output.
   - Check that referenced functions, types, and CLI flags exist in the codebase.
   - Confirm version numbers and feature claims match current implementation.
3. Classify each issue:
   - **SAFE_AUTO_FIX**: Typos, outdated function names, wrong flag names — fix directly.
   - **NEEDS_REVIEW**: Behavioral claims that may be intentionally aspirational — flag with `<!-- REVIEW: ... -->`.
   - **INFO_MISSING**: Documented feature exists but has no usage example — add example.
4. Apply SAFE_AUTO_FIX corrections directly. Mark NEEDS_REVIEW items with comments.
5. Run `go test -race ./...` to confirm no regressions (documentation changes should not affect tests).

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections documentation,naming
go-stats-generator diff baseline.json post.json
```
Confirm: doc coverage did not decrease.

## Thresholds
- Minimum doc coverage: 80%
- All exported symbols must have GoDoc comments
- All markdown code examples must reference existing symbols

## Issue Classification
| Category | Action | Example |
|----------|--------|---------|
| SAFE_AUTO_FIX | Fix directly | Wrong function name in example |
| NEEDS_REVIEW | Flag with comment | "Supports 50k+ files" claim without benchmark |
| INFO_MISSING | Add documentation | Exported function without GoDoc |

## Doc Coverage Targets
| Level | Target |
|-------|--------|
| Package-level doc | 100% (every package has a doc.go or package comment) |
| Exported functions | >=80% coverage |
| Exported types | >=80% coverage |
| Methods | >=80% coverage |

## Markdown Fix Rules
- Do not modify auto-generated files (detected by generator signatures like `godocdown` headers).
- Preserve existing accurate content — only fix what is wrong or missing.
- When adding GoDoc, follow Go convention: first sentence starts with the symbol name.
- For CLI flag documentation, verify against `--help` output.

## Output Format
```
[SAFE_AUTO_FIX] [file:line] — [what was wrong] -> [what was fixed]
[NEEDS_REVIEW] [file:line] — [claim] — [reason for uncertainty]
[INFO_MISSING] [symbol] — added GoDoc comment
Doc coverage: [before]% -> [after]%
```

## Tiebreaker
Fix the file with the lowest doc coverage first.
## Validation Checklist
- [ ] Doc coverage did not decrease
- [ ] All SAFE_AUTO_FIX changes are correct
- [ ] All NEEDS_REVIEW items are flagged with comments
- [ ] No auto-generated files were modified
- [ ] All tests still pass
