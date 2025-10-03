package aviatrix

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestEnvironment holds cloud provider configurations for testing
type TestEnvironment struct {
	SkipAWS   bool
	SkipAzure bool
	SkipGCP   bool
	SkipOCI   bool

	// AWS
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	AWSAccountNumber   string

	// Azure
	ARMClientID       string
	ARMClientSecret   string
	ARMSubscriptionID string
	ARMTenantID       string

	// GCP
	GoogleCredentials string
	GoogleProject     string

	// OCI
	OCIUserID      string
	OCITenancyID   string
	OCIFingerprint string
	OCIPrivateKey  string
	OCIRegion      string

	// Aviatrix Controller
	ControllerIP string
	Username     string
	Password     string
}

// NewTestEnvironment creates a test environment from environment variables
func NewTestEnvironment() *TestEnvironment {
	return &TestEnvironment{
		// Skip flags
		SkipAWS:   os.Getenv("SKIP_ACCOUNT_AWS") == "yes",
		SkipAzure: os.Getenv("SKIP_ACCOUNT_AZURE") == "yes",
		SkipGCP:   os.Getenv("SKIP_ACCOUNT_GCP") == "yes",
		SkipOCI:   os.Getenv("SKIP_ACCOUNT_OCI") == "yes",

		// AWS
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          getEnvWithDefault("AWS_DEFAULT_REGION", "us-east-1"),
		AWSAccountNumber:   os.Getenv("AWS_ACCOUNT_NUMBER"),

		// Azure
		ARMClientID:       os.Getenv("ARM_CLIENT_ID"),
		ARMClientSecret:   os.Getenv("ARM_CLIENT_SECRET"),
		ARMSubscriptionID: os.Getenv("ARM_SUBSCRIPTION_ID"),
		ARMTenantID:       os.Getenv("ARM_TENANT_ID"),

		// GCP
		GoogleCredentials: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		GoogleProject:     os.Getenv("GOOGLE_PROJECT"),

		// OCI
		OCIUserID:      os.Getenv("OCI_USER_ID"),
		OCITenancyID:   os.Getenv("OCI_TENANCY_ID"),
		OCIFingerprint: os.Getenv("OCI_FINGERPRINT"),
		OCIPrivateKey:  os.Getenv("OCI_PRIVATE_KEY_PATH"),
		OCIRegion:      os.Getenv("OCI_REGION"),

		// Aviatrix
		ControllerIP: os.Getenv("AVIATRIX_CONTROLLER_IP"),
		Username:     os.Getenv("AVIATRIX_USERNAME"),
		Password:     os.Getenv("AVIATRIX_PASSWORD"),
	}
}

// ValidateAWSCredentials checks if AWS credentials are available
func (te *TestEnvironment) ValidateAWSCredentials(t *testing.T) {
	if te.SkipAWS {
		t.Skip("Skipping AWS tests as SKIP_ACCOUNT_AWS is set")
	}
	if te.AWSAccessKeyID == "" || te.AWSSecretAccessKey == "" {
		t.Skip("AWS credentials not available for testing")
	}
}

// ValidateAzureCredentials checks if Azure credentials are available
func (te *TestEnvironment) ValidateAzureCredentials(t *testing.T) {
	if te.SkipAzure {
		t.Skip("Skipping Azure tests as SKIP_ACCOUNT_AZURE is set")
	}
	if te.ARMClientID == "" || te.ARMClientSecret == "" || te.ARMSubscriptionID == "" || te.ARMTenantID == "" {
		t.Skip("Azure credentials not available for testing")
	}
}

// ValidateGCPCredentials checks if GCP credentials are available
func (te *TestEnvironment) ValidateGCPCredentials(t *testing.T) {
	if te.SkipGCP {
		t.Skip("Skipping GCP tests as SKIP_ACCOUNT_GCP is set")
	}
	if te.GoogleCredentials == "" || te.GoogleProject == "" {
		t.Skip("GCP credentials not available for testing")
	}
}

// ValidateOCICredentials checks if OCI credentials are available
func (te *TestEnvironment) ValidateOCICredentials(t *testing.T) {
	if te.SkipOCI {
		t.Skip("Skipping OCI tests as SKIP_ACCOUNT_OCI is set")
	}
	if te.OCIUserID == "" || te.OCITenancyID == "" || te.OCIFingerprint == "" || te.OCIPrivateKey == "" {
		t.Skip("OCI credentials not available for testing")
	}
}

// ValidateControllerCredentials checks if Aviatrix controller credentials are available
func (te *TestEnvironment) ValidateControllerCredentials(t *testing.T) {
	if te.ControllerIP == "" || te.Username == "" || te.Password == "" {
		t.Fatal("Aviatrix controller credentials must be set for acceptance tests")
	}
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CheckResourceAttrSet is a helper to check if a resource attribute is set
func CheckResourceAttrSet(name, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		if v, ok := rs.Primary.Attributes[key]; !ok || v == "" {
			return fmt.Errorf("%s: attribute '%s' not set", name, key)
		}

		return nil
	}
}

// TestCheckResourceAttrNotEmpty checks that an attribute exists and is not empty
func CheckResourceAttrNotEmpty(name, key string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrSet(name, key)
}

// ComposeTestCheckFunc composes multiple test check functions
func ComposeTestCheckFunc(fs ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(fs...)
}

// ComposeAggregateTestCheckFunc composes multiple test check functions with detailed error reporting
func ComposeAggregateTestCheckFunc(fs ...resource.TestCheckFunc) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(fs...)
}

// PreCheckCloud is a convenience function for cloud-specific pre-checks
type CloudPreCheck func(t *testing.T, env *TestEnvironment)

var (
	// PreCheckAWS validates AWS environment for testing
	PreCheckAWS CloudPreCheck = func(t *testing.T, env *TestEnvironment) {
		env.ValidateControllerCredentials(t)
		env.ValidateAWSCredentials(t)
	}

	// PreCheckAzure validates Azure environment for testing
	PreCheckAzure CloudPreCheck = func(t *testing.T, env *TestEnvironment) {
		env.ValidateControllerCredentials(t)
		env.ValidateAzureCredentials(t)
	}

	// PreCheckGCP validates GCP environment for testing
	PreCheckGCP CloudPreCheck = func(t *testing.T, env *TestEnvironment) {
		env.ValidateControllerCredentials(t)
		env.ValidateGCPCredentials(t)
	}

	// PreCheckOCI validates OCI environment for testing
	PreCheckOCI CloudPreCheck = func(t *testing.T, env *TestEnvironment) {
		env.ValidateControllerCredentials(t)
		env.ValidateOCICredentials(t)
	}
)

// TestCase is a wrapper for resource.TestCase with common configurations
type TestCase struct {
	*resource.TestCase
}

// NewTestCase creates a new test case with default providers
func NewTestCase() *TestCase {
	return &TestCase{
		TestCase: &resource.TestCase{
			Providers: testAccProviders,
		},
	}
}

// WithSteps adds test steps to the test case
func (tc *TestCase) WithSteps(steps ...resource.TestStep) *TestCase {
	tc.TestCase.Steps = steps
	return tc
}

// WithPreCheck adds a pre-check function
func (tc *TestCase) WithPreCheck(f func()) *TestCase {
	tc.TestCase.PreCheck = f
	return tc
}

// WithCloudPreCheck adds a cloud-specific pre-check
func (tc *TestCase) WithCloudPreCheck(t *testing.T, preCheck CloudPreCheck) *TestCase {
	env := NewTestEnvironment()
	tc.TestCase.PreCheck = func() {
		preCheck(t, env)
	}
	return tc
}

// Run executes the test case
func (tc *TestCase) Run(t *testing.T) {
	resource.Test(t, *tc.TestCase)
}

// RandomString generates a random string for test resource names (reuse existing utility if available)
// This is a placeholder - should be replaced with actual implementation
func RandomString(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, os.Getpid())
}

// SkipIfNotAcceptance skips the test if not running acceptance tests
func SkipIfNotAcceptance(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Skipping acceptance test")
	}
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}
