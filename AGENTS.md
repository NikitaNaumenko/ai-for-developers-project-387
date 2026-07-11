# Repository Guidelines

## Project Structure & Module Organization

This repository contains a Go calendar API, PostgreSQL schema, and Vite React frontend. The API entrypoint is `cmd/api`, with implementation code under `internal/`: `handler` for API implementations, `httpserver` for routing, `storage` for database lifecycle, and `config` for environment loading. API contracts live in `api/typespec` and emit `api/openapi.yaml`; generated API types are in `internal/api/gen`. SQL migrations are in `db/migrations`, sqlc query sources are in `db/queries`, and generated database code is in `internal/db`. Frontend source, styles, and API client code live under `frontend/src`.

## Build, Test, and Development Commands

- `make tools` installs pinned `oapi-codegen` and `sqlc` CLIs.
- `make generate` regenerates OpenAPI server types and sqlc database code.
- `make test` runs all Go tests with `go test ./...`.
- `make run` starts the Go API from `cmd/api`.
- `make build` writes the API binary to `bin/calendar-api`.
- `docker compose up -d postgres` starts the local PostgreSQL service.
- `make frontend-install`, `make frontend-dev`, and `make frontend-build` install, run, and build the Vite frontend.
- `make typespec-install`, `make typespec-build`, `make typespec-check`, and `make mock-api` manage the TypeSpec contract and Prism mock API.

## Coding Style & Naming Conventions

Format Go code with `gofmt`; keep package names short, lowercase, and aligned with directory names. Name tests like `TestAvailableSlotsSkipsOccupiedSlot`. Treat `api/typespec/main.tsp`, `api/openapi.yaml`, and `db/queries/*.sql` as sources for generated code; run `make generate` after changing contracts or queries. Frontend code uses TypeScript, React function components, Mantine UI, and PascalCase component/type names.

## Testing Guidelines

Backend tests use Go’s standard `testing` package. Place tests beside the code they exercise as `*_test.go`, prefer behavior-focused names, and run `make test` before submitting backend changes. For frontend changes, run `npm run typecheck --prefix frontend` or `make frontend-build`; no separate frontend test runner is configured.

## Commit & Pull Request Guidelines

Recent commits are short imperative messages such as `init`, `add step2`, and `step 4`. Keep commit subjects concise and focused on one change. Pull requests should describe the API, database, or UI impact; mention generated files when included; link related issues when available; and add screenshots for visible frontend changes.

## Configuration & Generated Files

`Makefile` includes `.env.example` if present, so keep local secrets out of committed files. Do not manually edit generated files in `internal/api/gen` or `internal/db`; change the OpenAPI/TypeSpec contract or SQL query source and regenerate instead.
