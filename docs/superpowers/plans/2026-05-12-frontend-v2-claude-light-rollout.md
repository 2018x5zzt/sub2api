# frontend-v2 Claude-light Rollout Notes

Date: 2026-05-12
Branch target: `test/xlabapi`
Scope: visual refresh, default build path migration to `frontend-v2`, and parity risk tracking.

## What Changed In This Slice

- Moved default root Dockerfile, deploy Dockerfile, Makefile, and `deploy.sh` frontend build path from legacy Vue `frontend` to React `frontend-v2`.
- Kept legacy `frontend/` source in the repo as the production parity reference; it is no longer the default embedded frontend build target.
- Reworked `frontend-v2` design tokens to Claude-style light palette while keeping old Tailwind class names as aliases to reduce churn.
- Updated global primitives (`.btn`, `.input`, `.card`, `.badge`, `.pill-nav`, `.data-table`) to use warm canvas, white surfaces, deep ink, and accessible orange accent text.
- Updated shared landing primitives (`PillBtn`, `Wordmark`, `SectionFrame`, `PlasmaBlob`) so landing/auth/dashboard can share the light visual system.
- Converted the routed landing page away from dark full-page backgrounds and plasma glow decoration without changing routes or i18n keys.

## Explicit Non-Goals

- This slice does not claim production functional parity for `frontend-v2`.
- This slice does not rewrite auth callback flows, route guards, sidebar feature gating, setup wizard, public key usage, custom pages, or subscription product config.
- This slice does not remove the legacy Vue `frontend/` directory because it is still the reference baseline for parity work.

## P0 Parity Blockers Found During Audit

`frontend-v2` cannot be considered production-equivalent until these are closed:

1. WeChat/OIDC/payment OAuth callbacks still route to a generic callback page instead of production-equivalent automatic handling.
2. Login does not expose all production OAuth/Turnstile paths.
3. Route guards lack production rules for backend mode, payment enabled state, simple mode, and some pending auth routes.
4. Sidebar navigation is static and does not apply production feature flags, simple-mode hiding, or custom menu items.
5. `/setup`, `/key-usage`, `/custom/:id`, and `/admin/subscription-product-config` still use `ParityPlaceholder`.

## Validation Gate For This Slice

Run from repo root:

```sh
npm --prefix frontend-v2 run typecheck
npm --prefix frontend-v2 run build
```

Before any production rollout, add a separate parity slice for the P0 blockers above and verify against `origin/xlabapi:frontend` behavior.
