#!/bin/sh
set -e
cd "$(dirname "$0")"
if ! command -v go >/dev/null 2>&1; then
  echo "Go 1.23+ required. Install: pkg install golang" >&2
  exit 1
fi
CGO_ENABLED=0 go build -ldflags="-s -w" -o gemini-api .
echo "Built: ./gemini-api"
