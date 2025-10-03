package aviatrix

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestingConfig holds common test configuration
type TestingConfig struct {
	ControllerIP string
	Username     string
	Password     string
	SkipAWS      bool
	SkipAzure    bool
	SkipGCP      bool
	SkipOCI      bool
}

// GetTestConfig returns test configuration from environment variables
func GetTestConfig(t *testing.T) *TestingConfig {
	t.Helper()

	return &TestingConfig{
		ControllerIP: os.Getenv("AVIATRIX_CONTROLLER_IP"),
		Username:     os.Getenv("AVIATRIX_USERNAME"),
		Password:     os.Getenv("AVIATRIX_PASSWORD"),
		SkipAWS:      os.Getenv("SKIP_ACCOUNT_AWS") == "yes",
		SkipAzure:    os.Getenv("SKIP_ACCOUNT_AZURE") == "yes",
		SkipGCP:      os.Getenv("SKIP_ACCOUNT_GCP") == "yes",
		SkipOCI:      os.Getenv("SKIP_ACCOUNT_OCI") == "yes",
	}
}

// PreCheckFunc returns a pre-check function for the given cloud provider
func PreCheckFunc(provider string) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		testAccPreCheck(t)

		switch provider {
		case "aws":
			preCheckAWS(t)
		case "azure":
			preCheckAzure(t)
		case "gcp":
			preCheckGCP(t)
		case "oci":
			preCheckOCI(t)
		}
	}
}

// preCheckAWS verifies AWS environment variables
func preCheckAWS(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_ACCOUNT_AWS") == "yes" {
		t.Skip("Skipping AWS tests as SKIP_ACCOUNT_AWS is set")
	}

	required := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"AWS_ACCOUNT_NUMBER",
	}

	for _, envVar := range required {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for AWS acceptance tests", envVar)
		}
	}
}

// preCheckAzure verifies Azure environment variables
func preCheckAzure(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_ACCOUNT_AZURE") == "yes" {
		t.Skip("Skipping Azure tests as SKIP_ACCOUNT_AZURE is set")
	}

	required := []string{
		"ARM_CLIENT_ID",
		"ARM_CLIENT_SECRET",
		"ARM_SUBSCRIPTION_ID",
		"ARM_TENANT_ID",
	}

	for _, envVar := range required {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for Azure acceptance tests", envVar)
		}
	}
}

// preCheckGCP verifies GCP environment variables
func preCheckGCP(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_ACCOUNT_GCP") == "yes" {
		t.Skip("Skipping GCP tests as SKIP_ACCOUNT_GCP is set")
	}

	required := []string{
		"GOOGLE_APPLICATION_CREDENTIALS",
		"GOOGLE_PROJECT",
	}

	for _, envVar := range required {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for GCP acceptance tests", envVar)
		}
	}
}

// preCheckOCI verifies OCI environment variables
func preCheckOCI(t *testing.T) {
	t.Helper()

	if os.Getenv("SKIP_ACCOUNT_OCI") == "yes" {
		t.Skip("Skipping OCI tests as SKIP_ACCOUNT_OCI is set")
	}

	required := []string{
		"OCI_USER_ID",
		"OCI_TENANCY_ID",
		"OCI_FINGERPRINT",
		"OCI_PRIVATE_KEY_PATH",
		"OCI_REGION",
	}

	for _, envVar := range required {
		if v := os.Getenv(envVar); v == "" {
			t.Fatalf("%s must be set for OCI acceptance tests", envVar)
		}
	}
}

// TestStepConfig represents a single test step configuration
type TestStepConfig struct {
	Config string
	Check  resource.TestCheckFunc
}

// NewResourceTestCase creates a new resource.TestCase with common defaults
func NewResourceTestCase(t *testing.T, steps []TestStepConfig) resource.TestCase {
	t.Helper()

	testSteps := make([]resource.TestStep, len(steps))
	for i, step := range steps {
		testSteps[i] = resource.TestStep{
			Config: step.Config,
			Check:  step.Check,
		}
	}

	return resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil, // Set by specific tests
		Steps:        testSteps,
	}
}

// GetTestProvider returns the test provider instance
func GetTestProvider() *schema.Provider {
	return testAccProvider
}

// GetTestProviders returns the test providers map
func GetTestProviders() map[string]*schema.Provider {
	return testAccProviders
}

// CheckResourceAttrWithValueFunc returns a TestCheckFunc that validates an attribute using a custom function
func CheckResourceAttrWithValueFunc(name, key string, f func(string) error) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		value, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("attribute %s not found in resource %s", key, name)
		}

		return f(value)
	}
}

// SkipTestIfEnvSet skips the test if the given environment variable is set to "yes"
func SkipTestIfEnvSet(t *testing.T, envVar string) {
	t.Helper()
	if os.Getenv(envVar) == "yes" {
		t.Skipf("Skipping test as %s is set", envVar)
	}
}

// RequireEnvVars fails the test if any required environment variables are not set
func RequireEnvVars(t *testing.T, vars ...string) {
	t.Helper()
	var missing []string

	for _, v := range vars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		t.Fatalf("Required environment variables not set: %v", missing)
	}
}

// TestAcceptance checks if acceptance tests should run
func TestAcceptance(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1 for acceptance tests")
	}
}
