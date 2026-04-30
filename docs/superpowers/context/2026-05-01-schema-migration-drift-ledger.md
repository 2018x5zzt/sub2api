# Schema Migration Drift Ledger

**Date:** 2026-05-01
**Branch:** integrate/xlabapi-upstream-main-rebase-20260430

## Classification Legend

- upstream baseline: keep upstream-main schema or migration as-is
- xlabapi replay: add xlabapi-only schema or SQL behavior
- fused: combine upstream and xlabapi behavior into one target schema
- migration preserved: keep production SQL history but do not duplicate runtime behavior
- generated only: regenerate from schema; do not hand-edit

## High-Risk Schema Families

| Family | Initial classification | Required decision |
| --- | --- | --- |
| auth identity and pending auth | upstream baseline | keep upstream schema unless xlabapi auth tests require local fields |
| payment order/provider/audit | upstream baseline | keep compile and migrations coherent even if payment product is not the focus |
| channel monitor/request templates | upstream baseline | keep upstream schema; defer runtime feature behavior to channel plan |
| subscription products | xlabapi replay | preserve shared subscription products and user product subscriptions |
| daily carryover | xlabapi replay | preserve user subscription carryover fields and cleanup behavior |
| affiliate/invite compatibility | fused | keep upstream affiliate hardening and xlabapi affiliate-only invite migration behavior |
| channel pricing platform | fused | keep upstream channel features and xlabapi pricing platform behavior |
| usage log billing and metadata | fused | preserve upstream fields plus xlabapi billing, product, cache, endpoint, and model metadata |

## Initial File Drift

The source drift was captured with:

```bash
git -C /root/sub2api-src diff --name-status upstream/main..xlabapi -- backend/ent/schema backend/migrations backend/internal/repository/migrations_schema_integration_test.go > /tmp/schema-migration-drift.txt
```

### Ent Schema Drift

| File or family | Drift | Initial action |
| --- | --- | --- |
| `backend/ent/schema/auth_identity*.go` | present upstream, absent in xlabapi | upstream baseline |
| `backend/ent/schema/channel_monitor*.go` | present upstream, absent in xlabapi | upstream baseline |
| `backend/ent/schema/payment_*.go` | present upstream, absent in xlabapi | upstream baseline |
| `backend/ent/schema/pending_auth_session.go` | present upstream, absent in xlabapi | upstream baseline |
| `backend/ent/schema/subscription_plan.go` | present upstream, absent in xlabapi | upstream baseline unless superseded by xlabapi products |
| `backend/ent/schema/subscription_product*.go` | xlabapi-only | replay or fuse with upstream subscription plans |
| `backend/ent/schema/user_product_subscription.go` | xlabapi-only | replay |
| `backend/ent/schema/product_subscription_migration_source.go` | xlabapi-only | replay if migration compatibility needs source tracking |
| `backend/ent/schema/group_health_snapshot.go` | xlabapi-only | replay if runtime stability/channel health still depends on it |
| `backend/ent/schema/invite_*` | xlabapi-only invite history | fuse with upstream affiliate schema |
| `backend/ent/schema/api_key.go` | changed in both source histories | inspect field-level diff before edit |
| `backend/ent/schema/group.go` | changed in both source histories | fuse RPM, channel, subscription, and visibility behavior |
| `backend/ent/schema/usage_log.go` | changed in both source histories | preserve upstream plus xlabapi product/billing/cache metadata |
| `backend/ent/schema/user.go` | changed in both source histories | preserve upstream identity fields plus xlabapi compatibility fields |
| `backend/ent/schema/user_subscription.go` | changed in both source histories | preserve upstream behavior plus carryover fields |

### Migration Drift

| Migration family | Drift | Initial action |
| --- | --- | --- |
| `077..089` xlabapi renumbered usage/channel migrations | xlabapi renumbered several upstream-era migrations | preserve upstream numbering in new baseline; do not rename existing upstream files |
| `081_add_group_account_filter.sql` | upstream-only deletion in xlabapi diff | keep upstream baseline unless replaced by xlabapi group visibility semantics |
| `086_channel_platform_pricing.sql` | upstream-only deletion in xlabapi diff | cover final behavior with compatibility migration if xlabapi channel pricing platform still required |
| `087_usage_log_billing_mode.sql` | upstream-only deletion in xlabapi diff | keep upstream baseline or map to xlabapi `108_add_usage_log_billing_mode.sql` |
| `090_drop_sora.sql` | upstream-only deletion in xlabapi diff | do not blindly apply Sora removal if xlabapi gateway compatibility still exposes Sora |
| `092..096` payment migrations | upstream-only deletion in xlabapi diff | keep upstream baseline for compile and migration coherence |
| `101_*` account stats, balance notify, channel features, payment mode | upstream-only deletion in xlabapi diff | keep upstream baseline and map runtime behavior in feature plans |
| `104_migrate_notify_emails_to_struct.sql` | upstream-only deletion in xlabapi diff | keep upstream baseline for notification feature |
| `105_migrate_websearch_emulation_to_tristate.sql` | upstream-only deletion in xlabapi diff | keep upstream baseline for WebSearch feature |
| `108..124` auth identity and payment hardening | upstream-only deletion in xlabapi diff | keep upstream baseline |
| `125..129` channel monitor/request templates/RPM | upstream-only deletion in xlabapi diff | keep upstream baseline and reconcile runtime later |
| `130..133` affiliate hardening | upstream-only deletion in xlabapi diff | fuse with xlabapi affiliate-only invite migration |
| `134_xlabapi_invite_to_affiliate_compat.sql` | xlabapi-only | replay after high-water mark as compatibility migration |
| `135_add_channel_model_pricing_platform.sql` | xlabapi-only | replay or fuse after high-water mark |
| `136_add_subscription_gpt_image_group.sql` | xlabapi-only | replay seed behavior if still product-required |
| `137_clear_legacy_subscription_carryover.sql` | xlabapi-only | preserve cleanup behavior if daily carryover is replayed |

## Next Decisions

- Decide whether xlabapi shared subscription products coexist with upstream subscription plans or replace the relevant runtime surface.
- Decide whether `090_drop_sora.sql` can remain in upstream baseline while xlabapi keeps Sora API compatibility through separate tables or non-schema code.
- Decide final compatibility migration numbering after `137_clear_legacy_subscription_carryover.sql`.
- Write failing migration tests before adding compatibility SQL.
