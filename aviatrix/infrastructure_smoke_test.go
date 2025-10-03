package aviatrix

import (
	"os"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
)

// TestInfrastructureSmokeTest validates the basic test infrastructure setup
func TestInfrastructureSmokeTest(t *testing.T) {
	if !IsAcceptanceTest() {
		t.Skip("Skipping smoke test - not in acceptance test mode (TF_ACC not set)")
	}

	t.Run("EnvironmentVariables", testEnvironmentVariables)
	t.Run("ProviderConfiguration", testProviderConfiguration)
	t.Run("CloudProviderCredentials", testCloudProviderCredentials)
	t.Run("TestArtifactDirectories", testTestArtifactDirectories)
	t.Run("DockerEnvironment", testDockerEnvironment)
}

// testEnvironmentVariables validates required environment variables are set
func testEnvironmentVariables(t *testing.T) {
	required := []string{
		"TF_ACC",
		"AVIATRIX_CONTROLLER_IP",
		"AVIATRIX_USERNAME",
		"AVIATRIX_PASSWORD",
	}

	for _, envVar := range required {
		if os.Getenv(envVar) == "" {
			t.Errorf("Required environment variable not set: %s", envVar)
		} else {
			t.Logf("✓ %s is set", envVar)
		}
	}

	// Check optional variables
	optional := []string{
		"TEST_ARTIFACT_DIR",
		"GO_TEST_TIMEOUT",
		"TEST_RESOURCE_PREFIX",
	}

	for _, envVar := range optional {
		if os.Getenv(envVar) != "" {
			t.Logf("✓ Optional %s is set: %s", envVar, os.Getenv(envVar))
		} else {
			t.Logf("⚠ Optional %s not set (using default)", envVar)
		}
	}
}

// testProviderConfiguration validates the provider can be initialized
func testProviderConfiguration(t *testing.T) {
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider initialization returned nil")
	}

	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("Provider validation failed: %v", err)
	}

	t.Log("✓ Provider initialized and validated successfully")

	// Test provider schema
	schema := provider.Schema
	requiredFields := []string{
		"controller_ip",
		"username",
		"password",
	}

	for _, field := range requiredFields {
		if _, ok := schema[field]; !ok {
			t.Errorf("Provider schema missing required field: %s", field)
		} else {
			t.Logf("✓ Provider schema contains %s", field)
		}
	}
}

// testCloudProviderCredentials validates cloud provider credentials
func testCloudProviderCredentials(t *testing.T) {
	providers := GetCloudProviderConfigs()

	enabledCount := 0
	for name, config := range providers {
		if config.Enabled {
			enabledCount++
			t.Logf("✓ %s provider is enabled (region: %s)", name, config.Region)

			// Validate provider-specific credentials
			switch name {
			case "aws":
				validateAWSCredentials(t)
			case "azure":
				validateAzureCredentials(t)
			case "gcp":
				validateGCPCredentials(t)
			case "oci":
				validateOCICredentials(t)
			}
		} else {
			t.Logf("⊗ %s provider is disabled", name)
		}
	}

	if enabledCount == 0 {
		t.Log("⚠  No cloud providers are enabled - consider enabling at least one provider for integration testing")
	} else {
		t.Logf("✓ %d cloud provider(s) enabled for testing", enabledCount)
	}
}

func validateAWSCredentials(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_AWS") == "yes" {
		return
	}

	required := []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_ACCOUNT_NUMBER"}
	for _, v := range required {
		if os.Getenv(v) == "" {
			t.Errorf("AWS enabled but %s not set", v)
		} else {
			t.Logf("  ✓ %s is set", v)
		}
	}

	region := GetEnvOrDefault("AWS_DEFAULT_REGION", "us-east-1")
	t.Logf("  ✓ Using AWS region: %s", region)
}

func validateAzureCredentials(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_AZURE") == "yes" {
		return
	}

	required := []string{
		"ARM_CLIENT_ID",
		"ARM_CLIENT_SECRET",
		"ARM_SUBSCRIPTION_ID",
		"ARM_TENANT_ID",
	}
	for _, v := range required {
		if os.Getenv(v) == "" {
			t.Errorf("Azure enabled but %s not set", v)
		} else {
			t.Logf("  ✓ %s is set", v)
		}
	}
}

func validateGCPCredentials(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_GCP") == "yes" {
		return
	}

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		t.Error("GCP enabled but GOOGLE_APPLICATION_CREDENTIALS not set")
		return
	}

	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		t.Errorf("GCP credentials file not found: %s", credsPath)
	} else {
		t.Logf("  ✓ GCP credentials file exists: %s", credsPath)
	}

	if os.Getenv("GOOGLE_PROJECT") == "" {
		t.Error("GCP enabled but GOOGLE_PROJECT not set")
	} else {
		t.Logf("  ✓ GOOGLE_PROJECT is set")
	}
}

func validateOCICredentials(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_OCI") == "yes" {
		return
	}

	required := []string{
		"OCI_USER_ID",
		"OCI_TENANCY_ID",
		"OCI_FINGERPRINT",
		"OCI_REGION",
	}
	for _, v := range required {
		if os.Getenv(v) == "" {
			t.Errorf("OCI enabled but %s not set", v)
		} else {
			t.Logf("  ✓ %s is set", v)
		}
	}

	keyPath := os.Getenv("OCI_PRIVATE_KEY_PATH")
	if keyPath == "" {
		t.Error("OCI enabled but OCI_PRIVATE_KEY_PATH not set")
		return
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("OCI private key file not found: %s", keyPath)
	} else {
		t.Logf("  ✓ OCI private key file exists: %s", keyPath)
	}
}

// testTestArtifactDirectories validates test artifact directories exist or can be created
func testTestArtifactDirectories(t *testing.T) {
	artifactDir := TestArtifactDir()
	t.Logf("Test artifact directory: %s", artifactDir)

	// Check if directory exists
	if info, err := os.Stat(artifactDir); err == nil {
		if !info.IsDir() {
			t.Errorf("%s exists but is not a directory", artifactDir)
		} else {
			t.Logf("✓ Test artifact directory exists")
		}
	} else if os.IsNotExist(err) {
		// Try to create it
		if err := os.MkdirAll(artifactDir, 0755); err != nil {
			t.Errorf("Failed to create test artifact directory: %v", err)
		} else {
			t.Logf("✓ Created test artifact directory")
		}
	} else {
		t.Errorf("Error checking test artifact directory: %v", err)
	}

	// Check subdirectories
	subdirs := []string{"logs", "coverage", "reports"}
	for _, subdir := range subdirs {
		path := artifactDir + "/" + subdir
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Errorf("Failed to create %s directory: %v", subdir, err)
		} else {
			t.Logf("✓ %s directory ready", subdir)
		}
	}
}

// testDockerEnvironment validates Docker-specific environment if running in container
func testDockerEnvironment(t *testing.T) {
	// Check if running in Docker
	if _, err := os.Stat("/.dockerenv"); err == nil {
		t.Log("✓ Running in Docker container")

		// Validate Docker-specific configurations
		if os.Getenv("CI") == "true" {
			t.Log("✓ Running in CI environment")
		}

		// Check for volume mounts
		if _, err := os.Stat("/app"); err == nil {
			t.Log("✓ /app directory exists (Docker mount)")
		}
	} else {
		t.Log("⊗ Not running in Docker container (local environment)")
	}
}

// TestProviderVersionValidation tests provider version validation
func TestProviderVersionValidation(t *testing.T) {
	if !IsAcceptanceTest() {
		t.Skip("Skipping provider version validation test - not in acceptance test mode")
	}

	testAccPreCheck(t)

	// This tests that we can create a provider without version validation
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider initialization failed")
	}

	// Test version validation provider
	versionProvider := testAccProviderVersionValidation
	if versionProvider == nil {
		t.Fatal("Version validation provider initialization failed")
	}

	t.Log("✓ Provider version validation configuration successful")
}

// TestClientConnection validates that we can establish a connection to the Aviatrix controller
func TestClientConnection(t *testing.T) {
	if !IsAcceptanceTest() {
		t.Skip("Skipping client connection test - not in acceptance test mode")
	}

	testAccPreCheck(t)

	// Create a client
	client, err := goaviatrix.NewClient(
		os.Getenv("AVIATRIX_USERNAME"),
		os.Getenv("AVIATRIX_PASSWORD"),
		os.Getenv("AVIATRIX_CONTROLLER_IP"),
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("Failed to create Aviatrix client: %v", err)
	}

	if client == nil {
		t.Fatal("Client is nil after creation")
	}

	t.Logf("✓ Successfully created Aviatrix client")
	t.Logf("  Controller IP: %s", os.Getenv("AVIATRIX_CONTROLLER_IP"))
	t.Logf("  Username: %s", os.Getenv("AVIATRIX_USERNAME"))

	// Note: We don't test actual login here to avoid affecting test quotas
	// The actual connection will be tested by integration tests
}

// TestTestHelpers validates the test helper functions work correctly
func TestTestHelpers(t *testing.T) {
	t.Run("GetEnvOrDefault", func(t *testing.T) {
		// Test with existing env var
		os.Setenv("TEST_HELPER_VAR", "test_value")
		result := GetEnvOrDefault("TEST_HELPER_VAR", "default")
		if result != "test_value" {
			t.Errorf("Expected 'test_value', got '%s'", result)
		}

		// Test with non-existing env var
		result = GetEnvOrDefault("NON_EXISTENT_VAR", "default")
		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}

		os.Unsetenv("TEST_HELPER_VAR")
		t.Log("✓ GetEnvOrDefault works correctly")
	})

	t.Run("IsCloudProviderEnabled", func(t *testing.T) {
		// Test AWS
		os.Setenv("SKIP_ACCOUNT_AWS", "yes")
		if IsCloudProviderEnabled("AWS") {
			t.Error("Expected AWS to be disabled")
		}
		os.Unsetenv("SKIP_ACCOUNT_AWS")

		if !IsCloudProviderEnabled("AWS") {
			t.Error("Expected AWS to be enabled")
		}

		t.Log("✓ IsCloudProviderEnabled works correctly")
	})

	t.Run("GetTestTimeout", func(t *testing.T) {
		timeout := GetTestTimeout()
		if timeout <= 0 {
			t.Error("Test timeout should be positive")
		}
		t.Logf("✓ Test timeout: %v", timeout)
	})

	t.Run("RandomTestName", func(t *testing.T) {
		name1 := RandomTestName("test")
		name2 := RandomTestName("test")
		if name1 == name2 {
			t.Error("Random test names should be unique")
		}
		t.Logf("✓ Generated unique test names: %s, %s", name1, name2)
	})
}

// TestTerraformPluginSDKVersion validates we're using the correct SDK version
func TestTerraformPluginSDKVersion(t *testing.T) {
	// This is a compile-time check - if we can import and use the SDK, we're good
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider is nil")
	}

	// Check that we have the v2 SDK features
	if provider.Schema == nil {
		t.Fatal("Provider schema is nil - SDK v2 should have schema")
	}

	if provider.ResourcesMap == nil {
		t.Fatal("Provider resources map is nil - SDK v2 should have resources")
	}

	if provider.DataSourcesMap == nil {
		t.Fatal("Provider data sources map is nil - SDK v2 should have data sources")
	}

	t.Log("✓ Terraform Plugin SDK v2 is properly configured")
	t.Logf("  Resources: %d", len(provider.ResourcesMap))
	t.Logf("  Data Sources: %d", len(provider.DataSourcesMap))
}
