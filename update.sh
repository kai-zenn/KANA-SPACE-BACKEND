#!/bin/bash
# update.sh - Auto update dari GitHub (tanpa interaksi)

echo "Checking for updates..."

# Cek apakah ada perubahan di remote
git fetch origin main
LOCAL=$(git rev-parse HEAD)
REMOTE=$(git rev-parse origin/main)

if [ "$LOCAL" = "$REMOTE" ]; then
    echo "✅Sudah versi terbaru. Tidak ada update."
    exit 0
fi

echo "Ada update! Pulling changes..."
git pull origin main

echo "Rebuild & restart container..."
docker-compose down
docker-compose build --no-cache
docker-compose up -d

echo "Running migration..."
docker-compose exec app ./main migrate

echo "✅ Update selesai! Aplikasi berjalan di http://localhost:9090"
