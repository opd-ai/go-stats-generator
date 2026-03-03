package metrics

import (
	"testing"
)

func TestDetectBurdenRegressions(t *testing.T) {
	config := DefaultThresholdConfig()
	config.BurdenMetrics.FileMBIThreshold = 10.0
	config.BurdenMetrics.PackageMBIThreshold = 5.0
	config.BurdenMetrics.MaxDuplicationRatio = 0.10
	config.BurdenMetrics.MaxNamingViolations = 10

	tests := []struct {
		name              string
		baseline          Report
		current           Report
		expectedCount     int
		expectedTypes     []RegressionType
		expectedSeverity  []SeverityLevel
		expectedLocations []string
	}{
		{
			name: "file MBI increase above threshold",
			baseline: Report{
				Scores: ScoringMetrics{
					FileScores: []FileScore{
						{File: "file1.go", Score: 20.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			current: Report{
				Scores: ScoringMetrics{
					FileScores: []FileScore{
						{File: "file1.go", Score: 35.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			expectedCount:     1,
			expectedTypes:     []RegressionType{BurdenRegression},
			expectedSeverity:  []SeverityLevel{SeverityLevelWarning},
			expectedLocations: []string{"file1.go"},
		},
		{
			name: "package MBI increase above threshold",
			baseline: Report{
				Scores: ScoringMetrics{
					PackageScores: []PackageScore{
						{Package: "pkg1", Score: 10.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			current: Report{
				Scores: ScoringMetrics{
					PackageScores: []PackageScore{
						{Package: "pkg1", Score: 18.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			expectedCount:     1,
			expectedTypes:     []RegressionType{BurdenRegression},
			expectedSeverity:  []SeverityLevel{SeverityLevelWarning},
			expectedLocations: []string{"pkg1"},
		},
		{
			name: "duplication ratio exceeds threshold",
			baseline: Report{
				Scores:      ScoringMetrics{},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			current: Report{
				Scores:      ScoringMetrics{},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.15},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			expectedCount:     1,
			expectedTypes:     []RegressionType{DuplicationRegression},
			expectedSeverity:  []SeverityLevel{SeverityLevelError},
			expectedLocations: []string{"global"},
		},
		{
			name: "naming violations exceed threshold",
			baseline: Report{
				Scores:      ScoringMetrics{},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    2,
					IdentifierViolations:  3,
					PackageNameViolations: 1,
				},
			},
			current: Report{
				Scores:      ScoringMetrics{},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    5,
					IdentifierViolations:  8,
					PackageNameViolations: 3,
				},
			},
			expectedCount:     1,
			expectedTypes:     []RegressionType{NamingRegression},
			expectedSeverity:  []SeverityLevel{SeverityLevelWarning},
			expectedLocations: []string{"global"},
		},
		{
			name: "no regressions",
			baseline: Report{
				Scores: ScoringMetrics{
					FileScores: []FileScore{
						{File: "file1.go", Score: 20.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			current: Report{
				Scores: ScoringMetrics{
					FileScores: []FileScore{
						{File: "file1.go", Score: 22.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming: NamingMetrics{
					FileNameViolations:    1,
					IdentifierViolations:  2,
					PackageNameViolations: 1,
				},
			},
			expectedCount: 0,
		},
		{
			name: "critical file MBI increase",
			baseline: Report{
				Scores: ScoringMetrics{
					FileScores: []FileScore{
						{File: "file1.go", Score: 20.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming:      NamingMetrics{},
			},
			current: Report{
				Scores: ScoringMetrics{
					FileScores: []FileScore{
						{File: "file1.go", Score: 55.0},
					},
				},
				Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
				Naming:      NamingMetrics{},
			},
			expectedCount:     1,
			expectedTypes:     []RegressionType{BurdenRegression},
			expectedSeverity:  []SeverityLevel{SeverityLevelCritical},
			expectedLocations: []string{"file1.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regressions := DetectBurdenRegressions(tt.baseline, tt.current, config)

			if len(regressions) != tt.expectedCount {
				t.Errorf("expected %d regressions, got %d", tt.expectedCount, len(regressions))
			}

			if tt.expectedCount > 0 {
				for i, reg := range regressions {
					if i < len(tt.expectedTypes) && reg.Type != tt.expectedTypes[i] {
						t.Errorf("regression %d: expected type %v, got %v", i, tt.expectedTypes[i], reg.Type)
					}
					if i < len(tt.expectedSeverity) && reg.Severity != tt.expectedSeverity[i] {
						t.Errorf("regression %d: expected severity %v, got %v", i, tt.expectedSeverity[i], reg.Severity)
					}
					if i < len(tt.expectedLocations) && reg.Location != tt.expectedLocations[i] {
						t.Errorf("regression %d: expected location %v, got %v", i, tt.expectedLocations[i], reg.Location)
					}
				}
			}
		})
	}
}

func TestDetectBurdenRegressionsMultiple(t *testing.T) {
	config := DefaultThresholdConfig()
	config.BurdenMetrics.FileMBIThreshold = 10.0
	config.BurdenMetrics.PackageMBIThreshold = 5.0
	config.BurdenMetrics.MaxDuplicationRatio = 0.10
	config.BurdenMetrics.MaxNamingViolations = 10

	baseline := Report{
		Scores: ScoringMetrics{
			FileScores: []FileScore{
				{File: "file1.go", Score: 20.0},
				{File: "file2.go", Score: 15.0},
			},
			PackageScores: []PackageScore{
				{Package: "pkg1", Score: 10.0},
			},
		},
		Duplication: DuplicationMetrics{DuplicationRatio: 0.05},
		Naming: NamingMetrics{
			FileNameViolations:    2,
			IdentifierViolations:  3,
			PackageNameViolations: 1,
		},
	}

	current := Report{
		Scores: ScoringMetrics{
			FileScores: []FileScore{
				{File: "file1.go", Score: 35.0},
				{File: "file2.go", Score: 30.0},
			},
			PackageScores: []PackageScore{
				{Package: "pkg1", Score: 20.0},
			},
		},
		Duplication: DuplicationMetrics{DuplicationRatio: 0.15},
		Naming: NamingMetrics{
			FileNameViolations:    5,
			IdentifierViolations:  10,
			PackageNameViolations: 3,
		},
	}

	regressions := DetectBurdenRegressions(baseline, current, config)

	// Should have: 2 file MBI + 1 package MBI + 1 duplication + 1 naming = 5 total
	if len(regressions) < 4 {
		t.Errorf("expected at least 4 regressions, got %d", len(regressions))
	}

	// Verify regressions are sorted by priority
	for i := 1; i < len(regressions); i++ {
		if regressions[i].Priority > regressions[i-1].Priority {
			t.Errorf("regressions not sorted by priority: %d > %d at index %d", regressions[i].Priority, regressions[i-1].Priority, i)
		}
	}
}
