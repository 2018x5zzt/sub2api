# Invite Admin Rollout Verification Design

## Context

The backend invite system now spans two migrations:

- `081_add_invite_growth_foundation.sql` adds invite binding fields on `users`, `source_type` on `redeem_codes`, and the `invite_reward_records` table.
- `082_add_invite_admin_ops.sql` adds `invite_relationship_events`, backfills `register_bind` events from existing `users.invited_by_user_id` data, introduces `invite_admin_actions`, and extends `invite_reward_records` with optional `admin_action_id` plus nullable `trigger_redeem_code_id`.

Recent validation focused on the user-facing and backend semantic risk, not administrator usability. The key conclusion was that existing invite bindings and historical base rewards should not drift passively when the admin invite pages are introduced, unless an administrator explicitly performs actions such as rebind or recompute.

What is still missing is an operational verification artifact that can be run before and after rollout to confirm that conclusion against production data.

## Problem

The rollout needs a low-risk, reusable verification artifact that answers the following questions without modifying data:

1. Did the number of already-bound users drift unexpectedly?
2. Does the `register_bind` event backfill align with existing `users.invited_by_user_id` data?
3. Did historical `base_invite_reward` rows or totals change unexpectedly?
4. If reward rows changed, can that change be attributed to explicit admin actions rather than passive system drift?

The artifact must work even in environments where there is no dedicated admin verification tool, no API wrapper, and no shell helper around the database.

## Goals

- Provide a single read-only SQL checklist that can be run manually before and after deployment.
- Produce fixed, stable output blocks so operators can compare snapshots reliably.
- Separate passive-drift signals from expected admin-triggered changes.
- Surface a small amount of anomaly detail for quick investigation.

## Non-Goals

- No automatic diffing between pre-deploy and post-deploy snapshots.
- No repair SQL and no mutation of production data.
- No backend API, CLI command, or admin page for this verification.
- No attempt to reconstruct full business causality across all invite history.
- No attempt to distinguish normal new business growth from historical drift without operator judgment about the deployment window.

## Chosen Approach

Use a single SQL file stored in the repository as an ops artifact, not as a migration.

Why this approach:

- It is the lowest-risk option because it is read-only and introduces no runtime dependency.
- It can be executed in any environment that already has database access.
- It keeps the rollout check focused on semantic correctness rather than tool-building.

Alternatives considered and rejected:

- Split SQL into separate summary and anomaly files. Rejected because operators may run only part of the checklist.
- Add shell or `psql` wrapper logic. Rejected because the selected deliverable is intentionally SQL-only and should avoid executor-specific behavior.
- Build an API or CLI entrypoint. Rejected because it expands implementation scope and execution surface beyond rollout verification needs.

## Deliverable

Add one SQL file under an ops-oriented resource path:

- `backend/resources/sql/ops/invite_admin_rollout_verification.sql`

This file will contain only comments plus `WITH` and `SELECT` statements. It must not contain any `INSERT`, `UPDATE`, `DELETE`, `ALTER`, or transaction control statements.

## SQL Output Design

The SQL file will emit four result groups in a fixed order.

### 1. Fixed Metric Overview

This block returns a two-column summary table:

- `metric_name`
- `metric_value`

The block includes these fixed metrics:

- `bound_users_total`
- `register_bind_event_rows_total`
- `register_bind_distinct_invitees_total`
- `base_invite_reward_rows_total`
- `base_invite_reward_amount_total`
- `admin_rebind_event_rows_total`
- `manual_invite_grant_rows_total`
- `recompute_delta_rows_total`

Rationale:

- The first five metrics expose the core rollout safety signals.
- The last three metrics separate admin-triggered changes from passive system drift.
- A row-based metric table is preferred over a wide single row because it remains stable when copied into release notes or compared manually.

### 2. Binding Alignment Checks

This block answers whether already-bound users and `register_bind` events still line up.

The SQL derives two logical datasets:

- `bound_users`
  - source: `users`
  - filter: `users.invited_by_user_id IS NOT NULL`
  - columns:
    - `invitee_user_id`
    - `current_inviter_user_id`
    - `expected_effective_at = COALESCE(users.invite_bound_at, users.created_at)`
- `register_bind_events`
  - source: `invite_relationship_events`
  - filter: `event_type = 'register_bind'`
  - columns:
    - `invitee_user_id`
    - `event_inviter_user_id = new_inviter_user_id`
    - `effective_at`
    - `created_at`

The block outputs these fixed metrics:

- `bound_users_missing_register_bind_total`
- `register_bind_without_bound_user_total`
- `register_bind_duplicate_invitee_total`
- `register_bind_inviter_mismatch_total`
- `register_bind_effective_at_mismatch_total`

The block then outputs three anomaly samples, each capped at 50 rows:

- missing event samples
  - `invitee_user_id`
  - `current_inviter_user_id`
  - `expected_effective_at`
- duplicate `register_bind` samples
  - `invitee_user_id`
  - `register_bind_count`
  - `first_effective_at`
  - `last_effective_at`
- inviter or timestamp mismatch samples
  - `invitee_user_id`
  - `current_inviter_user_id`
  - `event_inviter_user_id`
  - `expected_effective_at`
  - `event_effective_at`

Interpretation:

- If no administrator explicitly runs invite admin write operations during the comparison window, these alignment metrics should not become worse after rollout.
- Any increase in mismatch-style metrics should be treated as a semantic drift signal and investigated before treating the rollout as safe.

### 3. Reward Attribution and Integrity Checks

This block separates natural base rewards from admin-triggered reward adjustments.

It first outputs a grouped summary table with these columns:

- `reward_type`
- `status`
- `rows_total`
- `reward_amount_total`
- `distinct_invitees_total`
- `distinct_reward_targets_total`
- `rows_with_admin_action_total`
- `rows_with_null_trigger_code_total`

The expected reward types of interest are:

- `base_invite_reward`
- `manual_invite_grant`
- `recompute_delta`

It then outputs these fixed anomaly metrics:

- `base_reward_with_admin_action_total`
- `manual_grant_without_admin_action_total`
- `recompute_delta_without_admin_action_total`
- `base_reward_without_trigger_code_total`

Finally, it outputs two anomaly samples, each capped at 50 rows:

- reward attribution anomaly samples
  - `id`
  - `reward_type`
  - `reward_role`
  - `reward_amount`
  - `admin_action_id`
  - `trigger_redeem_code_id`
  - `created_at`
- base reward observation samples
  - `id`
  - `inviter_user_id`
  - `invitee_user_id`
  - `reward_target_user_id`
  - `reward_role`
  - `reward_amount`
  - `trigger_redeem_code_id`
  - `created_at`

Interpretation:

- If no administrator explicitly performs invite correction actions, the `base_invite_reward` row count and total amount should not show unexplained historical drift.
- New `manual_invite_grant` or `recompute_delta` rows are not automatically considered regressions. They should first be attributed to explicit admin actions.
- `base_invite_reward` rows should continue to look like natural redeem-triggered rewards rather than admin-linked corrections.

### 4. Execution Notes

The SQL file will begin with comments that define how operators should use the output:

- This is a pre-deploy and post-deploy manual verification checklist, not a migration.
- If operators executed `rebind_inviter`, `manual_reward_grant`, or `recompute_rewards` during the comparison window, increases in related admin metrics may be expected.
- Without explicit admin actions, binding alignment metrics should remain stable and `base_invite_reward` totals should not drift unexpectedly.

The file intentionally avoids:

- `psql` meta-commands such as `\echo` or `\pset`
- shell-specific formatting assumptions
- parameters that require templating before execution

This keeps the artifact portable across environments and suitable for direct execution with standard SQL tooling.

## Expected Operator Workflow

1. Run the SQL file before deployment and save the result snapshot.
2. Deploy the admin invite rollout.
3. Run the same SQL file after deployment.
4. Compare the fixed metric block first.
5. If any drift appears, inspect the binding alignment and reward anomaly samples.
6. Correlate any admin-related changes with actual admin actions during the deployment window.

## Risks and Mitigations

- Risk: operators may misread expected admin-triggered changes as rollout regressions.
  - Mitigation: include explicit admin metrics in the summary and call out interpretation rules in the file header comments.
- Risk: large anomaly outputs may reduce usability.
  - Mitigation: cap sample outputs at 50 rows per anomaly query.
- Risk: future schema changes may silently make the SQL stale.
  - Mitigation: bind the design to current canonical names used by migrations and service constants, and keep the SQL file in-repo so it evolves with code review.

## Out of Scope for This Design

This design stops at the SQL verification artifact. It does not yet define:

- an implementation plan for authoring and validating the SQL file
- test fixtures for snapshot verification
- release checklist automation

Those belong to the next planning step after this design is reviewed.
