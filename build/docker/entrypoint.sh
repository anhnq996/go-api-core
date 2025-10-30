#!/bin/sh
set -e

echo "🚀 Entrypoint starting..."

# Create required directories if missing
mkdir -p /app/storages/app || true
mkdir -p /app/storages/log || true

# Optional: auto migrate when flag is set
if [ "$AUTO_MIGRATE" = "true" ]; then
  echo "📦 Running migrations via make (AUTO_MIGRATE=true) ..."
  if command -v make >/dev/null 2>&1; then
    make -C /app migrate
    echo "✅ Migrations done"
  else
    echo "❌ make not found; cannot run migrations"
  fi
fi

echo "🎯 Executing: $@"
exec "$@"


