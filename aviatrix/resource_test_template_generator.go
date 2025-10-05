package aviatrix

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceTestTemplateGenerator generates comprehensive integration test files for resources
type ResourceTestTemplateGenerator struct {
	resourceName     string
	resourceType     string
	resource         *schema.Resource
	templateEngine   *TestTemplateEngine
	schemaAnalyzer   *ResourceSchemaAnalyzer
	testDataProvider TestDataProvider
}

// NewResourceTestTemplateGenerator creates a new test template generator
func NewResourceTestTemplateGenerator(resourceType string, resource *schema.Resource) *ResourceTestTemplateGenerator {
	return &ResourceTestTemplateGenerator{
		resourceName:     extractResourceName(resourceType),
		resourceType:     resourceType,
		resource:         resource,
		templateEngine:   NewTestTemplateEngine(),
		schemaAnalyzer:   NewResourceSchemaAnalyzer(resource),
		testDataProvider: NewTestDataProvider(),
	}
}

// GenerateTestFile generates a complete test file for a resource
func (g *ResourceTestTemplateGenerator) GenerateTestFile(outputPath string) error {
	testCases := g.generateAllTestCases()

	content, err := g.templateEngine.RenderTestFile(TestFileParams{
		ResourceName:  g.resourceName,
		ResourceType:  g.resourceType,
		TestCases:     testCases,
		RequiredEnvs:  g.schemaAnalyzer.ExtractRequiredEnvironmentVars(),
		HasPreCheck:   true,
	})

	if err != nil {
		return fmt.Errorf("failed to render test file: %w", err)
	}

	// Format the generated code
	formattedContent, err := format.Source([]byte(content))
	if err != nil {
		// If formatting fails, use unformatted content with a warning comment
		formattedContent = []byte(fmt.Sprintf("// WARNING: Code formatting failed: %v\n\n%s", err, content))
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write the test file
	if err := os.WriteFile(outputPath, formattedContent, 0644); err != nil {
		return fmt.Errorf("failed to write test file: %w", err)
	}

	return nil
}

// generateAllTestCases generates all test cases for a resource
func (g *ResourceTestTemplateGenerator) generateAllTestCases() []TestCaseDefinition {
	testCases := []TestCaseDefinition{}

	// 1. Basic CRUD test
	testCases = append(testCases, g.generateCRUDTest())

	// 2. Import test
	if g.resource.Importer != nil {
		testCases = append(testCases, g.generateImportTest())
	}

	// 3. Error handling tests
	testCases = append(testCases, g.generateErrorHandlingTests()...)

	// 4. Update tests for individual attributes
	testCases = append(testCases, g.generateUpdateTests()...)

	// 5. Dependency tests (if applicable)
	if deps := g.schemaAnalyzer.IdentifyDependencies(); len(deps) > 0 {
		testCases = append(testCases, g.generateDependencyTests()...)
	}

	return testCases
}

// generateCRUDTest generates a basic CRUD test
func (g *ResourceTestTemplateGenerator) generateCRUDTest() TestCaseDefinition {
	requiredFields := g.schemaAnalyzer.GetRequiredFields()
	optionalFields := g.schemaAnalyzer.GetOptionalFields()

	// Build create config with required fields only
	createAttrs := g.buildAttributeMap(requiredFields)

	// Build update config with required + some optional fields
	updateAttrs := g.buildAttributeMap(requiredFields)
	for i, field := range optionalFields {
		if i < 3 { // Add up to 3 optional fields for update test
			updateAttrs[field.Name] = g.generateAttributeValue(field)
		}
	}

	return TestCaseDefinition{
		Name:        fmt.Sprintf("TestAccAviatrix%s_basic", g.resourceName),
		Description: fmt.Sprintf("Basic CRUD test for %s resource", g.resourceType),
		Type:        "crud",
		CreateConfig: g.buildTerraformConfig("test", createAttrs),
		UpdateConfig: g.buildTerraformConfig("test", updateAttrs),
		Checks:       g.generateTestChecks(createAttrs, updateAttrs),
		SkipImport:   g.resource.Importer == nil,
		ImportStateVerifyIgnore: g.schemaAnalyzer.GetSensitiveFields(),
	}
}

// generateImportTest generates an import-specific test
func (g *ResourceTestTemplateGenerator) generateImportTest() TestCaseDefinition {
	requiredFields := g.schemaAnalyzer.GetRequiredFields()
	attrs := g.buildAttributeMap(requiredFields)

	return TestCaseDefinition{
		Name:        fmt.Sprintf("TestAccAviatrix%s_import", g.resourceName),
		Description: fmt.Sprintf("Import test for %s resource", g.resourceType),
		Type:        "import",
		CreateConfig: g.buildTerraformConfig("test", attrs),
		ImportStateVerifyIgnore: g.schemaAnalyzer.GetSensitiveFields(),
	}
}

// generateErrorHandlingTests generates tests for error scenarios
func (g *ResourceTestTemplateGenerator) generateErrorHandlingTests() []TestCaseDefinition {
	testCases := []TestCaseDefinition{}
	requiredFields := g.schemaAnalyzer.GetRequiredFields()

	// Test 1: Missing required fields
	for _, field := range requiredFields {
		attrs := g.buildAttributeMap(requiredFields)
		delete(attrs, field.Name)

		if len(attrs) > 0 { // Only if there are other required fields
			testCases = append(testCases, TestCaseDefinition{
				Name:        fmt.Sprintf("TestAccAviatrix%s_missingRequired_%s", g.resourceName, field.Name),
				Description: fmt.Sprintf("Test missing required field: %s", field.Name),
				Type:        "error",
				CreateConfig: g.buildTerraformConfig("test", attrs),
				ExpectError: true,
				ErrorPattern: field.Name,
			})
		}
	}

	// Test 2: Invalid values for validated fields
	validatedFields := g.schemaAnalyzer.GetValidatedFields()
	for _, field := range validatedFields {
		attrs := g.buildAttributeMap(requiredFields)
		attrs[field.Name] = g.generateInvalidValue(field)

		testCases = append(testCases, TestCaseDefinition{
			Name:        fmt.Sprintf("TestAccAviatrix%s_invalid_%s", g.resourceName, field.Name),
			Description: fmt.Sprintf("Test invalid value for field: %s", field.Name),
			Type:        "error",
			CreateConfig: g.buildTerraformConfig("test", attrs),
			ExpectError: true,
			ErrorPattern: field.Name,
		})
	}

	return testCases
}

// generateUpdateTests generates tests for updating individual attributes
func (g *ResourceTestTemplateGenerator) generateUpdateTests() []TestCaseDefinition {
	testCases := []TestCaseDefinition{}
	requiredFields := g.schemaAnalyzer.GetRequiredFields()
	updatableFields := g.schemaAnalyzer.GetUpdatableFields()

	baseAttrs := g.buildAttributeMap(requiredFields)

	for _, field := range updatableFields {
		if field.ForceNew {
			continue // Skip fields that require resource recreation
		}

		createAttrs := copyMap(baseAttrs)
		updateAttrs := copyMap(baseAttrs)

		createAttrs[field.Name] = g.generateAttributeValue(field)
		updateAttrs[field.Name] = g.generateAlternativeValue(field)

		testCases = append(testCases, TestCaseDefinition{
			Name:        fmt.Sprintf("TestAccAviatrix%s_update_%s", g.resourceName, field.Name),
			Description: fmt.Sprintf("Test updating field: %s", field.Name),
			Type:        "update",
			CreateConfig: g.buildTerraformConfig("test", createAttrs),
			UpdateConfig: g.buildTerraformConfig("test", updateAttrs),
			Checks:       g.generateTestChecks(createAttrs, updateAttrs),
			SkipImport:   true,
		})
	}

	return testCases
}

// generateDependencyTests generates tests for resource dependencies
func (g *ResourceTestTemplateGenerator) generateDependencyTests() []TestCaseDefinition {
	// This would be enhanced based on resource-specific dependency analysis
	return []TestCaseDefinition{}
}

// buildAttributeMap builds a map of attributes with generated values
func (g *ResourceTestTemplateGenerator) buildAttributeMap(fields []SchemaFieldInfo) map[string]interface{} {
	attrs := make(map[string]interface{})

	for _, field := range fields {
		attrs[field.Name] = g.generateAttributeValue(field)
	}

	return attrs
}

// generateAttributeValue generates a realistic value for a schema field
func (g *ResourceTestTemplateGenerator) generateAttributeValue(field SchemaFieldInfo) interface{} {
	// Check for environment variable reference
	if envVar := g.getEnvVarForField(field.Name); envVar != "" {
		return fmt.Sprintf("${%s}", envVar)
	}

	// Generate based on type
	switch field.Type {
	case schema.TypeString:
		return g.generateStringValue(field)
	case schema.TypeInt:
		return g.generateIntValue(field)
	case schema.TypeBool:
		return false
	case schema.TypeList:
		return g.generateListValue(field)
	case schema.TypeSet:
		return g.generateSetValue(field)
	case schema.TypeMap:
		return g.generateMapValue(field)
	default:
		return nil
	}
}

// generateStringValue generates a string value based on field characteristics
func (g *ResourceTestTemplateGenerator) generateStringValue(field SchemaFieldInfo) string {
	// Use environment variable if available
	if envVar := g.getEnvVarForField(field.Name); envVar != "" {
		return fmt.Sprintf("${%s}", envVar)
	}

	// Check for specific patterns in field name
	lowerName := strings.ToLower(field.Name)

	if strings.Contains(lowerName, "name") {
		return fmt.Sprintf("tf-test-%s", g.testDataProvider.GenerateResourceName(g.resourceName))
	}
	if strings.Contains(lowerName, "cidr") {
		return "10.0.0.0/16"
	}
	if strings.Contains(lowerName, "ip") {
		return "10.0.0.1"
	}
	if strings.Contains(lowerName, "region") {
		return "us-east-1"
	}
	if strings.Contains(lowerName, "zone") {
		return "us-east-1a"
	}
	if strings.Contains(lowerName, "size") {
		return "t2.micro"
	}

	// Default string value
	return fmt.Sprintf("test-value-%s", field.Name)
}

// generateIntValue generates an integer value
func (g *ResourceTestTemplateGenerator) generateIntValue(field SchemaFieldInfo) int {
	if strings.Contains(strings.ToLower(field.Name), "cloud_type") {
		return 1 // AWS
	}
	if strings.Contains(strings.ToLower(field.Name), "port") {
		return 443
	}
	return 1
}

// generateListValue generates a list value
func (g *ResourceTestTemplateGenerator) generateListValue(field SchemaFieldInfo) []interface{} {
	return []interface{}{"item1", "item2"}
}

// generateSetValue generates a set value
func (g *ResourceTestTemplateGenerator) generateSetValue(field SchemaFieldInfo) []interface{} {
	return []interface{}{"item1"}
}

// generateMapValue generates a map value
func (g *ResourceTestTemplateGenerator) generateMapValue(field SchemaFieldInfo) map[string]interface{} {
	return map[string]interface{}{
		"key1": "value1",
	}
}

// generateAlternativeValue generates an alternative value for update tests
func (g *ResourceTestTemplateGenerator) generateAlternativeValue(field SchemaFieldInfo) interface{} {
	val := g.generateAttributeValue(field)

	switch v := val.(type) {
	case string:
		return v + "-updated"
	case int:
		return v + 1
	case bool:
		return !v
	default:
		return val
	}
}

// generateInvalidValue generates an intentionally invalid value
func (g *ResourceTestTemplateGenerator) generateInvalidValue(field SchemaFieldInfo) interface{} {
	switch field.Type {
	case schema.TypeInt:
		return -9999
	case schema.TypeString:
		return "!@#$%^&*()_invalid_value_!@#$%^&*()"
	default:
		return "invalid"
	}
}

// buildTerraformConfig builds a Terraform configuration string
func (g *ResourceTestTemplateGenerator) buildTerraformConfig(resourceLabel string, attrs map[string]interface{}) string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", g.resourceType, resourceLabel))

	// Sort keys for consistent output
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := attrs[key]
		buf.WriteString(g.formatAttribute(key, value, 1))
	}

	buf.WriteString("}\n")

	return buf.String()
}

// formatAttribute formats a single attribute for Terraform config
func (g *ResourceTestTemplateGenerator) formatAttribute(key string, value interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)

	switch v := value.(type) {
	case string:
		// Check if it's an environment variable reference
		if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
			envVar := strings.TrimSuffix(strings.TrimPrefix(v, "${"), "}")
			return fmt.Sprintf("%s%s = %s\n", indentStr, key, envVar)
		}
		return fmt.Sprintf("%s%s = \"%s\"\n", indentStr, key, v)
	case int:
		return fmt.Sprintf("%s%s = %d\n", indentStr, key, v)
	case bool:
		return fmt.Sprintf("%s%s = %t\n", indentStr, key, v)
	case []interface{}:
		if len(v) == 0 {
			return fmt.Sprintf("%s%s = []\n", indentStr, key)
		}
		result := fmt.Sprintf("%s%s = [\n", indentStr, key)
		for _, item := range v {
			result += fmt.Sprintf("%s  \"%v\",\n", indentStr, item)
		}
		result += fmt.Sprintf("%s]\n", indentStr)
		return result
	case map[string]interface{}:
		result := fmt.Sprintf("%s%s = {\n", indentStr, key)
		for k, val := range v {
			result += g.formatAttribute(k, val, indent+1)
		}
		result += fmt.Sprintf("%s}\n", indentStr)
		return result
	default:
		return fmt.Sprintf("%s%s = %v\n", indentStr, key, v)
	}
}

// generateTestChecks generates test check functions
func (g *ResourceTestTemplateGenerator) generateTestChecks(createAttrs, updateAttrs map[string]interface{}) TestChecks {
	return TestChecks{
		Create: g.buildCheckList(createAttrs),
		Update: g.buildCheckList(updateAttrs),
	}
}

// buildCheckList builds a list of test checks for attributes
func (g *ResourceTestTemplateGenerator) buildCheckList(attrs map[string]interface{}) []string {
	checks := []string{}
	resourceName := fmt.Sprintf("%s.test", g.resourceType)

	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			if !strings.HasPrefix(v, "${") {
				checks = append(checks, fmt.Sprintf("resource.TestCheckResourceAttr(%q, %q, %q)", resourceName, key, v))
			} else {
				checks = append(checks, fmt.Sprintf("resource.TestCheckResourceAttrSet(%q, %q)", resourceName, key))
			}
		case int:
			checks = append(checks, fmt.Sprintf("resource.TestCheckResourceAttr(%q, %q, %q)", resourceName, key, fmt.Sprintf("%d", v)))
		case bool:
			checks = append(checks, fmt.Sprintf("resource.TestCheckResourceAttr(%q, %q, %q)", resourceName, key, fmt.Sprintf("%t", v)))
		}
	}

	return checks
}

// getEnvVarForField returns the environment variable name for a field if applicable
func (g *ResourceTestTemplateGenerator) getEnvVarForField(fieldName string) string {
	envVarMap := map[string]string{
		"aws_account_number": "AWS_ACCOUNT_NUMBER",
		"aws_access_key":     "AWS_ACCESS_KEY",
		"aws_secret_key":     "AWS_SECRET_KEY",
		"vpc_id":             "AWS_VPC_ID",
		"vpc_reg":            "AWS_REGION",
		"subnet":             "AWS_SUBNET",
		"gw_size":            "AWS_GW_SIZE",
	}

	return envVarMap[fieldName]
}

// extractResourceName extracts the resource name from resource type
func extractResourceName(resourceType string) string {
	// Remove "aviatrix_" prefix and convert to PascalCase
	name := strings.TrimPrefix(resourceType, "aviatrix_")
	parts := strings.Split(name, "_")

	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}

// copyMap creates a deep copy of a map
func copyMap(original map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{})
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// TestCaseDefinition defines a single test case
type TestCaseDefinition struct {
	Name                    string
	Description             string
	Type                    string // crud, import, error, update, dependency
	CreateConfig            string
	UpdateConfig            string
	Checks                  TestChecks
	SkipImport              bool
	ImportStateVerifyIgnore []string
	ExpectError             bool
	ErrorPattern            string
}

// TestChecks defines check functions for create and update steps
type TestChecks struct {
	Create []string
	Update []string
}

// TestFileParams defines parameters for test file template
type TestFileParams struct {
	ResourceName  string
	ResourceType  string
	TestCases     []TestCaseDefinition
	RequiredEnvs  []string
	HasPreCheck   bool
}

// TestTemplateEngine handles test file template rendering
type TestTemplateEngine struct {
	templates *template.Template
}

// NewTestTemplateEngine creates a new template engine
func NewTestTemplateEngine() *TestTemplateEngine {
	engine := &TestTemplateEngine{}
	engine.loadTemplates()
	return engine
}

// loadTemplates loads all test templates
func (e *TestTemplateEngine) loadTemplates() {
	funcMap := template.FuncMap{
		"title": strings.Title,
		"join":  strings.Join,
	}

	e.templates = template.New("test").Funcs(funcMap)

	// Main test file template
	template.Must(e.templates.New("testFile").Parse(testFileTemplate))
}

// RenderTestFile renders a complete test file
func (e *TestTemplateEngine) RenderTestFile(params TestFileParams) (string, error) {
	var buf bytes.Buffer

	if err := e.templates.ExecuteTemplate(&buf, "testFile", params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// testFileTemplate is the main template for generating test files
const testFileTemplate = `package aviatrix

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

{{range .TestCases}}
// {{.Name}} - {{.Description}}
func {{.Name}}(t *testing.T) {
	{{if .ExpectError}}
	// Error handling test
	skipMsg := os.Getenv("SKIP_{{$.ResourceName | title}}")
	if skipMsg == "yes" {
		t.Skip("Skipping {{$.ResourceName}} test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: GetTestProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAcc{{$.ResourceName}}_{{.Type}}(),
				ExpectError: regexp.MustCompile("{{.ErrorPattern}}"),
			},
		},
	})
	{{else}}
	skipMsg := os.Getenv("SKIP_{{$.ResourceName | title}}")
	if skipMsg == "yes" {
		t.Skip("Skipping {{$.ResourceName}} test")
	}

	resourceName := "{{$.ResourceType}}.test"
	rName := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: GetTestProviderFactories(),
		CheckDestroy:      testAccCheck{{$.ResourceName}}Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAcc{{$.ResourceName}}_{{.Type}}(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{{$.ResourceName}}Exists(resourceName),
					{{range .Checks.Create}}{{.}},
					{{end}}
				),
			},
			{{if not .SkipImport}}
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				{{if .ImportStateVerifyIgnore}}ImportStateVerifyIgnore: []string{ {{range .ImportStateVerifyIgnore}}"{{.}}", {{end}} },{{end}}
			},
			{{end}}
			{{if .UpdateConfig}}
			{
				Config: testAcc{{$.ResourceName}}_{{.Type}}_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck{{$.ResourceName}}Exists(resourceName),
					{{range .Checks.Update}}{{.}},
					{{end}}
				),
			},
			{{end}}
		},
	})
	{{end}}
}

func testAcc{{$.ResourceName}}_{{.Type}}(rName string) string {
	return fmt.Sprintf(` + "`" + `
{{.CreateConfig}}
` + "`" + `, rName)
}

{{if .UpdateConfig}}
func testAcc{{$.ResourceName}}_{{.Type}}_update(rName string) string {
	return fmt.Sprintf(` + "`" + `
{{.UpdateConfig}}
` + "`" + `, rName)
}
{{end}}

{{end}}

// Helper functions for {{.ResourceName}} tests

func testAccCheck{{.ResourceName}}Destroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*goaviatrix.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "{{.ResourceType}}" {
			continue
		}

		// Add resource-specific destroy check here
		// This is a template - implement actual check based on resource
	}

	return nil
}

func testAccCheck{{.ResourceName}}Exists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("{{.ResourceType}} not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no {{.ResourceType}} ID is set")
		}

		// Add resource-specific existence check here
		// This is a template - implement actual check based on resource

		return nil
	}
}
`
