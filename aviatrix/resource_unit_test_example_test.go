package aviatrix

import (
	"context"
	"errors"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestResourceUnitTestFramework_AccountExample demonstrates comprehensive unit testing
// for the Account resource using the unit test framework
// Note: This is a demonstration of the framework structure. In practice, you would need
// to extend the ClientInterfaceMock to include all necessary CRUD functions.
func TestResourceUnitTestFramework_AccountExample(t *testing.T) {
	resource := resourceAviatrixAccount()
	template := NewCRUDTestTemplate(t, "aviatrix_account", resource)

	// Note: Create tests would require CreateAccountFunc to be added to the mock
	// This example shows the pattern - actual implementation needs mock extension

	// Test Read operation
	template.TestRead(ReadTestConfig{
		TestName: "ReadSuccess",
		ResourceData: map[string]interface{}{
			"account_name": "test-account",
		},
		SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
			mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
				return goaviatrix.Account{
					AccountName:      "test-account",
					CloudType:        1,
					AwsAccountNumber: "123456789012",
				}, nil
			}
		},
		ValidateState: func(t *testing.T, rd *schema.ResourceData) {
			assert.Equal(t, "test-account", rd.Get("account_name"))
			assert.Equal(t, 1, rd.Get("cloud_type"))
		},
	})

	template.TestRead(ReadTestConfig{
		TestName: "ReadNotFound",
		ResourceData: map[string]interface{}{
			"account_name": "test-account",
		},
		SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
			mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
				return goaviatrix.Account{}, errors.New("account not found")
			}
		},
		ExpectedNotFound: true,
	})

	// Test Delete operation
	template.TestDelete(DeleteTestConfig{
		TestName: "DeleteSuccess",
		ResourceData: map[string]interface{}{
			"account_name": "test-account",
		},
		SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
			mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
				assert.Equal(t, "test-account", account.AccountName)
				return nil
			}
		},
	})
}

// TestSchemaValidation_AccountExample demonstrates schema validation testing
func TestSchemaValidation_AccountExample(t *testing.T) {
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
}

// TestInputValidation_AccountExample demonstrates input validation testing
func TestInputValidation_AccountExample(t *testing.T) {
	template := NewInputValidationTemplate(t)

	// Test AWS account number validation
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
			Name:          "InvalidAccountNumber_TooLong",
			Value:         "1234567890123",
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

// TestErrorHandling_AccountExample demonstrates error handling testing
func TestErrorHandling_AccountExample(t *testing.T) {
	resource := resourceAviatrixAccount()
	template := NewErrorHandlingTemplate(t, "aviatrix_account", resource)

	// Test API error during Delete (using available mock function)
	template.APIErrorTest("Delete", func(mock *goaviatrix.ClientInterfaceMock) {
		mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
			return errors.New("account in use")
		}
	}, "failed to delete")

	// Test not found error
	template.NotFoundErrorTest(func(mock *goaviatrix.ClientInterfaceMock) {
		mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
			return goaviatrix.Account{}, errors.New("account not found")
		}
	})
}

// TestStateManagement_AccountExample demonstrates state management testing
func TestStateManagement_AccountExample(t *testing.T) {
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
			"cloud_type":   1,
			"aws_iam":      false,
		},
		map[string]interface{}{
			"aws_iam": true,
		},
		[]string{"aws_iam"},
	)
}

// TestHelpers_SchemaValidation demonstrates schema field testing helpers
func TestHelpers_SchemaValidation(t *testing.T) {
	resource := resourceAviatrixAccount()
	helper := NewSchemaFieldTestHelper(t, resource.Schema)

	// Test field existence
	helper.AssertFieldExists("account_name")
	helper.AssertFieldExists("cloud_type")

	// Test field properties
	helper.AssertFieldRequired("account_name")
	helper.AssertFieldOptional("aws_iam")
	helper.AssertFieldType("account_name", schema.TypeString)
	helper.AssertFieldType("cloud_type", schema.TypeInt)
}

// TestHelpers_ResourceState demonstrates resource state testing helpers
func TestHelpers_ResourceState(t *testing.T) {
	resource := resourceAviatrixAccount()
	rd := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"account_name": "test-account",
		"cloud_type":   1,
	})

	helper := NewResourceStateTestHelper(t, rd)

	// Test set and verify
	helper.SetAndVerify("aws_iam", true)

	// Test change detection
	helper.AssertHasChange("aws_iam")
	helper.AssertNoChange("account_name")
}

// TestHelpers_EdgeCases demonstrates edge case testing helpers
func TestHelpers_EdgeCases(t *testing.T) {
	helper := NewEdgeCaseTestHelper(t)

	// Test empty string handling
	helper.TestEmptyString(func(s string) error {
		if s == "" {
			return errors.New("empty string not allowed")
		}
		return nil
	}, true)

	// Test nil value handling
	helper.TestNilValue(func(v interface{}) error {
		if v == nil {
			return errors.New("nil value not allowed")
		}
		return nil
	}, true)

	// Test maximum length
	helper.TestMaxLength(func(s string) error {
		if len(s) > 100 {
			return errors.New("string too long")
		}
		return nil
	}, 100, true)
}

// TestResourceTestCases demonstrates using test case runner
func TestResourceTestCases(t *testing.T) {
	resource := resourceAviatrixAccount()
	framework := NewResourceUnitTestFramework(t, resource, nil)

	testCases := []ResourceTestCase{
		{
			Name: "DeleteAccount_Success",
			ResourceData: map[string]interface{}{
				"account_name": "test-account",
			},
			SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
				mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
					return nil
				}
			},
			Operation: func(ctx context.Context, rd *schema.ResourceData, client goaviatrix.ClientInterface) interface{} {
				return resourceAviatrixAccountDelete(ctx, rd, client)
			},
			ValidateResult: func(t *testing.T, result interface{}, rd *schema.ResourceData) {
				assert.Empty(t, result)
			},
			ExpectError: false,
		},
		{
			Name: "DeleteAccount_Error",
			ResourceData: map[string]interface{}{
				"account_name": "test-account",
			},
			SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
				mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
					return errors.New("API error")
				}
			},
			Operation: func(ctx context.Context, rd *schema.ResourceData, client goaviatrix.ClientInterface) interface{} {
				return resourceAviatrixAccountDelete(ctx, rd, client)
			},
			ExpectError:   true,
			ErrorContains: "failed to delete",
		},
	}

	framework.RunTestCases(testCases)
}
