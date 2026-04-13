# Invite Foundation Completion Design

Date: 2026-04-11
Status: Approved in conversation, written for implementation planning

## Summary

This design completes the base invite layer so it is not only user-visible, but also operationally manageable and auditable.

The existing foundation already moves the product away from one-time `redeem_codes(type=invitation)` semantics toward permanent per-user invite codes, registration-time inviter binding, and `3% + 3%` recharge rewards on qualifying commercial `balance` redeem codes.

This follow-up design closes the remaining base-layer gaps:

- complete the user-facing referral loop
- add admin visibility into invite relationships and reward settlement
- add guarded admin write operations for exceptional correction cases
- keep all accounting auditable and append-only

This phase still does **not** implement invite campaigns, leaderboards, first-recharge bonus events, or sign-in activity mechanics.

## Confirmed Product Decisions

### Base-Layer Scope

- The base invite layer includes both user-facing referral behavior and admin operational tooling.
- The user-facing side includes permanent invite codes, invite links, registration binding, qualifying recharge rewards, invite summary, and reward history.
- The admin side includes relationship lookup, reward ledger lookup, invite statistics, manual reward grant, inviter rebind, and historical recompute.

### Roles and Permissions

- The project keeps the current two-role model: `admin` and `user`.
- This phase does **not** introduce `super_admin`.
- High-risk invite operations remain available to `admin`.
- High-risk operations must be guarded by product-level safety controls rather than new role tiers.

### High-Risk Operation Safety

- High-risk admin entry points remain visible in the admin UI.
- The UI must display prominent warning copy near those actions at all times.
- Every high-risk write operation requires:
  - non-empty reason
  - explicit secondary confirmation
  - durable audit logging
- This phase does not require an extra password or TOTP challenge for invite operations.

### Rebind Semantics

- Admin inviter rebind changes only the current and future referral relationship.
- Rebind does **not** automatically change already-issued historical rewards.
- Historical correction is a separate explicit action via manual grant or recompute.

### Recompute Strategy

- Recompute is intentionally conservative.
- It does not rewrite or delete historical reward rows.
- It produces append-only delta records and matching balance adjustments.
- It is limited to controlled scopes rather than full-system replay.

## Goals

1. Make the base invite system complete enough to run without manual database work.
2. Keep reward accounting traceable, auditable, and explainable.
3. Preserve simple read paths for normal invite summary and settlement.
4. Support exceptional operator corrections without hidden side effects.
5. Avoid expanding this phase into a campaign platform or a broad authorization refactor.

## Explicit Non-Goals

- No invite campaign CRUD in this phase.
- No first-recharge campaign bonus.
- No recharge leaderboard or ranking payout.
- No sign-in or attendance activity work.
- No `super_admin` role introduction.
- No full-system global recompute.
- No automatic historical reward rewrite during inviter rebind.

## Current System Context

The current in-progress foundation already covers most of the base invite path:

- users receive permanent invite codes
- registration and first-time OAuth registration can bind an inviter
- invite links use the configured frontend URL when available
- base rewards are triggered only by explicit commercial `balance` redeem codes
- the user invite center exposes invite code, invite link, summary totals, and reward history

What remains incomplete is the operational layer:

- admin query surfaces for relationships and rewards
- auditable correction workflows
- guarded write operations for rare exception handling
- bounded recompute semantics for historical repair

## Architecture

### Read Model

The base invite layer should keep the current read model simple:

- `users.invite_code` remains the permanent outward-facing code
- `users.invited_by_user_id` remains the current effective inviter
- `users.invite_bound_at` remains the timestamp of the original binding
- `invite_reward_records` remains the canonical reward ledger

This allows the ordinary product path to keep using direct reads from `users` plus the reward ledger without replaying historical events.

### Write Model

Exceptional admin operations must be modeled as explicit events and audited actions:

- relationship changes are stored as invite relationship events
- operator actions are stored as admin invite action audit entries
- reward corrections are stored as append-only reward ledger rows

This separates normal product reads from operator correction history without losing auditability.

## Data Model

### Existing User Fields

Keep the current invite fields on `users`:

- `invite_code`
- `invited_by_user_id`
- `invite_bound_at`

Rules:

- `invite_code` remains immutable
- `invited_by_user_id` represents the current effective inviter
- `invite_bound_at` records the original binding time and is not overwritten by admin rebind

### Existing Reward Ledger Extension

Extend `invite_reward_records` so it can represent both automatic and manual settlement work.

Suggested additions and rules:

- keep existing fields for inviter, invitee, trigger redeem code, target user, reward role, reward type, rate, amount, status, notes
- add `admin_action_id` nullable, referencing the admin action that caused manual grant or recompute delta
- allow `reward_amount` to be positive or negative
- preserve append-only semantics

Reward types in this phase:

- `base_invite_reward`
- `manual_invite_grant`
- `recompute_delta`

Semantics:

- automatic settlement continues to write `base_invite_reward`
- manual correction grants write `manual_invite_grant`
- historical recompute writes delta-only `recompute_delta` rows

### New Relationship Event Table

Add `invite_relationship_events`.

Suggested fields:

- `id`
- `invitee_user_id`
- `previous_inviter_user_id` nullable
- `new_inviter_user_id` nullable
- `event_type`
- `effective_at`
- `operator_user_id` nullable
- `reason` nullable for initial bind, required for admin rebind
- `created_at`

Event types in this phase:

- `register_bind`
- `admin_rebind`

Purpose:

- explain how the current relationship was reached
- provide a relationship timeline in admin UI
- give recompute logic the inviter relation effective at a point in time

### New Admin Action Audit Table

Add `invite_admin_actions`.

Suggested fields:

- `id`
- `action_type`
- `operator_user_id`
- `target_user_id`
- `reason`
- `request_snapshot_json`
- `result_snapshot_json`
- `created_at`

Action types in this phase:

- `rebind_inviter`
- `manual_reward_grant`
- `recompute_rewards`

Purpose:

- identify who did what and why
- connect high-risk UI actions to resulting rows
- preserve operator intent and resulting effect snapshots

## Business Semantics

### Qualifying Base Reward

Base reward semantics stay unchanged:

- trigger only on successful redemption of qualifying commercial `balance` redeem codes
- inviter receives `3%`
- invitee receives `3%`
- settlement remains idempotent per triggering redeem event

### Admin Rebind

Admin rebind changes the current inviter for an invitee.

Execution rules:

- validate invitee exists
- validate new inviter exists
- forbid self-invite
- forbid no-op rebind to the same inviter
- update `users.invited_by_user_id`
- do not modify `users.invite_bound_at`
- append `invite_relationship_events(event_type=admin_rebind)`
- append `invite_admin_actions(action_type=rebind_inviter)`

Accounting rules:

- do not modify historical reward ledger rows
- do not modify historical balances
- future qualifying reward settlement uses the new inviter

### Manual Reward Grant

Manual reward grant is the explicit tool for exceptional correction without historical replay.

Execution rules:

- target user must exist
- amount must be non-zero
- reason is required
- one action may create one or more ledger rows if both inviter and invitee need correction

Accounting rules:

- append `invite_reward_records(reward_type=manual_invite_grant)`
- connect rows to `admin_action_id`
- apply matching balance updates in the same transaction

### Historical Recompute

Historical recompute is the controlled repair tool when manual grant is not sufficient.

Execution model:

- `preview` computes but does not write
- `execute` recomputes and writes delta-only corrections

Accounting rules:

- do not delete old rows
- do not overwrite old rows
- compare expected reward totals to existing reward totals in the selected scope
- append only the net positive or negative difference as `recompute_delta`
- apply matching positive or negative balance delta

Relationship rule:

- recompute uses the inviter relationship effective at the time of the original qualifying recharge event, not the current relationship

## Recompute Scope

To keep the base layer conservative, recompute supports only bounded scopes:

- single invitee
- single inviter plus time range

Not supported in this phase:

- all users
- all inviters
- unrestricted full historical replay

This is sufficient for operational repair while keeping complexity and blast radius controlled.

## API Design

### Existing User-Facing Endpoints

Keep the existing invite endpoints:

- `GET /api/v1/invite/summary`
- `GET /api/v1/invite/rewards`

No new user-facing invite activity endpoints are introduced in this phase.

### New Admin Read Endpoints

Add:

- `GET /api/v1/admin/invites/stats`
- `GET /api/v1/admin/invites/relationships`
- `GET /api/v1/admin/invites/rewards`
- `GET /api/v1/admin/invites/actions`

Capabilities:

- pagination
- inviter / invitee filters
- reward type filters
- invite code and email search
- time-range filters where relevant
- CSV export using the same filter model

### New Admin Write Endpoints

Add:

- `POST /api/v1/admin/invites/rebind`
- `POST /api/v1/admin/invites/manual-grants`
- `POST /api/v1/admin/invites/recompute/preview`
- `POST /api/v1/admin/invites/recompute/execute`

Request requirements:

- all write requests require `reason`
- all write requests should run through admin write idempotency handling
- recompute execute must validate it still matches the preview scope and expectations

## Admin UI Design

### Route and Page Structure

Add a dedicated page:

- `/admin/invites`

This should follow the existing admin table-page patterns already used elsewhere in the project.

### Page Sections

Recommended sections:

- summary cards
- relationship explorer
- reward ledger explorer
- admin action audit log
- high-risk operations panel

Suggested summary cards:

- total invited users
- users with qualifying reward activity
- total base rewards
- total manual grants
- total recompute adjustments

### High-Risk Operations UX

High-risk actions stay visible, but with explicit warning presentation.

Requirements:

- warning copy is always visible near the dangerous actions
- warning copy clearly states these operations directly affect invite accounting
- each action opens a dedicated modal or panel
- each modal requires reason input
- each modal requires an explicit confirm step

Action-specific copy requirements:

- rebind must clearly state that it affects only future rewards
- manual grant must clearly state that it will append reward ledger rows and adjust balance immediately
- recompute must clearly state that it will append positive or negative delta records rather than rewrite history

### Recompute UX

Recompute must be a two-step flow:

1. preview
2. execute

Preview should show:

- selected scope
- qualifying recharge event count
- current ledger total
- recomputed expected total
- resulting net deltas

Execute should require a second confirmation phrase such as:

- `REBIND`
- `GRANT`
- `RECOMPUTE`

The exact confirmation phrase can be action-specific.

## Transactions and Idempotency

### Transaction Boundaries

Each high-risk write operation must be atomic.

Within a single transaction:

- create the admin action audit row
- create any relationship event rows
- create any reward ledger rows
- apply resulting balance updates
- commit

Any failure rolls back the entire operation.

### Idempotency

All admin invite writes must use idempotency protection comparable to other admin write paths.

This is required to prevent:

- duplicate manual grants
- duplicate recompute delta application
- duplicate rebind execution from repeated clicks or retries

## Validation Rules

### Rebind Validation

- invitee must exist
- new inviter must exist
- invitee cannot equal new inviter
- new inviter must differ from current inviter
- reason must be non-empty

### Manual Grant Validation

- target user must exist
- amount must be non-zero
- reason must be non-empty

### Recompute Validation

- scope must be supported and bounded
- preview must be generated before execute
- execute must verify preview inputs still match current request
- no-op execute may return success-without-write or explicit no-op, but must not emit meaningless delta rows

## Reporting and Export

The admin system should support CSV export for:

- filtered reward ledger results
- filtered relationship results
- filtered admin action audit results

Export uses the same filtering semantics as the table currently displayed so the operator sees and exports the same slice.

## Testing Strategy

### Backend Unit Tests

Add or extend unit tests for:

- rebind updates only future relationship state
- rebind does not rewrite historical rewards
- manual grant writes ledger rows and applies balance updates
- recompute preview produces correct positive, negative, and no-op outcomes
- execute writes only delta rows
- idempotent retries do not double-apply writes

### Backend Integration Tests

Add integration coverage for:

- relationship event persistence
- admin action audit persistence
- reward ledger append-only corrections
- transaction rollback on failure
- consistency between ledger delta and resulting balance change

### Backend Handler Tests

Add handler coverage for:

- admin-only access control
- required reason validation
- rebind warning semantics in responses where applicable
- recompute preview-before-execute flow

### Frontend Tests

Add frontend coverage for:

- admin invite page rendering and filtering
- warning copy visibility
- rebind confirmation flow
- manual grant confirmation flow
- recompute preview and execute flow

## Rollout Notes

- This design is intentionally the last step of the base invite layer, not the first step of the activity layer.
- Implementation should preserve compatibility with the already-started foundation work in the current branch.
- Campaign mechanics should be designed later on top of this audited, append-only base.
