package aviatrix

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TestLogger provides structured logging for test execution
type TestLogger struct {
	logFile      *os.File
	artifactDir  string
	testName     string
	startTime    time.Time
	mu           sync.Mutex
	metadata     map[string]interface{}
	enableStdout bool
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	TestName  string                 `json:"test_name"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewTestLogger creates a new test logger instance
func NewTestLogger(testName string) (*TestLogger, error) {
	artifactDir := GetEnvOrDefault("TEST_ARTIFACT_DIR", "./test-results")
	logDir := filepath.Join(artifactDir, "logs")

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("%s-%s.log", testName, timestamp)
	logFilePath := filepath.Join(logDir, logFileName)

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	logger := &TestLogger{
		logFile:      logFile,
		artifactDir:  artifactDir,
		testName:     testName,
		startTime:    time.Now(),
		metadata:     make(map[string]interface{}),
		enableStdout: parseBool(os.Getenv("ENABLE_DETAILED_LOGS"), false),
	}

	logger.Info("Test logger initialized")
	return logger, nil
}

// Info logs an informational message
func (l *TestLogger) Info(message string, args ...interface{}) {
	l.log("INFO", message, args...)
}

// Debug logs a debug message
func (l *TestLogger) Debug(message string, args ...interface{}) {
	l.log("DEBUG", message, args...)
}

// Warn logs a warning message
func (l *TestLogger) Warn(message string, args ...interface{}) {
	l.log("WARN", message, args...)
}

// Error logs an error message
func (l *TestLogger) Error(message string, args ...interface{}) {
	l.log("ERROR", message, args...)
}

// log writes a log entry
func (l *TestLogger) log(level, message string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	formattedMessage := fmt.Sprintf(message, args...)
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		TestName:  l.testName,
		Level:     level,
		Message:   formattedMessage,
		Metadata:  l.metadata,
	}

	// Write JSON entry to log file
	if l.logFile != nil {
		jsonData, _ := json.Marshal(entry)
		l.logFile.WriteString(string(jsonData) + "\n")
	}

	// Also write to stdout if enabled
	if l.enableStdout {
		fmt.Printf("[%s] [%s] %s: %s\n", entry.Timestamp, level, l.testName, formattedMessage)
	}
}

// AddMetadata adds metadata to all subsequent log entries
func (l *TestLogger) AddMetadata(key string, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.metadata[key] = value
}

// RemoveMetadata removes metadata
func (l *TestLogger) RemoveMetadata(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.metadata, key)
}

// LogTestStep logs a test step execution
func (l *TestLogger) LogTestStep(stepName string, stepFunc func() error) error {
	l.Info("Starting test step: %s", stepName)
	start := time.Now()

	err := stepFunc()
	duration := time.Since(start)

	if err != nil {
		l.Error("Test step failed: %s (duration: %s, error: %v)", stepName, duration, err)
		return err
	}

	l.Info("Test step completed: %s (duration: %s)", stepName, duration)
	return nil
}

// SaveArtifact saves a test artifact (e.g., Terraform state, configuration)
func (l *TestLogger) SaveArtifact(name string, content []byte) error {
	artifactPath := filepath.Join(l.artifactDir, name)
	artifactDir := filepath.Dir(artifactPath)

	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	if err := os.WriteFile(artifactPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write artifact: %w", err)
	}

	l.Info("Saved artifact: %s", name)
	return nil
}

// SaveArtifactFromReader saves a test artifact from an io.Reader
func (l *TestLogger) SaveArtifactFromReader(name string, reader io.Reader) error {
	artifactPath := filepath.Join(l.artifactDir, name)
	artifactDir := filepath.Dir(artifactPath)

	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifact directory: %w", err)
	}

	outFile, err := os.Create(artifactPath)
	if err != nil {
		return fmt.Errorf("failed to create artifact file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to copy artifact data: %w", err)
	}

	l.Info("Saved artifact from reader: %s", name)
	return nil
}

// LogResourceCreation logs resource creation with details
func (l *TestLogger) LogResourceCreation(resourceType, resourceName string, attributes map[string]interface{}) {
	l.AddMetadata("resource_type", resourceType)
	l.AddMetadata("resource_name", resourceName)
	l.Info("Creating resource: %s.%s", resourceType, resourceName)

	if len(attributes) > 0 {
		l.Debug("Resource attributes: %+v", attributes)
	}
}

// LogResourceDestruction logs resource destruction
func (l *TestLogger) LogResourceDestruction(resourceType, resourceName string) {
	l.Info("Destroying resource: %s.%s", resourceType, resourceName)
}

// LogAPICall logs an API call
func (l *TestLogger) LogAPICall(method, endpoint string, params map[string]interface{}) {
	l.Debug("API Call: %s %s (params: %+v)", method, endpoint, params)
}

// LogAPIResponse logs an API response
func (l *TestLogger) LogAPIResponse(statusCode int, duration time.Duration, response interface{}) {
	l.Debug("API Response: status=%d, duration=%s", statusCode, duration)
	if l.enableStdout {
		l.Debug("Response data: %+v", response)
	}
}

// GenerateTestReport generates a summary report of the test execution
func (l *TestLogger) GenerateTestReport() error {
	duration := time.Since(l.startTime)

	report := map[string]interface{}{
		"test_name":      l.testName,
		"start_time":     l.startTime.Format(time.RFC3339),
		"end_time":       time.Now().Format(time.RFC3339),
		"duration":       duration.String(),
		"duration_ms":    duration.Milliseconds(),
		"artifact_dir":   l.artifactDir,
		"final_metadata": l.metadata,
	}

	reportData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	reportPath := filepath.Join(l.artifactDir, fmt.Sprintf("%s-report.json", l.testName))
	if err := os.WriteFile(reportPath, reportData, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	l.Info("Test report generated: %s", reportPath)
	return nil
}

// Close closes the test logger and generates final report
func (l *TestLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Info("Test execution completed")

	// Generate test report
	if err := l.GenerateTestReport(); err != nil {
		return err
	}

	// Close log file
	if l.logFile != nil {
		return l.logFile.Close()
	}

	return nil
}

// CaptureOutput captures stdout/stderr during test execution
func (l *TestLogger) CaptureOutput(name string, fn func() error) error {
	// Create capture file
	captureFile := filepath.Join(l.artifactDir, "logs", fmt.Sprintf("%s-output.log", name))
	f, err := os.Create(captureFile)
	if err != nil {
		return fmt.Errorf("failed to create capture file: %w", err)
	}
	defer f.Close()

	// Redirect stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = f
	os.Stderr = f

	// Restore on exit
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	l.Info("Capturing output for: %s", name)
	return fn()
}

// TestLoggerFactory provides a factory for creating test loggers
type TestLoggerFactory struct {
	artifactDir string
}

// NewTestLoggerFactory creates a new test logger factory
func NewTestLoggerFactory() *TestLoggerFactory {
	return &TestLoggerFactory{
		artifactDir: GetEnvOrDefault("TEST_ARTIFACT_DIR", "./test-results"),
	}
}

// CreateLogger creates a new test logger
func (f *TestLoggerFactory) CreateLogger(testName string) (*TestLogger, error) {
	return NewTestLogger(testName)
}

// CleanupOldLogs removes log files older than the specified duration
func (f *TestLoggerFactory) CleanupOldLogs(olderThan time.Duration) error {
	logDir := filepath.Join(f.artifactDir, "logs")
	cutoff := time.Now().Add(-olderThan)

	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read log directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			logPath := filepath.Join(logDir, entry.Name())
			if err := os.Remove(logPath); err != nil {
				fmt.Printf("Failed to remove old log file %s: %v\n", logPath, err)
			}
		}
	}

	return nil
}
