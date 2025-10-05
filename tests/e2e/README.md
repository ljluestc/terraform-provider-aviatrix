# End-to-End (E2E) Tests

End-to-end tests for complex, multi-resource workflows in the Terraform Provider Aviatrix.

## Overview

E2E tests validate complete workflows that span multiple resources and represent real-world use cases. Unlike integration tests that focus on individual resources, E2E tests verify:

- Multi-resource orchestration
- Complex dependency chains
- Complete networking workflows
- Cross-cloud provider scenarios
- Production-like configurations

## Directory Structure

```
e2e/
├── README.md                      # This file
├── multi_cloud_transit_test.go    # Multi-cloud transit network E2E tests
├── site2cloud_workflow_test.go    # Site2Cloud workflow tests
├── firewall_policy_test.go        # Complete firewall policy workflows
└── ... (other E2E workflows)
```

## E2E Test Categories

### Multi-Cloud Transit Networks
Complete transit network setup across multiple cloud providers:
- Transit gateways in hub clouds
- Spoke gateways in various regions
- Peering connections
- Route management

### Site2Cloud Workflows
End-to-end site-to-cloud connectivity:
- VPC/VNet creation
- Gateway deployment
- Tunnel configuration
- Route propagation

### Firewall Policy Management
Complete firewall policy lifecycle:
- Policy creation and management
- Rule sets and tagging
- Policy attachments
- Traffic validation

### High Availability Workflows
Complete HA deployment and failover:
- Primary gateway setup
- HA gateway configuration
- Failover testing
- Recovery validation

## Writing E2E Tests

### Test Structure

```go
package e2e

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestE2E_MultiCloudTransit(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    resource.Test(t, resource.TestCase{
        PreCheck:     func() { testE2EPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testE2ECheckTransitNetworkDestroy,
        Steps: []resource.TestStep{
            {
                Config: testE2EMultiCloudTransitConfig(),
                Check: resource.ComposeTestCheckFunc(
                    testE2ECheckTransitGatewayExists("aviatrix_transit_gateway.aws"),
                    testE2ECheckTransitGatewayExists("aviatrix_transit_gateway.azure"),
                    testE2ECheckSpokeGatewayExists("aviatrix_spoke_gateway.app"),
                    testE2ECheckPeeringExists("aviatrix_transit_gateway_peering.aws_azure"),
                    testE2EValidateConnectivity(),
                ),
            },
        },
    })
}
```

### Test Naming Convention

- `TestE2E_<Workflow>_<scenario>` - E2E test naming
- Examples:
  - `TestE2E_MultiCloudTransit_basic`
  - `TestE2E_Site2Cloud_ha`
  - `TestE2E_FirewallPolicy_multiRule`

## Running E2E Tests

### Run All E2E Tests
```bash
# Run all E2E tests
make teste2e

# Or directly with go test
TF_ACC=1 go test -v -timeout 120m ./tests/e2e
```

### Run Specific E2E Tests
```bash
# Run specific workflow
TF_ACC=1 go test -v -timeout 120m ./tests/e2e -run TestE2E_MultiCloudTransit

# Run with extended timeout
TF_ACC=1 go test -v -timeout 180m ./tests/e2e -run TestE2E_ComplexNetworking
```

### Run in Docker
```bash
# Build and run E2E tests in Docker
docker-compose -f docker-compose.test.yml run --rm e2e-tests

# Run specific E2E test in Docker
docker-compose -f docker-compose.test.yml run --rm e2e-tests \
  go test -v -timeout 120m ./tests/e2e -run TestE2E_MultiCloudTransit
```

## Environment Requirements

### All Integration Test Requirements Plus:
- Extended timeouts (E2E tests can take 30-120 minutes)
- Multiple cloud provider credentials configured simultaneously
- Sufficient cloud provider quotas for multi-resource deployments
- Network connectivity for cross-cloud testing

### Recommended Configuration
```bash
# Extended timeout
export GO_TEST_TIMEOUT=120m

# Enable all cloud providers
export SKIP_ACCOUNT_AWS=no
export SKIP_ACCOUNT_AZURE=no
export SKIP_ACCOUNT_GCP=no
export SKIP_ACCOUNT_OCI=no

# Enable detailed logging
export TF_LOG=DEBUG
export ENABLE_DETAILED_LOGS=true
```

## E2E Test Patterns

### Multi-Step Workflows
```go
Steps: []resource.TestStep{
    // Step 1: Create base infrastructure
    {
        Config: testE2EConfigStep1_BaseInfra(),
        Check: resource.ComposeTestCheckFunc(
            testE2ECheckVPCExists(),
        ),
    },
    // Step 2: Deploy gateways
    {
        Config: testE2EConfigStep2_Gateways(),
        Check: resource.ComposeTestCheckFunc(
            testE2ECheckGatewayDeployed(),
        ),
    },
    // Step 3: Configure connectivity
    {
        Config: testE2EConfigStep3_Connectivity(),
        Check: resource.ComposeTestCheckFunc(
            testE2EValidateFullConnectivity(),
        ),
    },
}
```

### Cross-Cloud Validation
```go
func testE2EValidateConnectivity() resource.TestCheckFunc {
    return func(s *terraform.State) error {
        // Validate AWS to Azure connectivity
        if err := validateCrossCloudPing("aws-spoke", "azure-spoke"); err != nil {
            return fmt.Errorf("AWS to Azure connectivity failed: %v", err)
        }
        // Validate route propagation
        if err := validateRoutePropagation(); err != nil {
            return fmt.Errorf("route propagation failed: %v", err)
        }
        return nil
    }
}
```

## E2E Test Scenarios

### 1. Multi-Cloud Transit Network
**File**: `multi_cloud_transit_test.go`
**Duration**: ~45-60 minutes
**Resources**:
- 2 Transit Gateways (AWS, Azure)
- 4 Spoke Gateways (various clouds/regions)
- Transit peering
- Spoke attachments

### 2. Site2Cloud with HA
**File**: `site2cloud_workflow_test.go`
**Duration**: ~30-45 minutes
**Resources**:
- VPC/VNet creation
- Primary and HA gateways
- Site2Cloud connections
- Route tables

### 3. Complete Firewall Policy
**File**: `firewall_policy_test.go`
**Duration**: ~20-30 minutes
**Resources**:
- Firewall policies
- Multiple rule sets
- Gateway attachments
- Traffic validation

### 4. FQDN Filtering Workflow
**File**: `fqdn_filtering_test.go`
**Duration**: ~25-35 minutes
**Resources**:
- FQDN tags
- Filtering rules
- Gateway attachments
- DNS validation

## Best Practices

1. **Cleanup**: Implement thorough CheckDestroy for all resources
2. **Timeouts**: Use generous timeouts (60-120 minutes)
3. **Ordering**: Test resource creation in dependency order
4. **Validation**: Validate actual connectivity, not just resource state
5. **Idempotency**: Ensure workflows can be re-run safely
6. **Error Handling**: Gracefully handle partial failures
7. **Logging**: Enable detailed logging for debugging
8. **Isolation**: Use unique naming to avoid conflicts

## Common Test Utilities

### Connectivity Validation
```go
// Validate cross-cloud connectivity
func validateCrossCloudPing(source, dest string) error {
    // Implementation
}

// Validate route propagation
func validateRoutePropagation() error {
    // Implementation
}
```

### Resource Cleanup
```go
// Ensure all resources are destroyed
func testE2ECheckTransitNetworkDestroy(s *terraform.State) error {
    client := testAccProvider.Meta().(*goaviatrix.Client)

    // Check transit gateways destroyed
    // Check spoke gateways destroyed
    // Check peering destroyed

    return nil
}
```

## Troubleshooting

### Test Timeouts
- Increase timeout with `-timeout 180m`
- Check cloud provider API rate limits
- Verify network connectivity

### Partial Failures
- Review detailed logs with `TF_LOG=DEBUG`
- Check resource dependencies
- Manually clean up orphaned resources

### Connectivity Validation Failures
- Verify security groups allow required traffic
- Check route table configurations
- Validate gateway status

## CI/CD Integration

E2E tests run in GitHub Actions:
- **On PR**: Limited E2E test suite (smoke tests)
- **On merge**: Full E2E test suite
- **Nightly**: Complete E2E validation across all clouds

### CI Configuration
```yaml
- name: Run E2E Tests
  run: |
    TF_ACC=1 go test -v -timeout 120m ./tests/e2e
  env:
    ENABLE_DETAILED_LOGS: true
    # All cloud provider credentials
```

## Performance Considerations

- E2E tests are resource-intensive
- Run in parallel where possible
- Use `t.Parallel()` for independent workflows
- Monitor cloud provider costs
- Clean up resources promptly

## Documentation

- [Complete Testing PRD](../../COMPLETE_TESTING_PRD.md)
- [Test Infrastructure](../../TEST_INFRASTRUCTURE.md)
- [Integration Tests](../integration/README.md)
