.PHONY: up down logs test test-api test-go test-worker-py lint fmt seed integration

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

test: test-api test-go test-worker-py test-web

test-api:
	cd api && .venv/bin/python -m pytest tests/ -v

test-go:
	cd scheduler && go test ./...
	cd worker-go && go test ./...

test-worker-py:
	cd worker-py && .venv/bin/python -m pytest tests/ -v

test-web:
	cd web && npm run test -- --run

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
	docker compose up -d --build
	@test -d tests/integration/.venv || python3 -m venv tests/integration/.venv
	tests/integration/.venv/bin/pip install -q -r tests/integration/requirements.txt
	tests/integration/.venv/bin/pytest tests/integration/ -v
