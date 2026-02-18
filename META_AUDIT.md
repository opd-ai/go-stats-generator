TASK: Select one unaudited Go sub-package, perform implementation audit against Go best practices, generate findings document, and update root audit tracker.

EXECUTION MODE: Autonomous action with structured reporting.

WORKFLOW:

1. PACKAGE SELECTION
   a) Read root `AUDIT.md` to identify audited packages
   b) If all packages audited, report completion and exit
   c) Select ONE unaudited sub-package from `pkg/`, `internal/`, or `cmd/` prioritizing:
      - Packages listed in root `AUDIT.md` but unchecked
      - High integration surface (many imports/importers)
      - Core business logic packages
   d) State chosen package path and 1-sentence rationale

2. CODE AUDIT
   Run checks on selected package, citing `file.go:LINE` for every finding:

   **Stub/Incomplete Code**
   - Functions returning only `nil`/zero values
   - `TODO`/`FIXME`/`placeholder` comments
   - Empty method bodies or unimplemented interfaces

   **API Design**
   - Exported types/functions follow Go naming conventions
   - Interfaces are minimal and focused
   - No unnecessary concrete type exposure

   **Concurrency Safety**
   - Shared state protected by mutexes or channels
   - No race conditions (verify with `go test -race`)
   - Context cancellation properly handled

   **Determinism & Reproducibility**
   - No direct use of `time.Now()` in build logic
   - Random number generation uses explicit seeds when determinism required for reproducible builds
   - No reliance on OS-specific or environment-dependent behavior

   **Error Handling**
   - All returned errors checked
   - No swallowed errors (`_ = err`)
   - Errors wrapped with context (`fmt.Errorf` with `%w`)
   - Critical errors logged with structured context

   **Test Coverage**
   - Run: `go test -cover ./path/to/pkg/...`
   - Flag if below 65% target
   - Note missing table-driven tests or benchmarks

   **Documentation**
   - Exported types/functions have godoc comments
   - Package has `doc.go` file
   - Complex algorithms have inline explanations

   **Dependencies**
   - No circular import dependencies
   - External dependencies justified and minimal
   - Standard library preferred where possible

3. FILE OPERATIONS (execute in this order)

   a) Create `<package-dir>/AUDIT.md`:
   ```markdown
   # Audit: <package-import-path>
   **Date**: YYYY-MM-DD
   **Status**: Complete | Incomplete | Needs Work

   ## Summary
   <2-3 sentences: scope, overall health, critical risks>

   ## Issues Found
   - [ ] <high|med|low> <category> — <description> (`file.go:LINE`)

   ## Test Coverage
   <percentage>% (target: 65%)

   ## Dependencies
   <External dependencies and integration points>

   ## Recommendations
   1. <highest-priority fix>
   2. <next priority fix>
   ```

   b) Update root `AUDIT.md`:
      - If package listed: change `[ ]` to `[x]`, append: `— <Status> — <N> issues (<H> high, <M> med, <L> low)`
      - If package not listed: append new checked entry with status

   c) Run: `go vet ./path/to/pkg/...`

4. CHAT REPORT (max 500 words)
   - Created file path: `<package-dir>/AUDIT.md`
   - Test coverage: `<N>%`
   - Top 3-5 critical findings with `file.go:LINE` citations
   - `go vet` result: PASS/FAIL
   - Updated root `AUDIT.md`: YES

OUTPUT FORMAT:
- Package AUDIT.md: Use exact template structure
- Root AUDIT.md: Single-line update only
- Chat report: Bullet list, no prose paragraphs

SUCCESS CRITERIA:
✓ Exactly one package audited
✓ All issues cite specific `file.go:LINE` locations
✓ Test coverage percentage reported
✓ Root `AUDIT.md` contains checked entry for audited package
✓ `go vet` executed and result documented
✓ Chat report under 500 words
