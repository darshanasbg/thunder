# Constants.
VERSION_FILE=version.txt
BINARY_NAME=thunder
REPOSITORY_DIR=cmd/server/repository
OUTPUT_DIR=target
BUILD_DIR=$(OUTPUT_DIR)/.build

# Variable constants.
VERSION=$(shell cat $(VERSION_FILE))
# ZIP_FILE_NAME=${BINARY_NAME_PREFIX}-$(VERSION)
PRODUCT_FOLDER=$(BINARY_NAME)-$(VERSION)

# Default target.
all: clean build test

lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# Clean up build artifacts.
clean:
	rm -rf $(OUTPUT_DIR)

# Build project and package it.
build: _build _package

# Build the Go project.
_build:
	mkdir -p $(BUILD_DIR) && \
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Package the binary and repository directory into a zip file.
_package:
	mkdir -p $(OUTPUT_DIR)/$(PRODUCT_FOLDER) && \
	cp $(BUILD_DIR)/$(BINARY_NAME) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp -r $(REPOSITORY_DIR) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp $(VERSION_FILE) $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp -r scripts $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cp -r dbscripts $(OUTPUT_DIR)/$(PRODUCT_FOLDER)/ && \
	cd $(OUTPUT_DIR) && zip -r $(PRODUCT_FOLDER).zip $(PRODUCT_FOLDER) && \
	rm -rf $(PRODUCT_FOLDER) && \
	rm -rf $(BUILD_DIR)

# Run all tests.
test: _integration-test

# Run integration tests.
_integration-test:
	@echo "Running integration tests..."
	@go run ./tests/integration/run_tests.go

help:
	@echo "Makefile targets:"
	@echo "  all          - Clean, build, and test the project."
	@echo "  clean        - Remove build artifacts."
	@echo "  build        - Build the Go project."
	@echo "  test         - Run all tests."
	@echo "  help         - Show the help message."

.PHONY: all clean lint build _build _package test _integration-test help
