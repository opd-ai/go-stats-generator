## Objective
Autonomously clean up the repository by removing binary files, redundant reports, and consolidating tests. Update `.gitignore` to prevent reintroduction of removed file types.

## Execution Mode
**Autonomous Action** — Execute all steps directly. No user approval required between steps.

## Task Steps (execute in order)

### 1. Remove Binary Files
- Scan the repository for committed binary artifacts (e.g., `.o`, `.so`, `.dll`, `.exe`, `.bin`, `.pyc`, `.class`, `.wasm`, compiled executables, and similar non-source files).
- Delete all identified binary files from the repository.
- **Do not remove** binaries that serve as intentional test fixtures or are documented as required assets.

### 2. Remove Redundant Reports
- Identify duplicate, outdated, or auto-generated report files (e.g., repeated coverage reports, stale build logs, duplicate analysis outputs).
- Delete files that are exact duplicates or superseded by newer versions.
- **Do not remove** the most recent version of any report that has no replacement.

### 3. Consolidate Tests
- Merge test files that contain overlapping or duplicate test cases into single cohesive test files, grouped by module or feature.
- Remove duplicate test cases while preserving full test coverage (no test logic lost).
- Ensure all consolidated test files pass after merging.

### 4. Update `.gitignore`
- Append entries to `.gitignore` (or create it if absent) to prevent future commits of:
  - Binary artifact types removed in Step 1 (e.g., `*.o`, `*.so`, `*.dll`, `*.exe`, `*.bin`, `*.pyc`, `*.class`, `*.wasm`)
  - Auto-generated report directories/patterns removed in Step 2
- Do not duplicate entries already present in `.gitignore`.
- Group new entries under a `# Repository cleanup` comment header.

## Success Criteria
- All committed binary artifacts are removed (except intentional fixtures)
- No actively used or necessary files are deleted
- Full test coverage is preserved — no test cases lost
- `.gitignore` prevents reintroduction of all removed file types
- Repository builds and tests pass after all changes

## Output Format
After execution, provide a summary:
```
### Cleanup Summary
- **Binaries removed**: <count> files (<total size>)
- **Reports removed**: <count> files
- **Tests consolidated**: <count> files merged into <count> files
- **`.gitignore` entries added**: <count> new patterns
- **Tests passing**: Yes/No
```
