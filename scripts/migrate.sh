#!/usr/bin/env bash
set -euo pipefail
if [ -z "${DATABASE_URL:-}" ]; then
  echo "set DATABASE_URL env var"
  exit 1
fi
docker run --rm -v "$PWD/internal/migrate":/migrations --network host migrate/migrate \
  -path=/migrations -database "${DATABASE_URL}" up
