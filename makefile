.PHONY: help build run test test-unit test-integration clean install-deps fmt lint

help:
    @echo "Available commands:"
    @echo "  make install-deps       - Install dependencies"
    @echo "  make build              - Build the application"
    @echo "  make run                - Run the application"
    @echo "  make test               - Run all tests"
    @echo "  make test-unit          - Run unit tests"
    @echo "  make test-integration   - Run integration tests"
    @echo "  make test-coverage      - Run tests with coverage report"
    @echo "  make fmt                - Format code"
    @echo "  make lint               - Run linter"
    @echo "  make clean              - Clean build artifacts"

install-deps:
    go mod download
    go mod tidy

build:
    go build -o bin/digicert ./cmd/main.go

run:
    go run ./cmd/main.go

test:
    go test ./... -v

test-unit:
    go test ./internal/handler -v

test-integration:
    go test ./test -v

test-coverage:
    go test ./... -v -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

fmt:
    go fmt ./...
    goimports -w .

lint:
    golangci-lint run ./...

clean:
    rm -f bin/digicert
    rm -f coverage.out coverage.html
    go clean

db-up:
    docker run --name digicert-db -e POSTGRES_PASSWORD=password -e POSTGRES_DB=digicert -p 5432:5432 -d postgres:15

db-down:
    docker stop digicert-db
    docker rm digicert-db