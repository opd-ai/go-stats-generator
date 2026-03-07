package metrics

import "time"

// Test helper functions to create test data

func newTestFunctionMetrics(name, pkg string, cyclomatic, totalLines int) FunctionMetrics {
	return FunctionMetrics{
		Name:    name,
		Package: pkg,
		Complexity: ComplexityScore{
			Cyclomatic: cyclomatic,
		},
		Lines: LineMetrics{
			Total: totalLines,
			Code:  totalLines - 2,
		},
	}
}

func newTestStructMetrics(name, pkg string, totalFields int) StructMetrics {
	return StructMetrics{
		Name:        name,
		Package:     pkg,
		TotalFields: totalFields,
	}
}

func newTestPackageMetrics(path string, coupling, cohesion float64) PackageMetrics {
	return PackageMetrics{
		Path:          path,
		CouplingScore: coupling,
		CohesionScore: cohesion,
	}
}

func newTestSnapshot(id string, funcs []FunctionMetrics, structs []StructMetrics, packages []PackageMetrics) Snapshot {
	return Snapshot{
		ID: id,
		Metadata: SnapshotMetadata{
			Timestamp: time.Now(),
		},
		Report: Report{
			Functions: funcs,
			Structs:   structs,
			Packages:  packages,
			Complexity: ComplexityMetrics{
				AverageFunction: 5.0,
			},
		},
	}
}
