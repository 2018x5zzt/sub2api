# Core Upgrade v0.1.122 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Integrate the upstream `v0.1.121..v0.1.122` core range into an isolated `xlabapi` upgrade branch while preserving the current 8081 production baseline and all xlab-specific boundaries.

**Architecture:** Start from current `xlabapi`, create a dedicated worktree, capture the live rollback image, inspect the upstream range, then integrate only `v0.1.121..v0.1.122` with strict source-safety checkpoints. Validate core backend, xlab-backend, frontend-v2, and a non-production runtime container before asking for approval to merge or deploy.

**Tech Stack:** Git worktrees, Go 1.26.2 backend, Ent migrations, React/Vite frontend-v2, Vue/Vite legacy frontend, Docker, curl health checks.

---

## File Structure

This plan is primarily a controlled upstream integration. The exact source files changed by the upstream range are discovered and reviewed during execution, but the known upstream range diff from `v0.1.121..v0.1.122` includes these areas:

- `backend/internal/service/openai_*`, `backend/internal/handler/openai_*`, `backend/internal/pkg/apicompat/**`, `backend/internal/pkg/openai_compat/**`
  - OpenAI compatible gateway behavior, raw chat completions path, stream draining, usage recording, WS passthrough metadata, and API-key upstream capability probing.
- `backend/internal/handler/admin/*`, `backend/internal/server/routes/admin.go`, `backend/internal/service/admin_service.go`, `backend/internal/service/affiliate_service.go`, `backend/internal/repository/affiliate_repo.go`
  - Upstream affiliate/admin records and balance history changes. These need careful review because xlab has its own affiliate/product subscription customizations.
- `backend/migrations/134_affiliate_ledger_audit_snapshots.sql`, `backend/migrations/auth_identity_payment_migrations_regression_test.go`
  - Upstream migration and migration regression tests. These must be reviewed before accepting any production schema impact.
- `frontend/src/**`
  - Legacy Vue frontend admin affiliate pages and account bulk-edit compact fields. These should not modify `frontend-v2/**`.
- `backend/cmd/server/VERSION`
  - Upstream version marker. Accepting this in the upgrade branch is expected; production deployment still requires explicit approval.

Files and directories that must not be deleted or unintentionally rewritten:

- `backend/internal/enterprisebff/**`
- `frontend-v2/**`
- `xlab-backend/**`
- xlab product-subscription/payment/affiliate/redeem/quota custom logic unless a step explicitly surfaces and approves the exact diff.

Plan output files created in the main worktree:

- `docs/superpowers/specs/2026-06-08-core-upgrade-v0.1.122-design.md`
- `docs/superpowers/plans/2026-06-08-core-upgrade-v0.1.122.md`

---

### Task 1: Capture the production baseline and create the upgrade worktree

**Files:**
- No source changes.
- Create runtime notes file in the upgrade worktree: `docs/superpowers/runtime/core-upgrade-v0.1.122-rollback.md`

- [ ] **Step 1: Verify the main worktree starts clean except approved planning docs**

Run from `/root/sub2api-src`:

```bash
git status --short --branch
```

Expected: branch is `xlabapi...origin/xlabapi`. The only untracked or modified files should be the approved design/plan docs:

```text
## xlabapi...origin/xlabapi
?? docs/superpowers/specs/2026-06-08-core-upgrade-v0.1.122-design.md
?? docs/superpowers/plans/2026-06-08-core-upgrade-v0.1.122.md
```

If any unrelated modified or untracked source files appear, stop and ask the user whether to keep, stash, or inspect them.

- [ ] **Step 2: Confirm the production baseline commit**

Run:

```bash
git rev-parse --short=12 HEAD && git log --oneline --decorate -n 1
```

Expected output includes current production baseline:

```text
f5a637be...
f5a637be (HEAD -> xlabapi, origin/xlabapi) feat(xlab): add subscription read mirror phase 3a
```

If `HEAD` is not `f5a637be`, update the runtime notes with the actual commit and ask the user to confirm it is still the desired baseline before continuing.

- [ ] **Step 3: Capture the live 8081 container image and status**

Run:

```bash
docker ps --filter name='^/sub2api$' --format 'name={{.Names}} image={{.Image}} ports={{.Ports}} status={{.Status}}'
docker inspect sub2api --format 'image={{.Config.Image}} id={{.Image}} created={{.Created}}'
curl -fsS http://127.0.0.1:8081/health
```

Expected output includes a healthy `sub2api` container and health JSON:

```text
name=sub2api image=sub2api-local:xlabapi-e08099a1-20260530-101540 ... status=Up ... (healthy)
image=sub2api-local:xlabapi-e08099a1-20260530-101540 id=sha256:...
{"status":"ok"}
```

If `/health` fails or the container is not healthy, stop. Do not begin the upgrade integration until the current production baseline is healthy or the user explicitly accepts proceeding without a healthy rollback baseline.

- [ ] **Step 4: Create the isolated upgrade worktree**

Run:

```bash
mkdir -p /root/.config/superpowers/worktrees/sub2api-src
git worktree add \
  /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608 \
  -b core-upgrade-v0.1.122-xlabapi-20260608 \
  xlabapi
```

Expected output includes:

```text
Preparing worktree (new branch 'core-upgrade-v0.1.122-xlabapi-20260608')
HEAD is now at f5a637be feat(xlab): add subscription read mirror phase 3a
```

If the branch or directory already exists, run:

```bash
git worktree list
git branch --list core-upgrade-v0.1.122-xlabapi-20260608
```

Then stop and ask the user whether to reuse the existing worktree or create a timestamped replacement branch.

- [ ] **Step 5: Confirm the upgrade worktree baseline**

Run:

```bash
git -C /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608 status --short --branch
git -C /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608 log --oneline --decorate -n 1
```

Expected:

```text
## core-upgrade-v0.1.122-xlabapi-20260608
f5a637be (HEAD -> core-upgrade-v0.1.122-xlabapi-20260608, origin/xlabapi, xlabapi) feat(xlab): add subscription read mirror phase 3a
```

- [ ] **Step 6: Record rollback notes inside the upgrade worktree**

Create the runtime notes directory and file:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608
mkdir -p docs/superpowers/runtime
cat > docs/superpowers/runtime/core-upgrade-v0.1.122-rollback.md <<'EOF'
# Core Upgrade v0.1.122 Rollback Notes

## Baseline

- Branch before upgrade: xlabapi
- Baseline commit: f5a637be feat(xlab): add subscription read mirror phase 3a
- Production container: sub2api
- Production port: 127.0.0.1:8081 -> 8080
- Production health URL: http://127.0.0.1:8081/health

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
EOF
```

Expected: file exists with the baseline and rollback command template.

- [ ] **Step 7: Commit rollback notes in the upgrade branch**

Run from the upgrade worktree root:

```bash
git add docs/superpowers/runtime/core-upgrade-v0.1.122-rollback.md
git commit -m "$(cat <<'EOF'
docs(core): capture v0.1.122 rollback baseline

Record the current xlabapi production baseline and rollback command template before integrating the first upstream tag range.
EOF
)"
```

Expected: commit succeeds. This commit is local to the upgrade branch and does not change the main `xlabapi` worktree.

---

### Task 2: Audit the upstream range before applying changes

**Files:**
- Create: `docs/superpowers/audits/2026-06-08-core-upgrade-v0.1.122-range.md`

- [ ] **Step 1: Verify upstream tags and range commits are available**

Run from the upgrade worktree root:

```bash
git tag --list 'v0.1.121' 'v0.1.122' --sort=version:refname
git log --oneline --decorate v0.1.121..v0.1.122
```

Expected tags:

```text
v0.1.121
v0.1.122
```

Expected commit list includes these commits at minimum:

```text
c129825f (tag: v0.1.122) Merge pull request #2116 from KnowSky404/fix/openai-bulk-edit-compact-config
ff50b8b6 Merge pull request #2170 from deqiying/fix/openai-ws-passthrough-reasoning-effort
4cbf518f fix: preserve raw chat completions usage billing
dc09b367 Merge pull request #2143 from alfadb/fix/openai-apikey-cc-default-routing
0b84d12d fix: correct affiliate audit record sources
47fb38bc fix: record zero OpenAI usage logs
72d5ee4c fix: drain OpenAI compat streams for usage
b2bdba78 stabilize image request handling
48912014 chore: sync VERSION to 0.1.121 [skip ci]
3953dc9c fix: add OpenAI compact bulk edit fields
```

If either tag is missing, run `git fetch upstream --tags` and repeat the command. If fetching fails, stop and report the network or permission error.

- [ ] **Step 2: Record changed file names and stats**

Run:

```bash
git diff --name-status v0.1.121..v0.1.122 > /tmp/core-upgrade-v0.1.122-name-status.txt
git diff --stat v0.1.121..v0.1.122 > /tmp/core-upgrade-v0.1.122-stat.txt
```

Expected: `/tmp/core-upgrade-v0.1.122-name-status.txt` contains backend OpenAI, affiliate/admin, one migration, and legacy `frontend/src/**` changes. It must not contain `frontend-v2/`, `xlab-backend/`, or `backend/internal/enterprisebff/`.

- [ ] **Step 3: Write the range audit document**

Run:

```bash
cat > docs/superpowers/audits/2026-06-08-core-upgrade-v0.1.122-range.md <<'EOF'
# Core Upgrade v0.1.122 Range Audit

## Range

- Base tag: v0.1.121
- Target tag: v0.1.122
- Production branch baseline: xlabapi at f5a637be

## Expected impact areas

- Backend OpenAI gateway and API compatibility paths.
- Backend affiliate/admin records and balance history paths.
- One affiliate ledger audit snapshot migration.
- Legacy frontend admin affiliate and account bulk-edit pages.

## Protected xlab areas

The range must not delete or unintentionally rewrite:

- backend/internal/enterprisebff/**
- frontend-v2/**
- xlab-backend/**
- xlab product-subscription/payment/redeem/quota semantics

## Name-status capture

Paste the output of:

```bash
git diff --name-status v0.1.121..v0.1.122
```

below this line during execution.

## Stat capture

Paste the output of:

```bash
git diff --stat v0.1.121..v0.1.122
```

below this line during execution.
EOF
```

Expected: audit file is created with explicit protected areas and commands for captured output.

- [ ] **Step 4: Append the captured command output to the audit**

Run:

```bash
{
  printf '\n### Captured name-status\n\n```text\n'
  sed -n '1,220p' /tmp/core-upgrade-v0.1.122-name-status.txt
  printf '```\n\n### Captured stat\n\n```text\n'
  sed -n '1,220p' /tmp/core-upgrade-v0.1.122-stat.txt
  printf '```\n'
} >> docs/superpowers/audits/2026-06-08-core-upgrade-v0.1.122-range.md
```

Expected: audit file now contains concrete diff output.

- [ ] **Step 5: Run protected-path precheck**

Run:

```bash
if rg '^(A|M|D|R[0-9]*)\s+(frontend-v2/|xlab-backend/|backend/internal/enterprisebff/)' /tmp/core-upgrade-v0.1.122-name-status.txt; then
  echo 'Protected path touched by upstream range; stop for review.'
  exit 1
fi
```

Expected: command exits `0` with no output. If it prints a protected path, stop and do not apply the range.

- [ ] **Step 6: Commit the range audit**

Run:

```bash
git add docs/superpowers/audits/2026-06-08-core-upgrade-v0.1.122-range.md
git commit -m "$(cat <<'EOF'
docs(core): audit upstream v0.1.122 range

Capture the changed files, protected xlab areas, and review scope before applying the first upstream core upgrade range.
EOF
)"
```

Expected: commit succeeds.

---

### Task 3: Apply the upstream v0.1.122 range into the upgrade branch

**Files:**
- Modify: upstream range files accepted from `v0.1.121..v0.1.122`
- Preserve: `backend/internal/enterprisebff/**`, `frontend-v2/**`, `xlab-backend/**`

- [ ] **Step 1: Create a safety branch label before applying upstream changes**

Run from the upgrade worktree root:

```bash
git branch safety/core-upgrade-v0.1.122-before-merge
git status --short --branch
```

Expected: status is clean on `core-upgrade-v0.1.122-xlabapi-20260608`.

- [ ] **Step 2: Attempt a no-commit merge of the target tag**

Run:

```bash
git merge --no-commit --no-ff v0.1.122
```

Expected: either a clean staged merge or merge conflicts. Do not commit yet.

If Git reports conflicts, continue to the next step. If Git reports an automatic merge without conflicts, still run all source-safety checks before committing.

- [ ] **Step 3: List conflicts and changed files**

Run:

```bash
git status --short
git diff --name-only --diff-filter=U
git diff --name-status HEAD
```

Expected: any conflicts are reviewable and do not include protected xlab paths. If conflicts include `backend/internal/enterprisebff/**`, `frontend-v2/**`, or `xlab-backend/**`, run:

```bash
git merge --abort
```

Then stop and report that the protected boundary was touched.

- [ ] **Step 4: Run protected-path post-merge check**

Run:

```bash
git diff --name-status HEAD > /tmp/core-upgrade-v0.1.122-merged-name-status.txt
if rg '^(A|M|D|R[0-9]*)\s+(frontend-v2/|xlab-backend/|backend/internal/enterprisebff/)' /tmp/core-upgrade-v0.1.122-merged-name-status.txt; then
  echo 'Protected xlab path changed after merge; aborting for review.'
  git merge --abort
  exit 1
fi
```

Expected: command exits `0` with no protected path output.

- [ ] **Step 5: Review sensitive-path changes before resolving or accepting them**

Run:

```bash
git diff --name-status HEAD | rg 'backend/(migrations|ent/schema)|payment|subscription|affiliate|redeem|quota|frontend/src' || true
git diff -- backend/migrations backend/internal/service/payment_fulfillment.go backend/internal/repository/affiliate_repo.go backend/internal/service/affiliate_service.go backend/internal/handler/admin/affiliate_handler.go frontend/src || true
```

Expected: review output shows the upstream affiliate/admin/legacy-frontend/migration changes. If the diff changes xlab product-subscription writes, payment fulfillment semantics, redeem product grants, or quota enforcement, run:

```bash
git merge --abort
```

Then stop and propose a selective-port plan instead of a full tag merge.

- [ ] **Step 6: Resolve merge conflicts if any exist**

For each conflicted file from `git diff --name-only --diff-filter=U`, resolve with these rules:

```text
1. Keep xlab-specific product subscription, payment, affiliate settlement, redeem product grant, enterprise BFF, and frontend-v2 behavior.
2. Accept upstream OpenAI gateway fixes only when they do not remove xlab-specific image diagnostics, compact keepalive behavior, usage logging customizations, or route compatibility.
3. Accept upstream legacy frontend changes only under frontend/src/**, not frontend-v2/**.
4. Accept upstream migrations only after confirming they do not rewrite existing xlab migration semantics.
5. Delete all conflict markers before staging.
```

After editing conflicts, run:

```bash
git diff --check
git diff --name-only --diff-filter=U
```

Expected:

```text
```

No output from unresolved-conflict command. `git diff --check` exits `0`.

- [ ] **Step 7: Run source-safety checks before committing the merge**

Run:

```bash
git diff --name-status HEAD > /tmp/core-upgrade-v0.1.122-final-name-status.txt
git diff --check
if rg '^(D|R[0-9]*)\s+backend/internal/enterprisebff/' /tmp/core-upgrade-v0.1.122-final-name-status.txt; then exit 1; fi
if rg '^(A|M|D|R[0-9]*)\s+frontend-v2/' /tmp/core-upgrade-v0.1.122-final-name-status.txt; then exit 1; fi
if rg '^(A|M|D|R[0-9]*)\s+xlab-backend/' /tmp/core-upgrade-v0.1.122-final-name-status.txt; then exit 1; fi
```

Expected: command exits `0`. If it exits non-zero, do not commit; inspect the offending paths and stop for user review.

- [ ] **Step 8: Commit the upstream range merge**

Run:

```bash
git add -A
git commit -m "$(cat <<'EOF'
merge(core): integrate upstream v0.1.122 range

Merge the first upstream tag range after the xlabapi fork baseline while preserving xlab-specific frontend-v2, xlab-backend, and enterprise boundaries.
EOF
)"
```

Expected: commit succeeds. If the merge had no changes or commit fails because there is nothing to commit, stop and inspect `git status --short --branch` before continuing.

---

### Task 4: Verify backend and xlab-backend behavior

**Files:**
- No planned source changes unless tests expose a necessary conflict-resolution fix.

- [ ] **Step 1: Run core backend targeted tests**

Run from the upgrade worktree root:

```bash
cd backend
go test ./internal/pkg/openai_compat ./internal/pkg/apicompat ./internal/handler ./internal/service ./internal/repository -run 'OpenAI|ChatCompletions|Raw|Usage|Gateway|Affiliate|Balance|Account|Payment|Migration|APIKey'
```

Expected: command exits `0`. Package output may include lines like:

```text
ok  github.com/Wei-Shaw/sub2api/backend/internal/service  ...
```

If failures point to conflict-resolution mistakes in touched files, fix the smallest possible issue, rerun this command, and commit with:

```bash
git add <fixed-files>
git commit -m "$(cat <<'EOF'
fix(core): resolve v0.1.122 integration test failures

Adjust the upstream v0.1.122 merge result only where tests showed a conflict with existing xlabapi behavior.
EOF
)"
```

If failures are pre-existing or unrelated, capture the failing output and ask the user before continuing.

- [ ] **Step 2: Run core backend broader package tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608/backend
go test ./internal/handler/... ./internal/service/... ./internal/repository/...
```

Expected: command exits `0`. If it fails, do not proceed to frontend or runtime checks until the failure is classified and resolved or explicitly accepted by the user.

- [ ] **Step 3: Run xlab-backend tests**

Run:

```bash
cd /root/.config/superpowers/worktrees/sub2api-src/core-upgrade-v0.1.122-xlabapi-20260608/xlab-backend
go test ./...
```

Expected: command exits `0`, proving the xlab read-mirror service still builds and tests against the upgraded repository state.

- [ ] **Step 4: Check backend and xlab-backend lints in the IDE**

Use the IDE diagnostics tool for these paths:

```text
backend/internal/pkg/openai_compat
backend/internal/pkg/apicompat
backend/internal/handler
backend/internal/service
backend/internal/repository
xlab-backend
```

Expected: no new diagnostics caused by the merge. If diagnostics are clearly from the merge and easy to fix, fix them, rerun the relevant Go tests, and commit the fix. If diagnostics are pre-existing or ambiguous, record them in the final handoff.

---

### Task 5: Verify frontend-v2 and legacy frontend impact

**Files:**
- No planned source changes unless tests expose a necessary conflict-resolution fix.

- [ ] **Step 1: Confirm frontend-v2 was not changed by the merge**

Run from the upgrade worktree root:

```bash
git diff safety/core-upgrade-v0.1.122-before-merge..HEAD --name-status | rg '^.*\sfrontend-v2/' || true
```

Expected: no output. If any `frontend-v2/**` file appears, inspect it. Unless the change is an explicitly approved adapter fix, revert only that unintended frontend-v2 change and commit the correction.

- [ ] **Step 2: Run frontend-v2 typecheck**

Run:

```bash
npm --prefix frontend-v2 run typecheck
```

Expected:

```text
> sub2api-frontend-v2@2.0.0 typecheck
> tsc -b
```

Command exits `0`. If dependencies are missing, run `npm --prefix frontend-v2 ci` and repeat the typecheck.

- [ ] **Step 3: Run frontend-v2 build**

Run:

```bash
npm --prefix frontend-v2 run build
```

Expected:

```text
> sub2api-frontend-v2@2.0.0 build
> tsc -b && vite build
```

Command exits `0` and Vite writes a production build.

- [ ] **Step 4: Run legacy frontend targeted tests for upstream-touched UI files**

Run:

```bash
npm --prefix frontend exec -- vitest --root frontend run \
  src/components/account/__tests__/BulkEditAccountModal.spec.ts
```

Expected: Vitest exits `0`. If dependencies are missing, run `npm --prefix frontend ci` and repeat this command.

- [ ] **Step 5: Run legacy frontend typecheck**

Run:

```bash
npm --prefix frontend run typecheck
```

Expected:

```text
> sub2api-frontend@1.0.0 typecheck
> vue-tsc --noEmit
```

Command exits `0`. If typecheck fails only in upstream-touched legacy frontend files, fix the smallest possible issue, rerun the command, and commit. If it fails elsewhere, stop and ask the user whether to proceed with a known pre-existing legacy frontend issue.

---

### Task 6: Build and run a non-production upgrade container

**Files:**
- No source changes.

- [ ] **Step 1: Build a local upgrade image**

Run from the upgrade worktree root:

```bash
docker build \
  -f Dockerfile \
  --build-arg VERSION=0.1.122-xlabapi-20260608 \
  --build-arg COMMIT=$(git rev-parse --short=12 HEAD) \
  -t sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608 \
  .
```

Expected: Docker build exits `0` and produces image `sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608`.

- [ ] **Step 2: Capture existing test container state if present**

Run:

```bash
docker ps -a --filter name='^/sub2api-core-upgrade-v0122-test$' --format 'name={{.Names}} image={{.Image}} status={{.Status}}'
```

Expected: either no output or a previous test container. If a previous test container exists, remove only that test container:

```bash
docker rm -f sub2api-core-upgrade-v0122-test
```

- [ ] **Step 3: Run the upgrade image on a non-production port**

Run:

```bash
docker run -d \
  --name sub2api-core-upgrade-v0122-test \
  --restart no \
  -p 127.0.0.1:18081:8080 \
  sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608
```

Expected: command prints a container ID. This uses a non-production port and does not replace the live 8081 container.

- [ ] **Step 4: Verify test container health**

Run:

```bash
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS http://127.0.0.1:18081/health; then
    exit 0
  fi
  sleep 3
done
docker logs sub2api-core-upgrade-v0122-test --tail 120
exit 1
```

Expected:

```text
{"status":"ok"}
```

If health fails, keep the test container logs and stop for review.

- [ ] **Step 5: Confirm the production 8081 container is still untouched and healthy**

Run:

```bash
docker ps --filter name='^/sub2api$' --format 'name={{.Names}} image={{.Image}} ports={{.Ports}} status={{.Status}}'
curl -fsS http://127.0.0.1:8081/health
```

Expected: production `sub2api` is still the baseline image and health returns `{"status":"ok"}`.

- [ ] **Step 6: Stop the non-production test container after verification**

Run:

```bash
docker rm -f sub2api-core-upgrade-v0122-test
```

Expected: container is removed. The image remains available for later staging or deployment approval.

---

### Task 7: Summarize verification and prepare merge/deploy decision

**Files:**
- Create: `docs/superpowers/runtime/core-upgrade-v0.1.122-verification.md`

- [ ] **Step 1: Write verification summary**

Run from the upgrade worktree root:

```bash
cat > docs/superpowers/runtime/core-upgrade-v0.1.122-verification.md <<'EOF'
# Core Upgrade v0.1.122 Verification Summary

## Branch

- Upgrade branch: core-upgrade-v0.1.122-xlabapi-20260608
- Baseline branch: xlabapi
- Baseline commit: f5a637be feat(xlab): add subscription read mirror phase 3a
- Upstream range: v0.1.121..v0.1.122

## Source safety gates

- backend/internal/enterprisebff/** deletion check: not changed
- frontend-v2/** unintended change check: not changed
- xlab-backend/** unintended change check: not changed
- payment/subscription/redeem/quota semantic review: completed during merge review
- migration/schema review: completed during merge review

## Verification commands

Record exact pass/fail output for each command below during execution:

```bash
go test ./internal/pkg/openai_compat ./internal/pkg/apicompat ./internal/handler ./internal/service ./internal/repository -run 'OpenAI|ChatCompletions|Raw|Usage|Gateway|Affiliate|Balance|Account|Payment|Migration|APIKey'
go test ./internal/handler/... ./internal/service/... ./internal/repository/...
go test ./...
npm --prefix frontend-v2 run typecheck
npm --prefix frontend-v2 run build
npm --prefix frontend exec -- vitest --root frontend run src/components/account/__tests__/BulkEditAccountModal.spec.ts
npm --prefix frontend run typecheck
docker build -f Dockerfile --build-arg VERSION=0.1.122-xlabapi-20260608 --build-arg COMMIT=<commit> -t sub2api-local:core-upgrade-v0.1.122-xlabapi-20260608 .
curl -fsS http://127.0.0.1:18081/health
curl -fsS http://127.0.0.1:8081/health
```

## Deployment decision

Do not merge to xlabapi or deploy to 8081 until the user reviews this summary and explicitly approves the merge/deploy step.
EOF
```

Expected: summary file exists and contains all gates.

- [ ] **Step 2: Append final git diff summary**

Run:

```bash
{
  printf '\n## Final diff summary against xlabapi\n\n```text\n'
  git diff --stat xlabapi..HEAD
  printf '```\n\n## Final changed files against xlabapi\n\n```text\n'
  git diff --name-status xlabapi..HEAD
  printf '```\n'
} >> docs/superpowers/runtime/core-upgrade-v0.1.122-verification.md
```

Expected: verification summary includes a concrete final diff summary and changed file list.

- [ ] **Step 3: Commit verification summary**

Run:

```bash
git add docs/superpowers/runtime/core-upgrade-v0.1.122-verification.md
git commit -m "$(cat <<'EOF'
docs(core): summarize v0.1.122 upgrade verification

Record source safety, test, frontend, runtime, and deployment-decision gates for the first upstream core upgrade range.
EOF
)"
```

Expected: commit succeeds.

- [ ] **Step 4: Show final branch state**

Run:

```bash
git status --short --branch
git log --oneline --decorate -n 8
```

Expected: upgrade branch is clean and contains local commits for rollback notes, range audit, upstream integration, any conflict fixes, and verification summary.

- [ ] **Step 5: Ask user for merge/deploy approval**

Report the results with:

```text
The core-upgrade-v0.1.122-xlabapi-20260608 branch is ready for review. It has not been merged into xlabapi and has not replaced the live 8081 container. Please choose:

1. Review only: leave the branch as-is.
2. Merge to xlabapi: merge locally, no deployment.
3. Merge and deploy: merge to xlabapi, build/deploy to 8081 with rollback command ready.
```

Expected: wait for explicit user approval before merging or deploying.

---

### Task 8: Optional merge back to xlabapi after explicit user approval

**Files:**
- Modify: whatever files are present in the approved upgrade branch.

Only run this task if the user explicitly chooses merge or merge-and-deploy.

- [ ] **Step 1: Switch to the main worktree and verify status**

Run:

```bash
cd /root/sub2api-src
git status --short --branch
```

Expected: no unrelated source changes. Approved planning docs may still be untracked if the user has not asked to commit them separately. If untracked planning docs exist, add them as part of the merge documentation commit only after user approval.

- [ ] **Step 2: Merge the upgrade branch**

Run:

```bash
git merge --no-ff core-upgrade-v0.1.122-xlabapi-20260608 -m "$(cat <<'EOF'
merge(core): upgrade upstream core to v0.1.122

Integrate the first upstream tag range into xlabapi with source-safety, backend, frontend, xlab-backend, and runtime verification gates.
EOF
)"
```

Expected: merge succeeds. If conflicts appear in the main worktree, stop and ask for review; do not force resolution.

- [ ] **Step 3: Run final post-merge smoke checks**

Run:

```bash
git status --short --branch
cd backend && go test ./internal/pkg/openai_compat ./internal/pkg/apicompat ./internal/handler ./internal/service ./internal/repository -run 'OpenAI|ChatCompletions|Raw|Usage|Gateway|Affiliate|Balance|Account|Payment|Migration|APIKey'
cd ../xlab-backend && go test ./...
cd .. && npm --prefix frontend-v2 run typecheck
```

Expected: all commands exit `0`.

---

### Task 9: Optional deploy to 8081 after explicit user approval

**Files:**
- No source changes.

Only run this task if the user explicitly chooses merge-and-deploy.

- [ ] **Step 1: Capture production image immediately before deployment**

Run from `/root/sub2api-src`:

```bash
docker inspect sub2api --format '{{.Config.Image}}' > /tmp/sub2api-8081-rollback-image.txt
docker inspect sub2api --format '{{.Image}}' > /tmp/sub2api-8081-rollback-image-id.txt
curl -fsS http://127.0.0.1:8081/health
```

Expected: health returns `{"status":"ok"}` and rollback image files are populated.

- [ ] **Step 2: Build production candidate image**

Run:

```bash
docker build \
  -f Dockerfile \
  --build-arg VERSION=0.1.122-xlabapi-20260608 \
  --build-arg COMMIT=$(git rev-parse --short=12 HEAD) \
  -t sub2api-local:xlabapi-core-v0.1.122-20260608 \
  .
```

Expected: build exits `0`.

- [ ] **Step 3: Replace the 8081 container only after rollback image is captured**

Run:

```bash
docker stop sub2api
docker rm sub2api
docker run -d \
  --name sub2api \
  --restart unless-stopped \
  -p 127.0.0.1:8081:8080 \
  sub2api-local:xlabapi-core-v0.1.122-20260608
```

Expected: new container starts and binds `127.0.0.1:8081`.

- [ ] **Step 4: Verify production health after deployment**

Run:

```bash
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS http://127.0.0.1:8081/health; then
    exit 0
  fi
  sleep 3
done
docker logs sub2api --tail 120
exit 1
```

Expected:

```text
{"status":"ok"}
```

- [ ] **Step 5: Roll back if health fails**

Run only if Step 4 fails:

```bash
ROLLBACK_IMAGE=$(cat /tmp/sub2api-8081-rollback-image.txt)
docker rm -f sub2api || true
docker run -d \
  --name sub2api \
  --restart unless-stopped \
  -p 127.0.0.1:8081:8080 \
  "$ROLLBACK_IMAGE"
curl -fsS http://127.0.0.1:8081/health
```

Expected: rollback container returns `{"status":"ok"}`.

- [ ] **Step 6: Report deployment status**

Run:

```bash
docker ps --filter name='^/sub2api$' --format 'name={{.Names}} image={{.Image}} ports={{.Ports}} status={{.Status}}'
git log --oneline --decorate -n 3
```

Expected: report the active image, health result, and current commit to the user.
