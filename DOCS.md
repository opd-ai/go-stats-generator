# TASK: Fix all undocumented exported symbols and inaccurate GoDoc comments to achieve >=80% documentation coverage.

## Execution Mode
**Autonomous action** — apply all changes without prompting for approval. Skip auto-generated files.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections documentation,naming --min-doc-coverage 0.80
go-stats-generator analyze . --skip-tests --min-doc-coverage 0.80
```

### Phase 2: Fix Documentation
1. From baseline JSON, extract `.documentation` section:
   - List all packages with doc coverage below 80%.
   - List all undocumented exported symbols (functions, types, methods, constants).
2. Process packages in priority order (most exported symbols first):
   - Add missing `doc.go` files for packages without package-level documentation.
   - Add GoDoc comments to undocumented exported symbols.
   - Fix stale GoDoc that references renamed/removed functions or incorrect behavior.
3. GoDoc conventions:
   - First sentence: `// SymbolName ...` (starts with the symbol name).
   - Describe what the function does, not how it does it.
   - Document parameters and return values for complex signatures.
   - Document panic conditions if any.
4. Run `go test -race ./...` after each package to confirm no regressions.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections documentation,naming --min-doc-coverage 0.80
go-stats-generator diff baseline.json post.json
```
Confirm: doc coverage increased, zero regressions.

## Thresholds
| Metric | Target |
|--------|--------|
| Overall doc coverage | >=80% |
| Package-level doc | 100% |
| Exported functions | >=80% |
| Exported types | >=80% |
| Methods | >=80% |

## Priority Rules
| Priority | Criteria |
|----------|----------|
| CRITICAL | Exported symbol with no GoDoc at all |
| HIGH | GoDoc exists but is stale/inaccurate (references wrong behavior) |
| MEDIUM | GoDoc exists but does not start with symbol name |
| LOW | Internal/unexported symbol documentation |

## Skip Rules
- Do not modify auto-generated files (detected by generator signatures).
- Do not modify test files for documentation purposes.
- Do not change function logic — documentation only.

## Output Format
```
[package]: [before]% -> [after]% doc coverage
  Added: [N] GoDoc comments ([list of symbols])
  Fixed: [N] stale GoDoc comments ([list of symbols])
  Created: doc.go (if needed)
Overall: [before]% -> [after]%
```

## Tiebreaker
Process the package with the most undocumented exported symbols first. If tied, lowest current coverage first.
## GoDoc Quality Standards
- First sentence is a complete, grammatical sentence starting with the symbol name.
- Describes behavior, not implementation details.
- Documents error conditions and edge cases for complex functions.
- Uses `// Deprecated:` prefix for deprecated symbols.
- References related symbols with their full path when in different packages.

## Validation Checklist
- [ ] All packages at >=80% doc coverage
- [ ] All packages have doc.go or package-level comment
- [ ] No auto-generated files modified
- [ ] All tests still pass
