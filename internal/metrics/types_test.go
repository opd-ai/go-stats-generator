package metrics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuplicationMetricsJSONSerialization(t *testing.T) {
	metrics := DuplicationMetrics{
		ClonePairs:       5,
		DuplicatedLines:  250,
		DuplicationRatio: 0.15,
		LargestCloneSize: 80,
		Clones: []ClonePair{
			{
				Hash:      "abc123",
				Type:      CloneTypeExact,
				LineCount: 25,
				Instances: []CloneInstance{
					{
						File:      "file1.go",
						StartLine: 10,
						EndLine:   35,
						NodeCount: 12,
					},
					{
						File:      "file2.go",
						StartLine: 42,
						EndLine:   67,
						NodeCount: 12,
					},
				},
			},
		},
	}

	data, err := json.Marshal(metrics)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled DuplicationMetrics
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, metrics.ClonePairs, unmarshaled.ClonePairs)
	assert.Equal(t, metrics.DuplicatedLines, unmarshaled.DuplicatedLines)
	assert.Equal(t, metrics.DuplicationRatio, unmarshaled.DuplicationRatio)
	assert.Equal(t, metrics.LargestCloneSize, unmarshaled.LargestCloneSize)
	assert.Len(t, unmarshaled.Clones, 1)
}

func TestCloneTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		cloneType CloneType
		expected  string
	}{
		{"Exact clone", CloneTypeExact, "exact"},
		{"Renamed clone", CloneTypeRenamed, "renamed"},
		{"Near clone", CloneTypeNear, "near"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.cloneType))
		})
	}
}

func TestReportWithDuplication(t *testing.T) {
	report := Report{
		Duplication: DuplicationMetrics{
			ClonePairs:       3,
			DuplicatedLines:  120,
			DuplicationRatio: 0.08,
			LargestCloneSize: 40,
		},
	}

	data, err := json.Marshal(report)
	require.NoError(t, err)
	assert.Contains(t, string(data), "duplication")
	assert.Contains(t, string(data), "clone_pairs")

	var unmarshaled Report
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, 3, unmarshaled.Duplication.ClonePairs)
	assert.Equal(t, 120, unmarshaled.Duplication.DuplicatedLines)
}

func TestClonePairStructure(t *testing.T) {
	clone := ClonePair{
		Hash:      "hash123",
		Type:      CloneTypeRenamed,
		LineCount: 15,
		Instances: []CloneInstance{
			{File: "a.go", StartLine: 1, EndLine: 15, NodeCount: 8},
			{File: "b.go", StartLine: 20, EndLine: 35, NodeCount: 8},
		},
	}

	assert.Equal(t, "hash123", clone.Hash)
	assert.Equal(t, CloneTypeRenamed, clone.Type)
	assert.Equal(t, 15, clone.LineCount)
	assert.Len(t, clone.Instances, 2)
}
