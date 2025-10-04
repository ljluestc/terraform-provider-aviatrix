package aviatrix

import (
	"errors"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestUnitTestFramework_SmokeTests validates the unit test framework with 5 sample resources
func TestUnitTestFramework_SmokeTests(t *testing.T) {
	t.Run("1_AccountResource_SchemaValidation", func(t *testing.T) {
		resource := resourceAviatrixAccount()
		template := NewSchemaValidationTemplate(t, resource)

		// Test required fields
		template.RequiredFieldTest("account_name")
		template.RequiredFieldTest("cloud_type")

		// Test optional fields
		template.OptionalFieldTest("aws_iam")

		// Test field types
		template.FieldTypeTest("account_name", schema.TypeString)
		template.FieldTypeTest("cloud_type", schema.TypeInt)
		template.FieldTypeTest("aws_iam", schema.TypeBool)
	})

	t.Run("2_AccountResource_ReadOperation", func(t *testing.T) {
		resource := resourceAviatrixAccount()
		template := NewCRUDTestTemplate(t, "aviatrix_account", resource)

		// Test successful Read
		template.TestRead(ReadTestConfig{
			TestName: "ReadSuccess",
			ResourceData: map[string]interface{}{
				"account_name": "test-account",
			},
			SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
				mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
					return goaviatrix.Account{
						AccountName:      "test-account",
						CloudType:        goaviatrix.AWS,
						AwsAccountNumber: "123456789012",
					}, nil
				}
			},
			ValidateState: func(t *testing.T, rd *schema.ResourceData) {
				assert.Equal(t, "test-account", rd.Get("account_name"))
			},
		})

		// Test Read not found
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
	})

	t.Run("3_AccountResource_DeleteOperation", func(t *testing.T) {
		resource := resourceAviatrixAccount()
		template := NewCRUDTestTemplate(t, "aviatrix_account", resource)

		// Test successful Delete
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

		// Test Delete with error
		template.TestDelete(DeleteTestConfig{
			TestName: "DeleteError",
			ResourceData: map[string]interface{}{
				"account_name": "test-account",
			},
			SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
				mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
					return errors.New("account in use")
				}
			},
			ExpectedError: "failed to delete",
		})
	})

	t.Run("4_InputValidation_AWSAccountNumber", func(t *testing.T) {
		template := NewInputValidationTemplate(t)

		template.TestValidator(validateAwsAccountNumber, []ValidatorTest{
			{
				Name:        "Valid_12Digits",
				Value:       "123456789012",
				Key:         "aws_account_number",
				ExpectError: false,
			},
			{
				Name:          "Invalid_TooShort",
				Value:         "12345678901",
				Key:           "aws_account_number",
				ExpectError:   true,
				ErrorContains: "must be 12 digits",
			},
			{
				Name:          "Invalid_TooLong",
				Value:         "1234567890123",
				Key:           "aws_account_number",
				ExpectError:   true,
				ErrorContains: "must be 12 digits",
			},
			{
				Name:          "Invalid_NonNumeric",
				Value:         "12345678901A",
				Key:           "aws_account_number",
				ExpectError:   true,
				ErrorContains: "must be 12 digits",
			},
		})
	})

	t.Run("5_StateManagement_Tests", func(t *testing.T) {
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

		// Test state helper directly
		rd := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
			"account_name": "test-account",
			"cloud_type":   1,
			"aws_iam":      false,
		})
		helper := NewResourceStateTestHelper(t, rd)
		helper.SetAndVerify("aws_iam", true)
	})
}

// TestUnitTestFramework_Helpers validates helper utilities
func TestUnitTestFramework_Helpers(t *testing.T) {
	t.Run("SchemaFieldTestHelper", func(t *testing.T) {
		resource := resourceAviatrixAccount()
		helper := NewSchemaFieldTestHelper(t, resource.Schema)

		helper.AssertFieldExists("account_name")
		helper.AssertFieldRequired("account_name")
		helper.AssertFieldOptional("aws_iam")
		helper.AssertFieldType("account_name", schema.TypeString)
	})

	t.Run("ResourceStateTestHelper", func(t *testing.T) {
		resource := resourceAviatrixAccount()
		rd := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
			"account_name": "test-account",
			"cloud_type":   1,
			"aws_iam":      false,
		})
		rd.SetId("test-id")

		helper := NewResourceStateTestHelper(t, rd)
		helper.SetAndVerify("aws_iam", true)
		helper.AssertHasChange("aws_iam")
		helper.AssertNoChange("account_name")
	})

	t.Run("EdgeCaseTestHelper", func(t *testing.T) {
		helper := NewEdgeCaseTestHelper(t)

		// Test empty string validation
		helper.TestEmptyString(func(s string) error {
			if s == "" {
				return errors.New("empty string not allowed")
			}
			return nil
		}, true)

		// Test nil value validation
		helper.TestNilValue(func(v interface{}) error {
			if v == nil {
				return errors.New("nil value not allowed")
			}
			return nil
		}, true)
	})

	t.Run("ComputedFieldTestHelper", func(t *testing.T) {
		resource := resourceAviatrixAccount()
		helper := NewComputedFieldTestHelper(t, resource.Schema)

		testData := map[string]interface{}{
			"account_name": "test-account",
		}

		helper.AssertComputedFieldBehavior("rbac_groups", testData)
	})
}

// TestUnitTestFramework_ErrorHandling validates error handling patterns
func TestUnitTestFramework_ErrorHandling(t *testing.T) {
	resource := resourceAviatrixAccount()
	template := NewErrorHandlingTemplate(t, "aviatrix_account", resource)

	t.Run("APIError_Delete", func(t *testing.T) {
		template.APIErrorTest("Delete", func(mock *goaviatrix.ClientInterfaceMock) {
			mock.DeleteAccountFunc = func(account *goaviatrix.Account) error {
				return errors.New("API connection timeout")
			}
		}, "failed to delete")
	})

	t.Run("NotFoundError_Read", func(t *testing.T) {
		template.NotFoundErrorTest(func(mock *goaviatrix.ClientInterfaceMock) {
			mock.GetAccountFunc = func(account *goaviatrix.Account) (goaviatrix.Account, error) {
				return goaviatrix.Account{}, errors.New("account not found")
			}
		})
	})
}
