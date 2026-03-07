package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDiffOptions(t *testing.T) {
	opts := DefaultDiffOptions()
	
	assert.Equal(t, 5.0, opts.ThresholdPercent)
	assert.False(t, opts.ShowOnlyChanges)
}

func TestCompareSnapshots_Success(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseline := newTestSnapshot("baseline-1", 
		[]FunctionMetrics{newTestFunctionMetrics("TestFunc", "pkg", 5, 20)},
		[]StructMetrics{newTestStructMetrics("TestStruct", "pkg", 3)},
		[]PackageMetrics{newTestPackageMetrics("pkg", 2.0, 0.8)},
	)
	
	current := newTestSnapshot("current-1",
		[]FunctionMetrics{newTestFunctionMetrics("TestFunc", "pkg", 8, 30)},
		[]StructMetrics{newTestStructMetrics("TestStruct", "pkg", 5)},
		[]PackageMetrics{newTestPackageMetrics("pkg", 3.0, 0.7)},
	)
	
	diff, err := CompareSnapshots(baseline, current, config)
	require.NoError(t, err)
	require.NotNil(t, diff)
	
	assert.Equal(t, "baseline-1", diff.Baseline.ID)
	assert.Equal(t, "current-1", diff.Current.ID)
}

func TestCompareSnapshots_InvalidIDs(t *testing.T) {
	config := DefaultThresholdConfig()
	
	tests := []struct {
		name       string
		baselineID string
		currentID  string
	}{
		{"empty baseline", "", "valid"},
		{"empty current", "valid", ""},
		{"both empty", "", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseline := newTestSnapshot(tt.baselineID, nil, nil, nil)
			current := newTestSnapshot(tt.currentID, nil, nil, nil)
			
			_, err := CompareSnapshots(baseline, current, config)
			assert.Error(t, err)
		})
	}
}

func TestCompareFunctionMetrics(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseline := []FunctionMetrics{
		newTestFunctionMetrics("Func1", "pkg", 5, 20),
		newTestFunctionMetrics("Func2", "pkg", 3, 10),
	}
	
	current := []FunctionMetrics{
		newTestFunctionMetrics("Func1", "pkg", 8, 25),
		newTestFunctionMetrics("Func3", "pkg", 2, 5),
	}
	
	changes := compareFunctionMetrics(baseline, current, config)
	assert.NotEmpty(t, changes)
}

func TestBuildFunctionMaps(t *testing.T) {
	baseline := []FunctionMetrics{
		newTestFunctionMetrics("Func1", "pkg", 5, 20),
		newTestFunctionMetrics("Func2", "pkg", 3, 10),
	}
	
	current := []FunctionMetrics{
		newTestFunctionMetrics("Func1", "pkg", 8, 25),
		newTestFunctionMetrics("Func3", "pkg", 2, 5),
	}
	
	baseMap, currMap := buildFunctionMaps(baseline, current)
	
	assert.Len(t, baseMap, 2)
	assert.Len(t, currMap, 2)
	assert.Contains(t, baseMap, "pkg.Func1")
	assert.Contains(t, baseMap, "pkg.Func2")
	assert.Contains(t, currMap, "pkg.Func1")
	assert.Contains(t, currMap, "pkg.Func3")
}

func TestCollectAllFunctionKeys(t *testing.T) {
	baseMap := map[string]FunctionMetrics{
		"pkg.Func1": newTestFunctionMetrics("Func1", "pkg", 5, 20),
		"pkg.Func2": newTestFunctionMetrics("Func2", "pkg", 3, 10),
	}
	
	currMap := map[string]FunctionMetrics{
		"pkg.Func1": newTestFunctionMetrics("Func1", "pkg", 8, 25),
		"pkg.Func3": newTestFunctionMetrics("Func3", "pkg", 2, 5),
	}
	
	allKeys := collectAllFunctionKeys(baseMap, currMap)
	
	assert.Len(t, allKeys, 3)
	assert.True(t, allKeys["pkg.Func1"])
	assert.True(t, allKeys["pkg.Func2"])
	assert.True(t, allKeys["pkg.Func3"])
}

func TestCompareFunctionsByKey(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseMap := map[string]FunctionMetrics{
		"pkg.Func1": newTestFunctionMetrics("Func1", "pkg", 5, 20),
		"pkg.Func2": newTestFunctionMetrics("Func2", "pkg", 3, 10),
	}
	
	currMap := map[string]FunctionMetrics{
		"pkg.Func1": newTestFunctionMetrics("Func1", "pkg", 8, 25),
		"pkg.Func3": newTestFunctionMetrics("Func3", "pkg", 2, 5),
	}
	
	allKeys := map[string]bool{
		"pkg.Func1": true,
		"pkg.Func2": true,
		"pkg.Func3": true,
	}
	
	changes := compareFunctionsByKey(baseMap, currMap, allKeys, config)
	assert.NotEmpty(t, changes)
}

func TestBuildFunctionRemovedChange(t *testing.T) {
	baseFunc := newTestFunctionMetrics("RemovedFunc", "pkg", 5, 20)
	
	change := buildFunctionRemovedChange(baseFunc)
	
	assert.Equal(t, "function", change.Category)
	assert.Equal(t, "pkg.RemovedFunc", change.Path)
	assert.Equal(t, ImpactLevelMedium, change.Impact)
	assert.Equal(t, "Function removed", change.Description)
}

func TestBuildFunctionAddedChange(t *testing.T) {
	currFunc := newTestFunctionMetrics("NewFunc", "pkg", 3, 10)
	
	change := buildFunctionAddedChange(currFunc)
	
	assert.Equal(t, "function", change.Category)
	assert.Equal(t, "pkg.NewFunc", change.Path)
	assert.Equal(t, ImpactLevelLow, change.Impact)
}

func TestCompareFunctionComplexity(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseline := newTestFunctionMetrics("TestFunc", "pkg", 5, 20)
	current := newTestFunctionMetrics("TestFunc", "pkg", 12, 35)
	
	changes := compareFunctionComplexity(baseline, current, config)
	assert.NotEmpty(t, changes)
}

func TestCreateCyclomaticChange(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseline := newTestFunctionMetrics("TestFunc", "pkg", 5, 20)
	current := newTestFunctionMetrics("TestFunc", "pkg", 12, 20)
	
	change := createCyclomaticChange(baseline, current, config)
	
	assert.NotEmpty(t, change.Description)
	assert.NotZero(t, change.Delta.Absolute)
}

func TestCompareStructMetrics(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseline := []StructMetrics{
		newTestStructMetrics("Struct1", "pkg", 3),
		newTestStructMetrics("Struct2", "pkg", 5),
	}
	
	current := []StructMetrics{
		newTestStructMetrics("Struct1", "pkg", 6),
		newTestStructMetrics("Struct3", "pkg", 2),
	}
	
	changes := compareStructMetrics(baseline, current, config)
	assert.NotEmpty(t, changes)
}

func TestBuildStructMaps(t *testing.T) {
	baseline := []StructMetrics{
		newTestStructMetrics("Struct1", "pkg", 3),
		newTestStructMetrics("Struct2", "pkg", 5),
	}
	
	current := []StructMetrics{
		newTestStructMetrics("Struct1", "pkg", 6),
	}
	
	baseMap, currMap := buildStructMaps(baseline, current)
	
	assert.Len(t, baseMap, 2)
	assert.Len(t, currMap, 1)
}

func TestMergeStructKeys(t *testing.T) {
	baseMap := map[string]StructMetrics{
		"pkg.Struct1": newTestStructMetrics("Struct1", "pkg", 3),
		"pkg.Struct2": newTestStructMetrics("Struct2", "pkg", 5),
	}
	
	currMap := map[string]StructMetrics{
		"pkg.Struct1": newTestStructMetrics("Struct1", "pkg", 6),
		"pkg.Struct3": newTestStructMetrics("Struct3", "pkg", 2),
	}
	
	allKeys := mergeStructKeys(baseMap, currMap)
	
	assert.Len(t, allKeys, 3)
	assert.True(t, allKeys["pkg.Struct1"])
	assert.True(t, allKeys["pkg.Struct2"])
	assert.True(t, allKeys["pkg.Struct3"])
}

func TestCreateStructRemovedChange(t *testing.T) {
	baseStruct := newTestStructMetrics("RemovedStruct", "pkg", 5)
	
	change := createStructRemovedChange(baseStruct)
	
	assert.Equal(t, "struct", change.Category)
	assert.Equal(t, ImpactLevelMedium, change.Impact)
	assert.Equal(t, "Struct removed", change.Description)
}

func TestCreateStructAddedChange(t *testing.T) {
	currStruct := newTestStructMetrics("NewStruct", "pkg", 3)
	
	change := createStructAddedChange(currStruct)
	
	assert.Equal(t, "struct", change.Category)
	assert.Equal(t, ImpactLevelLow, change.Impact)
}

func TestComparePackageMetrics(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseline := []PackageMetrics{
		newTestPackageMetrics("pkg1", 2.0, 0.8),
		newTestPackageMetrics("pkg2", 1.5, 0.9),
	}
	
	current := []PackageMetrics{
		newTestPackageMetrics("pkg1", 3.0, 0.7),
		newTestPackageMetrics("pkg3", 1.0, 0.95),
	}
	
	changes := comparePackageMetrics(baseline, current, config)
	assert.NotEmpty(t, changes)
}

func TestBuildPackageMaps(t *testing.T) {
	baseline := []PackageMetrics{
		newTestPackageMetrics("pkg1", 2.0, 0.8),
		newTestPackageMetrics("pkg2", 1.5, 0.9),
	}
	
	current := []PackageMetrics{
		newTestPackageMetrics("pkg1", 3.0, 0.7),
	}
	
	baseMap, currMap := buildPackageMaps(baseline, current)
	
	assert.Len(t, baseMap, 2)
	assert.Len(t, currMap, 1)
}

func TestMergePackagePaths(t *testing.T) {
	baseMap := map[string]PackageMetrics{
		"pkg1": newTestPackageMetrics("pkg1", 2.0, 0.8),
		"pkg2": newTestPackageMetrics("pkg2", 1.5, 0.9),
	}
	
	currMap := map[string]PackageMetrics{
		"pkg1": newTestPackageMetrics("pkg1", 3.0, 0.7),
		"pkg3": newTestPackageMetrics("pkg3", 1.0, 0.95),
	}
	
	allPaths := mergePackagePaths(baseMap, currMap)
	
	assert.Len(t, allPaths, 3)
	assert.True(t, allPaths["pkg1"])
	assert.True(t, allPaths["pkg2"])
	assert.True(t, allPaths["pkg3"])
}

func TestCompareComplexityMetrics(t *testing.T) {
	config := DefaultThresholdConfig()
	
	baseline := ComplexityMetrics{
		AverageFunction: 5.0,
		AverageStruct:   3.0,
	}
	
	current := ComplexityMetrics{
		AverageFunction: 7.0,
		AverageStruct:   4.0,
	}
	
	changes := compareComplexityMetrics(baseline, current, config)
	assert.NotEmpty(t, changes)
}

func TestCategorizeChanges(t *testing.T) {
	config := DefaultThresholdConfig()
	
	changes := []MetricChange{
		{
			Category:    "complexity",
			Description: "Complexity increased",
			Impact:      ImpactLevelHigh,
			Severity:    SeverityLevelWarning,
			Delta:       Delta{Absolute: 5, Percentage: 50, Direction: ChangeDirectionIncrease, Significant: true},
		},
		{
			Category:    "complexity",
			Description: "Complexity decreased",
			Impact:      ImpactLevelLow,
			Severity:    SeverityLevelInfo,
			Delta:       Delta{Absolute: -3, Percentage: -30, Direction: ChangeDirectionDecrease, Significant: true},
		},
	}
	
	regressions, improvements := categorizeChanges(changes, config)
	
	assert.NotEmpty(t, regressions)
	assert.NotEmpty(t, improvements)
}

func TestBuildRegression(t *testing.T) {
	change := MetricChange{
		Category:    "function",
		Description: "Complexity increased",
		Path:        "pkg.Func",
		Impact:      ImpactLevelHigh,
		Delta:       Delta{Absolute: 5},
	}
	
	regression := buildRegression(change)
	
	assert.Equal(t, "Complexity increased", regression.Description)
	assert.Equal(t, "pkg.Func", regression.Location)
}

func TestBuildImprovement(t *testing.T) {
	change := MetricChange{
		Category:    "function",
		Description: "Complexity decreased",
		Path:        "pkg.Func",
		Impact:      ImpactLevelLow,
		Delta:       Delta{Absolute: -3},
	}
	
	improvement := buildImprovement(change)
	
	assert.Equal(t, "Complexity decreased", improvement.Description)
	assert.Equal(t, -3.0, improvement.Delta.Absolute)
}

func TestGenerateDiffSummary(t *testing.T) {
	changes := []MetricChange{
		{Impact: ImpactLevelHigh, Delta: Delta{Absolute: 5}},
		{Impact: ImpactLevelLow, Delta: Delta{Absolute: -2}},
	}
	
	regressions := []Regression{
		{Severity: SeverityLevelCritical},
	}
	
	improvements := []Improvement{
		{Delta: Delta{Absolute: -2.0}},
	}
	
	summary := generateDiffSummary(changes, regressions, improvements)
	
	assert.Equal(t, 2, summary.TotalChanges)
	assert.Equal(t, 1, summary.RegressionCount)
	assert.Equal(t, 1, summary.ImprovementCount)
}

func TestCountSignificantChanges(t *testing.T) {
	changes := []MetricChange{
		{Impact: ImpactLevelHigh, Delta: Delta{Significant: true}},
		{Impact: ImpactLevelMedium, Delta: Delta{Significant: true}},
		{Impact: ImpactLevelLow, Delta: Delta{Significant: false}},
	}
	
	count := countSignificantChanges(changes)
	assert.Equal(t, 2, count)
}

func TestCountCriticalIssues(t *testing.T) {
	regressions := []Regression{
		{Severity: SeverityLevelCritical},
		{Severity: SeverityLevelError},
		{Severity: SeverityLevelWarning},
	}
	
	count := countCriticalIssues(regressions)
	assert.Equal(t, 1, count)
}

func TestDetermineOverallTrend(t *testing.T) {
	tests := []struct {
		name         string
		regressions  []Regression
		improvements []Improvement
		expected     TrendDirection
	}{
		{
			name:         "more regressions",
			regressions:  []Regression{{}, {}, {}},
			improvements: []Improvement{{}},
			expected:     TrendDegrading,
		},
		{
			name:         "more improvements",
			regressions:  []Regression{{}},
			improvements: []Improvement{{}, {}, {}},
			expected:     TrendImproving,
		},
		{
			name:         "equal",
			regressions:  []Regression{{}, {}},
			improvements: []Improvement{{}, {}},
			expected:     TrendStable,
		},
		{
			name:         "no changes",
			regressions:  []Regression{},
			improvements: []Improvement{},
			expected:     TrendStable,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := determineOverallTrend(tt.regressions, tt.improvements)
			assert.Equal(t, tt.expected, trend)
		})
	}
}

func TestCalculateQualityScore(t *testing.T) {
	improvements := []Improvement{
		{Delta: Delta{Absolute: -2.0}},
		{Delta: Delta{Absolute: -3.0}},
		{Delta: Delta{Absolute: -1.0}},
	}
	
	score := calculateQualityScore(5, improvements)
	
	assert.GreaterOrEqual(t, score, 0.0)
	assert.LessOrEqual(t, score, 100.0)
}

func TestCalculateDelta(t *testing.T) {
	tests := []struct {
		name      string
		oldValue  float64
		newValue  float64
		threshold float64
		wantAbs   float64
		wantPerc  float64
	}{
		{
			name:      "increase",
			oldValue:  10.0,
			newValue:  15.0,
			threshold: 5.0,
			wantAbs:   5.0,
			wantPerc:  50.0,
		},
		{
			name:      "decrease",
			oldValue:  20.0,
			newValue:  15.0,
			threshold: 5.0,
			wantAbs:   -5.0,
			wantPerc:  25.0,
		},
		{
			name:      "no change",
			oldValue:  10.0,
			newValue:  10.0,
			threshold: 5.0,
			wantAbs:   0.0,
			wantPerc:  0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := calculateDelta(tt.oldValue, tt.newValue, tt.threshold)
			
			assert.Equal(t, tt.wantAbs, delta.Absolute)
			assert.Equal(t, tt.wantPerc, delta.Percentage)
		})
	}
}

func TestCalculatePercentageChange(t *testing.T) {
	tests := []struct {
		name     string
		oldValue float64
		newValue float64
		expected float64
	}{
		{
			name:     "50% increase",
			oldValue: 10.0,
			newValue: 15.0,
			expected: 50.0,
		},
		{
			name:     "25% decrease",
			oldValue: 20.0,
			newValue: 15.0,
			expected: 25.0,
		},
		{
			name:     "zero old value",
			oldValue: 0.0,
			newValue: 10.0,
			expected: 100.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pct := calculatePercentageChange(tt.oldValue, tt.newValue)
			assert.Equal(t, tt.expected, pct)
		})
	}
}

func TestDetermineChangeDirection(t *testing.T) {
	tests := []struct {
		name     string
		absolute float64
		expected ChangeDirection
	}{
		{
			name:     "positive",
			absolute: 5.0,
			expected: ChangeDirectionIncrease,
		},
		{
			name:     "negative",
			absolute: -5.0,
			expected: ChangeDirectionDecrease,
		},
		{
			name:     "zero",
			absolute: 0.0,
			expected: ChangeDirectionNeutral,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := determineChangeDirection(tt.absolute)
			assert.Equal(t, tt.expected, dir)
		})
	}
}
