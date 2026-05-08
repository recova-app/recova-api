.PHONY: preflight fmt lint test test-integration test-e2e test-performance release-validation build run migrate-up migrate-down migrate-status migrate-check migrate-force seed openapi-generate openapi-check security-scan compose-smoke staging-deploy cutover-wave cutover-all stabilization-gate rollback-rehearsal runtime-decommission post-migration-maintenance

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

test-e2e:
	@echo "[test-e2e] running critical flow e2e suite"
	@./scripts/e2e-critical.sh

test-performance:
	@echo "[test-performance] running performance smoke suite"
	@./scripts/performance-smoke.sh

release-validation: test-e2e test-performance

build:
	@echo "[build] building api binary"
	@mkdir -p bin
	@go build -o ./bin/recova-api ./cmd/api

run:
	@echo "[run] starting api"
	@./scripts/with-env.sh go run ./cmd/api

migrate-up:
	@./scripts/with-env.sh ./scripts/migrate.sh up

migrate-down:
	@./scripts/with-env.sh ./scripts/migrate.sh down 1

migrate-status:
	@./scripts/with-env.sh ./scripts/migrate.sh status

migrate-check:
	@./scripts/with-env.sh ./scripts/migrate.sh check

migrate-force:
	@./scripts/with-env.sh ./scripts/migrate.sh force $(VERSION)

seed:
	@./scripts/with-env.sh ./scripts/seed.sh

openapi-generate:
	@./scripts/openapi.sh generate

openapi-check:
	@./scripts/openapi.sh check

security-scan:
	@./scripts/security-scan.sh ./...

compose-smoke:
	@./scripts/compose-smoke.sh

staging-deploy:
	@./scripts/staging-deploy.sh

cutover-wave:
	@./scripts/cutover-wave.sh $(WAVE)

cutover-all:
	@./scripts/cutover-wave.sh all

stabilization-gate:
	@./scripts/stabilization-gate.sh

rollback-rehearsal:
	@./scripts/rollback-rehearsal.sh

runtime-decommission:
	@./scripts/runtime-decommission.sh

post-migration-maintenance:
	@./scripts/post-migration-maintenance.sh
