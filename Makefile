 .PHONY: kvdb
kvdb:
	go run ./cmd/kvdb

 .PHONY: build_client
build_client:
	go build ./cmd/client 

 .PHONY: client
client:
	$(MAKE) build_client
	./client --username $(user) --password $(pass)

 .PHONY: lint
lint:
	golangci-lint run

 .PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out

 .PHONY: coverage_html
coverage_html:
	go tool cover -html=coverage.out -o coverage.html