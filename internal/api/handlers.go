package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/opd-ai/go-stats-generator/internal/api/storage"
	"github.com/opd-ai/go-stats-generator/internal/config"
)

// Server handles REST API requests.
type Server struct {
	storage storage.ResultStore
	version string
}

// NewServer creates a new API server with in-memory storage.
func NewServer(version string) *Server {
	return NewServerWithStorage(version, storage.NewMemory())
}

// NewServerWithStorage creates a new API server with custom storage.
func NewServerWithStorage(version string, store storage.ResultStore) *Server {
	return &Server{
		storage: store,
		version: version,
	}
}

// HandleHealth returns API health status.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resp := HealthResponse{
		Status:  "ok",
		Version: s.version,
	}
	writeJSON(w, http.StatusOK, resp)
}

// HandleAnalyze processes analysis requests asynchronously.
func (s *Server) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	analysisID := uuid.New().String()
	s.storage.Store(&AnalysisResult{
		ID:     analysisID,
		Status: "pending",
	})

	go s.runAnalysis(analysisID, &req)

	resp := AnalyzeResponse{
		ID:     analysisID,
		Status: "pending",
	}
	writeJSON(w, http.StatusAccepted, resp)
}

// HandleReport retrieves analysis results by ID.
func (s *Server) HandleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.URL.Path[len("/api/v1/report/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, "analysis ID required")
		return
	}

	result, ok := s.storage.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "analysis not found")
		return
	}

	resp := ReportResponse{
		ID:     result.ID,
		Status: result.Status,
		Report: result.Report,
	}
	if result.Error != nil {
		resp.Error = result.Error.Error()
	}

	writeJSON(w, http.StatusOK, resp)
}

// runAnalysis executes the analysis in a goroutine.
func (s *Server) runAnalysis(id string, req *AnalyzeRequest) {
	s.storage.Store(&AnalysisResult{
		ID:     id,
		Status: "running",
	})

	cfg := buildConfig(req)
	report, err := executeAnalysisWithPath(cfg, req.Path)

	status := "completed"
	if err != nil {
		status = "failed"
	}

	s.storage.Store(&AnalysisResult{
		ID:     id,
		Status: status,
		Report: report,
		Error:  err,
	})
}

// buildConfig creates configuration from API request.
func buildConfig(req *AnalyzeRequest) *config.Config {
	cfg := config.DefaultConfig()

	if len(req.Include) > 0 {
		cfg.Filters.IncludePatterns = req.Include
	}
	if len(req.Exclude) > 0 {
		cfg.Filters.ExcludePatterns = req.Exclude
	}
	cfg.Filters.SkipTestFiles = req.SkipTests

	if req.MaxFunctionLength > 0 {
		cfg.Analysis.MaxFunctionLength = req.MaxFunctionLength
	}
	if req.MaxComplexity > 0 {
		cfg.Analysis.MaxCyclomaticComplexity = req.MaxComplexity
	}

	cfg.Performance.Timeout = 5 * time.Minute

	return cfg
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
