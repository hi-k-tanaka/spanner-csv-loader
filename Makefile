export PATH := $(PWD)/_tools/bin:$(PATH)

COMMAND_NAME ?= spannercsvimporter

TESTPKGS ?= $(shell go list ./src/... | grep -v -e mock -e test/e2e)
COVERPKGS ?= $(shell go list ./src/... | grep -v -e mock -e model -e test/)
FMTPKGS = $(subst $(REPOSITORY)/,,$(shell go list ./... | grep -v mock))
VETPKGS = $(shell go list ./... | grep -v mock)
LINTPKGS = $(shell go list ./... | grep -v mock)

GOTEST ?= go test

all: dep build

## environment

.PHONY: env
env:  ## show environment variables in make
	@$(foreach var,$(EXPORT_ENV),echo $(var)=`echo $($(var))`;)

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
	go build -o spannercsvimporter src/main.go
	#./spannercsvimporter

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

.PHONY: test
test:
	$(GOTEST) -v -race -p=1 -run=$(TARGET) $(TESTPKGS)

#
# cover
#

cover:  ## take unit test coverage
	$(GOTEST) -v -race -p=1 -run=$(TARGET) -covermode=atomic -coverprofile=$@.out -coverpkg=$(REPOSITORY)/... $(COVERPKGS)

codecov: SHELL=/usr/bin/env bash
codecov:  ## send coverage result to codecov.io
	bash <(curl -s https://codecov.io/bash)

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
	golint src/... | grep -v 'should have comment or be unexported' | grep -v 'comment on exported function'

#
# reviewdog
#

reviewdog:  ## run reviewdog
	reviewdog -diff="git diff master"

reviewdog/ci:  ## run reviewdog on ci
	reviewdog -ci="circle-ci"

#
# makefile
#

help:  ## show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[\/a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

