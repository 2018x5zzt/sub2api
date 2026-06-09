# Core Upgrade v0.1.122 Verification Summary

## Branch

- Upgrade branch: `core-upgrade-v0.1.122-xlabapi-20260608`
- Baseline branch: `xlabapi`
- Baseline commit: `f5a637be feat(xlab): add subscription read mirror phase 3a`
- Merge commit before this summary: `ed3bdadaad01 merge(core): integrate upstream v0.1.122 range`
- Upstream range: `v0.1.121..v0.1.122`

## Source safety gates

- `backend/internal/enterprisebff/**` deletion check: not changed in final branch diff.
- `frontend-v2/**` unintended change check: not changed in final branch diff.
- `xlab-backend/**` unintended change check: not changed in final branch diff.
- Payment/subscription/redeem/quota semantic review: no final branch diff in these protected xlab semantic areas.
- Migration/schema review: upstream range audit was captured; final branch diff does not include unreviewed production schema changes beyond the documented audit path.

## Verification commands and evidence

### Backend

Targeted core backend test command, run from the upgrade worktree backend module:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608/backend
go test ./internal/pkg/openai_compat ./internal/pkg/apicompat ./internal/handler ./internal/service ./internal/repository -run 'OpenAI|ChatCompletions|Raw|Usage|Gateway|Affiliate|Balance|Account|Payment|Migration|APIKey'
```

Evidence: command exited `0`; output included:

```text
ok   github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat  0.005s [no tests to run]
ok   github.com/Wei-Shaw/sub2api/internal/pkg/apicompat       0.016s
```

Broader backend package test command:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608/backend
go test ./internal/handler/... ./internal/service/... ./internal/repository/...
```

Evidence: command exited `0`; output included:

```text
ok   github.com/Wei-Shaw/sub2api/internal/handler        20.955s
ok   github.com/Wei-Shaw/sub2api/internal/handler/admin   0.173s
ok   github.com/Wei-Shaw/sub2api/internal/handler/dto     0.022s
```

### Xlab backend

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608/xlab-backend
go test ./...
```

Evidence: command exited `0`; output included:

```text
ok   github.com/2018x5zzt/xlab-backend/cmd/server             0.004s
ok   github.com/2018x5zzt/xlab-backend/internal/config        0.004s
ok   github.com/2018x5zzt/xlab-backend/internal/core          0.011s
ok   github.com/2018x5zzt/xlab-backend/internal/httpapi       0.007s
ok   github.com/2018x5zzt/xlab-backend/internal/storage       0.006s
ok   github.com/2018x5zzt/xlab-backend/internal/subscriptions 0.008s
```

IDE diagnostics were checked for the touched backend/xlab paths and returned `totalDiagnostics: 0`.

### Frontend-v2

Unintended change check:

```bash
git diff safety/core-upgrade-v0.1.122-before-merge..HEAD --name-status | rg '^.*\sfrontend-v2/' || true
```

Evidence: no output.

Typecheck:

```bash
npm --prefix frontend-v2 ci
npm --prefix frontend-v2 run typecheck
```

Evidence: dependency install completed; typecheck command exited `0`.

Build:

```bash
npm --prefix frontend-v2 run build
```

Evidence: command exited `0`; output included:

```text
> sub2api-frontend-v2@2.0.0 build
> tsc -b && vite build

vite v5.4.21 building for production...
✓ 1819 modules transformed.
```

### Legacy frontend

`frontend/` has `pnpm-lock.yaml` but no `package-lock.json`, so `npm ci` failed as expected for lack of npm lockfile. Dependencies were installed temporarily without writing a lockfile:

```bash
npm --prefix frontend install --package-lock=false --no-save
npm --prefix frontend exec -- vitest --root frontend run src/components/account/__tests__/BulkEditAccountModal.spec.ts
npm --prefix frontend run typecheck
rm -rf frontend/node_modules
```

Evidence: targeted Vitest and typecheck exited `0`; Vitest output included:

```text
✓ src/components/account/__tests__/BulkEditAccountModal.spec.ts (12 tests) 270ms

Test Files  1 passed (1)
Tests       12 passed (12)
```

After cleanup, `git status --short --branch` for the upgrade worktree was clean.

### Runtime container

Upgrade image build/inspect:

```bash
docker build --progress=plain -f Dockerfile \
  --build-arg VERSION=0.1.122-xlabapi-20260608 \
  --build-arg COMMIT=$(git rev-parse --short=12 HEAD) \
  -t sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608 .
docker image inspect sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608 --format 'id={{.Id}} created={{.Created}} tags={{.RepoTags}}'
```

Evidence: image exists:

```text
id=sha256:b38b7e1983ee3d332e38911c4c36131e61116ea2e055abac1757b25d0d6f66f7
tags=[sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608]
```

Port and setup notes:

- Planned non-production port `18081` was occupied by `gpt-image-playground-local`.
- Alternate non-production port `18082` was occupied by `epusdt-epusdt-1`.
- A plain test container on `18083` entered first-run setup wizard and returned `404` for `/health`; logs showed setup wizard mode.
- Final runtime verification used the existing non-production `test_xlab_sub2api-network` and `sub2api-test-xlab` volumes/env, mapped to `127.0.0.1:18083`, without touching production.

Runtime verification command:

```bash
docker rm -f sub2api-core-upgrade-v0122-test 2>/dev/null || true
docker run -d \
  --name sub2api-core-upgrade-v0122-test \
  --restart no \
  --network test_xlab_sub2api-network \
  --volumes-from sub2api-test-xlab \
  -p 127.0.0.1:18083:8080 \
  <filtered env copied from sub2api-test-xlab> \
  sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608
curl -fsS http://127.0.0.1:18083/health
```

Evidence: command exited `0`; output included:

```text
{"status":"ok"}
```

The non-production test container was then removed:

```bash
docker rm -f sub2api-core-upgrade-v0122-test
docker ps -a --filter name='^/sub2api-core-upgrade-v0122-test$' --format 'name={{.Names}} image={{.Image}} status={{.Status}}'
```

Evidence: no container remained after removal.

### Production container untouched

Production health check after test cleanup:

```bash
docker ps --filter name='^/sub2api$' --format 'name={{.Names}} image={{.Image}} ports={{.Ports}} status={{.Status}}'
curl -fsS http://127.0.0.1:8081/health
```

Evidence:

```text
name=sub2api image=sub2api-local:xlabapi-e08099a1-20260530-101540 ports=127.0.0.1:8081->8080/tcp status=Up 36 hours (healthy)
{"status":"ok"}
```

## Final diff summary against xlabapi

```text
 .../service/openai_gateway_chat_completions.go     |   3 +
 .../2026-06-08-core-upgrade-v0.1.122-range.md      | 170 +++++++++++++++++++++
 .../runtime/core-upgrade-v0.1.122-rollback.md      |  40 +++++
 3 files changed, 213 insertions(+)
```

## Final changed files against xlabapi

```text
M	backend/internal/service/openai_gateway_chat_completions.go
A	docs/superpowers/audits/2026-06-08-core-upgrade-v0.1.122-range.md
A	docs/superpowers/runtime/core-upgrade-v0.1.122-rollback.md
```

## Deployment decision

This branch has not been merged into `xlabapi` and has not replaced the live `8081` container.

Do not merge to `xlabapi` or deploy to `8081` until the user reviews this summary and explicitly approves the merge/deploy step.
