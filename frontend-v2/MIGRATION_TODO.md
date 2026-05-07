# frontend-v2 — migration TODO

Phase 1 (this PR) ships the spine: Landing, auth, user console core (Dashboard/Keys/Usage/Profile), admin shell with one populated view (Users). Everything below still lives in the original Vue `frontend/` and needs to be ported.

Backend stays untouched — these are all UI-only ports. API modules already exist for most of these (see `frontend/src/api/`).

## User console — remaining views

- [ ] `KeyUsageView` (public per-key usage page) — `/key-usage`
- [ ] `ModelHubView` — `/models`
- [ ] `RedeemView` (full flow with redeem history)
- [ ] `SubscriptionsView` — `/subscriptions`
- [ ] `PurchaseSubscriptionView` — `/purchase-subscription`
- [ ] `SoraView` — `/sora`
- [ ] `CustomPageView` (admin-defined custom menu items)

## Auth — remaining flows

- [ ] 2FA / TOTP login flow (Login currently bails with a placeholder)
- [ ] Email verify view (`/email-verify`)
- [ ] OAuth callback (`/auth/callback`)
- [ ] Linux.do OAuth callback (`/auth/linuxdo/callback`) + `LinuxDoOAuthSection`
- [ ] Reset password page (`/reset-password`)
- [ ] Turnstile widget integration

## Admin — remaining views (each is a full page)

- [ ] `AccountsView` — upstream account management (large)
- [ ] `AnnouncementsView`
- [ ] `BackupView`
- [ ] `DataManagementView`
- [ ] `GroupsView`
- [ ] `OpsDashboard` + 18 sub-components (alerts, error logs, latency/throughput charts, etc.)
- [ ] `PromoCodesView`
- [ ] `ProxiesView`
- [ ] `RedeemView` (admin-side generation)
- [ ] `SettingsView`
- [ ] `SubscriptionsView` (admin)
- [ ] `UsageView` (admin)

## Setup wizard

- [ ] `SetupWizardView` (`/setup`) — first-run DB/Redis/admin config

## Cross-cutting capabilities not yet wired

- [ ] Charts (`vue-chartjs` → `recharts` or `chart.js` + react wrapper)
- [ ] Markdown rendering (`marked` + `dompurify`)
- [ ] CSV/XLSX export (`xlsx`)
- [ ] Virtualized tables (`@tanstack/react-virtual`)
- [ ] Drag & drop (`vue-draggable-plus` → `@dnd-kit/core`)
- [ ] QR codes (`qrcode`)
- [ ] Onboarding tour (`driver.js`)
- [ ] Announcement popup system + read tracking
- [ ] Theme (light/dark) toggle — currently always light Plato
- [ ] Document title resolver (`router/title.ts` equivalent)
- [ ] User custom menu items in sidebar
- [ ] Sticky-session, ops-monitoring-disabled handling

## API modules to port

The new `src/api/*` only contains the slice used by Phase 1. Still needed:

- `groups.ts`, `subscriptions.ts`, `announcements.ts`, `redeem.ts`, `sora.ts`, `totp.ts`, `setup.ts`, plus the full `admin/*` tree (proxies, accounts, ops, etc.)

Lifting from `frontend/src/api/` is mostly mechanical — change Pinia/`type` imports to point to the v2 paths.
