#!/usr/bin/env bash
set -euo pipefail

SERVER="${SERVER:-root@152.53.39.161}"
CORE_CONTAINER="${CORE_CONTAINER:-sub2api}"
CONTAINER="${XLAB_CONTAINER:-xlab-backend}"
IMAGE="${XLAB_IMAGE:-xlab-backend:local}"
REMOTE_DIR="${XLAB_REMOTE_DIR:-/tmp/xlab-backend-deploy}"
HOST_PORT="${XLAB_HOST_PORT:-8090}"
SERVER_ADDR="${XLAB_SERVER_ADDR:-:8090}"
CORE_API_BASE_URL="${CORE_API_BASE_URL:-http://${CORE_CONTAINER}:8080/api/v1}"
TIMEOUT_SECONDS="${XLAB_CORE_TIMEOUT_SECONDS:-10}"
XLAB_DATABASE_URL="${XLAB_DATABASE_URL:-}"
CORE_DATABASE_URL="${CORE_DATABASE_URL:-}"
SUBSCRIPTION_READ_SOURCE="${XLAB_SUBSCRIPTION_READ_SOURCE:-core}"
SUBSCRIPTION_SYNC_ENABLED="${XLAB_SUBSCRIPTION_SYNC_ENABLED:-false}"
SUBSCRIPTION_SYNC_INTERVAL_SECONDS="${XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS:-300}"
SUBSCRIPTION_SYNC_STALE_SECONDS="${XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS:-600}"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${ROOT_DIR}/xlab-backend/bin"
BIN_PATH="${BIN_DIR}/xlab-backend"

echo "▶ [1/5] Building xlab-backend linux/amd64 binary..."
mkdir -p "${BIN_DIR}"
(cd "${ROOT_DIR}/xlab-backend" && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o "${BIN_PATH}" ./cmd/server)

echo "▶ [2/5] Preparing remote directory on ${SERVER}..."
ssh "${SERVER}" "rm -rf '${REMOTE_DIR}' && mkdir -p '${REMOTE_DIR}'"

echo "▶ [3/5] Uploading binary and Dockerfile..."
scp "${BIN_PATH}" "${SERVER}:${REMOTE_DIR}/xlab-backend"
scp "${ROOT_DIR}/xlab-backend/Dockerfile" "${SERVER}:${REMOTE_DIR}/Dockerfile"

echo "▶ [4/5] Building image and restarting ${CONTAINER}..."
ssh "${SERVER}" bash -s <<EOF_REMOTE
set -euo pipefail
NETWORK="\$(docker inspect '${CORE_CONTAINER}' --format '{{range \$name, \$_ := .NetworkSettings.Networks}}{{\$name}}{{end}}')"
if [ -z "\${NETWORK}" ]; then
  echo "Could not detect Docker network for ${CORE_CONTAINER}" >&2
  exit 1
fi
docker build -t '${IMAGE}' '${REMOTE_DIR}'
docker rm -f '${CONTAINER}' >/dev/null 2>&1 || true
docker run -d \
  --name '${CONTAINER}' \
  --restart unless-stopped \
  --network "\${NETWORK}" \
  -p 127.0.0.1:${HOST_PORT}:8090 \
  -e XLAB_SERVER_ADDR='${SERVER_ADDR}' \
  -e CORE_API_BASE_URL='${CORE_API_BASE_URL}' \
  -e XLAB_CORE_TIMEOUT_SECONDS='${TIMEOUT_SECONDS}' \
  -e XLAB_DATABASE_URL='${XLAB_DATABASE_URL}' \
  -e CORE_DATABASE_URL='${CORE_DATABASE_URL}' \
  -e XLAB_SUBSCRIPTION_READ_SOURCE='${SUBSCRIPTION_READ_SOURCE}' \
  -e XLAB_SUBSCRIPTION_SYNC_ENABLED='${SUBSCRIPTION_SYNC_ENABLED}' \
  -e XLAB_SUBSCRIPTION_SYNC_INTERVAL_SECONDS='${SUBSCRIPTION_SYNC_INTERVAL_SECONDS}' \
  -e XLAB_SUBSCRIPTION_SYNC_STALE_SECONDS='${SUBSCRIPTION_SYNC_STALE_SECONDS}' \
  '${IMAGE}'
EOF_REMOTE

echo "▶ [5/5] Checking health..."
sleep 2
ssh "${SERVER}" "curl -fsS 'http://127.0.0.1:${HOST_PORT}/health' && echo && docker ps --filter name='${CONTAINER}' --format '{{.Names}}\t{{.Status}}'"

echo "✅ xlab-backend deploy complete. Configure reverse proxy /xapi/v1 -> http://127.0.0.1:${HOST_PORT}/xapi/v1 before switching frontend-v2."
