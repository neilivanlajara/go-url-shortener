.PHONY: up down build run logs tidy test

up:
	docker compose up -d

down:
	docker compose down -v

build:
	docker compose build --no-cache

run:
	go run ./cmd/server

logs:
	docker compose logs -f app

tidy:
	go mod tidy

test:
	go test ./...
