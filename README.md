# arcgo-rbac-template

A backend RBAC template based on ArcGo ecosystem, now organized with a DDD-style IAM module.

## Features

- Multi-database support via config: `sqlite`, `mysql`, `postgres`
- Fiber-based HTTP runtime with `httpx` endpoint registration
- JWT auth flow (`login`, `refresh`, `logout`, `me`) with revocation support
- Valkey integration via `kvx` (including distributed scheduler lock use-cases)
- Embedded SQL migrations and dedicated `cmd/migrate` process
- Modular IAM architecture (`domain/application/infrastructure/interfaces`)

## Architecture Docs

- Project architecture: `ARCHITECTURE.md`
- Contribution rules and checklists: `CONTRIBUTING.md`

## Requirements

- Go 1.26+
- Optional:
  - MySQL or Postgres for external DB mode
  - Valkey for token/session revocation and distributed lock mode
  - Docker + Docker Compose for containerized runs

## Quick Start

### 1) Install dependencies

```bash
go mod tidy
```

### 2) Configure environment

Copy and edit values:

```bash
cp .env.example .env
```

Typical local SQLite setup:

```env
APP_DB_DRIVER=sqlite
APP_DB_DSN=file:backend.db?cache=shared
APP_SERVER_PORT=8080
```

### 3) Run migrations

```bash
go run ./cmd/migrate
```

### 4) Run server

```bash
go run ./cmd/server
```

Default API base path: `http://localhost:8080/api`

## Common Commands

Using Taskfile:

```bash
task fmt
task lint
task test
task check
task build:bins
```

## Binaries

Build outputs (by default):

- `dist/server` (or `dist/server.exe` on Windows)
- `dist/migrate` (or `dist/migrate.exe` on Windows)

## Docker

- Multi-stage Docker build is configured for server/migrate targets.
- Production orchestration is provided via `docker-compose.prod.yml`.
- Existing local compose files remain unchanged.

## Main Runtime Modules

- `cmd/server`: app bootstrap and DI container startup
- `cmd/migrate`: migration bootstrap
- `internal/modules/iam`: IAM bounded context
  - `domain`
  - `application`
  - `infrastructure/persistence`
  - `interfaces/http`
- `internal/api`: endpoint aggregation layer
- `internal/http`: Fiber + middleware + httpx runtime wiring

## Notes

- Legacy `internal/service` and `internal/repo` layers were removed after DDD migration.
- Add new business capabilities under `internal/modules/*` in the same layered style.
