# TASK: Perform a focused audit of arithmetic and numeric logic errors in Go code, identifying integer overflow, incorrect division, sign handling bugs, precision loss, bit manipulation errors, and incorrect numeric comparisons while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the arithmetic logic audit report
2. **`GAPS.md`** — gaps in numeric correctness relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Numeric Domain
1. Read the project README to understand its purpose and what numeric values it computes (line counts, complexity scores, percentages, ratios, offsets, sizes).
2. Examine `go.mod` for module path, Go version, and dependencies that involve numeric computation or fixed-size encoding.
3. List packages (`go list ./...`) and identify which packages perform arithmetic on user-supplied or unbounded values.
4. Build a **numeric inventory** by scanning for:
   - Integer arithmetic on potentially large values: multiplication, addition in loop accumulators, `*` and `+` on `int`, `int32`, `int64`, `uint`, `uint32`, `uint64`
   - Division and modulo: `/` and `%` operators, especially with user-controlled divisors (divide-by-zero risk)
   - Mixed-type arithmetic: `int` * `int64`, `float64` + `int`, `uint` - `int`
   - Explicit numeric type conversions: `int(x)`, `int32(x)`, `uint(x)`, `float64(x)`
   - Bit operations: `<<`, `>>`, `&`, `|`, `^`, `&^`
   - Percentage/ratio computations: `a / b * 100` vs `a * 100 / b` (integer truncation order)
   - Floating-point comparisons: `f == 0.0`, `f == 1.0`, `f1 == f2`
   - Conversions between `float64` and integer types for values that may exceed the integer range
   - Use of `math/big` or `math` package functions that return special values (`NaN`, `+Inf`, `-Inf`)
5. Identify the project's numeric conventions — does it check for overflow? Does it use `math/big` for large numbers? Does it convert to `float64` for averages?
6. Map which functions produce numeric values consumed by others as sizes, counts, or indices — a correct-looking consumer may rely on an overflowed producer.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "wrong count", "negative value", "overflow", "NaN", "divide by zero", "incorrect percentage", "incorrect average", or "wrong metric" to understand known numeric bugs.
2. Research key dependencies from `go.mod` for any known numeric behavior (e.g., SQLite integer column sizes, JSON number precision limits).
3. Look up Go-specific numeric pitfalls relevant to the project's domain (e.g., `token.Pos` arithmetic in Go AST processing, file size computation with `os.Stat`, complexity score accumulation).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's numeric computations.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > /tmp/arithmetic-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -count=1 ./... 2>&1 | tee /tmp/arithmetic-test-results.txt
go vet ./... 2>&1 | tee /tmp/arithmetic-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Arithmetic Logic Audit

#### 3a. Integer Overflow and Underflow
For every arithmetic operation on integer types, verify the result cannot silently wrap:

- [ ] Multiplication of two `int` values that could each be large: `a * b` where both `a` and `b` can grow with input size (e.g., file count × average file size). Verify the product fits in the result type.
- [ ] Loop accumulators that add per-iteration values without overflow checks — an accumulator that processes millions of items may silently wrap.
- [ ] `uint` subtraction: `a - b` where `b > a` wraps to a large positive value rather than going negative. This is especially dangerous when the result is used as a count or size.
- [ ] `int32` / `int16` / `uint32` used for values that can exceed their max on large inputs (e.g., file size in `int32` wraps at ~2GB; line number in `int16` wraps at 32767).
- [ ] `len(slice)` returns `int` — on 32-bit platforms this limits slice length to ~2 billion; on 64-bit it is effectively unlimited. Verify the code does not store slice lengths in `int32` fields.
- [ ] Bit shift operations: `1 << n` where `n` can be ≥ 64 — in Go, the shift count must be less than the width of the type, or the result is zero (not a panic, but a silent bug).
- [ ] Negation of the minimum signed integer value: `-math.MinInt64` overflows back to `math.MinInt64`.

#### 3b. Integer Division Truncation
For every division operation producing a result intended as a real-number ratio, verify truncation is handled:

- [ ] Integer division used to compute a percentage: `a / b * 100` truncates to zero when `a < b`. The correct form is `a * 100 / b` (or use `float64`), but verify which is appropriate.
- [ ] Average computation: `total / count` where the intent is a rounded or fractional average. An average of `[1, 2]` computed as `(1+2)/2` gives `1`, not `1.5`.
- [ ] Ceiling division: code that needs `ceil(a / b)` but computes `a / b` (floor division) — the correct Go idiom is `(a + b - 1) / b`.
- [ ] `len(s) / 2` for midpoint computation: if `len(s)` is odd, the midpoint is truncated — verify whether the rounding direction matters for the algorithm.
- [ ] Division results used as loop bounds or slice indices where the truncated value causes the loop to skip the last element.
- [ ] `time.Duration` arithmetic: dividing a `time.Duration` by an integer gives a `time.Duration`, not a float. `time.Second / 3` is `333333333 ns`, not `333333333.33...`. Verify this is the intended behavior.

#### 3c. Division by Zero
For every division and modulo operation, verify the divisor cannot be zero:

- [ ] `a / b` and `a % b` where `b` is derived from user input, configuration, or a computed value — no check that `b != 0`.
- [ ] Average computation over a potentially empty collection: `total / len(items)` panics when `items` is empty.
- [ ] Divisors derived from metrics or statistics that could legitimately be zero: a file with zero functions, a package with zero dependencies, a codebase with zero test files.
- [ ] `math.Mod`, `big.Int.Div`, `big.Int.Mod` — these do not panic but produce `NaN` or undefined results for zero divisor; verify the behavior matches the intent.
- [ ] Reciprocal computation: `1.0 / x` in floating-point produces `+Inf` or `-Inf` for `x == 0.0` (no panic), which then propagates through further computations silently.

#### 3d. Sign and Signedness Errors
For every mixed signed/unsigned operation or sign-dependent comparison, verify:

- [ ] Conversion from `int` to `uint`: a negative `int` value becomes a very large `uint` — if this value is then used as a size or count, it will cause allocation failures or infinite loops.
- [ ] `len()` returns `int`; comparison with an `uint` value does not sign-extend correctly without explicit conversion. `uint(len(s)) > someUint` is safe; `len(s) > int(someUint)` may not be if `someUint > math.MaxInt`.
- [ ] Subtraction producing a negative result stored in a `uint` field: `uint(a - b)` where `a < b` wraps.
- [ ] `sort.Search` and `sort.Slice` use `int` indices; results used as unsigned values or indices into large slices need explicit bounds checks.
- [ ] Comparison of signed and unsigned values: Go requires explicit conversion; but logical errors where a negative sentinel value (e.g., `-1` meaning "not found") is stored in a field later compared as unsigned.
- [ ] Absolute value computation: `abs := x; if x < 0 { abs = -x }` — the negation of `math.MinInt64` overflows.

#### 3e. Floating-Point Logic Errors
For every floating-point computation, verify the logic accounts for floating-point semantics:

- [ ] Equality comparison of floating-point values: `f == 0.0`, `f == 1.0`, or `f1 == f2` after arithmetic — floating-point arithmetic is not exact. Use a tolerance comparison (`math.Abs(f-target) < epsilon`) where exactness is required.
- [ ] `NaN` comparisons: `NaN != NaN` in IEEE 754; `f == NaN` is always false. Use `math.IsNaN(f)` explicitly. Code that uses `NaN` as a sentinel "not computed yet" value will silently fail equality checks.
- [ ] `math.IsInf` is not checked after operations that can produce infinity (division by zero in float, overflow from large inputs to `math.Exp`, etc.), and `+Inf` or `-Inf` is then stored or reported as a metric.
- [ ] Accumulation of floating-point errors in a loop: summing many small values with a naive `total += item` accumulates error. For statistical metrics, verify the precision is adequate for the expected range of values.
- [ ] `float64` to `int` conversion: `int(f)` truncates toward zero, not rounds. `int(0.9999999)` is `0`, not `1`. Verify truncation vs rounding is intentional.
- [ ] Integer values larger than 2^53 stored in `float64` lose precision — all integers in the range [2^53, 2^64) cannot be represented exactly in `float64`.

#### 3f. Bit Manipulation Errors
For every bitwise operation, verify the operation produces the intended result:

- [ ] Setting a bit: `flags |= (1 << n)` — verify `1` has sufficient width. `1 << 32` on a system where `int` is 32 bits is zero, not `0x100000000`. Use `1 << uint(n)` or `uint64(1) << n` for portable code.
- [ ] Testing a bit: `flags & (1 << n) != 0` — in Go, `&` has higher precedence than `!=`, so this parses correctly as `(flags & (1 << n)) != 0`. Verify the logic is as intended and that the bit width is sufficient (see "Setting a bit" above).
- [ ] Clearing a bit: `flags &^= (1 << n)` — verify the same width issue as above.
- [ ] Right shift of signed integers: `x >> n` for negative `x` in Go performs arithmetic right shift (sign-extending), not logical right shift. If logical shift is intended, use `uint(x) >> n`.
- [ ] Using `^x` (bitwise complement) on signed integers: for `int8`, `^0` is `-1`, not `255`. Verify the intent is signed complement, not unsigned complement.
- [ ] Masking with a constant: `x & 0xFF` — verify the constant is the right width for the type; `int32(x) & 0xFFFF_FFFF` discards nothing, which may be the intent or may be a typo for `0xFF`.

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the value range**: An overflow can only occur if the operands can actually reach the critical range. Trace the maximum possible value of each operand through the call chain. A multiplication of two values each bounded at 1000 cannot overflow `int64`.
2. **Check for upstream clamping or validation**: A divisor that appears unguarded may be validated by the caller (e.g., configuration validation that rejects zero). Trace upward before flagging a divide-by-zero risk.
3. **Verify the integer type width**: An apparent overflow in `int` on a 64-bit platform where values are bounded at millions is not a real overflow risk. Check the platform assumptions of the project.
4. **Read surrounding comments**: If a comment explicitly acknowledges a numeric decision (e.g., `// safe: count cannot exceed 10000`, `// truncation intentional`, `//nolint:`), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Check the consequence**: A floating-point precision issue in a human-readable metric reported to two decimal places is LOW. A precision issue that causes a correctness failure in a binary protocol is CRITICAL.
6. **Verify floating-point equality intent**: Some code intentionally compares floats for exact equality when the values come from identical computation paths (not from user input or external sources). Do not flag as a bug unless the equality check can actually fail.

**Rule**: If you cannot demonstrate a concrete numeric value that triggers the arithmetic bug, do NOT report it. Numeric findings require a specific example value that causes the error.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# ARITHMETIC LOGIC AUDIT — [date]

## Project Numeric Profile
[Summary: types of numeric values computed, integer types used, presence of division/modulo, floating-point usage, overflow-sensitive paths, and stated correctness goals for metrics]

## Numeric Inventory
| Package | Integer Arithmetic | Division/Modulo | Float Ops | Bit Ops | Type Conversions | Overflow Risk |
|---------|-------------------|----------------|-----------|---------|-----------------|---------------|
| [pkg]   | N                 | N              | N         | N       | N               | ✅/❌          |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: concrete value that triggers the bug, result] — [impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why the arithmetic is actually safe for the values this code handles] |
```

Generate **`GAPS.md`**:
```markdown
# Arithmetic Logic Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about metric accuracy or numeric correctness]
- **Current State**: [what numeric operations exist without overflow/precision protection]
- **Risk**: [what input produces a wrong result: overflow, truncation, division by zero]
- **Closing the Gap**: [specific guard, type change, or formula correction needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Integer overflow silently producing a wrong metric or causing a panic; divide-by-zero panic on a reachable code path with valid input; sign error causing a large negative value used as a size |
| HIGH | Division truncation producing systematically wrong results (e.g., percentage always shows 0% for small numerators); `uint` underflow used as an allocation size; float `NaN` or `Inf` propagated into output |
| MEDIUM | Average or ratio computed with truncating division where rounding is intended; float equality comparison that fails on valid inputs due to precision; bit shift with insufficient type width |
| LOW | Precision loss in floating-point accumulation that only affects many-decimal-place accuracy; integer truncation that rounds in the unexpected direction but is documented |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly which expression to change, what the correct formula is, and what type or guard to use. Do not recommend "consider checking for overflow."
2. **Include the triggering value**: Every fix description must identify the specific numeric value or range that triggers the bug.
3. **Respect project idioms**: If the project uses `int64` throughout, recommend widening to `int64` not switching to `math/big`. If it uses `float64`, recommend `float64` fixes, not switching to `decimal` libraries.
4. **Verifiable**: Include a test case or example computation demonstrating the bug and the corrected result.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for function complexity and length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete triggering value demonstrating the arithmetic error — no speculative findings.
- Evaluate the code against its **own numeric conventions** and stated metric accuracy goals, not arbitrary external standards.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: panic-causing arithmetic (divide by zero, nil deref from overflow) → silent wrong results that affect all inputs of a given type → wrong results only on boundary inputs → precision issues affecting output accuracy → potential-but-unlikely overflow. Within a level, prioritize by the visibility of the wrong result: a metric shown to users is more important than an internal intermediate value.
