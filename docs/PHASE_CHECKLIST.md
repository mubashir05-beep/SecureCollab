# SecureCollab Phase Checklist

Last updated: March 21, 2026

Full product scope: `docs/PRODUCT_SCOPE.md`

## Status Legend
- `[x]` Done
- `[~]` In Progress
- `[ ]` Not Started

## Phase 1 - Foundation and API Gateway
Overall: `[x]` Complete

- `[x]` Mono-repo structure and docs/code separation established.
- `[x]` Dev environment baseline (`devbox.json`, `Taskfile.yml`, `.editorconfig`, `.gitignore`).
- `[x]` Local stack with Postgres, Redis, Prometheus, Grafana, Loki, Promtail, Gateway, Auth.
- `[x]` Gateway endpoints: `/healthz`, `/metrics`, protected route with JWT middleware.
- `[x]` Rate limiting: sliding window + token bucket, Redis-backed with in-memory fallback.
- `[x]` Gateway unit tests for auth, limiter, fallback, and metrics.
- `[x]` Auth endpoints: `/register`, `/login`, `/refresh`, `/healthz`.
- `[x]` Auth persistence to Postgres with in-memory fallback.
- `[x]` Auth unit and integration tests passing.
- `[x]` Load test tuned to constant-arrival-rate at 1200 req/s with `rate>=1000` threshold.
- `[x]` Phase 1 gate report in `docs/PHASE1_GATE_REPORT.md`.

## Phase 2 - Cryptographic Messaging Core
Overall: `[x]` Complete

- `[x]` Rust crypto: X25519 keygen, ChaCha20-Poly1305 encrypt/decrypt with roundtrip tests.
- `[x]` Tauri bridge: command wrappers, tauri.conf.json, main.rs with invoke handler.
- `[x]` Svelte crypto module: Tauri invoke with browser fallback.
- `[x]` Svelte auth flow wired to real auth API with persistent store.
- `[x]` API client module (auth, keydist, messaging).
- `[x]` Message send/receive UI with WebSocket real-time delivery.
- `[x]` Key Distribution Service (upload/fetch APIs with tests).
- `[x]` Messaging service E2E encrypted payload flow.
- `[x]` WebSocket delivery path with auth.
- `[x]` Full E2E encrypt/send/receive/decrypt integration test.
- `[x]` Zero-plaintext proof check in persistence path.
- `[x]` Key bootstrap flow: auto-generate keys on login, upload to keydist, localStorage persistence, peer key fetching.

## Phase 3 - CDC Analytics Pipeline
Overall: `[x]` Complete

- `[x]` Debezium connector configuration for `encrypted_messages`.
- `[x]` Kafka Connect runtime in local CDC overlay.
- `[x]` ClickHouse sink: Kafka engine + materialized views + hourly aggregation.
- `[x]` Analytics service with ClickHouse backend (falls back to Postgres → in-memory).
- `[x]` Grafana CDC analytics dashboard (6 panels).
- `[x]` Consumer lag metrics and alerting (Prometheus rules).
- `[x]` End-to-end CDC validation script (`task cdc:validate`).

## Phase 4 - Workspace and Channel System (Slack Core)
Overall: `[x]` Complete

### Backend
- `[x]` Workspace service: create, list, join (invite code), settings, member management APIs.
- `[x]` Channel service: create (public/private), list, archive, topic update APIs.
- `[x]` Role enforcement middleware: owner/admin required for member add/remove and archive.
- `[x]` Workspace + channel handler unit tests (8 tests passing).
- `[x]` DB migration 003: workspace/channel enhancements (is_private, topic, channel_members, invite_code).
- `[x]` Docker Compose + Taskfile integration for workspace service.

### Frontend (Brick 1-3)
- `[x]` App layout shell: sidebar + main content + top bar.
- `[x]` Workspace switcher component (dark Slack-style rail).
- `[x]` Channel list with unread indicators and private channel badges.
- `[x]` Create workspace modal + API wiring.
- `[x]` Create channel modal + API wiring.
- `[x]` Empty states: no workspaces, no channels, empty channel.
- `[x]` Member management UI panel (add/remove, role badges, admin controls).
- `[x]` Invite modal (share code + join by code).
- `[x]` Logout button in sidebar.
- `[x]` Onboarding flow (no workspaces → create or join prompt).
- `[x]` CORS middleware on all backend services.
- `[x]` Auto-migration on `docker compose up` (migrate service).
- `[x]` Auth returns user_id + username (UUID fix for messaging).

## Phase 5 - Rich Messaging Features
Overall: `[x]` Complete

### Backend
- `[x]` DB migration 004: threads (parent_message_id), reactions, pins tables.
- `[x]` Rich messaging store interface + in-memory implementation.
- `[x]` Thread reply POST/GET endpoints.
- `[x]` Reaction add/remove/list endpoints.
- `[x]` Pin/unpin/list-pins endpoints.
- `[x]` Edit and delete message endpoints (sender-only enforcement).
- `[x]` Rich messaging handler tests (5 tests passing).

### Frontend
- `[x]` EmojiPicker component (12 common emojis).
- `[x]` ThreadPanel component (parent + replies + reply input).
- `[x]` MessageBubble hover actions (react, thread, pin, delete).
- `[x]` Reaction display on messages.
- `[x]` Thread open/reply/close wired in App.svelte.
- `[x]` Pin and delete actions wired in App.svelte.
- `[x]` @mention users and @channel with autocomplete.
- `[x]` Markdown rendering and code blocks (lightweight custom parser).
- `[x]` Link previews (client-side URL card unfurl).

## Phase 6 - File Attachments
Overall: `[ ]` Not Started

- `[ ]` File service: client-side encrypted upload/download.
- `[ ]` S3-compatible storage backend (MinIO for local dev).
- `[ ]` Drag-and-drop upload in chat.
- `[ ]` Image/file preview components.
- `[ ]` File size limits and type validation.

## Phase 7 - Project Management (Kanban)
Overall: `[ ]` Not Started

- `[ ]` Projects service: board CRUD, column CRUD, task CRUD APIs.
- `[ ]` DB schema: boards, columns, tasks, task_comments, task_labels.
- `[ ]` Kanban board UI with drag-and-drop.
- `[ ]` Task card detail modal (description, assignee, due date, priority, labels).
- `[ ]` Task comments linked to messaging.
- `[ ]` Board views: Kanban and List.
- `[ ]` Task status tracking and filters.

## Phase 8 - Notifications and Presence
Overall: `[ ]` Not Started

- `[ ]` Notifications service: event-driven delivery.
- `[ ]` In-app notification center UI.
- `[ ]` Desktop notifications (Tauri).
- `[ ]` Online/offline/away presence indicators.
- `[ ]` Typing indicators in channels.
- `[ ]` Custom user status.

## Phase 9 - Search and Admin
Overall: `[ ]` Not Started

- `[ ]` Metadata search service (zero-knowledge — searches on metadata, not plaintext).
- `[ ]` Global search bar UI with filters.
- `[ ]` Workspace admin panel.
- `[ ]` Audit log.
- `[ ]` Data export.

## Phase 10 - Production Hardening and Release
Overall: `[ ]` Not Started

- `[ ]` Vault-based secrets management.
- `[ ]` mTLS/service mesh and cert automation.
- `[ ]` Full OWASP hardening pass.
- `[ ]` Helm packaging for full stack.
- `[ ]` Terraform modules for reference infra.
- `[ ]` Tauri build/release automation (Windows/macOS/Linux).
- `[ ]` Operator runbooks (backup, restore, upgrade, incident).
- `[ ]` Final architecture diagram and release-grade README.
