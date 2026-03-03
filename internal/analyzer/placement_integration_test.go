package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlacementAnalyzer_Integration_MisplacedFunction(t *testing.T) {
	// Test fixture: testdata/placement/misplaced_function/
	// ValidateUser in handler.go should be flagged as misplaced
	// because it heavily references Database which is defined in database.go

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "../../testdata/placement/misplaced_function", nil, parser.ParseComments)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	metrics := analyzer.Analyze(files, fset)

	// Should detect at least one misplaced function (ValidateUser)
	assert.Greater(t, metrics.MisplacedFunctions, 0, "should detect misplaced function")

	// Find the ValidateUser issue
	var foundValidateUser bool
	for _, issue := range metrics.FunctionIssues {
		if issue.Name == "ValidateUser" {
			foundValidateUser = true
			// Should be in handler.go currently
			assert.Contains(t, filepath.Base(issue.CurrentFile), "handler.go")
			// Should suggest database.go (or contain "database" in path)
			assert.Contains(t, filepath.Base(issue.SuggestedFile), "database.go")
			// Suggested affinity should be higher than current
			assert.Greater(t, issue.SuggestedAffinity, issue.CurrentAffinity)
			// Should have severity set
			assert.NotEmpty(t, issue.Severity)
		}
	}
	assert.True(t, foundValidateUser, "should find ValidateUser misplacement issue")
}

func TestPlacementAnalyzer_Integration_MisplacedMethod(t *testing.T) {
	// Test fixture: testdata/placement/misplaced_method/
	// Validate and IsAdmin methods on User are defined in validator.go
	// but should be in user.go with the User type definition

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "../../testdata/placement/misplaced_method", nil, parser.ParseComments)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	metrics := analyzer.Analyze(files, fset)

	// Should detect at least 2 misplaced methods (Validate and IsAdmin)
	assert.GreaterOrEqual(t, metrics.MisplacedMethods, 2, "should detect at least 2 misplaced methods")

	// Check that methods are flagged correctly
	methodsFound := make(map[string]bool)
	for _, issue := range metrics.MethodIssues {
		// Method names include receiver, e.g., "User.Validate"
		if issue.MethodName == "User.Validate" || issue.MethodName == "User.IsAdmin" {
			methodsFound[issue.MethodName] = true
			// Receiver type should be User
			assert.Equal(t, "User", issue.ReceiverType)
			// Current file should be validator.go
			assert.Contains(t, filepath.Base(issue.CurrentFile), "validator.go")
			// Receiver file should be user.go
			assert.Contains(t, filepath.Base(issue.ReceiverFile), "user.go")
			// Should be same package
			assert.Equal(t, "same_package", issue.Distance)
			// Should have severity
			assert.NotEmpty(t, issue.Severity)
		}
	}

	assert.True(t, methodsFound["User.Validate"], "should find Validate method issue")
	assert.True(t, methodsFound["User.IsAdmin"], "should find IsAdmin method issue")
}

func TestPlacementAnalyzer_Integration_LowCohesion(t *testing.T) {
	// Test fixture: testdata/placement/low_cohesion/
	// mixed.go contains unrelated User, Product, and Order declarations
	// File should have low cohesion score

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "../../testdata/placement/low_cohesion", nil, parser.ParseComments)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	metrics := analyzer.Analyze(files, fset)

	// Should detect at least one low cohesion file
	assert.Greater(t, metrics.LowCohesionFiles, 0, "should detect low cohesion file")

	// Find the mixed.go issue
	var foundMixed bool
	for _, issue := range metrics.CohesionIssues {
		if filepath.Base(issue.File) == "mixed.go" {
			foundMixed = true
			// Cohesion score should be below threshold (0.3)
			assert.Less(t, issue.CohesionScore, 0.3)
			// Should have suggested splits
			assert.Greater(t, len(issue.SuggestedSplits), 0, "should suggest file splits")
			// Should have severity
			assert.NotEmpty(t, issue.Severity)
		}
	}
	assert.True(t, foundMixed, "should find mixed.go cohesion issue")
}

func TestPlacementAnalyzer_Integration_HighCohesion(t *testing.T) {
	// Test fixture: testdata/placement/high_cohesion/
	// user.go contains only User-related declarations
	// File should have high cohesion score (>0.3) and NOT be flagged

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "../../testdata/placement/high_cohesion", nil, parser.ParseComments)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)

	var files []*ast.File
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	metrics := analyzer.Analyze(files, fset)

	// Average cohesion should be high
	assert.Greater(t, metrics.AvgFileCohesion, 0.3, "average cohesion should be above threshold")

	// user.go should NOT be in the low cohesion issues
	for _, issue := range metrics.CohesionIssues {
		assert.NotContains(t, filepath.Base(issue.File), "user.go",
			"user.go should not be flagged as low cohesion")
	}

	// Should have zero low cohesion files
	assert.Equal(t, 0, metrics.LowCohesionFiles, "should not detect any low cohesion files")
}

func TestPlacementAnalyzer_Integration_AllScenarios(t *testing.T) {
	// Integration test combining all scenarios to ensure they work together
	testCases := []struct {
		name                     string
		dir                      string
		expectMisplacedFunctions bool
		expectMisplacedMethods   bool
		expectLowCohesion        bool
	}{
		{
			name:                     "misplaced_function",
			dir:                      "../../testdata/placement/misplaced_function",
			expectMisplacedFunctions: true,
			expectMisplacedMethods:   false,
			expectLowCohesion:        false, // May have low cohesion but we test the function issue
		},
		{
			name:                     "misplaced_method",
			dir:                      "../../testdata/placement/misplaced_method",
			expectMisplacedFunctions: false,
			expectMisplacedMethods:   true,
			expectLowCohesion:        false, // May have low cohesion but we test the method issue
		},
		{
			name:                     "low_cohesion",
			dir:                      "../../testdata/placement/low_cohesion",
			expectMisplacedFunctions: false, // May have misplaced functions but we test cohesion
			expectMisplacedMethods:   false,
			expectLowCohesion:        true,
		},
		{
			name:                     "high_cohesion",
			dir:                      "../../testdata/placement/high_cohesion",
			expectMisplacedFunctions: false,
			expectMisplacedMethods:   false,
			expectLowCohesion:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			pkgs, err := parser.ParseDir(fset, tc.dir, nil, parser.ParseComments)
			require.NoError(t, err)

			var files []*ast.File
			for _, pkg := range pkgs {
				for _, file := range pkg.Files {
					files = append(files, file)
				}
			}

			analyzer := NewPlacementAnalyzer(0.25, 0.3)
			metrics := analyzer.Analyze(files, fset)

			if tc.expectMisplacedFunctions {
				assert.Greater(t, metrics.MisplacedFunctions, 0, "expected misplaced functions")
			}
			// Don't assert == 0 for other metrics since fixtures may have multiple issues

			if tc.expectMisplacedMethods {
				assert.Greater(t, metrics.MisplacedMethods, 0, "expected misplaced methods")
			}

			if tc.expectLowCohesion {
				assert.Greater(t, metrics.LowCohesionFiles, 0, "expected low cohesion files")
			}
		})
	}
}
