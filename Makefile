.PHONY: run build test tidy

run:
	go run ./cmd/api

build:
	go build -o bin/api.exe ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy
