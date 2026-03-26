.PHONY: all build build-local install test test-unit test-e2e test-coverage lint tidy clean deps version tag release-check

BINARY_NAME := yamlspec
DIST_DIR    := dist
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS     := -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

GOOS_ARCH := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

all: build-local

build-local:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

build:
	@mkdir -p $(DIST_DIR)
	@for platform in $(GOOS_ARCH); do \
		GOOS=$${platform%%/*} GOARCH=$${platform##*/} \
		go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$${platform%%/*}-$${platform##*/} . ; \
		echo "  Built: $${platform}"; \
	done

install: build-local
	sudo mv $(BINARY_NAME) /usr/local/bin/

test:
	go test ./... -v -race -timeout 120s

test-unit:
	go test -v -short ./services/...

test-e2e: build-local
	go test -v ./integration/... -timeout 120s

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run --timeout=5m

tidy:
	go mod tidy

clean:
	rm -rf $(DIST_DIR) $(BINARY_NAME) coverage.out coverage.html

deps:
	go mod tidy
	go mod download

version:
	@echo "Version: $(VERSION) | Build: $(BUILD_TIME) | Commit: $(GIT_COMMIT)"

tag:
	@if [ "$(V)" = "" ]; then echo "Usage: make tag V=v1.0.0"; exit 1; fi
	git tag -a $(V) -m "Release $(V)"

release-check: build test
	@echo "Ready for release $(VERSION)"
