package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	baselineID      string
	baselineMessage string
	tags            []string
	overwriteFlag   bool
)

var baselineCmd = &cobra.Command{
	Use:   "baseline [path]",
	Short: "Create or manage baseline snapshots for comparison",
	Long: `Create a new baseline snapshot from the current codebase state,
or manage existing baselines. Baselines are used as reference points
for complexity comparison and regression detection.`,
	RunE: runBaseline,
}

var createBaselineCmd = &cobra.Command{
	Use:   "create [path]",
	Short: "Create a new baseline snapshot",
	Long: `Create a new baseline snapshot from the current codebase state.
This snapshot can be used later for comparisons and trend analysis.`,
	RunE: runCreateBaseline,
}

var listBaselinesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available baseline snapshots",
	Long:  "List all available baseline snapshots with their metadata.",
	RunE:  runListBaselines,
}

var deleteBaselineCmd = &cobra.Command{
	Use:   "delete <baseline-id>",
	Short: "Delete a baseline snapshot",
	Long:  "Delete a specific baseline snapshot by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeleteBaseline,
}

// init registers the baseline command and its subcommands with the root command.
func init() {
	// Add baseline command to root
	rootCmd.AddCommand(baselineCmd)

	// Add subcommands to baseline
	baselineCmd.AddCommand(createBaselineCmd)
	baselineCmd.AddCommand(listBaselinesCmd)
	baselineCmd.AddCommand(deleteBaselineCmd)

	// Flags for create baseline
	createBaselineCmd.Flags().StringVarP(&baselineID, "id", "i", "", "Custom ID for the baseline (auto-generated if not specified)")
	createBaselineCmd.Flags().StringVarP(&baselineMessage, "message", "m", "", "Description message for the baseline")
	createBaselineCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags to associate with the baseline")
	createBaselineCmd.Flags().BoolVar(&overwriteFlag, "overwrite", false, "Overwrite existing baseline with same ID")

	// Global flags inherited from root
	createBaselineCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json, console)")
	createBaselineCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	createBaselineCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Flags for list baselines
	listBaselinesCmd.Flags().StringVarP(&outputFormat, "format", "f", "console", "Output format (json, console)")
	listBaselinesCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
}

// runBaseline executes the default baseline behavior by creating a new baseline snapshot.
func runBaseline(cmd *cobra.Command, args []string) error {
	// Default behavior is to create a baseline
	return runCreateBaseline(cmd, args)
}

// runCreateBaseline creates a new baseline snapshot from the analyzed codebase
func runCreateBaseline(cmd *cobra.Command, args []string) error {
	targetPath := extractTargetPath(args)
	if verbose {
		fmt.Fprintf(os.Stderr, "Creating baseline snapshot for: %s\n", targetPath)
	}

	// Initialize storage backend
	storageBackend, err := initializeStorageBackend()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storageBackend.Close()

	// Analyze the codebase
	report, err := analyzeCodebase(targetPath)
	if err != nil {
		return fmt.Errorf("failed to analyze codebase: %w", err)
	}

	// Generate baseline ID if not provided
	if baselineID == "" {
		baselineID = generateBaselineID()
	}

	// Create and store snapshot
	snapshot := createSnapshot(baselineID, report)
	if err := storeSnapshotWithRetry(storageBackend, snapshot); err != nil {
		return err
	}

	// Output the result
	return outputBaselineResult(snapshot)
}

// extractTargetPath gets the target directory from command arguments
func extractTargetPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "."
}

// initializeStorageBackend creates and configures the metrics storage backend based on application configuration.
// It reads storage type and path settings from viper configuration (respecting config files and environment variables),
// instantiates the appropriate storage implementation (SQLite or JSON), and initializes the database schema/file structure.
// Defaults to SQLite with metrics.db in current directory if not configured. Returns configured storage instance ready
// for baseline snapshot persistence, or error if initialization fails due to permissions or invalid configuration.
func initializeStorageBackend() (storage.MetricsStorage, error) {
	// Get storage configuration from viper (respects loaded config files)
	storageType := viper.GetString("storage.type")
	if storageType == "" {
		storageType = "sqlite" // Default to SQLite if not specified
	}

	// Get path configuration (works for both SQLite and JSON)
	storagePath := viper.GetString("storage.path")
	if storagePath == "" {
		if storageType == "json" {
			storagePath = ".go-stats-generator/snapshots" // Default directory for JSON
		} else {
			storagePath = ".go-stats-generator/metrics.db" // Default path for SQLite
		}
	}

	// Check for more specific JSON configuration
	jsonDir := viper.GetString("storage.json.directory")
	if jsonDir != "" {
		storagePath = jsonDir
	}

	compression := viper.GetBool("storage.compression")
	jsonCompression := viper.GetBool("storage.json.compression")
	if jsonCompression {
		compression = jsonCompression
	}

	jsonPretty := viper.GetBool("storage.json.pretty")

	// Convert to storage.Config format
	storageConfig := storage.Config{
		Type: storageType,
	}

	// Configure SQLite settings based on configuration
	storageConfig.SQLite = storage.SQLiteConfig{
		Path:              storagePath,
		EnableWAL:         true, // Sensible default for performance
		MaxConnections:    10,   // Sensible default for CLI tool
		EnableFK:          true, // Sensible default for data integrity
		EnableCompression: compression,
	}

	// Configure JSON settings
	storageConfig.JSON = storage.JSONConfig{
		Directory:   storagePath,
		Compression: compression,
		Pretty:      jsonPretty,
	}

	// Use the proper storage factory function that respects configuration
	return storage.NewStorage(storageConfig)
}

// convertToStorageConfig converts config.StorageConfig to storage.Config.
func convertToStorageConfig(cfg config.StorageConfig) storage.Config {
	storageConfig := storage.Config{
		Type: cfg.Type,
	}

	// Configure SQLite settings based on configuration
	storageConfig.SQLite = storage.SQLiteConfig{
		Path:              cfg.Path,
		EnableWAL:         true, // Sensible default for performance
		MaxConnections:    10,   // Sensible default for CLI tool
		EnableFK:          true, // Sensible default for data integrity
		EnableCompression: cfg.Compression,
	}

	// Configure JSON settings (for when JSON storage is implemented)
	storageConfig.JSON = storage.JSONConfig{
		Directory:   cfg.Path,
		Compression: cfg.Compression,
		Pretty:      false, // Compact JSON for storage efficiency
	}

	return storageConfig
}

// createSnapshot builds a snapshot with metadata from the analysis report
func createSnapshot(baselineID string, report *metrics.Report) metrics.Snapshot {
	metadata := metrics.SnapshotMetadata{
		Timestamp:   time.Now(),
		GitBranch:   getCurrentBranch(),
		GitCommit:   getCurrentCommit(),
		Description: baselineMessage,
		Tags:        convertToTagMap(tags),
	}

	return metrics.Snapshot{
		ID:       baselineID,
		Report:   *report,
		Metadata: metadata,
	}
}

// storeSnapshotWithRetry stores the snapshot, retrying with overwrite if necessary
func storeSnapshotWithRetry(storageBackend storage.MetricsStorage, snapshot metrics.Snapshot) error {
	ctx := context.Background()
	err := storageBackend.Store(ctx, snapshot, snapshot.Metadata)
	if err == nil {
		return nil
	}

	if !overwriteFlag {
		return fmt.Errorf("failed to store baseline snapshot: %w", err)
	}

	// Attempt overwrite by deleting first
	if err := storageBackend.Delete(ctx, snapshot.ID); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: could not delete existing baseline: %v\n", err)
	}

	if err := storageBackend.Store(ctx, snapshot, snapshot.Metadata); err != nil {
		return fmt.Errorf("failed to overwrite baseline snapshot: %w", err)
	}

	return nil
}

// outputBaselineResult writes the creation result in the specified format
func outputBaselineResult(snapshot metrics.Snapshot) error {
	if outputFormat == "console" {
		return outputConsoleBaselineResult(snapshot)
	}
	return outputJSONBaselineResult(snapshot)
}

// outputConsoleBaselineResult writes a human-readable baseline creation result
func outputConsoleBaselineResult(snapshot metrics.Snapshot) error {
	fmt.Printf("✓ Baseline snapshot created successfully\n")
	fmt.Printf("  ID: %s\n", snapshot.ID)
	fmt.Printf("  Timestamp: %s\n", snapshot.Metadata.Timestamp.Format("2006-01-02 15:04:05"))
	if snapshot.Metadata.Description != "" {
		fmt.Printf("  Message: %s\n", snapshot.Metadata.Description)
	}
	if len(snapshot.Metadata.Tags) > 0 {
		fmt.Printf("  Tags: %v\n", snapshot.Metadata.Tags)
	}
	return nil
}

// outputJSONBaselineResult writes the baseline creation result as JSON
func outputJSONBaselineResult(snapshot metrics.Snapshot) error {
	output := map[string]interface{}{
		"status":   "success",
		"baseline": snapshot,
	}

	outputWriter := os.Stdout
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		outputWriter = file
	}

	encoder := json.NewEncoder(outputWriter)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// runListBaselines retrieves and displays all stored baseline snapshots.
func runListBaselines(cmd *cobra.Command, args []string) error {
	storageBackend, err := initializeStorageBackend()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storageBackend.Close()

	snapshots, err := retrieveSnapshots(storageBackend)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	return outputBaselines(snapshots)
}

// retrieveSnapshots fetches all snapshots from storage with default filtering
func retrieveSnapshots(storageBackend storage.MetricsStorage) ([]storage.SnapshotInfo, error) {
	ctx := context.Background()
	filter := storage.SnapshotFilter{
		Limit: 100,
	}
	return storageBackend.List(ctx, filter)
}

// outputBaselines formats and outputs the baseline snapshots based on format
func outputBaselines(snapshots []storage.SnapshotInfo) error {
	if outputFormat == "console" {
		return outputBaselinesConsole(snapshots)
	}
	return outputBaselinesJSON(snapshots)
}

// outputBaselinesConsole outputs baseline snapshots in human-readable console format
func outputBaselinesConsole(snapshots []storage.SnapshotInfo) error {
	if len(snapshots) == 0 {
		fmt.Println("No baseline snapshots found.")
		return nil
	}

	fmt.Printf("Found %d baseline snapshot(s):\n\n", len(snapshots))
	for _, info := range snapshots {
		printSnapshotInfo(info)
	}
	return nil
}

// printSnapshotInfo prints detailed information about a single snapshot
func printSnapshotInfo(info storage.SnapshotInfo) {
	fmt.Printf("ID: %s\n", info.ID)
	fmt.Printf("  Timestamp: %s\n", info.Timestamp.Format("2006-01-02 15:04:05"))

	if info.Description != "" {
		fmt.Printf("  Message: %s\n", info.Description)
	}
	if len(info.Tags) > 0 {
		fmt.Printf("  Tags: %v\n", info.Tags)
	}
	if info.GitBranch != "" {
		fmt.Printf("  Branch: %s\n", info.GitBranch)
	}
	if info.GitCommit != "" {
		fmt.Printf("  Commit: %s\n", info.GitCommit)
	}

	fmt.Printf("  Size: %d bytes\n", info.Size)
	fmt.Println()
}

// outputBaselinesJSON outputs baseline snapshots in JSON format
func outputBaselinesJSON(snapshots []storage.SnapshotInfo) error {
	outputWriter, err := createOutputWriter()
	if err != nil {
		return fmt.Errorf("failed to create output writer: %w", err)
	}
	defer outputWriter.Close()

	encoder := json.NewEncoder(outputWriter)
	encoder.SetIndent("", "  ")
	return encoder.Encode(map[string]interface{}{
		"snapshots": snapshots,
		"count":     len(snapshots),
	})
}

// createOutputWriter creates the appropriate output writer based on outputFile setting
func createOutputWriter() (io.WriteCloser, error) {
	if outputFile == "" {
		return &nopCloser{os.Stdout}, nil
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// nopCloser wraps an io.Writer to provide a no-op Close method
type nopCloser struct {
	io.Writer
}

// Close implements io.Closer with a no-op for nopCloser.
func (nopCloser) Close() error { return nil }

// runDeleteBaseline removes a stored baseline snapshot from the storage backend
// by its ID, providing user feedback on success or failure.
func runDeleteBaseline(cmd *cobra.Command, args []string) error {
	baselineID := args[0]

	storageBackend, err := initializeStorageBackend()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer storageBackend.Close()

	if err := deleteSnapshot(storageBackend, baselineID); err != nil {
		return err
	}

	return outputDeleteSuccess(baselineID)
}

// deleteSnapshot removes a baseline snapshot from storage.
func deleteSnapshot(storageBackend storage.MetricsStorage, baselineID string) error {
	ctx := context.Background()
	if err := storageBackend.Delete(ctx, baselineID); err != nil {
		return fmt.Errorf("failed to delete baseline snapshot: %w", err)
	}
	return nil
}

// outputDeleteSuccess writes success message for baseline deletion.
func outputDeleteSuccess(baselineID string) error {
	if outputFormat == "console" {
		fmt.Printf("✓ Baseline snapshot '%s' deleted successfully\n", baselineID)
		return nil
	}

	return writeDeleteJSON(baselineID)
}

// writeDeleteJSON writes JSON output for successful deletion.
func writeDeleteJSON(baselineID string) error {
	output := map[string]interface{}{
		"status":     "success",
		"message":    "Baseline snapshot deleted",
		"baselineId": baselineID,
	}

	outputWriter, err := getOutputWriter()
	if err != nil {
		return err
	}
	if outputWriter != os.Stdout {
		defer outputWriter.Close()
	}

	encoder := json.NewEncoder(outputWriter)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// getOutputWriter returns the appropriate output writer.
func getOutputWriter() (*os.File, error) {
	if outputFile == "" {
		return os.Stdout, nil
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	return file, nil
}

// Helper functions

// analyzeCodebase performs code analysis on the target directory using the full workflow.
// This ensures all metrics (including scores and documentation) are properly populated.
func analyzeCodebase(targetPath string) (*metrics.Report, error) {
	// Load default configuration for full analysis
	// DefaultConfig enables all analyzers by default
	cfg := config.DefaultConfig()

	// Disable progress output for baseline creation
	cfg.Output.ShowProgress = false
	cfg.Output.Verbose = false

	// Use the full analysis workflow to ensure all metrics are populated
	report, err := runAnalysisWorkflow(context.Background(), targetPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	return report, nil
}

// generateBaselineID creates a unique baseline identifier based on the current timestamp.
func generateBaselineID() string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("baseline-%s", timestamp)
}

// getCurrentBranch attempts to retrieve the current git branch name.
func getCurrentBranch() string {
	// Try to get current git branch
	// This is a placeholder - in real implementation you'd use git commands
	return ""
}

// getCurrentCommit attempts to retrieve the current git commit hash.
func getCurrentCommit() string {
	// Try to get current git commit hash
	// This is a placeholder - in real implementation you'd use git commands
	return ""
}

// convertToTagMap converts a slice of key=value tag strings to a map.
func convertToTagMap(tagSlice []string) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range tagSlice {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) == 2 {
			tagMap[parts[0]] = parts[1]
		} else {
			tagMap[tag] = "true"
		}
	}
	return tagMap
}
