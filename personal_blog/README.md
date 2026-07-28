# Personal Blog

A Go web application backed by PostgreSQL.

## Prerequisites

- Go 1.25 or newer
- Docker with Docker Compose

## Quick start with Docker Compose

Generate a bcrypt hash for the admin password:

```sh
docker run --rm httpd:2.4-alpine htpasswd -nbB admin 'change-me' | cut -d: -f2
```

Create `.env` in the project root and paste the generated hash. Keep the single
quotes: they prevent Docker Compose from interpreting the hash's `$` characters.

```dotenv
ADMIN_PASSWORD_HASH='$2y$05$replace.with.the.generated.hash'
```

Start PostgreSQL, run all migrations, build the application, and start it:

```sh
docker compose up --build
```

After the app reports that it has started, open <http://localhost:3030>. Sign in
at <http://localhost:3030/login> with username `admin123` and the password used
to generate the hash. Stop the services with `docker compose down`.

## Run locally

Start only PostgreSQL:

```sh
docker compose up -d db
```

Set the required application configuration:

```sh
export DATABASE_URL='postgres://blog:blog@localhost:5432/blog?sslmode=disable'
export ADMIN_LOGIN='admin'
export ADMIN_PASSWORD_HASH='$2y$05$replace.with.a.real.bcrypt.hash'
```

Run the migrations and start the app:

```sh
make migrate-up
make run
```

The local app is available at <http://localhost:8080>. Migration commands use
Goose at the version set by `GOOSE_VERSION` (default `v3.26.0`) and may download
it on first use. `make migrate-down` rolls back the most recent migration.

## Environment variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `DATABASE_URL` | Yes | None | PostgreSQL connection URL. Compose constructs it from the `POSTGRES_*` variables. |
| `ADMIN_LOGIN` | Yes | `admin123` in Compose | Administrator login. No default for a direct run. |
| `ADMIN_PASSWORD_HASH` | Yes | None | Bcrypt hash of the administrator password. |
| `PORT` | No | `8080` locally, `3030` in Compose | HTTP listen port. |
| `SESSION_TTL` | No | `24h` | Session lifetime as a positive Go duration, such as `30m` or `24h`. |
| `POSTGRES_DB` | Compose only | `blog` | Database created by the Compose `db` service. |
| `POSTGRES_USER` | Compose only | `blog` | Database user created by Compose. |
| `POSTGRES_PASSWORD` | Compose only | `blog` | Database password used by Compose. |
| `TEST_DATABASE_URL` | Integration tests only | None | Test PostgreSQL URL. Repository integration tests skip when unset. |

## Tests and checks

Run the full test suite, including PostgreSQL integration tests:

```sh
docker compose run --rm test
```

For the normal local checks (database tests skip if `TEST_DATABASE_URL` is not
set):

```sh
go test ./...
go vet ./...
go mod tidy
git diff --exit-code -- go.mod go.sum
```

GitHub Actions runs these checks on every push and pull request.
