GOOSE_VERSION ?= v3.26.0
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)
MIGRATIONS_DIR ?= db/migrations

.PHONY: run test vet migrate-up migrate-down

run:
	go run ./cmd/blog

test:
	go test ./...

vet:
	go vet ./...

migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" >&2; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" >&2; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down
