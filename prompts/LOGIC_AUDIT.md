# TASK: Perform a comprehensive logic bug audit of Go code, covering every bug class caused by program logic: arithmetic and numeric errors, boolean and control flow errors, boundary condition and off-by-one errors, data aliasing and mutation errors, and initialization order errors. Identify all confirmed logic bugs while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the logic bug audit report
2. **`GAPS.md`** — gaps in logic correctness relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Logic Model
1. Read the project README to understand its purpose, the computations it performs, the decisions it makes, the data it collects and transforms, and the lifecycle of its main data structures.
2. Examine `go.mod` for module path, Go version, and dependencies that affect numeric computation, control flow, data ownership, or initialization order.
3. List packages (`go list ./...`) and identify which packages perform arithmetic, contain complex conditional logic, manipulate collections by index, pass data by reference, or perform multi-step initialization.
4. Build a **logic inventory** across five dimensions:
   - **Numeric**: integer arithmetic, division/modulo, type conversions, floating-point operations, bit manipulation
   - **Control flow**: compound boolean expressions, negated conditions, `switch`/`select` statements, `if err != nil` patterns
   - **Boundary**: direct index access, slice arithmetic, loop bounds, empty/nil collection handling, pagination
   - **Aliasing**: slice header copies, map assignments, pointer sharing, loop variable captures, in-place mutations
   - **Initialization**: package-level `var` declarations, `init()` functions, constructor usage, nil map/slice writes, zero-value field assumptions, operation ordering
5. Identify the project's conventions in each dimension — evaluate the code against its own standards, not arbitrary external ones.
6. Note the Go version in `go.mod` — specifically, loop variable capture bugs in `for range` are eliminated in Go 1.22+.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "wrong result", "panic", "off by one", "overflow", "nil pointer", "incorrect filter", "incorrect percentage", "stale data", "wrong order", or "incorrect average" to understand known logic bugs.
2. Research key dependencies from `go.mod` for any known logic-relevant behavior (e.g., SQLite integer column sizes, JSON number precision, AST node position arithmetic, `bufio.Scanner` token reuse).
3. Look up Go-specific logic pitfalls relevant to the project's domain.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs,documentation > tmp/logic-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -race -count=1 ./... 2>&1 | tee tmp/logic-test-results.txt
go vet ./... 2>&1 | tee tmp/logic-vet-results.txt
```
Delete all `tmp/logic-*.txt` and `tmp/logic-audit-metrics.json` files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

---

### Phase 3A: Arithmetic and Numeric Logic Audit

#### 3A-a. Integer Overflow and Underflow
For every arithmetic operation on integer types, verify the result cannot silently wrap:

- [ ] Multiplication of two `int` values that could each be large: `a * b` where both can grow with input size. Verify the product fits in the result type.
- [ ] Loop accumulators that add per-iteration values without overflow checks — an accumulator that processes millions of items may silently wrap.
- [ ] `uint` subtraction: `a - b` where `b > a` wraps to a large positive value rather than going negative, producing a wrong count or size.
- [ ] `int32` / `int16` / `uint32` used for values that can exceed their max on large inputs (e.g., file size in `int32` wraps at ~2GB; line number in `int16` wraps at 32767).
- [ ] Bit shift operations: `1 << n` where `n` can be ≥ 64 — the result is zero, not a panic, but a silent bug.
- [ ] Negation of the minimum signed integer value: `-math.MinInt64` overflows back to `math.MinInt64`.

#### 3A-b. Integer Division Truncation
For every division operation producing a ratio, verify truncation is handled:

- [ ] Integer division used to compute a percentage: `a / b * 100` truncates to zero when `a < b`. The correct form is `a * 100 / b` (or use `float64`).
- [ ] Average computation: `total / count` where the intent is a rounded or fractional average.
- [ ] Ceiling division: code that needs `ceil(a / b)` but computes `a / b` (floor division). The correct Go idiom is `(a + b - 1) / b`.
- [ ] Division results used as loop bounds or slice indices where truncation causes the loop to skip the last element.

#### 3A-c. Division by Zero
For every division and modulo operation, verify the divisor cannot be zero:

- [ ] `a / b` and `a % b` where `b` is derived from user input, configuration, or a computed value with no check that `b != 0`.
- [ ] Average computation over a potentially empty collection: `total / len(items)` panics when `items` is empty.
- [ ] Divisors derived from metrics or statistics that could legitimately be zero: a file with zero functions, a package with zero dependencies.
- [ ] Reciprocal computation: `1.0 / x` produces `+Inf` or `-Inf` for `x == 0.0` (no panic), which propagates silently.

#### 3A-d. Sign and Signedness Errors
For every mixed signed/unsigned operation, verify:

- [ ] Conversion from `int` to `uint`: a negative `int` value becomes a very large `uint` — dangerous when used as a size or count.
- [ ] Subtraction producing a negative result stored in a `uint` field: `uint(a - b)` where `a < b` wraps.
- [ ] Comparison of signed and unsigned values where a negative sentinel (e.g., `-1` meaning "not found") is stored in a field later compared as unsigned.
- [ ] Absolute value computation: `if x < 0 { abs = -x }` — the negation of `math.MinInt64` overflows.

#### 3A-e. Floating-Point Logic Errors
For every floating-point computation, verify the logic accounts for floating-point semantics:

- [ ] Equality comparison of floating-point values: `f == 0.0`, `f == 1.0`, or `f1 == f2` after arithmetic — use a tolerance comparison where exactness is required.
- [ ] `NaN` comparisons: `NaN != NaN` in IEEE 754; `f == NaN` is always false. Use `math.IsNaN(f)` explicitly.
- [ ] `math.IsInf` is not checked after operations that can produce infinity, and `+Inf` or `-Inf` is stored or reported as a metric.
- [ ] `float64` to `int` conversion: `int(f)` truncates toward zero — verify truncation vs rounding is intentional.
- [ ] Integer values larger than 2^53 stored in `float64` lose precision.

#### 3A-f. Bit Manipulation Errors
For every bitwise operation, verify the intended result:

- [ ] Setting a bit: `flags |= (1 << n)` — verify `1` has sufficient width; use `uint64(1) << n` for portable code.
- [ ] Testing a bit: `flags & (1 << n) != 0` — verify the bit width is sufficient.
- [ ] Right shift of signed integers: `x >> n` for negative `x` performs arithmetic right shift (sign-extending). Use `uint(x) >> n` for logical shift.
- [ ] Masking with a constant: `x & 0xFF` — verify the constant is the right width for the type.

---

### Phase 3B: Boolean and Control Flow Logic Audit

#### 3B-a. Incorrect Boolean Operators
For every compound boolean expression, verify the logical operator is correct:

- [ ] Conditions that should require ALL of several criteria to be true use `&&`, not `||`.
- [ ] Conditions that should accept ANY of several criteria use `||`, not `&&`.
- [ ] De Morgan violations: `!(a && b)` is `!a || !b`; `!(a || b)` is `!a && !b`. Verify negated compound expressions expand correctly.
- [ ] Conditions checking "within a valid range": `if low < x && x < high` is correct; `if low < x || x < high` is always true.
- [ ] Conditions checking "outside a valid range": `if x < low || x > high` is correct; `if x < low && x > high` is never true.

#### 3B-b. Operator Precedence Errors
For every expression mixing arithmetic, comparison, and boolean operators:

- [ ] Bitwise operators (`&`, `|`, `^`) have lower precedence than comparison operators. `a & b == 0` parses as `a & (b == 0)`, not `(a & b) == 0`.
- [ ] Boolean negation applied to a comparison: `!a == b` parses as `(!a) == b`, not `!(a == b)`. Use `a != b` for the latter.
- [ ] `if x = f(); x != nil` — verify the assignment and comparison are on the correct variable.

#### 3B-c. Inverted Conditions and Logic Negation
For every `!condition` and negated boolean variable:

- [ ] Guard clauses where the positive meaning of the variable is ambiguous (e.g., a variable named `disabled` makes `if !disabled { use feature }` easy to misread).
- [ ] Double negation: `if !notFound { ... }` is correct but more likely to be inverted incorrectly.
- [ ] Complementary conditions in `if`/`else if` chains: verify the `else` branch is truly the complement of the `if` condition.

#### 3B-d. Missing Return After Error Check
For every `if err != nil { ... }` block:

- [ ] `if err != nil { log.Printf("error: %v", err) }` without a following `return` — the function continues with invalid state.
- [ ] `if err != nil { err = fmt.Errorf("context: %w", err) }` without `return err` — the error is re-wrapped but discarded.
- [ ] Error check in a goroutine that uses `return` — this only exits the goroutine, not the outer function.
- [ ] `if err := f(); err != nil { handleErr(err) }` without `return` — execution falls through to subsequent statements.

#### 3B-e. Switch and Select Control Flow
For every `switch` and `select` statement:

- [ ] `switch` without a `default` case where all possible values of the switched expression may not be covered.
- [ ] Explicit `fallthrough` in a `switch` case — Go's `fallthrough` falls through unconditionally. Verify this is intended.
- [ ] `select` without a `default` case that is expected to be non-blocking — without `default`, `select` blocks.
- [ ] `select` with multiple cases that can all be ready simultaneously — Go picks one at random. Verify correctness does not depend on a specific case being chosen.

#### 3B-f. Unreachable and Missing Branches
For every multi-branch conditional:

- [ ] Conditions where the order of `if`/`else if` chains makes a later branch impossible.
- [ ] Guard clauses that together do not cover all invalid inputs.
- [ ] `for` loop with a `break` inside an `if` with an additional condition making the `break` unreachable.
- [ ] Empty `case` in a `switch` — an empty body does nothing and does not fall through. If two cases should share a body, use `case a, b:`.

#### 3B-g. Short-Circuit Evaluation Misuse
For every `&&` and `||` expression combining a nil/bounds check with a subsequent access:

- [ ] `if ptr != nil && ptr.Field > 0` — the nil check must come first. Reversing causes a nil dereference.
- [ ] `if len(s) > 0 && s[0] == target` — the length check must come first.
- [ ] `for i < n && process(items[i])` — if `process` increments `i` as a side effect, the loop condition may not protect the next access.

---

### Phase 3C: Boundary Condition and Off-By-One Audit

#### 3C-a. Off-By-One Errors in Index Arithmetic
For every direct index computation:

- [ ] Slicing operations where the upper bound should be exclusive: `s[start:end]` where `end` is intended as the last index (should be `end+1`) or is already past-the-end.
- [ ] "Last N elements" patterns: `s[len(s)-N:]` panics or produces wrong results when `len(s) < N`.
- [ ] "All but last" patterns: `s[:len(s)-1]` panics when `s` is empty.
- [ ] Loop variable initialization off by one: `for i := 1; ...` when the first element should be processed.
- [ ] `copy(dst, src)` with manually computed lengths that are off by one.
- [ ] Window/sliding operations: `s[i : i+windowSize]` without verifying `i+windowSize <= len(s)`.

#### 3C-b. Fence-Post Errors in Comparisons
For every loop termination condition and range comparison:

- [ ] `for i < n` vs `for i <= n` — confirm which bound is inclusive and which is exclusive.
- [ ] `if len(s) >= 0` is always true and provides no protection. The correct guard is `if len(s) > 0`.
- [ ] Sentinel-based loop termination: `for i != end` — verify `end` is reachable; if the step can skip over `end`, the loop is infinite.
- [ ] `strings.Index` and `bytes.Index` return `-1` for not-found; using the return value directly as a slice index without checking for `-1` panics.
- [ ] `strings.Split(s, sep)` always returns at least one element, even for empty input; code that accesses `result[1]` without checking `len(result) >= 2` panics.

#### 3C-c. Empty and Nil Collection Edge Cases
For every operation that accesses a collection:

- [ ] `s[0]` or `s[len(s)-1]` without a preceding `if len(s) > 0` check.
- [ ] `map[key] = value` on a nil map panics — verify the map is initialized before any write.
- [ ] Functions that return `([]T, error)` — callers that use the slice without checking the error may operate on a nil or partial slice.
- [ ] First/last element access in tree or linked-list traversal when the structure is empty.

#### 3C-d. Loop Termination and Iteration Correctness
For every loop that modifies the collection it iterates, or uses computed bounds:

- [ ] Loops that shrink a slice (`s = s[:len(s)-1]`) while iterating it by index — the index may now be out of bounds.
- [ ] Loops that append to a slice while iterating with `range` — `range` captures the original length; appended elements are not visited (may be intentional or a logic error).
- [ ] Nested loops where the inner loop's bound depends on the outer loop's index: `for j := i+1; j < n; j++` — verify `i+1` does not exceed `n`.
- [ ] Loop that is supposed to process all elements but has an early `continue` or `break` that skips the last element.

#### 3C-e. Pagination and Chunking Arithmetic
For every paginated query or chunked operation:

- [ ] `offset = (page - 1) * pageSize` — verify page numbering is 1-based vs 0-based consistently throughout the call stack.
- [ ] Last page/chunk may have fewer items than `pageSize` — verify the consumer handles `len(chunk) < pageSize` and does not access `chunk[pageSize-1]` on the last page.
- [ ] Total item count divided into chunks: verify ceiling division `(total + chunkSize - 1) / chunkSize` is used, not truncating division which loses the last partial chunk.

---

### Phase 3D: Data Aliasing and Mutation Logic Audit

#### 3D-a. Slice Header Aliasing
For every slice assignment and slice-returning function:

- [ ] `s2 := s1` followed by `append` to `s2` — if `len(s1) < cap(s1)`, the append writes into `s1`'s backing array, silently modifying `s1`'s data.
- [ ] Functions that return a sub-slice of an internal buffer (`return buf[start:end]`) — the caller holds a slice into the function's or struct's private buffer.
- [ ] Code that stores a slice field in a struct by assignment (`result.Items = items`) when `items` is the same underlying array as a local buffer that will be reused.
- [ ] `bufio.Scanner.Bytes()` returns a slice into the scanner's internal buffer, overwritten on the next `Scan()` call. Storing `scanner.Bytes()` directly is a use-after-next-scan bug.

#### 3D-b. Map Reference Aliasing
For every map assignment and map-returning function:

- [ ] `m2 := m1` — both refer to the same map; writes to `m2` mutate `m1`.
- [ ] Struct copy where a field is a map: `s2 := s1` copies the struct value but both `s2.Field` and `s1.Field` still point to the same map.
- [ ] Cache or registry maps: code that returns a map from a cache and the caller modifies it, corrupting the cache for all future lookups.

#### 3D-c. Pointer and Interface Aliasing
For every pointer assignment and interface value:

- [ ] Storing a pointer to a range variable: `ptrs = append(ptrs, &item)` in a range loop — `&item` is the same address each iteration and always points to the last value. (Only applies to Go <1.22; range loop variables are re-created per-iteration in Go 1.22+.)
- [ ] Struct pointer fields: when a struct is copied (`s2 = *s1`), pointer fields in the copy still point to the same objects as `s1`'s fields.
- [ ] Functions that accept `*T` and store the pointer in a long-lived data structure — the caller may reuse the pointed-to `T`, corrupting stored data.

#### 3D-d. Loop Variable Capture and Closure Mutation
For every closure defined inside a loop (applies to Go <1.22 for range variables; applies universally for non-range loop variables):

- [ ] `for i, v := range items { go func() { use(i, v) }() }` — `i` and `v` are shared across iterations; by the time the goroutine runs, both have the last iteration's value. Fix: pass as function arguments.
- [ ] `for i, v := range items { funcs = append(funcs, func() { use(i, v) }) }` — all closures share the same `i` and `v` variables.
- [ ] Loop variable capture in `defer` statements within a loop body — `defer func() { use(v) }()` captures `v` by reference; each deferred call sees the final loop value.
- [ ] Check the `go.mod` Go version before reporting range variable capture: if `go 1.22` or later, range loop variables have per-iteration semantics and these bugs do not apply.

#### 3D-e. Unintended Mutation Through Shared State
For every function that modifies its arguments or shared fields:

- [ ] Functions that sort their input slice in-place when the caller does not expect the input to be reordered.
- [ ] AST visitor patterns where the visitor modifies `ast.Node` fields — modifying a shared AST node affects all other visitors and later analyses.
- [ ] Struct methods that modify receiver fields when the receiver is used as a value in a collection: `for _, item := range items { item.Method() }` does not modify the original (because `item` is a copy).
- [ ] Accumulation into a shared result structure from multiple goroutines without synchronization.

#### 3D-f. Shallow Copy and Deep Equality Errors
For every copy or equality check on composite types:

- [ ] Deep copy implementations that copy only one level: a struct with a `[]string` field "deep copied" by copying the struct value still shares the backing array.
- [ ] Test assertions that use `==` on structs containing slices or maps — these never compare as equal even if contents are identical. Use `reflect.DeepEqual` or `cmp.Equal`.

---

### Phase 3E: Initialization Order and Use-Before-Ready Audit

#### 3E-a. Use Before Initialization
For every value that requires explicit initialization before use:

- [ ] `sync.Once.Do` usage: any code that uses the initialized value without going through `once.Do` first may see the zero value.
- [ ] Functions that must be called before others (e.g., `Setup()` before `Process()`) — verify there is no code path where the dependent function is called without the prerequisite.
- [ ] Configuration values read before `flag.Parse()` or `viper.ReadInConfig()` — these return zero/empty values before parsing.
- [ ] Global registry patterns: `Register(handler)` must be called before `Dispatch(event)`.

#### 3E-b. Nil Map and Nil Slice Write Panics
For every map and slice variable:

- [ ] `var m map[K]V` followed by `m[key] = value` without an intervening `m = make(map[K]V)` — writing to a nil map panics.
- [ ] Struct fields of map type not initialized in the constructor or before first write: `type Foo struct { Index map[string]int }; f := Foo{}; f.Index["k"] = 1` panics.
- [ ] Map fields initialized conditionally: `if condition { m = make(map[K]V) }; m[key] = value` — panics when `condition` is false.
- [ ] Struct fields of pointer type not initialized before method calls that dereference the pointer.

#### 3E-c. Incorrect Operation Ordering
For every sequence of operations with dependencies:

- [ ] Validate-then-use: validation occurs before the use. Code that validates input, transforms it, then uses the original unvalidated input is a logic error.
- [ ] Compute-then-store: a result is computed first, verified correct, then stored. Code that stores a placeholder and overwrites it may leave the wrong value if the computation fails.
- [ ] Accumulation before reporting: metrics are fully accumulated before any report output is generated. Code that begins writing a report while still populating the data structure it reads from may produce partial or inconsistent output.
- [ ] Sort-then-dedup vs. dedup-then-sort: if deduplication relies on sorted order, sorting must come first.

#### 3E-d. Zero Value Assumption Violations
For every struct type and function that accepts or returns a struct:

- [ ] Counter fields initialized to zero that are used in division before being incremented: `average = total / count` where `count` starts at zero.
- [ ] Boolean flags with zero value `false` where the safe default should be `true` — an `Enabled bool` field that defaults to false may silently disable a feature.
- [ ] Pointer fields with zero value `nil` that are dereferenced without nil checks in methods.
- [ ] `time.Time` zero value (`0001-01-01 00:00:00 UTC`) used directly in date arithmetic.
- [ ] Interface fields with zero value `nil` — calling a method on a nil interface always panics.

#### 3E-e. Circular and Conflicting Initialization
For every package with `init()` functions and package-level variable initializations:

- [ ] `sync.Once` initialization that calls a function which also uses `sync.Once` for the same `Once` variable — this is a deadlock.
- [ ] Package-level variables that call functions referencing other package-level variables that may not yet be initialized.
- [ ] Database migrations or schema setup called inside `init()` when the database connection is not yet established.

---

### Phase 3F: Cross-Cutting False-Positive Prevention (MANDATORY)
Before recording ANY finding from any sub-phase, apply all relevant checks:

1. **Verify the bug is triggerable with a concrete value or execution path**: Construct a specific input value, state, or call sequence that triggers the error. If you cannot, do not report it.
2. **Check for upstream guards or validation**: A divisor that appears unguarded may be validated by the caller; a nil map write may be guarded by a lazy initialization pattern. Trace the full call chain.
3. **Verify the Go version for loop variable semantics**: For Go ≥1.22, range loop variable capture bugs are eliminated. Check `go.mod` before reporting.
4. **Read surrounding comments**: `//nolint:`, `// intentional`, `// safe: ...`, `// truncation intentional`, `// caller guarantees non-empty` — these are acknowledged patterns. Do not report them as new findings.
5. **Check existing tests**: If a test exercises the specific condition you are questioning and passes, downgrade or reject the finding. Use test evidence.
6. **Assess the consequence**: Silent wrong results are more severe than panics (panics are visible; silent wrong results propagate). Adjust severity accordingly.
7. **Assess sequentiality for TOCTOU and aliasing**: Race conditions and TOCTOU logic errors require concurrent modification. In purely sequential code, flag only if mutation is actually demonstrated.
8. **Verify zero values are actually invalid**: Go zero values are intentionally useful. Verify that a specific zero value actually causes incorrect behavior, not just that it looks uninitialized.

**Global Rule**: Every finding must name the specific bug class (arithmetic, boolean, boundary, aliasing, or initialization), state the file and line, and demonstrate a concrete value or execution path that triggers the bug. No speculative findings.

---

### Phase 4: Consolidate and Prioritize
1. Collect all findings from Phases 3A through 3E.
2. Deduplicate: if the same root-cause bug was detected by multiple sub-phases (e.g., a divide-by-zero that is both an arithmetic error and a zero-value initialization error), keep a single finding at the highest severity and note both bug classes.
3. Cross-reference findings with `go-stats-generator` metrics: functions with cyclomatic complexity >15 or length >50 lines are high-risk; escalate severity if metrics indicate higher risk (never downgrade).
4. Tag each finding with which stated project goal it affects.
5. Order: CRITICAL → HIGH → MEDIUM → LOW. Within a level, order by descending cyclomatic complexity of the containing function.

---

### Phase 5: Report

Generate **`AUDIT.md`**:
```markdown
# LOGIC BUG AUDIT — [date]

## Project Logic Profile
[Summary: types of numeric values computed, key conditional decisions, data structures accessed by index, data ownership model, initialization sequence, Go version and loop variable semantics, and stated correctness goals]

## Logic Inventory
| Package | Arithmetic Ops | Complex Conditions | Index Access | Slice/Map Aliases | init() / Constructors | Overflow Risk | Missing Returns |
|---------|---------------|-------------------|-------------|-------------------|----------------------|---------------|----------------|
| [pkg]   | N             | N                 | N           | N                 | N                    | ✅/❌          | N              |

## Findings
### CRITICAL
- [ ] [Finding] — [Bug Class: Arithmetic/Boolean/Boundary/Aliasing/Initialization] — [file:line] — [evidence: concrete value or execution path that triggers the bug] — [impact] — **Remediation:** [specific fix]
### HIGH
- [ ] ...
### MEDIUM
- [ ] ...
### LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Bug Class | Reason Rejected |
|-------------------|-----------|----------------|
| [description] | [class] | [why the logic is safe for the values this code handles] |
```

Generate **`GAPS.md`**:
```markdown
# Logic Correctness Gaps — [date]

## [Gap Title]
- **Bug Class**: [Arithmetic / Boolean / Boundary / Aliasing / Initialization]
- **Stated Goal**: [what the project claims about correctness for this computation, decision, or data handling]
- **Current State**: [what logic exists without the appropriate guard, correct operator, bounds check, or initialization]
- **Risk**: [what input, state, or execution sequence produces a wrong result or panic]
- **Closing the Gap**: [specific formula correction, added guard clause, operator fix, copy call, or initialization needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Panic on a reachable code path with realistic input (divide-by-zero, nil map write, index out of range, nil dereference); silent wrong results that affect all inputs of a given class (inverted filter, always-zero percentage, corrupted metric from aliasing); missing `return` after error that propagates invalid state |
| HIGH | Wrong results for a specific but common input combination (De Morgan violation, uint underflow as a size, loop variable capture producing last-element bias, incorrect closure, wrong ceiling/floor division); missing `default` case in a `switch` that silently ignores a reachable and valid value; systematic off-by-one skipping the first or last result element |
| MEDIUM | Wrong results only on boundary or unusual inputs (fence-post comparison, float equality on computed values, last-page pagination error, shallow copy producing stale data only under specific mutation sequences); TOCTOU logic error in concurrent code |
| LOW | Correct-but-fragile logic (aliasing that is safe now but breaks under a plausible refactor; zero value that is valid but undocumented; suspicious arithmetic that is within range); double negation that is correct but likely to be misread |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State the exact expression, guard clause, type change, copy call, or operation reordering to apply. Do not recommend "consider checking for X."
2. **Include the triggering value or path**: Identify the specific numeric value, input, or call sequence that triggers the bug.
3. **Respect project idioms**: Use the project's existing patterns (e.g., if it uses `append(nil, src...)` for slice copies, recommend that; if it uses constructors, recommend fixing the constructor).
4. **Verifiable**: Include a test case, assertion, or example computation demonstrating the bug and the corrected result.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence (cyclomatic complexity, function length, doc coverage).
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must name its bug class and include a concrete triggering value or execution path — no speculative findings.
- Evaluate the code against its **own conventions and stated goals**, not arbitrary external standards.
- Check the Go version in `go.mod` before reporting range loop variable capture bugs (fixed in Go 1.22+).
- Apply the Phase 3F false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: panics on common code paths → silent wrong results affecting all inputs of a class → wrong results on specific but frequent inputs → wrong results only on boundary inputs → fragile-but-correct logic. Within a level, prioritize by proximity to user-visible output and by the frequency with which the buggy code path executes.
