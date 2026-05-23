#!/usr/bin/env bash
# Assigns the Keycloak realm-level "admin" role to a user in the ss-project realm.
# Runs inside the keycloak container via docker exec (no host port exposure needed).
# Usage: ./scripts/assign-admin-role.sh <username>

set -euo pipefail

USERNAME="${1:?Usage: $0 <username>}"
CONTAINER="${KEYCLOAK_CONTAINER:-keycloak}"
REALM="ss-project"
ADMIN_USER="${KEYCLOAK_ADMIN_USER:-admin}"
ADMIN_PASS="${KEYCLOAK_ADMIN_PASS:-admin}"
KCADM="/opt/keycloak/bin/kcadm.sh"
CA_CERT="/opt/keycloak/conf/certs/ca.pem"
TRUSTSTORE="/tmp/kcadm-truststore.jks"
TRUSTPASS="changeit"

echo "-> Building truststore from CA cert..."
docker exec "$CONTAINER" sh -c "
  keytool -import -noprompt -alias local-ca \
    -keystore $TRUSTSTORE \
    -storepass $TRUSTPASS \
    -file $CA_CERT 2>/dev/null || true
"

KC_OPTS="-Djavax.net.ssl.trustStore=$TRUSTSTORE -Djavax.net.ssl.trustStorePassword=$TRUSTPASS"

echo "-> Authenticating as Keycloak admin..."
docker exec "$CONTAINER" sh -c "
  KC_OPTS='$KC_OPTS' $KCADM config credentials \
    --server https://localhost:8443/auth \
    --realm master \
    --user '$ADMIN_USER' \
    --password '$ADMIN_PASS'
"

echo "-> Assigning 'admin' role to '$USERNAME' in realm '$REALM'..."
docker exec "$CONTAINER" sh -c "
  KC_OPTS='$KC_OPTS' $KCADM add-roles \
    --uusername '$USERNAME' \
    --rolename admin \
    -r '$REALM'
"

echo "Done. '$USERNAME' now has the 'admin' role."
echo "Log out and back in so the new token includes the role."

