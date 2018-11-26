export PATH := $(PWD)/_tools/bin:$(PATH)

REPOSITORY := github.com/hi-k-tanaka/spanner-csv-loader
TESTPKGS ?= $(shell go list ./...)
COVERPKGS ?= $(shell go list ./...)
FMTPKGS = $(subst $(REPOSITORY)/,,$(shell go list ./...))
VETPKGS = $(shell go list ./...)
LINTPKGS = $(shell go list ./...)

GOTEST ?= go test

#
# dep
#

dep:
	dep ensure -v

dep/ci:
	dep ensure --vendor-only

#
# build
#
build:
	go build -o spanner-csv-loader main.go

#
# tools
#

tools: ## install executable tools
	go get -u github.com/twitchtv/retool
	retool sync

#
# test
#

TARGET ?= .

test:
	$(GOTEST) -v -race -p=1 -run=$(TARGET) $(TESTPKGS)

#
# cover
#

cover:  ## take unit test coverage
	$(GOTEST) -v -race -p=1 -run=$(TARGET) -covermode=atomic -coverprofile=$@.out -coverpkg=$(REPOSITORY)/... $(COVERPKGS)

#
# static check
#

lint: fmt vet staticcheck errcheck  ## lint all

fmt:  ## run goimports and gofmt
	goimports -l $(FMTPKGS) | grep -E '.'; test $$? -eq 1
	gofmt -l $(FMTPKGS) | grep -E '.'; test $$? -eq 1

vet:  ## run go vet
	go vet $(VETPKGS)

staticcheck:  ## run staticcheck
	staticcheck $(LINTPKGS)

errcheck:  ## run errcheck
	errcheck $(LINTPKGS)

golint:  ## run golint
	golint ./... | grep -v 'should have comment or be unexported' | grep -v 'comment on exported function'

#
# makefile
#

help:  ## show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[\/a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

