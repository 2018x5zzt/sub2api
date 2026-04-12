# Invite System Consolidation and Legacy Code Retirement Design

**Date:** 2026-04-12

**Goal:** Formalize the existing invite implementation already present in the worktree into a self-contained, mergeable feature set, while retiring the legacy one-time registration invitation-code flow and hardening invite reward writes to be transactionally consistent.

## Scope

This change consolidates the invite system into one product and data model:

- permanent user invite codes become the only valid invitation code type
- user invite links and invite binding remain the primary acquisition path
- only qualifying balance purchases trigger bilateral invite rewards
- the new admin invite operations module is promoted to a first-class supported feature
- legacy one-time registration invitation codes and promo-code invite behavior are retired from active use
- base invite reward writes are made transactionally consistent

Included:

- formalize the currently untracked invite migrations, ent schema, repositories, services, handlers, routes, user pages, and admin pages into the feature branch
- keep the baseline reward rate at `3%` for inviter and `3%` for invitee
- keep the reward basis as credited balance amount from commercial balance redeem codes
- keep the user-facing invite center and admin invite operations pages as official product surfaces
- reject legacy one-time invitation codes with an explicit product error
- remove user-facing legacy invitation-code entry points
- keep admin-facing legacy registration invitation-code surfaces only as clearly marked removed functionality
- ensure invite reward ledger rows and balance updates succeed or fail together

Excluded:

- no invite reward on subscription redemption or subscription purchase
- no first-invite double-reward campaign logic
- no campaign scheduler or temporary promotional engine
- no automatic rewrite of historical 5% reward records
- no silent compatibility path for legacy one-time invitation codes

## Current State

The repository currently contains two overlapping realities:

1. A newer invite growth implementation exists in the worktree but is not yet fully formalized into a self-contained branch. It includes:
   - permanent user invite codes
   - user invite center APIs and UI
   - invite reward ledgers
   - relationship event history
   - admin invite operations (rebind, manual grant, recompute)

2. A legacy one-time registration invitation-code / promo-code concept still exists in product language and parts of the admin surface, which conflicts with the permanent user invite-code model.

That overlap creates three problems:

- product ambiguity: “邀请码” can mean two different systems
- operational ambiguity: admins can still think the legacy path is active
- branch packaging risk: invite code, handler, and page changes rely on files that have not yet been formalized into a clean, mergeable unit

## Approved Product Rules

### 1. Invitation identity

The only valid invitation code going forward is the permanent invite code owned by a user.

That permanent code is used for:

- invite links
- registration-time inviter binding
- first-time OAuth registration inviter binding
- user invite-center sharing

Legacy one-time registration invitation codes are no longer part of the active product model.

### 2. Legacy invitation-code retirement

Legacy one-time registration invitation codes and their associated promo-code invite behavior are retired.

Retirement means:

- end users no longer see legacy invitation-code entry points as a supported feature
- backend validation no longer accepts legacy one-time invitation codes as valid input
- if a legacy code is submitted, the backend returns an explicit product error indicating the legacy invitation-code feature has been removed
- admin UI may retain a placeholder position for the old feature, but it must be clearly labeled as removed and must not allow continued creation or operation

This is a hard retirement, not a soft deprecation.

### 3. Invite reward trigger

Invite rewards are triggered only when the bound invitee purchases balance.

In the current data model, that remains the exact qualifying rule:

- `redeem_code.type = balance`
- `redeem_code.source_type = commercial`

The following do not trigger base invite rewards:

- subscription redeem codes
- subscription purchases
- benefit grants
- compensation grants
- system grants
- manual invite grants
- recompute delta rows
- concurrency codes

### 4. Reward amounts

For each qualifying balance purchase:

- inviter receives `3%` of the credited balance amount
- invitee receives `3%` of the credited balance amount

The system continues to write two base reward rows:

- one row for the inviter
- one row for the invitee

User-facing copy must not show the percentage.

### 5. User-facing copy

The invite center should describe the product in conservative language:

> 邀请好友注册并购买余额后，双方同时获赠奖励

The English mirror should communicate the same meaning without naming the numeric percentage.

The UI must not imply that subscription purchases or subscription redeem codes trigger invite rewards.

## Target Product Shape

### User side

The user invite center becomes the single active invite product surface.

It remains responsible for:

- showing the user’s permanent invite code
- showing the invite link
- showing invited user count
- showing invitees’ qualifying balance-purchase total
- showing the user’s invite reward total
- showing reward history

Registration and first-time OAuth registration continue to support inviter binding through the permanent invite code.

### Admin side

The new admin invite operations module becomes the official invite back-office surface.

It remains responsible for:

- invite relationship inspection
- reward ledger inspection
- admin action audit log
- inviter rebinding for future qualifying rewards
- manual invite grants
- recompute preview and execution

Legacy registration invitation-code management is no longer an active tool. If its navigation position remains, it must visibly state that the feature has been removed.

## Design Changes

### Backend

The backend change has four parts:

1. **Formalize invite foundation**
   - promote the current invite migrations, ent schemas, repositories, services, handlers, DTOs, routes, and tests into a coherent tracked implementation

2. **Retire legacy one-time invitation-code flow**
   - remove legacy code acceptance from registration validation
   - return a dedicated product error when a removed legacy code is submitted
   - close any remaining legacy write paths so the retired feature cannot keep operating through the API

3. **Preserve the active invite reward rule**
   - continue to compute rewards only for commercial balance redeem codes
   - continue to use `InviteBaseRewardRate = 0.03`
   - continue to derive admin recompute math from the same shared baseline rate constant

4. **Make base reward writes transactional**
   - base reward ledger inserts and balance updates must happen within one database transaction
   - on any failure, neither the ledger rows nor the balance mutations may partially persist
   - idempotent duplicate-trigger protection must remain intact

### Frontend

The frontend change has five parts:

1. **Promote invite center to official user surface**
   - keep the current invite-center structure
   - use the approved conservative copy
   - avoid any fixed-percentage copy

2. **Clarify purchase scope**
   - make it clear through wording that rewards are tied to invitees registering and purchasing balance
   - do not imply subscriptions trigger rewards

3. **Promote admin invite operations page**
   - keep the current admin invite operations page as the official operational tool

4. **Retire legacy invitation-code UI**
   - remove legacy invitation-code entry points from the end-user registration product flow
   - if legacy admin navigation remains, label it clearly as removed

5. **Tighten regression tests**
   - test the invite center using real locale bundles
   - enforce the “no percentage in user-facing invite copy” rule in both zh and en

### Database and Records

The invite data model introduced by the current worktree becomes the authoritative model:

- invite relationship events
- invite reward records
- invite admin actions

Existing historical data remains valid:

- historical invite rewards are not rewritten automatically
- historical admin corrections remain append-only records
- legacy invitation-code data may remain stored, but it is no longer active product input

## Error Handling

Legacy code retirement requires a distinct error path.

If a submitted code matches the removed legacy one-time invitation-code mode, the backend should return a dedicated, explicit error that the frontend can present as a feature-retirement message rather than a generic validation failure.

This is important because the user action is semantically different from:

- entering an unknown invite code
- entering an inactive inviter code
- leaving the field blank

## Transaction Consistency Requirement

Base invite rewards are financial writes. The system must not allow:

- reward rows written without matching balance updates
- balance updates applied without matching reward rows

The target implementation must guarantee atomicity across:

- inviter reward ledger row
- invitee reward ledger row
- inviter balance increment
- invitee balance increment

That write path should follow the same operational standard already used by the admin invite write flows.

## Rollout Consequences

After this change:

- invite semantics become simpler because only one invitation-code system remains
- only balance purchase events continue to create invite rewards
- subscription sales remain excluded from invite rewards
- admin users can use the new invite operations page as the only supported back-office invite tool
- legacy one-time invitation-code submissions fail explicitly instead of quietly behaving like a supported path

## Testing

The completed implementation should verify:

- permanent user invite codes remain valid across registration and OAuth binding
- legacy one-time invitation-code submissions return the removed-feature error
- only `balance + commercial` redeem codes trigger base invite rewards
- subscription redeem codes do not trigger invite rewards
- inviter and invitee each receive a `3%` base reward row for qualifying balance purchases
- base reward ledger writes and balance updates are atomic
- user invite-center copy in zh and en contains no percentage
- admin invite operations pages and APIs still load and execute against the formalized invite backend

## Rollout Safety

This is not a net-new feature design; it is a consolidation and hardening design for invite functionality that already exists in the worktree.

Primary rollout checks:

- clean checkout of the branch builds successfully
- user invite center loads and returns live data
- admin invite operations routes are wired and usable
- new registrations still bind inviters through permanent invite codes
- legacy one-time invitation codes are rejected explicitly
- qualifying balance purchases create exactly two base reward rows and matching balance updates inside one atomic write
