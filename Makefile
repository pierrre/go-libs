.DEFAULT_GOAL=noop
.DELETE_ON_ERROR:

.PHONY: noop
noop:

.PHONY: test
test:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out -o=coverage.txt
	cat coverage.txt
	go tool cover -html=coverage.out -o=coverage.html

.PHONY: lint
lint::
	$(MAKE) golangci-lint

GOLANGCI_LINT_VERSION=v1.43.0
GOLANGCI_LINT_DIR=$(shell go env GOPATH)/pkg/golangci-lint/$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT_BIN=$(GOLANGCI_LINT_DIR)/golangci-lint

$(GOLANGCI_LINT_BIN):
	GOBIN=$(GOLANGCI_LINT_DIR) go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: install-golangci-lint
install-golangci-lint: $(GOLANGCI_LINT_BIN)

GOLANGCI_LINT_RUN=$(GOLANGCI_LINT_BIN) -v run
.PHONY: golangci-lint
golangci-lint: install-golangci-lint
	$(GOLANGCI_LINT_RUN) --fix

.PHONY: golangci-lint-cache-clean
golangci-lint-cache-clean: install-golangci-lint
	$(GOLANGCI_LINT_BIN) cache clean

.PHONY: mod-update
mod-update:
	go get -v -u -d all
	$(MAKE) mod-tidy

.PHONY: mod-tidy
mod-tidy:
	go mod tidy -v

.PHONY: clean
clean::
	git clean -fdX
	go clean -cache -testcache
