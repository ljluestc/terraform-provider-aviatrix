# Secrets Management Guide

This document describes how to manage secrets for the Terraform Provider Aviatrix testing framework.

## Overview

The testing framework requires credentials for:
- Aviatrix Controller access
- Cloud provider APIs (AWS, Azure, GCP, OCI)
- CI/CD systems (GitHub Actions)

## Secret Storage Methods

### 1. Environment Variables (Local Development)

Create a `.env` file in the project root (already gitignored):

```bash
# Aviatrix Controller
export AVIATRIX_CONTROLLER_IP="controller.example.com"
export AVIATRIX_USERNAME="admin"
export AVIATRIX_PASSWORD="your-secure-password"

# AWS
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_ACCOUNT_NUMBER="123456789012"
export AWS_DEFAULT_REGION="us-east-1"

# Azure
export ARM_CLIENT_ID="12345678-1234-1234-1234-123456789012"
export ARM_CLIENT_SECRET="your-client-secret"
export ARM_SUBSCRIPTION_ID="12345678-1234-1234-1234-123456789012"
export ARM_TENANT_ID="12345678-1234-1234-1234-123456789012"

# GCP
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/gcp-credentials.json"
export GOOGLE_PROJECT="your-gcp-project"

# OCI
export OCI_USER_ID="ocid1.user.oc1..example"
export OCI_TENANCY_ID="ocid1.tenancy.oc1..example"
export OCI_FINGERPRINT="aa:bb:cc:dd:ee:ff:gg:hh:ii:jj:kk:ll:mm:nn:oo:pp"
export OCI_PRIVATE_KEY_PATH="/path/to/oci-private-key.pem"
export OCI_REGION="us-ashburn-1"

# Test Configuration
export TF_ACC="1"  # Enable acceptance tests
```

Source the file:
```bash
source .env
```

### 2. GitHub Secrets (CI/CD)

Configure secrets in GitHub repository settings:

**Settings → Secrets and variables → Actions → New repository secret**

#### Required Secrets

**Aviatrix:**
- `AVIATRIX_CONTROLLER_IP`
- `AVIATRIX_USERNAME`
- `AVIATRIX_PASSWORD`

**AWS:**
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_ACCOUNT_NUMBER`
- `AWS_DEFAULT_REGION` (optional, defaults to us-east-1)

**Azure:**
- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`

**GCP:**
- `GOOGLE_APPLICATION_CREDENTIALS` (base64-encoded JSON)
- `GOOGLE_PROJECT`

**OCI:**
- `OCI_USER_ID`
- `OCI_TENANCY_ID`
- `OCI_FINGERPRINT`
- `OCI_PRIVATE_KEY` (base64-encoded PEM file)
- `OCI_REGION`

#### Encoding Files for GitHub Secrets

For GCP credentials:
```bash
base64 -w 0 gcp-credentials.json > gcp-credentials-base64.txt
# Copy contents of gcp-credentials-base64.txt to GOOGLE_APPLICATION_CREDENTIALS secret
```

For OCI private key:
```bash
base64 -w 0 oci-private-key.pem > oci-private-key-base64.txt
# Copy contents of oci-private-key-base64.txt to OCI_PRIVATE_KEY secret
```

### 3. Docker Secrets

Pass secrets to Docker containers via environment variables:

```bash
docker run --rm \
  -e AVIATRIX_CONTROLLER_IP="$AVIATRIX_CONTROLLER_IP" \
  -e AVIATRIX_USERNAME="$AVIATRIX_USERNAME" \
  -e AVIATRIX_PASSWORD="$AVIATRIX_PASSWORD" \
  -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
  -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
  -v $(pwd)/gcp-credentials.json:/app/gcp-credentials.json:ro \
  -v $(pwd)/oci-private-key.pem:/app/oci-private-key.pem:ro \
  aviatrix-test:latest
```

Or use docker-compose with `.env` file:
```bash
docker-compose -f tests/docker-compose.yml up
```

### 4. Terraform Variables

**Never commit `terraform.tfvars` to version control!**

Create `terraform.tfvars` from `terraform.tfvars.example`:

```bash
cp tests/fixtures/configs/terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your credentials
```

Add to `.gitignore`:
```
terraform.tfvars
*.tfvars
!*.tfvars.example
```

## Secret Rotation

### Recommended Rotation Schedule

- **Aviatrix Controller**: 90 days
- **Cloud Provider Keys**: 90 days
- **CI/CD Secrets**: After team member changes

### Rotation Process

1. Generate new credentials in cloud provider console
2. Update local `.env` file
3. Update GitHub Secrets
4. Update any running test environments
5. Revoke old credentials
6. Verify tests still pass

## Security Best Practices

### Development

1. **Never commit secrets to git**
   - Always use `.gitignore` for `.env`, `*.tfvars`, credential files
   - Use `git-secrets` or pre-commit hooks to scan for secrets

2. **Use least privilege**
   - Grant only required permissions for testing
   - Use dedicated test accounts when possible

3. **Encrypt at rest**
   - Use encrypted filesystems for credential files
   - Consider using `gpg` to encrypt local `.env` files

4. **Rotate regularly**
   - Follow rotation schedule
   - Automate rotation where possible

### CI/CD

1. **Use environment-specific secrets**
   - Separate secrets for dev/staging/prod
   - Use GitHub environments for additional protection

2. **Limit secret exposure**
   - Only expose secrets to jobs that need them
   - Use `if` conditions to restrict secret access

3. **Audit secret usage**
   - Review workflow logs regularly
   - Enable audit logging in cloud providers

4. **Enable secret scanning**
   - Use GitHub secret scanning
   - Enable push protection

## Skip Provider Configuration

To skip testing specific cloud providers, set these environment variables:

```bash
export SKIP_ACCOUNT_AWS=yes      # Skip AWS tests
export SKIP_ACCOUNT_AZURE=yes    # Skip Azure tests
export SKIP_ACCOUNT_GCP=yes      # Skip GCP tests
export SKIP_ACCOUNT_OCI=yes      # Skip OCI tests
```

This is useful when:
- You don't have credentials for a provider
- Testing a specific provider only
- Reducing CI/CD costs

## Troubleshooting

### Invalid Credentials

**Symptom**: Tests fail with authentication errors

**Solutions**:
1. Verify credentials are current and not expired
2. Check credentials have required permissions
3. Verify environment variables are set correctly
4. Test credentials directly with cloud provider CLI

### Missing Secrets in CI/CD

**Symptom**: GitHub Actions fail with "secret not found"

**Solutions**:
1. Verify secret is created in repository settings
2. Check secret name matches exactly (case-sensitive)
3. Ensure workflow has permission to access secrets
4. Check if using environment-specific secrets correctly

### File Permission Issues

**Symptom**: "Permission denied" errors for credential files

**Solutions**:
```bash
# Set correct permissions for private keys
chmod 600 oci-private-key.pem
chmod 600 ~/.ssh/gcp-key.json

# Verify ownership
ls -l oci-private-key.pem
```

## Example: Complete Setup

### Local Development

```bash
# 1. Copy example files
cp .env.test.example .env
cp tests/fixtures/configs/terraform.tfvars.example terraform.tfvars

# 2. Edit files with your credentials
vim .env
vim terraform.tfvars

# 3. Set up GCP credentials
mv ~/Downloads/gcp-credentials.json .
chmod 600 gcp-credentials.json

# 4. Set up OCI private key
mv ~/Downloads/oci-private-key.pem .
chmod 600 oci-private-key.pem

# 5. Source environment
source .env

# 6. Verify setup
env | grep AVIATRIX
env | grep AWS
env | grep ARM
env | grep GOOGLE
env | grep OCI

# 7. Run tests
make test
```

### GitHub Actions

1. Go to repository settings
2. Navigate to "Secrets and variables" → "Actions"
3. Add all required secrets (see list above)
4. Commit workflow file
5. Push to trigger tests

## Additional Resources

- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [AWS IAM Best Practices](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html)
- [Azure AD Best Practices](https://docs.microsoft.com/en-us/azure/active-directory/fundamentals/identity-secure-score)
- [GCP Service Account Best Practices](https://cloud.google.com/iam/docs/best-practices-for-using-and-managing-service-accounts)
- [OCI Security Best Practices](https://docs.oracle.com/en-us/iaas/Content/Security/Concepts/security_guide.htm)
