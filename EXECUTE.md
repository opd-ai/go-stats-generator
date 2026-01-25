# TASK: Execute Next Planned Item for Go Project #codebase

## OBJECTIVE:
Review PLAN.md, ROADMAP.md, or AUDIT.md to identify unfinished task(s) and implement them following Go best practices with comprehensive testing and documentation. AUDIT.md is for immediate action. if AUDIT.md exists, ignore PLAN.md and ROADMAP.md entirely. PLAN.md is medium-term, ROADMAP.md is long-term, and AUDIT.md is short-term. Always execute tasks from AUDIT.md first, then PLAN.md, and ROADMAP.md last. If PLAN.md or AUDIT.md appears to be complete, make sure. If PLAN.md or AUDIT.md is actually complete, delete the file. Execute one task, or optionally 2-3 closely related tasks together, with no regressions. Work by completing incomplete implementations. Prioritize important tasks before resorting to trivial tasks. Major tasks are in-scope. We should always be working toward a finished product. ALWAYS proceed in order, first AUDIT.md, second PLAN.md, third ROADMAP.md, never skip. Never proceed out of order. Never execute a PLAN.md task before an AUDIT.md task. Tasks completed out-of-order will be summarily rejected without consideration. ALWAYS choose items from AUDIT.md, if it exists, no matter what.

Do not second guess the order of the files. The humans put them in that order deliberately. If you think these files are misplaced, you're wrong. Go fuck yourself, then do the thing in the order the humans told you to.

### Grouping Related Tasks:
You may execute 1-3 tasks together if they meet ALL of these criteria:
- **Same Component**: Tasks affect the same file, function, or module
- **Shared Context**: Changes require understanding the same code area
- **Dependent Changes**: Completing one task naturally leads to or enables the next
- **Similar Scope**: Each task is small and together they stay under 500lines of changes
- **Common Testing**: The tasks can be validated with a shared test suite

**Examples of Related Tasks:**
- Adding 2-3 similar struct fields and their getter/setter methods
- Implementing multiple related interface methods for the same type
- Fixing multiple validation issues in the same function

**Examples of Unrelated Tasks (execute separately):**
- Changes to different packages or unrelated modules
- Tasks requiring different testing strategies
- Large refactoring combined with new feature development

## IMPLEMENTATION REQUIREMENTS:

### Code Standards:
- Use standard library first, then well-maintained libraries (>1000 GitHub stars, updated within 6 months)
- Keep functions under 30 lines with single responsibility
- Handle all errors explicitly - no ignored error returns
- Write self-documenting code with descriptive names over abbreviations

### Execution Process:
1. **Analysis**: Read AUDIT.md, PLAN.md or ROADMAP.md and identify incomplete item(s) with clear acceptance criteria. If multiple tasks are closely related (see criteria above), you may group 2-3 together.
2. **Design**: Before coding, document your approach and library choices in comments. For grouped tasks, ensure they share a coherent design.
3. **Implementation**: Write the minimal viable solution using existing libraries where possible. For grouped tasks, implement them in logical order.
4. **Testing**: Create unit tests with >80% coverage for business logic, include error case testing. For grouped tasks, ensure comprehensive test coverage across all changes.
5. **Documentation**: Add GoDoc comments for exported functions and update README if needed
6. **Reporting**: Update AUDIT.md, PLAN.md or ROADMAP.md to reflect all completed tasks and changes.

### Validation Checklist:
- [ ] Solution uses existing libraries instead of custom implementations
- [ ] All error paths tested and handled
- [ ] Code readable by junior developers without extensive context
- [ ] Tests demonstrate both success and failure scenarios
- [ ] Documentation explains WHY decisions were made, not just WHAT
- [ ] AUDIT.md, PLAN.md or ROADMAP.md is up-to-date with all completed task(s)
- [ ] If multiple tasks were grouped: they meet all the "closely related" criteria
- [ ] If multiple tasks were grouped: total changes stay focused and under 500lines

**SIMPLICITY RULE**: If your solution requires more than 3 levels of abstraction or clever patterns, redesign it for clarity. Choose boring, maintainable solutions over elegant complexity.
