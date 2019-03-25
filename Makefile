#
# Copyright (c) 2017-2018 Software AG, Darmstadt, Germany and/or Software AG USA
# Inc., Reston, VA, USA, and/or its subsidiaries and/or its affiliates
# and/or their licensors.
# Use, reproduction, transfer, publication or disclosure is prohibited except
# as specifically provided for in your License Agreement with Software AG.
#
#

GOARCH      ?= $(shell $(GO) env GOARCH)
GOOS        ?= $(shell $(GO) env GOOS)

PACKAGE      = softwareag.com
TESTPKGSDIR  = cmd/database
DATE        ?= $(shell date +%FT%T%z)
VERSION     ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
BIN          = $(CURDIR)/bin/$(GOOS)_$(GOARCH)
LOGPATH      = $(CURDIR)/logs
CURLOGPATH   = $(CURDIR)/logs
NETWORK     ?= emon:30011
TESTOUTPUT   = $(CURDIR)/test
MESSAGES     = $(CURDIR)/messages
EXECS        = $(BIN)/cmd/adabas-rest-client
OBJECTS      =  client/*.go cmd/database/*.go cmd/job/*.go cmd/filebrowser/*.go
SWAGGER_SPEC = $(CURDIR)/swagger/swagger.yaml
ENABLE_DEBUG = 2

export ADABAS_AIF_COMMAND

include $(CURDIR)/make/common.mk

generatemodels: $(SWAGGER_SPEC) $(CURDIR)/models

$(CURDIR)/models: $(GOSWAGGER) $(SWAGGER_SPEC) 
	if [ ! -d $(CURDIR)/models ]; then \
		GOPATH=$(CURDIR) $(GOSWAGGER) generate client -A AdabasAdmin -f $(SWAGGER_SPEC) -t $(CURDIR) -r copyright; \
	fi

.PHONY: clean
clean: cleanModules cleanVendor cleanModels cleanCommon ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf database.test
	@rm -rf bin pkg logs test
	@rm -rf test/tests.* test/coverage.*
	@rm -rf $(CURDIR)/cmd/adabas-rest-server/logs

cleanVendor: ; $(info $(M) cleaning vendor…)    @ ## Cleanup vendor
	@rm -rf $(CURDIR)/vendor

cleanModels: ; $(info $(M) cleaning models…)    @ ## Cleanup models
	@rm -rf $(CURDIR)/models
	@rm -rf $(CURDIR)/restapi/[!c]*
	@rm -rf $(CURDIR)/restapi/operations

$(BIN)/server: prepare generatemodels fmt lint lib $(EXECS)

startClient:
	$(GO) run $(GO_FLAGS) -ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.BuildDate=$(DATE)' \
	./$(EXECS:$(BIN)/%=%)
