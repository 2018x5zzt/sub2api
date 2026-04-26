# Upstream Affiliate And Available Channels Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace xlabapi's local Model Hub and Invite runtime surfaces with upstream's Available Channels and Affiliate implementations while preserving production invite data through a compatibility migration.

**Architecture:** Use upstream's `AvailableChannelHandler`, `ChannelService.ListAvailable`, frontend `/available-channels`, and feature flag wiring as the model browsing surface. Use upstream's `AffiliateService`, `AffiliateRepository`, user/admin affiliate handlers, frontend `/affiliate`, and `aff_code` registration path as the invite rebate surface. Add xlabapi-only SQL that migrates old `users.invite_code`, `users.invited_by_user_id`, and `invite_reward_records` into upstream `user_affiliates` and `user_affiliate_ledger`, preserves the old 3% business default, and keeps old `?invite=` registration links working.

**Tech Stack:** Go 1.26, Ent, PostgreSQL forward-only SQL migrations, Gin handlers, Wire DI, Vue 3, Pinia, Vue Router, Vitest.

---

### Task 1: Compatibility Migration

**Files:**
- Create: `backend/migrations/134_xlabapi_invite_to_affiliate_compat.sql`
- Test: `backend/migrations/affiliate_compat_migration_test.go`

- [ ] Write a migration test that applies upstream affiliate tables plus the compatibility migration against an in-memory SQL fixture.
- [ ] Verify RED: the test must fail before the migration exists or before it copies old invite data.
- [ ] Implement `134_xlabapi_invite_to_affiliate_compat.sql`:
  - Create missing upstream affiliate tables if needed.
  - Seed `user_affiliates` from existing users.
  - Copy `users.invite_code` into `aff_code` where possible.
  - Set `inviter_id` from `users.invited_by_user_id`.
  - Convert old applied invite rewards into `user_affiliate_ledger` rows.
  - Enable `affiliate_enabled`.
  - Set `affiliate_rebate_rate` to `3`.
- [ ] Verify GREEN with the migration test.

### Task 2: Backend Affiliate Runtime

**Files:**
- Create/port: `backend/internal/service/affiliate_service.go`
- Create/port: `backend/internal/repository/affiliate_repo.go`
- Create/port: `backend/internal/handler/admin/affiliate_handler.go`
- Modify: `backend/internal/service/auth_service.go`
- Modify: `backend/internal/service/redeem_service.go`
- Modify: `backend/internal/handler/auth_handler.go`
- Modify: `backend/internal/handler/user_handler.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/server/routes/admin.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`
- Modify: `backend/internal/repository/wire.go`
- Modify: `backend/internal/service/wire.go`
- Regenerate: `backend/cmd/server/wire_gen.go`

- [ ] Add/port upstream affiliate unit tests for code validation, binding, quota transfer, and admin custom settings.
- [ ] Verify RED against current xlabapi.
- [ ] Port upstream affiliate service/repository/handlers.
- [ ] Wire `aff_code` through email registration, OAuth registration, and user/admin APIs.
- [ ] Keep xlabapi's redeem-code recharge path as an affiliate accrual trigger because full upstream payment fulfillment is not present locally.
- [ ] Remove local Invite route/handler/service use from runtime wiring while leaving historical tables untouched.
- [ ] Verify with focused Go tests.

### Task 3: Backend Available Channels Runtime

**Files:**
- Create/port: `backend/internal/service/channel_available.go`
- Create/port: `backend/internal/handler/available_channel_handler.go`
- Modify: `backend/internal/service/channel.go`
- Modify: `backend/internal/service/channel_service.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`
- Modify: `backend/internal/service/domain_constants.go`
- Modify: `backend/internal/service/setting_service.go`
- Modify: `backend/internal/service/settings_view.go`
- Modify: `backend/internal/handler/dto/settings.go`

- [ ] Add/port tests that prove the user endpoint filters channels by current user's groups and strips internal fields.
- [ ] Verify RED against current xlabapi.
- [ ] Port upstream available-channel service and handler.
- [ ] Add `available_channels_enabled` public/admin setting and default it enabled for xlabapi migration if required by rollout.
- [ ] Verify with focused Go tests.

### Task 4: Frontend Replacement

**Files:**
- Create/port: `frontend/src/api/channels.ts`
- Create/port: `frontend/src/components/channels/AvailableChannelsTable.vue`
- Create/port: `frontend/src/components/channels/SupportedModelChip.vue`
- Create/port: `frontend/src/views/user/AvailableChannelsView.vue`
- Create/port: `frontend/src/api/admin/affiliates.ts`
- Create/port: `frontend/src/views/user/AffiliateView.vue`
- Create/port: `frontend/src/utils/oauthAffiliate.ts`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/components/layout/AppSidebar.vue`
- Modify: `frontend/src/views/auth/RegisterView.vue`
- Modify: `frontend/src/api/auth.ts`
- Modify: `frontend/src/api/user.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Remove runtime use: `frontend/src/views/user/ModelHubView.vue`
- Remove runtime use: `frontend/src/views/user/InviteView.vue`
- Remove runtime use: `frontend/src/views/admin/InvitesView.vue`

- [ ] Add/port tests for `oauthAffiliate` and route/query compatibility.
- [ ] Verify RED against current xlabapi.
- [ ] Port upstream Available Channels and Affiliate pages.
- [ ] Redirect `/models` to `/available-channels`.
- [ ] Accept both `?aff=` and legacy `?invite=` in registration, with new links emitted as `?aff=`.
- [ ] Remove sidebar entries for local Model Hub and Invite Ops.
- [ ] Verify with Vitest focused suites and typecheck if dependency state allows.

### Task 5: Final Verification And Commit

**Files:**
- All changed files.

- [ ] Run `go test -count=1 ./internal/service ./internal/handler ./internal/server`.
- [ ] Run focused frontend tests for affiliate/available-channels/register compatibility.
- [ ] Run `git diff --check`.
- [ ] Review `git status --short` and commit the implementation.
