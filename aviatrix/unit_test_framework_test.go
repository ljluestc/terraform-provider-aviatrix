package aviatrix

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestNewUnitTestFramework tests framework initialization
func TestNewUnitTestFramework(t *testing.T) {
	framework := NewUnitTestFramework(t)
	assert.NotNil(t, framework, "Framework should not be nil")
	assert.Equal(t, t, framework.t, "Framework should store test reference")
}

// TestValidateFieldType tests field type validation
func TestValidateFieldType(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"string_field": {
			Type:     schema.TypeString,
			Required: true,
		},
		"int_field": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"bool_field": {
			Type:     schema.TypeBool,
			Computed: true,
		},
	}

	t.Run("ValidateStringType", func(t *testing.T) {
		ValidateFieldType(t, testSchema, "string_field", schema.TypeString)
	})

	t.Run("ValidateIntType", func(t *testing.T) {
		ValidateFieldType(t, testSchema, "int_field", schema.TypeInt)
	})

	t.Run("ValidateBoolType", func(t *testing.T) {
		ValidateFieldType(t, testSchema, "bool_field", schema.TypeBool)
	})
}

// TestValidateFieldRequired tests required field validation
func TestValidateFieldRequired(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"required_field": {
			Type:     schema.TypeString,
			Required: true,
		},
		"optional_field": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	t.Run("ValidateRequiredField", func(t *testing.T) {
		ValidateFieldRequired(t, testSchema, "required_field", true)
	})

	t.Run("ValidateNonRequiredField", func(t *testing.T) {
		ValidateFieldRequired(t, testSchema, "optional_field", false)
	})
}

// TestValidateFieldOptional tests optional field validation
func TestValidateFieldOptional(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"optional_field": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"required_field": {
			Type:     schema.TypeString,
			Required: true,
		},
	}

	t.Run("ValidateOptionalField", func(t *testing.T) {
		ValidateFieldOptional(t, testSchema, "optional_field", true)
	})

	t.Run("ValidateNonOptionalField", func(t *testing.T) {
		ValidateFieldOptional(t, testSchema, "required_field", false)
	})
}

// TestValidateFieldComputed tests computed field validation
func TestValidateFieldComputed(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"computed_field": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"regular_field": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	t.Run("ValidateComputedField", func(t *testing.T) {
		ValidateFieldComputed(t, testSchema, "computed_field", true)
	})

	t.Run("ValidateNonComputedField", func(t *testing.T) {
		ValidateFieldComputed(t, testSchema, "regular_field", false)
	})
}

// TestValidateFieldSensitive tests sensitive field validation
func TestValidateFieldSensitive(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"password": {
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
		"username": {
			Type:     schema.TypeString,
			Required: true,
		},
	}

	t.Run("ValidateSensitiveField", func(t *testing.T) {
		ValidateFieldSensitive(t, testSchema, "password", true)
	})

	t.Run("ValidateNonSensitiveField", func(t *testing.T) {
		ValidateFieldSensitive(t, testSchema, "username", false)
	})
}

// TestValidateConflictsWith tests ConflictsWith validation
func TestValidateConflictsWith(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"option_a": {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"option_b", "option_c"},
		},
		"option_b": {
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"option_a"},
		},
		"option_c": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	t.Run("ValidateConflictsWithMultiple", func(t *testing.T) {
		ValidateConflictsWith(t, testSchema, "option_a", []string{"option_b", "option_c"})
	})

	t.Run("ValidateConflictsWithSingle", func(t *testing.T) {
		ValidateConflictsWith(t, testSchema, "option_b", []string{"option_a"})
	})
}

// TestValidateExactlyOneOf tests ExactlyOneOf validation
func TestValidateExactlyOneOf(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"method_a": {
			Type:         schema.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"method_a", "method_b", "method_c"},
		},
		"method_b": {
			Type:         schema.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"method_a", "method_b", "method_c"},
		},
		"method_c": {
			Type:         schema.TypeString,
			Optional:     true,
			ExactlyOneOf: []string{"method_a", "method_b", "method_c"},
		},
	}

	t.Run("ValidateExactlyOneOf", func(t *testing.T) {
		ValidateExactlyOneOf(t, testSchema, "method_a", []string{"method_a", "method_b", "method_c"})
	})
}

// TestCreateMockResourceData tests mock ResourceData creation
func TestCreateMockResourceData(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"count": {
			Type:     schema.TypeInt,
			Optional: true,
		},
	}

	testData := map[string]interface{}{
		"name":  "test-resource",
		"count": 5,
	}

	d := CreateMockResourceData(t, testSchema, testData)
	assert.NotNil(t, d, "ResourceData should not be nil")
	assert.Equal(t, "test-resource", d.Get("name"))
	assert.Equal(t, 5, d.Get("count"))
}

// TestValidateResourceData tests ResourceData validation
func TestValidateResourceData(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"enabled": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}

	testData := map[string]interface{}{
		"name":    "test-resource",
		"enabled": true,
	}

	d := CreateMockResourceData(t, testSchema, testData)

	expected := map[string]interface{}{
		"name":    "test-resource",
		"enabled": true,
	}

	ValidateResourceData(t, d, expected)
}

// TestValidateSchemaCompliance tests schema compliance validation
func TestValidateSchemaCompliance(t *testing.T) {
	t.Run("ValidSchema", func(t *testing.T) {
		validSchema := map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		}

		ValidateSchemaCompliance(t, validSchema)
	})

	t.Run("ValidNestedSchema", func(t *testing.T) {
		nestedSchema := map[string]*schema.Schema{
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		}

		ValidateSchemaCompliance(t, nestedSchema)
	})
}

// TestTestValidationFunction tests validation function testing
func TestTestValidationFunction(t *testing.T) {
	// Example validation function
	validatePositiveInt := func(val interface{}, key string) ([]string, []error) {
		v, ok := val.(int)
		if !ok {
			return nil, []error{fmt.Errorf("%s must be an integer", key)}
		}
		if v <= 0 {
			return nil, []error{fmt.Errorf("%s must be positive, got: %d", key, v)}
		}
		return nil, nil
	}

	testCases := []ValidationTestCase{
		{
			Name:           "ValidPositiveInt",
			Input:          10,
			Key:            "test_field",
			ExpectedErrors: nil,
		},
		{
			Name:           "InvalidNegativeInt",
			Input:          -5,
			Key:            "test_field",
			ExpectedErrors: []error{fmt.Errorf("test_field must be positive, got: -5")},
		},
		{
			Name:           "InvalidZero",
			Input:          0,
			Key:            "test_field",
			ExpectedErrors: []error{fmt.Errorf("test_field must be positive, got: 0")},
		},
	}

	TestValidationFunction(t, validatePositiveInt, testCases)
}

// TestMockContext tests mock context creation
func TestMockContext(t *testing.T) {
	ctx := MockContext()
	assert.NotNil(t, ctx, "Context should not be nil")
}

// TestAssertDiagnosticsEmpty tests diagnostics empty assertion
func TestAssertDiagnosticsEmpty(t *testing.T) {
	// This test validates that AssertDiagnosticsEmpty works correctly
	// We can't test failure cases easily, but we can test success
	t.Run("EmptyDiagnostics", func(t *testing.T) {
		// This should pass
		AssertDiagnosticsEmpty(t, nil)
	})
}
