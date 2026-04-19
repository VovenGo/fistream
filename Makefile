.PHONY: test build up down

test:
	go test ./...

build:
	go build ./cmd/api

up:
	docker compose up --build

down:
	docker compose down

