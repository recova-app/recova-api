.PHONY: preflight fmt lint test test-integration build run migrate-up migrate-down

preflight:
	@./scripts/preflight.sh

fmt:
	@echo "[fmt] running gofmt"
	@gofmt -w .

lint:
	@echo "[lint] running go vet"
	@go vet ./...

test:
	@echo "[test] running go test"
	@go test ./...

test-integration:
	@echo "[test-integration] running shell integration checks"
	@./test/integration/scripts_test.sh

build:
	@echo "[build] building api binary"
	@mkdir -p bin
	@go build -o ./bin/recova-api ./cmd/api

run:
	@echo "[run] starting api"
	@go run ./cmd/api

migrate-up:
	@./scripts/migrate.sh up

migrate-down:
	@./scripts/migrate.sh down 1
