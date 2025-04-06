VERSION ?= dev
BUILD_TIME := $(shell date +%Y-%m-%d_%T)
GIT_HASH := $(shell git rev-parse --short HEAD)

BIN_DIR := ./bin
SERVER_BIN := $(BIN_DIR)/kvdb-server
CLIENT_BIN := $(BIN_DIR)/kvdb-client

DOCKER_IMAGE := kvdb-server
DOCKER_TAG ?= latest
DOCKERFILE := build/Dockerfile

CONFIG_FILE ?= config.yml
DATA_VOLUME := kvdb_data

LDFLAGS := -X main.version=$(VERSION) \
           -X main.buildTime=$(BUILD_TIME) \
           -X main.gitHash=$(GIT_HASH)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  Build & Run:"
	@echo "    all           - Build server and client"
	@echo "    build-server  - Build server only"
	@echo "    build-client  - Build client only"
	@echo "    server        - Build and run main server"
	@echo "    server-slave  - Build and run slave server"
	@echo "    client        - Run client (usage: make client user=USER pass=PASS)"
	@echo "  Docker:"
	@echo "    docker-build  - Build Docker image"
	@echo "    docker-run    - Build and run container"
	@echo "    docker-stop   - Stop and remove container"
	@echo "    docker-logs   - Show container logs"
	@echo "    docker-shell  - Enter container shell"
	@echo "  Maintenance:"
	@echo "    clean         - Remove binaries"
	@echo "    version       - Show version info"
	@echo "    lint          - Run linter"
	@echo "    coverage      - Generate test coverage"
	@echo "    help          - Show this help message"

.PHONY: build-server
build-server:
	@mkdir -p $(BIN_DIR)
	@echo "Building server (version: $(VERSION))..."
	@go build -ldflags "$(LDFLAGS)" -o $(SERVER_BIN) ./cmd/server

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

 .PHONY: lint
lint:
	@golangci-lint run

 .PHONY: coverage
coverage:
	@go test ./... -coverprofile=coverage.out

 .PHONY: coverage_html
coverage_html:
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: docker-build
docker-build:
	@echo "Building $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@docker build \
		-f $(DOCKERFILE) \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME="$(BUILD_TIME)" \
		--build-arg GIT_HASH=$(GIT_HASH) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		.

.PHONY: docker-run
docker-run: docker-build
	@echo "Start container $(DOCKER_IMAGE)"
	@docker run -d \
		--name $(DOCKER_IMAGE) \
		-p 3223:3223 \
		-v $(DATA_VOLUME):/data \
		-v $(PWD)/$(CONFIG_FILE):/kvdb/etc/config.yml \
		-e KVDB_DATA_DIR=${KVDB_DATA_DIR} \
		-e KVDB_LOG_DIR=${KVDB_LOG_DIR} \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-stop
docker-stop: 
	@echo "Stop container"
	@docker stop $(DOCKER_IMAGE) 2>/dev/null || true
	@docker rm $(DOCKER_IMAGE) 2>/dev/null || true

.PHONY: docker-logs
docker-logs:
	@docker logs -f $(DOCKER_IMAGE)

.PHONY: docker-shell
docker-shell:
	@docker exec -it $(DOCKER_IMAGE) sh
