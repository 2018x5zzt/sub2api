# Core Upgrade v0.1.137 Rollback Notes

## Deployment

- Deployed at: 2026-06-16
- Branch: xlabapi @ selective merge v0.1.137
- New image: (filled after deploy)
- Production port: `127.0.0.1:8081 -> 8080`
- Health URL: http://127.0.0.1:8081/health
- Legacy admin (8084): rebuilt if frontend changed

## Rollback baseline (pre-deploy)

- Previous image: `sub2api-local:xlabapi-upstream-main-20260616-145115`
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
  sub2api-local:xlabapi-upstream-main-20260616-145115
curl -fsS http://127.0.0.1:8081/health
```

## Version range merged (v0.1.136..v0.1.137)

| Area | Key changes |
|------|-------------|
| Billing | Chinese LLM fallback pricing (GLM/Kimi/MiniMax/DeepSeek), thinking-enabled default reasoning_effort |
| Gateway | Protocol-aware thinking-block filtering, DeepSeek max→xhigh, MiniMax M-series thinking rewrite |
| OpenAI | cyber_policy passthrough/audit/billing, quota rate-limit credits query/reset, image server-error failover |
| Anthropic | Preserve 429 window cooldowns |
| Auth | IP ACL denial message includes client IP, OAuth signup promo code |
| Admin/UI | Account list shows account ID, channel monitor jitter config |

## Xlab semantics preserved

- `productAwareSubscriptionAssigner` retained across payment/redeem/auth/subscription
- `ProductSettlementFromContext` and product billing columns unchanged
- `ops_upstream_request_body` diagnostics retained
- `frontend-v2/`, `xlab-backend/`, `enterprisebff/` protected (no unintended changes)
- Product subscription auth tests retained (`newAuthTestRouterWithProduct`)

## Post-merge fixes

- `api_key_auth_test.go` — kept xlab product-subscription test helper + upstream IP ACL assertion helper
- `gateway_forward_as_chat_completions.go` — combined xlab implicitReasoningEffort with upstream ApplyThinkingEnabledFallback
- `openai_gateway_service_test.go` — aligned cyber policy streaming test with v0.1.137 pass-through behavior
- `responses_to_chatcompletions_codex_events_test.go` — removed duplicate test name
- `auth_identity_payment_migrations_regression_test.go` — restored missing closing brace
- `wire_gen.go` — restored proxyExpiry + quotaFlusher cleanup wiring

## Verification

- Backend tests: PASS (`GOFLAGS=-buildvcs=false go test ./...`)
- xlab-backend tests: PASS
