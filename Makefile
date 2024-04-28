.PHONY: build docker-app docker-memcache clean restart

BIN_DIR := bin

build:
	@echo "Building the Go application..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BIN_DIR)/app .
	
docker-app: build
	@echo "Building the Docker image for the Go application..."
	docker build -f Dockerfile.app -t chi-memcache-app-img .

docker-memcache:
	@echo "Building the Docker image for Memcache..."
	docker build -f Dockerfile.memcache -t memcache-img .

clean:
	@echo "Cleaning up..."
	rm -rf $(BIN_DIR)

restart: clean docker-app docker-memcache
	@echo "Restarting the application..."
	docker-compose down --rmi all --volumes
	docker-compose up --build -d
