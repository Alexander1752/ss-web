#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ ! -f "${ROOT_DIR}/.env" ]]; then
  echo "Missing .env in ${ROOT_DIR}. Copy or create it before starting the stack." >&2
  exit 1
fi

echo "Starting docker compose services..."
cd "${ROOT_DIR}"
env "UID=$(id -u)" "GID=$(id -g)" docker compose up -d

echo "Stack is ready. Press Ctrl+C has no effect; use scripts/dev-stop.sh to stop everything."

