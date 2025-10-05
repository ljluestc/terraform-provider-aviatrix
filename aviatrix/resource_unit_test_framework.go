package aviatrix

import (
	"context"
	"fmt"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// ResourceUnitTestFramework provides comprehensive unit testing utilities for Terraform resources
type ResourceUnitTestFramework struct {
	t        *testing.T
	resource *schema.Resource
	client   goaviatrix.ClientInterface
}

// NewResourceUnitTestFramework creates a new resource unit test framework instance
func NewResourceUnitTestFramework(t *testing.T, resource *schema.Resource, client goaviatrix.ClientInterface) *ResourceUnitTestFramework {
	return &ResourceUnitTestFramework{
		t:        t,
		resource: resource,
		client:   client,
	}
}

// ResourceTestData wraps schema.TestResourceDataRaw with additional utilities
type ResourceTestData struct {
	*schema.ResourceData
	t *testing.T
}

// NewResourceTestData creates test resource data from a map of attributes
func (f *ResourceUnitTestFramework) NewResourceTestData(attributes map[string]interface{}) *ResourceTestData {
	rd := schema.TestResourceDataRaw(f.t, f.resource.Schema, attributes)
	return &ResourceTestData{
		ResourceData: rd,
		t:            f.t,
	}
}

// AssertAttribute asserts that a resource attribute has the expected value
func (rtd *ResourceTestData) AssertAttribute(key string, expected interface{}) {
	actual := rtd.Get(key)
	assert.Equal(rtd.t, expected, actual, "Attribute %s should match expected value", key)
}

// AssertAttributeSet asserts that a resource attribute is set (not nil/zero)
func (rtd *ResourceTestData) AssertAttributeSet(key string) {
	actual := rtd.Get(key)
	assert.NotNil(rtd.t, actual, "Attribute %s should be set", key)
}

// AssertAttributeNotSet asserts that a resource attribute is not set
func (rtd *ResourceTestData) AssertAttributeNotSet(key string) {
	_, ok := rtd.GetOk(key)
	assert.False(rtd.t, ok, "Attribute %s should not be set", key)
}

// AssertNoError asserts that there are no errors in diagnostics
func (rtd *ResourceTestData) AssertNoError(diags interface{}) {
	assert.Empty(rtd.t, diags, "Expected no errors")
}

// MockClientBuilder helps build mock clients for testing
type MockClientBuilder struct {
	mock *goaviatrix.ClientInterfaceMock
}

// NewMockClientBuilder creates a new mock client builder
func NewMockClientBuilder() *MockClientBuilder {
	return &MockClientBuilder{
		mock: &goaviatrix.ClientInterfaceMock{},
	}
}

// Build returns the constructed mock client
func (b *MockClientBuilder) Build() goaviatrix.ClientInterface {
	return b.mock
}

// WithGetAccountFunc configures the GetAccount mock function
func (b *MockClientBuilder) WithGetAccountFunc(fn func(*goaviatrix.Account) (goaviatrix.Account, error)) *MockClientBuilder {
	b.mock.GetAccountFunc = fn
	return b
}

// WithDeleteAccountFunc configures the DeleteAccount mock function
func (b *MockClientBuilder) WithDeleteAccountFunc(fn func(*goaviatrix.Account) error) *MockClientBuilder {
	b.mock.DeleteAccountFunc = fn
	return b
}

// WithAuditAccountFunc configures the AuditAccount mock function
func (b *MockClientBuilder) WithAuditAccountFunc(fn func(context.Context, *goaviatrix.Account) error) *MockClientBuilder {
	b.mock.AuditAccountFunc = fn
	return b
}

// ResourceTestCase defines a unit test case for resource operations
type ResourceTestCase struct {
	Name           string
	ResourceData   map[string]interface{}
	SetupMock      func(*goaviatrix.ClientInterfaceMock)
	Operation      func(context.Context, *schema.ResourceData, goaviatrix.ClientInterface) interface{}
	ValidateResult func(*testing.T, interface{}, *schema.ResourceData)
	ExpectError    bool
	ErrorContains  string
}

// RunTestCases executes a series of resource test cases
func (f *ResourceUnitTestFramework) RunTestCases(cases []ResourceTestCase) {
	for _, tc := range cases {
		f.t.Run(tc.Name, func(t *testing.T) {
			// Create mock client
			mockClient := &goaviatrix.ClientInterfaceMock{}
			if tc.SetupMock != nil {
				tc.SetupMock(mockClient)
			}

			// Create resource data
			rd := schema.TestResourceDataRaw(t, f.resource.Schema, tc.ResourceData)

			// Execute operation
			ctx := context.Background()
			result := tc.Operation(ctx, rd, mockClient)

			// Validate results
			if tc.ExpectError {
				assert.NotNil(t, result, "Expected an error")
				if tc.ErrorContains != "" && result != nil {
					assert.Contains(t, fmt.Sprintf("%v", result), tc.ErrorContains)
				}
			} else {
				if tc.ValidateResult != nil {
					tc.ValidateResult(t, result, rd)
				}
			}
		})
	}
}

// SchemaValidationTest defines a test case for schema validation
type SchemaValidationTest struct {
	Name          string
	Key           string
	Value         interface{}
	ExpectError   bool
	ErrorContains string
}

// ValidateSchemaField tests a schema field validation function
func (f *ResourceUnitTestFramework) ValidateSchemaField(fieldName string, validator schema.SchemaValidateFunc, tests []SchemaValidationTest) {
	for _, tc := range tests {
		f.t.Run(tc.Name, func(t *testing.T) {
			warns, errs := validator(tc.Value, tc.Key)

			assert.Empty(t, warns, "Expected no warnings")

			if tc.ExpectError {
				assert.NotEmpty(t, errs, "Expected validation error")
				if tc.ErrorContains != "" && len(errs) > 0 {
					assert.Contains(t, errs[0].Error(), tc.ErrorContains)
				}
			} else {
				assert.Empty(t, errs, "Expected no validation errors")
			}
		})
	}
}

// SchemaFieldTestHelper helps test schema field configurations
type SchemaFieldTestHelper struct {
	t      *testing.T
	schema map[string]*schema.Schema
}

// NewSchemaFieldTestHelper creates a helper for testing schema fields
func NewSchemaFieldTestHelper(t *testing.T, resourceSchema map[string]*schema.Schema) *SchemaFieldTestHelper {
	return &SchemaFieldTestHelper{
		t:      t,
		schema: resourceSchema,
	}
}

// AssertFieldExists asserts that a field exists in the schema
func (h *SchemaFieldTestHelper) AssertFieldExists(fieldName string) *schema.Schema {
	field, exists := h.schema[fieldName]
	assert.True(h.t, exists, "Field %s should exist in schema", fieldName)
	return field
}

// AssertFieldRequired asserts that a field is required
func (h *SchemaFieldTestHelper) AssertFieldRequired(fieldName string) {
	field := h.AssertFieldExists(fieldName)
	assert.True(h.t, field.Required, "Field %s should be required", fieldName)
}

// AssertFieldOptional asserts that a field is optional
func (h *SchemaFieldTestHelper) AssertFieldOptional(fieldName string) {
	field := h.AssertFieldExists(fieldName)
	assert.True(h.t, field.Optional, "Field %s should be optional", fieldName)
}

// AssertFieldComputed asserts that a field is computed
func (h *SchemaFieldTestHelper) AssertFieldComputed(fieldName string) {
	field := h.AssertFieldExists(fieldName)
	assert.True(h.t, field.Computed, "Field %s should be computed", fieldName)
}

// AssertFieldType asserts the type of a field
func (h *SchemaFieldTestHelper) AssertFieldType(fieldName string, expectedType schema.ValueType) {
	field := h.AssertFieldExists(fieldName)
	assert.Equal(h.t, expectedType, field.Type, "Field %s should have correct type", fieldName)
}

// AssertFieldDefault asserts the default value of a field
func (h *SchemaFieldTestHelper) AssertFieldDefault(fieldName string, expectedDefault interface{}) {
	field := h.AssertFieldExists(fieldName)
	assert.Equal(h.t, expectedDefault, field.Default, "Field %s should have correct default", fieldName)
}

// AssertFieldConflictsWith asserts that a field has ConflictsWith set
func (h *SchemaFieldTestHelper) AssertFieldConflictsWith(fieldName string, conflictingFields []string) {
	field := h.AssertFieldExists(fieldName)
	assert.ElementsMatch(h.t, conflictingFields, field.ConflictsWith, "Field %s should have correct ConflictsWith", fieldName)
}

// ResourceStateTestHelper helps test resource state management
type ResourceStateTestHelper struct {
	t  *testing.T
	rd *schema.ResourceData
}

// NewResourceStateTestHelper creates a helper for testing resource state
func NewResourceStateTestHelper(t *testing.T, rd *schema.ResourceData) *ResourceStateTestHelper {
	return &ResourceStateTestHelper{
		t:  t,
		rd: rd,
	}
}

// SetAndVerify sets a value and verifies it was set correctly
func (h *ResourceStateTestHelper) SetAndVerify(key string, value interface{}) {
	err := h.rd.Set(key, value)
	assert.NoError(h.t, err, "Setting %s should not error", key)
	actual := h.rd.Get(key)
	assert.Equal(h.t, value, actual, "Value for %s should match", key)
}

// AssertHasChange asserts that a field has changed
func (h *ResourceStateTestHelper) AssertHasChange(key string) {
	assert.True(h.t, h.rd.HasChange(key), "Field %s should have changed", key)
}

// AssertNoChange asserts that a field has not changed
func (h *ResourceStateTestHelper) AssertNoChange(key string) {
	assert.False(h.t, h.rd.HasChange(key), "Field %s should not have changed", key)
}

// GetOldNew retrieves both old and new values for a changed field
func (h *ResourceStateTestHelper) GetOldNew(key string) (interface{}, interface{}) {
	old, new := h.rd.GetChange(key)
	return old, new
}

// ComputedFieldTestHelper helps test computed field behavior
type ComputedFieldTestHelper struct {
	t      *testing.T
	schema map[string]*schema.Schema
}

// NewComputedFieldTestHelper creates a helper for testing computed fields
func NewComputedFieldTestHelper(t *testing.T, resourceSchema map[string]*schema.Schema) *ComputedFieldTestHelper {
	return &ComputedFieldTestHelper{
		t:      t,
		schema: resourceSchema,
	}
}

// AssertComputedFieldBehavior tests that a computed field behaves correctly
func (h *ComputedFieldTestHelper) AssertComputedFieldBehavior(fieldName string, testData map[string]interface{}) {
	field, exists := h.schema[fieldName]
	assert.True(h.t, exists, "Field %s should exist", fieldName)
	assert.True(h.t, field.Computed, "Field %s should be computed", fieldName)

	// Computed fields should not be required
	assert.False(h.t, field.Required, "Computed field %s should not be required", fieldName)
}

// DefaultValueTestHelper helps test default value behavior
type DefaultValueTestHelper struct {
	t        *testing.T
	resource *schema.Resource
}

// NewDefaultValueTestHelper creates a helper for testing default values
func NewDefaultValueTestHelper(t *testing.T, resource *schema.Resource) *DefaultValueTestHelper {
	return &DefaultValueTestHelper{
		t:        t,
		resource: resource,
	}
}

// AssertDefaultApplied tests that default values are applied correctly
func (h *DefaultValueTestHelper) AssertDefaultApplied(fieldName string, emptyConfig map[string]interface{}) {
	rd := schema.TestResourceDataRaw(h.t, h.resource.Schema, emptyConfig)
	field := h.resource.Schema[fieldName]

	if field.Default != nil {
		// For fields with explicit defaults
		actual := rd.Get(fieldName)
		assert.Equal(h.t, field.Default, actual, "Default value should be applied for %s", fieldName)
	} else if field.DefaultFunc != nil {
		// For fields with default functions
		defaultVal, err := field.DefaultFunc()
		assert.NoError(h.t, err, "Default function should not error for %s", fieldName)
		actual := rd.Get(fieldName)
		assert.Equal(h.t, defaultVal, actual, "Default function value should be applied for %s", fieldName)
	}
}

// EdgeCaseTestHelper provides utilities for testing edge cases
type EdgeCaseTestHelper struct {
	t *testing.T
}

// NewEdgeCaseTestHelper creates a helper for testing edge cases
func NewEdgeCaseTestHelper(t *testing.T) *EdgeCaseTestHelper {
	return &EdgeCaseTestHelper{t: t}
}

// TestEmptyString tests behavior with empty string values
func (h *EdgeCaseTestHelper) TestEmptyString(operation func(string) error, shouldError bool) {
	err := operation("")
	if shouldError {
		assert.Error(h.t, err, "Empty string should cause error")
	} else {
		assert.NoError(h.t, err, "Empty string should not cause error")
	}
}

// TestNilValue tests behavior with nil values
func (h *EdgeCaseTestHelper) TestNilValue(operation func(interface{}) error, shouldError bool) {
	err := operation(nil)
	if shouldError {
		assert.Error(h.t, err, "Nil value should cause error")
	} else {
		assert.NoError(h.t, err, "Nil value should not cause error")
	}
}

// TestMaxLength tests behavior with maximum length strings
func (h *EdgeCaseTestHelper) TestMaxLength(operation func(string) error, maxLength int, shouldError bool) {
	longString := randomString(maxLength + 1)
	err := operation(longString)
	if shouldError {
		assert.Error(h.t, err, "Exceeding max length should cause error")
	} else {
		assert.NoError(h.t, err, "Exceeding max length should not cause error")
	}
}

// randomString generates a random string of specified length for testing
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
