# TASK: Convert Un-Optimized LLM Prompts to Use go-stats-generator

## EXECUTION MODE
**Autonomous action** — Rewrite each target file in-place. No user approval between files.

## OBJECTIVE
This repository contains LLM prompt files stored as ALL_CAPS markdown files. Two exemplar prompts — `BREAKDOWN.md` (function complexity refactoring) and `DEDUPLICATE.md` (duplication elimination) — have been optimized to use `go-stats-generator` for data-driven, measurable guidance. Convert all remaining eligible prompt files to follow the same pattern: use `go-stats-generator` analysis as the primary evidence source, define concrete metrics and thresholds, and validate outcomes with differential analysis.

## CONSTRAINT
The tool is called `go-stats-generator` — never `gostats`. Correct any occurrences.

## PREREQUISITES
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
which jq || sudo apt-get install -y jq
```

## FILE CLASSIFICATION

### Exemplars (DO NOT MODIFY — use as structural templates)
- `BREAKDOWN.md` — Function complexity breakdown with go-stats-generator
- `DEDUPLICATE.md` — Duplication elimination with go-stats-generator

### Already Optimized (REVIEW ONLY — update if structure diverges from exemplars)
- `SELF_BREAKDOWN.md` — Self-analysis variant of BREAKDOWN
- `LOOP_BREAKDOWN.md` — Iterative wrapper around BREAKDOWN
- `PERFORM_AUDIT.md` — Audit with go-stats-generator

### Non-Prompts (SKIP — these are reports, documentation, or plans, not agent prompts)
- `ENHANCEMENT_SUMMARY.md`, `MIGRATION.md`, `TEMPLATE_FIX.md`, `SINGLE_FILE_ANALYSIS.md`
- `DEFERRED_PLAN_DO_LATER.md`, `ROADMAP.md`, `README.md`

### Conversion Targets (REWRITE to use go-stats-generator)
Every ALL_CAPS `.md` file not listed above. Specifically:

| File | Purpose | Primary go-stats-generator Features |
|------|---------|-------------------------------------|
| `BASE_AUDIT.md` | Codebase audit vs README | `analyze --format json`, complexity hotspots, doc coverage |
| `BUGS.md` | Bug detection and fixing | `analyze` complexity/nesting for high-risk functions, `diff` for regression checks |
| `CLEANUP.md` | Repo cleanup | `analyze --skip-tests` before/after metrics, `diff` validation |
| `COLLATE_AUDIT.md` | Collate audit findings | `analyze --format json` for metric-backed prioritization of findings |
| `DOC.md` | Documentation accuracy audit | `analyze` doc coverage metrics, `--min-doc-coverage` threshold |
| `DOCS.md` | Godoc/markdown updates | `analyze` doc coverage, naming analysis for exported symbols |
| `EXECUTE.md` | Execute next planned item | `analyze` + `diff` to validate each change introduces no regressions |
| `FAIL.md` | Test failure analysis | `analyze` to correlate failures with complexity, `diff` to verify fixes |
| `GAPS.md` | Implementation gap analysis | `analyze --format json` full metrics vs README promises |
| `LOOP.md` | Iterative refactoring meta-prompt | Generic wrapper using `analyze` + `diff` continuation criteria |
| `MAKE_PLAN.md` | Generate implementation plan | `analyze --format json` to data-drive priority ordering |
| `META_AUDIT.md` | Per-package audit | `analyze [pkg-dir]` per-package metrics, `baseline` for tracking |
| `ORGANIZE.md` | Code reorganization | `analyze` package cohesion/coupling, `diff` before/after validation |
| `RESPOND.md` | Execute roadmap item | `analyze` + `diff` for measurable completion validation |
| `REVIEW.md` | Production readiness analysis | `analyze` full report as evidence base for readiness assessment |
| `TESTS.md` | Test coverage discovery | `analyze --format json` complexity ranking to prioritize test targets |

## go-stats-generator CAPABILITY REFERENCE

### Commands
| Command | Use |
|---------|-----|
| `analyze [dir\|file]` | Comprehensive metrics: functions, structs, packages, interfaces, concurrency, duplication, naming, docs |
| `diff <base.json> <new.json>` | Differential comparison with regression detection |
| `baseline create\|list\|compare` | Historical snapshot management |
| `trend analyze\|forecast\|regressions` | Trend analysis over time |

### Key Flags (analyze)
| Flag | Default | Purpose |
|------|---------|---------|
| `--format` | console | Output: console, json, csv, html, markdown |
| `--output` | stdout | Write to file |
| `--skip-tests` | false | Exclude `*_test.go` |
| `--max-function-length` | 30 | Length warning threshold |
| `--max-complexity` | 10 | Cyclomatic complexity threshold |
| `--min-doc-coverage` | 0.7 | Doc coverage threshold |
| `--min-block-lines` | 6 | Duplication minimum block size |
| `--similarity-threshold` | 0.80 | Near-clone similarity cutoff |
| `--workers` | CPU cores | Concurrency level |
| `--verbose` | false | Detailed output |

### Metrics Available in JSON Output
- `.functions[]` — complexity, lines, cyclomatic, nesting, signature, parameters
- `.structs[]` — members, methods, embedded types, tags
- `.packages[]` — dependencies, cohesion, coupling, circular refs
- `.interfaces[]` — implementations, embedding depth, signature complexity
- `.concurrency` — goroutines, channels, sync primitives, patterns
- `.duplication` — clone pairs, duplicated lines, ratio, clone details
- `.naming` — convention violations by category
- `.documentation` — coverage ratio, missing docs, quality scores

## CONVERSION INSTRUCTIONS

For each target file, apply these steps in order:

### Step 1: Read the Existing Prompt
Understand its core intent, execution mode, phases, and output format. Preserve the original purpose entirely.

### Step 2: Determine the Optimal go-stats-generator Integration
Each prompt has a different logical workflow. Match go-stats-generator features to the prompt's needs:
- **Audit/review prompts** → `analyze` as evidence-gathering phase before human-style review
- **Bug-fixing/refactoring prompts** → `analyze` baseline → fix → `analyze` + `diff` validation
- **Planning/prioritization prompts** → `analyze --format json` to data-drive ordering
- **Meta/loop prompts** → `analyze` + `diff` as continuation/termination criteria
- **Documentation prompts** → `analyze` doc coverage + naming metrics
- **Test prompts** → `analyze` complexity ranking for test target prioritization
- **Organization prompts** → `analyze` package metrics for cohesion-driven decisions

### Step 3: Rewrite Following Exemplar Structure
Mirror the structure of `BREAKDOWN.md` / `DEDUPLICATE.md`:

1. **TASK DESCRIPTION** — Single sentence stating the data-driven objective
2. **CONSTRAINT** — "Use only `go-stats-generator` and existing tests" (adapt per prompt's actual constraints)
3. **PREREQUISITES** — Installation check + `jq` recommendation
4. **CONTEXT** — Role description referencing go-stats-generator as primary analysis engine
5. **INSTRUCTIONS** — Phased workflow:
   - Phase 1: Baseline data collection with specific `go-stats-generator` commands
   - Phase 2: Analysis/action guided by metrics (adapt to prompt's purpose)
   - Phase 3: Validation via `diff` or post-analysis comparison
   - Phase 4: Quality verification with concrete pass/fail criteria
6. **OUTPUT FORMAT** — Structured template with metric placeholders
7. **THRESHOLDS** — Concrete numeric criteria (adapt thresholds to prompt's domain)
8. **EXAMPLE WORKFLOW** — Realistic `go-stats-generator` command/output sequence

### Step 4: Adapt Thresholds and Criteria to the Prompt's Domain
Do NOT blindly copy BREAKDOWN.md thresholds. Example adaptations:
- **BUGS.md**: Prioritize functions by `cyclomatic > 15` as high-risk bug candidates
- **DOCS.md**: Use `--min-doc-coverage 0.8` and flag exported symbols with missing godoc
- **TESTS.md**: Rank untested files by complexity score for highest-value test targets
- **ORGANIZE.md**: Use package coupling/cohesion metrics to guide reorganization
- **REVIEW.md**: Define production-readiness gates across all metric dimensions

### Step 5: Validate the Rewritten Prompt
Ensure:
- [ ] Original intent is fully preserved
- [ ] go-stats-generator commands are syntactically correct (match the flags above)
- [ ] JSON field paths match the actual output schema (`.functions[]`, `.duplication`, etc.)
- [ ] Execution mode matches the original (autonomous/report/interactive)
- [ ] No references to `gostats` — only `go-stats-generator`
- [ ] Phased workflow is logically ordered for THIS specific task
- [ ] Thresholds are appropriate for the prompt's domain, not copied from BREAKDOWN.md

## SUCCESS CRITERIA
- All target files rewritten with go-stats-generator integration
- Each rewritten prompt includes: prerequisites, phased workflow with concrete commands, validation via `diff`, output format template, and numeric thresholds
- Original intent of each prompt is preserved
- No `gostats` references anywhere in the repository
- All go-stats-generator commands use valid flags and syntax
- Already-optimized files left unchanged (or minimally updated for structural consistency)