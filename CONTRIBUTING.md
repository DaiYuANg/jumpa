# Contributing Guide

## Development Principles

- Keep business code in bounded contexts under `internal/modules/*`.
- Follow DDD layering:
  - `domain` -> business language and rules
  - `application` -> use-case orchestration
  - `infrastructure` -> db/kv/adapter implementations
  - `interfaces` -> transport/http DTO and endpoint wiring
- Do not reintroduce removed legacy layers (`internal/service`, `internal/repo`).

## Dependency Rules

- `domain` must not depend on `application`, `infrastructure`, or HTTP packages.
- `application` can depend on `domain` and persistence abstractions only.
- `infrastructure` implements abstractions; it should not contain HTTP concerns.
- `interfaces/http` can depend on `application` but not on concrete DB implementations.

## Feature Checklist

When adding a new business capability:

1. Add or update domain models/rules in `domain`.
2. Add application contract/use-case method in `application`.
3. Add persistence interface and implementation in `infrastructure/persistence`.
4. Add endpoint DTO mapping and route registration in `interfaces/http`.
5. Wire providers in module files and expose endpoint slices for API aggregation.
6. Add/adjust SQL migration if schema changes are required.

## API Conventions

- Use unified response envelope:
  - `Result[T]`
  - `Result[PageResult[T]]` for list APIs
- Keep DTOs in interface layer, not in domain/application.
- Keep handler functions thin: parse/validate -> call application -> map response.

## Testing and Validation

Before opening a PR:

- Run `go test ./...`
- Run lint checks used by the project task workflow.
- Verify migrations can run with configured local database.
- Verify changed endpoints still align with frontend contract.

## Pull Request Expectations

- Explain **why** the change is needed.
- List impacted modules and layers.
- Include migration notes for schema or config changes.
- Keep refactors and behavior changes clearly separated when possible.

## Commit Style

- Prefer concise conventional messages:
  - `feat: ...`
  - `fix: ...`
  - `refactor: ...`
  - `docs: ...`
- Make commit scope obvious (example: `refactor: move iam persistence to module dbx layer`).
