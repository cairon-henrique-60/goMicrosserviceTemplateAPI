SHELL := /bin/sh

.PHONY: tidy test fmt up down logs migrate-up migrate-down build

tidy:
	cd pkg/platform && go mod tidy
	cd apps/gateway && go mod tidy
	cd apps/auth-service && go mod tidy
	cd apps/catalog-service && go mod tidy
	cd apps/order-service && go mod tidy
	cd apps/notification-worker && go mod tidy

fmt:
	gofmt -w apps pkg

test:
	go test ./pkg/platform/...
	go test ./apps/gateway/...
	go test ./apps/auth-service/...
	go test ./apps/catalog-service/...
	go test ./apps/order-service/...
	go test ./apps/notification-worker/...

up:
	docker compose --env-file .env up --build -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=200

migrate-up:
	docker compose run --rm migrations-up

migrate-down:
	docker compose run --rm migrations-down

build:
	docker compose build
