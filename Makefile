.PHONY: run build test tidy migrate-up migrate-down docker-up docker-down docker-logs

run:
	go run ./cmd/api

build:
	go build -o bin/api.exe ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

migrate-up:
	go run ./cmd/migrate -direction up

migrate-down:
	go run ./cmd/migrate -direction down

docker-up:
	docker compose -f deployments/docker-compose.yml up --build -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

docker-logs:
	docker compose -f deployments/docker-compose.yml logs -f api
