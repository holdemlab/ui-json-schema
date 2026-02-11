.PHONY: build run test test-cover lint bench fmt clean

APP_NAME := ui-json-schema
BUILD_DIR := ./bin
MAIN := ./cmd/server

## build: Compile the application
build:
	@echo "==> Building..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN)

## run: Build and run the application
run: build
	$(BUILD_DIR)/$(APP_NAME)

## test: Run all tests
test:
	go test -race -count=1 ./...

## test-cover: Run tests with coverage report
test-cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "==> To view HTML report: go tool cover -html=coverage.out -o coverage.html"

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## bench: Run benchmarks
bench:
	go test -bench=. -benchmem -benchtime=3s ./parser/

## fmt: Format code
fmt:
	gofmt -s -w .
	goimports -w .

## clean: Remove build artefacts
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

## help: Show this help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
