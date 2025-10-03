package aviatrix

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	// Seed random number generator for test resource name generation
	rand.Seed(time.Now().UnixNano())
}

// RandomString generates a random string of the given length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RandomInt generates a random integer between min and max (inclusive)
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

// GenerateTestResourceName generates a unique test resource name with prefix
func GenerateTestResourceName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, RandomString(8))
}

// GenerateCIDR generates a test CIDR block
func GenerateCIDR() string {
	return fmt.Sprintf("10.%d.%d.0/24", RandomInt(0, 255), RandomInt(0, 255))
}

// GenerateVPCCIDR generates a VPC-appropriate CIDR block
func GenerateVPCCIDR() string {
	return fmt.Sprintf("10.%d.0.0/16", RandomInt(0, 255))
}

// CloudProviderRegions maps cloud providers to their test regions
var CloudProviderRegions = map[string][]string{
	"aws": {
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"eu-west-1",
		"eu-central-1",
	},
	"azure": {
		"East US",
		"West US",
		"West Europe",
		"Southeast Asia",
	},
	"gcp": {
		"us-central1",
		"us-west1",
		"us-east1",
		"europe-west1",
	},
	"oci": {
		"us-ashburn-1",
		"us-phoenix-1",
		"eu-frankfurt-1",
	},
}

// GetTestRegion returns a test region for the given cloud provider
func GetTestRegion(provider string) string {
	regions, ok := CloudProviderRegions[provider]
	if !ok || len(regions) == 0 {
		return "us-east-1" // Default fallback
	}
	return regions[0]
}

// GetRandomTestRegion returns a random test region for the given cloud provider
func GetRandomTestRegion(provider string) string {
	regions, ok := CloudProviderRegions[provider]
	if !ok || len(regions) == 0 {
		return "us-east-1" // Default fallback
	}
	return regions[rand.Intn(len(regions))]
}

// TestAccountNames provides consistent test account names across tests
var TestAccountNames = map[string]string{
	"aws":   "test-aws-account",
	"azure": "test-azure-account",
	"gcp":   "test-gcp-account",
	"oci":   "test-oci-account",
}

// GetTestAccountName returns the test account name for the given cloud provider
func GetTestAccountName(provider string) string {
	name, ok := TestAccountNames[provider]
	if !ok {
		return "test-account"
	}
	return name
}

// TestVPCNames provides test VPC naming patterns
func GetTestVPCName(provider string) string {
	return GenerateTestResourceName(fmt.Sprintf("test-%s-vpc", provider))
}

// GetTestGatewayName generates a test gateway name
func GetTestGatewayName(provider string) string {
	return GenerateTestResourceName(fmt.Sprintf("test-%s-gw", provider))
}

// WaitForDuration waits for the specified duration with a message
func WaitForDuration(d time.Duration, message string) {
	if message != "" {
		fmt.Printf("Waiting %v: %s\n", d, message)
	}
	time.Sleep(d)
}

// RetryOperation retries an operation with exponential backoff
func RetryOperation(attempts int, delay time.Duration, operation func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		if i < attempts-1 {
			backoff := delay * time.Duration(1<<uint(i))
			fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", i+1, err, backoff)
			time.Sleep(backoff)
		}
	}
	return fmt.Errorf("operation failed after %d attempts: %w", attempts, err)
}
