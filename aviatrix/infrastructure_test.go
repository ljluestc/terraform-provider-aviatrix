package aviatrix

import (
	"os"
	"testing"
)

// TestInfrastructureSetup validates the test infrastructure is properly configured
func TestInfrastructureSetup(t *testing.T) {
	t.Run("TestEnvironmentCreation", func(t *testing.T) {
		env := NewTestEnvironment()
		if env == nil {
			t.Fatal("Failed to create test environment")
		}
		t.Log("Test environment created successfully")
	})

	t.Run("TestConfigCreation", func(t *testing.T) {
		config := DefaultTestConfig()
		if config == nil {
			t.Fatal("Failed to create test config")
		}
		t.Logf("Test config created successfully: artifact_dir=%s", config.ArtifactDir)
	})

	t.Run("TestDirectoryCreation", func(t *testing.T) {
		config := DefaultTestConfig()
		err := config.EnsureDirectories()
		if err != nil {
			t.Fatalf("Failed to ensure directories: %v", err)
		}

		// Check if directories exist
		if _, err := os.Stat(config.ArtifactDir); os.IsNotExist(err) {
			t.Errorf("Artifact directory not created: %s", config.ArtifactDir)
		}
		if _, err := os.Stat(config.LogDir); os.IsNotExist(err) {
			t.Errorf("Log directory not created: %s", config.LogDir)
		}

		t.Log("All test directories created successfully")
	})

	t.Run("TestLoggerCreation", func(t *testing.T) {
		logger, err := NewTestLogger(t, "infrastructure_test")
		if err != nil {
			t.Fatalf("Failed to create test logger: %v", err)
		}
		defer logger.Close()

		logger.Info("Test logger working correctly")
		logger.Debug("Debug message test")
		logger.Warn("Warning message test")
		logger.Step(1, "Testing step logging")
		logger.Resource("CREATE", "test_resource", "test_name")

		t.Log("Test logger created and working successfully")
	})

	t.Run("TestMetricsTracking", func(t *testing.T) {
		metrics := NewTestMetrics("infrastructure_test")
		if metrics == nil {
			t.Fatal("Failed to create test metrics")
		}

		metrics.RecordResourceCreated()
		metrics.RecordResourceCreated()
		metrics.RecordResourceDeleted()
		metrics.RecordAPICall()
		metrics.RecordAPICall()
		metrics.RecordAPICall()

		metrics.Finalize()

		if metrics.ResourcesCreated != 2 {
			t.Errorf("Expected 2 resources created, got %d", metrics.ResourcesCreated)
		}
		if metrics.ResourcesDeleted != 1 {
			t.Errorf("Expected 1 resource deleted, got %d", metrics.ResourcesDeleted)
		}
		if metrics.APICallCount != 3 {
			t.Errorf("Expected 3 API calls, got %d", metrics.APICallCount)
		}

		t.Logf("Metrics tracking working: %s", metrics.Summary())
	})

	t.Run("TestResourceNaming", func(t *testing.T) {
		naming := NewResourceNamingConfig()
		if naming == nil {
			t.Fatal("Failed to create resource naming config")
		}

		name := naming.GenerateName("gateway")
		if name == "" {
			t.Error("Generated name is empty")
		}

		t.Logf("Resource naming working: %s", name)
	})

	t.Run("TestCloudProviderConfig", func(t *testing.T) {
		cloudConfig := DefaultCloudProviderTestConfig()
		if cloudConfig == nil {
			t.Fatal("Failed to create cloud provider config")
		}

		if cloudConfig.AWSTestRegion == "" {
			t.Error("AWS test region not set")
		}
		if cloudConfig.AzureTestRegion == "" {
			t.Error("Azure test region not set")
		}
		if cloudConfig.GCPTestRegion == "" {
			t.Error("GCP test region not set")
		}
		if cloudConfig.OCITestRegion == "" {
			t.Error("OCI test region not set")
		}

		t.Log("Cloud provider config created successfully")
	})

	t.Run("TestEnvironmentValidation", func(t *testing.T) {
		env := NewTestEnvironment()

		// These should not panic, just skip if credentials are missing
		t.Run("SkipValidation", func(t *testing.T) {
			// We're testing that validation functions work, not that credentials exist
			// So we'll just verify the functions can be called
			hasAWS := env.AWSAccessKeyID != ""
			hasAzure := env.ARMClientID != ""
			hasGCP := env.GoogleCredentials != ""
			hasOCI := env.OCIUserID != ""

			t.Logf("Credential availability - AWS: %v, Azure: %v, GCP: %v, OCI: %v",
				hasAWS, hasAzure, hasGCP, hasOCI)
		})
	})
}

// TestDockerBuildSmoke is a placeholder for Docker build validation
// This would be run separately to validate Docker builds work
func TestDockerBuildSmoke(t *testing.T) {
	if os.Getenv("RUN_DOCKER_TESTS") != "1" {
		t.Skip("Skipping Docker tests (set RUN_DOCKER_TESTS=1 to run)")
	}

	// This test is intentionally simple - it just validates the test can run
	t.Log("Docker smoke test placeholder")
	// In a real scenario, you might use docker client to validate builds
}

// TestGitHubActionsWorkflow is a placeholder for CI/CD validation
func TestGitHubActionsWorkflow(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping CI workflow test (only runs in CI)")
	}

	// Validate CI-specific environment variables
	t.Run("CIEnvironment", func(t *testing.T) {
		if os.Getenv("GITHUB_ACTIONS") == "true" {
			t.Log("Running in GitHub Actions")
		}
	})
}

// TestProviderInitialization validates the provider can be initialized
func TestProviderInitialization(t *testing.T) {
	t.Run("ProviderInstance", func(t *testing.T) {
		provider := Provider()
		if provider == nil {
			t.Fatal("Provider is nil")
		}

		err := provider.InternalValidate()
		if err != nil {
			t.Fatalf("Provider validation failed: %v", err)
		}

		t.Log("Provider initialized and validated successfully")
	})

	t.Run("TestProviders", func(t *testing.T) {
		if testAccProviders == nil {
			t.Fatal("testAccProviders is nil")
		}

		if _, ok := testAccProviders["aviatrix"]; !ok {
			t.Error("aviatrix provider not found in testAccProviders")
		}

		t.Log("Test providers configured correctly")
	})
}

// BenchmarkTestHelpers benchmarks the test helper functions
func BenchmarkTestHelpers(b *testing.B) {
	b.Run("NewTestEnvironment", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewTestEnvironment()
		}
	})

	b.Run("DefaultTestConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = DefaultTestConfig()
		}
	})

	b.Run("ResourceNaming", func(b *testing.B) {
		naming := NewResourceNamingConfig()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = naming.GenerateName("gateway")
		}
	})
}
