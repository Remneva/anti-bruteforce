hello:
	echo "Hello world"

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/main.go

run: build
	$(BIN) -config ./configs/config.toml

test:
	go test -race -count 93 -timeout=30s ./internal/app/...

lint:
	golangci-lint run ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.30.0

lint: install-lint-deps
	golangci-lint run ./...

lint-fix:
	gofmt -w ./..
	gci -w ./..

generate:
	mkdir -p internal/server/pb
	protoc --go_out=internal/server/pb  --go-grpc_out=internal/server/pb  api/*.proto

BIN := "./bin/cli"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

bd:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./internal/cli/client

bd2:
	go build -v -o ./bin -ldflags "$(LDFLAGS)" ./internal/cli/server.go