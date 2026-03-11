# SecureCollab
## Zero-Knowledge Team Messaging Infrastructure

> **E2E Encrypted · Self-Hostable · On-Premises Deployable**  
> Desktop App: Tauri | Web: **Out of Scope**  
> Author: Muhammad Mubashir Munir Khan | Version 1.0 | March 2026

---

## 1. Project Overview

SecureCollab is a production-grade, zero-knowledge team messaging platform designed to be self-hosted on-premises by organizations that cannot trust third-party messaging infrastructure. It combines Signal Protocol end-to-end encryption, a CDC-driven audit pipeline, institutional-grade rate limiting, and a Tauri-based desktop client into a single deployable stack.

This project demonstrates mastery across four high-value engineering domains simultaneously:

- Distributed systems & API gateway engineering (rate limiting, routing, auth)
- Cryptographic messaging infrastructure (Signal Protocol, key management)
- Real-time data pipelines (CDC with Debezium, Kafka, ClickHouse)
- Production infrastructure (Kubernetes, Terraform, Helm, Ansible, Observability)

> Web client is explicitly out of scope. All user-facing work is confined to the Tauri desktop application.

| Field | Value |
|---|---|
| **Project Name** | SecureCollab |
| **Type** | Open-Source Personal / Portfolio Project |
| **Platform** | Desktop (Tauri) — Web is OUT OF SCOPE |
| **Deployment Model** | Self-hosted, on-premises via Docker Compose and Kubernetes Helm chart |
| **Primary Language** | Go (backend), Rust/TypeScript (Tauri frontend) |
| **Total Estimated Duration** | ~20–28 weeks across 5 phases |
| **Resume Label** | SecureCollab — Zero-Knowledge Team Messaging Infrastructure |

---

## 2. Complete Technology Stack

### Backend & Core Services

| Category | Technology |
|---|---|
| **Primary Language** | Go (Golang) — all backend microservices |
| **API Framework** | Gin — HTTP REST API layer |
| **gRPC** | gRPC + protobuf — inter-service communication |
| **WebSockets** | Gorilla WebSocket — real-time messaging delivery |
| **Auth** | JWT (access + refresh tokens), PASETO v2 (server-to-server) |
| **Rate Limiting** | Custom Go implementation — sliding window + token bucket in Redis |
| **API Gateway** | Custom-built in Go — routing, auth middleware, rate limit enforcement |

### Cryptography & Security

| Category | Technology |
|---|---|
| **E2E Encryption** | Signal Protocol — Double Ratchet Algorithm + X3DH key exchange |
| **MLS (Future)** | Messaging Layer Security (RFC 9420) — group key agreement |
| **Key Storage** | Zero-knowledge: private keys NEVER leave client device |
| **Key Exchange** | Curve25519 ECDH (via libsodium bindings in Go) |
| **Server-Side** | AES-256-GCM for at-rest metadata; server never sees plaintext messages |
| **Transport** | TLS 1.3 (mutual TLS between services), Cert-Manager for cert automation |

### Desktop Client (Tauri)

| Category | Technology |
|---|---|
| **Framework** | Tauri v2 (Rust backend, webview frontend) |
| **Frontend (Webview)** | React + TypeScript + Tailwind CSS + ShadCN |
| **State Management** | Zustand — client state; TanStack Query — server state & caching |
| **Crypto in Client** | Rust crates: x25519-dalek, aes-gcm, hkdf, sha2 (via Tauri Rust backend) |
| **Local Storage** | SQLite (via sqlx in Rust) — encrypted message cache & key store |
| **Build/Packaging** | Tauri bundler — .msi (Windows), .dmg (macOS), .AppImage (Linux) |

### Messaging & Real-Time

| Category | Technology |
|---|---|
| **Message Broker** | Apache Kafka — event backbone for all system events |
| **Kafka Schema** | Confluent Schema Registry + Avro — typed event schemas |
| **WebSocket Gateway** | Custom Go service — fan-out message delivery to connected clients |
| **Presence System** | Redis Pub/Sub — online/offline/typing indicators |
| **Push Delivery** | Kafka consumer group per workspace — guaranteed delivery ordering |

### Databases & Storage

| Category | Technology |
|---|---|
| **Primary DB** | PostgreSQL 16 — users, workspaces, channels, message metadata (never plaintext) |
| **Cache / Rate Limit** | Redis 7 — rate limit counters (sliding window), sessions, presence |
| **Analytics DB** | ClickHouse — append-only analytics events from CDC pipeline |
| **Search** | Meilisearch — encrypted-metadata search (no plaintext content indexed) |
| **File Storage** | MinIO (S3-compatible, self-hosted) — encrypted file attachments |
| **ORM / Migrations** | sqlc (Go) + golang-migrate — type-safe queries, versioned schemas |

### CDC & Analytics Pipeline

| Category | Technology |
|---|---|
| **CDC Tool** | Debezium (Kafka Connect plugin) — captures PostgreSQL WAL changes |
| **Transport** | Kafka — CDC events flow: Postgres → Debezium → Kafka → ClickHouse |
| **Sink Connector** | ClickHouse Kafka Connect sink — consumes CDC events into ClickHouse |
| **Analytics Layer** | ClickHouse — message volume, user activity, workspace KPIs (no plaintext) |
| **Schema Registry** | Confluent Schema Registry — Avro schemas for CDC event types |

### Infrastructure & Cloud

| Category | Technology |
|---|---|
| **Containerization** | Docker + Docker Compose (local dev & single-node on-prem deploy) |
| **Orchestration** | Kubernetes (K3s for on-prem, EKS for cloud reference deploy) |
| **Helm** | Helm charts — all services packaged for one-command deployment |
| **IaC** | Terraform — VPC, EKS/EC2, RDS, ElastiCache, MSK, MinIO, security groups |
| **Config Mgmt** | Ansible — automated server provisioning, K3s cluster setup, cert install |
| **Service Mesh** | Istio (lightweight) OR Linkerd — mTLS between pods, traffic policies |
| **Ingress** | Nginx Ingress Controller — TLS termination, routing rules |
| **Secrets** | HashiCorp Vault — dynamic secrets, PKI, Kafka credentials |
| **Cert Automation** | cert-manager — automatic TLS cert issuance + renewal |

### Observability Stack

| Category | Technology |
|---|---|
| **Metrics** | Prometheus — scrapes all services via /metrics endpoint |
| **Dashboards** | Grafana — system metrics, business KPIs, rate limit graphs, Kafka lag |
| **Log Aggregation** | Loki + Promtail — structured JSON logs from all services |
| **Tracing** | OpenTelemetry (OTLP) + Jaeger — distributed trace across all services |
| **Alerting** | Grafana Alerting — PagerDuty/webhook alerts on error rate, latency SLOs |
| **Error Tracking** | Sentry (self-hosted) — runtime error aggregation per service |

### Developer Tooling

| Category | Technology |
|---|---|
| **Dev Environment** | Devbox (Nix-based) — declarative, reproducible toolchain; `devbox shell` pins Go, Rust, Node, protoc, sqlc, k6, Terraform, Helm, kubectl to exact versions in `devbox.json` |
| **Task Runner** | Task (Taskfile.yml) — replaces Makefile entirely; all dev workflows are named tasks: `task dev`, `task test`, `task lint`, `task migrate:up`, `task load-test`, `task deploy:local` |
| **Code Quality** | golangci-lint, staticcheck, gosec — invoked via `task lint`; enforced in CI |
| **API Testing** | Bruno (git-friendly) — API collections committed to repo; run via `task api-test` |
| **Load Testing** | k6 — invoked via `task load-test` and `task load-test:ws` |
| **Testing** | Go testing + testify — `task test`; testcontainers-go for integration tests via `task test:integration` |

### DevOps & CI/CD

| Category | Technology |
|---|---|
| **CI/CD** | GitHub Actions — on push: `task lint`, `task test`, Docker build+push, Helm lint |
| **Container Registry** | GitHub Container Registry (GHCR) — Docker image storage |
| **Release Pipeline** | Tag push triggers: Docker multi-arch build, Helm package+publish, Tauri binary build |
| **Environment Parity** | Devbox ensures CI uses identical tool versions as local dev — no "works on my machine" |

---

## 3. Developer Environment Convention

Every service and the repo root carries two files that define the complete developer experience.

### `devbox.json`

```json
{
  "$schema": "https://raw.githubusercontent.com/jetify-com/devbox/0.13.1/.schema/devbox.schema.json",
  "packages": [
    "go@1.22",
    "rust@latest",
    "nodejs@20",
    "protoc@25",
    "sqlc@1.26",
    "golang-migrate@4.17",
    "k6@0.50",
    "terraform@1.8",
    "helm@3.14",
    "kubectl@1.29",
    "golangci-lint@1.57",
    "docker-compose@2.27"
  ],
  "shell": {
    "init_hook": ["echo 'SecureCollab dev environment ready'"],
    "scripts": {
      "dev": "task dev",
      "test": "task test"
    }
  }
}
```

### `Taskfile.yml`

```yaml
version: '3'

tasks:
  dev:
    desc: Start all services locally via Docker Compose
    cmd: docker compose -f deploy/docker-compose.yml up --build

  dev:gateway:
    desc: Run API Gateway with hot reload
    cmd: air -c services/gateway/.air.toml

  test:
    desc: Run all unit tests
    cmd: go test ./... -race -cover

  test:integration:
    desc: Run integration tests (requires Docker)
    cmd: go test ./tests/integration/... -tags=integration -v

  lint:
    desc: Run all linters
    cmds:
      - golangci-lint run ./...
      - staticcheck ./...
      - gosec -quiet ./...

  migrate:up:
    desc: Run all pending DB migrations
    cmd: migrate -path db/migrations -database "$DATABASE_URL" up

  migrate:down:
    desc: Roll back last migration
    cmd: migrate -path db/migrations -database "$DATABASE_URL" down 1

  migrate:create:
    desc: Create a new migration file
    cmd: migrate create -ext sql -dir db/migrations -seq {{.NAME}}

  generate:
    desc: Run sqlc + protoc code generation
    cmds:
      - sqlc generate
      - protoc --go_out=. --go-grpc_out=. proto/**/*.proto

  load-test:
    desc: Run k6 load test against API Gateway
    cmd: k6 run tests/load/gateway.js

  load-test:ws:
    desc: Run k6 WebSocket load test
    cmd: k6 run tests/load/websocket.js

  deploy:local:
    desc: Deploy full stack to local K3s via Helm
    cmd: helm upgrade --install securecollab deploy/helm/securecollab -f deploy/helm/values.local.yaml

  deploy:prod:
    desc: Deploy to production K8s cluster
    cmd: helm upgrade --install securecollab deploy/helm/securecollab -f deploy/helm/values.prod.yaml

  infra:init:
    desc: Terraform init
    dir: infra/terraform
    cmd: terraform init

  infra:plan:
    desc: Terraform plan
    dir: infra/terraform
    cmd: terraform plan

  infra:apply:
    desc: Terraform apply
    dir: infra/terraform
    cmd: terraform apply

  tauri:dev:
    desc: Start Tauri desktop app in dev mode
    dir: client
    cmd: cargo tauri dev

  tauri:build:
    desc: Build Tauri desktop binaries
    dir: client
    cmd: cargo tauri build

  api-test:
    desc: Run Bruno API test collection
    cmd: bru run collections/securecollab --env local

  clean:
    desc: Tear down local Docker Compose stack and volumes
    cmd: docker compose -f deploy/docker-compose.yml down -v
```

> Any contributor runs `devbox shell` then `task dev` — full stack up, no manual tool installation, no version mismatches.

---

## 4. System Architecture

SecureCollab follows a microservices architecture with a clear separation between the zero-knowledge client layer and the server layer. The server processes only encrypted ciphertext and metadata — it never has access to plaintext messages or private keys.

### 4.1 Services Map

- **API Gateway** (Go/Gin) — single entry point, rate limiting, auth verification, routing
- **Auth Service** (Go) — registration, login, JWT issuance, key bundle upload
- **Messaging Service** (Go) — encrypted message ingestion, Kafka publish, WebSocket fan-out
- **Key Distribution Service** (Go) — Signal Protocol prekey bundle distribution (zero-trust)
- **Channel Service** (Go) — workspace/channel CRUD, membership management
- **File Service** (Go) — client-encrypted file upload/download via MinIO
- **Notification Service** (Go) — Kafka consumer, push/in-app notification delivery
- **Analytics Service** (Go) — ClickHouse query layer, exposes dashboard APIs
- **CDC Pipeline** (Debezium + Kafka) — PostgreSQL WAL → Kafka → ClickHouse

### 4.2 Zero-Knowledge Guarantee

The most critical architectural constraint: **the server is cryptographically unable to read message content.**

- Client generates Curve25519 identity key pair locally (Tauri Rust layer)
- X3DH key exchange establishes shared session keys between clients only
- Double Ratchet advances keys per-message — compromise of one key exposes only that message
- Server stores and relays only ciphertext byte blobs — zero plaintext exposure
- CDC audit log captures metadata events only — message content field is always NULL on server

### 4.3 Mono-repo Structure

```
securecollab/
├── devbox.json                  # Root Devbox — shared tool versions
├── Taskfile.yml                 # Root Taskfile — all dev workflows
├── services/
│   ├── gateway/                 # API Gateway (Go)
│   ├── auth/                    # Auth Service (Go)
│   ├── messaging/               # Messaging Service (Go)
│   ├── key-distribution/        # Key Distribution Service (Go)
│   ├── channel/                 # Channel Service (Go)
│   ├── file/                    # File Service (Go)
│   ├── notification/            # Notification Service (Go)
│   └── analytics/               # Analytics Service (Go)
├── client/                      # Tauri desktop app (Rust + React/TS)
├── pipeline/                    # Debezium connectors, ClickHouse schemas
├── proto/                       # Protobuf definitions (shared)
├── db/
│   └── migrations/              # golang-migrate SQL files
├── deploy/
│   ├── docker-compose.yml       # Local full-stack dev (`task dev`)
│   ├── helm/                    # Helm chart (`task deploy:local`)
│   └── ansible/                 # On-prem provisioning playbooks
├── infra/
│   └── terraform/               # AWS reference deploy (`task infra:apply`)
├── tests/
│   ├── integration/             # testcontainers-go (`task test:integration`)
│   └── load/                    # k6 scripts (`task load-test`)
└── collections/                 # Bruno API collections (`task api-test`)
```

---

## 5. Development Phases

> Each phase produces a standalone, functional milestone. Never leave a phase "half-done" — complete the milestone before advancing.

---

### Phase 1 — Foundation & API Gateway
**Duration: 3–4 weeks**

**Phase Goal:** A production-ready API Gateway in Go with custom rate limiting (sliding window + token bucket) backed by Redis, JWT auth middleware, and full observability. This phase alone demonstrates the Distributed Rate Limiter project requirement.

| Component | Tech Stack | Deliverable |
|---|---|---|
| Mono-repo & Devbox Setup | Devbox, Go workspaces, golangci-lint, GitHub Actions CI | `devbox shell` + `task dev` gives working environment; CI green |
| Taskfile Workflow | Taskfile.yml — all tasks defined: dev, test, lint, migrate, load-test | Every dev action is `task <name>`, no raw commands needed |
| PostgreSQL Schema | PostgreSQL 16, golang-migrate (`task migrate:up`), sqlc (`task generate`) | Versioned schema: users, workspaces, channels, keys |
| Redis Setup | Redis 7, go-redis/v9 | Redis running, connection pooling configured |
| Auth Service | Go, Gin, JWT, bcrypt, PASETO | Register/login/refresh endpoints with tested JWT flow |
| API Gateway Core | Go, Gin, reverse proxy middleware | Request routing to downstream services |
| Rate Limiter — Sliding Window | Go, Redis (sorted sets), Lua scripts | Per-user/IP sliding window with atomic Redis Lua |
| Rate Limiter — Token Bucket | Go, Redis (hash), atomic operations | Burst-capable token bucket for bot/API consumers |
| Auth Middleware | Go middleware, JWT validation | All routes protected; unauthenticated requests rejected |
| Prometheus Metrics | prometheus/client_golang | /metrics exposed: request count, latency, rate limit hits |
| Grafana Dashboard | Grafana, Prometheus datasource | Live dashboard: RPS, P95 latency, rate limit heatmap |
| Structured Logging | Zap logger, Loki, Promtail | JSON logs shipping to Loki, queryable in Grafana |
| k6 Load Test | k6, `task load-test` | Load test proves rate limiter correctness under 1K RPS |
| Bruno API Collection | Bruno, `task api-test` | All auth + gateway endpoints documented and tested |
| Docker Compose Stack | Docker Compose, `task dev` | Single command spins up: Postgres, Redis, Gateway, Grafana |

> ✅ **PHASE MILESTONE:** `task dev` starts the full stack. `task load-test` proves the API Gateway sustains 1K+ RPS, correctly enforces both rate limiting algorithms, all endpoints are authenticated, Grafana dashboard is live, k6 passes with 0 errors.

---

### Phase 2 — Cryptographic Messaging Core
**Duration: 5–6 weeks** *(Hardest phase — allocate extra time)*

**Phase Goal:** Full Signal Protocol implementation: X3DH key exchange, Double Ratchet message encryption, Key Distribution Service, and real-time WebSocket delivery. The server becomes provably zero-knowledge.

| Component | Tech Stack | Deliverable |
|---|---|---|
| Tauri Project Bootstrap | Tauri v2, Rust, React, TypeScript, Tailwind, ShadCN, `task tauri:dev` | Tauri app shell: login screen, workspace sidebar, channel view |
| Client Crypto (Rust) | x25519-dalek, aes-gcm, hkdf, sha2, rand_core | Rust module: key generation, X3DH, Double Ratchet (unit tested) |
| SQLite Key Store | SQLite via sqlx in Rust, encrypted at rest | Private keys stored locally — never sent to server |
| Key Distribution Service | Go, Gin, PostgreSQL (stores only public keys) | Upload prekey bundle; fetch recipient's public key bundle |
| Messaging Service | Go, Gin, PostgreSQL, Kafka producer | Accepts encrypted blob, stores ciphertext, publishes to Kafka |
| WebSocket Gateway | Go, Gorilla WebSocket, Redis Pub/Sub | Authenticated WS connections; fan-out to online recipients |
| Kafka Integration | Apache Kafka, confluent-kafka-go, Avro schemas | Message events flowing: Messaging Service → Kafka → WS Gateway |
| Schema Registry | Confluent Schema Registry, Avro | All Kafka event types have versioned Avro schemas |
| E2E Encryption Flow | Tauri Rust + Key Distribution Service + Messaging Service | Alice sends encrypted message; Bob decrypts on his device only |
| Presence System | Redis Pub/Sub, Go, WebSocket | Online/offline/typing indicators working in Tauri UI |
| OpenTelemetry Tracing | otel-go, Jaeger | Distributed traces: Gateway → Messaging → Kafka → WS Gateway |
| Integration Test | testcontainers-go, `task test:integration` | Automated: register → key exchange → send → receive → decrypt |

> ✅ **PHASE MILESTONE:** `task test:integration` passes the full E2E crypto flow. A message sent from one Tauri instance is received by another, fully decrypted on the recipient's device. The server's database contains ZERO plaintext. Jaeger shows full distributed trace.

---

### Phase 3 — CDC Analytics Pipeline
**Duration: 3–4 weeks**

**Phase Goal:** A complete Change Data Capture pipeline: Debezium reads PostgreSQL WAL, publishes events to Kafka, ClickHouse consumes them for real-time analytics. Zero plaintext in the analytics layer.

| Component | Tech Stack | Deliverable |
|---|---|---|
| Debezium Setup | Debezium PostgreSQL connector, Kafka Connect | Debezium captures INSERT/UPDATE/DELETE from Postgres WAL |
| Kafka Connect Stack | Kafka Connect distributed mode, connector REST API | Connectors manageable via REST; schema evolution handled |
| ClickHouse Setup | ClickHouse server, ClickHouse Keeper | Tables defined: message_events, user_events, workspace_events |
| ClickHouse Kafka Sink | ClickHouse Kafka engine / Kafka Connect ClickHouse sink | CDC events land in ClickHouse within 1–2 seconds of DB write |
| Analytics Service | Go, Gin, ClickHouse Go driver | REST API: message volume, active users, channel activity (no content) |
| Grafana Analytics Board | Grafana, ClickHouse datasource plugin | Real-time workspace analytics dashboard (privacy-preserving) |
| Schema Evolution Test | Debezium schema history, Schema Registry | Adding a DB column does not break the pipeline |
| Backpressure Handling | Kafka consumer lag monitoring, Prometheus | Consumer lag alert fires when ClickHouse sink falls behind |
| Bruno CDC Collection | Bruno, `task api-test` | Analytics service endpoints tested and documented |

> ✅ **PHASE MILESTONE:** A new message sent in the Tauri app appears as an analytics event in ClickHouse within 2 seconds. Grafana analytics dashboard shows live workspace KPIs. Schema evolution test passes. Consumer lag monitored and alerted.

---

### Phase 4 — Production Features & Hardening
**Duration: 4–5 weeks**

**Phase Goal:** File sharing, search, notifications, advanced security hardening, and full test coverage. The system becomes feature-complete and production-safe.

| Component | Tech Stack | Deliverable |
|---|---|---|
| File Service | Go, MinIO (self-hosted), client-side AES-256-GCM | Files encrypted on client before upload; server stores ciphertext blobs |
| Meilisearch Integration | Meilisearch, Go client | Search over channel names, user names, metadata — never message content |
| Notification Service | Go, Kafka consumer, in-app WebSocket push | Mention/reply notifications delivered in real-time |
| HashiCorp Vault | Vault server, Go Vault SDK, Kubernetes auth method | All service secrets injected via Vault — nothing in env files |
| Cert-Manager | cert-manager, Let's Encrypt / self-signed CA | Automatic TLS cert issuance for all ingress endpoints |
| Istio / Linkerd mTLS | Istio or Linkerd service mesh | All pod-to-pod traffic encrypted via mTLS |
| Security Hardening | Go security headers, CORS, CSRF, input sanitisation | OWASP Top 10 checklist reviewed and addressed per service |
| Full Test Suite | testcontainers-go, `task test` + `task test:integration` | E2E test: registration → key exchange → message → decrypt |
| k6 Full Load Test | k6, `task load-test` + `task load-test:ws` | 500 concurrent WS connections sustained, P95 < 100ms |
| Sentry Self-Hosted | Sentry (Docker), Go Sentry SDK | Runtime errors from all services visible in Sentry |

> ✅ **PHASE MILESTONE:** `task test` and `task test:integration` both pass. `task load-test:ws` sustains 500 concurrent WebSocket connections. All secrets managed via Vault. mTLS enforced between all pods. k6 P95 latency under 100ms.

---

### Phase 5 — On-Premises Packaging & Release
**Duration: 3–4 weeks**

**Phase Goal:** Package the entire system for one-command self-hosted deployment. Publish Helm chart, Terraform modules, Ansible playbooks. Build and release Tauri desktop binaries. Write operator documentation.

| Component | Tech Stack | Deliverable |
|---|---|---|
| Helm Chart (Full Stack) | Helm v3, `task deploy:local` | `task deploy:local` deploys all services to any K8s cluster |
| Terraform Modules | Terraform, AWS provider, `task infra:apply` | `task infra:apply` provisions full AWS reference environment |
| Ansible Playbooks | Ansible, SSH provisioning | Playbooks: K3s setup, cert install, Vault init, initial admin user |
| Docker Compose (Single Node) | Docker Compose v2, `task dev` | `task dev` starts complete stack on a single Linux server |
| Tauri Desktop Build | Tauri v2 bundler, `task tauri:build`, GitHub Actions | Signed .msi, .dmg, .AppImage published in GitHub Releases |
| GitHub Actions Release CI | GitHub Actions, GHCR, Helm chart museum | Tag push triggers: Docker build+push, Helm publish, Tauri binary release |
| Operator Documentation | Markdown, MkDocs or Docusaurus | Docs: `devbox shell` + `task dev` quickstart, key rotation, backup/restore, upgrade path |
| Performance Benchmarks | k6, `task load-test`, Grafana, ClickHouse | Published benchmark: throughput, latency, Kafka lag, ClickHouse ingest rate |
| README & Architecture Diagram | Markdown, Excalidraw / draw.io | Architecture diagram, tech badges, one-command quickstart in README |

> ✅ **PHASE MILESTONE:** A fresh Linux server runs `devbox shell` → `task dev` = fully operational SecureCollab instance. Tauri binaries downloadable from GitHub Releases. `task deploy:local` deploys to any K8s cluster in one command. Architecture diagram published in README.

---

## 6. Resume Entries

> Copy these verbatim into your resume LaTeX once each phase is complete.

### After Phase 1 (API Gateway complete)
**SecureCollab — Zero-Knowledge Team Messaging Infrastructure**
- Built a production-grade custom API Gateway in Go with dual rate limiting algorithms (sliding window + token bucket) backed by Redis Lua scripts, sustaining 1K+ RPS with 0 errors under k6 load test.
- Standardised developer environment with Devbox (Nix-based) and Taskfile, enabling one-command (`devbox shell` + `task dev`) full-stack local setup with pinned tool versions across the entire project.

### After Phase 2 (E2E Encryption complete)
- Implemented Signal Protocol (X3DH + Double Ratchet) in Rust (Tauri), achieving cryptographic zero-knowledge: server stores only ciphertext, private keys never leave the device.
- Built WebSocket fan-out gateway in Go with Kafka event backbone and Redis Pub/Sub presence system for real-time message delivery.

### After Phase 3 (CDC Pipeline complete)
- Architected a CDC analytics pipeline: Debezium captures PostgreSQL WAL → Kafka → ClickHouse, delivering analytics events within 2s of DB write across 3 event types.
- Built privacy-preserving real-time analytics dashboard in Grafana over ClickHouse — zero plaintext in the analytics layer by design.

### After Phase 5 (Full release)
- Packaged full stack as a Helm chart + Terraform modules + Ansible playbooks; `task deploy:local` deploys the entire system to any Kubernetes cluster in one command.
- Shipped cross-platform Tauri desktop binaries (.msi, .dmg, .AppImage) via automated GitHub Actions release pipeline.

---

## 7. Explicit Scope Boundaries

| ✅ IN SCOPE | ❌ OUT OF SCOPE |
|---|---|
| Tauri desktop app (Windows, macOS, Linux) | Web browser client |
| Self-hosted, on-premises Kubernetes deployment | SaaS / managed cloud offering |
| E2E encryption via Signal Protocol | SMS / PSTN bridging |
| CDC pipeline (Debezium → Kafka → ClickHouse) | Mobile app (iOS / Android) |
| Custom Go API Gateway with rate limiting | Video/voice calls (WebRTC) |
| Helm chart for one-command deploy | AI/ML features |
| Terraform + Ansible for infra provisioning | Billing or subscription system |
| Real-time WebSocket messaging | Multi-region active-active clustering |
| Devbox + Taskfile developer toolchain | — |

---

## 8. Project Timeline Summary

| Phase | Title | Duration | Key Milestone |
|---|---|---|---|
| 1 | Foundation & API Gateway | 3–4 weeks | `task load-test` passes 1K RPS, rate limiter proven |
| 2 | Cryptographic Messaging | 5–6 weeks | `task test:integration` passes E2E crypto, zero plaintext on server |
| 3 | CDC Analytics Pipeline | 3–4 weeks | CDC events in ClickHouse within 2s, Grafana dashboard live |
| 4 | Production Hardening | 4–5 weeks | 500 concurrent WS, Vault secrets, mTLS, k6 P95 < 100ms |
| 5 | On-Prem Packaging | 3–4 weeks | `devbox shell` + `task dev` = running instance, Tauri binaries on GitHub |

**Total estimated duration:** 18–23 weeks (part-time alongside current job) or 12–16 weeks (dedicated focus).

---

*SecureCollab Project Specification · Muhammad Mubashir Munir Khan · v1.0 · March 2026*
