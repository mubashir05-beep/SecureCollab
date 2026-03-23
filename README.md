# SecureCollab

<p align="center">
	<img src="docs/assets/securecollab-logo.svg" alt="SecureCollab logo" width="140" />
</p>

<p align="center"><strong>SecureCollab</strong><br/>Zero-Knowledge Team Messaging & Project Management</p>

SecureCollab is a self-hosted, zero-knowledge collaboration platform — combining **Slack-like messaging** with **ClickUp-like project management**. Built for teams that need full control over their communication infrastructure.

## What It Does

- **End-to-end encrypted messaging** — X25519 key exchange + ChaCha20-Poly1305. Server never sees plaintext.
- **Workspaces & Channels** — Slack-style organization with public/private channels, invite codes, role-based access.
- **Rich messaging** — Threads, emoji reactions, pins, @mentions, markdown, link previews, edit/delete.
- **Real-time delivery** — WebSocket-based message push with encrypted payloads.
- **CDC Analytics** — Debezium → Redpanda → ClickHouse pipeline for operational analytics (no plaintext exposure).
- **Full observability** — Prometheus metrics, Grafana dashboards, Loki logs, alerting rules.

## Architecture

```text
┌─────────────┐     ┌──────────┐     ┌───────────┐     ┌───────────┐
│  Svelte UI  │────▸│ Gateway  │────▸│   Auth    │────▸│ Postgres  │
│  (Tauri)    │     │ (JWT+RL) │     │           │     │           │
└─────────────┘     └──────────┘     └───────────┘     └───────────┘
       │                                                      │
       ├──▸ Messaging Service (E2E encrypted) ◂───────────────┤
       ├──▸ Workspace Service (channels, members) ◂───────────┤
       ├──▸ KeyDist Service (public key exchange) ◂───────────┤
       └──▸ Analytics Service (ClickHouse + Postgres) ◂───────┘
```

### Services

| Service | Port | Description |
|---------|------|-------------|
| `gateway` | 8080 | API entry, JWT middleware, rate limiting, metrics |
| `auth` | 8081 | Register, login, refresh, JWT issuance |
| `messaging` | 8082 | E2E encrypted messages, threads, reactions, pins, WebSocket delivery |
| `keydist` | 8083 | Public key upload/fetch for client key exchange |
| `analytics` | 8084 | Message volume analytics (ClickHouse with Postgres fallback) |
| `workspace` | 8086 | Workspaces, channels, members, invite codes, role enforcement |

## Tech Stack

| Layer | Technologies |
|-------|-------------|
| **Backend** | Go 1.22, Gin, JWT, PostgreSQL 16, Redis 7 |
| **Client** | Svelte 4, Tailwind CSS, Vite |
| **Desktop** | Tauri (Rust), X25519 + ChaCha20-Poly1305 |
| **CDC Pipeline** | Debezium, Redpanda (Kafka), ClickHouse |
| **Observability** | Prometheus, Grafana, Loki + Promtail |
| **DevEx** | Devbox (Nix), Taskfile, Docker Compose |

## Quick Start

```bash
# Enter dev shell (optional, for pinned toolchain)
devbox shell

# Start everything (Postgres, Redis, all services, auto-migrations)
task dev

# Start the UI dev server
task ui:dev
# Open http://localhost:5173

# Run all tests
task test
```

Migrations run automatically on `task dev` — no manual setup needed.

## Useful Commands

```bash
task dev              # Start full Docker stack with auto-migrations
task dev:down         # Stop stack
task ui:dev           # Svelte dev server at localhost:5173
task test             # Run all Go + Rust tests
task load-test        # k6 gateway load test (1200 req/s target)
task cdc:up           # Start CDC overlay (Redpanda, Debezium, ClickHouse)
task cdc:register     # Register Debezium Postgres connector
task cdc:validate     # E2E CDC pipeline validation
task smoke:gateway    # Quick gateway health check
```

## App Flow

1. User registers/logs in → auth service issues JWT + user_id.
2. Client auto-generates X25519 keypair, uploads public key to keydist.
3. User creates or joins a workspace (via invite code).
4. User creates channels, invites members.
5. Messages are encrypted client-side before sending.
6. Server stores ciphertext + nonce only — zero plaintext.
7. Recipients decrypt on-device using shared key derivation.
8. Real-time delivery via authenticated WebSocket.

## Roles Model

| Role | Permissions |
|------|------------|
| `owner` | Full control — workspace settings, members, channels, billing |
| `admin` | Manage members and channels, moderate messages |
| `member` | Read/write in allowed channels |
| `viewer` | Read-only access for audit/observer scenarios |

## Project Status

| Phase | Status | Summary |
|-------|--------|---------|
| 1 - API Gateway | Complete | Gateway, auth, rate limiting, load tests |
| 2 - Crypto Core | Complete | Rust crypto, Tauri bridge, key bootstrap, browser fallback |
| 3 - CDC Pipeline | Complete | Debezium → Redpanda → ClickHouse, dashboards, validation |
| 4 - Workspaces | Complete | Workspace/channel CRUD, invites, members, onboarding UI |
| 5 - Rich Messaging | Complete | Threads, reactions, pins, @mentions, markdown, link previews |
| 6 - File Attachments | Not Started | Encrypted upload/download, MinIO, drag-drop |
| 7 - Kanban (ClickUp) | Not Started | Boards, columns, tasks, drag-drop, labels |
| 8 - Notifications | Not Started | Real-time notifications, typing indicators, presence |
| 9 - Search/Admin | Not Started | Metadata search, admin panel, audit log |
| 10 - Production | Not Started | Secrets, mTLS, Helm, Terraform, Tauri builds |

Full checklist: `docs/PHASE_CHECKLIST.md` | Product scope: `docs/PRODUCT_SCOPE.md`

## Repository Layout

```text
services/      Go microservices (gateway, auth, messaging, keydist, workspace, analytics)
client/ui/     Svelte 4 + Tailwind CSS frontend
client/src-tauri/  Rust crypto core (X25519, ChaCha20-Poly1305)
db/migrations/ Versioned SQL schema (auto-applied on docker compose up)
deploy/        Docker Compose, observability configs
pipeline/      CDC configs (Debezium connectors, ClickHouse init)
tests/         Load tests, CDC validation
docs/          Architecture, ADRs, runbooks, phase checklist
```

## Documentation

- [Phase Checklist](docs/PHASE_CHECKLIST.md)
- [Product Scope](docs/PRODUCT_SCOPE.md)
- [Architecture](docs/architecture.md)
- [Phase 1 Gate Report](docs/PHASE1_GATE_REPORT.md)
- [Local Dev Runbook](docs/runbooks/local-dev.md)
- [CDC Local Runbook](docs/runbooks/cdc-local.md)

## License

License to be added.
