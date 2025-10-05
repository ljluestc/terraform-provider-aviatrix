# Integration Tests

Integration tests for Terraform Provider Aviatrix resources and data sources.

## Overview

Integration tests verify that individual Terraform resources and data sources work correctly against real cloud infrastructure and the Aviatrix Controller. These tests require:

- Active Aviatrix Controller
- Cloud provider credentials (AWS, Azure, GCP, OCI)
- TF_ACC=1 environment variable set

## Directory Structure

```
integration/
├── README.md           # This file
├── account_test.go     # Account resource tests
├── gateway_test.go     # Gateway resource tests
├── spoke_test.go       # Spoke gateway tests
├── transit_test.go     # Transit gateway tests
└── ... (other resource tests)
```

## Writing Integration Tests

### Test Structure

```go
package integration

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAviatrixAccount_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testAccCheckAviatrixAccountDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccAviatrixAccountConfigBasic(),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckAviatrixAccountExists("aviatrix_account.test"),
                    resource.TestCheckResourceAttr("aviatrix_account.test", "account_name", "test-account"),
                ),
            },
        },
    })
}
```

### Test Naming Convention

- `TestAccAviatrix<Resource>_<scenario>` - Standard test naming
- Examples:
  - `TestAccAviatrixGateway_basic`
  - `TestAccAviatrixGateway_ha`
  - `TestAccAviatrixAccount_awsIAM`

### PreCheck Requirements

```go
func testAccPreCheck(t *testing.T) {
    if os.Getenv("TF_ACC") != "1" {
        t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
    }
    if os.Getenv("AVIATRIX_CONTROLLER_IP") == "" {
        t.Fatal("AVIATRIX_CONTROLLER_IP must be set for acceptance tests")
    }
    if os.Getenv("AVIATRIX_USERNAME") == "" {
        t.Fatal("AVIATRIX_USERNAME must be set for acceptance tests")
    }
    if os.Getenv("AVIATRIX_PASSWORD") == "" {
        t.Fatal("AVIATRIX_PASSWORD must be set for acceptance tests")
    }
}
```

## Running Integration Tests

### Run All Integration Tests
```bash
TF_ACC=1 make testacc
```

### Run Specific Resource Tests
```bash
# Test specific resource
TF_ACC=1 go test -v ./tests/integration -run TestAccAviatrixGateway

# Test specific scenario
TF_ACC=1 go test -v ./tests/integration -run TestAccAviatrixGateway_basic
```

### Run Tests for Specific Cloud Provider
```bash
# AWS only
SKIP_ACCOUNT_AZURE=yes SKIP_ACCOUNT_GCP=yes SKIP_ACCOUNT_OCI=yes \
TF_ACC=1 go test -v ./tests/integration -run TestAccAviatrixGateway

# Azure only
SKIP_ACCOUNT_AWS=yes SKIP_ACCOUNT_GCP=yes SKIP_ACCOUNT_OCI=yes \
TF_ACC=1 go test -v ./tests/integration -run TestAccAviatrixAzureGateway
```

### Run with Timeout
```bash
TF_ACC=1 go test -v -timeout 60m ./tests/integration
```

## Environment Variables

### Required
- `TF_ACC=1` - Enable acceptance tests
- `AVIATRIX_CONTROLLER_IP` - Controller IP/hostname
- `AVIATRIX_USERNAME` - Controller username
- `AVIATRIX_PASSWORD` - Controller password

### Cloud Provider Credentials

**AWS:**
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_ACCOUNT_NUMBER`

**Azure:**
- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`

**GCP:**
- `GOOGLE_APPLICATION_CREDENTIALS` (path to JSON key file)
- `GOOGLE_PROJECT`

**OCI:**
- `OCI_USER_ID`
- `OCI_TENANCY_ID`
- `OCI_FINGERPRINT`
- `OCI_PRIVATE_KEY_PATH`
- `OCI_REGION`

## Test Organization

### By Resource Type
- Account tests: `account_test.go`
- Gateway tests: `gateway_test.go`, `spoke_gateway_test.go`, `transit_gateway_test.go`
- VPC tests: `vpc_test.go`, `azure_vnet_test.go`, `gcp_vpc_test.go`
- Firewall tests: `firewall_test.go`, `firewall_policy_test.go`
- Site2Cloud tests: `site2cloud_test.go`

### By Test Complexity
- **Basic**: Single resource creation and verification
- **HA**: High availability configurations
- **Update**: Resource update and modification
- **Import**: Resource import functionality
- **Complex**: Multi-attribute and advanced configurations

## Best Practices

1. **Cleanup**: Always implement CheckDestroy functions
2. **Idempotency**: Verify resources can be created repeatedly
3. **Updates**: Test resource updates where applicable
4. **Import**: Test import functionality for all resources
5. **Error Cases**: Test validation and error handling
6. **Parallelization**: Use `t.Parallel()` where safe
7. **Timeouts**: Set appropriate timeouts for long-running operations

## Common Patterns

### Testing Resource Creation
```go
resource.TestCheckResourceAttr("aviatrix_gateway.test", "gw_name", "test-gateway"),
resource.TestCheckResourceAttr("aviatrix_gateway.test", "gw_size", "t2.micro"),
```

### Testing Resource Updates
```go
Steps: []resource.TestStep{
    {
        Config: testAccConfigBasic(),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr("aviatrix_gateway.test", "gw_size", "t2.micro"),
        ),
    },
    {
        Config: testAccConfigUpdated(),
        Check: resource.ComposeTestCheckFunc(
            resource.TestCheckResourceAttr("aviatrix_gateway.test", "gw_size", "t2.small"),
        ),
    },
}
```

### Testing Resource Import
```go
{
    ResourceName:      "aviatrix_gateway.test",
    ImportState:       true,
    ImportStateVerify: true,
}
```

## Troubleshooting

### Tests Hang
- Check network connectivity to Aviatrix Controller
- Verify cloud provider credentials are valid
- Increase timeout with `-timeout` flag

### Tests Fail with Auth Errors
- Verify `AVIATRIX_CONTROLLER_IP`, `AVIATRIX_USERNAME`, `AVIATRIX_PASSWORD`
- Check controller is accessible from test environment

### Resource Already Exists Errors
- Ensure proper cleanup from previous test runs
- Check CheckDestroy implementation
- Manually clean up resources if needed

## CI/CD Integration

Integration tests run automatically in GitHub Actions:
- On pull requests (limited provider matrix)
- On main branch pushes (full provider matrix)
- Nightly (complete test suite)

See `.github/workflows/test-matrix.yml` for configuration.
