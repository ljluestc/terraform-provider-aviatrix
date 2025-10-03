#!/bin/bash
# Test runner script for Terraform Provider Aviatrix
# This script orchestrates test execution with proper logging and artifact collection

set -e

# Default values
TEST_TYPE="${TEST_TYPE:-unit}"
OUTPUT_DIR="${OUTPUT_DIR:-test-results}"
VERBOSE="${VERBOSE:-false}"
PROVIDER="${PROVIDER:-}"
TIMEOUT="${TIMEOUT:-30m}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
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

# Create output directories
setup_directories() {
    log_info "Setting up output directories..."
    mkdir -p "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR/coverage"
    mkdir -p "$OUTPUT_DIR/artifacts"
    mkdir -p "$OUTPUT_DIR/logs"
}

# Check required environment variables
check_prerequisites() {
    log_info "Checking prerequisites..."

    if [ "$TEST_TYPE" = "acceptance" ] || [ "$TEST_TYPE" = "integration" ]; then
        if [ -z "$AVIATRIX_CONTROLLER_IP" ]; then
            log_error "AVIATRIX_CONTROLLER_IP is required for acceptance/integration tests"
            exit 1
        fi

        if [ -z "$AVIATRIX_USERNAME" ]; then
            log_error "AVIATRIX_USERNAME is required for acceptance/integration tests"
            exit 1
        fi

        if [ -z "$AVIATRIX_PASSWORD" ]; then
            log_error "AVIATRIX_PASSWORD is required for acceptance/integration tests"
            exit 1
        fi

        log_success "Aviatrix controller credentials configured"
    fi
}

# Run unit tests
run_unit_tests() {
    log_info "Running unit tests..."

    go test \
        -v \
        -race \
        -timeout="$TIMEOUT" \
        -coverprofile="$OUTPUT_DIR/coverage/coverage.out" \
        -covermode=atomic \
        ./... 2>&1 | tee "$OUTPUT_DIR/logs/unit-tests.log" | \
        go-junit-report -set-exit-code > "$OUTPUT_DIR/unit-tests.xml"

    # Generate coverage reports
    log_info "Generating coverage reports..."
    gocov convert "$OUTPUT_DIR/coverage/coverage.out" | gocov-xml > "$OUTPUT_DIR/coverage/coverage.xml"
    go tool cover -html="$OUTPUT_DIR/coverage/coverage.out" -o "$OUTPUT_DIR/coverage/coverage.html"

    # Coverage summary
    COVERAGE=$(go tool cover -func="$OUTPUT_DIR/coverage/coverage.out" | grep total | awk '{print $3}')
    log_success "Unit tests completed with $COVERAGE coverage"
}

# Run acceptance tests
run_acceptance_tests() {
    log_info "Running acceptance tests..."

    export TF_ACC=1

    if [ -n "$PROVIDER" ]; then
        log_info "Running tests for provider: $PROVIDER"
        SKIP_VAR="SKIP_ACCOUNT_${PROVIDER^^}"
        export "$SKIP_VAR"=no
    fi

    go test \
        -v \
        -timeout="$TIMEOUT" \
        ./aviatrix/... 2>&1 | tee "$OUTPUT_DIR/logs/acceptance-tests.log" | \
        go-junit-report -set-exit-code > "$OUTPUT_DIR/acceptance-tests.xml"

    log_success "Acceptance tests completed"
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests for $PROVIDER..."

    export TF_ACC=1

    # Set up test infrastructure if needed
    if [ -d "test-infra" ]; then
        log_info "Setting up test infrastructure..."
        cd test-infra
        terraform init
        terraform apply -auto-approve

        if [ -f "./cmdExportOutput.sh" ]; then
            source ./cmdExportOutput.sh
        fi
        cd ..
    fi

    # Run provider-specific tests
    if [ -f "test-infra/runAccTest.sh" ]; then
        ./test-infra/runAccTest.sh "$PROVIDER" 2>&1 | tee "$OUTPUT_DIR/logs/integration-${PROVIDER}.log"
    else
        go test \
            -v \
            -timeout="$TIMEOUT" \
            ./aviatrix/... 2>&1 | tee "$OUTPUT_DIR/logs/integration-${PROVIDER}.log" | \
            go-junit-report -set-exit-code > "$OUTPUT_DIR/integration-${PROVIDER}.xml"
    fi

    log_success "Integration tests completed for $PROVIDER"
}

# Generate test summary
generate_summary() {
    log_info "Generating test summary..."

    SUMMARY_FILE="$OUTPUT_DIR/summary.md"

    cat > "$SUMMARY_FILE" << EOF
# Test Execution Summary

**Test Type:** $TEST_TYPE
**Date:** $(date -u +"%Y-%m-%d %H:%M:%S UTC")
**Provider:** ${PROVIDER:-All}

## Results

EOF

    # Count test results from XML files
    if command -v xmllint &> /dev/null; then
        for xml_file in "$OUTPUT_DIR"/*.xml; do
            if [ -f "$xml_file" ]; then
                TESTS=$(xmllint --xpath "string(//testsuites/@tests)" "$xml_file" 2>/dev/null || echo "0")
                FAILURES=$(xmllint --xpath "string(//testsuites/@failures)" "$xml_file" 2>/dev/null || echo "0")
                ERRORS=$(xmllint --xpath "string(//testsuites/@errors)" "$xml_file" 2>/dev/null || echo "0")

                echo "- $(basename "$xml_file"): $TESTS tests, $FAILURES failures, $ERRORS errors" >> "$SUMMARY_FILE"
            fi
        done
    fi

    # Add coverage info if available
    if [ -f "$OUTPUT_DIR/coverage/coverage.out" ]; then
        COVERAGE=$(go tool cover -func="$OUTPUT_DIR/coverage/coverage.out" | grep total | awk '{print $3}')
        echo "" >> "$SUMMARY_FILE"
        echo "**Coverage:** $COVERAGE" >> "$SUMMARY_FILE"
    fi

    # Add log excerpts for failures
    if grep -q "FAIL" "$OUTPUT_DIR"/logs/*.log 2>/dev/null; then
        echo "" >> "$SUMMARY_FILE"
        echo "## Failed Tests" >> "$SUMMARY_FILE"
        echo '```' >> "$SUMMARY_FILE"
        grep -A 5 "FAIL" "$OUTPUT_DIR"/logs/*.log 2>/dev/null | head -n 50 >> "$SUMMARY_FILE" || true
        echo '```' >> "$SUMMARY_FILE"
    fi

    log_success "Test summary generated at $SUMMARY_FILE"

    # Print summary to console
    if [ "$VERBOSE" = "true" ]; then
        cat "$SUMMARY_FILE"
    fi
}

# Clean up resources
cleanup() {
    log_info "Cleaning up test resources..."

    if [ -d "test-infra" ] && [ "$TEST_TYPE" = "integration" ]; then
        cd test-infra
        terraform destroy -auto-approve || log_warning "Failed to destroy test infrastructure"
        cd ..
    fi
}

# Main execution
main() {
    log_info "Starting test execution: $TEST_TYPE"

    setup_directories
    check_prerequisites

    case "$TEST_TYPE" in
        unit)
            run_unit_tests
            ;;
        acceptance)
            run_acceptance_tests
            ;;
        integration)
            if [ -z "$PROVIDER" ]; then
                log_error "PROVIDER must be set for integration tests"
                exit 1
            fi
            run_integration_tests
            ;;
        all)
            run_unit_tests
            run_acceptance_tests
            ;;
        *)
            log_error "Unknown test type: $TEST_TYPE"
            log_info "Valid types: unit, acceptance, integration, all"
            exit 1
            ;;
    esac

    generate_summary

    log_success "Test execution completed successfully"
}

# Trap errors and cleanup
trap cleanup EXIT

# Run main function
main "$@"
