# OpenAI OAuth Expired Recovery Design

## Problem

The current invalid-account cleanup flow is log-driven and runs hourly. It is good at identifying permanently invalid `openai/oauth` accounts from repeated auth failures, but it does not cover a different class of accounts:

- accounts that were auto-paused only because `accounts.expires_at <= now()`
- accounts that still pass live admin testing and can still serve requests
- accounts that should be returned to scheduling instead of staying paused indefinitely

In the current backend behavior, scheduler exclusion is based on account-level expiry metadata:

- `auto_pause_on_expired = true`
- `expires_at is not null`
- `expires_at <= now()`

When that state is reached, the repository auto-pause job sets `schedulable = false`. The system does not have a matching recovery job for accounts that were paused by stale or misleading expiry data.

## Decision

Add a separate daily recovery flow for falsely expired `openai/oauth` accounts.

- Keep the existing hourly log-based cleanup unchanged.
- Add a new standalone script that scans database state instead of logs.
- Add a separate daily `systemd` timer for that script.
- Candidate accounts are verified with the existing admin account test API before any state change.
- Verified-good accounts are restored by setting `expires_at = null` and `schedulable = true`.
- Verified-bad accounts that clearly match the existing `401/402/403` soft-delete rules are soft-deleted immediately.
- `429` or ambiguous failures are not restored and not deleted.

This keeps responsibilities separate:

- hourly job: remove permanently invalid accounts from live error signals
- daily job: recover accounts paused only because of stale expiry state

## Scope

The change is limited to deploy-time automation and does not modify the request path or scheduler core rules.

Expected changes:

- add `scripts/recover_expired_accounts.py`
- add `scripts/recover-expired-accounts`
- add `scripts/recover-expired-accounts.env`
- add `/etc/systemd/system/sub2api-recover-expired-accounts.service`
- add `/etc/systemd/system/sub2api-recover-expired-accounts.timer`

Out of scope:

- changing backend `IsSchedulable()` expiry semantics
- changing the existing hourly log cleanup timer
- recovering already soft-deleted accounts
- introducing group-level recovery rules

## Candidate Selection

The recovery script should query only undeleted `openai/oauth` accounts that satisfy all of the following:

- `deleted_at is null`
- `platform = 'openai'`
- `type = 'oauth'`
- `status = 'active'`
- `schedulable = false`
- `auto_pause_on_expired = true`
- `expires_at is not null`
- `expires_at <= now()`

Optional CLI filters may further narrow the set:

- one or more `--account-id`
- `--limit`

This script does not filter by group. Recovery remains account-level, so any restored or deleted account affects all groups bound to that account.

## Verification Flow

Each candidate account is actively tested through the existing admin API:

- endpoint: `POST /api/v1/admin/accounts/{id}/test`
- auth: `x-api-key` preferred, JWT fallback only if needed
- default model: `gpt-5-codex-mini`
- default prompt: `hi`

The script must process candidates one by one and log each stage:

- candidate discovered
- test result classification
- final action taken

No account is restored or deleted without a live verification result from the admin test API.

## Classification And Actions

### Test Success

If the admin test succeeds:

- set `expires_at = null`
- set `schedulable = true`
- set `updated_at = now()`
- append an audit note to `notes`

`auto_pause_on_expired` remains unchanged. Future imports or legitimate expirations should still be eligible for auto-pause.

### Confirmed Permanent Failure

If verification clearly matches the existing cleanup rules:

- `401`
- `402`
- `403`

then the account is soft-deleted using the same audit-note style and deletion semantics already used by `clear_accounts.py`.

This prevents permanently invalid accounts from staying in the expired-paused pool forever.

### Rate Limited

If verification returns `429`, `usage_limit_reached`, or `rate_limit_exceeded`:

- do not restore
- do not soft-delete
- leave the account unchanged

These accounts are not permanently invalid. They may become usable again on a later run.

### Ambiguous Failure

If verification returns timeout, transport error, `5xx`, malformed output, or an unclassified error:

- do not restore
- do not soft-delete
- leave the account unchanged

This keeps the recovery job conservative and avoids accidental deletion from transient failures.

## Scheduler Refresh

The recovery script should avoid emitting one scheduler outbox event per account.

Instead:

- track whether at least one restore or soft-delete happened in the current run
- enqueue exactly one `scheduler_outbox` `full_rebuild` event at the end of the run if any data changed

This keeps the scheduler cache consistent without flooding the outbox.

## CLI Shape

The new script should support:

- `--action recover|report`
- `--dry-run`
- `--account-id` (repeatable)
- `--limit`
- `--model`
- `--prompt`
- `--curl-timeout`
- `--admin-api-key`
- `--jwt-ttl-seconds`
- `--postgres-container`
- `--sub2api-container`
- `--admin-base-url`
- `--database`
- `--db-user`

Behavior:

- `report`: verify only, print classification and proposed action
- `recover`: apply restore or soft-delete actions
- `--dry-run`: never write, regardless of action

## Logging

The script should follow the same line-oriented operational style as the existing cleanup scripts.

Per-account lines:

- `candidate account=... email=... expires_at=...`
- `tested account=... email=... result=... action=... http_status=... events=...`
- `recovered account=... email=...`
- `deleted account=... email=...`
- `unchanged account=... email=... reason=...`

Final summary line:

```text
summary candidates=... verified=... recovered=... deleted=... rate_limited=... unchanged=... errors=... action=... dry_run=... model=... source=expired_recovery
```

## Service And Timer

Introduce a separate daily service and timer:

- service: `sub2api-recover-expired-accounts.service`
- timer: `sub2api-recover-expired-accounts.timer`

Execution model:

- service runs as `root`
- service uses an env file containing `SUB2API_ADMIN_API_KEY`
- timer triggers every 24 hours

The timer cadence is intentionally lower than the hourly invalid-account cleanup because expiry recovery is database-state remediation, not live incident response.

## Safety Rules

- Only undeleted `openai/oauth` accounts are eligible.
- No account is restored without a successful live admin test.
- No account is soft-deleted unless the test clearly classifies it as the existing `401/402/403` invalid state.
- `429` and ambiguous failures are preserved unchanged.
- The script does not mutate group bindings.
- The script does not touch already soft-deleted accounts.
- The script emits at most one scheduler full rebuild event per run.

## Verification

Recommended rollout sequence:

1. Run a small sample in report mode.
2. Inspect classifications and candidate counts.
3. Run recover mode on a limited sample.
4. Validate database counts and scheduler visibility.
5. Enable the daily timer.

Minimum verification commands:

- report mode on a small limit
- recover mode on a small limit
- SQL checks for:
  - remaining expired-paused candidate count
  - number restored in the sample
  - number soft-deleted in the sample
- `systemctl status` for the new service and timer

## Risks

- If the admin test model or prompt becomes unstable, ambiguous failures may rise and reduce recovery effectiveness.
- If `accounts.expires_at` is intentionally used for a subset of accounts as a hard business expiry, this job will override that behavior for accounts that still pass live testing.
- If many accounts qualify in one run, verification can take time. The daily cadence limits operational impact.

These risks are acceptable because the script is conservative, verifies accounts live, and only restores accounts that demonstrate actual usability.
