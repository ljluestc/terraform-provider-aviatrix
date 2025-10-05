#!/bin/bash
# Infrastructure Validation Script
# Validates that the test infrastructure is properly set up

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Counters
PASS=0
FAIL=0
WARN=0

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

log_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASS++))
}

log_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAIL++))
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARN++))
}

log_info() {
    echo -e "  $1"
}

# Check directory structure
check_directory_structure() {
    echo -e "\n${GREEN}Checking Directory Structure${NC}"

    local dirs=(
        "tests/integration"
        "tests/e2e"
        "tests/fixtures"
        "tests/utils"
        "tests/fixtures/terraform"
        "tests/fixtures/configs"
        "test-infra"
        "test-results"
    )

    for dir in "${dirs[@]}"; do
        if [ -d "${PROJECT_ROOT}/${dir}" ]; then
            log_pass "Directory exists: ${dir}"
        else
            log_fail "Directory missing: ${dir}"
        fi
    done
}

# Check required files
check_required_files() {
    echo -e "\n${GREEN}Checking Required Files${NC}"

    local files=(
        "tests/README.md"
        "tests/integration/README.md"
        "tests/e2e/README.md"
        "tests/fixtures/README.md"
        "tests/utils/README.md"
        "tests/docker-compose.yml"
        "Dockerfile"
        "docker-compose.test.yml"
        ".github/workflows/test-matrix.yml"
        ".github/workflows/comprehensive-tests.yml"
        ".env.test.example"
        "GNUmakefile"
    )

    for file in "${files[@]}"; do
        if [ -f "${PROJECT_ROOT}/${file}" ]; then
            log_pass "File exists: ${file}"
        else
            log_fail "File missing: ${file}"
        fi
    done
}

# Check executable scripts
check_executable_scripts() {
    echo -e "\n${GREEN}Checking Executable Scripts${NC}"

    local scripts=(
        "tests/fixtures/terraform/test_env_setup.sh"
        "tests/fixtures/terraform/test_env_teardown.sh"
        "scripts/test-env-setup.sh"
        "scripts/test-runner.sh"
    )

    for script in "${scripts[@]}"; do
        if [ -f "${PROJECT_ROOT}/${script}" ]; then
            if [ -x "${PROJECT_ROOT}/${script}" ]; then
                log_pass "Script is executable: ${script}"
            else
                log_warn "Script exists but not executable: ${script}"
            fi
        else
            log_fail "Script missing: ${script}"
        fi
    done
}

# Check Docker configuration
check_docker_config() {
    echo -e "\n${GREEN}Checking Docker Configuration${NC}"

    # Check if Docker is installed
    if command -v docker &> /dev/null; then
        log_pass "Docker is installed: $(docker --version)"
    else
        log_warn "Docker is not installed"
    fi

    # Check if docker-compose is installed
    if command -v docker-compose &> /dev/null; then
        log_pass "docker-compose is installed: $(docker-compose --version)"
    elif docker compose version &> /dev/null; then
        log_pass "docker compose (v2) is installed: $(docker compose version)"
    else
        log_warn "docker-compose is not installed"
    fi

    # Validate Dockerfile
    if [ -f "${PROJECT_ROOT}/Dockerfile" ]; then
        if docker build -t aviatrix-test:validation --target test "${PROJECT_ROOT}" > /dev/null 2>&1; then
            log_pass "Dockerfile builds successfully"
            docker rmi aviatrix-test:validation > /dev/null 2>&1 || true
        else
            log_fail "Dockerfile build failed"
        fi
    fi

    # Validate docker-compose files
    local compose_files=(
        "docker-compose.test.yml"
        "tests/docker-compose.yml"
    )

    for file in "${compose_files[@]}"; do
        if [ -f "${PROJECT_ROOT}/${file}" ]; then
            if docker-compose -f "${PROJECT_ROOT}/${file}" config > /dev/null 2>&1; then
                log_pass "docker-compose file is valid: ${file}"
            elif docker compose -f "${PROJECT_ROOT}/${file}" config > /dev/null 2>&1; then
                log_pass "docker-compose file is valid: ${file}"
            else
                log_fail "docker-compose file is invalid: ${file}"
            fi
        fi
    done
}

# Check Go environment
check_go_environment() {
    echo -e "\n${GREEN}Checking Go Environment${NC}"

    # Check if Go is installed
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}')
        log_pass "Go is installed: ${go_version}"
    else
        log_fail "Go is not installed"
        return
    fi

    # Check go.mod
    if [ -f "${PROJECT_ROOT}/go.mod" ]; then
        log_pass "go.mod exists"
    else
        log_fail "go.mod missing"
    fi

    # Check if dependencies are downloaded
    cd "${PROJECT_ROOT}"
    if go list -m all > /dev/null 2>&1; then
        log_pass "Go dependencies are available"
    else
        log_warn "Go dependencies need to be downloaded (run: go mod download)"
    fi
}

# Check GitHub Actions workflows
check_github_workflows() {
    echo -e "\n${GREEN}Checking GitHub Actions Workflows${NC}"

    local workflows=(
        ".github/workflows/test-matrix.yml"
        ".github/workflows/comprehensive-tests.yml"
    )

    for workflow in "${workflows[@]}"; do
        if [ -f "${PROJECT_ROOT}/${workflow}" ]; then
            # Basic YAML syntax check
            if command -v python3 &> /dev/null; then
                if python3 -c "import yaml; yaml.safe_load(open('${PROJECT_ROOT}/${workflow}'))" 2>/dev/null; then
                    log_pass "Workflow YAML is valid: ${workflow}"
                else
                    log_fail "Workflow YAML is invalid: ${workflow}"
                fi
            else
                log_warn "Cannot validate YAML (python3 not installed): ${workflow}"
            fi
        fi
    done
}

# Check test infrastructure
check_test_infrastructure() {
    echo -e "\n${GREEN}Checking Test Infrastructure${NC}"

    local infra_dirs=(
        "test-infra/aws"
        "test-infra/azure"
        "test-infra/gcp"
        "test-infra/oci"
    )

    for dir in "${infra_dirs[@]}"; do
        if [ -d "${PROJECT_ROOT}/${dir}" ]; then
            log_pass "Infrastructure directory exists: ${dir}"

            # Check for main.tf
            if [ -f "${PROJECT_ROOT}/${dir}/main.tf" ]; then
                log_info "  ✓ main.tf found"
            else
                log_warn "  ⚠ main.tf missing in ${dir}"
            fi
        else
            log_fail "Infrastructure directory missing: ${dir}"
        fi
    done
}

# Check Makefile targets
check_makefile_targets() {
    echo -e "\n${GREEN}Checking Makefile Targets${NC}"

    if [ -f "${PROJECT_ROOT}/GNUmakefile" ]; then
        log_pass "GNUmakefile exists"

        local targets=(
            "test"
            "testacc"
            "build"
            "fmt"
            "fmtcheck"
        )

        for target in "${targets[@]}"; do
            if grep -q "^${target}:" "${PROJECT_ROOT}/GNUmakefile"; then
                log_info "  ✓ Target defined: ${target}"
            else
                log_warn "  ⚠ Target missing: ${target}"
            fi
        done
    else
        log_fail "GNUmakefile missing"
    fi
}

# Check environment configuration
check_environment_config() {
    echo -e "\n${GREEN}Checking Environment Configuration${NC}"

    if [ -f "${PROJECT_ROOT}/.env.test.example" ]; then
        log_pass ".env.test.example exists"
    else
        log_warn ".env.test.example missing"
    fi

    # Check if .env is in .gitignore
    if [ -f "${PROJECT_ROOT}/.gitignore" ]; then
        if grep -q "^\.env$" "${PROJECT_ROOT}/.gitignore"; then
            log_pass ".env is in .gitignore"
        else
            log_warn ".env should be added to .gitignore"
        fi
    fi
}

# Summary
print_summary() {
    echo -e "\n${GREEN}=====================================${NC}"
    echo -e "${GREEN}Validation Summary${NC}"
    echo -e "${GREEN}=====================================${NC}"
    echo -e "${GREEN}Passed:${NC}  ${PASS}"
    echo -e "${YELLOW}Warnings:${NC} ${WARN}"
    echo -e "${RED}Failed:${NC}  ${FAIL}"
    echo -e "${GREEN}=====================================${NC}\n"

    if [ ${FAIL} -eq 0 ]; then
        echo -e "${GREEN}✓ Infrastructure validation PASSED${NC}"
        echo -e "  All critical components are in place."
        if [ ${WARN} -gt 0 ]; then
            echo -e "  ${YELLOW}Note: ${WARN} warning(s) found. Review recommended.${NC}"
        fi
        return 0
    else
        echo -e "${RED}✗ Infrastructure validation FAILED${NC}"
        echo -e "  ${FAIL} critical issue(s) found. Please fix before proceeding."
        return 1
    fi
}

# Main execution
main() {
    echo -e "${GREEN}=====================================${NC}"
    echo -e "${GREEN}Test Infrastructure Validation${NC}"
    echo -e "${GREEN}=====================================${NC}"
    echo -e "Project Root: ${PROJECT_ROOT}\n"

    check_directory_structure
    check_required_files
    check_executable_scripts
    check_docker_config
    check_go_environment
    check_github_workflows
    check_test_infrastructure
    check_makefile_targets
    check_environment_config

    print_summary
}

main
exit $?
