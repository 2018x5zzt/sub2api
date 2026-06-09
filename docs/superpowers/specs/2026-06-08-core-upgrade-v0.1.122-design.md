# Core Upgrade v0.1.122 Design

## Context

`xlabapi` is the current production branch. The latest production baseline is:

- `f5a637be feat(xlab): add subscription read mirror phase 3a`
- `sub2api` container on `127.0.0.1:8081 -> 8080`
- `/health` returning `{"status":"ok"}`
- `xlab-backend` running as a separate healthy service

Phase 1 and Phase 3A of the xlab shell work have established a separate `/xapi/v1` boundary and moved product-subscription read traffic behind `xlab-backend` with safe fallback. This makes it reasonable to start reducing the core fork delta, but core still owns runtime authorization, payment fulfillment, redeem grants, gateway billing, usage logs, and many production-specific behaviors.

The upstream audit showed that directly merging `upstream/main` is still too risky. Later upstream ranges touch payment, subscription, quota, redeem, OAuth, schema migrations, frontend behavior, and gateway accounting. The first controlled target is therefore only the first real upstream range after the current fork baseline:

```text
v0.1.121..v0.1.122
```

## Goals

1. Start upgrading the sub2api core toward upstream using release-tag-sized steps.
2. Treat current `xlabapi` `f5a637be` and the live 8081 container as the rollback baseline.
3. Integrate only the `v0.1.121..v0.1.122` range in the first upgrade branch.
4. Preserve xlab production customizations, especially frontend-v2, enterprise BFF, product subscriptions, payment, affiliate, redeem product grants, and xlab backend wiring.
5. Produce a repeatable process for later tag ranges.

## Non-goals

- Do not merge `upstream/main` directly.
- Do not jump past `v0.1.122` in this first upgrade.
- Do not replace the live 8081 container until the upgrade branch passes verification.
- Do not migrate payment, subscription writes, redeem product grants, or gateway authorization out of current core in this task.
- Do not accept upstream deletions of xlab-specific code such as `backend/internal/enterprisebff/**`.
- Do not treat the paused selective integration worktree as the primary path; it is only a reference for salvaging scoped fixes.

## Recommended approach

Use a new isolated worktree and branch for the first upgrade step:

```text
core-upgrade-v0.1.122-xlabapi-20260608
```

The branch should start at current `xlabapi` and integrate only the upstream tag range `v0.1.121..v0.1.122`. The preferred method is a controlled merge/cherry-pick sequence that preserves upstream commit intent while rejecting unrelated changes that would roll back xlab behavior.

If the range produces conflicts that are isolated and understandable, resolve them in the upgrade branch. If conflicts spread into subscription/payment/quota/schema or xlab-specific surfaces, stop and split the range into a selective port plan instead of forcing a merge.

## Alternatives considered

### Option A: Directly merge upstream latest

This is the fastest way to reach upstream, but it is not acceptable for production risk. The audit shows very high-risk ranges after `v0.1.125`, including payment, subscription, quota, schema, frontend, and gateway refactors. A direct merge could change paid-user authorization and billing semantics.

### Option B: Continue only selective cherry-picks

This is safer in the short term, and the previous selective worktree remains useful as a source of known good patches. However, staying selective forever keeps `xlabapi` as a large fork and prevents sustainable upstream upgrades.

### Option C: Versioned tag upgrade with strict gates

This is the recommended path. It gives us upstream convergence while keeping each risk window small. It also creates a repeatable process for future ranges and aligns with the longer-term xlab shell architecture.

## Architecture boundaries

The upgrade must preserve these boundaries:

- `/api/v1/**` and `/v1/**`: current sub2api core APIs and gateway paths.
- `/xapi/v1/**`: xlab backend APIs, including product-subscription read mirror and fallback behavior.
- `frontend-v2/**`: xlab shell UI and adapters.
- `xlab-backend/**`: independent xlab service code and mirror/fallback logic.
- `backend/internal/enterprisebff/**`: xlab enterprise BFF, not removable by upstream.

Core may absorb upstream improvements for gateway, admin/account behavior, usage, API keys, and upstream-compatible fixes. Xlab business logic should not move further into core as part of this upgrade.

## Integration gates

Before merging the upgrade branch back into `xlabapi`, all gates below must be checked.

### Source safety gates

- No deletion of `backend/internal/enterprisebff/**`.
- No unintended rollback of `frontend-v2/**`.
- No unintended rollback of `xlab-backend/**`.
- No unreviewed changes to payment, product-subscription, affiliate, redeem product-grant, or quota semantics.
- No unreviewed migration/schema changes that could affect production data.

### Backend verification gates

- Relevant core Go tests for handler, service, repository, gateway, API key, account, and usage behavior.
- Product-subscription read mirror tests in `xlab-backend`.
- Any new or changed tests required by conflicts in the `v0.1.122` range.

### Frontend verification gates

- `frontend-v2` typecheck.
- `frontend-v2` build.
- Focused checks for subscription pages if API adapter files change.

### Runtime gates

- Build an upgrade image locally.
- Run a test/staging container on a non-production port first.
- Verify `/health`.
- Verify `/xapi/v1` still routes to `xlab-backend` as expected.
- Confirm rollback command/path to current 8081 production image before replacing the production container.

## Error handling and stop conditions

Stop the upgrade and report for review if any of these occur:

- Merge conflicts touch product subscription writes, payment fulfillment, affiliate settlement, redeem product grants, quota enforcement, or production migrations.
- Upstream deletes or rewrites xlab-only code.
- Tests fail in a way that cannot be isolated to the upgraded range.
- Frontend-v2 build/typecheck fails outside touched files.
- The test container is not healthy or `/health` fails.

When a stop condition occurs, preserve the worktree state for inspection unless the user asks to abort it.

## Rollback strategy

The rollback baseline is current `xlabapi` at `f5a637be` and the currently running healthy 8081 container image. The upgrade process should record:

- the pre-upgrade branch and commit,
- the current production image tag,
- the test upgrade image tag,
- the exact container replacement command,
- and the command to restore the previous 8081 container.

No production replacement should happen until the rollback information is captured.

## Success criteria

The first core upgrade step is successful when:

1. A dedicated upgrade branch contains the integrated `v0.1.121..v0.1.122` range or a documented selective equivalent.
2. The source safety gates show no xlab-specific rollback or deletion.
3. Backend, xlab-backend, and frontend-v2 verification gates pass.
4. A test container built from the upgrade branch is healthy.
5. The user approves merging the upgrade branch back into `xlabapi` and deploying it to 8081.

## Next step after approval

After this design is reviewed, write a detailed implementation plan for the `core-upgrade-v0.1.122-xlabapi-20260608` worktree. The plan should include exact git commands, conflict-review checkpoints, verification commands, image build/deploy steps, and rollback capture steps.
