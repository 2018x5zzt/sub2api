# Core Upgrade v0.1.122 Rollback Notes

## Baseline

- Branch before upgrade: xlabapi
- Baseline commit: f5a637be feat(xlab): add subscription read mirror phase 3a
- Production container: sub2api
- Production port: 127.0.0.1:8081 -> 8080
- Production health URL: http://127.0.0.1:8081/health

## Captured production container

- Container image: sub2api-local:xlabapi-e08099a1-20260530-101540
- Image ID: sha256:08bc9d87dad1b73e54a733fcec175c7f761aa99c1c9547db01bd7a06f1d8dce7
- Container status: Up 10 hours (healthy)
- Container created: 2026-05-30T08:18:04.619543105Z
- Health response: {"status":"ok"}

## Capture commands

```bash
docker ps --filter name='^/sub2api$' --format 'name={{.Names}} image={{.Image}} ports={{.Ports}} status={{.Status}}'
docker inspect sub2api --format 'image={{.Config.Image}} id={{.Image}} created={{.Created}}'
curl -fsS http://127.0.0.1:8081/health
```

## Production replacement rule

Do not replace the 8081 production container until the upgrade branch passes source-safety, backend, xlab-backend, frontend-v2, and test-container runtime gates, and the user explicitly approves deployment.

## Rollback command template

Replace `<baseline-image>` with the image captured from `docker inspect sub2api` before deployment:

```bash
docker stop sub2api
docker rm sub2api
docker run -d --name sub2api --restart unless-stopped -p 127.0.0.1:8081:8080 <baseline-image>
curl -fsS http://127.0.0.1:8081/health
```
