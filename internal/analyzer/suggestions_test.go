package analyzer

import (
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNewSuggestionGenerator(t *testing.T) {
	weights := config.ScoringWeights{
		Duplication:   0.20,
		Naming:        0.10,
		Placement:     0.15,
		Documentation: 0.15,
		Organization:  0.15,
		Burden:        0.25,
	}
	scorer := NewScoringAnalyzer(weights)
	sg := NewSuggestionGenerator(scorer)

	assert.NotNil(t, sg)
	assert.NotNil(t, sg.scorer)
}

func TestGenerateDuplicationSuggestions(t *testing.T) {
	weights := config.ScoringWeights{Duplication: 0.20, Naming: 0.10, Placement: 0.15, Documentation: 0.15, Organization: 0.15, Burden: 0.25}
	scorer := NewScoringAnalyzer(weights)
	sg := NewSuggestionGenerator(scorer)

	report := &metrics.Report{
		Duplication: metrics.DuplicationMetrics{
			Clones: []metrics.ClonePair{
				{
					LineCount: 15,
					Instances: []metrics.CloneInstance{
						{File: "test.go", StartLine: 10, EndLine: 25},
						{File: "test2.go", StartLine: 20, EndLine: 35},
					},
				},
			},
		},
	}

	suggestions := sg.generateDuplicationSuggestions(report)

	assert.Len(t, suggestions, 1)
	assert.Equal(t, ActionDeduplicate, suggestions[0].Action)
	assert.Equal(t, "duplication", suggestions[0].Category)
	assert.Greater(t, suggestions[0].MBIImpact, 0.0)
}

func TestGenerateComplexitySuggestions(t *testing.T) {
	weights := config.ScoringWeights{Duplication: 0.20, Naming: 0.10, Placement: 0.15, Documentation: 0.15, Organization: 0.15, Burden: 0.25}
	scorer := NewScoringAnalyzer(weights)
	sg := NewSuggestionGenerator(scorer)

	report := &metrics.Report{
		Functions: []metrics.FunctionMetrics{
			{
				Name: "ComplexFunction",
				File: "test.go",
				Line: 10,
				Lines: metrics.LineMetrics{
					Code: 50,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 20,
				},
			},
		},
	}

	suggestions := sg.generateComplexitySuggestions(report)

	assert.Len(t, suggestions, 1)
	assert.Equal(t, ActionReduceComplexity, suggestions[0].Action)
	assert.Equal(t, "burden", suggestions[0].Category)
	assert.Equal(t, "ComplexFunction", suggestions[0].Target)
}

func TestGenerateDocumentationSuggestions(t *testing.T) {
	weights := config.ScoringWeights{Duplication: 0.20, Naming: 0.10, Placement: 0.15, Documentation: 0.15, Organization: 0.15, Burden: 0.25}
	scorer := NewScoringAnalyzer(weights)
	sg := NewSuggestionGenerator(scorer)

	report := &metrics.Report{
		Functions: []metrics.FunctionMetrics{
			{
				Name:       "ExportedFunction",
				File:       "test.go",
				Line:       10,
				IsExported: true,
				Documentation: metrics.DocumentationInfo{
					HasComment: false,
				},
			},
		},
	}

	suggestions := sg.generateDocumentationSuggestions(report)

	assert.Len(t, suggestions, 1)
	assert.Equal(t, ActionAddDocumentation, suggestions[0].Action)
	assert.Equal(t, "documentation", suggestions[0].Category)
	assert.Equal(t, EffortLow, suggestions[0].Effort)
}

func TestGenerateSuggestions_Prioritization(t *testing.T) {
	weights := config.ScoringWeights{Duplication: 0.20, Naming: 0.10, Placement: 0.15, Documentation: 0.15, Organization: 0.15, Burden: 0.25}
	scorer := NewScoringAnalyzer(weights)
	sg := NewSuggestionGenerator(scorer)

	report := &metrics.Report{
		Functions: []metrics.FunctionMetrics{
			{
				Name:       "ComplexFunction",
				File:       "test.go",
				Line:       10,
				IsExported: true,
				Lines: metrics.LineMetrics{
					Code: 50,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 20,
				},
				Documentation: metrics.DocumentationInfo{
					HasComment: false,
				},
			},
		},
		Duplication: metrics.DuplicationMetrics{
			Clones: []metrics.ClonePair{
				{
					LineCount: 15,
					Instances: []metrics.CloneInstance{
						{File: "test.go", StartLine: 10, EndLine: 25},
						{File: "test2.go", StartLine: 20, EndLine: 35},
					},
				},
			},
		},
	}

	suggestions := sg.GenerateSuggestions(report)

	// Should have suggestions from multiple categories
	assert.Greater(t, len(suggestions), 0)

	// Verify all suggestions have impact/effort ratio calculated
	for _, s := range suggestions {
		assert.Greater(t, s.ImpactEffort, 0.0)
	}

	// Verify sorted by impact/effort (descending)
	for i := 1; i < len(suggestions); i++ {
		assert.GreaterOrEqual(t, suggestions[i-1].ImpactEffort, suggestions[i].ImpactEffort)
	}
}

func TestCalculateImpactEffortRatio(t *testing.T) {
	weights := config.ScoringWeights{Duplication: 0.20, Naming: 0.10, Placement: 0.15, Documentation: 0.15, Organization: 0.15, Burden: 0.25}
	scorer := NewScoringAnalyzer(weights)
	sg := NewSuggestionGenerator(scorer)

	tests := []struct {
		name     string
		effort   EffortLevel
		impact   float64
		expected float64
	}{
		{"Low effort, high impact", EffortLow, 10.0, 20.0},
		{"Medium effort, high impact", EffortMedium, 10.0, 5.0},
		{"High effort, high impact", EffortHigh, 10.0, 1.666666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RefactoringSuggestion{
				Effort:    tt.effort,
				MBIImpact: tt.impact,
			}
			ratio := sg.calculateImpactEffortRatio(s)
			assert.InDelta(t, tt.expected, ratio, 0.01)
		})
	}
}

func TestIsExported(t *testing.T) {
	assert.True(t, isExported("ExportedFunction"))
	assert.False(t, isExported("unexportedFunction"))
	assert.False(t, isExported(""))
}

func TestClassifyEffort(t *testing.T) {
	assert.Equal(t, EffortLow, classifyEffort(20))
	assert.Equal(t, EffortMedium, classifyEffort(50))
	assert.Equal(t, EffortHigh, classifyEffort(150))
}
