.PHONY: all build test clean run docker-up docker-down migrate-up migrate-down

# Service list
SERVICES := auth-service post-service

# Build all services
build:
	@for service in $(SERVICES); do \
		echo "Building $$service..." ; \
		cd $$service && go build -o bin/app ./cmd/main.go && cd .. ; \
	done

# Run tests for all services
test:
	@for service in $(SERVICES); do \
		echo "Testing $$service..." ; \
		cd $$service && go test ./... && cd .. ; \
	done

# Clean build artifacts
clean:
	@for service in $(SERVICES); do \
		echo "Cleaning $$service..." ; \
		rm -rf $$service/bin ; \
	done

# Start all services using docker-compose
docker-up:
	docker-compose up --build -d

# Stop all services
docker-down:
	docker-compose down

# Run database migrations up
migrate-up:
	@for service in $(SERVICES); do \
		echo "Running migrations for $$service..." ; \
		migrate -path $$service/migrations -database "postgresql://postgres:postgres@localhost:5432/blogbish?sslmode=disable" up ; \
	done

# Rollback database migrations
migrate-down:
	@for service in $(SERVICES); do \
		echo "Rolling back migrations for $$service..." ; \
		migrate -path $$service/migrations -database "postgresql://postgres:postgres@localhost:5432/blogbish?sslmode=disable" down ; \
	done

# Run a specific service locally
run-%:
	@echo "Running $*..."
	@cd $* && go run cmd/main.go 