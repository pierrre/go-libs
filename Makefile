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
