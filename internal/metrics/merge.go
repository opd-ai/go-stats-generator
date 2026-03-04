package metrics

// MergeGenericsData merges a single generic metric into the target aggregation, combining type parameter statistics.
// Accumulates counts, constraint usages, and instantiation lists from gen into merged for cross-file generic analysis.
// Used during multi-file analysis to build comprehensive metrics about Go 1.18+ generics usage patterns.
func MergeGenericsData(merged *GenericMetrics, gen GenericMetrics) {
	merged.TypeParameters.Count += gen.TypeParameters.Count
	merged.TypeParameters.Complexity = append(
		merged.TypeParameters.Complexity,
		gen.TypeParameters.Complexity...)

	for k, v := range gen.TypeParameters.Constraints {
		merged.TypeParameters.Constraints[k] += v
	}
	for k, v := range gen.ConstraintUsage {
		merged.ConstraintUsage[k] += v
	}

	merged.Instantiations.Functions = append(
		merged.Instantiations.Functions,
		gen.Instantiations.Functions...)
	merged.Instantiations.Types = append(
		merged.Instantiations.Types,
		gen.Instantiations.Types...)
	merged.Instantiations.Methods = append(
		merged.Instantiations.Methods,
		gen.Instantiations.Methods...)
}
