# Enterprise Visible Groups Design

## Problem

`enterprise-bff` currently derives enterprise-visible groups from the shared user visibility rules. That leaks ordinary public groups into the enterprise console, which is the opposite of the required behavior.

The product requirement is narrower:

- normal users keep the existing shared visibility semantics
- enterprise users are filtered only inside `enterprise-bff`
- enterprise users should see only groups explicitly allowed for their enterprise

## Decision

Add an `enterprise-bff` configuration mapping from normalized enterprise name to visible group IDs.

- Config lives only in `enterprise-bff`
- Visibility is default-deny for enterprise users
- Shared core rules such as public, exclusive, and subscription do not control enterprise console visibility
- Only active groups that appear in the configured enterprise mapping are returned

## Shape

Introduce an environment variable:

- `ENTERPRISE_BFF_VISIBLE_GROUP_IDS_BY_ENTERPRISE`

Expected JSON shape:

```json
{
  "bustest": [2, 9, 11, 16, 18, 19]
}
```

Keys are normalized with the same company-name normalization already used by enterprise auth.

## Scope

Only these areas change:

- `backend/internal/enterprisebff/config.go`
- `backend/internal/enterprisebff/enterprise.go`
- enterprise BFF tests

No database migration, no core user visibility change, no ordinary frontend behavior change.

## Safety Rules

- Missing enterprise mapping returns zero visible groups
- Unknown or inactive group IDs are ignored
- Ordering still follows group `sort_order`, then `id`
- Enterprise auth and key authorization keep using the same enterprise identity checks

## Verification

Focused verification is sufficient for this change:

- enterprise visibility unit tests
- enterprise pool-status and key-authorization tests that depend on visible groups
