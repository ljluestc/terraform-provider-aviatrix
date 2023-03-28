package aviatrix

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/terraform-providers/terraform-provider-aviatrix/goaviatrix"
)

func TestAccAwsTgwVpnConn_basic(t *testing.T) {
	t.Parallel()

	var awsTgwVpnConn goaviatrix.AwsTgwVpnConn

	resourceName := fmt.Sprintf("aviatrix_aws_tgw_vpn_conn.test-%s", random.UniqueId())

	skipAcc := os.Getenv("SKIP_AWS_TGW_VPN_CONN")
	if skipAcc == "yes" {
		t.Skip("Skipping AWS TGW VPN CONN test as SKIP_AWS_TGW_VPN_CONN is set")
	}

	awsSideAsNumber := "12"

	terraformOptions := &terraform.Options{
		TerraformDir: "./",
		Vars: map[string]interface{}{
			"aws_side_as_number":   awsSideAsNumber,
			"vpn_connection_name": resourceName,
			"public_ip":           "40.0.0.0",
			"route_domain_name":   "Default_Domain",
			"AWS_ACCESS_KEY":      os.Getenv("AWS_ACCESS_KEY"),
			"AWS_SECRET_KEY":      os.Getenv("AWS_SECRET_KEY"),
			"AWS_REGION":          os.Getenv("AWS_REGION"),
			"AWS_ACCOUNT_NUMBER":  os.Getenv("AWS_ACCOUNT_NUMBER"),
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	awsTgwVpnConn = goaviatrix.AwsTgwVpnConn{
		TgwName:  terraform.Output(t, terraformOptions, "tgw_name"),
		ConnName: resourceName,
		VpnID:    terraform.Output(t, terraformOptions, "vpn_id"),
	}

	err := verifyAwsTgwVpnConnExists(t, &awsTgwVpnConn)
	assert.Nil(t, err)

	importedTerraformOptions := terraformOptions
	importedTerraformOptions.Import = true
	importedTerraformOptions.ImportStateVerify = true
	importedTerraformOptions.ImportStateVerifyIgnore = []string{"vpn_tunnel_data"}

	terraform.Import(t, importedTerraformOptions)

	err = verifyAwsTgwVpnConnExists(t, &awsTgwVpnConn)
	assert.Nil(t, err)
}

func verifyAwsTgwVpnConnExists(t *testing.T, awsTgwVpnConn *goaviatrix.AwsTgwVpnConn) error {
	client := getAviatrixClient(t)

	foundAwsTgwVpnConn := &goaviatrix.AwsTgwVpnConn{
		TgwName: awsTgwVpnConn.TgwName,
		VpnID:   awsTgwVpnConn.VpnID,
	}

	foundAwsTgwVpnConn2, err := client.GetAwsTgwVpnConn(foundAwsTgwVpnConn)
	if err != nil {
		return err
	}
	if foundAwsTgwVpnConn2.TgwName != awsTgwVpnConn.TgwName {
		return fmt.Errorf("tgw_name Not found in created attributes")
	}
	if foundAwsTgwVpnConn2.ConnName != awsTgwVpnConn.ConnName {
		return fmt.Errorf("connection_name Not found in created attributes")
	}

	return nil
}

func getAviatrixClient(t *testing.T) *goaviatrix.Client {
	username := os.Getenv("AVIATRIX_USERNAME")
	password := os.Getenv("AVIATRIX_PASSWORD")
	controllerURL := os.Getenv("AVIATRIX_CONTROLLER_URL")

	client, err := goaviatrix.NewClient(username, password, controllerURL, "", "")
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func testAwsTgwVpnConnDestroy(t *testing.T, vpnConnID string) {
	terraformOptions := &terraform.Options{
		TerraformDir: "../path/to/terraform/directory",
		Vars: map[string]interface{}{
			"aws_region":        awsRegion,
			"aws_access_key":    awsAccessKey,
			"aws_secret_key":    awsSecretKey,
			"aws_account_num":   awsAccountNum,
			"connection_name":   vpnConnID,
			"tgw_vpc_id":        tgwVPCID,
			"customer_gateway":  customerGateway,
			"ipsec_dpd_timeout": ipsecDPDTimeout,
			"ike_version":       ikeVersion,
			"bgp_asn":           bgpASN,
			"static_routes":     staticRoutes,
			"vpn_user_enabled":  vpnUserEnabled,
			"vpn_user_name":     vpnUserName,
			"vpn_user_password": vpnUserPassword,
		},
	}

	terraform.Destroy(t, terraformOptions)

	// Check if the VPN connection is destroyed
	awsTgwVpnConn := &goaviatrix.AwsTgwVpnConn{
		TgwName: tgwName,
		VpnID:   vpnConnID,
	}

	client := awsTgwVpnConnClient(t)
	_, err := client.GetAwsTgwVpnConn(awsTgwVpnConn)
	if err == nil {
		t.Errorf("AWS TGW VPN CONN still exists")
	}
}
