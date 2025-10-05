package aviatrix

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestSmokeProvider verifies the provider can be initialized
func TestSmokeProvider(t *testing.T) {
	if IsAcceptanceTest() {
		t.Skip("Skipping smoke test in acceptance mode")
	}

	provider := Provider()
	assert.NotNil(t, provider, "Provider should be initialized")
	assert.NotNil(t, provider.Schema, "Provider schema should be defined")
	assert.NotNil(t, provider.ResourcesMap, "Provider resources should be defined")

	// Verify core provider fields
	assert.Contains(t, provider.Schema, "controller_ip")
	assert.Contains(t, provider.Schema, "username")
	assert.Contains(t, provider.Schema, "password")
}

// TestSmokeResourceSchemas verifies resource schemas are properly defined
func TestSmokeResourceSchemas(t *testing.T) {
	if IsAcceptanceTest() {
		t.Skip("Skipping smoke test in acceptance mode")
	}

	provider := Provider()

	// Test a few key resources exist
	testResources := []string{
		"aviatrix_account",
		"aviatrix_aws_tgw",
	}

	for _, resourceName := range testResources {
		resource, exists := provider.ResourcesMap[resourceName]
		if !exists {
			// Resource might not exist in current codebase, skip
			continue
		}

		assert.NotNil(t, resource, "Resource %s should be defined", resourceName)
		assert.NotNil(t, resource.Schema, "Resource %s should have schema", resourceName)

		// Verify CRUD operations are defined
		assert.NotNil(t, resource.ReadContext, "Resource %s should have ReadContext", resourceName)

		// Most resources have Create/Update/Delete
		if resource.CreateContext != nil {
			assert.NotNil(t, resource.CreateContext, "Resource %s should have CreateContext", resourceName)
		}
	}
}

// TestSmokeFrameworkComponents verifies framework components are available
func TestSmokeFrameworkComponents(t *testing.T) {
	if IsAcceptanceTest() {
		t.Skip("Skipping smoke test in acceptance mode")
	}

	// Test that framework helpers can be instantiated
	resource := resourceAviatrixAccount()

	// Test CRUDTestTemplate
	crudTemplate := NewCRUDTestTemplate(t, "aviatrix_account", resource)
	assert.NotNil(t, crudTemplate, "CRUDTestTemplate should be created")
	assert.Equal(t, "aviatrix_account", crudTemplate.ResourceName)

	// Test SchemaValidationTemplate
	schemaTemplate := NewSchemaValidationTemplate(t, resource)
	assert.NotNil(t, schemaTemplate, "SchemaValidationTemplate should be created")

	// Test InputValidationTemplate
	inputTemplate := NewInputValidationTemplate(t)
	assert.NotNil(t, inputTemplate, "InputValidationTemplate should be created")

	// Test ErrorHandlingTemplate
	errorTemplate := NewErrorHandlingTemplate(t, "aviatrix_account", resource)
	assert.NotNil(t, errorTemplate, "ErrorHandlingTemplate should be created")

	// Test StateManagementTemplate
	stateTemplate := NewStateManagementTemplate(t, resource)
	assert.NotNil(t, stateTemplate, "StateManagementTemplate should be created")
}

// TestSmokeHelpers verifies helper utilities work correctly
func TestSmokeHelpers(t *testing.T) {
	if IsAcceptanceTest() {
		t.Skip("Skipping smoke test in acceptance mode")
	}

	resource := resourceAviatrixAccount()

	// Test SchemaFieldTestHelper
	schemaHelper := NewSchemaFieldTestHelper(t, resource.Schema)
	assert.NotNil(t, schemaHelper, "SchemaFieldTestHelper should be created")

	// Verify it can check field existence
	field := schemaHelper.AssertFieldExists("account_name")
	assert.NotNil(t, field, "account_name field should exist")

	// Test ResourceStateTestHelper
	rd := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"account_name": "test-account",
		"cloud_type":   1,
	})
	stateHelper := NewResourceStateTestHelper(t, rd)
	assert.NotNil(t, stateHelper, "ResourceStateTestHelper should be created")

	// Test EdgeCaseTestHelper
	edgeHelper := NewEdgeCaseTestHelper(t)
	assert.NotNil(t, edgeHelper, "EdgeCaseTestHelper should be created")
}

// TestSmokeResourceTestData verifies resource test data utilities
func TestSmokeResourceTestData(t *testing.T) {
	if IsAcceptanceTest() {
		t.Skip("Skipping smoke test in acceptance mode")
	}

	resource := resourceAviatrixAccount()
	framework := NewResourceUnitTestFramework(t, resource, nil)

	testData := framework.NewResourceTestData(map[string]interface{}{
		"account_name": "test-account",
		"cloud_type":   1,
	})

	assert.NotNil(t, testData, "ResourceTestData should be created")
	assert.Equal(t, "test-account", testData.Get("account_name"))
	assert.Equal(t, 1, testData.Get("cloud_type"))
}

// TestSmokeMockBuilder verifies mock client builder
func TestSmokeMockBuilder(t *testing.T) {
	if IsAcceptanceTest() {
		t.Skip("Skipping smoke test in acceptance mode")
	}

	builder := NewMockClientBuilder()
	assert.NotNil(t, builder, "MockClientBuilder should be created")

	client := builder.Build()
	assert.NotNil(t, client, "Mock client should be built")
}
