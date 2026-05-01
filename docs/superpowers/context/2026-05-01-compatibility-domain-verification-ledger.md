# Compatibility Domain Verification Ledger

**Date:** 2026-05-01
**Branch:** integrate/xlabapi-upstream-main-rebase-20260430
**Baseline:** upstream/main `48912014`

## Verified Domains

| Domain | Verification | Result |
| --- | --- | --- |
| account bulk edit | `go test -tags=unit ./internal/service -run 'BulkUpdateAccounts' -count=1` | pass |
| account bulk edit + mixed channel handler | `go test ./internal/handler/admin -run 'BulkUpdate|MixedChannel' -count=1` | pass |
| Vertex service accounts | `go test ./internal/service -run 'Vertex|ServiceAccount|AnthropicVertex' -count=1` | pass |
| WebSearch and notifications | `go test -tags=unit ./internal/service -run 'WebSearch|BalanceNotify|QuotaNotify|Notify' -count=1` | pass |
| WebSearch providers | `go test ./internal/pkg/websearch -count=1` | pass |
| OpenAI/Codex/Claude transforms | `go test ./internal/pkg/apicompat ./internal/pkg/antigravity ./internal/service -run 'WebSearch|Codex|Claude|Responses|Tools|GeminiMessages|Antigravity|Anthropic' -count=1` | pass |
| available channels handler | `go test -tags=unit ./internal/handler -run 'AvailableChannel|FilterUserVisible|ToUserSupportedModels' -count=1` | pass |
| frontend bulk edit, Vertex account controls, settings, router/sidebar | `pnpm vitest run src/views/admin/__tests__/AccountsView.bulkEdit.spec.ts src/components/account/__tests__/BulkEditAccountModal.spec.ts src/components/account/__tests__/AccountUsageCell.spec.ts src/components/account/__tests__/EditAccountModal.spec.ts src/components/layout/__tests__/AppSidebar.spec.ts src/router/__tests__/guards.spec.ts src/views/admin/__tests__/SettingsView.spec.ts` | pass |
| model plaza legacy routes | `pnpm vitest run src/router/__tests__/model-plaza-compat.spec.ts` | red then green |

## Excluded Or Classified Errors

| Finding | Classification | Action |
| --- | --- | --- |
| Unquoted `|` in `go test -run` commands executed shell pipelines such as `MixedChannel` as commands. | test command error, not product failure | Use quoted regex in every `-run` command. |
| Some Go tests returned `[no tests to run]` without `-tags=unit`. | test command error, not product failure | Use `-tags=unit` for unit-tagged service/handler tests. |
| Old `xlabapi` ModelHub full page and `/groups/models` data path are absent from the upstream-main integration branch. | intentionally not replayed runtime surface | Primary surface is upstream Available Channels. Legacy model plaza paths redirect to `/available-channels` and are covered by `model-plaza-compat.spec.ts`. |
| `SettingsView` frontend tests print unresolved `router-link` warnings. | existing test harness warning | Do not classify stderr alone as failure when Vitest exits 0 and reports all tests passed. |

## Final Verification

| Verification | Result |
| --- | --- |
| `go test ./...` | pass |
| `pnpm typecheck` | pass |
| `pnpm test:run` | pass, 92 files and 545 tests |
