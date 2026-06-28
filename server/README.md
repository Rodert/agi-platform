# AGI Platform Server

Backend service for AGI Platform.

## Stack

- Go
- Gin
- GORM
- MySQL

## Architecture

```text
cmd/server
  -> internal/app
  -> internal/router
  -> internal/handler
  -> internal/service
  -> internal/repository
  -> internal/model
```

Provider adapters are isolated under `internal/provider`.

The main dependency direction is:

```text
handler -> service -> repository -> database
service -> provider registry -> provider adapter
```

This keeps HTTP, persistence, and upstream image providers loosely coupled.
`service` does not import Gin or GORM. Database transactions are exposed through
the repository `TransactionManager` interface.

## Environment

```text
APP_ENV=local
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
JWT_SECRET=local-dev-secret-change-me
JWT_EXPIRE_SECONDS=604800
REGISTER_GIFT_CREDITS=0

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=
MYSQL_DATABASE=agi_platform
```

## Database

Run the schema and seed SQL from the repository root:

```text
database/mysql/001_schema.sql
database/mysql/002_seed.sql
```

The seed data creates:

- Demo user: `id = 1`
- Mock provider
- One enabled image model: `general-high-quality`

The default admin account is created from `config.yaml` or `ADMIN_*` environment variables on service startup. Existing admin passwords are not reset on every restart unless `ADMIN_RESET_PASSWORD_ON_STARTUP=true`.

## Run

```bash
go run -buildvcs=false ./cmd/server
```

Health check:

```bash
curl http://127.0.0.1:8080/health
```

Register:

```bash
curl -X POST http://127.0.0.1:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"user@example.com\",\"password\":\"secret123\",\"nickname\":\"User\"}"
```

Login:

```bash
curl -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"user@example.com\",\"password\":\"secret123\"}"
```

Generate mock image with JWT:

```bash
curl -X POST http://127.0.0.1:8080/api/images/generate \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d "{\"model\":\"general-high-quality\",\"prompt\":\"a cinematic AI poster\",\"size\":\"1024x1024\",\"n\":1}"
```

Create API Key:

```bash
curl -X POST http://127.0.0.1:8080/api/api-keys \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"local dev\"}"
```

OpenAI-style image generation with API Key:

```bash
curl -X POST http://127.0.0.1:8080/v1/images/generations \
  -H "Authorization: Bearer <api_key>" \
  -H "Content-Type: application/json" \
  -d "{\"model\":\"general-high-quality\",\"prompt\":\"a cinematic AI poster\",\"size\":\"1024x1024\",\"n\":1}"
```

Admin login:

```bash
curl -X POST http://127.0.0.1:8080/admin/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"admin123\"}"
```

Adjust user credits:

```bash
curl -X POST http://127.0.0.1:8080/admin/users/1/credits \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d "{\"amount\":100,\"remark\":\"manual bonus\"}"
```

## Current API

- `GET /health`
- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/models`
- `GET /api/me`
- `GET /api/api-keys`
- `POST /api/api-keys`
- `DELETE /api/api-keys/:id`
- `POST /api/images/generate`
- `GET /api/images/tasks/:task_no`
- `POST /v1/images/generations`
- `POST /admin/auth/login`
- `GET /admin/me`
- `GET /admin/users`
- `POST /admin/users/:id/credits`
- `GET /admin/providers`
- `POST /admin/providers`
- `PUT /admin/providers/:id`
- `GET /admin/providers/:id/keys`
- `POST /admin/providers/:id/keys`
- `DELETE /admin/provider-keys/:id`
- `GET /admin/image-models`
- `POST /admin/image-models`
- `PUT /admin/image-models/:id`
- `GET /admin/image-models/:id/routes`
- `POST /admin/image-models/:id/routes`
- `PUT /admin/image-model-routes/:id`
- `GET /admin/image-tasks`
