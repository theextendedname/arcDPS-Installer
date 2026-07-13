#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_dir"

echo "Building Windows binary..."
GOOS=windows GOARCH=amd64 go build \
  -o ./Release/arcDPS-Installer.exe \
  -ldflags "-X main.version=0.1.6 -s -w" .

echo "Building Linux binary..."
GOOS=linux GOARCH=amd64 go build \
  -o ./Release/arcDPS-Installer-linux \
  -ldflags "-X main.version=0.1.6 -s -w" .

echo "Build complete:"
ls -lh ./Release/arcDPS-Installer*
