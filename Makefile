.PHONY: build test clean run all docker docker-run test-docker

# Binary name
BINARY_NAME=fakessh

# Main directories
CMD_DIR=./cmd/fakessh
BUILD_DIR=./build

# Build variables
GO=go
GOFLAGS=-ldflags="-s -w"

# Docker variables
DOCKER_IMAGE=fakessh

all: clean build test

build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

test:
	@echo "Running tests..."
	$(GO) test -v ./...

test-docker:
	@echo "Running Docker integration tests..."
	$(GO) test -v -tags=docker ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

run: build
	@echo "Starting fake SSH server..."
	@$(BUILD_DIR)/$(BINARY_NAME) --generate-key=false

# Run with custom parameters
run-custom: build
	@echo "Starting fake SSH server with custom parameters..."
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Docker target
docker:
	@echo "Building Docker image..."
	docker build -f Dockerfile.alpine -t $(DOCKER_IMAGE) .

docker-run: docker
	@echo "Running Docker container..."
	docker run -p 2222:2222 $(DOCKER_IMAGE)

# Example: make run-custom ARGS="--port 2223 --log /tmp/credentials.log" 