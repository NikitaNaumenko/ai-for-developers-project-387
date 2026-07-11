-include .env.example
export

APP_NAME := calendar-api

.PHONY: tools generate test run build frontend-install frontend-dev frontend-build frontend-e2e typespec-install typespec-build typespec-check mock-api

tools:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.0
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

generate:
	oapi-codegen --config oapi-codegen.yaml api/openapi.yaml
	sqlc generate

test:
	go test ./...

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

frontend-install:
	npm install --prefix frontend

frontend-dev:
	npm run dev --prefix frontend

frontend-build:
	npm run build --prefix frontend

frontend-e2e:
	npm run test:e2e --prefix frontend

typespec-install:
	npm install --prefix api/typespec

typespec-build:
	npm run build --prefix api/typespec

typespec-check:
	npm run check --prefix api/typespec

mock-api:
	npm run mock --prefix api/typespec
