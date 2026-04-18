#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Stopping docker compose services..."
cd "${ROOT_DIR}"
env "UID=$(id -u)" "GID=$(id -g)" docker compose down

echo "All services stopped."

