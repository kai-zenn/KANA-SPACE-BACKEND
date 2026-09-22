#!/bin/bash
# E2E Test: Post #CariMaterial → trigger NLP match → verify notifications

set -e

API="http://localhost:9090/api"
NLP_HEALTH="http://localhost:8001/health"

json_print() {
  if command -v jq >/dev/null 2>&1; then
    jq '.'
  else
    python -c 'import sys, json; print(json.dumps(json.load(sys.stdin), indent=2, ensure_ascii=False))'
  fi
}

json_get_field() {
  key="$1"
  if command -v jq >/dev/null 2>&1; then
    jq -r "$key"
  else
    python - "$key" <<'PY'
import sys, json
key = sys.argv[1]
obj = json.load(sys.stdin)
value = obj
for part in key.split('.'):
    if isinstance(value, dict):
        value = value.get(part)
    else:
        value = None
        break
if value is None:
    print('')
else:
    print(value)
PY
  fi
}

echo "=== 1. Cek NLP Service Health ==="
curl -s $NLP_HEALTH | json_print
echo ""

echo "=== 2. Cek Backend Health ==="
# backend belum punya endpoint /api/health; skip bila 404
curl -s -o /tmp/kana_backend_health.txt -w '%{http_code}' $API/health || true
printf '\n'
cat /tmp/kana_backend_health.txt 2>/dev/null || true
echo ""

echo "=== 3. Register user (kalau belum ada) ==="
REGISTER_RESP=$(curl -s -X POST $API/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Test",
    "last_name": "E2E",
    "username": "test_e2e",
    "phone_number": "08123456789",
    "email": "test_e2e@kana.id",
    "password": "TestPassword123!"
  }' || echo '{"data":{}}')
printf '%s\n' "$REGISTER_RESP" | json_print || true
echo ""

echo "=== 4. Login ==="
LOGIN_RESP=$(curl -s -X POST $API/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_e2e",
    "password": "TestPassword123!"
  }')

TOKEN=$(printf '%s\n' "$LOGIN_RESP" | python -c 'import sys, json; d = json.load(sys.stdin); print((d.get("data") or {}).get("token") or d.get("token") or "")')
if [ -z "$TOKEN" ]; then
  echo "ERROR: Login gagal. Response: $LOGIN_RESP"
  exit 1
fi
echo "Token berhasil didapat: ${TOKEN:0:30}..."
echo ""

echo "=== 5. Post #CariMaterial ==="
POST_RESP=$(curl -s -X POST $API/space/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "butuh kain perca batik untuk produksi tas, 5kg minimal",
    "tag": "CariMaterial",
    "latitude": -6.225,
    "longitude": 106.9
  }')

POST_ID=$(printf '%s\n' "$POST_RESP" | python -c 'import sys, json; d = json.load(sys.stdin); print((d.get("data") or {}).get("id") or d.get("id") or "")')
if [ -z "$POST_ID" ] || [ "$POST_ID" = "null" ]; then
  echo "ERROR: Post gagal dibuat. Response: $POST_RESP"
  exit 1
fi
echo "Post berhasil dibuat: $POST_ID"
echo ""

echo "=== 6. Tunggu async NLP matching (5 detik) ==="
sleep 5

echo "=== 7. Cek notifikasi user ==="
NOTIF_RESP=$(curl -s $API/notifications \
  -H "Authorization: Bearer $TOKEN")
if command -v jq >/dev/null 2>&1; then
  echo "$NOTIF_RESP" | jq '.data[:3]' || echo "Response: $NOTIF_RESP"
else
  printf '%s\n' "$NOTIF_RESP" | python -c 'import sys, json; d=json.load(sys.stdin); arr = (d.get("data") or [])[:3]; print(json.dumps(arr, indent=2, ensure_ascii=False) if isinstance(arr, list) else json.dumps(d, indent=2, ensure_ascii=False))'
fi
echo ""

echo "=== 8. Cek matches di database (via SQL) ==="
echo "Jalankan manual: "
echo "  psql \$DATABASE_URL -c \"SELECT id, request_id, listing_id, final_score, rank FROM matches WHERE request_id = '$POST_ID' ORDER BY rank;\""
echo ""

echo "=== TEST SELESAI ==="
