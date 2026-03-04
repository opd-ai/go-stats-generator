package metrics

// MergeGenericsData merges a single generic metric into the target
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
