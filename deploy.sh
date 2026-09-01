#!/bin/bash
set -e

echo "Deploy KANA Backend"

# Cek apakah .env ada
if [ ! -f .env ]; then
  echo "❌ File .env tidak ditemukan! Copy .env.example ke .env dan isi secret."
  exit 1
fi

# Pull dari Git
if [ -d ".git" ]; then
  echo "Pull from Git..."
  git pull origin main 2>/dev/null || echo "⚠️Git pull gagal"
fi

# Build & restart container
echo "Building Docker images..."
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Migrasi
echo "Running database migration..."
docker-compose exec app ./main migrate

# Health check
sleep 3
echo "Health check..."
if curl -f http://localhost:9090/health > /dev/null 2>&1; then
  echo "Health check OK✅"
else
  echo "❌ Health check gagal! Cek log: docker-compose logs app"
  exit 1
fi

echo "✅Deploy selesai! Aplikasi berjalan di http://localhost:9090"
