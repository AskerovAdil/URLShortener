.PHONY: run build test tidy docker-up docker-down docker-logs

run:
	go run ./cmd/api

build:
	go build -o bin/api.exe ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

docker-up:
	docker compose -f deployments/docker-compose.yml up --build -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

docker-logs:
	docker compose -f deployments/docker-compose.yml logs -f api
