#!/usr/bin/env bash

set -euo pipefail

REALM=${REALM:-ss-project}
SERVICE=${SERVICE:-keycloak}
COMPOSE_FILES=${COMPOSE_FILES:-"-f docker-compose.yml -f docker-compose.optimized.yml"}
OUT_DIR=${OUT_DIR:-./keycloak/imports}

mkdir -p "$OUT_DIR"

echo "-> stopping service '$SERVICE' so export can acquire the DB"
docker compose $COMPOSE_FILES stop "$SERVICE" >/dev/null

echo "-> running one-shot export container"
docker compose $COMPOSE_FILES run --rm \
  --entrypoint /opt/keycloak/bin/kc.sh \
  -v "$(pwd)/${OUT_DIR}:/tmp/realm-out" \
  "$SERVICE" \
  export --dir /tmp/realm-out --realm "$REALM" --users realm_file

echo "-> restarting service '$SERVICE'"
docker compose $COMPOSE_FILES start "$SERVICE" >/dev/null

echo "Exported: ${OUT_DIR}/${REALM}-realm.json"
ls -l "${OUT_DIR}/${REALM}-realm.json"