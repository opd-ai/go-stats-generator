#!/usr/bin/env bash
# loop.sh — Autonomous iterative Go project completion orchestrator
#
# Drives an arbitrary Go project from its current state to bug-free,
# slop-free completion by delegating the repository's markdown prompt
# files to GitHub Copilot CLI.
#
# Usage:
#   bash scripts/loop.sh [PROMPT_DIR]
#
# Arguments:
#   PROMPT_DIR  Path to the directory containing prompt .md files
#               (default: prompts/ directory at the repository root)
#
# Environment:
#   PROMPT_DIR      Override the prompt directory (takes precedence over $1)
#   MAX_ITERATIONS  Maximum loop iterations (default: 50)
#
# The script is idempotent on re-run and safe to interrupt (Ctrl+C).

set -euo pipefail

# ─── Trap: graceful shutdown on Ctrl+C / SIGTERM ─────────────────────────────

cleanup() {
    echo ""
    echo "=== LOOP INTERRUPTED ==="
    echo "Iterations completed: $ITERATION"
    echo "Tasks executed: $TASKS_EXECUTED"
    echo "Test failures encountered: $TEST_FAILURES"
    echo "Re-run this script to resume from current state."
    exit 130
}
trap cleanup INT TERM

# ─── Constants and defaults ──────────────────────────────────────────────────

MAX_ITERATIONS="${MAX_ITERATIONS:-50}"
MAINTENANCE_INTERVAL=3          # Run periodic maintenance every N iterations
LOG_FILE="loop.log"
TEST_OUTPUT="test-output.txt"
ANALYSIS_OUTPUT=".loop-analysis.json"

# Counters
ITERATION=0
TASKS_EXECUTED=0
TEST_FAILURES=0
FINAL_TEST_STATUS="UNKNOWN"

# ─── Logging helpers ─────────────────────────────────────────────────────────

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
    echo "$msg" >> "$LOG_FILE"
}

log_and_print() {
    local msg="$*"
    log "$msg"
    echo "$msg"
}

iteration_summary() {
    local prompt="$1"
    local test_status="$2"
    local line="[iteration $ITERATION] executed $prompt — tests: $test_status"
    log "$line"
    echo "$line"
}

# ─── 1. Entry Gate ───────────────────────────────────────────────────────────

# Verify ROADMAP.md exists in the current working directory
if [ ! -f "ROADMAP.md" ]; then
    echo "ERROR: ROADMAP.md not found in $(pwd)." >&2
    echo "The target Go project must have a ROADMAP.md in its root directory." >&2
    exit 1
fi

# Resolve the prompt directory
# Default: prompts/ directory next to the scripts/ directory containing this script.
# Override via PROMPT_DIR env var or $1.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROMPT_DIR="${PROMPT_DIR:-${1:-$REPO_ROOT/prompts}}"

# Resolve to absolute path
PROMPT_DIR="$(cd "$PROMPT_DIR" && pwd)"

# Validate prompt directory contains required files
if [ ! -f "$PROMPT_DIR/EXECUTE.md" ]; then
    echo "ERROR: $PROMPT_DIR does not contain EXECUTE.md." >&2
    echo "Set PROMPT_DIR or pass the prompt directory as an argument." >&2
    exit 1
fi

log_and_print "=== LOOP STARTING ==="
log_and_print "Working directory: $(pwd)"
log_and_print "Prompt directory:  $PROMPT_DIR"
log_and_print "Max iterations:    $MAX_ITERATIONS"

# ─── Helper: delegate a prompt to copilot ────────────────────────────────────

delegate() {
    local prompt_name="$1"
    local prompt_file="$PROMPT_DIR/$prompt_name"

    if [ ! -f "$prompt_file" ]; then
        log_and_print "WARNING: Prompt file $prompt_file not found, skipping."
        return 1
    fi

    log "Delegating: $prompt_name"
    yes | copilot -p "$(cat "$prompt_file")" --allow-all-tools --deny-tool sudo
    local rc=$?
    log "Delegation complete: $prompt_name (exit code: $rc)"
    return $rc
}

# ─── Helper: run tests and return pass/fail ──────────────────────────────────

run_tests() {
    log "Running tests..."
    set +e
    xvfb-run go test -race -count=1 ./... 2>&1 | tee "$TEST_OUTPUT"
    local rc=${PIPESTATUS[0]}
    set -e

    if [ "$rc" -eq 0 ]; then
        log "Tests PASSED"
        return 0
    else
        log "Tests FAILED (exit code: $rc)"
        return 1
    fi
}

# ─── Helper: check if backlog is empty ───────────────────────────────────────

backlog_empty() {
    # Backlog is empty when AUDIT.md and PLAN.md are both absent or empty
    local audit_exists=false
    local plan_exists=false

    if [ -f "AUDIT.md" ] && [ -s "AUDIT.md" ]; then
        audit_exists=true
    fi
    if [ -f "PLAN.md" ] && [ -s "PLAN.md" ]; then
        plan_exists=true
    fi

    if [ "$audit_exists" = false ] && [ "$plan_exists" = false ]; then
        return 0  # backlog is empty
    fi
    return 1  # backlog has items
}

# ─── Helper: run analysis and extract metrics ────────────────────────────────

run_analysis() {
    # Run go-stats-generator analysis and capture JSON output
    # Returns 0 if analysis succeeded, 1 otherwise
    set +e
    go-stats-generator analyze . --skip-tests --format json --output "$ANALYSIS_OUTPUT" 2>/dev/null
    local rc=$?
    set -e
    return $rc
}

# ─── Helper: check complexity thresholds ─────────────────────────────────────

needs_breakdown() {
    # Check if any functions exceed complexity thresholds
    if [ ! -f "$ANALYSIS_OUTPUT" ]; then
        return 1  # no data, skip
    fi

    # Check for functions with overall > 10.0, lines > 30, or cyclomatic > 10
    # Use jq to inspect the analysis JSON
    set +e
    local violators
    violators=$(jq -r '
        [.. | objects | select(
            (.overall_complexity // 0) > 10.0 or
            (.lines // .code_lines // 0) > 30 or
            (.cyclomatic_complexity // 0) > 10
        )] | length
    ' "$ANALYSIS_OUTPUT" 2>/dev/null)
    set -e

    if [ -n "$violators" ] && [ "$violators" -gt 0 ] 2>/dev/null; then
        log "Complexity violations found: $violators functions exceed thresholds"
        return 0
    fi
    return 1
}

needs_deduplication() {
    # Check if duplication ratio exceeds 5% or clone pairs > 10
    if [ ! -f "$ANALYSIS_OUTPUT" ]; then
        return 1
    fi

    set +e
    local dup_ratio
    dup_ratio=$(jq -r '
        (.duplication_ratio // .duplication.ratio // 0)
    ' "$ANALYSIS_OUTPUT" 2>/dev/null)

    local clone_pairs
    clone_pairs=$(jq -r '
        (.clone_pairs // .duplication.clone_pairs // []) | length
    ' "$ANALYSIS_OUTPUT" 2>/dev/null)
    set -e

    # Check duplication ratio > 5%
    if [ -n "$dup_ratio" ] && [ "$dup_ratio" != "null" ] && [ "$dup_ratio" != "0" ]; then
        local exceeds
        exceeds=$(echo "$dup_ratio > 5.0" | bc -l 2>/dev/null || echo "0")
        if [ "$exceeds" = "1" ]; then
            log "Duplication ratio $dup_ratio% exceeds 5% threshold"
            return 0
        fi
    fi

    # Check clone pairs > 10
    if [ -n "$clone_pairs" ] && [ "$clone_pairs" != "null" ] && [ "$clone_pairs" -gt 10 ] 2>/dev/null; then
        log "Clone pairs ($clone_pairs) exceed threshold of 10"
        return 0
    fi

    return 1
}

needs_reorganization() {
    # Check if any package has cohesion < 0.3 or coupling > 0.7
    if [ ! -f "$ANALYSIS_OUTPUT" ]; then
        return 1
    fi

    set +e
    local problem_packages
    problem_packages=$(jq -r '
        [.. | objects | select(
            (has("cohesion") and .cohesion < 0.3) or
            (has("coupling") and .coupling > 0.7)
        )] | length
    ' "$ANALYSIS_OUTPUT" 2>/dev/null)
    set -e

    if [ -n "$problem_packages" ] && [ "$problem_packages" -gt 0 ] 2>/dev/null; then
        log "Package organization issues found: $problem_packages packages below thresholds"
        return 0
    fi
    return 1
}

# ─── 2. One-Time Initialization ──────────────────────────────────────────────

log_and_print "--- Initialization ---"

# 2a. AUDIT.md check
if [ ! -f "AUDIT.md" ] || [ ! -s "AUDIT.md" ]; then
    log_and_print "AUDIT.md not found — delegating BASE_AUDIT.md..."
    delegate "BASE_AUDIT.md" || true

    if [ -f "AUDIT.md" ] && [ -s "AUDIT.md" ]; then
        log_and_print "AUDIT.md created successfully."
    else
        log_and_print "WARNING: AUDIT.md was not created by BASE_AUDIT.md delegation."
    fi
else
    log_and_print "AUDIT.md already exists — skipping BASE_AUDIT.md."
fi

# 2b. PLAN.md check
if [ ! -f "PLAN.md" ] || [ ! -s "PLAN.md" ]; then
    log_and_print "PLAN.md not found — delegating MAKE_PLAN.md..."
    delegate "MAKE_PLAN.md" || true

    if [ -f "PLAN.md" ] && [ -s "PLAN.md" ]; then
        log_and_print "PLAN.md created successfully."
    else
        log_and_print "WARNING: PLAN.md was not created by MAKE_PLAN.md delegation."
    fi
else
    log_and_print "PLAN.md already exists — skipping MAKE_PLAN.md."
fi

# ─── 3. Main Loop ────────────────────────────────────────────────────────────

log_and_print "--- Entering main loop (max $MAX_ITERATIONS iterations) ---"

while [ "$ITERATION" -lt "$MAX_ITERATIONS" ]; do
    ITERATION=$((ITERATION + 1))
    log_and_print ""
    log_and_print "=== Iteration $ITERATION / $MAX_ITERATIONS ==="

    # ── Step A: Execute next task ────────────────────────────────────────
    log "Step A: Executing next task..."
    if delegate "EXECUTE.md"; then
        TASKS_EXECUTED=$((TASKS_EXECUTED + 1))
    else
        log_and_print "WARNING: EXECUTE.md delegation returned non-zero."
        TASKS_EXECUTED=$((TASKS_EXECUTED + 1))
    fi

    # ── Step B: Run tests and handle failures ────────────────────────────
    log "Step B: Running tests..."
    if run_tests; then
        FINAL_TEST_STATUS="PASS"
        iteration_summary "EXECUTE.md" "PASS"
    else
        FINAL_TEST_STATUS="FAIL"
        TEST_FAILURES=$((TEST_FAILURES + 1))
        log_and_print "Tests failed — delegating FAIL.md to diagnose and fix..."

        delegate "FAIL.md" || true

        # Re-run tests after FAIL.md fix attempt
        if run_tests; then
            FINAL_TEST_STATUS="PASS"
            iteration_summary "EXECUTE.md + FAIL.md" "PASS (fixed)"
        else
            FINAL_TEST_STATUS="FAIL"
            log_and_print "Tests still failing after FAIL.md — continuing to next iteration."
            iteration_summary "EXECUTE.md + FAIL.md" "FAIL (persisted)"
        fi
    fi

    # ── Step C: Periodic maintenance (every Nth iteration) ───────────────
    if [ $((ITERATION % MAINTENANCE_INTERVAL)) -eq 0 ]; then
        log_and_print "Step C: Periodic maintenance check (iteration $ITERATION)..."

        # Run analysis to gather current metrics
        if run_analysis; then
            # C1: Function breakdown
            if needs_breakdown; then
                log_and_print "  Complexity violations detected — delegating BREAKDOWN.md..."
                delegate "BREAKDOWN.md" || true
                TASKS_EXECUTED=$((TASKS_EXECUTED + 1))

                if ! run_tests; then
                    FINAL_TEST_STATUS="FAIL"
                    TEST_FAILURES=$((TEST_FAILURES + 1))
                    log_and_print "  Tests failed after BREAKDOWN.md — delegating FAIL.md..."
                    delegate "FAIL.md" || true
                    run_tests && FINAL_TEST_STATUS="PASS" || FINAL_TEST_STATUS="FAIL"
                fi
                iteration_summary "BREAKDOWN.md" "$FINAL_TEST_STATUS"
            fi

            # C2: Deduplication
            if needs_deduplication; then
                log_and_print "  Duplication threshold exceeded — delegating DEDUPLICATE.md..."
                delegate "DEDUPLICATE.md" || true
                TASKS_EXECUTED=$((TASKS_EXECUTED + 1))

                if ! run_tests; then
                    FINAL_TEST_STATUS="FAIL"
                    TEST_FAILURES=$((TEST_FAILURES + 1))
                    log_and_print "  Tests failed after DEDUPLICATE.md — delegating FAIL.md..."
                    delegate "FAIL.md" || true
                    run_tests && FINAL_TEST_STATUS="PASS" || FINAL_TEST_STATUS="FAIL"
                fi
                iteration_summary "DEDUPLICATE.md" "$FINAL_TEST_STATUS"
            fi

            # C3: Package reorganization
            if needs_reorganization; then
                log_and_print "  Package organization issues — delegating ORGANIZE.md..."
                delegate "ORGANIZE.md" || true
                TASKS_EXECUTED=$((TASKS_EXECUTED + 1))

                if ! run_tests; then
                    FINAL_TEST_STATUS="FAIL"
                    TEST_FAILURES=$((TEST_FAILURES + 1))
                    log_and_print "  Tests failed after ORGANIZE.md — delegating FAIL.md..."
                    delegate "FAIL.md" || true
                    run_tests && FINAL_TEST_STATUS="PASS" || FINAL_TEST_STATUS="FAIL"
                fi
                iteration_summary "ORGANIZE.md" "$FINAL_TEST_STATUS"
            fi
        else
            log_and_print "  go-stats-generator analysis failed — skipping maintenance checks."
        fi

        # Clean up analysis output
        rm -f "$ANALYSIS_OUTPUT"
    fi

    # ── Step D: Progress check ───────────────────────────────────────────
    if backlog_empty; then
        log_and_print "Backlog is empty (AUDIT.md and PLAN.md absent/empty)."

        # Run final test pass to confirm
        if run_tests; then
            FINAL_TEST_STATUS="PASS"
            log_and_print ""
            log_and_print "=== LOOP COMPLETE ==="
            log_and_print "Iterations: $ITERATION"
            log_and_print "Tasks executed: $TASKS_EXECUTED"
            log_and_print "Test failures encountered: $TEST_FAILURES"
            log_and_print "Final test status: PASS"
            rm -f "$TEST_OUTPUT" "$ANALYSIS_OUTPUT"
            exit 0
        else
            FINAL_TEST_STATUS="FAIL"
            TEST_FAILURES=$((TEST_FAILURES + 1))
            log_and_print "Backlog empty but tests failing — delegating FAIL.md..."
            delegate "FAIL.md" || true

            if run_tests; then
                FINAL_TEST_STATUS="PASS"
                log_and_print ""
                log_and_print "=== LOOP COMPLETE ==="
                log_and_print "Iterations: $ITERATION"
                log_and_print "Tasks executed: $TASKS_EXECUTED"
                log_and_print "Test failures encountered: $TEST_FAILURES"
                log_and_print "Final test status: PASS"
                rm -f "$TEST_OUTPUT" "$ANALYSIS_OUTPUT"
                exit 0
            else
                log_and_print "Tests still failing after FAIL.md on empty backlog — continuing."
            fi
        fi
    fi
done

# ─── 5. Max iterations reached ───────────────────────────────────────────────

# Run one final test to get the definitive status
run_tests && FINAL_TEST_STATUS="PASS" || FINAL_TEST_STATUS="FAIL"

log_and_print ""
log_and_print "=== LOOP COMPLETE (max iterations reached) ==="
log_and_print "Iterations: $ITERATION"
log_and_print "Tasks executed: $TASKS_EXECUTED"
log_and_print "Test failures encountered: $TEST_FAILURES"
log_and_print "Final test status: $FINAL_TEST_STATUS"

rm -f "$ANALYSIS_OUTPUT"

if [ "$FINAL_TEST_STATUS" = "PASS" ]; then
    exit 0
else
    exit 1
fi
