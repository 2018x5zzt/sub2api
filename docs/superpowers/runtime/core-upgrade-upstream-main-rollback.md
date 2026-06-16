# Core Upgrade upstream/main Rollback Notes

## Deployment

- Deployed at: 2026-06-16
- Branch: xlabapi @ sync with upstream/main (VERSION 0.1.136)
- New image: `sub2api-local:xlabapi-upstream-main-20260616-145115`
- Production port: `127.0.0.1:8081 -> 8080`
- Health URL: http://127.0.0.1:8081/health
- Legacy admin (8084): rebuilt to `/root/sub2api-deploy/admin-legacy/dist`

## Rollback baseline (pre-deploy)

- Previous image: `sub2api-local:xlabapi-core-v0.1.128-20260615-094243`
- Rollback snapshot: `/tmp/sub2api-8081-rollback-image.txt`

## Rollback command

```bash
docker stop sub2api
docker rm sub2api
docker run -d \
  --name sub2api \
  --restart unless-stopped \
  --network sub2api-deploy_sub2api-network \
  -p 127.0.0.1:8081:8080 \
  -v /root/sub2api-deploy/data:/app/data \
  --env-file /tmp/sub2api-8081.env \
  sub2api-local:xlabapi-core-v0.1.128-20260615-094243
curl -fsS http://127.0.0.1:8081/health
```

## Version ranges merged (v0.1.128 → upstream/main)

| Range | Risk | Key changes |
|-------|------|-------------|
| v0.1.128..v0.1.129 | Medium | API key usage daily detail, group/API key denial fixes |
| v0.1.129..v0.1.130 | High | Redeem batch update, Bedrock, OIDC, risk control, subscription expiry email toggle |
| v0.1.130..v0.1.131 | Very High | User-platform USD quota, HTTP/2 response header timeout |
| v0.1.131..v0.1.132 | High | Ops classification, WS failover, scheduler model cooldown |
| v0.1.132..v0.1.133 | High | apicompat usage/token fixes, endpoint capability gating, account auto-pause |
| v0.1.133..v0.1.136 | Very High | Scheduler outbox dedup, gateway refactor, Codex Responses bridge, quota flusher |
| v0.1.136..upstream/main | — | Final upstream/main tip (f069c9ae) |

## Xlab semantics preserved

- `productAwareSubscriptionAssigner` retained across payment/redeem/auth/subscription
- `ProductSettlementFromContext` and product billing columns unchanged
- `ops_upstream_request_body` diagnostics retained
- `frontend-v2/`, `xlab-backend/`, `enterprisebff/` protected (no unintended changes)
- Redeem product grant still uses xlab product_id semantics

## Post-merge fixes

- `frontend/src/views/admin/RedeemView.vue` — fixed merge corruption (duplicate `loadSubscriptionProducts` stub, `group.rate` → `rate_multiplier`)

## Verification

- Backend tests: PASS
- xlab-backend tests: PASS
- 8081 `/health` — OK
- 8084 `/health` + `/api/v1/settings/public` — OK
- `git log xlabapi..upstream/main` — 0 (fully synced)

## Notes

- Env snapshot: `/tmp/sub2api-8081.env`
- `xlab-backend` on `:8090` unchanged
- User-platform quota schema migrated; xlab product subscriptions use separate product settlement path
- Notification emails wired but inactive without SMTP config
