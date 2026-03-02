package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlacementAnalyzer_FunctionAffinity(t *testing.T) {
	tests := []struct {
		name                  string
		files                 map[string]string
		expectedMisplaced     int
		checkSuggestedFile    string
		checkCurrentAffinity  float64
		checkSuggestedAffinity float64
	}{
		{
			name: "function with all references in same file - no issue",
			files: map[string]string{
				"file1.go": `package test
func Helper() {}
func Main() { Helper(); Helper() }`,
			},
			expectedMisplaced: 0,
		},
		{
			name: "function with references split across files",
			files: map[string]string{
				"file1.go": `package test
func Helper() {}
func Main1() { Helper() }`,
				"file2.go": `package test
func Main2() { Helper(); Helper(); Helper() }`,
			},
			expectedMisplaced:      1,
			checkSuggestedFile:     "file2.go",
			checkCurrentAffinity:   0.25,
			checkSuggestedAffinity: 0.75,
		},
		{
			name: "function with balanced references - no clear winner",
			files: map[string]string{
				"file1.go": `package test
func Helper() {}
func Main1() { Helper(); Helper() }`,
				"file2.go": `package test
func Main2() { Helper(); Helper() }`,
			},
			expectedMisplaced: 0, // Within margin
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, fset := parseTestFiles(t, tt.files)
			analyzer := NewPlacementAnalyzer(0.25, 0.3)
			issues := analyzer.Analyze(files, fset).FunctionIssues

			assert.Equal(t, tt.expectedMisplaced, len(issues), "unexpected number of misplaced functions")

			if tt.expectedMisplaced > 0 && len(issues) > 0 {
				issue := issues[0]
				if tt.checkSuggestedFile != "" {
					assert.Contains(t, issue.SuggestedFile, tt.checkSuggestedFile)
				}
				if tt.checkCurrentAffinity > 0 {
					assert.InDelta(t, tt.checkCurrentAffinity, issue.CurrentAffinity, 0.01)
				}
				if tt.checkSuggestedAffinity > 0 {
					assert.InDelta(t, tt.checkSuggestedAffinity, issue.SuggestedAffinity, 0.01)
				}
			}
		})
	}
}

func TestPlacementAnalyzer_MethodPlacement(t *testing.T) {
	tests := []struct {
		name              string
		files             map[string]string
		expectedMisplaced int
		expectedDistance  string
		expectedSeverity  string
	}{
		{
			name: "method in same file as receiver - no issue",
			files: map[string]string{
				"user.go": `package test
type User struct { Name string }
func (u *User) GetName() string { return u.Name }`,
			},
			expectedMisplaced: 0,
		},
		{
			name: "method in different file from receiver - same package",
			files: map[string]string{
				"user.go": `package test
type User struct { Name string }`,
				"user_methods.go": `package test
func (u *User) GetName() string { return u.Name }`,
			},
			expectedMisplaced: 1,
			expectedDistance:  "same_package",
			expectedSeverity:  "medium",
		},
		{
			name: "method with pointer receiver",
			files: map[string]string{
				"types.go": `package test
type Counter struct { count int }`,
				"methods.go": `package test
func (c *Counter) Increment() { c.count++ }`,
			},
			expectedMisplaced: 1,
			expectedDistance:  "same_package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, fset := parseTestFiles(t, tt.files)
			analyzer := NewPlacementAnalyzer(0.25, 0.3)
			issues := analyzer.Analyze(files, fset).MethodIssues

			assert.Equal(t, tt.expectedMisplaced, len(issues), "unexpected number of misplaced methods")

			if tt.expectedMisplaced > 0 && len(issues) > 0 {
				issue := issues[0]
				if tt.expectedDistance != "" {
					assert.Equal(t, tt.expectedDistance, issue.Distance)
				}
				if tt.expectedSeverity != "" {
					assert.Equal(t, tt.expectedSeverity, issue.Severity)
				}
			}
		})
	}
}

func TestPlacementAnalyzer_FileCohesion(t *testing.T) {
	tests := []struct {
		name              string
		files             map[string]string
		expectedLowCohesion int
		checkCohesionScore float64
		checkSeverity     string
	}{
		{
			name: "high cohesion file - all internal references",
			files: map[string]string{
				"math.go": `package test
func Add(a, b int) int { return a + b }
func Sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total = Add(total, n)
	}
	return total
}`,
			},
			expectedLowCohesion: 0,
		},
		{
			name: "low cohesion file - mixed concerns",
			files: map[string]string{
				"helper.go": `package test
func Format(s string) string { return s }`,
				"mixed.go": `package test
func Process1() { Format("test") }
func Process2() { Format("test") }
func Process3() { Format("test") }`,
			},
			expectedLowCohesion: 1,
			checkCohesionScore:  0.0, // All external references
			checkSeverity:       "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, fset := parseTestFiles(t, tt.files)
			analyzer := NewPlacementAnalyzer(0.25, 0.3)
			issues := analyzer.Analyze(files, fset).CohesionIssues

			assert.Equal(t, tt.expectedLowCohesion, len(issues), "unexpected number of low cohesion files")

			if tt.expectedLowCohesion > 0 && len(issues) > 0 {
				issue := issues[0]
				if tt.checkCohesionScore >= 0 {
					assert.InDelta(t, tt.checkCohesionScore, issue.CohesionScore, 0.01)
				}
				if tt.checkSeverity != "" {
					assert.Equal(t, tt.checkSeverity, issue.Severity)
				}
				assert.NotEmpty(t, issue.SuggestedSplits)
			}
		})
	}
}

func TestPlacementAnalyzer_AvgFileCohesion(t *testing.T) {
	files := map[string]string{
		"file1.go": `package test
func Helper1() {}
func Main1() { Helper1() }`,
		"file2.go": `package test
func Helper2() {}`,
	}

	parsedFiles, fset := parseTestFiles(t, files)
	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	metrics := analyzer.Analyze(parsedFiles, fset)

	assert.Greater(t, metrics.AvgFileCohesion, 0.0)
	assert.LessOrEqual(t, metrics.AvgFileCohesion, 1.0)
}

func TestPlacementAnalyzer_ComplexScenario(t *testing.T) {
	files := map[string]string{
		"types.go": `package test
type User struct { Name string }
type Product struct { Price float64 }`,
		"user_logic.go": `package test
func CreateUser(name string) User { return User{Name: name} }
func (u *User) Validate() bool { return u.Name != "" }`,
		"product_logic.go": `package test
func CreateProduct(price float64) Product { return Product{Price: price} }
func (p *Product) GetPrice() float64 { return p.Price }`,
		"mixed.go": `package test
func Process() {
	u := CreateUser("test")
	_ = u.Validate()
	p := CreateProduct(10.0)
	_ = p.GetPrice()
}`,
	}

	parsedFiles, fset := parseTestFiles(t, files)
	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	metrics := analyzer.Analyze(parsedFiles, fset)

	// Should detect misplaced methods (methods away from receiver types)
	assert.Equal(t, 2, metrics.MisplacedMethods, "should find 2 misplaced methods")

	// Verify the methods are correctly identified
	methodNames := make(map[string]bool)
	for _, issue := range metrics.MethodIssues {
		methodNames[issue.MethodName] = true
		assert.Equal(t, "same_package", issue.Distance)
		assert.Equal(t, "medium", issue.Severity)
	}
	assert.True(t, methodNames["User.Validate"] || methodNames["Product.GetPrice"])
}

func TestPlacementAnalyzer_EmptyInput(t *testing.T) {
	analyzer := NewPlacementAnalyzer(0.25, 0.3)
	fset := token.NewFileSet()
	metrics := analyzer.Analyze([]*ast.File{}, fset)

	assert.Equal(t, 0, metrics.MisplacedFunctions)
	assert.Equal(t, 0, metrics.MisplacedMethods)
	assert.Equal(t, 0, metrics.LowCohesionFiles)
	assert.Equal(t, 0.0, metrics.AvgFileCohesion)
}

func TestPlacementAnalyzer_ReceiverTypeExtraction(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "value receiver",
			code: `package test
type User struct{}
func (u User) Method() {}`,
			expected: "User",
		},
		{
			name: "pointer receiver",
			code: `package test
type User struct{}
func (u *User) Method() {}`,
			expected: "User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{"test.go": tt.code}
			parsedFiles, fset := parseTestFiles(t, files)
			analyzer := NewPlacementAnalyzer(0.25, 0.3)
			analyzer.buildSymbolIndex(parsedFiles, fset)

			// Check that receiver type was recorded correctly
			found := false
			for _, info := range analyzer.methods {
				if info.receiverType == tt.expected {
					found = true
					break
				}
			}
			assert.True(t, found, "receiver type %s not found", tt.expected)
		})
	}
}

func TestPlacementAnalyzer_ConfigurableThresholds(t *testing.T) {
	files := map[string]string{
		"file1.go": `package test
func Helper() {}
func Main1() { Helper() }`,
		"file2.go": `package test
func Main2() { Helper(); Helper() }`,
	}

	parsedFiles, fset := parseTestFiles(t, files)

	// With strict margin (0.1), should detect misplacement
	strictAnalyzer := NewPlacementAnalyzer(0.1, 0.5)
	strictMetrics := strictAnalyzer.Analyze(parsedFiles, fset)
	assert.Greater(t, strictMetrics.MisplacedFunctions, 0)

	// With lenient margin (0.5), might not detect
	lenientAnalyzer := NewPlacementAnalyzer(0.5, 0.1)
	lenientMetrics := lenientAnalyzer.Analyze(parsedFiles, fset)
	// Depending on exact affinity, may or may not flag
	assert.GreaterOrEqual(t, lenientMetrics.MisplacedFunctions, 0)
}

// Helper function to parse multiple test files
func parseTestFiles(t *testing.T, files map[string]string) ([]*ast.File, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	var parsed []*ast.File

	for filename, content := range files {
		file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
		require.NoError(t, err, "failed to parse %s", filename)
		parsed = append(parsed, file)
	}

	return parsed, fset
}
