#
# Copyright © 2018 Software AG, Darmstadt, Germany and/or its licensors
#
# SPDX-License-Identifier: Apache-2.0
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#

GOARCH    ?= $(shell $(GO) env GOARCH)
GOOS      ?= $(shell $(GO) env GOOS)
PACKAGE    = softwareag.com
DATE      ?= $(shell date +%FT%T%z)
VERSION   ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
TMPPATH    ?= /tmp
GOPATH      = $(TMPPATH)/tmp_adabas-go-client.$(shell id -u)
ADAPATH    = $(CURDIR)/../AdabasAccess
BIN        = $(CURDIR)/bin
LOGPATH    = $(CURDIR)/logs
CURLOGPATH = $(CURDIR)/logs
NETWORK   ?= localhost:50001
TESTOUTPUT = $(CURDIR)/test
EXECS      = cmd/client
LIBS       = slib/adaapi
BASESRC    = $(CURDIR)/src/$(PACKAGE)
BASE       = $(GOPATH)/src/$(PACKAGE)
PKGS       = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH):$(ADAPATH) $(GO) list ./... | grep -v "^$(PACKAGE)/vendor/"))
TESTPKGS   = $(shell env GOPATH=$(GOPATH):$(ADAPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
GO_FLAGS   = $(if $(debug),"-x",)

export GOPATH

GO      = go
GODOC   = godoc
GOFMT   = gofmt
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

export GOOS GOARCH
GOEXE    ?= $(shell GOOS=$(GOOS) $(GO) env GOEXE)
export GOEXE

.PHONY: all
all: prepare fmt lint vendor generatemodels $(EXECS)

lib: all $(LIBS)

execs: $(EXECS)

gospecs:
	@echo "Build architecture ${GOOS}_${GOARCH} extension ${GOEXE}"


prepare: $(LOGPATH) $(CURLOGPATH) $(BIN)
	@echo "Build architecture ${GOOS}_${GOARCH} network=${NETWORK} GOFLAGS=$(GO_FLAGS)"

$(LIBS): | $(BASE) ; $(info $(M) building libraries…) @ ## Build program binary
	$Q cd $(BASE) && GOPATH=$(GOPATH) \
	    $(GO) build $(GO_FLAGS) \
		-tags release -buildmode=c-shared \
		-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
		-o $(BIN)/$@ $@.go

$(EXECS): | $(BASE) ; $(info $(M) building executable…) @ ## Build program binary
	$Q cd $(BASE) && \
	echo $(BIN)/$(GOOS)_$(GOARCH)/$@$(GOEXE) && GOPATH=$(GOPATH):$(ADAPATH) \
	    $(GO) build $(GO_FLAGS) \
		-tags release \
		-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
		-o $(BIN)/$(GOOS)_$(GOARCH)/$@$(GOEXE) $@.go

$(BASE): ; $(info $(M) setting GOPATH…)
	@mkdir -p $(dir $@)
	ln -sf $(CURDIR)/src/$(PACKAGE) $@

# Tools

$(LOGPATH):
	@mkdir -p $@

$(CURLOGPATH):
	@mkdir -p $@

$(BIN):
	@mkdir -p $@/$(GOOS)_$(GOARCH)
$(BIN)/%: $(BIN) | $(BASE) ; $(info $(M) building $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GOPATH=$$tmp go get $(REPOSITORY) && cp -r $$tmp/bin/* $(BIN)/.) || ret=$$?; \
		rm -rf $$tmp ; exit $$ret

GODEP = $(BIN)/dep
$(BIN)/dep: REPOSITORY=github.com/golang/dep/cmd/dep

GOLINT = $(BIN)/golint
$(BIN)/golint: REPOSITORY=golang.org/x/lint/golint

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: REPOSITORY=github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov: REPOSITORY=github.com/axw/gocov/...

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml: REPOSITORY=github.com/AlekSi/gocov-xml

GO2XUNIT = $(BIN)/go2xunit
$(BIN)/go2xunit: REPOSITORY=github.com/tebeka/go2xunit

# Tests
$(TESTOUTPUT):
	mkdir $(TESTOUTPUT)

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test-xml check test tests
test-bench:   ARGS=-run=__absolutelynothing__ -bench=. ## Run benchmarks
test-short:   ARGS=-short        ## Run only short tests
test-verbose: ARGS=-v            ## Run tests in verbose mode with coverage reporting
test-race:    ARGS=-race         ## Run tests with race detector
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: fmt lint vendor | $(BASE) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
	$Q cd $(BASE) && $(GO) test -timeout $(TIMEOUT)s $(ARGS) $(TESTPKGS)

test-xml: prepare fmt lint vendor generatemodels $(TESTOUTPUT) | $(BASE) $(GO2XUNIT) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests with xUnit output
	$Q cd $(BASE) && \
	echo $(BIN)/$(GOOS)_$(GOARCH)/$@$(GOEXE) && GOPATH=$(GOPATH):$(ADAPATH) \
	    $(GO) test -timeout 20s -v $(GO_FLAGS) $(TESTPKGS) | tee $(TESTOUTPUT)/tests.output
#	$Q cd $(BASE) && 2>&1 LOGPATH=$(LOGPATH) GOPATH=$(GOPATH) \
#	                      ENABLE_DEBUG=0 NETWORK=$(NETWORK) \
#	                      $(GO) test -timeout 20s -v $(GO_FLAGS) $(TESTPKGS) | tee $(TESTOUTPUT)/tests.output
	$(GO2XUNIT) -input $(TESTOUTPUT)/tests.output -output $(TESTOUTPUT)/tests.xml

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt lint vendor test-coverage-tools | $(BASE) ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)/coverage
	$Q cd $(BASE) && for pkg in $(TESTPKGS); do \
		$(GO) test \
			-coverpkg=$$($(GO) list -f '{{ join .Deps "\n" }}' $$pkg | \
					grep '^$(PACKAGE)/' | grep -v '^$(PACKAGE)/vendor/' | \
					tr '\n' ',')$$pkg \
			-covermode=$(COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$Q $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: vendor | $(BASE) $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
		test -z "$$(GOPATH=$(GOPATH):$(ADAPATH) $(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	 done ; exit $$ret

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Dependency management

.PHONY: vendor
vendor: Gopkg.toml Gopkg.lock | $(BASE) $(GODEP) ; $(info $(M) retrieving dependencies…)
	$Q cd $(BASE) && GOPATH=$(GOPATH):$(ADAPATH) $(GODEP) ensure -v
#	@ln -nsf . vendor/src
	@touch $@
.PHONY: vendor-install
vendor-install: vendor | $(BASE) $(GODEP) ; $(info $(M) vendor update…)
ifeq "$(origin PKG)" "command line"
	$(info $(M) installing $(PKG) dependency…)
	$Q cd $(BASE) && $(GODEP) ensure $(PKG)
else
	$(info $(M) installing all dependencies…)
	$Q cd $(BASE) && $(GODEP) ensure 
endif
	@ln -nsf . vendor/src
	@touch vendor
.PHONY: vendor-update
vendor-update: vendor | $(BASE) $(GODEP) ; $(info $(M) vendor update…)
ifeq "$(origin PKG)" "command line"
	$(info $(M) updating $(PKG) dependency…)
	$Q cd $(BASE) && $(GODEP) ensure -update $(PKG)
else
	$(info $(M) updating all dependencies…)
	$Q cd $(BASE) && $(GODEP) ensure -update
endif

# Misc

.PHONY: clean
clean: cleanModels cleanVendor; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(GOPATH)
	@rm -rf bin pkg logs test
	@rm -rf test/tests.* test/coverage.*
	@rm -rf src/softwareag.com/cmd/adabas-admin-server/logs
	@rm -rf src/softwareag.com/logs


cleanVendor: ; $(info $(M) cleaning vendor…)	@ ## Cleanup vendor
	@rm -rf src/softwareag.com/vendor vendor

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)

checkstyle:
	mkdir -p $(GOPATH)/checkstyle; cd $(GOPATH)/checkstyle && export GOPATH=`pwd` && \
	 go get github.com/qiniu/checkstyle/gocheckstyle
	cd $(CURDIR); $(GOPATH)/checkstyle/bin/gocheckstyle -reporter=xml -config=go_style src 2>gostyle.xml

$(BASE)/vendor: vendor-install

generatemodels: $(CURDIR)/swagger/aif-swagger.yaml ; $(info $(M) generate models …) @ ## Generate model
	if [ ! -d $(BASESRC)/models ]; then \
	  GOPATH=$(GOPATH) go get -u github.com/go-swagger/go-swagger/cmd/swagger; \
	    if [ -r ../AdabasRestServer/swagger/swagger.yaml ]; then \
	     grep -v application/xml ../AdabasRestServer/swagger/swagger.yaml >$(CURDIR)/swagger/aif-swagger.yaml; \
	    fi; \
	  GOPATH=$(GOPATH) $(GOPATH)/bin/swagger generate client -A AdabasAdmin -f $(CURDIR)/swagger/aif-swagger.yaml \
	 -t $(BASE) -r copyright; \
	fi

cleanModels: ; $(info $(M) cleaning models…)	@ ## Cleanup vendor
	@rm -rf $(BASESRC)/models
	@rm -rf $(BASESRC)/client

generate: cleanModels generatemodels  # $(BASESRC)/models

