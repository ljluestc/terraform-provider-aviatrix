package aviatrix

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestFramework provides a comprehensive test framework for Terraform Plugin SDK v2
type TestFramework struct {
	Provider    *schema.Provider
	Config      *TestConfig
	Logger      *TestLogger
	TestContext context.Context
}

// NewTestFramework creates a new test framework instance
func NewTestFramework(t *testing.T) (*TestFramework, error) {
	// Load test configuration
	config, err := LoadTestConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load test configuration: %w", err)
	}

	// Create test logger
	logger, err := NewTestLogger(t.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to create test logger: %w", err)
	}

	// Create provider instance
	provider := Provider()

	framework := &TestFramework{
		Provider:    provider,
		Config:      config,
		Logger:      logger,
		TestContext: context.Background(),
	}

	logger.Info("Test framework initialized")
	logger.AddMetadata("test_name", t.Name())
	logger.AddMetadata("go_version", os.Getenv("GO_VERSION"))
	logger.AddMetadata("tf_version", os.Getenv("TF_VERSION"))
	logger.AddMetadata("enabled_providers", config.GetEnabledProviders())

	return framework, nil
}

// ProviderFactories returns the provider factories for testing
func (tf *TestFramework) ProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"aviatrix": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}
}

// ConfigureProvider configures the provider with test credentials
func (tf *TestFramework) ConfigureProvider() (*terraform.ResourceConfig, error) {
	configMap := map[string]interface{}{
		"controller_ip": tf.Config.ControllerIP,
		"username":      tf.Config.Username,
		"password":      tf.Config.Password,
	}

	tf.Logger.Info("Configuring provider with controller: %s", tf.Config.ControllerIP)

	return terraform.NewResourceConfigRaw(configMap), nil
}

// PreCheck performs pre-test validation
func (tf *TestFramework) PreCheck(t *testing.T) {
	if !IsAcceptanceTest() {
		t.Skip("Skipping acceptance test (TF_ACC not set)")
	}

	if err := tf.Config.Validate(); err != nil {
		t.Fatalf("Test configuration validation failed: %v", err)
	}

	tf.Logger.Info("Pre-check validation passed")
}

// Cleanup performs cleanup after test execution
func (tf *TestFramework) Cleanup() error {
	if tf.Logger != nil {
		return tf.Logger.Close()
	}
	return nil
}

// GetProviderConfig returns provider configuration for specific cloud provider
func (tf *TestFramework) GetProviderConfig(provider string) (map[string]interface{}, error) {
	switch provider {
	case "aws":
		if !tf.Config.AWS.Enabled {
			return nil, fmt.Errorf("AWS provider is not enabled")
		}
		return map[string]interface{}{
			"access_key": tf.Config.AWS.AccessKeyID,
			"secret_key": tf.Config.AWS.SecretAccessKey,
			"region":     tf.Config.AWS.DefaultRegion,
		}, nil

	case "azure":
		if !tf.Config.Azure.Enabled {
			return nil, fmt.Errorf("Azure provider is not enabled")
		}
		return map[string]interface{}{
			"client_id":       tf.Config.Azure.ClientID,
			"client_secret":   tf.Config.Azure.ClientSecret,
			"subscription_id": tf.Config.Azure.SubscriptionID,
			"tenant_id":       tf.Config.Azure.TenantID,
		}, nil

	case "gcp":
		if !tf.Config.GCP.Enabled {
			return nil, fmt.Errorf("GCP provider is not enabled")
		}
		return map[string]interface{}{
			"credentials": tf.Config.GCP.ApplicationCredentialsPath,
			"project":     tf.Config.GCP.Project,
			"region":      tf.Config.GCP.Region,
		}, nil

	case "oci":
		if !tf.Config.OCI.Enabled {
			return nil, fmt.Errorf("OCI provider is not enabled")
		}
		return map[string]interface{}{
			"user_id":          tf.Config.OCI.UserID,
			"tenancy_id":       tf.Config.OCI.TenancyID,
			"fingerprint":      tf.Config.OCI.Fingerprint,
			"private_key_path": tf.Config.OCI.PrivateKeyPath,
			"region":           tf.Config.OCI.Region,
		}, nil

	default:
		return nil, fmt.Errorf("unknown cloud provider: %s", provider)
	}
}

// TestAccPreCheck is a convenience function for acceptance test pre-checks
func TestAccPreCheck(t *testing.T) {
	if !IsAcceptanceTest() {
		t.Skip("Skipping acceptance test (TF_ACC not set)")
	}

	// Check core configuration
	if os.Getenv("AVIATRIX_CONTROLLER_IP") == "" {
		t.Fatal("AVIATRIX_CONTROLLER_IP must be set for acceptance tests")
	}
	if os.Getenv("AVIATRIX_USERNAME") == "" {
		t.Fatal("AVIATRIX_USERNAME must be set for acceptance tests")
	}
	if os.Getenv("AVIATRIX_PASSWORD") == "" {
		t.Fatal("AVIATRIX_PASSWORD must be set for acceptance tests")
	}
}

// ProviderMeta returns provider metadata for testing
type ProviderMeta struct {
	Client interface{}
}

// TestProviderConfigure configures the provider for testing
func TestProviderConfigure(d *schema.ResourceData) (interface{}, error) {
	// This function provides a test-friendly way to configure the provider
	// It can be overridden in tests to inject mock clients
	return &ProviderMeta{
		Client: nil, // Will be replaced with actual client in real tests
	}, nil
}

// SkipIfNoCloudProvider skips test if no cloud providers are configured
func SkipIfNoCloudProvider(t *testing.T) {
	hasProvider := false
	providers := []string{"AWS", "AZURE", "GCP", "OCI"}

	for _, provider := range providers {
		skipVar := fmt.Sprintf("SKIP_ACCOUNT_%s", provider)
		if os.Getenv(skipVar) != "yes" {
			hasProvider = true
			break
		}
	}

	if !hasProvider {
		t.Skip("Skipping test: no cloud providers are configured")
	}
}

// GetTestProviderFactories returns standard provider factories for tests
func GetTestProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"aviatrix": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}
}

// TestConfig holds configuration for test execution
type TestConfig struct {
	ControllerIP string
	Username     string
	Password     string
	AWS          AWSConfig
	Azure        AzureConfig
	GCP          GCPConfig
	OCI          OCIConfig
	TestTimeout  time.Duration
	ArtifactDir  string
}

// AWSConfig holds AWS-specific test configuration
type AWSConfig struct {
	Enabled           bool
	AccessKeyID       string
	SecretAccessKey   string
	DefaultRegion     string
	AccountNumber     string
}

// AzureConfig holds Azure-specific test configuration
type AzureConfig struct {
	Enabled        bool
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	TenantID       string
}

// GCPConfig holds GCP-specific test configuration
type GCPConfig struct {
	Enabled                      bool
	ApplicationCredentialsPath   string
	Project                      string
	Region                       string
}

// OCIConfig holds OCI-specific test configuration
type OCIConfig struct {
	Enabled        bool
	UserID         string
	TenancyID      string
	Fingerprint    string
	PrivateKeyPath string
	Region         string
}

// LoadTestConfig loads test configuration from environment variables
func LoadTestConfig() (*TestConfig, error) {
	config := &TestConfig{
		ControllerIP: os.Getenv("AVIATRIX_CONTROLLER_IP"),
		Username:     os.Getenv("AVIATRIX_USERNAME"),
		Password:     os.Getenv("AVIATRIX_PASSWORD"),
		ArtifactDir:  getEnvOrDefault("TEST_ARTIFACT_DIR", "./test-results"),
		TestTimeout:  parseTimeout(getEnvOrDefault("GO_TEST_TIMEOUT", "30m")),
	}

	// Load AWS configuration
	config.AWS = AWSConfig{
		Enabled:           os.Getenv("SKIP_ACCOUNT_AWS") != "yes",
		AccessKeyID:       os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
		DefaultRegion:     getEnvOrDefault("AWS_DEFAULT_REGION", "us-east-1"),
		AccountNumber:     os.Getenv("AWS_ACCOUNT_NUMBER"),
	}

	// Load Azure configuration
	config.Azure = AzureConfig{
		Enabled:        os.Getenv("SKIP_ACCOUNT_AZURE") != "yes",
		ClientID:       os.Getenv("ARM_CLIENT_ID"),
		ClientSecret:   os.Getenv("ARM_CLIENT_SECRET"),
		SubscriptionID: os.Getenv("ARM_SUBSCRIPTION_ID"),
		TenantID:       os.Getenv("ARM_TENANT_ID"),
	}

	// Load GCP configuration
	config.GCP = GCPConfig{
		Enabled:                    os.Getenv("SKIP_ACCOUNT_GCP") != "yes",
		ApplicationCredentialsPath: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		Project:                    os.Getenv("GOOGLE_PROJECT"),
		Region:                     getEnvOrDefault("GOOGLE_REGION", "us-central1"),
	}

	// Load OCI configuration
	config.OCI = OCIConfig{
		Enabled:        os.Getenv("SKIP_ACCOUNT_OCI") != "yes",
		UserID:         os.Getenv("OCI_USER_ID"),
		TenancyID:      os.Getenv("OCI_TENANCY_ID"),
		Fingerprint:    os.Getenv("OCI_FINGERPRINT"),
		PrivateKeyPath: os.Getenv("OCI_PRIVATE_KEY_PATH"),
		Region:         os.Getenv("OCI_REGION"),
	}

	return config, nil
}

// Validate validates the test configuration
func (tc *TestConfig) Validate() error {
	if tc.ControllerIP == "" {
		return fmt.Errorf("AVIATRIX_CONTROLLER_IP must be set")
	}
	if tc.Username == "" {
		return fmt.Errorf("AVIATRIX_USERNAME must be set")
	}
	if tc.Password == "" {
		return fmt.Errorf("AVIATRIX_PASSWORD must be set")
	}

	// Validate enabled cloud providers have required credentials
	if tc.AWS.Enabled {
		if tc.AWS.AccessKeyID == "" || tc.AWS.SecretAccessKey == "" {
			return fmt.Errorf("AWS credentials required when AWS tests are enabled")
		}
	}

	if tc.Azure.Enabled {
		if tc.Azure.ClientID == "" || tc.Azure.ClientSecret == "" ||
			tc.Azure.SubscriptionID == "" || tc.Azure.TenantID == "" {
			return fmt.Errorf("Azure credentials required when Azure tests are enabled")
		}
	}

	if tc.GCP.Enabled {
		if tc.GCP.ApplicationCredentialsPath == "" || tc.GCP.Project == "" {
			return fmt.Errorf("GCP credentials required when GCP tests are enabled")
		}
	}

	if tc.OCI.Enabled {
		if tc.OCI.UserID == "" || tc.OCI.TenancyID == "" ||
			tc.OCI.Fingerprint == "" || tc.OCI.PrivateKeyPath == "" {
			return fmt.Errorf("OCI credentials required when OCI tests are enabled")
		}
	}

	return nil
}

// GetEnabledProviders returns a list of enabled cloud providers
func (tc *TestConfig) GetEnabledProviders() []string {
	var providers []string
	if tc.AWS.Enabled {
		providers = append(providers, "aws")
	}
	if tc.Azure.Enabled {
		providers = append(providers, "azure")
	}
	if tc.GCP.Enabled {
		providers = append(providers, "gcp")
	}
	if tc.OCI.Enabled {
		providers = append(providers, "oci")
	}
	return providers
}

// TestLogger provides structured logging for tests
type TestLogger struct {
	testName    string
	logFile     *os.File
	logFilePath string
	metadata    map[string]interface{}
}

// NewTestLogger creates a new test logger
func NewTestLogger(testName string) (*TestLogger, error) {
	// Create logs directory
	artifactDir := getEnvOrDefault("TEST_ARTIFACT_DIR", "./test-results")
	logsDir := filepath.Join(artifactDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file
	timestamp := time.Now().Format("20060102-150405")
	sanitizedName := strings.ReplaceAll(testName, "/", "_")
	logFileName := fmt.Sprintf("%s-%s.log", sanitizedName, timestamp)
	logFilePath := filepath.Join(logsDir, logFileName)

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	logger := &TestLogger{
		testName:    testName,
		logFile:     logFile,
		logFilePath: logFilePath,
		metadata:    make(map[string]interface{}),
	}

	logger.Info("Test logger initialized for: %s", testName)
	return logger, nil
}

// Info logs an informational message
func (tl *TestLogger) Info(format string, args ...interface{}) {
	tl.log("INFO", format, args...)
}

// Error logs an error message
func (tl *TestLogger) Error(format string, args ...interface{}) {
	tl.log("ERROR", format, args...)
}

// Warn logs a warning message
func (tl *TestLogger) Warn(format string, args ...interface{}) {
	tl.log("WARN", format, args...)
}

// Debug logs a debug message
func (tl *TestLogger) Debug(format string, args ...interface{}) {
	if os.Getenv("ENABLE_DETAILED_LOGS") == "true" {
		tl.log("DEBUG", format, args...)
	}
}

// log writes a formatted log message
func (tl *TestLogger) log(level, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, level, tl.testName, message)

	if tl.logFile != nil {
		tl.logFile.WriteString(logLine)
	}
}

// AddMetadata adds metadata to the logger
func (tl *TestLogger) AddMetadata(key string, value interface{}) {
	tl.metadata[key] = value
	tl.Info("Added metadata: %s = %v", key, value)
}

// Close closes the logger and flushes any remaining data
func (tl *TestLogger) Close() error {
	if tl.logFile != nil {
		tl.Info("Test logger closing")
		return tl.logFile.Close()
	}
	return nil
}

// IsAcceptanceTest checks if acceptance tests are enabled
func IsAcceptanceTest() bool {
	return os.Getenv("TF_ACC") == "1"
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseTimeout parses a timeout string (e.g., "30m", "1h")
func parseTimeout(timeout string) time.Duration {
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return 30 * time.Minute // Default to 30 minutes
	}
	return duration
}
