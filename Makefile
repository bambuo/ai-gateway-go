.PHONY: all build test vet fix lint clean run

GOFLAGS ?= -ldflags="-s -w"
BINARY  := gateway

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

clean:
	rm -f $(BINARY)

docker:
	docker build -t cc-gateway:latest .
