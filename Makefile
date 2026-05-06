.PHONY: up down logs test test-api test-go lint fmt seed integration

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

test: test-api test-go

test-api:
	cd api && python -m pytest

test-go:
	cd scheduler && go test ./...
	cd worker-go && go test ./...

lint:
	cd api && ruff check . && mypy .
	cd scheduler && golangci-lint run
	cd worker-go && golangci-lint run
	cd web && npm run lint

fmt:
	cd api && ruff format . && black .
	cd scheduler && gofmt -w .
	cd worker-go && gofmt -w .
	cd web && npm run format

seed:
	@echo "TODO: implement seed data"

integration:
	docker compose up -d
	@echo "TODO: run integration test suite"
