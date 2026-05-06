# User Subscription Balance Fallback UI

## Context

Product subscription balance fallback already exists as a user-level backend feature. The user model stores:

- `subscription_balance_fallback_enabled`
- `subscription_balance_fallback_group_id`
- `subscription_balance_fallback_limit_usd`
- `subscription_balance_fallback_used_usd`

The profile update API already accepts these fields, and the service validates that an enabled fallback has a positive limit and a valid active standard group visible to the user.

The current user subscription page has the right controls, but the fallback switch saves immediately. A first-time user cannot enable the switch cleanly because they have not yet selected the fallback group or entered the limit, so validation fails and the switch is turned off again.

## Goal

Make the existing user subscription page provide a usable binding flow:

1. The user can turn on the local switch.
2. The page opens the fallback configuration fields.
3. The user selects which standard balance group should be used for fallback.
4. The user sets a positive cumulative fallback cap.
5. The user explicitly saves the setting.

## Scope

In scope:

- Update `frontend/src/views/user/SubscriptionsView.vue`.
- Keep the control on the user subscription page.
- Use `userGroupsAPI.getAvailable()` as the source of user-visible groups.
- Offer only active standard groups as fallback targets.
- Preserve backend validation as the final authority.
- Update existing `SubscriptionsView` tests.

Out of scope:

- Changing billing semantics.
- Adding new backend fields or migrations.
- Changing admin user-edit behavior.
- Changing product subscription quota resolution.

## UX Design

The balance fallback card remains at the top of the user subscription page.

The switch becomes a local form control instead of an immediate save trigger. When enabled, it reveals:

- A standard balance group selector.
- A cumulative USD fallback limit input.
- Current fallback used amount.
- Remaining fallback budget.
- A warning that negative balance blocks future requests until recharge.
- Save and cancel/reset actions.

When the current saved setting is disabled, clicking the switch should keep it visibly on and allow the user to complete required fields before saving. It must not immediately call the backend or turn itself off.

When saving an enabled configuration:

- Missing fallback group shows a frontend validation error.
- Non-positive fallback limit shows a frontend validation error.
- Valid input calls `updateProfile` with enabled, limit, and group id.

When saving a disabled configuration:

- The frontend calls `updateProfile` with `subscription_balance_fallback_enabled: false`.
- The fallback group sent to the backend is `null`.
- The limit is normalized to a non-negative number.
- The already-used amount remains server-owned and is not reset by the user UI.

Cancel/reset restores the form from the current authenticated user profile.

## Data Flow

On page load, fetch in parallel:

- active legacy subscriptions
- active product subscriptions
- refreshed user profile
- available user groups

Initialize fallback form state from the refreshed profile.

Build fallback group options by filtering available groups to:

- `status === "active"`
- `subscription_type === "standard"`

After a successful save:

- Replace `authStore.user` with the updated profile.
- Rehydrate fallback form state from the updated profile.
- Show the existing saved toast.

After a failed save:

- Show an error.
- Rehydrate fallback form state from `authStore.user`.

## Testing

Update `frontend/src/views/user/__tests__/SubscriptionsView.spec.ts` to cover:

1. A first-time user can enable the local switch without an immediate API call or immediate validation failure.
2. Saving while enabled without a group is rejected.
3. Saving while enabled without a positive limit is rejected.
4. Saving a valid enabled configuration sends the selected group id and positive limit.
5. Saving a disabled configuration sends disabled state and clears the fallback group id.

Existing product subscription display tests should continue to pass.
