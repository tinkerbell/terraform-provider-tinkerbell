# Build parameters
CGO_ENABLED=0
LD_FLAGS="-extldflags '-static'"

# Go parameters
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

.PHONY: all
all: build build-test test lint

.PHONY: download
download:
	$(GOMOD) download

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BIN_PATH) $(GOLANGCI_LINT_VERSION)

.PHONY: install-cc-test-reporter
install-cc-test-reporter:
	curl -L https://codeclimate.com/downloads/test-reporter/test-reporter-latest-linux-amd64 > $(BIN_PATH)/cc-test-reporter
	chmod +x $(BIN_PATH)/cc-test-reporter

.PHONY: install-ci
install-ci: install-golangci-lint install-cc-test-reporter

.PHONY: build
build:
	$(GOBUILD)

.PHONY: test
test: build-test
	TF_ACC=$(TF_ACC) TINKERBELL_GRPC_AUTHORITY=$(TINKERBELL_GRPC_AUTHORITY) TINKERBELL_CERT_URL=$(TINKERBELL_CERT_URL) $(GOTEST) -run $(GO_TESTS) $(GO_PACKAGES)

.PHONY: lint
lint: build build-test
	golangci-lint run --enable-all --disable=$(DISABLED_LINTERS) --max-same-issues=0 --max-issues-per-linter=0 --build-tags integration --timeout 10m --exclude-use-default=false $(GO_PACKAGES)

.PHONY: build-test
build-test:
	$(GOTEST) -run=nope $(GO_PACKAGES)

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(OUTPUT_FILE) || true
	rm -f $(OUTPUT_FILE).sig || true

.PHONY: update
update:
	$(GOGET) -u $(GO_PACKAGES)
	$(GOMOD) tidy

.PHONY: all-cover
all-cover: build build-test test-cover lint

.PHONY: test-cover
test-cover: build-test
	$(GOTEST) -run $(GO_TESTS) -coverprofile=$(PROFILEFILE) $(GO_PACKAGES)

.PHONY: cover-upload
cover-upload: codecov
	# Make codeclimate as command, as we need to run test-cover twice and make deduplicates that.
	# Go test results are cached anyway, so it's fine to run it multiple times.
	make codeclimate

.PHONY: codecov
codecov: PROFILEFILE=coverage.txt
codecov: SHELL=/bin/bash
codecov: test-cover
codecov:
	bash <(curl -s https://codecov.io/bash)

.PHONY: testacc
testacc: TF_ACC=true
testacc: test

.PHONY: release-env-check
release-env-check:
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

install-tools: download
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go get %
