# Local Development Runbook

## Prerequisites
- Devbox installed (see [SETUP.md](../SETUP.md) for installation)
- Docker Desktop running
- Go 1.22+ (provided by Devbox or manual install)
- Task (go-task) installed (provided by Devbox or manual install)

## First-Time Setup
If you haven't installed Devbox yet, see [docs/SETUP.md](../SETUP.md) for complete installation instructions.

## Bootstrap
```bash
devbox shell
task dev
```

## Verify Services
```bash
# Gateway health
curl http://localhost:8080/healthz

# Gateway metrics
curl http://localhost:8080/metrics

# Auth health
curl http://localhost:8081/healthz
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
```

## Quality Checks
```bash
task lint
task test
task test:gateway
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

## Notes
- Gateway rate limiting uses Redis when `REDIS_ADDR` is set (for example `localhost:6379`).
- Gateway falls back to in-memory limiter if Redis is unavailable.
- Auth service uses in-memory token generation (no persistence yet).

