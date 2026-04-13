# Taskflow

A REST API for managing projects and tasks, built with Go.

---

## Overview

Taskflow is a JSON REST API that lets authenticated users create projects, manage tasks within those projects, and assign tasks to other users. It is a backend-only service — there is no frontend.

**Tech Stack**

| Layer | Choice |
|---|---|
| Language | Go 1.24 |
| HTTP router | [chi v5](https://github.com/go-chi/chi) |
| Database | PostgreSQL 16 |
| DB driver | [pgx v5](https://github.com/jackc/pgx) with connection pooling |
| Query layer | [sqlc](https://sqlc.dev) (type-safe generated code from raw SQL) |
| Authentication | HS256 JWT via [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) |
| Password hashing | bcrypt (DefaultCost) via `golang.org/x/crypto` |
| Config | Environment variables via [caarlos0/env](https://github.com/caarlos0/env) + `.env` file |
| Logging | Structured JSON via stdlib `log/slog` |
| Containerisation | Docker + Docker Compose |
| Task runner | [just](https://github.com/casey/just) |

---

## Architecture Decisions

### Layered architecture

The code is split into four layers with strict one-way dependencies:

```
delivery/http        ← HTTP concerns: routing, parsing, response shaping
internal/usecase     ← Business logic and authorization
internal/repository  ← Database interface (one per domain)
repository/driver    ← sqlc-generated Postgres queries
```

Each layer communicates only through the domain interfaces defined in `internal/domain/`. Nothing in `usecase` imports `delivery`; nothing in `repository` imports `usecase`. This makes the business logic easy to test in isolation and swappable at every boundary.

### sqlc over an ORM

All queries are hand-written SQL in `internal/repository/*/sql/query.sql`. sqlc generates type-safe Go wrappers from them. This keeps the SQL fully visible and auditable, avoids the hidden query generation of ORMs, and makes it straightforward to write efficient queries (e.g. the DISTINCT JOIN used for listing projects a user is involved in).

### No migration framework

Migrations are plain numbered `.sql` files applied manually via `psql`. This is intentional for a small project — no extra dependency, nothing to learn. The trade-off is no rollback tooling and no migration state tracking (though the SQL uses `IF NOT EXISTS` / `IF EXISTS` guards to be idempotent).

### Ownership-based authorization

Authorization is kept simple: only the project `owner_id` can modify or delete a project. Tasks can be modified or deleted by the project owner or the task's `assignee_id`. There are no roles or team memberships, which would be the natural next step.

### What was intentionally left out

- **No refresh tokens.** The JWT is long-lived (`JWT_EXPIRATION`, default 24 h). Refresh token rotation adds significant complexity and was out of scope.
- **No tests.** The time constraint made comprehensive test coverage impractical. See [What You'd Do With More Time](#what-youd-do-with-more-time).
- **No rate limiting or request size limits.** Both would be important before exposing this to the public internet.
- **No pagination metadata in responses.** Paginated list endpoints accept `page` and `page_size` query params but do not return total counts or next-page links.

---

## Running Locally

**Prerequisites:** Docker (with the Compose plugin). Nothing else is required.

```bash
git clone https://github.com/sovajitroy-fam/taskflow-SovajitRoy
cd taskflow-SovajitRoy
cp .env.example .env
docker compose up --build
```

The database schema is applied automatically on first start (the init SQL is mounted into the Postgres container). The API is available at `http://localhost:8080`.

> **Note:** The `api` container may log a connection error on the first attempt while Postgres is still initialising. Docker Compose will restart it automatically and it will connect once the database is ready.

To load seed data (optional — see [Test Credentials](#test-credentials) below):

```bash
# In a separate terminal, once the stack is up
docker compose cp migrations/seed.sql db:/tmp/seed.sql
docker compose exec db psql -U postgres -d taskflow -f /tmp/seed.sql
```

Verify the API is up:

```bash
curl http://localhost:8080/health
```

---

## Running Migrations

Migrations run automatically when the database container starts for the first time (via `docker-entrypoint-initdb.d`). If you need to apply them manually against a running database:

```bash
# Requires psql installed locally and the DB reachable
just migrate

# Or directly:
psql "postgresql://postgres:postgres@localhost:5432/taskflow?sslmode=disable" \
  -f migrations/001_init.up.sql
```

---

## Test Credentials

Run `just seed` (requires `psql` locally and the DB running) or the Docker command in [Running Locally](#running-locally) to load the seed data, then log in with:

```
Email:    test@example.com
Password: password123
```

---

## API Reference

### Conventions

- All request and response bodies are JSON.
- All authenticated endpoints require an `Authorization: Bearer <token>` header. Obtain a token from `/auth/register` or `/auth/login`.
- Success responses are enveloped as `{ "code": "SUCCESS", "data": { ... } }`.
- Error responses are enveloped as `{ "error_code": "TF-XXX-NN", "message": "...", "data": { ... } }`.

---

### Health

#### `GET /health`

```bash
curl http://localhost:8080/health
```

```json
{
  "code": "SUCCESS",
  "data": { "status": "ok", "service": "taskflow", "timestamp": 1713000000 }
}
```

---

### Auth

#### `POST /auth/register`

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "email": "alice@example.com", "password": "secret123"}'
```

**Request**

| Field | Type | Rules |
|---|---|---|
| `name` | string | required, min 2 chars |
| `email` | string | required, valid email |
| `password` | string | required, min 6 chars |

**Response `201`**

```json
{
  "code": "SUCCESS",
  "data": {
    "token": "eyJ...",
    "user": {
      "id": "uuid",
      "name": "Alice",
      "email": "alice@example.com",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

**Errors:** `409` email already registered · `400` validation failure

---

#### `POST /auth/login`

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "secret123"}'
```

**Request**

| Field | Type | Rules |
|---|---|---|
| `email` | string | required, valid email |
| `password` | string | required |

**Response `200`** — same shape as `/auth/register`.

**Errors:** `401` wrong credentials · `400` validation failure

---

### Projects

All project endpoints require authentication.

#### `GET /v1/projects`

Returns projects where the caller is the owner or is assigned to at least one task.

**Query params:** `page` (default 1) · `page_size` (default 20, max 100)

```bash
curl http://localhost:8080/v1/projects \
  -H "Authorization: Bearer <token>"
```

**Response `200`**

```json
{
  "code": "SUCCESS",
  "data": {
    "projects": [
      {
        "id": "uuid",
        "name": "Demo Project",
        "description": "A sample project",
        "owner_id": "uuid",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

---

#### `POST /v1/projects`

```bash
curl -X POST http://localhost:8080/v1/projects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Project", "description": "Optional description"}'
```

**Request**

| Field | Type | Rules |
|---|---|---|
| `name` | string | required, min 1 char |
| `description` | string | optional |

**Response `201`** — single project object.

---

#### `GET /v1/projects/{id}`

```bash
curl http://localhost:8080/v1/projects/<project-uuid> \
  -H "Authorization: Bearer <token>"
```

**Response `200`** — single project object.

**Errors:** `404` not found

---

#### `PATCH /v1/projects/{id}`

All fields are optional. Only supplied non-null fields are updated.

```bash
curl -X PATCH http://localhost:8080/v1/projects/<project-uuid> \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Renamed Project"}'
```

**Request**

| Field | Type |
|---|---|
| `name` | string \| null |
| `description` | string \| null |

**Response `200`** — updated project object.

**Errors:** `403` not the owner · `404` not found

---

#### `DELETE /v1/projects/{id}`

Cascade-deletes all tasks in the project.

```bash
curl -X DELETE http://localhost:8080/v1/projects/<project-uuid> \
  -H "Authorization: Bearer <token>"
```

**Response `204`** — no body.

**Errors:** `403` not the owner · `404` not found

---

### Tasks

All task endpoints require authentication.

#### `GET /v1/projects/{projectID}/tasks`

**Query params:**

| Param | Type | Description |
|---|---|---|
| `status` | string | Filter by `todo`, `in_progress`, or `done` |
| `assignee` | UUID | Filter by assignee |
| `page` | int | Default 1 |
| `page_size` | int | Default 20, max 100 |

```bash
curl "http://localhost:8080/v1/projects/<project-uuid>/tasks?status=in_progress" \
  -H "Authorization: Bearer <token>"
```

**Response `200`**

```json
{
  "code": "SUCCESS",
  "data": {
    "tasks": [
      {
        "id": "uuid",
        "project_id": "uuid",
        "title": "Implement authentication",
        "status": "in_progress",
        "priority": "high",
        "assignee_id": "uuid",
        "due_date": "2024-06-01",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

`assignee_id` and `due_date` are omitted when null.

---

#### `POST /v1/projects/{projectID}/tasks`

New tasks always start with `status: "todo"`.

```bash
curl -X POST http://localhost:8080/v1/projects/<project-uuid>/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Write tests",
    "priority": "high",
    "assignee_id": "<user-uuid>",
    "due_date": "2024-06-01"
  }'
```

**Request**

| Field | Type | Rules |
|---|---|---|
| `title` | string | required, min 1 char |
| `priority` | string | `low` \| `medium` \| `high` · default `medium` |
| `assignee_id` | UUID string | optional |
| `due_date` | `YYYY-MM-DD` | optional |

**Response `201`** — single task object.

**Errors:** `404` project not found

---

#### `PATCH /v1/tasks/{id}`

All fields are optional. Only supplied non-null fields are updated.

```bash
curl -X PATCH http://localhost:8080/v1/tasks/<task-uuid> \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status": "done"}'
```

**Request**

| Field | Type |
|---|---|
| `title` | string \| null |
| `status` | `todo` \| `in_progress` \| `done` \| null |
| `priority` | `low` \| `medium` \| `high` \| null |
| `assignee_id` | UUID string \| null |
| `due_date` | `YYYY-MM-DD` \| null |

**Response `200`** — updated task object.

**Errors:** `403` not the project owner or assignee · `404` not found

---

#### `DELETE /v1/tasks/{id}`

```bash
curl -X DELETE http://localhost:8080/v1/tasks/<task-uuid> \
  -H "Authorization: Bearer <token>"
```

**Response `204`** — no body.

**Errors:** `403` not the project owner or assignee · `404` not found

---

## What You'd Do With More Time

**Tests.** There are none. The layered architecture with domain interfaces was designed to make testing straightforward — usecases can be tested against a fake repository, handlers can be tested with `httptest` — but I ran out of time to write them. This is the biggest gap.

**Proper migration tooling.** The current approach (numbered SQL files run via `psql`) works but has no state tracking. A tool like [goose](https://github.com/pressly/goose) or [atlas](https://atlasgo.io) would add rollback support and make it safe to run migrations in CI.

**Docker startup ordering.** The `api` service has no healthcheck on the `db` service, so it can fail on the first connection attempt. A proper `healthcheck` in `docker-compose.yml` with `depends_on: condition: service_healthy` would fix this cleanly.

**Dockerfile Go version mismatch.** The builder stage uses `golang:1.23-alpine` but `go.mod` declares `go 1.24.0`. It should be `golang:1.24-alpine`.

**Pagination metadata.** Paginated endpoints accept `page` and `page_size` but return no total count or next-page cursor. Consumers have no way to know when they've reached the last page without getting an empty response.

**Refresh tokens.** The current JWT is long-lived and cannot be revoked. Short-lived access tokens paired with refresh tokens stored server-side would be the right fix.

**Rate limiting and input size caps.** There is no protection against large payloads or request flooding.

**Roles / team membership.** The ownership model is intentionally minimal. A real product would need at least viewer/editor/admin roles per project.
