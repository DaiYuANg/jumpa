# jumpa

A bastion and jump-host control-plane backend based on the ArcGo ecosystem, organized with DDD-style modules.

## Features

- Multi-database support via config: `sqlite`, `mariadb`, `postgres`
- Fiber-based HTTP runtime with `httpx` endpoint registration
- Dedicated `cmd/gateway` runtime for the bastion SSH entrypoint
- JWT auth flow (`login`, `refresh`, `logout`, `me`) with revocation support
- Valkey integration via `kvx` (including distributed scheduler lock use-cases)
- Embedded SQL migrations and dedicated `cmd/migrate` process
- Modular IAM architecture plus bastion scaffolding (`domain/application/interfaces`)
- Identity-source abstraction for `local` and `os` login modes
- OS identity planning for Linux PAM, Windows local/domain accounts, and macOS OpenDirectory
- Bastion schema baseline for hosts, access policies, sessions, access requests, and command audit records

## Architecture Docs

- Project architecture: `ARCHITECTURE.md`
- Contribution rules and checklists: `CONTRIBUTING.md`

## Requirements

- Go 1.26+
- Optional:
  - MariaDB or Postgres for external DB mode
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

Start the bastion gateway scaffold:

```bash
go run ./cmd/gateway
```

Current SSH login format:

```text
principal#host
principal#host#account
```

Examples:

```text
alice#prod-web-01
alice#prod-web-01#ubuntu
```

Target host credential conventions:

- `authentication_type=passthrough`: reuse the password the user entered to log into jumpa
- `credential_ref=env:VAR_NAME`: use the password stored in environment variable `VAR_NAME`
- `credential_ref=file:/path/to/private_key`: use the private key file for downstream SSH auth

Target host key verification:

- `APP_BASTION_SSH_HOST_KEY_POLICY=insecure`: skip downstream host key verification
- `APP_BASTION_SSH_HOST_KEY_POLICY=known_hosts`: verify downstream host keys against `APP_BASTION_SSH_KNOWN_HOSTS_PATH`
- Production deployments should use `known_hosts`

Useful early bastion endpoints:

- `GET /api/bastion/overview`
- `GET /api/assets/hosts`
- `GET /api/access-policies`
- `GET /api/access-requests`
- `POST /api/access-requests/{id}/approve`
- `POST /api/access-requests/{id}/reject`
- `GET /api/sessions`

Approval workflow defaults:

- Approved access requests expire after `APP_BASTION_ACCESS_APPROVAL_TTL_MIN` minutes
- Approved access requests are consumed on the first successful SSH session start

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
- `cmd/gateway`: bastion SSH gateway runtime
- `cmd/migrate`: migration bootstrap
- `internal/modules/iam`: IAM bounded context
- `internal/modules/bastion`: bastion control-plane scaffolding
- `internal/identity`: application-managed vs OS-backed identity source selection
  - `domain`
  - `application`
  - `interfaces/http`
- `internal/api`: endpoint aggregation layer
- `internal/http`: Fiber + middleware + httpx runtime wiring

## Notes

- Legacy `internal/service` and `internal/repo` layers were removed after DDD migration.
- The SSH gateway now authenticates through the configured identity provider, resolves target hosts/accounts, enforces bastion access policies, and opens downstream SSH client connections.
- `local` identity mode can already authenticate against the legacy `users` table when it contains bcrypt `password_hash` values.
- `os` identity mode now has explicit Linux/Windows/macOS backend slots (`pam`, `winlogon`, `opendirectory`).
- Linux PAM and macOS OpenDirectory paths are wired behind platform build tags. Linux requires cgo plus PAM headers and libraries; macOS requires cgo plus the OpenDirectory framework.
- Windows `winlogon` authentication is wired through `LogonUserW`, with support for `DOMAIN\user`, `user@domain`, and local account forms.
- The gateway resolves target hosts and host accounts from `bastion_hosts` and `bastion_host_accounts`, then opens a downstream SSH client connection.
- Access policy enforcement is wired into the live proxy path, with `subjectType` support for `user`, `principal`, `email`, `role`, and `*`.
- Policies marked `approvalRequired=true` now create bastion access requests in `bastion_access_requests`; approved requests are reusable in the current minimal workflow.
- Add new business capabilities under `internal/modules/*` in the same layered style.
