package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// Postgres provides PostgreSQL-backed storage for analysis results.
type Postgres struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewPostgres creates a new PostgreSQL storage instance.
// connectionString format: "postgres://user:password@host:port/dbname?sslmode=disable"
func NewPostgres(connectionString string) (*Postgres, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	p := &Postgres{db: db}
	if err := p.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return p, nil
}

// initSchema creates the analysis_results table if it doesn't exist.
func (p *Postgres) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS analysis_results (
		id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		report JSONB,
		error TEXT
	)`

	_, err := p.db.Exec(query)
	return err
}

// Store saves an analysis result.
func (p *Postgres) Store(result *AnalysisResult) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var reportJSON []byte
	var errorText *string

	if result.Report != nil {
		reportJSON, _ = json.Marshal(result.Report)
	}
	if result.Error != nil {
		errStr := result.Error.Error()
		errorText = &errStr
	}

	query := `
	INSERT INTO analysis_results (id, status, report, error)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE SET
		status = EXCLUDED.status,
		report = EXCLUDED.report,
		error = EXCLUDED.error`

	p.db.Exec(query, result.ID, result.Status, reportJSON, errorText)
}

// Get retrieves an analysis result by ID.
func (p *Postgres) Get(id string) (*AnalysisResult, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	query := `SELECT id, status, report, error FROM analysis_results WHERE id = $1`

	var result AnalysisResult
	var reportJSON []byte
	var errorText *string

	err := p.db.QueryRow(query, id).Scan(
		&result.ID,
		&result.Status,
		&reportJSON,
		&errorText,
	)

	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	if len(reportJSON) > 0 {
		var report metrics.Report
		if json.Unmarshal(reportJSON, &report) == nil {
			result.Report = &report
		}
	}
	if errorText != nil {
		result.Error = fmt.Errorf(*errorText)
	}

	return &result, true
}

// List returns all stored analysis results.
func (p *Postgres) List() []*AnalysisResult {
	p.mu.RLock()
	defer p.mu.RUnlock()

	query := `SELECT id, status, report, error FROM analysis_results`

	rows, err := p.db.Query(query)
	if err != nil {
		return []*AnalysisResult{}
	}
	defer rows.Close()

	return collectResults(rows)
}

func collectResults(rows *sql.Rows) []*AnalysisResult {
	results := []*AnalysisResult{}
	for rows.Next() {
		if result := scanResult(rows); result != nil {
			results = append(results, result)
		}
	}
	return results
}

func scanResult(rows *sql.Rows) *AnalysisResult {
	var result AnalysisResult
	var reportJSON []byte
	var errorText *string

	if err := rows.Scan(&result.ID, &result.Status, &reportJSON, &errorText); err != nil {
		return nil
	}

	populateReport(&result, reportJSON)
	populateError(&result, errorText)

	return &result
}

func populateReport(result *AnalysisResult, reportJSON []byte) {
	if len(reportJSON) > 0 {
		var report metrics.Report
		if json.Unmarshal(reportJSON, &report) == nil {
			result.Report = &report
		}
	}
}

func populateError(result *AnalysisResult, errorText *string) {
	if errorText != nil {
		result.Error = fmt.Errorf(*errorText)
	}
}

// Delete removes an analysis result by ID.
func (p *Postgres) Delete(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	query := `DELETE FROM analysis_results WHERE id = $1`
	result, err := p.db.Exec(query, id)
	if err != nil {
		return false
	}

	rows, _ := result.RowsAffected()
	return rows > 0
}

// Clear removes all stored analysis results.
func (p *Postgres) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.db.Exec(`DELETE FROM analysis_results`)
}

// Close closes the database connection.
func (p *Postgres) Close() error {
	return p.db.Close()
}
