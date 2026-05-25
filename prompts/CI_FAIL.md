# CI Failure Resolution

**Objective:** Analyze the most recent CI/CD pipeline failure in the current repository and create a comprehensive resolution plan.

## Instructions

### Step 1: Identify the Most Recent CI Failure

1. Use GitHub Actions tools to list recent workflow runs
2. Identify the most recent failed run on the main/master branch
3. Document:
   - CI Run ID and URL
   - Date and time of failure
   - Branch name
   - Commit SHA
   - Failed step/job name

### Step 2: Analyze the Failure

1. Retrieve the detailed logs for the failed job
2. Identify all errors, warnings, and failure messages
3. Categorize issues by type:
   - Build failures
   - Test failures
   - Linting/style issues
   - Security vulnerabilities
   - Deprecation warnings
   - Integration/deployment issues
   - Other issues

### Step 3: Document Each Issue

For each identified issue, document:
- **File and line number** where the issue occurs
- **Error type/category** (e.g., errcheck, unused, gosimple, staticcheck)
- **Error message** from the CI logs
- **Brief description** of what's wrong

### Step 4: Create Resolution Strategy

Prioritize issues based on severity and impact:

**Priority 1: Critical Failures**
- Build failures that prevent compilation
- Critical security vulnerabilities
- Breaking test failures

**Priority 2: High-Impact Issues**
- Error handling issues
- Resource leaks
- Race conditions
- Integration test failures

**Priority 3: Code Quality Issues**
- Unused code
- Code simplification opportunities
- Style/formatting issues
- Deprecation warnings

**Priority 4: Documentation and Warnings**
- Documentation gaps
- Minor deprecations
- Non-critical warnings

### Step 5: Provide Resolution Guidance

For each category of issues, provide:
- Specific fix recommendations
- Code examples where appropriate
- Links to relevant documentation
- Estimated effort/complexity

### Step 6: Validation Commands

List the commands needed to validate fixes:
```bash
# Example commands (adjust based on the repository)
make lint
make build
make test
# Or:
npm test
npm run lint
# Or:
cargo test
cargo clippy
```

## Output Format

Create a detailed report with the following sections:

1. **Executive Summary**: Brief overview of the CI failure
2. **Failure Details**: Complete information about the failed run
3. **Issues Breakdown**: Categorized list of all issues found
4. **Resolution Plan**: Prioritized action items with specific guidance
5. **Validation Steps**: Commands to run after fixes
6. **Additional Notes**: Any repository-specific considerations

## Example Structure

```markdown
# CI Failure Analysis - [Repository Name]

## Executive Summary
[Brief description of what failed and why]

## Failure Details
- **CI Run:** [#12345](link)
- **Date:** YYYY-MM-DD HH:MM:SS
- **Branch:** main
- **Commit:** abc123def
- **Failed Step:** [step name]

## Issues Found

### Category 1: [Issue Type]
1. `path/to/file.ext:123` - [description]
2. `path/to/file.ext:456` - [description]

### Category 2: [Issue Type]
[Continue for each category...]

## Resolution Plan

### Priority 1: [Category]
**Issues to address:**
- [Issue 1]
- [Issue 2]

**Resolution guidance:**
[Specific instructions...]

## Validation
```bash
[Commands to run]
```

## Notes
[Any additional context]
```

## Tips

- Use GitHub Actions API tools to retrieve logs programmatically
- Parse logs carefully to extract all unique error messages
- Group related errors together for efficient fixing
- Consider root causes - fixing one issue may resolve multiple errors
- Check if failures are related to recent changes vs. existing issues
- Note any flaky tests or intermittent failures
