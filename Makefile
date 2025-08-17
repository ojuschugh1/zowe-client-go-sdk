.PHONY: build test clean run-example install deps

# Build the SDK
build:
	go build -o bin/zowe-sdk ./...

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run the example
run-example:
	go run examples/profile_management.go

# Run job management example
run-job-example:
	go run examples/job_management.go

# Install the SDK
install:
	go install ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	godoc -http=:6060

# Build for different platforms
build-all: build-linux build-windows build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/zowe-sdk-linux-amd64 ./...

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/zowe-sdk-windows-amd64.exe ./...

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/zowe-sdk-darwin-amd64 ./...

# Development helpers
dev-setup: deps
	mkdir -p bin

# Run with sample config
run-with-sample:
	cp examples/sample_zowe_config.json ~/.zowe/zowe.config.json
	go run examples/profile_management.go 

# Run dataset management example
run-dataset-example:
	@echo "Running Dataset Management Example..."
	@go run examples/dataset_management.go

# Run all examples
run-all-examples: run-example run-job-example run-dataset-example 