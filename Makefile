 .PHONY: kvdb
kvdb:
	go run ./cmd/kvdb

 .PHONY: client
client:
	go run ./cmd/client

 .PHONY: lint
lint:
	golangci-lint run

 .PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out

 .PHONY: coverage_html
coverage_html:
	go tool cover -html=coverage.out -o coverage.html