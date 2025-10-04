package aviatrix

import (
	"context"
	"fmt"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// CRUDTestTemplate provides templates for testing CRUD operations
type CRUDTestTemplate struct {
	ResourceName string
	Resource     *schema.Resource
	t            *testing.T
}

// NewCRUDTestTemplate creates a new CRUD test template
func NewCRUDTestTemplate(t *testing.T, resourceName string, resource *schema.Resource) *CRUDTestTemplate {
	return &CRUDTestTemplate{
		ResourceName: resourceName,
		Resource:     resource,
		t:            t,
	}
}

// CreateTestConfig holds configuration for Create operation tests
type CreateTestConfig struct {
	TestName      string
	ResourceData  map[string]interface{}
	SetupMock     func(*goaviatrix.ClientInterfaceMock)
	ExpectedError string
	ValidateState func(*testing.T, *schema.ResourceData)
}

// TestCreate tests the resource Create operation
func (tt *CRUDTestTemplate) TestCreate(config CreateTestConfig) {
	tt.t.Run(config.TestName, func(t *testing.T) {
		if tt.Resource.CreateContext == nil {
			t.Skip("Resource does not implement CreateContext")
			return
		}

		// Setup mock client
		mockClient := &goaviatrix.ClientInterfaceMock{}
		if config.SetupMock != nil {
			config.SetupMock(mockClient)
		}

		// Create resource data
		rd := schema.TestResourceDataRaw(t, tt.Resource.Schema, config.ResourceData)

		// Execute Create
		ctx := context.Background()
		diags := tt.Resource.CreateContext(ctx, rd, mockClient)

		// Validate results
		if config.ExpectedError != "" {
			assert.True(t, diags.HasError(), "Expected error during Create")
			if diags.HasError() {
				assert.Contains(t, diags[0].Summary, config.ExpectedError)
			}
		} else {
			assert.False(t, diags.HasError(), "Create should not error: %v", diags)
			if config.ValidateState != nil {
				config.ValidateState(t, rd)
			}
		}
	})
}

// ReadTestConfig holds configuration for Read operation tests
type ReadTestConfig struct {
	TestName         string
	ResourceData     map[string]interface{}
	SetupMock        func(*goaviatrix.ClientInterfaceMock)
	ExpectedError    string
	ValidateState    func(*testing.T, *schema.ResourceData)
	ExpectedNotFound bool
}

// TestRead tests the resource Read operation
func (tt *CRUDTestTemplate) TestRead(config ReadTestConfig) {
	tt.t.Run(config.TestName, func(t *testing.T) {
		if tt.Resource.ReadContext == nil {
			t.Skip("Resource does not implement ReadContext")
			return
		}

		// Setup mock client
		mockClient := &goaviatrix.ClientInterfaceMock{}
		if config.SetupMock != nil {
			config.SetupMock(mockClient)
		}

		// Create resource data
		rd := schema.TestResourceDataRaw(t, tt.Resource.Schema, config.ResourceData)
		rd.SetId("test-id") // Ensure resource has an ID for Read

		// Execute Read
		ctx := context.Background()
		diags := tt.Resource.ReadContext(ctx, rd, mockClient)

		// Validate results
		if config.ExpectedError != "" {
			assert.True(t, diags.HasError(), "Expected error during Read")
			if diags.HasError() {
				assert.Contains(t, diags[0].Summary, config.ExpectedError)
			}
		} else if config.ExpectedNotFound {
			// Resource should be removed from state when not found
			assert.Equal(t, "", rd.Id(), "Resource ID should be cleared when not found")
		} else {
			assert.False(t, diags.HasError(), "Read should not error: %v", diags)
			if config.ValidateState != nil {
				config.ValidateState(t, rd)
			}
		}
	})
}

// UpdateTestConfig holds configuration for Update operation tests
type UpdateTestConfig struct {
	TestName      string
	InitialData   map[string]interface{}
	UpdatedData   map[string]interface{}
	SetupMock     func(*goaviatrix.ClientInterfaceMock)
	ExpectedError string
	ValidateState func(*testing.T, *schema.ResourceData)
}

// TestUpdate tests the resource Update operation
func (tt *CRUDTestTemplate) TestUpdate(config UpdateTestConfig) {
	tt.t.Run(config.TestName, func(t *testing.T) {
		if tt.Resource.UpdateContext == nil {
			t.Skip("Resource does not implement UpdateContext")
			return
		}

		// Setup mock client
		mockClient := &goaviatrix.ClientInterfaceMock{}
		if config.SetupMock != nil {
			config.SetupMock(mockClient)
		}

		// Create resource data with initial state
		rd := schema.TestResourceDataRaw(t, tt.Resource.Schema, config.InitialData)
		rd.SetId("test-id")

		// Apply updates
		for key, value := range config.UpdatedData {
			err := rd.Set(key, value)
			assert.NoError(t, err, "Setting %s should not error", key)
		}

		// Execute Update
		ctx := context.Background()
		diags := tt.Resource.UpdateContext(ctx, rd, mockClient)

		// Validate results
		if config.ExpectedError != "" {
			assert.True(t, diags.HasError(), "Expected error during Update")
			if diags.HasError() {
				assert.Contains(t, diags[0].Summary, config.ExpectedError)
			}
		} else {
			assert.False(t, diags.HasError(), "Update should not error: %v", diags)
			if config.ValidateState != nil {
				config.ValidateState(t, rd)
			}
		}
	})
}

// DeleteTestConfig holds configuration for Delete operation tests
type DeleteTestConfig struct {
	TestName      string
	ResourceData  map[string]interface{}
	SetupMock     func(*goaviatrix.ClientInterfaceMock)
	ExpectedError string
}

// TestDelete tests the resource Delete operation
func (tt *CRUDTestTemplate) TestDelete(config DeleteTestConfig) {
	tt.t.Run(config.TestName, func(t *testing.T) {
		if tt.Resource.DeleteContext == nil {
			t.Skip("Resource does not implement DeleteContext")
			return
		}

		// Setup mock client
		mockClient := &goaviatrix.ClientInterfaceMock{}
		if config.SetupMock != nil {
			config.SetupMock(mockClient)
		}

		// Create resource data
		rd := schema.TestResourceDataRaw(t, tt.Resource.Schema, config.ResourceData)
		rd.SetId("test-id")

		// Execute Delete
		ctx := context.Background()
		diags := tt.Resource.DeleteContext(ctx, rd, mockClient)

		// Validate results
		if config.ExpectedError != "" {
			assert.True(t, diags.HasError(), "Expected error during Delete")
			if diags.HasError() {
				assert.Contains(t, diags[0].Summary, config.ExpectedError)
			}
		} else {
			assert.False(t, diags.HasError(), "Delete should not error: %v", diags)
		}
	})
}

// ImportTestConfig holds configuration for Import operation tests
type ImportTestConfig struct {
	TestName      string
	ImportID      string
	SetupMock     func(*goaviatrix.ClientInterfaceMock)
	ExpectedError string
	ValidateState func(*testing.T, *schema.ResourceData)
}

// TestImport tests the resource Import operation
func (tt *CRUDTestTemplate) TestImport(config ImportTestConfig) {
	if tt.Resource.Importer == nil {
		tt.t.Skip("Resource does not support import")
		return
	}

	tt.t.Run(config.TestName, func(t *testing.T) {
		// Setup mock client
		mockClient := &goaviatrix.ClientInterfaceMock{}
		if config.SetupMock != nil {
			config.SetupMock(mockClient)
		}

		// Create empty resource data for import
		rd := schema.TestResourceDataRaw(t, tt.Resource.Schema, map[string]interface{}{})

		// Execute Import via StateFunc if available
		if tt.Resource.Importer.State != nil {
			results, err := tt.Resource.Importer.State(rd, mockClient)

			if config.ExpectedError != "" {
				assert.Error(t, err, "Expected error during Import")
				if err != nil {
					assert.Contains(t, err.Error(), config.ExpectedError)
				}
			} else {
				assert.NoError(t, err, "Import should not error")
				assert.NotEmpty(t, results, "Import should return resource data")
				if len(results) > 0 && config.ValidateState != nil {
					config.ValidateState(t, results[0])
				}
			}
		}
	})
}

// SchemaValidationTemplate provides templates for schema validation tests
type SchemaValidationTemplate struct {
	Resource *schema.Resource
	t        *testing.T
}

// NewSchemaValidationTemplate creates a new schema validation template
func NewSchemaValidationTemplate(t *testing.T, resource *schema.Resource) *SchemaValidationTemplate {
	return &SchemaValidationTemplate{
		Resource: resource,
		t:        t,
	}
}

// RequiredFieldTest validates that required fields are properly configured
func (svt *SchemaValidationTemplate) RequiredFieldTest(fieldName string) {
	svt.t.Run(fmt.Sprintf("RequiredField_%s", fieldName), func(t *testing.T) {
		field, exists := svt.Resource.Schema[fieldName]
		assert.True(t, exists, "Field %s should exist in schema", fieldName)
		assert.True(t, field.Required, "Field %s should be required", fieldName)
		assert.False(t, field.Optional, "Required field %s should not be optional", fieldName)
	})
}

// OptionalFieldTest validates that optional fields are properly configured
func (svt *SchemaValidationTemplate) OptionalFieldTest(fieldName string) {
	svt.t.Run(fmt.Sprintf("OptionalField_%s", fieldName), func(t *testing.T) {
		field, exists := svt.Resource.Schema[fieldName]
		assert.True(t, exists, "Field %s should exist in schema", fieldName)
		assert.True(t, field.Optional, "Field %s should be optional", fieldName)
		assert.False(t, field.Required, "Optional field %s should not be required", fieldName)
	})
}

// ComputedFieldTest validates that computed fields are properly configured
func (svt *SchemaValidationTemplate) ComputedFieldTest(fieldName string) {
	svt.t.Run(fmt.Sprintf("ComputedField_%s", fieldName), func(t *testing.T) {
		field, exists := svt.Resource.Schema[fieldName]
		assert.True(t, exists, "Field %s should exist in schema", fieldName)
		assert.True(t, field.Computed, "Field %s should be computed", fieldName)
		assert.False(t, field.Required, "Computed field %s should not be required", fieldName)
	})
}

// FieldTypeTest validates field types
func (svt *SchemaValidationTemplate) FieldTypeTest(fieldName string, expectedType schema.ValueType) {
	svt.t.Run(fmt.Sprintf("FieldType_%s", fieldName), func(t *testing.T) {
		field, exists := svt.Resource.Schema[fieldName]
		assert.True(t, exists, "Field %s should exist in schema", fieldName)
		assert.Equal(t, expectedType, field.Type, "Field %s should have correct type", fieldName)
	})
}

// DefaultValueTest validates default values
func (svt *SchemaValidationTemplate) DefaultValueTest(fieldName string, expectedDefault interface{}) {
	svt.t.Run(fmt.Sprintf("DefaultValue_%s", fieldName), func(t *testing.T) {
		field, exists := svt.Resource.Schema[fieldName]
		assert.True(t, exists, "Field %s should exist in schema", fieldName)

		if field.DefaultFunc != nil {
			defaultVal, err := field.DefaultFunc()
			assert.NoError(t, err, "DefaultFunc should not error for %s", fieldName)
			assert.Equal(t, expectedDefault, defaultVal, "Field %s should have correct default from DefaultFunc", fieldName)
		} else {
			assert.Equal(t, expectedDefault, field.Default, "Field %s should have correct default value", fieldName)
		}
	})
}

// ConflictsWithTest validates ConflictsWith configuration
func (svt *SchemaValidationTemplate) ConflictsWithTest(fieldName string, conflictingFields []string) {
	svt.t.Run(fmt.Sprintf("ConflictsWith_%s", fieldName), func(t *testing.T) {
		field, exists := svt.Resource.Schema[fieldName]
		assert.True(t, exists, "Field %s should exist in schema", fieldName)
		assert.ElementsMatch(t, conflictingFields, field.ConflictsWith, "Field %s should have correct ConflictsWith", fieldName)
	})
}

// InputValidationTemplate provides templates for input validation tests
type InputValidationTemplate struct {
	t *testing.T
}

// NewInputValidationTemplate creates a new input validation template
func NewInputValidationTemplate(t *testing.T) *InputValidationTemplate {
	return &InputValidationTemplate{t: t}
}

// ValidatorTest tests a validation function with various inputs
type ValidatorTest struct {
	Name          string
	Value         interface{}
	Key           string
	ExpectError   bool
	ErrorContains string
}

// TestValidator tests a schema validation function
func (ivt *InputValidationTemplate) TestValidator(validatorFunc schema.SchemaValidateFunc, tests []ValidatorTest) {
	for _, test := range tests {
		ivt.t.Run(test.Name, func(t *testing.T) {
			warns, errs := validatorFunc(test.Value, test.Key)

			assert.Empty(t, warns, "Should not produce warnings")

			if test.ExpectError {
				assert.NotEmpty(t, errs, "Expected validation error for %s", test.Name)
				if len(errs) > 0 && test.ErrorContains != "" {
					found := false
					for _, err := range errs {
						if assert.Contains(t, err.Error(), test.ErrorContains) {
							found = true
							break
						}
					}
					assert.True(t, found, "Error should contain expected message")
				}
			} else {
				assert.Empty(t, errs, "Should not produce errors for %s", test.Name)
			}
		})
	}
}

// ErrorHandlingTemplate provides templates for error handling tests
type ErrorHandlingTemplate struct {
	ResourceName string
	Resource     *schema.Resource
	t            *testing.T
}

// NewErrorHandlingTemplate creates a new error handling template
func NewErrorHandlingTemplate(t *testing.T, resourceName string, resource *schema.Resource) *ErrorHandlingTemplate {
	return &ErrorHandlingTemplate{
		ResourceName: resourceName,
		Resource:     resource,
		t:            t,
	}
}

// APIErrorTest tests handling of API errors
func (eht *ErrorHandlingTemplate) APIErrorTest(operation string, setupError func(*goaviatrix.ClientInterfaceMock), expectedErrorMsg string) {
	eht.t.Run(fmt.Sprintf("APIError_%s", operation), func(t *testing.T) {
		mockClient := &goaviatrix.ClientInterfaceMock{}
		setupError(mockClient)

		rd := schema.TestResourceDataRaw(t, eht.Resource.Schema, map[string]interface{}{})
		rd.SetId("test-id")

		ctx := context.Background()
		var diags diag.Diagnostics

		switch operation {
		case "Create":
			diags = eht.Resource.CreateContext(ctx, rd, mockClient)
		case "Read":
			diags = eht.Resource.ReadContext(ctx, rd, mockClient)
		case "Update":
			diags = eht.Resource.UpdateContext(ctx, rd, mockClient)
		case "Delete":
			diags = eht.Resource.DeleteContext(ctx, rd, mockClient)
		}

		assert.True(t, diags.HasError(), "Should return error from API")
		if diags.HasError() && expectedErrorMsg != "" {
			assert.Contains(t, diags[0].Summary, expectedErrorMsg)
		}
	})
}

// NotFoundErrorTest tests handling of "not found" errors during Read
func (eht *ErrorHandlingTemplate) NotFoundErrorTest(setupNotFound func(*goaviatrix.ClientInterfaceMock)) {
	eht.t.Run("NotFoundError_Read", func(t *testing.T) {
		mockClient := &goaviatrix.ClientInterfaceMock{}
		setupNotFound(mockClient)

		rd := schema.TestResourceDataRaw(t, eht.Resource.Schema, map[string]interface{}{})
		rd.SetId("test-id")

		ctx := context.Background()
		diags := eht.Resource.ReadContext(ctx, rd, mockClient)

		// Should not error but should clear the ID
		assert.False(t, diags.HasError(), "Not found should not return error")
		assert.Equal(t, "", rd.Id(), "Resource ID should be cleared when not found")
	})
}

// StateManagementTemplate provides templates for state management tests
type StateManagementTemplate struct {
	Resource *schema.Resource
	t        *testing.T
}

// NewStateManagementTemplate creates a new state management template
func NewStateManagementTemplate(t *testing.T, resource *schema.Resource) *StateManagementTemplate {
	return &StateManagementTemplate{
		Resource: resource,
		t:        t,
	}
}

// TestStateUpdate validates that state updates work correctly
func (smt *StateManagementTemplate) TestStateUpdate(initialState, updates map[string]interface{}) {
	smt.t.Run("StateUpdate", func(t *testing.T) {
		rd := schema.TestResourceDataRaw(t, smt.Resource.Schema, initialState)

		for key, value := range updates {
			err := rd.Set(key, value)
			assert.NoError(t, err, "Setting %s should not error", key)

			actual := rd.Get(key)
			assert.Equal(t, value, actual, "Value for %s should be updated", key)
		}
	})
}

// TestStateChange validates change detection
func (smt *StateManagementTemplate) TestStateChange(initialState, updates map[string]interface{}, expectedChanges []string) {
	smt.t.Run("StateChange", func(t *testing.T) {
		rd := schema.TestResourceDataRaw(t, smt.Resource.Schema, initialState)
		rd.SetId("test-id")

		// Apply updates
		for key, value := range updates {
			rd.Set(key, value)
		}

		// Verify expected changes
		for _, key := range expectedChanges {
			assert.True(t, rd.HasChange(key), "Field %s should have changed", key)
		}
	})
}
