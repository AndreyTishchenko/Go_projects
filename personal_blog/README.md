# Personal Blog

A Go web application backed by PostgreSQL.

## Environment variables

The application reads configuration from the process environment. Docker Compose
also reads the variables in the project-level `.env` file when interpolating
`docker-compose.yml`.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `PORT` | No | `8080` when running the application directly; `3030` with Docker Compose | HTTP port on which the application listens. Docker Compose publishes the same port on the host. |
| `DATABASE_URL` | Yes | None | PostgreSQL connection string used by the application. For a direct run, set the complete URL, for example `postgres://blog:blog@localhost:5432/blog?sslmode=disable`. Docker Compose constructs this value from the `POSTGRES_*` variables. |
| `ADMIN_LOGIN` | Yes | `admin123` with Docker Compose | Login name for the blog administrator. There is no default when running the application directly. |
| `ADMIN_PASSWORD` | Yes | `admin` with Docker Compose | Password for the blog administrator. There is no default when running the application directly. Change the Compose default outside local development. |
| `POSTGRES_DB` | Docker Compose only | `blog` | Name of the PostgreSQL database created by the `db` service and used to construct `DATABASE_URL`. |
| `POSTGRES_USER` | Docker Compose only | `blog` | PostgreSQL user created by the `db` service and used to construct `DATABASE_URL`. |
| `POSTGRES_PASSWORD` | Docker Compose only | `blog` | Password for `POSTGRES_USER`, also used to construct `DATABASE_URL`. Change the default outside local development. |
| `TEST_DATABASE_URL` | Database tests only | None | PostgreSQL connection string used by repository integration tests. Those tests are skipped when it is unset. The Compose `test` service supplies it automatically. |

Example configuration for running the application directly:

```dotenv
PORT=8080
DATABASE_URL=postgres://blog:blog@localhost:5432/blog?sslmode=disable
ADMIN_LOGIN=admin
ADMIN_PASSWORD=replace-me
```

Start the application and its database with the Compose defaults:

```sh
docker compose up --build
```

Override any Compose default by exporting the variable or placing it in `.env`
before running the command. Run the full test suite, including PostgreSQL
integration tests, with:

```sh
docker compose run --rm test
```
