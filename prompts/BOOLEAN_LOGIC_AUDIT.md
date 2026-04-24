# TASK: Perform a focused audit of boolean and control flow logic errors in Go code, identifying incorrect logical operators, De Morgan violations, operator precedence bugs, switch/select fallthrough omissions, missing branches causing unintended execution, and short-circuit evaluation misuse while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the boolean and control flow logic audit report
2. **`GAPS.md`** — gaps in control flow correctness relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Decision Logic
1. Read the project README to understand its purpose, the decisions it makes (which files to analyze, which metrics to compute, which findings to report), and the conditions that govern them.
2. Examine `go.mod` for module path, Go version, and dependencies that introduce decision logic (filter functions, condition evaluators, rule engines).
3. List packages (`go list ./...`) and identify which packages contain complex conditional logic, filtering, classification, or state machine logic.
4. Build a **control flow inventory** by scanning for:
   - Complex boolean expressions: `&&`, `||`, `!` with three or more operands
   - Negated conditions: `!condition` especially when the condition is itself a compound expression
   - `switch` statements: cases without explicit `break` (Go does not fall through by default), explicit `fallthrough` statements
   - `select` statements: cases for channel operations, especially `default` cases
   - Nested `if`/`else if`/`else` chains with many branches
   - `if err != nil` followed by code that continues to execute — missing `return` after error check
   - Early return patterns: `if !valid { return }` — verify the complementary condition is the correct inverse
   - Guard clause patterns: multiple early returns that should collectively cover all invalid input cases
   - Ternary-like patterns: Go has no ternary operator; code that simulates it may have logic errors
   - `for` loops with complex conditions or multiple `break`/`continue` paths
5. Identify the project's conditional logic conventions — does it use guard clauses? Does it use positive or negative conditions? Does it use complex multi-condition `if` statements?
6. Map which boolean expressions determine what gets included or excluded from reports — a logic inversion here silently omits or duplicates findings.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "wrong filter", "missing results", "duplicate results", "always true", "always false", "never reached", or "incorrect condition" to understand known logic bugs.
2. Research key dependencies from `go.mod` for any documented behavior that depends on boolean logic (e.g., filter interfaces, predicate functions, rule matching).
3. Look up Go-specific control flow pitfalls (e.g., `switch` with no expression, `select` without `default`, `for` with complex multi-value conditions).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's decision logic.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > /tmp/boolean-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -count=1 ./... 2>&1 | tee /tmp/boolean-test-results.txt
go vet ./... 2>&1 | tee /tmp/boolean-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Boolean and Control Flow Logic Audit

#### 3a. Incorrect Boolean Operators
For every compound boolean expression, verify the logical operator (`&&` vs `||`) is correct:

- [ ] Conditions that should require ALL of several criteria to be true use `&&`, not `||`: `if a && b && c { /* all three required */ }`.
- [ ] Conditions that should accept ANY of several criteria use `||`, not `&&`: `if a || b || c { /* any one sufficient */ }`.
- [ ] De Morgan violations: negating a compound expression incorrectly. `!(a && b)` is `!a || !b`; `!(a || b)` is `!a && !b`. A common mistake is writing `!a && !b` when the intent is `!(a || b)`, or `!a || !b` when the intent is `!(a && b)`.
- [ ] Conditions checking a value is "within a valid range": `if low < x && x < high` is correct; `if low < x || x < high` is always true for any `x` (every number is either greater than `low` or less than `high`).
- [ ] Conditions checking a value is "outside a valid range": `if x < low || x > high` is correct; `if x < low && x > high` is never true.
- [ ] Boolean expressions where one sub-expression is always true or always false due to type constraints (e.g., `uint >= 0` is always true; `uint < 0` is always false — Go reports these as compilation errors, but similar patterns with implicit conversions may slip through).

#### 3b. Operator Precedence Errors
For every expression mixing arithmetic, comparison, and boolean operators, verify precedence is explicit:

- [ ] Bitwise operators (`&`, `|`, `^`) have lower precedence than comparison operators (`==`, `!=`, `<`, `>`), but many developers expect the reverse. `a & b == 0` parses as `a & (b == 0)`, not `(a & b) == 0`. This is a frequent source of bugs in flag testing.
- [ ] Mixed arithmetic and comparison: `a + b > c + d` is unambiguous in Go, but `a | b > c` parses as `a | (b > c)` due to precedence. Use explicit parentheses when mixing bitwise and comparison operators.
- [ ] Boolean negation applied to a comparison: `!a == b` parses as `(!a) == b`, not `!(a == b)`. Use `a != b` for the latter.
- [ ] String and range comparisons: `a < b && b < c` is correct; ensure the middle value `b` is not accidentally repeated (`a < b && a < c` misses the upper bound on `b`).
- [ ] `if x = f(); x != nil` — the assignment form of `if` is valid in Go; verify the assignment and comparison are on the correct variable and the logic is not accidentally `if x := f(); y != nil` comparing the wrong variable.

#### 3c. Inverted Conditions and Logic Negation
For every `!condition` and negated boolean variable, verify the inversion is correct:

- [ ] `if !valid { return err }` guard clauses — verify the positive meaning of `valid` is what it appears. A variable named `disabled` that means "the feature is disabled" makes `if !disabled { use feature }` easy to misread or misname.
- [ ] Double negation: `if !notFound { ... }` is equivalent to `if found { ... }` but is harder to reason about and more likely to be inverted incorrectly.
- [ ] Boolean function names where the positive form is counterintuitive: `IsEmpty()` returning `true` for a non-empty collection, or `ShouldSkip()` returning `false` when the item should be skipped.
- [ ] Complementary conditions in `if`/`else if` chains: verify that the `else` branch is truly the complement of the `if` condition, not a different (possibly overlapping or gapped) condition.
- [ ] Negated compound conditions that gate important logic: `if !(a && b) { skip }` — if this should only skip when both `a` and `b` are false (not just one), the correct condition is `if !a || !b { skip }` — which is what `!(a && b)` means. Verify the intent matches the De Morgan expansion.

#### 3d. Missing Return After Error Check
For every `if err != nil { ... }` block, verify the function returns or panics on the error path:

- [ ] `if err != nil { log.Printf("error: %v", err) }` without a following `return` — the function continues executing with an invalid state after logging the error.
- [ ] `if err != nil { err = fmt.Errorf("context: %w", err) }` without `return err` — the error is re-wrapped but then discarded; the function continues as if no error occurred.
- [ ] Error checks in the middle of a function where the zero value of a subsequent local variable would cause incorrect behavior if execution continues: `x, err := f(); if err != nil { logError(err) }; use(x)` — `x` is the zero value after the error.
- [ ] Error check in a goroutine that uses `return` — this only exits the goroutine, not the outer function. The outer function may continue executing assuming success.
- [ ] `if err := f(); err != nil { handleErr(err) }` without `return` — the `handleErr` call may not exit; execution falls through to subsequent statements.

#### 3e. Switch and Select Control Flow
For every `switch` and `select` statement, verify all cases and fallthrough behavior:

- [ ] `switch` without a `default` case — verify that all possible values of the switched expression are covered. A missing case for an enum-like value silently does nothing.
- [ ] Explicit `fallthrough` in a `switch` case — Go's `fallthrough` falls through to the next case's body unconditionally, without re-evaluating the next case's expression. Verify this unconditional fallthrough is intended.
- [ ] `switch` on an interface type or `any` value without a `default` case — new types implementing the interface are silently ignored.
- [ ] `select` without a `default` case that is expected to be non-blocking — without `default`, `select` blocks until a case is ready; with `default`, it returns immediately. Verify the presence or absence of `default` matches the intent.
- [ ] `select` with multiple cases that can all be ready simultaneously — Go picks one at random. Verify the algorithm's correctness does not depend on a specific case being chosen (priority select must be implemented explicitly).
- [ ] `switch` on a boolean expression with only `case true` — the `case false` or `default` is implicitly a no-op. Verify this is intentional and not a missing branch.

#### 3f. Unreachable and Missing Branches
For every multi-branch conditional, verify all intended cases are covered and no branches are unreachable:

- [ ] `if a { ... } else if a { ... }` — the second branch is unreachable because `a` was already false when reaching the `else if`.
- [ ] Conditions where the order of `if`/`else if` chains makes a later branch impossible: `if x > 10 { ... } else if x > 5 { ... } else if x > 8 { ... }` — the third branch is unreachable because `x > 8` is impossible when `x <= 5`.
- [ ] Guard clauses that together do not cover all invalid inputs: a function with `if x < 0 { return errNegative }` but no check for `x == 0` when zero is also invalid.
- [ ] `for` loop with a `break` inside an `if` that has an additional condition making the `break` unreachable: `for { if err != nil && alwaysTrue { break } }` — if `alwaysTrue` is a constant, the `break` always executes or never executes.
- [ ] Functions that return without a value on some paths in a function declared to return a value — the Go compiler catches most of these, but complex `switch` or `if` chains may have paths that the compiler considers covered (via `return` in `default`) but that are logically impossible.
- [ ] Empty `case` in a `switch` — in Go, an empty case body does nothing and does not fall through. If the intent was to handle the case identically to the next case, use a comma-separated case list: `case a, b:`.

#### 3g. Short-Circuit Evaluation Misuse
For every `&&` and `||` expression that combines a nil/bounds check with a subsequent access, verify the order is correct:

- [ ] `if ptr != nil && ptr.Field > 0` — the nil check must come first. Reversing to `if ptr.Field > 0 && ptr != nil` causes a nil dereference when `ptr` is nil.
- [ ] `if len(s) > 0 && s[0] == target` — the length check must come first. Reversing panics on empty slices.
- [ ] `if err == nil || fallbackOK` — if `err == nil` is true, `fallbackOK` is not evaluated (short-circuit). Verify `fallbackOK` has no required side effects that must execute when `err == nil`.
- [ ] `if expensiveCheck() || cheapCheck()` — the expensive check runs first even though either condition suffices. This is a performance issue, not a correctness bug (unless `expensiveCheck` has required side effects). However, if the order matters for correctness (e.g., `expensiveCheck` modifies state that `cheapCheck` reads), it is a logic error.
- [ ] `for i < n && process(items[i])` — `process` is called with `i < n` verified, but if `process` increments `i` as a side effect, the loop condition may not protect the next iteration's `items[i]` access correctly.

#### 3h. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the condition actually fires incorrectly**: Construct a concrete test case where the boolean expression produces a different result than the correct logic requires. If you cannot construct one, do not report it.
2. **Check the semantics of the variables**: A variable named `skip` set to `true` when items should be skipped makes `if skip { continue }` correct. Read the variable's documentation or usage context before inferring its semantics from the name alone.
3. **Verify the switch/select exhaustiveness**: Before flagging a missing `default` in a `switch`, verify whether the switched value can actually take a value not covered by the cases. If the type is a defined type with a fixed set of constants, and all constants are covered, there is no bug.
4. **Read surrounding comments**: If a comment explicitly explains the control flow decision (e.g., `// intentional fallthrough`, `// only valid states reach here`, `//nolint:`, or a TODO), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Check existing tests**: If a test for the specific condition exists and passes, the condition may be correct. Verify that the test exercises the specific boolean sub-expression you are questioning.
6. **Assess the De Morgan application**: Before reporting a De Morgan violation, write out both the original and the De Morgan expansion explicitly and verify they are actually different. It is easy to make an error in the expansion itself.

**Rule**: If you cannot construct a concrete input where the boolean expression produces an incorrect result (wrong branch taken, wrong item filtered, wrong action performed), do NOT report it. Logic findings require a specific counterexample.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# BOOLEAN AND CONTROL FLOW LOGIC AUDIT — [date]

## Project Decision Logic Profile
[Summary: key filtering, classification, and branching decisions in the codebase; conditional complexity; use of negation and compound conditions; Go version and relevant language features]

## Control Flow Inventory
| Package | Complex Conditions (3+ operands) | Negated Conditions | Switch Statements | Select Statements | Missing Returns After Error |
|---------|----------------------------------|-------------------|-------------------|-------------------|-----------------------------|
| [pkg]   | N                                | N                 | N                 | N                 | N                           |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: concrete input that triggers wrong branch] — [impact: what is silently omitted or incorrectly included] — **Remediation:** [specific fix with corrected expression]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why the logic is correct for the values this code handles] |
```

Generate **`GAPS.md`**:
```markdown
# Boolean and Control Flow Logic Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims its filtering, classification, or branching logic should do]
- **Current State**: [what the boolean expression or control flow actually does]
- **Risk**: [what input causes incorrect behavior: wrong branch, missing case, silent skip]
- **Closing the Gap**: [specific expression correction, added case, or added return needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Logic error that silently produces wrong results for all inputs of a given class (e.g., a filter always includes what it should exclude, or always excludes what it should include); missing `return` after error that causes invalid state propagation |
| HIGH | Missing `default` case that silently ignores a valid and reachable value; incorrect `&&`/`||` that produces wrong results for a specific but common input combination; unreachable branch that represents missing functionality |
| MEDIUM | De Morgan violation that only manifests on unusual input combinations; operator precedence error that is accidentally masked by typical input values; `fallthrough` used incorrectly but only affecting uncommon cases |
| LOW | Double negation that is correct but likely to be misread; missing `default` case where all reachable values are covered; short-circuit order that is correct but executes an expensive check unnecessarily |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Corrected expression**: Write the exact corrected boolean expression or control flow structure. Do not recommend "fix the logic" — show the fix.
2. **Include the counterexample**: Identify the specific input value or condition combination that triggers the wrong behavior.
3. **Verify with De Morgan**: For any boolean restructuring, include the De Morgan expansion to confirm the corrected expression is equivalent to the intent.
4. **Verifiable**: Include a test case with the specific counterexample input that fails before the fix and passes after.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for function complexity and cyclomatic complexity (high cyclomatic complexity functions warrant closer inspection of their branching logic).
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete counterexample input demonstrating the logic error — no speculative findings.
- Evaluate the code against its **own stated logic** and the conditions it is supposed to represent, not arbitrary external standards.
- Apply the Phase 3h false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: logic errors causing wrong results for all inputs of a given class → missing return causing invalid state continuation → errors affecting only specific input combinations → unreachable branches representing missing functionality → confusing-but-correct logic. Within a level, prioritize by visibility of the wrong result to users and by frequency of the triggering input.
