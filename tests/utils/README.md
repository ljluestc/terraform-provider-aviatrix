# Test Utilities

Shared test utilities and helper functions for Terraform Provider Aviatrix tests.

## Overview

The utils directory contains reusable helper functions, test data generators, assertion utilities, and common testing patterns used across all test types.

## Directory Structure

```
utils/
├── README.md              # This file
├── helpers.go             # General test helper functions
├── resource_helpers.go    # Resource-specific helpers
├── assertions.go          # Custom assertion functions
├── generators.go          # Test data generators
├── cleanup.go             # Resource cleanup utilities
├── validators.go          # Validation functions
└── test_config.go         # Test configuration utilities
```

## Core Utilities

### Test Helper Functions

**File**: `helpers.go`

```go
package utils

import (
    "fmt"
    "os"
    "testing"
)

// PreCheckFunc returns a function that performs common pre-checks
func PreCheckFunc(t *testing.T) func() {
    return func() {
        if v := os.Getenv("TF_ACC"); v != "1" {
            t.Skip("Acceptance tests skipped unless env 'TF_ACC' is set")
        }
        if v := os.Getenv("AVIATRIX_CONTROLLER_IP"); v == "" {
            t.Fatal("AVIATRIX_CONTROLLER_IP must be set for acceptance tests")
        }
        if v := os.Getenv("AVIATRIX_USERNAME"); v == "" {
            t.Fatal("AVIATRIX_USERNAME must be set for acceptance tests")
        }
        if v := os.Getenv("AVIATRIX_PASSWORD"); v == "" {
            t.Fatal("AVIATRIX_PASSWORD must be set for acceptance tests")
        }
    }
}

// PreCheckAWS checks AWS-specific environment variables
func PreCheckAWS(t *testing.T) {
    if v := os.Getenv("SKIP_ACCOUNT_AWS"); v == "yes" {
        t.Skip("Skipping AWS tests")
    }
    if v := os.Getenv("AWS_ACCOUNT_NUMBER"); v == "" {
        t.Fatal("AWS_ACCOUNT_NUMBER must be set for AWS tests")
    }
}

// PreCheckAzure checks Azure-specific environment variables
func PreCheckAzure(t *testing.T) {
    if v := os.Getenv("SKIP_ACCOUNT_AZURE"); v == "yes" {
        t.Skip("Skipping Azure tests")
    }
    if v := os.Getenv("ARM_SUBSCRIPTION_ID"); v == "" {
        t.Fatal("ARM_SUBSCRIPTION_ID must be set for Azure tests")
    }
}

// PreCheckGCP checks GCP-specific environment variables
func PreCheckGCP(t *testing.T) {
    if v := os.Getenv("SKIP_ACCOUNT_GCP"); v == "yes" {
        t.Skip("Skipping GCP tests")
    }
    if v := os.Getenv("GOOGLE_PROJECT"); v == "" {
        t.Fatal("GOOGLE_PROJECT must be set for GCP tests")
    }
}

// PreCheckOCI checks OCI-specific environment variables
func PreCheckOCI(t *testing.T) {
    if v := os.Getenv("SKIP_ACCOUNT_OCI"); v == "yes" {
        t.Skip("Skipping OCI tests")
    }
    if v := os.Getenv("OCI_TENANCY_ID"); v == "" {
        t.Fatal("OCI_TENANCY_ID must be set for OCI tests")
    }
}

// GetEnvOrDefault returns environment variable value or default
func GetEnvOrDefault(key, defaultValue string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultValue
}

// SkipIfNotAcceptance skips test if not in acceptance mode
func SkipIfNotAcceptance(t *testing.T) {
    if os.Getenv("TF_ACC") != "1" {
        t.Skip("Skipping acceptance test")
    }
}
```

### Resource Helper Functions

**File**: `resource_helpers.go`

```go
package utils

import (
    "fmt"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// GetResourceState retrieves resource from state by address
func GetResourceState(s *terraform.State, address string) (*terraform.ResourceState, error) {
    rs, ok := s.RootModule().Resources[address]
    if !ok {
        return nil, fmt.Errorf("resource not found: %s", address)
    }
    return rs, nil
}

// GetResourceAttr retrieves an attribute from resource state
func GetResourceAttr(s *terraform.State, address, attr string) (string, error) {
    rs, err := GetResourceState(s, address)
    if err != nil {
        return "", err
    }

    val, ok := rs.Primary.Attributes[attr]
    if !ok {
        return "", fmt.Errorf("attribute not found: %s", attr)
    }
    return val, nil
}

// ResourceExists checks if resource exists in state
func ResourceExists(s *terraform.State, address string) bool {
    _, err := GetResourceState(s, address)
    return err == nil
}

// CountResources counts resources of a specific type
func CountResources(s *terraform.State, resourceType string) int {
    count := 0
    for _, rs := range s.RootModule().Resources {
        if rs.Type == resourceType {
            count++
        }
    }
    return count
}
```

### Custom Assertions

**File**: `assertions.go`

```go
package utils

import (
    "fmt"
    "testing"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// AssertResourceAttrMatch checks if attribute matches expected value
func AssertResourceAttrMatch(address, attr, expected string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        val, err := GetResourceAttr(s, address, attr)
        if err != nil {
            return err
        }
        if val != expected {
            return fmt.Errorf("%s.%s: expected %q, got %q", address, attr, expected, val)
        }
        return nil
    }
}

// AssertResourceAttrSet checks if attribute is set (non-empty)
func AssertResourceAttrSet(address, attr string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        val, err := GetResourceAttr(s, address, attr)
        if err != nil {
            return err
        }
        if val == "" {
            return fmt.Errorf("%s.%s: attribute not set", address, attr)
        }
        return nil
    }
}

// AssertResourceCount checks if expected number of resources exist
func AssertResourceCount(resourceType string, expected int) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        count := CountResources(s, resourceType)
        if count != expected {
            return fmt.Errorf("expected %d resources of type %s, got %d",
                expected, resourceType, count)
        }
        return nil
    }
}

// AssertNoError checks that no error occurred
func AssertNoError(t *testing.T, err error, message string) {
    t.Helper()
    if err != nil {
        t.Fatalf("%s: %v", message, err)
    }
}

// AssertEqual checks if values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
    t.Helper()
    if expected != actual {
        t.Fatalf("%s: expected %v, got %v", message, expected, actual)
    }
}
```

### Test Data Generators

**File**: `generators.go`

```go
package utils

import (
    "fmt"
    "math/rand"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

// RandomString generates random alphanumeric string
func RandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

// RandomInt generates random integer in range [min, max]
func RandomInt(min, max int) int {
    return min + rand.Intn(max-min+1)
}

// GenerateResourceName generates unique resource name with prefix
func GenerateResourceName(prefix string) string {
    timestamp := time.Now().Unix()
    random := RandomString(6)
    return fmt.Sprintf("%s-%d-%s", prefix, timestamp, random)
}

// GenerateTestAccountName generates test account name
func GenerateTestAccountName() string {
    return GenerateResourceName("test-acc")
}

// GenerateTestGatewayName generates test gateway name
func GenerateTestGatewayName() string {
    return GenerateResourceName("test-gw")
}

// GenerateTestVPCName generates test VPC name
func GenerateTestVPCName() string {
    return GenerateResourceName("test-vpc")
}

// GenerateCIDR generates random CIDR block
func GenerateCIDR(prefix string, prefixLength int) string {
    if prefix == "10" {
        second := RandomInt(0, 255)
        return fmt.Sprintf("10.%d.0.0/%d", second, prefixLength)
    }
    return fmt.Sprintf("192.168.%d.0/%d", RandomInt(0, 255), prefixLength)
}
```

### Cleanup Utilities

**File**: `cleanup.go`

```go
package utils

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
)

// CleanupGateway ensures gateway is deleted
func CleanupGateway(client *goaviatrix.Client, gatewayName string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    gateway := &goaviatrix.Gateway{GwName: gatewayName}
    err := client.DeleteGateway(gateway)
    if err != nil {
        // Check if already deleted
        if _, getErr := client.GetGateway(gateway); getErr != nil {
            log.Printf("[INFO] Gateway %s already deleted", gatewayName)
            return nil
        }
        return fmt.Errorf("failed to delete gateway %s: %v", gatewayName, err)
    }

    // Wait for deletion to complete
    return waitForGatewayDeletion(ctx, client, gatewayName)
}

// CleanupAccount ensures account is deleted
func CleanupAccount(client *goaviatrix.Client, accountName string) error {
    account := &goaviatrix.Account{AccountName: accountName}
    err := client.DeleteAccount(account)
    if err != nil {
        // Check if already deleted
        if _, getErr := client.GetAccount(account); getErr != nil {
            log.Printf("[INFO] Account %s already deleted", accountName)
            return nil
        }
        return fmt.Errorf("failed to delete account %s: %v", accountName, err)
    }
    return nil
}

// waitForGatewayDeletion polls until gateway is deleted
func waitForGatewayDeletion(ctx context.Context, client *goaviatrix.Client, name string) error {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    gateway := &goaviatrix.Gateway{GwName: name}
    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("timeout waiting for gateway deletion")
        case <-ticker.C:
            _, err := client.GetGateway(gateway)
            if err != nil {
                // Gateway not found, deletion complete
                return nil
            }
        }
    }
}

// CleanupTestResources removes all test resources with prefix
func CleanupTestResources(client *goaviatrix.Client, prefix string) {
    log.Printf("[INFO] Cleaning up test resources with prefix: %s", prefix)
    // Implementation for bulk cleanup
}
```

### Validation Functions

**File**: `validators.go`

```go
package utils

import (
    "fmt"
    "net"
    "regexp"
)

// ValidateCIDR validates CIDR notation
func ValidateCIDR(cidr string) error {
    _, _, err := net.ParseCIDR(cidr)
    if err != nil {
        return fmt.Errorf("invalid CIDR: %v", err)
    }
    return nil
}

// ValidateIPAddress validates IP address
func ValidateIPAddress(ip string) error {
    if net.ParseIP(ip) == nil {
        return fmt.Errorf("invalid IP address: %s", ip)
    }
    return nil
}

// ValidateResourceName validates resource name format
func ValidateResourceName(name string) error {
    // Must be alphanumeric with hyphens, 1-255 chars
    matched, _ := regexp.MatchString("^[a-zA-Z0-9-]{1,255}$", name)
    if !matched {
        return fmt.Errorf("invalid resource name: %s", name)
    }
    return nil
}

// ValidateCloudType validates cloud type value
func ValidateCloudType(cloudType int) error {
    validTypes := map[int]string{
        1:  "AWS",
        4:  "GCP",
        8:  "Azure",
        16: "OCI",
    }
    if _, ok := validTypes[cloudType]; !ok {
        return fmt.Errorf("invalid cloud type: %d", cloudType)
    }
    return nil
}
```

### Test Configuration

**File**: `test_config.go`

```go
package utils

import (
    "os"
)

// TestConfig holds test configuration
type TestConfig struct {
    ControllerIP string
    Username     string
    Password     string

    AWSAccountNumber string
    AzureSubID       string
    GCPProject       string
    OCITenancyID     string

    SkipAWS   bool
    SkipAzure bool
    SkipGCP   bool
    SkipOCI   bool
}

// LoadTestConfig loads configuration from environment
func LoadTestConfig() *TestConfig {
    return &TestConfig{
        ControllerIP: os.Getenv("AVIATRIX_CONTROLLER_IP"),
        Username:     os.Getenv("AVIATRIX_USERNAME"),
        Password:     os.Getenv("AVIATRIX_PASSWORD"),

        AWSAccountNumber: os.Getenv("AWS_ACCOUNT_NUMBER"),
        AzureSubID:       os.Getenv("ARM_SUBSCRIPTION_ID"),
        GCPProject:       os.Getenv("GOOGLE_PROJECT"),
        OCITenancyID:     os.Getenv("OCI_TENANCY_ID"),

        SkipAWS:   os.Getenv("SKIP_ACCOUNT_AWS") == "yes",
        SkipAzure: os.Getenv("SKIP_ACCOUNT_AZURE") == "yes",
        SkipGCP:   os.Getenv("SKIP_ACCOUNT_GCP") == "yes",
        SkipOCI:   os.Getenv("SKIP_ACCOUNT_OCI") == "yes",
    }
}

// IsAWSEnabled returns true if AWS tests should run
func (c *TestConfig) IsAWSEnabled() bool {
    return !c.SkipAWS && c.AWSAccountNumber != ""
}

// IsAzureEnabled returns true if Azure tests should run
func (c *TestConfig) IsAzureEnabled() bool {
    return !c.SkipAzure && c.AzureSubID != ""
}

// IsGCPEnabled returns true if GCP tests should run
func (c *TestConfig) IsGCPEnabled() bool {
    return !c.SkipGCP && c.GCPProject != ""
}

// IsOCIEnabled returns true if OCI tests should run
func (c *TestConfig) IsOCIEnabled() bool {
    return !c.SkipOCI && c.OCITenancyID != ""
}
```

## Usage Examples

### Using PreCheck Functions
```go
func TestAccAviatrixGateway_aws(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck: func() {
            utils.PreCheckFunc(t)()
            utils.PreCheckAWS(t)
        },
        // ...
    })
}
```

### Using Generators
```go
func testAccConfig() string {
    gwName := utils.GenerateTestGatewayName()
    cidr := utils.GenerateCIDR("10", 16)
    return fmt.Sprintf(`
        resource "aviatrix_gateway" "test" {
            gw_name = "%s"
            vpc_id  = "%s"
        }
    `, gwName, cidr)
}
```

### Using Assertions
```go
Check: resource.ComposeTestCheckFunc(
    utils.AssertResourceAttrSet("aviatrix_gateway.test", "public_ip"),
    utils.AssertResourceCount("aviatrix_gateway", 1),
),
```

## Best Practices

1. **Reusability**: Write generic, reusable functions
2. **Error Handling**: Always return meaningful errors
3. **Documentation**: Add GoDoc comments to all exported functions
4. **Testing**: Unit test utility functions
5. **Consistency**: Follow Go conventions and patterns

## Contributing

When adding utilities:
1. Place in appropriate file
2. Add GoDoc comments
3. Include usage examples
4. Write unit tests
5. Update this README
