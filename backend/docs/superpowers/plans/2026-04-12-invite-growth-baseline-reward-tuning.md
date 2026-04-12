# Invite Growth Baseline Reward Tuning Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Change the baseline invite reward from 5% to 3% on credited commercial balance codes, while updating user-facing invite copy to avoid exposing a fixed percentage.

**Architecture:** Reuse the existing invite growth pipeline instead of introducing a campaign layer. The only production backend behavior change is the shared baseline rate constant, which automatically affects live reward issuance and recompute previews. Frontend behavior stays structurally identical; only locale-driven copy changes so the invite center describes bilateral rewards without promising a fixed percentage.

**Tech Stack:** Go, `testify`, repository/service unit tests, Vue 3, Vitest, Vue I18n, Vite

---

## File Structure

- `backend/internal/service/invite.go`
  - Owns `InviteBaseRewardRate`, the single source of truth for baseline invite reward math.
- `backend/internal/service/invite_service_test.go`
  - Verifies live reward issuance credits inviter/invitee balances and persists the correct reward amounts/rates.
- `backend/internal/service/admin_service_invite_test.go`
  - Verifies recompute preview math derives deltas from the same baseline rate constant.
- `backend/internal/handler/invite_handler_test.go`
  - Verifies invite summary back-calculates credited recharge totals from stored base reward totals using the baseline rate.
- `frontend/src/i18n/locales/zh.ts`
  - Chinese invite-center copy shown to end users.
- `frontend/src/i18n/locales/en.ts`
  - English mirror of the invite-center copy.
- `frontend/src/views/user/__tests__/InviteView.spec.ts`
  - Frontend regression coverage for the invite-center description and the “no fixed percentage in UI copy” rule.

## Task 1: Tune Backend Baseline Reward Math to 3%

**Files:**
- Modify: `backend/internal/service/invite.go`
- Modify: `backend/internal/service/invite_service_test.go`
- Modify: `backend/internal/service/admin_service_invite_test.go`
- Modify: `backend/internal/handler/invite_handler_test.go`
- Test: `backend/internal/service/invite_service_test.go`
- Test: `backend/internal/service/admin_service_invite_test.go`
- Test: `backend/internal/handler/invite_handler_test.go`

- [ ] **Step 1: Write the failing tests**

Update the existing invite-focused tests so they encode the approved 3% baseline:

```go
// backend/internal/service/invite_service_test.go
func TestInviteService_ApplyBaseRechargeRewardsCreditsInviterAndInvitee(t *testing.T) {
	inviterID := int64(7)
	userRepo := &inviteSettlementUserRepoStub{
		users: map[int64]*User{
			7: {ID: 7, Email: "inviter@test.com", Balance: 0, Status: StatusActive},
			8: {ID: 8, Email: "invitee@test.com", Balance: 10, Status: StatusActive, InvitedByUserID: &inviterID},
		},
	}
	rewardRepo := &inviteRewardRepoStub{}
	svc := &InviteService{userRepo: userRepo, rewardRepo: rewardRepo}

	err := svc.ApplyBaseRechargeRewards(context.Background(), 8, &RedeemCode{
		ID: 101, Type: RedeemTypeBalance, SourceType: RedeemSourceCommercial, Value: 100,
	})
	require.NoError(t, err)
	require.Equal(t, 3.0, userRepo.users[7].Balance)
	require.Equal(t, 13.0, userRepo.users[8].Balance)
	require.Len(t, rewardRepo.created, 2)
	require.NotNil(t, rewardRepo.created[0].RewardRate)
	require.NotNil(t, rewardRepo.created[1].RewardRate)
	require.Equal(t, 3.0, rewardRepo.created[0].RewardAmount)
	require.Equal(t, 3.0, rewardRepo.created[1].RewardAmount)
	require.Equal(t, 0.03, *rewardRepo.created[0].RewardRate)
	require.Equal(t, 0.03, *rewardRepo.created[1].RewardRate)
}

// backend/internal/service/admin_service_invite_test.go
func TestPreviewInviteRecompute_ReturnsPositiveDelta(t *testing.T) {
	inviterID := int64(7)
	svc := &adminServiceImpl{
		inviteQualifyingRechargeRepo: &inviteQualifyingRechargeRepoStub{
			events: []InviteQualifyingRecharge{
				{InviteeUserID: 8, TriggerRedeemCodeID: 101, TriggerRedeemCodeValue: 100, UsedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)},
			},
		},
		inviteRelationshipEventRepo: &inviteRelationshipEventRepoStub{
			effectiveInviters: map[int64]int64{8: inviterID},
		},
		inviteRewardRecordRepo: &inviteWriteRewardRepoStub{
			scopeTotals: map[string]float64{
				"7:8:7:inviter": 0,
				"7:8:8:invitee": 0,
			},
		},
	}

	preview, err := svc.PreviewInviteRecompute(context.Background(), InviteRecomputeInput{
		OperatorUserID: 1,
		Reason:         "rebuild base rewards",
		Scope:          InviteRecomputeScope{InviteeUserID: inviteAdminInt64Ptr(8)},
	})
	require.NoError(t, err)
	require.Equal(t, 1, preview.QualifyingEventCount)
	require.Len(t, preview.Deltas, 2)
	got := []float64{preview.Deltas[0].DeltaAmount, preview.Deltas[1].DeltaAmount}
	require.ElementsMatch(t, []float64{3.0, 3.0}, got)
}

// backend/internal/handler/invite_handler_test.go
func TestInviteHandler_GetSummary(t *testing.T) {
	// ... existing setup unchanged ...
	require.Equal(t, 300.0, envelope.Data.InviteesRechargeTotal)
	require.Equal(t, 9.0, envelope.Data.BaseRewardsTotal)
}
```

- [ ] **Step 2: Run the focused tests to verify they fail**

Run:

```bash
cd backend
go test -tags=unit ./internal/service ./internal/handler -run 'TestInviteService_ApplyBaseRechargeRewardsCreditsInviterAndInvitee|TestPreviewInviteRecompute_ReturnsPositiveDelta|TestInviteHandler_GetSummary' -count=1
```

Expected: FAIL because the production baseline rate is still `0.05`, so balances, recompute deltas, and summary recharge totals still reflect 5%.

- [ ] **Step 3: Write the minimal implementation**

Change the shared baseline rate constant and leave the rest of the invite pipeline untouched:

```go
// backend/internal/service/invite.go
const (
	InviteRewardRoleInviter = "inviter"
	InviteRewardRoleInvitee = "invitee"

	InviteRewardTypeBase           = "base_invite_reward"
	InviteRewardTypeManualGrant    = "manual_invite_grant"
	InviteRewardTypeRecomputeDelta = "recompute_delta"

	InviteRelationshipEventTypeRegisterBind = "register_bind"
	InviteRelationshipEventTypeAdminRebind  = "admin_rebind"

	InviteAdminActionTypeRebind      = "rebind_inviter"
	InviteAdminActionTypeManualGrant = "manual_reward_grant"
	InviteAdminActionTypeRecompute   = "recompute_rewards"

	InviteBaseRewardRate = 0.03
)
```

No additional backend branching should be introduced. `ApplyBaseRechargeRewards`, `GetSummary`, and recompute preview logic should continue to read the single shared constant.

- [ ] **Step 4: Run invite-focused backend tests to verify they pass**

Run the focused regression first:

```bash
cd backend
go test -tags=unit ./internal/service ./internal/handler -run 'TestInviteService_ApplyBaseRechargeRewardsCreditsInviterAndInvitee|TestPreviewInviteRecompute_ReturnsPositiveDelta|TestInviteHandler_GetSummary' -count=1
```

Expected: PASS

Then run the broader invite-focused suites:

```bash
cd backend
go test -tags=unit ./internal/service -run 'TestInviteService_|TestPreviewInviteRecompute_|TestExecuteInviteRecompute_' -count=1
go test -tags=unit ./internal/handler -run 'TestInviteHandler_' -count=1
go test -tags=unit ./internal/service ./internal/handler -run 'TestAuthService_RegisterWithVerification_|TestAuthService_LoginOrRegisterOAuthWithTokenPair_RequiresInviteCodeWhenSettingEnabled|TestValidateInvitationCode_' -count=1
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/invite.go backend/internal/service/invite_service_test.go backend/internal/service/admin_service_invite_test.go backend/internal/handler/invite_handler_test.go
git commit -m "feat: tune baseline invite reward to three percent"
```

## Task 2: Remove Fixed Percentage Copy From the Invite Center

**Files:**
- Modify: `frontend/src/i18n/locales/zh.ts`
- Modify: `frontend/src/i18n/locales/en.ts`
- Modify: `frontend/src/views/user/__tests__/InviteView.spec.ts`
- Test: `frontend/src/views/user/__tests__/InviteView.spec.ts`

- [ ] **Step 1: Write the failing frontend regression test**

Replace the current mocked-key I18n setup with real locale messages and assert the invite-center description communicates bilateral rewards without any numeric percentage:

```ts
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { describe, expect, it, vi } from 'vitest'

import InviteView from '@/views/user/InviteView.vue'
import zh from '@/i18n/locales/zh'

vi.mock('@/api/invite', () => ({
  inviteAPI: {
    getSummary: vi.fn().mockResolvedValue({
      invite_code: 'HELLO123',
      invite_link: 'https://example.com/register?invite=HELLO123',
      invited_users_total: 4,
      invitees_recharge_total: 300,
      base_rewards_total: 9
    }),
    listRewards: vi.fn().mockResolvedValue({
      items: [
        {
          reward_role: 'invitee',
          reward_type: 'base_invite_reward',
          reward_amount: 3,
          created_at: '2026-04-11T08:00:00Z'
        }
      ],
      total: 1,
      page: 1,
      page_size: 20,
      pages: 1
    })
  }
}))

describe('InviteView', () => {
  it('renders bilateral reward copy without exposing a fixed percentage', async () => {
    const i18n = createI18n({
      legacy: false,
      locale: 'zh',
      messages: { zh }
    })

    const wrapper = mount(InviteView, {
      global: {
        plugins: [i18n],
        stubs: { AppLayout: { template: '<div><slot /></div>' } }
      }
    })

    await flushPromises()

    expect(wrapper.text()).toContain('HELLO123')
    expect(wrapper.text()).toContain('双方同时获赠奖励')
    expect(wrapper.text()).not.toContain('5%')
    expect(wrapper.text()).not.toContain('3%')
  })
})
```

- [ ] **Step 2: Run the frontend test to verify it fails**

Run:

```bash
cd frontend
npm test -- --run src/views/user/__tests__/InviteView.spec.ts
```

Expected: FAIL because the current invite-center description still says "分享邀请链接并查看充值返利", which does not contain the approved bilateral-reward wording.

- [ ] **Step 3: Write the minimal implementation**

Update the locale strings only; keep the component structure unchanged:

```ts
// frontend/src/i18n/locales/zh.ts
invite: {
  title: '邀请中心',
  description: '分享邀请链接，成功绑定后双方同时获赠奖励',
  myCode: '我的邀请码',
  invitedUsers: '已邀请用户数',
  totalRecharge: '邀请用户累计充值',
  totalRewards: '我的邀请奖励',
  totalRewardsCount: '{total} 条奖励记录',
  link: '邀请链接',
  copyLink: '复制链接',
  rewardHistory: '奖励记录',
  emptyRewards: '暂时还没有邀请奖励',
  roles: {
    inviter: '邀请人奖励',
    invitee: '被邀请人奖励'
  }
},

// frontend/src/i18n/locales/en.ts
invite: {
  title: 'Invite Center',
  description: 'Share your invite link so both sides receive rewards after a successful bind',
  myCode: 'My Invite Code',
  invitedUsers: 'Invited Users',
  totalRecharge: 'Invitees Recharge Total',
  totalRewards: 'My Invite Rewards',
  totalRewardsCount: '{total} reward records',
  link: 'Invite Link',
  copyLink: 'Copy Link',
  rewardHistory: 'Reward History',
  emptyRewards: 'No invite rewards yet',
  roles: {
    inviter: 'Inviter Reward',
    invitee: 'Invitee Bonus'
  }
},
```

Do not add the numeric rate to any user-facing string in this task.

- [ ] **Step 4: Run frontend verification plus a final cross-stack smoke check**

Run the focused frontend test:

```bash
cd frontend
npm test -- --run src/views/user/__tests__/InviteView.spec.ts
```

Expected: PASS

Then rebuild the frontend bundle:

```bash
cd frontend
npm run build
```

Expected: PASS

Then re-run the invite-focused backend smoke checks so the final branch state is verified together:

```bash
cd backend
go test -tags=unit ./internal/service ./internal/handler -run 'TestInviteService_ApplyBaseRechargeRewardsCreditsInviterAndInvitee|TestPreviewInviteRecompute_ReturnsPositiveDelta|TestInviteHandler_GetSummary' -count=1
go test -tags=unit ./internal/service ./internal/handler -run 'TestAuthService_RegisterWithVerification_|TestAuthService_LoginOrRegisterOAuthWithTokenPair_RequiresInviteCodeWhenSettingEnabled|TestValidateInvitationCode_' -count=1
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/i18n/locales/zh.ts frontend/src/i18n/locales/en.ts frontend/src/views/user/__tests__/InviteView.spec.ts
git commit -m "feat: soften invite center reward copy"
```
