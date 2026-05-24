#!/bin/bash
# Convenience wrapper for starting the development environment

install_mkcert() {
    sudo apt update && sudo apt install libnss3-tools -y
    curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/amd64"
    chmod +x mkcert-v*-linux-amd64
    sudo mv mkcert-v*-linux-amd64 /usr/local/bin/mkcert
    mkcert -install
}

test "mkcert --version" &> /dev/null || {
    echo "mkcert not found, installing..."
    install_mkcert
}

rm -rf ./secrets 2>/dev/null
mkdir -p ./secrets
cp $(mkcert -CAROOT)/rootCA.pem ./secrets/ca.pem

mkcert -cert-file ./secrets/web.crt -key-file ./secrets/web.key "lf4a.com" "www.lf4a.com" "frontend" "localhost" "127.0.0.1"
mkcert -cert-file ./secrets/server.crt -key-file ./secrets/server.key "api.lf4a.com" "go-api" "localhost" "127.0.0.1"
mkcert -cert-file ./secrets/broker.crt -key-file ./secrets/broker.key "broker.lf4a.com" "broker" "localhost" "127.0.0.1"
mkcert -cert-file ./secrets/minio.crt -key-file ./secrets/minio.key "minio.lf4a.com" "minio" "localhost" "127.0.0.1"
mkcert -cert-file ./secrets/express.crt -key-file ./secrets/express.key "express.lf4a.com" "express" "localhost" "127.0.0.1"
mkcert -cert-file ./secrets/keycloak.crt -key-file ./secrets/keycloak.key "auth.lf4a.com" "keycloak" "localhost" "127.0.0.1"

openssl req -new -newkey rsa:2048 -nodes \
  -keyout ./secrets/client.key \
  -out ./secrets/client.csr \
  -subj "/CN=web" \
  -addext "subjectAltName=DNS:web,DNS:go-api,DNS:api.lf4a.com" \
  -addext "extendedKeyUsage=clientAuth"

mkcert -csr ./secrets/client.csr -cert-file ./secrets/client.crt

openssl req -new -newkey rsa:2048 -nodes \
  -keyout ./secrets/ocr-service.key \
  -out ./secrets/ocr-service.csr \
  -subj "/CN=ocr-service" \
  -addext "subjectAltName=DNS:ocr-service" \
  -addext "extendedKeyUsage=clientAuth"

mkcert -csr ./secrets/ocr-service.csr -cert-file ./secrets/ocr-service.crt

env "UID=$(id -u)" "GID=$(id -g)" "COMPOSE_BAKE=true" docker compose -f docker-compose.yml up --build -d
