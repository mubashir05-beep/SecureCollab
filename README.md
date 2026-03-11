# SecureCollab

<p align="center">
	<img src="docs/assets/securecollab-logo.svg" alt="SecureCollab logo" width="140" />
</p>

<p align="center"><strong>SecureCollab</strong><br/>Zero-Knowledge Team Messaging Infrastructure</p>

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
- Svelte 4 UI shell (`client/ui`)
- Tailwind CSS design system tokens and reusable components
- Vite build tooling
- Vitest + Testing Library frontend smoke tests
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
- Tauri integration between Svelte UI and Rust core
- Kafka + Debezium + ClickHouse CDC pipeline
- Vault / mTLS hardening for production deployment

## Quick Start

```bash
# 1) Enter pinned development shell
cd SecureCollab
devbox shell

# 2) Start local stack
task dev

# 3) Run backend + client-core tests
task test

# 4) Run Svelte UI tests
task ui:test

# 5) Smoke check gateway
task smoke:gateway
```

## Run UI Locally

```bash
# One-time install
task ui:install

# Start Svelte dev server
task ui:dev
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
- Phase 2 has started in code with a tested Rust crypto foundation in `client/src-tauri` and a clean Svelte + Tailwind UI shell in `client/ui`.

## App Flow

1. User signs in through client UI.
2. Auth service issues JWT access token.
3. Gateway validates JWT and enforces rate limits.
4. Client initializes/loads local identity keys (Phase 2 crypto core).
5. Client encrypts outbound payloads before sending (in-progress in Phase 2).
6. Server stores ciphertext + metadata and routes events.
7. Recipient client decrypts on-device.

## Roles Model

Current role intent for workspace-level access:

- `owner`: Workspace creator, full admin control.
- `admin`: Manage members/channels and moderation actions.
- `member`: Default collaborator role; read/write in allowed channels.
- `viewer` (planned): Read-only access for audit/observer scenarios.

Current DB schema already supports role values via `workspace_members.role` and will be enforced by service authorization in upcoming steps.

## Repository Layout

```text
services/      Go microservices
client/        Svelte UI + Rust core for Tauri integration
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
