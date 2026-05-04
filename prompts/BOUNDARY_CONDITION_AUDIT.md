# TASK: Perform a focused audit of boundary condition and off-by-one logic errors in Go code, identifying incorrect index arithmetic, fence-post errors, loop termination bugs, and edge-case mishandling while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the boundary condition audit report
2. **`GAPS.md`** — gaps in boundary condition handling relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Data Structures and Iteration Patterns
1. Read the project README to understand its purpose, users, and the nature of the data it processes (files, ASTs, indices, counts).
2. Examine `go.mod` for module path, Go version, and dependencies that introduce iteration, pagination, or indexed access.
3. List packages (`go list ./...`) and identify which packages perform indexed access, range iteration, or slice/array manipulation.
4. Build a **boundary inventory** by scanning for:
   - Direct slice/array index access: `s[i]`, `s[i:j]`, `s[:n]`, `s[n:]`
   - Loop bounds derived from lengths or counts: `for i := 0; i < len(s); i++`
   - Off-by-one-prone comparisons: `<` vs `<=`, `>` vs `>=`, `!=` as a loop terminator
   - Slice operations that compute indices: `s[start:end]`, `copy(dst, src[offset:])`
   - Length/count arithmetic used as bounds: `n-1`, `n+1`, `len(s)-1`
   - Nil and empty collection access: accessing element 0 of an empty slice, iterating an empty map
   - `strings.Index`, `bytes.Index`, `strings.Cut`, `strings.Split` return value handling (`-1` for not-found)
   - `binary.Read`, `binary.Write`, and fixed-size buffer access with computed offsets
   - Pagination: `offset + limit`, `page * pageSize`, `(page-1) * pageSize`
5. Identify the project's conventions for bounds checking — does it check length before access? Does it use `range` exclusively, or does it use index arithmetic?
6. Map which functions produce indices that are later used by other functions as array/slice subscripts — a correct-looking consumer may rely on a buggy producer.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "index out of range", "panic", "off by one", "wrong result", "missing last", "skipped first", or "fence post" to understand known boundary bugs.
2. Research key dependencies from `go.mod` for any known boundary or index-related bugs in their APIs.
3. Look up Go-specific boundary condition pitfalls relevant to the project's domain (e.g., AST node position arithmetic, file line/column numbering conventions, zero-based vs one-based indexing in reporting).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's data access patterns.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/boundary-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -count=1 ./... 2>&1 | tee tmp/boundary-test-results.txt
go vet ./... 2>&1 | tee tmp/boundary-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Boundary Condition Audit

#### 3a. Off-By-One Errors in Index Arithmetic
For every direct index computation, verify the arithmetic is correct:

- [ ] Loop bounds that use `i <= len(s)-1` instead of `i < len(s)` — both are correct but the subtraction form panics when `len(s) == 0` due to unsigned underflow if `len` were a uint (it is int in Go, so it returns -1; still logically suspicious).
- [ ] Slicing operations where the upper bound should be exclusive: `s[start:end]` where `end` is intended as the last index (should be `end+1`) or is already past-the-end.
- [ ] "Last N elements" patterns: `s[len(s)-N:]` panics or produces wrong results when `len(s) < N`.
- [ ] "All but last" patterns: `s[:len(s)-1]` panics when `s` is empty.
- [ ] Loop variable initialization off by one: `for i := 1; ...` when the first element should be processed, or `for i := 0; ...` when the first element must be skipped.
- [ ] `copy(dst, src)` with manually computed lengths that are off by one, copying one too few or one too many elements.
- [ ] Window/sliding operations: `s[i : i+windowSize]` without verifying `i+windowSize <= len(s)`.
- [ ] Binary search implementations: verify the pivot computation `mid = (lo + hi) / 2` is used correctly and that `lo`, `hi` converge without skipping the target.

#### 3b. Fence-Post Errors in Comparisons
For every loop termination condition and range comparison, verify:

- [ ] `for i < n` vs `for i <= n` — confirm which bound is intended as inclusive and which as exclusive based on what the loop body accesses.
- [ ] Comparison operators on lengths used as limits: `if len(s) > 0` before `s[0]` is correct; `if len(s) >= 0` is always true and provides no protection.
- [ ] Sentinel-based loop termination: `for i != end` — verify `end` is reachable from the start direction; if the step can skip over `end`, the loop is infinite or terminates with wrong state.
- [ ] "At least N" checks: `if n >= threshold` vs `if n > threshold-1` — confirm whether the boundary value is included in the "valid" set.
- [ ] Exclusive-end ranges used as inclusive in error messages or output (e.g., reporting line numbers where the AST uses 0-based offsets but the user expects 1-based).
- [ ] `strings.Index` and `bytes.Index` return `-1` for not-found; using the return value directly as a slice index without checking for `-1` causes a panic.
- [ ] `strings.Split(s, sep)` always returns at least one element, even for empty input; code that assumes `len(result) >= 2` before accessing `result[1]` may panic on single-segment input.

#### 3c. Empty and Nil Collection Edge Cases
For every operation that accesses a collection, verify the empty/nil case is handled:

- [ ] `s[0]` or `s[len(s)-1]` without a preceding `if len(s) > 0` check.
- [ ] Range over a nil slice is safe in Go (produces zero iterations), but range over a nil map is also safe — verify the code does not assume a nil map and a non-nil empty map are interchangeable when later writing to the map.
- [ ] `map[key]` on a nil map returns the zero value without panic (read is safe), but `map[key] = value` on a nil map panics — verify the map is initialized before any write.
- [ ] Functions that return `([]T, error)` — callers that use the slice without checking the error may operate on a nil or partial slice.
- [ ] `append` to a nil slice is valid and returns a new slice, but code that compares the appended slice to the original by identity will see they differ.
- [ ] First/last element access in tree or linked-list traversal when the structure is empty.

#### 3d. Loop Termination and Iteration Correctness
For every loop that modifies the collection it iterates, or uses computed bounds, verify:

- [ ] Loops that shrink a slice (`s = s[:len(s)-1]`) while iterating it by index — the index may now be out of bounds.
- [ ] Loops that append to a slice while iterating with `range` — `range` captures the original length; appended elements are not visited (this is correct Go behavior, but may be a logic error if the intent was to visit all elements including appended ones).
- [ ] Loops that delete from a map while iterating with `range` — Go allows this, but verify the loop logic still produces correct results (e.g., counting deleted items).
- [ ] Two-pointer or sliding window loops where both pointers advance: verify the invariant that `left <= right` is maintained and that both pointers stay within bounds.
- [ ] Nested loops where the inner loop's bound depends on the outer loop's index: `for j := i+1; j < n; j++` — verify `i+1` does not exceed `n` before the inner loop starts.
- [ ] Loop that is supposed to process all elements but has an early `continue` or `break` that skips the last element.

#### 3e. Pagination and Chunking Arithmetic
For every paginated query, chunked read, or batched operation, verify:

- [ ] `offset = (page - 1) * pageSize` — verify page numbering is 1-based vs 0-based consistently throughout the call stack.
- [ ] Last page/chunk may have fewer items than `pageSize` — verify the consumer handles `len(chunk) < pageSize` correctly and does not access `chunk[pageSize-1]` on the last page.
- [ ] Total item count divided into chunks: `numChunks = (total + chunkSize - 1) / chunkSize` — verify this ceiling division is used, not truncating division which loses the last partial chunk.
- [ ] Buffer reuse across chunks: if the same buffer is reused for each chunk, verify the slice is re-sliced to the actual number of items read, not always to `cap(buf)`.
- [ ] `io.ReadFull` vs `io.Read` — `io.Read` may return fewer bytes than requested without error; using its return value `n` as a fixed-size record length may process partial records.

#### 3f. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the code path is reachable**: Confirm the slice/index access is actually reachable and that the collection can actually be empty or shorter than the code assumes. Trace the data source.
2. **Check for upstream length enforcement**: A slice access at index `n` may be safe because a caller always ensures `len(s) > n`. Trace the full call chain before flagging.
3. **Distinguish panic from logic error**: A missing bounds check that causes a panic is a CRITICAL finding. A bounds check that uses `<=` instead of `<` but never causes a panic in practice because the data is always well-formed is MEDIUM.
4. **Read surrounding comments**: If a comment explicitly acknowledges a boundary decision (e.g., `// caller guarantees non-empty`, `// safe: length checked above`, `//nolint:`, or a TODO), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Check test coverage for edge cases**: If the test suite exercises the empty-collection and single-element cases and they pass, downgrade speculative findings. Use test evidence, not absence of evidence.
6. **Assess the consequence**: A boundary error that produces a wrong result silently is more severe than one that panics — panics are visible and recoverable; silent wrong results may propagate.

**Rule**: If you cannot demonstrate a concrete input that triggers the boundary error (either a panic or a silent wrong result), do NOT report it. Off-by-one findings require a specific counterexample.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# BOUNDARY CONDITION AUDIT — [date]

## Project Data Access Profile
[Summary: data structures accessed by index, iteration patterns, pagination usage, indexing conventions (0-based vs 1-based), and empty/nil handling approach]

## Boundary Inventory
| Package | Direct Index Access | Slice Arithmetic | Pagination | Empty Checks | String Index Use |
|---------|---------------------|-----------------|------------|--------------|-----------------|
| [pkg]   | N                   | N               | N          | ✅/❌         | N               |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: concrete input that triggers panic or wrong result] — [impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why the boundary condition is actually safe in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Boundary Condition Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about correctness or data handling]
- **Current State**: [what boundary checking exists]
- **Risk**: [what input causes incorrect behavior: panic or silent wrong result]
- **Closing the Gap**: [specific bounds checks, guard clauses, or index arithmetic corrections needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Missing bounds check causing a runtime panic on a reachable code path with realistic input, or an off-by-one that causes silent data corruption (wrong element written/read) |
| HIGH | Off-by-one that skips the first or last element of a result, causing systematically incomplete output on all inputs of certain shapes |
| MEDIUM | Missing nil/empty collection guard that panics only on unusual but valid input, or incorrect pagination arithmetic producing wrong results on the last page |
| LOW | Suspicious-looking boundary arithmetic that is actually correct but should be documented, or a fence-post comparison that works but is confusingly written |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what index arithmetic, guard clause, or comparison operator to change. Include the corrected expression, not just a description.
2. **Include the counterexample**: Every fix description must identify the specific input or state that triggers the bug.
3. **Verifiable**: Include a test case or `go test -run` command that reproduces the bug and passes after the fix.
4. **Minimal scope**: Fix the boundary condition without restructuring unrelated logic.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for function complexity and length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete counterexample input demonstrating the bug — no speculative findings.
- Evaluate the code against its **own conventions** and the actual data it processes, not arbitrary external standards.
- Apply the Phase 3f false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: runtime panics on common inputs → silent wrong results on all inputs of a given shape → panics on uncommon but valid inputs → wrong results only on edge cases → confusing-but-correct arithmetic. Within a level, prioritize by proximity to user-visible output and by how frequently the buggy code path executes.
