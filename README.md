# MiniFlow

A minimal distributed workflow engine. Define DAG-based workflows in YAML, dispatch tasks to polyglot workers via Redis, visualize runs in a web UI.

> Learning project — Airflow/Temporal concepts in ~2 weeks.

## Stack

| Layer | Technology |
|-------|-----------|
| API | Python 3.12, FastAPI, SQLAlchemy |
| Scheduler | Go 1.22 |
| Worker (Go) | Go 1.22 — `shell`, `http` tasks |
| Worker (Python) | Python 3.12 — `python`, `ml` tasks |
| Web UI | React 18, Vite, TypeScript, react-flow |
| Database | PostgreSQL 16 |
| Queue | Redis 7 |

## Quickstart

```bash
# Start infrastructure
make up

# Verify services are healthy
docker compose ps

# Tail logs
make logs
```

## Development

```bash
make test        # run all test suites
make lint        # lint all languages
make fmt         # format all code
make seed        # load example workflows
make integration # run integration tests
```

## Architecture

All services communicate via **PostgreSQL** (state) and **Redis** (task queue + event stream). No direct HTTP calls between scheduler and workers.

See `docs/adr/` for architecture decisions and `shared/schemas/` for inter-service message contracts.
