.PHONY: all build test vet fix lint clean run web-install web-dev web-build admin

GOFLAGS ?= -ldflags="-s -w"
BINARY  := gateway
WEB_DIR := web

all: vet test build

build:
	go build $(GOFLAGS) -o $(BINARY) ./cmd/gateway

test:
	go test -v -race -count=1 ./...

vet:
	go vet ./...

fix:
	go fix ./...
	go mod tidy

lint:
	golangci-lint run ./...

run: build
	./$(BINARY) serve config.yaml

admin: web-build
	go build $(GOFLAGS) -o $(BINARY) ./cmd/gateway
	./$(BINARY) admin --db ./data/admin.db --static ./web/dist

web-install:
	cd $(WEB_DIR) && bun install

web-dev:
	cd $(WEB_DIR) && bunx vite

web-build:
	cd $(WEB_DIR) && bunx vue-tsc -b && bunx vite build

clean:
	rm -f $(BINARY)
	rm -rf $(WEB_DIR)/dist

docker:
	docker build -t cc-gateway:latest .
