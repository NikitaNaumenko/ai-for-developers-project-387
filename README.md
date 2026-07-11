### Hexlet tests and linter status:
[![Actions Status](https://github.com/NikitaNaumenko/ai-for-developers-project-386/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/NikitaNaumenko/ai-for-developers-project-386/actions)
# Calendar API

Go backend template with PostgreSQL, `sqlc`, and HTTP server code generated from OpenAPI.

## Stack

- Go
- PostgreSQL
- `sqlc` for type-safe SQL
- `oapi-codegen` for server interfaces and request/response types from OpenAPI
- `chi` for routing
- `pgx` for PostgreSQL access
- `zap` for structured logs
- React
- Mantine
- Vite
- TanStack Query
- Stoplight Prism for OpenAPI mock API

## Layout

- `api/openapi.yaml` - HTTP contract, the source of truth for handlers.
- `db/migrations` - PostgreSQL migrations.
- `db/queries` - SQL queries consumed by `sqlc`.
- `cmd/api` - application entrypoint.
- `internal/config` - environment configuration.
- `internal/handler` - implementation of generated API interfaces.
- `internal/httpserver` - router, JSON errors, middleware.
- `internal/storage` - database connection lifecycle.
- `frontend` - React + Mantine application.
- `api/typespec` - new TypeSpec contract workspace.

Generated packages:

- `internal/api/gen` from `api/openapi.yaml`
- `internal/db` from `db/queries/*.sql`

## Commands

```sh
make tools
make generate
make test
make run
```

Frontend:

```sh
make frontend-install
make frontend-dev
make frontend-build
```

The Vite dev server proxies `/api/*` to `http://localhost:8080`.

TypeSpec:

```sh
make typespec-install
make typespec-build
make typespec-check
make mock-api
```

TypeSpec output is generated to `api/openapi.yaml`. The TypeSpec workspace is the source for the OpenAPI contract; run `make generate` after updating operations that should affect the Go server.

Mock API:

```sh
make typespec-install
make mock-api
```

The Prism mock server runs on `http://localhost:4010` and serves the OpenAPI paths without the `/api` prefix, matching the Go router. For example, `GET http://localhost:4010/event-types`.

To point the Vite dev proxy at the mock API instead of the Go backend:

```sh
VITE_API_PROXY_TARGET=http://localhost:4010 make frontend-dev
```

PostgreSQL:

```sh
docker compose up -d postgres
```

Migrations are intentionally tool-agnostic. The SQL files are compatible with common migration tools such as `golang-migrate`, `tern`, and `atlas`.

## Render deploy

The repository includes a `render.yaml` Blueprint that creates:

- `calendar-api` - Docker web service with the Go API and built React frontend.
- `calendar-db` - Render Postgres database.

After committing and pushing changes to GitHub, create a new Render Blueprint from this repository. Render injects `DATABASE_URL` from `calendar-db`; the app reads Render's `PORT` automatically and serves the frontend from the Docker image.
