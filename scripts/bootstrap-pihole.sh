#!/usr/bin/env bash
set -euo pipefail

PIHOLES=("pihole1" "pihole2")
PASSWORDS=("pihole1" "pihole2")

BLOCK_DOMAINS=("ads.badsite.com" "tracker.example.net")
ALLOW_DOMAINS=("example.org" "openai.com")
TEST_DOMAINS=("${BLOCK_DOMAINS[@]}" "${ALLOW_DOMAINS[@]}")

echo "🔧 Bootstrapping Pi-hole development data..."

for i in "${!PIHOLES[@]}"; do
  HOST="${PIHOLES[$i]}"
  PORT=80
  PASS="${PASSWORDS[$i]}"
  BASE_URL="http://$HOST:$PORT"

  echo "🟡 Authenticating with $HOST..."

  # Authenticate and extract SID
  RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth" \
    -H "Content-Type: application/json" \
    -d "{\"password\": \"$PASS\"}")
  SID=$(echo "$RESPONSE" | jq -r '.session.sid')

  if [[ "$SID" == "null" || -z "$SID" ]]; then
    echo "❌ Failed to authenticate with $HOST"
    continue
  fi

  echo "✅ Authenticated with $HOST — SID: $SID"

  echo "🚀 Populating block/allow lists..."
  BLOCKED=0
  ALLOWED=0

  for domain in "${BLOCK_DOMAINS[@]}"; do
    RESPONSE=$(curl -s -w "|||%{http_code}" "$BASE_URL/api/domains/deny/exact" \
      -H "Content-Type: application/json" \
      -H "X-FTL-SID: $SID" \
      -d "{\"domain\": \"$domain\"}")
    BODY="${RESPONSE%|||*}"
    STATUS="${RESPONSE##*|||}"
    [[ "$STATUS" == "200" || "$STATUS" == "201" ]] && ((BLOCKED+=1)) || {
      echo "  ❌ Failed to block $domain (HTTP $STATUS)"
      echo "     ↳ $BODY"
    }
  done

  for domain in "${ALLOW_DOMAINS[@]}"; do
    RESPONSE=$(curl -s -w "|||%{http_code}" -X POST "$BASE_URL/api/domains/allow/exact" \
      -H "Content-Type: application/json" \
      -H "X-FTL-SID: $SID" \
      -d "{\"domain\": \"$domain\"}")
    BODY="${RESPONSE%|||*}"
    STATUS="${RESPONSE##*|||}"
    [[ "$STATUS" == "200" || "$STATUS" == "201" ]] && ((ALLOWED+=1)) || {
      echo "  ❌ Failed to allow $domain (HTTP $STATUS)"
      echo "     ↳ $BODY"
    }
  done

  echo "✅ Added $BLOCKED blocked and $ALLOWED allowed domains."

  echo "🔍 Verifying domain presence in config..."
  BLOCKED_DOMAINS=$(curl -s "$BASE_URL/api/domains/deny/exact" \
    -H "X-FTL-SID: $SID" | jq -r '.domains[].domain // empty')
  ALLOWED_DOMAINS=$(curl -s "$BASE_URL/api/domains/allow/exact" \
    -H "X-FTL-SID: $SID" | jq -r '.domains[].domain // empty')

  for domain in "${BLOCK_DOMAINS[@]}"; do
    if echo "$BLOCKED_DOMAINS" | grep -qx "$domain"; then
      echo "   ✅ $domain is present in block list"
    else
      echo "   ❌ $domain is NOT found in block list"
    fi
  done

  for domain in "${ALLOW_DOMAINS[@]}"; do
    if echo "$ALLOWED_DOMAINS" | grep -qx "$domain"; then
      echo "   ✅ $domain is present in allow list"
    else
      echo "   ❌ $domain is NOT found in allow list"
    fi
  done

  # Query log

  echo "🧪 Issuing test queries..."
  
  START_TIME=$(date +%s)
  sleep 1
  for domain in "${TEST_DOMAINS[@]}"; do
    dig @"$HOST" "$domain" > /dev/null 2>&1 || true
  done
  sleep 1
  END_TIME=$(date +%s)

  echo "📜 Verifying domain behavior in recent query log..."

  QUERY_LOG=$(curl -s "$BASE_URL/api/queries?from=$START_TIME&until=$END_TIME" -H "X-FTL-SID: $SID")
  echo "🛠️ Full raw query log between $START_TIME and $END_TIME:"

  for domain in "${BLOCK_DOMAINS[@]}"; do
    STATUS=$(echo "$QUERY_LOG" | jq -r --arg d "$domain" '.queries[] | select(.domain == $d) | .status' | head -n1)
    if [[ "$STATUS" == "BLOCKED" || "$STATUS" == "GRAVITY" || "$STATUS" == "DENYLIST" ]]; then
      echo "   ✅ $domain was blocked (status: $STATUS)"
    elif [[ -z "$STATUS" ]]; then
      echo "   ⚠️  $domain not found in query log"
    else
      echo "   ❌ $domain was NOT blocked (status: $STATUS)"
    fi
  done

  for domain in "${ALLOW_DOMAINS[@]}"; do
    STATUS=$(echo "$QUERY_LOG" | jq -r --arg d "$domain" '.queries[] | select(.domain == $d) | .status' | head -n1)
    if [[ "$STATUS" == "BLOCKED" || "$STATUS" == "GRAVITY" || "$STATUS" == "DENYLIST" ]]; then
      echo "   ❌ $domain was unexpectedly blocked (status: $STATUS)"
    elif [[ -z "$STATUS" ]]; then
      echo "   ⚠️  $domain not found in query log"
    else
      echo "   ✅ $domain was allowed (status: $STATUS)"
    fi
  done
done

echo "🔚 Logging out of $HOST..."
curl -s -X POST "$BASE_URL/api/logout" \
  -H "X-FTL-SID: $SID" > /dev/null || true

echo "✅ Pi-hole dev data bootstrapped."
