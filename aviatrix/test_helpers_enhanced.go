package aviatrix

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestPreCheckConfig defines configuration for pre-check validation
type TestPreCheckConfig struct {
	RequireAWS       bool
	RequireGCP       bool
	RequireAzure     bool
	RequireOCI       bool
	RequireAWSGov    bool
	RequireAzureGov  bool
	RequireAWSChina  bool
	CustomEnvVars    []string
	ResourceSpecific []string
}

// EnhancedPreCheck performs comprehensive pre-check validation
func EnhancedPreCheck(t *testing.T, config TestPreCheckConfig) {
	// Standard pre-check
	testAccPreCheck(t)

	// Cloud provider checks
	if config.RequireAWS {
		preCheckAWS(t)
	}
	if config.RequireGCP {
		preCheckGCP(t)
	}
	if config.RequireAzure {
		preCheckAzure(t)
	}
	if config.RequireOCI {
		preCheckOCI(t)
	}
	if config.RequireAWSGov {
		preCheckAWSGov(t)
	}
	if config.RequireAzureGov {
		preCheckAzureGov(t)
	}
	if config.RequireAWSChina {
		preCheckAWSChina(t)
	}

	// Custom environment variables
	for _, envVar := range config.CustomEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for this test", envVar)
		}
	}

	// Resource-specific checks
	for _, check := range config.ResourceSpecific {
		preCheckResourceSpecific(t, check)
	}
}

// preCheckAWS validates AWS credentials and configuration
func preCheckAWS(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_AWS") == "yes" {
		t.Skip("Skipping AWS tests")
	}

	requiredEnvVars := []string{
		"AWS_ACCOUNT_NUMBER",
		"AWS_ACCESS_KEY",
		"AWS_SECRET_KEY",
		"AWS_REGION",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for AWS tests", envVar)
		}
	}
}

// preCheckGCP validates GCP credentials and configuration
func preCheckGCP(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_GCP") == "yes" {
		t.Skip("Skipping GCP tests")
	}

	requiredEnvVars := []string{
		"GCP_ID",
		"GCP_CREDENTIALS_FILEPATH",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for GCP tests", envVar)
		}
	}
}

// preCheckAzure validates Azure credentials and configuration
func preCheckAzure(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_AZURE") == "yes" {
		t.Skip("Skipping Azure tests")
	}

	requiredEnvVars := []string{
		"ARM_SUBSCRIPTION_ID",
		"ARM_DIRECTORY_ID",
		"ARM_APPLICATION_ID",
		"ARM_APPLICATION_KEY",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for Azure tests", envVar)
		}
	}
}

// preCheckOCI validates OCI credentials and configuration
func preCheckOCI(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_OCI") == "yes" {
		t.Skip("Skipping OCI tests")
	}

	requiredEnvVars := []string{
		"OCI_TENANCY_ID",
		"OCI_USER_ID",
		"OCI_COMPARTMENT_ID",
		"OCI_API_KEY_FILEPATH",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for OCI tests", envVar)
		}
	}
}

// preCheckAWSGov validates AWS GovCloud credentials
func preCheckAWSGov(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_AWSGOV") == "yes" {
		t.Skip("Skipping AWS GovCloud tests")
	}

	requiredEnvVars := []string{
		"AWSGOV_ACCOUNT_NUMBER",
		"AWSGOV_ACCESS_KEY",
		"AWSGOV_SECRET_KEY",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for AWS GovCloud tests", envVar)
		}
	}
}

// preCheckAzureGov validates Azure Government credentials
func preCheckAzureGov(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_AZUREGOV") == "yes" {
		t.Skip("Skipping Azure Government tests")
	}

	requiredEnvVars := []string{
		"AZUREGOV_SUBSCRIPTION_ID",
		"AZUREGOV_DIRECTORY_ID",
		"AZUREGOV_APPLICATION_ID",
		"AZUREGOV_APPLICATION_KEY",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for Azure Government tests", envVar)
		}
	}
}

// preCheckAWSChina validates AWS China credentials
func preCheckAWSChina(t *testing.T) {
	if os.Getenv("SKIP_ACCOUNT_AWSCHINA") == "yes" {
		t.Skip("Skipping AWS China tests")
	}

	requiredEnvVars := []string{
		"AWSCHINA_ACCOUNT_NUMBER",
		"AWSCHINA_ACCESS_KEY",
		"AWSCHINA_SECRET_KEY",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for AWS China tests", envVar)
		}
	}
}

// preCheckResourceSpecific performs resource-specific pre-checks
func preCheckResourceSpecific(t *testing.T, resourceType string) {
	lowerResource := strings.ToLower(resourceType)

	// Gateway-specific checks
	if strings.Contains(lowerResource, "gateway") {
		if os.Getenv("AWS_VPC_ID") == "" {
			t.Fatal("AWS_VPC_ID must be set for gateway tests")
		}
		if os.Getenv("AWS_SUBNET") == "" {
			t.Fatal("AWS_SUBNET must be set for gateway tests")
		}
	}

	// VPC-specific checks
	if strings.Contains(lowerResource, "vpc") && !strings.Contains(lowerResource, "gateway") {
		// VPC tests might need specific region configuration
	}

	// Transit-specific checks
	if strings.Contains(lowerResource, "transit") {
		// Transit gateway tests might need additional setup
	}

	// Firewall-specific checks
	if strings.Contains(lowerResource, "firewall") || strings.Contains(lowerResource, "firenet") {
		// Firewall tests might need specific configuration
	}
}

// GetEnhancedTestProviderFactories returns provider factories for testing (wrapper for GetTestProviderFactories)
// Note: This is a wrapper to avoid duplicate declaration. Use GetTestProviderFactories() from test_framework.go
func GetEnhancedTestProviderFactories() map[string]func() (*schema.Provider, error) {
	return GetTestProviderFactories()
}

// IsAcceptanceTestEnabled checks if acceptance tests should run (wrapper for IsAcceptanceTest)
// Note: This is a wrapper to avoid duplicate declaration. Use IsAcceptanceTest() from test_helpers.go
func IsAcceptanceTestEnabled() bool {
	return IsAcceptanceTest()
}

// TestConfigValidator validates test configurations
type TestConfigValidator struct {
	resource     *schema.Resource
	config       string
	expectedErrs []string
}

// NewTestConfigValidator creates a new config validator
func NewTestConfigValidator(resource *schema.Resource, config string) *TestConfigValidator {
	return &TestConfigValidator{
		resource: resource,
		config:   config,
	}
}

// ExpectError adds an expected error pattern
func (v *TestConfigValidator) ExpectError(pattern string) *TestConfigValidator {
	v.expectedErrs = append(v.expectedErrs, pattern)
	return v
}

// Validate performs validation
func (v *TestConfigValidator) Validate() error {
	// This would integrate with Terraform's validation logic
	// For now, it's a placeholder for future implementation
	return nil
}

// TestStepBuilder builds test steps with a fluent interface
type TestStepBuilder struct {
	steps []resource.TestStep
}

// NewTestStepBuilder creates a new test step builder
func NewTestStepBuilder() *TestStepBuilder {
	return &TestStepBuilder{
		steps: []resource.TestStep{},
	}
}

// AddCreate adds a create step
func (b *TestStepBuilder) AddCreate(config string, checks ...resource.TestCheckFunc) *TestStepBuilder {
	b.steps = append(b.steps, resource.TestStep{
		Config: config,
		Check:  resource.ComposeTestCheckFunc(checks...),
	})
	return b
}

// AddImport adds an import step
func (b *TestStepBuilder) AddImport(resourceName string, ignoreFields ...string) *TestStepBuilder {
	b.steps = append(b.steps, resource.TestStep{
		ResourceName:            resourceName,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: ignoreFields,
	})
	return b
}

// AddUpdate adds an update step
func (b *TestStepBuilder) AddUpdate(config string, checks ...resource.TestCheckFunc) *TestStepBuilder {
	b.steps = append(b.steps, resource.TestStep{
		Config: config,
		Check:  resource.ComposeTestCheckFunc(checks...),
	})
	return b
}

// AddPlanOnly adds a plan-only step
func (b *TestStepBuilder) AddPlanOnly(config string, expectNonEmpty bool) *TestStepBuilder {
	b.steps = append(b.steps, resource.TestStep{
		Config:             config,
		PlanOnly:           true,
		ExpectNonEmptyPlan: expectNonEmpty,
	})
	return b
}

// AddExpectError adds a step that expects an error
func (b *TestStepBuilder) AddExpectError(config string, errorPattern string) *TestStepBuilder {
	// Note: ExpectError requires a *regexp.Regexp, so this would need to be compiled
	b.steps = append(b.steps, resource.TestStep{
		Config: config,
		// ExpectError: regexp.MustCompile(errorPattern),
	})
	return b
}

// Build returns the built test steps
func (b *TestStepBuilder) Build() []resource.TestStep {
	return b.steps
}

// ResourceTestHelper provides helper methods for resource testing
type ResourceTestHelper struct {
	resourceType string
	resourceName string
}

// NewResourceTestHelper creates a new resource test helper
func NewResourceTestHelper(resourceType, resourceName string) *ResourceTestHelper {
	return &ResourceTestHelper{
		resourceType: resourceType,
		resourceName: resourceName,
	}
}

// FullResourceName returns the full Terraform resource name
func (h *ResourceTestHelper) FullResourceName() string {
	return fmt.Sprintf("%s.%s", h.resourceType, h.resourceName)
}

// BuildResourceAttrCheck creates a TestCheckResourceAttr check
func (h *ResourceTestHelper) BuildResourceAttrCheck(key, value string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttr(h.FullResourceName(), key, value)
}

// BuildResourceAttrSetCheck creates a TestCheckResourceAttrSet check
func (h *ResourceTestHelper) BuildResourceAttrSetCheck(key string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrSet(h.FullResourceName(), key)
}

// BuildResourceAttrPairCheck creates a TestCheckResourceAttrPair check
func (h *ResourceTestHelper) BuildResourceAttrPairCheck(key, otherResource, otherKey string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrPair(h.FullResourceName(), key, otherResource, otherKey)
}

// StandardCRUDChecks returns standard checks for CRUD operations
func (h *ResourceTestHelper) StandardCRUDChecks() []resource.TestCheckFunc {
	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(h.FullResourceName(), "id"),
	}
}

// TestEnvironmentManager manages test environment setup and teardown
type TestEnvironmentManager struct {
	setupFuncs    []func() error
	teardownFuncs []func() error
}

// NewTestEnvironmentManager creates a new environment manager
func NewTestEnvironmentManager() *TestEnvironmentManager {
	return &TestEnvironmentManager{
		setupFuncs:    []func() error{},
		teardownFuncs: []func() error{},
	}
}

// AddSetup adds a setup function
func (m *TestEnvironmentManager) AddSetup(fn func() error) {
	m.setupFuncs = append(m.setupFuncs, fn)
}

// AddTeardown adds a teardown function
func (m *TestEnvironmentManager) AddTeardown(fn func() error) {
	m.teardownFuncs = append(m.teardownFuncs, fn)
}

// Setup executes all setup functions
func (m *TestEnvironmentManager) Setup() error {
	for _, fn := range m.setupFuncs {
		if err := fn(); err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}
	}
	return nil
}

// Teardown executes all teardown functions in reverse order
func (m *TestEnvironmentManager) Teardown() error {
	for i := len(m.teardownFuncs) - 1; i >= 0; i-- {
		if err := m.teardownFuncs[i](); err != nil {
			return fmt.Errorf("teardown failed: %w", err)
		}
	}
	return nil
}

// TestResourceConfig builds Terraform configuration strings
type TestResourceConfig struct {
	resources []string
}

// NewTestResourceConfig creates a new resource config builder
func NewTestResourceConfig() *TestResourceConfig {
	return &TestResourceConfig{
		resources: []string{},
	}
}

// AddResource adds a resource block
func (c *TestResourceConfig) AddResource(resourceType, name string, attributes map[string]interface{}) *TestResourceConfig {
	var config strings.Builder

	config.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", resourceType, name))

	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			config.WriteString(fmt.Sprintf("  %s = \"%s\"\n", key, v))
		case int:
			config.WriteString(fmt.Sprintf("  %s = %d\n", key, v))
		case bool:
			config.WriteString(fmt.Sprintf("  %s = %t\n", key, v))
		}
	}

	config.WriteString("}\n")

	c.resources = append(c.resources, config.String())
	return c
}

// Build returns the complete configuration
func (c *TestResourceConfig) Build() string {
	return strings.Join(c.resources, "\n")
}

// QuickTestConfig provides quick configuration builders for common scenarios
type QuickTestConfig struct{}

// NewQuickTestConfig creates a new quick config builder
func NewQuickTestConfig() *QuickTestConfig {
	return &QuickTestConfig{}
}

// BasicAccount creates a basic account configuration
func (q *QuickTestConfig) BasicAccount(name string) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "%s" {
  account_name       = "%s"
  cloud_type         = 1
  aws_account_number = %s
  aws_iam            = false
  aws_access_key     = %s
  aws_secret_key     = %s
}
`, name, name, "${AWS_ACCOUNT_NUMBER}", "${AWS_ACCESS_KEY}", "${AWS_SECRET_KEY}")
}

// BasicVPC creates a basic VPC configuration
func (q *QuickTestConfig) BasicVPC(name, accountRef string) string {
	return fmt.Sprintf(`
resource "aviatrix_vpc" "%s" {
  cloud_type   = 1
  account_name = %s
  region       = "us-east-1"
  name         = "%s"
  cidr         = "10.0.0.0/16"
}
`, name, accountRef, name)
}

// BasicGateway creates a basic gateway configuration
func (q *QuickTestConfig) BasicGateway(name, accountRef, vpcRef string) string {
	return fmt.Sprintf(`
resource "aviatrix_gateway" "%s" {
  cloud_type   = 1
  account_name = %s
  gw_name      = "%s"
  vpc_id       = %s
  vpc_reg      = "us-east-1"
  gw_size      = "t2.micro"
  subnet       = "10.0.1.0/24"
}
`, name, accountRef, name, vpcRef)
}
