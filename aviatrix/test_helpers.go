package aviatrix

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestHelper provides common test utilities and helper functions
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper instance
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// SkipIfEnvSet skips the test if the given environment variable is set
func (h *TestHelper) SkipIfEnvSet(envVar string) {
	if os.Getenv(envVar) != "" {
		h.t.Skipf("Skipping test because %s is set", envVar)
	}
}

// RequireEnvVar fails the test if the required environment variable is not set
func (h *TestHelper) RequireEnvVar(envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		h.t.Fatalf("Required environment variable %s is not set", envVar)
	}
	return value
}

// GetEnvOrDefault returns the environment variable value or a default if not set
func GetEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvOrSkip returns the environment variable value or skips the test
func GetEnvOrSkip(t *testing.T, envVar string) string {
	value := os.Getenv(envVar)
	if value == "" {
		t.Skipf("Skipping test because %s is not set", envVar)
	}
	return value
}

// CloudProviderConfig contains cloud provider specific test configuration
type CloudProviderConfig struct {
	Provider string
	Enabled  bool
	Region   string
}

// GetCloudProviderConfigs returns configured cloud providers for testing
func GetCloudProviderConfigs() map[string]*CloudProviderConfig {
	return map[string]*CloudProviderConfig{
		"aws": {
			Provider: "aws",
			Enabled:  os.Getenv("SKIP_ACCOUNT_AWS") != "yes",
			Region:   GetEnvOrDefault("AWS_DEFAULT_REGION", "us-east-1"),
		},
		"azure": {
			Provider: "azure",
			Enabled:  os.Getenv("SKIP_ACCOUNT_AZURE") != "yes",
			Region:   GetEnvOrDefault("ARM_LOCATION", "East US"),
		},
		"gcp": {
			Provider: "gcp",
			Enabled:  os.Getenv("SKIP_ACCOUNT_GCP") != "yes",
			Region:   GetEnvOrDefault("GOOGLE_REGION", "us-central1"),
		},
		"oci": {
			Provider: "oci",
			Enabled:  os.Getenv("SKIP_ACCOUNT_OCI") != "yes",
			Region:   GetEnvOrDefault("OCI_REGION", "us-ashburn-1"),
		},
	}
}

// IsCloudProviderEnabled checks if a specific cloud provider is enabled for testing
func IsCloudProviderEnabled(provider string) bool {
	skipVar := fmt.Sprintf("SKIP_ACCOUNT_%s", provider)
	return os.Getenv(skipVar) != "yes"
}

// SkipUnlessCloudProvider skips test unless the specified cloud provider is enabled
func SkipUnlessCloudProvider(t *testing.T, provider string) {
	skipVar := fmt.Sprintf("SKIP_ACCOUNT_%s", provider)
	if os.Getenv(skipVar) == "yes" {
		t.Skipf("Skipping test because %s provider is disabled", provider)
	}
}

// WaitForResourceState waits for a resource to reach the expected state
func WaitForResourceState(
	resourceName string,
	pending []string,
	target []string,
	timeout time.Duration,
	stateFunc resource.StateRefreshFunc,
) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    stateFunc,
		Timeout:    timeout,
		MinTimeout: 3 * time.Second,
		Delay:      5 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return err
}

// CheckResourceAttrWithFunc checks a resource attribute using a custom validation function
func CheckResourceAttrWithFunc(name, key string, validationFunc func(string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		value, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("attribute not found: %s", key)
		}

		return validationFunc(value)
	}
}

// ComposeTestCheckFuncWithRetry retries the test check function with exponential backoff
func ComposeTestCheckFuncWithRetry(f resource.TestCheckFunc, maxRetries int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var lastErr error
		for i := 0; i < maxRetries; i++ {
			err := f(s)
			if err == nil {
				return nil
			}
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
		}
		return fmt.Errorf("check failed after %d retries: %w", maxRetries, lastErr)
	}
}

// TestArtifactDir returns the directory for test artifacts
func TestArtifactDir() string {
	return GetEnvOrDefault("TEST_ARTIFACT_DIR", "./test-results")
}

// IsAcceptanceTest returns true if running acceptance tests
func IsAcceptanceTest() bool {
	return os.Getenv("TF_ACC") != ""
}

// GetTestTimeout returns the configured test timeout
func GetTestTimeout() time.Duration {
	timeout := GetEnvOrDefault("GO_TEST_TIMEOUT", "30m")
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return 30 * time.Minute
	}
	return duration
}

// ImportStateIDFunc is a helper for generating import IDs
type ImportStateIDFunc func(*terraform.State) (string, error)

// LogTestProgress logs test progress to both stdout and artifact file
func LogTestProgress(t *testing.T, message string, args ...interface{}) {
	msg := fmt.Sprintf(message, args...)
	t.Logf("[TEST] %s", msg)

	// Also write to artifact log if enabled
	if logFile := os.Getenv("TEST_LOG_FILE"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			fmt.Fprintf(f, "[%s] %s: %s\n", timestamp, t.Name(), msg)
		}
	}
}

// RandomTestName generates a random test resource name
func RandomTestName(prefix string) string {
	return resource.PrefixedUniqueId(prefix)
}

// TestPreCheckFuncs aggregates multiple precheck functions
func TestPreCheckFuncs(funcs ...func(*testing.T)) func(*testing.T) {
	return func(t *testing.T) {
		for _, f := range funcs {
			f(t)
		}
	}
}

// PreCheckAWS checks AWS-specific prerequisites
func PreCheckAWS(t *testing.T) {
	SkipUnlessCloudProvider(t, "AWS")
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Fatal("AWS_ACCESS_KEY_ID must be set for AWS tests")
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		t.Fatal("AWS_SECRET_ACCESS_KEY must be set for AWS tests")
	}
}

// PreCheckAzure checks Azure-specific prerequisites
func PreCheckAzure(t *testing.T) {
	SkipUnlessCloudProvider(t, "AZURE")
	if os.Getenv("ARM_CLIENT_ID") == "" {
		t.Fatal("ARM_CLIENT_ID must be set for Azure tests")
	}
	if os.Getenv("ARM_CLIENT_SECRET") == "" {
		t.Fatal("ARM_CLIENT_SECRET must be set for Azure tests")
	}
	if os.Getenv("ARM_SUBSCRIPTION_ID") == "" {
		t.Fatal("ARM_SUBSCRIPTION_ID must be set for Azure tests")
	}
	if os.Getenv("ARM_TENANT_ID") == "" {
		t.Fatal("ARM_TENANT_ID must be set for Azure tests")
	}
}

// PreCheckGCP checks GCP-specific prerequisites
func PreCheckGCP(t *testing.T) {
	SkipUnlessCloudProvider(t, "GCP")
	if os.Getenv("GOOGLE_PROJECT") == "" {
		t.Fatal("GOOGLE_PROJECT must be set for GCP tests")
	}
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		t.Fatal("GOOGLE_APPLICATION_CREDENTIALS must be set for GCP tests")
	}
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		t.Fatalf("GCP credentials file not found: %s", credsPath)
	}
}

// PreCheckOCI checks OCI-specific prerequisites
func PreCheckOCI(t *testing.T) {
	SkipUnlessCloudProvider(t, "OCI")
	if os.Getenv("OCI_USER_ID") == "" {
		t.Fatal("OCI_USER_ID must be set for OCI tests")
	}
	if os.Getenv("OCI_TENANCY_ID") == "" {
		t.Fatal("OCI_TENANCY_ID must be set for OCI tests")
	}
	if os.Getenv("OCI_FINGERPRINT") == "" {
		t.Fatal("OCI_FINGERPRINT must be set for OCI tests")
	}
	keyPath := os.Getenv("OCI_PRIVATE_KEY_PATH")
	if keyPath == "" {
		t.Fatal("OCI_PRIVATE_KEY_PATH must be set for OCI tests")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatalf("OCI private key file not found: %s", keyPath)
	}
}
