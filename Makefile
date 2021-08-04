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

GOARCH     ?= $(shell $(GO) env GOARCH)
GOOS       ?= $(shell $(GO) env GOOS)

PACKAGE     = softwareag.com
TESTPKGSDIR  = cmd/database
DATE       ?= $(shell date +%FT%T%z)
VERSION    ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
BIN         = $(CURDIR)/bin/$(GOOS)_$(GOARCH)
BINTOOLS    = $(CURDIR)/bin/tools/$(GOOS)_$(GOARCH)
LOGPATH     = $(CURDIR)/logs
CURLOGPATH  = $(CURDIR)/logs
NETWORK    ?= emon:30011
TESTOUTPUT  = $(CURDIR)/test
MESSAGES    = $(CURDIR)/messages
EXECS       = $(BIN)/cmd/adabas-restful-client
OBJECTS     =  cmd/*/*.go
SWAGGER_SPEC = $(CURDIR)/swagger/swagger.yaml
ENABLE_DEBUG = 2

include $(CURDIR)/make/common.mk

generatemodels: $(SWAGGER_SPEC) $(CURDIR)/models

execs: $(EXECS)

$(CURDIR)/models: $(GOSWAGGER) $(SWAGGER_SPEC) 
	if [ ! -d $(CURDIR)/models ]; then \
		$(GOSWAGGER) generate client -A AdabasAdmin -f $(SWAGGER_SPEC) -t $(CURDIR) -r copyright; \
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
