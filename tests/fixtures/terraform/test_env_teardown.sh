#!/bin/bash
# Test Environment Teardown Script
# Destroys test infrastructure across all cloud providers

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
FORCE="${FORCE:-false}"
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

# Confirm destruction
confirm_destruction() {
    if [ "${FORCE}" = "true" ]; then
        return 0
    fi

    echo -e "${YELLOW}WARNING: This will destroy all test infrastructure!${NC}"
    read -p "Are you sure you want to continue? (yes/no): " -r
    echo

    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Teardown cancelled"
        exit 0
    fi
}

# Destroy AWS infrastructure
destroy_aws() {
    if [ "${SKIP_ACCOUNT_AWS:-no}" = "yes" ]; then
        log_info "Skipping AWS teardown"
        return 0
    fi

    log_info "Destroying AWS infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/aws"

    if [ -f "terraform.tfstate" ]; then
        terraform destroy -auto-approve || log_error "AWS destruction failed"
        log_info "AWS infrastructure destroyed"
    else
        log_warn "No AWS state file found"
    fi
}

# Destroy Azure infrastructure
destroy_azure() {
    if [ "${SKIP_ACCOUNT_AZURE:-no}" = "yes" ]; then
        log_info "Skipping Azure teardown"
        return 0
    fi

    log_info "Destroying Azure infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/azure"

    if [ -f "terraform.tfstate" ]; then
        terraform destroy -auto-approve || log_error "Azure destruction failed"
        log_info "Azure infrastructure destroyed"
    else
        log_warn "No Azure state file found"
    fi
}

# Destroy GCP infrastructure
destroy_gcp() {
    if [ "${SKIP_ACCOUNT_GCP:-no}" = "yes" ]; then
        log_info "Skipping GCP teardown"
        return 0
    fi

    log_info "Destroying GCP infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/gcp"

    if [ -f "terraform.tfstate" ]; then
        terraform destroy -auto-approve || log_error "GCP destruction failed"
        log_info "GCP infrastructure destroyed"
    else
        log_warn "No GCP state file found"
    fi
}

# Destroy OCI infrastructure
destroy_oci() {
    if [ "${SKIP_ACCOUNT_OCI:-no}" = "yes" ]; then
        log_info "Skipping OCI teardown"
        return 0
    fi

    log_info "Destroying OCI infrastructure..."

    cd "${PROJECT_ROOT}/test-infra/oci"

    if [ -f "terraform.tfstate" ]; then
        terraform destroy -auto-approve || log_error "OCI destruction failed"
        log_info "OCI infrastructure destroyed"
    else
        log_warn "No OCI state file found"
    fi
}

# Clean up local artifacts
cleanup_artifacts() {
    log_info "Cleaning up local artifacts..."

    # Remove test results
    if [ -d "${PROJECT_ROOT}/test-results" ]; then
        rm -rf "${PROJECT_ROOT}/test-results"
        log_info "Removed test results"
    fi

    # Remove credentials files
    local cred_files=(
        "${PROJECT_ROOT}/gcp-credentials.json"
        "${PROJECT_ROOT}/oci-private-key.pem"
    )

    for file in "${cred_files[@]}"; do
        if [ -f "$file" ]; then
            rm -f "$file"
            log_info "Removed $file"
        fi
    done
}

# Main execution
main() {
    log_info "Starting test environment teardown..."

    confirm_destruction

    if [ "${PARALLEL}" = "true" ]; then
        log_info "Destroying infrastructure in parallel..."
        destroy_aws &
        destroy_azure &
        destroy_gcp &
        destroy_oci &
        wait
    else
        log_info "Destroying infrastructure sequentially..."
        destroy_aws
        destroy_azure
        destroy_gcp
        destroy_oci
    fi

    cleanup_artifacts

    log_info "Test environment teardown complete!"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --force)
            FORCE=true
            shift
            ;;
        --parallel)
            PARALLEL=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --force          Skip confirmation prompt"
            echo "  --parallel       Destroy cloud providers in parallel"
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
