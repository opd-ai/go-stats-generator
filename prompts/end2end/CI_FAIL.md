# TASK: Analyze CI/CD failures specific to end-to-end testing scenarios

## Execution Mode
**Report generation only** — do NOT modify any source code unless specifically instructed.

## Objective
Analyze CI/CD pipeline failures that occur during end-to-end (E2E) testing phases. E2E tests validate that multiple components work together correctly, so failures often indicate integration issues rather than unit-level bugs.

## Key Differences from Unit Tests

- **Scope:** E2E tests exercise multiple components or subsystems together
- **Failure patterns:** Often environmental, timing-dependent, or integration-related
- **Root causes:** May include flaky tests, external service dependencies, configuration issues
- **Debugging:** Requires understanding of component interactions and system state

## Workflow

### Phase 1: Identify E2E Test Failures

1. **Locate E2E test step:** Find which CI step runs E2E tests (may be labeled as "integration tests", "E2E tests", "end-to-end", etc.)

2. **Gather failure information:**
   - Which specific E2E tests failed?
   - Did they fail consistently or intermittently?
   - Did failures appear after recent code changes?
   - Are failures related to specific environments (OS, architecture)?

3. **Check failure frequency:** 
   - Is this a new failure?
   - Do other branches have the same failure?
   - Is this a known flaky test?

### Phase 2: Analyze E2E Failure Types

E2E tests can fail for different reasons than unit tests:

1. **Component integration issues** — Components don't work together correctly
2. **External dependency issues** — Database, API, service not available/responding
3. **Flaky tests** — Timing-dependent, race conditions in parallel execution
4. **Environmental issues** — Missing config, wrong setup, resource constraints
5. **Configuration issues** — Test fixtures not properly initialized
6. **Data issues** — Test data not in expected state
7. **Port/resource conflicts** — Port already in use, insufficient memory

### Phase 3: Create Resolution Plan

For each E2E failure:

1. **Reproduce locally:** Can you reproduce the failure consistently on your machine?
   - If yes: Likely a code or configuration issue
   - If no: Likely environmental, timing, or flaky

2. **Isolate the failure:**
   - Is it a specific test or all E2E tests?
   - Does it fail every time or intermittently?
   - Does it fail on all branches or just specific ones?

3. **Determine root cause:**
   - Component integration issue?
   - External service dependency?
   - Timing/race condition?
   - Configuration or setup issue?

4. **Implement fix:**
   - Fix the underlying code issue
   - Or adjust test to handle timing/ordering
   - Or mock/stub external dependency
   - Or improve test setup/teardown

5. **Validate:**
   - Run E2E tests locally multiple times
   - Run in CI to confirm fix
   - Monitor for regression

## Common E2E Test Failure Patterns

### Pattern 1: External Service Dependency
- Test expects external service to be available
- Test fails when service is down/slow
- **Solution:** Mock the service, use test fixtures, or add retry logic

### Pattern 2: Flaky Due to Timing
- Test assumes operations complete within timeout
- Test fails intermittently under load
- **Solution:** Increase timeouts, add explicit waits, use polling

### Pattern 3: State Initialization
- Test expects data/resources to exist
- Previous test didn't clean up properly
- **Solution:** Improve setup/teardown, use transaction rollback, isolate test state

### Pattern 4: Resource Exhaustion
- Test tries to allocate resources but none available
- Multiple tests run in parallel competing for resources
- **Solution:** Reduce resource usage, run tests serially, increase resource limits

### Pattern 5: Order Dependencies
- Test 1 must run before Test 2
- Tests assumed to run in specific order
- **Solution:** Make tests independent, don't rely on state from other tests

## Output Format

```markdown
# E2E Test Failure Analysis

## Summary
[Brief overview of which E2E tests are failing and why]

## Failures
[List each failing test with error message and logs]

## Root Cause Analysis
[What is causing the failures]

## Resolution Plan
[Specific steps to fix]

## Validation
[How to verify the fix]
```

## Tips for E2E Debugging

- Run tests multiple times to determine if flaky
- Check CI logs for timing/ordering information
- Look for resource contention (port conflicts, memory, etc.)
- Verify external service dependencies are available
- Check test setup/teardown for state management issues
- Consider running tests in isolation vs. parallel
- Check for hard-coded timeouts that may be too short
- Review recent changes to test fixtures or setup code
