# Build parameters.
CGO_ENABLED=0
LD_FLAGS="-extldflags '-static'"

# Go parameters.
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOBUILD=CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) build -v -buildmode=exe -ldflags $(LD_FLAGS)
GO_PACKAGES=./...
GO_TESTS=^.*$

GOLANGCI_LINT_VERSION=v1.30.0

# Disabled linters:
#
# - gci as we use gofmt for formatting.
#
# - goerr113 as this code do not export any API returning errors
#   and internally there is no need to use typed errors.
#
# - testpackage as Terraform testing convention do not encourage to use them.
#
# - godox as it is OK to have TODOs in the code.
DISABLED_LINTERS=gci,goerr113,testpackage,godox

BIN_PATH=$$HOME/bin
TF_ACC=
TINKERBELL_GRPC_AUTHORITY=127.0.0.1:42113
TINKERBELL_CERT_URL=http://127.0.0.1:42114/cert

GITHUB_TOKEN=
GPG_FINGERPRINT=
RELEASE_VERSION=

.PHONY: all ## Build the binary, run unit tests and run linter.
all: build build-test test lint

.PHONY: download
download: ## Download Go module dependencies required for building and testing.
	$(GOMOD) download

.PHONY: install-golangci-lint
install-golangci-lint: ## Installs golangci-lint binary into BIN_PATH.
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BIN_PATH) $(GOLANGCI_LINT_VERSION)

.PHONY: install-cc-test-reporter
install-cc-test-reporter: ## Installs Code Climate test reporter binary into BIN_PATH.
	curl -L https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64 > $(BIN_PATH)/cc-test-reporter
	chmod +x $(BIN_PATH)/cc-test-reporter

.PHONY: install-ci ## Installs binaries required for CI.
install-ci: install-golangci-lint install-cc-test-reporter

.PHONY: build
build: ## Build Terraform provider binary.
	$(GOBUILD)

.PHONY: test
test: build-test ## Run unit tests matching GO_TESTS in GO_PACKAGES.
	TF_ACC=$(TF_ACC) TINKERBELL_GRPC_AUTHORITY=$(TINKERBELL_GRPC_AUTHORITY) TINKERBELL_CERT_URL=$(TINKERBELL_CERT_URL) $(GOTEST) -run $(GO_TESTS) $(GO_PACKAGES)

.PHONY: lint
lint: build build-test ## Compile code and run linter.
	golangci-lint run --enable-all --disable=$(DISABLED_LINTERS) --max-same-issues=0 --max-issues-per-linter=0 --build-tags integration --timeout 10m --exclude-use-default=false $(GO_PACKAGES)

.PHONY: build-test
build-test: # Compile unit tests. Useful for checking syntax errors before running unit tests.
	$(GOTEST) -run=nope $(GO_PACKAGES)

.PHONY: update
update: ## Updates all Go module dependencies.
	$(GOGET) -u $(GO_PACKAGES)
	$(GOMOD) tidy

.PHONY: all-cover
all-cover: build build-test test-cover lint ## Builds the binary, runs unit tests with coverage report and runs linter.

.PHONY: test-cover
test-cover: build-test ## Runs unit tests and writes coverage report to a PROFILEFILE for given GO_PACKAGES.
	$(GOTEST) -run $(GO_TESTS) -coverprofile=$(PROFILEFILE) $(GO_PACKAGES)

.PHONY: cover-upload
cover-upload: codecov ## Runs unit tests with coverage report and uploads it to Codecov and Code Climate.
	# Make codeclimate as command, as we need to run test-cover twice and make deduplicates that.
	# Go test results are cached anyway, so it's fine to run it multiple times.
	make codeclimate

.PHONY: codecov
codecov: PROFILEFILE=coverage.txt
codecov: SHELL=/bin/bash
codecov: test-cover
codecov: ## Runs unit tests with coverage report and uploads it to Codecov.
	bash <(curl -s https://codecov.io/bash)

.PHONY: testacc
testacc: TF_ACC=true
testacc: test ## Runs Terraform acceptance tests.

.PHONY: release-env-check
release-env-check: ## Checks if required environment variables are set to create a release.
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is undefined)
endif
ifndef GPG_FINGERPRINT
	$(error GPG_FINGERPRINT is undefined)
endif
ifndef RELEASE_VERSION
	$(error RELEASE_VERSION is undefined)
endif

.PHONY: release
release: release-env-check all
release: ## Creates a GitHub release using goreleaser.
	GITHUB_TOKEN=$(GITHUB_TOKEN) GPG_FINGERPRINT=$(GPG_FINGERPRINT) bash -c 'go run github.com/goreleaser/goreleaser release --release-notes <(go run github.com/rcmachado/changelog show $(RELEASE_VERSION))'

.PHONY: install-tools
install-tools: ## Installs development tools required for creating a release.
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go get %

.PHONY: test-up
test-up: ## Starts testing tink-server instance in Docker container using docker-compose.
	docker-compose -f test/docker-compose.yml up -d

.PHONY: test-down
test-down: ## Tears down testing tink-server instance created by 'test-up'.
	docker-compose -f test/docker-compose.yml down

.PHONY: help
help: ## Prints help message.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
