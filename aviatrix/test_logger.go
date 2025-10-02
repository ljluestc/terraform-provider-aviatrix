package aviatrix

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestLogger provides enhanced logging capabilities for tests
type TestLogger struct {
	t           *testing.T
	logFile     *os.File
	logger      *log.Logger
	mu          sync.Mutex
	startTime   time.Time
	testName    string
	artifactDir string
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T, testName string) (*TestLogger, error) {
	config := DefaultTestConfig()

	// Ensure directories exist
	if err := config.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Create log file
	logPath := config.LogPath(testName)
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Create multi-writer to write to both file and test output
	multiWriter := io.MultiWriter(logFile, &testWriter{t: t})
	logger := log.New(multiWriter, "", log.LstdFlags|log.Lmicroseconds)

	tl := &TestLogger{
		t:           t,
		logFile:     logFile,
		logger:      logger,
		startTime:   time.Now(),
		testName:    testName,
		artifactDir: config.ArtifactDir,
	}

	tl.Info("Test started: %s", testName)
	return tl, nil
}

// testWriter is a writer that writes to testing.T.Log
type testWriter struct {
	t *testing.T
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

// Info logs an informational message
func (tl *TestLogger) Info(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.logger.Printf("[INFO] "+format, args...)
}

// Debug logs a debug message
func (tl *TestLogger) Debug(format string, args ...interface{}) {
	if os.Getenv("ENABLE_DETAILED_LOGS") == "true" {
		tl.mu.Lock()
		defer tl.mu.Unlock()
		tl.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Warn logs a warning message
func (tl *TestLogger) Warn(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.logger.Printf("[WARN] "+format, args...)
}

// Error logs an error message
func (tl *TestLogger) Error(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.logger.Printf("[ERROR] "+format, args...)
}

// Fatal logs a fatal error and fails the test
func (tl *TestLogger) Fatal(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.logger.Printf("[FATAL] "+format, args...)
	tl.t.Fatalf(format, args...)
}

// Step logs a test step
func (tl *TestLogger) Step(stepNum int, description string) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.logger.Printf("[STEP %d] %s", stepNum, description)
}

// Resource logs resource-related information
func (tl *TestLogger) Resource(action, resourceType, resourceName string) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.logger.Printf("[RESOURCE] %s %s.%s", action, resourceType, resourceName)
}

// Duration logs the duration of an operation
func (tl *TestLogger) Duration(operation string, duration time.Duration) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.logger.Printf("[DURATION] %s took %s", operation, duration)
}

// Close closes the log file and logs test summary
func (tl *TestLogger) Close() {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	duration := time.Since(tl.startTime)
	tl.logger.Printf("[INFO] Test completed: %s (duration: %s)", tl.testName, duration)

	if tl.logFile != nil {
		tl.logFile.Close()
	}
}

// SaveArtifact saves a test artifact (e.g., state file, config)
func (tl *TestLogger) SaveArtifact(name string, content []byte) error {
	artifactPath := filepath.Join(tl.artifactDir, fmt.Sprintf("%s_%s", tl.testName, name))

	tl.Info("Saving artifact: %s", artifactPath)

	if err := os.WriteFile(artifactPath, content, 0644); err != nil {
		tl.Error("Failed to save artifact %s: %v", artifactPath, err)
		return err
	}

	return nil
}

// TestMetrics tracks test execution metrics
type TestMetrics struct {
	mu               sync.Mutex
	TestName         string
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
	ResourcesCreated int
	ResourcesDeleted int
	APICallCount     int
	Errors           []string
	Warnings         []string
}

// NewTestMetrics creates a new test metrics tracker
func NewTestMetrics(testName string) *TestMetrics {
	return &TestMetrics{
		TestName:  testName,
		StartTime: time.Now(),
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
	}
}

// RecordResourceCreated increments the created resources counter
func (tm *TestMetrics) RecordResourceCreated() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.ResourcesCreated++
}

// RecordResourceDeleted increments the deleted resources counter
func (tm *TestMetrics) RecordResourceDeleted() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.ResourcesDeleted++
}

// RecordAPICall increments the API call counter
func (tm *TestMetrics) RecordAPICall() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.APICallCount++
}

// RecordError records an error message
func (tm *TestMetrics) RecordError(err string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Errors = append(tm.Errors, err)
}

// RecordWarning records a warning message
func (tm *TestMetrics) RecordWarning(warning string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.Warnings = append(tm.Warnings, warning)
}

// Finalize calculates final metrics
func (tm *TestMetrics) Finalize() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.EndTime = time.Now()
	tm.Duration = tm.EndTime.Sub(tm.StartTime)
}

// Summary returns a formatted summary of the metrics
func (tm *TestMetrics) Summary() string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	return fmt.Sprintf(`
Test Metrics for %s:
  Duration: %s
  Resources Created: %d
  Resources Deleted: %d
  API Calls: %d
  Errors: %d
  Warnings: %d
`, tm.TestName, tm.Duration, tm.ResourcesCreated, tm.ResourcesDeleted,
		tm.APICallCount, len(tm.Errors), len(tm.Warnings))
}

// WriteMetricsToFile writes metrics to a JSON file
func (tm *TestMetrics) WriteMetricsToFile(path string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	content := fmt.Sprintf(`{
  "test_name": "%s",
  "start_time": "%s",
  "end_time": "%s",
  "duration_seconds": %.2f,
  "resources_created": %d,
  "resources_deleted": %d,
  "api_call_count": %d,
  "error_count": %d,
  "warning_count": %d,
  "errors": %v,
  "warnings": %v
}`, tm.TestName, tm.StartTime.Format(time.RFC3339), tm.EndTime.Format(time.RFC3339),
		tm.Duration.Seconds(), tm.ResourcesCreated, tm.ResourcesDeleted,
		tm.APICallCount, len(tm.Errors), len(tm.Warnings),
		tm.Errors, tm.Warnings)

	return os.WriteFile(path, []byte(content), 0644)
}
