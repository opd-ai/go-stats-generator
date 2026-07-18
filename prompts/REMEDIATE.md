TASK: Perform an exhaustive end-to-end audit of the entire Go codebase for code quality improvements and generate a complete remediation plan in PLAN.md.

EXECUTION MODE: Report generation only - no code changes will be made.

SCOPE: Analyze ALL Go files in the codebase with no finding limits.

ANALYSIS CATEGORIES:
1. Code duplication (identical or similar functions/blocks ≥10 lines)
2. DRY violations (repeated logic patterns across 3+ locations)
3. Simplification opportunities (complexity reduction without functionality loss)
4. Adherence to Go idioms and existing project conventions

OUTPUT STRUCTURE (PLAN.md):

## Audit Summary
- Total files analyzed: [count]
- Total issues found: [count by category]
- Estimated total effort: [hours]
- Priority breakdown: [high/medium/low counts]

## High Priority Issues
For each issue:
- **Location**: `package/file.go:lines`
- **Type**: [Duplication|DRY|Simplification]
- **Description**: [brief explanation]
- **Proposed fix**: [specific approach]
- **Impact**: [lines reduced, complexity improvement]

## Medium Priority Issues
[Same format as above]

## Low Priority Issues
[Same format as above]

## Implementation Checklist
- [ ] Issue 1: [location + brief description]
- [ ] Issue 2: [location + brief description]
[Continue for ALL identified issues]

ANALYSIS STANDARDS:
- Duplication: ≥10 lines of similar code in 2+ locations
- DRY violation: same logic pattern in 3+ locations
- Simplification: cyclomatic complexity >10 OR nesting depth >4
- Must preserve all functionality, public APIs, and existing behavior
- Must respect current project code style and error handling patterns

REQUIREMENTS:
- Document EVERY instance found across the entire codebase
- Do not cap or limit the number of findings
- Scan all .go files including tests
- Prioritize issues by impact and effort