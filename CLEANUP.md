## Objective
Perform a repository cleanup by removing binary files, redundant/duplicate reports, and consolidating test files.

## Execution Mode
**Report/Plan Generation** *(assumed — original prompt did not specify an execution mode)*

## Task Breakdown

1. **Identify binary files**: Scan the repository for committed binary files (e.g., compiled executables, `.o`, `.so`, `.dll`, `.exe`, `.bin`, `.pyc`, `.class`, and similar artifacts) that should not be tracked in version control. List each file with its path and size.

2. **Identify redundant reports**: Find duplicate or outdated report files (e.g., repeated coverage reports, stale build logs, duplicate analysis outputs). Flag files that are duplicates or superseded by newer versions.

3. **Identify consolidation opportunities in tests**: Detect test files that contain overlapping or duplicate test cases, tests split unnecessarily across multiple files, or test utilities that could be merged. Group related findings by module or directory.

## Expected Output Format

Produce a structured report with three sections:

```
### 1. Binary Files to Remove
| File Path | File Type | Size |
|-----------|-----------|------|

### 2. Redundant Reports to Remove
| File Path | Reason (duplicate/outdated/superseded by) |
|-----------|-------------------------------------------|

### 3. Test Consolidation Recommendations
| Current Files | Proposed Action | Rationale |
|---------------|-----------------|-----------|
```

## Success Criteria
- All committed binary artifacts are identified
- No actively used or necessary files are flagged for removal
- Test consolidation suggestions preserve full test coverage (no test cases lost)
- Each recommendation includes a clear rationale
