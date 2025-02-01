 .PHONY: kvdb
kvdb:
	go run ./cmd/kvdb

 .PHONY: client
client:
	go run ./cmd/client

 .PHONY: lint
lint:
	golangci-lint run