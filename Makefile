 .PHONY: kvdb
kvdb:
	go build ./cmd/kvdb && ./kvdb --config config.yml

 .PHONY: kvdb-slave
kvdb-slave:
	go build ./cmd/kvdb && ./kvdb --config config.slave.yml

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