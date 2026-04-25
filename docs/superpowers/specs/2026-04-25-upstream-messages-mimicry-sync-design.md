# Xlabapi Upstream Messages Mimicry Sync Design

**Date:** 2026-04-25

**Goal:** Absorb a narrow set of upstream `/v1/messages`-related compatibility fixes into `xlabapi` to improve Claude-client parity, cache behavior, and reasoning-effort observability without pulling in the broader `cc-mimicry-parity` change set.

## Summary

This design intentionally limits the integration scope to four upstream commits:

- `165553cf` `fix(gateway): use full beta list in buildUpstreamRequest mimicry path`
- `66d64545` `feat(claude): add ttl to cache_control with default 5m`
- `f3233db0` `fix(gateway): apply D/E/F mimicry to native /v1/messages and count_tokens paths`
- `6530776a` `fix: support xhigh reasoning effort in usage records for Claude Messages API`

The integration does not attempt to absorb the full `cc-mimicry-parity` branch or any unrelated upstream gateway changes.

The objective is to reduce the gap between `xlabapi` and current upstream behavior exactly where recent user feedback points: Claude-compatible clients using `/v1/messages`, cache behavior around `cache_control`, and request/usage visibility for reasoning effort.

## Problem

`xlabapi` already contains earlier fixes for OpenAI-compatible `/v1/messages`, prompt cache key propagation, and several thinking/cache accounting issues.

However, current branch inspection still shows three meaningful drifts from upstream:

1. native `/v1/messages` OAuth mimicry is not running the same D/E/F body-rewrite pipeline as the newer upstream code
2. gateway-generated `cache_control` blocks still omit `ttl`, making proxied Claude-style traffic diverge from real Claude Code traffic
3. reasoning-effort normalization for Claude request parsing still ignores `xhigh`, weakening usage observability and incident diagnosis

These drifts are small enough to integrate selectively, but they sit on sensitive gateway code. A broad upstream merge would increase regression risk more than necessary.

## Decision

Three strategies were considered.

### Option 1: selective upstream sync of the four identified commits

Advantages:

- smallest behavior change that matches the current incident hypothesis
- easiest to audit and test
- least likely to disturb `xlabapi`-specific gateway customizations
- leaves room for a second phase of branch-specific adaptation after verification

Disadvantages:

- does not fully close the overall `cc-mimicry-parity` gap
- may still leave unrelated Claude-client mismatches for later work

### Option 2: absorb the whole recent Claude Code mimicry chain

Advantages:

- highest upstream parity
- fewer follow-up syncs if the full chain is eventually desired

Disadvantages:

- significantly larger diff surface
- higher chance of pulling in unrelated request-shaping behavior
- harder to attribute regressions if behavior changes after rollout

### Option 3: ignore upstream commit boundaries and hand-reimplement the symptoms

Advantages:

- maximum flexibility
- can be tailored to current branch assumptions

Disadvantages:

- weak traceability
- easier to miss important edge cases embedded in upstream tests
- invites accidental behavior drift from upstream intent

### Chosen approach

Use **Option 1: selective upstream sync of the four identified commits**.

This preserves the current `xlabapi` business baseline while directly targeting the highest-signal compatibility gaps.

## Goals

- Bring native `/v1/messages` OAuth mimicry closer to upstream without importing the full mimicry stack.
- Ensure gateway-generated `cache_control` blocks carry a default `ttl` compatible with the newer upstream policy.
- Apply the native-message D/E/F body rewrite pipeline to both `/v1/messages` and `/v1/messages/count_tokens`.
- Preserve and expose `xhigh` reasoning effort in Claude request parsing and usage recording.
- Keep the integration auditable by referencing upstream commit intent directly.
- Avoid touching unrelated in-progress local working tree changes.

## Non-Goals

- Do not absorb the complete `cc-mimicry-parity` series.
- Do not change the product decision that Claude-compatible clients may target OpenAI-backed groups.
- Do not redesign OpenAI-compatible Anthropic model mapping semantics in this phase.
- Do not merge broad upstream `gateway_service.go` refactors unless required for the four selected behaviors.
- Do not deploy in this phase.

## Scope Boundary

This phase is complete when `xlabapi` includes the behavior equivalent of the four selected commits and passes targeted regression tests.

This phase is not responsible for:

- full Claude Code header/body parity against upstream main
- broader `messages` admin-model-mapping features
- changes to `AnthropicToResponses` effort semantics
- any follow-up branch-specific functional adaptations beyond what is strictly needed to land the four selected patches safely

## Current State

The current branch already has:

- prompt cache key injection for OpenAI-compatible `messages -> responses`
- SSE-to-JSON handling for non-streaming upstream responses
- a buffered Responses stream accumulator for `/v1/messages`
- earlier fixes for default OpenAI reasoning restoration and cache usage accounting

The current branch still differs from upstream in relevant ways:

- native `/v1/messages` mimicry path does not run the same D/E/F pipeline now used upstream for OAuth-parity requests
- gateway-generated `cache_control` payloads only include `type`, not `ttl`
- OAuth mimicry still uses a reduced `anthropic-beta` set rather than the newer full set used upstream
- `NormalizeClaudeOutputEffort` accepts `low|medium|high|max` but not `xhigh`

## File Boundary

### Primary backend files

- Modify [gateway_service.go](/root/sub2api-src/backend/internal/service/gateway_service.go)
  - adopt the selected upstream mimicry behavior in the native `/v1/messages` and `count_tokens` paths
  - switch OAuth mimicry beta handling to the chosen upstream-equivalent full beta set
  - add `ttl` support to gateway-generated `cache_control` payloads

- Modify [constants.go](/root/sub2api-src/backend/internal/pkg/claude/constants.go)
  - add a default cache-control TTL constant used by the gateway-generated payloads
  - add any helper used by the narrowed beta-list integration if not already present

- Modify [types.go](/root/sub2api-src/backend/internal/pkg/apicompat/types.go)
  - add `cache_control.ttl` support where needed so typed Anthropic payloads can carry the new field cleanly

- Modify [gateway_request.go](/root/sub2api-src/backend/internal/service/gateway_request.go)
  - accept and normalize `xhigh` in Claude `output_config.effort`

### Test files

- Modify or add tests under:
  - [gateway_request_test.go](/root/sub2api-src/backend/internal/service/gateway_request_test.go)
  - existing `gateway_service` test files that cover:
    - OAuth mimicry beta headers
    - system/message `cache_control` serialization
    - native `/v1/messages` mimicry pipeline
    - native `/v1/messages/count_tokens` mimicry pipeline

The exact test file list should be chosen to match the current branch’s test organization rather than forcing upstream file layout.

## Integration Strategy

The integration should be done as a **selective behavioral transplant**, not as a raw multi-commit cherry-pick.

Reason:

- `gateway_service.go` has diverged in `xlabapi`
- straight cherry-picks are likely to create noisy conflicts
- only a subset of the upstream behavior is intentionally in scope

Implementation should still use the upstream commits as the source of truth for intent:

1. inspect each upstream patch
2. identify the exact behavior change
3. implement the equivalent change in the current `xlabapi` structure
4. add or adapt tests proving the selected behavior exists here

## Testing Strategy

This work should follow TDD with targeted regression coverage.

At minimum, add or update tests proving:

1. **full beta list on native OAuth mimic path**
   native `/v1/messages` OAuth mimic requests include the expected upstream-style required beta tokens

2. **default cache-control TTL**
   gateway-generated `cache_control` blocks serialize with `ttl="5m"` unless the client explicitly supplied a different supported TTL

3. **D/E/F pipeline on native `/v1/messages`**
   native OAuth `/v1/messages` requests run:
   - message cache-control stripping where intended
   - message cache breakpoint injection
   - tool-name rewrite or tools-last breakpoint handling

4. **same D/E/F pipeline on `/v1/messages/count_tokens`**
   the count-tokens path applies the same narrowed mimicry behavior

5. **Claude `xhigh` reasoning effort parsing**
   `output_config.effort="xhigh"` is preserved through normalization and is available for usage recording

After targeted tests pass, run the relevant gateway/service package tests for the touched files.

## Risks

### Risk: mimicry patch changes request bytes enough to alter upstream routing or third-party detection outcomes

Mitigation:

- limit scope to the four selected behaviors
- verify serialized body/header deltas with focused tests instead of broad gateway rewrites

### Risk: native `/v1/messages` D/E/F integration conflicts with current `xlabapi` custom request normalization

Mitigation:

- apply upstream behavior after the branch’s existing normalization step, matching the current control flow
- avoid moving unrelated logic in `gateway_service.go`

### Risk: adding `ttl` to generated cache-control blocks changes cache consumption characteristics

Mitigation:

- preserve upstream’s chosen default `5m`
- do not alter client-supplied TTL semantics in this phase beyond the selected upstream behavior

### Risk: `xhigh` support changes observability but not actual upstream execution behavior

Mitigation:

- document this explicitly as an observability correctness fix
- keep it in scope because accurate `reasoning_effort` data materially improves future debugging

## Success Criteria

This phase succeeds when:

- the branch includes equivalent behavior for the four selected upstream commits
- targeted regression tests for beta headers, cache TTL, native D/E/F mimicry, and `xhigh` parsing all pass
- no unrelated local working-tree changes are modified or committed
- the resulting diff remains narrow enough to review as a focused compatibility sync

## Follow-Up

If user feedback improves after this narrow sync, the next phase can focus on branch-specific functional adaptation.

If the same symptom persists, the next highest-probability investigation target is the OpenAI-compatible Anthropic conversion layer, especially the current rule that `thinking.type` is ignored for `AnthropicToResponses` effort derivation.
