package metrics

// DefaultMinPatternConfidence is the minimum confidence score for a pattern to be included in reports.
// Patterns below this threshold are considered likely false positives and are filtered out.
const DefaultMinPatternConfidence = 0.5

// FilterLowConfidencePatterns removes pattern instances with confidence scores below the given
// threshold from all pattern collections in the report. This reduces false positives while
// preserving all true positives (which have higher confidence scores). A threshold of 0
// disables filtering, preserving all results.
func (r *Report) FilterLowConfidencePatterns(minConfidence float64) {
	if minConfidence <= 0 {
		return
	}
	r.filterDesignPatterns(minConfidence)
	r.filterConcurrencyPatterns(minConfidence)
}

// filterDesignPatterns filters design pattern instances below the confidence threshold
func (r *Report) filterDesignPatterns(minConfidence float64) {
	dp := &r.Patterns.DesignPatterns
	dp.Singleton = filterPatternInstances(dp.Singleton, minConfidence)
	dp.Factory = filterPatternInstances(dp.Factory, minConfidence)
	dp.Builder = filterPatternInstances(dp.Builder, minConfidence)
	dp.Observer = filterPatternInstances(dp.Observer, minConfidence)
	dp.Strategy = filterPatternInstances(dp.Strategy, minConfidence)
}

// filterConcurrencyPatterns filters concurrency pattern instances below the confidence threshold
func (r *Report) filterConcurrencyPatterns(minConfidence float64) {
	cp := &r.Patterns.ConcurrencyPatterns
	cp.WorkerPools = filterPatternInstances(cp.WorkerPools, minConfidence)
	cp.Pipelines = filterPatternInstances(cp.Pipelines, minConfidence)
	cp.FanOut = filterPatternInstances(cp.FanOut, minConfidence)
	cp.FanIn = filterPatternInstances(cp.FanIn, minConfidence)
	cp.Semaphores = filterPatternInstances(cp.Semaphores, minConfidence)
}

// filterPatternInstances returns only the pattern instances at or above the minimum confidence score
func filterPatternInstances(patterns []PatternInstance, minConfidence float64) []PatternInstance {
	if len(patterns) == 0 {
		return patterns
	}
	filtered := make([]PatternInstance, 0, len(patterns))
	for _, p := range patterns {
		if p.ConfidenceScore >= minConfidence {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
