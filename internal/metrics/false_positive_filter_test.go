package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterLowConfidencePatterns(t *testing.T) {
	tests := []struct {
		name          string
		minConfidence float64
		report        Report
		wantSingleton int
		wantFactory   int
		wantBuilder   int
		wantObserver  int
		wantStrategy  int
		wantWorker    int
		wantPipeline  int
		wantFanOut    int
		wantFanIn     int
		wantSemaphore int
	}{
		{
			name:          "filters low confidence patterns",
			minConfidence: 0.7,
			report: Report{
				Patterns: PatternMetrics{
					DesignPatterns: DesignPatternMetrics{
						Singleton: []PatternInstance{{Name: "Singleton", ConfidenceScore: 0.95}},
						Factory:   []PatternInstance{{Name: "Factory", ConfidenceScore: 0.85}},
						Builder:   []PatternInstance{{Name: "Builder", ConfidenceScore: 0.9}},
						Observer:  []PatternInstance{{Name: "Observer", ConfidenceScore: 0.6}},
						Strategy:  []PatternInstance{{Name: "Strategy", ConfidenceScore: 0.4}},
					},
					ConcurrencyPatterns: ConcurrencyPatternMetrics{
						WorkerPools: []PatternInstance{{Name: "Worker Pool", ConfidenceScore: 0.8}},
						Pipelines:   []PatternInstance{{Name: "Pipeline", ConfidenceScore: 0.5}},
						FanOut:      []PatternInstance{{Name: "Fan-Out", ConfidenceScore: 0.9}},
						FanIn:       []PatternInstance{{Name: "Fan-In", ConfidenceScore: 0.3}},
						Semaphores:  []PatternInstance{{Name: "Semaphore", ConfidenceScore: 0.75}},
					},
				},
			},
			wantSingleton: 1,
			wantFactory:   1,
			wantBuilder:   1,
			wantObserver:  0,
			wantStrategy:  0,
			wantWorker:    1,
			wantPipeline:  0,
			wantFanOut:    1,
			wantFanIn:     0,
			wantSemaphore: 1,
		},
		{
			name:          "zero threshold preserves all",
			minConfidence: 0,
			report: Report{
				Patterns: PatternMetrics{
					DesignPatterns: DesignPatternMetrics{
						Strategy: []PatternInstance{{Name: "Strategy", ConfidenceScore: 0.1}},
					},
					ConcurrencyPatterns: ConcurrencyPatternMetrics{
						Pipelines:  []PatternInstance{{Name: "Pipeline", ConfidenceScore: 0.1}},
						Semaphores: []PatternInstance{{Name: "Semaphore", ConfidenceScore: 0.1}},
					},
				},
			},
			wantStrategy:  1,
			wantPipeline:  1,
			wantSemaphore: 1,
		},
		{
			name:          "empty patterns unaffected",
			minConfidence: 0.5,
			report: Report{
				Patterns: PatternMetrics{
					DesignPatterns:      DesignPatternMetrics{},
					ConcurrencyPatterns: ConcurrencyPatternMetrics{},
				},
			},
		},
		{
			name:          "exact threshold is inclusive",
			minConfidence: 0.5,
			report: Report{
				Patterns: PatternMetrics{
					DesignPatterns: DesignPatternMetrics{
						Factory: []PatternInstance{
							{Name: "Factory A", ConfidenceScore: 0.5},
							{Name: "Factory B", ConfidenceScore: 0.49},
						},
					},
					ConcurrencyPatterns: ConcurrencyPatternMetrics{
						Semaphores: []PatternInstance{
							{Name: "Semaphore A", ConfidenceScore: 0.5},
							{Name: "Semaphore B", ConfidenceScore: 0.49},
						},
					},
				},
			},
			wantFactory:   1,
			wantSemaphore: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.report.FilterLowConfidencePatterns(tt.minConfidence)

			assert.Len(t, tt.report.Patterns.DesignPatterns.Singleton, tt.wantSingleton)
			assert.Len(t, tt.report.Patterns.DesignPatterns.Factory, tt.wantFactory)
			assert.Len(t, tt.report.Patterns.DesignPatterns.Builder, tt.wantBuilder)
			assert.Len(t, tt.report.Patterns.DesignPatterns.Observer, tt.wantObserver)
			assert.Len(t, tt.report.Patterns.DesignPatterns.Strategy, tt.wantStrategy)
			assert.Len(t, tt.report.Patterns.ConcurrencyPatterns.WorkerPools, tt.wantWorker)
			assert.Len(t, tt.report.Patterns.ConcurrencyPatterns.Pipelines, tt.wantPipeline)
			assert.Len(t, tt.report.Patterns.ConcurrencyPatterns.FanOut, tt.wantFanOut)
			assert.Len(t, tt.report.Patterns.ConcurrencyPatterns.FanIn, tt.wantFanIn)
			assert.Len(t, tt.report.Patterns.ConcurrencyPatterns.Semaphores, tt.wantSemaphore)
		})
	}
}

func TestFilterPatternInstances(t *testing.T) {
	patterns := []PatternInstance{
		{Name: "High", ConfidenceScore: 0.95},
		{Name: "Medium", ConfidenceScore: 0.7},
		{Name: "Low", ConfidenceScore: 0.3},
		{Name: "VeryLow", ConfidenceScore: 0.1},
	}

	result := filterPatternInstances(patterns, 0.5)
	assert.Len(t, result, 2)
	assert.Equal(t, "High", result[0].Name)
	assert.Equal(t, "Medium", result[1].Name)
}

func TestFilterPatternInstancesPreservesHighConfidence(t *testing.T) {
	// Verify 100% true positive preservation: all high-confidence patterns must survive
	patterns := []PatternInstance{
		{Name: "Singleton", ConfidenceScore: 0.95},
		{Name: "Factory", ConfidenceScore: 0.85},
		{Name: "Builder", ConfidenceScore: 0.9},
	}

	result := filterPatternInstances(patterns, DefaultMinPatternConfidence)
	assert.Len(t, result, 3, "all high-confidence patterns should be preserved")
	for _, p := range result {
		assert.GreaterOrEqual(t, p.ConfidenceScore, DefaultMinPatternConfidence)
	}
}

func TestDefaultMinPatternConfidenceBehavior(t *testing.T) {
	// Verify that the default threshold correctly filters instances below it
	patterns := []PatternInstance{
		{Name: "Below", ConfidenceScore: DefaultMinPatternConfidence - 0.01},
		{Name: "Equal", ConfidenceScore: DefaultMinPatternConfidence},
		{Name: "Above", ConfidenceScore: DefaultMinPatternConfidence + 0.01},
	}

	result := filterPatternInstances(patterns, DefaultMinPatternConfidence)

	// Instances at or above the default threshold should be preserved,
	// while those below it should be filtered out.
	assert.Len(t, result, 2)
	assert.Equal(t, "Equal", result[0].Name)
	assert.Equal(t, "Above", result[1].Name)
	for _, p := range result {
		assert.GreaterOrEqual(t, p.ConfidenceScore, DefaultMinPatternConfidence)
	}
}
