# Frontend V2 XlabAPI Parity Design

**Goal:** Make the test-environment React `frontend-v2` expose the same functional entry points as the existing xlabapi frontend, while keeping production `/root/sub2api-deploy` and port `8081` untouched.

**Scope Guardrails**

- Only operate under `/root/sub2api-src`, `/root/test_xlab`, and the `sub2api-test-xlab` container.
- Develop from the current `test/xlabapi` branch and deploy only to the test image used by `/root/test_xlab/docker-compose.override.yml`.
- Do not edit `/root/sub2api-deploy`, do not restart or reconfigure the production `sub2api` container, and do not use port `8081` for validation.

## Findings

The old Vue frontend has a wider surface than React `frontend-v2`. The main gaps are not only page implementation depth; several old routes currently 404 in `frontend-v2`, and several existing API clients call paths that do not match the Go router.

### P0 Route And Navigation Parity

User routes to restore:

- `/available-channels`
- `/monitor`
- `/purchase`
- `/orders`
- `/affiliate`
- `/invite -> /affiliate`
- `/custom/:id`
- `/payment/qrcode`
- `/payment/result`
- `/payment/stripe`
- `/payment/stripe-popup`
- `/auth/wechat/callback`
- `/auth/wechat/payment/callback`
- `/auth/oidc/callback`
- `/oauth/consent`

Admin routes to restore:

- `/admin/dashboard`
- `/admin -> /admin/dashboard`
- `/admin/ops`
- `/admin/channels -> /admin/channels/pricing`
- `/admin/channels/pricing`
- `/admin/channels/monitor`
- `/admin/proxies`
- `/admin/subscription-products -> /admin/subscriptions`
- `/admin/subscription-product-config`
- `/admin/invites -> /admin/users`
- `/admin/affiliates -> /admin/affiliates/invites`
- `/admin/affiliates/invites`
- `/admin/affiliates/rebates`
- `/admin/affiliates/transfers`
- `/admin/orders/dashboard`
- `/admin/orders`
- `/admin/orders/plans`
- `/admin/backup` in admin navigation

Public route to restore:

- `/docs`, because the landing page links to it.
- `/setup`, the legacy setup wizard route.
- `/key-usage`, the public API key usage query route.

Compatibility user route to restore:

- `/image-studio`, the legacy Sora/image creation entry.

### P0 API Contract Fixes

Existing `frontend-v2` clients should be aligned to the backend router before adding deeper UI:

- User usage: `/usage`, `/usage/stats`, `/usage/dashboard/stats`, `/usage/dashboard/trend`, `/usage/dashboard/models`.
- Admin usage: `/admin/usage`, `/admin/usage/stats`.
- Admin dashboard: `/admin/dashboard/stats`.
- Backup: `/admin/backups`, `/admin/backups/s3-config`, `/admin/backups/s3-config/test`, `/admin/backups/schedule`, `/admin/backups/:id/download-url`.
- Settings SMTP: `/admin/settings/test-smtp`, `/admin/settings/send-test-email`.
- Accounts schedulable and refresh: `POST /admin/accounts/:id/schedulable`, `POST /admin/accounts/:id/refresh`.
- Admin subscriptions revoke: `DELETE /admin/subscriptions/:id`.
- User visible groups: `/groups/available`, `/groups/rates`.

### P1 Functional Closures

After entry parity, the first real business closures should be:

- Auth/OAuth: WeChat callback, OIDC callback, pending OAuth create/bind, Xlab OAuth consent.
- Payment: checkout, create order, QR/status, result recovery, orders/refunds.
- User profile: TOTP, identity binding, notification email, avatar/account binding cards.
- User commercial pages: affiliate, available channels, channel monitor, subscriptions balance fallback.
- Admin operations: ops dashboard, channels, channel monitor, proxies, payment orders/plans, affiliate records, subscription products/config.

### P2 Experience And Compatibility

- Feature-flagged sidebar items based on public/admin settings.
- Custom menu iframe behavior for `/custom/:id`.
- Column preferences, auto-refresh, export, cleanup tasks, graph distributions.
- Setup wizard, key usage public page, and image studio are restored as route-compatible placeholders in the first batch, then migrated fully in later phases.

## Architecture

Use a route-matrix-driven migration. First, every old route gets a React route. Pages that are not fully migrated yet use a shared parity placeholder that names the missing old function and exposes the expected API/client surface rather than pretending to be complete. This makes test-environment QA possible without hiding gaps.

API clients should be thin and path-accurate. A client may be added before a full page is built if it documents the backend contract and lets future pages use the right endpoint from the start.

The first deployable batch should:

- Add the missing route and navigation entries.
- Fix existing P0 API path mismatches.
- Add minimal clients for payment, channels, channel monitor, affiliate, OAuth consent, and admin missing domains.
- Keep visual style consistent with current `frontend-v2` console components.

## Verification

Verification for the first batch:

- `npm run build` inside `/root/sub2api-src/frontend-v2`.
- Build the test image from `/root/sub2api-src` using the existing test deployment flow.
- Update only `/root/test_xlab/docker-compose.override.yml` so `sub2api-test-xlab` uses the new image.
- Recreate only the test container.
- Validate `http://127.0.0.1:11454/health`.
- Spot-check representative routes on port `11454`: `/login`, `/setup`, `/key-usage`, `/dashboard`, `/image-studio`, `/available-channels`, `/purchase`, `/admin/dashboard`, `/admin/ops`, `/admin/channels/pricing`, `/admin/orders`.
