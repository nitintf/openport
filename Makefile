BINARY_SERVER=openport-server
BINARY_CLIENT=openport
BINARY_OP=op

.PHONY: all build server client op clean test lint

all: build

build: server op

server:
	go build -o bin/$(BINARY_SERVER) ./cmd/server

client:
	go build -o bin/$(BINARY_CLIENT) ./cmd/client

op:
	go build -o bin/$(BINARY_OP) ./cmd/op

clean:
	rm -rf bin/

test:
	go test ./...

lint:
	golangci-lint run ./...

run-server:
	go run ./cmd/server

run-client:
	go run ./cmd/client

run-op:
	go run ./cmd/op -- 3000
