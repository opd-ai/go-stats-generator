# TASK: Fix all undocumented exported symbols, achieve >=80% GoDoc coverage, and audit all markdown documents for accuracy and completeness.

## Execution Mode
**Autonomous action** — apply all changes without prompting for approval. Skip auto-generated files.
This prompt operates on a **third-party Go project**, not on go-stats-generator itself. Every decision must serve the target project's own stated goals and conventions.

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
6. Discover whether the project uses auto-generated docs (look for generator signatures, `go generate` directives, or `godocdown` headers).

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
Audit every markdown file discovered in Phase 0 step 5 for accuracy and completeness.

#### Discovery and Categorization
Classify each discovered markdown file by its role in the project:

| Category | How to Identify | Action |
|----------|----------------|--------|
| Project README | Root `README.md` and any `README.md` in subdirectories | Verify feature claims, CLI examples, and API usage match current implementation |
| Technical/user guides | Files in `docs/`, `doc/`, or similar documentation directories | Verify configuration instructions, code examples, and referenced endpoints exist |
| API/usage examples | Files in `examples/`, or markdown with code blocks demonstrating API usage | Verify code examples compile and reference existing symbols |
| Changelog/release notes | `CHANGELOG.md`, `RELEASES.md`, or similar | Verify most recent entries reference existing code and versions |
| Architecture/design docs | ADRs, design documents, specification files | Verify architectural claims reflect current code structure |
| Contributing guides | `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, or similar | Verify build/test instructions work and match actual project setup |
| Any other markdown | All remaining `.md` files | Verify referenced symbols, files, and claims are current |

#### Audit Steps
1. For each markdown file:
   - Verify that referenced function names, type names, and CLI flags still exist in the codebase.
   - Check that code examples use valid syntax and reference real symbols.
   - Confirm version numbers, dependency names, and feature claims match `go.mod` and current code.
   - Check internal cross-references between markdown files resolve to existing files.
   - Verify CLI usage examples against `--help` output where applicable.
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
- Do not modify markdown template files unless template variables are provably incorrect.
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
Process the package with the most undocumented exported symbols first. If tied, lowest current coverage first. For markdown audit, process project README and user-facing documentation before internal or auxiliary files.
