package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestAnalyzeQuality(t *testing.T) {
	tests := []struct {
		name                string
		code                string
		wantInlineMin       int
		wantBlockMin        int
		wantCodeExamplesMin int
		wantAvgLengthMin    float64
		wantQualityMin      float64
	}{
		{
			name: "detects inline and block comments",
			code: `package test
// This is an inline comment
// Another inline comment
func example() {}

/*
This is a block comment
spanning multiple lines
with some content
*/
type Example struct{}
`,
			wantInlineMin: 2,
			wantBlockMin:  1,
		},
		{
			name: "detects code examples",
			code: `package test
// Example:
//   result := example()
//   fmt.Println(result)
func example() int { return 42 }
`,
			wantCodeExamplesMin: 1,
		},
		{
			name: "calculates average comment length",
			code: `package test
// Short comment
// This is a longer comment with more words and content to analyze
func example() {}
`,
			wantAvgLengthMin: 10,
		},
		{
			name: "calculates quality score",
			code: `package test
// Example is a well-documented function that demonstrates quality
// Example:
//   result := Example()
// Usage:
//   fmt.Println(result)
func Example() int { return 42 }
`,
			wantQualityMin: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			analyzer := NewDocumentationAnalyzer(fset, nil)
			metrics := analyzer.Analyze([]*ast.File{f}, nil)

			if metrics.Quality.InlineComments < tt.wantInlineMin {
				t.Errorf("InlineComments = %d, want >= %d",
					metrics.Quality.InlineComments, tt.wantInlineMin)
			}

			if metrics.Quality.BlockComments < tt.wantBlockMin {
				t.Errorf("BlockComments = %d, want >= %d",
					metrics.Quality.BlockComments, tt.wantBlockMin)
			}

			if metrics.Quality.CodeExamples < tt.wantCodeExamplesMin {
				t.Errorf("CodeExamples = %d, want >= %d",
					metrics.Quality.CodeExamples, tt.wantCodeExamplesMin)
			}

			if metrics.Quality.AverageLength < tt.wantAvgLengthMin {
				t.Errorf("AverageLength = %.2f, want >= %.2f",
					metrics.Quality.AverageLength, tt.wantAvgLengthMin)
			}

			if metrics.Quality.QualityScore < tt.wantQualityMin {
				t.Errorf("QualityScore = %.2f, want >= %.2f",
					metrics.Quality.QualityScore, tt.wantQualityMin)
			}
		})
	}
}

func TestContainsCodeExample(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "detects Example keyword",
			text: "Example:\n  result := function()",
			want: true,
		},
		{
			name: "detects Usage keyword",
			text: "Usage:\n  call this function",
			want: true,
		},
		{
			name: "detects func declaration",
			text: "func example() int",
			want: true,
		},
		{
			name: "detects var declaration",
			text: "var x = 10",
			want: true,
		},
		{
			name: "detects type declaration",
			text: "type Example struct{}",
			want: true,
		},
		{
			name: "no code example",
			text: "This is just a regular comment",
			want: false,
		},
	}

	analyzer := NewDocumentationAnalyzer(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyzer.containsCodeExample(tt.text)
			if got != tt.want {
				t.Errorf("containsCodeExample() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateQualityScore(t *testing.T) {
	tests := []struct {
		name         string
		coverage     float64
		avgLength    float64
		codeExamples int
		wantScoreMin float64
		wantScoreMax float64
	}{
		{
			name:         "high coverage and quality",
			coverage:     80.0,
			avgLength:    60.0,
			codeExamples: 10,
			wantScoreMin: 50.0,
			wantScoreMax: 100.0,
		},
		{
			name:         "medium coverage",
			coverage:     50.0,
			avgLength:    40.0,
			codeExamples: 5,
			wantScoreMin: 20.0,
			wantScoreMax: 60.0,
		},
		{
			name:         "low quality",
			coverage:     20.0,
			avgLength:    10.0,
			codeExamples: 0,
			wantScoreMin: 0.0,
			wantScoreMax: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewDocumentationAnalyzer(fset, nil)
			f, _ := parser.ParseFile(fset, "test.go", "package test", parser.ParseComments)
			metrics := analyzer.Analyze([]*ast.File{f}, nil)

			metrics.Coverage.Overall = tt.coverage
			metrics.Quality.AverageLength = tt.avgLength
			metrics.Quality.CodeExamples = tt.codeExamples

			score := analyzer.calculateQualityScore(metrics)

			if score < tt.wantScoreMin || score > tt.wantScoreMax {
				t.Errorf("calculateQualityScore() = %.2f, want between %.2f and %.2f",
					score, tt.wantScoreMin, tt.wantScoreMax)
			}
		})
	}
}
