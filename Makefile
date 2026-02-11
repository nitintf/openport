VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -s -w \
	-X github.com/nitintf/openport/internal/version.Version=$(VERSION) \
	-X github.com/nitintf/openport/internal/version.Commit=$(COMMIT) \
	-X github.com/nitintf/openport/internal/version.Date=$(DATE)

PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: all build server op client clean test lint release tag

all: build

build: server op

server:
	go build -ldflags "$(LDFLAGS)" -o bin/openport-server ./cmd/server

op:
	go build -ldflags "$(LDFLAGS)" -o bin/op ./cmd/op

client:
	go build -ldflags "$(LDFLAGS)" -o bin/openport ./cmd/client

clean:
	rm -rf bin/ dist/

test:
	go test ./...

lint:
	golangci-lint run ./...

run-server:
	go run -ldflags "$(LDFLAGS)" ./cmd/server

run-op:
	go run -ldflags "$(LDFLAGS)" ./cmd/op -- 3000

# Create a version tag (usage: make tag BUMP=patch|minor|major)
BUMP ?= patch
tag:
	@./scripts/release.sh $(BUMP)

# Build release binaries for all platforms
release: clean
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o dist/op-$$os-$$arch ./cmd/op; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o dist/openport-server-$$os-$$arch ./cmd/server; \
	done
	@echo "Release binaries in dist/"
	@ls -lh dist/
