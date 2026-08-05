package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/opd-ai/go-stats-generator/internal/api/storage"
	"github.com/opd-ai/go-stats-generator/internal/config"
)

// Server handles REST API requests.
type Server struct {
	storage     storage.ResultStore
	version     string
	allowedRoot string
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewServer creates a new API server with in-memory storage for handling HTTP-based analysis requests.
// Exposes RESTful endpoints for triggering code analysis, retrieving results, and health checks.
// Uses non-persistent storage by default; for baseline retention across restarts, use NewServerWithStorage with SQLite/JSON backend.
func NewServer(version string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		storage: storage.NewMemory(),
		version: version,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// NewServerWithStorage creates a new API server with custom storage backend for durable result persistence.
// Accepts any storage.ResultStore implementation (Memory, SQLite, PostgreSQL, MongoDB) to control data retention strategy.
// Use this when analysis results must survive server restarts or be shared across multiple server instances.
func NewServerWithStorage(version string, store storage.ResultStore) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		storage: store,
		version: version,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// NewServerWithConfig creates a new API server with custom storage and allowed root directory for path validation.
// The allowedRoot restricts analysis to paths within this directory, preventing path traversal attacks.
// Use this for production deployments to sandbox the API server.
func NewServerWithConfig(version string, store storage.ResultStore, allowedRoot string) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		storage:     store,
		version:     version,
		allowedRoot: allowedRoot,
		ctx:         ctx,
		cancel:      cancel,
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

	// Check if server is shutting down
	select {
	case <-s.ctx.Done():
		writeError(w, http.StatusServiceUnavailable, "server is shutting down")
		return
	default:
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

	// Validate path against allowed root to prevent path traversal
	validatedPath, err := s.validateAnalysisPath(req.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	analysisID := uuid.New().String()
	s.storage.Store(&AnalysisResult{
		ID:     analysisID,
		Status: "pending",
	})

	go s.runAnalysis(analysisID, &req, validatedPath)

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
func (s *Server) runAnalysis(id string, req *AnalyzeRequest, validatedPath string) {
	s.wg.Add(1)
	defer s.wg.Done()

	select {
	case <-s.ctx.Done():
		// Server is shutting down, mark analysis as cancelled
		s.storage.Store(&AnalysisResult{
			ID:     id,
			Status: "cancelled",
			Error:  fmt.Errorf("server shutdown"),
		})
		return
	default:
	}

	s.storage.Store(&AnalysisResult{
		ID:     id,
		Status: "running",
	})

	cfg := buildConfig(req)

	// Run analysis with cancellation support
	report, err := executeAnalysisWithPath(cfg, validatedPath)

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

// Shutdown gracefully stops the server by cancelling the context and waiting for
// all in-flight analysis goroutines to complete.
func (s *Server) Shutdown() {
	s.cancel()
	s.wg.Wait()
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
	cfg.Filters.OnlyTestFiles = req.OnlyTests

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
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "api: failed to encode JSON response: %v\n", err)
	}
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// validateAnalysisPath validates that the requested path is within the allowed root directory.
// Returns the cleaned absolute path on success, or an error if the path is outside the allowed root.
func (s *Server) validateAnalysisPath(requestedPath string) (string, error) {
	// If no allowed root is configured, allow any path (backward compatibility)
	if s.allowedRoot == "" {
		return requestedPath, nil
	}

	// Reject absolute paths in the request
	if filepath.IsAbs(requestedPath) {
		return "", fmt.Errorf("absolute paths not allowed; use relative paths within allowed root")
	}

	// Clean the requested path to resolve . and .. components
	cleaned := filepath.Clean(requestedPath)

	// Check for directory traversal attempts
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, string(filepath.Separator)+"..") {
		return "", fmt.Errorf("path traversal not allowed: %s", requestedPath)
	}

	// Get absolute path of allowed root
	allowedAbs, err := filepath.Abs(s.allowedRoot)
	if err != nil {
		return "", fmt.Errorf("failed to resolve allowed root: %w", err)
	}

	// Resolve requested path relative to allowed root
	requestedAbs := filepath.Join(allowedAbs, cleaned)

	// Ensure requested path is within allowed root
	rel, err := filepath.Rel(allowedAbs, requestedAbs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("requested path outside allowed root: %s", requestedPath)
	}

	return requestedAbs, nil
}
