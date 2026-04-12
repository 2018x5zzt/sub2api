# Invite System Consolidation and Legacy Code Retirement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Formalize the invite system already present in the worktree into a clean, mergeable feature set, retire the legacy one-time invitation-code mode, lock invite-link registration UX, and make base invite reward settlement atomic.

**Architecture:** Treat the current invite implementation as the source of truth and promote it into tracked backend and frontend code, rather than rewriting it from scratch. Then remove legacy invitation-code settings and validation semantics so permanent user invite codes are the only active invite identity, while wrapping base reward ledger writes and balance updates in the same Ent transaction model already used by admin invite write flows.

**Tech Stack:** Go, Ent, Wire, Gin, PostgreSQL, Vitest, Vue 3, TypeScript

---

## File Structure

- `backend/migrations/081_add_invite_growth_foundation.sql`
  - Adds `users.invite_code`, `users.invited_by_user_id`, `users.invite_bound_at`, `redeem_codes.source_type`, and the `invite_reward_records` table.
- `backend/migrations/082_add_invite_admin_ops.sql`
  - Adds `invite_relationship_events`, `invite_admin_actions`, backfills `register_bind` rows, and links admin actions to invite reward records.
- `backend/ent/schema/invite_admin_action.go`
  - Ent schema for append-only admin invite operation audit rows.
- `backend/ent/schema/invite_relationship_event.go`
  - Ent schema for inviter binding history and effective-time lookups.
- `backend/ent/schema/invite_reward_record.go`
  - Ent schema for invite reward ledger rows.
- `backend/ent/inviteadminaction*.go`
  - Generated Ent model/query/create/update files for invite admin actions.
- `backend/ent/inviterelationshipevent*.go`
  - Generated Ent model/query/create/update files for invite relationship events.
- `backend/ent/inviterewardrecord*.go`
  - Generated Ent model/query/create/update files for invite reward records.
- `backend/ent/client.go`
  - Generated Ent client entrypoint that must include the three invite entities.
- `backend/ent/migrate/schema.go`
  - Generated migration schema snapshot that must include invite tables and columns.
- `backend/internal/repository/invite_reward_record_repo.go`
  - Invite reward ledger repository used by both user invite read APIs and admin recompute/manual-grant flows.
- `backend/internal/repository/invite_relationship_event_repo.go`
  - Invite relationship history repository, including "effective inviter at time" lookup.
- `backend/internal/repository/invite_admin_action_repo.go`
  - Repository for creating admin invite action audit rows.
- `backend/internal/repository/invite_admin_query_repo.go`
  - Admin-facing read models for stats, relationships, rewards, and actions.
- `backend/internal/repository/invite_growth_integration_test.go`
  - Integration coverage for invite growth foundation persistence.
- `backend/internal/repository/invite_admin_integration_test.go`
  - Integration coverage for admin invite persistence and migration backfill semantics.
- `backend/internal/repository/invite_admin_service_integration_test.go`
  - Integration coverage for admin invite service rollback semantics.
- `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`
  - Existing rollout verification SQL contract coverage that must continue to pass after consolidation.
- `backend/internal/repository/wire.go`
  - Repository provider set; must expose all invite repositories so clean `wire` generation works.
- `backend/internal/service/invite.go`
  - Invite domain constants and service-facing structs.
- `backend/internal/service/invite_service.go`
  - User invite summary/list logic plus base reward settlement path.
- `backend/internal/service/admin_invite.go`
  - Admin invite service contracts and recompute scope/delta models.
- `backend/internal/service/admin_service_invite.go`
  - Admin invite write/read orchestration using transaction-bound repository calls.
- `backend/internal/service/invite_service_test.go`
  - Unit tests for invite code generation, summary math, and base reward settlement behavior.
- `backend/internal/service/auth_service.go`
  - Registration and OAuth registration semantics that currently still carry legacy invitation-code behavior.
- `backend/internal/service/auth_service_register_test.go`
  - Unit tests for registration and OAuth invite handling semantics.
- `backend/internal/service/settings_view.go`
  - API-facing settings view models that still expose `invitation_code_enabled`.
- `backend/internal/service/setting_service.go`
  - Public/admin settings assembly and persistence logic that must stop reading/writing the retired invitation-code toggle.
- `backend/internal/service/domain_constants.go`
  - Setting-key constants; the retired `invitation_code_enabled` contract should not remain active.
- `backend/internal/service/wire.go`
  - Service provider set; `InviteService` must receive the Ent client for transactional settlement.
- `backend/internal/handler/dto/invite.go`
  - User invite API DTOs.
- `backend/internal/handler/dto/admin_invite.go`
  - Admin invite API DTOs and filter parsing helpers.
- `backend/internal/handler/dto/settings.go`
  - Admin/public settings DTOs that currently still include `invitation_code_enabled`.
- `backend/internal/handler/invite_handler.go`
  - User invite endpoints.
- `backend/internal/handler/admin/invite_handler.go`
  - Admin invite endpoints.
- `backend/internal/handler/auth_handler.go`
  - Public invite-code validation endpoint and registration request handling.
- `backend/internal/handler/auth_handler_invite_test.go`
  - Handler tests for public invite-code validation responses.
- `backend/internal/handler/admin/setting_handler.go`
  - Admin settings request/response mapping that must stop exposing the retired toggle.
- `backend/internal/handler/handler.go`
  - Top-level handler struct that must include the invite handlers in tracked code.
- `backend/internal/handler/wire.go`
  - Handler provider set that must include user/admin invite handlers in tracked code.
- `backend/internal/server/routes/user.go`
  - User API routing for `/invite/summary` and `/invite/rewards`.
- `backend/internal/server/routes/admin.go`
  - Admin API routing for `/admin/invites/*`.
- `backend/internal/server/api_contract_test.go`
  - Settings contract fixture that must drop `invitation_code_enabled`.
- `backend/cmd/server/wire_gen.go`
  - Generated Wire injector output; must be regenerated after provider-set changes.
- `frontend/src/api/invite.ts`
  - User invite API client.
- `frontend/src/api/admin/invites.ts`
  - Admin invite API client.
- `frontend/src/api/auth.ts`
  - Public invite-code validation client used by registration.
- `frontend/src/api/index.ts`
  - Frontend API barrel export that must expose `inviteAPI`.
- `frontend/src/api/admin/index.ts`
  - Admin API barrel export that must expose `adminInvitesAPI`.
- `frontend/src/api/admin/settings.ts`
  - Admin settings API types that still expose `invitation_code_enabled`.
- `frontend/src/types/index.ts`
  - Shared frontend DTOs for invite APIs and public/admin settings.
- `frontend/src/stores/app.ts`
  - Public settings fallback object that still includes `invitation_code_enabled`.
- `frontend/src/utils/inviteQuery.ts`
  - Query-string normalization helper for invite links.
- `frontend/src/utils/__tests__/inviteQuery.spec.ts`
  - Unit tests for invite query normalization.
- `frontend/src/views/user/InviteView.vue`
  - User invite center UI.
- `frontend/src/views/user/__tests__/InviteView.spec.ts`
  - Invite-center copy regression tests in zh and en.
- `frontend/src/views/admin/InvitesView.vue`
  - Admin invite operations page.
- `frontend/src/views/admin/__tests__/InvitesView.spec.ts`
  - Basic regression tests for admin invite operations page.
- `frontend/src/views/auth/RegisterView.vue`
  - Registration page that must support locked invite-link-prefilled invite codes.
- `frontend/src/views/auth/__tests__/RegisterView.spec.ts`
  - Registration UX regression tests for locked invite-link state versus normal editable state.
- `frontend/src/views/admin/SettingsView.vue`
  - Admin settings page that must mark legacy invitation-code registration as removed.
- `frontend/src/router/index.ts`
  - Official user/admin route definitions for `/invite` and `/admin/invites`.
- `frontend/src/components/layout/AppSidebar.vue`
  - Sidebar entries for user invite center and admin invite operations.
- `frontend/src/i18n/locales/zh.ts`
  - Chinese invite/register/settings copy, including removed-feature and locked-link messages.
- `frontend/src/i18n/locales/en.ts`
  - English invite/register/settings copy, including removed-feature and locked-link messages.

Important implementation note for this plan: many invite-specific files already exist in the current worktree as untracked files or tracked-but-uncommitted edits. Task 1 is intentionally about promoting those existing files into tracked source control and restoring reproducible code generation, not rewriting the invite system from scratch.

## Task 1: Formalize Invite Foundation and Reproducible Wiring

**Files:**
- Add: `backend/migrations/081_add_invite_growth_foundation.sql`
- Add: `backend/migrations/082_add_invite_admin_ops.sql`
- Add: `backend/ent/schema/invite_admin_action.go`
- Add: `backend/ent/schema/invite_relationship_event.go`
- Add: `backend/ent/schema/invite_reward_record.go`
- Add: `backend/ent/inviteadminaction.go`
- Add: `backend/ent/inviteadminaction/inviteadminaction.go`
- Add: `backend/ent/inviteadminaction_create.go`
- Add: `backend/ent/inviteadminaction_delete.go`
- Add: `backend/ent/inviteadminaction_query.go`
- Add: `backend/ent/inviteadminaction_update.go`
- Add: `backend/ent/inviterelationshipevent.go`
- Add: `backend/ent/inviterelationshipevent/inviterelationshipevent.go`
- Add: `backend/ent/inviterelationshipevent_create.go`
- Add: `backend/ent/inviterelationshipevent_delete.go`
- Add: `backend/ent/inviterelationshipevent_query.go`
- Add: `backend/ent/inviterelationshipevent_update.go`
- Add: `backend/ent/inviterewardrecord.go`
- Add: `backend/ent/inviterewardrecord/inviterewardrecord.go`
- Add: `backend/ent/inviterewardrecord_create.go`
- Add: `backend/ent/inviterewardrecord_delete.go`
- Add: `backend/ent/inviterewardrecord_query.go`
- Add: `backend/ent/inviterewardrecord_update.go`
- Add: `backend/internal/repository/invite_reward_record_repo.go`
- Add: `backend/internal/repository/invite_relationship_event_repo.go`
- Add: `backend/internal/repository/invite_admin_action_repo.go`
- Add: `backend/internal/repository/invite_admin_query_repo.go`
- Add: `backend/internal/repository/invite_growth_integration_test.go`
- Add: `backend/internal/repository/invite_admin_integration_test.go`
- Add: `backend/internal/repository/invite_admin_service_integration_test.go`
- Add: `backend/internal/repository/invite_rollout_verification_sql_integration_test.go`
- Add: `backend/internal/service/invite.go`
- Add: `backend/internal/service/invite_service.go`
- Add: `backend/internal/service/admin_invite.go`
- Add: `backend/internal/service/admin_service_invite.go`
- Add: `backend/internal/handler/dto/invite.go`
- Add: `backend/internal/handler/dto/admin_invite.go`
- Add: `backend/internal/handler/invite_handler.go`
- Add: `backend/internal/handler/invite_handler_test.go`
- Add: `backend/internal/handler/admin/invite_handler.go`
- Add: `backend/internal/handler/admin/invite_handler_test.go`
- Add: `frontend/src/api/invite.ts`
- Add: `frontend/src/api/admin/invites.ts`
- Add: `frontend/src/utils/inviteQuery.ts`
- Add: `frontend/src/utils/__tests__/inviteQuery.spec.ts`
- Add: `frontend/src/views/user/InviteView.vue`
- Add: `frontend/src/views/admin/InvitesView.vue`
- Add: `frontend/src/views/admin/__tests__/InvitesView.spec.ts`
- Modify: `backend/internal/repository/wire.go`
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/handler/wire.go`
- Modify: `backend/internal/server/routes/user.go`
- Modify: `backend/internal/server/routes/admin.go`
- Modify: `frontend/src/api/index.ts`
- Modify: `frontend/src/api/admin/index.ts`
- Modify: `frontend/src/router/index.ts`
- Modify: `frontend/src/components/layout/AppSidebar.vue`
- Regenerate: `backend/ent/client.go`
- Regenerate: `backend/ent/ent.go`
- Regenerate: `backend/ent/hook/hook.go`
- Regenerate: `backend/ent/intercept/intercept.go`
- Regenerate: `backend/ent/migrate/schema.go`
- Regenerate: `backend/ent/mutation.go`
- Regenerate: `backend/ent/predicate/predicate.go`
- Regenerate: `backend/ent/redeemcode.go`
- Regenerate: `backend/ent/redeemcode/redeemcode.go`
- Regenerate: `backend/ent/redeemcode/where.go`
- Regenerate: `backend/ent/redeemcode_create.go`
- Regenerate: `backend/ent/redeemcode_update.go`
- Regenerate: `backend/ent/runtime/runtime.go`
- Regenerate: `backend/ent/tx.go`
- Regenerate: `backend/ent/user.go`
- Regenerate: `backend/ent/user/user.go`
- Regenerate: `backend/ent/user/where.go`
- Regenerate: `backend/ent/user_create.go`
- Regenerate: `backend/ent/user_update.go`
- Regenerate: `backend/cmd/server/wire_gen.go`
- Test: `backend/internal/repository/invite_growth_integration_test.go`
- Test: `backend/internal/repository/invite_admin_integration_test.go`
- Test: `backend/internal/repository/invite_admin_service_integration_test.go`
- Test: `backend/internal/handler/invite_handler_test.go`
- Test: `backend/internal/handler/admin/invite_handler_test.go`
- Test: `frontend/src/views/user/__tests__/InviteView.spec.ts`
- Test: `frontend/src/views/admin/__tests__/InvitesView.spec.ts`

- [ ] **Step 1: Run the reproducibility checks before formalizing the foundation**

Run:

```bash
cd /root/sub2api-src
git status --short -- \
  backend/migrations/081_add_invite_growth_foundation.sql \
  backend/migrations/082_add_invite_admin_ops.sql \
  backend/internal/service/invite_service.go \
  backend/internal/repository/invite_growth_integration_test.go \
  frontend/src/views/user/InviteView.vue

cd /root/sub2api-src/backend
go generate ./cmd/server
go test -tags=integration ./internal/repository -run 'TestInvite(Admin|Growth)RepoSuite' -count=1 -v
```

Expected: `git status --short` shows the invite foundation is still living in untracked or modified files, which is the actual reason this work is not yet formalized as a reproducible tracked feature set. `go generate ./cmd/server` and the invite repository suite may or may not succeed in the dirty worktree, but they are informational only at this step and should not be treated as the failing gate.

- [ ] **Step 2: Promote the existing invite foundation files into tracked source control and fix the missing repository providers**

Update `backend/internal/repository/wire.go` so clean Wire generation can construct the invite admin path:

```go
var ProviderSet = wire.NewSet(
	NewUserRepository,
	NewInviteRewardRecordRepository,
	NewInviteRelationshipEventRepository,
	NewInviteAdminActionRepository,
	NewInviteAdminQueryRepository,
	NewAPIKeyRepository,
	NewGroupRepository,
	// ...
	NewRedeemCodeRepository,
	// ...
)
```

Keep the tracked handler/routes/export wiring aligned with the already-prepared invite implementation:

```go
// backend/internal/server/routes/user.go
invite := authenticated.Group("/invite")
{
	invite.GET("/summary", h.Invite.GetSummary)
	invite.GET("/rewards", h.Invite.ListRewards)
}
```

```go
// backend/internal/server/routes/admin.go
func registerInviteRoutes(admin *gin.RouterGroup, h *handler.Handlers) {
	invites := admin.Group("/invites")
	{
		invites.GET("/stats", h.Admin.Invite.GetStats)
		invites.GET("/relationships", h.Admin.Invite.ListRelationships)
		invites.GET("/rewards", h.Admin.Invite.ListRewards)
		invites.GET("/actions", h.Admin.Invite.ListActions)
		invites.POST("/rebind", h.Admin.Invite.Rebind)
		invites.POST("/manual-grants", h.Admin.Invite.CreateManualGrant)
		invites.POST("/recompute/preview", h.Admin.Invite.PreviewRecompute)
		invites.POST("/recompute/execute", h.Admin.Invite.ExecuteRecompute)
	}
}
```

```ts
// frontend/src/router/index.ts
{
  path: '/invite',
  name: 'Invite',
  component: () => import('@/views/user/InviteView.vue'),
  meta: {
    requiresAuth: true,
    requiresAdmin: false,
    title: 'Invite Center',
    titleKey: 'invite.title',
    descriptionKey: 'invite.description'
  }
},
{
  path: '/admin/invites',
  name: 'AdminInvites',
  component: () => import('@/views/admin/InvitesView.vue'),
  meta: {
    requiresAuth: true,
    requiresAdmin: true,
    title: 'Invite Operations',
    titleKey: 'admin.invites.title',
    descriptionKey: 'admin.invites.description'
  }
}
```

```ts
// frontend/src/components/layout/AppSidebar.vue
const userNavItems = computed(() => [
  { path: '/dashboard', label: t('nav.dashboard'), icon: DashboardIcon },
  { path: '/keys', label: t('nav.apiKeys'), icon: KeyIcon },
  { path: '/models', label: t('nav.modelHub'), icon: GridIcon },
  { path: '/usage', label: t('nav.usage'), icon: ChartIcon },
  { path: '/redeem', label: t('nav.redeem'), icon: GiftIcon },
  { path: '/invite', label: t('nav.invite'), icon: UsersIcon, hideInSimpleMode: true },
  { path: '/profile', label: t('nav.profile'), icon: UserIcon }
])

const adminNavItems = computed(() => [
  { path: '/admin/dashboard', label: t('admin.dashboard.title'), icon: DashboardIcon },
  // ...
  { path: '/admin/invites', label: t('nav.inviteOps'), icon: GiftIcon, hideInSimpleMode: true },
  { path: '/admin/redeem', label: t('nav.redeemCodes'), icon: GiftIcon }
])
```

Stage the existing invite foundation instead of rewriting it:

```bash
cd /root/sub2api-src
git add \
  backend/migrations/081_add_invite_growth_foundation.sql \
  backend/migrations/082_add_invite_admin_ops.sql \
  backend/ent/schema/invite_admin_action.go \
  backend/ent/schema/invite_relationship_event.go \
  backend/ent/schema/invite_reward_record.go \
  backend/ent/inviteadminaction.go \
  backend/ent/inviteadminaction \
  backend/ent/inviteadminaction_create.go \
  backend/ent/inviteadminaction_delete.go \
  backend/ent/inviteadminaction_query.go \
  backend/ent/inviteadminaction_update.go \
  backend/ent/inviterelationshipevent.go \
  backend/ent/inviterelationshipevent \
  backend/ent/inviterelationshipevent_create.go \
  backend/ent/inviterelationshipevent_delete.go \
  backend/ent/inviterelationshipevent_query.go \
  backend/ent/inviterelationshipevent_update.go \
  backend/ent/inviterewardrecord.go \
  backend/ent/inviterewardrecord \
  backend/ent/inviterewardrecord_create.go \
  backend/ent/inviterewardrecord_delete.go \
  backend/ent/inviterewardrecord_query.go \
  backend/ent/inviterewardrecord_update.go \
  backend/internal/repository/invite_reward_record_repo.go \
  backend/internal/repository/invite_relationship_event_repo.go \
  backend/internal/repository/invite_admin_action_repo.go \
  backend/internal/repository/invite_admin_query_repo.go \
  backend/internal/repository/invite_growth_integration_test.go \
  backend/internal/repository/invite_admin_integration_test.go \
  backend/internal/repository/invite_admin_service_integration_test.go \
  backend/internal/repository/invite_rollout_verification_sql_integration_test.go \
  backend/internal/service/invite.go \
  backend/internal/service/invite_service.go \
  backend/internal/service/admin_invite.go \
  backend/internal/service/admin_service_invite.go \
  backend/internal/handler/dto/invite.go \
  backend/internal/handler/dto/admin_invite.go \
  backend/internal/handler/invite_handler.go \
  backend/internal/handler/invite_handler_test.go \
  backend/internal/handler/admin/invite_handler.go \
  backend/internal/handler/admin/invite_handler_test.go \
  frontend/src/api/invite.ts \
  frontend/src/api/admin/invites.ts \
  frontend/src/utils/inviteQuery.ts \
  frontend/src/utils/__tests__/inviteQuery.spec.ts \
  frontend/src/views/user/InviteView.vue \
  frontend/src/views/user/__tests__/InviteView.spec.ts \
  frontend/src/views/admin/InvitesView.vue \
  frontend/src/views/admin/__tests__/InvitesView.spec.ts \
  backend/internal/repository/wire.go \
  backend/internal/handler/handler.go \
  backend/internal/handler/wire.go \
  backend/internal/server/routes/user.go \
  backend/internal/server/routes/admin.go \
  frontend/src/api/index.ts \
  frontend/src/api/admin/index.ts \
  frontend/src/router/index.ts \
  frontend/src/components/layout/AppSidebar.vue
```

- [ ] **Step 3: Regenerate Ent and Wire output so the branch is self-contained**

Run:

```bash
cd /root/sub2api-src/backend
go generate ./ent
go generate ./cmd/server
```

Stage the generated files:

```bash
cd /root/sub2api-src
git add \
  backend/ent/client.go \
  backend/ent/ent.go \
  backend/ent/hook/hook.go \
  backend/ent/intercept/intercept.go \
  backend/ent/migrate/schema.go \
  backend/ent/mutation.go \
  backend/ent/predicate/predicate.go \
  backend/ent/redeemcode.go \
  backend/ent/redeemcode/redeemcode.go \
  backend/ent/redeemcode/where.go \
  backend/ent/redeemcode_create.go \
  backend/ent/redeemcode_update.go \
  backend/ent/runtime/runtime.go \
  backend/ent/tx.go \
  backend/ent/user.go \
  backend/ent/user/user.go \
  backend/ent/user/where.go \
  backend/ent/user_create.go \
  backend/ent/user_update.go \
  backend/cmd/server/wire_gen.go
```

- [ ] **Step 4: Run focused backend and frontend invite checks**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=integration ./internal/repository -run 'TestInvite(Admin|Growth)RepoSuite' -count=1 -v
go test -tags=unit ./internal/handler ./internal/service -run 'TestInviteHandler_GetSummary|TestInviteHandler_ListRewards|TestInviteHandler_GetStats|TestInviteHandler_RebindRequiresReason' -count=1

cd /root/sub2api-src/frontend
npm test -- --run src/views/user/__tests__/InviteView.spec.ts src/views/admin/__tests__/InvitesView.spec.ts
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /root/sub2api-src
git commit -m "feat: formalize invite foundation"
```

## Task 2: Retire Legacy Invitation-Code Semantics and Lock Invite-Link Registration

**Files:**
- Modify: `backend/internal/service/auth_service.go`
- Modify: `backend/internal/service/auth_service_register_test.go`
- Modify: `backend/internal/handler/auth_handler.go`
- Modify: `backend/internal/handler/auth_handler_invite_test.go`
- Modify: `frontend/src/views/auth/RegisterView.vue`
- Modify: `frontend/src/views/auth/__tests__/RegisterView.spec.ts`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Test: `backend/internal/service/auth_service_register_test.go`
- Test: `backend/internal/handler/auth_handler_invite_test.go`
- Test: `frontend/src/views/auth/__tests__/RegisterView.spec.ts`

- [ ] **Step 1: Write the failing backend and frontend regression tests**

Replace the legacy-toggle tests with retirement and locked-link coverage.

In `backend/internal/service/auth_service_register_test.go`, add:

```go
type legacyInvitationLookupStub struct {
	codesByCode map[string]*RedeemCode
}

func (s *legacyInvitationLookupStub) GetByCode(_ context.Context, code string) (*RedeemCode, error) {
	if s.codesByCode == nil {
		return nil, ErrRedeemCodeNotFound
	}
	row, ok := s.codesByCode[code]
	if !ok {
		return nil, ErrRedeemCodeNotFound
	}
	return row, nil
}

func TestAuthService_RegisterWithVerification_ReturnsRemovedErrorForLegacyInvitationRedeemCode(t *testing.T) {
	repo := &inviteAuthUserRepoStub{userRepoStub: userRepoStub{nextID: 12}}
	inviteSvc := &InviteService{
		userRepo: repo,
		codeGenerator: func() (string, error) {
			return "NEWCODE12", nil
		},
	}
	authSvc := newAuthServiceWithInvite(repo, inviteSvc, map[string]string{
		SettingKeyRegistrationEnabled: "true",
	}, nil)
	authSvc.redeemRepo = &legacyInvitationLookupStub{
		codesByCode: map[string]*RedeemCode{
			"LEGACY001": {Code: "LEGACY001", Type: RedeemTypeInvitation},
		},
	}

	_, _, err := authSvc.RegisterWithVerification(context.Background(), "legacy@test.com", "password", "", "", "LEGACY001")
	require.ErrorIs(t, err, ErrInvitationCodeRemoved)
}

func TestAuthService_LoginOrRegisterOAuthWithTokenPair_IgnoresRetiredInvitationToggle(t *testing.T) {
	repo := &inviteAuthUserRepoStub{userRepoStub: userRepoStub{nextID: 13}}
	inviteSvc := &InviteService{
		userRepo: repo,
		codeGenerator: func() (string, error) {
			return "NEWCODE13", nil
		},
	}
	authSvc := newAuthServiceWithInvite(repo, inviteSvc, map[string]string{
		SettingKeyRegistrationEnabled:   "true",
		SettingKeyInvitationCodeEnabled: "true",
	}, nil)
	authSvc.refreshTokenCache = &refreshTokenCacheStub{}
	authSvc.cfg.JWT.RefreshTokenExpireDays = 7

	tokenPair, user, err := authSvc.LoginOrRegisterOAuthWithTokenPair(context.Background(), "oauth@test.com", "oauth-user", "")
	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	require.NotNil(t, user)
	require.Nil(t, user.InvitedByUserID)
}
```

In `backend/internal/handler/auth_handler_invite_test.go`, add:

```go
type inviteValidationRedeemRepoStub struct {
	codesByCode map[string]*service.RedeemCode
}

func (s *inviteValidationRedeemRepoStub) GetByCode(_ context.Context, code string) (*service.RedeemCode, error) {
	row, ok := s.codesByCode[code]
	if !ok {
		return nil, service.ErrRedeemCodeNotFound
	}
	return row, nil
}

func TestValidateInvitationCode_ReturnsRemovedErrorForLegacyInvitationRedeemCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/validate-invitation-code", strings.NewReader(`{"code":"LEGACY001"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	inviteSvc := service.NewInviteService(&inviteValidationUserRepoStub{}, &inviteValidationRewardRepoStub{})
	authSvc := service.NewAuthService(
		nil,
		nil,
		&inviteValidationRedeemRepoStub{
			codesByCode: map[string]*service.RedeemCode{
				"LEGACY001": {Code: "LEGACY001", Type: service.RedeemTypeInvitation},
			},
		},
		nil,
		&config.Config{},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		inviteSvc,
	)

	handler := &AuthHandler{
		authService:   authSvc,
		inviteService: inviteSvc,
	}

	handler.ValidateInvitationCode(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"valid":false`)
	require.Contains(t, rec.Body.String(), `"error_code":"INVITATION_CODE_REMOVED"`)
}
```

In `frontend/src/views/auth/__tests__/RegisterView.spec.ts`, replace the retired-toggle assertions with:

```ts
it('locks invite code when the registration page is entered from an invite link', async () => {
  routeMock.query = { invite: 'hello123' }

  const wrapper = mount(RegisterView, {
    global: {
      stubs: {
        AuthLayout: { template: '<div><slot /></div>' },
        LinuxDoOAuthSection: { template: '<div />' },
        TurnstileWidget: { template: '<div />' },
        Icon: { template: '<span />' },
        RouterLink: { template: '<a><slot /></a>' }
      }
    }
  })

  await flushPromises()

  const inviteInput = wrapper.find('#invitation_code')
  expect((inviteInput.element as HTMLInputElement).value).toBe('HELLO123')
  expect(inviteInput.attributes('readonly')).toBeDefined()
  expect(wrapper.text()).toContain('auth.invitationCodeLockedFromLink')
})

it('keeps invite code editable during ordinary registration when no invite link is present', async () => {
  routeMock.query = {}

  const wrapper = mount(RegisterView, {
    global: {
      stubs: {
        AuthLayout: { template: '<div><slot /></div>' },
        LinuxDoOAuthSection: { template: '<div />' },
        TurnstileWidget: { template: '<div />' },
        Icon: { template: '<span />' },
        RouterLink: { template: '<a><slot /></a>' }
      }
    }
  })

  await flushPromises()

  const inviteInput = wrapper.find('#invitation_code')
  expect(inviteInput.attributes('readonly')).toBeUndefined()
  await inviteInput.setValue('manual123')
  expect((inviteInput.element as HTMLInputElement).value).toBe('manual123')
})
```

- [ ] **Step 2: Run the targeted tests to verify the new expectations fail**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/service ./internal/handler -run 'TestAuthService_RegisterWithVerification_ReturnsRemovedErrorForLegacyInvitationRedeemCode|TestAuthService_LoginOrRegisterOAuthWithTokenPair_IgnoresRetiredInvitationToggle|TestValidateInvitationCode_ReturnsRemovedErrorForLegacyInvitationRedeemCode' -count=1

cd /root/sub2api-src/frontend
npm test -- --run src/views/auth/__tests__/RegisterView.spec.ts
```

Expected: FAIL because the backend still treats the legacy invitation toggle as active and the frontend still leaves the prefilled invite input editable.

- [ ] **Step 3: Implement retired-legacy semantics and locked invite-link UI**

Shrink the legacy-code lookup dependency in `backend/internal/service/auth_service.go` to the one method this service actually uses, then centralize invite-code resolution:

```go
type InvitationCodeLookupRepository interface {
	GetByCode(ctx context.Context, code string) (*RedeemCode, error)
}

var (
	ErrInvitationCodeInvalid  = infraerrors.BadRequest("INVITATION_CODE_INVALID", "invalid invite code")
	ErrInvitationCodeRemoved  = infraerrors.BadRequest("INVITATION_CODE_REMOVED", "legacy invitation code registration has been removed")
)

type AuthService struct {
	entClient          *dbent.Client
	userRepo           UserRepository
	redeemRepo         InvitationCodeLookupRepository
	refreshTokenCache  RefreshTokenCache
	cfg                *config.Config
	settingService     *SettingService
	emailService       *EmailService
	turnstileService   *TurnstileService
	emailQueueService  *EmailQueueService
	promoService       *PromoService
	defaultSubAssigner DefaultSubscriptionAssigner
	inviteService      *InviteService
}

func NewAuthService(
	entClient *dbent.Client,
	userRepo UserRepository,
	redeemRepo InvitationCodeLookupRepository,
	refreshTokenCache RefreshTokenCache,
	cfg *config.Config,
	settingService *SettingService,
	emailService *EmailService,
	turnstileService *TurnstileService,
	emailQueueService *EmailQueueService,
	promoService *PromoService,
	defaultSubAssigner DefaultSubscriptionAssigner,
	inviteService *InviteService,
) *AuthService {
	return &AuthService{
		entClient:          entClient,
		userRepo:           userRepo,
		redeemRepo:         redeemRepo,
		refreshTokenCache:  refreshTokenCache,
		cfg:                cfg,
		settingService:     settingService,
		emailService:       emailService,
		turnstileService:   turnstileService,
		emailQueueService:  emailQueueService,
		promoService:       promoService,
		defaultSubAssigner: defaultSubAssigner,
		inviteService:      inviteService,
	}
}

func (s *AuthService) resolveSubmittedInvitationCode(ctx context.Context, invitationCode string) (*User, error) {
	normalized := strings.ToUpper(strings.TrimSpace(invitationCode))
	if normalized == "" {
		return nil, nil
	}

	inviter, err := s.inviteService.ResolveInviterByCode(ctx, normalized)
	if err == nil {
		return inviter, nil
	}

	if s.redeemRepo != nil {
		code, codeErr := s.redeemRepo.GetByCode(ctx, normalized)
		if codeErr == nil && code != nil && code.Type == RedeemTypeInvitation {
			return nil, ErrInvitationCodeRemoved
		}
	}

	return nil, ErrInvitationCodeInvalid
}

func (s *AuthService) ValidateInvitationCode(ctx context.Context, code string) error {
	_, err := s.resolveSubmittedInvitationCode(ctx, code)
	return err
}
```

Apply that helper in both registration paths and retire the old required-toggle behavior:

```go
func (s *AuthService) RegisterWithVerification(ctx context.Context, email, password, verifyCode, promoCode, invitationCode string) (string, *User, error) {
	// registration_enabled and email policy checks stay unchanged

	if s.inviteService == nil {
		return "", nil, ErrServiceUnavailable
	}

	inviter, err := s.resolveSubmittedInvitationCode(ctx, invitationCode)
	if err != nil {
		return "", nil, err
	}

	// continue with user creation and bind inviter when inviter != nil
}

func (s *AuthService) LoginOrRegisterOAuthWithTokenPair(ctx context.Context, email, username, invitationCode string) (*TokenPair, *User, error) {
	// keep registration_enabled and account lookup logic unchanged

	inviter, err := s.resolveSubmittedInvitationCode(ctx, invitationCode)
	if err != nil {
		return nil, nil, err
	}

	// continue with new user creation when user not found
}
```

Update the public validation endpoint to return removed-feature semantics:

```go
func (h *AuthHandler) ValidateInvitationCode(c *gin.Context) {
	var req ValidateInvitationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if h.authService == nil {
		response.InternalError(c, "Auth service unavailable")
		return
	}

	err := h.authService.ValidateInvitationCode(c.Request.Context(), req.Code)
	if err == nil {
		response.Success(c, ValidateInvitationCodeResponse{Valid: true})
		return
	}

	errorCode := "INVITATION_CODE_NOT_FOUND"
	if errors.Is(err, service.ErrInvitationCodeRemoved) {
		errorCode = "INVITATION_CODE_REMOVED"
	}
	response.Success(c, ValidateInvitationCodeResponse{
		Valid:     false,
		ErrorCode: errorCode,
	})
}
```

Lock invite-link-prefilled codes in `frontend/src/views/auth/RegisterView.vue` while keeping ordinary manual entry editable:

```ts
const lockedInviteCode = computed(() => getInviteCodeFromQuery(route.query.invite))
const inviteCodeLocked = computed(() => lockedInviteCode.value !== '')

onMounted(async () => {
  const settings = await getPublicSettings()
  registrationEnabled.value = settings.registration_enabled
  emailVerifyEnabled.value = settings.email_verify_enabled
  promoCodeEnabled.value = settings.promo_code_enabled
  linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
  turnstileEnabled.value = settings.turnstile_enabled
  turnstileSiteKey.value = settings.turnstile_site_key
  siteName.value = settings.site_name
  registrationEmailSuffixWhitelist.value = settings.registration_email_suffix_whitelist || []

  if (lockedInviteCode.value) {
    formData.invitation_code = lockedInviteCode.value
    await validateInvitationCodeDebounced(lockedInviteCode.value)
  }
})

function handleInvitationCodeInput(): void {
  if (inviteCodeLocked.value) {
    return
  }
  const code = formData.invitation_code.trim()
  // existing debounce and validation behavior stays unchanged
}

function getInvitationErrorMessage(errorCode?: string): string {
  switch (errorCode) {
    case 'INVITATION_CODE_REMOVED':
      return t('auth.invitationCodeRemoved')
    case 'INVITATION_CODE_NOT_FOUND':
    case 'INVITATION_CODE_INVALID':
    case 'INVITATION_CODE_USED':
    default:
      return t('auth.invitationCodeInvalid')
  }
}
```

Update the input and helper text:

```vue
<input
  id="invitation_code"
  v-model="formData.invitation_code"
  type="text"
  :readonly="inviteCodeLocked"
  :disabled="isLoading"
  class="input pl-11 pr-10"
  :class="{
    'cursor-not-allowed bg-gray-50 dark:bg-dark-800': inviteCodeLocked,
    'border-green-500 focus:border-green-500 focus:ring-green-500': invitationValidation.valid,
    'border-red-500 focus:border-red-500 focus:ring-red-500': invitationValidation.invalid || errors.invitation_code
  }"
  :placeholder="t('auth.invitationCodePlaceholder')"
  @input="handleInvitationCodeInput"
/>

<p v-if="inviteCodeLocked" class="input-hint">
  {{ t('auth.invitationCodeLockedFromLink') }}
</p>
```

Add copy in `frontend/src/i18n/locales/zh.ts` and `frontend/src/i18n/locales/en.ts`:

```ts
// zh.ts
invitationCodeInvalid: '邀请码无效',
invitationCodeRemoved: '旧邀请码注册功能已下线，请使用用户分享的邀请码链接重新进入',
invitationCodeLockedFromLink: '该邀请码来自邀请链接。如需更换，请修改链接后重新进入注册页。',

// en.ts
invitationCodeInvalid: 'Invalid invite code',
invitationCodeRemoved: 'Legacy invitation-code registration has been removed. Please use a shared user invite link instead.',
invitationCodeLockedFromLink: 'This invite code came from the shared invite link. To change it, edit the link and reopen the registration page.',
```

- [ ] **Step 4: Run the focused backend and frontend tests again**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/service ./internal/handler -run 'TestAuthService_RegisterWithVerification_AssignsInviteCodeAndBindsInviter|TestAuthService_RegisterWithVerification_ReturnsRemovedErrorForLegacyInvitationRedeemCode|TestAuthService_RegisterWithVerification_RejectsUnknownPermanentInviteCode|TestAuthService_LoginOrRegisterOAuthWithTokenPair_IgnoresRetiredInvitationToggle|TestValidateInvitationCode_UsesPermanentInviteCodes|TestValidateInvitationCode_ReturnsRemovedErrorForLegacyInvitationRedeemCode' -count=1

cd /root/sub2api-src/frontend
npm test -- --run src/views/auth/__tests__/RegisterView.spec.ts
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /root/sub2api-src
git add \
  backend/internal/service/auth_service.go \
  backend/internal/service/auth_service_register_test.go \
  backend/internal/handler/auth_handler.go \
  backend/internal/handler/auth_handler_invite_test.go \
  frontend/src/views/auth/RegisterView.vue \
  frontend/src/views/auth/__tests__/RegisterView.spec.ts \
  frontend/src/i18n/locales/zh.ts \
  frontend/src/i18n/locales/en.ts
git commit -m "feat: retire legacy invitation code flow"
```

## Task 3: Make Base Invite Reward Settlement Atomic

**Files:**
- Modify: `backend/internal/service/invite_service.go`
- Modify: `backend/internal/service/wire.go`
- Modify: `backend/internal/service/invite_service_test.go`
- Add: `backend/internal/repository/invite_growth_service_integration_test.go`
- Test: `backend/internal/service/invite_service_test.go`
- Test: `backend/internal/repository/invite_growth_service_integration_test.go`

- [ ] **Step 1: Write the failing rollback coverage**

Add `backend/internal/repository/invite_growth_service_integration_test.go`:

```go
//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type failOnSecondInviteBalanceUserRepo struct {
	*userRepository
	calls int
	err   error
}

func (r *failOnSecondInviteBalanceUserRepo) UpdateBalance(ctx context.Context, id int64, amount float64) error {
	r.calls++
	if r.calls == 2 {
		return r.err
	}
	return r.userRepository.UpdateBalance(ctx, id, amount)
}

func TestInviteService_ApplyBaseRechargeRewards_RollsBackOnBalanceFailure(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	baseUserRepo := NewUserRepository(client, nil).(*userRepository)
	rewardRepo := NewInviteRewardRecordRepository(client)
	userRepo := &failOnSecondInviteBalanceUserRepo{
		userRepository: baseUserRepo,
		err:            errors.New("invitee balance update failed"),
	}

	inviter := &service.User{
		Email:        "atomic-inviter@example.com",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		InviteCode:   "ATOMIC01",
	}
	require.NoError(t, baseUserRepo.Create(ctx, inviter))

	inviterID := inviter.ID
	invitee := &service.User{
		Email:           "atomic-invitee@example.com",
		PasswordHash:    "hash",
		Role:            service.RoleUser,
		Status:          service.StatusActive,
		InviteCode:      "ATOMIC02",
		InvitedByUserID: &inviterID,
	}
	require.NoError(t, baseUserRepo.Create(ctx, invitee))

	svc := service.NewInviteService(userRepo, rewardRepo, client)

	err := svc.ApplyBaseRechargeRewards(ctx, invitee.ID, &service.RedeemCode{
		ID:         901,
		Type:       service.RedeemTypeBalance,
		SourceType: service.RedeemSourceCommercial,
		Value:      100,
	})
	require.ErrorContains(t, err, "invitee balance update failed")

	reloadedInviter, err := baseUserRepo.GetByID(ctx, inviter.ID)
	require.NoError(t, err)
	reloadedInvitee, err := baseUserRepo.GetByID(ctx, invitee.ID)
	require.NoError(t, err)
	require.Equal(t, 0.0, reloadedInviter.Balance)
	require.Equal(t, 0.0, reloadedInvitee.Balance)

	rows, _, err := rewardRepo.ListByRewardTarget(ctx, inviter.ID, pagination.PaginationParams{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Empty(t, rows)
}
```

Add a small constructor assertion to `backend/internal/service/invite_service_test.go`:

```go
func TestNewInviteService_WithEntClientSetsTransactionalClient(t *testing.T) {
	svc := NewInviteService(&inviteSettlementUserRepoStub{}, &inviteRewardRepoStub{}, nil)
	require.Nil(t, svc.entClient)

	svc = NewInviteService(&inviteSettlementUserRepoStub{}, &inviteRewardRepoStub{}, &dbent.Client{})
	require.NotNil(t, svc.entClient)
}

func TestInviteService_ApplyBaseRechargeRewardsSkipsSubscriptionRedeemCode(t *testing.T) {
	userRepo := &inviteSettlementUserRepoStub{
		users: map[int64]*User{
			8: {ID: 8, Status: StatusActive, InvitedByUserID: int64Ptr(7)},
		},
	}
	rewardRepo := &inviteRewardRepoStub{}
	svc := &InviteService{userRepo: userRepo, rewardRepo: rewardRepo}

	err := svc.ApplyBaseRechargeRewards(context.Background(), 8, &RedeemCode{
		ID: 103, Type: RedeemTypeSubscription, SourceType: RedeemSourceCommercial, Value: 100,
	})
	require.NoError(t, err)
	require.Empty(t, rewardRepo.created)
	require.Empty(t, userRepo.balanceUpdates)
}
```

- [ ] **Step 2: Run the new tests to verify the current write path fails the rollback expectation**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=integration ./internal/repository -run 'TestInviteService_ApplyBaseRechargeRewards_RollsBackOnBalanceFailure' -count=1 -v
go test -tags=unit ./internal/service -run 'TestNewInviteService_WithEntClientSetsTransactionalClient|TestInviteService_ApplyBaseRechargeRewardsSkipsSubscriptionRedeemCode' -count=1
```

Expected: FAIL because `ApplyBaseRechargeRewards` currently inserts reward rows before balance updates without a shared transaction.

- [ ] **Step 3: Wrap base reward settlement in an Ent transaction**

Extend `InviteService` so it can participate in transaction-bound repository calls:

```go
type InviteService struct {
	userRepo        InviteUserRepository
	rewardRepo      InviteRewardRecordRepository
	settingService  *SettingService
	registerURLBase string
	codeGenerator   func() (string, error)
	entClient       *dbent.Client
}

func NewInviteService(userRepo InviteUserRepository, rewardRepo InviteRewardRecordRepository, entClient ...*dbent.Client) *InviteService {
	svc := &InviteService{
		userRepo:        userRepo,
		rewardRepo:      rewardRepo,
		registerURLBase: "/register",
		codeGenerator:   defaultInviteCodeGenerator,
	}
	if len(entClient) > 0 {
		svc.entClient = entClient[0]
	}
	return svc
}

func ProvideInviteService(userRepo InviteUserRepository, rewardRepo InviteRewardRecordRepository, settingService *SettingService, entClient *dbent.Client) *InviteService {
	svc := NewInviteService(userRepo, rewardRepo, entClient)
	svc.settingService = settingService
	return svc
}

func (s *InviteService) withInviteWriteTx(ctx context.Context, fn func(context.Context) error) error {
	if dbent.TxFromContext(ctx) != nil || s.entClient == nil {
		return fn(ctx)
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
```

Then wrap the reward and balance writes:

```go
func (s *InviteService) ApplyBaseRechargeRewards(ctx context.Context, inviteeID int64, redeemCode *RedeemCode) error {
	// existing qualifying checks stay unchanged

	return s.withInviteWriteTx(ctx, func(txCtx context.Context) error {
		if err := s.rewardRepo.CreateBatch(txCtx, records); err != nil {
			if errors.Is(err, ErrInviteRewardAlreadyRecorded) {
				return nil
			}
			return err
		}

		if err := s.userRepo.UpdateBalance(txCtx, inviterID, inviterRewardAmount); err != nil {
			return err
		}
		if err := s.userRepo.UpdateBalance(txCtx, inviteeID, inviteeRewardAmount); err != nil {
			return err
		}
		return nil
	})
}
```

- [ ] **Step 4: Run the unit and integration coverage again**

Run:

```bash
cd /root/sub2api-src/backend
go test -tags=unit ./internal/service -run 'TestNewInviteService_WithEntClientSetsTransactionalClient|TestInviteService_ApplyBaseRechargeRewardsCreditsInviterAndInvitee|TestInviteService_ApplyBaseRechargeRewardsSkipsNonCommercialRecharge|TestInviteService_ApplyBaseRechargeRewardsSkipsEmptySourceType|TestInviteService_ApplyBaseRechargeRewardsSkipsSubscriptionRedeemCode' -count=1
go test -tags=integration ./internal/repository -run 'TestInviteService_ApplyBaseRechargeRewards_RollsBackOnBalanceFailure' -count=1 -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /root/sub2api-src
git add \
  backend/internal/service/invite_service.go \
  backend/internal/service/wire.go \
  backend/internal/service/invite_service_test.go \
  backend/internal/repository/invite_growth_service_integration_test.go
git commit -m "fix: make invite reward settlement atomic"
```

## Task 4: Remove Retired Invitation Toggle Surfaces and Finalize Verification

**Files:**
- Modify: `backend/internal/service/settings_view.go`
- Modify: `backend/internal/service/setting_service.go`
- Modify: `backend/internal/service/domain_constants.go`
- Modify: `backend/internal/handler/dto/settings.go`
- Modify: `backend/internal/handler/admin/setting_handler.go`
- Modify: `backend/internal/server/api_contract_test.go`
- Modify: `frontend/src/api/admin/settings.ts`
- Modify: `frontend/src/types/index.ts`
- Modify: `frontend/src/stores/app.ts`
- Modify: `frontend/src/views/auth/RegisterView.vue`
- Modify: `frontend/src/views/admin/SettingsView.vue`
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Test: `backend/internal/server/api_contract_test.go`
- Test: `frontend/src/views/auth/__tests__/RegisterView.spec.ts`
- Test: `frontend/src/views/user/__tests__/InviteView.spec.ts`
- Test: `frontend/src/views/admin/__tests__/InvitesView.spec.ts`

- [ ] **Step 1: Write the failing settings-contract expectation**

In `backend/internal/server/api_contract_test.go`, remove the retired field from the expected settings payload:

```go
{
  "registration_enabled": true,
  "email_verify_enabled": false,
  "registration_email_suffix_whitelist": [],
  "promo_code_enabled": true,
  "password_reset_enabled": false,
  "totp_enabled": false,
  "turnstile_enabled": false,
  "turnstile_site_key": "",
  "site_name": "Sub2API",
  "site_logo": "",
  "site_subtitle": "Subtitle",
  "api_base_url": "https://api.example.com",
  "contact_info": "support",
  "doc_url": "https://docs.example.com",
  "home_content": "",
  "hide_ccs_import_button": false,
  "purchase_subscription_enabled": false,
  "purchase_subscription_url": "",
  "sora_client_enabled": false,
  "linuxdo_oauth_enabled": false,
  "backend_mode_enabled": false,
  "version": "",
  "custom_menu_items": []
}
```

- [ ] **Step 2: Run the contract test and frontend build to verify cleanup is still incomplete**

Run:

```bash
cd /root/sub2api-src/backend
go test ./internal/server -run 'TestAPIContracts' -count=1

cd /root/sub2api-src/frontend
npm run build
```

Expected: FAIL because the backend and frontend settings contracts still expose `invitation_code_enabled`.

- [ ] **Step 3: Remove the retired toggle from contracts, keep a clear removed marker in admin settings, and leave historical DB rows untouched**

Stop exposing or persisting `invitation_code_enabled`, but do not add a data migration for historical `settings` rows. The retired key can remain in the database unused.

Remove the retired field from service/admin/public settings models:

```go
// backend/internal/service/settings_view.go
type SystemSettings struct {
	RegistrationEnabled              bool
	EmailVerifyEnabled               bool
	RegistrationEmailSuffixWhitelist []string
	PromoCodeEnabled                 bool
	PasswordResetEnabled             bool
	FrontendURL                      string
	TotpEnabled                      bool
	// ...
}

type PublicSettings struct {
	RegistrationEnabled              bool
	EmailVerifyEnabled               bool
	RegistrationEmailSuffixWhitelist []string
	PromoCodeEnabled                 bool
	PasswordResetEnabled             bool
	TotpEnabled                      bool
	// ...
}
```

Remove the key from settings assembly and persistence:

```go
// backend/internal/service/setting_service.go
keys := []string{
	SettingKeyRegistrationEnabled,
	SettingKeyEmailVerifyEnabled,
	SettingKeyRegistrationEmailSuffixWhitelist,
	SettingKeyPromoCodeEnabled,
	SettingKeyPasswordResetEnabled,
	SettingKeyTotpEnabled,
	SettingKeyTurnstileEnabled,
	// ...
}

return &PublicSettings{
	RegistrationEnabled:              settings[SettingKeyRegistrationEnabled] == "true",
	EmailVerifyEnabled:               emailVerifyEnabled,
	RegistrationEmailSuffixWhitelist: registrationEmailSuffixWhitelist,
	PromoCodeEnabled:                 settings[SettingKeyPromoCodeEnabled] != "false",
	PasswordResetEnabled:             passwordResetEnabled,
	TotpEnabled:                      settings[SettingKeyTotpEnabled] == "true",
	// ...
}

updates[SettingKeyRegistrationEnabled] = strconv.FormatBool(settings.RegistrationEnabled)
updates[SettingKeyEmailVerifyEnabled] = strconv.FormatBool(settings.EmailVerifyEnabled)
updates[SettingKeyPromoCodeEnabled] = strconv.FormatBool(settings.PromoCodeEnabled)
updates[SettingKeyPasswordResetEnabled] = strconv.FormatBool(settings.PasswordResetEnabled)
updates[SettingKeyFrontendURL] = settings.FrontendURL
updates[SettingKeyTotpEnabled] = strconv.FormatBool(settings.TotpEnabled)

// Remove the retired helper completely instead of leaving a dangling
// compile-time reference to SettingKeyInvitationCodeEnabled.
// func (s *SettingService) IsInvitationCodeEnabled(ctx context.Context) bool { ... }
```

Also remove the retired field from the public-settings injection payload:

```go
return &struct {
	RegistrationEnabled              bool            `json:"registration_enabled"`
	EmailVerifyEnabled               bool            `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string        `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool            `json:"promo_code_enabled"`
	PasswordResetEnabled             bool            `json:"password_reset_enabled"`
	TotpEnabled                      bool            `json:"totp_enabled"`
	TurnstileEnabled                 bool            `json:"turnstile_enabled"`
	TurnstileSiteKey                 string          `json:"turnstile_site_key,omitempty"`
	SiteName                         string          `json:"site_name"`
	// no invitation_code_enabled field remains here
}{
	RegistrationEnabled: settings.RegistrationEnabled,
	EmailVerifyEnabled:  settings.EmailVerifyEnabled,
	PromoCodeEnabled:    settings.PromoCodeEnabled,
	PasswordResetEnabled: settings.PasswordResetEnabled,
	TotpEnabled:         settings.TotpEnabled,
}
```

Remove the now-unused setting key constant from `backend/internal/service/domain_constants.go`:

```go
const (
	SettingKeyRegistrationEnabled              = "registration_enabled"
	SettingKeyEmailVerifyEnabled               = "email_verify_enabled"
	SettingKeyRegistrationEmailSuffixWhitelist = "registration_email_suffix_whitelist"
	SettingKeyPromoCodeEnabled                 = "promo_code_enabled"
	SettingKeyPasswordResetEnabled             = "password_reset_enabled"
	SettingKeyFrontendURL                      = "frontend_url"
	SettingKeyTotpEnabled                      = "totp_enabled"
	// no invitation_code_enabled entry remains here
)
```

Remove the field from admin handler DTOs and request/response mapping:

```go
// backend/internal/handler/dto/settings.go
type PublicSettings struct {
	RegistrationEnabled              bool     `json:"registration_enabled"`
	EmailVerifyEnabled               bool     `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool     `json:"password_reset_enabled"`
	TotpEnabled                      bool     `json:"totp_enabled"`
	// ...
}

type UpdateSettingsRequest struct {
	RegistrationEnabled              bool     `json:"registration_enabled"`
	EmailVerifyEnabled               bool     `json:"email_verify_enabled"`
	RegistrationEmailSuffixWhitelist []string `json:"registration_email_suffix_whitelist"`
	PromoCodeEnabled                 bool     `json:"promo_code_enabled"`
	PasswordResetEnabled             bool     `json:"password_reset_enabled"`
	TotpEnabled                      bool     `json:"totp_enabled"`
	// ...
}
```

Update frontend settings types and fallbacks:

```ts
// frontend/src/api/admin/settings.ts and frontend/src/types/index.ts
export interface SystemSettings {
  registration_enabled: boolean
  email_verify_enabled: boolean
  registration_email_suffix_whitelist: string[]
  promo_code_enabled: boolean
  password_reset_enabled: boolean
  totp_enabled: boolean
  // no invitation_code_enabled field
}

export interface PublicSettings {
  registration_enabled: boolean
  email_verify_enabled: boolean
  registration_email_suffix_whitelist: string[]
  promo_code_enabled: boolean
  password_reset_enabled: boolean
  turnstile_enabled: boolean
  // no invitation_code_enabled field
}
```

```ts
// frontend/src/stores/app.ts
return {
  registration_enabled: false,
  email_verify_enabled: false,
  registration_email_suffix_whitelist: [],
  promo_code_enabled: true,
  password_reset_enabled: false,
  turnstile_enabled: false,
  turnstile_site_key: '',
  site_name: siteName.value,
  site_logo: siteLogo.value,
  site_subtitle: '',
  api_base_url: apiBaseUrl.value,
  contact_info: contactInfo.value,
  doc_url: docUrl.value,
  home_content: '',
  hide_ccs_import_button: false,
  purchase_subscription_enabled: false,
  purchase_subscription_url: '',
  custom_menu_items: [],
  linuxdo_oauth_enabled: false,
  sora_client_enabled: false,
  backend_mode_enabled: false,
  version: ''
}
```

Update `frontend/src/views/auth/RegisterView.vue` to stop depending on the retired public setting while preserving the invite-link lock from Task 2:

```ts
const registrationEnabled = ref<boolean>(true)
const emailVerifyEnabled = ref<boolean>(false)
const promoCodeEnabled = ref<boolean>(true)
const turnstileEnabled = ref<boolean>(false)

onMounted(async () => {
  const settings = await getPublicSettings()
  registrationEnabled.value = settings.registration_enabled
  emailVerifyEnabled.value = settings.email_verify_enabled
  promoCodeEnabled.value = settings.promo_code_enabled
  turnstileEnabled.value = settings.turnstile_enabled
  // no invitation_code_enabled read remains here
})

function validateForm(): boolean {
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''
  errors.invitation_code = ''

  let isValid = true
  // existing email/password/turnstile checks stay unchanged
  return isValid
}
```

```vue
<label for="invitation_code" class="input-label">
  {{ t('auth.invitationCodeLabel') }}
  <span class="ml-1 text-xs font-normal text-gray-400 dark:text-dark-500">
    ({{ t('common.optional') }})
  </span>
</label>

<input
  id="invitation_code"
  v-model="formData.invitation_code"
  type="text"
  :readonly="inviteCodeLocked"
  :disabled="isLoading"
  class="input pl-11 pr-10"
  :placeholder="t('auth.invitationCodePlaceholder')"
  @input="handleInvitationCodeInput"
/>
```

Mark the admin feature as removed in `frontend/src/views/admin/SettingsView.vue` instead of exposing a live toggle:

```vue
<div class="card border-amber-200 bg-amber-50 dark:border-amber-900/50 dark:bg-amber-950/20">
  <div class="border-b border-amber-200 px-6 py-4 dark:border-amber-900/50">
    <h2 class="text-lg font-semibold text-amber-800 dark:text-amber-300">
      {{ t('admin.settings.legacyInvitation.title') }}
      <span class="ml-2 text-xs font-medium uppercase tracking-[0.18em]">
        {{ t('admin.settings.legacyInvitation.removedBadge') }}
      </span>
    </h2>
    <p class="mt-1 text-sm text-amber-700 dark:text-amber-200">
      {{ t('admin.settings.legacyInvitation.description') }}
    </p>
  </div>
</div>
```

Add locale entries:

```ts
// zh.ts
legacyInvitation: {
  title: '邀请码注册',
  removedBadge: '已移除',
  description: '旧的一次性邀请码注册功能已下线。当前仅保留用户邀请码分享和邀请关系运营能力。'
}

// en.ts
legacyInvitation: {
  title: 'Invitation Code Registration',
  removedBadge: 'Removed',
  description: 'The legacy one-time invitation-code registration flow has been retired. Only user invite sharing and invite operations remain active.'
}
```

- [ ] **Step 4: Run the contract, regression, and build verification sweep**

Run:

```bash
cd /root/sub2api-src/backend
go generate ./ent
go generate ./cmd/server
go test -tags=unit ./internal/service ./internal/handler ./internal/server -run 'TestAuthService_|TestInviteService_|TestInviteHandler_|TestValidateInvitationCode|TestAPIContracts' -count=1
go test -tags=integration ./internal/repository -run 'TestInvite(Admin|Growth)RepoSuite|TestInviteService_ApplyBaseRechargeRewards_RollsBackOnBalanceFailure|TestInviteRolloutVerificationSQL_' -count=1 -v

cd /root/sub2api-src/frontend
npm test -- --run src/views/auth/__tests__/RegisterView.spec.ts src/views/user/__tests__/InviteView.spec.ts src/views/admin/__tests__/InvitesView.spec.ts
npm run build
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /root/sub2api-src
git add \
  backend/internal/service/settings_view.go \
  backend/internal/service/setting_service.go \
  backend/internal/service/domain_constants.go \
  backend/internal/handler/dto/settings.go \
  backend/internal/handler/admin/setting_handler.go \
  backend/internal/server/api_contract_test.go \
  frontend/src/api/admin/settings.ts \
  frontend/src/types/index.ts \
  frontend/src/stores/app.ts \
  frontend/src/views/auth/RegisterView.vue \
  frontend/src/views/admin/SettingsView.vue \
  frontend/src/i18n/locales/zh.ts \
  frontend/src/i18n/locales/en.ts
git commit -m "chore: remove legacy invitation setting surfaces"
```
