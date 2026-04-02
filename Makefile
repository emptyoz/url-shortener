SHELL := /bin/bash

.PHONY: up down clean run-tests test migrate-up migrate-down

up:
	docker compose up --build -d

down:
	docker compose down

clean:
	docker compose down -v --remove-orphans

test:
	go test -v -cover ./...

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down
