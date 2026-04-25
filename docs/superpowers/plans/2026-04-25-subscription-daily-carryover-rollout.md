# Subscription Daily Carryover Rollout Note

## Purpose

This note captures the operational rollout guidance for subscription daily carryover and the one-off daily compensation reset.

## Product Semantics

- Carryover applies to unused subscription daily quota, not to account balance.
- Only one day of carryover is allowed.
- Carryover is consumed before the current day's fresh quota.
- Unused carryover expires at the end of the current day and never rolls into a third day.
- Carryover is only generated when the subscription is still active on the next day.

## Rolling Deployment Guidance

- The schema change is additive only: two non-null columns with default `0`.
- Old pods do not understand the new daily carryover semantics.
- During a rolling deployment, old and new pods can coexist for a short period.
- Because of that, treat the feature as fully effective only after all old pods are drained.

## Compensation Reset

The one-off compensation reset is an admin-only operational tool. It is not intended to become a normal user-facing product feature.

Endpoint:

- `POST /api/v1/admin/subscriptions/reset-daily`

Request body:

- `{}` resets the daily state for all active subscriptions.
- `{"group_id": 88}` resets only active subscriptions in one group.

Response shape:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "reset_count": 123,
    "window_start": "2026-04-25T00:00:00Z"
  }
}
```

Effect:

- Resets daily usage for the targeted active subscriptions.
- Clears daily carryover state for the targeted active subscriptions.
- Invalidates subscription read caches so the next request reloads authoritative state.

## Recommended Rollout Order

1. Deploy the backend version that includes the additive migration and carryover logic.
2. Deploy the frontend that explains carryover-aware totals and copy.
3. Wait until all old backend pods are drained from the rolling deployment.
4. If operationally needed, run `POST /api/v1/admin/subscriptions/reset-daily`.
5. Wait roughly 10 to 15 seconds for per-process L1 subscription caches to settle naturally.
6. Spot-check several subscription users and one target group in the admin UI.

## When To Run Compensation

Use the reset only if you want a clean operational boundary after rollout or if you need to compensate for potentially confusing same-day state during the transition window.

Do not run the reset before old pods are gone. Doing that early weakens the whole point of the reset because old pods can still continue operating with the old daily semantics.

## Residual Risks

- The existing system is still optimistic at request time: eligibility is checked before billing write completion.
- During rollout, a request that lands on an old pod will still use the old behavior until that pod is drained.
- Subscription L1 caches are process-local, so invalidation is strongest on the writing pod and converges on other pods shortly after reload.

## Example Commands

Reset all active subscriptions:

```bash
curl -X POST \
  http://<host>/api/v1/admin/subscriptions/reset-daily \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <admin-token>' \
  -d '{}'
```

Reset one group:

```bash
curl -X POST \
  http://<host>/api/v1/admin/subscriptions/reset-daily \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <admin-token>' \
  -d '{"group_id":88}'
```
