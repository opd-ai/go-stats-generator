# TASK: Fix all undocumented exported symbols, achieve >=80% GoDoc coverage, and audit all markdown documents for accuracy and completeness.

## Execution Mode
**Autonomous action** — apply all changes without prompting for approval. Skip auto-generated files.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
1. Read the project README to understand its purpose, domain, and API surface.
2. Examine `go.mod` for module path.
3. List packages (`go list ./...`) and identify the public API packages vs. internal packages.
4. Discover the project's existing GoDoc style: how existing comments are phrased, whether they document parameters, and whether they reference related symbols.
5. Inventory all markdown files in the repository (`find . -name '*.md' -not -path './.git/*'`).

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections documentation,naming --min-doc-coverage 0.80
go-stats-generator analyze . --skip-tests --min-doc-coverage 0.80
```

### Phase 2: Fix GoDoc Documentation
1. From baseline JSON, extract `.documentation` section:
   - List all packages with doc coverage below target.
   - List all undocumented exported symbols.
2. Process packages in priority order (most exported symbols first):
   - Add missing `doc.go` files for packages without package-level documentation.
   - Add GoDoc comments to undocumented exported symbols.
   - Fix stale GoDoc that references renamed/removed functions or incorrect behavior.
3. GoDoc conventions (match the project's existing style):
   - First sentence: `// SymbolName ...` (starts with the symbol name).
   - Describe what the function does, not how it does it.
   - Document parameters and return values for complex signatures.
   - Document panic conditions if any.
4. Run `go test -race ./...` after each package.

### Phase 3: Audit Markdown Documents
Audit every markdown file in the repository for accuracy and completeness.

#### Markdown File Categories

| Category | Files | Action |
|----------|-------|--------|
| Project documentation | `README.md`, `examples/README.md`, `SINGLE_FILE_ANALYSIS.md` | Verify feature claims, CLI examples, and API usage match current implementation |
| Technical guides | `docs/ci-cd-integration.md`, `docs/postgres-backend.md`, `docs/mongodb-backend.md`, `docs/GITHUB_PAGES_SETUP.md` | Verify configuration instructions, code examples, and referenced endpoints exist |
| Architecture & performance | `docs/LLM_SLOP_PREVENTION.md`, `docs/PERFORMANCE.md`, `docs/performance-validation.md` | Verify architectural claims and benchmark data reflect current code |
| Task prompts | All uppercase `.md` files in the root directory (`EXECUTE.md`, `RESPOND.md`, `REVIEW.md`, `MAKE_PLAN.md`, `BREAKDOWN.md`, `BUGS.md`, `TESTS.md`, `FAIL.md`, `DOC.md`, and others — 25 files total) | Verify referenced CLI commands and flags exist, prerequisite install commands are correct, and workflow steps are valid |
| Reports & assessments | `FUNCTIONAL_AUDIT.md`, `AUDIT_RESOLUTION.md`, `ENHANCEMENT_SUMMARY.md`, `GAPS.md`, `PLAN.md`, `ROADMAP.md`, `PRODUCTION_READINESS_2026-03-07.md`, `MIGRATION.md`, `TEMPLATE_FIX.md` | Verify referenced files and symbols still exist; flag stale claims |
| Package audits | `*/AUDIT.md` files in `cmd/`, `pkg/`, `internal/`, `testdata/` | Verify audit findings reference current code; flag resolved items still marked open |
| Templates | `internal/reporter/templates/markdown/report.md`, `internal/reporter/templates/markdown/diff.md` | Verify template variables match current report generation code |
| Copilot config | `.github/copilot-instructions.md` | Verify project description and technical stack match current implementation |

#### Audit Steps
1. For each markdown file:
   - Verify that referenced function names, type names, and CLI flags still exist in the codebase.
   - Check that code examples (```go, ```bash blocks) use valid syntax and reference real symbols.
   - Confirm version numbers, dependency names, and feature claims match `go.mod` and current code.
   - Check internal cross-references between markdown files (e.g., "see ROADMAP.md") resolve to existing files.
2. Classify each issue found:
   - **SAFE_FIX**: Wrong function name, outdated flag, broken cross-reference — fix directly.
   - **STALE**: Content references removed code or completed work — update or remove.
   - **NEEDS_REVIEW**: Behavioral or performance claim that cannot be verified from code alone — flag with `<!-- REVIEW: ... -->`.
3. Apply SAFE_FIX and STALE corrections directly. Mark NEEDS_REVIEW items with inline comments.

### Phase 4: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections documentation,naming --min-doc-coverage 0.80
go-stats-generator diff baseline.json post.json
```
Confirm: doc coverage increased, zero regressions, no broken cross-references.

## Default Thresholds (calibrate to project)
| Metric | Target |
|--------|--------|
| Overall GoDoc coverage | >=80% |
| Package-level doc | 100% |
| Exported functions | >=80% |
| Exported types | >=80% |
| Markdown code examples | All reference existing symbols |
| Markdown cross-references | All resolve to existing files |

## Priority Rules
| Priority | Criteria |
|----------|----------|
| CRITICAL | Exported symbol with no GoDoc at all |
| CRITICAL | Markdown references a function or type that no longer exists |
| HIGH | GoDoc exists but is stale/inaccurate |
| HIGH | Markdown code example uses removed CLI flags or wrong syntax |
| MEDIUM | GoDoc exists but does not start with symbol name |
| MEDIUM | Markdown cross-reference points to renamed or moved file |
| LOW | Internal/unexported symbol documentation |
| LOW | Minor markdown formatting or style inconsistencies |

## Skip Rules
- Do not modify auto-generated files.
- Do not modify test files for documentation purposes.
- Do not change function logic — documentation only.
- Do not modify markdown template files (`internal/reporter/templates/`) unless template variables reference non-existent report fields or contain syntax errors.
- Do not delete markdown files — fix or flag them.

## Output Format
```
=== GoDoc Coverage ===
[package]: [before]% -> [after]% doc coverage
  Added: [N] GoDoc comments ([list of symbols])
  Fixed: [N] stale GoDoc comments ([list of symbols])
  Created: doc.go (if needed)
Overall: [before]% -> [after]%

=== Markdown Audit ===
[SAFE_FIX] [file:line] — [what was wrong] -> [what was fixed]
[STALE] [file:line] — [description] -> [action taken]
[NEEDS_REVIEW] [file:line] — [claim] — [reason for uncertainty]
Markdown files audited: [N]
Issues fixed: [N] | Flagged for review: [N]
```

## Tiebreaker
Process the package with the most undocumented exported symbols first. If tied, lowest current coverage first. For markdown audit, process project documentation and technical guides before task prompts and reports.
