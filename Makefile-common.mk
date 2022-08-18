.DEFAULT_GOAL=noop
.DELETE_ON_ERROR:

.PHONY: noop
noop:

CI?=false

VERSION?=$(shell (git describe --tags --exact-match 2> /dev/null || git rev-parse HEAD) | sed "s/^v//")
.PHONY: version
version:
	@echo $(VERSION)

GO_BUILD_DIR=build
.PHONY: build
build:
ifneq ($(wildcard ./cmd/*),)
	mkdir -p $(GO_BUILD_DIR)
	go build -v -ldflags="-s -w -X main.version=$(VERSION)" -o $(GO_BUILD_DIR) ./cmd/...
endif

.PHONY: test
test:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out -o=coverage.txt
	cat coverage.txt
	go tool cover -html=coverage.out -o=coverage.html

.PHONY: generate
generate::
	go generate -v ./...

.PHONY: lint
lint:
	$(MAKE) golangci-lint

GOLANGCI_LINT_VERSION=v1.48.0
GOLANGCI_LINT_DIR=$(shell go env GOPATH)/pkg/golangci-lint/$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT_BIN=$(GOLANGCI_LINT_DIR)/golangci-lint

$(GOLANGCI_LINT_BIN):
	curl -vfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOLANGCI_LINT_DIR) $(GOLANGCI_LINT_VERSION)

.PHONY: install-golangci-lint
install-golangci-lint: $(GOLANGCI_LINT_BIN)

GOLANGCI_LINT_RUN=$(GOLANGCI_LINT_BIN) -v run
.PHONY: golangci-lint
golangci-lint: install-golangci-lint
ifeq ($(CI),true)
	$(GOLANGCI_LINT_RUN)
else
# Fix errors if possible.
	$(GOLANGCI_LINT_RUN) --fix
endif

.PHONY: golangci-lint-cache-clean
golangci-lint-cache-clean: install-golangci-lint
	$(GOLANGCI_LINT_BIN) cache clean

.PHONY: mod-update
mod-update:
	go get -v -u all
	$(MAKE) mod-tidy

.PHONY: mod-tidy
mod-tidy:
	go mod tidy -v

.PHONY: git-latest-release
git-latest-release:
	@git tag --list --sort=v:refname --format="%(refname:short) => %(creatordate:short)" | tail -n 1

.PHONY: clean
clean:
	git clean -fdX
	go clean -cache -testcache
	$(MAKE) golangci-lint-cache-clean

ifeq ($(CI),true)

CI_LOG_GROUP_START=@echo "::group::$(1)"
CI_LOG_GROUP_END=@echo "::endgroup::"

.PHONY: ci
ci:
	$(call CI_LOG_GROUP_START,build)
	$(MAKE) build
	$(call CI_LOG_GROUP_END)

	$(call CI_LOG_GROUP_START,test)
	$(MAKE) test
	$(call CI_LOG_GROUP_END)

	$(call CI_LOG_GROUP_START,lint)
	$(MAKE) lint
	$(call CI_LOG_GROUP_END)

endif # CI end
