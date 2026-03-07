// Package placement provides test fixtures for code placement analysis.
//
// This package intentionally contains code organization anti-patterns to validate
// the go-stats-generator's ability to detect misplaced functions, misplaced methods,
// and low cohesion issues. The package demonstrates four key scenarios:
//
// 1. High Cohesion (high_cohesion/): Example of well-organized code where all
// declarations in a file are closely related. The User struct and all its methods
// are properly co-located, demonstrating the correct approach to code organization.
//
// 2. Low Cohesion (low_cohesion/): Example of poorly organized code where functions
// in a file primarily reference external symbols rather than each other. The
// mixed.go file contains functions that all call handlers from handlers.go,
// creating weak internal relationships and low cohesion scores.
//
// 3. Misplaced Functions (misplaced_function/): Example where functions have high
// affinity to files other than where they are defined. ValidateUser and ProcessUser
// in handler.go heavily reference the Database type and should be moved to
// database.go for better organization.
//
// 4. Misplaced Methods (misplaced_method/): Example where methods are defined in
// files separate from their receiver type. The User type is defined in user.go,
// but some of its methods (Validate, IsAdmin) are defined in validator.go,
// violating Go's convention of co-locating methods with their receiver types.
//
// The package is used by the placement analyzer tests to ensure correct detection
// of code organization issues. All anti-patterns are intentional and serve as
// negative test cases.
package placement
