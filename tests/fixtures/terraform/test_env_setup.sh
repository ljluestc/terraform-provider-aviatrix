#!/bin/bash
# Test Environment Setup Script
# Provisions test infrastructure across multiple cloud providers

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"

# Configuration
TEST_ENV="${TEST_ENV:-dev}"
CLEANUP_ON_ERROR="${CLEANUP_ON_ERROR:-true}"
PARALLEL="${PARALLEL:-false}"

# Log functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function
cleanup() {
    if [ "${CLEANUP_ON_ERROR}" = "true" ] && [ $? -ne 0 ]; then
        log_warn "Error detected. Cleaning up resources..."
        terraform destroy -auto-approve || log_error "Cleanup failed"
    fi
}

trap cleanup EXIT

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        log_error "Terraform is not installed"
        exit 1
    fi

    # Check required environment variables
    local required_vars=(
        "AVIATRIX_CONTROLLER_IP"
        "AVIATRIX_USERNAME"
        "AVIATRIX_PASSWORD"
    )

    for var in "${required_vars[@]}"; do
        if [ -z "${!var:-}" ]; then
            log_error "Required environment variable $var is not set"
            exit 1
        fi
    done

    log_info "Prerequisites check passed"
}

# Setup AWS infrastructure
setup_aws() {
    if [ "${SKIP_ACCOUNT_AWS:-no}" = "yes" ]; then
        log_info "Skipping AWS setup"
        return 0
    fi

    log_info "Setting up AWS infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/aws"

    terraform init -upgrade
    terraform plan -out=tfplan
    terraform apply tfplan

    log_info "AWS infrastructure setup complete"
}

# Setup Azure infrastructure
setup_azure() {
    if [ "${SKIP_ACCOUNT_AZURE:-no}" = "yes" ]; then
        log_info "Skipping Azure setup"
        return 0
    fi

    log_info "Setting up Azure infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/azure"

    terraform init -upgrade
    terraform plan -out=tfplan
    terraform apply tfplan

    log_info "Azure infrastructure setup complete"
}

# Setup GCP infrastructure
setup_gcp() {
    if [ "${SKIP_ACCOUNT_GCP:-no}" = "yes" ]; then
        log_info "Skipping GCP setup"
        return 0
    fi

    log_info "Setting up GCP infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/gcp"

    terraform init -upgrade
    terraform plan -out=tfplan
    terraform apply tfplan

    log_info "GCP infrastructure setup complete"
}

# Setup OCI infrastructure
setup_oci() {
    if [ "${SKIP_ACCOUNT_OCI:-no}" = "yes" ]; then
        log_info "Skipping OCI setup"
        return 0
    fi

    log_info "Setting up OCI infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/oci"

    terraform init -upgrade
    terraform plan -out=tfplan
    terraform apply tfplan

    log_info "OCI infrastructure setup complete"
}

# Export environment variables from Terraform outputs
export_outputs() {
    log_info "Exporting Terraform outputs..."

    cd "${PROJECT_ROOT}/test-infra"

    if [ -f "cmdExportOutput.sh" ]; then
        source cmdExportOutput.sh
        log_info "Environment variables exported"
    else
        log_warn "cmdExportOutput.sh not found"
    fi
}

# Validate setup
validate_setup() {
    log_info "Validating test environment setup..."

    # Check if resources were created
    local errors=0

    if [ "${SKIP_ACCOUNT_AWS:-no}" != "yes" ]; then
        cd "${PROJECT_ROOT}/test-infra/aws"
        if ! terraform output vpc_id &> /dev/null; then
            log_error "AWS VPC not created"
            ((errors++))
        fi
    fi

    if [ "${SKIP_ACCOUNT_AZURE:-no}" != "yes" ]; then
        cd "${PROJECT_ROOT}/test-infra/azure"
        if ! terraform output vnet_name &> /dev/null; then
            log_error "Azure VNet not created"
            ((errors++))
        fi
    fi

    if [ $errors -eq 0 ]; then
        log_info "Validation passed"
        return 0
    else
        log_error "Validation failed with $errors error(s)"
        return 1
    fi
}

# Main execution
main() {
    log_info "Starting test environment setup..."
    log_info "Environment: ${TEST_ENV}"

    check_prerequisites

    if [ "${PARALLEL}" = "true" ]; then
        log_info "Setting up infrastructure in parallel..."
        setup_aws &
        setup_azure &
        setup_gcp &
        setup_oci &
        wait
    else
        log_info "Setting up infrastructure sequentially..."
        setup_aws
        setup_azure
        setup_gcp
        setup_oci
    fi

    export_outputs
    validate_setup

    log_info "Test environment setup complete!"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --parallel)
            PARALLEL=true
            shift
            ;;
        --no-cleanup)
            CLEANUP_ON_ERROR=false
            shift
            ;;
        --env)
            TEST_ENV="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --parallel       Setup cloud providers in parallel"
            echo "  --no-cleanup     Don't cleanup on error"
            echo "  --env ENV        Set test environment (default: dev)"
            echo "  --help           Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

main
