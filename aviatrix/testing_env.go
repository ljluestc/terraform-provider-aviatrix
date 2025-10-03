package aviatrix

import (
	"fmt"
	"os"
	"strings"
)

// CloudProviderCreds represents cloud provider credentials configuration
type CloudProviderCreds struct {
	Provider string
	EnvVars  map[string]string
	Required []string
	Optional []string
}

// GetAllCloudProviderCreds returns credentials configuration for all supported providers
func GetAllCloudProviderCreds() []CloudProviderCreds {
	return []CloudProviderCreds{
		GetAWSCreds(),
		GetAzureCreds(),
		GetGCPCreds(),
		GetOCICreds(),
	}
}

// GetAWSCreds returns AWS credentials configuration
func GetAWSCreds() CloudProviderCreds {
	return CloudProviderCreds{
		Provider: "aws",
		Required: []string{
			"AWS_ACCESS_KEY_ID",
			"AWS_SECRET_ACCESS_KEY",
			"AWS_ACCOUNT_NUMBER",
		},
		Optional: []string{
			"AWS_DEFAULT_REGION",
			"AWS_SESSION_TOKEN",
		},
		EnvVars: map[string]string{
			"AWS_ACCESS_KEY_ID":     os.Getenv("AWS_ACCESS_KEY_ID"),
			"AWS_SECRET_ACCESS_KEY": os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"AWS_DEFAULT_REGION":    getEnvWithDefault("AWS_DEFAULT_REGION", "us-east-1"),
			"AWS_ACCOUNT_NUMBER":    os.Getenv("AWS_ACCOUNT_NUMBER"),
			"AWS_SESSION_TOKEN":     os.Getenv("AWS_SESSION_TOKEN"),
		},
	}
}

// GetAzureCreds returns Azure credentials configuration
func GetAzureCreds() CloudProviderCreds {
	return CloudProviderCreds{
		Provider: "azure",
		Required: []string{
			"ARM_CLIENT_ID",
			"ARM_CLIENT_SECRET",
			"ARM_SUBSCRIPTION_ID",
			"ARM_TENANT_ID",
		},
		Optional: []string{
			"ARM_ENVIRONMENT",
		},
		EnvVars: map[string]string{
			"ARM_CLIENT_ID":       os.Getenv("ARM_CLIENT_ID"),
			"ARM_CLIENT_SECRET":   os.Getenv("ARM_CLIENT_SECRET"),
			"ARM_SUBSCRIPTION_ID": os.Getenv("ARM_SUBSCRIPTION_ID"),
			"ARM_TENANT_ID":       os.Getenv("ARM_TENANT_ID"),
			"ARM_ENVIRONMENT":     getEnvWithDefault("ARM_ENVIRONMENT", "public"),
		},
	}
}

// GetGCPCreds returns GCP credentials configuration
func GetGCPCreds() CloudProviderCreds {
	return CloudProviderCreds{
		Provider: "gcp",
		Required: []string{
			"GOOGLE_APPLICATION_CREDENTIALS",
			"GOOGLE_PROJECT",
		},
		Optional: []string{
			"GOOGLE_REGION",
			"GOOGLE_ZONE",
		},
		EnvVars: map[string]string{
			"GOOGLE_APPLICATION_CREDENTIALS": os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
			"GOOGLE_PROJECT":                 os.Getenv("GOOGLE_PROJECT"),
			"GOOGLE_REGION":                  getEnvWithDefault("GOOGLE_REGION", "us-central1"),
			"GOOGLE_ZONE":                    getEnvWithDefault("GOOGLE_ZONE", "us-central1-a"),
		},
	}
}

// GetOCICreds returns OCI credentials configuration
func GetOCICreds() CloudProviderCreds {
	return CloudProviderCreds{
		Provider: "oci",
		Required: []string{
			"OCI_USER_ID",
			"OCI_TENANCY_ID",
			"OCI_FINGERPRINT",
			"OCI_PRIVATE_KEY_PATH",
			"OCI_REGION",
		},
		Optional: []string{
			"OCI_COMPARTMENT_ID",
		},
		EnvVars: map[string]string{
			"OCI_USER_ID":          os.Getenv("OCI_USER_ID"),
			"OCI_TENANCY_ID":       os.Getenv("OCI_TENANCY_ID"),
			"OCI_FINGERPRINT":      os.Getenv("OCI_FINGERPRINT"),
			"OCI_PRIVATE_KEY_PATH": os.Getenv("OCI_PRIVATE_KEY_PATH"),
			"OCI_REGION":           os.Getenv("OCI_REGION"),
			"OCI_COMPARTMENT_ID":   os.Getenv("OCI_COMPARTMENT_ID"),
		},
	}
}

// GetAviatrixCreds returns Aviatrix controller credentials configuration
func GetAviatrixCreds() map[string]string {
	return map[string]string{
		"AVIATRIX_CONTROLLER_IP": os.Getenv("AVIATRIX_CONTROLLER_IP"),
		"AVIATRIX_USERNAME":      os.Getenv("AVIATRIX_USERNAME"),
		"AVIATRIX_PASSWORD":      os.Getenv("AVIATRIX_PASSWORD"),
	}
}

// ValidateProviderCreds validates that required credentials are present
func ValidateProviderCreds(provider string) error {
	var creds CloudProviderCreds

	switch provider {
	case "aws":
		creds = GetAWSCreds()
	case "azure":
		creds = GetAzureCreds()
	case "gcp":
		creds = GetGCPCreds()
	case "oci":
		creds = GetOCICreds()
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}

	var missing []string
	for _, req := range creds.Required {
		if creds.EnvVars[req] == "" {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required credentials for %s: %s", provider, strings.Join(missing, ", "))
	}

	return nil
}

// ValidateAviatrixCreds validates Aviatrix controller credentials
func ValidateAviatrixCreds() error {
	creds := GetAviatrixCreds()
	var missing []string

	required := []string{
		"AVIATRIX_CONTROLLER_IP",
		"AVIATRIX_USERNAME",
		"AVIATRIX_PASSWORD",
	}

	for _, req := range required {
		if creds[req] == "" {
			missing = append(missing, req)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required Aviatrix credentials: %s", strings.Join(missing, ", "))
	}

	return nil
}

// IsProviderEnabled checks if a provider is enabled (not skipped)
func IsProviderEnabled(provider string) bool {
	skipVar := fmt.Sprintf("SKIP_ACCOUNT_%s", strings.ToUpper(provider))
	return os.Getenv(skipVar) != "yes"
}

// GetEnabledProviders returns a list of enabled cloud providers
func GetEnabledProviders() []string {
	providers := []string{"aws", "azure", "gcp", "oci"}
	var enabled []string

	for _, provider := range providers {
		if IsProviderEnabled(provider) {
			enabled = append(enabled, provider)
		}
	}

	return enabled
}

// LoadEnvFile loads environment variables from a file
func LoadEnvFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		// Only set if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}

// getEnvWithDefault returns the environment variable value or a default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// PrintCredentialsSummary prints a summary of available credentials (masked)
func PrintCredentialsSummary() {
	fmt.Println("=== Credential Configuration Summary ===")

	// Aviatrix
	avxCreds := GetAviatrixCreds()
	fmt.Println("\nAviatrix Controller:")
	for key, value := range avxCreds {
		fmt.Printf("  %s: %s\n", key, maskValue(value))
	}

	// Cloud Providers
	for _, creds := range GetAllCloudProviderCreds() {
		fmt.Printf("\n%s:\n", strings.ToUpper(creds.Provider))
		if !IsProviderEnabled(creds.Provider) {
			fmt.Println("  [SKIPPED]")
			continue
		}

		for _, req := range creds.Required {
			value := creds.EnvVars[req]
			fmt.Printf("  %s: %s\n", req, maskValue(value))
		}
	}
	fmt.Println()
}

// maskValue masks credential values for display
func maskValue(value string) string {
	if value == "" {
		return "[NOT SET]"
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}
