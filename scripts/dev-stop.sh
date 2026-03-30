#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${ROOT_DIR}/.dev-runtime"
PID_FILE="${LOG_DIR}/client.pid"

if [[ -f "${PID_FILE}" ]]; then
  CLIENT_PID="$(cat "${PID_FILE}")"
  if kill -0 "${CLIENT_PID}" >/dev/null 2>&1; then
    echo "Stopping client dev server (PID ${CLIENT_PID})..."
    kill "${CLIENT_PID}" || true
    wait "${CLIENT_PID}" 2>/dev/null || true
  else
    echo "Client dev server process ${CLIENT_PID} not running."
  fi
  rm -f "${PID_FILE}"
else
  echo "No client dev server PID file found; assuming it is not running."
fi

echo "Stopping docker compose services..."
cd "${ROOT_DIR}"
env "UID=$(id -u)" "GID=$(id -g)" docker compose down

echo "All services stopped."

