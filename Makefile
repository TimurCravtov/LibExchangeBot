APP_NAME := libexchangebot
BIN_DIR := bin

.PHONY: all build run test vet fmt lint tidy clean

all: fmt vet test build

build:
	go build -trimpath -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

run:
	go run ./cmd/$(APP_NAME)

test:
	go test -v -race ./...

vet:
	go vet ./...

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR) coverage.*
