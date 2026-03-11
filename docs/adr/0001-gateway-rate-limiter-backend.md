# ADR 0001: Gateway Rate Limiter Backend Selection

## Context
Phase 1 requires sliding-window and token-bucket rate limiting with Redis backing. During local development, Redis may be unavailable temporarily, and hard-failing all protected routes would block development and tests.

## Decision
Implement both algorithms with a Redis backend and keep an in-memory fallback used when Redis is not configured or cannot be reached at startup.

## Consequences
- Production-like behavior is available by setting `REDIS_ADDR`.
- Local development remains operational if Redis is down.
- Rate-limit state consistency differs in fallback mode, so milestone/load validation should run with Redis enabled.

## Alternatives Considered
- Redis-only with startup hard fail: strongest consistency, but poor local DX and brittle test runs.
- In-memory only: simple, but does not meet the Redis-backed Phase 1 target.
