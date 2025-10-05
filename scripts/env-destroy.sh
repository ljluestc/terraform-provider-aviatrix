#!/bin/bash
# Environment Destruction Script
# Destroys an isolated test environment and cleans up all resources

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
ENV_NAME="${1:-}"
FORCE="${FORCE:-false}"
SKIP_CONFIRMATION="${SKIP_CONFIRMATION:-false}"

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

# Validate arguments
if [ -z "${ENV_NAME}" ]; then
    log_error "Environment name is required"
    echo "Usage: $0 <environment-name> [--force]"
    echo "Example: $0 aws-1234567890"
    exit 1
fi

# Check for --force flag
if [ "${2:-}" = "--force" ]; then
    FORCE=true
    SKIP_CONFIRMATION=true
fi

STATE_DIR="${PROJECT_ROOT}/test-state/${ENV_NAME}"
METADATA_FILE="${STATE_DIR}/metadata.json"

# Print header
echo "=============================================="
echo "Test Environment Destruction"
echo "=============================================="
echo "Environment Name: ${ENV_NAME}"
echo "State Directory: ${STATE_DIR}"
echo "=============================================="
echo ""

# Validate environment exists
validate_environment() {
    log_step "Validating environment"

    if [ ! -d "${STATE_DIR}" ]; then
        log_error "Environment not found: ${ENV_NAME}"
        log_error "State directory does not exist: ${STATE_DIR}"
        exit 1
    fi

    if [ ! -f "${METADATA_FILE}" ]; then
        log_warn "Metadata file not found: ${METADATA_FILE}"
        log_warn "Will attempt to destroy anyway"
    else
        log_info "Environment found"

        # Display environment info
        if command -v jq &> /dev/null; then
            echo ""
            echo "Environment Details:"
            jq '.' "${METADATA_FILE}"
            echo ""
        fi
    fi
}

# Confirm destruction
confirm_destruction() {
    if [ "${SKIP_CONFIRMATION}" = "true" ]; then
        return 0
    fi

    echo -e "${YELLOW}WARNING: This will destroy all resources in environment ${ENV_NAME}!${NC}"
    read -p "Are you sure you want to continue? (yes/no): " -r
    echo

    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log_info "Destruction cancelled"
        exit 0
    fi
}

# Get metadata value
get_metadata() {
    local key=$1

    if [ ! -f "${METADATA_FILE}" ]; then
        return 1
    fi

    if command -v jq &> /dev/null; then
        jq -r ".${key} // empty" "${METADATA_FILE}"
    else
        return 1
    fi
}

# Destroy Terraform resources
destroy_terraform() {
    log_step "Destroying Terraform resources"

    # Get module directory from metadata
    MODULE_DIR=$(get_metadata "module_dir" 2>/dev/null || echo "")

    if [ -z "${MODULE_DIR}" ]; then
        log_warn "Module directory not found in metadata"

        # Try to determine from environment name
        PROVIDER=$(echo "${ENV_NAME}" | cut -d'-' -f1)
        MODULE_DIR="${PROJECT_ROOT}/test-infra/${PROVIDER}"

        if [ ! -d "${MODULE_DIR}" ]; then
            log_error "Could not determine module directory"
            if [ "${FORCE}" = "false" ]; then
                exit 1
            fi
            log_warn "Skipping Terraform destroy (force mode)"
            return 0
        fi
    fi

    log_info "Using module directory: ${MODULE_DIR}"

    # Check if state file exists
    STATE_FILE="${STATE_DIR}/terraform.tfstate"
    if [ ! -f "${STATE_FILE}" ]; then
        log_warn "State file not found: ${STATE_FILE}"
        log_warn "Skipping Terraform destroy"
        return 0
    fi

    # Destroy resources
    cd "${MODULE_DIR}"

    export TF_DATA_DIR="${STATE_DIR}/.terraform"
    export TF_STATE="${STATE_FILE}"

    if terraform destroy -auto-approve 2>&1 | tee "${STATE_DIR}/terraform-destroy.log"; then
        log_info "Terraform resources destroyed successfully"
    else
        log_error "Terraform destroy failed"

        if [ "${FORCE}" = "false" ]; then
            log_error "Use --force to skip errors and continue cleanup"
            exit 1
        fi

        log_warn "Continuing cleanup despite errors (force mode)"
    fi
}

# Remove Docker network
remove_docker_network() {
    log_step "Removing Docker network"

    DOCKER_NETWORK=$(get_metadata "docker_network" 2>/dev/null || echo "")

    if [ -z "${DOCKER_NETWORK}" ]; then
        log_info "No Docker network to remove"
        return 0
    fi

    if ! command -v docker &> /dev/null; then
        log_warn "Docker not installed, skipping network removal"
        return 0
    fi

    # Check if network exists
    if docker network ls | grep -q "${DOCKER_NETWORK}"; then
        if docker network rm "${DOCKER_NETWORK}" 2>&1; then
            log_info "Docker network removed: ${DOCKER_NETWORK}"
        else
            log_warn "Failed to remove Docker network: ${DOCKER_NETWORK}"

            if [ "${FORCE}" = "false" ]; then
                exit 1
            fi
        fi
    else
        log_info "Docker network does not exist: ${DOCKER_NETWORK}"
    fi
}

# Remove Docker containers
remove_docker_containers() {
    log_step "Removing Docker containers"

    if ! command -v docker &> /dev/null; then
        log_info "Docker not installed, skipping container removal"
        return 0
    fi

    # Find containers with environment name
    CONTAINERS=$(docker ps -a --filter "name=${ENV_NAME}" -q)

    if [ -z "${CONTAINERS}" ]; then
        log_info "No Docker containers to remove"
        return 0
    fi

    log_info "Removing ${CONTAINERS}"

    if docker rm -f ${CONTAINERS} 2>&1; then
        log_info "Docker containers removed"
    else
        log_warn "Failed to remove some Docker containers"

        if [ "${FORCE}" = "false" ]; then
            exit 1
        fi
    fi
}

# Remove state directory
remove_state_directory() {
    log_step "Removing state directory"

    if [ -d "${STATE_DIR}" ]; then
        if rm -rf "${STATE_DIR}"; then
            log_info "State directory removed: ${STATE_DIR}"
        else
            log_error "Failed to remove state directory: ${STATE_DIR}"

            if [ "${FORCE}" = "false" ]; then
                exit 1
            fi
        fi
    else
        log_info "State directory does not exist"
    fi
}

# Main execution
main() {
    validate_environment
    confirm_destruction
    destroy_terraform
    remove_docker_containers
    remove_docker_network
    remove_state_directory

    echo ""
    echo "=============================================="
    log_info "Environment destruction completed!"
    echo "=============================================="
    echo "Environment ${ENV_NAME} has been destroyed"
    echo "=============================================="
}

main "$@"
