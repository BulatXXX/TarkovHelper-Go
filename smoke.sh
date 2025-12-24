#!/usr/bin/env bash
set -euo pipefail

PORT="${APP_PORT:-8080}"
BASE="http://localhost:${PORT}"

EMAIL="${TEST_EMAIL:-user@mail.com}"
PASS="${TEST_PASS:-secret12345}"
NAME="${TEST_NAME:-Singularity}"

MODE="${TEST_MODE:-pvp}"

# маленькая утилита: достать поле из JSON через python
json_get() {
  local key="$1"
  python3 -c 'import json,sys
key=sys.argv[1]
try:
    data=json.load(sys.stdin)
except Exception:
    print("")
    sys.exit(0)

cur=data
for part in key.split("."):
    if isinstance(cur, dict) and part in cur:
        cur=cur[part]
    else:
        print("")
        sys.exit(0)

print("" if cur is None else cur)
' "$key"
}

# ассерт, что в ответе есть подстрока
assert_contains() {
  local hay="$1"
  local needle="$2"
  if [[ "$hay" != *"$needle"* ]]; then
    echo "❌ ASSERT FAILED: expected to contain: $needle"
    echo "---- actual ----"
    echo "$hay"
    echo "----------------"
    exit 1
  fi
}

# ассерт равенства (строго)
assert_eq() {
  local a="$1"
  local b="$2"
  if [[ "$a" != "$b" ]]; then
    echo "❌ ASSERT FAILED: expected '$b', got '$a'"
    exit 1
  fi
}

echo "==> 0) Ensure .env exists (optional)"
if [[ ! -f .env ]]; then
  if [[ -f .env.example ]]; then
    cp .env.example .env
    echo "   created .env from .env.example"
  else
    echo "   .env not found (ok), continuing"
  fi
fi

echo "==> 1) Build & up"
docker compose up --build -d

echo "==> 2) Wait for /healthz"
for i in {1..60}; do
  if curl -sS "${BASE}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

HEALTH="$(curl -sS "${BASE}/healthz")"
assert_contains "$HEALTH" '"status"'
assert_contains "$HEALTH" '"ok"'
echo "   health: $HEALTH"

echo "==> 3) Register (or accept conflict)"
REG_BODY="$(curl -sS -X POST "${BASE}/auth/register" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\",\"name\":\"${NAME}\"}" \
  || true)"

# Если email уже есть, ваш API вернёт JSON с code EMAIL_TAKEN и статус 409.
# curl по умолчанию не падает на 409, но может упасть из-за сетевых.
if [[ "$REG_BODY" == *'"token"'* ]]; then
  TOKEN="$(printf '%s' "$REG_BODY" | json_get token)"
  echo "   registered ok"
else
  # делаем login
  echo "   register not returned token (maybe already registered). logging in..."
  LOGIN_BODY="$(curl -sS -X POST "${BASE}/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")"
  TOKEN="$(printf '%s' "$LOGIN_BODY" | json_get token)"
fi

if [[ -z "${TOKEN}" ]]; then
  echo "❌ Failed to obtain JWT token"
  echo "REG_BODY: $REG_BODY"
  exit 1
fi
echo "   token ok (len=${#TOKEN})"

echo "==> 4) /me"
ME="$(curl -sS "${BASE}/me" -H "Authorization: Bearer ${TOKEN}")"
ME_EMAIL="$(printf '%s' "$ME" | json_get user.email)"
ME_NAME="$(printf '%s' "$ME" | json_get user.name)"
assert_eq "$ME_EMAIL" "$EMAIL"
assert_eq "$ME_NAME" "$NAME"
echo "   me ok: email=$ME_EMAIL name=$ME_NAME"

echo "==> 5) GET /tracked (should be array, may be empty)"
T0="$(curl -sS "${BASE}/tracked?mode=${MODE}" -H "Authorization: Bearer ${TOKEN}")"
assert_contains "$T0" '"items"'
echo "   tracked initial: $T0"

echo "==> 6) PUT /tracked"
NOW_MS="$(python3 - <<'PY'
import time
print(int(time.time()*1000))
PY
)"

PUT_PAYLOAD="$(cat <<JSON
{
  "items": [
    {
      "id": "5c94bbff86f7747ee735c08f",
      "iconLink": "https://assets.tarkov.dev/5c94bbff86f7747ee735c08f-icon.webp",
      "updatedAt": ${NOW_MS}
    },
    {
      "id": "59faff1d86f7746c51718c9c",
      "iconLink": null,
      "updatedAt": ${NOW_MS}
    }
  ]
}
JSON
)"

PUT_RES="$(curl -sS -X PUT "${BASE}/tracked?mode=${MODE}" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "${PUT_PAYLOAD}")"

assert_contains "$PUT_RES" '5c94bbff86f7747ee735c08f'
assert_contains "$PUT_RES" '59faff1d86f7746c51718c9c'
echo "   put ok"

echo "==> 7) GET /tracked (must contain both ids)"
T1="$(curl -sS "${BASE}/tracked?mode=${MODE}" -H "Authorization: Bearer ${TOKEN}")"
assert_contains "$T1" '5c94bbff86f7747ee735c08f'
assert_contains "$T1" '59faff1d86f7746c51718c9c'
echo "   get after put ok"

echo "==> 8) Restart containers (down/up without -v) to test persistence"
docker compose down
docker compose up -d

echo "==> 9) Wait for /healthz again"
for i in {1..60}; do
  if curl -sS "${BASE}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "==> 10) Login again (token might be invalid after restart if secret changed)"
LOGIN_BODY2="$(curl -sS -X POST "${BASE}/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASS}\"}")"
TOKEN2="$(printf '%s' "$LOGIN_BODY2" | json_get token)"
if [[ -z "${TOKEN2}" ]]; then
  echo "❌ Failed to login after restart"
  echo "$LOGIN_BODY2"
  exit 1
fi

echo "==> 11) GET /tracked again - MUST still contain ids"
T2="$(curl -sS "${BASE}/tracked?mode=${MODE}" -H "Authorization: Bearer ${TOKEN2}")"
assert_contains "$T2" '5c94bbff86f7747ee735c08f'
assert_contains "$T2" '59faff1d86f7746c51718c9c'

echo
echo "✅ SMOKE TEST PASSED"
echo "Tracked persisted across container recreate (volume OK)."