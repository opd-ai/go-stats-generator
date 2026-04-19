#!/usr/bin/env bash
# audit-loop.sh — Systematic audit-driven Go project hardening orchestrator
#
# Runs every *AUDIT.md prompt in a logical order (implementation gaps first,
# product completeness last), delegating each audit exactly once and
# immediately following it with an EXECUTE.md call so that findings are acted
# upon before moving to the next dimension of quality.
#
# Usage:
#   bash scripts/audit-loop.sh [PROMPT_DIR]
#
# Arguments:
#   PROMPT_DIR  Path to the directory containing prompt .md files
#               (default: prompts/ directory at the repository root)
#
# Environment:
#   PROMPT_DIR  Override the prompt directory (takes precedence over $1)
#
# The script is idempotent on re-run and safe to interrupt (Ctrl+C).

set -euo pipefail

# ─── Trap: graceful shutdown on Ctrl+C / SIGTERM ─────────────────────────────

cleanup() {
    echo ""
    echo "=== AUDIT LOOP INTERRUPTED ==="
    echo "Audits completed: $AUDITS_COMPLETED"
    echo "Tasks executed:   $TASKS_EXECUTED"
    echo "Test failures:    $TEST_FAILURES"
    echo "Re-run this script to resume from current state."
    exit 130
}
trap cleanup INT TERM

# ─── Constants and defaults ───────────────────────────────────────────────────

LOG_FILE="audit-loop.log"
TEST_OUTPUT="test-output.txt"

# Counters
AUDITS_COMPLETED=0
TASKS_EXECUTED=0
TEST_FAILURES=0
FINAL_TEST_STATUS="UNKNOWN"
LAST_TEST_RC=0

# ─── Audit sequence: implementation gaps → product completeness ───────────────
#
# Logical progression toward production software:
#   1. What is not yet implemented (gaps)
#   2. Foundational health: dependencies, resource/memory/concurrency/error safety
#   3. Networking correctness
#   4. Security (standard then adversarial red-team)
#   5. Performance
#   6. Observability / logging
#   7. API design quality
#   8. Go best practices
#   9. Test quality and test-code metrics
#  10. Maintainability / technical debt
#  11. Goal-focused functional metrics (production code then test code)
#  12. Per-package deep dive
#  13. Consolidated findings
#  14. Overall functional audit against stated goals
#  15. Product completeness (final gate)

AUDIT_SEQUENCE=(
    "IMPLEMENTATION_GAP_AUDIT.md"
    "DEPENDENCY_AUDIT.md"
    "RESOURCE_AUDIT.md"
    "MEMORY_AUDIT.md"
    "SYNC_AUDIT.md"
    "ERROR_AUDIT.md"
    "NETWORK_AUDIT.md"
    "SECURITY_AUDIT.md"
    "REDTEAM_AUDIT.md"
    "PERFORMANCE_AUDIT.md"
    "LOGGING_AUDIT.md"
    "API_AUDIT.md"
    "BEST_PRACTICES_AUDIT.md"
    "TESTING_AUDIT.md"
    "TEST_PERFORM_AUDIT.md"
    "MAINTAINABILITY_AUDIT.md"
    "PERFORM_AUDIT.md"
    "META_AUDIT.md"
    "COLLATE_AUDIT.md"
    "BASE_AUDIT.md"
    "PRODUCT_COMPLETENESS_AUDIT.md"
)

# ─── Logging helpers ──────────────────────────────────────────────────────────

log() {
    local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $*"
    echo "$msg" >> "$LOG_FILE"
}

log_and_print() {
    local msg="$*"
    log "$msg"
    echo "$msg"
}

audit_summary() {
    local audit="$1"
    local test_status="$2"
    local line="[audit $AUDITS_COMPLETED/${#AUDIT_SEQUENCE[@]}] $audit → EXECUTE.md — tests: $test_status"
    log "$line"
    echo "$line"
}

# ─── 1. Entry Gate ────────────────────────────────────────────────────────────

if [ ! -f "ROADMAP.md" ]; then
    echo "ERROR: ROADMAP.md not found in $(pwd)." >&2
    echo "The target Go project must have a ROADMAP.md in its root directory." >&2
    exit 1
fi

# Resolve the prompt directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROMPT_DIR="${PROMPT_DIR:-${1:-$REPO_ROOT/prompts}}"

if [ ! -d "$PROMPT_DIR" ]; then
    echo "ERROR: Prompt directory '$PROMPT_DIR' does not exist or is not a directory." >&2
    echo "Set PROMPT_DIR or pass the prompt directory as an argument." >&2
    exit 1
fi
PROMPT_DIR="$(cd "$PROMPT_DIR" && pwd)"

if [ ! -f "$PROMPT_DIR/EXECUTE.md" ]; then
    echo "ERROR: $PROMPT_DIR does not contain EXECUTE.md." >&2
    echo "Set PROMPT_DIR or pass the prompt directory as an argument." >&2
    exit 1
fi

log_and_print "=== AUDIT LOOP STARTING ==="
log_and_print "Working directory: $(pwd)"
log_and_print "Prompt directory:  $PROMPT_DIR"
log_and_print "Audit sequence:    ${#AUDIT_SEQUENCE[@]} audits"

# ─── Helper: delegate a prompt to copilot ────────────────────────────────────

delegate() {
    local prompt_name="$1"
    local prompt_file="$PROMPT_DIR/$prompt_name"

    if [ ! -f "$prompt_file" ]; then
        log_and_print "WARNING: Prompt file $prompt_file not found, skipping."
        return 1
    fi

    log "Delegating: $prompt_name"
    set +e
    yes | copilot --model claude-opus-4.5 -p "$(cat "$prompt_file")" --allow-all-tools --deny-tool sudo
    local rc=${PIPESTATUS[1]}
    set -e
    log "Delegation complete: $prompt_name (exit code: $rc)"
    return $rc
}

# ─── Helper: run tests and return pass/fail ───────────────────────────────────

run_tests() {
    log "Running tests..."
    local rc
    set +e
    if command -v xvfb-run >/dev/null 2>&1; then
        # -a: auto-select a free server number to avoid :99 conflicts
        xvfb-run -a -- go test -race -count=1 ./... 2>&1 | tee "$TEST_OUTPUT"
        rc=${PIPESTATUS[0]}
        # If Xvfb itself failed to start, fall back to running without a display
        if grep -q 'xvfb-run: error:' "$TEST_OUTPUT"; then
            log "WARNING: xvfb-run failed to start Xvfb — retrying without display wrapper..."
            go test -race -count=1 ./... 2>&1 | tee "$TEST_OUTPUT"
            rc=${PIPESTATUS[0]}
        fi
    else
        log "WARNING: xvfb-run not available — running tests without display wrapper..."
        go test -race -count=1 ./... 2>&1 | tee "$TEST_OUTPUT"
        rc=${PIPESTATUS[0]}
    fi
    set -e
    LAST_TEST_RC="$rc"

    if [ "$rc" -eq 0 ]; then
        log "Tests PASSED"
        return 0
    else
        log "Tests FAILED (exit code: $rc)"
        return 1
    fi
}

is_relevant_test_failure() {
    # Exclude xvfb infrastructure errors — not actionable by FAIL.md.
    if [ -f "$TEST_OUTPUT" ] && grep -q 'xvfb-run: error:' "$TEST_OUTPUT"; then
        return 1
    fi
    # Actionable failures are real go test package/test failures.
    if [ "${LAST_TEST_RC:-0}" -eq 1 ] && [ -f "$TEST_OUTPUT" ] &&
        grep -Eq '^(--- FAIL:|FAIL[[:space:]\t])' "$TEST_OUTPUT"; then
        return 0
    fi
    return 1
}

ensure_relevant_test_failure_or_exit() {
    if is_relevant_test_failure; then
        return 0
    fi
    log_and_print "ERROR: Non-actionable test command failure (exit code: ${LAST_TEST_RC:-unknown}); skipping FAIL.md."
    log_and_print "Inspect $TEST_OUTPUT for details."
    exit "${LAST_TEST_RC:-1}"
}

# ─── 2. Audit sequence ────────────────────────────────────────────────────────

log_and_print "--- Beginning audit sequence ---"

for AUDIT_PROMPT in "${AUDIT_SEQUENCE[@]}"; do
    AUDITS_COMPLETED=$((AUDITS_COMPLETED + 1))
    log_and_print ""
    log_and_print "=== Audit $AUDITS_COMPLETED / ${#AUDIT_SEQUENCE[@]}: $AUDIT_PROMPT ==="

    # ── Step A: Run the audit ─────────────────────────────────────────────
    log "Step A: Running audit $AUDIT_PROMPT..."
    delegate "$AUDIT_PROMPT" || log_and_print "WARNING: $AUDIT_PROMPT delegation returned non-zero (continuing)."

    # ── Step B: Execute findings ──────────────────────────────────────────
    log "Step B: Executing findings from $AUDIT_PROMPT..."
    if delegate "EXECUTE.md"; then
        TASKS_EXECUTED=$((TASKS_EXECUTED + 1))
    else
        log_and_print "WARNING: EXECUTE.md delegation returned non-zero after $AUDIT_PROMPT."
        TASKS_EXECUTED=$((TASKS_EXECUTED + 1))
    fi

    # ── Step C: Run tests and handle failures ─────────────────────────────
    log "Step C: Running tests after $AUDIT_PROMPT + EXECUTE.md..."
    if run_tests; then
        FINAL_TEST_STATUS="PASS"
        audit_summary "$AUDIT_PROMPT" "PASS"
    else
        FINAL_TEST_STATUS="FAIL"
        TEST_FAILURES=$((TEST_FAILURES + 1))
        ensure_relevant_test_failure_or_exit
        log_and_print "Tests failed — delegating FAIL.md to diagnose and fix..."

        delegate "FAIL.md" || true

        if run_tests; then
            FINAL_TEST_STATUS="PASS"
            audit_summary "$AUDIT_PROMPT" "PASS (fixed)"
        else
            ensure_relevant_test_failure_or_exit
            FINAL_TEST_STATUS="FAIL"
            log_and_print "Tests still failing after FAIL.md — continuing to next audit."
            audit_summary "$AUDIT_PROMPT" "FAIL (persisted)"
        fi
    fi
done

# ─── 3. Final summary ─────────────────────────────────────────────────────────

# Run one final test to get the definitive status
run_tests && FINAL_TEST_STATUS="PASS" || FINAL_TEST_STATUS="FAIL"

log_and_print ""
log_and_print "=== AUDIT LOOP COMPLETE ==="
log_and_print "Audits completed: $AUDITS_COMPLETED"
log_and_print "Tasks executed:   $TASKS_EXECUTED"
log_and_print "Test failures:    $TEST_FAILURES"
log_and_print "Final test status: $FINAL_TEST_STATUS"

rm -f "$TEST_OUTPUT"

if [ "$FINAL_TEST_STATUS" = "PASS" ]; then
    exit 0
else
    exit 1
fi
