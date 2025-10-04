# Resource Unit Test Framework and Templates

## Overview

This unit test framework provides comprehensive templates and utilities for testing Terraform resource CRUD operations, schema validation, error handling, and edge cases. The framework uses the Go testing package with Terraform Plugin SDK v2 and testify/assert for assertions.

## Framework Components

### Core Files

1. **resource_unit_test_framework.go** - Core framework with helper classes
2. **resource_unit_test_templates.go** - Reusable test templates for CRUD operations
3. **resource_unit_test_example_test.go** - Complete examples demonstrating usage

## Quick Start

### 1. Basic CRUD Testing

```go
func TestResourceAccount_CRUD(t *testing.T) {
    resource := resourceAviatrixAccount()
    template := NewCRUDTestTemplate(t, "aviatrix_account", resource)

    // Test Create
    template.TestCreate(CreateTestConfig{
        TestName: "CreateSuccess",
        ResourceData: map[string]interface{}{
            "account_name": "test-account",
            "cloud_type":   1,
        },
        SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
            mock.CreateAccountFunc = func(account *goaviatrix.Account) error {
                return nil
            }
        },
        ValidateState: func(t *testing.T, rd *schema.ResourceData) {
            assert.Equal(t, "test-account", rd.Get("account_name"))
        },
    })

    // Test Read
    template.TestRead(ReadTestConfig{
        TestName: "ReadSuccess",
        ResourceData: map[string]interface{}{
            "account_name": "test-account",
        },
        SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
            mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
                return goaviatrix.Account{
                    AccountName: "test-account",
                }, nil
            }
        },
    })

    // Test Update
    template.TestUpdate(UpdateTestConfig{
        TestName: "UpdateSuccess",
        InitialData: map[string]interface{}{
            "account_name": "test-account",
            "aws_iam":      false,
        },
        UpdatedData: map[string]interface{}{
            "aws_iam": true,
        },
        SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
            mock.UpdateAccountFunc = func(account *goaviatrix.Account) error {
                return nil
            }
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

    // Test Import
    template.TestImport(ImportTestConfig{
        TestName: "ImportSuccess",
        ImportID: "test-account",
        SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
            mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
                return goaviatrix.Account{
                    AccountName: "test-account",
                }, nil
            }
        },
    })
}
```

### 2. Schema Validation Testing

```go
func TestResourceAccount_SchemaValidation(t *testing.T) {
    resource := resourceAviatrixAccount()
    template := NewSchemaValidationTemplate(t, resource)

    // Test required fields
    template.RequiredFieldTest("account_name")
    template.RequiredFieldTest("cloud_type")

    // Test optional fields
    template.OptionalFieldTest("aws_iam")
    template.OptionalFieldTest("aws_role_arn")

    // Test computed fields
    template.ComputedFieldTest("rbac_groups")

    // Test field types
    template.FieldTypeTest("account_name", schema.TypeString)
    template.FieldTypeTest("cloud_type", schema.TypeInt)
    template.FieldTypeTest("aws_iam", schema.TypeBool)

    // Test default values
    template.DefaultValueTest("aws_iam", false)

    // Test ConflictsWith
    template.ConflictsWithTest("aws_account_number", []string{"aws_role_arn"})
}
```

### 3. Input Validation Testing

```go
func TestResourceAccount_InputValidation(t *testing.T) {
    template := NewInputValidationTemplate(t)

    // Test custom validator
    template.TestValidator(validateAwsAccountNumber, []ValidatorTest{
        {
            Name:        "ValidAccountNumber",
            Value:       "123456789012",
            Key:         "aws_account_number",
            ExpectError: false,
        },
        {
            Name:          "InvalidAccountNumber_TooShort",
            Value:         "12345678901",
            Key:           "aws_account_number",
            ExpectError:   true,
            ErrorContains: "must be 12 digits",
        },
        {
            Name:          "InvalidAccountNumber_NonNumeric",
            Value:         "12345678901A",
            Key:           "aws_account_number",
            ExpectError:   true,
            ErrorContains: "must be 12 digits",
        },
    })
}
```

### 4. Error Handling Testing

```go
func TestResourceAccount_ErrorHandling(t *testing.T) {
    resource := resourceAviatrixAccount()
    template := NewErrorHandlingTemplate(t, "aviatrix_account", resource)

    // Test API error during Create
    template.APIErrorTest("Create", func(mock *goaviatrix.ClientInterfaceMock) {
        mock.CreateAccountFunc = func(account *goaviatrix.Account) error {
            return errors.New("API connection timeout")
        }
    }, "failed to create")

    // Test API error during Delete
    template.APIErrorTest("Delete", func(mock *goaviatrix.ClientInterfaceMock) {
        mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
            return errors.New("account in use")
        }
    }, "failed to delete")

    // Test not found error during Read
    template.NotFoundErrorTest(func(mock *goaviatrix.ClientInterfaceMock) {
        mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
            return goaviatrix.Account{}, errors.New("account not found")
        }
    })
}
```

### 5. State Management Testing

```go
func TestResourceAccount_StateManagement(t *testing.T) {
    resource := resourceAviatrixAccount()
    template := NewStateManagementTemplate(t, resource)

    // Test state updates
    template.TestStateUpdate(
        map[string]interface{}{
            "account_name": "test-account",
            "cloud_type":   1,
        },
        map[string]interface{}{
            "aws_iam": true,
        },
    )

    // Test change detection
    template.TestStateChange(
        map[string]interface{}{
            "account_name": "test-account",
            "aws_iam":      false,
        },
        map[string]interface{}{
            "aws_iam": true,
        },
        []string{"aws_iam"}, // Expected changed fields
    )
}
```

## Advanced Usage

### Using Helper Classes

#### SchemaFieldTestHelper

```go
func TestResourceSchema(t *testing.T) {
    resource := resourceAviatrixAccount()
    helper := NewSchemaFieldTestHelper(t, resource.Schema)

    // Test field properties
    helper.AssertFieldExists("account_name")
    helper.AssertFieldRequired("account_name")
    helper.AssertFieldOptional("aws_iam")
    helper.AssertFieldType("account_name", schema.TypeString)
    helper.AssertFieldDefault("aws_iam", false)
    helper.AssertFieldConflictsWith("aws_account_number", []string{"aws_role_arn"})
}
```

#### ResourceStateTestHelper

```go
func TestResourceState(t *testing.T) {
    resource := resourceAviatrixAccount()
    rd := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
        "account_name": "test-account",
    })

    helper := NewResourceStateTestHelper(t, rd)

    // Set and verify values
    helper.SetAndVerify("aws_iam", true)

    // Check for changes
    helper.AssertHasChange("aws_iam")
    helper.AssertNoChange("account_name")

    // Get old and new values
    old, new := helper.GetOldNew("aws_iam")
    assert.Equal(t, false, old)
    assert.Equal(t, true, new)
}
```

#### EdgeCaseTestHelper

```go
func TestEdgeCases(t *testing.T) {
    helper := NewEdgeCaseTestHelper(t)

    // Test empty string
    helper.TestEmptyString(func(s string) error {
        if s == "" {
            return errors.New("empty not allowed")
        }
        return nil
    }, true) // Should error

    // Test nil values
    helper.TestNilValue(func(v interface{}) error {
        if v == nil {
            return errors.New("nil not allowed")
        }
        return nil
    }, true) // Should error

    // Test maximum length
    helper.TestMaxLength(func(s string) error {
        if len(s) > 100 {
            return errors.New("too long")
        }
        return nil
    }, 100, true) // Should error when exceeding
}
```

### Using Test Case Runner

```go
func TestResourceWithTestCases(t *testing.T) {
    resource := resourceAviatrixAccount()
    framework := NewResourceUnitTestFramework(t, resource, nil)

    testCases := []ResourceTestCase{
        {
            Name: "ValidInput",
            ResourceData: map[string]interface{}{
                "account_name": "test-account",
            },
            SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
                // Setup mock behavior
            },
            Operation: func(ctx context.Context, rd *schema.ResourceData, client goaviatrix.ClientInterface) interface{} {
                return resourceAviatrixAccountCreate(ctx, rd, client)
            },
            ValidateResult: func(t *testing.T, result interface{}, rd *schema.ResourceData) {
                assert.Empty(t, result) // No errors
            },
            ExpectError: false,
        },
        {
            Name: "InvalidInput",
            ResourceData: map[string]interface{}{
                "account_name": "",
            },
            Operation: func(ctx context.Context, rd *schema.ResourceData, client goaviatrix.ClientInterface) interface{} {
                return resourceAviatrixAccountCreate(ctx, rd, client)
            },
            ExpectError:   true,
            ErrorContains: "account_name cannot be empty",
        },
    }

    framework.RunTestCases(testCases)
}
```

## Testing Patterns

### Pattern 1: Comprehensive Resource Test Suite

Create a complete test suite for a resource covering all operations:

```go
func TestResourceAviatrixGateway_Complete(t *testing.T) {
    resource := resourceAviatrixGateway()

    t.Run("CRUD", func(t *testing.T) {
        template := NewCRUDTestTemplate(t, "aviatrix_gateway", resource)
        // Add all CRUD tests
    })

    t.Run("Schema", func(t *testing.T) {
        template := NewSchemaValidationTemplate(t, resource)
        // Add all schema validation tests
    })

    t.Run("Validation", func(t *testing.T) {
        template := NewInputValidationTemplate(t)
        // Add all input validation tests
    })

    t.Run("Errors", func(t *testing.T) {
        template := NewErrorHandlingTemplate(t, "aviatrix_gateway", resource)
        // Add all error handling tests
    })

    t.Run("State", func(t *testing.T) {
        template := NewStateManagementTemplate(t, resource)
        // Add all state management tests
    })
}
```

### Pattern 2: Testing Computed Fields

```go
func TestComputedFields(t *testing.T) {
    resource := resourceAviatrixGateway()
    helper := NewComputedFieldTestHelper(t, resource.Schema)

    testData := map[string]interface{}{
        "gw_name": "test-gateway",
    }

    helper.AssertComputedFieldBehavior("public_ip", testData)
    helper.AssertComputedFieldBehavior("private_ip", testData)
}
```

### Pattern 3: Testing Default Values

```go
func TestDefaultValues(t *testing.T) {
    resource := resourceAviatrixGateway()
    helper := NewDefaultValueTestHelper(t, resource)

    emptyConfig := map[string]interface{}{
        "gw_name": "test-gateway",
        "vpc_id":  "vpc-123",
    }

    helper.AssertDefaultApplied("enable_snat", emptyConfig)
    helper.AssertDefaultApplied("single_az_ha", emptyConfig)
}
```

## Best Practices

### 1. Organize Tests by Functionality

Group related tests together using `t.Run()`:

```go
func TestResourceAviatrixAccount(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        // All create tests
    })

    t.Run("Read", func(t *testing.T) {
        // All read tests
    })

    t.Run("Update", func(t *testing.T) {
        // All update tests
    })
}
```

### 2. Use Descriptive Test Names

```go
template.TestCreate(CreateTestConfig{
    TestName: "CreateSuccess_WithAWSIAM",
    // ...
})

template.TestCreate(CreateTestConfig{
    TestName: "CreateError_MissingRequiredField",
    // ...
})
```

### 3. Test Both Success and Error Cases

Always test both happy path and error conditions:

```go
// Success case
template.TestCreate(CreateTestConfig{
    TestName: "CreateSuccess",
    // ...
})

// Error case
template.TestCreate(CreateTestConfig{
    TestName: "CreateError_APIFailure",
    ExpectedError: "failed to create",
    // ...
})
```

### 4. Validate State After Operations

Always validate that state is correctly updated:

```go
template.TestCreate(CreateTestConfig{
    TestName: "CreateSuccess",
    ValidateState: func(t *testing.T, rd *schema.ResourceData) {
        assert.Equal(t, "expected-value", rd.Get("field_name"))
        assert.NotEmpty(t, rd.Id())
    },
})
```

### 5. Test Edge Cases

Use the EdgeCaseTestHelper for comprehensive edge case coverage:

```go
helper := NewEdgeCaseTestHelper(t)

// Test empty strings
helper.TestEmptyString(validationFunc, true)

// Test nil values
helper.TestNilValue(validationFunc, true)

// Test max length
helper.TestMaxLength(validationFunc, 256, true)
```

## Running Tests

### Run all unit tests (without TF_ACC):

```bash
TF_ACC=0 go test -v ./aviatrix/...
```

### Run specific test:

```bash
TF_ACC=0 go test -v ./aviatrix/ -run TestResourceAccount_CRUD
```

### Run with coverage:

```bash
TF_ACC=0 go test -v -cover ./aviatrix/...
```

### Run with race detection:

```bash
TF_ACC=0 go test -v -race ./aviatrix/...
```

## Integration with CI/CD

These unit tests are designed to run quickly without external dependencies:

```yaml
# .github/workflows/unit-tests.yml
- name: Run Unit Tests
  run: |
    TF_ACC=0 go test -v -race -coverprofile=coverage.txt ./aviatrix/...
  env:
    GO_TEST_TIMEOUT: "5m"
```

## Troubleshooting

### Mock Setup Issues

If your mocks aren't being called:

1. Ensure you're setting up the correct function on the mock
2. Verify the resource is using the context-aware functions (CreateContext, ReadContext, etc.)
3. Check that you're passing the mock client to the operation

### State Validation Failures

If state validation fails:

1. Ensure you're setting the ID on the ResourceData for Read/Update/Delete tests
2. Check that the mock is returning the expected data structure
3. Verify the resource's Read function is properly setting state

### Test Timeouts

If tests are timing out:

1. These are unit tests and should run quickly (< 1 second each)
2. Ensure you're not making actual API calls
3. Check for infinite loops in retry logic

## Examples for Common Resources

See `resource_unit_test_example_test.go` for complete working examples including:

- Account resource (basic CRUD)
- Schema validation for all field types
- Input validation for custom validators
- Error handling for various scenarios
- State management and change detection
- Edge case testing

## Contributing

When adding new test utilities:

1. Add them to the appropriate file (framework, templates, or helpers)
2. Document the function with clear examples
3. Add example usage in `resource_unit_test_example_test.go`
4. Update this README with usage patterns
