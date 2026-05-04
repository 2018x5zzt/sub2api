#!/usr/bin/env bash
set -euo pipefail

# === 配置 ===
SERVER="root@152.53.39.161"
CONTAINER="sub2api"
REMOTE_TMP="/tmp/sub2api"
CONTAINER_BIN="/app/sub2api"

FRONTEND_DIR="frontend"
BACKEND_DIR="backend"

# === 1. 构建前端 ===
echo "▶ [1/4] Building frontend..."
pnpm --dir "$FRONTEND_DIR" run build

# === 2. 编译 Go 二进制（embed 前端） ===
echo "▶ [2/4] Building backend (linux/amd64 + embed)..."
cd "$BACKEND_DIR"
VERSION=$(tr -d '\r\n' < cmd/server/VERSION)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -tags embed \
  -ldflags="-s -w -X main.Version=${VERSION}" \
  -trimpath -o bin/sub2api ./cmd/server
cd ..

# === 3. 上传到服务器 ===
echo "▶ [3/4] Uploading to ${SERVER}..."
scp "${BACKEND_DIR}/bin/sub2api" "${SERVER}:${REMOTE_TMP}"

# === 4. 替换并重启容器 ===
echo "▶ [4/4] Deploying to container ${CONTAINER}..."
ssh "$SERVER" "docker cp ${REMOTE_TMP} ${CONTAINER}:${CONTAINER_BIN} && docker restart ${CONTAINER}"

echo ""
echo "✅ Deploy complete! Checking status..."
sleep 3
ssh "$SERVER" "docker ps --filter name=${CONTAINER} --format '{{.Names}}\t{{.Status}}'"
