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

## Phase 3A product subscription read mirror

Rollback mode keeps xlab-backend proxying core:

```bash
cd /root/sub2api-src
XLAB_SUBSCRIPTION_READ_SOURCE=core \
XLAB_SUBSCRIPTION_SYNC_ENABLED=false \
./deploy-xlab-backend.sh
```

Sync-only rollout requires DSNs to be exported first:

```bash
export XLAB_DATABASE_URL="$XLAB_DATABASE_URL"
export CORE_DATABASE_URL="$CORE_DATABASE_URL"
test -n "$XLAB_DATABASE_URL"
test -n "$CORE_DATABASE_URL"

cd /root/sub2api-src
XLAB_SUBSCRIPTION_SYNC_ENABLED=true \
XLAB_SUBSCRIPTION_READ_SOURCE=core \
./deploy-xlab-backend.sh
```

Hybrid read rollout keeps fallback enabled:

```bash
export XLAB_DATABASE_URL="$XLAB_DATABASE_URL"
export CORE_DATABASE_URL="$CORE_DATABASE_URL"
test -n "$XLAB_DATABASE_URL"
test -n "$CORE_DATABASE_URL"

cd /root/sub2api-src
XLAB_SUBSCRIPTION_SYNC_ENABLED=true \
XLAB_SUBSCRIPTION_READ_SOURCE=hybrid \
./deploy-xlab-backend.sh
```

Runtime rollback:

```bash
cd /root/sub2api-src
XLAB_SUBSCRIPTION_READ_SOURCE=core \
XLAB_SUBSCRIPTION_SYNC_ENABLED=false \
./deploy-xlab-backend.sh
```
