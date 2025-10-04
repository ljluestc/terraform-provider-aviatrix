package aviatrix

import (
	"os"
	"testing"
)

// TestSmokeProvider tests basic provider initialization
func TestSmokeProvider(t *testing.T) {
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider() returned nil")
	}

	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("Provider validation failed: %s", err)
	}
}

// TestSmokeTestHelpers validates test helper functions
func TestSmokeTestHelpers(t *testing.T) {
	// Test GetEnvOrDefault
	os.Setenv("TEST_VAR_123", "test_value")
	defer os.Unsetenv("TEST_VAR_123")

	if got := GetEnvOrDefault("TEST_VAR_123", "default"); got != "test_value" {
		t.Errorf("GetEnvOrDefault() = %v, want %v", got, "test_value")
	}

	if got := GetEnvOrDefault("NON_EXISTENT_VAR", "default"); got != "default" {
		t.Errorf("GetEnvOrDefault() = %v, want %v", got, "default")
	}
}

// TestSmokeCloudProviderConfig tests cloud provider configuration helpers
func TestSmokeCloudProviderConfig(t *testing.T) {
	// Save original values
	origSkipAWS := os.Getenv("SKIP_ACCOUNT_AWS")
	defer os.Setenv("SKIP_ACCOUNT_AWS", origSkipAWS)

	// Test IsCloudProviderEnabled
	os.Setenv("SKIP_ACCOUNT_AWS", "yes")
	if IsCloudProviderEnabled("AWS") {
		t.Error("IsCloudProviderEnabled(AWS) should be false when SKIP_ACCOUNT_AWS=yes")
	}

	os.Setenv("SKIP_ACCOUNT_AWS", "")
	if !IsCloudProviderEnabled("AWS") {
		t.Error("IsCloudProviderEnabled(AWS) should be true when SKIP_ACCOUNT_AWS is not set")
	}

	// Test GetCloudProviderConfigs
	configs := GetCloudProviderConfigs()
	if configs == nil {
		t.Fatal("GetCloudProviderConfigs() returned nil")
	}

	if _, ok := configs["aws"]; !ok {
		t.Error("GetCloudProviderConfigs() missing aws configuration")
	}
}
