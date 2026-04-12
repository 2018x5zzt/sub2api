# Upgrade v0.1.106 Incremental Absorption Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Absorb the remaining effective changes from `upgrade-v0.1.106-merge` into `xlabapi` in five user-approved batches, while preserving the current `xlabapi` invite, redeem, and legacy invitation-code retirement semantics.

**Architecture:** Treat `xlabapi` as the semantic base and absorb upgrade work incrementally with scoped `git cherry-pick -n` operations plus manual hunk ports for mixed commits. Use clean batch commits and push after every completed batch. Keep migrations and generated-code reconciliation last, with explicit migration renumbering to fit the current `xlabapi` sequence and a final `git merge -s ours` marker only after all content has already been absorbed and verified.

**Tech Stack:** Git, Go, Wire, Ent, PostgreSQL migrations, Vue 3, TypeScript, Vitest

---

## File Structure

- `backend/docs/superpowers/specs/2026-04-12-upgrade-v0106-incremental-absorption-design.md`
  - Approved design spec. Use it as the scope guard for every batch.
- `backend/internal/pkg/apicompat/*.go`
  - Batch 1 compat conversion layer for Responses, Chat Completions, and Anthropic bridging.
- `backend/internal/service/gateway_*.go`
  - Batch 1 gateway endpoint plumbing and Batch 2 resilience/billing behavior.
- `backend/internal/service/openai_*.go`
  - Batch 1 OpenAI routing/model normalization and Batch 2 requested-model billing fixes.
- `backend/internal/handler/gateway_handler_*.go`
  - Batch 1 endpoint handlers for `/v1/responses` and `/v1/chat/completions`.
- `backend/internal/handler/openai_*.go`
  - Batch 1 OpenAI compatibility fixes and Batch 2 failover behavior.
- `backend/internal/server/routes/gateway.go`
  - Batch 1 route split registration.
- `backend/internal/service/antigravity_internal500_penalty.go`
  - Batch 2 extracted progressive internal-500 penalty logic.
- `backend/internal/repository/internal500_counter_cache.go`
  - Batch 2 counter cache backing the penalty logic.
- `backend/internal/service/usage_log_helpers.go`
  - Batch 2 requested-model billing/logging alignment.
- `backend/internal/service/pricing_service.go`
  - Batch 2 pricing-file refresh fix.
- `backend/migrations/083_add_tls_fingerprint_profiles.sql`
  - New Batch 3 migration number for the TLS profile schema. Do not reuse upstream `080_*`; `xlabapi` already owns `080`, `081`, and `082`.
- `backend/ent/schema/tls_fingerprint_profile.go`
  - Batch 3 Ent schema for TLS fingerprint profiles.
- `backend/internal/repository/tls_fingerprint_profile_*.go`
  - Batch 3 repository/cache layer for TLS profiles.
- `backend/internal/handler/admin/tls_fingerprint_profile_handler.go`
  - Batch 3 admin handler for TLS profile CRUD.
- `backend/internal/handler/admin/account_handler.go`
  - Batch 3 admin account filters, privacy actions, and bulk account actions.
- `backend/internal/service/admin_service.go`
  - Batch 3 admin account behavior.
- `frontend/src/components/account/BulkEditAccountModal.vue`
  - Batch 3 bulk passthrough/WS mode editing.
- `frontend/src/components/admin/account/AccountTableFilters.vue`
  - Batch 3 account privacy-mode filter UI.
- `frontend/src/components/keys/EndpointPopover.vue`
  - Batch 3 custom endpoint display in the keys UI.
- `frontend/src/views/user/ModelHubView.vue`
  - Batch 3 model-hub pricing presentation.
- `backend/internal/service/promo_service.go`
  - Batch 4 promo behavior ported from the mixed upstream merge commit.
- `backend/internal/service/redeem_service.go`
  - Batch 4 redeem behavior. Must keep current `xlabapi` invite-source semantics.
- `backend/internal/handler/redeem_handler.go`
  - Batch 4 benefit leaderboard and redeem behavior.
- `backend/internal/handler/admin/promo_handler.go`
  - Batch 4 admin promo endpoints.
- `backend/internal/handler/admin/redeem_handler.go`
  - Batch 4 admin redeem endpoints.
- `frontend/src/views/user/RedeemView.vue`
  - Batch 4 user redeem page.
- `frontend/src/views/admin/RedeemView.vue`
  - Batch 4 admin redeem page.
- `frontend/src/views/admin/PromoCodesView.vue`
  - Batch 4 admin promo-code page.
- `backend/migrations/084_add_usage_log_requested_model.sql`
  - New Batch 5 migration number for requested-model persistence.
- `backend/migrations/085_add_usage_log_requested_model_index_notx.sql`
  - New Batch 5 non-transactional index migration.
- `backend/ent/schema/usage_log.go`
  - Batch 5 requested-model schema uplift.
- `backend/ent/**`
  - Batch 5 regenerated Ent artifacts after all prior feature batches are in place.
- `backend/cmd/server/wire_gen.go`
  - Batch 3 and Batch 5 generated dependency-injection output.
- `backend/internal/server/api_contract_test.go`
  - Batch 3/5 contract guard for settings and DTO changes.

Important execution constraints for this plan:

- `xlabapi` wins on business semantics whenever a conflict touches invite, redeem, promo, or legacy invitation-code retirement.
- Do not directly merge `upgrade-v0.1.106-merge` before the final bookkeeping step.
- Do not merge `archive/main-before-runtime-redpacket-20260409` or `archive/origin-main-before-runtime-redpacket-20260409`.
- Do not cherry-pick `05edb551` (`feat(redeem): support subscription type in create-and-redeem API`); current `xlabapi` must keep subscription redeem behavior separate from commercial balance invite rewards.
- Do not cherry-pick `192efb84` (`feat(promo-code): complete promo code feature implementation`) as a whole; it collides with current invite/register/settings semantics. Batch 4 must manually port only the promo/redeem-compatible behavior from `b88ae98a`.
- `b88ae98a` is a mixed merge commit. Use it as a hunk source only; never cherry-pick it directly.

## Task 0: Preflight the Mainline and Create a Recovery Checkpoint

**Files:**
- Inspect: `backend/docs/superpowers/specs/2026-04-12-upgrade-v0106-incremental-absorption-design.md`
- Inspect: `backend/docs/superpowers/plans/2026-04-12-upgrade-v0106-incremental-absorption.md`
- Inspect: `backend/migrations/*`
- Inspect: `backend/internal/service/openai_gateway_service_test.go`
- Inspect: `backend/internal/service/gateway_multiplatform_test.go`

- [ ] **Step 1: Confirm `xlabapi` is the only working base and create a backup branch**

Run:

```bash
cd /root/sub2api-src
git checkout xlabapi
git pull --ff-only origin xlabapi
git status --short --branch
git branch backup/xlabapi-pre-upgrade-v0106-20260412
git branch --no-merged xlabapi
```

Expected:

- `git status --short --branch` shows `xlabapi` as the active branch.
- Only docs commits may be ahead of `origin/xlabapi`; there must be no unrelated dirty code changes in `backend/` or `frontend/`.
- `git branch --no-merged xlabapi` shows `upgrade-v0.1.106-merge` plus the two archive branches before absorption begins.

- [ ] **Step 2: Run the current baseline checks before touching upgrade commits**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/pkg/apicompat -count=1
go test -tags=unit ./internal/service -run 'TestApplyCodexOAuthTransform|TestOpenAIGatewayService_OAuthPassthrough|TestPromoServiceAllocateBenefitRandomBonus|TestRankBenefitUsages_' -count=1
go test -tags=unit ./internal/handler -run 'TestOpenAIResponses_|TestRedeemHandler_' -count=1
go test -tags=unit ./internal/handler/admin -run 'TestAccountHandler|TestRedeemHandlerEndpoints' -count=1
go test -tags=unit ./internal/server -run 'TestAPIContracts' -count=1

cd /root/sub2api-src/frontend
npm run build
```

Expected: PASS. Important detail: the service package must be tested with `-tags=unit`; without that tag, `openai_gateway_service_test.go` cannot see the unit-only test stub defined in `gateway_multiplatform_test.go`.

- [ ] **Step 3: Snapshot the exact source commits each batch will absorb**

Run:

```bash
cd /root/sub2api-src
git show --stat --name-only 8afa8c10 4feacf22 68f151f5 4321adab 31660c4c d927c0e4 | sed -n '1,240p'
git show --stat --name-only 4617ef2b 8c109411 ff5b467f f2c2abe6 70a9d0d3 c729ee42 8fcd819e 941c469a ab3e44e4 | sed -n '1,240p'
git show --stat --name-only 093a5a26 7cca69a1 3ee6f085 d563eb23 eb94342f 08108fdf ce8520c9 58755712 f5764d8d 50288e6b | sed -n '1,260p'
git show --stat --name-only 1854050d 4838ab74 c13c81f0 c489f238 47a54423 c2965c0f f6fd7c83 ccd42c1d e298a718 b6527523 | sed -n '1,280p'
git show --stat --name-only 73d72651 adbedd48 995bee14 1f39bf8a 2e6c02ac | sed -n '1,220p'
git show --stat --name-only b88ae98a fe60412a 0b845c25 efe8401e | sed -n '1,320p'
```

Expected: the inspected file lists match the task boundaries below. If a commit obviously drags unrelated invite/redeem semantics into the wrong batch, stop and re-scope before applying anything.

## Task 1: Absorb Batch 1A Compat Plumbing and New Gateway Endpoints

**Files:**
- Modify: `backend/internal/pkg/apicompat/anthropic_to_responses.go`
- Modify: `backend/internal/pkg/apicompat/anthropic_to_responses_response.go`
- Modify: `backend/internal/pkg/apicompat/responses_to_anthropic_request.go`
- Modify: `backend/internal/pkg/apicompat/chatcompletions_to_responses.go`
- Modify: `backend/internal/pkg/apicompat/anthropic_responses_test.go`
- Modify: `backend/internal/pkg/apicompat/chatcompletions_responses_test.go`
- Modify: `backend/internal/service/gateway_forward_as_chat_completions.go`
- Modify: `backend/internal/service/gateway_forward_as_responses.go`
- Modify: `backend/internal/handler/gateway_handler_chat_completions.go`
- Modify: `backend/internal/handler/gateway_handler_responses.go`
- Modify: `backend/internal/server/routes/gateway.go`

- [ ] **Step 1: Cherry-pick the Batch 1A compat/endpoint commits without committing yet**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 8afa8c10 4feacf22 68f151f5 4321adab 31660c4c d927c0e4
```

Expected: the worktree contains only Batch 1A compat and endpoint-plumbing changes.

- [ ] **Step 2: Resolve conflicts with `xlabapi` semantics**

Keep these rules while resolving:

- Preserve current `xlabapi` route wiring outside gateway-path registration.
- Only add the new `/v1/responses` and `/v1/chat/completions` plumbing; do not pull unrelated redeem/invite/admin routes.
- In `backend/internal/pkg/apicompat/*`, keep the upstream conversion logic from these commits unless it directly collides with current `xlabapi` custom model semantics.

- [ ] **Step 3: Run focused Batch 1A verification**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/pkg/apicompat -run 'TestAnthropicToResponses_|TestResponsesToAnthropic_|TestChatCompletionsToResponses_' -count=1
go test -tags=unit ./internal/service -run 'TestForwardAsChatCompletions_|TestForwardAsAnthropic_' -count=1
go test -tags=unit ./internal/handler -run 'TestOpenAIResponses_|TestOpenAIMissingResponsesDependencies|TestOpenAIEnsureResponsesDependencies' -count=1
```

Expected: PASS

- [ ] **Step 4: Commit and push the Batch 1A slice**

Run:

```bash
cd /root/sub2api-src
git add backend/internal/pkg/apicompat backend/internal/service/gateway_forward_as_chat_completions.go backend/internal/service/gateway_forward_as_responses.go backend/internal/handler/gateway_handler_chat_completions.go backend/internal/handler/gateway_handler_responses.go backend/internal/server/routes/gateway.go
git commit -m "merge: absorb upgrade batch 1a compat endpoints"
git push origin xlabapi
```

Expected: one reviewable commit containing only the compat and endpoint-plumbing slice.

## Task 2: Absorb Batch 1B Gateway, OAuth, and Message Compatibility Fixes

**Files:**
- Modify: `backend/internal/handler/openai_chat_completions.go`
- Modify: `backend/internal/handler/openai_gateway_handler.go`
- Modify: `backend/internal/handler/openai_gateway_handler_test.go`
- Modify: `backend/internal/pkg/oauth/oauth.go`
- Modify: `backend/internal/service/gateway_request.go`
- Modify: `backend/internal/service/gateway_request_test.go`
- Modify: `backend/internal/service/gateway_service.go`
- Modify: `backend/internal/service/gateway_prompt_test.go`
- Modify: `backend/internal/service/header_util.go`
- Modify: `backend/internal/service/openai_compat_model.go`
- Modify: `backend/internal/service/openai_compat_model_test.go`
- Modify: `backend/internal/service/openai_gateway_messages.go`
- Modify: `backend/internal/service/openai_model_mapping.go`
- Modify: `backend/internal/service/openai_model_mapping_test.go`

- [ ] **Step 1: Cherry-pick the Batch 1B compatibility commits**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 4617ef2b 8c109411 ff5b467f f2c2abe6 70a9d0d3 c729ee42 8fcd819e 941c469a ab3e44e4
```

Expected: model normalization, gateway-request cleanup, OAuth scope, PKCE, and Claude session-header compatibility are staged together without Batch 2 resiliency changes.

- [ ] **Step 2: Resolve conflicts in the gateway/OpenAI hotspot files**

Resolve carefully in:

- `backend/internal/service/gateway_service.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/service/openai_model_mapping.go`
- `backend/internal/service/openai_compat_model.go`

Keep these rules while resolving:

- Preserve current `xlabapi` invite/redeem behavior entirely; Batch 1 must remain gateway/OAuth-only.
- Keep `gpt-5.4-xhigh` normalization scoped to messages only, matching the `8c109411` -> `ff5b467f` -> `f2c2abe6` fix chain.
- Keep the current `xlabapi` Codex and passthrough flags unless the incoming change is explicitly about header or model compatibility.

- [ ] **Step 3: Run focused Batch 1B verification**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/service -run 'TestApplyCodexOAuthTransform|TestResolveOpenAIForwardModel|TestOpenAIGatewayService_OAuthPassthrough|TestOpenAIGatewayService_APIKeyPassthrough|TestOpenAIGatewayService_CodexCLIOnly|TestOpenAIGatewayService_Forward_WSv2' -count=1
go test -tags=unit ./internal/handler -run 'TestOpenAIResponses_|TestOpenAIHandler_|TestOpenAIResponsesWebSocket_' -count=1
go test -tags=unit ./internal/pkg/apicompat -run 'TestAnthropicToResponses_|TestChatCompletionsToResponses_' -count=1
```

Expected: PASS

- [ ] **Step 4: Commit and push the Batch 1B slice**

Run:

```bash
cd /root/sub2api-src
git add backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_gateway_handler_test.go backend/internal/pkg/oauth/oauth.go backend/internal/service/gateway_request.go backend/internal/service/gateway_request_test.go backend/internal/service/gateway_service.go backend/internal/service/gateway_prompt_test.go backend/internal/service/header_util.go backend/internal/service/openai_compat_model.go backend/internal/service/openai_compat_model_test.go backend/internal/service/openai_gateway_messages.go backend/internal/service/openai_model_mapping.go backend/internal/service/openai_model_mapping_test.go
git commit -m "merge: absorb upgrade batch 1b gateway oauth compatibility"
git push origin xlabapi
```

Expected: the full Batch 1 gateway/OAuth/messages scope is now on `origin/xlabapi`.

## Task 3: Absorb Batch 2 Resilience, Internal-500 Penalty, and Requested-Model Billing

**Files:**
- Modify: `backend/cmd/server/wire_gen.go`
- Modify: `backend/internal/repository/internal500_counter_cache.go`
- Modify: `backend/internal/repository/wire.go`
- Modify: `backend/internal/service/antigravity_gateway_service.go`
- Add: `backend/internal/service/antigravity_internal500_penalty.go`
- Add: `backend/internal/service/antigravity_internal500_penalty_test.go`
- Modify: `backend/internal/service/internal500_counter.go`
- Modify: `backend/internal/service/openai_gateway_service.go`
- Modify: `backend/internal/service/openai_gateway_service_test.go`
- Modify: `backend/internal/service/openai_gateway_record_usage_test.go`
- Modify: `backend/internal/service/openai_oauth_passthrough_test.go`
- Modify: `backend/internal/service/usage_log_helpers.go`
- Modify: `backend/internal/service/ratelimit_service.go`
- Modify: `backend/internal/service/pricing_service.go`
- Modify: `backend/internal/handler/openai_chat_completions.go`
- Modify: `backend/internal/handler/openai_gateway_handler.go`
- Modify: `backend/internal/handler/openai_gateway_handler_test.go`
- Add: `backend/internal/handler/openai_rate_limit_failover.go`

- [ ] **Step 1: Cherry-pick the internal-500 penalty chain**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 093a5a26 7cca69a1 3ee6f085 d563eb23 eb94342f
```

Expected: the worktree now contains the internal-500 counter cache, extracted penalty file, tests, and the timing fixups.

- [ ] **Step 2: Run the internal-500 penalty tests before layering more changes**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/service -run 'TestIsAntigravityInternalServerError|TestApplyInternal500Penalty|TestHandleInternal500RetryExhausted|TestResetInternal500Counter' -count=1
```

Expected: PASS

- [ ] **Step 3: Cherry-pick the 429 failover and requested-model billing commits**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 08108fdf ce8520c9 58755712 f5764d8d 50288e6b
```

Expected: the worktree now also contains 429 silent failover, persistent OpenAI 429 snapshots, abnormal-account marking for invalidated/revoked/deactivated OpenAI accounts, requested-model billing/logging, and the pricing-file refresh fix.

- [ ] **Step 4: Resolve the shared gateway-service conflicts**

Resolve carefully in:

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/handler/openai_gateway_handler.go`
- `backend/internal/service/usage_log_helpers.go`
- `backend/internal/service/ratelimit_service.go`

Keep these rules while resolving:

- Do not pull Batch 3 admin-account UI behavior into this batch.
- Keep billing based on the user-requested model, not the mapped upstream model.
- Keep 429 persistence and requested-model logging both active after the merge; they are not alternatives.

- [ ] **Step 5: Run focused Batch 2 verification**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/service -run 'TestIsAntigravityInternalServerError|TestApplyInternal500Penalty|TestHandleInternal500RetryExhausted|TestResetInternal500Counter|TestHandle429_OpenAIPersistsCodexSnapshotImmediately|TestOpenAIGatewayService_Forward_WSv2Handshake429PersistsRateLimit|TestOpenAIGatewayService_ProxyResponsesWebSocketFromClient_ErrorEventUsageLimitPersistsRateLimit|TestOpenAIGatewayServiceRecordUsage_UsesRequestedModelAndUpstreamModelMetadataFields|TestRateLimitService_HandleUpstreamError_OAuth401SetsTempUnschedulable' -count=1
go test -tags=unit ./internal/handler -run 'TestOpenAIResponses_|TestOpenAIEnsureForwardErrorResponse_|TestOpenAIRecoverResponsesPanic_' -count=1
go test -tags=unit ./internal/server -run 'TestAPIContracts' -count=1
```

Expected: PASS

- [ ] **Step 6: Commit and push Batch 2**

Run:

```bash
cd /root/sub2api-src
git add backend/cmd/server/wire_gen.go backend/internal/repository backend/internal/service backend/internal/handler/openai_chat_completions.go backend/internal/handler/openai_gateway_handler.go backend/internal/handler/openai_gateway_handler_test.go backend/internal/handler/openai_rate_limit_failover.go
git commit -m "merge: absorb upgrade batch 2 resilience billing"
git push origin xlabapi
```

Expected: Batch 2 lands as one backend-focused resiliency/billing commit.

## Task 4: Absorb Batch 3A TLS Fingerprint Profiles and Admin Account Privacy Actions

**Files:**
- Create: `backend/migrations/083_add_tls_fingerprint_profiles.sql`
- Add: `backend/ent/schema/tls_fingerprint_profile.go`
- Regenerate: `backend/ent/tlsfingerprintprofile*.go`
- Regenerate: `backend/ent/client.go`
- Regenerate: `backend/ent/ent.go`
- Regenerate: `backend/ent/hook/hook.go`
- Regenerate: `backend/ent/intercept/intercept.go`
- Regenerate: `backend/ent/migrate/schema.go`
- Regenerate: `backend/ent/mutation.go`
- Regenerate: `backend/ent/predicate/predicate.go`
- Regenerate: `backend/ent/runtime/runtime.go`
- Regenerate: `backend/ent/tx.go`
- Modify: `backend/internal/config/config.go`
- Modify: `backend/internal/repository/tls_fingerprint_profile_cache.go`
- Modify: `backend/internal/repository/tls_fingerprint_profile_repo.go`
- Modify: `backend/internal/repository/account_repo.go`
- Modify: `backend/internal/repository/account_repo_integration_test.go`
- Modify: `backend/internal/repository/wire.go`
- Modify: `backend/internal/handler/admin/account_data.go`
- Modify: `backend/internal/handler/admin/account_handler.go`
- Modify: `backend/internal/handler/admin/admin_service_stub_test.go`
- Add: `backend/internal/handler/admin/tls_fingerprint_profile_handler.go`
- Modify: `backend/internal/handler/dto/mappers.go`
- Modify: `backend/internal/handler/dto/types.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`
- Modify: `backend/internal/server/routes/admin.go`
- Modify: `backend/internal/server/api_contract_test.go`
- Modify: `backend/internal/service/admin_service.go`
- Modify: `backend/internal/service/account.go`
- Modify: `backend/internal/service/account_service.go`
- Modify: `backend/internal/service/account_service_delete_test.go`
- Modify: `backend/internal/service/account_test_service.go`
- Modify: `backend/internal/service/antigravity_oauth_service.go`
- Modify: `backend/internal/service/antigravity_privacy_service.go`
- Modify: `backend/internal/service/antigravity_privacy_service_test.go`
- Modify: `backend/internal/service/antigravity_subscription_service.go`
- Modify: `backend/internal/service/openai_privacy_service.go`
- Modify: `backend/internal/service/openai_privacy_retry_test.go`
- Modify: `backend/internal/service/token_refresh_service.go`
- Modify: `backend/internal/service/tls_fingerprint_profile_service.go`
- Modify: `backend/internal/service/wire.go`
- Modify: `backend/internal/pkg/antigravity/client.go`
- Modify: `backend/internal/pkg/antigravity/client_test.go`
- Modify: `backend/internal/pkg/tlsfingerprint/*.go`
- Modify: `frontend/src/api/admin/accounts.ts`
- Modify: `frontend/src/api/admin/index.ts`
- Add: `frontend/src/api/admin/tlsFingerprintProfile.ts`
- Modify: `frontend/src/components/account/CreateAccountModal.vue`
- Modify: `frontend/src/components/account/EditAccountModal.vue`
- Add: `frontend/src/components/admin/TLSFingerprintProfilesModal.vue`
- Add: `frontend/src/components/admin/account/AccountTableFilters.vue`
- Add: `frontend/src/components/admin/account/__tests__/AccountTableFilters.spec.ts`
- Modify: `frontend/src/components/common/PlatformTypeBadge.vue`
- Modify: `frontend/src/i18n/locales/en.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/views/admin/AccountsView.vue`

- [ ] **Step 1: Cherry-pick the TLS fingerprint commit and immediately renumber its migration**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 1854050d
mv backend/migrations/080_create_tls_fingerprint_profiles.sql backend/migrations/083_add_tls_fingerprint_profiles.sql
```

Expected: the TLS profile code is present, but the migration filename no longer collides with current `xlabapi` migration numbers.

- [ ] **Step 2: Cherry-pick the admin-account privacy/action commits**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 4838ab74 c13c81f0 c489f238 47a54423 c2965c0f f6fd7c83 ccd42c1d e298a718 b6527523
```

Expected: account privacy filters, privacy retry, custom forward URL, and admin retry plumbing are present on top of the TLS work.

- [ ] **Step 3: Resolve the migration, Ent, and account-admin conflicts**

Resolve carefully in:

- `backend/migrations/083_add_tls_fingerprint_profiles.sql`
- `backend/ent/migrate/schema.go`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/service/admin_service.go`
- `frontend/src/views/admin/AccountsView.vue`

Keep these rules while resolving:

- Preserve all current `xlabapi` invite/redeem semantics; this batch must remain account/TLS/admin-only.
- Keep the renumbered migration at `083_add_tls_fingerprint_profiles.sql`.
- Keep `frontend/src/views/admin/AccountsView.vue` aligned with current `xlabapi` styling while adding the new TLS/privacy controls.

- [ ] **Step 4: Run focused Batch 3A verification**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/pkg/tlsfingerprint -count=1
go test -tags=unit ./internal/handler/admin -run 'TestAccountHandlerGetAvailableModels_|TestAccountHandler_Create_AnthropicAPIKeyPassthroughExtraForwarded' -count=1
go test -tags=unit ./internal/service -run 'TestAccountTestService_OpenAISuccessPersistsSnapshotFromHeaders|TestAccountTestService_OpenAI429PersistsSnapshotAndRateLimit|TestAdminService_ListAccounts_WithSearch' -count=1
go test -tags=unit ./internal/server -run 'TestAPIContracts' -count=1

cd /root/sub2api-src/frontend
npm run test:run -- src/components/admin/account/__tests__/AccountTableFilters.spec.ts
npm run build
```

Expected: PASS

- [ ] **Step 5: Commit and push Batch 3A**

Run:

```bash
cd /root/sub2api-src
git add backend/migrations/083_add_tls_fingerprint_profiles.sql backend/ent backend/internal/config/config.go backend/internal/repository backend/internal/handler backend/internal/server backend/internal/service backend/internal/pkg/antigravity backend/internal/pkg/tlsfingerprint frontend/src/api/admin frontend/src/components/account frontend/src/components/admin frontend/src/components/common/PlatformTypeBadge.vue frontend/src/i18n frontend/src/types frontend/src/views/admin/AccountsView.vue
git commit -m "merge: absorb upgrade batch 3a tls admin privacy"
git push origin xlabapi
```

Expected: TLS profiles and admin privacy actions land together as one deployable batch slice.

## Task 5: Absorb Batch 3B Bulk Account Ops, Keys Endpoint UI, and Model-Hub Pricing

**Files:**
- Modify: `backend/internal/handler/admin/account_handler.go`
- Modify: `backend/internal/repository/api_key_repo.go`
- Modify: `backend/internal/handler/api_key_handler.go`
- Modify: `backend/internal/handler/api_key_handler_models_test.go`
- Modify: `backend/internal/handler/dto/settings.go`
- Modify: `backend/internal/handler/dto/types.go`
- Modify: `backend/internal/handler/admin/setting_handler.go`
- Modify: `backend/internal/handler/setting_handler.go`
- Modify: `backend/internal/service/setting_service.go`
- Modify: `backend/internal/service/settings_view.go`
- Modify: `backend/internal/service/domain_constants.go`
- Modify: `backend/internal/config/config.go`
- Modify: `deploy/config.example.yaml`
- Modify: `backend/internal/server/api_contract_test.go`
- Modify: `frontend/src/components/account/BulkEditAccountModal.vue`
- Modify: `frontend/src/components/account/__tests__/BulkEditAccountModal.spec.ts`
- Add: `frontend/src/components/keys/EndpointPopover.vue`
- Add: `frontend/src/components/keys/__tests__/EndpointPopover.spec.ts`
- Modify: `frontend/src/api/admin/settings.ts`
- Modify: `frontend/src/stores/app.ts`
- Modify: `frontend/src/views/admin/SettingsView.vue`
- Modify: `frontend/src/views/user/KeysView.vue`
- Modify: `frontend/src/views/user/ModelHubView.vue`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`

- [ ] **Step 1: Cherry-pick the bulk-account and keys/model-hub commits**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 73d72651 adbedd48 995bee14 1f39bf8a 2e6c02ac
```

Expected: the worktree contains bulk OpenAI passthrough / WS mode editing, custom endpoint display, API key uniqueness fix, and model-hub pricing presentation.

- [ ] **Step 2: Resolve the shared settings and types conflicts**

Resolve carefully in:

- `backend/internal/service/setting_service.go`
- `backend/internal/service/settings_view.go`
- `backend/internal/handler/dto/settings.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/server/api_contract_test.go`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`

Keep these rules while resolving:

- Preserve the current `xlabapi` removal of legacy one-time invitation-code registration.
- Only add the new keys endpoint and model-hub pricing fields that are compatible with current settings contracts.
- Keep `frontend/src/views/user/ModelHubView.vue` aligned with current `xlabapi` model naming and invite copy.

- [ ] **Step 3: Run focused Batch 3B verification**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/handler/admin -run 'TestAccountHandlerCheckMixedChannel|TestAccountHandlerBulkUpdateMixedChannel' -count=1
go test -tags=unit ./internal/handler -run 'TestFetchUpstreamModels_' -count=1
go test -tags=unit ./internal/service -run 'TestAdminService_BulkUpdateAccounts_|TestAdminService_ListAccounts_WithSearch' -count=1
go test -tags=unit ./internal/server -run 'TestAPIContracts' -count=1

cd /root/sub2api-src/frontend
npm run test:run -- src/components/account/__tests__/BulkEditAccountModal.spec.ts src/components/admin/account/__tests__/AccountTableFilters.spec.ts src/components/keys/__tests__/EndpointPopover.spec.ts
npm run build
```

Expected: PASS

- [ ] **Step 4: Commit and push Batch 3B**

Run:

```bash
cd /root/sub2api-src
git add backend/internal/config/config.go backend/internal/repository/api_key_repo.go backend/internal/handler backend/internal/server/api_contract_test.go backend/internal/service deploy/config.example.yaml frontend/src/components/account frontend/src/components/keys frontend/src/api/admin/settings.ts frontend/src/stores/app.ts frontend/src/views/admin/SettingsView.vue frontend/src/views/user/KeysView.vue frontend/src/views/user/ModelHubView.vue frontend/src/types/index.ts frontend/src/i18n
git commit -m "merge: absorb upgrade batch 3b bulk ops keys modelhub"
git push origin xlabapi
```

Expected: Batch 3 is now complete on `origin/xlabapi`.

## Task 6: Absorb Batch 4 Promo/Redeem/Benefit Behavior Without Regressing Current `xlabapi` Semantics

**Files:**
- Modify: `backend/internal/repository/promo_code_repo.go`
- Modify: `backend/internal/repository/redeem_code_repo.go`
- Modify: `backend/internal/service/promo_code.go`
- Modify: `backend/internal/service/promo_code_repository.go`
- Modify: `backend/internal/service/promo_service.go`
- Modify: `backend/internal/service/redeem_code.go`
- Modify: `backend/internal/service/redeem_service.go`
- Modify: `backend/internal/handler/redeem_handler.go`
- Modify: `backend/internal/handler/redeem_handler_test.go`
- Modify: `backend/internal/handler/admin/promo_handler.go`
- Modify: `backend/internal/handler/admin/redeem_handler.go`
- Modify: `backend/internal/handler/admin/redeem_handler_test.go`
- Modify: `backend/internal/handler/dto/types.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `frontend/src/api/redeem.ts`
- Modify: `frontend/src/api/admin/promo.ts`
- Modify: `frontend/src/api/admin/redeem.ts`
- Modify: `frontend/src/views/user/RedeemView.vue`
- Modify: `frontend/src/views/admin/RedeemView.vue`
- Modify: `frontend/src/views/admin/PromoCodesView.vue`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`

- [ ] **Step 1: Inspect the mixed upstream merge commit as a path-scoped hunk source**

Run:

```bash
cd /root/sub2api-src
git show --stat --name-only b88ae98a -- backend/internal/repository/promo_code_repo.go backend/internal/repository/redeem_code_repo.go backend/internal/service/promo_code.go backend/internal/service/promo_code_repository.go backend/internal/service/promo_service.go backend/internal/service/redeem_code.go backend/internal/service/redeem_service.go backend/internal/handler/redeem_handler.go backend/internal/handler/redeem_handler_test.go backend/internal/handler/admin/promo_handler.go backend/internal/handler/admin/redeem_handler.go backend/internal/handler/admin/redeem_handler_test.go backend/internal/handler/dto/types.go backend/internal/server/routes/user.go frontend/src/api/redeem.ts frontend/src/api/admin/promo.ts frontend/src/api/admin/redeem.ts frontend/src/views/user/RedeemView.vue frontend/src/views/admin/RedeemView.vue frontend/src/views/admin/PromoCodesView.vue frontend/src/types/index.ts frontend/src/i18n/locales/en.ts frontend/src/i18n/locales/zh.ts | sed -n '1,320p'
```

Expected: only the Batch 4 files above are considered. Do not pull gateway/model-hub/account hunks from the same merge commit.

- [ ] **Step 2: Restore isolated whole-file promo/admin files directly from `b88ae98a`**

Run:

```bash
cd /root/sub2api-src
git restore --source=b88ae98a --worktree --staged -- backend/internal/repository/promo_code_repo.go backend/internal/service/promo_code.go backend/internal/service/promo_code_repository.go backend/internal/service/promo_service.go backend/internal/handler/admin/promo_handler.go frontend/src/api/admin/promo.ts frontend/src/views/admin/PromoCodesView.vue
```

Expected: isolated promo/admin files are copied in full without dragging shared gateway or settings files.

- [ ] **Step 3: Manually port only the safe hunks into shared redeem and route files**

Manually port the benefit/redeem hunks from `b88ae98a` into:

- `backend/internal/repository/redeem_code_repo.go`
- `backend/internal/service/redeem_code.go`
- `backend/internal/service/redeem_service.go`
- `backend/internal/handler/redeem_handler.go`
- `backend/internal/handler/redeem_handler_test.go`
- `backend/internal/handler/admin/redeem_handler.go`
- `backend/internal/handler/admin/redeem_handler_test.go`
- `backend/internal/handler/dto/types.go`
- `backend/internal/server/routes/user.go`
- `frontend/src/api/redeem.ts`
- `frontend/src/api/admin/redeem.ts`
- `frontend/src/views/user/RedeemView.vue`
- `frontend/src/views/admin/RedeemView.vue`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/en.ts`
- `frontend/src/i18n/locales/zh.ts`

Keep these rules while resolving:

- Do not reintroduce legacy registration invitation-code logic anywhere in auth, settings, or register flows.
- Do not let subscription redeem flows count as commercial balance recharges for invite base rewards.
- Do not add new migration files in Batch 4; current `xlabapi` already owns the promo/redeem migration numbering.

- [ ] **Step 4: Run focused Batch 4 verification**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/service -run 'TestPromoServiceAllocateBenefitRandomBonus_DeterministicMinimumDraw|TestRankBenefitUsages_|TestRedeemService_InvalidateRedeemCaches_AuthCache' -count=1
go test -tags=unit ./internal/handler -run 'TestRedeemHandler_(RedeemCodeStillWorks|PromoFallbackRedeemsOncePerUser|PromoDisabledReturnsRedeemNotFound|BenefitFallbackWorksWhenPromoSettingDisabled|BenefitRedPacketReturnsBreakdownAndLeaderboardFlag|BenefitRedPacketRequiresUsername|GetBenefitLeaderboardAfterRedeem|GetBenefitLeaderboardDefaultsToTopTwentyEntries)$' -count=1
go test -tags=unit ./internal/handler/admin -run 'TestRedeemHandlerEndpoints' -count=1

cd /root/sub2api-src/frontend
npm run build
```

Expected: PASS

- [ ] **Step 5: Commit and push Batch 4**

Run:

```bash
cd /root/sub2api-src
git add backend/internal/repository/promo_code_repo.go backend/internal/repository/redeem_code_repo.go backend/internal/service/promo_code.go backend/internal/service/promo_code_repository.go backend/internal/service/promo_service.go backend/internal/service/redeem_code.go backend/internal/service/redeem_service.go backend/internal/handler/redeem_handler.go backend/internal/handler/redeem_handler_test.go backend/internal/handler/admin/promo_handler.go backend/internal/handler/admin/redeem_handler.go backend/internal/handler/admin/redeem_handler_test.go backend/internal/handler/dto/types.go backend/internal/server/routes/user.go frontend/src/api/redeem.ts frontend/src/api/admin/promo.ts frontend/src/api/admin/redeem.ts frontend/src/views/user/RedeemView.vue frontend/src/views/admin/RedeemView.vue frontend/src/views/admin/PromoCodesView.vue frontend/src/types/index.ts frontend/src/i18n
git commit -m "merge: absorb upgrade batch 4 redeem promo compatibility"
git push origin xlabapi
```

Expected: Batch 4 is absorbed without changing current `xlabapi` invite/redeem rules.

## Task 7: Absorb Batch 5 Requested-Model Migrations, Regenerate Ent/Wire, and Reconcile Contracts

**Files:**
- Create: `backend/migrations/084_add_usage_log_requested_model.sql`
- Create: `backend/migrations/085_add_usage_log_requested_model_index_notx.sql`
- Modify: `backend/ent/schema/usage_log.go`
- Regenerate: `backend/ent/**`
- Regenerate: `backend/cmd/server/wire_gen.go`
- Modify: `backend/internal/server/api_contract_test.go`
- Modify: `backend/internal/repository/migrations_schema_integration_test.go`

- [ ] **Step 1: Materialize the requested-model migrations under the `xlabapi` numbering**

Run:

```bash
cd /root/sub2api-src
git show fe60412a:backend/migrations/077_add_usage_log_requested_model.sql > backend/migrations/084_add_usage_log_requested_model.sql
git show fe60412a:backend/migrations/078_add_usage_log_requested_model_index_notx.sql > backend/migrations/085_add_usage_log_requested_model_index_notx.sql
```

Expected: the requested-model migrations exist under `084` and `085`, avoiding the `077`/`078` collisions that already exist on `xlabapi`.

- [ ] **Step 2: Cherry-pick the usage-log schema uplift and generated Ent baseline**

Run:

```bash
cd /root/sub2api-src
git cherry-pick -n 0b845c25 efe8401e
```

Expected: `backend/ent/schema/usage_log.go` and the upstream-generated Ent files are present in the worktree as a temporary baseline before local regeneration.

- [ ] **Step 3: Regenerate Ent and Wire from the combined `xlabapi` state**

Run:

```bash
cd /root/sub2api-src/backend
go generate ./ent
go generate ./cmd/server
```

Expected: local generation rewrites the Ent and Wire output to match the already-absorbed Batch 1-4 code plus the new requested-model schema.

- [ ] **Step 4: Resolve the remaining generated and contract drift**

Resolve carefully in:

- `backend/ent/migrate/schema.go`
- `backend/ent/runtime/runtime.go`
- `backend/cmd/server/wire_gen.go`
- `backend/internal/server/api_contract_test.go`

Keep these rules while resolving:

- Do not re-import upstream migration filenames that collide with existing `xlabapi` numbering.
- Keep the current `xlabapi` invite/settings contract fields intact unless a schema change in this task requires an additive update.
- If `b88ae98a` still contains unabsorbed promo-schema changes after Batch 4, port only the minimum schema-compatible hunks; do not add a second promo migration.

- [ ] **Step 5: Run the full post-absorption verification suite**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/pkg/apicompat ./internal/service ./internal/handler ./internal/handler/admin ./internal/server -count=1
go test -tags=integration ./internal/repository -run 'Test(AccountRepoSuite|APIKeyRepoSuite|InviteAdminRepoSuite|InviteGrowthRepoSuite|InviteRolloutVerificationSQL_|MigrationsRunner_IsIdempotent_AndSchemaIsUpToDate|MigrationsSchema_DynamicBudgetColumnsExist)' -count=1
go test ./cmd/server -run 'TestProvideServiceBuildInfo|TestProvideCleanup_WithMinimalDependencies_NoPanic' -count=1

cd /root/sub2api-src/frontend
npm run build
```

Expected: PASS

- [ ] **Step 6: Commit and push Batch 5**

Run:

```bash
cd /root/sub2api-src
git add backend/migrations/084_add_usage_log_requested_model.sql backend/migrations/085_add_usage_log_requested_model_index_notx.sql backend/ent backend/cmd/server/wire_gen.go backend/internal/server/api_contract_test.go backend/internal/repository/migrations_schema_integration_test.go
git commit -m "merge: absorb upgrade batch 5 generated cleanup"
git push origin xlabapi
```

Expected: the combined branch is now coherent, generated, and verified.

## Task 8: Mark `upgrade-v0.1.106-merge` as Absorbed and Verify Branch Inventory

**Files:**
- Inspect: entire repo diff between `xlabapi` and `upgrade-v0.1.106-merge`

- [ ] **Step 1: Confirm there are no unexpected code deltas left before the bookkeeping merge**

Run:

```bash
cd /root/sub2api-src
git diff --name-only xlabapi..upgrade-v0.1.106-merge
```

Expected: either empty output or only intentionally skipped leftovers that were explicitly rejected by the current `xlabapi` semantics. If unexpected code files remain, stop here and absorb them before continuing.

- [ ] **Step 2: Create a no-content ancestry merge so Git records the branch as absorbed**

Run:

```bash
cd /root/sub2api-src
git merge -s ours --no-ff upgrade-v0.1.106-merge -m "merge: mark upgrade-v0.1.106-merge absorbed"
git push origin xlabapi
```

Expected: Git ancestry now records the source branch as merged without changing the already-verified code content.

- [ ] **Step 3: Verify that only archive branches remain unmerged**

Run:

```bash
cd /root/sub2api-src
git branch --no-merged xlabapi
```

Expected:

```text
  archive/main-before-runtime-redpacket-20260409
  archive/origin-main-before-runtime-redpacket-20260409
```
