# Xlabapi OpenAI Image2 Compatibility Integration Design

**Date:** 2026-04-23

**Goal:** Integrate the upstream OpenAI Images and `image2` compatibility feature set into `xlabapi` while preserving existing branch-specific OpenAI, Sora, TLS, and account-testing behavior.

## Summary

`xlabapi` currently does not contain the upstream OpenAI Images API surface.

The relevant upstream work is not a single isolated patch. It is a feature chain that introduced:

- OpenAI Images API routes for `generations` and `edits`
- request parsing for JSON and multipart image requests
- account capability selection and scheduler integration
- usage and billing support for image generation
- a Codex/Responses-based compatibility path for OAuth image generation
- frontend account test support for OpenAI image models

This design does not merge all of `upstream/main`, and it does not blindly cherry-pick the full upstream commit chain into `xlabapi`.

Instead, it defines a selective manual integration that uses upstream behavior as the source of truth for the new image feature while keeping `xlabapi`-specific customizations intact.

## Problem

The current `xlabapi` branch is missing the upstream image generation compatibility layer the user wants to absorb.

At the same time, `xlabapi` has already diverged from upstream in sensitive areas:

- OpenAI gateway behavior and request compatibility handling
- account test service behavior
- Sora-specific test flow and cooldown behavior
- TLS fingerprint and direct-endpoint handling for account tests
- existing frontend account test UX

That means a direct cherry-pick strategy is risky. The image feature touches exactly the surfaces where `xlabapi` already differs from upstream.

The integration therefore needs to preserve the branch's existing custom behavior while still bringing over the upstream image feature end to end.

## Decision

Three implementation strategies were considered.

### Option 1: direct cherry-pick of the upstream image commit chain

Advantages:

- keeps upstream commit history
- low design effort if conflicts are small

Disadvantages:

- high conflict risk in `account_test_service`, OpenAI gateway, and frontend files
- likely to import unrelated upstream assumptions
- higher risk of silently dropping `xlabapi` custom behavior

### Option 2: selective manual integration on top of `xlabapi`

Advantages:

- keeps `xlabapi` as the behavioral baseline
- imports only the required image feature surfaces
- allows targeted adaptation in the known-diverged files
- easier to reason about regression risk

Disadvantages:

- requires more careful implementation work than a raw cherry-pick

### Option 3: merge a large slice of `upstream/main` first, then adapt

Advantages:

- potentially reduces future upstream drift

Disadvantages:

- far outside the requested scope
- mixes image compatibility with a broad upstream sync
- materially increases review and regression surface

### Chosen approach

Use **Option 2: selective manual integration on top of `xlabapi`**.

This keeps the branch's current behavior intact and limits the imported scope to the image compatibility feature set the user explicitly asked for.

## Upstream Source Boundary

The upstream commits that define the required behavior are:

- `c5480219 feat(openai): 同步生图 API 支持并接入图片计费调度`
- `00778dca fix openai image request handling`
- `0b85a8da fix: add io.LimitReader bounds to prevent OOM in image handling`
- `eea6f388 使用codex的生图接口代替web2api`
- `1e0d4660 feat: 补充gpt生图模型测试功能`

These commits are treated as behavior references, not as a mandatory one-shot cherry-pick sequence.

The integration will absorb:

- the backend OpenAI Images API surface
- the OAuth image compatibility path through Responses/Codex
- image request parsing and image result transformation
- image billing and usage recording
- frontend and backend account test support for OpenAI image models

The integration will not absorb unrelated upstream history outside these feature surfaces.

## Goals

- Add OpenAI-compatible `/images/generations` and `/images/edits` endpoints to `xlabapi`.
- Preserve `xlabapi`'s current OpenAI gateway behavior outside the new image flow.
- Preserve `xlabapi` custom account test logic, including Sora cooldown and TLS/profile behavior.
- Support both API key and OAuth OpenAI accounts for image generation.
- Bring over image-specific request parsing, usage recording, and billing behavior.
- Bring over frontend account-test support for OpenAI image models and image previews.
- Add targeted regression coverage for the new image flow and the adapted `xlabapi` test surfaces.

## Non-Goals

- Do not perform a general upstream synchronization of `xlabapi`.
- Do not redesign the current `xlabapi` frontend account-test experience.
- Do not remove or weaken current Sora-specific test behavior.
- Do not refactor unrelated gateway or billing code while integrating image support.
- Do not deploy in this phase.

## Current State

The current `xlabapi` branch already has:

- OpenAI gateway routing for `messages`, `responses`, and `chat/completions`
- custom endpoint normalization for those surfaces
- a customized `account_test_service.go` with branch-specific behavior
- a frontend account test modal that already supports Gemini image testing and Sora-specific UX

The current branch does not yet have:

- OpenAI Images routes in [gateway.go](/root/sub2api-src/backend/internal/server/routes/gateway.go)
- image endpoint normalization in [endpoint.go](/root/sub2api-src/backend/internal/handler/endpoint.go)
- a handler or service for OpenAI Images requests
- the Responses/Codex compatibility bridge for OAuth image generation
- image-specific usage recording fields in the new request flow
- frontend OpenAI image model test support

## Integration Architecture

The integration uses a five-layer design:

1. **Route entry**
   add OpenAI Images endpoints and reject unsupported platforms with OpenAI-style errors

2. **Request normalization**
   classify inbound `images/generations` and `images/edits` distinctly from `responses`

3. **Image orchestration**
   parse image requests, choose a compatible account, apply failover, and dispatch to the correct upstream strategy

4. **Compatibility transformation**
   for OAuth accounts, transform Images API requests into Responses/Codex image-generation tool calls and transform the results back to Images API responses

5. **Usage and UX integration**
   record image usage and billing correctly, and expose image-account test support in the frontend

This architecture keeps the image feature self-contained while minimizing disruption to existing text and chat gateway flows.

## File Boundary

### Backend route and endpoint normalization

- Modify [gateway.go](/root/sub2api-src/backend/internal/server/routes/gateway.go)
  - add `/v1/images/generations`
  - add `/v1/images/edits`
  - add non-`/v1` aliases for both endpoints
  - reject non-OpenAI platforms with OpenAI-compatible `not_found_error`

- Modify [endpoint.go](/root/sub2api-src/backend/internal/handler/endpoint.go)
  - add canonical inbound endpoints for image generations and edits
  - ensure OpenAI upstream endpoint derivation preserves image endpoints instead of collapsing them into `/v1/responses`

### Backend handler layer

- Add [openai_images.go](/root/sub2api-src/backend/internal/handler/openai_images.go)
  - parse request bodies
  - integrate with existing auth, subscription, failover, and ops logging behavior
  - acquire user/account concurrency slots using the same protection model as the existing OpenAI gateway
  - call image-specific service methods rather than reusing text-response flows

### Backend service layer

- Add [openai_images.go](/root/sub2api-src/backend/internal/service/openai_images.go)
  - parse JSON and multipart image requests
  - normalize defaults
  - classify required account capability
  - select upstream path
  - stream or return non-stream image responses
  - enforce bounded image upload and download sizes

- Add [openai_images_responses.go](/root/sub2api-src/backend/internal/service/openai_images_responses.go)
  - convert Images API requests to Responses/Codex image-generation tool calls
  - attach uploaded images or URL inputs for edit requests
  - transform Responses lifecycle events and completed outputs back into Images API-compatible payloads

- Modify [openai_gateway_service.go](/root/sub2api-src/backend/internal/service/openai_gateway_service.go)
  - add image-related result fields such as `ImageCount`, `ImageSize`, and image output usage data
  - add `ForwardImages`
  - integrate image response handling into existing account schedule and record-usage infrastructure

- Modify [openai_codex_transform.go](/root/sub2api-src/backend/internal/service/openai_codex_transform.go)
  - absorb upstream fixes needed so image requests are converted safely into the current branch's Codex/Responses compatibility path

### Backend account, scheduler, and billing surfaces

- Modify [account.go](/root/sub2api-src/backend/internal/service/account.go)
  - add OpenAI image capability checks
  - preserve current `xlabapi` account behavior while explicitly advertising image compatibility

- Modify [openai_account_scheduler.go](/root/sub2api-src/backend/internal/service/openai_account_scheduler.go)
  - integrate image capability filtering and account selection for image requests

- Modify [pricing_service.go](/root/sub2api-src/backend/internal/service/pricing_service.go)
  - add fallback pricing behavior for OpenAI image models
  - preserve billing behavior when exact pricing entries are unavailable

- Modify usage-recording related code and tests
  - ensure image requests are recorded with `billing_mode=image`
  - ensure per-image billing overrides token-based billing when group image pricing is configured

### Backend account test service

- Modify [account_test_service.go](/root/sub2api-src/backend/internal/service/account_test_service.go)
  - add a dedicated OpenAI image test branch
  - preserve current `xlabapi` custom behavior for:
    - Sora cooldown handling
    - TLS fingerprint profile resolution
    - direct endpoint logic
    - existing Gemini and Claude account test behavior

- Add [account_test_service_openai_image_test.go](/root/sub2api-src/backend/internal/service/account_test_service_openai_image_test.go)
  - test OpenAI image account test payload generation and result parsing

### Frontend account test UI

- Modify [AccountTestModal.vue](/root/sub2api-src/frontend/src/components/account/AccountTestModal.vue)
- Modify [AccountTestModal.vue](/root/sub2api-src/frontend/src/components/admin/account/AccountTestModal.vue)
- Modify [useModelWhitelist.ts](/root/sub2api-src/frontend/src/composables/useModelWhitelist.ts)

The frontend adaptation must:

- add OpenAI image models to the selectable test set
- show prompt input when the selected model is an OpenAI image model
- display generated image previews
- support image lightbox behavior where already used in the branch
- preserve the current branch's Gemini image test behavior
- preserve the current branch's Sora-specific UI behavior

## Request Flow Design

### Entry and parsing

The new route accepts:

- `POST /v1/images/generations`
- `POST /v1/images/edits`
- the existing no-prefix alias style already used by other OpenAI routes

The parser supports:

- JSON bodies for generation and URL-based edit requests
- multipart bodies for file-upload edit requests

Defaults follow upstream image behavior where no explicit model is supplied.

### Capability classification

Each image request is classified into one of two capability levels:

- `basic`
  - prompt-only or defaulted request
  - no explicit size
  - no mask
  - single image output
  - no advanced native options

- `native`
  - explicit model or explicit size
  - mask input
  - multiple images
  - stream mode
  - advanced native image options such as `output_format`, `input_fidelity`, `partial_images`, `output_compression`, `moderation`, `style`, or equivalent edit-only requirements

This classification lets the scheduler avoid selecting an account path that cannot satisfy the request shape.

### Account selection

Account selection follows the existing OpenAI scheduling model but adds image capability filtering.

The scheduler must only select accounts that:

- belong to the target group
- remain otherwise eligible by current branch rules
- support the required OpenAI image capability

Failover behavior follows the existing OpenAI gateway conventions:

- same-account retry where already supported by pool mode logic
- cross-account retry where the upstream error qualifies for failover

### Upstream path split

The image feature uses a deliberate dual-path design.

#### API key accounts

API key accounts forward to native OpenAI image endpoints:

- `/v1/images/generations`
- `/v1/images/edits`

Before forwarding:

- the request model may be rewritten through the branch's existing model-mapping logic
- content type and multipart structure must be preserved

#### OAuth accounts

OAuth accounts use the Responses/Codex compatibility path:

- convert the Images API request into a Responses request with `tool_choice=image_generation`
- attach image inputs for edit requests
- send the transformed request through the Codex-compatible OpenAI OAuth path
- transform lifecycle and completed output events back into Images API-compatible output

This is the core `image2` compatibility mechanism.

### Result translation

Image results may come back in multiple forms:

- inline base64 image data
- download URLs
- file-service or asset pointers
- partial-image streaming events

The compatibility layer must normalize these into OpenAI Images API output structures while preserving useful metadata such as:

- `revised_prompt`
- `output_format`
- `quality`
- `size`
- `background`
- `model`

## Billing and Usage Design

The image path must record usage in a way that downstream reporting can distinguish from token-only requests.

Requirements:

- persist `image_count`
- persist `image_size`
- persist image billing mode as `image`
- preserve image-output token accounting where available
- prefer per-image group billing when configured
- otherwise fall back to model pricing logic for image models

This design intentionally keeps image accounting inside the existing usage pipeline instead of inventing a parallel persistence path.

## Error Handling

The new flow must produce coherent failures at each layer.

### Client and validation errors

Return `400 invalid_request_error` when:

- the body is empty or malformed
- the model is not an image model
- `edits` requests do not include image input
- fields have incompatible types or values

### Platform mismatch

Return OpenAI-style `404 not_found_error` when a non-OpenAI platform group hits the Images routes.

### No compatible account

Return `503` with the branch's existing failover-exhausted style when no compatible account can serve the request.

### Upstream retry and failover

For retryable upstream failures:

- preserve existing OpenAI failover rules
- allow same-account retry where pool mode already supports it
- switch accounts when the existing failover classifier says the error is retryable across accounts

### Resource safety

Image handling must remain bounded:

- limit multipart upload part reads
- limit image download reads
- avoid unbounded buffering during abnormal upstream responses

The upstream 20 MB bound is preserved as part of the design requirement because it directly prevents avoidable OOM behavior.

## Compatibility Constraints

The integration must not regress existing `xlabapi` custom behavior in these areas:

- current OpenAI text/chat/Responses gateway flow
- Sora-specific account test UX and cooldown behavior
- TLS fingerprint profile resolution for account tests
- direct endpoint handling for supported account types
- current Gemini image test flow in the frontend

The image feature is additive. Any implementation approach that weakens these existing surfaces is a design failure.

## Verification Design

Verification is intentionally targeted. The goal is to prove the integrated image feature works without claiming full upstream parity.

### Backend verification

Required backend verification areas:

- route registration and endpoint normalization tests
- image request parsing tests for JSON and multipart
- image capability classification tests
- API key native-image forwarding tests
- OAuth Responses/Codex compatibility tests
- image response translation tests for:
  - base64 outputs
  - download URLs
  - asset pointers
  - partial image events
- image billing and usage-recording tests
- account test service tests for OpenAI image models

### Frontend verification

Required frontend verification areas:

- account test modal recognizes OpenAI image models
- prompt input appears only when appropriate
- generated image previews render correctly
- current Gemini image behavior is not regressed
- current Sora-specific UX is not regressed

### Branch integrity verification

Because the working tree is already dirty, implementation verification must also ensure:

- only intended image-related files are changed by this integration work
- unrelated user changes are not reverted or folded into the migration

## Risks

### Risk: direct upstream code assumes newer surrounding infrastructure than `xlabapi` has

Mitigation:

- integrate manually rather than force a raw cherry-pick
- adapt the image feature to the current branch's actual OpenAI gateway structure

### Risk: `account_test_service.go` loses branch-specific Sora or TLS logic

Mitigation:

- treat `account_test_service.go` as a manual adaptation file
- preserve current branch behavior and add the OpenAI image path incrementally

### Risk: image billing records are persisted but charged incorrectly

Mitigation:

- add explicit tests for `billing_mode=image`
- add explicit tests for per-image billing precedence

### Risk: OAuth image compatibility works for simple generate requests but fails for edit requests or streamed responses

Mitigation:

- include tests for both generation and edit parsing
- verify result translation for both completed and streaming shapes

### Risk: large image payloads cause memory pressure or response-body abuse

Mitigation:

- preserve bounded reads for uploads and downloads
- verify the read limits remain in place after adaptation

## Success Criteria

This design is complete only when all of the following are true:

- `xlabapi` exposes OpenAI-compatible image generation and edit routes
- API key OpenAI accounts can use native image endpoints
- OAuth OpenAI accounts can use the Responses/Codex compatibility path for image generation
- image usage and billing are recorded correctly
- frontend account test modals support OpenAI image models and image previews
- current `xlabapi` Gemini and Sora test behavior remains intact
- targeted backend and frontend verification for the touched surfaces passes
- no unrelated local branch customizations are silently removed

## Next Phase

After this design is approved, the next phase is not immediate code editing in the current conversation state.

The next phase is:

1. write the implementation plan for the integration
2. execute the plan with targeted tests
3. review the resulting diff against both `xlabapi` behavior and the referenced upstream image feature behavior
