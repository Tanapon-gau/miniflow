# MiniFlow

Distributed workflow engine. DAG-based task orchestration with polyglot workers.

This is a learning project to explore workflow engines (Airflow/Temporal style), distributed systems patterns, Docker, and Terraform.

---

## Project Goal

Build a minimal but production-shaped workflow engine where:
- Users define workflows as DAGs (YAML/JSON)
- A scheduler dispatches ready tasks to a queue
- Polyglot workers (Go and Python) consume tasks
- A web UI visualizes runs and streams logs
- The whole stack runs locally via Docker Compose, and can deploy to cloud via Terraform

Scope: 2-week MVP. Prioritize clarity over feature completeness.

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| API Server | Python 3.12, FastAPI, SQLAlchemy, Pydantic v2 |
| Scheduler | Go 1.22, Gin, sqlx |
| Worker (Go) | Go 1.22 — handles `shell`, `http` task types |
| Worker (Python) | Python 3.12 — handles `python`, `ml` task types |
| Web UI | React 18, Vite, TypeScript, react-flow |
| Database | PostgreSQL 16 |
| Queue | Redis 7 (Streams + Lists) |
| Containers | Docker, Docker Compose |
| IaC | Terraform 1.7+ |
| CI | GitHub Actions |

---

## Repository Structure

```
miniflow/
├── api/              # FastAPI service — workflow CRUD, run queries
│   ├── app/
│   ├── tests/
│   ├── Dockerfile
│   └── pyproject.toml
├── scheduler/        # Go scheduler — DAG resolution, task dispatch
│   ├── cmd/scheduler/
│   ├── internal/
│   ├── Dockerfile
│   └── go.mod
├── worker-go/        # Go worker — fast/concurrent task execution
│   ├── cmd/worker/
│   ├── internal/
│   ├── Dockerfile
│   └── go.mod
├── worker-py/        # Python worker — Python/ML tasks
│   ├── worker/
│   ├── tests/
│   ├── Dockerfile
│   └── pyproject.toml
├── web/              # React UI — DAG visualization, log streaming
│   ├── src/
│   ├── Dockerfile
│   └── package.json
├── infra/            # Terraform — cloud deployment
│   ├── modules/
│   ├── envs/dev/
│   └── envs/prod/
├── shared/           # Shared schemas (JSON Schema for tasks/events)
├── docs/             # ADRs, architecture diagrams
├── scripts/          # Dev helpers (seed data, load tests)
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## Core Concepts (read before writing code)

### Workflow
A DAG of tasks. Defined in YAML. Stored in `workflows` table as JSON.

### Run
An instance of a workflow being executed. One workflow can have many runs.

### Task
A single node in the DAG. Has a `type` (shell, http, python, ml) that determines which worker class picks it up.

### Task State Machine
`pending` → `queued` → `running` → (`success` | `failed` | `retrying`)

A task only becomes `queued` when ALL its upstream dependencies are `success`.

### Routing
Each task has a `type`. Workers subscribe to specific types via Redis queue names:
- `tasks:shell`, `tasks:http` → consumed by `worker-go`
- `tasks:python`, `tasks:ml` → consumed by `worker-py`

---

## Conventions

### Commits
Conventional Commits, lowercase scope:

```
feat(api): add workflow validation endpoint
fix(scheduler): handle empty DAG gracefully
chore(infra): bump terraform aws provider
docs(adr): record queue choice rationale
test(worker-go): add retry edge cases
refactor(worker-py): extract task runner interface
```

Commit early and often. Each commit should leave the repo in a working state (tests pass, services start).

### Branches
- `main` — protected, always deployable
- `feat/<area>-<short-desc>` — feature work (e.g. `feat/scheduler-retry`)
- `fix/<area>-<short-desc>` — bug fixes

### Pull Requests
- Squash merge into main
- PR title follows Conventional Commits
- Must include: what changed, why, how to test

### Code Style
- **Python**: ruff + black, type hints required, mypy strict on `api/` and `worker-py/`
- **Go**: gofmt, golangci-lint, errors.Is/As over string matching
- **TypeScript**: eslint + prettier, strict mode on
- **SQL**: snake_case tables and columns, plural table names

### Testing
- Every PR must add or update tests for changed behavior
- API: pytest + httpx test client
- Go services: stdlib testing + testify
- Web: vitest + react-testing-library
- Integration tests live in `tests/integration/` at repo root, run against docker-compose

---

## Inter-Service Contract

All services communicate via two channels only:

1. **PostgreSQL** — source of truth for workflow/run/task state
2. **Redis** — task queue (push) and event stream (pub/sub for logs)

**No direct HTTP calls between scheduler and workers.** This keeps services decoupled and language-agnostic.

### Task Message Format (Redis)

```json
{
  "task_id": "uuid",
  "run_id": "uuid",
  "type": "shell",
  "payload": { "command": "echo hello" },
  "timeout_seconds": 300,
  "max_retries": 3,
  "attempt": 1
}
```

### Event Format (Redis stream `events`)

```json
{
  "task_id": "uuid",
  "event": "started" | "log" | "succeeded" | "failed",
  "timestamp": "ISO-8601",
  "data": { ... }
}
```

Schemas live in `shared/schemas/` and are validated on both ends.

---

## Common Commands

> **Local port note:** postgres binds to `5433` and redis to `6380` to avoid conflicts with other projects on this machine.

```bash
# Local dev
docker compose up -d              # start all services
docker compose logs -f scheduler  # tail one service
make seed                         # load example workflows

# Testing
make test                         # run all test suites
make test-api                     # python tests only
make test-go                      # go tests only
make integration                  # spin up compose, run integration suite

# Linting
make lint                         # lint all languages
make fmt                          # format all code

# Infra
cd infra/envs/dev && terraform plan
cd infra/envs/dev && terraform apply
```

---

## Working with Claude Code on This Repo

### Preferred workflow
1. **Plan before code.** For any non-trivial change, propose the approach first (files to touch, contract changes, test plan). Wait for confirmation before editing.
2. **One concern per commit.** If a change spans API + scheduler + schema, that is acceptable as one commit only if they are tightly coupled. Otherwise split.
3. **Tests with the change.** Don't add code without tests in the same commit.
4. **Update docs.** If a change affects the inter-service contract, README, or this file, update them in the same PR.

### Things to ask before doing
- Adding a new external dependency (library, service)
- Changing the inter-service contract (queue names, message shape, DB schema)
- Introducing a new task type
- Changing how Terraform state is stored

### Things you can do without asking
- Refactor within a single service (no contract change)
- Add tests
- Fix lint/format issues
- Improve error messages and logs
- Update inline code comments

### When stuck
Read in this order before guessing:
1. `docs/adr/` — architecture decisions
2. `shared/schemas/` — message contracts
3. Existing tests — they document expected behavior
4. Then ask the user

---

## Roadmap (current state and next steps)

Update this section as work progresses. Mark items `[x]` when shipped.

### Week 1 — Core Engine
- [x] Monorepo skeleton + docker-compose
- [x] DB schema + migrations (workflows, runs, tasks)
- [x] FastAPI: workflow CRUD, run trigger, run query
- [x] Go scheduler: poll runs, resolve DAG, dispatch
- [x] Go worker: shell + http task types
- [ ] Python worker: python task type
- [ ] End-to-end happy path test

### Week 2 — Polish + Cloud
- [ ] React UI: DAG view, run list, log stream
- [ ] Retry logic + exponential backoff
- [ ] Timeouts and cancellation
- [ ] Terraform modules (network, db, redis, services)
- [ ] Deploy to cloud dev environment
- [ ] README with quickstart and demo workflow

---

## Out of Scope (do not build, do not suggest)

To keep scope tight, these are explicitly deferred:
- Authentication / multi-tenancy
- Workflow versioning beyond a single `version` field
- Distributed scheduler (single scheduler instance is fine for MVP)
- Custom DSL — YAML/JSON only
- Plugin system for task types — hardcoded enum is fine
- Metrics dashboards (Prometheus/Grafana) — basic logs only

If a requirement seems to need any of the above, surface it as a question rather than building it.

---

## Glossary

- **DAG** — Directed Acyclic Graph. The shape of a workflow.
- **Topological sort** — Ordering of DAG nodes such that every node comes after its dependencies. Used to figure out which tasks are ready.
- **Idempotent** — Running a task twice with the same input produces the same outcome. Critical for safe retries.
- **Backpressure** — When workers are overloaded, the queue grows. Scheduler must respect queue length before dispatching more.