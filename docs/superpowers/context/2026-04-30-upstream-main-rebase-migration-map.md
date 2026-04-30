# Upstream Main Rebase Migration Map

**Date:** 2026-04-30

## Policy

- Upstream schema is the desired target schema.
- xlabapi production migration history is preserved.
- Existing xlabapi migration files are not overwritten by upstream files.
- New compatibility migrations are added after the xlabapi high-water mark.
- Generated Ent code is regenerated after schema and SQL decisions are complete.

## Current High-Water Marks

| Source | High-water mark | Notes |
| --- | --- | --- |
| xlabapi | 137_clear_legacy_subscription_carryover.sql | includes local shared subscription products, GPT image group, and carryover cleanup |
| upstream/main | 133_affiliate_rebate_freeze.sql | includes payment, auth identity, channel monitor, affiliate, WebSearch/notification-related schema |

## Required Mapping Categories

| Category | Action |
| --- | --- |
| identical migration | record as already equivalent |
| upstream migration absent from xlabapi | port as compatibility migration after xlabapi high-water mark if schema is required |
| xlabapi migration absent from upstream | preserve and ensure target schema still includes behavior |
| overlapping semantic migration | write one compatibility migration that reaches target schema without breaking existing installs |
| generated schema drift | resolve schema files first, then regenerate Ent |

## Initial Required Local Migrations To Preserve

| xlabapi migration | Required outcome |
| --- | --- |
| 109_add_subscription_daily_carryover.sql | daily carryover fields remain available |
| 110_add_shared_subscription_products.sql | shared subscription product schema remains available |
| 134_xlabapi_invite_to_affiliate_compat.sql | affiliate compatibility remains available |
| 135_add_channel_model_pricing_platform.sql | channel pricing platform remains available |
| 136_add_subscription_gpt_image_group.sql | GPT image subscription group seed remains available |
| 137_clear_legacy_subscription_carryover.sql | legacy carryover cleanup remains safe |

## Initial Required Upstream Schema Families

| Upstream family | Required outcome |
| --- | --- |
| account stats pricing | needed for WebSearch/pricing/account stats work |
| balance notify fields | needed for notifications |
| notify email entries | needed for notification recipient toggles |
| WebSearch tri-state | needed for WebSearch migration |
| auth identity and pending auth | included by upstream baseline; preserve if required by settings/auth flow |
| payment provider/order/audit | included by upstream baseline; do not regress compile or migrations even if not product-focused |
| channel monitor/request templates | included by upstream baseline; keep compile and migrations coherent |
| affiliate rebate hardening | included by upstream baseline; reconcile with xlabapi affiliate-only invite semantics |

## Final Mapping Table

Populate this table during the schema child plan before writing SQL:

| Upstream migration | xlabapi equivalent | Final action | Verification |
| --- | --- | --- | --- |
| 101_add_balance_notify_fields.sql | none confirmed | classify in schema child plan | migration test |
| 104_migrate_notify_emails_to_struct.sql | none confirmed | classify in schema child plan | migration test |
| 105_migrate_websearch_emulation_to_tristate.sql | none confirmed | classify in schema child plan | migration test |
| 108_auth_identity_foundation_core.sql | none confirmed | classify in schema child plan | migration test |
| 125_add_channel_monitors.sql | none confirmed | classify in schema child plan | migration test |
| 130_add_user_affiliates.sql | 134_xlabapi_invite_to_affiliate_compat.sql partial | classify in schema child plan | migration test |

## Schema Child Plan Notes

- The schema child plan owns final migration numbering after `137_clear_legacy_subscription_carryover.sql`.
- New SQL must be idempotent where it can be run on databases that already received xlabapi local migrations.
- `docs/superpowers` is ignored by upstream `.gitignore`; stage context updates with `git add -f`.
