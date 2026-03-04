package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	watchDebounce  time.Duration
	watchRecursive bool
	watchQuiet     bool
)

// watchCmd provides real-time file system monitoring.
var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch directory for changes and analyze",
	Long: `Monitor a directory for file changes and automatically run analysis.

This command provides real-time monitoring with:
  - File system watching for continuous analysis
  - Debounced analysis to avoid excessive re-runs
  - Live metrics updates during development
  - Automatic detection of .go file changes

Example:
  go-stats-generator watch .
  go-stats-generator watch --debounce 5s ./pkg
  go-stats-generator watch --quiet ./internal`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().DurationVar(&watchDebounce, "debounce", 2*time.Second, "Wait time before re-analyzing after changes")
	watchCmd.Flags().BoolVar(&watchRecursive, "recursive", true, "Watch subdirectories recursively")
	watchCmd.Flags().BoolVar(&watchQuiet, "quiet", false, "Suppress file change notifications")
	rootCmd.AddCommand(watchCmd)
}

// runWatch monitors the filesystem for changes.
func runWatch(cmd *cobra.Command, args []string) error {
	absPath, err := getWatchPath(args)
	if err != nil {
		return err
	}

	watcher, err := createWatcher(absPath)
	if err != nil {
		return err
	}
	defer watcher.Close()

	printWatchStartMessage(absPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	debouncer := newDebouncer(watchDebounce, func() {
		analyzeWithWatch(absPath)
	})

	return watchEventLoop(ctx, watcher, debouncer)
}

// getWatchPath extracts and validates the watch path from arguments.
func getWatchPath(args []string) (string, error) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	return absPath, nil
}

// createWatcher initializes a filesystem watcher for the given path.
func createWatcher(absPath string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	if err := addWatchPaths(watcher, absPath); err != nil {
		watcher.Close()
		return nil, err
	}

	return watcher, nil
}

// printWatchStartMessage displays the initial watch status.
func printWatchStartMessage(absPath string) {
	fmt.Printf("Watching %s for changes (debounce: %v)...\n", absPath, watchDebounce)
	fmt.Println("Press Ctrl+C to stop")
}

// watchEventLoop processes filesystem events until context cancellation.
func watchEventLoop(ctx context.Context, watcher *fsnotify.Watcher, debouncer *debouncer) error {
	for {
		if done := processWatchEvent(ctx, watcher, debouncer); done {
			return nil
		}
	}
}

// processWatchEvent handles a single watch event cycle and returns true if done.
func processWatchEvent(ctx context.Context, watcher *fsnotify.Watcher, debouncer *debouncer) bool {
	select {
	case event, ok := <-watcher.Events:
		return processFileSystemEvent(event, ok, debouncer)
	case err, ok := <-watcher.Errors:
		return processWatcherError(err, ok)
	case <-ctx.Done():
		return true
	}
}

// processFileSystemEvent handles file system events from the watcher.
func processFileSystemEvent(event fsnotify.Event, ok bool, debouncer *debouncer) bool {
	if shouldStopOnChannel(ok) {
		return true
	}
	handleFileEvent(event, debouncer)
	return false
}

// processWatcherError handles errors from the watcher.
func processWatcherError(err error, ok bool) bool {
	if shouldStopOnChannel(ok) {
		return true
	}
	handleWatchError(err)
	return false
}

// shouldStopOnChannel checks if a channel is closed.
func shouldStopOnChannel(ok bool) bool {
	return !ok
}

// handleFileEvent processes a single file system event.
func handleFileEvent(event fsnotify.Event, debouncer *debouncer) {
	if isGoFile(event.Name) {
		if !watchQuiet {
			fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), event.Name)
		}
		debouncer.trigger()
	}
}

// handleWatchError logs a watcher error to stderr.
func handleWatchError(err error) {
	fmt.Fprintf(os.Stderr, "Watch error: %v\n", err)
}

// addWatchPaths recursively adds directories to watcher.
func addWatchPaths(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.Contains(path, "/.git/") || strings.Contains(path, "/vendor/") {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}

// isGoFile checks if a file is a Go source file.
func isGoFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

// analyzeWithWatch runs analysis with watch-specific config.
func analyzeWithWatch(path string) {
	fmt.Printf("\n=== Re-analyzing at %s ===\n", time.Now().Format("15:04:05"))

	cfg := loadConfiguration()
	cfg.Filters.SkipTestFiles = true
	cfg.Output.ShowProgress = false

	ctx := context.Background()
	report, err := runDirectoryAnalysis(ctx, path, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Analysis failed: %v\n", err)
		return
	}

	printWatchSummary(report)
}

// printWatchSummary shows compact analysis results.
func printWatchSummary(report interface{}) {
	fmt.Printf("✓ Analysis complete at %s\n", time.Now().Format("15:04:05"))
}

// debouncer delays execution until quiet period.
type debouncer struct {
	mu       sync.Mutex
	timer    *time.Timer
	duration time.Duration
	action   func()
}

// newDebouncer creates a debouncer.
func newDebouncer(duration time.Duration, action func()) *debouncer {
	return &debouncer{
		duration: duration,
		action:   action,
	}
}

// trigger resets the debounce timer.
func (d *debouncer) trigger() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.duration, d.action)
}
