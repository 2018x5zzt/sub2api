# Available Channels Public Surface Design

**Date:** 2026-04-27

**Goal:** Align `xlabapi` user-facing discovery with upstream by exposing only `Available Channels` as the primary browsing surface, while keeping supported models visible as channel details instead of restoring a separate local `Model Hub`.

## Summary

`xlabapi` previously had a local `Model Hub` experience centered on browsing models first and treating groups or pricing sources as secondary context.

After the upstream absorption that introduced affiliate and available-channel support, the runtime now already behaves differently:

- the old user `ModelHubView` was removed
- `/models` redirects to `/available-channels`
- user-facing data now comes from `/channels/available`
- backend filtering is channel-centric and scoped by the current user's accessible groups

The remaining design question is not whether to expose upstream channels at all. That is already the effective implementation direction. The real question is what public product shape should be considered correct going forward.

This design chooses a strict upstream-aligned answer:

- users browse **channels**
- models remain visible, searchable, and priceable **inside channel details**
- there is **no separate first-class model-plaza surface** in this phase
- `/models` remains only as a compatibility redirect, not a product concept

## Problem

The current branch has a product-language mismatch.

From the implementation side, upstream `Available Channels` has already replaced the old local model-plaza runtime. From the user-expectation side, there is still concern that "users must not lose the ability to see what models are available."

If this mismatch is left unresolved, the product will drift in one of two bad directions:

1. runtime and copy say different things
   - backend/admin centers on channels
   - frontend/navigation still talks about a model plaza
   - future upstream syncs become harder because product language and data shape disagree

2. local compatibility layers reintroduce a second browsing concept
   - users see both "model plaza" and "available channels"
   - the same data is represented twice
   - future maintenance cost grows without adding real capability

The system needs one explicit public contract so navigation, naming, backend filtering, and future upstream merges all point in the same direction.

## Current State

The current codebase already contains the core pieces for an upstream-aligned channel surface.

### Frontend state

- `frontend/src/views/user/ModelHubView.vue` has been removed.
- `frontend/src/views/user/AvailableChannelsView.vue` is the current user-facing discovery page.
- `frontend/src/router/index.ts` redirects `/models` to `/available-channels`.
- `frontend/src/components/layout/AppSidebar.vue` already drives visibility through the `available_channels_enabled` public setting.
- `frontend/src/components/channels/AvailableChannelsTable.vue` presents channel rows, platform sections, user-visible groups, and supported models.

### Backend state

- `backend/internal/handler/available_channel_handler.go` exposes `/channels/available`.
- `backend/internal/service/channel_available.go` builds the user-facing available-channel DTO.
- `ChannelService.ListAvailable` uses channel entities plus active groups to produce a public browsing shape.
- `AvailableChannelHandler.List` already filters channels to:
  - authenticated user context
  - active channels only
  - groups available to the current user
  - non-empty platform sections

### Product state

The implementation direction is already "channel-first," but the product contract has not been explicitly documented. That is the gap this design closes.

## Options Considered

### Option 1: full upstream alignment with `Available Channels`

Users browse channels as the top-level discovery object. Supported models are shown within each channel and platform section. No separate model plaza is restored.

Advantages:

- matches current upstream structure and naming
- preserves the cleanest path for future upstream merges
- keeps backend/admin/runtime/product language aligned
- avoids reintroducing a local branch-only surface

Disadvantages:

- users who were accustomed to browsing by model first must adjust
- "find a model" becomes "find the channel that exposes the model"

### Option 2: keep the name `Model Hub` but back it with channel data

The runtime remains channel-driven, but the user-facing navigation and title continue to say `Model Hub`.

Advantages:

- softens user-facing terminology changes
- reduces short-term migration shock

Disadvantages:

- product language becomes misleading
- backend/admin say "channel," frontend says "model hub"
- future upstream syncs will repeatedly require local naming forks

### Option 3: expose both `Model Hub` and `Available Channels`

The system keeps two separate user entry points that draw on overlapping data.

Advantages:

- maximum short-term familiarity
- easiest to preserve historical terminology

Disadvantages:

- duplicated navigation and mental models
- unclear difference between two pages that represent the same capabilities
- unnecessary maintenance surface for a branch trying to converge with upstream

## Decision

Choose **Option 1: full upstream alignment with `Available Channels`**.

This decision is based on three constraints that matter more than short-term naming familiarity:

1. `xlabapi` has already absorbed the upstream runtime shape.
2. the user explicitly approved following the official product direction first.
3. future upstream compatibility is more valuable than preserving the retired local model-plaza abstraction.

## Public Product Contract

The user-facing browsing contract in this phase is:

- the primary discovery surface is **Available Channels**
- the page exposes only channels the current user can actually access
- each channel shows the groups and platforms through which it is usable
- supported models are displayed within each channel section
- pricing remains visible at the model level where data exists
- there is no separate standalone "model plaza" page in the live product

In plain terms:

Users browse **what channel they can use**, and within that context they inspect **what models the channel supports**.

## Public Exposure Rules

The frontend should expose only channels that satisfy all of the following:

1. the feature is globally enabled through `available_channels_enabled`
2. the channel status is `active`
3. the current user has at least one accessible group attached to the channel
4. the channel still yields at least one visible platform section after filtering

The frontend should not expose:

- inactive channels
- channels with no user-visible groups
- admin-only internal management data
- a synthetic global model list detached from channels

This means the public surface is **not** "all channels in the system." It is "the subset of active channels that the current user can actually reach."

## Information Architecture

The user-facing hierarchy should be:

1. **Channel**
   - channel name
   - channel description

2. **Platform section**
   - `openai`, `anthropic`, `gemini`, and similar platform partitions

3. **Visible groups**
   - the specific user-accessible groups bound to that channel and platform
   - group metadata such as subscription type, exclusivity, and rate multiplier

4. **Supported models**
   - model identifiers available under that platform section
   - pricing information when present

This structure preserves the ability to discover models without making models the top-level abstraction.

## Navigation And Naming

The product should use upstream terminology directly in this phase.

### Required naming

- sidebar entry: `Available Channels` / `可用渠道`
- route: `/available-channels`
- page title: `Available Channels` / `可用渠道`

### Compatibility behavior

- `/models` remains a redirect to `/available-channels`
- the redirect exists for compatibility only
- UI copy should not present `Model Hub` as an active product surface

### Explicit non-goal for this phase

Do not keep `Model Hub` as a visible alias in navigation, titles, or primary empty states.

If a later product pass wants a branded or more user-friendly label, that should be treated as a separate copy/design decision after the upstream-aligned baseline is stable.

## Search And Discoverability

The old concern behind "users cannot lose the model plaza" is really a discoverability concern, not a requirement for a separate page type.

Discoverability should be preserved through the available-channels page itself.

The page should support finding content by:

- channel name
- channel description
- platform name
- group name
- model name

The current `AvailableChannelsView` filtering direction already supports this structure and should remain channel-centric rather than being reshaped into a global model browser.

## Data Semantics

Supported models are part of the public contract, but they are **secondary objects**.

That means:

- models must remain visible in the UI
- models may still display pricing, billing mode, and platform hints
- models do not become standalone navigation items
- models are always interpreted in the context of a channel and platform section

This is important because the same model name can have different operational meaning depending on channel mapping, billing source, platform grouping, or pricing restrictions.

Keeping models contextual prevents the UI from implying a false global guarantee.

## Settings Contract

The `available_channels_enabled` public setting remains the feature gate for the browsing surface.

If the flag is disabled:

- sidebar entry should be hidden
- the public channel page should not be presented as a usable feature
- the handler may return an empty list as it does today

This design does not introduce a new `model_hub_enabled` setting and does not preserve dual feature switches for old and new browsing concepts.

## Why No Additional Visibility Flag Yet

One possible extension would be adding a channel-level field such as:

- `user_visible`
- `public_visible`

That would allow operations to keep some active channels usable for routing while hiding them from the user-facing list.

This design intentionally does **not** add that field yet.

Reason:

- the current problem is conceptual alignment, not missing filtering primitives
- the existing rules are sufficient for the approved channel-first public surface
- adding local-only visibility semantics now would create another branch-specific divergence from upstream

If real production usage later proves that some active channels must remain hidden even when attached to user-visible groups, that should be handled as a separate design and implementation step.

## Goals

- make `Available Channels` the single official user discovery surface
- preserve model discoverability within channel details
- keep runtime behavior, product naming, and navigation aligned
- minimize future merge friction with upstream
- avoid restoring the removed local `Model Hub` runtime

## Non-Goals

- do not restore `frontend/src/views/user/ModelHubView.vue`
- do not add a second user-facing page that duplicates channel data as a model-first catalog
- do not redesign backend channel filtering behavior beyond what the current handler already does
- do not introduce new channel visibility fields in this phase
- do not deploy copy customization that intentionally diverges from upstream naming in this phase

## Risks

### Risk: users miss the old model-first browsing pattern

Mitigation:

- keep models visible within each channel section
- preserve model-name search on the available-channels page
- keep `/models` redirect compatibility so old bookmarks do not break

### Risk: internal or operational channels appear in the public list unintentionally

Mitigation:

- rely on current group access filtering first
- treat any leakage as a signal for a later explicit visibility-field design, not for reintroducing a model-plaza layer

### Risk: product copy remains half old and half new

Mitigation:

- standardize on `Available Channels` in navigation and page-level copy
- avoid visible `Model Hub` aliases in this phase

## Success Criteria

This design is successful when all of the following are true:

- `Available Channels` is the only first-class user browsing entry for model access discovery
- users still see supported models and pricing inside channel details
- `/models` exists only as a compatibility redirect
- product terminology, frontend navigation, and backend filtering all describe the same public contract
- no new local model-plaza runtime is introduced

## Implementation Direction

The follow-up implementation plan should stay narrow and pragmatic.

Expected work is primarily:

- audit user-facing copy to ensure `Available Channels` is the visible term
- verify sidebar and route behavior remain consistent with the feature flag
- verify search, empty states, and page descriptions do not imply a separate model plaza
- avoid broad structural frontend rewrites because the current runtime shape already matches the chosen design

This is a convergence task, not a redesign-from-scratch task.
