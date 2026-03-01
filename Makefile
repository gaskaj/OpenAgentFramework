.PHONY: build test lint run clean fmt vet

BINARY := agentctl
PKG := ./...

build:
	go build -o bin/$(BINARY) ./cmd/agentctl

test:
	go test -race -count=1 $(PKG)

test-cover:
	go test -race -coverprofile=coverage.out $(PKG)
	go tool cover -html=coverage.out -o coverage.html

lint: vet
	@which golangci-lint > /dev/null 2>&1 || echo "golangci-lint not installed"
	golangci-lint run $(PKG)

vet:
	go vet $(PKG)

fmt:
	gofmt -s -w .

run: build
	./bin/$(BINARY) start --config configs/config.example.yaml

clean:
	rm -rf bin/ coverage.out coverage.html
