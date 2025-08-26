# Makefile for common development tasks.
#
# Targets:
#   build        Compile the Go application into a binary named `forum`.
#   run          Run the application directly with `go run`.
#   build-docker Build the Docker image tagged `forum`.
#   run-docker   Run the Docker image, mapping port 8080.

.PHONY: build run build-docker run-docker

build:
	@echo "Building forum binary..."
	go build -o forum

run:
	@echo "Running application..."
	go run main.go

build-docker:
	@echo "Building Docker image..."
	docker build -t forum .

run-docker:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 forum
