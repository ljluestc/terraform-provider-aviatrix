package aviatrix

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Example: Integration test for Account resource using the framework
func TestIntegrationFramework_Account_CRUD(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_FRAMEWORK_EXAMPLE") == "yes" {
		t.Skip("Skipping integration framework example test")
	}

	// Initialize framework
	framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

	// Generate test data
	testData := framework.testDataProvider.GenerateAccountData(1) // AWS
	accountName := testData["account_name"].(string)
	resourceName := "aviatrix_account.test"

	// Build test configuration
	createConfig := fmt.Sprintf(`
resource "aviatrix_account" "test" {
  account_name       = "%s"
  cloud_type         = 1
  aws_account_number = "%s"
  aws_iam            = false
  aws_access_key     = "%s"
  aws_secret_key     = "%s"
}
	`, accountName,
		os.Getenv("AWS_ACCOUNT_NUMBER"),
		os.Getenv("AWS_ACCESS_KEY"),
		os.Getenv("AWS_SECRET_KEY"))

	updateConfig := fmt.Sprintf(`
resource "aviatrix_account" "test" {
  account_name       = "%s"
  cloud_type         = 1
  aws_account_number = "%s"
  aws_iam            = false
  aws_access_key     = "%s"
  aws_secret_key     = "%s"
  audit_account      = false
}
	`, accountName,
		os.Getenv("AWS_ACCOUNT_NUMBER"),
		os.Getenv("AWS_ACCESS_KEY"),
		os.Getenv("AWS_SECRET_KEY"))

	// Build test checks
	createChecks := NewTestCheckFuncBuilder().
		AddResourceAttr(resourceName, "account_name", accountName).
		AddResourceAttr(resourceName, "cloud_type", "1").
		AddResourceAttrSet(resourceName, "aws_account_number").
		Build()

	updateChecks := NewTestCheckFuncBuilder().
		AddResourceAttr(resourceName, "account_name", accountName).
		AddResourceAttr(resourceName, "audit_account", "false").
		Build()

	// Configure CRUD test
	crudConfig := CRUDTestConfig{
		ResourceName:                resourceName,
		PreCheck:                    func() { testAccPreCheck(t); preAccountCheck(t, "") },
		CreateConfig:                createConfig,
		UpdateConfig:                updateConfig,
		ImportStateVerifyIgnore:     []string{"aws_secret_key", "aws_access_key", "audit_account"},
		CheckDestroy:                testAccCheckAccountDestroy,
		CreateChecks:                []resource.TestCheckFunc{createChecks},
		UpdateChecks:                []resource.TestCheckFunc{updateChecks},
	}

	// Run the test
	testFunc := framework.GenerateCRUDTest(crudConfig)
	testFunc(t)
}

// Example: Error handling tests using the framework
func TestIntegrationFramework_Account_ErrorHandling(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_FRAMEWORK_EXAMPLE") == "yes" {
		t.Skip("Skipping integration framework example test")
	}

	framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

	errorConfigs := []ErrorHandlingTestConfig{
		{
			Name: "missing_required_field_account_name",
			Config: `
resource "aviatrix_account" "test" {
  cloud_type = 1
}
			`,
			ExpectError:       true,
			ErrorMessageMatch: "account_name",
			PreCheck:          func() { testAccPreCheck(t) },
		},
		{
			Name: "invalid_cloud_type",
			Config: `
resource "aviatrix_account" "test" {
  account_name = "test-invalid-cloud"
  cloud_type   = 999
}
			`,
			ExpectError:       true,
			ErrorMessageMatch: "cloud_type",
			PreCheck:          func() { testAccPreCheck(t) },
		},
	}

	testFunc := framework.GenerateErrorHandlingTests(errorConfigs)
	testFunc(t)
}

// Example: Dependency test using the framework
func TestIntegrationFramework_Gateway_DependsOnAccount(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_FRAMEWORK_EXAMPLE") == "yes" {
		t.Skip("Skipping integration framework example test")
	}
	if os.Getenv("SKIP_ACCOUNT_AWS") == "yes" {
		t.Skip("Skipping AWS tests")
	}

	framework := NewIntegrationTestFramework(t, "aviatrix_gateway", resourceAviatrixGateway())

	accountData := framework.testDataProvider.GenerateAccountData(1)
	gatewayData := framework.testDataProvider.GenerateGatewayData(1)

	dependencyConfig := fmt.Sprintf(`
resource "aviatrix_account" "test" {
  account_name       = "%s"
  cloud_type         = 1
  aws_account_number = "%s"
  aws_iam            = false
  aws_access_key     = "%s"
  aws_secret_key     = "%s"
}
	`, accountData["account_name"],
		os.Getenv("AWS_ACCOUNT_NUMBER"),
		os.Getenv("AWS_ACCESS_KEY"),
		os.Getenv("AWS_SECRET_KEY"))

	dependentConfig := fmt.Sprintf(`
resource "aviatrix_gateway" "test" {
  cloud_type   = 1
  account_name = aviatrix_account.test.account_name
  gw_name      = "%s"
  vpc_id       = "%s"
  vpc_reg      = "%s"
  gw_size      = "t2.micro"
  subnet       = "%s"
}
	`, gatewayData["gw_name"],
		os.Getenv("AWS_VPC_ID"),
		os.Getenv("AWS_REGION"),
		os.Getenv("AWS_SUBNET"))

	checks := NewTestCheckFuncBuilder().
		AddResourceAttrPair("aviatrix_gateway.test", "account_name", "aviatrix_account.test", "account_name").
		AddResourceAttr("aviatrix_gateway.test", "cloud_type", "1").
		Build()

	depConfig := DependencyTestConfig{
		ResourceName:     "aviatrix_gateway.test",
		DependentConfig:  dependentConfig,
		DependencyConfig: dependencyConfig,
		Checks:           []resource.TestCheckFunc{checks},
		PreCheck: func() {
			testAccPreCheck(t)
			preAccountCheck(t, "")
			preGatewayCheck(t, "")
		},
	}

	testFunc := framework.GenerateDependencyTest(depConfig)
	testFunc(t)
}

// Example: Import test using the framework
func TestIntegrationFramework_Account_Import(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_FRAMEWORK_EXAMPLE") == "yes" {
		t.Skip("Skipping integration framework example test")
	}

	framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())
	testData := framework.testDataProvider.GenerateAccountData(1)
	accountName := testData["account_name"].(string)
	resourceName := "aviatrix_account.test"

	config := fmt.Sprintf(`
resource "aviatrix_account" "test" {
  account_name       = "%s"
  cloud_type         = 1
  aws_account_number = "%s"
  aws_iam            = false
  aws_access_key     = "%s"
  aws_secret_key     = "%s"
}
	`, accountName,
		os.Getenv("AWS_ACCOUNT_NUMBER"),
		os.Getenv("AWS_ACCESS_KEY"),
		os.Getenv("AWS_SECRET_KEY"))

	importConfig := IntegrationImportTestConfig{
		ResourceName:                resourceName,
		Config:                      config,
		ImportID:                    accountName,
		ImportStateVerifyIgnore:     []string{"aws_secret_key", "aws_access_key", "audit_account"},
		PreCheck:                    func() { testAccPreCheck(t); preAccountCheck(t, "") },
		CheckDestroy:                testAccCheckAccountDestroy,
	}

	testFunc := framework.GenerateImportTest(importConfig)
	testFunc(t)
}

// Example: Using TestConfigBuilder
func TestIntegrationFramework_ConfigBuilder(t *testing.T) {
	builder := NewTestConfigBuilder()

	config := builder.
		AddResourceBlock("aviatrix_account", "test", map[string]interface{}{
			"account_name":       "test-account",
			"cloud_type":         1,
			"aws_account_number": "123456789012",
			"aws_iam":            false,
		}).
		AddResourceBlock("aviatrix_vpc", "test", map[string]interface{}{
			"cloud_type":   1,
			"account_name": "aviatrix_account.test.account_name",
			"region":       "us-east-1",
			"name":         "test-vpc",
			"cidr":         "10.0.0.0/16",
		}).
		Build()

	// Verify config is not empty
	if config == "" {
		t.Error("Generated config should not be empty")
	}

	// Verify config contains expected resources
	if !integrationTestContains(config, "resource \"aviatrix_account\" \"test\"") {
		t.Error("Config should contain account resource")
	}

	if !integrationTestContains(config, "resource \"aviatrix_vpc\" \"test\"") {
		t.Error("Config should contain VPC resource")
	}
}

// Example: Using TestConfigTemplateEngine
func TestIntegrationFramework_TemplateEngine(t *testing.T) {
	engine := NewTestConfigTemplateEngine()

	data := map[string]interface{}{
		"Name":             "test",
		"AccountName":      "test-account",
		"AWSAccountNumber": "123456789012",
		"AWSIam":           "false",
		"AWSAccessKey":     "AKIAIOSFODNN7EXAMPLE",
		"AWSSecretKey":     "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	config, err := engine.RenderTemplate("aws_account", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Verify config contains expected values
	if !integrationTestContains(config, "test-account") {
		t.Error("Config should contain account name")
	}

	if !integrationTestContains(config, "123456789012") {
		t.Error("Config should contain account number")
	}
}

// Example: Using ResourceDependencyGraph
func TestIntegrationFramework_DependencyGraph(t *testing.T) {
	graph := NewResourceDependencyGraph()

	// Add dependencies: Gateway depends on Account and VPC
	graph.AddDependency("aviatrix_gateway.test", "aviatrix_account.test")
	graph.AddDependency("aviatrix_gateway.test", "aviatrix_vpc.test")
	graph.AddDependency("aviatrix_vpc.test", "aviatrix_account.test")

	// Valid order: Account -> VPC -> Gateway
	validOrder := []string{
		"aviatrix_account.test",
		"aviatrix_vpc.test",
		"aviatrix_gateway.test",
	}

	err := graph.ValidateDependencyOrder(validOrder)
	if err != nil {
		t.Errorf("Valid order should not produce error: %v", err)
	}

	// Invalid order: Gateway before Account
	invalidOrder := []string{
		"aviatrix_gateway.test",
		"aviatrix_account.test",
		"aviatrix_vpc.test",
	}

	err = graph.ValidateDependencyOrder(invalidOrder)
	if err == nil {
		t.Error("Invalid order should produce error")
	}
}

// Example: Using TestDataProvider
func TestIntegrationFramework_TestDataProvider(t *testing.T) {
	provider := NewTestDataProvider()

	// Test AWS account data generation
	awsData := provider.GenerateAccountData(1)
	if awsData["cloud_type"] != 1 {
		t.Error("AWS cloud type should be 1")
	}

	if awsData["account_name"] == "" {
		t.Error("Account name should not be empty")
	}

	// Test GCP account data generation
	gcpData := provider.GenerateAccountData(4)
	if gcpData["cloud_type"] != 4 {
		t.Error("GCP cloud type should be 4")
	}

	// Test Azure account data generation
	azureData := provider.GenerateAccountData(8)
	if azureData["cloud_type"] != 8 {
		t.Error("Azure cloud type should be 8")
	}

	// Test OCI account data generation
	ociData := provider.GenerateAccountData(16)
	if ociData["cloud_type"] != 16 {
		t.Error("OCI cloud type should be 16")
	}
}

// Helper functions for tests that need string containment checking
func integrationTestContains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || integrationTestHasSubstring(s, substr)))
}

func integrationTestHasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
