# Gopi Go Backend API

A production-ready Go backend that replaces the prior Python/Django API. It provides REST endpoints, WebSocket chat, JWT auth, storage (local/S3), email, Redis-backed features, and Swagger documentation.

## Features

- **HTTP API**: Gin-powered REST endpoints (see Swagger for full contract)
- **WebSocket chat**: Realtime group chat endpoint
- **Auth**: JWT Bearer tokens via `Authorization: Bearer <token>`
- **Docs**: Swagger UI at `/swagger/index.html` and JSON at `/doc.json`
- **Data**: GORM with SQLite/MySQL drivers and automatic migrations
- **Redis**: Token blacklist and password reset flows
- **Email**: SMTP via gomail
- **Storage**: Local uploads or S3-compatible object storage
- **Logging**: Structured logs with slog and rotating file logs via lumberjack

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

### Option 1: Native Go Development

- Go 1.25+
- Redis (recommended; used for JWT blacklist and password reset)
- Database:
  - SQLite (default, embedded)
  - MySQL/PostgreSQL (driver included; can be wired via config and code paths)

### Option 2: Docker Development

- Docker and Docker Compose
- No additional dependencies needed (everything runs in containers)

## Quick start

1. Create your env file:

```
cp .env.example .env
```

2. Configure minimal env in `.env` (example):

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

3. Build and run:

```
make build
./bin/go-server
```

Or use hot reload (auto-installs Air locally if missing):

```
make dev
```

4. Open the API docs:

- Swagger UI: http://localhost:8080/swagger/index.html
- Swagger JSON: http://localhost:8080/doc.json

5. Health check:

```
curl -s http://localhost:8080/health | jq
```

## Docker Setup

For easier development and deployment, use Docker and Docker Compose:

### Quick Docker Start

1. **Start all services:**

```bash
make docker-compose-up
```

This will start:

- ✅ GoPadi backend (port 8080)
- ✅ PostgreSQL database (port 5432)
- ✅ Redis cache (port 6379)
- ✅ Adminer database GUI (port 8081)

2. **Check services:**

```bash
make docker-compose-logs
```

3. **Stop services:**

```bash
make docker-compose-down
```

### Docker Commands

```bash
# Build Docker image
make docker-build

# Run single container (without database)
make docker-run

# View container logs
make docker-logs

# Stop container
make docker-stop

# Clean Docker artifacts
make docker-clean
```

### Docker Environment

The Docker setup includes:

- **Database**: PostgreSQL 15 with automatic initialization
- **Cache**: Redis 7 for JWT blacklist and password reset
- **Admin Interface**: Adminer web UI for database management
- **Volume Mounts**: Persistent logs and uploads directories
- **Health Checks**: Automatic service health monitoring

### Accessing Services

- **API**: http://localhost:8080
- **Swagger**: http://localhost:8080/swagger/index.html
- **Adminer**: http://localhost:8081 (Database GUI)
- **PostgreSQL**: localhost:5432 (from host machine)
- **Redis**: localhost:6379 (from host machine)

### Docker Configuration

The `docker-compose.yml` includes optimized settings:

- Health checks for service dependencies
- Proper restart policies
- Volume persistence for data
- Network isolation
- Environment variable configuration

## Usage Scenarios

### Development Scenarios

#### 1. Local Go Development (Recommended for Contributors)

**Best for:** Active development, debugging, and code contributions

```bash
# 1. Clone and setup
git clone <repository>
cd gopi
cp .env.example .env

# 2. Configure local environment
# Edit .env with your local settings
# - SQLite for simplicity
# - Local Redis (brew install redis)
# - Local email logging

# 3. Install dependencies
go mod download

# 4. Run with hot reload
make dev

# 5. Access services
# - API: http://localhost:8080
# - Swagger: http://localhost:8080/swagger/index.html
# - Health: http://localhost:8080/health
```

**Pros:**

- ✅ Fastest development cycle
- ✅ Full debugging capabilities
- ✅ Easy code hot-reload
- ✅ Direct access to logs

**Cons:**

- ❌ Requires local Go installation
- ❌ Manual dependency management
- ❌ Platform-specific setup

#### 2. Docker Development (Recommended for Teams)

**Best for:** Team development, consistent environments, onboarding

```bash
# 1. Setup environment
cp docker-env.example .env

# 2. Start full development stack
make docker-compose-up

# 3. Access services
# - API: http://localhost:8080
# - Swagger: http://localhost:8080/swagger/index.html
# - Adminer (DB GUI): http://localhost:8081
# - PostgreSQL: localhost:5432
# - Redis: localhost:6379

# 4. View logs
make docker-compose-logs

# 5. Stop services
make docker-compose-down
```

**Pros:**

- ✅ Zero local setup required
- ✅ Consistent across all machines
- ✅ Full service stack included
- ✅ Easy onboarding for new developers

**Cons:**

- ❌ Slightly slower than native Go
- ❌ More resource intensive
- ❌ Less debugging capabilities

#### 3. Hybrid Development (Recommended for Advanced Users)

**Best for:** Performance-focused development with containerized dependencies

```bash
# 1. Run only database and cache in Docker
docker-compose up db redis adminer -d

# 2. Run Go app natively
make dev

# 3. Access services
# - API: http://localhost:8080 (native Go)
# - Adminer: http://localhost:8081 (Docker)
# - PostgreSQL: localhost:5432 (Docker)
# - Redis: localhost:6379 (Docker)
```

**Pros:**

- ✅ Fast Go development
- ✅ Containerized dependencies
- ✅ Best of both worlds
- ✅ Easy database management

**Cons:**

- ❌ More complex setup
- ❌ Network configuration needed

### Testing Scenarios

#### 1. Unit Testing

```bash
# Run all unit tests
make test

# Run specific package tests
make test-user

# Run with coverage
make test-coverage

# Run with race detection
make test-race
```

#### 2. Integration Testing

```bash
# Run integration tests (requires Docker)
make test-integration

# Run in Docker environment
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

#### 3. End-to-End Testing

```bash
# Start services
make docker-compose-up

# Run E2E tests against running services
npm test  # Assuming frontend tests

# Or use tools like Postman/Newman
newman run api-tests.postman_collection.json
```

### Production Deployment Scenarios

#### 1. Docker Compose Production

```bash
# 1. Update docker-compose.yml for production
# - Change secrets and passwords
# - Update ports if needed
# - Configure proper logging
# - Add reverse proxy (nginx/traefik)

# 2. Use production environment file
cp docker-env.prod .env

# 3. Deploy
docker-compose -f docker-compose.prod.yml up -d

# 4. Setup monitoring
# - Health checks
# - Log aggregation
# - Metrics collection
```

#### 2. Kubernetes Deployment

```yaml
# Use Kubernetes manifests
kubectl apply -f k8s/
# Services include:
# - GoPadi deployment
# - PostgreSQL statefulset
# - Redis deployment
# - Ingress configuration
# - ConfigMaps and Secrets
```

#### 3. Cloud Platform Deployment

**AWS:**

```bash
# Use AWS Fargate or ECS
aws ecs create-service --service-name gopi-backend \
  --task-definition gopi-task \
  --desired-count 2
```

**Google Cloud:**

```bash
# Use Cloud Run
gcloud run deploy gopi-backend \
  --source . \
  --platform managed \
  --port 8080
```

### CI/CD Scenarios

#### 1. GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.25"
      - run: make test
      - run: make docker-build

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - run: make docker-compose-up
```

#### 2. GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  image: golang:1.25-alpine
  script:
    - make test

build:
  stage: build
  script:
    - make docker-build
    - docker tag gopi-backend registry.gitlab.com/project/gopi-backend:$CI_COMMIT_SHA

deploy:
  stage: deploy
  script:
    - docker-compose pull
    - docker-compose up -d
```

### Environment-Specific Scenarios

#### 1. Development Environment

**Configuration Focus:**

- Local email logging (`USE_LOCAL_EMAIL=true`)
- Debug logging (`LOG_LEVEL=debug`)
- SQLite database for simplicity
- Hot reload enabled

**Sample .env:**

```bash
GIN_MODE=debug
LOG_LEVEL=debug
USE_LOCAL_EMAIL=true
DB_DRIVER=sqlite
```

#### 2. Staging Environment

**Configuration Focus:**

- Production-like settings
- External email service
- PostgreSQL database
- Enhanced logging and monitoring

**Sample .env:**

```bash
GIN_MODE=release
LOG_LEVEL=info
USE_LOCAL_EMAIL=false
DB_DRIVER=postgres
```

#### 3. Production Environment

**Configuration Focus:**

- Maximum security
- External services
- Monitoring and alerting
- Performance optimization

**Sample .env:**

```bash
GIN_MODE=release
LOG_LEVEL=warn
USE_LOCAL_EMAIL=false
DB_DRIVER=postgres
# Additional production settings...
```

### Troubleshooting Scenarios

#### 1. Database Connection Issues

```bash
# Check if database is running
make docker-compose-logs db

# Test database connection
docker-compose exec db pg_isready -U gopi_user -d gopi_db

# Reset database
make docker-compose-down
docker volume rm gopi_postgres_data
make docker-compose-up
```

#### 2. Email Not Working

```bash
# Check email configuration
echo $USE_LOCAL_EMAIL
echo $EMAIL_LOG_PATH

# Test email service
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"test","password":"password123","first_name":"Test","last_name":"User","height":175,"weight":70}'

# Check email logs
tail -f logs/emails.log
```

#### 3. Container Issues

```bash
# Check container status
docker-compose ps

# View container logs
docker-compose logs app

# Restart specific service
docker-compose restart app

# Debug container
docker-compose exec app sh
```

## API overview

- **Base HTTP server**: see route wiring in `api/http/router/router.go`
- **Swagger**: UI at `/swagger/index.html`, JSON at `/doc.json`
- **Health**: `GET /health`
- **Uploads**: static files served from `/uploads` (local storage)
- **Chat WebSocket**: `GET /ws/chat/groups/{groupSlug}`
  - Headers: `Authorization: Bearer <JWT>`

For the complete list of endpoints, request/response schemas, auth requirements, and tags, consult the Swagger UI.

## Configuration

All configuration is centralized in `config/env.go`. Key variables:

- **Server**

  - `PUBLIC_HOST` — external host used in links/emails
  - `PORT` — server port (default `8080`)
  - `GIN_MODE` — `debug` or `release`

- **Database**

  - `DB_DRIVER` — `sqlite` (default) or `mysql`
  - `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` — used primarily for MySQL

- **Logging**

  - `LOG_LEVEL` — e.g., `debug`, `info`
  - `LOG_FILE_ENABLED` — `true|false` (write logs to file)
  - `LOG_FILE` — e.g., `logs/app.log`

- **JWT**

  - `JWT_SECRET` — signing key for tokens

- **Sessions (optional)**

  - `SESSION_SECRET`, `SESSION_NAME`, `SESSION_SECURE`, `SESSION_DOMAIN`, `SESSION_MAX_AGE`

- **Redis**

  - `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`

- **Email**

  - `EMAIL_HOST`, `EMAIL_PORT`, `EMAIL_USERNAME`, `EMAIL_PASSWORD`, `EMAIL_FROM`

- **Storage**
  - `STORAGE_BACKEND` — `local` (default) or `s3`
  - Local: `UPLOAD_BASE_DIR` (default `./uploads`), `UPLOAD_PUBLIC_BASE_URL` (default `/uploads`)
  - S3: `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_USE_SSL`, `S3_FORCE_PATH_STYLE`, `S3_PUBLIC_BASE_URL`

## Development workflow

- **Hot reload**: `make dev` (runs Air; installs to `./tmp/bin` if needed)
- **Generate Swagger**: `make swag` (installs `swag` locally if needed)
- **Build**: `make build`
- **Run**: `make run` (after build)
- **Clean**: `make clean`

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

- **Swagger not reflecting changes**: run `make swag` to regenerate docs in `docs/`
- **CORS errors**: update allowed origins in `api/http/router/router.go`
- **Redis issues**: verify `REDIS_ADDR` and network; password may be empty for local
- **Uploads missing**: ensure `UPLOAD_BASE_DIR` exists and your process has write perms

## License

Proprietary. All rights reserved (or update this section with your actual license).

Windows Packaging in Makefile
New Makefile targets:
✅ make build-windows - Cross-compile for Windows 64-bit
✅ make package-windows - Build + create complete package
✅ make clean-windows - Clean Windows build artifacts
Package contents (auto-generated):
✅ gopi-windows.exe - Main executable
✅ env.example.txt - Config template with EMAIL_LOG_PATH
✅ README_Windows.txt - Complete setup instructions
✅ start-server.bat - Easy Windows startup script
✅ gopi-backend-windows.zip - Ready-to-deliver package
