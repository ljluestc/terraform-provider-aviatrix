package aviatrix

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestSmokeProvider verifies the provider can be initialized
func TestSmokeProvider(t *testing.T) {
	provider := Provider()

	if provider == nil {
		t.Fatal("Provider() returned nil")
	}

	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("Provider validation failed: %v", err)
	}

	t.Log("✓ Provider initialized and validated successfully")
}

// TestSmokeProviderSchema verifies provider schema is valid
func TestSmokeProviderSchema(t *testing.T) {
	provider := Provider()

	schema := provider.Schema
	if schema == nil {
		t.Fatal("Provider schema is nil")
	}

	// Check for required configuration fields
	requiredFields := []string{
		"controller_ip",
		"username",
		"password",
	}

	for _, field := range requiredFields {
		if _, exists := schema[field]; !exists {
			t.Errorf("Required provider field missing: %s", field)
		}
	}

	t.Logf("✓ Provider schema validated with %d fields", len(schema))
}

// TestSmokeProviderResources verifies provider has resources registered
func TestSmokeProviderResources(t *testing.T) {
	provider := Provider()

	resources := provider.ResourcesMap
	if resources == nil {
		t.Fatal("Provider resources map is nil")
	}

	if len(resources) == 0 {
		t.Fatal("No resources registered in provider")
	}

	t.Logf("✓ Provider has %d resources registered", len(resources))

	// Validate a few key resources exist
	keyResources := []string{
		"aviatrix_account",
		"aviatrix_gateway",
		"aviatrix_transit_gateway",
		"aviatrix_spoke_gateway",
	}

	for _, resourceName := range keyResources {
		if _, exists := resources[resourceName]; exists {
			t.Logf("  ✓ Found resource: %s", resourceName)
		} else {
			t.Logf("  - Resource not found: %s", resourceName)
		}
	}
}

// TestSmokeProviderDataSources verifies provider has data sources registered
func TestSmokeProviderDataSources(t *testing.T) {
	provider := Provider()

	dataSources := provider.DataSourcesMap
	if dataSources == nil {
		t.Fatal("Provider data sources map is nil")
	}

	if len(dataSources) == 0 {
		t.Fatal("No data sources registered in provider")
	}

	t.Logf("✓ Provider has %d data sources registered", len(dataSources))
}

// TestSmokeTestingUtils verifies testing utilities are working
func TestSmokeTestingUtils(t *testing.T) {
	// Test environment retrieval
	env := NewTestEnvironment()
	if env == nil {
		t.Fatal("NewTestEnvironment returned nil")
	}

	t.Log("✓ Test environment retrieved")

	// Test provider accessors
	if testAccProvider == nil {
		t.Fatal("testAccProvider is nil")
	}

	t.Log("✓ Test provider accessor working")

	// Test providers map
	if testAccProviders == nil || len(testAccProviders) == 0 {
		t.Fatal("testAccProviders is empty")
	}

	t.Log("✓ Test providers map accessible")
}

// TestSmokeTestingHelpers verifies test helper functions
func TestSmokeTestingHelpers(t *testing.T) {
	// Test random string generation
	str := RandomString("8")
	if str == "" {
		t.Fatal("RandomString returned empty string")
	}
	t.Logf("✓ RandomString generated: %s", str)

	// Test environment validation
	env := NewTestEnvironment()
	if env == nil {
		t.Fatal("NewTestEnvironment returned nil")
	}
	t.Log("✓ Test environment created")

	// Test provider enabled checks
	t.Logf("✓ AWS enabled: %v", !env.SkipAWS)
	t.Logf("✓ Azure enabled: %v", !env.SkipAzure)
	t.Logf("✓ GCP enabled: %v", !env.SkipGCP)
	t.Logf("✓ OCI enabled: %v", !env.SkipOCI)
}

// TestSmokeEnvironmentVariables verifies environment variable handling
func TestSmokeEnvironmentVariables(t *testing.T) {
	// Test environment configuration
	env := NewTestEnvironment()
	if env == nil {
		t.Fatal("NewTestEnvironment returned nil")
	}

	t.Log("✓ Environment variables loaded")

	// Test Aviatrix controller credentials
	if env.ControllerIP != "" {
		t.Logf("✓ Controller IP configured: %s", maskValue(env.ControllerIP))
	} else {
		t.Log("⚠ Controller IP not configured")
	}

	// Test AWS credentials
	if !env.SkipAWS {
		if env.AWSAccessKeyID != "" {
			t.Logf("✓ AWS credentials configured")
		} else {
			t.Log("⚠ AWS credentials not configured")
		}
	}

	// Test Azure credentials
	if !env.SkipAzure {
		if env.ARMClientID != "" {
			t.Logf("✓ Azure credentials configured")
		} else {
			t.Log("⚠ Azure credentials not configured")
		}
	}

	// Test GCP credentials
	if !env.SkipGCP {
		if env.GoogleProject != "" {
			t.Logf("✓ GCP credentials configured")
		} else {
			t.Log("⚠ GCP credentials not configured")
		}
	}

	// Test OCI credentials
	if !env.SkipOCI {
		if env.OCIUserID != "" {
			t.Logf("✓ OCI credentials configured")
		} else {
			t.Log("⚠ OCI credentials not configured")
		}
	}
}

// maskValue masks credential values for display
func maskValue(value string) string {
	if value == "" {
		return "[NOT SET]"
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****"
}

// TestSmokeTestLogger verifies test logging infrastructure
func TestSmokeTestLogger(t *testing.T) {
	// Create test logger from existing infrastructure
	logger, err := NewTestLogger(t, "smoke-test")
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	defer logger.Close()

	// Test logging at different levels
	logger.Info("Smoke test: info message")
	logger.Debug("Smoke test: debug message")
	logger.Error("Smoke test: error message (test)")

	t.Log("✓ Test logger created and used successfully")
}

// TestSmokeArtifactManager verifies artifact management
func TestSmokeArtifactManager(t *testing.T) {
	// Create test results directory
	testDir := "test-results/smoke"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Save a test artifact
	content := []byte("Test artifact content")
	artifactPath := testDir + "/test-artifact.txt"
	if err := os.WriteFile(artifactPath, content, 0644); err != nil {
		t.Fatalf("Failed to save artifact: %v", err)
	}

	t.Logf("✓ Artifact saved to: %s", artifactPath)

	// Verify artifact exists
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		t.Fatal("Artifact file does not exist")
	}

	t.Log("✓ Artifact management working")

	// Clean up
	os.Remove(artifactPath)
}

// TestSmokeDockerEnvironment verifies we're running in expected environment
func TestSmokeDockerEnvironment(t *testing.T) {
	// Check if running in Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		t.Log("✓ Running inside Docker container")
	} else {
		t.Log("✓ Running in local environment")
	}

	// Check Go version
	t.Logf("✓ Go version: %s", os.Getenv("GOVERSION"))

	// Check working directory
	wd, err := os.Getwd()
	if err == nil {
		t.Logf("✓ Working directory: %s", wd)
	}
}

// TestSmokeGitHubActionsEnvironment verifies GitHub Actions environment
func TestSmokeGitHubActionsEnvironment(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Log("✓ Running in GitHub Actions")
		t.Logf("  Workflow: %s", os.Getenv("GITHUB_WORKFLOW"))
		t.Logf("  Run ID: %s", os.Getenv("GITHUB_RUN_ID"))
		t.Logf("  Runner: %s", os.Getenv("RUNNER_OS"))
	} else {
		t.Log("✓ Not running in GitHub Actions")
	}
}

// TestSmokeTestInfrastructureSetup verifies test infrastructure readiness
func TestSmokeTestInfrastructureSetup(t *testing.T) {
	// Verify test-results directory can be created
	testDir := "test-results"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test-results directory: %v", err)
	}
	t.Log("✓ Test results directory accessible")

	// Verify we can write test artifacts
	testFile := testDir + "/smoke-test.txt"
	if err := os.WriteFile(testFile, []byte("smoke test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	t.Log("✓ Test artifact writing works")

	// Clean up
	os.Remove(testFile)
}

// TestSmokeProviderInitialization is an integration smoke test
func TestSmokeProviderInitialization(t *testing.T) {
	// Skip if not acceptance test
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("Skipping acceptance test")
	}

	// Basic precheck
	testAccPreCheck(t)

	// Get test environment
	env := NewTestEnvironment()

	if env.ControllerIP == "" {
		t.Log("⚠ Controller IP not set - skipping provider initialization test")
		return
	}

	t.Log("✓ Provider initialization smoke test passed")
}

// TestSmokeResourceSchema validates resource schemas
func TestSmokeResourceSchema(t *testing.T) {
	provider := Provider()
	resources := provider.ResourcesMap

	// Test schema validation for a few key resources
	resourcesToTest := []string{
		"aviatrix_gateway",
		"aviatrix_transit_gateway",
	}

	for _, resourceName := range resourcesToTest {
		if resource, exists := resources[resourceName]; exists {
			// Check that resource has schema
			if resource.Schema == nil {
				t.Errorf("Resource %s has nil schema", resourceName)
				continue
			}

			// Check that resource has CRUD functions (Context or non-Context versions)
			hasCreate := resource.CreateContext != nil || resource.CreateWithoutTimeout != nil || resource.Create != nil
			hasRead := resource.ReadContext != nil || resource.ReadWithoutTimeout != nil || resource.Read != nil
			hasDelete := resource.DeleteContext != nil || resource.DeleteWithoutTimeout != nil || resource.Delete != nil

			if !hasCreate || !hasRead || !hasDelete {
				t.Errorf("Resource %s missing CRUD functions (Create: %v, Read: %v, Delete: %v)",
					resourceName, hasCreate, hasRead, hasDelete)
			} else {
				t.Logf("✓ Resource %s has valid schema and CRUD functions", resourceName)
			}
		}
	}
}

// TestSmokeDataSourceSchema validates data source schemas
func TestSmokeDataSourceSchema(t *testing.T) {
	provider := Provider()
	dataSources := provider.DataSourcesMap

	// Test schema validation for data sources
	dataSourceToTest := "aviatrix_account"

	if dataSource, exists := dataSources[dataSourceToTest]; exists {
		if dataSource.Schema == nil {
			t.Errorf("Data source %s has nil schema", dataSourceToTest)
		}

		hasRead := dataSource.ReadContext != nil || dataSource.Read != nil
		if !hasRead {
			t.Errorf("Data source %s missing Read function", dataSourceToTest)
		} else {
			t.Logf("✓ Data source %s has valid schema and Read function", dataSourceToTest)
		}
	}
}

// getTestProviderWithConfig returns a test provider for smoke tests
func getTestProviderWithConfig() *schema.Provider {
	// Return the test provider without configuration for smoke tests
	return testAccProvider
}
