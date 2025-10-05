package aviatrix

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

// TestDataProvider generates synthetic test data for integration tests
type TestDataProvider interface {
	GenerateResourceName(prefix string) string
	GenerateAccountData(cloudType int) map[string]interface{}
	GenerateVPCData(cloudType int) map[string]interface{}
	GenerateGatewayData(cloudType int) map[string]interface{}
	GenerateTransitGatewayData(cloudType int) map[string]interface{}
	GenerateSpokeGatewayData(cloudType int) map[string]interface{}
}

// DefaultTestDataProvider implements TestDataProvider with default logic
type DefaultTestDataProvider struct {
	randomSeed int64
	rand       *rand.Rand
}

// NewTestDataProvider creates a new test data provider
func NewTestDataProvider() TestDataProvider {
	seed := time.Now().UnixNano()
	return &DefaultTestDataProvider{
		randomSeed: seed,
		rand:       rand.New(rand.NewSource(seed)),
	}
}

// GenerateResourceName generates a unique resource name
func (p *DefaultTestDataProvider) GenerateResourceName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, acctest.RandString(5))
}

// GenerateAccountData generates test data for account resources
func (p *DefaultTestDataProvider) GenerateAccountData(cloudType int) map[string]interface{} {
	accountName := p.GenerateResourceName("tf-acc")

	switch cloudType {
	case 1: // AWS
		return map[string]interface{}{
			"account_name":       accountName,
			"cloud_type":         1,
			"aws_account_number": integrationGetEnvOrDefault("AWS_ACCOUNT_NUMBER", "123456789012"),
			"aws_iam":            false,
			"aws_access_key":     integrationGetEnvOrDefault("AWS_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE"),
			"aws_secret_key":     integrationGetEnvOrDefault("AWS_SECRET_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		}

	case 4: // GCP
		return map[string]interface{}{
			"account_name":                 accountName,
			"cloud_type":                   4,
			"gcloud_project_id":            integrationGetEnvOrDefault("GCP_ID", "aviatrix-project"),
			"gcloud_project_credentials":   integrationGetEnvOrDefault("GCP_CREDENTIALS_FILEPATH", "/path/to/creds.json"),
		}

	case 8: // Azure
		return map[string]interface{}{
			"account_name":           accountName,
			"cloud_type":             8,
			"arm_subscription_id":    integrationGetEnvOrDefault("ARM_SUBSCRIPTION_ID", "00000000-0000-0000-0000-000000000000"),
			"arm_directory_id":       integrationGetEnvOrDefault("ARM_DIRECTORY_ID", "00000000-0000-0000-0000-000000000000"),
			"arm_application_id":     integrationGetEnvOrDefault("ARM_APPLICATION_ID", "00000000-0000-0000-0000-000000000000"),
			"arm_application_key":    integrationGetEnvOrDefault("ARM_APPLICATION_KEY", "example-key"),
		}

	case 16: // OCI
		return map[string]interface{}{
			"account_name":       accountName,
			"cloud_type":         16,
			"oci_tenancy_id":     integrationGetEnvOrDefault("OCI_TENANCY_ID", "ocid1.tenancy.example"),
			"oci_user_id":        integrationGetEnvOrDefault("OCI_USER_ID", "ocid1.user.example"),
			"oci_compartment_id": integrationGetEnvOrDefault("OCI_COMPARTMENT_ID", "ocid1.compartment.example"),
			"oci_api_private_key_filepath": integrationGetEnvOrDefault("OCI_API_KEY_FILEPATH", "/path/to/key.pem"),
		}

	default:
		return map[string]interface{}{
			"account_name": accountName,
			"cloud_type":   cloudType,
		}
	}
}

// GenerateVPCData generates test data for VPC resources
func (p *DefaultTestDataProvider) GenerateVPCData(cloudType int) map[string]interface{} {
	vpcName := p.GenerateResourceName("tf-vpc")

	switch cloudType {
	case 1: // AWS
		return map[string]interface{}{
			"cloud_type":   1,
			"account_name": p.GenerateResourceName("tf-acc"),
			"region":       integrationGetEnvOrDefault("AWS_REGION", "us-east-1"),
			"name":         vpcName,
			"cidr":         "10.0.0.0/16",
		}

	case 4: // GCP
		return map[string]interface{}{
			"cloud_type":   4,
			"account_name": p.GenerateResourceName("tf-acc"),
			"name":         vpcName,
			"subnets": []map[string]interface{}{
				{
					"region": integrationGetEnvOrDefault("GCP_ZONE", "us-central1-a"),
					"cidr":   "10.0.0.0/24",
				},
			},
		}

	case 8: // Azure
		return map[string]interface{}{
			"cloud_type":   8,
			"account_name": p.GenerateResourceName("tf-acc"),
			"region":       integrationGetEnvOrDefault("AZURE_REGION", "East US"),
			"name":         vpcName,
			"cidr":         "10.0.0.0/16",
		}

	case 16: // OCI
		return map[string]interface{}{
			"cloud_type":   16,
			"account_name": p.GenerateResourceName("tf-acc"),
			"region":       integrationGetEnvOrDefault("OCI_REGION", "us-ashburn-1"),
			"name":         vpcName,
			"cidr":         "10.0.0.0/16",
		}

	default:
		return map[string]interface{}{
			"cloud_type":   cloudType,
			"account_name": p.GenerateResourceName("tf-acc"),
			"name":         vpcName,
		}
	}
}

// GenerateGatewayData generates test data for gateway resources
func (p *DefaultTestDataProvider) GenerateGatewayData(cloudType int) map[string]interface{} {
	gwName := p.GenerateResourceName("tf-gw")

	baseData := map[string]interface{}{
		"cloud_type":   cloudType,
		"account_name": p.GenerateResourceName("tf-acc"),
		"gw_name":      gwName,
		"vpc_id":       integrationGetEnvOrDefault("AWS_VPC_ID", "vpc-0123456789abcdef0"),
		"vpc_reg":      integrationGetEnvOrDefault("AWS_REGION", "us-east-1"),
		"subnet":       integrationGetEnvOrDefault("AWS_SUBNET", "10.0.1.0/24"),
	}

	switch cloudType {
	case 1: // AWS
		baseData["gw_size"] = integrationGetEnvOrDefault("AWS_GW_SIZE", "t2.micro")

	case 4: // GCP
		baseData["gw_size"] = integrationGetEnvOrDefault("GCP_GW_SIZE", "n1-standard-1")
		baseData["vpc_id"] = integrationGetEnvOrDefault("GCP_VPC_ID", "gcp-vpc")
		baseData["vpc_reg"] = integrationGetEnvOrDefault("GCP_ZONE", "us-central1-a")
		baseData["subnet"] = integrationGetEnvOrDefault("GCP_SUBNET", "10.0.1.0/24")

	case 8: // Azure
		baseData["gw_size"] = integrationGetEnvOrDefault("AZURE_GW_SIZE", "Standard_B1s")
		baseData["vpc_id"] = integrationGetEnvOrDefault("AZURE_VNET_ID", "azure-vnet:resource-group")
		baseData["vpc_reg"] = integrationGetEnvOrDefault("AZURE_REGION", "East US")
		baseData["subnet"] = integrationGetEnvOrDefault("AZURE_SUBNET", "10.0.1.0/24")

	case 16: // OCI
		baseData["gw_size"] = integrationGetEnvOrDefault("OCI_GW_SIZE", "VM.Standard2.2")
		baseData["vpc_id"] = integrationGetEnvOrDefault("OCI_VPC_ID", "ocid1.vcn.example")
		baseData["vpc_reg"] = integrationGetEnvOrDefault("OCI_REGION", "us-ashburn-1")
		baseData["subnet"] = integrationGetEnvOrDefault("OCI_SUBNET", "10.0.1.0/24")
	}

	return baseData
}

// GenerateTransitGatewayData generates test data for transit gateway resources
func (p *DefaultTestDataProvider) GenerateTransitGatewayData(cloudType int) map[string]interface{} {
	data := p.GenerateGatewayData(cloudType)
	data["gw_name"] = p.GenerateResourceName("tf-transit-gw")
	data["enable_hybrid_connection"] = true
	data["connected_transit"] = true

	return data
}

// GenerateSpokeGatewayData generates test data for spoke gateway resources
func (p *DefaultTestDataProvider) GenerateSpokeGatewayData(cloudType int) map[string]interface{} {
	data := p.GenerateGatewayData(cloudType)
	data["gw_name"] = p.GenerateResourceName("tf-spoke-gw")
	data["single_az_ha"] = false

	return data
}

// TestConfigTemplateEngine generates Terraform configurations from templates
type TestConfigTemplateEngine struct {
	templates map[string]string
}

// NewTestConfigTemplateEngine creates a new template engine
func NewTestConfigTemplateEngine() *TestConfigTemplateEngine {
	engine := &TestConfigTemplateEngine{
		templates: make(map[string]string),
	}
	engine.loadDefaultTemplates()
	return engine
}

// loadDefaultTemplates loads commonly used test configuration templates
func (e *TestConfigTemplateEngine) loadDefaultTemplates() {
	// AWS Account Template
	e.templates["aws_account"] = `
resource "aviatrix_account" "{{.Name}}" {
  account_name       = "{{.AccountName}}"
  cloud_type         = 1
  aws_account_number = "{{.AWSAccountNumber}}"
  aws_iam            = {{.AWSIam}}
  aws_access_key     = "{{.AWSAccessKey}}"
  aws_secret_key     = "{{.AWSSecretKey}}"
}
`

	// AWS VPC Template
	e.templates["aws_vpc"] = `
resource "aviatrix_vpc" "{{.Name}}" {
  cloud_type   = 1
  account_name = {{.AccountRef}}
  region       = "{{.Region}}"
  name         = "{{.VPCName}}"
  cidr         = "{{.CIDR}}"
}
`

	// AWS Gateway Template
	e.templates["aws_gateway"] = `
resource "aviatrix_gateway" "{{.Name}}" {
  cloud_type   = 1
  account_name = {{.AccountRef}}
  gw_name      = "{{.GatewayName}}"
  vpc_id       = {{.VPCRef}}
  vpc_reg      = "{{.Region}}"
  gw_size      = "{{.Size}}"
  subnet       = "{{.Subnet}}"
}
`

	// Transit Gateway Template
	e.templates["transit_gateway"] = `
resource "aviatrix_transit_gateway" "{{.Name}}" {
  cloud_type               = {{.CloudType}}
  account_name             = {{.AccountRef}}
  gw_name                  = "{{.GatewayName}}"
  vpc_id                   = {{.VPCRef}}
  vpc_reg                  = "{{.Region}}"
  gw_size                  = "{{.Size}}"
  subnet                   = "{{.Subnet}}"
  enable_hybrid_connection = {{.EnableHybrid}}
  connected_transit        = {{.ConnectedTransit}}
}
`

	// Spoke Gateway Template
	e.templates["spoke_gateway"] = `
resource "aviatrix_spoke_gateway" "{{.Name}}" {
  cloud_type   = {{.CloudType}}
  account_name = {{.AccountRef}}
  gw_name      = "{{.GatewayName}}"
  vpc_id       = {{.VPCRef}}
  vpc_reg      = "{{.Region}}"
  gw_size      = "{{.Size}}"
  subnet       = "{{.Subnet}}"
  transit_gw   = {{.TransitGWRef}}
}
`

	// Spoke Transit Attachment Template
	e.templates["spoke_transit_attachment"] = `
resource "aviatrix_spoke_transit_attachment" "{{.Name}}" {
  spoke_gw_name   = {{.SpokeGWRef}}
  transit_gw_name = {{.TransitGWRef}}
}
`
}

// RenderTemplate renders a template with provided data
func (e *TestConfigTemplateEngine) RenderTemplate(templateName string, data map[string]interface{}) (string, error) {
	template, ok := e.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result, nil
}

// AddTemplate adds a custom template
func (e *TestConfigTemplateEngine) AddTemplate(name, template string) {
	e.templates[name] = template
}

// TestFixtureManager manages test fixtures and cleanup
type TestFixtureManager struct {
	fixtures map[string]interface{}
	cleanup  []func() error
}

// NewTestFixtureManager creates a new fixture manager
func NewTestFixtureManager() *TestFixtureManager {
	return &TestFixtureManager{
		fixtures: make(map[string]interface{}),
		cleanup:  []func() error{},
	}
}

// RegisterFixture registers a fixture with an optional cleanup function
func (m *TestFixtureManager) RegisterFixture(name string, fixture interface{}, cleanupFunc func() error) {
	m.fixtures[name] = fixture
	if cleanupFunc != nil {
		m.cleanup = append(m.cleanup, cleanupFunc)
	}
}

// GetFixture retrieves a fixture by name
func (m *TestFixtureManager) GetFixture(name string) (interface{}, bool) {
	fixture, ok := m.fixtures[name]
	return fixture, ok
}

// Cleanup executes all registered cleanup functions
func (m *TestFixtureManager) Cleanup() []error {
	var errors []error

	// Execute cleanup in reverse order (LIFO)
	for i := len(m.cleanup) - 1; i >= 0; i-- {
		if err := m.cleanup[i](); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// TestDataValidator validates test data against schema constraints
type TestDataValidator struct{}

// NewTestDataValidator creates a new data validator
func NewTestDataValidator() *TestDataValidator {
	return &TestDataValidator{}
}

// ValidateCIDR validates CIDR notation
func (v *TestDataValidator) ValidateCIDR(cidr string) error {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid CIDR format: %s", cidr)
	}

	// Basic IP validation
	ipParts := strings.Split(parts[0], ".")
	if len(ipParts) != 4 {
		return fmt.Errorf("invalid IP address in CIDR: %s", parts[0])
	}

	return nil
}

// ValidateAccountName validates account name format
func (v *TestDataValidator) ValidateAccountName(name string) error {
	if name == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("account name too long: %d characters (max 255)", len(name))
	}

	return nil
}

// ValidateRegion validates cloud region format
func (v *TestDataValidator) ValidateRegion(cloudType int, region string) error {
	if region == "" {
		return fmt.Errorf("region cannot be empty")
	}

	// Cloud-specific validation could be added here
	return nil
}

// integrationGetEnvOrDefault is a helper function for test data generation
func integrationGetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
