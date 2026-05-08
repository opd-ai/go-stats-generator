# TASK: Audit a target third-party codebase for proper code organization and extensible, library-first architecture.

## Execution Mode
**Audit only** — identify findings and remediation actions; do not implement code changes.

## Core Audit Principles
1. **Library-forward design**: business logic should live in reusable packages, not in `main`.
2. **Thin entrypoint**: if `main` exists, it should orchestrate library calls and CLI wiring only.
3. **Interface-driven APIs**: public functions should accept interfaces instead of concrete implementations.
4. **Struct/interface pairing**: exported structs should have corresponding exported interfaces where abstraction is expected.
5. **Separation of concerns**: boundaries must be clear in directories, files, and data structures.

## Workflow

### Phase 0: Discover Organization Model
1. Map top-level directories and package responsibilities.
2. Identify entrypoints (`main` packages) and trace where core functionality is implemented.
3. Catalog public structs, public interfaces, and public function signatures.
4. Identify directory/file naming and placement conventions.

### Phase 1: Library-Forward Audit
1. Flag logic implemented directly in `main` that belongs in libraries.
2. Verify `main` is limited to argument parsing, dependency construction, orchestration, and output handling.
3. Record cases where feature logic is tightly coupled to CLI or process lifecycle concerns.

### Phase 2: Interface and API Boundary Audit
1. For each exported struct, verify whether an exported interface abstraction exists when needed for extensibility.
2. For each public function, check whether parameters use interfaces at package boundaries instead of concrete types.
3. Flag over-concrete APIs that make testing, substitution, or extension difficult.
4. Identify redundant interfaces that do not improve boundaries.

### Phase 3: Separation-of-Concerns Audit
1. Verify directories represent clear domain or architectural boundaries.
2. Verify files group related concerns and avoid mixed responsibilities.
3. Verify data structures are placed in packages matching their ownership and usage.
4. Flag circular or cross-layer dependencies that violate intended boundaries.

### Phase 4: Findings and Remediation
Report findings by severity:
- **CRITICAL**: architecture blocks extensibility or places core logic in entrypoints.
- **HIGH**: public APIs over-expose concrete implementations across boundaries.
- **MEDIUM**: unclear package/file responsibility or misplaced data structures.
- **LOW**: naming/placement inconsistencies with low architectural impact.

For every finding include:
1. **Evidence**: file/path and concrete example.
2. **Impact**: why it harms extensibility, maintainability, or testability.
3. **Remediation**: specific reorganization action (what to move/split/abstract).
4. **Validation**: command(s) to confirm no regressions after refactor.

## Output Format
```markdown
# Organization Audit

## Summary
- Library-forward score: [good/partial/poor]
- Entrypoint thinness: [good/partial/poor]
- Interface boundary health: [good/partial/poor]
- Separation-of-concerns health: [good/partial/poor]

## Findings
### CRITICAL
- [ ] [Title]
  - Evidence: [path]
  - Impact: [...]
  - Remediation: [...]
  - Validation: [...]

### HIGH
- [ ] ...

### MEDIUM
- [ ] ...

### LOW
- [ ] ...
```

## Constraints
- Evaluate organization quality against the codebase’s own conventions first, then against these architecture principles.
- Do not conflate stylistic preferences with architecture flaws.
- Prefer minimal, high-leverage structural remediation recommendations.
