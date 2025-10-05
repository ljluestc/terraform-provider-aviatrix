# Resource Test Template and Generation Framework Guide

## Overview

This guide documents the comprehensive test template and generation framework for creating standardized integration tests for all 282 Aviatrix Terraform Provider resources.

## Table of Contents

1. [Architecture](#architecture)
2. [Core Components](#core-components)
3. [Usage Guide](#usage-guide)
4. [Test Templates](#test-templates)
5. [Code Generation](#code-generation)
6. [Best Practices](#best-practices)
7. [Examples](#examples)
8. [Troubleshooting](#troubleshooting)

## Architecture

### Framework Components

```
Resource Test Framework
├── Template Generation
│   ├── ResourceTestTemplateGenerator - Main generator
│   ├── TestTemplateEngine - Template rendering
│   └── TestCaseDefinition - Test case structure
├── Schema Analysis
│   ├── ResourceSchemaAnalyzer - Schema introspection
│   ├── SchemaFieldInfo - Field metadata
│   └── SchemaComplexityReport - Complexity analysis
├── Test Helpers
│   ├── TestDataProvider - Test data generation
│   ├── TestCheckFuncBuilder - Check function building
│   └── EnhancedPreCheck - Environment validation
└── Code Generation
    ├── TestGeneratorCLI - Command-line interface
    └── TestScaffoldConfig - Quick scaffolding
```

## Core Components

### 1. ResourceTestTemplateGenerator

Generates complete test files for resources with comprehensive CRUD, import, error handling, and update tests.

**Location**: `aviatrix/resource_test_template_generator.go`

**Key Methods**:
- `GenerateTestFile(outputPath string)` - Generates complete test file
- `generateCRUDTest()` - Creates basic CRUD test
- `generateImportTest()` - Creates import test
- `generateErrorHandlingTests()` - Creates error scenario tests
- `generateUpdateTests()` - Creates individual attribute update tests

### 2. ResourceSchemaAnalyzer

Analyzes resource schemas to extract metadata for intelligent test generation.

**Location**: `aviatrix/resource_schema_analyzer.go`

**Key Methods**:
- `GetRequiredFields()` - Returns required fields
- `GetOptionalFields()` - Returns optional fields
- `GetSensitiveFields()` - Returns sensitive fields to ignore in import
- `AnalyzeComplexity()` - Calculates complexity score
- `SuggestTestCount()` - Suggests number of test cases based on complexity
- `IdentifyDependencies()` - Identifies resource dependencies

### 3. Test Data Providers

Generates realistic test data based on cloud provider and resource type.

**Location**: `aviatrix/integration_test_data_generator.go`

**Key Functions**:
- `GenerateResourceName(prefix)` - Generates unique resource names
- `GenerateAccountData(cloudType)` - Generates account test data
- `GenerateVPCData(cloudType)` - Generates VPC test data
- `GenerateGatewayData(cloudType)` - Generates gateway test data

### 4. Enhanced PreCheck Utilities

Comprehensive environment validation before test execution.

**Location**: `aviatrix/test_helpers_enhanced.go`

**Key Functions**:
- `EnhancedPreCheck(t, config)` - Comprehensive pre-check validation
- `preCheckAWS(t)` - AWS-specific validation
- `preCheckGCP(t)` - GCP-specific validation
- `preCheckAzure(t)` - Azure-specific validation
- `preCheckOCI(t)` - OCI-specific validation

## Usage Guide

### Generating Tests for a Single Resource

```go
// Import the generator
import "github.com/AviatrixSystems/terraform-provider-aviatrix/v3/aviatrix"

// Get the provider and resource
provider := aviatrix.Provider()
resource := provider.ResourcesMap["aviatrix_account"]

// Create generator
generator := aviatrix.NewResourceTestTemplateGenerator("aviatrix_account", resource)

// Generate test file
err := generator.GenerateTestFile("./aviatrix/resource_aviatrix_account_test.go")
if err != nil {
    log.Fatal(err)
}
```

### Using the CLI Tool

```bash
# Generate test for a single resource
go run ./cmd/test-generator -resource aviatrix_account -output ./aviatrix

# Generate tests for multiple resources
go run ./cmd/test-generator \
    -resource aviatrix_account \
    -resource aviatrix_gateway \
    -resource aviatrix_vpc \
    -output ./aviatrix

# Generate tests for all resources
go run ./cmd/test-generator -all -output ./aviatrix

# Dry run to preview what would be generated
go run ./cmd/test-generator -all -dry-run -verbose

# Generate with verbose output
go run ./cmd/test-generator -resource aviatrix_account -verbose
```

### Analyzing Resource Complexity

```go
// Analyze a resource
analyzer := aviatrix.NewResourceSchemaAnalyzer(resource)
report := analyzer.GenerateComplexityReport()

fmt.Printf("Total Fields: %d\n", report.TotalFields)
fmt.Printf("Complexity Score: %d\n", report.ComplexityScore)
fmt.Printf("Suggested Tests: %d\n", report.SuggestedTests)
```

### Generating Test Scaffold

```go
// Quick scaffold for manual customization
config := aviatrix.TestScaffoldConfig{
    ResourceType:   "aviatrix_account",
    IncludeCRUD:    true,
    IncludeImport:  true,
    IncludeErrors:  true,
    IncludeUpdates: true,
}

err := aviatrix.GenerateTestScaffold(config, "./aviatrix/resource_aviatrix_account_test.go")
```

## Test Templates

### Standard Test Naming Convention

All generated tests follow the pattern:
```
TestAccAviatrix{ResourceName}_{scenario}
```

Examples:
- `TestAccAviatrixAccount_basic` - Basic CRUD test
- `TestAccAviatrixAccount_import` - Import test
- `TestAccAviatrixAccount_missingRequired_account_name` - Error test
- `TestAccAviatrixAccount_update_cloud_type` - Update test

### Test Configuration Pattern

```go
func testAccAviatrixAccount_basic(rName string) string {
    return fmt.Sprintf(`
resource "aviatrix_account" "test" {
  account_name       = "%s"
  cloud_type         = 1
  aws_account_number = %s
  aws_iam            = false
  aws_access_key     = %s
  aws_secret_key     = %s
}
    `, rName,
       os.Getenv("AWS_ACCOUNT_NUMBER"),
       os.Getenv("AWS_ACCESS_KEY"),
       os.Getenv("AWS_SECRET_KEY"))
}
```

### Test Check Pattern

```go
check := resource.ComposeTestCheckFunc(
    resource.TestCheckResourceAttr("aviatrix_account.test", "account_name", accountName),
    resource.TestCheckResourceAttr("aviatrix_account.test", "cloud_type", "1"),
    resource.TestCheckResourceAttrSet("aviatrix_account.test", "id"),
)
```

## Code Generation

### Template Structure

The framework uses Go's `text/template` package with the following structure:

```go
type TestFileParams struct {
    ResourceName  string              // Account, Gateway, etc.
    ResourceType  string              // aviatrix_account, etc.
    TestCases     []TestCaseDefinition
    RequiredEnvs  []string           // Required environment variables
    HasPreCheck   bool
}

type TestCaseDefinition struct {
    Name                    string
    Description             string
    Type                    string // crud, import, error, update
    CreateConfig            string
    UpdateConfig            string
    Checks                  TestChecks
    SkipImport              bool
    ImportStateVerifyIgnore []string
    ExpectError             bool
    ErrorPattern            string
}
```

### Generated Test Structure

Each generated test file contains:

1. **Package Declaration and Imports**
2. **Test Functions** - One per test case
3. **Configuration Functions** - Terraform config builders
4. **Helper Functions**:
   - `testAccCheck{ResourceName}Destroy` - Verify resource deletion
   - `testAccCheck{ResourceName}Exists` - Verify resource exists

### Customization Points

Generated tests can be customized by:

1. **Modifying Templates** - Edit template strings in `TestTemplateEngine`
2. **Custom Data Generators** - Implement `TestDataProvider` interface
3. **Custom PreChecks** - Add resource-specific validation
4. **Custom Assertions** - Extend `TestCheckFuncBuilder`

## Best Practices

### 1. Environment Variables

Always use environment variables for sensitive data:

```go
// Good
aws_account_number = os.Getenv("AWS_ACCOUNT_NUMBER")

// Bad - Never hardcode
aws_account_number = "123456789012"
```

### 2. PreCheck Functions

Use comprehensive pre-checks to fail fast:

```go
func preAccountCheck(t *testing.T) {
    testAccPreCheck(t)

    config := TestPreCheckConfig{
        RequireAWS: true,
        CustomEnvVars: []string{"AWS_VPC_ID", "AWS_SUBNET"},
    }

    EnhancedPreCheck(t, config)
}
```

### 3. Resource Cleanup

Always implement proper cleanup:

```go
func testAccCheckAccountDestroy(s *terraform.State) error {
    client := testAccProvider.Meta().(*goaviatrix.Client)

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "aviatrix_account" {
            continue
        }

        // Verify resource is deleted
        _, err := client.GetAccount(&goaviatrix.Account{
            AccountName: rs.Primary.ID,
        })

        if err != goaviatrix.ErrNotFound {
            return fmt.Errorf("Account still exists")
        }
    }

    return nil
}
```

### 4. Import State Verification

Exclude sensitive and computed fields from import verification:

```go
ImportStateVerifyIgnore: []string{
    "aws_secret_key",
    "aws_access_key",
    "audit_account",  // Computed field
}
```

### 5. Test Data Uniqueness

Use unique identifiers to avoid conflicts:

```go
func TestAccAviatrixAccount_basic(t *testing.T) {
    rName := acctest.RandString(5)
    accountName := fmt.Sprintf("tf-acc-%s", rName)

    // Use accountName in config
}
```

### 6. Parallel Testing

Enable parallel execution for independent tests:

```go
func TestAccAviatrixAccount_basic(t *testing.T) {
    t.Parallel()  // Only for independent tests

    // Test implementation
}
```

## Examples

### Example 1: Simple Resource Test

```go
func TestAccAviatrixAccount_basic(t *testing.T) {
    resourceName := "aviatrix_account.test"
    rName := acctest.RandString(5)

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t); preAccountCheck(t, "") },
        ProviderFactories: GetTestProviderFactories(),
        CheckDestroy:      testAccCheckAccountDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccAccountConfigBasic(rName),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckAccountExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tf-acc-%s", rName)),
                    resource.TestCheckResourceAttr(resourceName, "cloud_type", "1"),
                ),
            },
            {
                ResourceName:            resourceName,
                ImportState:             true,
                ImportStateVerify:       true,
                ImportStateVerifyIgnore: []string{"aws_secret_key", "aws_access_key"},
            },
        },
    })
}
```

### Example 2: Using Integration Test Framework

```go
func TestIntegrationFramework_Account_CRUD(t *testing.T) {
    framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

    testData := framework.testDataProvider.GenerateAccountData(1)
    resourceName := "aviatrix_account.test"

    createConfig := fmt.Sprintf(`
resource "aviatrix_account" "test" {
  account_name       = "%s"
  cloud_type         = 1
  aws_account_number = "%s"
  aws_iam            = false
  aws_access_key     = "%s"
  aws_secret_key     = "%s"
}
    `, testData["account_name"],
        os.Getenv("AWS_ACCOUNT_NUMBER"),
        os.Getenv("AWS_ACCESS_KEY"),
        os.Getenv("AWS_SECRET_KEY"))

    crudConfig := CRUDTestConfig{
        ResourceName:            resourceName,
        PreCheck:                func() { testAccPreCheck(t); preAccountCheck(t, "") },
        CreateConfig:            createConfig,
        ImportStateVerifyIgnore: []string{"aws_secret_key", "aws_access_key"},
        CheckDestroy:            testAccCheckAccountDestroy,
        CreateChecks: []resource.TestCheckFunc{
            resource.TestCheckResourceAttr(resourceName, "account_name", testData["account_name"].(string)),
        },
    }

    testFunc := framework.GenerateCRUDTest(crudConfig)
    testFunc(t)
}
```

### Example 3: Error Handling Test

```go
func TestAccAviatrixAccount_missingRequiredField(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: GetTestProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: `
resource "aviatrix_account" "test" {
  cloud_type = 1
  // Missing required account_name
}
                `,
                ExpectError: regexp.MustCompile("account_name"),
            },
        },
    })
}
```

### Example 4: Using Test Step Builder

```go
func TestAccAviatrixAccount_withBuilder(t *testing.T) {
    resourceName := "aviatrix_account.test"
    rName := acctest.RandString(5)

    steps := NewTestStepBuilder().
        AddCreate(
            testAccAccountConfigBasic(rName),
            testAccCheckAccountExists(resourceName),
            resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tf-acc-%s", rName)),
        ).
        AddImport(resourceName, "aws_secret_key", "aws_access_key").
        AddUpdate(
            testAccAccountConfigUpdated(rName),
            resource.TestCheckResourceAttr(resourceName, "audit_account", "false"),
        ).
        Build()

    resource.Test(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: GetTestProviderFactories(),
        CheckDestroy:      testAccCheckAccountDestroy,
        Steps:             steps,
    })
}
```

## Running Generated Tests

### Prerequisites

Set required environment variables:

```bash
# Controller connection
export AVIATRIX_CONTROLLER_IP="controller.example.com"
export AVIATRIX_USERNAME="admin"
export AVIATRIX_PASSWORD="password"

# AWS credentials
export AWS_ACCOUNT_NUMBER="123456789012"
export AWS_ACCESS_KEY="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-east-1"
export AWS_VPC_ID="vpc-0123456789abcdef0"
export AWS_SUBNET="10.0.1.0/24"

# Enable acceptance tests
export TF_ACC=1
```

### Running Tests

```bash
# Run all generated tests
go test -v ./aviatrix/...

# Run specific resource tests
go test -v -run TestAccAviatrixAccount ./aviatrix

# Run with timeout
go test -v -timeout 30m -run TestAccAviatrixAccount ./aviatrix

# Run in parallel
go test -v -parallel 4 ./aviatrix

# Skip specific test scenarios
export SKIP_ACCOUNT_AWS=yes
go test -v -run TestAccAviatrixAccount ./aviatrix
```

## Troubleshooting

### Common Issues

#### 1. Missing Environment Variables

**Error**: `AWS_ACCOUNT_NUMBER must be set for aws acceptance tests`

**Solution**: Export all required environment variables before running tests

#### 2. Import State Verification Failures

**Error**: `ImportStateVerifyIgnore attribute mismatch`

**Solution**: Add computed or sensitive fields to `ImportStateVerifyIgnore`:

```go
ImportStateVerifyIgnore: []string{"aws_secret_key", "computed_field"}
```

#### 3. Resource Not Destroyed

**Error**: `Resource still exists after destroy`

**Solution**: Implement proper CheckDestroy function:

```go
func testAccCheckAccountDestroy(s *terraform.State) error {
    // Properly verify resource deletion
    client := testAccProvider.Meta().(*goaviatrix.Client)

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "aviatrix_account" {
            continue
        }

        _, err := client.GetAccount(&goaviatrix.Account{
            AccountName: rs.Primary.ID,
        })

        if err == nil {
            return fmt.Errorf("Account %s still exists", rs.Primary.ID)
        }
    }

    return nil
}
```

#### 4. Test Data Conflicts

**Error**: `Resource already exists`

**Solution**: Use unique identifiers:

```go
rName := acctest.RandString(5)
accountName := fmt.Sprintf("tf-acc-%s-%d", rName, time.Now().Unix())
```

## Advanced Topics

### Custom Test Data Providers

Implement custom data generation:

```go
type CustomTestDataProvider struct {
    *DefaultTestDataProvider
}

func (p *CustomTestDataProvider) GenerateAccountData(cloudType int) map[string]interface{} {
    // Custom implementation
    data := p.DefaultTestDataProvider.GenerateAccountData(cloudType)
    data["custom_field"] = "custom_value"
    return data
}
```

### Extending the Template Engine

Add custom templates:

```go
engine := NewTestTemplateEngine()
engine.AddTemplate("custom_test", `
func TestCustom{{.ResourceName}}(t *testing.T) {
    // Custom test implementation
}
`)
```

### Batch Test Generation

Generate tests for multiple resources:

```go
resources := []string{
    "aviatrix_account",
    "aviatrix_gateway",
    "aviatrix_vpc",
}

for _, resourceType := range resources {
    if err := GenerateTestForResourceType(resourceType, "./aviatrix"); err != nil {
        log.Printf("Failed to generate test for %s: %v", resourceType, err)
    }
}
```

## Contributing

When adding new features to the test framework:

1. Add unit tests in `*_test.go` files
2. Update this documentation
3. Add examples for new functionality
4. Ensure backward compatibility
5. Run all tests before submitting

## Reference

### Files

- `resource_test_template_generator.go` - Main template generator
- `resource_schema_analyzer.go` - Schema analysis
- `test_generator_cli.go` - CLI tool
- `test_helpers_enhanced.go` - Enhanced helpers
- `integration_test_framework.go` - Integration framework
- `integration_test_data_generator.go` - Data generation
- `integration_test_example_test.go` - Usage examples

### Environment Variables

See [test_helpers_enhanced.go:14](aviatrix/test_helpers_enhanced.go:14) for complete list of supported environment variables.

### Test Naming Conventions

- Basic CRUD: `TestAccAviatrix{Resource}_basic`
- Import: `TestAccAviatrix{Resource}_import`
- Error: `TestAccAviatrix{Resource}_missingRequired_{field}`
- Update: `TestAccAviatrix{Resource}_update_{field}`
- Dependency: `TestAccAviatrix{Resource}_dependsOn{Dependency}`

---

**Version**: 1.0.0
**Last Updated**: 2025-10-05
**Maintainer**: Aviatrix Provider Team
