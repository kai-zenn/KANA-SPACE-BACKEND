#!/bin/bash
set -e

echo "🚀 Deploy KANA Backend"

# Pull terbaru dari Git
git pull origin main

# Build & restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Migrasi
docker-compose exec app ./main migrate

# Seeding (opsional)
# docker-compose exec app ./main seed

echo "✅ Deploy selesai! App running on port 9090"
