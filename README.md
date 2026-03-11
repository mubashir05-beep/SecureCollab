# SecureCollab

<p align="center">
	<img src="docs/assets/securecollab-logo.svg" alt="SecureCollab logo" width="760" />
</p>

SecureCollab is a self-hosted, zero-knowledge team messaging platform built for organizations that cannot trust third-party chat infrastructure with sensitive internal communication.

## The Problem

Most collaboration tools optimize for convenience over control. That creates hard problems for security-conscious teams:

- Message content is often visible to service providers.
- Compliance and audit requirements are difficult in black-box SaaS systems.
- On-premises deployment paths are limited or expensive.
- Reproducible developer environments are hard to maintain across teams.

## What SecureCollab Solves

SecureCollab is designed to provide:

- End-to-end encryption with zero-knowledge server constraints.
- Self-hosted deployment for private networks and regulated environments.
- A practical operations stack (metrics, logs, dashboards).
- A reproducible local developer workflow (`devbox` + `task`).

## Core Architecture

- `services/gateway`: Entry point, auth middleware, rate limiting, metrics.
- `services/auth`: Register/login/refresh API and JWT flow.
- `db/migrations`: Versioned schema changes.
- `deploy/docker-compose.yml`: Local infrastructure and services.
- `docs/`: Architecture notes, ADRs, runbooks.

## Tech Stack

### Implemented Now

#### Backend and APIs
- Go 1.22
- Gin
- JWT (`golang-jwt/jwt`)

#### Data and State
- PostgreSQL 16
- Redis 7

#### Observability
- Prometheus
- Grafana
- Loki + Promtail

#### Client (Phase 2 Started)
- Rust client core crate (`client/src-tauri`)
- X25519 identity key generation (`x25519-dalek`)
- Hash/key utilities (`sha2`, `base64`, `hex`)

#### Developer Experience
- Devbox (Nix-based, pinned toolchain)
- Taskfile (`task`) workflows
- Docker Compose local stack
- k6 load testing (`tests/load/gateway.js`)
- golang-migrate tasks (`migrate:up`, `migrate:down`, `migrate:create`)

### Planned Next Technologies
- Tauri (Rust + TypeScript) desktop client
- React + TypeScript client UI shell
- Kafka + Debezium + ClickHouse CDC pipeline
- Vault / mTLS hardening for production deployment

## Quick Start

```bash
# 1) Enter pinned development shell
cd SecureCollab
devbox shell

# 2) Start local stack
task dev

# 3) Run tests
task test

# 4) Smoke check gateway
task smoke:gateway
```

## Useful Commands

```bash
# Gateway only tests
task test:gateway

# Run database migrations (set DATABASE_URL first)
export DATABASE_URL="postgres://securecollab:securecollab@localhost:5432/securecollab?sslmode=disable"
task migrate:up

# Gateway load test scaffolding
task load-test
```

## Current Project Status

- Phase 1 is in active completion work with gateway + auth implemented and tested.
- Redis-backed gateway rate limiting is implemented with in-memory fallback for local resilience.
- Phase 2 has started in code with a tested Rust crypto foundation in `client/src-tauri`.

## Repository Layout

```text
services/      Go microservices
client/        Tauri desktop app (Phase 2+)
db/            SQL migrations
deploy/        Docker Compose and deployment assets
docs/          Architecture, ADRs, runbooks
tests/         Integration and load tests
```

## Documentation

- Setup: `docs/SETUP.md`
- Architecture: `docs/architecture.md`
- Local dev runbook: `docs/runbooks/local-dev.md`
- ADR index: `docs/adr/README.md`

## License

License to be added.
