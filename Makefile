.DEFAULT_GOAL := dev

# Build Docker image
build:
	@docker compose build

# Start container in dev mode (hot reload with Air)
dev:
	@docker compose up

# Run without Docker (local go run main.go)
run:
	@go run ./cmd/api/main.go

# Build local binary
build-local:
	@go build -o bin/app ./cmd/api

# Stop container & remove image
stop:
	@docker compose down
	@docker rmi ticket-booking || true

# View logs
logs:
	@docker compose logs -f app

# Clean up Docker environment
prune:
	@docker system prune -f
