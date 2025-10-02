# Test Infrastructure Documentation

## Overview

This document describes the comprehensive test infrastructure for the Terraform Provider Aviatrix, including Docker containerization, CI/CD pipeline configuration, and test framework utilities.

## Architecture

### Components

1. **Docker Multi-Stage Build** (`Dockerfile`)
   - Builder stage: Compiles the provider
   - Test stage: Runs unit tests
   - Production stage: Creates minimal runtime image
   - CI-test stage: Full integration testing with cloud provider tools

2. **Docker Compose** (`docker-compose.test.yml`)
   - Unit test orchestration
   - Multi-cloud integration test services (AWS, Azure, GCP, OCI)
   - Isolated test environments

3. **GitHub Actions** (`.github/workflows/test-matrix.yml`)
   - Matrix testing across Go versions and cloud providers
   - Automated Docker builds
   - Test artifact collection
   - Coverage reporting
   - Security scanning

4. **Test Framework** (`aviatrix/test_*.go`)
   - Test helpers and utilities
   - Environment management
   - Logging and metrics
   - Cloud provider configurations

5. **Scripts** (`scripts/`)
   - `setup-test-env.sh`: Environment setup and validation
   - `run-tests.sh`: Flexible test execution

## Quick Start

### Prerequisites

- Go 1.23+ installed
- Docker and Docker Compose (for containerized tests)
- Cloud provider credentials (for acceptance tests)

### Setup

1. **Copy and configure environment file:**
   ```bash
   cp .env.test.example .env.test
   # Edit .env.test with your actual credentials
   ```

2. **Run setup script:**
   ```bash
   ./scripts/setup-test-env.sh
   ```

3. **Run tests:**
   ```bash
   # Unit tests only
   ./scripts/run-tests.sh unit

   # Acceptance tests
   ./scripts/run-tests.sh acceptance

   # Cloud-specific tests
   ./scripts/run-tests.sh aws
   ./scripts/run-tests.sh azure
   ./scripts/run-tests.sh gcp
   ./scripts/run-tests.sh oci

   # All tests
   ./scripts/run-tests.sh all
   ```

## Test Execution Methods

### 1. Direct Go Test

```bash
# Unit tests
go test -v ./aviatrix/

# With coverage
go test -v -coverprofile=coverage.out ./aviatrix/

# Specific test
go test -v -run TestInfrastructureSetup ./aviatrix/

# Acceptance tests
TF_ACC=1 go test -v -timeout 30m ./aviatrix/
```

### 2. Using Test Script

```bash
# Basic usage
./scripts/run-tests.sh unit

# With options
./scripts/run-tests.sh unit --verbose --coverage

# Filter tests
./scripts/run-tests.sh --filter "Gateway" acceptance

# Docker-based execution
./scripts/run-tests.sh --docker unit
```

### 3. Docker Compose

```bash
# Unit tests
docker-compose -f docker-compose.test.yml up unit-tests

# AWS integration tests
docker-compose -f docker-compose.test.yml up integration-tests-aws

# All tests
docker-compose -f docker-compose.test.yml up
```

### 4. GitHub Actions

Tests run automatically on:
- Pull requests to `main`/`master`
- Pushes to `main`/`master`
- Nightly at 2 AM UTC

## Environment Variables

### Required for Acceptance Tests

```bash
# Aviatrix Controller
AVIATRIX_CONTROLLER_IP=<controller-ip>
AVIATRIX_USERNAME=<username>
AVIATRIX_PASSWORD=<password>
```

### Cloud Provider Credentials

#### AWS
```bash
AWS_ACCESS_KEY_ID=<key-id>
AWS_SECRET_ACCESS_KEY=<secret-key>
AWS_DEFAULT_REGION=us-east-1
AWS_ACCOUNT_NUMBER=<account-number>
```

#### Azure
```bash
ARM_CLIENT_ID=<client-id>
ARM_CLIENT_SECRET=<client-secret>
ARM_SUBSCRIPTION_ID=<subscription-id>
ARM_TENANT_ID=<tenant-id>
```

#### GCP
```bash
GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
GOOGLE_PROJECT=<project-id>
```

#### OCI
```bash
OCI_USER_ID=<user-ocid>
OCI_TENANCY_ID=<tenancy-ocid>
OCI_FINGERPRINT=<fingerprint>
OCI_PRIVATE_KEY_PATH=/path/to/private-key.pem
OCI_REGION=us-ashburn-1
```

### Test Control Variables

```bash
# Skip specific cloud providers
SKIP_ACCOUNT_AWS=yes
SKIP_ACCOUNT_AZURE=yes
SKIP_ACCOUNT_GCP=yes
SKIP_ACCOUNT_OCI=yes

# Test configuration
TF_ACC=1                          # Enable acceptance tests
GO_TEST_TIMEOUT=30m               # Test timeout
GOMAXPROCS=4                      # CPU cores for tests
TEST_ARTIFACT_DIR=./test-results  # Artifact directory

# Feature flags
ENABLE_PARALLEL_TESTS=true
ENABLE_DETAILED_LOGS=false
ENABLE_SCREENSHOTS=false

# Special tests
SKIP_CID_EXPIRY=yes              # Skip 1.5 hour CID expiry test
```

## Test Framework Components

### Test Helpers (`test_helpers.go`)

```go
// Create test environment
env := NewTestEnvironment()

// Validate credentials
env.ValidateAWSCredentials(t)
env.ValidateAzureCredentials(t)

// Pre-check functions
PreCheckAWS(t, env)
PreCheckAzure(t, env)

// Test case builder
tc := NewTestCase().
    WithCloudPreCheck(t, PreCheckAWS).
    WithSteps(step1, step2).
    Run(t)
```

### Test Configuration (`test_config.go`)

```go
// Get default configuration
config := DefaultTestConfig()

// Ensure directories exist
config.EnsureDirectories()

// Cloud provider configuration
cloudConfig := DefaultCloudProviderTestConfig()
```

### Test Logging (`test_logger.go`)

```go
// Create logger
logger, _ := NewTestLogger(t, "my_test")
defer logger.Close()

// Log messages
logger.Info("Starting test")
logger.Debug("Debug information")
logger.Warn("Warning message")
logger.Error("Error occurred")
logger.Step(1, "Create resource")
logger.Resource("CREATE", "gateway", "test-gateway")

// Track metrics
metrics := NewTestMetrics("my_test")
metrics.RecordResourceCreated()
metrics.RecordAPICall()
metrics.Finalize()
logger.Info(metrics.Summary())
```

## Directory Structure

```
terraform-provider-aviatrix/
├── .github/workflows/
│   └── test-matrix.yml          # CI/CD pipeline
├── aviatrix/
│   ├── test_helpers.go          # Test helper functions
│   ├── test_config.go           # Test configuration
│   ├── test_logger.go           # Logging and metrics
│   ├── infrastructure_test.go   # Infrastructure validation
│   └── *_test.go                # Test files
├── scripts/
│   ├── setup-test-env.sh        # Environment setup
│   └── run-tests.sh             # Test execution
├── test-results/                # Test artifacts (created)
│   ├── logs/                    # Test logs
│   ├── screenshots/             # Test screenshots
│   ├── coverage.out             # Coverage data
│   └── coverage.html            # Coverage report
├── test-data/                   # Test data files
├── Dockerfile                   # Multi-stage build
├── docker-compose.test.yml      # Test orchestration
├── .env.test.example            # Environment template
└── TEST_INFRASTRUCTURE.md       # This file
```

## GitHub Actions Workflow

### Jobs

1. **changes**: Detect which files changed
2. **unit-tests**: Run unit tests on Go 1.23 and 1.24
3. **docker-build**: Build all Docker stages
4. **integration-tests**: Matrix test across AWS/Azure/GCP/OCI
5. **security-scan**: Run Gosec security scanner
6. **test-summary**: Aggregate results

### Matrix Testing

- **Go Versions**: 1.23, 1.24
- **Cloud Providers**: AWS, Azure, GCP, OCI
- **Docker Stages**: builder, test, production, ci-test

### Artifacts

- Test results (logs, XML reports)
- Coverage reports (HTML, XML)
- Docker images (cached)
- Security scan results (SARIF)

## Docker Stages

### 1. Builder Stage
- Base: `golang:1.23-alpine`
- Purpose: Compile provider binary
- Output: `terraform-provider-aviatrix` binary

### 2. Test Stage
- Base: `golang:1.23-alpine`
- Purpose: Run unit tests
- Tools: go-junit-report, gocov, gocov-xml
- Command: `go test -v ./...`

### 3. Production Stage
- Base: `alpine:3.19`
- Purpose: Minimal runtime image
- Size: ~20MB
- User: Non-root (terraform:1001)

### 4. CI-Test Stage
- Base: `golang:1.23`
- Purpose: Integration testing
- Tools: Terraform, AWS CLI, Azure CLI, gcloud, OCI CLI
- Full test environment with all cloud provider tools

## CI/CD Pipeline Features

### Optimizations

- Path filtering (skip docs-only changes)
- Build caching (Go modules, Docker layers)
- Parallel execution (multiple Go versions, cloud providers)
- Conditional execution (skip if no relevant changes)

### Security

- Gosec static analysis
- SARIF report upload to GitHub Security
- Secrets management via GitHub Secrets
- Non-root Docker containers

### Reporting

- JUnit XML test results
- Coverage reports (XML, HTML)
- Test duration tracking
- Failure summary in PR comments

## Test Coverage Goals

Current infrastructure supports:

- ✅ Unit test execution
- ✅ Integration test execution
- ✅ Multi-cloud testing (AWS, Azure, GCP, OCI)
- ✅ Coverage reporting
- ✅ Test artifact collection
- ✅ Logging and metrics
- ✅ CI/CD automation
- ✅ Docker isolation
- ✅ Security scanning

## Troubleshooting

### Common Issues

1. **Missing credentials**
   ```bash
   ./scripts/setup-test-env.sh
   # Check output for which providers are configured
   ```

2. **Docker build fails**
   ```bash
   # Check Docker daemon
   docker info

   # Rebuild without cache
   docker build --no-cache -t terraform-provider-aviatrix:test .
   ```

3. **Test timeout**
   ```bash
   # Increase timeout
   ./scripts/run-tests.sh unit -t 60m
   ```

4. **Permission denied on scripts**
   ```bash
   chmod +x scripts/*.sh
   ```

### Debug Mode

```bash
# Enable detailed logging
ENABLE_DETAILED_LOGS=true ./scripts/run-tests.sh unit -v

# Run single test with verbose output
go test -v -run TestSpecificTest ./aviatrix/
```

## Best Practices

### Writing Tests

1. **Use test helpers**
   ```go
   env := NewTestEnvironment()
   PreCheckAWS(t, env)
   ```

2. **Log test steps**
   ```go
   logger.Step(1, "Create resource")
   logger.Resource("CREATE", "gateway", name)
   ```

3. **Track metrics**
   ```go
   metrics := NewTestMetrics("test_name")
   metrics.RecordResourceCreated()
   ```

4. **Use proper cleanup**
   ```go
   defer logger.Close()
   defer metrics.Finalize()
   ```

### CI/CD

1. Always use GitHub Secrets for credentials
2. Tag Docker images with commit SHA
3. Archive test artifacts for debugging
4. Set appropriate timeouts (default: 30m)

### Local Development

1. Use `.env.test` for local credentials
2. Run unit tests before pushing
3. Test Docker builds locally
4. Validate CI workflow changes with `act` (GitHub Actions locally)

## Future Enhancements

- [ ] Add performance benchmarking
- [ ] Implement test retry logic
- [ ] Add mutation testing
- [ ] Create test data generators
- [ ] Implement chaos testing
- [ ] Add smoke tests for common scenarios
- [ ] Create test result dashboard
- [ ] Implement automatic test selection based on code changes

## Support

For issues or questions:
1. Check test logs in `test-results/logs/`
2. Review GitHub Actions workflow runs
3. Open an issue with test output

## References

- [Terraform Plugin SDK Testing](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing)
- [GitHub Actions Documentation](https://docs.github.com/actions)
- [Docker Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Go Testing Package](https://pkg.go.dev/testing)
