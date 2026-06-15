# Core Upgrade v0.1.128 Rollback Notes

## Deployment

- Deployed at: 2026-06-15
- Branch: xlabapi @ 2497d9c4
- New image: `sub2api-local:xlabapi-core-v0.1.128-20260615-094243`
- Production port: `127.0.0.1:8081 -> 8080`
- Health URL: http://127.0.0.1:8081/health
- Legacy admin (8084): rebuilt to `/root/sub2api-deploy/admin-legacy/dist`

## Rollback baseline (pre-deploy)

- Previous image: `sub2api-local:xlabapi-core-v0.1.127-20260615-082808`
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
  sub2api-local:xlabapi-core-v0.1.127-20260615-082808
curl -fsS http://127.0.0.1:8081/health
```

## Migrations added

- `136_remove_ops_retry_replay.sql` — removes ops retry/replay tables
- `138_channel_monitor_openai_api_mode.sql` — channel monitor API mode column
- `139_seed_openai_monitor_templates.sql` — OpenAI detection monitor templates

## Upstream capabilities introduced

| Area | Capability | Notes |
|------|-----------|-------|
| Channel monitor | OpenAI API mode selection, protocol templates | Full upstream port |
| Gateway | `/v1/responses` force chat completions fallback | Full upstream port |
| Gateway | OpenAI images `n` param passthrough | Full upstream port |
| Gateway | Reasoning model temperature/top_p strip in Responses | Full upstream port |
| Gateway | Codex OAuth browser UA rewrite (Cloudflare bypass) | Code present; env-dependent |
| Gateway | gemini-3.5-flash model support | Full upstream port |
| Channels | Model pricing one-click sync | Full upstream port |
| Risk control | Keyword interception in content moderation | Full upstream port |
| Email | Notification email template service + editor | Code present; SMTP/env-dependent |
| Email | Balance/subscription/payment success notification hooks | Wired but inactive without SMTP |
| Ops | Remove retry/replay controls | Upstream simplification accepted |
| Frontend (legacy) | Email template editor, channel monitor UI, payment QR flow UI | Rebuilt to 8084 |

## Xlab semantics preserved

| Area | Decision |
|------|----------|
| `productAwareSubscriptionAssigner` | Kept in payment/redeem/auth/subscription wiring |
| `subscriptionAssigner` in `PaymentService` | Merged alongside notification email service |
| `ops_upstream_request_body` diagnostics | Restored from xlab baseline (upstream removed) |
| `ProductSettlementFromContext` | Unchanged in gateway handlers |
| `frontend-v2/`, `xlab-backend/`, `enterprisebff/` | No changes (protected paths clean) |
| Payment fulfillment product grant | xlab assigner path retained |

## Skipped / minimized

| Area | Decision |
|------|----------|
| Payment mobile QR force | UI code merged; no payment provider configured in env |
| Airwallex / multi-currency | Not introduced (already present from v0.1.126, still unconfigured) |
| Notification emails | Service wired; delivery inactive until SMTP + template config |

## Conflict resolution summary

5 conflicts resolved:

1. `openai_images.go` — kept xlab ops body capture + detachUpstreamContext
2. `payment_service.go` — merged subscriptionAssigner + notificationEmailService
3. `handler/wire.go` — kept InviteHandler + ProvideAdminSettingHandler
4. `service/wire.go` — merged ProvidePaymentService with both assigner and email
5. `wire_gen.go` — aligned with merged ProvidePaymentService signature

Additionally restored `ops_upstream_context.go` from xlab baseline after auto-merge dropped ops request body capture.

## Verification

- Backend tests: `go test ./internal/handler/... ./internal/service/... ./internal/repository/...` — PASS
- xlab-backend tests: `go test ./...` — PASS
- 8081 `/health` — `{"status":"ok"}`
- 8084 `/health` — `{"status":"ok"}`
- 8084 `/api/v1/settings/public` — proxy OK

## Notes

- Env snapshot: `/tmp/sub2api-8081.env`
- `xlab-backend` on `:8090` unchanged; still proxies core via Docker network.
- Worktree: `/root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.128-xlabapi-20260615`
- Safety branch: `safety/core-upgrade-v0.1.128-before-merge`
