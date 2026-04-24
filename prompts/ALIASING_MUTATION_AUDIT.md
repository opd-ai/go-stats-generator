# TASK: Perform a focused audit of data aliasing and mutation logic errors in Go code, identifying unintended shared state from slice/map aliasing, shallow copy bugs, unexpected mutation through pointers, iteration variable capture errors, and incorrect equality semantics while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the aliasing and mutation audit report
2. **`GAPS.md`** — gaps in data isolation relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Data Ownership Model
1. Read the project README to understand its purpose and the types of data it collects, transforms, and stores (AST nodes, file stats, metric structs, report data).
2. Examine `go.mod` for module path, Go version, and dependencies that pass data by reference or use shared data structures.
3. List packages (`go list ./...`) and identify which packages store, copy, or transform aggregated data structures.
4. Build a **data ownership inventory** by scanning for:
   - Struct assignments: `a = b` where both are struct types (copies the value but not any pointed-to data)
   - Slice header copies: `s2 := s1` (copies the header; both share the backing array)
   - Map assignments: `m2 := m1` (both refer to the same underlying map)
   - Pointer aliasing: `p2 := p1` (both point to the same memory)
   - Function parameters passed by pointer that are later modified
   - Return values that are slices or maps sharing the caller's backing storage
   - Struct fields that are pointers, slices, or maps — copied structs still share these
   - `append` operations on shared slice headers
   - Loop variable capture: closures that reference `i`, `v` from an enclosing `for range`
   - `copy(dst, src)` — shallow copy of slice elements that are themselves pointers or contain slices
5. Identify the project's data ownership conventions — does it document which fields are owned vs borrowed? Does it use `copy` when storing slices? Does it clone structs before storing?
6. Map the data flow from collection (AST walking, file reading) through transformation (aggregation, scoring) to storage (results structs, output) — unexpected aliasing most commonly occurs at handoff points.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "wrong result", "modified after return", "stale data", "overwritten", "shared state", or "unexpected mutation" to understand known aliasing bugs.
2. Research key dependencies from `go.mod` for data ownership contracts (e.g., does the AST library share node memory? does the SQLite library reuse row buffers?).
3. Look up common Go aliasing pitfalls relevant to the project's domain (e.g., `go/ast` node reuse in visitors, `bufio.Scanner` token reuse, `encoding/json` decoder buffer reuse).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's data handling patterns.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > /tmp/aliasing-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -race -count=1 ./... 2>&1 | tee /tmp/aliasing-test-results.txt
go vet ./... 2>&1 | tee /tmp/aliasing-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

**Note**: The race detector (`-race`) can catch some aliasing bugs in concurrent code, but most aliasing bugs are sequential logic errors that the race detector will not find. This audit focuses primarily on non-concurrent aliasing.

### Phase 3: Aliasing and Mutation Audit

#### 3a. Slice Header Aliasing
For every slice assignment and slice-returning function, verify the caller and callee do not share the backing array when independent modification is needed:

- [ ] `s2 := s1` followed by `append` to `s2` — if `len(s1) < cap(s1)`, the append writes into `s1`'s backing array, silently modifying `s1`'s data.
- [ ] Functions that return a sub-slice of an internal buffer (`return buf[start:end]`) — the caller may hold a slice into the function's or struct's private buffer; later operations on that buffer corrupt the caller's data.
- [ ] `copy` vs assignment: code that stores a slice field in a struct by assignment (`result.Items = items`) when `items` is the same underlying array as a local buffer that will be reused or modified.
- [ ] `append` to a function parameter that is a slice: in Go, `append` may or may not modify the caller's array depending on capacity; if the function assumes it is always creating a fresh copy, it is wrong half the time.
- [ ] `json.Unmarshal` into a `[]byte` that is then stored — the JSON decoder may hold a reference to the input buffer; mutations to the buffer corrupt the decoded data.
- [ ] `bufio.Scanner.Bytes()` returns a slice into the scanner's internal buffer, which is overwritten on the next `Scan()` call. Storing `scanner.Bytes()` directly (not `append([]byte{}, scanner.Bytes()...)` or `string(scanner.Bytes())`) is a use-after-next-scan bug.

#### 3b. Map Reference Aliasing
For every map assignment and map-returning function, verify the caller and callee do not share the same underlying map when independent modification is needed:

- [ ] `m2 := m1` — both `m2` and `m1` refer to the same map; writes to `m2` mutate `m1` and vice versa.
- [ ] Struct copy where a field is a map: `s2 := s1` copies the struct value but both `s2.Field` and `s1.Field` still point to the same map.
- [ ] Functions that return a map field from a struct by value — the caller holds a direct reference to the struct's map; writes to the returned map mutate the struct.
- [ ] Merging maps: code that appends one map's entries into another correctly, but leaves the source map's entries pointing to the same value objects as the destination (shallow merge).
- [ ] Cache or registry maps: code that returns a map from a cache and the caller modifies it, corrupting the cache for all future lookups.

#### 3c. Pointer and Interface Aliasing
For every pointer assignment and interface value, verify the sharing semantics are correct:

- [ ] Storing a pointer to a loop variable: `ptrs = append(ptrs, &items[i])` in a range loop — the address of `items[i]` is stable because it indexes a slice, but `&item` (where `item` is the range variable) is the same address each iteration and always points to the last value.
- [ ] Struct pointer fields: when a struct is copied (`s2 = *s1`), pointer fields in the copy still point to the same objects as `s1`'s fields — mutating through `s2.Ptr` also mutates `s1`'s referenced data.
- [ ] `interface{}` / `any` values holding mutable types: the interface value stores a copy of a value type (safe) or a reference to a reference type (shares the underlying data). Verify callers do not mutate interface-wrapped values they do not own.
- [ ] Functions that accept `*T` and store the pointer in a long-lived data structure — the caller may reuse the pointed-to `T` for a different purpose, corrupting stored data. Verify whether the function should copy `*T` before storing.
- [ ] Returning `&localVar` from a function is valid in Go (the variable escapes to the heap), but returning a pointer to an element of a local array is dangerous if the array is also returned or reused.

#### 3d. Loop Variable Capture and Closure Mutation
For every closure defined inside a loop, verify variable capture is correct:

- [ ] `for i, v := range items { go func() { use(i, v) }() }` — `i` and `v` are loop variables reused each iteration; by the time the goroutine runs, both have the value from the last iteration. Fix: `go func(i int, v T) { use(i, v) }(i, v)`.
- [ ] `for i, v := range items { funcs = append(funcs, func() { use(i, v) }) }` — same capture bug in synchronous closures; all closures in `funcs` share the same `i` and `v` variables.
- [ ] Loop variable capture in callbacks, event handlers, and `defer` statements within a loop body — `defer func() { use(v) }()` captures `v` by reference; each deferred call sees the final loop value, not the value at the time `defer` was called.
- [ ] **Note**: As of Go 1.22, loop variables are re-created each iteration (per-iteration semantics), eliminating the classic capture bug for `for range` loops. Verify the project's `go.mod` Go version. If `go 1.22` or later, do NOT report range variable capture as a bug — it is fixed at the language level. Report only closures over non-range variables or Go <1.22 range variables.
- [ ] Closures that capture a slice header by reference: modifying the slice variable inside the closure does not modify the outer variable (slices are passed by value as headers), but modifying slice elements does affect the outer data.

#### 3e. Unintended Mutation Through Shared State
For every function that modifies its arguments or shared fields, verify mutations are intentional:

- [ ] Functions that sort their input slice in-place when the caller does not expect the input to be reordered — the function should document the side effect or sort a copy.
- [ ] Functions that normalize or modify string/byte slices in-place (e.g., lowercasing) when the input is shared with another data structure.
- [ ] AST visitor patterns where the visitor modifies `ast.Node` fields — modifying a shared AST node affects all other visitors and later analyses that use the same AST.
- [ ] Struct methods that modify receiver fields when the receiver is used as a value in a collection: `items[i].Method()` modifies the struct, but `for _, item := range items { item.Method() }` does not (because `item` is a copy). Verify which is intended.
- [ ] Configuration or options structs passed by value that contain pointer/slice fields — the callee modifying a slice field of its options copy still modifies the caller's underlying data.
- [ ] Accumulation into a shared result structure from multiple goroutines without synchronization — the mutation appears sequential but is actually concurrent.

#### 3f. Shallow Copy and Deep Equality Errors
For every copy or equality check on composite types, verify the depth is appropriate:

- [ ] `reflect.DeepEqual` vs `==` for structs containing pointers — `==` on structs compares pointer addresses, not pointed-to values; two logically identical structs with different pointer addresses compare as unequal.
- [ ] `bytes.Equal(a, b)` vs `a == b` for `[]byte` — `a == b` always false for slices (slices are not comparable with `==`; this is a compile error). But `string(a) == string(b)` works correctly and allocates; `bytes.Equal` is correct and allocation-free.
- [ ] Deep copy implementations that copy only one level: a struct with a `[]string` field is "deep copied" by copying the struct value, but the new struct's slice field still shares the backing array with the original. True deep copy requires `copy(new.Field, old.Field)`.
- [ ] Test assertions that use `==` on structs containing slices or maps — these never compare as equal even if contents are identical. Use `reflect.DeepEqual` or `cmp.Equal` (from `github.com/google/go-cmp`).
- [ ] Sorting and deduplication functions that use `==` for pointer types when value equality is intended — two objects with identical fields but different addresses compare as unequal, causing duplicates to remain.

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the mutation path is reachable**: Confirm that both the aliased copy and the original are actually written to after the aliasing occurs. Read-only aliasing is not a bug.
2. **Check for documented ownership transfer**: A function that returns a slice into its internal buffer may document that the caller must copy the data before the next call. If so, callers that do copy are correct and callers that do not copy are bugs — report only the incorrect callers.
3. **Check the Go version for loop variable semantics**: For Go ≥1.22, range loop variable capture bugs are eliminated by the language. Do not report them for projects using Go 1.22+.
4. **Verify the concurrency context**: Aliasing bugs in single-goroutine sequential code only matter if both the alias and the original are actually used after the mutation. In concurrent code, the race detector may already catch them.
5. **Read surrounding comments**: If a comment explicitly acknowledges a sharing decision (e.g., `// caller must copy before next call`, `// intentional alias — read-only`, `//nolint:`, or a TODO), treat it as an acknowledged pattern — do not report it as a new finding.
6. **Check for copy-on-write or immutability patterns**: Some code intentionally shares backing storage for efficiency and copies only when a write is needed. Verify whether the code actually implements this correctly before flagging it.

**Rule**: If you cannot demonstrate that both the alias and the original are written to (or that the alias is read after the original is mutated), do NOT report it as an aliasing bug.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# ALIASING AND MUTATION AUDIT — [date]

## Project Data Ownership Profile
[Summary: data flow from collection through transformation to storage, slice/map passing conventions, presence of shared mutable state, Go version and loop variable semantics]

## Data Ownership Inventory
| Package | Slice Aliases | Map Aliases | Pointer Aliases | Loop Captures | In-Place Mutations | Deep Copy Needs |
|---------|--------------|-------------|-----------------|---------------|-------------------|----------------|
| [pkg]   | N            | N           | N               | N             | N                 | ✅/❌           |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: alias path, mutation point, affected consumers] — [impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why the aliasing is safe or intentional in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Data Aliasing and Mutation Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about data correctness or result independence]
- **Current State**: [what data sharing exists without explicit isolation]
- **Risk**: [what mutation causes incorrect results: stale data, overwritten output, wrong accumulation]
- **Closing the Gap**: [specific copy, clone, or ownership transfer needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Aliasing causes silent data corruption in results — a metric is computed incorrectly because an intermediate buffer was overwritten; or a loop capture causes all closures to operate on the wrong data |
| HIGH | Aliasing causes incorrect results on specific inputs (e.g., the last element is always reused, overwriting the previous result), or shared map corruption causes incorrect cache behavior |
| MEDIUM | Aliasing causes a problem only under concurrent use (but currently sequential code may become concurrent), or shallow copy leaves pointer fields shared when deep copy is required for correctness |
| LOW | Aliasing is safe currently but fragile — a future refactor could introduce mutation; or equality comparison uses pointer identity when value equality is more appropriate |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly where to add `copy()`, `append([]byte{}, src...)`, `maps.Clone()`, a deep copy function, or a parameter copy. Do not recommend "consider copying the data."
2. **Include the mutation path**: Identify exactly which code writes to the shared data and what the observable effect on the aliased copy is.
3. **Respect project idioms**: If the project copies slices with `append(nil, src...)`, recommend the same pattern. If it has a `Clone()` method, recommend using it.
4. **Verifiable**: Include a test case with two operations on the same data that demonstrates the aliasing bug: perform operation A, then operation B, then verify A's result is unchanged.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for struct complexity and function length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete mutation path demonstrating the aliasing bug — no speculative findings.
- Evaluate the code against its **own data ownership conventions** and stated correctness goals, not arbitrary external standards.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.
- Check the Go version in `go.mod` before reporting loop variable capture bugs (fixed in Go 1.22+).

## Tiebreaker
Prioritize: silent corruption of reported results → aliasing that causes all output to reflect only the last-processed item → aliasing that causes incorrect results on specific inputs → aliasing in concurrent code caught by the race detector → fragile aliasing that is currently safe. Within a level, prioritize by how many output values are affected and how visible the corruption is.
