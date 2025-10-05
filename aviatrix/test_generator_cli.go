package aviatrix

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestGeneratorCLI provides command-line interface for test generation
type TestGeneratorCLI struct {
	outputDir     string
	resourceTypes []string
	generateAll   bool
	verbose       bool
	dryRun        bool
}

// NewTestGeneratorCLI creates a new CLI instance
func NewTestGeneratorCLI() *TestGeneratorCLI {
	return &TestGeneratorCLI{}
}

// ParseFlags parses command-line flags
func (cli *TestGeneratorCLI) ParseFlags() error {
	flag.StringVar(&cli.outputDir, "output", "./aviatrix", "Output directory for generated test files")
	flag.BoolVar(&cli.generateAll, "all", false, "Generate tests for all resources")
	flag.BoolVar(&cli.verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&cli.dryRun, "dry-run", false, "Show what would be generated without writing files")

	flag.Parse()

	// Remaining args are resource types
	cli.resourceTypes = flag.Args()

	if !cli.generateAll && len(cli.resourceTypes) == 0 {
		return fmt.Errorf("must specify resource types or use -all flag")
	}

	return nil
}

// Run executes the test generator
func (cli *TestGeneratorCLI) Run() error {
	provider := Provider()
	resourceMap := provider.ResourcesMap

	if cli.generateAll {
		// Generate for all resources
		for resourceType := range resourceMap {
			if err := cli.generateTestForResource(resourceType, resourceMap[resourceType]); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating test for %s: %v\n", resourceType, err)
				continue
			}
		}
	} else {
		// Generate for specified resources
		for _, resourceType := range cli.resourceTypes {
			resource, ok := resourceMap[resourceType]
			if !ok {
				fmt.Fprintf(os.Stderr, "Resource not found: %s\n", resourceType)
				continue
			}

			if err := cli.generateTestForResource(resourceType, resource); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating test for %s: %v\n", resourceType, err)
				continue
			}
		}
	}

	return nil
}

// generateTestForResource generates a test file for a single resource
func (cli *TestGeneratorCLI) generateTestForResource(resourceType string, resource *schema.Resource) error {
	generator := NewResourceTestTemplateGenerator(resourceType, resource)

	// Determine output file path
	fileName := fmt.Sprintf("%s_test.go", resourceType)
	outputPath := filepath.Join(cli.outputDir, fileName)

	if cli.verbose {
		fmt.Printf("Generating test for %s...\n", resourceType)
		analyzer := NewResourceSchemaAnalyzer(resource)
		report := analyzer.GenerateComplexityReport()
		fmt.Printf("  Complexity: %d (suggested tests: %d)\n", report.ComplexityScore, report.SuggestedTests)
		fmt.Printf("  Fields: %d required, %d optional, %d computed\n",
			report.RequiredFields, report.OptionalFields, report.ComputedFields)
	}

	if cli.dryRun {
		fmt.Printf("Would generate: %s\n", outputPath)
		return nil
	}

	// Generate the test file
	if err := generator.GenerateTestFile(outputPath); err != nil {
		return fmt.Errorf("failed to generate test file: %w", err)
	}

	if cli.verbose {
		fmt.Printf("  Generated: %s\n", outputPath)
	}

	return nil
}

// GenerateTestSuite generates a complete test suite for all resources
func GenerateTestSuite(outputDir string) error {
	provider := Provider()
	resourceMap := provider.ResourcesMap

	var succeeded, failed int

	for resourceType, resource := range resourceMap {
		generator := NewResourceTestTemplateGenerator(resourceType, resource)

		fileName := fmt.Sprintf("%s_test.go", resourceType)
		outputPath := filepath.Join(outputDir, fileName)

		if err := generator.GenerateTestFile(outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", resourceType, err)
			failed++
			continue
		}

		succeeded++
	}

	fmt.Printf("\nTest generation complete:\n")
	fmt.Printf("  Succeeded: %d\n", succeeded)
	fmt.Printf("  Failed: %d\n", failed)
	fmt.Printf("  Total: %d\n", succeeded+failed)

	return nil
}

// GenerateTestForResourceType generates a test file for a specific resource type
func GenerateTestForResourceType(resourceType, outputDir string) error {
	provider := Provider()
	resource, ok := provider.ResourcesMap[resourceType]
	if !ok {
		return fmt.Errorf("resource type not found: %s", resourceType)
	}

	generator := NewResourceTestTemplateGenerator(resourceType, resource)

	fileName := fmt.Sprintf("%s_test.go", resourceType)
	outputPath := filepath.Join(outputDir, fileName)

	return generator.GenerateTestFile(outputPath)
}

// ListResources lists all available resources
func ListResources() []string {
	provider := Provider()
	resources := make([]string, 0, len(provider.ResourcesMap))

	for resourceType := range provider.ResourcesMap {
		resources = append(resources, resourceType)
	}

	return resources
}

// AnalyzeResource provides detailed analysis of a resource
func AnalyzeResource(resourceType string) (*SchemaComplexityReport, error) {
	provider := Provider()
	resource, ok := provider.ResourcesMap[resourceType]
	if !ok {
		return nil, fmt.Errorf("resource type not found: %s", resourceType)
	}

	analyzer := NewResourceSchemaAnalyzer(resource)
	report := analyzer.GenerateComplexityReport()

	return &report, nil
}

// GenerateTestMatrix generates a test matrix configuration
func GenerateTestMatrix(resources []string, outputPath string) error {
	var matrix strings.Builder

	matrix.WriteString("# Test Matrix Configuration\n\n")
	matrix.WriteString("This file contains test scenarios for resources.\n\n")

	for _, resourceType := range resources {
		report, err := AnalyzeResource(resourceType)
		if err != nil {
			continue
		}

		matrix.WriteString(fmt.Sprintf("## %s\n\n", resourceType))
		matrix.WriteString(fmt.Sprintf("- **Complexity Score**: %d\n", report.ComplexityScore))
		matrix.WriteString(fmt.Sprintf("- **Suggested Tests**: %d\n", report.SuggestedTests))
		matrix.WriteString(fmt.Sprintf("- **Required Fields**: %d\n", report.RequiredFields))
		matrix.WriteString(fmt.Sprintf("- **Optional Fields**: %d\n", report.OptionalFields))
		matrix.WriteString(fmt.Sprintf("- **Computed Fields**: %d\n", report.ComputedFields))
		matrix.WriteString(fmt.Sprintf("- **Nested Resources**: %d\n", report.NestedResources))
		matrix.WriteString("\n")
	}

	return os.WriteFile(outputPath, []byte(matrix.String()), 0644)
}

// TestScaffoldConfig defines configuration for test scaffolding
type TestScaffoldConfig struct {
	ResourceType    string
	IncludeCRUD     bool
	IncludeImport   bool
	IncludeErrors   bool
	IncludeUpdates  bool
	CustomPreCheck  string
	CustomFixtures  []string
}

// GenerateTestScaffold generates a minimal test scaffold for quick customization
func GenerateTestScaffold(config TestScaffoldConfig, outputPath string) error {
	provider := Provider()
	resource, ok := provider.ResourcesMap[config.ResourceType]
	if !ok {
		return fmt.Errorf("resource type not found: %s", config.ResourceType)
	}

	analyzer := NewResourceSchemaAnalyzer(resource)
	requiredFields := analyzer.GetRequiredFields()

	var scaffold strings.Builder

	scaffold.WriteString(fmt.Sprintf("package aviatrix\n\n"))
	scaffold.WriteString("import (\n")
	scaffold.WriteString("\t\"testing\"\n")
	scaffold.WriteString("\t\"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource\"\n")
	scaffold.WriteString(")\n\n")

	// Basic test function
	scaffold.WriteString(fmt.Sprintf("func TestAcc%s_basic(t *testing.T) {\n", extractResourceName(config.ResourceType)))
	scaffold.WriteString("\t// TODO: Implement test\n")
	scaffold.WriteString("\tresource.Test(t, resource.TestCase{\n")
	scaffold.WriteString("\t\tPreCheck: func() { testAccPreCheck(t) },\n")
	scaffold.WriteString("\t\tProviderFactories: GetTestProviderFactories(),\n")
	scaffold.WriteString("\t\tSteps: []resource.TestStep{\n")
	scaffold.WriteString("\t\t\t{\n")
	scaffold.WriteString("\t\t\t\tConfig: testAccConfig_basic(),\n")
	scaffold.WriteString("\t\t\t},\n")
	scaffold.WriteString("\t\t},\n")
	scaffold.WriteString("\t})\n")
	scaffold.WriteString("}\n\n")

	// Config function with required fields
	scaffold.WriteString("func testAccConfig_basic() string {\n")
	scaffold.WriteString("\treturn `\n")
	scaffold.WriteString(fmt.Sprintf("resource \"%s\" \"test\" {\n", config.ResourceType))

	for _, field := range requiredFields {
		scaffold.WriteString(fmt.Sprintf("  %s = \"TODO: Set value\"\n", field.Name))
	}

	scaffold.WriteString("}\n")
	scaffold.WriteString("`\n")
	scaffold.WriteString("}\n")

	return os.WriteFile(outputPath, []byte(scaffold.String()), 0644)
}
