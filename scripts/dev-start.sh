#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLIENT_DIR="${ROOT_DIR}/client"
LOG_DIR="${ROOT_DIR}/.dev-runtime"
PID_FILE="${LOG_DIR}/client.pid"
LOG_FILE="${LOG_DIR}/client.log"
UPLOAD_DIR="${ROOT_DIR}/uploads"

mkdir -p "${LOG_DIR}"

mkdir -p "${UPLOAD_DIR}"
chmod 755 "${UPLOAD_DIR}"

if [[ ! -f "${ROOT_DIR}/.env" ]]; then
  echo "Missing .env in ${ROOT_DIR}. Copy or create it before starting the stack." >&2
  exit 1
fi

echo "Installing client dependencies (if needed)..."
cd "${CLIENT_DIR}"
yarn install --check-files >/dev/null

echo "Starting docker compose services..."
cd "${ROOT_DIR}"
# Detect host IP for MQTT broker info
export HOST_IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}' || echo "localhost")
echo "Detected HOST_IP: ${HOST_IP}"
env "UID=$(id -u)" "GID=$(id -g)" docker compose up -d --build

if [[ -f "${PID_FILE}" ]]; then
  EXISTING_PID="$(cat "${PID_FILE}")"
  if kill -0 "${EXISTING_PID}" >/dev/null 2>&1; then
    echo "Client dev server already running with PID ${EXISTING_PID}."
    echo "Logs: ${LOG_FILE}"
    exit 0
  else
    rm -f "${PID_FILE}"
  fi
fi

echo "Starting Vite dev server on http://127.0.0.1:5173 ..."
cd "${CLIENT_DIR}"
CHOKIDAR_USEPOLLING=true yarn dev:poll --host 127.0.0.1 --port 5173 >"${LOG_FILE}" 2>&1 &
CLIENT_PID=$!
echo "${CLIENT_PID}" > "${PID_FILE}"

echo "Client dev server PID: ${CLIENT_PID}"
echo "Logs: ${LOG_FILE}"
echo "Stack is ready. Press Ctrl+C has no effect; use scripts/dev-stop.sh to stop everything."

