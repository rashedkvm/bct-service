##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
GOOS=$(shell go env GOOS)
else
GOBIN=$(shell go env GOBIN)
GOOS=$(shell go env GOOS)
endif

# Suppress kapp prompts with KAPP_ARGS="--yes"
KAPP_ARGS ?= "--yes=false"

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Tools
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

YTT ?= $(LOCALBIN)/ytt
KCTRL ?= $(LOCALBIN)/kctrl
KAPP ?= $(LOCALBIN)/kapp
IMGPKG ?= $(LOCALBIN)/imgpkg
KO ?= $(LOCALBIN)/ko

KO_VERSION ?= 0.16.0

.PHONY: carvel-tools
carvel-tools: $(LOCALBIN) ## Downloads Carvel CLI tools locally
	if [[ ! -f $(YTT) ]]; then \
		curl -L https://carvel.dev/install.sh | K14SIO_INSTALL_BIN_DIR=$(LOCALBIN) bash; \
	fi

.PHONY: ko-setup
ko-setup: $(KO) ## Setup for ko binary
$(KO): $(LOCALBIN)
	@if [ ! -f $(KO) ]; then \
		echo curl -sSfL "https://github.com/ko-build/ko/releases/download/v$(KO_VERSION)/ko_$(KO_VERSION)_$(GOOS)_x86_64.tar.gz"; \
		curl -sSfL "https://github.com/ko-build/ko/releases/download/v$(KO_VERSION)/ko_$(KO_VERSION)_$(GOOS)_x86_64.tar.gz" > $(LOCALBIN)/ko.tar.gz; \
		tar xzf $(LOCALBIN)/ko.tar.gz -C $(LOCALBIN)/; \
		chmod +x $(LOCALBIN)/ko; \
	fi;

.PHONY: tools
tools: carvel-tools ko-setup ## Setup tools for local build & development

##@ Development
.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy -v

.PHONY: build
build: fmt vet tidy ## Builds the binary under bin folder
	mkdir -p "bin"
	go build -o bin/bct-service cmd/main.go

.PHONY: test
test: fmt vet ## Run unit tests only.
	go test ./... -short -coverprofile cover.out

.PHONY: deploy
deploy: test ## Deploy bct-service to the K8s cluster specified in ~/.kube/config.
	$(KAPP) deploy -a bct-service -n kube-system -f <($(KO) resolve -f <( $(YTT) -f app-config)) $(KAPP_ARGS)

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KAPP) delete -a bct-service -n kube-system $(KAPP_ARGS)
