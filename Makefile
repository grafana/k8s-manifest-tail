BINARY := k8s-manifest-tail
BUILD_DIR := build

.PHONY: all build clean test

all: build

build: $(BUILD_DIR)/$(BINARY)

$(BUILD_DIR)/$(BINARY): $(shell find . -name '*.go') go.mod go.sum
	@mkdir -p $(BUILD_DIR)
	go build -o $@ .

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...
