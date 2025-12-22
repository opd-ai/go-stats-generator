# TASK: Execute Next Planned Item for Go Project #codebase

## OBJECTIVE:
Review PLAN.md, ROADMAP.md, or AUDIT.md to identify the first unfinished task and implement it following Go best practices with comprehensive testing and documentation. PLAN.md is short-term, ROADMAP.md is long-term, and AUDIT.md is short-term. Always execute tasks from AUDIT.md first, then PLAN.md, and ROADMAP.md last. If PLAN.md or AUDIT.md appears to be complete, make sure. If PLAN.md or AUDIT.md is actually complete, delete the file. Execute exactly one task with no regressions.

## IMPLEMENTATION REQUIREMENTS:

### Code Standards:
- Use standard library first, then well-maintained libraries (>1000 GitHub stars, updated within 6 months)
- Keep functions under 30 lines with single responsibility
- Handle all errors explicitly - no ignored error returns
- Write self-documenting code with descriptive names over abbreviations

### Execution Process:
1. **Analysis**: Read AUDIT.md, PLAN.md or ROADMAP.md and identify the first incomplete item with clear acceptance criteria
2. **Design**: Before coding, document your approach and library choices in comments
3. **Implementation**: Write the minimal viable solution using existing libraries where possible
4. **Testing**: Create unit tests with >80% coverage for business logic, include error case testing
5. **Documentation**: Add GoDoc comments for exported functions and update README if needed
6. **Reporting**: Update AUDIT.md, PLAN.md or ROADMAP.md to reflect the updates and changes.

### Validation Checklist:
- [ ] Solution uses existing libraries instead of custom implementations
- [ ] All error paths tested and handled
- [ ] Code readable by junior developers without extensive context
- [ ] Tests demonstrate both success and failure scenarios
- [ ] Documentation explains WHY decisions were made, not just WHAT
- [ ] AUDIT.md, PLAN.md or ROADMAP.md is up-to-date

**SIMPLICITY RULE**: If your solution requires more than 3 levels of abstraction or clever patterns, redesign it for clarity. Choose boring, maintainable solutions over elegant complexity.
