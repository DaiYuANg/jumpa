# Architecture Overview

This project uses a DDD-style modular architecture centered on the `iam` bounded context.

## High-level Structure

- `cmd/server`: server bootstrap and container startup.
- `cmd/migrate`: migration entrypoint (embedded SQL migrations).
- `internal/modules/iam`: core IAM domain module.
- `internal/api`: API aggregation for system/auth/dashboard and module endpoint composition.
- `internal/http`: Fiber + Huma server runtime and authz middleware wiring.
- `internal/{config,db,kv,scheduler,event,auth,schema}`: platform and cross-cutting modules.

## IAM Module Layout

`internal/modules/iam` is split by DDD layers:

- `domain`
  - Domain models and core business language (`User`, `Role`, `Permission`, `PermissionGroup`).
- `application`
  - Use-case style service contracts and implementations.
  - Orchestrates domain logic and calls persistence abstractions.
- `infrastructure/persistence`
  - Repository interfaces exposed to application.
  - `dbx` implementation (`infrastructure/persistence/dbx`) for MySQL/Postgres/SQLite.
- `interfaces/http`
  - IAM HTTP endpoints, DTOs, mapping, paging/response helpers.
  - Registers IAM routes via `httpx.Endpoint`.

## Dependency Direction

Within IAM:

- `interfaces/http` -> `application` -> `infrastructure/persistence` -> `dbx/sql`
- `application` -> `domain`
- `domain` has no dependency on infrastructure or transport.

Across system:

- `internal/http` depends on `internal/api` and auth middleware.
- `internal/api` composes system/auth/dashboard endpoints plus IAM endpoints from `modules/iam/interfaces/http`.

## Module Wiring (DI / dix)

- `cmd/server` starts container with modules: config, event, db, kv, iam, scheduler, http.
- `internal/modules/iam/module.go` wires:
  - application services
  - persistence module imports
  - event bus integration where needed.
- `internal/modules/iam/interfaces/http/module.go` wires IAM endpoints.
- `internal/api/module.go` aggregates endpoint slices into one `[]httpx.Endpoint`.

## API Response Conventions

- Success envelope: `Result[T]`.
- Paging envelope: `Result[PageResult[T]]`.
- IAM list endpoints support page/pageSize and simple query filters.

## Migration Notes

The legacy `internal/service` and `internal/repo` layers were removed and replaced by IAM module layers.
All new business features should be added under `internal/modules/iam` (or new bounded contexts with the same style), not reintroduced into removed legacy paths.

## Extension Guidelines

When adding new business capabilities:

1. Add domain model/rules in `domain`.
2. Add use-case contract + implementation in `application`.
3. Add persistence interfaces/implementation in `infrastructure/persistence`.
4. Add transport DTO + endpoint wiring in `interfaces/http`.
5. Register providers in module files and expose endpoint slices through API aggregation.

When adding another bounded context (example: `audit`), mirror the IAM folder shape:

- `internal/modules/audit/domain`
- `internal/modules/audit/application`
- `internal/modules/audit/infrastructure`
- `internal/modules/audit/interfaces/http`

Keep the same dependency direction and DI composition pattern.
