# Test Infrastructure Documentation

This document describes the test infrastructure for the Terraform Provider Aviatrix, including setup, execution, and troubleshooting.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Setup](#setup)
- [Running Tests](#running-tests)
- [Docker-Based Testing](#docker-based-testing)
- [CI/CD Integration](#cicd-integration)
- [Test Utilities](#test-utilities)
- [Troubleshooting](#troubleshooting)

## Overview

The test infrastructure provides:

- **Multi-stage Docker builds** for isolated test environments
- **GitHub Actions workflows** with matrix testing across cloud providers
- **Comprehensive test utilities** for common testing patterns
- **Environment validation** to ensure proper configuration
- **Automated test reporting** and artifact storage

## Architecture

### Components

```
terraform-provider-aviatrix/
├── .github/workflows/
│   └── test-matrix.yml          # CI/CD pipeline configuration
├── aviatrix/
│   ├── testing_utils.go         # Test utility functions
│   └── infrastructure_smoke_test.go  # Infrastructure validation tests
├── scripts/
│   └── test-env-setup.sh        # Environment setup and validation
├── test-infra/                  # Terraform test infrastructure
├── Dockerfile                   # Multi-stage build for testing
├── docker-compose.test.yml      # Test orchestration
└── .env.test.example            # Environment variable template
```

### Test Stages

1. **Builder Stage**: Builds the provider binary
2. **Test Stage**: Runs unit tests with coverage
3. **Production Stage**: Creates minimal runtime image
4. **CI/CD Stage**: Full integration testing with cloud providers

## Setup

### Prerequisites

- Go 1.23+ ([installation guide](https://golang.org/doc/install))
- Terraform 1.6.6+ ([installation guide](https://www.terraform.io/downloads))
- Docker (optional, for containerized testing)
- Cloud provider credentials (AWS, Azure, GCP, and/or OCI)

### Environment Configuration

1. Copy the example environment file:

```bash
cp .env.test.example .env.test
```

2. Edit `.env.test` and fill in your credentials:

```bash
# Core configuration
export TF_ACC=1
export AVIATRIX_CONTROLLER_IP=your-controller-ip
export AVIATRIX_USERNAME=your-username
export AVIATRIX_PASSWORD=your-password

# AWS (if testing AWS resources)
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_ACCOUNT_NUMBER=your-account-number

# Azure (if testing Azure resources)
export ARM_CLIENT_ID=your-client-id
export ARM_CLIENT_SECRET=your-client-secret
export ARM_SUBSCRIPTION_ID=your-subscription-id
export ARM_TENANT_ID=your-tenant-id

# GCP (if testing GCP resources)
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
export GOOGLE_PROJECT=your-project-id

# OCI (if testing OCI resources)
export OCI_USER_ID=your-user-ocid
export OCI_TENANCY_ID=your-tenancy-ocid
export OCI_FINGERPRINT=your-fingerprint
export OCI_PRIVATE_KEY_PATH=/path/to/private-key.pem
export OCI_REGION=your-region
```

3. Validate your environment:

```bash
./scripts/test-env-setup.sh
```

This script will:
- Check all required environment variables
- Validate cloud provider credentials
- Create necessary test directories
- Verify tool installations

## Running Tests

### Local Testing

#### Unit Tests

Run all unit tests with coverage:

```bash
make test
```

Or with Go directly:

```bash
go test -v -race -coverprofile=coverage.out ./...
```

Generate coverage report:

```bash
go tool cover -html=coverage.out -o coverage.html
```

#### Acceptance Tests

Run acceptance tests for all providers:

```bash
make testacc
```

Run tests for a specific resource:

```bash
TF_ACC=1 go test -v ./aviatrix -run TestAccAviatrixGateway
```

#### Smoke Tests

Run infrastructure smoke tests:

```bash
go test -v ./aviatrix -run TestInfrastructure
```

### Using Test Utilities

The `testing_utils.go` file provides helper functions:

```go
import "github.com/AviatrixSystems/terraform-provider-aviatrix/v3/aviatrix"

func TestMyResource(t *testing.T) {
    // Get test configuration
    config := aviatrix.GetTestConfig(t)

    // Generate random resource name
    name := aviatrix.RandomResourceName("gateway", "test")

    // Check if provider tests should be skipped
    aviatrix.PreCheckFunc("aws")(t)

    // Create artifact directory for this test
    dir := aviatrix.CreateTestArtifactDir(t)

    // Save test artifact
    aviatrix.SaveTestArtifact(t, "output.txt", "test output")
}
```

## Docker-Based Testing

### Build Test Images

Build specific stage:

```bash
docker build --target test -t terraform-provider-aviatrix:test .
docker build --target ci-test -t terraform-provider-aviatrix:ci-test .
```

### Run Tests in Docker

#### Unit Tests

```bash
docker-compose -f docker-compose.test.yml up unit-tests
```

#### Integration Tests

Run tests for specific provider:

```bash
# AWS
docker-compose -f docker-compose.test.yml up integration-tests-aws

# Azure
docker-compose -f docker-compose.test.yml up integration-tests-azure

# GCP
docker-compose -f docker-compose.test.yml up integration-tests-gcp

# OCI
docker-compose -f docker-compose.test.yml up integration-tests-oci
```

Run all tests:

```bash
docker-compose -f docker-compose.test.yml up
```

### Test Orchestration

The `docker-compose.test.yml` file provides:

- **Isolated test networks** for each test suite
- **Automatic dependency management** (integration tests wait for unit tests)
- **Volume mounting** for test results and artifacts
- **Health checks** to monitor test progress
- **Test aggregation** service for result reporting

## CI/CD Integration

### GitHub Actions Workflow

The `.github/workflows/test-matrix.yml` workflow provides:

- **Path filtering** to skip unnecessary test runs
- **Matrix testing** across Go versions (1.23, 1.24)
- **Docker build caching** for faster builds
- **Parallel test execution** across cloud providers
- **Test result publishing** with JUnit format
- **Coverage reporting** in multiple formats
- **Artifact retention** for 30 days
- **Security scanning** with Gosec

### Workflow Triggers

- **Pull requests** to `main` or `master` branches
- **Pushes** to `main` or `master` branches
- **Nightly schedule** at 2 AM UTC
- **Manual dispatch** via GitHub Actions UI

### Test Matrix

```yaml
Unit Tests:
  - Go 1.23
  - Go 1.24

Docker Builds:
  - builder
  - test
  - production
  - ci-test

Integration Tests:
  - AWS
  - Azure
  - GCP
  - OCI
```

### Required Secrets

Configure these in GitHub repository settings:

#### Aviatrix Controller
- `AVIATRIX_CONTROLLER_IP`
- `AVIATRIX_USERNAME`
- `AVIATRIX_PASSWORD`

#### AWS
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_ACCOUNT_NUMBER`
- `AWS_DEFAULT_REGION` (optional, defaults to us-east-1)

#### Azure
- `ARM_CLIENT_ID`
- `ARM_CLIENT_SECRET`
- `ARM_SUBSCRIPTION_ID`
- `ARM_TENANT_ID`

#### GCP
- `GOOGLE_APPLICATION_CREDENTIALS` (base64-encoded JSON)
- `GOOGLE_PROJECT`

#### OCI
- `OCI_USER_ID`
- `OCI_TENANCY_ID`
- `OCI_FINGERPRINT`
- `OCI_PRIVATE_KEY_PATH` (base64-encoded PEM)
- `OCI_REGION`

## Test Utilities

### Available Functions

#### Configuration Management

```go
// GetTestConfig returns test configuration from environment
func GetTestConfig(t *testing.T) *TestingConfig

// PreCheckFunc returns a pre-check function for cloud provider
func PreCheckFunc(provider string) func(*testing.T)

// TestAcceptance checks if acceptance tests should run
func TestAcceptance(t *testing.T)
```

#### Resource Management

```go
// RandomResourceName generates random resource name
func RandomResourceName(prefix, suffix string) string

// NewResourceTestCase creates resource.TestCase with defaults
func NewResourceTestCase(t *testing.T, steps []TestStepConfig) resource.TestCase
```

#### Test Artifacts

```go
// CreateTestArtifactDir creates directory for test artifacts
func CreateTestArtifactDir(t *testing.T) string

// SaveTestArtifact saves content to test artifact file
func SaveTestArtifact(t *testing.T, filename, content string) error
```

#### Environment Checks

```go
// RequireEnvVars fails test if environment variables not set
func RequireEnvVars(t *testing.T, vars ...string)

// SkipTestIfEnvSet skips test if environment variable is set
func SkipTestIfEnvSet(t *testing.T, envVar string)

// ParallelTestAllowed checks if parallel testing is enabled
func ParallelTestAllowed() bool

// DetailedLogsEnabled checks if detailed logging is enabled
func DetailedLogsEnabled() bool

// CheckTestTimeout checks if test is approaching timeout
func CheckTestTimeout(t *testing.T) bool
```

### Environment Variables

#### Test Configuration

- `TF_ACC`: Set to `1` to enable acceptance tests (required)
- `GO_TEST_TIMEOUT`: Test timeout duration (default: `30m`)
- `TEST_ARTIFACT_DIR`: Directory for test artifacts (default: `./test-results`)
- `TEST_DATA_DIR`: Directory for test data (default: `./test-data`)
- `ENABLE_PARALLEL_TESTS`: Enable parallel test execution (default: `true`)
- `ENABLE_DETAILED_LOGS`: Enable detailed test logging (default: `false`)

#### Resource Naming

- `TEST_RESOURCE_PREFIX`: Prefix for test resources (default: `tf-test`)
- `TEST_RESOURCE_SUFFIX`: Suffix for test resources (optional)

#### Provider Skip Flags

- `SKIP_ACCOUNT_AWS`: Set to `yes` to skip AWS tests
- `SKIP_ACCOUNT_AZURE`: Set to `yes` to skip Azure tests
- `SKIP_ACCOUNT_GCP`: Set to `yes` to skip GCP tests
- `SKIP_ACCOUNT_OCI`: Set to `yes` to skip OCI tests

## Troubleshooting

### Common Issues

#### "TF_ACC must be set to 1"

Acceptance tests require the `TF_ACC` environment variable:

```bash
export TF_ACC=1
go test ./...
```

#### "Required environment variables not set"

Run the environment setup script to identify missing variables:

```bash
./scripts/test-env-setup.sh
```

#### Docker Build Failures

Clear Docker cache and rebuild:

```bash
docker system prune -a
docker-compose -f docker-compose.test.yml build --no-cache
```

#### Test Timeouts

Increase the test timeout:

```bash
export GO_TEST_TIMEOUT=60m
go test -timeout=60m ./...
```

Or in docker-compose:

```yaml
environment:
  - GO_TEST_TIMEOUT=60m
```

#### Cloud Provider Authentication Failures

##### AWS
```bash
# Verify credentials
aws sts get-caller-identity

# Check environment variables
echo $AWS_ACCESS_KEY_ID
```

##### Azure
```bash
# Login and verify
az login
az account show
```

##### GCP
```bash
# Verify credentials file exists and is valid
cat $GOOGLE_APPLICATION_CREDENTIALS
gcloud auth application-default print-access-token
```

##### OCI
```bash
# Verify private key file exists
ls -l $OCI_PRIVATE_KEY_PATH

# Check permissions
chmod 600 $OCI_PRIVATE_KEY_PATH
```

### Test Artifacts

Test results are stored in `./test-results/`:

```
test-results/
├── unit-tests.log          # Unit test execution log
├── unit-tests.xml          # JUnit format test results
├── coverage.out            # Coverage profile
├── coverage.xml            # Coverage in XML format
├── coverage/
│   └── coverage.html       # HTML coverage report
├── integration-aws.log     # AWS integration test log
├── integration-azure.log   # Azure integration test log
├── integration-gcp.log     # GCP integration test log
├── integration-oci.log     # OCI integration test log
└── summary/
    └── report.md           # Test summary report
```

### Debugging Tests

#### Enable Detailed Logging

```bash
export ENABLE_DETAILED_LOGS=true
export TF_LOG=DEBUG
go test -v ./...
```

#### Run Specific Test

```bash
go test -v ./aviatrix -run TestAccAviatrixGateway_basic
```

#### Run Tests with Race Detector

```bash
go test -race ./...
```

#### Generate Test Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Getting Help

- **GitHub Issues**: [Report bugs or request features](https://github.com/AviatrixSystems/terraform-provider-aviatrix/issues)
- **Documentation**: [Provider documentation](https://registry.terraform.io/providers/AviatrixSystems/aviatrix/latest/docs)
- **Community**: [Aviatrix Community](https://community.aviatrix.com/)

## Best Practices

1. **Always run smoke tests** before full test suite:
   ```bash
   go test -v ./aviatrix -run TestInfrastructure
   ```

2. **Use environment validation** before CI/CD setup:
   ```bash
   ./scripts/test-env-setup.sh
   ```

3. **Run tests locally** before pushing:
   ```bash
   make test
   ```

4. **Check coverage** for new code:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out
   ```

5. **Clean up test resources** after failures:
   ```bash
   # Check for dangling test resources in cloud providers
   # Use TEST_RESOURCE_PREFIX to identify them
   ```

6. **Use parallel testing** for faster results:
   ```bash
   export ENABLE_PARALLEL_TESTS=true
   go test -parallel=4 ./...
   ```

7. **Monitor test artifacts** for debugging:
   ```bash
   tail -f test-results/unit-tests.log
   ```

## Next Steps

- Review [PRD_100_Percent_Testing.md](../PRD_100_Percent_Testing.md) for the comprehensive testing roadmap
- Check [TEST_INFRASTRUCTURE.md](../TEST_INFRASTRUCTURE.md) for infrastructure details
- Explore [test-infra/README_accep_test.md](../test-infra/README_accep_test.md) for acceptance test setup
