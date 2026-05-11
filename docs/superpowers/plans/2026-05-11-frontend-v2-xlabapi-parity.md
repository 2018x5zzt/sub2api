# Frontend V2 XlabAPI Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore xlabapi route, navigation, and P0 API contract parity in the test-environment React `frontend-v2`.

**Architecture:** Use a route-matrix-driven migration. Missing routes receive either a minimal functional page or a clearly labeled parity placeholder. Existing API clients are corrected to the Go backend contract before deeper page work.

**Tech Stack:** React 18, React Router, TanStack Query, TypeScript, Vite, Tailwind, existing `frontend-v2` UI components, Go backend `/api/v1` router.

---

### Task 1: P0 API Contract Corrections

**Files:**
- Modify: `frontend-v2/src/api/usage.ts`
- Modify: `frontend-v2/src/api/admin.ts`
- Modify: `frontend-v2/src/api/admin/usage.ts`
- Modify: `frontend-v2/src/api/admin/backup.ts`
- Modify: `frontend-v2/src/api/admin/settings.ts`
- Modify: `frontend-v2/src/api/admin/accounts.ts`
- Modify: `frontend-v2/src/api/admin/subscriptions.ts`
- Modify: `frontend-v2/src/api/models.ts`

- [ ] Change user usage endpoints to `/usage`, `/usage/stats`, and `/usage/dashboard/*`.
- [ ] Change admin usage endpoint to `/admin/usage`.
- [ ] Change admin dashboard endpoint to `/admin/dashboard/stats`.
- [ ] Change backup endpoints from `/admin/backup` to `/admin/backups`.
- [ ] Change SMTP endpoints to `/admin/settings/test-smtp` and `/admin/settings/send-test-email`.
- [ ] Change account schedulable to `POST /admin/accounts/:id/schedulable`.
- [ ] Change account credential refresh to `POST /admin/accounts/:id/refresh`.
- [ ] Change subscription revoke to `DELETE /admin/subscriptions/:id`.
- [ ] Change user group list to `/groups/available`.
- [ ] Run `npm run build` from `frontend-v2` and fix compile errors.

### Task 2: Route And Navigation Entry Parity

**Files:**
- Modify: `frontend-v2/src/router/index.tsx`
- Modify: `frontend-v2/src/components/layout/ConsoleLayout.tsx`
- Modify: `frontend-v2/src/i18n/locales/en.ts`
- Modify: `frontend-v2/src/i18n/locales/zh.ts`
- Create: `frontend-v2/src/pages/Docs.tsx`
- Create: `frontend-v2/src/pages/ParityPlaceholder.tsx`

- [ ] Add `/docs`.
- [ ] Add public compatibility routes `/setup` and `/key-usage`.
- [ ] Add user routes `/available-channels`, `/monitor`, `/purchase`, `/orders`, `/affiliate`, `/custom/:id`.
- [ ] Add user compatibility route `/image-studio` and the corresponding sidebar entry.
- [ ] Add `/invite -> /affiliate`.
- [ ] Add payment routes `/payment/qrcode`, `/payment/result`, `/payment/stripe`, `/payment/stripe-popup`.
- [ ] Add auth routes `/auth/wechat/callback`, `/auth/wechat/payment/callback`, `/auth/oidc/callback`, `/oauth/consent`.
- [ ] Add admin routes and redirects listed in the design.
- [ ] Add missing user/admin navigation entries using lucide icons.
- [ ] Add English and Chinese nav/page labels.
- [ ] Run `npm run build` from `frontend-v2`.

### Task 3: Minimal Missing API Clients

**Files:**
- Create: `frontend-v2/src/api/payment.ts`
- Create: `frontend-v2/src/api/channels.ts`
- Create: `frontend-v2/src/api/channelMonitor.ts`
- Create: `frontend-v2/src/api/affiliate.ts`
- Create: `frontend-v2/src/api/oauth.ts`
- Create: `frontend-v2/src/api/admin/channels.ts`
- Create: `frontend-v2/src/api/admin/channelMonitor.ts`
- Create: `frontend-v2/src/api/admin/proxies.ts`
- Create: `frontend-v2/src/api/admin/payment.ts`
- Create: `frontend-v2/src/api/admin/affiliate.ts`
- Create: `frontend-v2/src/api/admin/ops.ts`

- [ ] Add thin methods for the backend endpoints without implementing full UI.
- [ ] Use broad but type-safe enough response types where DTOs are not yet modeled.
- [ ] Export API objects consistently with existing `frontend-v2` style.
- [ ] Run `npm run build` from `frontend-v2`.

### Task 4: Test Deployment

**Files:**
- Modify only if needed: `/root/test_xlab/docker-compose.override.yml`

- [ ] Build the new test image from `/root/sub2api-src`.
- [ ] Update `/root/test_xlab/docker-compose.override.yml` so only the test service uses the new image.
- [ ] Recreate only `sub2api-test-xlab`.
- [ ] Verify `curl -fsS http://127.0.0.1:11454/health`.
- [ ] Spot-check representative frontend routes on `http://127.0.0.1:11454`.
