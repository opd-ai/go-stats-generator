// Package api provides REST API tests.
package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleHealth(t *testing.T) {
	server := NewServer("1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	server.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", resp.Version)
	}
}

func TestHandleAnalyze_ValidRequest(t *testing.T) {
	server := NewServer("1.0.0")

	reqBody := AnalyzeRequest{
		Path:      "../../testdata/simple",
		SkipTests: true,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.HandleAnalyze(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var resp AnalyzeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID == "" {
		t.Error("expected non-empty analysis ID")
	}
	if resp.Status != "pending" {
		t.Errorf("expected status pending, got %s", resp.Status)
	}
}

func TestHandleAnalyze_MissingPath(t *testing.T) {
	server := NewServer("1.0.0")

	reqBody := AnalyzeRequest{}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.HandleAnalyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleReport_NotFound(t *testing.T) {
	server := NewServer("1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/report/nonexistent", nil)
	w := httptest.NewRecorder()

	server.HandleReport(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleReport_CompletedAnalysis(t *testing.T) {
	server := NewServer("1.0.0")

	// Store a completed analysis
	result := &AnalysisResult{
		ID:     "test-id",
		Status: "completed",
		Report: nil,
		Error:  nil,
	}
	server.storage.Store(result)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/report/test-id", nil)
	w := httptest.NewRecorder()

	server.HandleReport(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp ReportResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", resp.ID)
	}
	if resp.Status != "completed" {
		t.Errorf("expected status completed, got %s", resp.Status)
	}
}

func TestAnalysisWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	server := NewServer("1.0.0")

	reqBody := AnalyzeRequest{
		Path:              "../../testdata/simple",
		SkipTests:         true,
		MaxFunctionLength: 30,
		MaxComplexity:     10,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.HandleAnalyze(w, req)

	var analyzeResp AnalyzeResponse
	json.NewDecoder(w.Body).Decode(&analyzeResp)

	// Wait for analysis to complete (with timeout)
	maxWait := 10 * time.Second
	start := time.Now()

	for time.Since(start) < maxWait {
		result, ok := server.storage.Get(analyzeResp.ID)
		if !ok {
			t.Fatal("analysis result not found")
		}

		if result.Status == "completed" || result.Status == "failed" {
			if result.Status == "failed" {
				t.Fatalf("analysis failed: %v", result.Error)
			}

			// Analysis completed successfully
			if result.Report == nil {
				t.Error("expected non-nil report")
			}
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Error("analysis did not complete within timeout")
}
