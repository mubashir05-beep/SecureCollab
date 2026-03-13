# Local Development Runbook

## Prerequisites
- Devbox installed (see [SETUP.md](../SETUP.md) for installation)
- Docker Desktop running
- Go 1.22+ (provided by Devbox or manual install)
- Task (go-task) installed (provided by Devbox or manual install)

## First-Time Setup
If you haven't installed Devbox yet, see [docs/SETUP.md](../SETUP.md) for complete installation instructions.

Create a local environment file from template:
```bash
cp .env.example .env
```

## Bootstrap
```bash
devbox shell
task dev

# If port 8080 is already in use on your machine:
GATEWAY_PORT=8082 task dev
```

## Verify Services
```bash
# Gateway health
curl http://localhost:8080/healthz

# If you started with a custom gateway port:
curl http://localhost:8082/healthz

# Gateway metrics
curl http://localhost:8080/metrics

# Auth health
curl http://localhost:8081/healthz

# Key Distribution health
curl http://localhost:8082/healthz

# Messaging health
curl http://localhost:8083/healthz

# Analytics health
curl http://localhost:8084/healthz

# Analytics message-volume endpoint
curl http://localhost:8084/v1/analytics/messages/volume
```

## Test Auth API
```bash
# Register new user
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"secret"}'

# Login
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"secret"}'

# Refresh access token (replace <access_token>)
curl -X POST http://localhost:8081/refresh \
  -H "Authorization: Bearer <access_token>"
```

## Quality Checks
```bash
task lint
task test
task test:gateway

# Auth integration tests for DB store + HTTP handlers
task test:auth:integration

# Key Distribution integration test for key bootstrap flow
task test:keydist:integration

# Messaging integration test for encrypted payload send/inbox flow
task test:messaging:integration

# Analytics tests
task test:analytics

# Analytics integration test for Postgres-backed message volume
task test:analytics:integration
```

## Database Migrations
```bash
export DATABASE_URL="postgres://securecollab:securecollab@localhost:5432/securecollab?sslmode=disable"
task migrate:up
```

## Gateway Load Test
```bash
# Option A: validate protected route under real auth token
export GATEWAY_JWT="<access_token>"
task load-test

# Option B: run without token to validate unauthorized path stability
unset GATEWAY_JWT
task load-test
```

## Observability
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090
- Loki: http://localhost:3100

## CDC Bootstrap (Phase 3)
```bash
# Start CDC overlay services (Redpanda, Connect, ClickHouse)
task cdc:up

# Register the Postgres Debezium connector
task cdc:register

# Check connector status
task cdc:status

# Stop CDC overlay when done
task cdc:down
```

Detailed CDC steps: `docs/runbooks/cdc-local.md`

## Notes
- Gateway rate limiting uses Redis when `REDIS_ADDR` is set (for example `localhost:6379`).
- Gateway falls back to in-memory limiter if Redis is unavailable.
- Auth service uses PostgreSQL persistence when `DATABASE_URL` is set.
- Auth service falls back to in-memory user storage when `DATABASE_URL` is not set.
- Key Distribution service stores public keys in PostgreSQL when `DATABASE_URL` is set.
- Messaging service stores ciphertext/nonce in PostgreSQL when `DATABASE_URL` is set.
- Analytics service reads aggregated message volume from PostgreSQL when `DATABASE_URL` is set.
- Task commands load variables from repo `.env` automatically.

