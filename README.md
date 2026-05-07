# MiniFlow

A minimal distributed workflow engine. Define DAGs in JSON, dispatch tasks to polyglot workers via Redis, visualize runs in a React UI.

> Learning project — Airflow/Temporal concepts built from scratch in ~2 weeks.

---

## Architecture

```
┌─────────────┐     HTTP      ┌───────────────┐
│   Web UI    │ ────────────► │   FastAPI      │  (workflow CRUD, run queries)
│  React 18   │               │   (Python)    │
└─────────────┘               └───────┬───────┘
                                      │ PostgreSQL
                               ┌──────▼───────┐
                               │  Scheduler   │  (DAG resolution, task dispatch)
                               │    (Go)      │
                               └──────┬───────┘
                                      │ Redis Streams
                     ┌────────────────┴────────────────┐
              ┌──────▼──────┐                  ┌───────▼──────┐
              │  Worker Go  │                  │  Worker Py   │
              │ shell · http│                  │ python · ml  │
              └─────────────┘                  └──────────────┘
```

All services communicate via **PostgreSQL** (source of truth) and **Redis** (task queue + event stream). No direct HTTP between scheduler and workers.

---

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
| IaC | Terraform 1.7+, AWS |

---

## Quickstart

**Prerequisites:** Docker, Docker Compose

```bash
# 1. Start all services
make up

# 2. Verify everything is healthy (api, scheduler, workers, web should all be Up)
docker compose ps

# 3. Open the web UI
open http://localhost:5173
```

> **Port note:** Postgres binds to `5433` and Redis to `6380` to avoid conflicts with other local services. The API is on `8000`, the web UI on `5173`.

---

## Demo Workflow

Create and run a three-task pipeline that shows dependency ordering:

```bash
# 1. Create the workflow
curl -s -X POST http://localhost:8000/workflows \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "demo-pipeline",
    "dag": {
      "tasks": [
        {
          "name": "fetch",
          "type": "shell",
          "command": "echo fetching data && sleep 1",
          "timeout_seconds": 30
        },
        {
          "name": "process",
          "type": "shell",
          "command": "echo processing && sleep 2",
          "depends_on": ["fetch"],
          "timeout_seconds": 30
        },
        {
          "name": "notify",
          "type": "shell",
          "command": "echo done",
          "depends_on": ["process"],
          "timeout_seconds": 10
        }
      ]
    }
  }' | python3 -m json.tool

# 2. Trigger a run (replace <workflow-id> with the id from step 1)
curl -s -X POST http://localhost:8000/workflows/<workflow-id>/runs | python3 -m json.tool

# 3. Poll until complete (replace <run-id>)
curl -s http://localhost:8000/runs/<run-id> | python3 -m json.tool
```

Or skip the terminal entirely and use the web UI at `http://localhost:5173`.

### Task types

| Type | Worker | Example payload |
|------|--------|----------------|
| `shell` | worker-go | `"command": "echo hello"` |
| `http` | worker-go | `"url": "https://example.com", "method": "GET"` |
| `python` | worker-py | `"code": "print('hello')"` |

### Task options

```jsonc
{
  "name": "my-task",
  "type": "shell",
  "command": "...",
  "depends_on": ["other-task"],   // upstream dependencies (default: none)
  "timeout_seconds": 60,          // default: 300
  "max_retries": 3                // retries on failure with exponential backoff (default: 0)
}
```

---

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workflows/` | List all workflows |
| `POST` | `/workflows` | Create a workflow |
| `GET` | `/workflows/{id}` | Get a workflow |
| `POST` | `/workflows/{id}/runs` | Trigger a run |
| `GET` | `/runs` | List runs (optional `?workflow_id=`) |
| `GET` | `/runs/{id}` | Get run detail with task statuses |
| `POST` | `/runs/{id}/cancel` | Cancel a pending/running run |
| `GET` | `/health` | Health check |

---

## Development

```bash
make test        # run all test suites (api, go, worker-py, web)
make test-api    # Python unit tests only
make test-go     # Go unit tests only
make integration # spin up compose, run integration suite
make lint        # lint all languages
make fmt         # format all code
make logs        # tail all service logs
make down        # stop all services
```

### Task state machine

```
pending → queued → running → success
                           → failed → retrying → (running again, up to max_retries)
                                    → failed (retries exhausted)
                 → cancelled
```

A task only moves to `queued` when all upstream dependencies are `success`. Cancelling a run marks `pending` and `queued` tasks as `cancelled`; already-running tasks complete normally.

---

## Cloud Deployment (AWS)

Terraform modules under `infra/` deploy the full stack to AWS ECS Fargate.

```
infra/
├── modules/
│   ├── network/   # VPC, subnets, NAT gateway, security groups
│   ├── db/        # RDS PostgreSQL 16
│   ├── redis/     # ElastiCache Redis 7
│   └── services/  # ECR repos, ECS cluster, ALB, task definitions
└── envs/dev/      # dev environment wiring
```

**Deploy to dev:**

```bash
# 1. Configure credentials
export AWS_PROFILE=your-profile   # or set AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY

# 2. Create your tfvars
cd infra/envs/dev
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars — set db_password and region

# 3. Init and apply
terraform init
terraform plan
terraform apply

# 4. Push Docker images to ECR (use URLs from terraform output)
terraform output ecr_repository_urls

# Build and push each service:
# docker build -t <ecr-url>/miniflow/api:latest ./api
# docker push <ecr-url>/miniflow/api:latest
# ... repeat for scheduler, worker-go, worker-py, web
```

After images are pushed, ECS will pull and start the containers. The ALB DNS is in `terraform output web_url`.

> **Teardown:** `terraform destroy` removes all resources. RDS has `skip_final_snapshot = true` and `deletion_protection = false` for easy cleanup in dev.

---

## Project Structure

```
miniflow/
├── api/          # FastAPI — workflow CRUD, run queries
├── scheduler/    # Go — DAG resolution, task dispatch
├── worker-go/    # Go — shell and http task execution
├── worker-py/    # Python — python task execution
├── web/          # React UI — DAG visualization, run list
├── infra/        # Terraform — AWS deployment
├── shared/       # JSON Schema contracts (task messages, events)
├── tests/        # Integration tests (run against docker-compose)
├── docs/adr/     # Architecture Decision Records
├── docker-compose.yml
└── Makefile
```

See `docs/adr/` for design decisions and `shared/schemas/` for inter-service message contracts.
