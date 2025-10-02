#!/bin/bash
# Test Execution Script
# Provides various test execution modes

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

usage() {
    cat << EOF
Usage: $0 [OPTIONS] [TEST_TYPE]

Test execution script for Terraform Provider Aviatrix

TEST_TYPE:
    unit            Run unit tests only
    acceptance      Run acceptance tests
    aws             Run AWS-specific tests
    azure           Run Azure-specific tests
    gcp             Run GCP-specific tests
    oci             Run OCI-specific tests
    all             Run all tests (default)

OPTIONS:
    -h, --help      Show this help message
    -v, --verbose   Enable verbose output
    -c, --coverage  Generate coverage report
    -p, --parallel  Run tests in parallel
    -f, --filter    Filter tests by pattern (regex)
    -t, --timeout   Set test timeout (default: 30m)
    --docker        Run tests in Docker container
    --no-cache      Disable Go build cache

EXAMPLES:
    $0 unit                     # Run unit tests
    $0 acceptance --verbose     # Run acceptance tests with verbose output
    $0 aws -c                   # Run AWS tests with coverage
    $0 --filter "Gateway"       # Run tests matching "Gateway"
    $0 --docker unit            # Run unit tests in Docker

EOF
    exit 0
}

# Default values
TEST_TYPE="all"
VERBOSE=false
COVERAGE=false
PARALLEL=false
FILTER=""
TIMEOUT="30m"
USE_DOCKER=false
NO_CACHE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        -f|--filter)
            FILTER="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --docker)
            USE_DOCKER=true
            shift
            ;;
        --no-cache)
            NO_CACHE=true
            shift
            ;;
        unit|acceptance|aws|azure|gcp|oci|all)
            TEST_TYPE="$1"
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Load environment if .env.test exists
if [ -f .env.test ]; then
    log_info "Loading environment from .env.test..."
    export $(grep -v '^#' .env.test | xargs)
fi

# Ensure test directories exist
mkdir -p test-results
mkdir -p test-results/logs

# Build test command
build_test_cmd() {
    local cmd="go test"

    if [ "$VERBOSE" = true ]; then
        cmd="$cmd -v"
    fi

    if [ "$COVERAGE" = true ]; then
        cmd="$cmd -coverprofile=test-results/coverage.out"
    fi

    if [ "$PARALLEL" = true ]; then
        cmd="$cmd -parallel=4"
    fi

    cmd="$cmd -timeout=$TIMEOUT"

    if [ -n "$FILTER" ]; then
        cmd="$cmd -run=$FILTER"
    fi

    if [ "$NO_CACHE" = true ]; then
        cmd="$cmd -count=1"
    fi

    echo "$cmd"
}

# Run tests in Docker
run_docker_tests() {
    local target="$1"

    log_info "Building Docker image for testing..."
    docker build -t terraform-provider-aviatrix:test --target=test .

    log_info "Running tests in Docker container..."
    docker run --rm \
        -v $(pwd)/test-results:/app/test-results \
        -e TF_ACC \
        -e AWS_ACCESS_KEY_ID \
        -e AWS_SECRET_ACCESS_KEY \
        -e AWS_DEFAULT_REGION \
        -e ARM_CLIENT_ID \
        -e ARM_CLIENT_SECRET \
        -e ARM_SUBSCRIPTION_ID \
        -e ARM_TENANT_ID \
        -e GOOGLE_APPLICATION_CREDENTIALS \
        -e GOOGLE_PROJECT \
        -e OCI_USER_ID \
        -e OCI_TENANCY_ID \
        -e OCI_FINGERPRINT \
        -e OCI_PRIVATE_KEY_PATH \
        -e OCI_REGION \
        -e AVIATRIX_CONTROLLER_IP \
        -e AVIATRIX_USERNAME \
        -e AVIATRIX_PASSWORD \
        terraform-provider-aviatrix:test \
        sh -c "$(build_test_cmd) ./..."
}

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."

    if [ "$USE_DOCKER" = true ]; then
        run_docker_tests "test"
    else
        local cmd=$(build_test_cmd)
        log_info "Command: $cmd ./..."
        eval "$cmd ./..." 2>&1 | tee test-results/logs/unit-tests.log
    fi

    if [ "$COVERAGE" = true ]; then
        log_info "Generating coverage report..."
        go tool cover -html=test-results/coverage.out -o test-results/coverage.html
        log_success "Coverage report: test-results/coverage.html"
    fi
}

# Run acceptance tests
run_acceptance_tests() {
    log_info "Running acceptance tests..."

    export TF_ACC=1

    if [ "$USE_DOCKER" = true ]; then
        run_docker_tests "ci-test"
    else
        local cmd=$(build_test_cmd)
        log_info "Command: TF_ACC=1 $cmd ./..."
        eval "$cmd ./..." 2>&1 | tee test-results/logs/acceptance-tests.log
    fi
}

# Run cloud-specific tests
run_cloud_tests() {
    local cloud="$1"
    log_info "Running $cloud tests..."

    export TF_ACC=1

    case $cloud in
        aws)
            export SKIP_ACCOUNT_AZURE=yes
            export SKIP_ACCOUNT_GCP=yes
            export SKIP_ACCOUNT_OCI=yes
            FILTER="${FILTER:-.*AWS.*}"
            ;;
        azure)
            export SKIP_ACCOUNT_AWS=yes
            export SKIP_ACCOUNT_GCP=yes
            export SKIP_ACCOUNT_OCI=yes
            FILTER="${FILTER:-.*Azure.*}"
            ;;
        gcp)
            export SKIP_ACCOUNT_AWS=yes
            export SKIP_ACCOUNT_AZURE=yes
            export SKIP_ACCOUNT_OCI=yes
            FILTER="${FILTER:-.*GCP.*}"
            ;;
        oci)
            export SKIP_ACCOUNT_AWS=yes
            export SKIP_ACCOUNT_AZURE=yes
            export SKIP_ACCOUNT_GCP=yes
            FILTER="${FILTER:-.*OCI.*}"
            ;;
    esac

    local cmd=$(build_test_cmd)
    log_info "Command: $cmd ./..."
    eval "$cmd ./..." 2>&1 | tee test-results/logs/$cloud-tests.log
}

# Main execution
log_info "Starting test execution (type: $TEST_TYPE)"

case $TEST_TYPE in
    unit)
        run_unit_tests
        ;;
    acceptance)
        run_acceptance_tests
        ;;
    aws|azure|gcp|oci)
        run_cloud_tests "$TEST_TYPE"
        ;;
    all)
        log_info "Running all tests..."
        run_unit_tests
        run_acceptance_tests
        ;;
    *)
        log_error "Unknown test type: $TEST_TYPE"
        usage
        ;;
esac

log_success "Test execution completed!"
log_info "Test results saved to: test-results/"
