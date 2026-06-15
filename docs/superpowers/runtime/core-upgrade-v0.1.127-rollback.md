# Core Upgrade v0.1.127 Rollback Notes

## Deployment

- Deployed at: 2026-06-15
- Branch: xlabapi @ merge commit after v0.1.127 integration
- New image: `sub2api-local:xlabapi-core-v0.1.127-20260615-082808`
- Production port: `127.0.0.1:8081 -> 8080`
- Health URL: http://127.0.0.1:8081/health
- Legacy admin (8084): rebuilt to `/root/sub2api-deploy/admin-legacy/dist`

## Rollback baseline (pre-deploy)

- Previous image: `sub2api-local:xlabapi-core-v0.1.126-20260615-054537`

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
  sub2api-local:xlabapi-core-v0.1.126-20260615-054537
curl -fsS http://127.0.0.1:8081/health
```

## Notes

- Migrations added: DingTalk provider type, usage_log image size metadata, redeem_code expires_at.
- Product subscription semantics preserved via productAwareSubscriptionAssigner and xlab product pages.
- DingTalk OAuth is code-complete but inactive until env/config is set.
