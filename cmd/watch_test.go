package cmd

import (
	"sync"
	"testing"
	"time"
)

func TestDebouncer(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		triggers int
		expected int
	}{
		{
			name:     "single trigger",
			duration: 10 * time.Millisecond,
			triggers: 1,
			expected: 1,
		},
		{
			name:     "multiple triggers within window",
			duration: 50 * time.Millisecond,
			triggers: 5,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			count := 0
			debouncer := newDebouncer(tt.duration, func() {
				mu.Lock()
				count++
				mu.Unlock()
			})

			for i := 0; i < tt.triggers; i++ {
				debouncer.trigger()
				time.Sleep(5 * time.Millisecond)
			}

			time.Sleep(tt.duration + 20*time.Millisecond)

			mu.Lock()
			actual := count
			mu.Unlock()

			if actual != tt.expected {
				t.Errorf("expected %d executions, got %d", tt.expected, actual)
			}
		})
	}
}

func TestIsGoFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"pkg/analyzer/function.go", true},
		{"main_test.go", false},
		{"pkg/analyzer/function_test.go", false},
		{"README.md", false},
		{"config.yaml", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isGoFile(tt.path)
			if result != tt.expected {
				t.Errorf("isGoFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}
