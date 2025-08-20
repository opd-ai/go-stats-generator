package analyzer

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLineClassificationComplexComments validates complex comment patterns work correctly
func TestLineClassificationComplexComments(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewFunctionAnalyzer(fset)

	// This used to be a bug: nested block comments were misclassified
	// Now it should work correctly
	inComment := false
	line := "/* outer /* inner comment */ still in comment"
	result := analyzer.classifyLine(line, &inComment)
	
	// This should be classified as "comment" because the entire line is within a comment
	// The inner /* */ should be treated as nested, leaving us still inside the outer comment
	assert.Equal(t, "comment", result, "Nested block comment line should be classified as comment")
	assert.True(t, inComment, "Should still be in block comment after processing nested comments")
	
	// Test that we properly find the real end of complex comments
	inComment = false
	line2 := "/* comment with */ pattern inside */ x := 1"
	result2 := analyzer.classifyLine(line2, &inComment)
	
	// This should properly find the real end of the comment and detect code after
	assert.Equal(t, "mixed", result2, "Line with */ pattern inside comment followed by code should be mixed")
	assert.False(t, inComment, "Should not be in comment after processing complete block comment")
}

// TestLineClassificationRegression ensures existing behavior is preserved
func TestLineClassificationRegression(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewFunctionAnalyzer(fset)

	tests := []struct {
		name     string
		line     string
		inBlock  bool
		expected string
	}{
		{"simple_line_comment", "// simple comment", false, "comment"},
		{"simple_block_comment", "/* simple block comment */", false, "comment"},
		{"mixed_line", "code(); // comment", false, "mixed"},
		{"mixed_block", "code(); /* comment */", false, "mixed"},
		{"code_only", "x := 1", false, "code"},
		{"blank_line", "", false, "blank"},
		{"in_block_comment", "text in block comment", true, "comment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inComment := tt.inBlock
			result := analyzer.classifyLine(tt.line, &inComment)
			assert.Equal(t, tt.expected, result, "Line classification should match expected value")
		})
	}
}
