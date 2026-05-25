# TASK: Analyze recent CI/CD pipeline failures and create a comprehensive resolution plan

## Execution Mode
**Report generation only** — do NOT modify any source code unless specifically instructed.

## Output
Write the following file in the repository root (the directory containing `go.mod`):
- **`CI_FAIL_ANALYSIS.md`** — the consolidated CI failure analysis report

If the file already exists, you may update it or create a fresh one.

## Objective
Analyze the most recent CI/CD pipeline failures on the main/master branch to identify persistent issues, categorize them by severity, and provide actionable resolution guidance.

## Workflow

### Phase 1: Identify Recent CI Failures

1. **List recent workflow runs:** Use GitHub Actions API to fetch recent runs from the primary CI workflow
   - Focus on the main/master branch
   - Identify all **failed** runs in the last 10-20 runs
   - Document:
     - Run ID and URL
     - Date and time of failure
     - Branch name
     - Commit SHA
     - Failed step/job name
     - Run number

2. **Determine failure pattern:** Identify if failures are:
   - Consistent (fail on every run)
   - Intermittent (flaky)
   - Related to specific recent changes
   - Affecting a specific CI step

### Phase 2: Analyze Failure Details

For each failed run, retrieve detailed logs and extract:

1. **Error messages and stack traces** — What exactly failed?
2. **Error categorization** — Is it:
   - Build failure (compilation error)
   - Test failure (test assertion failed)
   - Linting/style issue (code quality check failed)
   - Security vulnerability
   - Deprecation warning
   - Integration/deployment issue
   - Infrastructure/environment issue
   - Other

3. **Root cause analysis** — Why did it fail?
   - Code issue (logic error, missing error handling, etc.)
   - Configuration issue (missing env var, wrong settings)
   - Flaky test (timing-dependent, external service dependency)
   - Environmental issue (temporary infrastructure problem)
   - Dependency issue (transitive dependency update, CVE)

4. **Affected components** — Which part of the codebase?
   - Specific file paths
   - Specific functions or modules
   - Specific subsystems (API, storage, CLI, analyzer, etc.)

### Phase 3: Aggregate and Categorize

Group failures by:

1. **Error type** (e.g., errcheck, unused, staticcheck, test failures)
2. **Severity level**:
   - **Critical:** Build failures, breaking tests, security vulnerabilities
   - **High:** Error handling issues, resource leaks, race conditions
   - **Medium:** Code quality issues, deprecation warnings
   - **Low:** Style issues, minor warnings

3. **Frequency** — How many recent runs were affected?
4. **Related to recent changes** — Can failures be traced to specific commits or PRs?

### Phase 4: Create Resolution Plan

For each category of issues, provide:

1. **Issue Summary**
   - What is the issue?
   - Which files/functions are affected?
   - What is the impact?

2. **Root Cause**
   - Why is this happening?
   - Is it a code issue, configuration issue, or external factor?

3. **Resolution Guidance**
   - Specific steps to fix the issue
   - Code examples where appropriate
   - Links to relevant documentation
   - Estimated effort/complexity

4. **Validation Steps**
   - How to verify the fix works
   - Commands to run locally
   - Commands to validate in CI

5. **Priority** — Where should this be addressed first?

### Phase 5: Provide Implementation Roadmap

Create a prioritized list of actions:

1. **Priority 1: Critical Issues** (address immediately)
2. **Priority 2: High-Impact Issues** (address next)
3. **Priority 3: Medium-Impact Issues** (schedule soon)
4. **Priority 4: Low-Impact Issues** (nice-to-have improvements)

## Output Format

Structure the report as follows:

```markdown
# CI Failure Analysis Report

## Executive Summary
[Brief overview of what's failing and why]

## Failure Summary
- **Failed runs:** [Count and run numbers]
- **Time period:** [Date range]
- **Branch:** main/master
- **Most common failure:** [Type of failure]
- **Recency:** [How recent/frequent]

## Failures by Category

### Category 1: [Issue Type] (N issues)
[List of specific failures with file:line, error message, impact]

### Category 2: [Issue Type] (N issues)
[List of specific failures with file:line, error message, impact]

## Root Cause Analysis

### Contributing Factors
- [Factor 1]: [Explanation]
- [Factor 2]: [Explanation]

## Resolution Plan

### Priority 1: [Issue]
**Files affected:** [list]
**Resolution:** [specific guidance]
**Validation:** [commands to run]
**Effort:** [estimate]

### Priority 2: [Issue]
[Same structure]

## Recommendations

1. [Recommendation 1]
2. [Recommendation 2]

## Appendix: Raw Failure Data

[Include specific error messages, stack traces, and log excerpts]
```

## Key Questions to Answer

- **Are failures consistent or intermittent?** (Consistency impacts debugging strategy)
- **Did failures start recently?** (Recent changes may be the cause)
- **Are failures related to environment/infrastructure?** (May not be code issues)
- **Are there patterns across failures?** (Multiple failures with same root cause)
- **What is the critical path to fix?** (Which fixes unblock other issues?)
- **Are there actionable quick wins?** (Fix easy issues first to build momentum)

## Tips

- Use GitHub Actions API tools to retrieve logs programmatically
- Parse logs carefully to extract all unique error messages
- Group related errors together for efficient fixing
- Consider root causes — fixing one issue may resolve multiple CI failures
- Check if failures are related to recent changes vs. existing issues
- Note any flaky tests or intermittent failures — they need different treatment
- Look for patterns: same error appearing across multiple runs indicates a systemic issue
- Consider the broader context: has a dependency been updated? Did infrastructure change?
