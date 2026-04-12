# Invite Growth Baseline Reward Tuning Design

**Date:** 2026-04-12

**Goal:** Adjust the baseline invite reward to fit the current rollout without introducing a separate campaign engine. Keep the existing invite growth flow, but change the default reward rate from 5% to 3% and remove user-facing copy that hard-codes a percentage.

## Scope

This change is limited to the baseline invite reward flow that already exists in the codebase.

Included:

- Change the baseline invite reward rate from `5%` to `3%`.
- Keep the reward basis as the credited balance amount represented by commercial balance redeem codes.
- Keep the current rule that only commercial balance redeem codes trigger invite rewards.
- Update user-facing invite copy so it says both sides receive rewards, without exposing a fixed percentage.
- Keep storing actual reward amounts in backend records and admin views.

Excluded:

- No first-invite double-reward campaign logic.
- No event or campaign scheduling system.
- No switch to payment-order "actual paid amount" as the reward basis.
- No expansion to subscription codes, concurrency codes, admin grants, or system-grant balance codes.

## Current Behavior

The current invite growth implementation already supports:

- user-level permanent invite codes
- invite links derived from the user's invite code
- binding an inviter during registration and first-time OAuth registration
- automatic base reward issuance when a bound invitee redeems a commercial balance code
- reward ledgers for inviter and invitee

Today the reward rate is controlled by `InviteBaseRewardRate = 0.05`, and the user-facing invite page shows recharge/reward summaries without explicitly stating the percentage.

## Approved Product Rules

### 1. Reward basis

The baseline invite reward continues to use the balance amount credited by a redeem code, not any external order payment amount.

The authoritative input remains the existing redeem-code value already available in the reward pipeline.

### 2. What counts as a qualifying recharge

Only redeem codes that satisfy both conditions continue to trigger baseline invite rewards:

- `redeem_code.type = balance`
- `redeem_code.source_type = commercial`

This means the following do **not** count as qualifying recharge for baseline invite rewards:

- system grants
- admin compensation
- manual invite grants
- recompute delta rows
- subscription redeem codes
- concurrency redeem codes

### 3. Reward rate

The baseline reward rate changes from the current `5%` baseline to the following implementation rule:

- inviter receives `3%` of the credited balance amount
- invitee receives `3%` of the credited balance amount

The system continues to create two baseline reward rows per qualifying recharge:

- one row for the inviter
- one row for the invitee

### 4. User-facing copy

User-facing invite copy must avoid hard-coding a percentage. The invite center should communicate the rule at a product level:

- both sides receive rewards
- the invite link can be shared to bind new users
- reward history can be reviewed after qualifying recharge

This keeps the UI stable when later campaign layers such as "first invite double reward" are introduced.

## Design Changes

### Backend

- Update the baseline rate constant from `0.05` to `0.03`.
- Keep reward computation in the existing invite reward path.
- Keep storing exact computed amounts in `invite_reward_records`.
- Keep admin recompute logic derived from the same baseline rate constant so previews and repairs stay consistent with live issuance.

### Frontend

- Update the user invite-center copy to say both sides receive rewards, without showing `3%`.
- Keep existing summary cards, invite link, and reward history structure.
- Do not add campaign copy or temporary promotional badges in this change.

### Database / Records

No schema change is required.

Existing invite reward records remain valid historical data. This change only affects newly generated baseline invite rewards after deployment, plus any future recompute actions intentionally run by admins.

Historical rows created under the previous 5% rule are **not** rewritten automatically in this change.

## Operational Consequences

- After deployment, newly triggered baseline rewards will be smaller than before.
- Summary pages and admin views will still show real recorded amounts because they read ledger data, not a hard-coded percentage.
- If product later chooses to normalize historical records, that must be handled by a separate recompute or campaign design, not implicitly in this change.

## Testing

Update and extend existing tests to cover:

- baseline reward calculation uses `3%`
- inviter and invitee each receive their own `3%` row
- non-commercial balance codes still do not trigger invite rewards
- user-facing invite copy no longer promises a fixed percentage
- existing invite registration and binding flows still work unchanged

## Rollout Safety

This is a low-risk behavioral change because it reuses the current invite pipeline and does not add new tables, endpoints, or asynchronous jobs.

Primary rollout checks:

- confirm new registrations still bind correctly through invite links
- confirm commercial balance code redemption still creates two baseline reward rows
- confirm each row amount equals `credited_balance_amount * 0.03`
- confirm the user invite page copy no longer states a numeric percentage
