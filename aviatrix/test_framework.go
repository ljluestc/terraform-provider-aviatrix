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
