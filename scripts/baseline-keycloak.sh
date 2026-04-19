#!/usr/bin/env bash
#
# Usage:
#   # Dev baseline (default compose file, start-dev + H2)
#   ./scripts/baseline-keycloak.sh dev
#
#   # Optimized run (override file, start --optimized)
#   COMPOSE_FILES="-f docker-compose.yml -f docker-compose.optimized.yml" \
#       ./scripts/baseline-keycloak.sh optimized

set -euo pipefail

COMPOSE_FILES=${COMPOSE_FILES:-"-f docker-compose.yml"}
SERVICE=${SERVICE:-keycloak}
CONTAINER=${CONTAINER:-keycloak}
HEALTH_CMD=${HEALTH_CMD:-"docker exec $CONTAINER curl -fsS -o /dev/null http://localhost:9000/health/ready"}
TIMEOUT=${TIMEOUT:-300}  # seconds

LABEL=${1:-run}
TS=$(date +%Y%m%d-%H%M%S)
OUT_DIR=${OUT_DIR:-".bench/keycloak"}
mkdir -p "$OUT_DIR"
OUT_FILE="$OUT_DIR/${LABEL}-${TS}.txt"

log() { echo "$@" | tee -a "$OUT_FILE"; }

wait_ready() {
  local start_ts=$1
  while ! eval "$HEALTH_CMD" 2>/dev/null; do
    sleep 1
    local now=$(date +%s)
    if (( now - start_ts > TIMEOUT )); then
      log "ERROR: $HEALTH_URL not ready after ${TIMEOUT}s"
      docker compose $COMPOSE_FILES logs --tail=100 "$SERVICE" | tee -a "$OUT_FILE" || true
      exit 1
    fi
  done
}

log "============================================================"
log "Keycloak baseline run"
log "Label:         $LABEL"
log "Timestamp:     $TS"
log "Compose:       $COMPOSE_FILES"
log "Service:       $SERVICE"
log "Health URL:    $HEALTH_URL"
log "Docker host:   $(docker version -f '{{.Server.Os}}/{{.Server.Arch}}' 2>/dev/null || echo 'unknown')"
log "============================================================"

# Start from a clean slate for the measured service
log "-> tearing down service before cold start"
docker compose $COMPOSE_FILES rm -sfv "$SERVICE" >/dev/null 2>&1 || true

# Cold start 
log "-> cold start: docker compose up -d $SERVICE"
t0=$(date +%s)
docker compose $COMPOSE_FILES up -d "$SERVICE" >/dev/null
wait_ready "$t0"
t1=$(date +%s)
cold=$(( t1 - t0 ))
log "Cold start ready in: ${cold}s"

# Restart 
log "-> restart: docker compose restart $SERVICE"
r0=$(date +%s)
docker compose $COMPOSE_FILES restart "$SERVICE" >/dev/null
wait_ready "$r0"
r1=$(date +%s)
restart=$(( r1 - r0 ))
log "Restart ready in:    ${restart}s"

# Memory / CPU snapshot
sleep 5
log "-> docker stats snapshot (after 5s warm-up)"
docker stats --no-stream \
  --format "table {{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.CPUPerc}}" \
  "$CONTAINER" | tee -a "$OUT_FILE"

log "============================================================"
log "Done. Output: $OUT_FILE"
log "============================================================"