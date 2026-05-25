# CI Failure Analysis - go-stats-generator (End-to-End Tests)

## Overview

This file documents CI failures specific to the end-to-end testing suite for go-stats-generator. For the main CI failures and linting issues, see the root `/prompts/CI_FAIL.md` file.

## Recent Failures

When end-to-end tests fail, follow this resolution process:

1. **Identify the failure:** Check the GitHub Actions workflow logs to see which test suite failed
2. **Analyze root cause:** Determine if it's a code issue, configuration issue, or flaky test
3. **Document the issue:** Record file locations, error messages, and reproduction steps
4. **Implement fix:** Address the root cause based on the error category
5. **Validate:** Run the tests locally and in CI to confirm the fix

## Common End-to-End Test Failure Categories

### Integration Issues
- API endpoints returning unexpected responses
- Database connectivity problems
- Storage layer failures
- Multi-service interaction issues

### Configuration Issues
- Missing environment variables
- Invalid configuration values
- Incorrect test data setup

### Flaky Tests
- Race conditions in concurrent code
- Timing-dependent assertions
- External service dependencies

## Resolution Strategy

1. Check the most recent failed workflow run
2. Extract detailed error messages from CI logs
3. Categorize the failure type
4. Apply appropriate fix (see main CI_FAIL.md for code quality fixes)
5. Run end-to-end tests locally before pushing: `go test -race ./... -tags=integration`

## Validation Commands

```bash
# Run full test suite with race detection
make test

# Run specific integration tests
go test -race ./cmd -run TestAnalyze

# Run linting to check code quality
make lint

# Build the project
make build
```

## Notes

- Check if failures are intermittent (flaky) or consistent
- Review recent code changes that might have caused the failure
- Ensure test data and fixtures are up to date
