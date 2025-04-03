VERSION ?= dev
BUILD_TIME := $(shell date +%Y-%m-%d_%T)
GIT_HASH := $(shell git rev-parse --short HEAD)

BIN_DIR := bin
SERVER_BIN := $(BIN_DIR)/kvdb
CLIENT_BIN := $(BIN_DIR)/kvdb-client

LDFLAGS := -X main.version=$(VERSION) \
           -X main.buildTime=$(BUILD_TIME) \
           -X main.gitHash=$(GIT_HASH)

.PHONY: build-server
build-server:
	@mkdir -p $(BIN_DIR)
	@echo "Building server (version: $(VERSION))..."
	@go build -ldflags "$(LDFLAGS)" -o $(SERVER_BIN) ./cmd/kvdb

.PHONY: build-client
build-client:
	@mkdir -p $(BIN_DIR)
	@echo "Building client (version: $(VERSION))..."
	@go build -ldflags "$(LDFLAGS)" -o $(CLIENT_BIN) ./cmd/client

.PHONY: server
server: build-server
	@echo "Starting server..."
	@$(SERVER_BIN) run --config config.yml

.PHONY: server-slave
server-slave: build-server
	@echo "Starting slave server..."
	@$(SERVER_BIN) run --config config.slave.yml

.PHONY: client
client: build-client
	@echo "Starting client..."
	@$(CLIENT_BIN) run \
		--username $(user) \
		--password $(pass)

.PHONY: clean
clean:
	@echo "Cleaning binaries..."
	@rm -rf $(BIN_DIR)

.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git hash: $(GIT_HASH)"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Build server and client"
	@echo "  build-server  - Build server only"
	@echo "  build-client  - Build client only"
	@echo "  server        - Build and run main server"
	@echo "  server-slave  - Build and run slave server"
	@echo "  client        - Run client (usage: make client user=USER pass=PASS)"
	@echo "  clean         - Remove binaries"
	@echo "  version       - Show version info"
	@echo "  help          - Show this help message"

 .PHONY: lint
lint:
	golangci-lint run

 .PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out

 .PHONY: coverage_html
coverage_html:
	go tool cover -html=coverage.out -o coverage.html