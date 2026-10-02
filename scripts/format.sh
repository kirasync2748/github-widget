#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$ROOT_DIR"

if [ "${1:-}" = "--write" ]; then
  gofmt -w .
  echo "Files formatted."
else
  out=$(gofmt -l .)
  if [ -n "$out" ]; then
    echo "Unformatted files:"
    echo "$out"
    exit 1
  fi
  echo "All files formatted."
fi
