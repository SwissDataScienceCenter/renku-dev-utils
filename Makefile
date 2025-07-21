SHELL = bash
# Current git branch
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
# Current git commit hash
GIT_COMMIT_HASH := $(shell git show --no-patch --no-notes --pretty='%h' HEAD)
# Current git tag
GIT_TAG := $(shell git describe --tags --exact-match || "")
ifeq ($(GIT_TAG),)
VERSION := $(BRANCH).$(GIT_COMMIT_HASH)
else
VERSION := $(GIT_TAG)
endif
ifeq ($(shell git status --porcelain),)
DIRTY :=
else
VERSION := $(BRANCH).$(GIT_COMMIT_HASH)
DIRTY := "dev"
endif
LDFLAGS=--ldflags "-s -X github.com/SwissDataScienceCenter/renku-dev-utils/pkg/version.Version=$(VERSION) -X github.com/SwissDataScienceCenter/renku-dev-utils/pkg/version.VersionSuffix=$(DIRTY)"

.PHONY: all
all: help

.PHONY: vars
vars:  ## Show the Makefile vars
	@echo SHELL="'$(SHELL)'"
	@echo BRANCH="'$(BRANCH)'"
	@echo GIT_COMMIT_HASH="'$(GIT_COMMIT_HASH)'"
	@echo GIT_TAG="'$(GIT_TAG)'"
	@echo VERSION="'$(VERSION)'"
	@echo DIRTY="'$(DIRTY)'"

.PHONY: rdu
rdu: build/renku-dev-utils  ## Build and install renku-dev-utils
	mkdir -p `go env GOPATH`/bin/
	cp -av build/renku-dev-utils`go env GOEXE` `go env GOPATH`/bin/rdu`go env GOEXE`.new
	mv -v `go env GOPATH`/bin/rdu`go env GOEXE`.new `go env GOPATH`/bin/rdu`go env GOEXE`

.PHONY: build/renku-dev-utils
build/renku-dev-utils:
	go build -v -o build/ $(LDFLAGS)

# From the operator sdk Makefile
# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php
.PHONY: help
help:  ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: format
format:  ## Format source files
	gofmt -l -w .

.PHONY: check-format
check-format:  ## Check that sources are correctly formatted
	gofmt -d -s . && git diff --exit-code

.PHONY: check-vet
check-vet:  ## Check source files with `go vet`
	go vet ./...
