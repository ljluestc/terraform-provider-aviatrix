package aviatrix

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TestConfig holds global test configuration
type TestConfig struct {
	// Timeouts
	DefaultTimeout time.Duration
	CreateTimeout  time.Duration
	UpdateTimeout  time.Duration
	DeleteTimeout  time.Duration
	RetryTimeout   time.Duration

	// Retry settings
	MaxRetries int
	RetryDelay time.Duration

	// Test artifacts
	ArtifactDir   string
	LogDir        string
	ScreenshotDir string

	// Feature flags
	EnableParallel     bool
	EnableDetailedLogs bool
	EnableScreenshots  bool

	// Test data
	TestDataDir string
}

// DefaultTestConfig returns the default test configuration
func DefaultTestConfig() *TestConfig {
	artifactDir := getEnvWithDefault("TEST_ARTIFACT_DIR", "./test-results")

	return &TestConfig{
		// Timeouts
		DefaultTimeout: 30 * time.Minute,
		CreateTimeout:  20 * time.Minute,
		UpdateTimeout:  15 * time.Minute,
		DeleteTimeout:  20 * time.Minute,
		RetryTimeout:   5 * time.Minute,

		// Retry settings
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,

		// Test artifacts
		ArtifactDir:   artifactDir,
		LogDir:        filepath.Join(artifactDir, "logs"),
		ScreenshotDir: filepath.Join(artifactDir, "screenshots"),

		// Feature flags
		EnableParallel:     getEnvBool("ENABLE_PARALLEL_TESTS", true),
		EnableDetailedLogs: getEnvBool("ENABLE_DETAILED_LOGS", false),
		EnableScreenshots:  getEnvBool("ENABLE_SCREENSHOTS", false),

		// Test data
		TestDataDir: getEnvWithDefault("TEST_DATA_DIR", "./test-data"),
	}
}

// EnsureDirectories creates all required test directories
func (tc *TestConfig) EnsureDirectories() error {
	dirs := []string{
		tc.ArtifactDir,
		tc.LogDir,
		tc.ScreenshotDir,
		tc.TestDataDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// LogPath returns the path for a test log file
func (tc *TestConfig) LogPath(testName string) string {
	return filepath.Join(tc.LogDir, fmt.Sprintf("%s_%s.log", testName, time.Now().Format("20060102_150405")))
}

// ScreenshotPath returns the path for a test screenshot
func (tc *TestConfig) ScreenshotPath(testName string, step int) string {
	return filepath.Join(tc.ScreenshotDir, fmt.Sprintf("%s_step%d_%s.png", testName, step, time.Now().Format("20060102_150405")))
}

// getEnvBool returns a boolean environment variable value or default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// ResourceNamingConfig provides standardized resource naming for tests
type ResourceNamingConfig struct {
	Prefix    string
	Suffix    string
	Timestamp string
}

// NewResourceNamingConfig creates a new resource naming configuration
func NewResourceNamingConfig() *ResourceNamingConfig {
	return &ResourceNamingConfig{
		Prefix:    getEnvWithDefault("TEST_RESOURCE_PREFIX", "tf-test"),
		Suffix:    getEnvWithDefault("TEST_RESOURCE_SUFFIX", ""),
		Timestamp: time.Now().Format("20060102-150405"),
	}
}

// GenerateName creates a standardized test resource name
func (rnc *ResourceNamingConfig) GenerateName(resourceType string) string {
	if rnc.Suffix != "" {
		return fmt.Sprintf("%s-%s-%s-%s", rnc.Prefix, resourceType, rnc.Timestamp, rnc.Suffix)
	}
	return fmt.Sprintf("%s-%s-%s", rnc.Prefix, resourceType, rnc.Timestamp)
}

// CloudProviderTestConfig holds cloud-specific test configurations
type CloudProviderTestConfig struct {
	// AWS
	AWSTestRegion       string
	AWSTestVPCCIDR      string
	AWSTestSubnetCIDR   string
	AWSTestInstanceType string

	// Azure
	AzureTestRegion     string
	AzureTestVNetCIDR   string
	AzureTestSubnetCIDR string
	AzureTestVMSize     string

	// GCP
	GCPTestRegion      string
	GCPTestVPCCIDR     string
	GCPTestSubnetCIDR  string
	GCPTestMachineType string

	// OCI
	OCITestRegion        string
	OCITestVCNCIDR       string
	OCITestSubnetCIDR    string
	OCITestInstanceShape string
}

// DefaultCloudProviderTestConfig returns default cloud provider test configuration
func DefaultCloudProviderTestConfig() *CloudProviderTestConfig {
	return &CloudProviderTestConfig{
		// AWS defaults
		AWSTestRegion:       getEnvWithDefault("AWS_TEST_REGION", "us-east-1"),
		AWSTestVPCCIDR:      getEnvWithDefault("AWS_TEST_VPC_CIDR", "10.0.0.0/16"),
		AWSTestSubnetCIDR:   getEnvWithDefault("AWS_TEST_SUBNET_CIDR", "10.0.1.0/24"),
		AWSTestInstanceType: getEnvWithDefault("AWS_TEST_INSTANCE_TYPE", "t3.medium"),

		// Azure defaults
		AzureTestRegion:     getEnvWithDefault("AZURE_TEST_REGION", "East US"),
		AzureTestVNetCIDR:   getEnvWithDefault("AZURE_TEST_VNET_CIDR", "10.1.0.0/16"),
		AzureTestSubnetCIDR: getEnvWithDefault("AZURE_TEST_SUBNET_CIDR", "10.1.1.0/24"),
		AzureTestVMSize:     getEnvWithDefault("AZURE_TEST_VM_SIZE", "Standard_B2s"),

		// GCP defaults
		GCPTestRegion:      getEnvWithDefault("GCP_TEST_REGION", "us-central1"),
		GCPTestVPCCIDR:     getEnvWithDefault("GCP_TEST_VPC_CIDR", "10.2.0.0/16"),
		GCPTestSubnetCIDR:  getEnvWithDefault("GCP_TEST_SUBNET_CIDR", "10.2.1.0/24"),
		GCPTestMachineType: getEnvWithDefault("GCP_TEST_MACHINE_TYPE", "n1-standard-2"),

		// OCI defaults
		OCITestRegion:        getEnvWithDefault("OCI_TEST_REGION", "us-ashburn-1"),
		OCITestVCNCIDR:       getEnvWithDefault("OCI_TEST_VCN_CIDR", "10.3.0.0/16"),
		OCITestSubnetCIDR:    getEnvWithDefault("OCI_TEST_SUBNET_CIDR", "10.3.1.0/24"),
		OCITestInstanceShape: getEnvWithDefault("OCI_TEST_INSTANCE_SHAPE", "VM.Standard2.1"),
	}
}
