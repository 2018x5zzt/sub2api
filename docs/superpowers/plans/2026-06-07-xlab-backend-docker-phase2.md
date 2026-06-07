# Xlab Backend Docker Phase 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a safe Docker-based deployment path for `xlab-backend` so production can expose `/xapi/v1` without changing product subscription data or payment fulfillment.

**Architecture:** Keep `xlab-backend` as a separate container on the same Docker network as `sub2api`. Build a static Linux binary, package it into a minimal image, deploy it as `xlab-backend`, and expose it only on `127.0.0.1:8090` for a host reverse proxy. Frontend switching to `/xapi/v1` remains a separate guarded step.

**Tech Stack:** Go 1.26.2, Docker, Bash deployment scripts, frontend-v2 Vite environment variables.

---

## File Structure

- Create `xlab-backend/Dockerfile`
  - Minimal runtime image for local/CI image builds.
- Create `deploy-xlab-backend.sh`
  - Builds the xlab backend Linux binary locally, uploads it to the server, builds a small image on the server, runs/replaces `xlab-backend`, and verifies `/health`.
- Create `docs/superpowers/specs/2026-06-07-xlab-backend-docker-phase2-runbook.md`
  - Operator runbook for deployment, reverse proxy, switching frontend mode, and rollback.

This plan must not modify core backend code, core migrations, product subscription logic, payment fulfillment, or API key authorization.

---

### Task 1: Add xlab-backend Dockerfile

**Files:**
- Create: `xlab-backend/Dockerfile`

- [ ] **Step 1: Add Dockerfile**

Create `xlab-backend/Dockerfile`:

```Dockerfile
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget && rm -rf /var/cache/apk/*

WORKDIR /app
COPY xlab-backend /app/xlab-backend
RUN chmod +x /app/xlab-backend

EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -q -T 5 -O /dev/null http://127.0.0.1:8090/health || exit 1

CMD ["/app/xlab-backend"]
```

- [ ] **Step 2: Verify Dockerfile references the expected binary name**

Run:

```bash
rg 'COPY xlab-backend|CMD \["/app/xlab-backend"\]' xlab-backend/Dockerfile
```

Expected: both lines are present.

- [ ] **Step 3: Commit Dockerfile**

Run:

```bash
git add xlab-backend/Dockerfile
git commit -m "$(cat <<'EOF'
feat(xlab): add backend Docker runtime image

Provide a minimal runtime image for deploying the xlab backend as an independent container.
EOF
)"
```

Expected: commit succeeds.

---

### Task 2: Add xlab-backend deployment script

**Files:**
- Create: `deploy-xlab-backend.sh`

- [ ] **Step 1: Add deployment script**

Create `deploy-xlab-backend.sh`:

```bash
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
  '${IMAGE}'
EOF_REMOTE

echo "▶ [5/5] Checking health..."
sleep 2
ssh "${SERVER}" "curl -fsS 'http://127.0.0.1:${HOST_PORT}/health' && echo && docker ps --filter name='${CONTAINER}' --format '{{.Names}}\t{{.Status}}'"

echo "✅ xlab-backend deploy complete. Configure reverse proxy /xapi/v1 -> http://127.0.0.1:${HOST_PORT}/xapi/v1 before switching frontend-v2."
```

- [ ] **Step 2: Make script executable and syntax check**

Run:

```bash
chmod +x deploy-xlab-backend.sh
bash -n deploy-xlab-backend.sh
```

Expected: PASS.

- [ ] **Step 3: Verify xlab-backend binary builds**

Run:

```bash
cd xlab-backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/xlab-backend ./cmd/server
```

Expected: PASS.

- [ ] **Step 4: Commit deployment script**

Run:

```bash
git add deploy-xlab-backend.sh xlab-backend/bin/.gitkeep
git commit -m "$(cat <<'EOF'
feat(xlab): add backend container deploy script

Build and deploy xlab-backend as an independent container on the sub2api Docker network.
EOF
)"
```

If `xlab-backend/bin/.gitkeep` does not exist, create it before commit and ensure the built binary itself is not committed.

---

### Task 3: Add Phase 2 operator runbook

**Files:**
- Create: `docs/superpowers/specs/2026-06-07-xlab-backend-docker-phase2-runbook.md`

- [ ] **Step 1: Add runbook**

Create `docs/superpowers/specs/2026-06-07-xlab-backend-docker-phase2-runbook.md`:

```md
# Xlab Backend Docker Phase 2 Runbook

## Deploy xlab-backend

```bash
cd /root/sub2api-src
./deploy-xlab-backend.sh
```

## Reverse proxy

Map `/xapi/v1/` to `http://127.0.0.1:8090/xapi/v1/`.

Before switching frontend-v2, verify:

```bash
curl -i https://<domain>/xapi/v1/subscription-products/active
```

Expected: unauthenticated request returns JSON 401, not HTML or 404.

## Switch frontend-v2 to xlab mode

```bash
VITE_XLAB_API_BASE_URL=/xapi/v1 bash ./deploy.sh
```

## Roll back frontend-v2

```bash
VITE_XLAB_API_BASE_URL=/api/v1 bash ./deploy.sh
```

## Stop xlab-backend

```bash
ssh root@152.53.39.161 "docker stop xlab-backend"
```
```

- [ ] **Step 2: Commit runbook**

Run:

```bash
git add docs/superpowers/specs/2026-06-07-xlab-backend-docker-phase2-runbook.md
git commit -m "$(cat <<'EOF'
docs(xlab): add phase two deployment runbook

Document the operational steps for deploying xlab-backend and safely switching frontend-v2 to /xapi/v1.
EOF
)"
```

Expected: commit succeeds.

---

### Task 4: Verify and optionally deploy xlab-backend

**Files:**
- No source changes unless verification finds issues.

- [ ] **Step 1: Run xlab backend tests**

Run:

```bash
cd xlab-backend
```

Expected: PASS.

- [ ] **Step 2: Run script checks**

Run:

```bash
bash -n deploy-xlab-backend.sh
git status --short --branch
```

Expected: PASS and no built binary tracked.

- [ ] **Step 3: Probe remote deployment topology**

Run:

```bash
ssh root@152.53.39.161 "docker ps --format '{{.Names}}\t{{.Status}}'"
ssh root@152.53.39.161 "docker inspect sub2api --format '{{json .NetworkSettings.Networks}}'"
ssh root@152.53.39.161 "ss -lntp | grep -E ':80|:443|:8080|:8090' || true"
```

Expected: SSH succeeds. If SSH is unavailable, stop and report the exact network failure.

- [ ] **Step 4: Deploy xlab-backend container**

Run:

```bash
./deploy-xlab-backend.sh
```

Expected: remote container `xlab-backend` is running and `http://127.0.0.1:8090/health` returns JSON success.

- [ ] **Step 5: Do not switch frontend until reverse proxy is validated**

Run from a network path that reaches production domain:

```bash
curl -i https://<domain>/xapi/v1/subscription-products/active
```

Expected: JSON 401. If it returns HTML/404, keep frontend-v2 in compatibility mode (`/api/v1`).

---

## Self-Review

- Spec coverage: Dockerfile, deployment script, reverse proxy expectations, guarded frontend switch, and rollback are covered.
- Placeholder scan: The only `<domain>` value is intentionally operator-provided because the repo does not encode the production domain.
- Type consistency: Container name `xlab-backend`, port `8090`, env vars, and `/xapi/v1` prefix match the Phase 2 design.
