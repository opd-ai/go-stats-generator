# LLM Slop Prevention Architecture

## Mission Statement

`go-stats-generator` exists to make LLM-generated code **measurably accountable** to the same standards as expert-written code. It captures the metrics that experienced developers intuitively feel but rarely quantify — and makes them **machine-readable** so an LLM can act on them.

Traditional linters answer: *"Does this code follow rules?"*

`go-stats-generator` answers: *"Is this code getting worse, and where exactly?"*

---

## What Is "Slop"?

In the context of LLM-generated code, **slop** refers to code that:

- Compiles and nominally works, but is **structurally lazy, bloated, or non-idiomatic**
- Exhibits patterns an experienced developer would immediately flag in review
- Accumulates technical debt silently because it passes `go vet` and basic tests

Slop is not broken code. It is code that **degrades codebase health over time** in ways that are invisible to `go vet` or `golangci-lint` alone. LLMs produce slop at scale because they optimize for "code that compiles and passes tests" rather than "code that a senior engineer would approve."

---

## Concrete Slop Indicators in Go

| Category | Slop Signal | What `go-stats-generator` Detects |
|---|---|---|
| **Complexity bloat** | Single functions doing too much | Cyclomatic complexity > threshold, cognitive complexity spikes, function line count violations |
| **Copy-paste duplication** | Repeated logic with minor variations | AST-based Type 1/2/3 clone detection across files |
| **Missing documentation** | Exported API with zero GoDoc comments | Doc coverage < threshold on exported symbols, missing package docs |
| **Naming laziness** | `data`, `result`, `tmp`, `handleStuff` | Short/generic identifier detection, naming convention violations |
| **Over-nesting** | Deeply nested `if`/`switch`/`for` blocks | Nesting depth > threshold, cognitive complexity penalties |
| **Magic values** | Hardcoded numbers/strings with no explanation | Magic number/string literal detection in function bodies |
| **Dead code accumulation** | Unexported functions with no callers | Dead code indicators, unreachable function detection |
| **God structs/files** | One file or type doing everything | Types-per-file, functions-per-file, file line count, method count per type |
| **Naked error returns** | `if err != nil { return err }` with no context | Bare error returns without `fmt.Errorf` wrapping or annotation |
| **Concurrency naivety** | Goroutines with no synchronization or cancellation | Goroutine spawn density, missing `context.Context`, unsynchronized shared state |
| **Shallow interface usage** | Interfaces with one method and one implementor | Interface method count, implementor count, design pattern absence |
| **Parameter bloat** | Functions with 6+ parameters | Parameter count threshold violations |

---

## Three-Phase Workflow

### Phase 1: Pre-Generation Baseline

Before asking an LLM to generate or modify code, capture the codebase's current health as a **quantified snapshot**: complexity distributions, duplication levels, doc coverage, naming quality, organization metrics.

```bash
# Create a named baseline snapshot
go-stats-generator baseline create . --id "pre-llm-change"

# Or generate a JSON report for diff comparison
go-stats-generator analyze . --format json --output baseline.json
```

The baseline is the contract: *the LLM's output must not degrade these numbers.*

### Phase 2: Post-Generation Gate (Slop Prevention)

After the LLM produces code and it's written to disk, run a **regression check** — not just "does it compile" but:

- Did average cyclomatic complexity increase?
- Did duplication percentage rise?
- Did doc coverage drop?
- Did any new function exceed the complexity ceiling?
- Did naming quality scores degrade?
- Did maintenance burden index increase?

```bash
# Analyze the updated codebase
go-stats-generator analyze . --format json --output current.json

# Compare against the baseline
go-stats-generator diff baseline.json current.json --changes-only
```

For CI/CD enforcement, use threshold flags to block merges on regression:

```bash
go-stats-generator analyze . \
  --max-function-length 30 \
  --max-complexity 10 \
  --max-duplication-ratio 0.10 \
  --max-burden-score 50 \
  --min-doc-coverage 0.80 \
  --enforce-thresholds
```

If **any threshold is violated**, exit code 1 blocks the commit/merge. The output tells the LLM (or developer) **exactly what regressed and by how much**.

### Phase 3: Post-Facto Remediation (Slop Cleanup)

For codebases that already contain LLM-generated slop, the JSON report becomes a **remediation work order** that an LLM can consume:

```bash
go-stats-generator analyze . --format json --output slop-report.json
```

Feed the report to an LLM with structured instructions:

```
"Here is the analysis of my Go codebase. For each item flagged as
exceeding thresholds, refactor the code to bring it into compliance.
Prioritize by maintenance burden score descending."
```

The LLM receives structured, actionable data — not vague instructions — and can systematically address each violation.

---

## LLM Integration Architecture

### Feedback Loop Design

```
┌─────────────┐     ┌──────────────────────┐     ┌─────────────┐
│ LLM generates│───▶│  go-stats-generator   │───▶│ Violations? │
│ Go code      │     │  analyze + diff       │     │             │
└─────────────┘     └──────────────────────┘     └──────┬──────┘
       ▲                                                 │
       │              ┌──────────────────────┐     yes   │  no
       └──────────────│ LLM remediates with  │◀──────────┘   │
                      │ structured feedback   │               ▼
                      └──────────────────────┘          ✅ Accept
```

The tool is designed so its **JSON output is the LLM's input** in a remediation loop. This is not incidental — it is the primary design goal.

### JSON Output as LLM-Consumable Remediation Orders

Every violation in the JSON output includes the fields necessary for an LLM to act without ambiguity:

| Field | Purpose |
|---|---|
| `file` | Exact file path so the LLM knows what to edit |
| `line` | Line number for precise location |
| `item_name` | Function/type/interface name for context |
| `metric` | Which metric is violated (e.g., `cyclomatic_complexity`) |
| `actual_value` | Current measured value |
| `threshold` | The maximum allowed value |
| `severity` | `warning` or `violation` for prioritization |
| `suggestion` | Human/LLM-readable remediation hint |

Example violation entry:

```json
{
  "file": "internal/analyzer/complexity.go",
  "line": 83,
  "item_name": "calculateCognitiveComplexity",
  "metric": "cyclomatic_complexity",
  "actual_value": 22,
  "threshold": 10,
  "severity": "violation",
  "suggestion": "Extract switch cases into named helper functions or use a dispatch map"
}
```

An LLM receiving this knows **exactly** what to fix, **where**, and **how much** improvement is needed.

### CI/CD as Slop Firewall

```yaml
# .github/workflows/slop-gate.yml
name: Slop Prevention Gate

on:
  pull_request:
    branches: [main]

jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install go-stats-generator
        run: go install github.com/opd-ai/go-stats-generator@latest

      - name: Analyze current code
        run: |
          go-stats-generator analyze . --format json --output current.json

      - name: Check for slop regression
        run: |
          go-stats-generator analyze . \
            --max-function-length 30 \
            --max-complexity 10 \
            --max-duplication-ratio 0.10 \
            --max-burden-score 50 \
            --min-doc-coverage 0.80 \
            --enforce-thresholds

      - name: Diff against baseline (if available)
        if: hashFiles('.baseline.json') != ''
        run: |
          go-stats-generator diff .baseline.json current.json --changes-only
```

This turns the tool into an **automated slop firewall**: no PR with degraded metrics merges, whether the code was written by a human or an LLM.

---

## Go-Specific Slop Patterns

These are slop patterns **unique to Go** that `go-stats-generator` specifically detects because LLMs produce them frequently:

### 1. Bare Error Returns Without Wrapping

**Pattern:** `if err != nil { return err }` without adding context via `fmt.Errorf("doing X: %w", err)`.

**Why LLMs do this:** LLMs propagate errors mechanically without understanding call-chain context.

**Detection:** Bare `return err` density after `err != nil` checks.

**Remediation:** Wrap every error return with context: `return fmt.Errorf("loading config: %w", err)`.

### 2. Goroutine Leaks

**Pattern:** `go func()` spawned without cancellation, `sync.WaitGroup`, or `context.Context` propagation.

**Why LLMs do this:** LLMs spawn goroutines to satisfy concurrency requirements without modeling lifecycle.

**Detection:** Goroutine spawns without corresponding synchronization primitives in the same scope.

**Remediation:** Add `context.Context` propagation, `sync.WaitGroup` for fan-out, or `errgroup.Group` for error-aware concurrency.

### 3. `interface{}` / `any` Overuse

**Pattern:** Empty interfaces used to avoid designing proper types.

**Why LLMs do this:** LLMs use `interface{}` and `any` as escape hatches when they cannot infer the correct type.

**Detection:** `interface{}` and `any` parameter/return density outside of genuinely generic utility functions.

**Remediation:** Define concrete types or use generics (Go 1.18+) with type constraints.

### 4. `init()` Proliferation

**Pattern:** Setup logic placed in `init()` functions, creating hidden execution order dependencies.

**Why LLMs do this:** LLMs use `init()` as a convenient place for initialization without considering testability or dependency injection.

**Detection:** `init()` function count and complexity per package.

**Remediation:** Move initialization into explicit constructor functions or `main()`.

### 5. Stuttering Names

**Pattern:** `package user` with `type UserService struct` producing `user.UserService` at call sites.

**Why LLMs do this:** LLMs generate fully-qualified conceptual names without considering the package prefix already provides context.

**Detection:** Exported identifiers that repeat the package name prefix.

**Remediation:** Rename to `type Service struct` so the call site reads `user.Service`.

### 6. Naked Returns in Long Functions

**Pattern:** Named returns with naked `return` in functions beyond trivial length, harming readability.

**Why LLMs do this:** LLMs use named returns with naked `return` mechanically, regardless of function length.

**Detection:** Naked returns in functions exceeding a line threshold.

**Remediation:** Replace naked returns with explicit `return val1, val2, err` in any function longer than ~10 lines.

### 7. `panic()` in Library Code

**Pattern:** `panic()` or `log.Fatal()` used instead of returning errors in non-main packages.

**Why LLMs do this:** LLMs use `panic()` to simplify error handling, especially in initialization code.

**Detection:** `panic()` and `log.Fatal()` calls outside `main` packages and `init()` functions.

**Remediation:** Return errors instead; let the caller decide how to handle failures.

### 8. Giant `switch`/`if-else` Chains

**Pattern:** Long conditional chains instead of maps, interfaces, or table-driven logic.

**Why LLMs do this:** LLMs generate exhaustive case-by-case logic because it is the simplest pattern to produce.

**Detection:** Switch/if-else branch count per statement, flagged above threshold.

**Remediation:** Replace with a dispatch map (`map[string]func()`), strategy pattern, or table-driven approach.

### 9. Unused Receiver Names

**Pattern:** Method receivers like `func (s *Service) DoThing()` where `s` is never referenced in the method body.

**Why LLMs do this:** LLMs generate methods by default even when a plain function would suffice.

**Detection:** Method receivers that are unused in the method body.

**Remediation:** Convert to a function, or use `_` as the receiver name if the method is required to satisfy an interface.

### 10. Test-Only Exports

**Pattern:** Types and functions exported solely to make them testable from `_test.go` files, instead of using `export_test.go` patterns or restructuring.

**Why LLMs do this:** LLMs export symbols as the path of least resistance to test internal logic.

**Detection:** Exported symbols with zero cross-package references outside test files.

**Remediation:** Use `export_test.go` files for test-only access, or restructure to test via the public API.

---

## Alignment With `rust-stats-generator`

`go-stats-generator` and [rust-stats-generator](https://github.com/opd-ai/rust-stats-generator) are **companion tools** built on the same anti-slop philosophy. `go-stats-generator` was developed first and establishes the core workflow patterns that `rust-stats-generator` mirrors for the Rust ecosystem.

### Shared Principles

1. **Same philosophy** — Both tools exist to make LLM-generated code measurably accountable to expert-level standards.

2. **Same workflow integration** — The Baseline → Generate → Diff → Remediate loop is **the** workflow. Both tools support it identically at the CLI level.

3. **Same JSON schema compatibility** — Where metrics overlap (complexity, duplication, doc coverage, naming, organization), field names and structure match so that:
   - Dashboards work across both languages
   - CI/CD scripts share threshold logic
   - LLM remediation prompts are language-agnostic at the structural level

4. **Same threshold enforcement** — `--enforce-thresholds` with exit code 1 is the primary CI/CD integration point for both tools. Behavior is identical.

5. **Same remediation-first output** — Every metric in the report is actionable. No metric exists just for vanity — if it's measured, there is a clear remediation path an LLM can follow.

### Language-Specific Extensions

Each tool extends the shared core with language-specific slop detection. These extensions are additive — they do not replace or conflict with the shared metric set.

```
┌─────────────────────────────┐
│      Shared JSON Schema      │
│  (complexity, duplication,   │
│   doc coverage, naming,      │
│   organization, burden)      │
└──────────┬──────────────────┘
           │
     ┌─────┴─────┐
     │           │
     ▼           ▼
┌──────────┐ ┌──────────────────┐
│ go-stats │ │ rust-stats       │
│ generator│ │ generator        │
├──────────┤ ├──────────────────┤
│ Go-only: │ │ Rust-only:       │
│ goroutine│ │ .clone() density │
│ init()   │ │ .unwrap() abuse  │
│ stutter  │ │ derive spam      │
│ naked ret│ │ Box<dyn> overuse │
└──────────┘ └──────────────────┘
```

**Go-specific detections:**
- Goroutine leaks and concurrency naivety
- `init()` proliferation
- Stuttering names (package name repeated in exported identifiers)
- Naked returns in long functions
- `interface{}` / `any` overuse
- `panic()` in library code
- Bare error returns without wrapping

**Rust-specific detections:**
- `.clone()` proliferation
- `.unwrap()` abuse
- `Box<dyn Trait>` overuse
- Derive spam
- `String` everywhere (instead of `&str`)

Both tools feed the same LLM remediation loop. A team using Go and Rust gets **one unified quality methodology** across both languages.

---

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Analysis completed, all checks pass |
| 1 | Quality gate violation (when `--enforce-thresholds` is set) |
| 2 | Analysis error (invalid input, file not found, etc.) |

---

## Quick Reference

### Create a Baseline Before LLM Generation

```bash
go-stats-generator analyze . --format json --output baseline.json
```

### Run the Slop Gate After LLM Generation

```bash
go-stats-generator analyze . \
  --max-function-length 30 \
  --max-complexity 10 \
  --max-duplication-ratio 0.10 \
  --max-burden-score 50 \
  --enforce-thresholds
```

### Diff Against Baseline

```bash
go-stats-generator analyze . --format json --output current.json
go-stats-generator diff baseline.json current.json --changes-only
```

### Generate a Remediation Work Order for an LLM

```bash
go-stats-generator analyze . --format json --output slop-report.json
```

Then feed `slop-report.json` to the LLM with instructions to remediate each flagged item, prioritized by maintenance burden score.

### Track Trends Over Time

```bash
go-stats-generator trend analyze --days 30 --metric mbi_score
go-stats-generator trend forecast --metric duplication_ratio
```
