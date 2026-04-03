SHELL := /bin/bash

.PHONY: up down clean run-tests test migrate-up migrate-down

up:
	docker compose up --build -d

down:
	docker compose down

run-tests:
	docker run --rm --network=host tests:latest

clean:
	docker compose down -v

test:
	make clean
	make up
	@echo wait cluster to start && sleep 10
	make run-tests
	make clean
	@echo "test finished"

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down
