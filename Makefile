# Makefile for the New Orleans Port Information Service
# See README.md for prerequisites and usage.

MODULE      := github.com/djdees/port-simulation
BIN_DIR     := bin
SERVER_BIN  := $(BIN_DIR)/server
CLIENT_BIN  := $(BIN_DIR)/client

GO          := go
GOFLAGS     :=

# Pinned tool versions
GOLANGCI_VERSION  := v2.15.0
VACUUM_VERSION    := v0.35.4
GOVULNCHECK_VERSION := v1.1.4
GOSEC_VERSION     := v2.30.0

GOLANGCI    := $(shell go env GOPATH)/bin/golangci-lint
VACUUM      := $(shell go env GOPATH)/bin/vacuum
GOVULNCHECK := $(shell go env GOPATH)/bin/govulncheck
GOSEC       := $(shell go env GOPATH)/bin/gosec

# Docker
IMAGE_SERVER := port-server:latest
IMAGE_CLIENT := port-client:latest
COMPOSE_FILE := docker/docker-compose.yml

.PHONY: help fmt vet lint test test-race build build-server build-client \
        run-server run-client openapi-check security check clean \
        docker-build-server docker-build-client compose-up compose-client \
        install-tools

## help: Show this help message
help:
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'

## fmt: Check formatting (gofmt)
fmt:
	@echo "==> Checking formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

## vet: Run go vet
vet:
	@echo "==> Running go vet..."
	$(GO) vet ./...

## lint: Run golangci-lint
lint: $(GOLANGCI)
	@echo "==> Running golangci-lint..."
	$(GOLANGCI) run ./...

## test: Run unit and API tests (no SPIRE required)
test:
	@echo "==> Running tests..."
	$(GO) test ./...

## test-race: Run tests with race detector
test-race:
	@echo "==> Running tests with race detector..."
	$(GO) test -race ./...

## build: Build server and client binaries
build: build-server build-client

## build-server: Build the server binary
build-server:
	@echo "==> Building server..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(SERVER_BIN) ./cmd/server

## build-client: Build the client binary
build-client:
	@echo "==> Building client..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(CLIENT_BIN) ./cmd/client

## run-server: Run server locally (SPIFFE off, loopback only)
run-server: build-server
	@echo "==> Starting server (SPIFFE-off mode, loopback only)"
	PORT_SPIFFE_ENABLED=false $(SERVER_BIN) --listen 127.0.0.1:8080

## run-client: Run client query (SPIFFE off, loopback only)
## Usage: make run-client ARGS="query arrivals"
run-client: build-client
	@echo "==> Running client (SPIFFE-off mode)"
	PORT_SPIFFE_ENABLED=false PORT_SERVER_URL=http://127.0.0.1:8080 $(CLIENT_BIN) $(ARGS)

## openapi-check: Validate the OpenAPI specification with Vacuum
openapi-check: $(VACUUM)
	@echo "==> Validating OpenAPI spec..."
	$(VACUUM) lint openapi/openapi.yaml

## security: Run govulncheck and gosec
security: $(GOVULNCHECK) $(GOSEC)
	@echo "==> Running govulncheck..."
	$(GOVULNCHECK) ./...
	@echo "==> Running gosec..."
	$(GOSEC) -quiet ./...

## check: Run all quality checks (fmt vet lint test test-race openapi-check security build)
check: fmt vet lint test test-race openapi-check security build
	@echo "==> All checks passed."

## clean: Remove built binaries and collector output
clean:
	@echo "==> Cleaning..."
	rm -rf $(BIN_DIR) collector-output

## docker-build-server: Build the server Docker image
docker-build-server:
	@echo "==> Building server Docker image..."
	docker build -f docker/Dockerfile.server -t $(IMAGE_SERVER) .

## docker-build-client: Build the client Docker image
docker-build-client:
	@echo "==> Building client Docker image..."
	docker build -f docker/Dockerfile.client -t $(IMAGE_CLIENT) .

## compose-up: Start the port server via Docker Compose
## Requires: SPIFFE_SOCKET, PORT_AUTH_POLICY_FILE, PORT_SERVER_SPIFFE_ID
compose-up:
	@echo "==> Starting port server via Compose..."
	docker compose -f $(COMPOSE_FILE) up -d server

## compose-client: Run a one-shot client command via Compose
## Usage: make compose-client ARGS="query arrivals"
compose-client:
	@echo "==> Running client via Compose..."
	docker compose -f $(COMPOSE_FILE) run --rm client $(ARGS)

## install-tools: Install pinned development tools
install-tools:
	@echo "==> Installing golangci-lint $(GOLANGCI_VERSION)..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	@echo "==> Installing vacuum $(VACUUM_VERSION)..."
	go install github.com/daveshanley/vacuum@$(VACUUM_VERSION)
	@echo "==> Installing govulncheck $(GOVULNCHECK_VERSION)..."
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	@echo "==> Installing gosec $(GOSEC_VERSION)..."
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	@echo "==> Tools installed."

# Tool prerequisites — fail with a helpful message if not installed.
$(GOLANGCI):
	@echo "golangci-lint not found. Run: make install-tools"
	@exit 1

$(VACUUM):
	@echo "vacuum not found. Run: make install-tools"
	@exit 1

$(GOVULNCHECK):
	@echo "govulncheck not found. Run: make install-tools"
	@exit 1

$(GOSEC):
	@echo "gosec not found. Run: make install-tools"
	@exit 1
