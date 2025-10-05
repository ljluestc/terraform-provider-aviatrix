# Unit Test Framework Guide

## Overview

The Unit Test Framework provides comprehensive utilities for testing Terraform resources without requiring live infrastructure. It leverages the Terraform Plugin SDK v2 testing capabilities and includes mock provider interfaces, helper functions, and test templates.

## Core Components

### 1. Resource Unit Test Framework (`resource_unit_test_framework.go`)

**Location**: `aviatrix/resource_unit_test_framework.go`

**Key Features**:
- Mock client builder for creating test doubles
- Resource test data wrappers with assertion helpers
- Test case runners for CRUD operations
- Schema validation helpers
- State management utilities
- Edge case testing helpers

### 2. Test Templates (`resource_unit_test_templates.go`)

**Location**: `aviatrix/resource_unit_test_templates.go`

Provides high-level templates for common testing patterns:
- CRUD operation testing
- Schema validation
- Input validation
- Error handling
- State management

### 3. Test Helpers (`test_helpers.go`)

**Location**: `aviatrix/test_helpers.go`

Contains utility functions for:
- Environment variable management
- Cloud provider configuration
- Test skipping logic
- Retry mechanisms with backoff
- Test logging and progress tracking

## Quick Start

### Basic Test Structure

```go
func TestResourceAviatrixAccount_Delete(t *testing.T) {
    // 1. Create mock client
    mockClient := NewMockClientBuilder().
        WithDeleteAccountFunc(func(account *goaviatrix.Account) error {
            assert.Equal(t, "test-account", account.AccountName)
            return nil
        }).
        Build()

    // 2. Create test resource data
    resource := resourceAviatrixAccount()
    framework := NewResourceUnitTestFramework(t, resource, mockClient)
    testData := framework.NewResourceTestData(map[string]interface{}{
        "account_name": "test-account",
    })

    // 3. Execute operation
    diags := resourceAviatrixAccountDelete(context.Background(), testData.ResourceData, mockClient)

    // 4. Assert results
    testData.AssertNoError(diags)
}
```

### Using Test Templates

```go
func TestResourceAviatrixAccount_Comprehensive(t *testing.T) {
    resource := resourceAviatrixAccount()
    template := NewCRUDTestTemplate(t, "aviatrix_account", resource)

    // Test Delete operation
    template.TestDelete(DeleteTestConfig{
        TestName: "DeleteSuccess",
        ResourceData: map[string]interface{}{
            "account_name": "test-account",
        },
        SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
            mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
                return nil
            }
        },
    })
}
```

## Framework Components Reference

### MockClientBuilder

Build mock clients with specific behaviors:

```go
mockClient := NewMockClientBuilder().
    WithGetAccountFunc(func(account *goaviatrix.Account) (goaviatrix.Account, error) {
        return goaviatrix.Account{
            AccountName: "test-account",
            CloudType:   1,
        }, nil
    }).
    WithDeleteAccountFunc(func(account *goaviatrix.Account) error {
        return nil
    }).
    Build()
```

**Available Methods**:
- `WithGetAccountFunc(fn)` - Mock GetAccount operations
- `WithDeleteAccountFunc(fn)` - Mock DeleteAccount operations
- `WithAuditAccountFunc(fn)` - Mock AuditAccount operations
- `Build()` - Return configured mock client

### ResourceTestData

Wrapper around `schema.ResourceData` with assertion helpers:

```go
testData := framework.NewResourceTestData(map[string]interface{}{
    "account_name": "test-account",
    "cloud_type":   1,
})

// Assert attribute values
testData.AssertAttribute("account_name", "test-account")
testData.AssertAttributeSet("cloud_type")
testData.AssertAttributeNotSet("optional_field")

// Assert no errors in diagnostics
testData.AssertNoError(diags)
```

### Schema Validation Helpers

Test schema field properties:

```go
helper := NewSchemaFieldTestHelper(t, resource.Schema)

// Test field existence and properties
helper.AssertFieldExists("account_name")
helper.AssertFieldRequired("account_name")
helper.AssertFieldOptional("aws_iam")
helper.AssertFieldType("account_name", schema.TypeString)
helper.AssertFieldDefault("aws_iam", false)
```

### State Management Helpers

Test resource state changes:

```go
helper := NewResourceStateTestHelper(t, resourceData)

// Set and verify values
helper.SetAndVerify("aws_iam", true)

// Check for changes
helper.AssertHasChange("aws_iam")
helper.AssertNoChange("account_name")

// Get old and new values
old, new := helper.GetOldNew("aws_iam")
```

### Edge Case Testing

Test boundary conditions:

```go
helper := NewEdgeCaseTestHelper(t)

// Test empty string handling
helper.TestEmptyString(func(s string) error {
    return validateAccountName(s)
}, true) // expect error

// Test nil values
helper.TestNilValue(func(v interface{}) error {
    return validateValue(v)
}, true) // expect error

// Test max length
helper.TestMaxLength(func(s string) error {
    return validateName(s)
}, 100, true) // expect error for > 100 chars
```

## Test Templates

### CRUD Test Template

```go
template := NewCRUDTestTemplate(t, "aviatrix_account", resource)

// Test Read
template.TestRead(ReadTestConfig{
    TestName: "ReadSuccess",
    ResourceData: map[string]interface{}{
        "account_name": "test-account",
    },
    SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
        mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
            return goaviatrix.Account{AccountName: "test-account"}, nil
        }
    },
    ValidateState: func(t *testing.T, rd *schema.ResourceData) {
        assert.Equal(t, "test-account", rd.Get("account_name"))
    },
})

// Test Delete
template.TestDelete(DeleteTestConfig{
    TestName: "DeleteSuccess",
    ResourceData: map[string]interface{}{
        "account_name": "test-account",
    },
    SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
        mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
            return nil
        }
    },
})
```

### Schema Validation Template

```go
template := NewSchemaValidationTemplate(t, resource)

// Test field properties
template.RequiredFieldTest("account_name")
template.OptionalFieldTest("aws_iam")
template.ComputedFieldTest("rbac_groups")
template.FieldTypeTest("account_name", schema.TypeString)
template.DefaultValueTest("aws_iam", false)
```

### Input Validation Template

```go
template := NewInputValidationTemplate(t)

template.TestValidator(validateAwsAccountNumber, []ValidatorTest{
    {
        Name:        "ValidAccountNumber",
        Value:       "123456789012",
        Key:         "aws_account_number",
        ExpectError: false,
    },
    {
        Name:          "InvalidAccountNumber",
        Value:         "12345",
        Key:           "aws_account_number",
        ExpectError:   true,
        ErrorContains: "must be 12 digits",
    },
})
```

### Error Handling Template

```go
template := NewErrorHandlingTemplate(t, "aviatrix_account", resource)

// Test API errors
template.APIErrorTest("Delete", func(mock *goaviatrix.ClientInterfaceMock) {
    mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
        return errors.New("API error")
    }
}, "failed to delete")

// Test not found errors
template.NotFoundErrorTest(func(mock *goaviatrix.ClientInterfaceMock) {
    mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
        return goaviatrix.Account{}, errors.New("not found")
    }
})
```

### State Management Template

```go
template := NewStateManagementTemplate(t, resource)

// Test state updates
template.TestStateUpdate(
    map[string]interface{}{"account_name": "test", "cloud_type": 1},
    map[string]interface{}{"aws_iam": true},
)

// Test change detection
template.TestStateChange(
    map[string]interface{}{"aws_iam": false},
    map[string]interface{}{"aws_iam": true},
    []string{"aws_iam"},
)
```

## Running Tests

### Run All Unit Tests

```bash
TF_ACC=0 go test -v ./aviatrix -run "Test" -timeout 30s
```

### Run Specific Test Pattern

```bash
TF_ACC=0 go test -v ./aviatrix -run "TestResourceAviatrixAccount" -timeout 30s
```

### Run Framework Smoke Tests

```bash
TF_ACC=0 go test -v ./aviatrix -run "TestSmoke" -timeout 10s
```

### Run With Coverage

```bash
TF_ACC=0 go test -v ./aviatrix -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Best Practices

### 1. Test Naming Conventions

- **Unit tests**: `TestResourceAviatrix<Resource>_<Operation>`
- **Schema tests**: `TestSchemaValidation_<Resource>`
- **Input tests**: `TestInputValidation_<Resource>`
- **Helper tests**: `TestHelpers_<Component>`

### 2. Mock Setup

Always configure mocks to:
- Verify input parameters
- Return realistic data
- Simulate both success and error cases

```go
mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
    // Verify input
    assert.Equal(t, "expected-name", account.AccountName)

    // Return appropriate result
    return nil // or return errors.New("failure")
}
```

### 3. Test Organization

Group related tests using subtests:

```go
func TestResourceAviatrixAccount_Delete(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        // Success case
    })

    t.Run("NotFound", func(t *testing.T) {
        // Not found case
    })

    t.Run("APIError", func(t *testing.T) {
        // API error case
    })
}
```

### 4. Assertion Clarity

Use descriptive assertion messages:

```go
// Good
assert.Equal(t, expected, actual, "Account name should match input")

// Better - use helpers
testData.AssertAttribute("account_name", "expected-value")
```

### 5. Environment Variables

Skip tests based on configuration:

```go
func TestResourceAviatrixAccount_AWS(t *testing.T) {
    SkipUnlessCloudProvider(t, "AWS")
    // Test implementation
}
```

## Example Test Files

### Complete Resource Test Example

See `resource_aviatrix_account_unit_test.go` for a comprehensive example including:
- Field validation tests
- Delete operation tests with mocks
- Read operation tests with audit
- Error handling

### Framework Usage Examples

See `resource_unit_test_example_test.go` for:
- CRUD template usage
- Schema validation
- Input validation
- Error handling patterns
- State management testing

## Framework Extension

### Adding New Mock Functions

To add support for additional client methods:

1. Check if the method exists in `goaviatrix.ClientInterfaceMock`
2. Add builder method to `MockClientBuilder`:

```go
func (b *MockClientBuilder) WithCustomFunc(fn func(*goaviatrix.CustomType) error) *MockClientBuilder {
    b.mock.CustomFunc = fn
    return b
}
```

### Creating Custom Test Helpers

Extend the framework with domain-specific helpers:

```go
type CustomTestHelper struct {
    t *testing.T
}

func NewCustomTestHelper(t *testing.T) *CustomTestHelper {
    return &CustomTestHelper{t: t}
}

func (h *CustomTestHelper) AssertCustomBehavior(data interface{}) {
    // Custom assertions
}
```

## Troubleshooting

### Common Issues

**Issue**: Mock function not called
- **Solution**: Verify mock is passed to resource function, not the nil client

**Issue**: Assertion fails with unexpected value
- **Solution**: Check if field uses computed values or default functions

**Issue**: Test skipped unexpectedly
- **Solution**: Check environment variables like `TF_ACC`, `SKIP_ACCOUNT_*`

### Debug Tips

```go
// Print resource data for debugging
t.Logf("Resource data: %+v", resourceData.State())

// Print mock call counts
mock := mockClient.(*goaviatrix.ClientInterfaceMock)
t.Logf("DeleteAccount called %d times", len(mock.DeleteAccountCalls()))
```

## Additional Resources

- **Terraform Plugin SDK v2 Docs**: https://developer.hashicorp.com/terraform/plugin/sdkv2
- **Testing Guide**: https://developer.hashicorp.com/terraform/plugin/sdkv2/testing
- **Testify Assertions**: https://pkg.go.dev/github.com/stretchr/testify/assert
- **Go Testing**: https://golang.org/pkg/testing/

## Summary

The Unit Test Framework provides:
- ✅ Mock provider interface for testing without infrastructure
- ✅ Comprehensive test templates for common patterns
- ✅ Helper functions for assertions and validations
- ✅ State management testing utilities
- ✅ Schema validation helpers
- ✅ Edge case testing support
- ✅ Integration with existing test helpers

Use this framework to achieve high test coverage while maintaining fast, reliable tests that don't require live infrastructure or API credentials.
