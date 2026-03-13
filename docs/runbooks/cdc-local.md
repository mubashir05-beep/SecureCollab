# CDC Local Runbook

## Purpose
Run a local CDC pipeline scaffold for SecureCollab using Postgres WAL capture, Kafka-compatible transport, and ClickHouse.

## Prerequisites
- Docker Desktop running
- Base stack started with `task dev`
- `.env` present in repository root

## Start CDC Overlay
```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.cdc.yml up -d redpanda connect clickhouse
```

## Verify CDC Components
```bash
# Kafka-compatible broker metadata endpoint via Redpanda proxy
curl http://localhost:8082/brokers

# Kafka Connect API
curl http://localhost:${CONNECT_PORT:-8085}/connectors

# ClickHouse ping
curl "http://localhost:${CLICKHOUSE_HTTP_PORT:-8123}/ping"
```

## Register Debezium Connector
```bash
curl -X POST http://localhost:${CONNECT_PORT:-8085}/connectors \
  -H "Content-Type: application/json" \
  --data @pipeline/debezium/connectors/postgres-encrypted-messages.json
```

## Check Connector Status
```bash
curl http://localhost:${CONNECT_PORT:-8085}/connectors/securecollab-postgres-encrypted-messages/status
```

## Stop CDC Overlay
```bash
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.cdc.yml down
```

## Notes
- This is a Phase 3 bootstrap scaffold for CDC infrastructure.
- ClickHouse sink wiring is intentionally deferred to a later step.
- Base Postgres is configured for Debezium with `wal_level=logical`, `max_wal_senders=10`, and `max_replication_slots=10` in `deploy/docker-compose.yml`.
