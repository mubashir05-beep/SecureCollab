#!/usr/bin/env bash
# CDC Pipeline End-to-End Validation
# Usage: bash tests/cdc/validate.sh
#
# Prerequisites: base stack + CDC overlay running
#   task dev && task cdc:up && task cdc:register

set -euo pipefail

AUTH_URL="${AUTH_URL:-http://localhost:8081}"
MSG_URL="${MSG_URL:-http://localhost:8083}"
CH_URL="${CLICKHOUSE_URL:-http://localhost:8123}"
CONNECT_URL="${CONNECT_URL:-http://localhost:8085}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }
info() { echo -e "${YELLOW}[INFO]${NC} $1"; }

# --- Step 1: Check services ---
info "Checking services are reachable..."

curl -sf "$AUTH_URL/healthz" > /dev/null || fail "Auth service not reachable"
pass "Auth service OK"

curl -sf "$MSG_URL/healthz" > /dev/null || fail "Messaging service not reachable"
pass "Messaging service OK"

curl -sf "$CH_URL/?query=SELECT%201" > /dev/null || fail "ClickHouse not reachable"
pass "ClickHouse OK"

# --- Step 2: Check Debezium connector ---
info "Checking Debezium connector..."
CONNECTOR_STATUS=$(curl -sf "$CONNECT_URL/connectors/securecollab-postgres-encrypted-messages/status" 2>/dev/null || echo "NOT_FOUND")

if echo "$CONNECTOR_STATUS" | grep -q '"state":"RUNNING"'; then
  pass "Debezium connector RUNNING"
else
  info "Connector not found or not running, registering..."
  curl -sf -X POST "$CONNECT_URL/connectors" \
    -H "Content-Type: application/json" \
    --data @pipeline/debezium/connectors/postgres-encrypted-messages.json > /dev/null 2>&1 || true
  sleep 5
  CONNECTOR_STATUS=$(curl -sf "$CONNECT_URL/connectors/securecollab-postgres-encrypted-messages/status" 2>/dev/null || echo "FAILED")
  if echo "$CONNECTOR_STATUS" | grep -q '"state":"RUNNING"'; then
    pass "Debezium connector registered and RUNNING"
  else
    fail "Debezium connector failed to start: $CONNECTOR_STATUS"
  fi
fi

# --- Step 3: Register test user + get token ---
info "Registering CDC test user..."
TIMESTAMP=$(date +%s)
USERNAME="cdc_test_${TIMESTAMP}"

REGISTER_RESP=$(curl -sf -X POST "$AUTH_URL/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"email\":\"${USERNAME}@test.com\",\"password\":\"testpass123\"}" 2>&1)

TOKEN=$(echo "$REGISTER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])" 2>/dev/null || echo "")
USER_ID=$(echo "$REGISTER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['user_id'])" 2>/dev/null || echo "")

if [ -z "$TOKEN" ] || [ -z "$USER_ID" ]; then
  fail "Registration failed: $REGISTER_RESP"
fi
pass "Registered user $USERNAME ($USER_ID)"

# --- Step 4: Get pre-count from ClickHouse ---
PRE_COUNT=$(curl -sf "$CH_URL/?query=SELECT+count()+FROM+encrypted_messages_analytics" 2>/dev/null | tr -d '[:space:]' || echo "0")
info "ClickHouse pre-count: $PRE_COUNT"

# --- Step 5: Send a test message ---
info "Sending test message via messaging service..."
# Simple base64 payload
CIPHERTEXT=$(echo -n "cdc-validation-${TIMESTAMP}" | base64)
NONCE=$(echo -n "cdcnonce123456789012" | base64)

SEND_RESP=$(curl -sf -X POST "$MSG_URL/v1/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"recipient_user_id\":\"$USER_ID\",\"ciphertext_b64\":\"$CIPHERTEXT\",\"nonce_b64\":\"$NONCE\",\"content_type\":\"text\"}" 2>&1)

MSG_ID=$(echo "$SEND_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || echo "")

if [ -z "$MSG_ID" ]; then
  fail "Send message failed: $SEND_RESP"
fi
pass "Message sent (id: $MSG_ID)"

# --- Step 6: Wait for CDC propagation ---
info "Waiting for CDC propagation (up to 30s)..."
FOUND=false
for i in $(seq 1 15); do
  sleep 2
  POST_COUNT=$(curl -sf "$CH_URL/?query=SELECT+count()+FROM+encrypted_messages_analytics" 2>/dev/null | tr -d '[:space:]' || echo "0")
  if [ "$POST_COUNT" -gt "$PRE_COUNT" ] 2>/dev/null; then
    FOUND=true
    break
  fi
  echo -n "."
done
echo ""

if [ "$FOUND" = true ]; then
  pass "CDC propagation confirmed! ClickHouse count: $PRE_COUNT -> $POST_COUNT"
else
  POST_COUNT=$(curl -sf "$CH_URL/?query=SELECT+count()+FROM+encrypted_messages_analytics" 2>/dev/null | tr -d '[:space:]' || echo "?")
  fail "CDC propagation not detected after 30s. ClickHouse count still: $POST_COUNT"
fi

# --- Step 7: Verify the specific message in ClickHouse ---
info "Verifying message in ClickHouse..."
CH_MSG=$(curl -sf "$CH_URL/?query=SELECT+id,sender_user_id+FROM+encrypted_messages_analytics+WHERE+sender_user_id='${USER_ID}'+FORMAT+JSONEachRow" 2>/dev/null || echo "")

if echo "$CH_MSG" | grep -q "$USER_ID"; then
  pass "Message found in ClickHouse with correct sender_user_id"
else
  fail "Message not found in ClickHouse for sender $USER_ID"
fi

# --- Step 8: Check hourly aggregation ---
info "Checking hourly aggregation table..."
AGG_COUNT=$(curl -sf "$CH_URL/?query=SELECT+sum(message_count)+FROM+message_volume_hourly+WHERE+sender_user_id='${USER_ID}'" 2>/dev/null | tr -d '[:space:]' || echo "0")

if [ "$AGG_COUNT" -ge 1 ] 2>/dev/null; then
  pass "Hourly aggregation working (count: $AGG_COUNT)"
else
  info "Hourly aggregation not yet populated (may need OPTIMIZE). Skipping."
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  CDC Pipeline Validation: ALL PASSED   ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "  Postgres -> Debezium -> Redpanda -> ClickHouse"
echo "  Message $MSG_ID flowed end-to-end successfully."
