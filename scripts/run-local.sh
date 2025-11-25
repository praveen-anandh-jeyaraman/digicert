#!/usr/bin/env bash
set -euo pipefail
export DATABASE_URL=${DATABASE_URL:-"postgres://library:librarypass@localhost:5432/library?sslmode=disable"}
export PORT=${PORT:-8080}
go run ./cmd/library-api
