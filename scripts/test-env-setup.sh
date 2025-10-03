#!/bin/bash
# Test Environment Setup and Validation Script
# This script validates that all required environment variables are set
# and cloud provider credentials are properly configured.

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track validation status
VALIDATION_PASSED=true
WARNINGS=()
ERRORS=()

echo "=========================================="
echo "Test Environment Setup and Validation"
echo "=========================================="
echo ""

# Function to check required variable
check_required() {
    local var_name=$1
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        ERRORS+=("❌ Required: $var_name is not set")
        VALIDATION_PASSED=false
        return 1
    else
        echo -e "${GREEN}✓${NC} $var_name is set"
        return 0
    fi
}

# Function to check optional variable
check_optional() {
    local var_name=$1
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        WARNINGS+=("⚠️  Optional: $var_name is not set")
        return 1
    else
        echo -e "${GREEN}✓${NC} $var_name is set"
        return 0
    fi
}

# Function to check skip flag
check_skip() {
    local provider=$1
    local var_name="SKIP_ACCOUNT_${provider}"
    local var_value="${!var_name}"

    if [ "$var_value" = "yes" ]; then
        echo -e "${YELLOW}⊗${NC} $provider tests will be skipped"
        return 0
    else
        echo -e "${GREEN}✓${NC} $provider tests are enabled"
        return 1
    fi
}

echo "=== Core Test Configuration ==="
check_required "TF_ACC"
check_optional "GO_TEST_TIMEOUT"
check_optional "TEST_ARTIFACT_DIR"
echo ""

echo "=== Aviatrix Controller Configuration ==="
check_required "AVIATRIX_CONTROLLER_IP"
check_required "AVIATRIX_USERNAME"
check_required "AVIATRIX_PASSWORD"
echo ""

echo "=== AWS Configuration ==="
if ! check_skip "AWS"; then
    check_required "AWS_ACCESS_KEY_ID"
    check_required "AWS_SECRET_ACCESS_KEY"
    check_required "AWS_ACCOUNT_NUMBER"
    check_optional "AWS_DEFAULT_REGION"

    # Validate AWS credentials if available
    if command -v aws &> /dev/null; then
        if aws sts get-caller-identity &> /dev/null; then
            echo -e "${GREEN}✓${NC} AWS credentials are valid"
        else
            ERRORS+=("❌ AWS credentials are invalid or expired")
            VALIDATION_PASSED=false
        fi
    fi
fi
echo ""

echo "=== Azure Configuration ==="
if ! check_skip "AZURE"; then
    check_required "ARM_CLIENT_ID"
    check_required "ARM_CLIENT_SECRET"
    check_required "ARM_SUBSCRIPTION_ID"
    check_required "ARM_TENANT_ID"

    # Validate Azure credentials if available
    if command -v az &> /dev/null; then
        if az account show &> /dev/null; then
            echo -e "${GREEN}✓${NC} Azure credentials are valid"
        else
            WARNINGS+=("⚠️  Azure credentials may be invalid (az login required)")
        fi
    fi
fi
echo ""

echo "=== GCP Configuration ==="
if ! check_skip "GCP"; then
    check_required "GOOGLE_APPLICATION_CREDENTIALS"
    check_required "GOOGLE_PROJECT"

    # Validate GCP credentials if available
    if [ -n "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
        if [ -f "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
            echo -e "${GREEN}✓${NC} GCP credentials file exists"

            if command -v gcloud &> /dev/null; then
                export GOOGLE_APPLICATION_CREDENTIALS
                if gcloud auth application-default print-access-token &> /dev/null; then
                    echo -e "${GREEN}✓${NC} GCP credentials are valid"
                else
                    WARNINGS+=("⚠️  GCP credentials may be invalid")
                fi
            fi
        else
            ERRORS+=("❌ GCP credentials file not found: $GOOGLE_APPLICATION_CREDENTIALS")
            VALIDATION_PASSED=false
        fi
    fi
fi
echo ""

echo "=== OCI Configuration ==="
if ! check_skip "OCI"; then
    check_required "OCI_USER_ID"
    check_required "OCI_TENANCY_ID"
    check_required "OCI_FINGERPRINT"
    check_required "OCI_PRIVATE_KEY_PATH"
    check_required "OCI_REGION"

    # Validate OCI private key file
    if [ -n "$OCI_PRIVATE_KEY_PATH" ]; then
        if [ -f "$OCI_PRIVATE_KEY_PATH" ]; then
            echo -e "${GREEN}✓${NC} OCI private key file exists"
        else
            ERRORS+=("❌ OCI private key file not found: $OCI_PRIVATE_KEY_PATH")
            VALIDATION_PASSED=false
        fi
    fi
fi
echo ""

echo "=== Test Infrastructure Validation ==="

# Check if Go is installed
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Go is installed: $GO_VERSION"
else
    ERRORS+=("❌ Go is not installed")
    VALIDATION_PASSED=false
fi

# Check if Terraform is installed
if command -v terraform &> /dev/null; then
    TF_VERSION=$(terraform version -json | jq -r '.terraform_version')
    echo -e "${GREEN}✓${NC} Terraform is installed: v$TF_VERSION"
else
    ERRORS+=("❌ Terraform is not installed")
    VALIDATION_PASSED=false
fi

# Check if Docker is installed (optional for local dev)
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker version --format '{{.Server.Version}}' 2>/dev/null || echo "unknown")
    echo -e "${GREEN}✓${NC} Docker is installed: $DOCKER_VERSION"
else
    WARNINGS+=("⚠️  Docker is not installed (optional for local development)")
fi

echo ""
echo "=== Creating Test Directories ==="

# Create test directories if they don't exist
TEST_ARTIFACT_DIR="${TEST_ARTIFACT_DIR:-./test-results}"
TEST_DATA_DIR="${TEST_DATA_DIR:-./test-data}"

mkdir -p "$TEST_ARTIFACT_DIR"
echo -e "${GREEN}✓${NC} Created $TEST_ARTIFACT_DIR"

mkdir -p "$TEST_DATA_DIR"
echo -e "${GREEN}✓${NC} Created $TEST_DATA_DIR"

mkdir -p "$TEST_ARTIFACT_DIR/logs"
echo -e "${GREEN}✓${NC} Created $TEST_ARTIFACT_DIR/logs"

mkdir -p "$TEST_ARTIFACT_DIR/coverage"
echo -e "${GREEN}✓${NC} Created $TEST_ARTIFACT_DIR/coverage"

echo ""
echo "=========================================="
echo "Validation Summary"
echo "=========================================="

# Print warnings
if [ ${#WARNINGS[@]} -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}Warnings:${NC}"
    for warning in "${WARNINGS[@]}"; do
        echo "  $warning"
    done
fi

# Print errors
if [ ${#ERRORS[@]} -gt 0 ]; then
    echo ""
    echo -e "${RED}Errors:${NC}"
    for error in "${ERRORS[@]}"; do
        echo "  $error"
    done
fi

echo ""
if [ "$VALIDATION_PASSED" = true ]; then
    echo -e "${GREEN}✓ Environment validation passed!${NC}"
    echo ""
    echo "You can now run tests with:"
    echo "  make test          # Run all tests"
    echo "  make testacc       # Run acceptance tests"
    echo "  go test ./...      # Run unit tests"
    exit 0
else
    echo -e "${RED}✗ Environment validation failed!${NC}"
    echo ""
    echo "Please fix the errors above before running tests."
    echo "See .env.test.example for required environment variables."
    exit 1
fi
