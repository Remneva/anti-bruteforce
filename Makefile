BIN := "./bin/antifrod"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

.PHONY: build run test lint

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd

run: build
	$(BIN) -config ./configs/config.toml

version: build
	$(BIN) version

test:
	go test -race -count 90 -timeout=30s ./internal/app/...

lint:
	golangci-lint run ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.34.0

lint: install-lint-deps
	golangci-lint run ./...

lint-fix:
	gofmt -w ./..
	gci -w ./..

generate:
	mkdir -p internal/server/pb
	protoc --go_out=internal/server/pb  --go-grpc_out=internal/server/pb  api/*.proto

build-cli:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./internal/cli/client

up:
	docker-compose up -d --build