package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestCheckGiantBranchingChains(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		expectedCount    int
		expectedTypes    []string
		expectedSeverity metrics.SeverityLevel
	}{
		{
			name: "giant switch statement",
			code: `package main
func process(x int) string {
	switch x {
	case 1: return "one"
	case 2: return "two"
	case 3: return "three"
	case 4: return "four"
	case 5: return "five"
	case 6: return "six"
	case 7: return "seven"
	case 8: return "eight"
	case 9: return "nine"
	case 10: return "ten"
	case 11: return "eleven"
	case 12: return "twelve"
	default: return "other"
	}
}`,
			expectedCount:    1,
			expectedTypes:    []string{"giant_switch"},
			expectedSeverity: metrics.SeverityLevelWarning,
		},
		{
			name: "acceptable switch statement",
			code: `package main
func process(x int) string {
	switch x {
	case 1: return "one"
	case 2: return "two"
	case 3: return "three"
	case 4: return "four"
	case 5: return "five"
	default: return "other"
	}
}`,
			expectedCount: 0,
		},
		{
			name: "giant if-else chain",
			code: `package main
func categorize(x int) string {
	if x == 1 {
		return "one"
	} else if x == 2 {
		return "two"
	} else if x == 3 {
		return "three"
	} else if x == 4 {
		return "four"
	} else if x == 5 {
		return "five"
	} else if x == 6 {
		return "six"
	} else if x == 7 {
		return "seven"
	} else if x == 8 {
		return "eight"
	} else if x == 9 {
		return "nine"
	} else if x == 10 {
		return "ten"
	} else if x == 11 {
		return "eleven"
	} else {
		return "other"
	}
}`,
			expectedCount:    1,
			expectedTypes:    []string{"giant_if_else_chain"},
			expectedSeverity: metrics.SeverityLevelWarning,
		},
		{
			name: "acceptable if-else chain",
			code: `package main
func categorize(x int) string {
	if x == 1 {
		return "one"
	} else if x == 2 {
		return "two"
	} else if x == 3 {
		return "three"
	} else {
		return "other"
	}
}`,
			expectedCount: 0,
		},
		{
			name: "giant type switch",
			code: `package main
func handleValue(v interface{}) string {
	switch v.(type) {
	case int: return "int"
	case string: return "string"
	case bool: return "bool"
	case float64: return "float64"
	case []int: return "[]int"
	case []string: return "[]string"
	case map[string]int: return "map[string]int"
	case map[int]string: return "map[int]string"
	case struct{}: return "struct"
	case *int: return "*int"
	case *string: return "*string"
	case chan int: return "chan int"
	default: return "unknown"
	}
}`,
			expectedCount:    1,
			expectedTypes:    []string{"giant_type_switch"},
			expectedSeverity: metrics.SeverityLevelWarning,
		},
		{
			name: "multiple giant branching structures",
			code: `package main
func process(x int, v interface{}) string {
	switch x {
	case 1: return "one"
	case 2: return "two"
	case 3: return "three"
	case 4: return "four"
	case 5: return "five"
	case 6: return "six"
	case 7: return "seven"
	case 8: return "eight"
	case 9: return "nine"
	case 10: return "ten"
	case 11: return "eleven"
	default: return "other"
	}

	if x == 1 {
		return "one"
	} else if x == 2 {
		return "two"
	} else if x == 3 {
		return "three"
	} else if x == 4 {
		return "four"
	} else if x == 5 {
		return "five"
	} else if x == 6 {
		return "six"
	} else if x == 7 {
		return "seven"
	} else if x == 8 {
		return "eight"
	} else if x == 9 {
		return "nine"
	} else if x == 10 {
		return "ten"
	} else if x == 11 {
		return "eleven"
	} else {
		return "other"
	}
}`,
			expectedCount: 2,
			expectedTypes: []string{"giant_switch", "giant_if_else_chain"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			analyzer := NewAntipatternAnalyzer(fset)
			patterns := analyzer.Analyze(file)

			// Filter to only branching-related patterns
			branchingPatterns := []string{"giant_switch", "giant_type_switch", "giant_if_else_chain"}
			var filtered []string
			for _, p := range patterns {
				for _, bp := range branchingPatterns {
					if p.Type == bp {
						filtered = append(filtered, p.Type)
						if tt.expectedSeverity != "" {
							assert.Equal(t, tt.expectedSeverity, p.Severity, "Severity mismatch for pattern %s", p.Type)
						}
						assert.NotEmpty(t, p.Suggestion, "Suggestion should not be empty for pattern %s", p.Type)
					}
				}
			}

			assert.Equal(t, tt.expectedCount, len(filtered), "Expected %d branching patterns, got %d", tt.expectedCount, len(filtered))

			if tt.expectedTypes != nil && len(tt.expectedTypes) > 0 {
				for _, expectedType := range tt.expectedTypes {
					assert.Contains(t, filtered, expectedType, "Expected pattern type %s not found", expectedType)
				}
			}
		})
	}
}

func TestCountSwitchBranches(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedCount int
	}{
		{
			name: "switch with 5 cases",
			code: `package main
func f(x int) {
	switch x {
	case 1:
	case 2:
	case 3:
	case 4:
	default:
	}
}`,
			expectedCount: 5,
		},
		{
			name: "switch with 12 cases",
			code: `package main
func f(x int) {
	switch x {
	case 1, 2, 3:
	case 4:
	case 5:
	case 6:
	case 7:
	case 8:
	case 9:
	case 10:
	case 11:
	case 12:
	case 13:
	default:
	}
}`,
			expectedCount: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			analyzer := NewAntipatternAnalyzer(fset)

			// Find the switch statement
			var switchStmt *ast.SwitchStmt
			ast.Inspect(file, func(n ast.Node) bool {
				if s, ok := n.(*ast.SwitchStmt); ok {
					switchStmt = s
					return false
				}
				return true
			})

			assert.NotNil(t, switchStmt, "Switch statement not found")
			count := analyzer.countSwitchBranches(switchStmt)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestCountIfElseChainLength(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		expectedCount int
	}{
		{
			name: "simple if statement",
			code: `package main
func f(x int) {
	if x == 1 {
	}
}`,
			expectedCount: 1,
		},
		{
			name: "if-else",
			code: `package main
func f(x int) {
	if x == 1 {
	} else {
	}
}`,
			expectedCount: 2,
		},
		{
			name: "if-else-if chain of 5",
			code: `package main
func f(x int) {
	if x == 1 {
	} else if x == 2 {
	} else if x == 3 {
	} else if x == 4 {
	} else {
	}
}`,
			expectedCount: 5,
		},
		{
			name: "if-else-if chain of 12",
			code: `package main
func f(x int) {
	if x == 1 {
	} else if x == 2 {
	} else if x == 3 {
	} else if x == 4 {
	} else if x == 5 {
	} else if x == 6 {
	} else if x == 7 {
	} else if x == 8 {
	} else if x == 9 {
	} else if x == 10 {
	} else if x == 11 {
	} else {
	}
}`,
			expectedCount: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			analyzer := NewAntipatternAnalyzer(fset)

			// Find the if statement
			var ifStmt *ast.IfStmt
			ast.Inspect(file, func(n ast.Node) bool {
				if i, ok := n.(*ast.IfStmt); ok && ifStmt == nil {
					ifStmt = i
					return false
				}
				return true
			})

			assert.NotNil(t, ifStmt, "If statement not found")
			count := analyzer.countIfElseChainLength(ifStmt)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestBranchingPatternSuggestions(t *testing.T) {
	code := `package main
func process(x int) string {
	switch x {
	case 1: return "one"
	case 2: return "two"
	case 3: return "three"
	case 4: return "four"
	case 5: return "five"
	case 6: return "six"
	case 7: return "seven"
	case 8: return "eight"
	case 9: return "nine"
	case 10: return "ten"
	case 11: return "eleven"
	default: return "other"
	}
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	assert.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Find the giant_switch pattern
	var giantSwitch *metrics.PerformanceAntipattern
	for i := range patterns {
		if patterns[i].Type == "giant_switch" {
			giantSwitch = &patterns[i]
			break
		}
	}

	assert.NotNil(t, giantSwitch, "giant_switch pattern not found")
	assert.Contains(t, giantSwitch.Suggestion, "dispatch map", "Suggestion should mention dispatch map")
	assert.Contains(t, giantSwitch.Suggestion, "strategy pattern", "Suggestion should mention strategy pattern")
	assert.Equal(t, metrics.SeverityLevelWarning, giantSwitch.Severity)
}
