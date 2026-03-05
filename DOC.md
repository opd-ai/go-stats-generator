# TASK: Audit documentation accuracy across all markdown files and fix discrepancies between documented behavior and actual code.

## Execution Mode
**Autonomous action** — apply safe fixes directly, flag ambiguous issues for review.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
1. Read the project README to understand what the project does, who uses it, and what it promises.
2. Examine `go.mod` for module path and dependencies.
3. Identify all markdown documentation files in the repository.
4. Discover whether the project uses auto-generated docs (look for generator signatures, `go generate` directives, or `godocdown` headers).

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

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections documentation,naming
go-stats-generator diff baseline.json post.json
```
Confirm: doc coverage did not decrease.

## Default Thresholds (calibrate to project)
- Minimum doc coverage: 80%
- All exported symbols should have GoDoc comments
- All markdown code examples should reference existing symbols

## Issue Classification
| Category | Action | Example |
|----------|--------|---------|
| SAFE_AUTO_FIX | Fix directly | Wrong function name in example |
| NEEDS_REVIEW | Flag with comment | Performance claim without benchmark |
| INFO_MISSING | Add documentation | Exported function without GoDoc |

## Markdown Fix Rules
- Do not modify auto-generated files.
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
