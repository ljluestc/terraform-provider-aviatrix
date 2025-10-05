# Test Fixtures

Reusable test fixtures, configuration templates, and mock data for Terraform Provider Aviatrix tests.

## Overview

The fixtures directory contains:
- Terraform configuration templates
- JSON/YAML test data files
- Mock API responses
- Provider configuration examples
- Common resource definitions

## Directory Structure

```
fixtures/
├── README.md                    # This file
├── terraform/                   # Terraform configuration templates
│   ├── account/                # Account resource configs
│   ├── gateway/                # Gateway resource configs
│   ├── vpc/                    # VPC/VNet resource configs
│   └── ...
├── data/                       # Test data files
│   ├── accounts.json           # Sample account data
│   ├── gateways.json           # Sample gateway configurations
│   └── ...
├── mocks/                      # Mock API responses
│   ├── controller/             # Aviatrix Controller mocks
│   └── cloud_providers/        # Cloud provider API mocks
└── configs/                    # Complete configuration examples
    ├── multi_cloud_transit.tf  # Multi-cloud transit setup
    ├── site2cloud_ha.tf        # Site2Cloud with HA
    └── ...
```

## Terraform Configuration Templates

### Account Templates

**File**: `terraform/account/aws_iam.tf`
```hcl
resource "aviatrix_account" "aws_iam" {
  account_name       = var.account_name
  cloud_type         = 1
  aws_account_number = var.aws_account_number
  aws_iam            = true
  aws_role_arn       = var.aws_role_arn
  aws_role_ec2       = var.aws_role_ec2
}
```

**File**: `terraform/account/azure_arm.tf`
```hcl
resource "aviatrix_account" "azure_arm" {
  account_name        = var.account_name
  cloud_type          = 8
  arm_subscription_id = var.subscription_id
  arm_directory_id    = var.directory_id
  arm_application_id  = var.application_id
  arm_application_key = var.application_key
}
```

### Gateway Templates

**File**: `terraform/gateway/aws_gateway.tf`
```hcl
resource "aviatrix_gateway" "aws" {
  cloud_type   = 1
  account_name = var.account_name
  gw_name      = var.gateway_name
  vpc_id       = var.vpc_id
  vpc_reg      = var.vpc_region
  gw_size      = var.gateway_size
  subnet       = var.subnet
}
```

**File**: `terraform/gateway/azure_gateway.tf`
```hcl
resource "aviatrix_gateway" "azure" {
  cloud_type   = 8
  account_name = var.account_name
  gw_name      = var.gateway_name
  vpc_id       = var.vnet_name
  vpc_reg      = var.vnet_region
  gw_size      = var.gateway_size
  subnet       = var.subnet
}
```

## Test Data Files

### Account Data

**File**: `data/accounts.json`
```json
{
  "aws": {
    "account_name": "test-aws-account",
    "cloud_type": 1,
    "aws_account_number": "123456789012"
  },
  "azure": {
    "account_name": "test-azure-account",
    "cloud_type": 8,
    "arm_subscription_id": "12345678-1234-1234-1234-123456789012"
  },
  "gcp": {
    "account_name": "test-gcp-account",
    "cloud_type": 4,
    "gcloud_project_id": "test-project-123456"
  }
}
```

### Gateway Configurations

**File**: `data/gateways.json`
```json
{
  "sizes": {
    "aws": ["t2.micro", "t2.small", "t2.medium", "c5.large", "c5.xlarge"],
    "azure": ["Standard_B1ms", "Standard_B2s", "Standard_D3_v2"],
    "gcp": ["n1-standard-1", "n1-standard-2", "n1-standard-4"]
  },
  "regions": {
    "aws": ["us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"],
    "azure": ["East US", "West US 2", "West Europe", "Southeast Asia"],
    "gcp": ["us-central1", "us-west1", "europe-west1", "asia-southeast1"]
  }
}
```

### VPC/VNet Data

**File**: `data/vpcs.json`
```json
{
  "aws": {
    "cidr": "10.0.0.0/16",
    "subnets": [
      { "cidr": "10.0.1.0/24", "az": "us-east-1a" },
      { "cidr": "10.0.2.0/24", "az": "us-east-1b" }
    ]
  },
  "azure": {
    "address_space": ["10.1.0.0/16"],
    "subnets": [
      { "address_prefix": "10.1.1.0/24", "name": "gateway-subnet" }
    ]
  }
}
```

## Mock API Responses

### Aviatrix Controller Mocks

**File**: `mocks/controller/login_response.json`
```json
{
  "return": true,
  "results": {
    "CID": "mock-cid-12345",
    "api_version": "v1"
  }
}
```

**File**: `mocks/controller/gateway_create_response.json`
```json
{
  "return": true,
  "results": {
    "gateway_name": "test-gateway",
    "public_ip": "1.2.3.4",
    "private_ip": "10.0.1.5"
  }
}
```

### Cloud Provider Mocks

**File**: `mocks/cloud_providers/aws_vpc_describe.json`
```json
{
  "Vpcs": [
    {
      "VpcId": "vpc-12345678",
      "CidrBlock": "10.0.0.0/16",
      "State": "available"
    }
  ]
}
```

## Complete Configuration Examples

### Multi-Cloud Transit Network

**File**: `configs/multi_cloud_transit.tf`
```hcl
# AWS Transit Gateway
resource "aviatrix_transit_gateway" "aws" {
  cloud_type   = 1
  account_name = "test-aws"
  gw_name      = "aws-transit-gw"
  vpc_id       = "vpc-12345"
  vpc_reg      = "us-east-1"
  gw_size      = "c5.xlarge"
  subnet       = "10.0.1.0/24"
}

# Azure Transit Gateway
resource "aviatrix_transit_gateway" "azure" {
  cloud_type   = 8
  account_name = "test-azure"
  gw_name      = "azure-transit-gw"
  vpc_id       = "test-vnet:test-rg"
  vpc_reg      = "East US"
  gw_size      = "Standard_D3_v2"
  subnet       = "10.1.1.0/24"
}

# Transit Gateway Peering
resource "aviatrix_transit_gateway_peering" "aws_azure" {
  transit_gateway_name1 = aviatrix_transit_gateway.aws.gw_name
  transit_gateway_name2 = aviatrix_transit_gateway.azure.gw_name
}
```

### Site2Cloud with HA

**File**: `configs/site2cloud_ha.tf`
```hcl
# Primary Gateway
resource "aviatrix_gateway" "primary" {
  cloud_type   = 1
  account_name = "test-aws"
  gw_name      = "s2c-primary-gw"
  vpc_id       = "vpc-12345"
  vpc_reg      = "us-east-1"
  gw_size      = "t2.medium"
  subnet       = "10.0.1.0/24"
}

# HA Gateway
resource "aviatrix_gateway" "ha" {
  cloud_type     = 1
  account_name   = "test-aws"
  gw_name        = "s2c-ha-gw"
  vpc_id         = "vpc-12345"
  vpc_reg        = "us-east-1"
  gw_size        = "t2.medium"
  subnet         = "10.0.2.0/24"
  peering_ha_gw  = aviatrix_gateway.primary.gw_name
}

# Site2Cloud Connection
resource "aviatrix_site2cloud" "connection" {
  vpc_id                     = "vpc-12345"
  connection_name            = "test-s2c"
  connection_type            = "unmapped"
  remote_gateway_type        = "generic"
  tunnel_type                = "route"
  primary_cloud_gateway_name = aviatrix_gateway.primary.gw_name
  remote_subnet_cidr         = "192.168.0.0/16"
  ha_enabled                 = true
}
```

## Using Fixtures in Tests

### Loading Terraform Templates
```go
func testAccAviatrixGatewayConfigBasic(name string) string {
    template, _ := os.ReadFile("tests/fixtures/terraform/gateway/aws_gateway.tf")
    return fmt.Sprintf(string(template), name)
}
```

### Loading Test Data
```go
func loadAccountFixture(provider string) map[string]interface{} {
    data, _ := os.ReadFile("tests/fixtures/data/accounts.json")
    var accounts map[string]map[string]interface{}
    json.Unmarshal(data, &accounts)
    return accounts[provider]
}
```

### Using Mock Responses
```go
func mockControllerLogin() *http.Response {
    data, _ := os.ReadFile("tests/fixtures/mocks/controller/login_response.json")
    return &http.Response{
        StatusCode: 200,
        Body:       io.NopCloser(bytes.NewReader(data)),
    }
}
```

## Fixture Maintenance

### Adding New Fixtures
1. Create fixture file in appropriate subdirectory
2. Follow naming convention: `<resource>_<variant>.tf` or `<resource>.json`
3. Include comments explaining usage
4. Add to this README

### Updating Fixtures
1. Update fixture file
2. Update any tests using the fixture
3. Document changes in this README
4. Ensure backward compatibility where possible

## Best Practices

1. **Modularity**: Create small, reusable fixtures
2. **Naming**: Use descriptive, consistent names
3. **Comments**: Document fixture purpose and usage
4. **Variables**: Use variables for configurable values
5. **Validation**: Ensure fixtures are syntactically correct
6. **Versioning**: Track fixture changes with tests

## Contributing

When adding fixtures:
1. Place in appropriate subdirectory
2. Follow existing patterns and conventions
3. Include inline documentation
4. Add usage examples to README
5. Ensure fixtures are tested
