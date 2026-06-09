# Core Upgrade v0.1.122 Range Audit

## Range

- Base tag: v0.1.121
- Target tag: v0.1.122
- Production branch baseline: xlabapi at f5a637be

## Expected impact areas

- Backend OpenAI gateway and API compatibility paths.
- Backend affiliate/admin records and balance history paths.
- One affiliate ledger audit snapshot migration.
- Legacy frontend admin affiliate and account bulk-edit pages.

## Protected xlab areas

The range must not delete or unintentionally rewrite:

- backend/internal/enterprisebff/**
- frontend-v2/**
- xlab-backend/**
- xlab product-subscription/payment/redeem/quota semantics

## Name-status capture

Paste the output of:

```bash
git diff --name-status v0.1.121..v0.1.122
```

below this line during execution.

## Stat capture

Paste the output of:

```bash
git diff --stat v0.1.121..v0.1.122
```

below this line during execution.

### Captured name-status

```text
M	backend/cmd/server/VERSION
M	backend/internal/handler/admin/account_handler.go
M	backend/internal/handler/admin/affiliate_handler.go
M	backend/internal/handler/admin/user_handler.go
M	backend/internal/handler/openai_chat_completions.go
M	backend/internal/handler/openai_gateway_handler.go
M	backend/internal/handler/openai_gateway_handler_test.go
M	backend/internal/pkg/apicompat/anthropic_responses_test.go
M	backend/internal/pkg/apicompat/chatcompletions_responses_test.go
M	backend/internal/pkg/apicompat/responses_to_anthropic.go
M	backend/internal/pkg/apicompat/responses_to_chatcompletions.go
M	backend/internal/pkg/apicompat/types.go
A	backend/internal/pkg/openai_compat/upstream_capability.go
A	backend/internal/pkg/openai_compat/upstream_capability_test.go
M	backend/internal/repository/affiliate_repo.go
M	backend/internal/repository/affiliate_repo_integration_test.go
A	backend/internal/repository/affiliate_repo_test.go
M	backend/internal/server/routes/admin.go
M	backend/internal/service/account_test_service.go
A	backend/internal/service/admin_balance_history_test.go
M	backend/internal/service/admin_service.go
M	backend/internal/service/affiliate_service.go
M	backend/internal/service/domain_constants.go
M	backend/internal/service/gateway_service.go
M	backend/internal/service/gateway_service_streaming_test.go
A	backend/internal/service/openai_apikey_responses_probe.go
M	backend/internal/service/openai_compat_model_test.go
M	backend/internal/service/openai_fast_policy_ws_test.go
M	backend/internal/service/openai_gateway_403_reset_test.go
M	backend/internal/service/openai_gateway_chat_completions.go
A	backend/internal/service/openai_gateway_chat_completions_raw.go
A	backend/internal/service/openai_gateway_chat_completions_raw_test.go
M	backend/internal/service/openai_gateway_chat_completions_test.go
M	backend/internal/service/openai_gateway_messages.go
M	backend/internal/service/openai_gateway_record_usage_test.go
M	backend/internal/service/openai_gateway_service.go
M	backend/internal/service/openai_images.go
M	backend/internal/service/openai_images_test.go
M	backend/internal/service/openai_oauth_passthrough_test.go
M	backend/internal/service/openai_ws_forwarder.go
M	backend/internal/service/openai_ws_forwarder_ingress_session_test.go
M	backend/internal/service/openai_ws_v2_passthrough_adapter.go
M	backend/internal/service/payment_fulfillment.go
A	backend/migrations/134_affiliate_ledger_audit_snapshots.sql
M	backend/migrations/auth_identity_payment_migrations_regression_test.go
M	frontend/src/api/admin/affiliates.ts
M	frontend/src/api/admin/users.ts
M	frontend/src/components/account/BulkEditAccountModal.vue
M	frontend/src/components/account/__tests__/BulkEditAccountModal.spec.ts
M	frontend/src/components/admin/user/UserBalanceHistoryModal.vue
M	frontend/src/components/layout/AppSidebar.vue
M	frontend/src/i18n/locales/en.ts
M	frontend/src/i18n/locales/zh.ts
M	frontend/src/router/index.ts
A	frontend/src/views/admin/affiliates/AdminAffiliateInvitesView.vue
A	frontend/src/views/admin/affiliates/AdminAffiliateRebatesView.vue
A	frontend/src/views/admin/affiliates/AdminAffiliateRecordsTable.vue
A	frontend/src/views/admin/affiliates/AdminAffiliateTransfersView.vue
```

### Captured stat

```text
backend/cmd/server/VERSION                         |   2 +-
 backend/internal/handler/admin/account_handler.go  |  40 ++
 .../internal/handler/admin/affiliate_handler.go    | 108 +++++
 backend/internal/handler/admin/user_handler.go     |   2 +-
 .../internal/handler/openai_chat_completions.go    |  16 +-
 backend/internal/handler/openai_gateway_handler.go |   1 +
 .../handler/openai_gateway_handler_test.go         | 316 ++++++++++++++
 .../pkg/apicompat/anthropic_responses_test.go      |  39 ++
 .../apicompat/chatcompletions_responses_test.go    |  43 ++
 .../pkg/apicompat/responses_to_anthropic.go        |   4 +-
 .../pkg/apicompat/responses_to_chatcompletions.go  |   4 +-
 backend/internal/pkg/apicompat/types.go            |   2 +-
 .../pkg/openai_compat/upstream_capability.go       |  75 ++++
 .../pkg/openai_compat/upstream_capability_test.go  |  55 +++
 backend/internal/repository/affiliate_repo.go      | 465 ++++++++++++++++++++-
 .../repository/affiliate_repo_integration_test.go  |  22 +-
 backend/internal/repository/affiliate_repo_test.go |  28 ++
 backend/internal/server/routes/admin.go            |   5 +
 backend/internal/service/account_test_service.go   |  12 +-
 .../internal/service/admin_balance_history_test.go |  86 ++++
 backend/internal/service/admin_service.go          | 200 ++++++++-
 backend/internal/service/affiliate_service.go      | 138 +++++-
 backend/internal/service/domain_constants.go       |   9 +-
 backend/internal/service/gateway_service.go        |   7 +
 .../service/gateway_service_streaming_test.go      |  13 +
 .../service/openai_apikey_responses_probe.go       | 149 +++++++
 .../internal/service/openai_compat_model_test.go   | 287 +++++++++++++
 .../internal/service/openai_fast_policy_ws_test.go |  56 +++
 .../service/openai_gateway_403_reset_test.go       |  19 +-
 .../service/openai_gateway_chat_completions.go     | 243 ++++++-----
 .../service/openai_gateway_chat_completions_raw.go | 437 +++++++++++++++++++
 .../openai_gateway_chat_completions_raw_test.go    | 260 ++++++++++++
 .../openai_gateway_chat_completions_test.go        | 262 ++++++++++++
 .../internal/service/openai_gateway_messages.go    | 376 ++++++++++++-----
 .../service/openai_gateway_record_usage_test.go    |  50 +++
 backend/internal/service/openai_gateway_service.go |  11 +-
 backend/internal/service/openai_images.go          |  62 ++-
 backend/internal/service/openai_images_test.go     | 103 +++++
 .../service/openai_oauth_passthrough_test.go       |  92 ++++
 backend/internal/service/openai_ws_forwarder.go    |   7 +-
 .../openai_ws_forwarder_ingress_session_test.go    |   4 +-
 .../service/openai_ws_v2_passthrough_adapter.go    | 100 ++++-
 backend/internal/service/payment_fulfillment.go    |   3 +-
 .../134_affiliate_ledger_audit_snapshots.sql       |  85 ++++
 ..._identity_payment_migrations_regression_test.go |  15 +
 frontend/src/api/admin/affiliates.ts               | 122 ++++++
 frontend/src/api/admin/users.ts                    |   2 +-
 .../components/account/BulkEditAccountModal.vue    | 151 ++++++-
 .../account/__tests__/BulkEditAccountModal.spec.ts |  38 ++
 .../admin/user/UserBalanceHistoryModal.vue         |   5 +-
 frontend/src/components/layout/AppSidebar.vue      |  13 +
 frontend/src/i18n/locales/en.ts                    |  49 +++
 frontend/src/i18n/locales/zh.ts                    |  49 +++
 frontend/src/router/index.ts                       |  40 ++
 .../admin/affiliates/AdminAffiliateInvitesView.vue |   7 +
 .../admin/affiliates/AdminAffiliateRebatesView.vue |   7 +
 .../affiliates/AdminAffiliateRecordsTable.vue      | 407 ++++++++++++++++++
 .../affiliates/AdminAffiliateTransfersView.vue     |   7 +
 58 files changed, 4940 insertions(+), 270 deletions(-)
```
