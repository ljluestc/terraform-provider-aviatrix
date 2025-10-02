#!/bin/bash
# Setup Test Environment Script
# This script sets up the test environment and validates credentials

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if .env.test exists
if [ ! -f .env.test ]; then
    log_warning ".env.test not found. Creating from template..."
    if [ -f .env.test.example ]; then
        cp .env.test.example .env.test
        log_success "Created .env.test from .env.test.example"
        log_warning "Please update .env.test with your actual credentials"
    else
        log_error ".env.test.example not found!"
        exit 1
    fi
fi

# Load environment variables
log_info "Loading environment variables from .env.test..."
export $(grep -v '^#' .env.test | xargs)

# Create test directories
log_info "Creating test directories..."
mkdir -p test-results
mkdir -p test-results/logs
mkdir -p test-results/screenshots
mkdir -p test-data

log_success "Test directories created"

# Validate Aviatrix Controller credentials
log_info "Validating Aviatrix Controller credentials..."
if [ -z "$AVIATRIX_CONTROLLER_IP" ] || [ -z "$AVIATRIX_USERNAME" ] || [ -z "$AVIATRIX_PASSWORD" ]; then
    log_error "Aviatrix Controller credentials are not set!"
    log_error "Please set AVIATRIX_CONTROLLER_IP, AVIATRIX_USERNAME, and AVIATRIX_PASSWORD in .env.test"
    exit 1
fi
log_success "Aviatrix Controller credentials found"

# Validate cloud provider credentials
validate_aws=false
validate_azure=false
validate_gcp=false
validate_oci=false

if [ "$SKIP_ACCOUNT_AWS" != "yes" ]; then
    log_info "Validating AWS credentials..."
    if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ]; then
        validate_aws=true
        log_success "AWS credentials found"
    else
        log_warning "AWS credentials not set. AWS tests will be skipped."
    fi
fi

if [ "$SKIP_ACCOUNT_AZURE" != "yes" ]; then
    log_info "Validating Azure credentials..."
    if [ -n "$ARM_CLIENT_ID" ] && [ -n "$ARM_CLIENT_SECRET" ] && [ -n "$ARM_SUBSCRIPTION_ID" ] && [ -n "$ARM_TENANT_ID" ]; then
        validate_azure=true
        log_success "Azure credentials found"
    else
        log_warning "Azure credentials not set. Azure tests will be skipped."
    fi
fi

if [ "$SKIP_ACCOUNT_GCP" != "yes" ]; then
    log_info "Validating GCP credentials..."
    if [ -n "$GOOGLE_APPLICATION_CREDENTIALS" ] && [ -n "$GOOGLE_PROJECT" ]; then
        if [ -f "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
            validate_gcp=true
            log_success "GCP credentials found"
        else
            log_warning "GCP credentials file not found at: $GOOGLE_APPLICATION_CREDENTIALS"
        fi
    else
        log_warning "GCP credentials not set. GCP tests will be skipped."
    fi
fi

if [ "$SKIP_ACCOUNT_OCI" != "yes" ]; then
    log_info "Validating OCI credentials..."
    if [ -n "$OCI_USER_ID" ] && [ -n "$OCI_TENANCY_ID" ] && [ -n "$OCI_FINGERPRINT" ] && [ -n "$OCI_PRIVATE_KEY_PATH" ]; then
        if [ -f "$OCI_PRIVATE_KEY_PATH" ]; then
            validate_oci=true
            log_success "OCI credentials found"
        else
            log_warning "OCI private key file not found at: $OCI_PRIVATE_KEY_PATH"
        fi
    else
        log_warning "OCI credentials not set. OCI tests will be skipped."
    fi
fi

# Summary
echo ""
log_info "=== Test Environment Setup Summary ==="
echo "Test Artifact Directory: ${TEST_ARTIFACT_DIR:-./test-results}"
echo "Test Data Directory: ${TEST_DATA_DIR:-./test-data}"
echo ""
echo "Cloud Provider Status:"
echo "  AWS:   $([ "$validate_aws" = true ] && echo -e "${GREEN}Enabled${NC}" || echo -e "${YELLOW}Disabled${NC}")"
echo "  Azure: $([ "$validate_azure" = true ] && echo -e "${GREEN}Enabled${NC}" || echo -e "${YELLOW}Disabled${NC}")"
echo "  GCP:   $([ "$validate_gcp" = true ] && echo -e "${GREEN}Enabled${NC}" || echo -e "${YELLOW}Disabled${NC}")"
echo "  OCI:   $([ "$validate_oci" = true ] && echo -e "${GREEN}Enabled${NC}" || echo -e "${YELLOW}Disabled${NC}")"
echo ""

# Check if at least one cloud provider is configured
if [ "$validate_aws" = false ] && [ "$validate_azure" = false ] && [ "$validate_gcp" = false ] && [ "$validate_oci" = false ]; then
    log_error "No cloud provider credentials are configured!"
    log_error "Please configure at least one cloud provider in .env.test"
    exit 1
fi

log_success "Test environment setup complete!"
echo ""
log_info "You can now run tests using:"
echo "  - make test          # Run unit tests"
echo "  - make testacc       # Run acceptance tests"
echo "  - docker-compose -f docker-compose.test.yml up  # Run tests in Docker"
