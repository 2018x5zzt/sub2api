# User Subscription Balance Fallback UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the user subscription page let users turn on subscription balance fallback, choose the fallback balance group, set a positive cap, and save explicitly.

**Architecture:** Keep the existing `SubscriptionsView.vue` card and backend API. Convert the fallback controls from immediate-save inputs into a local form with explicit save/reset actions, while continuing to hydrate state from the authenticated user profile and available groups.

**Tech Stack:** Vue 3 Composition API, TypeScript, Vitest, Vue Test Utils, existing `updateProfile`, `userGroupsAPI.getAvailable`, `useAuthStore`, and `useAppStore`.

---

### Task 1: Add Failing Tests For Explicit Save Behavior

**Files:**
- Modify: `frontend/src/views/user/__tests__/SubscriptionsView.spec.ts`

- [ ] **Step 1: Write the failing tests**

Add these tests inside the existing `describe('SubscriptionsView product subscriptions', () => { ... })` block, after the current balance fallback test:

```ts
  it('lets a first-time user enable fallback locally before choosing group and limit', async () => {
    authStore.user = {
      subscription_balance_fallback_enabled: false,
      subscription_balance_fallback_limit_usd: 0,
      subscription_balance_fallback_used_usd: 0,
      subscription_balance_fallback_group_id: null
    }
    authStore.refreshUser.mockResolvedValue(authStore.user)
    getAvailableGroups.mockResolvedValue([
      { id: 11, name: 'Balance Pool', status: 'active', subscription_type: 'standard' }
    ])

    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true
        }
      }
    })

    await flushPromises()

    const toggle = wrapper.get('input[type="checkbox"]')
    await toggle.setValue(true)
    await flushPromises()

    expect(updateProfileMock).not.toHaveBeenCalled()
    expect(showError).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Balance Pool')
    expect((wrapper.vm as any).fallbackEnabled).toBe(true)
  })

  it('validates enabled fallback only when the user saves', async () => {
    authStore.user = {
      subscription_balance_fallback_enabled: false,
      subscription_balance_fallback_limit_usd: 0,
      subscription_balance_fallback_used_usd: 0,
      subscription_balance_fallback_group_id: null
    }
    authStore.refreshUser.mockResolvedValue(authStore.user)
    getAvailableGroups.mockResolvedValue([
      { id: 11, name: 'Balance Pool', status: 'active', subscription_type: 'standard' }
    ])

    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true
        }
      }
    })

    await flushPromises()

    const vm = wrapper.vm as unknown as {
      fallbackEnabled: boolean
      fallbackLimit: number
      fallbackGroupId: number | null
      saveBalanceFallbackSettings: () => Promise<void>
    }

    vm.fallbackEnabled = true
    vm.fallbackLimit = 12
    vm.fallbackGroupId = null
    await vm.saveBalanceFallbackSettings()

    expect(updateProfileMock).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalled()
    expect(vm.fallbackEnabled).toBe(true)

    showError.mockClear()
    vm.fallbackGroupId = 11
    vm.fallbackLimit = 0
    await vm.saveBalanceFallbackSettings()

    expect(updateProfileMock).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalled()
    expect(vm.fallbackEnabled).toBe(true)
  })

  it('saves disabled fallback explicitly and clears the fallback group', async () => {
    authStore.user = {
      subscription_balance_fallback_enabled: true,
      subscription_balance_fallback_limit_usd: 12,
      subscription_balance_fallback_used_usd: 3.5,
      subscription_balance_fallback_group_id: 11
    }
    authStore.refreshUser.mockResolvedValue(authStore.user)
    getAvailableGroups.mockResolvedValue([
      { id: 11, name: 'Balance Pool', status: 'active', subscription_type: 'standard' }
    ])
    updateProfileMock.mockResolvedValue({
      subscription_balance_fallback_enabled: false,
      subscription_balance_fallback_limit_usd: 12,
      subscription_balance_fallback_used_usd: 3.5,
      subscription_balance_fallback_group_id: null
    })

    const wrapper = mount(SubscriptionsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true
        }
      }
    })

    await flushPromises()

    const vm = wrapper.vm as unknown as {
      fallbackEnabled: boolean
      saveBalanceFallbackSettings: () => Promise<void>
    }
    vm.fallbackEnabled = false
    await vm.saveBalanceFallbackSettings()

    expect(updateProfileMock).toHaveBeenCalledWith({
      subscription_balance_fallback_enabled: false,
      subscription_balance_fallback_limit_usd: 12,
      subscription_balance_fallback_group_id: null
    })
  })
```

Also update the existing test named `requires enabled balance fallback to choose a standard balance group and positive limit` so it expects validation not to turn `fallbackEnabled` off:

```ts
    expect(vm.fallbackEnabled).toBe(true)
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
npm run test:run -- src/views/user/__tests__/SubscriptionsView.spec.ts
```

Expected: FAIL because the checkbox still calls `saveBalanceFallbackSettings` on change and validation still forces `fallbackEnabled` to `false`.

- [ ] **Step 3: Commit failing tests**

Do not commit failing tests separately. Leave them staged only after implementation passes.

### Task 2: Convert Fallback Controls To Explicit Save

**Files:**
- Modify: `frontend/src/views/user/SubscriptionsView.vue`
- Test: `frontend/src/views/user/__tests__/SubscriptionsView.spec.ts`

- [ ] **Step 1: Update the template**

In the balance fallback card:

1. Remove `@change="saveBalanceFallbackSettings"` from the checkbox.
2. Remove `@change="saveBalanceFallbackSettings"` from the select.
3. Remove `@blur="saveBalanceFallbackSettings"` from the limit input.
4. Add action buttons inside the `v-if="fallbackEnabled"` configuration section after the warning paragraph:

```vue
            <div class="mt-4 flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="btn btn-primary btn-sm"
                :disabled="savingFallback"
                @click="saveBalanceFallbackSettings"
              >
                {{ savingFallback ? t('common.saving', 'Saving...') : t('common.save', 'Save') }}
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="savingFallback"
                @click="resetBalanceFallbackForm"
              >
                {{ t('common.cancel', 'Cancel') }}
              </button>
            </div>
```

Add a compact save row for the disabled state as well, directly after the `v-if="fallbackEnabled"` block:

```vue
          <div v-else class="border-t border-gray-100 bg-gray-50/50 px-5 py-4 dark:border-dark-700 dark:bg-dark-800/50">
            <div class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="btn btn-primary btn-sm"
                :disabled="savingFallback"
                @click="saveBalanceFallbackSettings"
              >
                {{ savingFallback ? t('common.saving', 'Saving...') : t('common.save', 'Save') }}
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="savingFallback"
                @click="resetBalanceFallbackForm"
              >
                {{ t('common.cancel', 'Cancel') }}
              </button>
            </div>
          </div>
```

- [ ] **Step 2: Update the script**

Add this helper near `loadSubscriptions`:

```ts
function resetBalanceFallbackForm() {
  fallbackEnabled.value = Boolean(authStore.user?.subscription_balance_fallback_enabled)
  fallbackLimit.value = authStore.user?.subscription_balance_fallback_limit_usd || 0
  fallbackGroupId.value = authStore.user?.subscription_balance_fallback_group_id || null
}
```

Replace the validation part of `saveBalanceFallbackSettings` with:

```ts
  if (fallbackEnabled.value && !hasFallbackGroup) {
    appStore.showError(t('userSubscriptions.balanceFallback.groupRequired', 'Please select a balance group'))
    return
  }
  if (fallbackEnabled.value && !hasPositiveLimit) {
    appStore.showError(t('userSubscriptions.balanceFallback.limitRequired', 'Please set a positive fallback limit'))
    return
  }
```

Replace duplicated catch-state restoration with:

```ts
    resetBalanceFallbackForm()
```

Keep the payload shape:

```ts
    const updated = await updateProfile({
      subscription_balance_fallback_enabled: fallbackEnabled.value,
      subscription_balance_fallback_limit_usd: Math.max(fallbackLimit.value || 0, 0),
      subscription_balance_fallback_group_id: fallbackEnabled.value ? fallbackGroupId.value : null
    })
```

- [ ] **Step 3: Run test to verify it passes**

Run:

```bash
npm run test:run -- src/views/user/__tests__/SubscriptionsView.spec.ts
```

Expected: PASS.

- [ ] **Step 4: Run typecheck**

Run:

```bash
npm run typecheck
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add frontend/src/views/user/SubscriptionsView.vue frontend/src/views/user/__tests__/SubscriptionsView.spec.ts
git commit -m "fix: make subscription fallback settings explicit"
```

### Task 3: Final Verification

**Files:**
- Inspect: `frontend/src/views/user/SubscriptionsView.vue`
- Inspect: `frontend/src/views/user/__tests__/SubscriptionsView.spec.ts`

- [ ] **Step 1: Run focused tests again**

Run:

```bash
npm run test:run -- src/views/user/__tests__/SubscriptionsView.spec.ts
```

Expected: PASS.

- [ ] **Step 2: Check git state**

Run:

```bash
git status --short
```

Expected: clean, or only unrelated user changes.

- [ ] **Step 3: Summarize**

Report:

- The spec commit hash.
- The implementation commit hash.
- Test commands run and results.
