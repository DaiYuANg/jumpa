# Architecture Overview

This project uses a DDD-style modular architecture for a bastion and jump-host control plane. The current codebase keeps `iam` as a bounded context and evolves `bastion` into an internal control-plane module with explicit asset, access, and session slices.

## High-level Structure

- `cmd/server`: server bootstrap and container startup.
- `cmd/gateway`: bastion SSH gateway bootstrap.
- `cmd/migrate`: migration entrypoint (embedded SQL migrations).
- `internal/modules/iam`: core IAM domain module.
- `internal/modules/bastion`: bastion control-plane module.
- `internal/identity`: identity source resolution for application-managed and OS-backed login modes.
- `internal/api`: API aggregation for system/auth/dashboard and module endpoint composition.
- `internal/http`: Fiber + Huma server runtime and authz middleware wiring.
- `internal/{config,db,kv,scheduler,event,auth,schema}`: platform and cross-cutting modules.

## Bounded Contexts

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

`internal/modules/bastion` currently provides the first control-plane slice and is now internally split across dedicated internal submodules:

- `overview`
  - wires runtime overview/readiness services.
- `asset`
  - wires host and host-account services plus gateway target resolution.
- `access`
  - wires policy matching and approval-request services.
- `session`
  - wires session query/runtime lifecycle services.

The shared business code remains grouped by layer under `application/*`, `domain/*`, and `ports/*`, but the root `bastion.Module` now acts as an aggregator instead of directly providing every service.

- `domain`
  - Host, access policy, session, access request, and overview models.
- `application`
  - Services for overview, assets, access policies, access requests, and sessions.
- `interfaces/http`
  - one `BastionEndpoint` registration surface, split internally into overview, asset, access, and session route files.
  - `/api/bastion/overview`, `/api/assets/hosts`, `/api/access-policies`, `/api/access-requests`, `/api/sessions`.

The gateway runtime currently uses a pragmatic login convention:

- `principal#host`
- `principal#host#account`

It authenticates `principal` through the configured identity source, resolves `host` and optional `account` from bastion tables, evaluates access policy, and then opens a downstream SSH client connection.
Policies can currently target `user`, `principal`, `email`, `role`, or `*` subjects. When a matched policy requires approval, the gateway persists an access request and denies login until that request is approved.

`internal/identity` resolves how authentication should be sourced:

- `local`
  - Users and credentials are managed in the application database.
- `os`
  - Authentication is delegated to Linux PAM, Windows local/domain accounts, or macOS OpenDirectory, while bastion authorization remains application-owned.
  - Linux PAM and macOS OpenDirectory are modeled as platform-native backends behind build tags; Windows uses a native `LogonUserW` backend.

## Dependency Direction

Within IAM and bastion:

- `interfaces/http` -> `application` -> `infrastructure/persistence` -> `dbx/sql`
- `application` -> `domain`
- `domain` has no dependency on infrastructure or transport.

For identity:

- `internal/identity` depends on config only.
- Authentication source selection is independent from bastion authorization and audit persistence.

Across system:

- `internal/http` depends on `internal/api` and auth middleware.
- `internal/api` composes system/auth/dashboard endpoints plus IAM endpoints from `modules/iam/interfaces/http`.

## Module Wiring (DI / dix)

- `cmd/server` starts container with modules: config, event, db, kv, iam, scheduler, http.
- `internal/modules/iam/module.go` wires:
  - application services
  - persistence module imports
  - event bus integration where needed.
- `internal/modules/bastion/module.go` imports `overview`, `asset`, `access`, and `session` submodules, plus shared persistence/config dependencies.
- `internal/modules/iam/interfaces/http/module.go` wires IAM endpoints.
- `internal/modules/bastion/interfaces/http/module.go` wires bastion endpoints.
- `internal/api/module.go` aggregates endpoint slices into one `[]httpx.Endpoint`.

## API Response Conventions

- Success envelope: `Result[T]`.
- Paging envelope: `Result[PageResult[T]]`.
- IAM list endpoints support page/pageSize and simple query filters.

## Migration Notes

The legacy `internal/service` and `internal/repo` layers were removed and replaced by module layers.
New business features should be added under `internal/modules/*` with the same dependency direction, not reintroduced into removed legacy paths.

## Extension Guidelines

When adding new business capabilities:

1. Add domain model/rules in `domain`.
2. Add use-case contract + implementation in `application`.
3. Add persistence interfaces/implementation in `infrastructure/persistence`.
4. Add transport DTO + endpoint wiring in `interfaces/http`.
5. Register providers in module files and expose endpoint slices through API aggregation.

Planned next bounded contexts after this landing:

- `internal/modules/assets`
- `internal/modules/access`
- `internal/modules/session`
- `internal/modules/audit`

Recommended runtime split:

1. `cmd/server` remains the HTTP control plane.
2. `cmd/gateway` now hosts the bastion SSH server and target connection proxy loop.
3. Both runtimes should share IAM, policy, and audit persistence modules.
