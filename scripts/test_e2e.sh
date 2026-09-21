#!/bin/bash
# E2E Test: Post #CariMaterial → trigger NLP match → verify notifications

set -e

API="http://localhost:9090/api"
NLP_HEALTH="http://localhost:8001/health"

echo "=== 1. Cek NLP Service Health ==="
curl -s $NLP_HEALTH | jq '.'
echo ""

echo "=== 2. Cek Backend Health ==="
curl -s $API/health || echo "Backend health endpoint belum ada, skip"
echo ""

echo "=== 3. Register user (kalau belum ada) ==="
REGISTER_RESP=$(curl -s -X POST $API/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User E2E",
    "username": "test_e2e",
    "phone": "08123456789",
    "email": "test_e2e@kana.id",
    "password": "TestPassword123!"
  }' || echo '{"data":{}}')
echo "$REGISTER_RESP" | jq '.' || true
echo ""

echo "=== 4. Login ==="
LOGIN_RESP=$(curl -s -X POST $API/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_e2e",
    "password": "TestPassword123!"
  }')

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.token // .token // empty')
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

POST_ID=$(echo "$POST_RESP" | jq -r '.data.id // empty')
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
echo "$NOTIF_RESP" | jq '.data[:3]' || echo "Response: $NOTIF_RESP"
echo ""

echo "=== 8. Cek matches di database (via SQL) ==="
echo "Jalankan manual: "
echo "  psql \$DATABASE_URL -c \"SELECT id, request_id, listing_id, final_score, rank FROM matches WHERE request_id = '$POST_ID' ORDER BY rank;\""
echo ""

echo "=== TEST SELESAI ==="
