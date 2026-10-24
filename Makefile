.PHONY: build run test lint clean migrate-up

APP_NAME := diagram-server
BUILD_DIR := ./bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

migrate-up:
	go run ./cmd/migrate
