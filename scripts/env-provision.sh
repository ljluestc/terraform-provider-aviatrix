#!/bin/bash
# Environment Provisioning Script
# Provisions a new isolated test environment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Configuration
PROVIDER="${1:-aws}"
ENV_ID="${2:-$(date +%s)}"
ENV_NAME="${PROVIDER}-${ENV_ID}"
STATE_DIR="${PROJECT_ROOT}/test-state/${ENV_NAME}"
MODULE_DIR="${PROJECT_ROOT}/test-infra/${PROVIDER}"
DOCKER_NETWORK="${DOCKER_NETWORK:-test-network-${ENV_ID}}"
ENABLE_DOCKER="${ENABLE_DOCKER:-false}"

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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Print header
echo "=============================================="
echo "Test Environment Provisioning"
echo "=============================================="
echo "Provider: ${PROVIDER}"
echo "Environment ID: ${ENV_ID}"
echo "Environment Name: ${ENV_NAME}"
echo "State Directory: ${STATE_DIR}"
echo "Module Directory: ${MODULE_DIR}"
echo "=============================================="
echo ""

# Validate provider
validate_provider() {
    log_step "Validating provider: ${PROVIDER}"

    case "${PROVIDER}" in
        aws|azure|gcp|oci)
            log_info "Provider ${PROVIDER} is valid"
            ;;
        *)
            log_error "Unknown provider: ${PROVIDER}"
            echo "Supported providers: aws, azure, gcp, oci"
            exit 1
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites"

    # Check Terraform
    if ! command -v terraform &> /dev/null; then
        log_error "Terraform is not installed"
        exit 1
    fi
    log_info "Terraform: $(terraform version -json | jq -r '.terraform_version')"

    # Check Docker (if enabled)
    if [ "${ENABLE_DOCKER}" = "true" ]; then
        if ! command -v docker &> /dev/null; then
            log_error "Docker is not installed"
            exit 1
        fi
        log_info "Docker: $(docker version --format '{{.Server.Version}}')"
    fi

    # Check module directory
    if [ ! -d "${MODULE_DIR}" ]; then
        log_error "Module directory not found: ${MODULE_DIR}"
        exit 1
    fi
    log_info "Module directory exists"
}

# Create state directory
create_state_directory() {
    log_step "Creating state directory"

    mkdir -p "${STATE_DIR}"
    mkdir -p "${STATE_DIR}/.terraform"

    log_info "State directory created: ${STATE_DIR}"
}

# Setup Docker network
setup_docker_network() {
    if [ "${ENABLE_DOCKER}" != "true" ]; then
        return 0
    fi

    log_step "Setting up Docker network"

    # Check if network exists
    if docker network ls | grep -q "${DOCKER_NETWORK}"; then
        log_info "Docker network ${DOCKER_NETWORK} already exists"
    else
        docker network create --driver bridge "${DOCKER_NETWORK}"
        log_info "Docker network ${DOCKER_NETWORK} created"
    fi
}

# Initialize Terraform
initialize_terraform() {
    log_step "Initializing Terraform"

    cd "${MODULE_DIR}"

    export TF_DATA_DIR="${STATE_DIR}/.terraform"
    export TF_STATE="${STATE_DIR}/terraform.tfstate"

    terraform init -upgrade 2>&1 | tee "${STATE_DIR}/terraform-init.log"

    log_info "Terraform initialized"
}

# Create Terraform plan
create_plan() {
    log_step "Creating Terraform plan"

    cd "${MODULE_DIR}"

    export TF_DATA_DIR="${STATE_DIR}/.terraform"
    export TF_STATE="${STATE_DIR}/terraform.tfstate"

    # Build variable args from environment
    VARS=""
    if [ -n "${AWS_REGION:-}" ]; then
        VARS="${VARS} -var aws_region=${AWS_REGION}"
    fi
    if [ -n "${AZURE_REGION:-}" ]; then
        VARS="${VARS} -var azure_region=${AZURE_REGION}"
    fi
    if [ -n "${GCP_REGION:-}" ]; then
        VARS="${VARS} -var gcp_region=${GCP_REGION}"
    fi

    terraform plan -out="${STATE_DIR}/tfplan" ${VARS} 2>&1 | tee "${STATE_DIR}/terraform-plan.log"

    log_info "Terraform plan created"
}

# Apply Terraform
apply_terraform() {
    log_step "Applying Terraform plan"

    cd "${MODULE_DIR}"

    export TF_DATA_DIR="${STATE_DIR}/.terraform"
    export TF_STATE="${STATE_DIR}/terraform.tfstate"

    terraform apply -auto-approve "${STATE_DIR}/tfplan" 2>&1 | tee "${STATE_DIR}/terraform-apply.log"

    log_info "Terraform applied successfully"
}

# Extract outputs
extract_outputs() {
    log_step "Extracting Terraform outputs"

    cd "${MODULE_DIR}"

    export TF_DATA_DIR="${STATE_DIR}/.terraform"
    export TF_STATE="${STATE_DIR}/terraform.tfstate"

    terraform output -json > "${STATE_DIR}/outputs.json"

    log_info "Outputs saved to ${STATE_DIR}/outputs.json"
}

# Create environment metadata
create_metadata() {
    log_step "Creating environment metadata"

    cat > "${STATE_DIR}/metadata.json" <<EOF
{
    "environment_id": "${ENV_ID}",
    "environment_name": "${ENV_NAME}",
    "provider": "${PROVIDER}",
    "created_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "state_file": "${STATE_DIR}/terraform.tfstate",
    "module_dir": "${MODULE_DIR}",
    "docker_network": "${DOCKER_NETWORK}",
    "status": "ready"
}
EOF

    log_info "Metadata created"
}

# Cleanup on error
cleanup_on_error() {
    if [ $? -ne 0 ]; then
        log_error "Provisioning failed, cleaning up..."

        # Attempt to destroy resources
        if [ -f "${STATE_DIR}/terraform.tfstate" ]; then
            cd "${MODULE_DIR}"
            export TF_DATA_DIR="${STATE_DIR}/.terraform"
            export TF_STATE="${STATE_DIR}/terraform.tfstate"
            terraform destroy -auto-approve || true
        fi

        # Remove state directory
        rm -rf "${STATE_DIR}"

        # Remove Docker network
        if [ "${ENABLE_DOCKER}" = "true" ]; then
            docker network rm "${DOCKER_NETWORK}" 2>/dev/null || true
        fi
    fi
}

trap cleanup_on_error EXIT

# Main execution
main() {
    validate_provider
    check_prerequisites
    create_state_directory
    setup_docker_network
    initialize_terraform
    create_plan
    apply_terraform
    extract_outputs
    create_metadata

    echo ""
    echo "=============================================="
    log_info "Environment provisioning completed!"
    echo "=============================================="
    echo "Environment Name: ${ENV_NAME}"
    echo "State Directory: ${STATE_DIR}"
    echo "Metadata: ${STATE_DIR}/metadata.json"
    echo "Outputs: ${STATE_DIR}/outputs.json"
    echo "=============================================="
}

main "$@"
