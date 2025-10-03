TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep "aviatrix/")
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=aviatrix
TF_PLUGIN_DIR=~/.terraform.d/plugins
AVIATRIX_PROVIDER_NAMESPACE=aviatrix.com/aviatrix/aviatrix

default: build

build: fmtcheck
	go install

build13: GOOS=$(shell go env GOOS)
build13: GOARCH=$(shell go env GOARCH)
ifeq ($(OS),Windows_NT)  # is Windows_NT on XP, 2000, 7, Vista, 10...
build13: DESTINATION=$(APPDATA)/terraform.d/plugins/$(AVIATRIX_PROVIDER_NAMESPACE)/99.0.0/$(GOOS)_$(GOARCH)
else
build13: DESTINATION=$(HOME)/.terraform.d/plugins/$(AVIATRIX_PROVIDER_NAMESPACE)/99.0.0/$(GOOS)_$(GOARCH)
endif
build13: fmtcheck
	@echo "==> Installing plugin to $(DESTINATION)"
	@mkdir -p $(DESTINATION)
	CGO_ENABLED=0 \
	go build \
		-ldflags "-X github.com/AviatrixSystems/terraform-provider-aviatrix/v3/aviatrix.Version=99.0.0" \
		-ldflags "-extldflags -static -s -w" \
		-o $(DESTINATION)/terraform-provider-aviatrix_v99.0.0

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=1800s -parallel=4

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 1200m

# Enhanced test targets for test infrastructure
test-unit: fmtcheck
	@echo "==> Running unit tests with coverage..."
	@mkdir -p test-results test-results/coverage
	go test -v -race -coverprofile=test-results/coverage.out -timeout=30m ./... 2>&1 | tee test-results/unit-tests.log

test-smoke: fmtcheck
	@echo "==> Running smoke tests..."
	@mkdir -p test-results
	TF_ACC=0 go test -v -run TestSmoke ./aviatrix/... -timeout=5m 2>&1 | tee test-results/smoke-tests.log

test-integration-aws: fmtcheck
	@echo "==> Running AWS integration tests..."
	@./scripts/test-env-setup.sh
	@export TF_ACC=1 SKIP_ACCOUNT_AZURE=yes SKIP_ACCOUNT_GCP=yes SKIP_ACCOUNT_OCI=yes && \
		./scripts/test-runner.sh

test-integration-azure: fmtcheck
	@echo "==> Running Azure integration tests..."
	@./scripts/test-env-setup.sh
	@export TF_ACC=1 SKIP_ACCOUNT_AWS=yes SKIP_ACCOUNT_GCP=yes SKIP_ACCOUNT_OCI=yes && \
		./scripts/test-runner.sh

test-integration-gcp: fmtcheck
	@echo "==> Running GCP integration tests..."
	@./scripts/test-env-setup.sh
	@export TF_ACC=1 SKIP_ACCOUNT_AWS=yes SKIP_ACCOUNT_AZURE=yes SKIP_ACCOUNT_OCI=yes && \
		./scripts/test-runner.sh

test-integration-oci: fmtcheck
	@echo "==> Running OCI integration tests..."
	@./scripts/test-env-setup.sh
	@export TF_ACC=1 SKIP_ACCOUNT_AWS=yes SKIP_ACCOUNT_AZURE=yes SKIP_ACCOUNT_GCP=yes && \
		./scripts/test-runner.sh

test-coverage: test-unit
	@echo "==> Generating coverage reports..."
	@go tool cover -html=test-results/coverage.out -o test-results/coverage.html
	@go tool cover -func=test-results/coverage.out | grep total | awk '{print "Total coverage: " $$3}'

docker-test:
	@echo "==> Running tests in Docker..."
	@docker-compose -f docker-compose.test.yml run --rm unit-tests

docker-test-clean:
	@echo "==> Cleaning up Docker test environment..."
	@docker-compose -f docker-compose.test.yml down -v
	@rm -rf test-results/*

test-env-validate:
	@echo "==> Validating test environment..."
	@./scripts/test-env-setup.sh

test-all: test-unit test-smoke
	@echo "==> All tests completed successfully"

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

imports:
	goimports -w $(GOFMT_FILES)

fmt:
	gofmt -s -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

tools:
	go get -u github.com/kardianos/govendor
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

vendor-status:
	@govendor status

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-lint:
	@echo "==> Checking website against linters..."
	@misspell -error -source=text website/

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: build test testacc vet fmt fmtcheck errcheck tools vendor-status test-compile website-lint website website-test \
	test-unit test-smoke test-integration-aws test-integration-azure test-integration-gcp test-integration-oci \
	test-coverage docker-test docker-test-clean test-env-validate test-all
