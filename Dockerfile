# Multi-stage Dockerfile for Terraform Provider Aviatrix Testing
# Stage 1: Build stage
FROM golang:1.23-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the provider
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o terraform-provider-aviatrix .

# Stage 2: Test stage
FROM golang:1.23-alpine AS test

# Install testing dependencies
RUN apk add --no-cache git ca-certificates tzdata make curl terraform

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code and test files
COPY . .

# Install additional test tools
RUN go install github.com/jstemmer/go-junit-report/v2@latest
RUN go install github.com/axw/gocov/gocov@latest
RUN go install github.com/AlekSi/gocov-xml@latest
RUN go install github.com/matm/gocov-html@latest
RUN go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest

# Create test results directory
RUN mkdir -p /app/test-results /app/test-artifacts /app/test-logs

# Set default command for tests
CMD ["go", "test", "-v", "./..."]

# Stage 3: Production stage
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Add non-root user
RUN addgroup -g 1001 -S terraform && \
    adduser -u 1001 -S terraform -G terraform

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/terraform-provider-aviatrix /usr/local/bin/terraform-provider-aviatrix

# Change ownership
RUN chown terraform:terraform /usr/local/bin/terraform-provider-aviatrix

# Switch to non-root user
USER terraform

# Default command
CMD ["/usr/local/bin/terraform-provider-aviatrix"]

# Stage 4: CI/CD testing stage with cloud provider tools
FROM golang:1.23 AS ci-test

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    python3 \
    python3-pip \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install Terraform
RUN curl -fsSL https://releases.hashicorp.com/terraform/1.6.6/terraform_1.6.6_linux_amd64.zip -o terraform.zip && \
    unzip terraform.zip && \
    mv terraform /usr/local/bin/ && \
    rm terraform.zip

# Install AWS CLI v2
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf aws awscliv2.zip

# Install Azure CLI
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# Install Google Cloud SDK
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg add - && \
    apt-get update && apt-get install -y google-cloud-sdk

# Install OCI CLI
RUN bash -c "$(curl -L https://raw.githubusercontent.com/oracle/oci-cli/master/scripts/install/install.sh)" -- --accept-all-defaults

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Install Go test tools
RUN go install github.com/jstemmer/go-junit-report/v2@latest && \
    go install github.com/axw/gocov/gocov@latest && \
    go install github.com/AlekSi/gocov-xml@latest && \
    go install github.com/matm/gocov-html@latest && \
    go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest && \
    go install gotest.tools/gotestsum@latest

# Create test directory structure
RUN mkdir -p /app/test-results /app/test-artifacts /app/test-logs

# Copy test helper scripts
COPY scripts/test-*.sh /usr/local/bin/ 2>/dev/null || true
RUN chmod +x /usr/local/bin/test-*.sh 2>/dev/null || true

# Default command for CI tests
CMD ["make", "test"]