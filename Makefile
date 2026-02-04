GOLANGCI_LINT ?= golangci-lint
BINARY := k8s-manifest-tail
BUILD_DIR := build

VERSION := 0.0.1
IMAGE ?= ghcr.io/grafana/$(BINARY)
PLATFORMS ?= linux/arm64,linux/amd64

.PHONY: all build clean lint lint-go lint-zizmor test help

##@ Build

all: build ## Build the project (default target)

build: $(BUILD_DIR)/$(BINARY) ## Compile the binary into the build directory

$(BUILD_DIR)/$(BINARY): $(shell find . -name '*.go') go.mod go.sum
	@mkdir -p $(BUILD_DIR)
	go build -o $@ .

build-image: $(BUILD_DIR)/image-built-$(VERSION)
$(BUILD_DIR)/image-built-$(VERSION): operations/container/Dockerfile $(shell find . -name '*.go') go.mod go.sum
	docker buildx build --platform $(PLATFORMS) --tag $(IMAGE):$(VERSION) --file operations/container/Dockerfile .
	mkdir -p $(BUILD_DIR) && touch $(BUILD_DIR)/image-built-${VERSION}

push-image: build/image-built-$(VERSION)
	docker push $(IMAGE):$(VERSION)

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

##@ Test

lint: lint-go lint-zizmor ## Run all lint checks

lint-go: ## Run golangci-lint against the codebase
	$(GOLANGCI_LINT) run ./...

lint-zizmor: ## Statically analyze GitHub Action workflows
	@if command -v zizmor >/dev/null 2>&1; then \
		zizmor .; \
	else \
		docker run --rm -v $(shell pwd):/src --workdir /src ghcr.io/zizmorcore/zizmor:latest .; \
	fi

test: ## Execute unit and integration tests
	go test ./...


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

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
