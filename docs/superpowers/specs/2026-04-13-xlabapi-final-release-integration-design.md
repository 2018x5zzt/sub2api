# Xlabapi Final Release Integration Design

**Date:** 2026-04-13

**Goal:** Produce one deployable `xlabapi` release on the current host that preserves the existing invite and registration feature set, absorbs the approved feature branches in order, and ships as a verified containerized deployment with a rollback path.

## Scope

This work consolidates the current `xlabapi` branch into a final release candidate and deploys it on the current machine.

Included:

- keep the current `xlabapi` branch as the only integration target
- preserve the invite and registration feature set already present on `xlabapi`
- treat invite and registration as compatibility surfaces to verify, not new feature work
- preserve the already-verified ops fixes currently on `xlabapi`
- absorb the approved feature branches into `xlabapi` in the user-specified order
- run fresh code-level verification after each merge gate and again after the final integrated state
- build one final local deployment image from `/root/sub2api-src`
- update `/root/sub2api-deploy` to point to the final integrated image
- generate release and rollback compose snapshots before the live switch
- perform the real container cutover on the current host
- run fresh post-deploy health checks and log checks

Excluded:

- no redesign of invite, registration, or reward product behavior
- no branch reshuffling or reprioritization
- no unrelated refactoring during conflict resolution
- no archive, backup, release, or `main` branch merges
- no silent trust in prior test results; all completion claims require fresh verification

## Current State

The repository is already on branch `xlabapi`, and that branch is not empty integration scaffolding.

It already contains:

- the invite growth and invite consolidation work
- invite admin rollout verification SQL and related tests
- user and admin invite pages and APIs
- the recent ops fixes and regression coverage

That means the deployment objective is not to implement invite and registration from scratch. The objective is to:

1. preserve those capabilities through branch integration
2. prove they still behave as expected after integration
3. publish the final integrated container on the current host

## Approved Integration Order

The merge order is fixed:

1. `api-key-quota-help-url-plus-oauth-itemrefs-20260412-084117`
2. `dynamic-group-budget-multiplier`
3. `enterprise-visible-groups`
4. `openai-oauth-store-false-itemrefs`
5. `smtp-outlook-starttls-20260412-151123`
6. `upgrade-v0.1.106-merge`

This order is part of the design. It is not adjusted during execution unless a blocker forces explicit user re-approval.

## Design Principles

1. Keep `xlabapi` as the single source of release truth.
2. Resolve merge conflicts with the minimum change set required to preserve approved behavior.
3. Verify after every integration step instead of deferring all risk to the end.
4. Preserve invite and registration behavior unless verification proves a regression.
5. Build and deploy one final image tag for the fully integrated branch instead of shipping a leftover feature image.
6. Prepare rollback artifacts before any live cutover.

## Compatibility Surfaces

The following surfaces are treated as release gates.

### Invite and Registration

Invite and registration are already implemented on `xlabapi`, so they are release compatibility surfaces rather than new work.

They must remain intact at three levels:

- backend: invite migrations, repositories, services, handlers, DTOs, and verification SQL remain present and wired
- frontend: user invite center, admin invite operations, and registration invite-link language remain present and buildable
- behavior contract: existing invite and registration tests continue to pass after integration

If integration breaks one of these surfaces, fix only the regression required to restore the current approved behavior.

### Ops Regressions

The current `ops` fixes already landed on `xlabapi` and were freshly verified before this design was written.

Those fixes remain protected release surfaces:

- upstream/requested model display fallback behavior
- runtime log controls overflow fix
- targeted frontend regression tests for those views

### Branch-Specific Feature Surfaces

Each merged branch introduces its own verification target. The execution plan will define a minimum proof command for each branch so that risk is contained at the branch that introduced it.

## Merge Gate Design

Each branch merge follows the same gate sequence:

1. merge the branch into `xlabapi`
2. if conflicts occur, resolve them with the smallest behavior-preserving edit set
3. inspect the resulting diff to ensure the branch intent survived the merge
4. run branch-specific verification for the merged feature
5. run cross-cutting protection checks for invite, registration, and current ops regressions when relevant
6. only proceed to the next branch if the current gate is green

This design intentionally avoids the "merge everything, then debug everything" pattern.

## Code-Level Verification Design

Verification is split into targeted gates and final integrated gates.

### Per-merge verification

After each branch merge, run the smallest command set that proves:

- the merged branch still provides its intended feature or fix
- the merge did not immediately break `xlabapi`

The exact command list will be captured in the implementation plan, but it must remain concrete and fresh for each merge.

### Final integrated verification

After the sixth branch is merged, run a fresh final verification pass for the final branch state.

That final pass must cover:

- invite and registration compatibility checks
- targeted ops regression tests already on `xlabapi`
- any branch-specific tests needed to prove the integrated state is coherent
- clean working tree before build and before deployment

## Container Design

The deployment image must be built from the final integrated `/root/sub2api-src` worktree, not from a historical feature override.

Requirements:

- use a new final image tag that clearly represents the integrated `xlabapi` release
- update `/root/sub2api-deploy/docker-compose.override.yml` to reference that final image build
- preserve the existing deployment topology in `/root/sub2api-deploy`
- validate the compose result before the live switch

Container verification must prove:

- image builds successfully
- compose renders successfully
- application container starts successfully
- dependent PostgreSQL and Redis services remain healthy
- application health endpoint succeeds
- startup logs do not show obvious migration, wiring, or missing-env failures

## Release and Rollback Design

The deployment flow on the current host must be reversible.

Before the live switch:

- snapshot the current deployment override/config into a new release record
- create a paired rollback snapshot representing the exact pre-cutover state

After the live switch:

- if health checks fail, use the rollback snapshot instead of manual ad hoc recovery

This keeps rollback deterministic and auditable.

## Deployment Flow

The deployment flow has four stages:

1. integration
   merge branches into `xlabapi` in the approved order with verification gates

2. final code verification
   run the final fresh verification commands against the final integrated branch state

3. image and compose preparation
   build the final image, update deployment override/config, and generate release and rollback snapshots

4. live switch and post-deploy checks
   run the real compose update on the current host, then verify health and logs

## Error Handling

### Merge conflicts

If a branch conflicts with the current `xlabapi` state:

- prefer minimal conflict resolution
- preserve already-approved `xlabapi` behavior unless the incoming branch intentionally supersedes it
- do not sneak in opportunistic cleanup or refactors

### Verification failures

If a verification command fails during integration:

- stop at the failing branch
- restore green status by fixing only the regression required for the intended integrated behavior
- do not continue to the next branch while red

### Deployment failures

If the final image builds but the live container does not become healthy:

- inspect logs and compose state
- if the issue is not immediately fixable without expanding risk, execute the prepared rollback

## Success Criteria

This design is complete only when all of the following are true:

- `xlabapi` contains the six approved branches in the approved order
- invite and registration compatibility checks still pass
- current ops regression checks still pass
- final integrated verification is fresh and green
- one final deployment image is built from the integrated branch
- `/root/sub2api-deploy` is updated to use the final release image/config
- release and rollback snapshots exist for this cutover
- the current host is running the new container successfully
- post-deploy health checks and log checks are green

## Risks and Mitigations

### Risk: invite or registration regression hidden by unrelated merges

Mitigation:

- treat invite and registration as explicit cross-cutting verification gates
- stop on first regression instead of merging all branches first

### Risk: final deployment still points at an older feature image

Mitigation:

- require a new final image tag and explicit override update before cutover

### Risk: live rollout without a clean rollback artifact

Mitigation:

- generate rollback compose snapshots before any live switch

### Risk: false completion claims based on stale evidence

Mitigation:

- all success claims require fresh verification run in the final integrated state
