# Selective Upstream Channel Management Sync Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Selectively absorb the remaining upstream channel-management changes that improve `xlabapi`'s public `Available Channels` surface and core channel semantics, without pulling in unrelated monitor, payment, or profile work.

**Architecture:** Port a narrow set of upstream channel behaviors as local behavioral transplants instead of broad branch merges. Keep the public contract channel-first: users browse channels, inspect models within channels, and the code stays aligned with upstream naming and handler semantics while preserving `xlabapi`-specific rollout choices such as seeding `available_channels_enabled=true` in the affiliate compatibility migration.

**Tech Stack:** Go 1.26, Gin, testify, Vue 3.4, Vitest, Vue Test Utils, TypeScript.

---

## Scope Boundary

This plan intentionally includes only upstream deltas that directly affect:

- `Available Channels` user exposure and feature gating
- channel billing-model-source defaults and downstream semantics
- channel UI details that help users understand groups and models

This plan intentionally excludes:

- channel monitor / channel insights
- payment, WeChat OAuth, and `PaymentChannel` refactors
- account-stats pricing rules
- websearch emulation and `features_config`
- any attempt to restore a model-first `Model Hub`

## Upstream Deltas To Absorb

The narrow upstream behaviors worth porting are:

- `9ba42aa5` `feat(channels): gate available channels behind feature switch (backend)`
- `375aefa2` `refactor(channels): centralize BillingModelSource normalization and exhaustive enum maps`
- `6cd7c605` `fix(channels): supported models = mapping ∪ pricing with global LiteLLM fallback`
- `9dae6c7a` `feat(sidebar+groups): available-channels above channel-status; show rate for subscription groups`
- `ff4ef1b5` `feat(channels): themed model popover + group-badge with rate, subscription & exclusivity`

Two of those are already partially present in `xlabapi`; this plan closes the remaining gaps instead of re-importing whole commits.

## File Map

### Backend

- Create: `backend/internal/handler/available_channel_handler_test.go`
  - focused unit tests for the user-facing available-channel handler helper behavior
- Modify: `backend/internal/handler/available_channel_handler.go`
  - safer feature-gate default and keep platform leakage prevention explicit
- Modify: `backend/internal/service/channel.go`
  - centralize `BillingModelSource` normalization and add `channel_mapped`
- Modify: `backend/internal/service/channel_service.go`
  - use normalized billing-model-source defaults in create and resolve paths
- Modify: `backend/internal/service/channel_test.go`
  - unit tests for `normalizeBillingModelSource`
- Modify: `backend/internal/service/channel_service_test.go`
  - update resolve/create expectations for `channel_mapped`
- Modify: `backend/internal/service/channel_available.go`
  - keep public supported-model derivation aligned with current mapping ∪ pricing behavior and global pricing fallback comments/expectations
- Modify: `backend/internal/service/channel_available_test.go`
  - cover the public fallback behavior so later upstream syncs do not regress it

### Frontend

- Create: `frontend/src/components/common/__tests__/GroupBadge.spec.ts`
  - focused tests for subscription-group rate display on the available-channels surface
- Modify: `frontend/src/components/common/GroupBadge.vue`
  - add an opt-in `alwaysShowRate` prop so subscription groups can still show their multiplier when the caller needs that
- Modify: `frontend/src/components/channels/AvailableChannelsTable.vue`
  - pass the opt-in rate-display prop on the available-channels page
- Modify: `frontend/src/components/channels/SupportedModelChip.vue`
  - keep platform-hint/themed-popover behavior aligned if the current file still drifts after the backend changes

## Task 1: Harden The Backend Available-Channels Public Gate

**Files:**
- Create: `backend/internal/handler/available_channel_handler_test.go`
- Modify: `backend/internal/handler/available_channel_handler.go`

- [ ] **Step 1: Write failing handler tests for the public gate and platform filtering**

```go
package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAvailableChannelFeatureDisabledByDefaultWithoutSettingService(t *testing.T) {
	h := &AvailableChannelHandler{settingService: nil}
	require.False(t, h.featureEnabled(nil))
}

func TestBuildPlatformSections_OnlyReturnsModelsForVisiblePlatforms(t *testing.T) {
	price := 0.000001
	ch := service.AvailableChannel{
		Name: "official",
		SupportedModels: []service.SupportedModel{
			{Name: "claude-sonnet-4-6", Platform: service.PlatformAnthropic, Pricing: &service.ChannelModelPricing{InputPrice: &price}},
			{Name: "gpt-5.4", Platform: service.PlatformOpenAI, Pricing: &service.ChannelModelPricing{InputPrice: &price}},
		},
	}
	visibleGroups := []userAvailableGroup{
		{ID: 1, Name: "anthropic-sub", Platform: service.PlatformAnthropic},
	}

	sections := buildPlatformSections(ch, visibleGroups)
	require.Len(t, sections, 1)
	require.Equal(t, service.PlatformAnthropic, sections[0].Platform)
	require.Len(t, sections[0].SupportedModels, 1)
	require.Equal(t, "claude-sonnet-4-6", sections[0].SupportedModels[0].Name)
}
```

- [ ] **Step 2: Run the new backend tests and confirm RED**

Run:

```bash
cd /root/sub2api-src/backend && go test -tags unit ./internal/handler -run 'AvailableChannelFeatureDisabledByDefaultWithoutSettingService|BuildPlatformSections_OnlyReturnsModelsForVisiblePlatforms' -count=1
```

Expected:

```text
FAIL
```

The first test should fail because `featureEnabled` currently returns `true` when `settingService` is `nil`.

- [ ] **Step 3: Change the handler to fail closed when the feature service is absent**

```go
func (h *AvailableChannelHandler) featureEnabled(c *gin.Context) bool {
	if h.settingService == nil {
		return false
	}
	return h.settingService.GetAvailableChannelsRuntime(c.Request.Context()).Enabled
}
```

Do not change the migration seed that writes `available_channels_enabled=true`; the safer default here is only for missing DI / isolated tests.

- [ ] **Step 4: Re-run the focused backend handler tests and confirm GREEN**

Run:

```bash
cd /root/sub2api-src/backend && go test -tags unit ./internal/handler -run 'AvailableChannelFeatureDisabledByDefaultWithoutSettingService|BuildPlatformSections_OnlyReturnsModelsForVisiblePlatforms' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the backend public-gate hardening**

```bash
git -C /root/sub2api-src add backend/internal/handler/available_channel_handler.go backend/internal/handler/available_channel_handler_test.go
git -C /root/sub2api-src commit -m "fix(channels): fail closed when available-channels gate is not wired"
```

## Task 2: Normalize Billing Model Source To `channel_mapped`

**Files:**
- Modify: `backend/internal/service/channel.go`
- Modify: `backend/internal/service/channel_service.go`
- Modify: `backend/internal/service/channel_test.go`
- Modify: `backend/internal/service/channel_service_test.go`

- [ ] **Step 1: Write failing unit tests for the normalized default**

Add to `backend/internal/service/channel_test.go`:

```go
func TestNormalizeBillingModelSource_DefaultsToChannelMapped(t *testing.T) {
	ch := &Channel{}
	ch.normalizeBillingModelSource()
	require.Equal(t, BillingModelSourceChannelMapped, ch.BillingModelSource)
}
```

Update `backend/internal/service/channel_service_test.go`:

```go
func TestResolveChannelMapping_DefaultBillingModelSource(t *testing.T) {
	ch := Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		BillingModelSource: "",
	}
	repo := makeStandardRepo(ch, map[int64]string{10: "anthropic"})
	svc := newTestChannelService(repo)

	result := svc.ResolveChannelMapping(context.Background(), 10, "claude-opus-4")
	require.Equal(t, BillingModelSourceChannelMapped, result.BillingModelSource)
}
```

- [ ] **Step 2: Run the focused service tests and confirm RED**

Run:

```bash
cd /root/sub2api-src/backend && go test -tags unit ./internal/service -run 'NormalizeBillingModelSource_DefaultsToChannelMapped|ResolveChannelMapping_DefaultBillingModelSource' -count=1
```

Expected:

```text
FAIL
```

The failing expectation should still show `requested`.

- [ ] **Step 3: Port the normalized billing-model-source default into the entity and service paths**

In `backend/internal/service/channel.go`:

```go
const (
	BillingModelSourceRequested     = "requested"
	BillingModelSourceUpstream      = "upstream"
	BillingModelSourceChannelMapped = "channel_mapped"
)

func (c *Channel) normalizeBillingModelSource() {
	if c == nil {
		return
	}
	if c.BillingModelSource == "" {
		c.BillingModelSource = BillingModelSourceChannelMapped
	}
}
```

In `backend/internal/service/channel_service.go`, replace ad hoc defaults with the entity helper:

```go
channel := &Channel{
	Name:               input.Name,
	Description:        input.Description,
	Status:             StatusActive,
	BillingModelSource: input.BillingModelSource,
	RestrictModels:     input.RestrictModels,
	GroupIDs:           input.GroupIDs,
	ModelPricing:       input.ModelPricing,
	ModelMapping:       input.ModelMapping,
}
channel.normalizeBillingModelSource()
```

and:

```go
result := ChannelMappingResult{
	MappedModel:        model,
	ChannelID:          ch.ID,
	BillingModelSource: ch.BillingModelSource,
}
if result.BillingModelSource == "" {
	result.BillingModelSource = BillingModelSourceChannelMapped
}
```

Do not change the existing `billingModelForRestriction` switch logic. Its default branch already treats `channel_mapped` correctly.

- [ ] **Step 4: Re-run the focused service tests and confirm GREEN**

Run:

```bash
cd /root/sub2api-src/backend && go test -tags unit ./internal/service -run 'NormalizeBillingModelSource_DefaultsToChannelMapped|ResolveChannelMapping_DefaultBillingModelSource' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the billing-model-source normalization**

```bash
git -C /root/sub2api-src add backend/internal/service/channel.go backend/internal/service/channel_service.go backend/internal/service/channel_test.go backend/internal/service/channel_service_test.go
git -C /root/sub2api-src commit -m "refactor(channels): default billing model source to channel_mapped"
```

## Task 3: Show Subscription Group Rates On The Available-Channels Surface

**Files:**
- Create: `frontend/src/components/common/__tests__/GroupBadge.spec.ts`
- Modify: `frontend/src/components/common/GroupBadge.vue`
- Modify: `frontend/src/components/channels/AvailableChannelsTable.vue`

- [ ] **Step 1: Write the failing Vue test for subscription groups that should still show rates**

```ts
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import GroupBadge from '../GroupBadge.vue'

describe('GroupBadge', () => {
  it('shows rate for subscription groups when alwaysShowRate is enabled', () => {
    const wrapper = mount(GroupBadge, {
      props: {
        name: 'subscription-openai',
        platform: 'openai',
        subscriptionType: 'subscription',
        rateMultiplier: 1.5,
        alwaysShowRate: true
      },
      global: {
        stubs: {
          PlatformIcon: { template: '<span />' }
        }
      }
    })

    expect(wrapper.text()).toContain('1.5x')
  })
})
```

- [ ] **Step 2: Run the focused frontend test and confirm RED**

Run:

```bash
cd /root/sub2api-src/frontend && npm run test:run -- src/components/common/__tests__/GroupBadge.spec.ts
```

Expected:

```text
FAIL
```

The component should still render the subscription label instead of the rate.

- [ ] **Step 3: Add an opt-in rate-display prop and use it from AvailableChannelsTable**

In `frontend/src/components/common/GroupBadge.vue`:

```ts
interface Props {
  name: string
  platform?: GroupPlatform
  subscriptionType?: SubscriptionType
  rateMultiplier?: number
  userRateMultiplier?: number | null
  showRate?: boolean
  daysRemaining?: number | null
  alwaysShowRate?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  subscriptionType: 'standard',
  showRate: true,
  daysRemaining: null,
  userRateMultiplier: null,
  alwaysShowRate: false
})

const showLabel = computed(() => {
  if (!props.showRate) return false
  if (props.alwaysShowRate) return props.rateMultiplier !== undefined || hasCustomRate.value
  if (isSubscription.value) return true
  return props.rateMultiplier !== undefined || hasCustomRate.value
})

const labelText = computed(() => {
  if (props.alwaysShowRate && props.rateMultiplier !== undefined) {
    return `${props.rateMultiplier}x`
  }
  if (isSubscription.value) {
    if (props.daysRemaining !== null && props.daysRemaining !== undefined) {
      if (props.daysRemaining <= 0) return t('admin.users.expired')
      return t('admin.users.daysRemaining', { days: props.daysRemaining })
    }
    return t('groups.subscription')
  }
  return props.rateMultiplier !== undefined ? `${props.rateMultiplier}x` : ''
})
```

In `frontend/src/components/channels/AvailableChannelsTable.vue`, pass the new prop in both exclusive and public group rows:

```vue
<GroupBadge
  v-for="group in exclusiveGroups(section)"
  :key="`exclusive-${group.id}`"
  :name="group.name"
  :platform="group.platform as GroupPlatform"
  :subscription-type="(group.subscription_type || 'standard') as SubscriptionType"
  :rate-multiplier="group.rate_multiplier"
  :user-rate-multiplier="userGroupRates[group.id] ?? null"
  always-show-rate
/>
```

- [ ] **Step 4: Re-run the focused frontend test and confirm GREEN**

Run:

```bash
cd /root/sub2api-src/frontend && npm run test:run -- src/components/common/__tests__/GroupBadge.spec.ts
```

Expected:

```text
PASS
```

- [ ] **Step 5: Commit the group-rate UI parity change**

```bash
git -C /root/sub2api-src add frontend/src/components/common/GroupBadge.vue frontend/src/components/common/__tests__/GroupBadge.spec.ts frontend/src/components/channels/AvailableChannelsTable.vue
git -C /root/sub2api-src commit -m "feat(channels): show subscription group rates in available channels"
```

## Task 4: Verify The Narrow Sync End-To-End

**Files:**
- All files changed in Tasks 1-3.

- [ ] **Step 1: Run the backend channel-focused unit suite**

Run:

```bash
cd /root/sub2api-src/backend && go test -tags unit ./internal/handler ./internal/service -run 'AvailableChannel|BillingModelSource|ResolveChannelMapping|SupportedModels' -count=1
```

Expected:

```text
ok
```

- [ ] **Step 2: Run the frontend focused tests**

Run:

```bash
cd /root/sub2api-src/frontend && npm run test:run -- src/components/common/__tests__/GroupBadge.spec.ts
```

Expected:

```text
PASS
```

- [ ] **Step 3: Run frontend typecheck**

Run:

```bash
cd /root/sub2api-src/frontend && npm run typecheck
```

Expected:

```text
exit code 0
```

- [ ] **Step 4: Run diff hygiene checks**

Run:

```bash
git -C /root/sub2api-src diff --check
git -C /root/sub2api-src status --short
```

Expected:

```text
no whitespace errors
only the intended channel-related files changed
```

- [ ] **Step 5: Create the integration commit**

```bash
git -C /root/sub2api-src add backend/internal/handler/available_channel_handler.go backend/internal/handler/available_channel_handler_test.go backend/internal/service/channel.go backend/internal/service/channel_service.go backend/internal/service/channel_test.go backend/internal/service/channel_service_test.go frontend/src/components/common/GroupBadge.vue frontend/src/components/common/__tests__/GroupBadge.spec.ts frontend/src/components/channels/AvailableChannelsTable.vue
git -C /root/sub2api-src commit -m "sync(channels): absorb selective upstream available-channel parity fixes"
```
