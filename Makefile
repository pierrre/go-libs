include Makefile-common.mk

ALL_COMMAND=cat projects.txt | xargs -I {} $(1)
ALL_RUN=$(call ALL_COMMAND,sh -c 'echo {} && cd {} && $(1)')
.PHONY: all-run
all-run:
	$(eval COMMAND?=ls)
	$(call ALL_RUN,$(COMMAND))

GIT_REPOSITORY_PATTERN=git@github.com:pierrre/{}.git
.PHONY: all-git-clone
all-git-clone:
	$(call ALL_COMMAND,sh -c "(ls ../{} > /dev/null 2>&1 || git -C .. clone $(GIT_REPOSITORY_PATTERN))")

.PHONY: all-copy-common
all-copy-common:
	$(call ALL_COMMAND,cp -r Makefile-common.mk LICENSE CODEOWNERS .gitignore .github .golangci.yml ../{})

.PHONY: all-build
all-build:
	$(call ALL_RUN,make build)

.PHONY: all-test
all-test:
	$(call ALL_RUN,make test)

.PHONY: all-generate
all-generate:
	$(call ALL_RUN,make generate)

.PHONY: all-lint
all-lint:
	$(call ALL_RUN,make lint)

.PHONY: all-golangci-lint
all-golangci-lint:
	$(call ALL_RUN,make golangci-lint)

.PHONY: all-lint-rules
all-lint-rules:
	$(call ALL_RUN,make lint-rules)

.PHONY: all-mod-update
all-mod-update: all-copy-common
	$(call ALL_RUN,make mod-update)

.PHONY: all-mod-update-pierrre
all-mod-update-pierrre: all-copy-common
	$(call ALL_RUN,make mod-update-pierrre)

.PHONY: all-mod-tidy
all-mod-tidy:
	$(call ALL_RUN,make mod-tidy)

.PHONY: all-clean
all-clean:
	$(call ALL_RUN,make clean)
