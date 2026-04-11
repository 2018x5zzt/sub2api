# Invite Growth Design

Date: 2026-04-11
Status: Draft approved in conversation, written for implementation planning

## Summary

This design replaces the current one-time "invitation redeem code" registration gate with a persistent single-layer referral system built around permanent per-user invite codes.

The system has two layers:

1. A stable base referral mechanism:
   - Every user gets one permanent invite code.
   - A new user can register with that code or via an invite link.
   - Registration creates a permanent `inviter -> invitee` binding.
   - When the invitee later redeems a qualifying paid `balance` redeem code, both sides receive the configured base reward.

2. A flexible campaign layer:
   - Campaigns never reward registration counts.
   - Campaigns only reward outcomes tied to qualifying paid recharge behavior.
   - First-phase campaigns are limited to safe patterns such as first-recharge bonus and recharge-amount leaderboard.

The immediate objective is growth from roughly 500 users toward 5000 users without introducing multi-level commission logic or easy registration-count abuse.

## Product Decisions

### Confirmed Decisions

- Each user has one unique permanent invite code.
- Invite code and invite link are both supported.
- Invite binding is permanent after registration and cannot be changed later.
- Base rewards are triggered only by successful redemption of qualifying `balance` redeem codes.
- Base rewards are `5%` to inviter and `5%` to invitee for each qualifying event.
- The existing registration input field `invitation_code` is reused.
- The existing public validation endpoint `/api/v1/auth/validate-invitation-code` is reused.
- Campaigns must never reward raw registration count.
- Campaigns may reward only qualifying recharge outcomes.

### Explicit Non-Goals

- No multi-level referral tree.
- No team-leader / distributor hierarchy.
- No "invite N registered users to earn X" campaigns.
- No sign-in campaign work in this phase.
- No manual "rebind inviter" operation.

## Current System Context

The current implementation uses `redeem_codes(type=invitation)` as a one-time registration gate:

- registration validates the code
- registration marks the code as used
- no persistent inviter/invitee relation is stored
- later balance changes have no referral linkage

Current balance-affecting paths include:

- user redeeming `balance` redeem codes
- external payment integrations using admin create-and-redeem
- admin manual balance adjustment
- promo / benefit / other balance grant paths

This design narrows referral reward triggering to qualifying paid `balance` redeem code redemption events only.

## Design Principles

1. Keep the base mechanism simple and stable.
2. Separate referral binding from redeem-code inventory.
3. Separate base reward accounting from campaign accounting.
4. Treat "registration" and "effective recharge" as different concepts.
5. Make campaign logic extensible without rewriting the base referral logic.
6. Avoid product mechanics that can be gamed by registration count alone.

## Recommended Architecture

### Base Referral Layer

The base layer provides:

- permanent invite code generation per user
- invite-code validation for registration
- one-time inviter binding during account creation
- reward settlement when qualifying paid recharge happens
- user-visible referral summary and reward history

### Campaign Layer

The campaign layer provides:

- optional, time-bounded campaign definitions
- campaign-specific qualifying conditions
- campaign-only reward settlement and leaderboards
- separation from base reward settlement and reporting

The campaign layer uses the same referral graph and recharge trigger events as the base layer. It does not redefine referral relationships.

## Data Model

### User Extensions

Add the following fields to `users`:

- `invite_code`: string, unique, permanent, generated once
- `invited_by_user_id`: nullable bigint, references `users.id`
- `invite_bound_at`: nullable timestamptz

Rationale:

- `invite_code` allows direct validation and link generation without touching `redeem_codes`
- `invited_by_user_id` makes the binding cheap to read for settlement and dashboard queries
- `invite_bound_at` gives a clear audit point for when the relation became effective

Rules:

- `invite_code` is immutable after initial generation
- `invited_by_user_id` can be set only once
- self-invite is forbidden

### Invite Reward Records

Add a dedicated settlement ledger table, for example `invite_reward_records`.

Suggested fields:

- `id`
- `inviter_user_id`
- `invitee_user_id`
- `trigger_redeem_code_id`
- `trigger_redeem_code_value`
- `reward_target_user_id`
- `reward_role` (`inviter`, `invitee`)
- `reward_type` (`base_invite_reward`, `campaign_bonus`, `campaign_rank_reward`)
- `campaign_id` nullable
- `reward_rate` nullable
- `reward_amount`
- `status`
- `notes` nullable
- `created_at`

Purpose:

- base and campaign reward accounting should not be reconstructed indirectly from user balance history
- this table is the source of truth for referral reward reporting

Uniqueness / idempotency expectations:

- base reward: unique on `trigger_redeem_code_id + reward_role + reward_type`
- campaign bonus: unique on `trigger_redeem_code_id + campaign_id + reward_role + reward_type`
- campaign rank reward: unique on `campaign_id + reward_target_user_id + reward_type`

### Invite Campaigns

Add a campaign configuration table, for example `invite_campaigns`.

Suggested fields:

- `id`
- `name`
- `type` (`first_recharge_bonus`, `recharge_amount_leaderboard`, future-safe extensions)
- `status`
- `start_at`
- `end_at`
- `config_json`
- `created_at`
- `updated_at`

Phase-one campaign types:

- `first_recharge_bonus`
- `recharge_amount_leaderboard`

### Redeem Code Qualification Metadata

Campaign safety depends on distinguishing real paid recharge codes from benefit or compensation grants.

Add qualification metadata to redeem codes. This can be done via one of the following:

- `source_type`
- or a narrower boolean such as `invite_reward_eligible`

Recommended shape:

- `source_type` with values such as:
  - `commercial`
  - `benefit`
  - `compensation`
  - `system_grant`

Qualification rule:

- only `source_type = commercial` participates in campaign statistics
- base referral reward may also use the same qualification rule to keep semantics clean and predictable

This prevents free benefit codes from generating invite campaign value.

## Core Business Definitions

### Invite Binding

Invite binding means:

- the registering user supplied a valid permanent invite code
- the code resolved to an inviter user
- registration completed successfully
- the new user record now stores `invited_by_user_id`

Binding is:

- one-time
- permanent
- not editable
- not retroactively backfillable

### Qualifying Recharge Event

A recharge event qualifies for referral settlement only if all of the following are true:

- redeem code type is `balance`
- redeem code redemption succeeded
- the redeem code is considered paid / commercial according to qualification metadata
- the redeemed user has a valid `invited_by_user_id`

Non-qualifying events include:

- invitation gating codes
- benefit / welfare codes
- promo code balance grants
- admin manual balance adjustments
- compensation grants
- subscription redeem codes
- concurrency redeem codes

### Campaign-Effective Recharge

Campaigns use a stricter unit than "registration" or even "any recharge":

- campaign statistics are based on qualifying recharge outcomes
- not on registration count
- not on invite bindings alone
- not on bonus-only or free-credit events

## User Flows

### 1. Invite Link and Invite Code

For an existing user:

- the system exposes their permanent invite code
- the UI can generate a registration URL such as `/register?invite=<code>`
- invite code can also be copied and shared directly

### 2. Registration

Registration continues to accept `invitation_code`.

Behavior:

- if invite code exists, resolve it against user invite codes
- if valid, continue registration
- on successful account creation, write `invited_by_user_id`
- if no invite code is supplied, registration still succeeds unless the site separately decides to require invitation-only registration

Note:

- this phase changes the meaning of invitation validation from "unused invitation redeem code exists" to "a user with that permanent invite code exists"

### 3. OAuth First Registration

`LoginOrRegisterOAuthWithTokenPair(...)` follows the same rule:

- existing user login ignores invite code
- first-time user registration accepts optional invite code
- successful first registration writes the same permanent inviter binding

### 4. Qualifying Recharge Settlement

When a user redeems a qualifying paid `balance` redeem code:

1. redeem succeeds
2. base recharge amount is credited to the user
3. if the user has `invited_by_user_id`, calculate:
   - inviter reward = recharge amount * 5%
   - invitee reward = recharge amount * 5%
4. credit those rewards
5. write separate `invite_reward_records`
6. run campaign qualification / settlement if any active campaign applies

## Reward Settlement Model

### Base Rewards

Base rewards are always-on system behavior:

- inviter receives `5%`
- invitee receives `5%`
- triggered only by qualifying paid recharge events

Recommended operational behavior:

- base rewards can settle immediately because they already require a real recharge event

### Campaign Rewards

Campaign rewards are extra overlays:

- they do not change invite bindings
- they do not replace base rewards
- they are tracked separately from base rewards

Phase-one allowed campaign patterns:

1. `first_recharge_bonus`
   - trigger: invitee's first qualifying recharge event
   - reward: fixed or configured extra bonus for inviter and/or invitee

2. `recharge_amount_leaderboard`
   - ranking metric: qualifying recharge amount generated by invitees during campaign window
   - payout: fixed rewards after campaign close

Explicitly disallowed in phase one:

- registration-count leaderboard
- registration-count milestones
- binding-only rewards
- invite-N-users campaigns without recharge requirements

## API Design

### Existing Endpoints to Reuse

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/validate-invitation-code`
- OAuth completion path for first-time registration

Behavior changes:

- validation resolves permanent invite codes instead of unused invitation redeem codes
- registration binds inviter relation instead of consuming invitation inventory

### New User-Facing Referral Endpoints

Recommended new endpoints:

- `GET /api/v1/invite/summary`
  - returns invite code, invite link, totals, active campaign summary

- `GET /api/v1/invite/rewards`
  - paginated reward history for current user

Optional future endpoints:

- `GET /api/v1/invite/campaigns/active`
- `GET /api/v1/invite/leaderboard`

### Recommended Summary Response Fields

Suggested fields for `/api/v1/invite/summary`:

- `invite_code`
- `invite_link`
- `invited_users_total`
- `qualified_recharge_users_total`
- `invitees_total_recharge_amount`
- `base_rewards_total`
- `campaign_rewards_total`
- `active_campaigns`

## Frontend Design

### Registration Page

Keep the existing invitation field on the registration page.

Enhancements:

- support `?invite=<code>` prefill
- use the existing public validation endpoint
- copy should explain the code as a referral code, not a one-time invitation inventory item

### Invite Center

Add a dedicated invite center page instead of burying all data inside profile:

- my invite code
- copy invite link
- total referred recharge amount
- total earned rewards
- current active campaign cards
- recent reward history

### Redemption Feedback

When recharge triggers an invitee reward, redemption success UI should clearly state:

- base recharge amount
- extra invite reward amount if granted

This gives users immediate feedback that the referral system is working.

## Admin / Operations Design

Phase-one admin capabilities:

- list referral relationships
- inspect user-level referral stats
- list invite reward records
- create and manage campaigns
- inspect campaign performance

Required campaign configuration controls:

- campaign name
- campaign type
- start / end time
- qualifying minimum recharge threshold if applicable
- reward configuration
- source-type qualification rules

Recommended admin visibility:

- inviter
- invitee
- triggering recharge amount
- reward amount
- reward type
- campaign linkage
- timestamps

## Anti-Abuse and Safety

This phase does not attempt to solve all registration abuse globally. Instead it hardens the referral and campaign surfaces where abuse matters most.

### Existing Controls to Reuse

Current project capabilities already include:

- Turnstile on registration
- email verification flow
- server-side registration rate limiting
- invite-code validation rate limiting
- optional email domain policy

These remain part of the baseline.

### Referral-Specific Abuse Prevention

The design intentionally avoids activity forms that are easy to game through raw registration volume:

- no registration-count rewards
- no binding-only rewards
- no registration leaderboards

As a result, a registration bot by itself does not produce referral reward value. Value is created only by qualifying paid recharge events.

### Campaign Safety Rules

Hard campaign rules:

- campaigns must never reward registration count
- campaigns must always derive from qualifying recharge outcomes
- free benefit codes must not count as campaign-eligible recharge
- leaderboard rewards should be fixed payouts after campaign close

This preserves room for fun future activities without reintroducing easy abuse vectors.

## Transactions and Idempotency

### Settlement Requirements

The settlement path must prevent duplicate reward issuance when:

- redeem requests retry
- external integrations retry
- processes crash mid-flow
- the same qualifying event is reprocessed

### Recommended Transaction Boundaries

For a qualifying redeem event:

1. mark redeem code as used
2. apply base recharge amount
3. create referral reward records
4. apply reward balances
5. commit

If campaign logic runs synchronously for the same event, it should either:

- be included in the same transaction when bounded and predictable
- or enqueue a post-commit settlement task keyed by the redeem event

The first phase should keep campaign types simple enough that synchronous settlement is still practical for first-recharge bonus, while leaderboard ranking can be computed later from ledger data.

### Uniqueness Rules

Enforce database uniqueness aligned with reward semantics:

- one base inviter reward per triggering redeem event
- one base invitee reward per triggering redeem event
- one campaign bonus per triggering event / campaign / role
- one leaderboard payout per campaign / target user

## Error Handling

### Registration

Possible user-facing validation outcomes:

- invalid invite code
- self-invite attempt
- registration closed
- email verification missing or failed
- Turnstile verification failed

### Settlement

Reward settlement failures should be treated carefully:

- recharge redemption itself must remain atomic and correct
- referral settlement must not silently duplicate on retry
- if reward application fails, a clear operational signal must exist for reconciliation

Recommended implementation behavior:

- base recharge and referral reward records belong in a reliable transactional path
- avoid "log and ignore" for referral settlement

## Testing Strategy

Required coverage for implementation:

### Registration Tests

- register with valid invite code binds inviter
- register without invite code does not bind
- self-invite rejected
- rebinding impossible
- OAuth first registration respects same rules

### Qualification Tests

- `balance + commercial` redeem triggers settlement
- `balance + benefit` redeem does not count for campaign
- non-balance redeem never triggers referral reward
- admin balance adjustment never triggers referral reward

### Settlement Tests

- inviter gets 5%
- invitee gets 5%
- both ledgers are written
- duplicate processing does not double reward

### Campaign Tests

- first recharge bonus only triggers once per invitee
- recharge amount leaderboard counts qualifying recharge amount only
- registration count is never used as a metric

### API / UI Contract Tests

- `/auth/validate-invitation-code` semantics updated correctly
- invite summary response shape
- invite reward history pagination

## Rollout Plan

Phase one scope:

- permanent invite codes
- permanent inviter binding
- base `5% + 5%` referral settlement
- qualifying recharge metadata for redeem codes
- invite center summary + reward history
- campaign infrastructure
- first-recharge bonus campaign
- recharge-amount leaderboard campaign

Deferred work:

- sign-in campaign mechanics
- more advanced anti-abuse heuristics
- manual review workflows
- richer campaign combinators
- multi-level referral

## Implementation Notes for Planning

Recommended implementation order:

1. data model and migrations
2. registration and invite validation switch-over
3. redeem settlement integration for base rewards
4. user-facing invite APIs
5. invite center UI
6. campaign infrastructure
7. first-recharge bonus
8. recharge-amount leaderboard
9. admin campaign UI and reporting

## Final Recommendation

Implement a single-layer permanent referral system with:

- permanent per-user invite code
- permanent registration-time inviter binding
- always-on `5% + 5%` rewards for qualifying paid `balance` recharges
- campaign layer restricted to qualifying recharge outcomes only

This gives the product a strong growth mechanism without turning it into a fragile distributor system and keeps future event design safe from registration-count abuse.
