package aviatrix

import (
	"fmt"
	"os"
	"testing"

	"github.com/AviatrixSystems/terraform-provider-aviatrix/v3/goaviatrix"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func preAccountCheck(t *testing.T, msgEnd string) {
	requiredEnvVars := []string{
		"AWS_ACCOUNT_NUMBER", "AWS_ACCESS_KEY", "AWS_SECRET_KEY",
		"GCP_ID", "GCP_CREDENTIALS_FILEPATH",
		"ARM_SUBSCRIPTION_ID", "ARM_DIRECTORY_ID", "ARM_APPLICATION_ID", "ARM_APPLICATION_KEY",
		"OCI_TENANCY_ID", "OCI_USER_ID", "OCI_COMPARTMENT_ID", "OCI_API_KEY_FILEPATH",
		"AWSGOV_ACCOUNT_NUMBER", "AWSGOV_ACCESS_KEY", "AWSGOV_SECRET_KEY",
		"AZUREGOV_SUBSCRIPTION_ID", "AZUREGOV_DIRECTORY_ID", "AZUREGOV_APPLICATION_ID", "AZUREGOV_APPLICATION_KEY",
		"AWSCHINA_IAM_ACCOUNT_NUMBER", "AWSCHINA_ACCOUNT_NUMBER", "AWSCHINA_ACCESS_KEY", "AWSCHINA_SECRET_KEY",
		"AZURECHINA_SUBSCRIPTION_ID", "AZURECHINA_DIRECTORY_ID", "AZURECHINA_APPLICATION_ID", "AZURECHINA_APPLICATION_KEY",
		"AWSTS_ACCOUNT_NUMBER", "AWSTS_CAP_URL", "AWSTS_CAP_AGENCY", "AWSTS_CAP_MISSION", "AWSTS_CAP_ROLE_NAME",
		"AWSTS_CAP_CERT", "AWSTS_CAP_CERT_KEY", "AWSTS_CA_CHAIN_CERT",
		"AWSS_ACCOUNT_NUMBER", "AWSS_CAP_URL", "AWSS_CAP_AGENCY", "AWSS_CAP_ACCOUNT_NAME", "AWSS_CAP_ROLE_NAME",
		"AWSS_CAP_CERT", "AWSS_CAP_CERT_KEY", "AWSS_CA_CHAIN_CERT",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(fmt.Sprintf("SKIP_ACCOUNT_%s", envVar)) != "no" {
			continue
		}

		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests. %s", envVar, msgEnd)
		}
	}
}

func TestAccAviatrixAccount_basic(t *testing.T) {
	terraformOptions := &terraform.Options{
		TerraformDir: "./terraform/account",
		Vars: map[string]interface{}{
			"aws_access_key":                 os.Getenv("AWS_ACCESS_KEY"),
			"aws_secret_key":                 os.Getenv("AWS_SECRET_KEY"),
			"account_name_aws":               fmt.Sprintf("tfa-aws-%d", acctest.RandInt()),
			"account_name_gcp":               fmt.Sprintf("tfa-gcp-%d", acctest.RandInt()),
			"gcloud_project_id":              os.Getenv("GCP_ID"),
			"gcloud_project_credentials_file": os.Getenv("GCP_CREDENTIALS_FILEPATH"),
		},
	}

	// Skip the test if SKIP_ACCOUNT is set
	if os.Getenv("SKIP_ACCOUNT") == "yes" {
		t.Skip("Skipping Access Account test as SKIP_ACCOUNT is set")
	}

	// Test AWS account
	if os.Getenv("SKIP_ACCOUNT_AWS") != "yes" {
		t.Run("Test AWS Account", func(t *testing.T) {
			terraformOptionsCopy := terraformOptions.Copy()
			terraformOptionsCopy.TerraformDir = "./terraform/account/aws"

			defer terraform.Destroy(t, terraformOptionsCopy)

			terraform.InitAndApply(t, terraformOptionsCopy)

			outputAccountId := terraform.Output(t, terraformOptionsCopy, "account_id")

			assert.NotEmpty(t, outputAccountId)
		})
	}

	// Test GCP account
	if os.Getenv("SKIP_ACCOUNT_GCP") != "yes" {
		t.Run("Test GCP Account", func(t *testing.T) {
			terraformOptionsCopy := terraformOptions.Copy()
			terraformOptionsCopy.TerraformDir = "./terraform/account/gcp"

			defer terraform.Destroy(t, terraformOptionsCopy)

			terraform.InitAndApply(t, terraformOptionsCopy)

			outputProjectId := terraform.Output(t, terraformOptionsCopy, "project_id")

			assert.NotEmpty(t, outputProjectId)
		})
	}

	if os.Getenv("SKIP_ACCOUNT_AZURE") == "yes" {
		t.Log("Skipping ARN Access Account test as SKIP_ACCOUNT_AZURE is set")
	} else {
		testAzureAccountConfig := testAccAccountConfigAZURE(rInt)
	
		// Create the Terraform options with the test directory and variable inputs
		terraformOptions := &terraform.Options{
			TerraformDir: "../../path/to/terraform/directory",
			Vars: map[string]interface{}{
				"account_name":        fmt.Sprintf("tfa-azure-%d", rInt),
				"arm_subscription_id": os.Getenv("ARM_SUBSCRIPTION_ID"),
				"arm_directory_id":    os.Getenv("ARM_DIRECTORY_ID"),
				"arm_application_id":  os.Getenv("ARM_APPLICATION_ID"),
				"arm_application_key": os.Getenv("ARM_APPLICATION_KEY"),
			},
		}
	
		// Defer the terraform destroy until the end of the test
		defer terraform.Destroy(t, terraformOptions)
	
		// Run terraform init and apply
		terraform.InitAndApply(t, terraformOptions)
	
		// Import the created resource state
		resourceName := "aviatrix_account.azure"
		importStateOptions := &terraform.ImportStateOpts{
			ResourceName: resourceName,
		}
		err := terraform.ImportState(importStateOptions)
		require.NoError(t, err)
	
		// Verify the imported resource state
		err = aviatrix.VerifyAccountExists(resourceName, &account)
		require.NoError(t, err)
	}

	if skipOCI == "yes" {
		t.Log("Skipping OCI Access Account test as SKIP_ACCOUNT_OCI is set")
	} else {
		resourceName := "aviatrix_account.oci"
		importStateVerifyIgnore = append(importStateVerifyIgnore, "oci_tenancy_id")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "oci_user_id")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "oci_compartment_id")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "oci_api_private_key_filepath")
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigOCI(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-oci-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "oci_tenancy_id", os.Getenv("OCI_TENANCY_ID")),
						resource.TestCheckResourceAttr(resourceName, "oci_user_id", os.Getenv("OCI_USER_ID")),
						resource.TestCheckResourceAttr(resourceName, "oci_compartment_id", os.Getenv("OCI_COMPARTMENT_ID")),
						resource.TestCheckResourceAttr(resourceName, "oci_api_private_key_filepath", os.Getenv("OCI_API_KEY_FILEPATH")),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}

	if skipAZUREGOV == "yes" {
		t.Log("Skipping AZUREGOV Access Account test as SKIP_ACCOUNT_AZUREGOV is set")
	} else {
		resourceName := "aviatrix_account.azuregov"
		importStateVerifyIgnore = append(importStateVerifyIgnore, "azuregov_directory_id")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "azuregov_application_id")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "azuregov_application_key")
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigAZUREGOV(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-azuregov-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "azuregov_subscription_id", os.Getenv("AZUREGOV_SUBSCRIPTION_ID")),
						resource.TestCheckResourceAttr(resourceName, "azuregov_directory_id", os.Getenv("AZUREGOV_DIRECTORY_ID")),
						resource.TestCheckResourceAttr(resourceName, "azuregov_application_id", os.Getenv("AZUREGOV_APPLICATION_ID")),
						resource.TestCheckResourceAttr(resourceName, "azuregov_application_key", os.Getenv("AZUREGOV_APPLICATION_KEY")),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}

	if skipAWSGOV == "yes" {
		t.Log("Skipping AWSGov Access Account test as SKIP_ACCOUNT_AWSGOV is set")
	} else {
		resourceName := "aviatrix_account.awsgov"
		importStateVerifyIgnore = append(importStateVerifyIgnore, "awsgov_access_key")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "awsgov_secret_key")
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
				preAccountCheck(t, ". Set SKIP_ACCOUNT to yes to skip account tests")
			},
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigAWSGOV(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-awsgov-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "awsgov_account_number", os.Getenv("AWSGOV_ACCOUNT_NUMBER")),
						resource.TestCheckResourceAttr(resourceName, "awsgov_access_key", os.Getenv("AWSGOV_ACCESS_KEY")),
						resource.TestCheckResourceAttr(resourceName, "awsgov_secret_key", os.Getenv("AWSGOV_SECRET_KEY")),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}

	if skipAWSCHINAIAM == "yes" {
		t.Log("Skipping AWS China IAM Access Account test as SKIP_ACCOUNT_AWSCHINA_IAM is set")
	} else {
		resourceName := "aviatrix_account.awschinaiam"

		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
				preAccountCheck(t, ". Set SKIP_ACCOUNT to yes to skip account tests")
			},
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigAWSCHINAIAM(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-awschinaiam-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "awschina_account_number", os.Getenv("AWSCHINA_IAM_ACCOUNT_NUMBER")),
						resource.TestCheckResourceAttr(resourceName, "awschina_iam", "true"),
						resource.TestCheckResourceAttr(resourceName, "awschina_role_app", fmt.Sprintf("arn:aws-cn:iam::%s:role/aviatrix-role-app", os.Getenv("AWSCHINA_IAM_ACCOUNT_NUMBER"))),
						resource.TestCheckResourceAttr(resourceName, "awschina_role_ec2", fmt.Sprintf("arn:aws-cn:iam::%s:role/aviatrix-role-ec2", os.Getenv("AWSCHINA_IAM_ACCOUNT_NUMBER"))),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}

	if skipAWSTS == "yes" {
		t.Log("Skipping AWS Top Secret Region (C2S) Access Account test as SKIP_ACCOUNT_AWSTS is set")
	} else {
		resourceName := "aviatrix_account.awsts"
		importStateVerifyIgnore = append(importStateVerifyIgnore, "awsts_cap_cert", "awsts_cap_cert_key", "awsts_ca_chain_cert")
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
				preAccountCheck(t, ". Set SKIP_ACCOUNT to yes to skip account tests")
			},
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigAWSTS(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-awsc2s-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "awsts_account_number", os.Getenv("AWSTS_ACCOUNT_NUMBER")),
						resource.TestCheckResourceAttr(resourceName, "awsts_cap_url", os.Getenv("AWSTS_CAP_URL")),
						resource.TestCheckResourceAttr(resourceName, "awsts_cap_agency", os.Getenv("AWSTS_CAP_AGENCY")),
						resource.TestCheckResourceAttr(resourceName, "awsts_cap_mission", os.Getenv("AWSTS_CAP_MISSION")),
						resource.TestCheckResourceAttr(resourceName, "awsts_cap_role_name", os.Getenv("AWSTS_CAP_ROLE_NAME")),
						resource.TestCheckResourceAttr(resourceName, "awsts_cap_cert", os.Getenv("AWSTS_CAP_CERT")),
						resource.TestCheckResourceAttr(resourceName, "awsts_cap_cert_key", os.Getenv("AWSTS_CAP_CERT_KEY")),
						resource.TestCheckResourceAttr(resourceName, "awsts_ca_chain_cert", os.Getenv("AWSTS_CA_CHAIN_CERT")),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}

	if skipAWSCHINA == "yes" {
		t.Log("Skipping AWS China Access Account test as SKIP_ACCOUNT_AWSCHINA is set")
	} else {
		resourceName := "aviatrix_account.awschina"
		importStateVerifyIgnore = append(importStateVerifyIgnore, "awschina_secret_key")

		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
				preAccountCheck(t, ". Set SKIP_ACCOUNT to yes to skip account tests")
			},
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigAWSCHINA(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-awschina-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "awschina_account_number", os.Getenv("AWSCHINA_IAM_ACCOUNT_NUMBER")),
						resource.TestCheckResourceAttr(resourceName, "awschina_iam", "false"),
						resource.TestCheckResourceAttr(resourceName, "awschina_access_key", os.Getenv("AWSCHINA_ACCESS_KEY")),
						resource.TestCheckResourceAttr(resourceName, "awschina_secret_key", os.Getenv("AWSCHINA_SECRET_KEY")),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}

	if skipAZURECHINA == "yes" {
		t.Log("Skipping AzureChina Access Account test as SKIP_ACCOUNT_AZURECHINA is set")
	} else {
		resourceName := "aviatrix_account.azurechina"
		importStateVerifyIgnore = append(importStateVerifyIgnore, "azurechina_directory_id")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "azurechina_application_id")
		importStateVerifyIgnore = append(importStateVerifyIgnore, "azurechina_application_key")
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigAZURECHINA(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-azurechina-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "azurechina_subscription_id", os.Getenv("AZURECHINA_SUBSCRIPTION_ID")),
						resource.TestCheckResourceAttr(resourceName, "azurechina_directory_id", os.Getenv("AZURECHINA_DIRECTORY_ID")),
						resource.TestCheckResourceAttr(resourceName, "azurechina_application_id", os.Getenv("AZURECHINA_APPLICATION_ID")),
						resource.TestCheckResourceAttr(resourceName, "azurechina_application_key", os.Getenv("AZURECHINA_APPLICATION_KEY")),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}

	if skipAWSS == "yes" {
		t.Log("Skipping AWS Secret Region (SC2S) Access Account test as SKIP_ACCOUNT_AWSS is set")
	} else {
		resourceName := "aviatrix_account.aws_s"
		importStateVerifyIgnore = append(importStateVerifyIgnore, "awss_cap_cert", "awss_cap_cert_key", "awss_ca_chain_cert")
		resource.Test(t, resource.TestCase{
			PreCheck: func() {
				testAccPreCheck(t)
				preAccountCheck(t, ". Set SKIP_ACCOUNT to yes to skip account tests")
			},
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckAccountDestroy,
			Steps: []resource.TestStep{
				{
					Config: testAccAccountConfigAWSS(rInt),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckAccountExists(resourceName, &account),
						resource.TestCheckResourceAttr(resourceName, "account_name", fmt.Sprintf("tfa-awssc2s-%d", rInt)),
						resource.TestCheckResourceAttr(resourceName, "awss_account_number", os.Getenv("AWSS_ACCOUNT_NUMBER")),
						resource.TestCheckResourceAttr(resourceName, "awss_cap_url", os.Getenv("AWSS_CAP_URL")),
						resource.TestCheckResourceAttr(resourceName, "awss_cap_agency", os.Getenv("AWSS_CAP_AGENCY")),
						resource.TestCheckResourceAttr(resourceName, "awss_cap_account_name", os.Getenv("AWSS_CAP_ACCOUNT_NAME")),
						resource.TestCheckResourceAttr(resourceName, "awss_cap_role_name", os.Getenv("AWSS_CAP_ROLE_NAME")),
						resource.TestCheckResourceAttr(resourceName, "awss_cap_cert", os.Getenv("AWSS_CAP_CERT")),
						resource.TestCheckResourceAttr(resourceName, "awss_cap_cert_key", os.Getenv("AWSS_CAP_CERT_KEY")),
						resource.TestCheckResourceAttr(resourceName, "awss_ca_chain_cert", os.Getenv("AWSS_CA_CHAIN_CERT")),
					),
				},
				{
					ResourceName:            resourceName,
					ImportState:             true,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: importStateVerifyIgnore,
				},
			},
		})
	}
}

func testAccAccountConfigAWS(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "aws" {
	account_name       = "tfa-aws-%d"
	cloud_type         = 1
	aws_account_number = "%s"
	aws_iam            = false
	aws_access_key     = "%s"
	aws_secret_key     = "%s"
}
	`, rInt, os.Getenv("AWS_ACCOUNT_NUMBER"), os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY"))
}


func testAccAccountConfigGCP(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "gcp" {
	account_name                        = "tfa-gcp-%d"
	cloud_type                          = 4
	gcloud_project_id                   = "%s"
	gcloud_project_credentials_filepath = "%s"
}
	`, rInt, os.Getenv("GCP_ID"), os.Getenv("GCP_CREDENTIALS_FILEPATH"))
}

func testAccAccountConfigAZURE(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "azure" {
	account_name        = "tfa-azure-%d"
	cloud_type          = 8
	arm_subscription_id = "%s"
	arm_directory_id    = "%s"
	arm_application_id  = "%s"
	arm_application_key = "%s"
}
	`, rInt, os.Getenv("ARM_SUBSCRIPTION_ID"), os.Getenv("ARM_DIRECTORY_ID"),
		os.Getenv("ARM_APPLICATION_ID"), os.Getenv("ARM_APPLICATION_KEY"))
}

func testAccAccountConfigOCI(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "oci" {
	account_name                 = "tfa-oci-%d"
	cloud_type                   = 16
	oci_tenancy_id               = "%s"
	oci_user_id                  = "%s"
	oci_compartment_id           = "%s"
	oci_api_private_key_filepath = "%s"
}
	`, rInt, os.Getenv("OCI_TENANCY_ID"), os.Getenv("OCI_USER_ID"),
		os.Getenv("OCI_COMPARTMENT_ID"), os.Getenv("OCI_API_KEY_FILEPATH"))
}

func testAccAccountConfigAZUREGOV(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "azuregov" {
	account_name             = "tfa-azuregov-%d"
	cloud_type             	 = 32
	azuregov_subscription_id = "%s"
	azuregov_directory_id    = "%s"
	azuregov_application_id  = "%s"
	azuregov_application_key = "%s"
}
	`, rInt, os.Getenv("AZUREGOV_SUBSCRIPTION_ID"), os.Getenv("AZUREGOV_DIRECTORY_ID"),
		os.Getenv("AZUREGOV_APPLICATION_ID"), os.Getenv("AZUREGOV_APPLICATION_KEY"))
}

func testAccAccountConfigAWSGOV(rInt int) string {
	return fmt.Sprintf(`
	resource "aviatrix_account" "awsgov" {
	account_name          = "tfa-awsgov-%d"
	cloud_type            = 256
	awsgov_account_number = "%s"
	awsgov_access_key     = "%s"
	awsgov_secret_key     = "%s"
}
	`, rInt, os.Getenv("AWSGOV_ACCOUNT_NUMBER"), os.Getenv("AWSGOV_ACCESS_KEY"), os.Getenv("AWSGOV_SECRET_KEY"))
}

func testAccAccountConfigAWSCHINAIAM(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "awschinaiam" {
	account_name            = "tfa-awschinaiam-%d"
	cloud_type              = 1024
	awschina_account_number = "%s"
	awschina_iam            = true
}
	`, rInt, os.Getenv("AWSCHINA_IAM_ACCOUNT_NUMBER"))
}

func testAccAccountConfigAWSCHINA(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "awschina" {
	account_name            = "tfa-awschina-%d"
	cloud_type              = 1024
	awschina_account_number = "%s"
	awschina_access_key     = "%s"
	awschina_secret_key     = "%s"
}
	`, rInt, os.Getenv("AWSCHINA_ACCOUNT_NUMBER"), os.Getenv("AWSCHINA_ACCESS_KEY"), os.Getenv("AWSCHINA_SECRET_KEY"))
}

func testAccAccountConfigAZURECHINA(rInt int) string {
	return fmt.Sprintf(`
resource "aviatrix_account" "azurechina" {
	account_name               = "tfa-azurechina-%d"
	cloud_type                 = 2048
	azurechina_subscription_id = "%s"
	azurechina_directory_id    = "%s"
	azurechina_application_id  = "%s"
	azurechina_application_key = "%s"
}
	`, rInt, os.Getenv("AZURECHINA_SUBSCRIPTION_ID"), os.Getenv("AZURECHINA_DIRECTORY_ID"),
		os.Getenv("AZURECHINA_APPLICATION_ID"), os.Getenv("AZURECHINA_APPLICATION_KEY"))
}

func TestAccAccountAWSTS(t *testing.T) {
	t.Parallel()

	rInt := rand.Intn(100)
	resourceName := fmt.Sprintf("aviatrix_account.awsts-%d", rInt)

	// Create the Terraform options
	terraformOptions := &terraform.Options{
		TerraformDir: "./",
		VarFiles: []string{
			"test-fixtures/account.tfvars",
		},
		Vars: map[string]interface{}{
			"rName": rInt,
		},
	}

	// Clean up resources after testing
	defer terraform.Destroy(t, terraformOptions)

	// Create the resources
	terraform.InitAndApply(t, terraformOptions)

	// Check if the account exists
	var account goaviatrix.Account
	err := testAccProvider.Meta().(*goaviatrix.Client).GetAccount(&goaviatrix.Account{AccountName: resourceName}, &account)
	assert.NoError(t, err)
	assert.Equal(t, resourceName, account.AccountName)

	// Check if the resources got created in AWS account
	assert.Eventually(t, func() bool {
		// Check if the resources got created in AWS account
		sess := session.Must(session.NewSession())
		awsService := sts.New(sess)
		input := &sts.GetCallerIdentityInput{}
		_, err := awsService.GetCallerIdentity(input)
		if err != nil {
			return false
		}
		return true
	}, time.Minute*5, time.Second*10)
}

func TestAccAccountAWSS(t *testing.T) {
	t.Parallel()

	rInt := rand.Intn(100)
	resourceName := fmt.Sprintf("aviatrix_account.aws_s-%d", rInt)

	// Create the Terraform options
	terraformOptions := &terraform.Options{
		TerraformDir: "./",
		VarFiles: []string{
			"test-fixtures/account.tfvars",
		},
		Vars: map[string]interface{}{
			"rName": rInt,
		},
	}

	// Clean up resources after testing
	defer terraform.Destroy(t, terraformOptions)

	// Create the resources
	terraform.InitAndApply(t, terraformOptions)

	// Check if the account exists
	var account goaviatrix.Account
	err := testAccProvider.Meta().(*goaviatrix.Client).GetAccount(&goaviatrix.Account{AccountName: resourceName}, &account)
	assert.NoError(t, err)
	assert.Equal(t, resourceName, account.AccountName)

	// Check if the resources got created in AWS account
	assert.Eventually(t, func() bool {
		// Check if the resources got created in AWS account
		sess := session.Must(session.NewSession())
		awsService := sts.New(sess)
		input := &sts.GetCallerIdentityInput{}
		_, err := awsService.GetCallerIdentity(input)
		if err != nil {
			return false
		}
		return true
	}, time.Minute*5, time.Second*10)
}
func testAccCheckAccountExists(n string, account *goaviatrix.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("account Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Account ID is set")
		}

		client := testAccProvider.Meta().(*goaviatrix.Client)

		foundAccount := &goaviatrix.Account{
			AccountName: rs.Primary.Attributes["account_name"],
		}

		_, err := client.GetAccount(foundAccount)
		if err != nil {
			return err
		}
		if foundAccount.AccountName != rs.Primary.ID {
			return fmt.Errorf("account not found")
		}

		*account = *foundAccount
		return nil
	}
}

func testAccCheckAccountDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*goaviatrix.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aviatrix_account" {
			continue
		}

		foundAccount := &goaviatrix.Account{
			AccountName: rs.Primary.Attributes["account_name"],
		}

		_, err := client.GetAccount(foundAccount)
		if err != goaviatrix.ErrNotFound {
			return fmt.Errorf("account still exists")
		}
	}

	return nil
}
