# Affiliate-Only Invitation Retirement Design

## Goal

Retire xlabapi's legacy invitation system and remove registration-gate semantics. The product should keep only the upstream-compatible affiliate rebate model:

- Users may register without any invitation or access code.
- A referral code is optional and only establishes an affiliate relationship.
- Affiliate relationships and rebates live in `user_affiliates` and `user_affiliate_ledger`.
- Legacy invite reward/admin/rebind flows stop being active production paths.

This design intentionally chooses a compatibility window: the backend may accept legacy `invitation_code` request fields as a deprecated alias for `aff_code`, but the frontend must stop sending or displaying `invitation_code`.

## Non-Goals

- Do not require invitation/access codes for registration.
- Do not preserve old invite reward issuance as an active business path.
- Do not keep `/invite` as a separate user-facing product surface.
- Do not drop old invite tables, generated Ent code, or historical data in the first implementation phase.
- Do not remove affiliate rebate features, admin affiliate controls, or affiliate transfer-to-balance behavior.
- Do not couple this work to channel management, payment-channel refactors, or account-stats pricing.

## Current State

The repository currently contains three overlapping concepts:

- Legacy xlabapi invite growth:
  - `InviteService`, `InviteHandler`, invite admin actions, invite reward records, relationship events, `users.invite_code`, and `users.invited_by_user_id`.
- Upstream affiliate rebate:
  - `AffiliateService`, affiliate handlers, `user_affiliates`, `user_affiliate_ledger`, `aff_code`, affiliate-specific admin controls, and rebate settings.
- Upstream invitation-code registration gate:
  - `invitation_code_enabled`, `invitation_code`, invitation-code validation, and registration/OAuth paths that may reject users when a required invitation code is missing.

The target state keeps the second concept and retires the first and third from active product behavior.

## Product Behavior

Registration no longer has an access-code gate. Missing, unknown, expired, or malformed invitation codes must not prevent a user from registering.

Affiliate referral remains optional:

- A valid `aff_code` binds the new user to an inviter through `AffiliateService.BindInviterByCode`.
- An invalid `aff_code` does not block registration; it only leaves the new user unbound.
- A user still gets or can lazily create their own affiliate profile through `AffiliateService.EnsureUserAffiliate`.
- Rebates continue to accrue through `AffiliateService.AccrueInviteRebate` according to affiliate settings.

The public user surface should present "Affiliate Rebates", not a separate legacy invite center. Any `/invite` route should redirect to `/affiliate` or be removed from navigation.

## Backend Design

### Registration Requests

`RegisterRequest` keeps `aff_code` as the canonical field.

`invitation_code` may remain temporarily as a deprecated alias. Backend request handling resolves the affiliate code as:

```go
affiliateCode := strings.TrimSpace(req.AffCode)
if affiliateCode == "" {
	affiliateCode = strings.TrimSpace(req.InvitationCode)
}
```

The alias must be documented in code comments as backwards compatibility only. New frontend code must not send `invitation_code`.

### AuthService

`AuthService.RegisterWithVerification` must not enforce `SettingKeyInvitationCodeEnabled` and must not return `ErrInvitationCodeRequired`.

Registration flow:

1. Validate normal registration requirements such as registration-enabled, email verification, password, promo code, and Turnstile.
2. Create the user.
3. Ensure the user has an affiliate profile when `AffiliateService` is available.
4. If an affiliate code was supplied, attempt `AffiliateService.BindInviterByCode`.
5. Log affiliate binding failures but do not fail registration.

OAuth and pending OAuth registration paths follow the same rule: `aff_code` is optional, `invitation_code` is at most a deprecated alias, and missing invitation data must not block account creation.

### Validation Endpoint

`/auth/validate-invitation-code` is no longer a registration-gate API. During the compatibility window, it may remain as a wrapper that validates an affiliate code:

- Existing affiliate code: `valid: true`
- Missing or unknown affiliate code: `valid: false`
- No registration-disabled, invite-required, used-code, or retired-gate semantics

A later cleanup may introduce `/auth/validate-affiliate-code` and remove the old endpoint. That should be a separate change after clients no longer depend on the legacy route.

### Settings

`invitation_code_enabled` must stop affecting registration and OAuth registration. Existing database values can remain for compatibility, but active service code should not read the setting to decide whether registration requires a code.

Affiliate settings remain active:

- `affiliate_enabled`
- `affiliate_rebate_rate`
- `affiliate_rebate_freeze_hours`
- `affiliate_rebate_duration_days`
- `affiliate_rebate_per_invitee_cap`

Admin settings APIs and frontend forms should remove or hide invitation-code gate controls while keeping affiliate controls.

### Legacy Invite Production Paths

The first implementation phase retires active legacy invite production paths:

- Remove `InviteHandler` from active route registration or make old routes unavailable.
- Remove legacy invite user surface from navigation.
- Remove legacy admin invite reward/rebind/recompute surfaces from active navigation/API exposure.
- Ensure new registrations and new rebate accruals do not write `users.invite_code`, `users.invited_by_user_id`, `invite_reward_records`, or `invite_relationship_events`.

Old tables and Ent-generated code may remain in phase one to preserve historical data and reduce migration risk.

## Frontend Design

### Register Flow

The register page must remove invitation-code input, validation debounce, validation errors, and submit blocking tied to invitation code state.

It keeps affiliate referral capture:

- Read `?aff=` and `?aff_code=` query parameters.
- Store referral codes using the existing affiliate referral storage utility.
- Submit `aff_code` when present.
- Do not submit `invitation_code`.

Registration must be possible with no affiliate code.

### OAuth Flow

OAuth components keep optional affiliate propagation and remove invitation-gate UX:

- Continue storing and submitting `aff_code`.
- Stop requiring users to complete an invitation-code form before OAuth registration.
- If pending OAuth pages still render an invitation-code field, replace it with an optional affiliate-code field or remove it.

### Navigation

User navigation should expose affiliate rebates. Legacy invite navigation should be removed or redirect `/invite` to `/affiliate`.

Admin navigation should expose affiliate custom-code and rebate controls. Legacy invite reward/admin/rebind/recompute screens should be removed from visible navigation in phase one.

## Data Migration Strategy

Phase one is non-destructive.

The existing compatibility migration that materializes `user_affiliates` and `user_affiliate_ledger` from legacy invite data remains the bridge for historical relationships and reward history. New business behavior must write only affiliate tables.

Do not drop legacy invite columns or tables in phase one. Physical cleanup belongs to phase two after production behavior and clients are confirmed to be affiliate-only.

Phase two may remove:

- Legacy invite service and repository files
- Legacy invite handlers and routes
- Legacy invite admin services
- Legacy frontend invite API/views/i18n
- Legacy Ent schemas and generated code
- Legacy invite migrations or future migration references, where safe

## Error Handling

Invalid affiliate codes should not fail registration. They should be logged at a low severity and leave the new user unbound.

Affiliate service unavailability should not prevent registration unless the user profile creation itself fails in a way that the current repository layer treats as fatal. The default posture is registration-first, rebate-binding-best-effort.

Compatibility validation endpoints should return structured `valid: false` responses instead of hard errors for missing or unknown affiliate codes.

## Testing Plan

Backend focused tests:

- Registration succeeds when no invitation or affiliate code is supplied.
- Registration succeeds even if old `invitation_code_enabled` setting is `true`.
- Registration with valid `aff_code` binds `user_affiliates.inviter_id`.
- Registration with invalid `aff_code` succeeds and does not bind an inviter.
- Registration with deprecated `invitation_code` alias binds an inviter only when `aff_code` is empty.
- OAuth registration succeeds without invitation code.
- OAuth registration with valid `aff_code` binds inviter.
- `ValidateInvitationCode` compatibility endpoint validates affiliate code and does not enforce registration-gate semantics.
- Active route registration no longer exposes legacy invite reward/admin endpoints unless intentionally left as retired wrappers.

Frontend focused tests:

- Register page does not render invitation-code input.
- Register submit payload omits `invitation_code`.
- `?aff=` and `?aff_code=` populate affiliate referral payload.
- Register can submit without affiliate code.
- OAuth affiliate storage still passes `aff_code`.
- `/invite` redirects to `/affiliate` or is no longer present in navigation.

Verification commands should include:

```bash
cd /root/sub2api-src/backend && go test -tags unit ./internal/service ./internal/handler ./internal/server -run 'Auth|Affiliate|Invite|Register|OAuth|InvitationCode' -count=1
cd /root/sub2api-src/frontend && npm run test:run -- RegisterView EmailVerifyView oauthAffiliate
cd /root/sub2api-src/frontend && npm run typecheck
git -C /root/sub2api-src diff --check
```

## Rollout And Rollback

Rollout should happen in two implementation phases:

1. Semantic retirement:
   - Stop enforcing registration invitation codes.
   - Stop exposing legacy invite product surfaces.
   - Keep backend alias compatibility for `invitation_code`.
   - Keep legacy tables intact.
2. Physical cleanup:
   - Remove legacy invite code, schemas, generated code, and stale UI after phase-one behavior is verified.

Rollback for phase one is straightforward because the data schema remains intact. If needed, route exposure and old frontend controls can be restored without data loss. Affiliate data created during phase one remains valid and should not be discarded.

## Open Decisions

The design fixes the first-phase compatibility choice: backend keeps `invitation_code` as a deprecated alias to `aff_code`, while frontend stops using it.

The only deferred decision is whether phase two should physically drop legacy invite tables or leave them permanently as archived historical data. That decision should be made after phase one has run successfully.
