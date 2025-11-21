BINARY_NAME = prmanager
CMD_DIR = ./cmd/pr-reviewer-service

.PHONY: build run-local test docker-build docker-up docker-down

build:
	go build -o bin/$(BINARY_NAME) $(CMD_DIR)

run-local:
	CONFIG_PATH=./config/local.yaml go run $(CMD_DIR)

test:
	go test ./...

docker-build:
	docker build -t $(BINARY_NAME) .

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
