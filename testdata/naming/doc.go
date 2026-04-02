// Package naming is a test fixture for go-stats-generator's naming convention analyzer.
//
// This package intentionally contains both correct and incorrect naming patterns to validate
// the analyzer's ability to detect violations of Go naming conventions. The package serves
// as a comprehensive test suite for naming analysis features including:
//
//   - File naming violations (non-snake_case filenames)
//   - Identifier naming violations (underscore usage, incorrect MixedCaps)
//   - Acronym casing violations (Http vs HTTP, Url vs URL, Json vs JSON)
//   - Package stuttering detection (e.g., naming.NewNamingService)
//   - Generic package names (util, helper, common)
//   - Exported symbol documentation requirements
//
// # Test Fixtures
//
// The package contains the following test files:
//
//   - bad_file_name.go: Demonstrates snake_case compliant file naming (good example)
//   - bad_identifiers.go: Contains intentional naming violations for detection testing
//   - good_identifiers.go: Contains correct naming patterns as positive test cases
//
// # Multi-Package Issue
//
// Note: bad_identifiers.go declares "package util" while other files use "package naming".
// This violates Go's one-package-per-directory rule and is a known issue that prevents
// go vet from running on this directory. This multi-package scenario is itself being
// evaluated as a potential test case for package organization analysis.
//
// # Usage
//
// These files should NOT be modified to fix naming violations - the violations are
// intentional test data. Changes should only update the test fixtures to cover
// additional naming convention scenarios or fix the multi-package build issue.
package naming
