# Core Upgrade v0.1.126 Rollback Notes

## Deployment

- Deployed at: 2026-06-15
- Branch: xlabapi @ 6f6455a88475
- New image: `sub2api-local:xlabapi-core-v0.1.126-20260615-054537`
- Production port: `127.0.0.1:8081 -> 8080`
- Health URL: http://127.0.0.1:8081/health
- Legacy admin (8084): rebuilt to `/root/sub2api-deploy/admin-legacy/dist`

## Rollback baseline (pre-deploy)

- Previous image: `sub2api-local:xlabapi-core-v0.1.125-20260615-051822`

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
  sub2api-local:xlabapi-core-v0.1.125-20260615-051822
curl -fsS http://127.0.0.1:8081/health
```

## Notes

- Env snapshot: `/tmp/sub2api-8081.env`
- Airwallex payment provider code is present but not configured in env; rollback does not affect xlab product subscriptions.
- `xlab-backend` on `:8090` unchanged; still proxies core via Docker network.
