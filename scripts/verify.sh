#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

echo "=== Formatting check ==="
scripts/format.sh

echo "=== go vet ==="
go vet ./...

echo "=== Tests ==="
go test ./...

echo "=== Race tests ==="
go test -race ./...

echo "=== Build ==="
go build ./...

echo "=== All checks passed ==="
