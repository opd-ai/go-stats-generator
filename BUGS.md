TASK: Analyze the Go codebase, identify bugs, and autonomously apply fixes.

SCOPE: Fix these bug categories:
1. Runtime errors (nil pointers, race conditions, deadlocks)
2. Resource leaks (unclosed files, goroutine leaks, memory leaks)
3. Logic errors (incorrect conditionals, off-by-one errors, improper error handling)
4. Concurrency issues (unsafe concurrent access, missing synchronization)
5. Security vulnerabilities (injection risks, unsafe input handling)

EXECUTION APPROACH:
1. Scan all .go files and identify bugs
2. For each bug: determine fix, apply changes, verify syntax
3. Prioritize fixes: CRITICAL > HIGH > MEDIUM > LOW
4. Skip fixes that require architectural changes or unclear requirements

CONSTRAINTS:
- Only fix bugs with clear, deterministic solutions
- Preserve existing functionality and API contracts
- Maintain code style consistency with surrounding code
- Do not modify comments, imports, or formatting unless necessary for the fix

OUTPUT FORMAT:
For each file modified, show:

**Fixed: `path/to/file.go`**
- Line X: [SEVERITY] [Brief issue description]
  - Applied: [Specific change made]

End with:
SUMMARY: Fixed X bugs across Y files (Critical: X, High: X, Medium: X, Low: X)

VERIFICATION:
- Ensure all modified files remain valid Go syntax
- Confirm fixes don't introduce new issues
- Flag any bugs that require manual review with: ⚠️ MANUAL REVIEW NEEDED
