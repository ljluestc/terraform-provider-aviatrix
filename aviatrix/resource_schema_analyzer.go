package aviatrix

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceSchemaAnalyzer analyzes a resource schema to extract metadata for test generation
type ResourceSchemaAnalyzer struct {
	resource *schema.Resource
}

// NewResourceSchemaAnalyzer creates a new schema analyzer
func NewResourceSchemaAnalyzer(resource *schema.Resource) *ResourceSchemaAnalyzer {
	return &ResourceSchemaAnalyzer{
		resource: resource,
	}
}

// SchemaFieldInfo contains information about a schema field
type SchemaFieldInfo struct {
	Name        string
	Type        schema.ValueType
	Required    bool
	Optional    bool
	Computed    bool
	ForceNew    bool
	Sensitive   bool
	Description string
	Default     interface{}
}

// GetRequiredFields returns all required fields in the schema
func (a *ResourceSchemaAnalyzer) GetRequiredFields() []SchemaFieldInfo {
	var fields []SchemaFieldInfo

	for name, field := range a.resource.Schema {
		if field.Required {
			fields = append(fields, a.buildFieldInfo(name, field))
		}
	}

	return fields
}

// GetOptionalFields returns all optional (non-computed) fields
func (a *ResourceSchemaAnalyzer) GetOptionalFields() []SchemaFieldInfo {
	var fields []SchemaFieldInfo

	for name, field := range a.resource.Schema {
		if field.Optional && !field.Computed {
			fields = append(fields, a.buildFieldInfo(name, field))
		}
	}

	return fields
}

// GetUpdatableFields returns fields that can be updated (not ForceNew)
func (a *ResourceSchemaAnalyzer) GetUpdatableFields() []SchemaFieldInfo {
	var fields []SchemaFieldInfo

	for name, field := range a.resource.Schema {
		if (field.Required || field.Optional) && !field.Computed && !field.ForceNew {
			fields = append(fields, a.buildFieldInfo(name, field))
		}
	}

	return fields
}

// GetSensitiveFields returns names of sensitive fields
func (a *ResourceSchemaAnalyzer) GetSensitiveFields() []string {
	var fields []string

	for name, field := range a.resource.Schema {
		if field.Sensitive {
			fields = append(fields, name)
		}
	}

	return fields
}

// GetValidatedFields returns fields that have validation functions
func (a *ResourceSchemaAnalyzer) GetValidatedFields() []SchemaFieldInfo {
	var fields []SchemaFieldInfo

	for name, field := range a.resource.Schema {
		if field.ValidateFunc != nil || field.ValidateDiagFunc != nil {
			fields = append(fields, a.buildFieldInfo(name, field))
		}
	}

	return fields
}

// GetComputedFields returns all computed fields
func (a *ResourceSchemaAnalyzer) GetComputedFields() []SchemaFieldInfo {
	var fields []SchemaFieldInfo

	for name, field := range a.resource.Schema {
		if field.Computed && !field.Optional {
			fields = append(fields, a.buildFieldInfo(name, field))
		}
	}

	return fields
}

// IdentifyDependencies identifies potential resource dependencies
func (a *ResourceSchemaAnalyzer) IdentifyDependencies() []string {
	var deps []string

	for name, field := range a.resource.Schema {
		// Look for fields that commonly reference other resources
		lowerName := strings.ToLower(name)

		if strings.Contains(lowerName, "account_name") {
			deps = append(deps, "aviatrix_account")
		} else if strings.Contains(lowerName, "vpc_id") {
			deps = append(deps, "aviatrix_vpc")
		} else if strings.Contains(lowerName, "transit_gw") {
			deps = append(deps, "aviatrix_transit_gateway")
		} else if strings.Contains(lowerName, "spoke_gw") {
			deps = append(deps, "aviatrix_spoke_gateway")
		}

		// Check field type for TypeMap or TypeSet which might contain references
		if field.Type == schema.TypeSet || field.Type == schema.TypeList {
			if elem, ok := field.Elem.(*schema.Resource); ok {
				// Nested resource - could have dependencies
				for nestedName := range elem.Schema {
					lowerNested := strings.ToLower(nestedName)
					if strings.Contains(lowerNested, "name") || strings.Contains(lowerNested, "id") {
						// Potential reference field
					}
				}
			}
		}
	}

	return uniqueStrings(deps)
}

// ExtractRequiredEnvironmentVars identifies environment variables needed for testing
func (a *ResourceSchemaAnalyzer) ExtractRequiredEnvironmentVars() []string {
	var envVars []string

	for name := range a.resource.Schema {
		lowerName := strings.ToLower(name)

		// Map field names to environment variables
		if strings.Contains(lowerName, "cloud_type") {
			envVars = append(envVars, "AWS_ACCOUNT_NUMBER", "AWS_ACCESS_KEY", "AWS_SECRET_KEY")
		}
		if strings.Contains(lowerName, "aws") {
			envVars = append(envVars, "AWS_ACCOUNT_NUMBER", "AWS_ACCESS_KEY", "AWS_SECRET_KEY", "AWS_VPC_ID", "AWS_REGION")
		}
		if strings.Contains(lowerName, "gcp") || strings.Contains(lowerName, "gcloud") {
			envVars = append(envVars, "GCP_ID", "GCP_CREDENTIALS_FILEPATH")
		}
		if strings.Contains(lowerName, "azure") || strings.Contains(lowerName, "arm_") {
			envVars = append(envVars, "ARM_SUBSCRIPTION_ID", "ARM_DIRECTORY_ID", "ARM_APPLICATION_ID", "ARM_APPLICATION_KEY")
		}
		if strings.Contains(lowerName, "oci") {
			envVars = append(envVars, "OCI_TENANCY_ID", "OCI_USER_ID", "OCI_COMPARTMENT_ID", "OCI_API_KEY_FILEPATH")
		}
	}

	return uniqueStrings(envVars)
}

// HasCreateTimeoutCustomization checks if resource has custom create timeout
func (a *ResourceSchemaAnalyzer) HasCreateTimeoutCustomization() bool {
	return a.resource.Timeouts != nil && a.resource.Timeouts.Create != nil
}

// HasUpdateTimeoutCustomization checks if resource has custom update timeout
func (a *ResourceSchemaAnalyzer) HasUpdateTimeoutCustomization() bool {
	return a.resource.Timeouts != nil && a.resource.Timeouts.Update != nil
}

// HasDeleteTimeoutCustomization checks if resource has custom delete timeout
func (a *ResourceSchemaAnalyzer) HasDeleteTimeoutCustomization() bool {
	return a.resource.Timeouts != nil && a.resource.Timeouts.Delete != nil
}

// IsImportable checks if the resource supports import
func (a *ResourceSchemaAnalyzer) IsImportable() bool {
	return a.resource.Importer != nil
}

// GetConflictingFields identifies fields with ConflictsWith constraints
func (a *ResourceSchemaAnalyzer) GetConflictingFields() map[string][]string {
	conflicts := make(map[string][]string)

	for name, field := range a.resource.Schema {
		if len(field.ConflictsWith) > 0 {
			conflicts[name] = field.ConflictsWith
		}
	}

	return conflicts
}

// GetRequiredWithFields identifies fields with RequiredWith constraints
func (a *ResourceSchemaAnalyzer) GetRequiredWithFields() map[string][]string {
	requiredWith := make(map[string][]string)

	for name, field := range a.resource.Schema {
		if len(field.RequiredWith) > 0 {
			requiredWith[name] = field.RequiredWith
		}
	}

	return requiredWith
}

// GetExactlyOneOfFields identifies fields with ExactlyOneOf constraints
func (a *ResourceSchemaAnalyzer) GetExactlyOneOfFields() map[string][]string {
	exactlyOneOf := make(map[string][]string)

	for name, field := range a.resource.Schema {
		if len(field.ExactlyOneOf) > 0 {
			exactlyOneOf[name] = field.ExactlyOneOf
		}
	}

	return exactlyOneOf
}

// GetAtLeastOneOfFields identifies fields with AtLeastOneOf constraints
func (a *ResourceSchemaAnalyzer) GetAtLeastOneOfFields() map[string][]string {
	atLeastOneOf := make(map[string][]string)

	for name, field := range a.resource.Schema {
		if len(field.AtLeastOneOf) > 0 {
			atLeastOneOf[name] = field.AtLeastOneOf
		}
	}

	return atLeastOneOf
}

// GetDeprecatedFields returns fields marked as deprecated
func (a *ResourceSchemaAnalyzer) GetDeprecatedFields() []SchemaFieldInfo {
	var fields []SchemaFieldInfo

	for name, field := range a.resource.Schema {
		if field.Deprecated != "" {
			fields = append(fields, a.buildFieldInfo(name, field))
		}
	}

	return fields
}

// GetDefaultValues returns fields with default values
func (a *ResourceSchemaAnalyzer) GetDefaultValues() map[string]interface{} {
	defaults := make(map[string]interface{})

	for name, field := range a.resource.Schema {
		if field.Default != nil {
			defaults[name] = field.Default
		}
	}

	return defaults
}

// AnalyzeComplexity provides a complexity score for the resource
func (a *ResourceSchemaAnalyzer) AnalyzeComplexity() int {
	complexity := 0

	// Base complexity from field count
	complexity += len(a.resource.Schema)

	// Add complexity for nested structures
	for _, field := range a.resource.Schema {
		if field.Type == schema.TypeSet || field.Type == schema.TypeList {
			if _, ok := field.Elem.(*schema.Resource); ok {
				complexity += 5 // Nested resources add significant complexity
			}
		}

		// Add complexity for validation
		if field.ValidateFunc != nil || field.ValidateDiagFunc != nil {
			complexity += 2
		}

		// Add complexity for constraints
		if len(field.ConflictsWith) > 0 || len(field.RequiredWith) > 0 {
			complexity += 3
		}
	}

	return complexity
}

// SuggestTestCount suggests number of test cases based on complexity
func (a *ResourceSchemaAnalyzer) SuggestTestCount() int {
	complexity := a.AnalyzeComplexity()

	switch {
	case complexity < 10:
		return 3 // Basic tests
	case complexity < 20:
		return 5 // Standard tests
	case complexity < 40:
		return 8 // Comprehensive tests
	default:
		return 12 // Extensive tests
	}
}

// buildFieldInfo creates a SchemaFieldInfo from a schema field
func (a *ResourceSchemaAnalyzer) buildFieldInfo(name string, field *schema.Schema) SchemaFieldInfo {
	return SchemaFieldInfo{
		Name:        name,
		Type:        field.Type,
		Required:    field.Required,
		Optional:    field.Optional,
		Computed:    field.Computed,
		ForceNew:    field.ForceNew,
		Sensitive:   field.Sensitive,
		Description: field.Description,
		Default:     field.Default,
	}
}

// uniqueStrings removes duplicates from a string slice
func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, val := range input {
		if !seen[val] {
			seen[val] = true
			result = append(result, val)
		}
	}

	return result
}

// SchemaComplexityReport provides detailed complexity analysis
type SchemaComplexityReport struct {
	TotalFields       int
	RequiredFields    int
	OptionalFields    int
	ComputedFields    int
	NestedResources   int
	ValidatedFields   int
	ConflictingFields int
	ComplexityScore   int
	SuggestedTests    int
}

// GenerateComplexityReport generates a detailed complexity report
func (a *ResourceSchemaAnalyzer) GenerateComplexityReport() SchemaComplexityReport {
	report := SchemaComplexityReport{
		TotalFields:     len(a.resource.Schema),
		RequiredFields:  len(a.GetRequiredFields()),
		OptionalFields:  len(a.GetOptionalFields()),
		ComputedFields:  len(a.GetComputedFields()),
		ValidatedFields: len(a.GetValidatedFields()),
		ComplexityScore: a.AnalyzeComplexity(),
		SuggestedTests:  a.SuggestTestCount(),
	}

	// Count nested resources
	for _, field := range a.resource.Schema {
		if field.Type == schema.TypeSet || field.Type == schema.TypeList {
			if _, ok := field.Elem.(*schema.Resource); ok {
				report.NestedResources++
			}
		}
	}

	// Count conflicting fields
	report.ConflictingFields = len(a.GetConflictingFields())

	return report
}
