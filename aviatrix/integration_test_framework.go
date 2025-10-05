package aviatrix

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// IntegrationTestFramework provides comprehensive integration testing for Terraform resources
type IntegrationTestFramework struct {
	t                *testing.T
	resource         *schema.Resource
	resourceName     string
	resourceType     string
	testDataProvider TestDataProvider
	dependencyGraph  *ResourceDependencyGraph
}

// NewIntegrationTestFramework creates a new integration test framework instance
func NewIntegrationTestFramework(t *testing.T, resourceType string, resource *schema.Resource) *IntegrationTestFramework {
	return &IntegrationTestFramework{
		t:                t,
		resource:         resource,
		resourceType:     resourceType,
		testDataProvider: NewTestDataProvider(),
		dependencyGraph:  NewResourceDependencyGraph(),
	}
}

// CRUDTestConfig defines configuration for CRUD operation testing
type CRUDTestConfig struct {
	ResourceName     string
	PreCheck         func()
	CreateConfig     string
	UpdateConfig     string
	ImportStateVerifyIgnore []string
	CheckDestroy     resource.TestCheckFunc
	CreateChecks     []resource.TestCheckFunc
	UpdateChecks     []resource.TestCheckFunc
	SkipImport       bool
	SkipUpdate       bool
	DependsOn        []string
}

// GenerateCRUDTest generates a complete CRUD test for a resource
func (f *IntegrationTestFramework) GenerateCRUDTest(config CRUDTestConfig) func(*testing.T) {
	return func(t *testing.T) {
		if !IsAcceptanceTest() {
			t.Skip("Skipping acceptance test (TF_ACC not set)")
		}

		steps := f.buildCRUDTestSteps(config)

		resource.Test(t, resource.TestCase{
			PreCheck:          config.PreCheck,
			ProviderFactories: GetTestProviderFactories(),
			CheckDestroy:      config.CheckDestroy,
			Steps:             steps,
		})
	}
}

// buildCRUDTestSteps builds the test steps for CRUD operations
func (f *IntegrationTestFramework) buildCRUDTestSteps(config CRUDTestConfig) []resource.TestStep {
	var steps []resource.TestStep

	// Step 1: Create
	createStep := resource.TestStep{
		Config: config.CreateConfig,
		Check:  resource.ComposeTestCheckFunc(config.CreateChecks...),
	}
	steps = append(steps, createStep)

	// Step 2: Import (if not skipped)
	if !config.SkipImport {
		importStep := resource.TestStep{
			ResourceName:            config.ResourceName,
			ImportState:             true,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: config.ImportStateVerifyIgnore,
		}
		steps = append(steps, importStep)
	}

	// Step 3: Update (if not skipped)
	if !config.SkipUpdate && config.UpdateConfig != "" {
		updateStep := resource.TestStep{
			Config: config.UpdateConfig,
			Check:  resource.ComposeTestCheckFunc(config.UpdateChecks...),
		}
		steps = append(steps, updateStep)
	}

	// Step 4: Read (implicit in all steps via provider refresh)

	// Delete is handled by CheckDestroy

	return steps
}

// ErrorHandlingTestConfig defines configuration for error handling tests
type ErrorHandlingTestConfig struct {
	Name              string
	Config            string
	ExpectError       bool
	ErrorMessageMatch string
	PreCheck          func()
}

// GenerateErrorHandlingTests generates tests for error scenarios
func (f *IntegrationTestFramework) GenerateErrorHandlingTests(configs []ErrorHandlingTestConfig) func(*testing.T) {
	return func(t *testing.T) {
		for _, tc := range configs {
			t.Run(tc.Name, func(t *testing.T) {
				if !IsAcceptanceTest() {
					t.Skip("Skipping acceptance test (TF_ACC not set)")
				}

				var expectErrorRegex *regexp.Regexp
				if tc.ExpectError && tc.ErrorMessageMatch != "" {
					expectErrorRegex = regexp.MustCompile(regexp.QuoteMeta(tc.ErrorMessageMatch))
				}

				resource.Test(t, resource.TestCase{
					PreCheck:          tc.PreCheck,
					ProviderFactories: GetTestProviderFactories(),
					Steps: []resource.TestStep{
						{
							Config:      tc.Config,
							ExpectError: expectErrorRegex,
						},
					},
				})
			})
		}
	}
}

// DependencyTestConfig defines configuration for dependency testing
type DependencyTestConfig struct {
	ResourceName     string
	DependentConfig  string
	DependencyConfig string
	Checks           []resource.TestCheckFunc
	PreCheck         func()
}

// GenerateDependencyTest generates tests for resource dependencies
func (f *IntegrationTestFramework) GenerateDependencyTest(config DependencyTestConfig) func(*testing.T) {
	return func(t *testing.T) {
		if !IsAcceptanceTest() {
			t.Skip("Skipping acceptance test (TF_ACC not set)")
		}

		// Combine dependency and dependent configs
		fullConfig := fmt.Sprintf("%s\n\n%s", config.DependencyConfig, config.DependentConfig)

		resource.Test(t, resource.TestCase{
			PreCheck:          config.PreCheck,
			ProviderFactories: GetTestProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: fullConfig,
					Check:  resource.ComposeTestCheckFunc(config.Checks...),
				},
			},
		})
	}
}

// StateDriftTestConfig defines configuration for state drift detection
type StateDriftTestConfig struct {
	ResourceName       string
	InitialConfig      string
	ModifyResourceFunc func(*terraform.State) error
	ExpectedChanges    []string
	PreCheck           func()
}

// GenerateStateDriftTest generates tests for state drift detection
func (f *IntegrationTestFramework) GenerateStateDriftTest(config StateDriftTestConfig) func(*testing.T) {
	return func(t *testing.T) {
		if !IsAcceptanceTest() {
			t.Skip("Skipping acceptance test (TF_ACC not set)")
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          config.PreCheck,
			ProviderFactories: GetTestProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: config.InitialConfig,
				},
				{
					PreConfig: func() {
						// Simulate external modification
						// This would be implemented with actual API calls
						t.Log("Simulating external resource modification")
					},
					Config:             config.InitialConfig,
					PlanOnly:           true,
					ExpectNonEmptyPlan: len(config.ExpectedChanges) > 0,
				},
			},
		})
	}
}

// IntegrationImportTestConfig defines configuration for import functionality testing
type IntegrationImportTestConfig struct {
	ResourceName            string
	Config                  string
	ImportID                string
	ImportStateVerifyIgnore []string
	PreCheck                func()
	CheckDestroy            resource.TestCheckFunc
}

// GenerateImportTest generates a test specifically for import functionality
func (f *IntegrationTestFramework) GenerateImportTest(config IntegrationImportTestConfig) func(*testing.T) {
	return func(t *testing.T) {
		if !IsAcceptanceTest() {
			t.Skip("Skipping acceptance test (TF_ACC not set)")
		}

		resource.Test(t, resource.TestCase{
			PreCheck:          config.PreCheck,
			ProviderFactories: GetTestProviderFactories(),
			CheckDestroy:      config.CheckDestroy,
			Steps: []resource.TestStep{
				{
					Config: config.Config,
				},
				{
					ResourceName:            config.ResourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: config.ImportStateVerifyIgnore,
					ImportStateIdFunc: func(s *terraform.State) (string, error) {
						if config.ImportID != "" {
							return config.ImportID, nil
						}
						rs, ok := s.RootModule().Resources[config.ResourceName]
						if !ok {
							return "", fmt.Errorf("resource not found: %s", config.ResourceName)
						}
						return rs.Primary.ID, nil
					},
				},
			},
		})
	}
}

// ResourceDependencyGraph manages resource dependency relationships
type ResourceDependencyGraph struct {
	dependencies map[string][]string
}

// NewResourceDependencyGraph creates a new dependency graph
func NewResourceDependencyGraph() *ResourceDependencyGraph {
	return &ResourceDependencyGraph{
		dependencies: make(map[string][]string),
	}
}

// AddDependency adds a dependency relationship
func (g *ResourceDependencyGraph) AddDependency(resource, dependsOn string) {
	if g.dependencies[resource] == nil {
		g.dependencies[resource] = []string{}
	}
	g.dependencies[resource] = append(g.dependencies[resource], dependsOn)
}

// GetDependencies returns all dependencies for a resource
func (g *ResourceDependencyGraph) GetDependencies(resource string) []string {
	return g.dependencies[resource]
}

// ValidateDependencyOrder validates that dependencies are created in correct order
func (g *ResourceDependencyGraph) ValidateDependencyOrder(resourceOrder []string) error {
	created := make(map[string]bool)

	for _, resource := range resourceOrder {
		deps := g.GetDependencies(resource)
		for _, dep := range deps {
			if !created[dep] {
				return fmt.Errorf("resource %s depends on %s which hasn't been created yet", resource, dep)
			}
		}
		created[resource] = true
	}

	return nil
}

// TestCheckFuncBuilder helps build complex test check functions
type TestCheckFuncBuilder struct {
	checks []resource.TestCheckFunc
}

// NewTestCheckFuncBuilder creates a new builder
func NewTestCheckFuncBuilder() *TestCheckFuncBuilder {
	return &TestCheckFuncBuilder{
		checks: []resource.TestCheckFunc{},
	}
}

// AddResourceAttr adds a resource attribute check
func (b *TestCheckFuncBuilder) AddResourceAttr(name, key, value string) *TestCheckFuncBuilder {
	b.checks = append(b.checks, resource.TestCheckResourceAttr(name, key, value))
	return b
}

// AddResourceAttrSet adds a check that attribute is set
func (b *TestCheckFuncBuilder) AddResourceAttrSet(name, key string) *TestCheckFuncBuilder {
	b.checks = append(b.checks, resource.TestCheckResourceAttrSet(name, key))
	return b
}

// AddResourceAttrPair adds a check comparing attributes between resources
func (b *TestCheckFuncBuilder) AddResourceAttrPair(name1, key1, name2, key2 string) *TestCheckFuncBuilder {
	b.checks = append(b.checks, resource.TestCheckResourceAttrPair(name1, key1, name2, key2))
	return b
}

// AddCustomCheck adds a custom check function
func (b *TestCheckFuncBuilder) AddCustomCheck(check resource.TestCheckFunc) *TestCheckFuncBuilder {
	b.checks = append(b.checks, check)
	return b
}

// Build returns the composite check function
func (b *TestCheckFuncBuilder) Build() resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(b.checks...)
}

// RetryableOperation wraps operations that may need retries
type RetryableOperation struct {
	MaxRetries int
	RetryDelay time.Duration
}

// NewRetryableOperation creates a new retryable operation
func NewRetryableOperation(maxRetries int, retryDelay time.Duration) *RetryableOperation {
	return &RetryableOperation{
		MaxRetries: maxRetries,
		RetryDelay: retryDelay,
	}
}

// Execute executes an operation with retry logic
func (r *RetryableOperation) Execute(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.RetryDelay):
			}
		}

		if err := operation(); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("operation failed after %d retries: %w", r.MaxRetries, lastErr)
}

// ResourceExists is a helper to check if a resource exists in state
func ResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource has no ID set: %s", resourceName)
		}

		return nil
	}
}

// ResourceDoesNotExist is a helper to check if a resource does not exist in state
func ResourceDoesNotExist(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if ok {
			return fmt.Errorf("resource still exists: %s", resourceName)
		}
		return nil
	}
}

// TestConfigBuilder helps build test configurations
type TestConfigBuilder struct {
	blocks []string
}

// NewTestConfigBuilder creates a new config builder
func NewTestConfigBuilder() *TestConfigBuilder {
	return &TestConfigBuilder{
		blocks: []string{},
	}
}

// AddResourceBlock adds a resource block to the configuration
func (b *TestConfigBuilder) AddResourceBlock(resourceType, name string, attributes map[string]interface{}) *TestConfigBuilder {
	block := fmt.Sprintf("resource \"%s\" \"%s\" {\n", resourceType, name)

	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			block += fmt.Sprintf("  %s = \"%s\"\n", key, v)
		case int:
			block += fmt.Sprintf("  %s = %d\n", key, v)
		case bool:
			block += fmt.Sprintf("  %s = %t\n", key, v)
		case []string:
			block += fmt.Sprintf("  %s = [\n", key)
			for _, item := range v {
				block += fmt.Sprintf("    \"%s\",\n", item)
			}
			block += "  ]\n"
		}
	}

	block += "}\n"
	b.blocks = append(b.blocks, block)
	return b
}

// AddDataSourceBlock adds a data source block to the configuration
func (b *TestConfigBuilder) AddDataSourceBlock(dataSourceType, name string, attributes map[string]interface{}) *TestConfigBuilder {
	block := fmt.Sprintf("data \"%s\" \"%s\" {\n", dataSourceType, name)

	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			block += fmt.Sprintf("  %s = \"%s\"\n", key, v)
		case int:
			block += fmt.Sprintf("  %s = %d\n", key, v)
		case bool:
			block += fmt.Sprintf("  %s = %t\n", key, v)
		}
	}

	block += "}\n"
	b.blocks = append(b.blocks, block)
	return b
}

// Build returns the complete configuration
func (b *TestConfigBuilder) Build() string {
	return strings.Join(b.blocks, "\n")
}
