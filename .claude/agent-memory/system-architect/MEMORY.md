# System Architect Memory

## Architecture Patterns (confirmed March 2026)
- Each Go service follows: `cmd/<name>/main.go` + `internal/handlers/` + `internal/store/`
- Store layer uses interface pattern: in-memory impl + Postgres impl + `NewXxxFromEnv()` factory
- All services init from env vars: `DATABASE_URL`, `JWT_SECRET`, `REDIS_ADDR`, `CLICKHOUSE_DSN`
- Each service has its own copy of authMiddleware, corsMiddleware, JWT claims struct (DUPLICATION)
- Services run on fixed ports: gateway:8080, auth:8081, keydist:8082, messaging:8083, analytics:8084, workspace:8086

## Key Security Issues Found
- Hardcoded dev JWT secret `securecollab-dev-secret-key` as exported const in auth/jwt.go (line 11)
- CORS is `Access-Control-Allow-Origin: *` in every service
- No password complexity validation in auth register handler
- Token in URL query string for WebSocket auth (access_token param)
- Private key stored in browser localStorage (keyStore.js)
- No token blacklist/revocation mechanism
- InMemoryUserStore has no mutex (race condition in tests)
- Gateway ValidateToken does NOT check signing method (auth service jwt.go line 43)
- Browser fallback crypto uses XOR (not real encryption) - acceptable for dev

## Duplicated Code (cross-service)
- authMiddleware: gateway, keydist, messaging, workspace (4 copies)
- corsMiddleware: auth, keydist, messaging, workspace (4 copies)
- JWT claims struct: gateway, keydist, messaging, workspace (4 copies)
- jwtSecretFromEnv: all services
- testToken helper: duplicated in every test file

## CDC Pipeline
- Postgres (WAL) -> Debezium -> Redpanda -> ClickHouse
- ClickHouse init.sql creates Kafka engine + MergeTree + materialized views
- CDC validation script in tests/cdc/validate.sh
- Analytics service has ClickHouse + Postgres + in-memory fallback chain

## Database Migrations
- 4 migration files in db/migrations/
- Migration runner in docker-compose is simple shell loop with `|| true` (errors swallowed)

## Client Architecture
- Svelte 4 + Tailwind in client/ui/
- Rust crypto core in client/src-tauri/src/lib.rs
- Tauri bridge commands: generate_keys, encrypt_message, decrypt_message
- Crypto: X25519 DH -> SHA256 hash -> ChaCha20-Poly1305
- All crypto logic properly isolated in Rust, exposed via Tauri commands

## File Paths
- Gateway: `/services/gateway/internal/httpserver/`
- Auth: `/services/auth/internal/`
- KeyDist: `/services/keydist/internal/`
- Messaging: `/services/messaging/internal/`
- Analytics: `/services/analytics/internal/`
- Workspace: `/services/workspace/internal/`
- Migrations: `/db/migrations/`
- Docker: `/deploy/docker-compose.yml`, `/deploy/docker-compose.cdc.yml`
- Taskfile: `/Taskfile.yml`
- CI: `/.github/workflows/ci.yml`

## Phase 4 & 5 Reality Check (March 23, 2026)
**Status**: Claimed complete, actually 70% complete with critical blockers

### Phase 4 - Workspace/Channel (Backend: 100%, Frontend: 90%)
**Implemented**:
- Workspace service fully built (create, list, join, members, channels endpoints)
- All frontend components exist (Sidebar, Workspace/Channel modals, MembersPanel)
- DB migration 003 deployed (workspace/channel enhancements)
- Role enforcement (owner/admin) working in handlers
- Both in-memory and Postgres backends in workspace store

**BLOCKER - Port Mapping Chaos**:
- docker-compose.yml has REVERSED port mappings:
  - keydist: internal 8082 → external 8083 (WRONG, should be 8082)
  - messaging: internal 8083 → external 8082 (WRONG, should be 8083)
- Frontend api.js hardcodes: MESSAGING_BASE=localhost:8082, KEYDIST_BASE=localhost:8083
- These are BACKWARDS relative to service internal ports
- When user runs `task ui:dev` alone, frontend tries to reach services and gets ERR_CONNECTION_REFUSED

**Missing Pieces**:
- Workspace service uses same hardcoded JWT secret pattern as auth (NOT YET FIXED)
- No algorithm check in workspace authMiddleware (auth already has it in line 386)

### Phase 5 - Rich Messaging (Backend: 100%, Frontend: 95%)
**Implemented**:
- All 9 rich message endpoints in messaging service (threads, reactions, pins, edit, delete)
- DB migration 004 deployed (threads, reactions, pins tables)
- API wiring complete in api.js (9 functions: postThreadReply, addReaction, pinMessage, editMessage, etc.)
- Frontend components: ThreadPanel, EmojiPicker, MessageBubble hover actions
- All routes registered in RegisterRichRoutes()

**INHERITED BLOCKER**: Same port confusion from Phase 4 breaks rich message APIs
- All rich endpoints use MESSAGING_BASE which is hardcoded to wrong port in api.js

**Missing Pieces**:
- ThreadPanel not visually tested (component exists but integration unclear)
- MarkdownText and LinkPreview components exist but integration testing needed
- Emoji picker styling may not match dark theme overhaul

## Immediate Phase 6 Blockers
1. **Port mapping fix** - Critical, blocks all service communication
2. **JWT secret removal** - workspace service still has hardcoded secret
3. **Algorithm check** - workspace authMiddleware missing algorithm validation
4. **Migration down.sql** - verify migration 004 has proper down.sql (check if exists)
5. **Full stack test** - `task dev` → verify all services respond on correct ports → `task ui:dev` should work

After these fixes, can proceed to Phase 6 (File Attachments).
