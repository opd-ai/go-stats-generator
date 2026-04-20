#!/usr/bin/env bash
# lib.sh — Shared helpers for loop.sh and audit-loop.sh
#
# Sourced by both orchestrators.  Callers must define the following variables
# and functions BEFORE sourcing this file (so the runtime references resolve):
#
#   Variables:
#     TEST_OUTPUT     — path to the test output capture file
#     LAST_TEST_RC    — integer holding the last `go test` exit code
#
#   Functions:
#     log()           — writes a timestamped line to the script's log file
#     log_and_print() — log() + echo to stdout

# ─── xvfb-run error sentinel ─────────────────────────────────────────────────

XVFB_ERROR_PATTERN="xvfb-run: error:"

# ─── Helper: run tests under xvfb-run if available ──────────────────────────
#
# Sets LAST_TEST_RC.  Returns 0 on pass, 1 on failure.

run_tests() {
    log "Running tests..."
    local rc
    set +e
    if command -v xvfb-run >/dev/null 2>&1; then
        # -a: auto-select a free server number to avoid :99 conflicts
        xvfb-run -a -- go test -race -count=1 ./... 2>&1 | tee "$TEST_OUTPUT"
        rc=${PIPESTATUS[0]}
        # If Xvfb itself failed to start, fall back to running without a display
        if grep -q "$XVFB_ERROR_PATTERN" "$TEST_OUTPUT"; then
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

# ─── Helper: distinguish real test failures from infrastructure errors ────────
#
# Returns 0 (true) only when the test output contains actual go test failure
# lines and the failure is not caused by an xvfb startup error.

is_relevant_test_failure() {
    # Exclude xvfb infrastructure errors — not actionable by FAIL.md.
    if [ -f "$TEST_OUTPUT" ] && grep -q "$XVFB_ERROR_PATTERN" "$TEST_OUTPUT"; then
        return 1
    fi
    # Actionable failures are real go test package/test failures (exit codes 1+).
    if [ "${LAST_TEST_RC:-0}" -ne 0 ] && [ -f "$TEST_OUTPUT" ] &&
        grep -Eq '^(--- FAIL:|FAIL[[:space:]\t])' "$TEST_OUTPUT"; then
        return 0
    fi
    return 1
}

# ─── Helper: abort on non-actionable failure ─────────────────────────────────
#
# Call after run_tests returns non-zero.  Exits the whole script if the failure
# is not a real go test failure (e.g. xvfb crash, build error without FAIL
# lines) so the orchestrator does not pointlessly delegate FAIL.md.

ensure_relevant_test_failure_or_exit() {
    if is_relevant_test_failure; then
        return 0
    fi
    log_and_print "ERROR: Non-actionable test command failure (exit code: ${LAST_TEST_RC:-unknown}); skipping FAIL.md."
    log_and_print "Inspect $TEST_OUTPUT for details."
    exit "${LAST_TEST_RC:-1}"
}
