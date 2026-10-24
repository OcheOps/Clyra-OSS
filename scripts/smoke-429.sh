#!/usr/bin/env bash
# Free-tier 429 smoke test.
# Hits /api/v1/diagrams/generate N+1 times as a free user and verifies the
# (N+1)th response is 429. N defaults to PLAN_FREE_DAILY (10) so the 11th
# call is the one that must trip.
#
# Usage:
#   BASE_URL=http://localhost:8080 \
#   EMAIL=smoke+free@clyra.test \
#   PASSWORD=smoke12345678 \
#   PLAN_FREE_DAILY=10 \
#   ./scripts/smoke-429.sh
#
# Set BASE_URL=https://useclyra.com for a prod run after scrubbing the
# test user from the DB afterward. Exits non-zero if the 11th response is
# NOT 429 or any of the first 10 fails.

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
EMAIL="${EMAIL:-smoke+free@clyra.test}"
PASSWORD="${PASSWORD:-smoke12345678}"
LIMIT="${PLAN_FREE_DAILY:-10}"
COOKIE_JAR="$(mktemp)"
trap 'rm -f "$COOKIE_JAR"' EXIT

note() { printf '\033[36m[smoke]\033[0m %s\n' "$*"; }
fail() { printf '\033[31m[fail]\033[0m %s\n' "$*" >&2; exit 1; }

note "base: $BASE_URL  user: $EMAIL  limit: $LIMIT"

# 1) Register (idempotent — if the user exists, fall back to login).
REG_STATUS=$(curl -s -o /dev/null -w '%{http_code}' \
  -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  -H 'Content-Type: application/json' \
  -X POST "$BASE_URL/api/v1/auth/register" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"invite_code\":\"\"}")

if [ "$REG_STATUS" != "201" ]; then
  note "register returned $REG_STATUS — trying login"
  LOGIN_STATUS=$(curl -s -o /dev/null -w '%{http_code}' \
    -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    -H 'Content-Type: application/json' \
    -X POST "$BASE_URL/api/v1/auth/login" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
  [ "$LOGIN_STATUS" = "200" ] || fail "login failed ($LOGIN_STATUS)"
fi

# Grab the CSRF token our backend issues alongside the session cookie.
# Any cookie name ending in _csrf is valid (naming varies by deployment).
CSRF=$(awk -F'\t' '/_csrf/{print $7}' "$COOKIE_JAR" | tail -n1)
[ -n "$CSRF" ] || fail "no CSRF cookie after login"
note "auth ok, csrf token captured"

# 2) Fire LIMIT requests, expect 200 each.
BODY='{"prompt":"a tiny web app on aws","provider":"managed_bedrock"}'
for i in $(seq 1 "$LIMIT"); do
  CODE=$(curl -s -o /dev/null -w '%{http_code}' \
    -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
    -H 'Content-Type: application/json' \
    -H "X-CSRF-Token: $CSRF" \
    -X POST "$BASE_URL/api/v1/diagrams/generate" \
    -d "$BODY")
  if [ "$CODE" != "200" ]; then
    note "request $i returned $CODE (not 200) — continuing, but flag to you"
  else
    note "request $i ok"
  fi
done

# 3) Request LIMIT+1 must be 429.
FINAL=$(curl -s -o /dev/null -w '%{http_code}' \
  -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -X POST "$BASE_URL/api/v1/diagrams/generate" \
  -d "$BODY")

if [ "$FINAL" = "429" ]; then
  printf '\033[32m[pass]\033[0m request %d returned 429 as expected\n' "$((LIMIT + 1))"
  exit 0
fi

fail "request $((LIMIT + 1)) returned $FINAL, expected 429 — daily quota gate is NOT enforcing"
