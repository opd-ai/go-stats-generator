# TASK DESCRIPTION:
Generate a data-driven implementation plan by combining `go-stats-generator` codebase metrics with project roadmap documentation to identify and prioritize the next incomplete development phase into an actionable `PLAN.md`.

## CONSTRAINT:

**Report generation only** — Produce `PLAN.md` (and optionally `GAPS.md`) on disk in the repository root directory. Do not commit files or open PRs. Use only `go-stats-generator`, `jq`, and repository documentation for your analysis.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive codebase analysis to inform planning decisions:

### Installation:
```bash
# First, check if go-stats-generator is already installed
which go-stats-generator
# If not, install it with `go install`
go install github.com/opd-ai/go-stats-generator@latest
```

## Recommendations:
```bash
# When long json outputs are encountered, use `jq`
go-stats-generator analyze --format json | jq .
# Check if it is installed
which jq
# If it is not, install it
sudo apt-get install jq
```

## CONTEXT:
You are an automated Go project planner using `go-stats-generator` for data-driven development planning. The tool provides precise codebase metrics — complexity hotspots, duplication ratios, documentation coverage, package coupling, and concurrency patterns — that you correlate with roadmap milestones to prioritize work, estimate scope, and ground every planning decision in quantitative evidence rather than subjective assessment.

## INSTRUCTIONS:

### Phase 1: Codebase Metrics Collection
1. **Run Comprehensive Analysis:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output metrics.json
  go-stats-generator analyze . --skip-tests
  ```
  - Capture the full metrics snapshot for planning reference
  - Review the console summary for an overview of codebase health

2. **Extract Key Planning Indicators:**
  ```bash
  # Complexity hotspots — which areas need most attention
  cat metrics.json | jq '[.functions[] | select(.complexity.overall > 9)] | sort_by(-.complexity.overall) | .[:10] | .[] | {name, file, complexity: .complexity.overall, lines: .lines.code}'

  # Duplication burden — where consolidation would reduce scope
  cat metrics.json | jq '.duplication | {clone_pairs, duplicated_lines, duplication_ratio, largest_clone_size}'

  # Documentation coverage — what needs docs before new work
  cat metrics.json | jq '.documentation'

  # Package coupling — which packages are most entangled
  cat metrics.json | jq '[.packages[] | {name, coupling_score, cohesion_score, dependencies: (.dependencies | length)}] | sort_by(-.coupling_score)'

  # Concurrency patterns — complexity of concurrent code
  cat metrics.json | jq '.concurrency'
  ```

### Phase 2: Documentation Review and Phase Selection
1. **Review Project Documentation (in priority order):**
  - `ROADMAP.md` — Identify the earliest incomplete milestone or phase
  - `README.md` — Understand project purpose, architecture, and current capabilities
  - `docs/*.md` — Gather implementation details, specs, and existing plans
  - Source code structure — Assess current state against documented goals

  **If `ROADMAP.md` does not exist**, infer the next phase from any planning documents found (e.g., `FOUNDATION.md`, `TODO.md`, `CHANGELOG.md`, project board references), then note this substitution in the output.

2. **Apply Phase Selection Rules:**
  - Select the **earliest incomplete milestone** in the roadmap
  - If all milestones are complete: propose the next logical enhancement based on project trajectory and metrics analysis
  - If no roadmap or planning document exists: state this clearly and derive a phase from metrics analysis — target the area with the worst metric scores

3. **Correlate Metrics with Roadmap Items:**
  ```bash
  # Identify which roadmap areas overlap with complexity hotspots
  cat metrics.json | jq '[.functions[] | select(.complexity.overall > 9)] | group_by(.file) | .[] | {file: .[0].file, count: length, max_complexity: (map(.complexity.overall) | max)}'

  # Identify packages with highest technical debt indicators
  cat metrics.json | jq '[.packages[] | select(.coupling_score > 0.7 or .cohesion_score < 0.3)] | .[] | {name, coupling_score, cohesion_score}'
  ```
  - Roadmap items touching high-complexity areas get **higher scope estimates**
  - Roadmap items touching high-duplication areas should include deduplication as a prerequisite step
  - Roadmap items touching poorly-documented areas should include documentation as a deliverable

### Phase 3: Data-Driven Plan Generation
1. **Estimate Scope Using Metrics:**
  - **Small**: Target area has <5 functions above complexity threshold, duplication ratio <3%, documentation coverage >70%
  - **Medium**: Target area has 5-15 functions above complexity threshold, duplication ratio 3-8%, or documentation coverage 40-70%
  - **Large**: Target area has >15 functions above complexity threshold, duplication ratio >8%, or documentation coverage <40%

2. **Prioritize Implementation Steps:**
  - Order steps by metric severity: highest-complexity or highest-duplication areas first
  - Include prerequisite cleanup steps when metrics indicate tech debt in the target area
  - Define validation criteria using specific metric thresholds from `go-stats-generator`

3. **Write PLAN.md to Disk:**
  Create `PLAN.md` in the repository root with the structured plan (see OUTPUT FORMAT below).

4. **Write GAPS.md if Needed:**
  If any information is missing or ambiguous, create `GAPS.md` in the repository root.

### Phase 4: Plan Validation
1. **Verify Plan Completeness:**
  - Every step is independently actionable by a developer
  - No step requires information that is undefined — flag it as a gap instead
  - Dependencies between steps are explicit
  - Deliverables are concrete artifacts (files, functions, passing tests), not vague outcomes
  - Technical specifications answer "how", not just "what"
  - Scope estimates are grounded in quantitative metrics from `go-stats-generator`
  - Validation criteria reference specific measurable thresholds

2. **Cross-Check Against Metrics:**
  ```bash
  # Confirm the selected phase addresses the highest-priority issues
  cat metrics.json | jq '{high_complexity_count: [.functions[] | select(.complexity.overall > 15)] | length, duplication_ratio: .duplication.duplication_ratio, doc_coverage: .documentation.coverage.overall}'
  ```
  - The plan should address the most impactful issues revealed by the metrics
  - If metrics reveal critical issues outside the selected phase, note them in Known Gaps

## OUTPUT FORMAT:

**Create the `PLAN.md` file on disk in the repository root directory.** The file must contain the structured plan described below.

Additionally, produce **three separate markdown code blocks** for reference, each labeled with its filename:

### 1. `PLAN.md`

```
# Implementation Plan: [Phase Name]

## Phase Overview
- **Objective**: [One-sentence goal]
- **Source Document**: [File used to identify this phase]
- **Prerequisites**: [Completed items required]
- **Estimated Scope**: [Small / Medium / Large] — based on go-stats-generator metrics

## Metrics Summary
- **Complexity Hotspots**: [n] functions above threshold in target area
- **Duplication Ratio**: [x.xx]% in target area
- **Documentation Coverage**: [x.xx]% in target area
- **Package Coupling**: [summary of coupling in target area]

## Implementation Steps
1. [Actionable task]
   - **Deliverable**: [Specific, verifiable output]
   - **Dependencies**: [If any]
   - **Metric Justification**: [Which go-stats-generator metric drives this priority]

2. [Actionable task]
   - **Deliverable**: [Specific, verifiable output]
   - **Dependencies**: [If any]
   - **Metric Justification**: [Which go-stats-generator metric drives this priority]

## Technical Specifications
- [Key technical decision or constraint]
- [Key technical decision or constraint]

## Validation Criteria
- [ ] [Measurable success criterion with specific metric threshold]
- [ ] [Measurable success criterion with specific metric threshold]
- [ ] go-stats-generator diff shows no regressions in unrelated areas

## Known Gaps
- [Gap description, or "None identified"]
```

### 2. `GAPS.md` (only if gaps exist; otherwise output "Not needed — no gaps identified")

```
# Implementation Gaps: [Phase Name]

## [Gap Title]
- **Description**: [What information is missing]
- **Impact**: [How it blocks implementation]
- **Metrics Context**: [What go-stats-generator data relates to this gap]
- **Resolution**: [What is needed to close this gap]
```

### 3. `ROADMAP.md` addition (exact text to append)

```
- [YYYY-MM-DD] PLAN.md created for [Phase Name]
```

## THRESHOLDS:
```
Scope Estimation (from go-stats-generator metrics):
  Small  = <5 functions above complexity 9.0, duplication <3%, doc coverage >70%
  Medium = 5-15 functions above complexity 9.0, duplication 3-8%, doc coverage 40-70%
  Large  = >15 functions above complexity 9.0, duplication >8%, doc coverage <40%

Priority Ordering (metric-driven):
  Critical = complexity >20.0 OR duplication ratio >10% OR doc coverage <20%
  High     = complexity 15.0-20.0 OR duplication ratio 5-10% OR doc coverage 20-40%
  Medium   = complexity 9.0-15.0 OR duplication ratio 3-5% OR doc coverage 40-70%

Quality Gates for Generated Plan:
  Every step independently actionable
  No undefined information — gaps flagged explicitly
  Explicit dependencies between steps
  Concrete deliverables (files, functions, passing tests)
  Scope estimates grounded in go-stats-generator metrics
  Validation criteria reference specific metric thresholds
```

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests
=== CODEBASE METRICS SUMMARY ===
Functions Analyzed: 142
  Above Complexity Threshold (9.0): 8
  Above Length Threshold (40 lines): 5
  Average Complexity: 6.2

Duplication:
  Clone Pairs: 7
  Duplicated Lines: 93
  Duplication Ratio: 4.18%

Documentation Coverage: 62.3%

Packages: 9
  High Coupling (>0.7): 2
  Low Cohesion (<0.3): 1

$ go-stats-generator analyze . --skip-tests --format json --output metrics.json

$ # Extract complexity hotspots for planning
$ cat metrics.json | jq '[.functions[] | select(.complexity.overall > 9)] | sort_by(-.complexity.overall) | .[:5] | .[] | {name, file, complexity: .complexity.overall, lines: .lines.code}'
{"name": "analyzeFile", "file": "internal/analyzer/file.go", "complexity": 22.4, "lines": 58}
{"name": "processResults", "file": "internal/reporter/console.go", "complexity": 16.1, "lines": 45}
{"name": "detectPatterns", "file": "internal/patterns/detect.go", "complexity": 14.8, "lines": 42}
{"name": "resolveImports", "file": "internal/scanner/imports.go", "complexity": 11.3, "lines": 35}
{"name": "buildMetrics", "file": "internal/metrics/builder.go", "complexity": 10.7, "lines": 31}

$ # Check duplication for prerequisite cleanup
$ cat metrics.json | jq '.duplication | {clone_pairs, duplicated_lines, duplication_ratio}'
{"clone_pairs": 7, "duplicated_lines": 93, "duplication_ratio": 0.0418}

$ # Review ROADMAP.md for next incomplete milestone
$ head -50 ROADMAP.md
# Roadmap
## Phase 1: Core Analysis Engine ✅
## Phase 2: Advanced Metrics ✅
## Phase 3: Pattern Detection (In Progress)
  - [ ] Design pattern recognition
  - [ ] Anti-pattern detection
  - [x] Concurrency pattern analysis

$ # Correlate: Phase 3 targets internal/patterns/ — detectPatterns has 14.8 complexity
$ # Scope estimate: 3 functions above threshold in target area, duplication <5%, docs 62% → Small/Medium

$ # Generate PLAN.md with metric-driven priorities
$ cat PLAN.md
# Implementation Plan: Phase 3 — Pattern Detection

## Phase Overview
- **Objective**: Complete design pattern recognition and anti-pattern detection
- **Source Document**: ROADMAP.md (Phase 3: Pattern Detection)
- **Prerequisites**: Phase 1 and Phase 2 complete
- **Estimated Scope**: Medium — 3 functions above complexity threshold, 4.18% duplication, 62.3% doc coverage

## Metrics Summary
- **Complexity Hotspots**: 3 functions above threshold in target packages
- **Duplication Ratio**: 4.18% overall
- **Documentation Coverage**: 62.3% overall
- **Package Coupling**: internal/patterns has 0.65 coupling (acceptable)

## Implementation Steps
1. Refactor detectPatterns (complexity 14.8 → target <9.0)
   - **Deliverable**: Refactored internal/patterns/detect.go with extracted helpers
   - **Dependencies**: None
   - **Metric Justification**: Highest complexity in target area (14.8)

2. Implement design pattern recognition
   - **Deliverable**: internal/patterns/design.go with Singleton, Factory, Observer detection
   - **Dependencies**: Step 1 (clean base to build on)
   - **Metric Justification**: ROADMAP.md incomplete item

3. Implement anti-pattern detection
   - **Deliverable**: internal/patterns/antipattern.go with God Object, Feature Envy detection
   - **Dependencies**: Step 2 (shared pattern infrastructure)
   - **Metric Justification**: ROADMAP.md incomplete item

## Technical Specifications
- Pattern detection via AST analysis using go/ast visitor pattern
- Results integrated into existing metrics JSON output under .patterns key

## Validation Criteria
- [ ] go-stats-generator analyze shows 0 functions above complexity 9.0 in internal/patterns/
- [ ] go-stats-generator diff baseline.json final.json shows no regressions
- [ ] Documentation coverage in target packages ≥ 70%
- [ ] All tests pass: go test ./internal/patterns/...

## Known Gaps
- None identified

$ # Verify plan addresses highest-priority metrics
$ cat metrics.json | jq '{high_complexity_count: [.functions[] | select(.complexity.overall > 15)] | length, duplication_ratio: .duplication.duplication_ratio}'
{"high_complexity_count": 2, "duplication_ratio": 0.0418}
```

This data-driven approach ensures planning decisions are grounded in quantitative codebase analysis from `go-stats-generator`, with scope estimates and priority ordering derived from actual metrics rather than subjective assessment.
