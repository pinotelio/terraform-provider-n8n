.PHONY: build install test clean fmt lint

# Detect OS and architecture
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
	OS := darwin
endif
ifeq ($(UNAME_S),Linux)
	OS := linux
endif

ifeq ($(UNAME_M),x86_64)
	ARCH := amd64
endif
ifeq ($(UNAME_M),arm64)
	ARCH := arm64
endif
ifeq ($(UNAME_M),aarch64)
	ARCH := arm64
endif

PROVIDER_NAME := terraform-provider-n8n
VERSION := 0.1.0
INSTALL_PATH := ~/.terraform.d/plugins/registry.terraform.io/pinotelio/n8n/$(VERSION)/$(OS)_$(ARCH)

build:
	@echo "Building provider..."
	go build -o $(PROVIDER_NAME)

install: build
	@echo "Installing provider to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(PROVIDER_NAME) $(INSTALL_PATH)/
	@echo "Provider installed successfully!"

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(PROVIDER_NAME)
	@rm -rf dist/

fmt:
	@echo "Formatting code..."
	go fmt ./...
	terraform fmt -recursive ./examples/

lint:
	@echo "Running linters..."
	golangci-lint run

docs:
	@echo "Generating documentation..."
	go generate

.DEFAULT_GOAL := build

