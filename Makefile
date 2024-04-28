.PHONY: all build docker-app docker-memcache clean rebuild

BIN_DIR := bin
APP_IMAGE_NAME := app-img
MEMCACHE_IMAGE_NAME := memcache-img

all: rebuild

build:
	@echo "Building the Go application..."
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BIN_DIR)/app cmd/main.go

docker-app: build
	@echo "Building the Docker image for the Go application..."
	@docker build -f Dockerfile.app -t $(APP_IMAGE_NAME) .

docker-memcache:
	@echo "Building the Docker image for Memcache..."
	@docker build -f Dockerfile.memcache -t $(MEMCACHE_IMAGE_NAME) .

clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR) || true

rebuild: clean docker-app docker-memcache
	@echo "Rebuilding and restarting the application..."
	@docker-compose down --rmi all --volumes
	@docker-compose up --build -d