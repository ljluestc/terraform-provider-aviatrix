# Integration Test Framework Guide

## Overview

The Integration Test Framework provides comprehensive testing capabilities for all 282 Terraform resources in the Aviatrix provider. It builds upon the Unit Test Framework (Task 41) and Test Infrastructure (Task 42) to enable automated CRUD operations, dependency testing, error handling validation, import testing, and state drift detection.

## Architecture

### Components

1. **IntegrationTestFramework** (`integration_test_framework.go`)
   - Core framework for generating and executing integration tests
   - CRUD test generation
   - Error handling test generation
   - Dependency testing
   - State drift detection
   - Import functionality testing

2. **TestDataProvider** (`integration_test_data_generator.go`)
   - Synthetic test data generation
   - Cloud provider-specific data templates
   - Test configuration builders
   - Template engine for Terraform configurations

3. **Supporting Utilities**
   - ResourceDependencyGraph: Manage resource relationships
   - TestCheckFuncBuilder: Build complex test assertions
   - RetryableOperation: Handle flaky operations
   - TestFixtureManager: Manage test fixtures and cleanup

## Quick Start

### Basic CRUD Test

```go
func TestAccAviatrixAccount_CRUD(t *testing.T) {
    // Initialize framework
    framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

    // Generate test data
    testData := framework.testDataProvider.GenerateAccountData(1) // AWS

    // Configure CRUD test
    crudConfig := CRUDTestConfig{
        ResourceName:     "aviatrix_account.test",
        PreCheck:         func() { testAccPreCheck(t) },
        CreateConfig:     buildCreateConfig(testData),
        UpdateConfig:     buildUpdateConfig(testData),
        CreateChecks:     buildCreateChecks(),
        UpdateChecks:     buildUpdateChecks(),
        CheckDestroy:     testAccCheckAccountDestroy,
    }

    // Run test
    testFunc := framework.GenerateCRUDTest(crudConfig)
    testFunc(t)
}
```

### Error Handling Test

```go
func TestAccAviatrixAccount_ErrorHandling(t *testing.T) {
    framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

    errorConfigs := []ErrorHandlingTestConfig{
        {
            Name: "missing_required_field",
            Config: `
resource "aviatrix_account" "test" {
  cloud_type = 1
}
            `,
            ExpectError:       true,
            ErrorMessageMatch: "account_name",
            PreCheck:          func() { testAccPreCheck(t) },
        },
    }

    testFunc := framework.GenerateErrorHandlingTests(errorConfigs)
    testFunc(t)
}
```

### Dependency Test

```go
func TestAccAviatrixGateway_DependsOnAccount(t *testing.T) {
    framework := NewIntegrationTestFramework(t, "aviatrix_gateway", resourceAviatrixGateway())

    depConfig := DependencyTestConfig{
        ResourceName:     "aviatrix_gateway.test",
        DependentConfig:  gatewayConfig,
        DependencyConfig: accountConfig,
        Checks:           buildDependencyChecks(),
        PreCheck:         func() { testAccPreCheck(t) },
    }

    testFunc := framework.GenerateDependencyTest(depConfig)
    testFunc(t)
}
```

### Import Test

```go
func TestAccAviatrixAccount_Import(t *testing.T) {
    framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

    importConfig := ImportTestConfig{
        ResourceName:                "aviatrix_account.test",
        Config:                      config,
        ImportID:                    accountName,
        ImportStateVerifyIgnore:     []string{"aws_secret_key"},
        PreCheck:                    func() { testAccPreCheck(t) },
        CheckDestroy:                testAccCheckAccountDestroy,
    }

    testFunc := framework.GenerateImportTest(importConfig)
    testFunc(t)
}
```

### State Drift Test

```go
func TestAccAviatrixAccount_StateDrift(t *testing.T) {
    framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

    driftConfig := StateDriftTestConfig{
        ResourceName:       "aviatrix_account.test",
        InitialConfig:      initialConfig,
        ModifyResourceFunc: modifyAccountExternally,
        ExpectedChanges:    []string{"audit_account"},
        PreCheck:           func() { testAccPreCheck(t) },
    }

    testFunc := framework.GenerateStateDriftTest(driftConfig)
    testFunc(t)
}
```

## Test Data Generation

### Using TestDataProvider

```go
provider := NewTestDataProvider()

// Generate AWS account data
awsAccount := provider.GenerateAccountData(1)

// Generate GCP VPC data
gcpVPC := provider.GenerateVPCData(4)

// Generate Azure gateway data
azureGateway := provider.GenerateGatewayData(8)

// Generate custom resource name
resourceName := provider.GenerateResourceName("tf-test")
```

### Cloud Type Constants

- `1` - AWS
- `4` - GCP
- `8` - Azure
- `16` - OCI
- `256` - AWS GovCloud
- `512` - Azure GovCloud
- `1024` - AWS China
- `2048` - Azure China

### Using TestConfigBuilder

```go
builder := NewTestConfigBuilder()

config := builder.
    AddResourceBlock("aviatrix_account", "test", map[string]interface{}{
        "account_name": "test-account",
        "cloud_type":   1,
    }).
    AddResourceBlock("aviatrix_vpc", "test", map[string]interface{}{
        "cloud_type":   1,
        "account_name": "aviatrix_account.test.account_name",
        "region":       "us-east-1",
    }).
    Build()
```

### Using Template Engine

```go
engine := NewTestConfigTemplateEngine()

// Add custom template
engine.AddTemplate("my_resource", `
resource "aviatrix_{{.ResourceType}}" "{{.Name}}" {
  name = "{{.ResourceName}}"
  type = "{{.Type}}"
}
`)

// Render template
config, err := engine.RenderTemplate("my_resource", map[string]interface{}{
    "ResourceType": "account",
    "Name":         "test",
    "ResourceName": "test-account",
    "Type":         "aws",
})
```

## Building Test Checks

### Using TestCheckFuncBuilder

```go
checks := NewTestCheckFuncBuilder().
    AddResourceAttr("aviatrix_account.test", "account_name", "test-acc").
    AddResourceAttr("aviatrix_account.test", "cloud_type", "1").
    AddResourceAttrSet("aviatrix_account.test", "id").
    AddResourceAttrPair(
        "aviatrix_gateway.test", "account_name",
        "aviatrix_account.test", "account_name",
    ).
    AddCustomCheck(customCheckFunction).
    Build()
```

### Common Test Check Functions

```go
// Check resource exists
ResourceExists("aviatrix_account.test")

// Check resource does not exist (for destroy)
ResourceDoesNotExist("aviatrix_account.test")

// Standard SDK checks
resource.TestCheckResourceAttr(name, key, value)
resource.TestCheckResourceAttrSet(name, key)
resource.TestCheckResourceAttrPair(name1, key1, name2, key2)
```

## Dependency Management

### Using ResourceDependencyGraph

```go
graph := NewResourceDependencyGraph()

// Define dependencies
graph.AddDependency("aviatrix_gateway.test", "aviatrix_account.test")
graph.AddDependency("aviatrix_gateway.test", "aviatrix_vpc.test")
graph.AddDependency("aviatrix_vpc.test", "aviatrix_account.test")

// Get dependencies for a resource
deps := graph.GetDependencies("aviatrix_gateway.test")

// Validate creation order
resourceOrder := []string{
    "aviatrix_account.test",
    "aviatrix_vpc.test",
    "aviatrix_gateway.test",
}

err := graph.ValidateDependencyOrder(resourceOrder)
```

## Test Fixtures

### Using TestFixtureManager

```go
fixtureManager := NewTestFixtureManager()

// Register fixture with cleanup
fixtureManager.RegisterFixture("test_account", accountData, func() error {
    // Cleanup logic
    return deleteAccount(accountData)
})

// Retrieve fixture
account, ok := fixtureManager.GetFixture("test_account")

// Cleanup all fixtures (deferred)
defer func() {
    errors := fixtureManager.Cleanup()
    for _, err := range errors {
        t.Errorf("Cleanup error: %v", err)
    }
}()
```

## Retryable Operations

### Using RetryableOperation

```go
retryOp := NewRetryableOperation(3, 5*time.Second)

err := retryOp.Execute(context.Background(), func() error {
    // Operation that might fail transiently
    return checkResourceStatus()
})
```

## Complete Test Example

### Account Resource CRUD Test

```go
func TestAccAviatrixAccount_Complete(t *testing.T) {
    // Skip conditions
    if os.Getenv("SKIP_ACCOUNT") == "yes" {
        t.Skip("Skipping account test")
    }

    // Initialize framework
    framework := NewIntegrationTestFramework(t, "aviatrix_account", resourceAviatrixAccount())

    // Generate test data
    testData := framework.testDataProvider.GenerateAccountData(1)
    accountName := testData["account_name"].(string)
    resourceName := "aviatrix_account.test"

    // Build configurations
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

    // Build checks
    createChecks := NewTestCheckFuncBuilder().
        AddResourceAttr(resourceName, "account_name", accountName).
        AddResourceAttr(resourceName, "cloud_type", "1").
        AddResourceAttrSet(resourceName, "aws_account_number").
        AddCustomCheck(ResourceExists(resourceName)).
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
        SkipImport:                  false,
        SkipUpdate:                  false,
    }

    // Run test
    testFunc := framework.GenerateCRUDTest(crudConfig)
    testFunc(t)
}
```

## Best Practices

### 1. Test Organization

```go
// Group related tests
func TestAccAviatrixAccount_AWS(t *testing.T) { /* AWS-specific tests */ }
func TestAccAviatrixAccount_GCP(t *testing.T) { /* GCP-specific tests */ }
func TestAccAviatrixAccount_Azure(t *testing.T) { /* Azure-specific tests */ }

// Separate error handling
func TestAccAviatrixAccount_ErrorHandling(t *testing.T) { /* Error tests */ }

// Separate import tests
func TestAccAviatrixAccount_Import(t *testing.T) { /* Import tests */ }
```

### 2. Use Skip Conditions

```go
if os.Getenv("SKIP_ACCOUNT_AWS") == "yes" {
    t.Skip("Skipping AWS account tests")
}

if !IsAcceptanceTest() {
    t.Skip("Skipping acceptance test (TF_ACC not set)")
}
```

### 3. Clean Test Data

```go
// Use unique names
accountName := framework.testDataProvider.GenerateResourceName("tf-acc")

// Verify cleanup
defer func() {
    errors := fixtureManager.Cleanup()
    if len(errors) > 0 {
        t.Logf("Cleanup errors: %v", errors)
    }
}()
```

### 4. Validate Configurations

```go
validator := NewTestDataValidator()

if err := validator.ValidateCIDR("10.0.0.0/16"); err != nil {
    t.Fatalf("Invalid CIDR: %v", err)
}

if err := validator.ValidateAccountName(accountName); err != nil {
    t.Fatalf("Invalid account name: %v", err)
}
```

### 5. Handle Provider-Specific Logic

```go
var config string

switch cloudType {
case 1: // AWS
    config = buildAWSConfig(testData)
case 4: // GCP
    config = buildGCPConfig(testData)
case 8: // Azure
    config = buildAzureConfig(testData)
default:
    t.Fatalf("Unsupported cloud type: %d", cloudType)
}
```

## Environment Variables

### Required

- `TF_ACC=1` - Enable acceptance tests
- `AVIATRIX_CONTROLLER_IP` - Controller IP address
- `AVIATRIX_USERNAME` - Controller username
- `AVIATRIX_PASSWORD` - Controller password

### AWS

- `AWS_ACCOUNT_NUMBER`
- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`
- `AWS_REGION`
- `AWS_VPC_ID` (for gateway tests)
- `AWS_SUBNET` (for gateway tests)

### GCP

- `GCP_ID` - Project ID
- `GCP_CREDENTIALS_FILEPATH`
- `GCP_ZONE`
- `GCP_VPC_ID` (for gateway tests)
- `GCP_SUBNET` (for gateway tests)

### Azure

- `ARM_SUBSCRIPTION_ID`
- `ARM_DIRECTORY_ID`
- `ARM_APPLICATION_ID`
- `ARM_APPLICATION_KEY`
- `AZURE_REGION`
- `AZURE_VNET_ID` (for gateway tests)
- `AZURE_SUBNET` (for gateway tests)

### OCI

- `OCI_TENANCY_ID`
- `OCI_USER_ID`
- `OCI_COMPARTMENT_ID`
- `OCI_API_KEY_FILEPATH`
- `OCI_REGION`

### Skip Flags

- `SKIP_ACCOUNT=yes` - Skip all account tests
- `SKIP_ACCOUNT_AWS=yes` - Skip AWS tests
- `SKIP_ACCOUNT_GCP=yes` - Skip GCP tests
- `SKIP_ACCOUNT_AZURE=yes` - Skip Azure tests
- `SKIP_ACCOUNT_OCI=yes` - Skip OCI tests

## Running Tests

### Run All Integration Tests

```bash
TF_ACC=1 go test -v -run TestIntegration ./aviatrix/
```

### Run Specific Resource Tests

```bash
TF_ACC=1 go test -v -run TestIntegrationFramework_Account ./aviatrix/
```

### Run with Timeout

```bash
TF_ACC=1 go test -v -timeout 30m -run TestIntegration ./aviatrix/
```

### Run in Parallel

```bash
TF_ACC=1 go test -v -parallel 4 -run TestIntegration ./aviatrix/
```

### Skip Provider-Specific Tests

```bash
SKIP_ACCOUNT_AZURE=yes SKIP_ACCOUNT_GCP=yes TF_ACC=1 go test -v -run TestIntegration ./aviatrix/
```

## Troubleshooting

### Issue: Tests Timeout

**Solution:** Increase timeout
```bash
TF_ACC=1 go test -v -timeout 60m -run TestIntegration ./aviatrix/
```

### Issue: Import Fails

**Solution:** Add fields to ImportStateVerifyIgnore
```go
ImportStateVerifyIgnore: []string{"aws_secret_key", "sensitive_field"}
```

### Issue: Flaky Tests

**Solution:** Use RetryableOperation
```go
retryOp := NewRetryableOperation(5, 10*time.Second)
err := retryOp.Execute(ctx, operationFunc)
```

### Issue: Dependency Errors

**Solution:** Verify dependency graph
```go
err := graph.ValidateDependencyOrder(resourceOrder)
if err != nil {
    t.Fatalf("Invalid dependency order: %v", err)
}
```

## Framework Extension

### Adding Custom Test Types

```go
// Add to IntegrationTestFramework
func (f *IntegrationTestFramework) GenerateCustomTest(config CustomTestConfig) func(*testing.T) {
    return func(t *testing.T) {
        // Custom test logic
    }
}
```

### Adding Custom Templates

```go
engine := NewTestConfigTemplateEngine()

engine.AddTemplate("custom_resource", `
resource "aviatrix_custom" "{{.Name}}" {
  // Template content
}
`)
```

### Adding Custom Validators

```go
type CustomValidator struct {
    *TestDataValidator
}

func (v *CustomValidator) ValidateCustomField(value string) error {
    // Custom validation logic
    return nil
}
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
- name: Run Integration Tests
  env:
    TF_ACC: 1
    AVIATRIX_CONTROLLER_IP: ${{ secrets.CONTROLLER_IP }}
    AVIATRIX_USERNAME: ${{ secrets.USERNAME }}
    AVIATRIX_PASSWORD: ${{ secrets.PASSWORD }}
    AWS_ACCESS_KEY: ${{ secrets.AWS_ACCESS_KEY }}
    AWS_SECRET_KEY: ${{ secrets.AWS_SECRET_KEY }}
  run: |
    go test -v -timeout 60m -run TestIntegration ./aviatrix/
```

## Resources

- [Terraform Plugin SDK v2 Testing](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing)
- [Unit Test Framework Guide](./UNIT_TEST_FRAMEWORK_GUIDE.md)
- [Test Infrastructure PRD](../docs/PRD_Task_11_Test_Infrastructure.md)
- [Example Tests](./integration_test_example_test.go)

## Summary

The Integration Test Framework provides:

✅ **Automated CRUD Testing** - Generate complete create, read, update, delete tests
✅ **Error Handling** - Test negative scenarios and edge cases
✅ **Dependency Testing** - Validate resource relationships
✅ **Import Testing** - Verify import functionality works correctly
✅ **State Drift Detection** - Test drift detection and correction
✅ **Test Data Generation** - Synthetic data for all cloud providers
✅ **Template Engine** - Reusable configuration templates
✅ **Fixture Management** - Manage test fixtures and cleanup

This comprehensive framework enables systematic testing of all 282 Terraform resources with minimal boilerplate code.
