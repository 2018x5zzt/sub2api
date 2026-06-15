# Core Upgrade v0.1.124 Rollback Notes

## Deployment

- Deployed at: 2026-06-15
- Branch: xlabapi @ 90f278aa6910
- New image: `sub2api-local:xlabapi-core-v0.1.124-20260615-050713`
- Production port: `127.0.0.1:8081 -> 8080`
- Health URL: http://127.0.0.1:8081/health

## Rollback baseline (pre-deploy)

- Previous image: `sub2api-local:xlabapi-core-v0.1.122-20260608`
- Previous image id: see `/tmp/sub2api-8081-rollback-image-id.txt` on server

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
  sub2api-local:xlabapi-core-v0.1.122-20260608
curl -fsS http://127.0.0.1:8081/health
```

## Notes

- Env snapshot saved at `/tmp/sub2api-8081.env` on the server.
- Migrations in v0.1.124 range run automatically on container start if configured by the app.
- `xlab-backend` on `:8090` was not changed; it still proxies core via Docker network.
