# Xlabapi Upstream Main Rebase Design

**Date:** 2026-04-30

**Goal:** Move `xlabapi` onto the latest `upstream/main` baseline while preserving `xlabapi` production database compatibility, OpenAI/Codex/Claude compatibility, local subscription and model behavior, and the requested upstream feature set.

## Summary

The selected direction is no longer a narrow selective sync. The target is to make `upstream/main` the new functional baseline, then replay and reconcile the important `xlabapi` local behavior on top.

As of 2026-04-30:

- `xlabapi` is based on the local branch at `363f5ab4 docs: add upstream v0121 stability sync plan`.
- `origin/xlabapi` is behind local `xlabapi` by two documentation commits.
- `upstream/main` is at `48912014 chore: sync VERSION to 0.1.121`.
- `xlabapi...upstream/main` shows roughly 245 local-side commits and 770 upstream-side commits.
- A dry merge conflict scan shows heavy overlap in Ent generated code, migrations, account/admin handlers, channels, affiliate/invite, settings, gateway compatibility, Sora/image paths, and frontend account/channel/settings/user surfaces.

The project will therefore use a structured rebase-style integration rather than a direct merge into the live `xlabapi` branch.

## Problem

`xlabapi` has accumulated production-specific behavior that must not be lost:

- OpenAI, Codex, Claude, Anthropic, Responses, Messages, WebSocket, and image compatibility patches.
- `gpt-5.4`, `gpt-5.5`, Claude/Sonnet, image, reasoning, and default-instructions behavior.
- shared subscription products, daily carryover, subscription pro/image groups, group multipliers, billing adjustments, and usage display behavior.
- available-channels and affiliate transitions already adapted to local product decisions.
- production migrations already applied or intended to be applied on `xlabapi`.
- local Sora/image API compatibility paths that upstream has partly removed or reshaped.

At the same time, upstream now contains required capabilities:

- account bulk edit improvements.
- Vertex service-account support.
- WebSearch tri-state and Anthropic API Key web-search emulation.
- balance/quota notification features.
- channel and available-channel improvements.
- OpenAI/Codex/Claude compatibility fixes.
- scheduler, sticky-session, compressed-body, stream-failover, cache-reset, and frontend persistence stability fixes.
- broader upstream schema, settings, auth identity, payment, channel monitor, and affiliate work.

A direct merge would mix all of this into one conflict resolution. That would make it hard to prove which behavior was preserved and would create high risk around migrations and generated code.

## Decision

Use **upstream-main baseline with staged `xlabapi` replay**.

The integration branch will start from `upstream/main`, not from current `xlabapi`. Then it will apply `xlabapi` behavior back in controlled domains:

1. migration and generated-code compatibility
2. gateway compatibility
3. channel/model-surface migration compatibility
4. account bulk edit
5. Vertex
6. WebSearch and notifications
7. runtime stability
8. final verification and cutover

This makes `upstream/main` the baseline while still treating existing `xlabapi` behavior as required product behavior, not disposable fork drift.

## Options Considered

### Option 1: direct merge `upstream/main` into `xlabapi`

Advantages:

- preserves upstream history in the simplest Git shape
- makes it obvious that upstream was merged

Disadvantages:

- very large conflict surface
- harder to isolate regressions by feature area
- generated code and migrations become especially risky
- local gateway behavior can be overwritten by conflict resolution mistakes

This option is rejected.

### Option 2: keep `xlabapi` as base and cherry-pick upstream in batches

Advantages:

- keeps the current production branch shape
- lower initial conflict cost

Disadvantages:

- does not satisfy the goal of tracking upstream as the new baseline
- keeps future upstream sync work expensive
- encourages selective patch drift

This option is rejected for this phase.

### Option 3: start from `upstream/main`, replay `xlabapi` by domain

Advantages:

- achieves the requested upstream-main baseline
- lets each local behavior be reintroduced deliberately
- makes migrations and generated code easier to reason about
- allows test gates after each domain

Disadvantages:

- larger planning and validation cost
- requires careful classification of local patches
- may need temporary compatibility shims during the transition

This option is selected.

## Product Direction

### Channels and model plaza compatibility

The long-term direction is channel-first, aligned with upstream `Available Channels`.

The current `xlabapi` model-plaza concept must not disappear abruptly. Compatibility requirements:

- Keep legacy `/models` and model-plaza entry points as redirects, aliases, or compatibility routes during the transition.
- Make the user-facing primary surface `Available Channels`.
- Ensure users can still find models by model name, platform, channel, group, and price.
- Keep model pricing, group multiplier, subscription-group, image-group, and local channel billing semantics intact.
- Do not reintroduce a separate long-term first-class model plaza if the same capability is already represented through channels.

### OpenAI, Codex, and Claude compatibility

Existing compatibility must be fully preserved.

Required behavior includes:

- OpenAI Responses, Chat Completions, Messages, image, and WebSocket compatibility.
- Codex compact payload support and Codex-specific payload normalization.
- Claude/Anthropic bridge and SSE behavior.
- default instructions behavior already added by `xlabapi`.
- reasoning relay and reasoning-item safety.
- model aliases and local support for `gpt-5.4`, `gpt-5.5`, Claude/Sonnet variants, and image-group routing.
- upstream fixes from `0.1.117..0.1.121`, including stream EOF failover, sanitized stream errors, previous-response item references, unsupported field stripping, and tool-choice compatibility.

Upstream compatibility fixes are additive unless they directly conflict with a stronger `xlabapi` behavior. In conflicts, preserve the union of behavior and add tests for the local contract.

## Required Upstream Features

### Account bulk edit

The following upstream chain is required:

- `65c27d2c docs: add account bulk edit scope design`
- `54de4e00 docs: add account bulk edit implementation plan`
- `25c7b0d9 feat: support filter-target account bulk update`
- `2ab6b34f feat: add filtered-result account bulk edit`
- `a161f9d0 feat: align OpenAI bulk edit compact settings`
- `53b24bc2 fix: tighten account bulk edit target typing`

The implementation must adapt bulk edit to all current account fields that matter in `xlabapi`, including platform, channel, quota, RPM, compact mode, WebSocket/Responses settings, model restrictions, priority, billing multiplier, and any new Vertex/WebSearch fields added during this integration.

### Vertex

The following upstream chain is required:

- `6d11f9ed Add Vertex service account support`
- `489a4d93 Show today stats for Vertex usage window`
- `93d91e20 fix(vertex): audit fixes for Vertex Service Account feature (#1977)`

Vertex must be integrated through account creation/editing, account tests, scheduling, credential handling, usage windows, and frontend visibility.

### WebSearch and notifications

The required upstream behavior includes:

- WebSearch tri-state.
- Anthropic API Key web-search emulation.
- quota and balance notifications.
- notify email entries.
- recharge URL in notification context.
- relevant public settings DTO fields.
- frontend settings and account form support.

This must be reconciled with current `xlabapi` Antigravity WebSearch handling, model pricing, account stats, email service behavior, and public settings injection.

### Runtime stability

The required upstream stability set includes:

- compressed request body decoding and decompression bomb protection.
- scheduler snapshot race fixes.
- sticky-session scheduling improvements.
- API key rate-limit cache reset behavior.
- Anthropic stream EOF failover before client output.
- Anthropic-standard SSE error events.
- stream error sanitization.
- OpenAI unsupported field stripping.
- previous-response item-reference inference.
- table page-size localStorage persistence.
- ops retention `0` behavior if it fits the upstream-main baseline.

## Migration And Schema Policy

Production database compatibility is mandatory.

The integration must not rewrite already-shipped `xlabapi` migration history in a way that breaks existing deployments.

Observed migration state:

- Current `xlabapi` contains local migrations through `137_clear_legacy_subscription_carryover.sql`.
- `upstream/main` contains migrations through `133_affiliate_rebate_freeze.sql`.
- The numbering ranges overlap semantically but not always by feature.
- Upstream has schema work for payment, auth identity, pending auth sessions, channel monitors, subscription plans, affiliate, WebSearch, notifications, and account stats.
- `xlabapi` has local subscription products, image groups, carryover cleanup, and production checksum compatibility work.

Policy:

- Treat upstream schema as the new desired schema baseline.
- Preserve existing `xlabapi` migration files that may already exist in production.
- Do not simply overwrite local migration files with upstream files of the same or nearby number.
- Add new compatibility migrations after the current `xlabapi` high-water mark, starting at the next safe number.
- Use idempotent SQL where practical for compatibility backfills and index creation.
- Regenerate Ent only after schema decisions are finalized for the batch.
- Keep a migration mapping document that records upstream migration, local equivalent, and final integration action.

## Integration Branching

Create a new isolated branch from `upstream/main`:

```text
integrate/xlabapi-upstream-main-rebase-20260430
```

Recommended worktree:

```text
/root/.config/superpowers/worktrees/sub2api-src/xlabapi-upstream-main-rebase-20260430
```

The existing `xlabapi` branch remains untouched until the integration branch passes verification.

## Implementation Phases

### Phase 1: baseline and classification

Create classification documents for:

- upstream migration map
- `xlabapi` local commits that must be replayed
- local commits that are superseded by upstream
- local commits that are documentation-only or obsolete
- local commits that need manual fusion

Exit criteria:

- every required domain has a source list and destination strategy.

### Phase 2: schema, migrations, Ent, and Wire

Reconcile schema first because later code depends on stable generated types.

Tasks:

- compare upstream and `xlabapi` Ent schema.
- identify local `xlabapi` schema not present in upstream.
- preserve production migrations and add compatibility migrations.
- regenerate Ent and Wire.
- run schema and migration-focused tests.

Exit criteria:

- generated code compiles.
- migration map is documented.
- no migration overwrites production history.

### Phase 3: gateway compatibility

Replay and reconcile OpenAI/Codex/Claude compatibility.

Tasks:

- port local model alias and default-instruction behavior.
- preserve image and Sora compatibility where still required.
- absorb upstream Responses, Anthropic, Codex, and WebSocket fixes.
- preserve local billing and usage metadata.
- add tests for each local compatibility contract.

Exit criteria:

- focused gateway, apicompat, OpenAI, Codex, Claude, image, and WebSocket tests pass.

### Phase 4: channels and model-surface compatibility

Make channels the primary user-facing discovery model while preserving model-plaza compatibility.

Tasks:

- adopt upstream available-channel behavior as the primary surface.
- preserve `/models` compatibility.
- keep searchable model visibility inside channel rows/sections.
- preserve `xlabapi` subscription, group, image, and billing behavior.
- reconcile frontend navigation and copy.

Exit criteria:

- users can discover models through available channels.
- old route compatibility remains.
- subscription/image/group billing behavior is preserved.

### Phase 5: account bulk edit

Port account bulk edit after account schema and frontend account models are stable.

Tasks:

- add filter-target bulk update.
- add filtered-result selection.
- add OpenAI compact settings bulk edit.
- adapt field typing for local and new upstream fields.
- cover backend service, handler, and frontend modal behavior.

Exit criteria:

- bulk edit works for visible filters and local account field set.

### Phase 6: Vertex

Port Vertex service-account support.

Tasks:

- support Vertex credential fields.
- support create/edit/test flows.
- support scheduling and usage window stats.
- expose frontend controls and translations.

Exit criteria:

- Vertex account can be configured, tested, scheduled, and shown in usage stats.

### Phase 7: WebSearch and notifications

Port WebSearch and notification systems.

Tasks:

- add WebSearch tri-state.
- add Anthropic API Key web-search emulation.
- add balance/quota notify settings and recipient entries.
- expose public settings and admin/user UI behavior.
- reconcile email service and recharge URL behavior.

Exit criteria:

- WebSearch behavior is configurable and notifications work without breaking existing email flows.

### Phase 8: runtime stability

Confirm and finish upstream stability fixes.

Tasks:

- verify compressed request body handling.
- verify scheduler and sticky-session behavior.
- verify stream error/failover behavior.
- verify API key rate-limit cache reset.
- verify frontend table page-size persistence.

Exit criteria:

- stability tests pass and no known upstream `0.1.121` stability fix is missing without a documented reason.

### Phase 9: final verification and cutover

Tasks:

- run targeted backend package tests.
- run frontend typecheck/build/focused Vitest tests.
- run migration compatibility tests.
- run smoke checks if local services are available.
- prepare final merge or fast-forward strategy for `xlabapi`.

Exit criteria:

- integration branch is ready to become the new `xlabapi`.

## Verification Strategy

Each phase must include focused tests before moving on. At minimum:

- `go test -tags=unit ./internal/pkg/apicompat`
- focused gateway/OpenAI/Codex/Claude tests under `backend/internal/service` and `backend/internal/handler`
- account bulk edit tests under `backend/internal/service` and `backend/internal/handler/admin`
- migration/schema tests under `backend/internal/repository`, `backend/migrations`, and `backend/ent`
- frontend `pnpm` typecheck/build or the repository's equivalent scripts
- focused Vitest tests for account, settings, channels, usage, and routing

Full-suite failures may be triaged when unrelated to the changed phase, but any failure in the touched domain blocks the phase.

## Non-Goals

- Do not deploy directly from the integration branch before verification.
- Do not discard `xlabapi` local OpenAI/Codex/Claude compatibility.
- Do not remove model-plaza compatibility routes during this integration.
- Do not rewrite production migration history.
- Do not manually edit generated Ent/Wire code as a long-term substitute for regeneration.
- Do not claim upstream parity until the required features and stability checks have been validated.

## Risks

- Migration conflicts may require careful compatibility SQL and production dry-runs.
- Some upstream payment/auth identity work may be structurally entangled with settings and public DTOs even if not directly requested.
- Sora/image behavior may need manual restoration because upstream removed or reshaped some local files.
- Frontend account forms may have multiple overlapping versions of account fields after bulk edit, Vertex, WebSearch, and local compact settings are combined.
- Gateway regressions are the highest operational risk and need the most focused tests.

## Success Criteria

The work is complete when:

- `integrate/xlabapi-upstream-main-rebase-20260430` is based on `upstream/main`.
- Required upstream features are present: account bulk edit, Vertex, WebSearch/notifications, available channels, and runtime stability.
- `xlabapi` OpenAI/Codex/Claude compatibility is preserved with tests.
- channel-first product direction is active while model-plaza compatibility remains.
- production migration compatibility is documented and implemented.
- targeted backend and frontend verification passes.
- the branch can be merged or promoted to `xlabapi` with a clear rollout note.
