.PHONY: generate up seed test coverage

generate:
	oapi-codegen --config oapi-codegen.yaml api.yaml

up:
	docker-compose up --build

seed:
	docker-compose exec postgres psql -U postgres -d avito -f /docker-entrypoint-initdb.d/init.sql

test:
	go test ./...

# Покрытие без автосгенерированного кода (internal/api)
coverage:
	go test -coverpkg=./cmd/...,./internal/domain/...,./internal/handler/...,./internal/job/...,./internal/repository/...,./internal/service/...,./internal/testutil/... -coverprofile=cover.out ./...
	go tool cover -func=cover.out | tail -1
