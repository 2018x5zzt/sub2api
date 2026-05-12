# P0 Contract Test Plan

- 目标：把 T004 中“先补 30–40 个 P0 golden contract”的建议细化成可执行测试设计。本文只做设计，不写真测试代码。
- 输入：`docs/superpowers/audits/2026-05-12-test-coverage-inventory.md`、现有 Go test 风格、`origin/test/xlabapi` 路由/contract test 现状。
- T007 状态：`docs/superpowers/contracts/API-CONTRACT.md` 当前未落盘；涉及字段级细节处以 `[待 T007 补]` 标记，后续应按 T007 的正式 API 契约校正。
- 设计边界：优先保护 frontend-v2 和后端重构前最容易漂移的 HTTP contract：method/path、status、认证、envelope、分页、字段名/类型、null vs empty、cookie/header/token、错误响应。

---

## 1. 测试框架选型

### 1.1 Go 端

建议：继续使用 stdlib `testing` + `net/http/httptest` + 现有 `testify/require`，不引入新的 BDD/contract 框架。

理由：

- 仓库已经有同类先例：`backend/internal/server/api_contract_test.go` 使用 table-driven case、`httptest.NewRecorder()`、`require.Equal`、固定 JSON 断言。
- 现有 Go 测试整体风格就是 stdlib + `require`，例如 `backend/internal/handler/request_body_limit_test.go` 直接构造 `gin.New()`、`httptest.NewRequest()`、`httptest.NewRecorder()`。
- P0 contract 的核心不是复杂 matcher，而是稳定、可 review 的输入/输出快照；引入 Ginkgo/Pact 之类第三方框架会增加维护面，不解决当前缺口。
- 后端已有大量 service/repository integration；新增 P0 应落在 server/HTTP 边界，不应先扩散为全量真实 DB e2e。

建议文件组织：

| 用途 | 建议路径 | 说明 |
|---|---|---|
| xlabapi envelope contract | `backend/internal/server/api_contract_test.go` 或拆成 `backend/internal/server/contract_api_test.go` | 覆盖 `/api/v1/*`，使用 stub deps + `httptest` |
| golden JSON | `backend/internal/server/testdata/contracts/api_v1/*.golden.json` | 文件名建议 `C001_auth_login_success.golden.json` |
| gateway compatibility | `backend/internal/server/gateway_contract_test.go` | 覆盖 `/v1/*`、root `/responses`、SSE/非 envelope 兼容响应 |
| gateway golden | `backend/internal/server/testdata/contracts/gateway/*.golden.json` | OpenAI/Anthropic/Gemini-compatible 响应单独存放，避免混入 xlabapi envelope |
| shared helpers | `backend/internal/server/contract_test_helpers_test.go` | `newContractTestServer`、`readGolden`、`assertJSONEqual`、`normalizeDynamicFields` |

Golden 比较规则：

- 固定时间、ID、key、token 前缀，或测试中 normalize 动态字段。
- JSON 先 unmarshal 再比较，避免格式/字段顺序造成噪音；但 golden 文件保留 pretty JSON 便于 review。
- 对 token、secret、签名、随机码只断言字段存在、类型、前缀/长度，不把真实随机值写入 golden。
- 错误响应也要 golden：P0 不能只钉 happy path。

### 1.2 前端

建议：frontend-v2 先引入 Vitest + React Testing Library + MSW；Playwright 只保留为后续少量 smoke/e2e，不作为第一批 contract 工作主力。

理由：

- frontend-v2 当前无 `*.test.ts(x)`、无 Vitest/Playwright 配置，第一步应低成本覆盖 API client 和 auth/data hooks。
- MSW 可以在前端侧模拟 `/api/v1` contract，验证 API client 对 envelope、401 refresh、pagination、null/empty 的解释是否与后端契约一致。
- Playwright 更适合跨页面冒烟，不适合快速锁定 30–40 个 HTTP contract；现在全量 e2e 会被 UI parity 重构频繁打断。

建议文件组织：

| 用途 | 建议路径 | 说明 |
|---|---|---|
| Vitest config | `frontend-v2/vitest.config.ts` | jsdom + setup file |
| MSW handlers | `frontend-v2/src/test/msw/handlers.ts` | 以后端 golden contract 为源，不另造一套字段 |
| API client tests | `frontend-v2/src/api/__tests__/*.spec.ts` | 先测 `client.ts`、`auth.ts`、`keys.ts`、`usage.ts`、`subscriptions.ts`、`payment.ts` |
| route smoke | `frontend-v2/src/router/__tests__/guards.spec.tsx` | auth guard、admin guard、not found |
| page smoke | `frontend-v2/src/pages/**/__tests__/*.spec.tsx` | 只测 loading/error/empty/success，不做大 UI snapshot |

### 1.3 Snapshot / Golden 存放原则

| 类型 | 存放 | 命名 |
|---|---|---|
| 后端 `/api/v1` JSON golden | `backend/internal/server/testdata/contracts/api_v1/` | `C001_auth_login_success.golden.json` |
| 后端 gateway JSON golden | `backend/internal/server/testdata/contracts/gateway/` | `C032_gateway_models_success.golden.json` |
| 后端 SSE golden | `backend/internal/server/testdata/contracts/gateway/` | `C034_gateway_chat_stream_success.golden.sse` |
| 前端 MSW fixtures | `frontend-v2/src/test/fixtures/contracts/` | 从后端 golden 派生，文件名保持编号一致 |
| 文档索引 | `docs/superpowers/contracts/CONTRACT-TEST-PLAN.md` | 本文件作为测试落地索引 |

原则：后端 golden 是单一事实源；前端 fixtures 应从后端 golden 拷贝/生成，不允许独立发明字段。

---

## 2. P0 契约测试清单（36 个）

伪代码约定：以下骨架只展示测试形态，不要求字段完整；正式实现时应按 T007 的 `API-CONTRACT.md` 和 golden fixture 补齐。公共 helper 假设为 `newContractTestServer(t)`、`doJSON`、`assertGolden`、`authHeader(userID)`、`adminHeader()`。

### C001 — Auth Login Success

- 端点：`POST /api/v1/auth/login`
- 类型：contract
- 检查点：`200 OK`；xlabapi envelope `code/message/data`；`data.access_token` 存在且为 string；refresh token/cookie 行为按 T007 固定；用户基础字段不缺失。
- 前置依赖：预置 active user `alice@example.com`、密码 hash、token version。
- 后置清理：无，使用 stub repo 或事务回滚。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedUser("alice@example.com", "pass")
body := `{"email":"alice@example.com","password":"pass"}`
rec := httptest.NewRecorder()
req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
req.Header.Set("Content-Type", "application/json")
srv.Router.ServeHTTP(rec, req)
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C001_auth_login_success.golden.json", rec.Body.Bytes())
```

### C002 — Auth Login Invalid Password

- 端点：`POST /api/v1/auth/login`
- 类型：contract
- 检查点：`401 Unauthorized` 或当前契约状态码；错误 envelope 固定；不返回 token；错误 message/code 固定；不设置成功 cookie。
- 前置依赖：预置 active user。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedUser("alice@example.com", "pass")
rec := doJSON(srv, http.MethodPost, "/api/v1/auth/login", `{"email":"alice@example.com","password":"bad"}`, nil)
require.Equal(t, http.StatusUnauthorized, rec.Code)
assertGolden(t, "C002_auth_login_invalid_password.golden.json", rec.Body.Bytes())
```

### C003 — Auth Register Success

- 端点：`POST /api/v1/auth/register`
- 类型：contract
- 检查点：`200 OK` 或当前契约状态码；envelope 固定；返回用户/token 字段按契约；默认 balance/concurrency/rpm/status 固定；invite/promo 字段处理固定。
- 前置依赖：注册开放；邮箱未存在；验证码/Turnstile 以 stub 通过。
- 后置清理：删除测试用户或事务回滚。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.AllowRegistration()
srv.AcceptVerifyCode("new@example.com", "123456")
body := `{"email":"new@example.com","password":"pass","code":"123456"}`
rec := doJSON(srv, http.MethodPost, "/api/v1/auth/register", body, nil)
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C003_auth_register_success.golden.json", rec.Body.Bytes())
```

### C004 — Auth Refresh Token

- 端点：`POST /api/v1/auth/refresh`
- 类型：behavioral
- 检查点：有效 refresh 返回新 access token；无效/过期 refresh 返回固定错误；token version 生效；cookie/header 输入方式按契约固定。
- 前置依赖：预置用户和 refresh token。
- 后置清理：撤销测试 refresh token。
- 伪代码：

```go
srv := newContractTestServer(t)
refresh := srv.SeedRefreshToken(userID)
req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refresh})
rec := httptest.NewRecorder()
srv.Router.ServeHTTP(rec, req)
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C004_auth_refresh_success.golden.json", rec.Body.Bytes())
```

### C005 — Auth Me Compatibility Fields

- 端点：`GET /api/v1/auth/me`
- 类型：contract
- 检查点：`id/email/username/role/balance/status`；`identities`、`identity_bindings`、`auth_bindings` 三套兼容字段；`allowed_groups` null/array 语义；`run_mode`。
- 前置依赖：预置 authenticated user，带 email identity，其他 provider 未绑定。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedUserProfile(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/auth/me", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C005_auth_me_success.golden.json", rec.Body.Bytes())
```

### C006 — Auth Login 2FA Challenge Complete

- 端点：`POST /api/v1/auth/login/2fa`
- 类型：behavioral
- 检查点：challenge/session id 输入字段固定；成功后 token/envelope 固定；错误 code 固定；challenge 一次性消费。
- 前置依赖：预置开启 TOTP 的用户和 login challenge。
- 后置清理：清理 challenge。
- 伪代码：

```go
srv := newContractTestServer(t)
challenge := srv.Seed2FAChallenge(userID)
body := fmt.Sprintf(`{"challenge_id":"%s","code":"123456"}`, challenge)
rec := doJSON(srv, http.MethodPost, "/api/v1/auth/login/2fa", body, nil)
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C006_auth_login_2fa_success.golden.json", rec.Body.Bytes())
```

### C007 — API Key Create

- 端点：`POST /api/v1/keys`
- 类型：contract
- 检查点：`key` 明文只在 create 返回；`group_id` null 语义；quota/rate limit/window 字段完整；status/created_at/updated_at 类型固定。
- 前置依赖：预置 user，balance 非负，默认 group 可用。
- 后置清理：删除创建的 key。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedUser(userID)
body := `{"name":"Key One","custom_key":"sk_custom_1234567890"}`
rec := doJSON(srv, http.MethodPost, "/api/v1/keys", body, authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C007_keys_create_success.golden.json", rec.Body.Bytes())
```

### C008 — API Key List Pagination

- 端点：`GET /api/v1/keys?page=1&page_size=10`
- 类型：contract
- 检查点：pagination envelope；items array；key fields；`last_used_at` null；`total/page/page_size` 字段命名固定。
- 前置依赖：预置 2 个 API keys。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAPIKeys(userID, 2)
rec := doJSON(srv, http.MethodGet, "/api/v1/keys?page=1&page_size=10", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C008_keys_list_paginated.golden.json", rec.Body.Bytes())
```

### C009 — API Key Update Group And Budget

- 端点：`PUT /api/v1/keys/:id`
- 类型：contract
- 检查点：group binding shape；dynamic budget fields；unchanged key secret 不重新暴露；invalid group 错误固定。
- 前置依赖：预置 user、key、available group。
- 后置清理：恢复 key group 或事务回滚。
- 伪代码：

```go
srv := newContractTestServer(t)
keyID := srv.SeedAPIKey(userID)
srv.SeedAvailableGroup(userID, groupID)
body := `{"name":"Key One","group_id":2,"budget_multiplier":1.5}`
rec := doJSON(srv, http.MethodPut, fmt.Sprintf("/api/v1/keys/%d", keyID), body, authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C009_keys_update_group.golden.json", rec.Body.Bytes())
```

### C010 — API Key Delete

- 端点：`DELETE /api/v1/keys/:id`
- 类型：behavioral
- 检查点：成功 status/envelope；二次删除错误；非 owner 删除错误；list 不再出现。
- 前置依赖：预置 user + key。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
keyID := srv.SeedAPIKey(userID)
rec := doJSON(srv, http.MethodDelete, fmt.Sprintf("/api/v1/keys/%d", keyID), "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C010_keys_delete_success.golden.json", rec.Body.Bytes())
```

### C011 — Available Groups

- 端点：`GET /api/v1/groups/available`
- 类型：contract
- 检查点：items array；standard/product/dynamic group 字段；rate/pricing 字段；空列表语义。
- 前置依赖：预置 user allowed groups 和 product subscription group。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAvailableGroups(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/groups/available", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C011_groups_available.golden.json", rec.Body.Bytes())
```

### C012 — User Group Rates

- 端点：`GET /api/v1/groups/rates`
- 类型：contract
- 检查点：倍率字段命名；default vs override；dynamic pricing fields；null/zero 区分。
- 前置依赖：预置 group rate multiplier 和 user override。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedGroupRates(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/groups/rates", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C012_groups_rates.golden.json", rec.Body.Bytes())
```

### C013 — Usage List

- 端点：`GET /api/v1/usage?page=1&page_size=10`
- 类型：contract
- 检查点：pagination；`request_type` vs legacy `stream/openai_ws_mode`；model/requested_model visibility；cost/token fields。
- 前置依赖：预置 usage logs（stream、ws_v2、non-stream 各一条）。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedUsageLogs(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/usage?page=1&page_size=10", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C013_usage_list.golden.json", rec.Body.Bytes())
```

### C014 — Usage Stats

- 端点：`GET /api/v1/usage/stats?start_date=2025-01-01&end_date=2025-01-02`
- 类型：contract
- 检查点：date filters；numeric zero/default；total tokens/cost/request count；request_type filter compatibility。
- 前置依赖：预置 usage aggregation data。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedUsageStats(userID)
path := "/api/v1/usage/stats?start_date=2025-01-01&end_date=2025-01-02"
rec := doJSON(srv, http.MethodGet, path, "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C014_usage_stats.golden.json", rec.Body.Bytes())
```

### C015 — Usage Dashboard Stats

- 端点：`GET /api/v1/usage/dashboard/stats`
- 类型：contract
- 检查点：dashboard cards fields；today/month/all totals；balance/subscription-related fields；zero state。
- 前置依赖：预置 user、balance、usage logs。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedDashboardUsage(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/usage/dashboard/stats", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C015_usage_dashboard_stats.golden.json", rec.Body.Bytes())
```

### C016 — Usage Dashboard Trend

- 端点：`GET /api/v1/usage/dashboard/trend`
- 类型：contract
- 检查点：series array；date bucket format；missing days zero fill；cost/token/request fields。
- 前置依赖：预置跨日 usage logs。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedUsageTrend(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/usage/dashboard/trend", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C016_usage_dashboard_trend.golden.json", rec.Body.Bytes())
```

### C017 — Usage Dashboard Models

- 端点：`GET /api/v1/usage/dashboard/models`
- 类型：contract
- 检查点：model distribution array；requested/upstream model policy；percent/total fields；empty state。
- 前置依赖：预置多模型 usage logs。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedModelUsage(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/usage/dashboard/models", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C017_usage_dashboard_models.golden.json", rec.Body.Bytes())
```

### C018 — Gateway Chat Completion Billing Path

- 端点：`POST /v1/chat/completions`
- 类型：behavioral
- 检查点：OpenAI-compatible response 不走 xlabapi envelope；usage 写入 task 被触发；billing fingerprint 固定；auth/balance failure status 固定。
- 前置依赖：预置 API key、user balance、schedulable account、stub upstream。
- 后置清理：删除 usage log/billing record 或事务回滚。
- 伪代码：

```go
srv := newGatewayContractServer(t)
srv.SeedAPIKeyWithBalance("sk_test", 10)
srv.StubOpenAIChatCompletion()
body := `{"model":"gpt-5.4","messages":[{"role":"user","content":"hi"}]}`
rec := doJSON(srv, http.MethodPost, "/v1/chat/completions", body, bearer("sk_test"))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C018_gateway_chat_completion.golden.json", rec.Body.Bytes())
```

### C019 — User Subscriptions List

- 端点：`GET /api/v1/subscriptions`
- 类型：contract
- 检查点：list shape；legacy subscription fields；expired/active status；quota windows。
- 前置依赖：预置 active and expired subscriptions。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedSubscriptions(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/subscriptions", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C019_subscriptions_list.golden.json", rec.Body.Bytes())
```

### C020 — User Subscription Active

- 端点：`GET /api/v1/subscriptions/active`
- 类型：contract
- 检查点：single/null active subscription semantics；group fields；quota fields；no active response。
- 前置依赖：预置 active subscription。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedActiveSubscription(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/subscriptions/active", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C020_subscriptions_active.golden.json", rec.Body.Bytes())
```

### C021 — User Subscription Progress

- 端点：`GET /api/v1/subscriptions/progress`
- 类型：contract
- 检查点：daily/weekly/monthly progress；percent clamp；reset seconds non-negative；over-limit behavior。
- 前置依赖：预置 subscription limits and usage。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedSubscriptionProgress(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/subscriptions/progress", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C021_subscriptions_progress.golden.json", rec.Body.Bytes())
```

### C022 — Product Subscription Summary

- 端点：`GET /api/v1/subscription-products/summary`
- 类型：contract
- 检查点：product family；shared product groups；active product subscription counts；balance fallback fields。
- 前置依赖：预置 GPT product subscription and standard group。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedProductSubscription(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/subscription-products/summary", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C022_subscription_products_summary.golden.json", rec.Body.Bytes())
```

### C023 — Redeem Success Balance Code

- 端点：`POST /api/v1/redeem`
- 类型：behavioral
- 检查点：success envelope；balance change；code consumed；duplicate redeem error stable。
- 前置依赖：预置 unused balance redeem code。
- 后置清理：回滚 redeem usage and balance。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedRedeemCode("BALANCE100")
rec := doJSON(srv, http.MethodPost, "/api/v1/redeem", `{"code":"BALANCE100"}`, authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C023_redeem_balance_success.golden.json", rec.Body.Bytes())
```

### C024 — Redeem Product Subscription Code

- 端点：`POST /api/v1/redeem`
- 类型：behavioral
- 检查点：subscription/product fields；validity days; assigned group/product; balance unaffected unless overage policy says so `[待 T007 补]`。
- 前置依赖：预置 product subscription redeem code and product。
- 后置清理：回滚 subscription/redeem usage。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedProductRedeemCode("PROD30")
rec := doJSON(srv, http.MethodPost, "/api/v1/redeem", `{"code":"PROD30"}`, authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C024_redeem_product_subscription_success.golden.json", rec.Body.Bytes())
```

### C025 — Redeem History

- 端点：`GET /api/v1/redeem/history?page=1&page_size=10`
- 类型：contract
- 检查点：pagination；code masking policy；type/source/status fields；created_at format。
- 前置依赖：预置 redeem history records。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedRedeemHistory(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/redeem/history?page=1&page_size=10", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C025_redeem_history.golden.json", rec.Body.Bytes())
```

### C026 — Affiliate Summary

- 端点：`GET /api/v1/user/aff`
- 类型：contract
- 检查点：invite code/link；effective invitees；claimed/frozen/available quota；rate percent。
- 前置依赖：预置 affiliate records and settings。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAffiliateSummary(userID)
rec := doJSON(srv, http.MethodGet, "/api/v1/user/aff", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C026_affiliate_summary.golden.json", rec.Body.Bytes())
```

### C027 — Affiliate Transfer Quota

- 端点：`POST /api/v1/user/aff/transfer`
- 类型：behavioral
- 检查点：available quota decreases；balance increases；ledger record returned/created；insufficient quota error stable。
- 前置依赖：预置 matured affiliate quota。
- 后置清理：回滚 balance and transfer record。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAffiliateQuota(userID, 50)
rec := doJSON(srv, http.MethodPost, "/api/v1/user/aff/transfer", `{"amount":10}`, authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C027_affiliate_transfer_success.golden.json", rec.Body.Bytes())
```

### C028 — Admin Users List

- 端点：`GET /api/v1/admin/users?page=1&page_size=20`
- 类型：contract
- 检查点：admin auth required；pagination；activity fields；balance/status/role；sort/filter echo policy `[待 T007 补]`。
- 前置依赖：预置 admin user and normal users。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAdmin(adminID)
srv.SeedUsers(3)
rec := doJSON(srv, http.MethodGet, "/api/v1/admin/users?page=1&page_size=20", "", adminHeader(adminID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C028_admin_users_list.golden.json", rec.Body.Bytes())
```

### C029 — Admin Groups List

- 端点：`GET /api/v1/admin/groups`
- 类型：contract
- 检查点：group dynamic pricing fields；rate multipliers；sort order；standard/product group type fields。
- 前置依赖：预置 groups of multiple types。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAdmin(adminID)
srv.SeedGroups()
rec := doJSON(srv, http.MethodGet, "/api/v1/admin/groups", "", adminHeader(adminID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C029_admin_groups_list.golden.json", rec.Body.Bytes())
```

### C030 — Admin Accounts List

- 端点：`GET /api/v1/admin/accounts?page=1&page_size=20`
- 类型：contract
- 检查点：account type/platform/status fields；group binding; quota/tier fields；secret redaction policy。
- 前置依赖：预置 accounts with mixed channels。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAdmin(adminID)
srv.SeedAccounts()
rec := doJSON(srv, http.MethodGet, "/api/v1/admin/accounts?page=1&page_size=20", "", adminHeader(adminID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C030_admin_accounts_list.golden.json", rec.Body.Bytes())
```

### C031 — Admin Settings Get

- 端点：`GET /api/v1/admin/settings`
- 类型：contract
- 检查点：auth source defaults；payment visible methods；omitted/implicit defaults；sensitive fields masking。
- 前置依赖：预置 settings partial values and config defaults。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAdmin(adminID)
srv.SeedSettingsPartial()
rec := doJSON(srv, http.MethodGet, "/api/v1/admin/settings", "", adminHeader(adminID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C031_admin_settings_get.golden.json", rec.Body.Bytes())
```

### C032 — Admin Redeem Codes List

- 端点：`GET /api/v1/admin/redeem-codes?page=1&page_size=20`
- 类型：contract
- 检查点：pagination；balance/subscription/product fields；status/source_type；sort/search params。
- 前置依赖：预置 redeem codes of multiple types。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAdmin(adminID)
srv.SeedRedeemCodes()
rec := doJSON(srv, http.MethodGet, "/api/v1/admin/redeem-codes?page=1&page_size=20", "", adminHeader(adminID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C032_admin_redeem_codes_list.golden.json", rec.Body.Bytes())
```

### C033 — Admin Payment Dashboard

- 端点：`GET /api/v1/admin/payment/dashboard`
- 类型：contract
- 检查点：summary numbers；provider distribution；order status counts；zero state。
- 前置依赖：预置 payment orders across statuses。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedAdmin(adminID)
srv.SeedPaymentOrders()
rec := doJSON(srv, http.MethodGet, "/api/v1/admin/payment/dashboard", "", adminHeader(adminID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C033_admin_payment_dashboard.golden.json", rec.Body.Bytes())
```

### C034 — Payment Config Public User

- 端点：`GET /api/v1/payment/config`
- 类型：contract
- 检查点：enabled methods；visible method routing；provider display fields；no sensitive config leaked。
- 前置依赖：预置 enabled providers and visible methods。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedPaymentConfig()
rec := doJSON(srv, http.MethodGet, "/api/v1/payment/config", "", authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C034_payment_config.golden.json", rec.Body.Bytes())
```

### C035 — Payment Create Order

- 端点：`POST /api/v1/payment/orders`
- 类型：behavioral
- 检查点：order id/out_trade_no；provider-specific payload；return_url/resume token；pending status；plan snapshot。
- 前置依赖：预置 user、payment plan、enabled provider。
- 后置清理：cancel/delete test order or transaction rollback。
- 伪代码：

```go
srv := newContractTestServer(t)
srv.SeedPaymentPlan("basic")
srv.SeedPaymentProvider("stripe")
body := `{"plan_id":1,"provider":"stripe"}`
rec := doJSON(srv, http.MethodPost, "/api/v1/payment/orders", body, authHeader(userID))
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C035_payment_order_create.golden.json", rec.Body.Bytes())
```

### C036 — Payment Public Resolve By Resume Token

- 端点：`POST /api/v1/payment/public/orders/resolve`
- 类型：contract
- 检查点：resume token validation；public-safe order fields；snapshot authority；invalid/expired token error。
- 前置依赖：预置 order and signed resume token。
- 后置清理：无。
- 伪代码：

```go
srv := newContractTestServer(t)
token := srv.SeedPaymentResumeToken(orderID)
rec := doJSON(srv, http.MethodPost, "/api/v1/payment/public/orders/resolve", fmt.Sprintf(`{"resume_token":"%s"}`, token), nil)
require.Equal(t, http.StatusOK, rec.Code)
assertGolden(t, "C036_payment_public_resolve.golden.json", rec.Body.Bytes())
```

---

## 3. 落地顺序

### 3.1 第一批：最小硬门槛（8 个）

必须最先落地：C001、C002、C005、C007、C008、C015、C018、C031。

| 编号 | 端点 | 为什么必须先做 |
|---|---|---|
| C001 | `POST /api/v1/auth/login` | frontend-v2 所有登录后页面的入口；token/cookie/envelope 漂移会让整体不可用 |
| C002 | `POST /api/v1/auth/login` invalid password | 错误 envelope 和状态码是前端表单/拦截器的核心依赖；只测 happy path 不够 |
| C005 | `GET /api/v1/auth/me` | 登录态恢复、用户菜单、权限、identity binding 均依赖此端点；已有 contract，可先标准化迁移成 golden |
| C007 | `POST /api/v1/keys` | 用户核心动作：创建 API key；明文 key 只返回一次，最容易被重构误伤 |
| C008 | `GET /api/v1/keys` | console key 列表基础；分页/null/窗口字段漂移会直接破页面 |
| C015 | `GET /api/v1/usage/dashboard/stats` | dashboard 首页最核心数据；frontend-v2 首屏依赖，能快速发现 usage contract 漂移 |
| C018 | `POST /v1/chat/completions` | 核心扣费路径和外部 API compatibility；不属于 frontend-v2，但属于业务生命线 |
| C031 | `GET /api/v1/admin/settings` | admin/settings 已有 partial contract，且控制 auth/payment/frontend 配置默认值；先钉住能降低后续配置重构风险 |

第一批退出条件：

- 8 个 contract 全部通过，含成功和至少一个关键错误路径。
- Golden 文件已 review，动态字段已 normalize，无真实 secret/token 泄漏。
- 手工冒烟：登录、创建 key、dashboard stats、一次 chat completion 代理请求、admin settings 页面能走通。
- 后端重构不得在第一批未通过时合入。

### 3.2 第二批：订阅 + Redeem + Affiliate

建议落地：C019、C020、C021、C022、C023、C024、C025、C026、C027。

理由：

- 订阅、产品订阅、redeem、affiliate 是当前语义双轨风险最高的区域之一。
- T004 已确认这些领域内部 service/repo 测试不少，但 HTTP contract 缺口明显。
- 第二批完成后，前端可以安全改造 subscriptions/redeem/affiliate 页面，后端也可以继续推进语义收敛。

第二批退出条件：

- user-side subscription/redeem/affiliate contract 全部通过。
- balance code 与 product subscription code 都有 golden。
- 手工冒烟：查看订阅、查看进度、兑换 balance code、兑换 subscription code、查看/转出返利。
- 若 T007 定义了 `/redeem/benefit-leaderboard`，补入本批；若 T007 确认为残留，frontend-v2 应删除或改指向正式端点。

### 3.3 第三批：admin 核心

建议落地：C028、C029、C030、C032、C033，并补充 admin payment config/list orders/plans/providers 的拆分用例 `[待 T007 补]`。

理由：

- admin 页面多，最容易出现分页、排序、过滤、字段名、空值语义漂移。
- 先覆盖 users/groups/accounts/redeem/payment dashboard，能保护主要管理入口。
- ops、backup、channel-monitor 可在第三批后半段追加，不要阻塞核心 admin parity。

第三批退出条件：

- admin list 类 contract 均覆盖分页和空列表。
- 至少一个 mutation/bulk mutation contract 已覆盖，例如现有 `POST /admin/accounts/bulk-update` 或 admin redeem generate。
- 手工冒烟：admin users/groups/accounts/redeem/payment dashboard 页面可读可筛选。

### 3.4 第四批：Payment 完整链路

建议落地：C034、C035、C036，再补 `POST /api/v1/payment/orders/verify`、admin payment config/plans/providers CRUD、webhook notify contract `[待 T007 补]`。

理由：

- Payment 复杂度最高，涉及 provider-specific payload、resume token、public resolve、webhook、订单状态机。
- 不应放在第一批阻塞所有工作，但在 payment UI 或 payment service 重构前必须补齐。
- webhook 很难做纯 contract，需要明确 provider stub 和签名 normalize 策略。

第四批退出条件：

- 支付配置、创建订单、公开 resolve、verify 至少覆盖一个主 provider 和一个错误路径。
- 不泄漏 provider secret；provider snapshot 字段按 T007 固定。
- 手工冒烟：用户能看到支付计划/渠道，创建订单，返回/刷新后能 resolve 订单状态。

### 3.5 暂缓项

| 项 | 暂缓原因 | 触发条件 |
|---|---|---|
| 全量 Playwright e2e | UI 正在 parity/风格重构，维护成本高 | 第一、二批 contract 稳定后，只补 3–5 条关键 flow |
| 所有 admin ops contract | 端点多，P0 用户路径优先 | admin ops 页面进入重构前 |
| 全量 gateway SSE snapshot | SSE normalize 复杂，先钉非流式和模型列表 | gateway stream 重构前 |
| 组件级 UI snapshot | 易产生噪音，不能保护 HTTP contract | design token 稳定后再考虑 |

---

## 4. 与重构的联动

### 4.1 前端重构门槛

| 前端改造阶段 | 开始前必须通过 | 允许开始的范围 | 不允许做的事 |
|---|---|---|---|
| Auth/Login/Register/Profile | C001、C002、C003、C004、C005、C006 | 登录页、注册页、登录态恢复、2FA UI、profile 基础展示 | 未钉 token/cookie/错误 envelope 前，不改 API client auth interceptor |
| User Console Shell + Dashboard | C005、C015、C016、C017 | dashboard cards、trend/model charts、用户菜单 | 未钉 dashboard stats/trend 前，不重排字段映射逻辑 |
| Keys 页面 | C007、C008、C009、C010、C011、C012 | key list/create/update/delete、group/rate select | 未钉 create key 明文语义前，不改 key create modal |
| Usage 页面 | C013、C014、C015、C016、C017 | usage table、filters、dashboard usage widgets | 未钉 request_type/stream 兼容前，不改 usage filter model |
| Subscriptions/Redeem/Affiliate 页面 | C019–C027 | subscriptions、redeem、affiliate 页面 | 未钉 product subscription/redeem 双轨前，不合并语义字段 |
| Payment 页面 | C034、C035、C036 + verify 用例 `[待 T007 补]` | purchase/orders/result/public resume | 未钉 provider payload/resume token 前，不改支付跳转和回调页 |
| Admin 核心页面 | C028–C033 | users/groups/accounts/redeem/settings/payment dashboard | 未钉 pagination/filter/sort 前，不重写 admin table adapter |

### 4.2 后端重构门槛

| 后端改造阶段 | 开始前必须通过 | 额外要求 |
|---|---|---|
| Auth/session/OAuth cleanup | C001–C006 | 覆盖 success + invalid credential + expired refresh；cookie/header 行为必须固定 |
| API key/group/rate multiplier refactor | C007–C012 | 覆盖 null group、dynamic group、budget multiplier、invalid group 错误 |
| Usage/billing/dashboard refactor | C013–C018 | 覆盖 request_type 兼容、usage dashboard、gateway billing path；C018 是硬门槛 |
| Subscription/redeem service refactor | C019–C025 | 覆盖 legacy subscription 和 product subscription 双路径 |
| Affiliate/invite refactor | C026–C027 | 覆盖 quota transfer、ledger/audit 字段 |
| Admin service split | C028–C033 | 至少 users/groups/accounts/settings/redeem/payment dashboard 通过 |
| Payment service/provider refactor | C034–C036 + verify/webhook `[待 T007 补]` | provider-specific payload 必须 golden；webhook 必须有 deterministic signature stub |
| Gateway compatibility refactor | C018 + gateway models/messages/responses 用例 `[待 T007 补]` | 明确 xlabapi envelope 与 OpenAI-compatible 非 envelope 的边界 |

### 4.3 语义去双轨门槛

| 双轨主题 | 必须先有的 contract | 保护点 |
|---|---|---|
| subscription vs product subscription | C019、C020、C021、C022、C024 | 前端和后端可以重命名/收敛内部模型，但 HTTP 响应中的兼容字段不能断 |
| redeem balance vs subscription code | C023、C024、C025、C032 | 兑换类型、权益字段、admin list/generate 字段必须稳定 |
| group rate multiplier / dynamic pricing | C011、C012、C029、C018 | group/rate 字段、key binding、gateway billing multiplier 行为必须稳定 |
| usage request_type vs legacy stream/openai_ws_mode | C013、C014、C017 | 旧字段与新字段的可见性、优先级和 fallback 不能漂移 |
| payment visible method vs provider instance | C031、C033、C034、C035、C036 | settings/payment config/order snapshot 的来源权威必须固定 |

### 4.4 每批测试的退出条件

| 批次 | 覆盖率/通过率 | 手工冒烟 | 文档同步 |
|---|---|---|---|
| 第一批 | 8/8 自动化通过；至少 2 个错误路径 golden | 登录、创建 key、dashboard stats、一次 gateway chat completion、admin settings | 更新 `API-CONTRACT.md` 的 auth/keys/usage/gateway/admin settings 条目 |
| 第二批 | 9/9 自动化通过；balance redeem + product redeem 都覆盖 | 订阅、进度、兑换、历史、返利 summary/transfer | 更新 subscription/redeem/affiliate contract 字段 |
| 第三批 | admin core 自动化通过；list/detail 至少覆盖 empty + non-empty | users/groups/accounts/redeem/payment dashboard 页面读写基本可用 | 更新 admin pagination/filter/sort 规则 |
| 第四批 | payment 自动化通过；至少一个 provider success + 一个 invalid token/error | payment config、create order、resolve/verify result | 更新 payment provider payload/resume token/webhook 规则 |

### 4.5 执行规则

- 每个 contract 用例必须有编号，编号进入 golden 文件名、测试名、文档索引，方便 review 和失败定位。
- 每个用例必须写清楚动态字段 normalize 策略；没有 normalize 策略的用例不得落地。
- 每批合入前必须跑对应 Go package 测试；frontend-v2 MSW tests 可以后置，但不能与后端 golden 字段冲突。
- 如果 T007 `API-CONTRACT.md` 与本文端点/字段不一致，以 T007 为权威；本文应更新编号或检查点，不允许测试实现私自选择第三种契约。
- 如果 `.gitignore` 继续忽略 `docs/superpowers/`，交付仍可本地读取，但最终是否入 repo 需要用户/foreman 决策。
