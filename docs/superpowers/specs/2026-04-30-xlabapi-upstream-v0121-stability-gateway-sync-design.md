# Xlabapi Upstream v0.1.121 Stability And Gateway Sync Design

**Date:** 2026-04-30

**Goal:** Selectively absorb upstream `v0.1.119..v0.1.121` stability, security, scheduler, and gateway/API compatibility fixes into `xlabapi` while preserving existing `xlabapi` product behavior and deferring large upstream feature migrations.

## Summary

`xlabapi` has diverged far enough from `upstream/main` that a broad merge is not a safe first step.

After refreshing remotes on 2026-04-30:

- `xlabapi` is at `cca30da8 fix: support sonnet 4.6 thinking model`
- `origin/xlabapi` matches local `xlabapi`
- `upstream/main` is at `48912014 chore: sync VERSION to 0.1.121`
- `xlabapi...upstream/main` shows roughly 243 `xlabapi`-only commits and 770 upstream-only commits
- `backend/cmd/server/VERSION` on `xlabapi` remains `0.1.105`, even though the branch has selectively absorbed later upstream work

The practical goal is therefore not to claim full `v0.1.121` parity. The goal is to land a controlled compatibility checkpoint that brings over the upstream fixes most likely to affect production stability and API compatibility.

## Problem

Current upstream contains many changes that are valuable in isolation but too broad to merge together:

- scheduler and sticky-session reliability fixes
- HTTP request body decompression and decompression bomb protection
- OpenAI Responses, Anthropic bridge, Codex transform, and WebSocket continuation fixes
- API key rate-limit cache reset behavior
- admin settings and small frontend persistence fixes
- OpenAI Fast/Flex Policy
- account bulk edit
- Vertex service accounts
- payment, monitor, profile identity, OAuth adoption, and settings-page refactors

Several of those areas overlap with `xlabapi` branch-specific behavior:

- shared subscription products and subscription group billing
- affiliate-only invite replacement
- available channels public surface
- async OpenAI image jobs
- `gpt-5.4`, `gpt-5.5`, and Claude/Sonnet compatibility changes
- production migration checksum compatibility

Mixing all upstream-only commits into one merge would make conflicts hard to reason about and would risk silently overwriting `xlabapi` customizations.

## Decision

Three strategies were considered.

### Option 1: full merge from `upstream/main`

Advantages:

- maximum upstream parity
- preserves upstream history in one operation

Disadvantages:

- very large conflict surface across backend schema, settings, payment, monitor, profile, and frontend pages
- hard to distinguish bug fixes from product changes
- high risk of overriding `xlabapi` custom business behavior
- expensive to verify before deployment

### Option 2: feature-priority selective sync

Advantages:

- can bring over visible upstream features such as OpenAI Fast/Flex Policy, account bulk edit, and Vertex service accounts
- avoids unrelated full-branch merge noise

Disadvantages:

- still mixes stability work with product migrations
- each feature touches different UI, settings, service, and test surfaces
- harder to produce a quick stable checkpoint

### Option 3: stability and gateway compatibility first

Advantages:

- narrows the first checkpoint to production reliability and protocol compatibility
- keeps the diff mostly in backend gateway, scheduler, HTTP utility, and focused settings/test files
- allows explicit duplicate detection against patches already ported into `xlabapi`
- creates a cleaner base for later feature-specific migrations

Disadvantages:

- does not provide full `v0.1.121` feature parity
- leaves larger upstream features for later plans

### Chosen approach

Use **Option 3: stability and gateway compatibility first**.

This phase selectively ports upstream behavior from `v0.1.119..v0.1.121` that is a clear stability, security, scheduler, or gateway/API compatibility fix. It does not update the branch version to `0.1.121`, and it does not import unrelated product features.

## Goals

- Absorb upstream HTTP body safety and compressed-request handling fixes that are applicable to `xlabapi`.
- Absorb upstream scheduler and sticky-session reliability fixes that reduce race conditions and bad account selection.
- Absorb upstream OpenAI/Anthropic/Codex/Responses compatibility fixes where they still apply.
- Absorb upstream API key rate-limit cache reset behavior if not already present.
- Preserve `xlabapi` gateway behavior for existing model aliases, reasoning relay handling, OAuth/image flows, and Claude/Sonnet compatibility.
- Detect and document upstream patches already ported under different local commits.
- Keep all implementation changes isolated on a new integration branch.
- Verify with targeted backend tests before broader package checks.

## Non-Goals

- Do not merge all of `upstream/main`.
- Do not claim full upstream `v0.1.121` parity.
- Do not update `backend/cmd/server/VERSION` to `0.1.121` in this phase.
- Do not import upstream payment, channel monitor, profile identity, OAuth adoption, or large settings-page refactors.
- Do not import OpenAI Fast/Flex Policy in this phase.
- Do not import Vertex service accounts in this phase.
- Do not import account bulk edit in this phase.
- Do not redesign `xlabapi` shared subscription, affiliate, available-channels, or image-job behavior.
- Do not deploy in this phase.

## Upstream Source Boundary

The source window is:

```text
v0.1.119..v0.1.121
```

Within that window, this phase targets the following categories.

### Include: stability, security, and gateway/API fixes

- `798fd673 feat(httputil): decode compressed request bodies (zstd/gzip/deflate)`
- `40feb86b fix(httputil): add decompression bomb guard and fix errcheck lint`
- `8bf2a7b8 fix(scheduler): resolve SetSnapshot race conditions and remove usage throttle`
- `733627cf fix: improve sticky session scheduling`
- `53f919f8 fix(api-key): reset rate limit usage cache`
- `30220903 fix(anthropic): drop empty Read.pages in responses-to-anthropic tool input`
- `615557ec fix(openai): avoid implicit image sticky sessions`
- `9fe02bba fix(openai): strip unsupported passthrough fields`
- `3d4ca5e8 fix(openai): preserve current Codex compact payload fields`
- `7452fad8 fix(openai): drop reasoning items from /v1/responses input on OAuth path`
- `04b2866f fix: use Responses-compatible function tool_choice format`
- `63275735 fix(gateway): wrap Anthropic stream EOF as failover error before client output`
- `4c474616 fix(gateway): emit Anthropic-standard SSE error events and failover body`
- `d78478e8 fix(gateway): sanitize stream errors to avoid leaking infrastructure topology`
- `28dc34b6 fix(openai): avoid inferred WS continuation on explicit tool replay`
- `094e1171 fix(openai): infer previous response for item references`

### Include only if low-risk after inspection

- `73b87299 feat: add Anthropic cache TTL injection switch`
- `f084d30d fix: restore table page-size localStorage persistence`
- `4b6954f9 feat(ops): allow retention days = 0 to wipe table on each scheduled cleanup`
- `9d801595 test: update admin settings contract fields`

These are useful, but they touch settings or frontend surfaces. They should be included only when they can land without pulling in the broader settings-page or product-feature refactors.

### Exclude from this phase

- `30f55a1f feat(openai): OpenAI Fast/Flex Policy`
- account bulk edit chain: `25c7b0d9`, `2ab6b34f`, `a161f9d0`, and related tests
- Vertex service account chain: `6d11f9ed`, `489a4d93`, `93d91e20`, and related tests
- payment, channel monitor, profile identity, OAuth adoption, and large admin settings migrations
- documentation-only, sponsor, CLA, asset, and upstream repository-maintenance changes unless required by tests

## Duplicate Detection Policy

Several patches may already be present in `xlabapi` under different local commits or as manually adapted code.

Before porting a commit, implementation must classify it as one of:

- **already equivalent:** no code change needed; record the local evidence
- **partially present:** add the missing behavior and tests only
- **applicable:** port behavior into current `xlabapi` structure
- **not applicable:** skip because the upstream code path does not exist or has been intentionally replaced in `xlabapi`
- **defer:** belongs to a later feature migration

Patch IDs and commit subjects are guides, not proof. Behavior and tests are the source of truth.

## File Boundary

The expected implementation surface is mostly backend.

### Backend HTTP utility

- Modify [body.go](/root/sub2api-src/backend/internal/pkg/httputil/body.go)
- Modify or add [body_test.go](/root/sub2api-src/backend/internal/pkg/httputil/body_test.go)

Responsibilities:

- decode supported compressed request bodies
- enforce decompressed body size protection
- keep error behavior compatible with existing callers

### Backend scheduler and sticky sessions

- Modify [scheduler_cache.go](/root/sub2api-src/backend/internal/repository/scheduler_cache.go)
- Modify [scheduler_cache.go](/root/sub2api-src/backend/internal/service/scheduler_cache.go)
- Modify [scheduler_snapshot_service.go](/root/sub2api-src/backend/internal/service/scheduler_snapshot_service.go)
- Modify [gateway_handler.go](/root/sub2api-src/backend/internal/handler/gateway_handler.go)
- Modify [gateway_service.go](/root/sub2api-src/backend/internal/service/gateway_service.go)
- Modify or add focused scheduler tests under:
  - [scheduler_cache_unit_test.go](/root/sub2api-src/backend/internal/repository/scheduler_cache_unit_test.go)
  - [scheduler_cache_integration_test.go](/root/sub2api-src/backend/internal/repository/scheduler_cache_integration_test.go)
  - [scheduler_snapshot_hydration_test.go](/root/sub2api-src/backend/internal/service/scheduler_snapshot_hydration_test.go)

Responsibilities:

- remove or adapt the upstream usage-throttle behavior only if the current branch still has the same race
- apply sticky-session scheduling fixes without breaking `xlabapi` group/account selection
- preserve branch-specific model and image routing behavior

### Backend OpenAI, Anthropic, Codex, and Responses compatibility

- Modify [responses_to_anthropic.go](/root/sub2api-src/backend/internal/pkg/apicompat/responses_to_anthropic.go)
- Modify [responses_to_anthropic_request.go](/root/sub2api-src/backend/internal/pkg/apicompat/responses_to_anthropic_request.go)
- Modify [anthropic_to_responses.go](/root/sub2api-src/backend/internal/pkg/apicompat/anthropic_to_responses.go)
- Modify [chatcompletions_to_responses.go](/root/sub2api-src/backend/internal/pkg/apicompat/chatcompletions_to_responses.go)
- Modify [openai_codex_transform.go](/root/sub2api-src/backend/internal/service/openai_codex_transform.go)
- Modify [openai_gateway_service.go](/root/sub2api-src/backend/internal/service/openai_gateway_service.go)
- Modify [openai_gateway_chat_completions.go](/root/sub2api-src/backend/internal/service/openai_gateway_chat_completions.go)
- Modify [openai_gateway_messages.go](/root/sub2api-src/backend/internal/service/openai_gateway_messages.go)
- Modify [openai_ws_forwarder.go](/root/sub2api-src/backend/internal/service/openai_ws_forwarder.go)
- Modify [openai_ws_v2_passthrough_adapter.go](/root/sub2api-src/backend/internal/service/openai_ws_v2_passthrough_adapter.go)
- Modify [openai_images.go](/root/sub2api-src/backend/internal/handler/openai_images.go) only if the implicit image sticky-session fix is still relevant

Responsibilities:

- strip or preserve upstream fields according to the selected upstream fix
- avoid dropping `xlabapi` local default-instructions and reasoning relay protections
- preserve current Codex compact payload fields
- handle tool-choice compatibility for Responses transformations
- improve Anthropic stream failover behavior before client output is committed
- sanitize infrastructure details in stream errors
- keep WebSocket continuation inference correct for explicit tool replay and item references

### Backend settings and API key behavior

- Modify [apikey_handler.go](/root/sub2api-src/backend/internal/handler/admin/apikey_handler.go)
- Modify [admin_service.go](/root/sub2api-src/backend/internal/service/admin_service.go)
- Modify [billing_cache_service.go](/root/sub2api-src/backend/internal/service/billing_cache_service.go)
- Modify [wire.go](/root/sub2api-src/backend/internal/service/wire.go)
- Regenerate [wire_gen.go](/root/sub2api-src/backend/cmd/server/wire_gen.go) if DI changes require it
- Modify settings files only if the low-risk settings subset is accepted:
  - [domain_constants.go](/root/sub2api-src/backend/internal/service/domain_constants.go)
  - [setting_service.go](/root/sub2api-src/backend/internal/service/setting_service.go)
  - [settings_view.go](/root/sub2api-src/backend/internal/service/settings_view.go)
  - [settings.go](/root/sub2api-src/backend/internal/handler/dto/settings.go)
  - [setting_handler.go](/root/sub2api-src/backend/internal/handler/admin/setting_handler.go)

Responsibilities:

- reset affected rate-limit cache entries when API key limits change
- avoid broad settings rewrites unless required by the included low-risk settings items

### Frontend

Frontend changes should be avoided in the first backend-focused pass except for the low-risk page-size persistence fix:

- Modify [usePersistedPageSize.ts](/root/sub2api-src/frontend/src/composables/usePersistedPageSize.ts)
- Modify [Pagination.vue](/root/sub2api-src/frontend/src/components/common/Pagination.vue)
- Modify corresponding tests only if the upstream fix still applies cleanly

No payment, monitor, account bulk edit, profile, or settings-page feature migration belongs in this phase.

## Integration Strategy

Use a new isolated branch from `xlabapi`:

```text
integrate/xlabapi-upstream-v0121-stability-gateway-20260430
```

Do not perform a broad merge from `upstream/main`.

For each selected upstream commit:

1. inspect the upstream patch and tests
2. inspect the current `xlabapi` equivalent code path
3. classify the patch using the duplicate detection policy
4. write or adapt a focused failing test when behavior is missing
5. port the minimal behavior into the current branch structure
6. run the focused test
7. keep commits grouped by concern, not by massive upstream ranges

Raw cherry-pick is acceptable only when:

- the touched files have not materially diverged
- conflicts are absent or trivial
- the resulting diff does not bring excluded feature code

Manual behavioral transplant is preferred for gateway and scheduler code because those files have meaningful `xlabapi` drift.

## Testing Strategy

Testing should prove branch coherence rather than full upstream parity.

### Focused backend tests

Run focused tests around each included area:

- `go test -tags unit ./internal/pkg/httputil -run 'Body|Compressed|Decompression|Limit' -count=1`
- `go test -tags unit ./internal/repository -run 'Scheduler|Snapshot|Sticky|Rate' -count=1`
- `go test -tags unit ./internal/service -run 'OpenAI|Codex|Responses|Anthropic|WebSocket|Sticky|Scheduler|Image' -count=1`
- `go test -tags unit ./internal/handler ./internal/handler/admin -run 'Gateway|Warmup|ApiKey|Setting' -count=1`

Exact test names should be adjusted to the current branch after patch inspection.

### Broader backend checks

After focused tests pass:

- `go test -tags unit ./internal/pkg/apicompat ./internal/pkg/httputil ./internal/service ./internal/handler ./internal/handler/admin -count=1`
- `go test ./internal/repository -run 'Scheduler|AccountRepo' -count=1` if integration tests are available in the current environment
- `go test ./internal/server -run 'API|Contract' -count=1` if route or DTO contracts changed

### Frontend checks

Only needed if the low-risk frontend page-size persistence fix is included:

- `npm test -- --run usePersistedPageSize Pagination`
- `npm run typecheck`

Use the repository's actual frontend scripts from `frontend/package.json`; do not invent scripts if they are unavailable.

### General checks

- `git diff --check`
- `git status --short`
- inspect final diff against `xlabapi` for excluded feature leakage

## Risks

### Risk: a gateway fix overwrites `xlabapi` model compatibility behavior

Mitigation:

- treat upstream as behavioral reference, not a forced file replacement
- inspect `gpt-5.4`, `gpt-5.5`, Sonnet, default-instructions, image, and OAuth paths after each gateway patch
- keep gateway tests focused on branch-specific protections

### Risk: scheduler changes alter account selection or billing behavior

Mitigation:

- apply only the race/sticky-session fix needed by the current branch
- run scheduler and gateway selection tests after the scheduler patch
- inspect behavior around group-specific subscriptions and image jobs

### Risk: settings-related low-risk items drag in broader settings refactors

Mitigation:

- include `Anthropic cache TTL` or `ops retention days = 0` only if they can be represented by small DTO/service/view additions
- defer if they require broad admin settings page restructuring

### Risk: duplicate patches are applied twice

Mitigation:

- classify every upstream candidate before editing
- compare local branch commits such as `xlabapi-upstream-v01119-selective-20260428`
- prefer tests that prove behavior rather than blindly replaying commit diffs

### Risk: version number creates false parity

Mitigation:

- do not bump `VERSION` in this phase
- document the result as a selective stability and gateway sync, not a full `v0.1.121` upgrade

## Deliverable

The deliverable is a verified integration branch containing a small series of focused commits.

The final branch should include:

- a patch classification note in commit messages or implementation notes
- targeted tests for newly ported behavior
- no broad upstream feature migrations
- no `VERSION` bump to `0.1.121`
- no deployment changes

After this branch is verified, larger upstream features can be planned independently:

- OpenAI Fast/Flex Policy
- account bulk edit
- Vertex service accounts
- payment and provider system
- channel monitor
- profile identity and OAuth adoption
- broader settings UI/API refactors

## Success Criteria

- Selected upstream stability and gateway/API compatibility behaviors are present or explicitly classified as already equivalent/not applicable.
- Focused backend tests for touched surfaces pass.
- Broader package tests for touched backend packages pass where available.
- `git diff --check` passes.
- Final diff does not include excluded upstream feature migrations.
- Existing `xlabapi` business behavior is preserved by inspection and targeted tests.
