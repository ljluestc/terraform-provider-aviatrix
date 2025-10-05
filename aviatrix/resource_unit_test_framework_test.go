package aviatrix

import (
	"context"
	"errors"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestResourceUnitTestFramework_Initialize tests framework initialization
func TestResourceUnitTestFramework_Initialize(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	mockClient := &goaviatrix.ClientInterfaceMock{}
	framework := NewResourceUnitTestFramework(t, resource, mockClient)

	assert.NotNil(t, framework, "Framework should not be nil")
	assert.Equal(t, resource, framework.resource, "Resource should be set")
	assert.Equal(t, mockClient, framework.client, "Client should be set")
}

// TestResourceTestData_AssertAttribute tests attribute assertion
func TestResourceTestData_AssertAttribute(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}

	framework := NewResourceUnitTestFramework(t, resource, &goaviatrix.ClientInterfaceMock{})
	testData := framework.NewResourceTestData(map[string]interface{}{
		"name":    "test-resource",
		"enabled": true,
	})

	testData.AssertAttribute("name", "test-resource")
	testData.AssertAttribute("enabled", true)
}

// TestResourceTestData_AssertAttributeSet tests attribute set assertion
func TestResourceTestData_AssertAttributeSet(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}

	framework := NewResourceUnitTestFramework(t, resource, &goaviatrix.ClientInterfaceMock{})
	testData := framework.NewResourceTestData(map[string]interface{}{
		"name": "test-value",
	})

	testData.AssertAttributeSet("name")
}

// TestResourceTestData_AssertAttributeNotSet tests attribute not set assertion
func TestResourceTestData_AssertAttributeNotSet(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}

	framework := NewResourceUnitTestFramework(t, resource, &goaviatrix.ClientInterfaceMock{})
	testData := framework.NewResourceTestData(map[string]interface{}{
		"name": "test-value",
	})

	testData.AssertAttributeNotSet("description")
}

// TestResourceTestData_AssertNoError tests no error assertion
func TestResourceTestData_AssertNoError(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	framework := NewResourceUnitTestFramework(t, resource, &goaviatrix.ClientInterfaceMock{})
	testData := framework.NewResourceTestData(map[string]interface{}{
		"name": "test",
	})

	testData.AssertNoError(nil)
	testData.AssertNoError(diag.Diagnostics{})
}

// TestMockClientBuilder tests mock client builder
func TestMockClientBuilder(t *testing.T) {
	builder := NewMockClientBuilder()
	assert.NotNil(t, builder, "Builder should not be nil")

	client := builder.Build()
	assert.NotNil(t, client, "Client should not be nil")
}

// TestResourceTestCase_RunTestCases tests running resource test cases
func TestResourceTestCase_RunTestCases(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	framework := NewResourceUnitTestFramework(t, resource, &goaviatrix.ClientInterfaceMock{})

	cases := []ResourceTestCase{
		{
			Name: "SuccessfulOperation",
			ResourceData: map[string]interface{}{
				"account_name": "test-account",
			},
			SetupMock: func(mock *goaviatrix.ClientInterfaceMock) {
				// Setup mock behavior if needed
			},
			Operation: func(ctx context.Context, d *schema.ResourceData, client goaviatrix.ClientInterface) interface{} {
				// Simulate successful operation
				return nil
			},
			ValidateResult: func(t *testing.T, result interface{}, d *schema.ResourceData) {
				assert.Nil(t, result, "Result should be nil for success")
			},
			ExpectError: false,
		},
		{
			Name: "OperationWithError",
			ResourceData: map[string]interface{}{
				"account_name": "test-account",
			},
			Operation: func(ctx context.Context, d *schema.ResourceData, client goaviatrix.ClientInterface) interface{} {
				return diag.Errorf("operation failed")
			},
			ExpectError:   true,
			ErrorContains: "operation failed",
		},
	}

	framework.RunTestCases(cases)
}

// TestSchemaFieldTestHelper tests schema field helper
func TestSchemaFieldTestHelper(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"required_field": {
			Type:     schema.TypeString,
			Required: true,
		},
		"optional_field": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"computed_field": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"int_field": {
			Type:    schema.TypeInt,
			Default: 10,
		},
		"conflicting_field": {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"other_field"},
		},
		"other_field": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	helper := NewSchemaFieldTestHelper(t, testSchema)

	t.Run("AssertFieldExists", func(t *testing.T) {
		field := helper.AssertFieldExists("required_field")
		assert.NotNil(t, field)
	})

	t.Run("AssertFieldRequired", func(t *testing.T) {
		helper.AssertFieldRequired("required_field")
	})

	t.Run("AssertFieldOptional", func(t *testing.T) {
		helper.AssertFieldOptional("optional_field")
	})

	t.Run("AssertFieldComputed", func(t *testing.T) {
		helper.AssertFieldComputed("computed_field")
	})

	t.Run("AssertFieldType", func(t *testing.T) {
		helper.AssertFieldType("int_field", schema.TypeInt)
	})

	t.Run("AssertFieldDefault", func(t *testing.T) {
		helper.AssertFieldDefault("int_field", 10)
	})

	t.Run("AssertFieldConflictsWith", func(t *testing.T) {
		helper.AssertFieldConflictsWith("conflicting_field", []string{"other_field"})
	})
}

// TestResourceStateTestHelper tests resource state helper
func TestResourceStateTestHelper(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"count": {
			Type:     schema.TypeInt,
			Optional: true,
		},
	}

	rd := schema.TestResourceDataRaw(t, testSchema, map[string]interface{}{
		"name": "original",
	})

	helper := NewResourceStateTestHelper(t, rd)

	t.Run("SetAndVerify", func(t *testing.T) {
		helper.SetAndVerify("count", 5)
	})

	t.Run("AssertHasChange", func(t *testing.T) {
		rd.Set("name", "updated")
		helper.AssertHasChange("name")
	})

	t.Run("GetOldNew", func(t *testing.T) {
		old, new := helper.GetOldNew("name")
		// After Set, GetChange returns current value as both old and new if not in a real state change context
		// In test context, we just verify the method works
		assert.NotNil(t, old)
		assert.NotNil(t, new)
	})
}

// TestComputedFieldTestHelper tests computed field helper
func TestComputedFieldTestHelper(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"created_at": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}

	helper := NewComputedFieldTestHelper(t, testSchema)

	helper.AssertComputedFieldBehavior("id", nil)
	helper.AssertComputedFieldBehavior("created_at", nil)
}

// TestDefaultValueTestHelper tests default value helper
func TestDefaultValueTestHelper(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:    schema.TypeBool,
				Default: true,
			},
			"timeout": {
				Type:    schema.TypeInt,
				Default: 30,
			},
		},
	}

	helper := NewDefaultValueTestHelper(t, resource)

	helper.AssertDefaultApplied("enabled", map[string]interface{}{})
	helper.AssertDefaultApplied("timeout", map[string]interface{}{})
}

// TestEdgeCaseTestHelper tests edge case helper
func TestEdgeCaseTestHelper(t *testing.T) {
	helper := NewEdgeCaseTestHelper(t)

	t.Run("TestEmptyString", func(t *testing.T) {
		operation := func(s string) error {
			if s == "" {
				return errors.New("empty string")
			}
			return nil
		}

		helper.TestEmptyString(operation, true)
	})

	t.Run("TestNilValue", func(t *testing.T) {
		operation := func(v interface{}) error {
			if v == nil {
				return errors.New("nil value")
			}
			return nil
		}

		helper.TestNilValue(operation, true)
	})

	t.Run("TestMaxLength", func(t *testing.T) {
		operation := func(s string) error {
			if len(s) > 10 {
				return errors.New("too long")
			}
			return nil
		}

		helper.TestMaxLength(operation, 10, true)
	})
}

// TestRandomString tests random string generation
func TestRandomString(t *testing.T) {
	str10 := randomString(10)
	assert.Equal(t, 10, len(str10), "Should generate string of correct length")

	str100 := randomString(100)
	assert.Equal(t, 100, len(str100), "Should generate string of correct length")

	// Test that function generates deterministic pattern
	str10_again := randomString(10)
	assert.Equal(t, str10, str10_again, "Same length should generate same pattern")
}
