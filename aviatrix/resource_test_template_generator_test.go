package aviatrix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestResourceTestTemplateGenerator_GenerateTestFile tests test file generation
func TestResourceTestTemplateGenerator_GenerateTestFile(t *testing.T) {
	// Create a simple test resource schema
	testResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}

	generator := NewResourceTestTemplateGenerator("aviatrix_test_resource", testResource)

	// Create temporary directory for output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "resource_aviatrix_test_resource_test.go")

	// Generate test file
	err := generator.GenerateTestFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to generate test file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Test file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)

	// Verify essential components
	expectedComponents := []string{
		"package aviatrix",
		"import",
		"func TestAcc",
		"resource.Test(",
		"testAccPreCheck",
		"GetTestProviderFactories",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(contentStr, component) {
			t.Errorf("Generated file missing component: %s", component)
		}
	}
}

// TestResourceSchemaAnalyzer_GetRequiredFields tests required field extraction
func TestResourceSchemaAnalyzer_GetRequiredFields(t *testing.T) {
	testResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
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
		},
	}

	analyzer := NewResourceSchemaAnalyzer(testResource)
	requiredFields := analyzer.GetRequiredFields()

	if len(requiredFields) != 1 {
		t.Errorf("Expected 1 required field, got %d", len(requiredFields))
	}

	if requiredFields[0].Name != "required_field" {
		t.Errorf("Expected required_field, got %s", requiredFields[0].Name)
	}
}

// TestResourceSchemaAnalyzer_GetOptionalFields tests optional field extraction
func TestResourceSchemaAnalyzer_GetOptionalFields(t *testing.T) {
	testResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"required_field": {
				Type:     schema.TypeString,
				Required: true,
			},
			"optional_field": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"computed_optional_field": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}

	analyzer := NewResourceSchemaAnalyzer(testResource)
	optionalFields := analyzer.GetOptionalFields()

	// Should only include truly optional fields (not computed+optional)
	if len(optionalFields) != 1 {
		t.Errorf("Expected 1 optional field, got %d", len(optionalFields))
	}
}

// TestResourceSchemaAnalyzer_GetSensitiveFields tests sensitive field identification
func TestResourceSchemaAnalyzer_GetSensitiveFields(t *testing.T) {
	testResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	analyzer := NewResourceSchemaAnalyzer(testResource)
	sensitiveFields := analyzer.GetSensitiveFields()

	if len(sensitiveFields) != 1 {
		t.Errorf("Expected 1 sensitive field, got %d", len(sensitiveFields))
	}

	if sensitiveFields[0] != "password" {
		t.Errorf("Expected password, got %s", sensitiveFields[0])
	}
}

// TestResourceSchemaAnalyzer_AnalyzeComplexity tests complexity analysis
func TestResourceSchemaAnalyzer_AnalyzeComplexity(t *testing.T) {
	simpleResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	complexResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"config": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"setting": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"validated_field": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(interface{}, string) ([]string, []error) {
					return nil, nil
				},
			},
		},
	}

	simpleAnalyzer := NewResourceSchemaAnalyzer(simpleResource)
	complexAnalyzer := NewResourceSchemaAnalyzer(complexResource)

	simpleComplexity := simpleAnalyzer.AnalyzeComplexity()
	complexComplexity := complexAnalyzer.AnalyzeComplexity()

	if complexComplexity <= simpleComplexity {
		t.Errorf("Complex resource should have higher complexity score. Simple: %d, Complex: %d",
			simpleComplexity, complexComplexity)
	}
}

// TestResourceSchemaAnalyzer_GenerateComplexityReport tests report generation
func TestResourceSchemaAnalyzer_GenerateComplexityReport(t *testing.T) {
	testResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"required1": {
				Type:     schema.TypeString,
				Required: true,
			},
			"required2": {
				Type:     schema.TypeString,
				Required: true,
			},
			"optional1": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"computed1": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

	analyzer := NewResourceSchemaAnalyzer(testResource)
	report := analyzer.GenerateComplexityReport()

	if report.TotalFields != 4 {
		t.Errorf("Expected 4 total fields, got %d", report.TotalFields)
	}

	if report.RequiredFields != 2 {
		t.Errorf("Expected 2 required fields, got %d", report.RequiredFields)
	}

	if report.OptionalFields != 1 {
		t.Errorf("Expected 1 optional field, got %d", report.OptionalFields)
	}

	if report.ComputedFields != 1 {
		t.Errorf("Expected 1 computed field, got %d", report.ComputedFields)
	}
}

// TestExtractResourceName tests resource name extraction
func TestExtractResourceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"aviatrix_account", "Account"},
		{"aviatrix_transit_gateway", "TransitGateway"},
		{"aviatrix_aws_tgw", "AwsTgw"},
		{"aviatrix_spoke_gateway", "SpokeGateway"},
	}

	for _, tt := range tests {
		result := extractResourceName(tt.input)
		if result != tt.expected {
			t.Errorf("extractResourceName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestGenerateStringValue tests string value generation
func TestGenerateStringValue(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString},
		},
	}

	generator := NewResourceTestTemplateGenerator("aviatrix_test", resource)

	field := SchemaFieldInfo{
		Name: "account_name",
		Type: schema.TypeString,
	}

	value := generator.generateStringValue(field)

	if value == "" {
		t.Error("Generated string value should not be empty")
	}

	if !strings.Contains(value, "tf-test") {
		t.Errorf("Expected value to contain 'tf-test', got %s", value)
	}
}

// TestBuildTerraformConfig tests Terraform config building
func TestBuildTerraformConfig(t *testing.T) {
	resource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {Type: schema.TypeString},
		},
	}

	generator := NewResourceTestTemplateGenerator("aviatrix_test", resource)

	attrs := map[string]interface{}{
		"name":    "test-resource",
		"enabled": true,
		"count":   5,
	}

	config := generator.buildTerraformConfig("test", attrs)

	expectedComponents := []string{
		"resource \"aviatrix_test\" \"test\"",
		"name = \"test-resource\"",
		"enabled = true",
		"count = 5",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(config, component) {
			t.Errorf("Config missing component: %s\nConfig:\n%s", component, config)
		}
	}
}

// TestTestCheckFuncBuilder tests check function building
func TestTestCheckFuncBuilder(t *testing.T) {
	builder := NewTestCheckFuncBuilder()

	builder.
		AddResourceAttr("aviatrix_account.test", "account_name", "test-account").
		AddResourceAttrSet("aviatrix_account.test", "id").
		AddResourceAttrPair("aviatrix_gateway.test", "account_name", "aviatrix_account.test", "account_name")

	check := builder.Build()

	if check == nil {
		t.Fatal("Built check function should not be nil")
	}
}

// TestCopyMap tests map copying
func TestCopyMap(t *testing.T) {
	original := map[string]interface{}{
		"name":  "test",
		"count": 5,
	}

	copied := copyMap(original)

	// Verify copy
	if len(copied) != len(original) {
		t.Error("Copied map has different length")
	}

	// Modify copy and verify original is unchanged
	copied["name"] = "modified"

	if original["name"] != "test" {
		t.Error("Original map was modified when copy was changed")
	}
}

// TestUniqueStrings tests string deduplication
func TestUniqueStrings(t *testing.T) {
	input := []string{"aws", "gcp", "aws", "azure", "gcp", "oci"}
	expected := []string{"aws", "gcp", "azure", "oci"}

	result := uniqueStrings(input)

	if len(result) != len(expected) {
		t.Errorf("Expected %d unique strings, got %d", len(expected), len(result))
	}

	// Verify all expected values are present
	resultMap := make(map[string]bool)
	for _, v := range result {
		resultMap[v] = true
	}

	for _, v := range expected {
		if !resultMap[v] {
			t.Errorf("Expected value %s not found in result", v)
		}
	}
}

// TestResourceSchemaAnalyzer_IdentifyDependencies tests dependency identification
func TestResourceSchemaAnalyzer_IdentifyDependencies(t *testing.T) {
	testResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"transit_gw_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}

	analyzer := NewResourceSchemaAnalyzer(testResource)
	deps := analyzer.IdentifyDependencies()

	expectedDeps := []string{"aviatrix_account", "aviatrix_vpc", "aviatrix_transit_gateway"}

	if len(deps) != len(expectedDeps) {
		t.Errorf("Expected %d dependencies, got %d", len(expectedDeps), len(deps))
	}

	depsMap := make(map[string]bool)
	for _, dep := range deps {
		depsMap[dep] = true
	}

	for _, expectedDep := range expectedDeps {
		if !depsMap[expectedDep] {
			t.Errorf("Expected dependency %s not found", expectedDep)
		}
	}
}

// TestResourceSchemaAnalyzer_SuggestTestCount tests test count suggestion
func TestResourceSchemaAnalyzer_SuggestTestCount(t *testing.T) {
	simpleResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}

	complexResource := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"field1":  {Type: schema.TypeString, Required: true},
			"field2":  {Type: schema.TypeString, Required: true},
			"field3":  {Type: schema.TypeString, Optional: true},
			"field4":  {Type: schema.TypeString, Optional: true},
			"field5":  {Type: schema.TypeBool, Optional: true},
			"field6":  {Type: schema.TypeInt, Optional: true},
			"field7":  {Type: schema.TypeString, Computed: true},
			"field8":  {Type: schema.TypeString, Computed: true},
			"field9":  {Type: schema.TypeString, Optional: true, ConflictsWith: []string{"field10"}},
			"field10": {Type: schema.TypeString, Optional: true},
		},
	}

	simpleAnalyzer := NewResourceSchemaAnalyzer(simpleResource)
	complexAnalyzer := NewResourceSchemaAnalyzer(complexResource)

	simpleCount := simpleAnalyzer.SuggestTestCount()
	complexCount := complexAnalyzer.SuggestTestCount()

	if complexCount <= simpleCount {
		t.Errorf("Complex resource should suggest more tests. Simple: %d, Complex: %d",
			simpleCount, complexCount)
	}

	if simpleCount < 3 {
		t.Errorf("Even simple resources should suggest at least 3 tests, got %d", simpleCount)
	}
}
