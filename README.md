# Gopi Go Backend API

A production-ready Go backend that replaces the prior Python/Django API. It provides REST endpoints, WebSocket chat, JWT auth, storage (local/S3), email, Redis-backed features, and Swagger documentation.

## Features
- __HTTP API__: Gin-powered REST endpoints (see Swagger for full contract)
- __WebSocket chat__: Realtime group chat endpoint
- __Auth__: JWT Bearer tokens via `Authorization: Bearer <token>`
- __Docs__: Swagger UI at `/swagger/index.html` and JSON at `/doc.json`
- __Data__: GORM with SQLite/MySQL drivers and automatic migrations
- __Redis__: Token blacklist and password reset flows
- __Email__: SMTP via gomail
- __Storage__: Local uploads or S3-compatible object storage
- __Logging__: Structured logs with slog and rotating file logs via lumberjack

## Project structure
- `cmd/api/main.go` — server bootstrap, DI wiring, migrations, Swagger metadata
- `api/http/` — HTTP handlers, middleware, and route registration
- `api/http/router/router.go` — Gin engine, CORS, health, Swagger, route wiring
- `api/ws/chat_handler.go` — WebSocket chat handler and endpoint annotations
- `internal/app/` — domain services (user, campaign, challenge, chat, post)
- `internal/data/` — GORM models and repositories per domain
- `internal/db/db.go` — DB factories for SQLite and MySQL
- `internal/lib/` — shared libs (jwt, email, storage, password reset)
- `config/env.go` — centralized environment configuration
- `docs/` — generated Swagger docs (JSON, YAML, docs.go)
- `Makefile`, `.air.toml` — dev tooling (build, run, hot reload, swagger)

## Prerequisites
- Go 1.25+
- Redis (recommended; used for JWT blacklist and password reset)
- Database:
  - SQLite (default, embedded)
  - MySQL (driver included; can be wired via config and code paths)

## Quick start
1) Create your env file:
```
cp .env.example .env
```

2) Configure minimal env in `.env` (example):
```
PUBLIC_HOST=http://localhost
PORT=8080
GIN_MODE=debug

# Auth
JWT_SECRET=dev-jwt-secret-change-me

# Redis
REDIS_ADDR=localhost:6379
REDIS_DB=0

# Email (optional for password reset)
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=
EMAIL_PASSWORD=
EMAIL_FROM=noreply@gopadi.com

# Storage (local or s3)
STORAGE_BACKEND=local
UPLOAD_BASE_DIR=./uploads
UPLOAD_PUBLIC_BASE_URL=/uploads
```

3) Build and run:
```
make build
./bin/go-server
```

Or use hot reload (auto-installs Air locally if missing):
```
make dev
```

4) Open the API docs:
- Swagger UI: http://localhost:8080/swagger/index.html
- Swagger JSON: http://localhost:8080/doc.json

5) Health check:
```
curl -s http://localhost:8080/health | jq
```

## API overview
- __Base HTTP server__: see route wiring in `api/http/router/router.go`
- __Swagger__: UI at `/swagger/index.html`, JSON at `/doc.json`
- __Health__: `GET /health`
- __Uploads__: static files served from `/uploads` (local storage)
- __Chat WebSocket__: `GET /ws/chat/groups/{groupSlug}`
  - Headers: `Authorization: Bearer <JWT>`

For the complete list of endpoints, request/response schemas, auth requirements, and tags, consult the Swagger UI.

## Configuration
All configuration is centralized in `config/env.go`. Key variables:

- __Server__
  - `PUBLIC_HOST` — external host used in links/emails
  - `PORT` — server port (default `8080`)
  - `GIN_MODE` — `debug` or `release`

- __Database__
  - `DB_DRIVER` — `sqlite` (default) or `mysql`
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` — used primarily for MySQL

- __Logging__
  - `LOG_LEVEL` — e.g., `debug`, `info`
  - `LOG_FILE_ENABLED` — `true|false` (write logs to file)
  - `LOG_FILE` — e.g., `logs/app.log`

- __JWT__
  - `JWT_SECRET` — signing key for tokens

- __Sessions (optional)__
  - `SESSION_SECRET`, `SESSION_NAME`, `SESSION_SECURE`, `SESSION_DOMAIN`, `SESSION_MAX_AGE`

- __Redis__
  - `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`

- __Email__
  - `EMAIL_HOST`, `EMAIL_PORT`, `EMAIL_USERNAME`, `EMAIL_PASSWORD`, `EMAIL_FROM`

- __Storage__
  - `STORAGE_BACKEND` — `local` (default) or `s3`
  - Local: `UPLOAD_BASE_DIR` (default `./uploads`), `UPLOAD_PUBLIC_BASE_URL` (default `/uploads`)
  - S3: `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_USE_SSL`, `S3_FORCE_PATH_STYLE`, `S3_PUBLIC_BASE_URL`

## Development workflow
- __Hot reload__: `make dev` (runs Air; installs to `./tmp/bin` if needed)
- __Generate Swagger__: `make swag` (installs `swag` locally if needed)
- __Build__: `make build`
- __Run__: `make run` (after build)
- __Clean__: `make clean`

CORS defaults to allow `http://localhost:3000` and common methods/headers. Adjust in `api/http/router/router.go` for your frontend origin(s).

## Data & migrations
`cmd/api/main.go` performs AutoMigrate across core models (user, challenge, campaign, chat, post). Drivers for SQLite and MySQL are included via GORM. For production, ensure your chosen DB and schema lifecycle match your operational needs.

## Authentication
- Bearer tokens via `Authorization: Bearer <JWT>`
- Swagger security: defined as `Bearer` apiKey in headers
- Password reset flow uses Redis TTL to manage reset tokens

## Storage
- Local storage: files written under `./uploads` and served at `/uploads`
- S3 storage: configure S3 envs; URLs can be exposed via `S3_PUBLIC_BASE_URL`

## Logging
- Structured logging via Go `slog`
- File rotation via `lumberjack` when `LOG_FILE_ENABLED=true` (path from `LOG_FILE`)

## Troubleshooting
- __Swagger not reflecting changes__: run `make swag` to regenerate docs in `docs/`
- __CORS errors__: update allowed origins in `api/http/router/router.go`
- __Redis issues__: verify `REDIS_ADDR` and network; password may be empty for local
- __Uploads missing__: ensure `UPLOAD_BASE_DIR` exists and your process has write perms

## License
Proprietary. All rights reserved (or update this section with your actual license).
