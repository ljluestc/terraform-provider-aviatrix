package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

func ConvertToTerratestTest(t *testing.T, testFunc interface{}) {
	testName := runtime.FuncForPC(reflect.ValueOf(testFunc).Pointer()).Name()

	if os.Getenv("SKIP_" + strings.ToUpper(testName)) == "YES" {
		t.Skip(fmt.Sprintf("Skipping %s test as SKIP_%s is set", testName, strings.ToUpper(testName)))
	}

	terraformDir := filepath.Join(".", "terraform")

	terraformOptions := &terraform.Options{
		TerraformDir: terraformDir,
		Vars: map[string]interface{}{
			"tgw_name":            fmt.Sprintf("aws-tgw-%s", random.UniqueId()),
			"connection_name":     fmt.Sprintf("aws-tgw-connect-%s", random.UniqueId()),
			"connect_peer_name":   fmt.Sprintf("connect-peer-%s", random.UniqueId()),
			"peer_as_number":      "65001",
			"peer_gre_address":    "172.31.1.11",
			"bgp_inside_cidrs":    []string{"169.254.6.0/29"},
			"tgw_gre_address":     "10.0.0.32",
		},
	}

	// Clean up resources at the end of the test
	defer terraform.Destroy(t, terraformOptions)

	// Apply the Terraform configuration
	terraform.InitAndApply(t, terraformOptions)

	// Call the original test function
	testFunc.(func(*testing.T))(t)
}
