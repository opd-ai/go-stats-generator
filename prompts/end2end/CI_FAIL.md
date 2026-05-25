# AUTONOMOUS E2E TEST FAILURE FIX AGENT

## Execution Mode
**FULLY AUTONOMOUS** — Automatically diagnose, fix, validate, and commit ALL E2E test failures.

## Objective
Fix ALL end-to-end (E2E) test failures autonomously. E2E tests validate component integration, so failures indicate integration issues, flaky tests, or environmental problems. This agent will autonomously resolve them.

**NO MANUAL INTERVENTION REQUIRED. FIX EVERY FAILING E2E TEST AUTONOMOUSLY.**

## Execution Strategy

### PHASE 1: IDENTIFY E2E TEST FAILURES

1. **Locate E2E test step:**
   - Find which CI step runs E2E tests (look for "E2E", "integration", "end-to-end", etc.)
   - Identify which test runner is used (Jest, pytest, Go test, etc.)

2. **Retrieve failure logs:**
   - Extract full E2E test output from CI
   - Parse which specific tests failed
   - Extract error messages, stack traces, assertion failures

3. **Categorize failures:**
   - Flaky test (timing-dependent)?
   - External dependency failure (service, database)?
   - Component integration issue?
   - Configuration/setup issue?
   - Resource conflict?

### PHASE 2: DIAGNOSE ROOT CAUSE

For EACH failing E2E test:

1. **Run test locally:**
   - Execute the exact failing test locally multiple times
   - Does it fail consistently or intermittently?
   - Can you reproduce the failure?

2. **Analyze failure pattern:**
   - **External Service Dependency:** Test expects service (API, database, etc.) to be available
   - **Timing/Race Condition:** Test fails due to timing issues or race conditions
   - **State Management:** Test data not in expected state, previous test didn't clean up
   - **Resource Conflict:** Port already in use, memory exhausted, file locks
   - **Order Dependency:** Test expects other tests to run first
   - **Component Integration:** Components don't work together correctly

3. **Extract diagnostic info:**
   - Full error message
   - Stack trace
   - Test setup/teardown code
   - Dependencies (services, databases, ports)
   - Timing windows

### PHASE 3: AUTONOMOUSLY FIX EACH E2E TEST FAILURE

#### FIX TYPE 1: EXTERNAL SERVICE DEPENDENCY
- Check if service is required (API, database, cache, etc.)
- Option 1: Mock/stub the service in the test
- Option 2: Use test fixtures or fake implementations
- Option 3: Add retry logic with backoff
- Option 4: Update CI config to start service before tests
- Implement appropriate fix

#### FIX TYPE 2: FLAKY TEST - TIMING ISSUES
- Identify what is timing-dependent (operations, HTTP requests, etc.)
- Increase timeouts to be more generous
- Add explicit waits for conditions instead of fixed delays
- Use polling with timeout instead of one-shot checks
- Remove/fix race conditions in test code
- Add synchronization mechanisms (semaphores, channels, etc.)

#### FIX TYPE 3: FLAKY TEST - ORDER DEPENDENCY
- Make each test independent
- Move state initialization into each test's setup
- Remove assumptions about test execution order
- Clean up properly in test teardown
- Use test isolation (separate data per test, rollback transactions, etc.)

#### FIX TYPE 4: STATE MANAGEMENT ISSUE
- Improve test setup to ensure proper initial state
- Improve test teardown to clean up after test
- Use transaction rollback for database tests
- Isolate test data (separate databases, buckets, etc.)
- Fix previous tests that don't clean up properly

#### FIX TYPE 5: RESOURCE CONFLICT
- Identify conflict (port, file, memory, etc.)
- Use dynamic port allocation instead of hard-coded ports
- Add resource cleanup in test teardown
- Increase resource limits (if environment-controlled)
- Run conflicting tests serially instead of parallel

#### FIX TYPE 6: CONFIGURATION/SETUP ISSUE
- Verify test configuration is correct
- Check environment variables are set correctly
- Verify test fixtures exist and are valid
- Check if service needs to be started
- Update CI config if needed

#### FIX TYPE 7: COMPONENT INTEGRATION ISSUE
- Read the actual test to understand what it's testing
- Review the components being tested
- Find the actual integration issue (usually a mismatch in API, data format, etc.)
- Fix the component implementation or test expectations
- Verify components work together correctly

### PHASE 4: VALIDATE FIXES

For EACH fix:

1. **Local validation:**
   - Run the failing test multiple times (at least 5 times for flaky tests)
   - Verify it passes consistently
   - Run related E2E tests to ensure no regression

2. **Commit and push:**
   - Stage changes: `git add .`
   - Commit with descriptive message: `git commit -m "Fix E2E: [description of what was fixed]"`
   - Push to branch

3. **Re-run CI:**
   - Trigger E2E test step in CI
   - Wait for completion
   - Verify previously-failed tests now pass
   - If new failures appear, loop back to PHASE 2

### PHASE 5: FINAL VALIDATION

Once ALL E2E test failures are fixed:

1. Verify ALL E2E tests pass consistently
2. Verify NO new failures introduced
3. Run full test suite to ensure no regressions
4. Document fixes applied

## CRITICAL SUCCESS CRITERIA

✅ **MUST ACHIEVE:**
- ALL E2E tests pass consistently (not just once)
- Flaky tests are truly fixed, not just passing randomly
- External dependencies properly mocked/handled
- NO manual intervention required
- Clear git history of what was fixed

## EXECUTION CHECKLIST

- [ ] Identify all failing E2E tests
- [ ] Extract failure logs and diagnostic info
- [ ] For EACH failing test:
  - [ ] Run test locally to understand failure
  - [ ] Diagnose root cause
  - [ ] Implement fix
  - [ ] Validate fix (run multiple times)
  - [ ] Commit fix with clear message
- [ ] Re-run full E2E suite in CI
- [ ] Verify ALL tests pass
- [ ] Document summary of fixes

## KEY RULES

1. **Be autonomous:** Do not ask for guidance. Diagnose and fix.
2. **Be thorough:** Fix flaky tests properly, not just mask them.
3. **Don't remove tests:** Fix the underlying issues instead.
4. **Don't skip fixtures:** Understand what the test needs and provide it.
5. **Test multiple times:** Run flaky tests at least 5 times to verify fix.
6. **Proper timeouts:** Use generous timeouts and explicit waits, not hard delays.
7. **Clean state:** Ensure proper setup/teardown to avoid state leakage.
8. **Committed to success:** Fix everything until ALL tests pass consistently.
