package aviatrix

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TestConfig represents the central test configuration
type TestConfig struct {
	// Core Aviatrix Configuration
	ControllerIP string
	Username     string
	Password     string

	// Cloud Provider Configurations
	AWS   *AWSTestConfig
	Azure *AzureTestConfig
	GCP   *GCPTestConfig
	OCI   *OCITestConfig

	// Test Execution Configuration
	TestTimeout       string
	TestArtifactDir   string
	EnableParallel    bool
	EnableDetailedLog bool
	MaxRetries        int
}

// AWSTestConfig contains AWS-specific test configuration
type AWSTestConfig struct {
	Enabled         bool
	AccessKeyID     string
	SecretAccessKey string
	DefaultRegion   string
	AccountNumber   string
}

// AzureTestConfig contains Azure-specific test configuration
type AzureTestConfig struct {
	Enabled        bool
	ClientID       string
	ClientSecret   string
	SubscriptionID string
	TenantID       string
	Location       string
}

// GCPTestConfig contains GCP-specific test configuration
type GCPTestConfig struct {
	Enabled                    bool
	ApplicationCredentialsPath string
	Project                    string
	Region                     string
}

// OCITestConfig contains OCI-specific test configuration
type OCITestConfig struct {
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

		TestTimeout:       GetEnvOrDefault("GO_TEST_TIMEOUT", "30m"),
		TestArtifactDir:   GetEnvOrDefault("TEST_ARTIFACT_DIR", "./test-results"),
		EnableParallel:    parseBool(os.Getenv("ENABLE_PARALLEL_TESTS"), true),
		EnableDetailedLog: parseBool(os.Getenv("ENABLE_DETAILED_LOGS"), false),
		MaxRetries:        parseInt(os.Getenv("TEST_MAX_RETRIES"), 3),

		AWS:   loadAWSConfig(),
		Azure: loadAzureConfig(),
		GCP:   loadGCPConfig(),
		OCI:   loadOCIConfig(),
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate checks if the required configuration is present
func (c *TestConfig) Validate() error {
	var errors []string

	// Validate core Aviatrix configuration
	if c.ControllerIP == "" {
		errors = append(errors, "AVIATRIX_CONTROLLER_IP is required")
	}
	if c.Username == "" {
		errors = append(errors, "AVIATRIX_USERNAME is required")
	}
	if c.Password == "" {
		errors = append(errors, "AVIATRIX_PASSWORD is required")
	}

	// Validate at least one cloud provider is enabled
	if !c.AWS.Enabled && !c.Azure.Enabled && !c.GCP.Enabled && !c.OCI.Enabled {
		errors = append(errors, "at least one cloud provider must be enabled")
	}

	// Validate enabled cloud providers have required credentials
	if c.AWS.Enabled {
		if err := c.AWS.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("AWS: %v", err))
		}
	}

	if c.Azure.Enabled {
		if err := c.Azure.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("Azure: %v", err))
		}
	}

	if c.GCP.Enabled {
		if err := c.GCP.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("GCP: %v", err))
		}
	}

	if c.OCI.Enabled {
		if err := c.OCI.Validate(); err != nil {
			errors = append(errors, fmt.Sprintf("OCI: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// Validate checks AWS configuration
func (a *AWSTestConfig) Validate() error {
	if a.AccessKeyID == "" {
		return fmt.Errorf("AWS_ACCESS_KEY_ID is required")
	}
	if a.SecretAccessKey == "" {
		return fmt.Errorf("AWS_SECRET_ACCESS_KEY is required")
	}
	return nil
}

// Validate checks Azure configuration
func (a *AzureTestConfig) Validate() error {
	if a.ClientID == "" {
		return fmt.Errorf("ARM_CLIENT_ID is required")
	}
	if a.ClientSecret == "" {
		return fmt.Errorf("ARM_CLIENT_SECRET is required")
	}
	if a.SubscriptionID == "" {
		return fmt.Errorf("ARM_SUBSCRIPTION_ID is required")
	}
	if a.TenantID == "" {
		return fmt.Errorf("ARM_TENANT_ID is required")
	}
	return nil
}

// Validate checks GCP configuration
func (g *GCPTestConfig) Validate() error {
	if g.Project == "" {
		return fmt.Errorf("GOOGLE_PROJECT is required")
	}
	if g.ApplicationCredentialsPath == "" {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS is required")
	}
	if _, err := os.Stat(g.ApplicationCredentialsPath); os.IsNotExist(err) {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS file not found: %s", g.ApplicationCredentialsPath)
	}
	return nil
}

// Validate checks OCI configuration
func (o *OCITestConfig) Validate() error {
	if o.UserID == "" {
		return fmt.Errorf("OCI_USER_ID is required")
	}
	if o.TenancyID == "" {
		return fmt.Errorf("OCI_TENANCY_ID is required")
	}
	if o.Fingerprint == "" {
		return fmt.Errorf("OCI_FINGERPRINT is required")
	}
	if o.PrivateKeyPath == "" {
		return fmt.Errorf("OCI_PRIVATE_KEY_PATH is required")
	}
	if _, err := os.Stat(o.PrivateKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("OCI_PRIVATE_KEY_PATH file not found: %s", o.PrivateKeyPath)
	}
	return nil
}

// loadAWSConfig loads AWS configuration from environment
func loadAWSConfig() *AWSTestConfig {
	return &AWSTestConfig{
		Enabled:         os.Getenv("SKIP_ACCOUNT_AWS") != "yes",
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		DefaultRegion:   GetEnvOrDefault("AWS_DEFAULT_REGION", "us-east-1"),
		AccountNumber:   os.Getenv("AWS_ACCOUNT_NUMBER"),
	}
}

// loadAzureConfig loads Azure configuration from environment
func loadAzureConfig() *AzureTestConfig {
	return &AzureTestConfig{
		Enabled:        os.Getenv("SKIP_ACCOUNT_AZURE") != "yes",
		ClientID:       os.Getenv("ARM_CLIENT_ID"),
		ClientSecret:   os.Getenv("ARM_CLIENT_SECRET"),
		SubscriptionID: os.Getenv("ARM_SUBSCRIPTION_ID"),
		TenantID:       os.Getenv("ARM_TENANT_ID"),
		Location:       GetEnvOrDefault("ARM_LOCATION", "East US"),
	}
}

// loadGCPConfig loads GCP configuration from environment
func loadGCPConfig() *GCPTestConfig {
	return &GCPTestConfig{
		Enabled:                    os.Getenv("SKIP_ACCOUNT_GCP") != "yes",
		ApplicationCredentialsPath: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		Project:                    os.Getenv("GOOGLE_PROJECT"),
		Region:                     GetEnvOrDefault("GOOGLE_REGION", "us-central1"),
	}
}

// loadOCIConfig loads OCI configuration from environment
func loadOCIConfig() *OCITestConfig {
	return &OCITestConfig{
		Enabled:        os.Getenv("SKIP_ACCOUNT_OCI") != "yes",
		UserID:         os.Getenv("OCI_USER_ID"),
		TenancyID:      os.Getenv("OCI_TENANCY_ID"),
		Fingerprint:    os.Getenv("OCI_FINGERPRINT"),
		PrivateKeyPath: os.Getenv("OCI_PRIVATE_KEY_PATH"),
		Region:         GetEnvOrDefault("OCI_REGION", "us-ashburn-1"),
	}
}

// parseBool parses a boolean environment variable with default
func parseBool(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return b
}

// parseInt parses an integer environment variable with default
func parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

// GetEnabledProviders returns a list of enabled cloud providers
func (c *TestConfig) GetEnabledProviders() []string {
	providers := []string{}
	if c.AWS.Enabled {
		providers = append(providers, "aws")
	}
	if c.Azure.Enabled {
		providers = append(providers, "azure")
	}
	if c.GCP.Enabled {
		providers = append(providers, "gcp")
	}
	if c.OCI.Enabled {
		providers = append(providers, "oci")
	}
	return providers
}

// IsProviderEnabled checks if a specific provider is enabled
func (c *TestConfig) IsProviderEnabled(provider string) bool {
	switch strings.ToLower(provider) {
	case "aws":
		return c.AWS.Enabled
	case "azure":
		return c.Azure.Enabled
	case "gcp":
		return c.GCP.Enabled
	case "oci":
		return c.OCI.Enabled
	default:
		return false
	}
}

// PrintSummary prints a summary of the test configuration
func (c *TestConfig) PrintSummary() {
	fmt.Println("=== Test Configuration Summary ===")
	fmt.Printf("Controller IP: %s\n", c.ControllerIP)
	fmt.Printf("Username: %s\n", c.Username)
	fmt.Printf("Test Timeout: %s\n", c.TestTimeout)
	fmt.Printf("Test Artifact Dir: %s\n", c.TestArtifactDir)
	fmt.Printf("Parallel Tests: %v\n", c.EnableParallel)
	fmt.Printf("Detailed Logs: %v\n", c.EnableDetailedLog)
	fmt.Println("\n=== Cloud Providers ===")
	for _, provider := range c.GetEnabledProviders() {
		fmt.Printf("  ✓ %s\n", strings.ToUpper(provider))
	}
}
