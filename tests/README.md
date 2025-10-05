# Test Infrastructure Foundation

This directory contains the comprehensive testing framework for the Terraform Provider Aviatrix.

## Directory Structure

```
tests/
├── integration/     # Integration tests for resource/data source testing
├── e2e/            # End-to-end workflow tests
├── fixtures/       # Test fixtures and mock data
├── utils/          # Shared test utilities and helpers
└── README.md       # This file
```

## Test Categories

### Integration Tests (`integration/`)
Integration tests verify individual Terraform resources and data sources against real cloud infrastructure. These tests use the TF_ACC=1 flag and require cloud provider credentials.

**Location**: `tests/integration/`
**Run Command**: `make testacc`
**Requirements**:
- Cloud provider credentials (AWS, Azure, GCP, OCI)
- Aviatrix Controller access
- TF_ACC=1 environment variable

### End-to-End Tests (`e2e/`)
End-to-end tests validate complete workflows and complex multi-resource scenarios that represent real-world use cases.

**Location**: `tests/e2e/`
**Run Command**: `make teste2e`
**Requirements**:
- All integration test requirements
- Extended timeout configurations
- Complete network infrastructure

### Fixtures (`fixtures/`)
Test fixtures provide reusable configuration templates, mock data, and test infrastructure definitions.

**Location**: `tests/fixtures/`
**Contents**:
- Terraform configuration templates
- JSON/YAML test data
- Mock API responses
- Provider configuration examples

### Utilities (`utils/`)
Shared test utilities and helper functions used across all test types.

**Location**: `tests/utils/`
**Contents**:
- Common test helper functions
- Resource cleanup utilities
- Test data generators
- Assertion helpers

## Running Tests

### Prerequisites
```bash
# Copy and configure environment variables
cp .env.example .env
# Edit .env with your credentials

# Source the environment
source .env
```

### Unit Tests
```bash
# Run all unit tests
make test

# Run unit tests with coverage
make test-coverage
```

### Integration Tests
```bash
# Run all integration tests
TF_ACC=1 make testacc

# Run specific resource tests
TF_ACC=1 go test -v ./aviatrix -run TestAccAviatrixGateway

# Run with specific provider only
SKIP_ACCOUNT_AZURE=yes SKIP_ACCOUNT_GCP=yes SKIP_ACCOUNT_OCI=yes \
TF_ACC=1 make testacc
```

### End-to-End Tests
```bash
# Run all E2E tests
make teste2e

# Run specific E2E workflow
go test -v ./tests/e2e -run TestE2E_MultiCloudTransit
```

### Docker-based Tests
```bash
# Run all tests in Docker
docker-compose -f docker-compose.test.yml up

# Run specific test suite
docker-compose -f docker-compose.test.yml up unit-tests
docker-compose -f docker-compose.test.yml up integration-tests-aws
```

## CI/CD Integration

Tests are automatically run via GitHub Actions on:
- Pull requests to main/master
- Pushes to main/master
- Nightly scheduled runs (2 AM UTC)

See `.github/workflows/test-matrix.yml` for the complete CI/CD pipeline configuration.

## Test Environment Setup

### Local Development
```bash
# Setup test infrastructure
cd test-infra
terraform init
terraform apply

# Export environment variables
source cmdExportOutput.sh
```

### Docker Environment
```bash
# Build test containers
docker build -t aviatrix-test:latest --target test .

# Run tests in container
docker run --rm -v $(pwd)/test-results:/app/test-results aviatrix-test:latest
```

## Secrets Management

Cloud provider credentials and Aviatrix Controller access are managed via:

1. **Local Development**: `.env` file (gitignored)
2. **CI/CD**: GitHub Secrets
3. **Docker**: Environment variables passed via docker-compose

Required secrets:
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
- `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID`
- `GOOGLE_APPLICATION_CREDENTIALS`, `GOOGLE_PROJECT`
- `OCI_USER_ID`, `OCI_TENANCY_ID`, `OCI_FINGERPRINT`, `OCI_PRIVATE_KEY_PATH`
- `AVIATRIX_CONTROLLER_IP`, `AVIATRIX_USERNAME`, `AVIATRIX_PASSWORD`

## Contributing

When adding new tests:

1. Place integration tests in `tests/integration/`
2. Place E2E tests in `tests/e2e/`
3. Add fixtures to `tests/fixtures/`
4. Add shared utilities to `tests/utils/`
5. Update this README with any new test patterns
6. Ensure tests pass in CI before merging

## Documentation

- [Complete Testing PRD](../COMPLETE_TESTING_PRD.md)
- [Test Infrastructure Details](../TEST_INFRASTRUCTURE.md)
- [Acceptance Test Guide](../test-infra/README_accep_test.md)
