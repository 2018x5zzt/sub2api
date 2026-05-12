# Test Coverage Inventory + P0 Contract Test Gaps

- 审计目标：只读盘点 `origin/test/xlabapi` 分支上现有后端/前端测试覆盖，定位 frontend-v2/后端重构前必须钉住的 P0 HTTP 契约测试缺口。
- 审计基线：当前工作树为 `main`，但本轮对齐目标在 `test/xlabapi`；因此后端测试和 frontend-v2 代码主要通过 `git ls-tree/show origin/test/xlabapi` 读取，老前端对照通过 `origin/xlabapi:frontend/` 读取。
- 总览：`origin/test/xlabapi` 后端共有 531 个 `*_test.go`；当前 `main` 后端为 386 个 `*_test.go`。后端测试集中在 service/repository/handler 层，已有少量 server contract/e2e。`frontend-v2` 没有测试框架和测试文件；老前端 `frontend/` 有 Vitest 测试 107 个。

---

## 1. 后端 Go 测试现状

### 1.1 文件分布与集成级别

- `find backend -name '*_test.go' | head -50` 在当前工作树能看到 handler/admin/gateway/redeem/auth 等测试；`origin/test/xlabapi` 上测试规模更大，包含 `backend/internal/server/api_contract_test.go`、`backend/internal/integration/e2e_gateway_test.go`、`backend/internal/integration/e2e_user_flow_test.go`。
- 目录分布（当前 `main` 快照）：`backend/internal/service` 209 个、`backend/internal/repository` 66 个、`backend/internal/handler` 25 个、`backend/internal/handler/admin` 23 个、`backend/internal/server/middleware` 10 个、`backend/internal/integration` 3 个。
- 集成/E2E 级别存在但不均衡：repository 层大量 `*_integration_test.go`；gateway 和用户注册/key 生命周期有 e2e；server contract 只有一个 `backend/internal/server/api_contract_test.go`，主要覆盖少量用户/管理端响应 JSON。
- 已有 contract 测试覆盖的 HTTP 路径：`GET /api/v1/auth/me`、`POST /api/v1/keys`、`GET /api/v1/keys`、`GET /api/v1/groups/available`、`GET /api/v1/subscriptions`、`GET /api/v1/redeem/history`、`GET /api/v1/usage/stats`、`GET /api/v1/usage`、`GET /api/v1/admin/settings`、`POST /api/v1/admin/accounts/bulk-update`。

### 1.2 子系统测试盘点

| 子系统 | test 文件数 | 总行数 | 关键函数覆盖 | 缺口 |
|---|---:|---:|---|---|
| billing | 16 | 3,851 | `CalculateCost*`、`CalculateImageCost*`、`BillingCacheService.CheckRPM`、usage billing dedupe、subscription billing command、gateway billing error mapping | 业务/仓储覆盖较强；缺少用户侧/admin 侧 billing HTTP 响应 golden；支付/订阅扣费跨端到端链路未形成稳定契约快照 |
| subscription | 13 | 3,436 | product subscription repo/service、assign/extend/idempotency、progress/reset quota、subscription auth quota、redeem/payment fulfillment | service/repo 覆盖强；仅 `GET /api/v1/subscriptions` 在 contract 中钉住，`active/progress/summary`、admin subscription/product routes 缺 golden |
| redeem | 8 | 1,187 | redeem cache/repo、admin generate/create-and-redeem validation、product subscription redeem validation | 用户 `GET /redeem/history` 有 contract；`POST /redeem` 和 admin redeem-code list/generate/export/delete/expire 缺 HTTP contract |
| affiliate | 3 | 726 | affiliate repo transfer/accrual/custom code/rebate rate、service rate/masking/format validation | 基本是 repository/service；用户 `/user/aff`、`/user/aff/transfer`、`/invite/*` 和 admin affiliate/invite HTTP 契约未钉 |
| group | 12 | 3,598 | group repo/sort、allowed_groups cascade、admin group create/update/dynamic pricing/rate multipliers、available groups、gateway group isolation | 有 `GET /groups/available` contract；admin groups CRUD、rates、capacity/usage summary、group API keys 缺 golden；动态倍率字段需强制钉响应 shape |
| auth | 47 | 22,249 | register/login/OAuth pending flow、LinuxDo/OIDC/WeChat/Xlab OAuth、session revocation、JWT/admin/api-key auth middleware、auth identity migrations/cache | 覆盖很广但偏 handler/service 单测；contract 只钉 `GET /auth/me`，未钉 `login/register/refresh/logout/2fa/oauth pending` 的状态码/cookie/token envelope |
| openai | 73 | 32,436 | OpenAI gateway responses/chat/images、Codex/WS/passthrough、OAuth/token provider、scheduler、rate limit、usage record、model mapping | 核心网关实现测试多；但 OpenAI-compatible 对外 HTTP 契约多在 handler/service 层，缺 `/v1/*` 与根路径 `/chat/completions`/`/responses` golden-file 响应/错误快照 |
| channel | 13 | 5,559 | channel repo/service、admin channel DTO/pricing、available channel whitelist、channel monitor、gateway/openai channel restriction | 用户 `/channels/available` 和 admin channels/monitors 缺 contract；字段白名单已有 handler 单测但未作为 HTTP snapshot 固定 |
| backup | 1 | 703 | `backup_service_test.go` 覆盖备份服务逻辑 | 只有 service 层；admin `/backups`、`/data-management/backups`、S3 config/test/download/restore HTTP 契约未钉 |
| multiplier | 2 | 90 | account billing rate multiplier default/zero/negative、billing negative multiplier clamp | 覆盖很薄；倍率散落在 billing/group/API key/service 逻辑，缺 HTTP 层对 group rate multipliers、dynamic pricing、usage cost 字段的契约测试 |

### 1.3 已存在的 integration / e2e

| 级别 | 文件 | 覆盖 |
|---|---|---|
| Contract/unit | `backend/internal/server/api_contract_test.go` | 少量 `/api/v1` 用户/管理端响应 JSON，使用 stub deps + httptest |
| E2E | `backend/internal/integration/e2e_gateway_test.go` | Claude/Gemini models/messages/count/generateContent，覆盖网关行为 |
| E2E | `backend/internal/integration/e2e_user_flow_test.go` | 用户注册登录、API key lifecycle |
| Repository integration | `backend/internal/repository/*_integration_test.go` | Ent/DB/cache/migration/排序/幂等/使用记录/订阅/支付相关仓储逻辑 |
| Middleware/server integration | `backend/internal/server/middleware/*_test.go`、`backend/internal/server/routes/*_test.go` | JWT/admin/API key/CORS/body limit/rate limit 等中间件和路由级逻辑 |

**结论**：后端单元和仓储集成测试并不少，问题不是“没测试”，而是 P0 外部 HTTP 契约只被零散覆盖。对 frontend-v2/后端重构最危险的是：已有 handler/service 测试能保证内部逻辑，却不能保证前端依赖的 status code、JSON envelope、字段名、null/empty 语义、分页结构、cookie/token 行为不变。

---

## 2. 前端测试现状

### 2.1 frontend-v2

| 项 | 现状 |
|---|---|
| 测试文件 | 无 `*.test.ts(x)`、`*.spec.ts(x)` |
| Vitest | 无 `vitest.config.*`，`package.json` 无 `test`/`test:run` 脚本，无 `vitest`/`jsdom`/testing-library 依赖 |
| Playwright | 无 `playwright.config.*`，无 e2e 脚本 |
| 当前可验证项 | 只有 `typecheck` (`tsc -b`) 和 `build` (`tsc -b && vite build`) |
| API client 规模 | `frontend-v2/src/api/` 已有 30 个 TS 文件，例如 `client.ts`、`auth.ts`、`payment.ts`、`usage.ts`、`keys.ts`、`admin/*`，但没有 API client tests |

一句话总结：`frontend-v2` 基本没有测试体系，连最小 API client 契约测试都没有；当前只能靠 TypeScript 编译和人工验证。

### 2.2 老前端 frontend/

| 项 | 现状 |
|---|---|
| 测试框架 | `frontend/vitest.config.ts` + `vitest`/`jsdom`/`@vue/test-utils` |
| 脚本 | `test`、`test:run`、`test:coverage` |
| 测试文件数 | `origin/xlabapi:frontend/` 下 107 个 `*.spec.*` / `*.test.*` |
| 覆盖类型 | api client、router guards/title、stores、composables、components、views、少量 integration navigation/data-import |
| API client tests | `frontend/src/api/__tests__/client.spec.ts`、`keys.spec.ts`、`payment.spec.ts`、`user.spec.ts`、`admin.redeem.spec.ts`、`admin.users.spec.ts`、settings/auth oauth adoption 等 |

一句话总结：老 Vue 前端有较完整的 Vitest 单元/组件/API client 测试，但不是 frontend-v2 可直接复用的执行资产；最有价值的是把 API client expectations 和页面流程场景迁移成 React/Vitest + MSW 用例。

---

## 3. P0 契约测试缺口

原则：frontend-v2 可以重写、后端可以重构，但对外 HTTP 契约不能变。这里的“契约”至少包括 method/path、认证方式、status code、统一响应 envelope、分页结构、字段名、字段类型、null vs empty array/object、错误码/message、cookie/header/token 行为。

### 3.1 现有 contract 覆盖基线

`backend/internal/server/api_contract_test.go` 已有 contract 用例，但范围很窄：

- 已覆盖：`GET /api/v1/auth/me`、`POST /api/v1/keys`、`GET /api/v1/keys`、`GET /api/v1/groups/available`、`GET /api/v1/subscriptions`、`GET /api/v1/redeem/history`、`GET /api/v1/usage/stats`、`GET /api/v1/usage`、`GET /api/v1/admin/settings`、`POST /api/v1/admin/accounts/bulk-update`。
- 未覆盖：大多数 auth mutation、user profile mutation、payment flow、redeem submit、affiliate/invite、admin CRUD/list/detail/export、gateway `/v1/*` OpenAI-compatible endpoints。
- 质量判断：这是好的起点，但还不是“P0 contract safety net”。它只保护少数 happy path 响应 JSON，没有形成按端点组的 golden-file/snapshot 矩阵。

### 3.2 按 API 端点组的缺口清单

| API 组 | 代表端点 | 当前契约测试 | 建议测试级别 | 必须钉住的契约 |
|---|---|---|---|---|
| auth | `POST /api/v1/auth/login`、`/register`、`/refresh`、`/logout`、`/login/2fa`、`/send-verify-code`、`/forgot-password`、`/reset-password`、`GET /auth/oauth/*/start/callback`、`POST /auth/oauth/pending/*`、`GET /api/v1/auth/me` | 部分：`GET /auth/me` 有 JSON contract；OAuth/2FA/login/register 多为 handler/service 单测 | contract + smoke；OAuth callback 可加 focused e2e | token 响应字段、refresh cookie/headers、401/403/429 envelope、2FA challenge shape、OAuth pending state、redirect URL/cookie 名称、`/auth/me` identity binding 字段 |
| user | `GET /api/v1/user/profile`、`PUT /api/v1/user`、`PUT /api/v1/user/password`、`/user/account-bindings/*`、`/user/totp/*`、`/user/notify-email/*` | 无专门 contract；有 `user_handler_test.go` 和 auth current user 单测 | contract | profile 字段兼容名、identity summaries、avatar/email fields、password change error shape、TOTP status/setup/enable/disable payload、account binding null/boolean 语义 |
| keys | `GET/POST/PUT/DELETE /api/v1/keys`、`GET /api/v1/keys/:id` | 部分：list/create contract；API key lifecycle e2e | contract + e2e smoke | list pagination shape、created key 一次性明文、quota/rate limit/window 字段、group_id null、IP whitelist/blacklist array/null、delete/update status |
| usage | `GET /api/v1/usage`、`GET /api/v1/usage/:id`、`/usage/stats`、`/usage/dashboard/*`、admin `/usage/*` | 部分：user list/stats contract；handler/repo/service 有 request_type/sort/filter 测试 | contract | list pagination、stats numeric fields、request_type vs stream 兼容、model/requested_model/upstream_model 字段可见性、dashboard trend/models/api-key usage shape、admin cleanup task shape |
| subscriptions | `GET /api/v1/subscriptions`、`/active`、`/progress`、`/summary`、`/subscription-products/*`、admin `/subscriptions/*`、admin `/subscription-products/*` | 部分：user `GET /subscriptions` contract；service/repo 覆盖强 | contract | active/progress/summary 字段、product family/group binding、quota/reset windows、expired window normalization、admin assign/bulk/extend/reset/revoke 响应和错误 |
| redeem | `POST /api/v1/redeem`、`GET /api/v1/redeem/history`、frontend-v2 还引用 `/redeem/benefit-leaderboard`、admin `/redeem-codes/*` | 部分：history contract；admin/service/repo 单测 | contract + smoke | redeem 成功/重复/过期/无效响应、balance vs subscription code 字段、history pagination、admin generate/create-and-redeem/list/export/delete/expire payload；确认 `/redeem/benefit-leaderboard` 是否仍存在或是 frontend-v2 残留 |
| affiliate / invite | user `/user/aff`、`/user/aff/transfer`、`/invite/summary`、`/invite/rewards`、admin `/affiliates/*`、`/invites/*` | 无 HTTP contract；repo/service 测试覆盖返利计算 | contract | summary/quota/claimed/frozen fields、transfer response、invite rewards pagination、admin records filters/sort、custom code/rate settings shape |
| payment | user `/payment/config`、`/checkout-info`、`/plans`、`/channels`、`/limits`、`/orders/*`、public `/payment/public/orders/*`、webhook `/payment/webhook/*`、admin `/admin/payment/*` | 无 server contract；service 层 payment 测试较多，handler resume/webhook 有局部测试 | contract + focused e2e | order create response、JSAPI/Stripe/Alipay/Wxpay provider payload、resume token、public verify/resolve shape、webhook success text/status、admin order/plan/provider CRUD payload、pending/paid/canceled/refund status |
| admin/* | `/admin/dashboard/*`、`/admin/users/*`、`/admin/groups/*`、`/admin/accounts/*`、`/admin/proxies/*`、`/admin/settings/*`、`/admin/backups/*`、`/admin/channels/*`、`/admin/channel-monitors/*`、`/admin/ops/*` | 部分：`GET /admin/settings`、`POST /admin/accounts/bulk-update` contract；大量 handler/service 单测 | contract，少量 smoke | 统一 admin envelope、pagination/filter/sort 参数、list/detail item shape、bulk mutation result shape、settings default injection、export/import response、ops dashboard snapshot/overview fields、backup S3 config/download/restore shape |
| gateway / openai-compatible | `/v1/messages`、`/v1/messages/count_tokens`、`/v1/models`、`/v1/usage`、`/v1/responses`、`/v1/chat/completions`、`/v1/images/*`、`/v1beta/models/*`、root `/responses`/`/chat/completions`/`/images/*` | 部分：gateway e2e + extensive service/handler tests；缺统一 golden contract suite | contract + e2e | OpenAI/Anthropic/Gemini-compatible non-envelope JSON/SSE shape、stream headers/events、error status/body、model list format、auth failure body、body limit behavior、endpoint normalization、root vs `/v1` compatibility |

### 3.3 P0 缺口排序

| 优先级 | 缺口 | 原因 |
|---|---|---|
| P0-1 | auth/login/register/refresh/logout/2FA/OAuth pending contract | 任何字段/cookie/token 变化都会直接让 frontend-v2 登录态断裂 |
| P0-2 | keys + usage + subscriptions/redeem user-side golden | 这是用户 console 核心工作流：创建 key、看用量、看订阅、兑换权益 |
| P0-3 | payment order create/verify/public resolve + admin payment config/plans | 支付路径状态复杂，前端强依赖 provider-specific payload |
| P0-4 | admin list/detail/mutation 基础 CRUD contract | admin 页面多，重构时最容易发生分页/字段/空值语义漂移 |
| P0-5 | gateway `/v1/*` compatibility snapshots | 对外 API 消费者不走 frontend-v2，但属于更高优先级的外部兼容红线 |

---

## 4. 建议路线

### 4.1 重构前必须先补的最小测试集

目标不是“测试覆盖率好看”，而是在 frontend-v2 和后端继续改之前，把最容易破坏外部行为的 HTTP 契约钉住。建议先补一组 `backend/internal/server/api_contract_test.go` 风格的 golden-file/snapshot 用例，不接真实外部服务，使用 stub deps + `httptest`。

| 批次 | 必补端点 | 建议形式 | 验收点 |
|---|---|---|---|
| 1 | `POST /api/v1/auth/login`、`POST /api/v1/auth/register`、`POST /api/v1/auth/refresh`、`POST /api/v1/auth/logout`、`POST /api/v1/auth/login/2fa`、`GET /api/v1/auth/me` | contract golden + auth smoke | 成功/失败 status、envelope、token/cookie/header、2FA challenge、identity binding 字段固定 |
| 2 | `GET/POST/PUT/DELETE /api/v1/keys`、`GET /api/v1/groups/available`、`GET /api/v1/groups/rates` | contract golden + lifecycle smoke | key list/create/update/delete、pagination、key 明文字段、group/rate/dynamic pricing 字段固定 |
| 3 | `GET /api/v1/usage`、`GET /api/v1/usage/stats`、`GET /api/v1/usage/dashboard/{stats,trend,models}` | contract golden | request_type/model/source/service_tier 字段、分页、时间序列、空列表/零值语义固定 |
| 4 | `GET /api/v1/subscriptions`、`/active`、`/progress`、`/summary`、`GET /api/v1/subscription-products/{active,progress,summary}` | contract golden | subscription/product subscription 双轨字段、quota/progress/window/reset 字段固定 |
| 5 | `POST /api/v1/redeem`、`GET /api/v1/redeem/history`、`GET /api/v1/user/aff`、`POST /api/v1/user/aff/transfer`、`GET /api/v1/invite/{summary,rewards}` | contract golden | 兑换成功/失败、权益类型、返利余额/冻结/可转出字段、列表 pagination 固定 |
| 6 | `GET /api/v1/payment/{config,plans,channels,limits,checkout-info}`、`POST /api/v1/payment/orders`、`POST /api/v1/payment/orders/verify`、`POST /api/v1/payment/public/orders/{resolve,verify}` | contract + one focused e2e | provider-specific payload、resume token、订单状态、支付回跳/公开查询响应固定 |
| 7 | admin P0：`GET /admin/settings`、`PUT /admin/settings`、`GET /admin/users`、`GET /admin/groups`、`GET /admin/accounts`、`GET /admin/redeem-codes`、`GET /admin/payment/{dashboard,config,orders,plans,providers}` | contract golden | admin envelope、pagination/filter/sort、settings default injection、bulk mutation result shape 固定 |
| 8 | gateway P0：`GET /v1/models`、`POST /v1/messages`、`POST /v1/messages/count_tokens`、`POST /v1/chat/completions`、`POST /v1/responses`、`GET /v1beta/models` | compatibility snapshots + existing e2e reuse | 非 xlabapi envelope 的兼容响应、SSE headers/events、auth/error body、root path alias 行为固定 |

### 4.2 测试框架选型建议

| 层 | 建议 | 理由 |
|---|---|---|
| 后端 contract | 继续 Go stdlib `testing` + `httptest` + `testify/require`；复用现有 `backend/internal/server/api_contract_test.go` 结构 | 已经存在先例，成本低；不引入新框架；适合 golden JSON、status/header/cookie 断言 |
| 后端 golden | JSON fixture 放 `backend/internal/server/testdata/contracts/*.golden.json`，测试中 normalize time/id 后比较 | 能直接 review diff；比散落 require 更适合防字段漂移 |
| 后端 integration | 保留现有 repository/service integration，不扩散到全量真实 DB e2e | 当前仓储集成已多，P0 缺口在 HTTP contract，不在 DB 逻辑 |
| frontend-v2 unit/API | Vitest + React Testing Library + MSW | 与 Vite/TS 栈兼容；MSW 可模拟 `/api/v1` 契约，优先保护 API client、hooks、关键页面数据态 |
| frontend-v2 route smoke | Vitest + MemoryRouter；少量 Playwright later | 先验证路由可渲染、auth guard、主要页面 loading/error/empty states；不要一开始拉起全浏览器矩阵 |
| 老前端迁移 | 从 `origin/xlabapi:frontend/src/api/__tests__` 和关键 view specs 提炼 expectations | 老前端 107 个测试是行为素材，不应照搬 Vue 测试实现 |

### 4.3 不建议此刻做的事

- 不建议先做全量 Playwright e2e：页面和 frontend-v2 parity 仍在变，维护成本高，且不能替代 HTTP contract。
- 不建议追求覆盖率阈值：当前问题是关键契约未钉，不是覆盖率数字低。
- 不建议把所有 handler/service 单测改成 contract 测试：已有内部测试有价值，补 P0 HTTP 层即可。
- 不建议在 frontend-v2 先写大量组件 snapshot：UI 还在风格对齐阶段，snapshot 会制造噪音。
- 不建议让 contract fixture 依赖真实时间、真实随机 key、真实外部支付/OAuth/OpenAI 服务：必须 stub/normalize，否则测试会不稳定。

### 4.4 推荐执行顺序

1. 先扩展 `backend/internal/server/api_contract_test.go`：auth + keys + usage + subscriptions + redeem，覆盖 frontend-v2 用户 console 的最小闭环。
2. 再补 payment contract：订单创建/verify/public resolve 是支付页面改造前的硬门槛。
3. 再补 admin P0 list/detail/mutation：先 dashboard/settings/users/groups/accounts/redeem/payment，不追全量 ops。
4. 同步给 frontend-v2 加 Vitest + MSW，先测 `src/api/client.ts` token refresh/error envelope，再测 `auth/keys/usage/subscriptions/payment` API modules。
5. 最后把 gateway compatibility snapshots 独立成 suite，避免与 xlabapi envelope contract 混在一起。

**最终判断**：后端已有大量内部测试，但 frontend-v2 重写最需要的是“HTTP 契约防漂移”。最小有效动作是先补 30 到 40 个 golden contract 用例，而不是马上铺开全量 e2e 或 UI snapshot。
