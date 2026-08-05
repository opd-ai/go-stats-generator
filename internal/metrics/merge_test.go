package metrics

import "testing"

func TestMergeGenericsData_NilMaps(t *testing.T) {
	// Test that merging into a zero-value GenericMetrics doesn't panic
	var merged GenericMetrics
	gen := GenericMetrics{
		TypeParameters: GenericTypeParameters{
			Count: 5,
			Constraints: map[string]int{
				"Comparable": 3,
				"Ordered":    2,
			},
			Complexity: []GenericComplexity{
				{Name: "Func1", ParameterCount: 2, ConstraintCount: 1, ComplexityScore: 1.5},
			},
		},
		ConstraintUsage: map[string]int{
			"Comparable": 10,
		},
		Instantiations: GenericInstantiations{
			Functions: []GenericInstantiation{{GenericName: "Func1", TypeArgs: []string{"int"}}},
			Types:     []GenericInstantiation{{GenericName: "Type1", TypeArgs: []string{"string"}}},
			Methods:   []GenericInstantiation{{GenericName: "Method1", TypeArgs: []string{"bool"}}},
		},
	}

	// This should not panic
	MergeGenericsData(&merged, gen)

	// Verify the merge worked
	if merged.TypeParameters.Count != 5 {
		t.Errorf("expected Count=5, got %d", merged.TypeParameters.Count)
	}
	if merged.TypeParameters.Constraints["Comparable"] != 3 {
		t.Errorf("expected Constraints[Comparable]=3, got %d", merged.TypeParameters.Constraints["Comparable"])
	}
	if merged.TypeParameters.Constraints["Ordered"] != 2 {
		t.Errorf("expected Constraints[Ordered]=2, got %d", merged.TypeParameters.Constraints["Ordered"])
	}
	if merged.ConstraintUsage["Comparable"] != 10 {
		t.Errorf("expected ConstraintUsage[Comparable]=10, got %d", merged.ConstraintUsage["Comparable"])
	}
	if len(merged.TypeParameters.Complexity) != 1 {
		t.Errorf("expected 1 complexity entry, got %d", len(merged.TypeParameters.Complexity))
	}
	if len(merged.Instantiations.Functions) != 1 {
		t.Errorf("expected 1 function instantiation, got %d", len(merged.Instantiations.Functions))
	}
	if len(merged.Instantiations.Types) != 1 {
		t.Errorf("expected 1 type instantiation, got %d", len(merged.Instantiations.Types))
	}
	if len(merged.Instantiations.Methods) != 1 {
		t.Errorf("expected 1 method instantiation, got %d", len(merged.Instantiations.Methods))
	}
}

func TestMergeGenericsData_MultipleMerges(t *testing.T) {
	// Test multiple merges into the same target
	var merged GenericMetrics

	gen1 := GenericMetrics{
		TypeParameters: GenericTypeParameters{
			Count: 3,
			Constraints: map[string]int{
				"Comparable": 2,
			},
		},
		ConstraintUsage: map[string]int{
			"Comparable": 5,
		},
	}

	gen2 := GenericMetrics{
		TypeParameters: GenericTypeParameters{
			Count: 2,
			Constraints: map[string]int{
				"Ordered": 1,
			},
		},
		ConstraintUsage: map[string]int{
			"Ordered": 3,
		},
	}

	MergeGenericsData(&merged, gen1)
	MergeGenericsData(&merged, gen2)

	if merged.TypeParameters.Count != 5 {
		t.Errorf("expected Count=5, got %d", merged.TypeParameters.Count)
	}
	if merged.TypeParameters.Constraints["Comparable"] != 2 {
		t.Errorf("expected Constraints[Comparable]=2, got %d", merged.TypeParameters.Constraints["Comparable"])
	}
	if merged.TypeParameters.Constraints["Ordered"] != 1 {
		t.Errorf("expected Constraints[Ordered]=1, got %d", merged.TypeParameters.Constraints["Ordered"])
	}
	if merged.ConstraintUsage["Comparable"] != 5 {
		t.Errorf("expected ConstraintUsage[Comparable]=5, got %d", merged.ConstraintUsage["Comparable"])
	}
	if merged.ConstraintUsage["Ordered"] != 3 {
		t.Errorf("expected ConstraintUsage[Ordered]=3, got %d", merged.ConstraintUsage["Ordered"])
	}
}
