# frontend-v2 功能 / 路由 / API 对齐矩阵

- 审计目标：对照 `origin/xlabapi:frontend` 旧 Vue 前端与 `origin/test/xlabapi:frontend-v2` React 前端，盘点旧功能入口、API 端点、frontend-v2 覆盖状态与迁移风险。
- 审计范围：只读 git；旧路由以 `origin/xlabapi:frontend/src/router/index.ts` 为准；新路由以 `origin/test/xlabapi:frontend-v2/src/router/index.tsx` 为准；API 以两侧 `src/api/**` 与页面直接调用为参考。
- 实现深度标记：`完整/接近完整` 表示页面和主要 API 已落地；`薄实现` 表示有页面和主要读写 API，但业务细节明显少于旧 Vue；`占位` 表示使用 `ParityPlaceholder` 或功能仅提示；`重定向` 表示兼容路径存在但不承载独立功能；`缺失/错配` 表示入口或 API 契约仍有 P0 风险。
- 参考：`docs/superpowers/audits/2026-05-12-frontend-v2-style-gap.md` 为并列 T002 视觉审计，本文件只盘点功能/路由/API。

---

## 1. 基线复用

来自 `origin/test/xlabapi:docs/superpowers/plans/2026-05-11-frontend-v2-xlabapi-parity.md` 与 `origin/test/xlabapi:docs/superpowers/specs/2026-05-11-frontend-v2-xlabapi-parity-design.md` 的既有 P0 结论如下，作为本审计起点，不重复推导。

### 1.1 P0 路由与导航缺口基线

| 分组 | 旧 xlabapi 入口 | 既有基线结论 |
|---|---|---|
| Public | `/docs` | Landing 有文档入口，frontend-v2 需提供公共文档页。 |
| Public | `/setup` | 旧安装向导入口必须恢复，首批允许占位。 |
| Public | `/key-usage` | 旧公开 API key 用量查询入口必须恢复，首批允许占位。 |
| User | `/available-channels` | 用户可用渠道/模型/价格入口缺失。 |
| User | `/monitor` | 用户渠道状态入口缺失。 |
| User | `/purchase` | 用户购买订阅入口缺失。 |
| User | `/orders` | 用户订单列表入口缺失。 |
| User | `/affiliate` | 用户邀请/返利入口缺失。 |
| User | `/invite -> /affiliate` | 旧邀请兼容重定向缺失。 |
| User | `/custom/:id` | 自定义菜单 iframe 行为缺失，首批允许占位。 |
| User | `/image-studio` | 旧 Sora/图片创作兼容入口缺失。 |
| Payment/Auth | `/payment/qrcode`, `/payment/result`, `/payment/stripe`, `/payment/stripe-popup` | 支付流程页缺失。 |
| Auth/OAuth | `/auth/wechat/callback`, `/auth/wechat/payment/callback`, `/auth/oidc/callback`, `/oauth/consent` | WeChat/OIDC/Xlab OAuth 回调和授权入口缺失。 |
| Admin | `/admin/dashboard`, `/admin -> /admin/dashboard` | 管理首页和兼容跳转需恢复。 |
| Admin | `/admin/ops` | 运维监控入口缺失。 |
| Admin | `/admin/channels -> /admin/channels/pricing`, `/admin/channels/pricing`, `/admin/channels/monitor` | 渠道价格和监控入口缺失。 |
| Admin | `/admin/proxies` | 代理管理入口缺失。 |
| Admin | `/admin/subscription-products -> /admin/subscriptions`, `/admin/subscription-product-config` | 订阅产品兼容入口和产品配置入口缺失。 |
| Admin | `/admin/invites -> /admin/users` | 邀请管理旧路径兼容跳转缺失。 |
| Admin | `/admin/affiliates -> /admin/affiliates/invites`, `/admin/affiliates/invites`, `/admin/affiliates/rebates`, `/admin/affiliates/transfers` | 管理端返利记录入口缺失。 |
| Admin | `/admin/orders/dashboard`, `/admin/orders`, `/admin/orders/plans` | 支付管理入口缺失。 |
| Admin | `/admin/backup` | 管理导航需包含备份入口。 |

### 1.2 P0 API 契约修正基线

| API 域 | 既有基线要求 |
|---|---|
| User usage | `/usage`, `/usage/stats`, `/usage/dashboard/stats`, `/usage/dashboard/trend`, `/usage/dashboard/models`。 |
| Admin usage | `/admin/usage`, `/admin/usage/stats`。 |
| Admin dashboard | `/admin/dashboard/stats`。 |
| Backup | `/admin/backups`, `/admin/backups/s3-config`, `/admin/backups/s3-config/test`, `/admin/backups/schedule`, `/admin/backups/:id/download-url`。 |
| Settings SMTP | `/admin/settings/test-smtp`, `/admin/settings/send-test-email`。 |
| Accounts | `POST /admin/accounts/:id/schedulable`, `POST /admin/accounts/:id/refresh`。 |
| Admin subscriptions | `DELETE /admin/subscriptions/:id`。 |
| Groups | `/groups/available`, `/groups/rates`。 |
| New thin clients | payment、channels、channel monitor、affiliate、OAuth、admin channels、admin channel monitor、admin proxies、admin payment、admin affiliate、admin ops。 |

### 1.3 与当前 `origin/test/xlabapi` 的差异快照

本次复核发现 `origin/test/xlabapi:frontend-v2` 已经包含一批基线要求产物：`/docs`、多数用户/支付/OAuth/admin 路由、`ParityPlaceholder`、以及 payment/channels/channelMonitor/affiliate/oauth/admin 相关 thin clients 已存在。剩余风险因此从“入口 404”转为“页面实现深度、兼容行为、API 覆盖完整度、路由守卫语义”四类。


---

## 2. 路由源清单

### 2.1 旧 xlabapi Vue 路由源

读取命令：`git ls-tree -r origin/xlabapi -- frontend/src/router`。实际路由定义集中在 `frontend/src/router/index.ts`，其他文件为 README、meta 类型、title helper 和测试。

| 文件 | 用途 |
|---|---|
| `frontend/src/router/index.ts` | 旧 Vue Router 主配置，包含 public/user/admin/payment/OAuth 路由、导航守卫、backend mode 例外、simple mode 限制、payment gate。 |
| `frontend/src/router/meta.d.ts` | route meta 类型，包含 `requiresAuth`、`requiresAdmin`、`requiresPayment`、`titleKey`、`descriptionKey`。 |
| `frontend/src/router/title.ts` | 页面标题生成 helper。 |
| `frontend/src/router/__tests__/*.spec.ts` | 路由守卫、标题、WeChat、admin subscriptions 等路由测试。 |

旧 Vue 路由总览：

| 分组 | 路由 |
|---|---|
| Setup/Public | `/setup`, `/home`, `/login`, `/register`, `/email-verify`, `/forgot-password`, `/reset-password`, `/key-usage`, `/` -> `/home` |
| Auth/OAuth | `/auth/callback`, `/auth/linuxdo/callback`, `/auth/wechat/callback`, `/auth/wechat/payment/callback`, `/auth/oidc/callback`, `/oauth/consent` |
| User Core | `/dashboard`, `/keys`, `/models`, `/usage`, `/redeem`, `/profile`, `/subscriptions` |
| User Commercial/Ops | `/invite` -> `/affiliate`, `/affiliate`, `/available-channels`, `/monitor`, `/purchase`, `/orders`, `/image-studio`, `/custom/:id` |
| Payment | `/payment/qrcode`, `/payment/result`, `/payment/stripe`, `/payment/stripe-popup` |
| Admin Core | `/admin` -> `/admin/dashboard`, `/admin/dashboard`, `/admin/users`, `/admin/groups`, `/admin/accounts`, `/admin/announcements`, `/admin/redeem`, `/admin/promo-codes`, `/admin/settings`, `/admin/usage` |
| Admin Ops/Channels | `/admin/ops`, `/admin/channels` -> `/admin/channels/pricing`, `/admin/channels/pricing`, `/admin/channels/monitor`, `/admin/proxies` |
| Admin Subscriptions | `/admin/subscriptions`, `/admin/subscription-products` -> `/admin/subscriptions`, `/admin/subscription-product-config` |
| Admin Affiliate/Payment | `/admin/invites` -> `/admin/users`, `/admin/affiliates` -> `/admin/affiliates/invites`, `/admin/affiliates/invites`, `/admin/affiliates/rebates`, `/admin/affiliates/transfers`, `/admin/orders/dashboard`, `/admin/orders`, `/admin/orders/plans` |
| 404 | `/:pathMatch(.*)*` |

旧 Vue 路由守卫行为也属于 parity 范围：未登录保护页跳 `/login?redirect=...`；admin route 需要 admin；payment route 受 public settings `payment_enabled` 控制；simple mode 限制部分订阅/兑换/group 路径；backend mode 仅允许 `/login`、`/key-usage`、`/setup`、`/payment/result` 和 OAuth callback 等白名单。

### 2.2 frontend-v2 React 路由源

读取命令：`git show origin/test/xlabapi:frontend-v2/src/router/index.tsx`。当前 React 路由已经包含以下覆盖：

| 分组 | frontend-v2 当前状态 |
|---|---|
| Public | `/`, `/home` -> `/`, `/docs`, `/setup` 占位, `/key-usage` 占位。 |
| Auth | `/login`, `/register`, `/email-verify`, `/forgot-password`, `/reset-password`, `/auth/callback`, `/auth/linuxdo/callback`, `/auth/wechat/callback`, `/auth/wechat/payment/callback`, `/auth/oidc/callback`。 |
| User | `/dashboard`, `/keys`, `/usage`, `/models`, `/subscriptions`, `/redeem`, `/profile`, `/invite` -> `/affiliate`, `/affiliate`, `/available-channels`, `/monitor`, `/image-studio`, `/purchase`, `/orders`, `/payment/qrcode`, `/custom/:id` 占位, `/oauth/consent`。 |
| Payment public-ish | `/payment/result`, `/payment/stripe`, `/payment/stripe-popup` 在 auth layout 外。 |
| Admin | `/admin` -> `/admin/dashboard`, `/admin/dashboard`, `/admin/users`, `/admin/groups`, `/admin/accounts`, `/admin/usage`, `/admin/announcements`, `/admin/redeem`, `/admin/promo-codes`, `/admin/subscriptions`, `/admin/subscription-products` -> `/admin/subscriptions`, `/admin/subscription-product-config` 占位, `/admin/backup`, `/admin/settings`, `/admin/ops`, `/admin/channels` -> `/admin/channels/pricing`, `/admin/channels/pricing`, `/admin/channels/monitor`, `/admin/proxies`, `/admin/invites` -> `/admin/users`, `/admin/affiliates` -> `/admin/affiliates/invites`, `/admin/affiliates/invites`, `/admin/affiliates/rebates`, `/admin/affiliates/transfers`, `/admin/orders/dashboard`, `/admin/orders`, `/admin/orders/plans`。 |
| 404 | `*` -> `NotFoundPage`。 |

关键差异：React 路由通过 `<RequireAuth>`、`<RequireAdmin>`、`<RedirectIfAuthed>` 做基础守卫，但当前路由文件未体现旧 Vue 的 `requiresPayment`、simple mode 限制、backend mode 细粒度白名单、custom page document title / iframe 语义。需要确认这些是否在 `guards` 或 store 层补齐，否则属于行为 parity 风险。

---

## 3. 路由 / 功能 / API 对齐矩阵

### 3.1 Public / Setup / Auth

| xlabapi 路由 | 功能简述 | 用到的 API 端点 | frontend-v2 是否有 | 实现深度 | 风险 | 建议 |
|---|---|---|---|---|---|---|
| `/` -> `/home` | 旧首页重定向到 Home。 | `GET /settings/public` 等 public settings。 | 有差异：frontend-v2 `/` 是 Landing，`/home` -> `/`。 | 完整但语义变更 | 低：营销页入口变更可接受，但若旧用户记忆 `/home`，已兼容。 | 保留当前 `/home` 兼容即可。 |
| `/home` | 旧公共首页。 | `GET /settings/public`。 | 有，重定向到 `/`。 | 完整 | 低。 | 无需立刻动。 |
| `/docs` | 新基线要求：Landing 文档入口。 | 主要静态内容。 | 有：`DocsPage`。 | 薄实现 | 中：旧 Vue 无该路由，但 Landing 依赖；若文档内容过浅影响转化但非 P0。 | 可延后补文档深度。 |
| `/setup` | 初始化安装向导，数据库/Redis 测试和安装。 | `GET /setup/status`, `POST /setup/test-db`, `POST /setup/test-redis`, `POST /setup/install`。 | 有：`ParityPlaceholder standalone`。 | 占位 | 高：入口不 404，但无法完成安装流程；旧 Vue 是完整 wizard。 | 必须立刻迁移最小安装向导，不能只展示占位。 |
| `/key-usage` | 公共 API key 用量查询页，用户输入 key 后查 `/v1/usage`。 | `GET /v1/usage` with `Authorization: Bearer <key>`。 | 有：`ParityPlaceholder standalone`，列的是 `/usage/dashboard/*`。 | 占位且 API 标注错配 | 高：旧功能是 public key query，不是登录态 dashboard usage；当前占位列出的 API 不能替代。 | 必须修正文案/API 面，并实现最小表单调用 `/v1/usage`。 |
| `/login` | 登录，支持 2FA/public settings/OAuth 入口。 | `POST /auth/login`, `POST /auth/login/2fa`, `GET /settings/public`；旧版还根据设置展示 OAuth。 | 有。 | 接近完整 | 中：frontend-v2 auth client 缺部分 OAuth/pending helpers，需确认 Login UI 是否覆盖 WeChat/OIDC/Linode-like flows。 | 与 OAuth callback 一起补齐第三方登录入口和 pending flow。 |
| `/register` | 注册、验证码、邀请码/返利码。 | `POST /auth/register`, `POST /auth/send-verify-code`；旧版还有 `/auth/validate-promo-code`, `/auth/validate-invitation-code`。 | 有。 | 薄实现 | 中：基础注册有，旧版邀请码/促销码校验可能缺。 | 商业闭环前补邀请码/促销码校验。 |
| `/email-verify` | 邮箱验证码验证/绑定流程。 | `POST /auth/send-verify-code`，以及页面内验证相关调用。 | 有。 | 薄实现 | 中：需与旧 pending OAuth 邮箱验证联动，否则第三方注册会断。 | 并入 OAuth/pending flow 修复。 |
| `/forgot-password` | 忘记密码。 | `GET /settings/public`, `POST /auth/forgot-password`。 | 有。 | 接近完整 | 低。 | 延后。 |
| `/reset-password` | 重置密码。 | `POST /auth/reset-password`。 | 有。 | 接近完整 | 低。 | 延后。 |
| `/auth/callback` | 通用 OAuth callback。 | 通常读取 URL token/code/state，部分流程需 `/auth/oauth/pending/exchange`。 | 有：generic code/state copy page。 | 薄实现 | 中：如果生产流程需要自动 token 交换，当前只展示 copy 值不足。 | 核对后端回调模式，必要时补自动交换。 |
| `/auth/linuxdo/callback` | LinuxDo OAuth callback，支持 token fragment、邀请码补全。 | `POST /auth/oauth/linuxdo/complete-registration`。 | 有。 | 接近完整 | 低-中：只覆盖 LinuxDo complete registration，未覆盖 bind/exchange 全套。 | 与 profile account binding 一起补。 |
| `/auth/wechat/callback` | WeChat 登录 callback。 | 旧版使用 WeChat 专页和 pending OAuth helpers：`/auth/oauth/wechat/complete-registration`, `/auth/oauth/pending/exchange` 等。 | 有，但复用 generic `OAuthCallbackPage`。 | 占位/错配 | 高：旧版是自动处理 WeChat token、pending、邀请码；当前只是展示 code/state。 | 必须立刻补 WeChat callback 专页或复用完整 OAuth handler。 |
| `/auth/wechat/payment/callback` | WeChat 支付绑定/授权回调。 | WeChat payment OAuth start/callback，支付恢复相关 API。 | 有，但复用 generic `OAuthCallbackPage`。 | 占位/错配 | 高：支付场景回调不能只展示 code/state，可能导致微信支付恢复失败。 | 必须立刻补支付回调处理和 resume 行为。 |
| `/auth/oidc/callback` | OIDC 登录 callback。 | `POST /auth/oauth/oidc/complete-registration`, pending exchange/bind。 | 有，但复用 generic `OAuthCallbackPage`。 | 占位/错配 | 高：OIDC 登录/注册闭环未恢复。 | 必须立刻补 OIDC callback 自动处理。 |
| `/oauth/consent` | Xlab OAuth 授权同意页。 | `POST /oauth/authorize`。 | 有：`XlabOAuthConsentPage`。 | 薄实现 | 中：基础 allow/deny 有；旧版测试覆盖 redirect_uri/state，React 需补同等测试。 | 补契约测试和错误态处理。 |

### 3.2 User Core / Commercial / Payment

| xlabapi 路由 | 功能简述 | 用到的 API 端点 | frontend-v2 是否有 | 实现深度 | 风险 | 建议 |
|---|---|---|---|---|---|---|
| `/dashboard` | 用户仪表盘：用量、成本、余额/订阅摘要。 | `GET /usage/dashboard/stats`, `/usage/dashboard/trend`, `/usage/dashboard/models`。 | 有。 | 薄实现 | 中：页面可用但比旧版统计/趋势少。 | P0 保入口和关键统计；P1 补趋势和模型分布。 |
| `/keys` | API key 管理。 | `GET/POST/PUT/DELETE /keys`, batch key usage stats `/usage/dashboard/api-keys-usage`。 | 有。 | 接近完整 | 中：旧版动态预算/批量统计更深，需确认 frontend-v2 是否全覆盖。 | P1 补 key 用量细节和测试。 |
| `/models` | 模型中心 / 可用模型展示。 | 旧版主要用 `/channels/available`, `/groups/available`, `/groups/rates`。 | 有：`ModelHubPage`。 | 薄实现 | 中：有展示，但旧版可用渠道/订阅价格联动可能未全。 | 与 available channels 一起补价格语义。 |
| `/usage` | 用户用量记录。 | `GET /usage`, `GET /usage/stats`, dashboard stats/trend/models。 | 有。 | 薄实现 | 中：当前列表仅基础分页，缺旧版 filters/stats/export/图表。 | P1 补 filters/stats；P0 确认 `/usage/stats` client 是否需要页面使用。 |
| `/redeem` | 兑换码。 | `POST /redeem`, `/redeem/history`, `/redeem/benefit-leaderboard`。 | 有。 | 接近完整 | 低-中：需确认 leaderboard/history UI 深度。 | 延后。 |
| `/profile` | 用户资料、密码、OAuth 绑定、通知邮箱、TOTP 等。 | `GET /user/profile`, `PUT /user`, `PUT /user/password`, `/user/notify-email/*`, `/user/account-bindings/*`, TOTP APIs, OAuth bind start。 | 有。 | 薄实现 | 高：React 当前只含 username/password；旧版 profile 是账号安全和绑定中心。 | 必须列入 P0/P1 之间：至少补 OAuth binding、TOTP、通知邮箱入口，避免安全设置回退。 |
| `/subscriptions` | 我的订阅、进度、产品摘要、默认组切换。 | `GET /subscriptions`, `/subscriptions/active`, `/subscriptions/progress`, `/subscriptions/summary`, `/subscription-products/*`, `PUT /user`。 | 有。 | 薄实现 | 中-高：React 有订阅页但 API client 缺 `subscriptionProductsAPI`，产品购买/默认组绑定可能不完整。 | 与 purchase/subscription product 一起补齐。 |
| `/invite` -> `/affiliate` | 邀请旧路径兼容。 | 同 `/affiliate`。 | 有重定向。 | 重定向 | 低。 | 保留。 |
| `/affiliate` | 邀请返利、邀请码、返利转余额。 | `GET /user/aff`, `POST /user/aff/transfer`。 | 有。 | 接近完整 | 中：基础功能在；旧版 invitee/冻结额度/倍率文案需核对。 | P1 补测试。 |
| `/available-channels` | 用户可见渠道、平台、模型、价格。 | `GET /channels/available`, `GET /groups/available`, `GET /groups/rates`。 | 有。 | 薄实现 | 中：React 仅用 `/channels/available`，未显式使用 `/groups/rates`；价格倍率语义可能不足。 | P1 补倍率解释和 groups/rates fallback。 |
| `/monitor` | 用户渠道状态。 | `GET /channel-monitors`, `GET /channel-monitors/:id/status`。 | 有。 | 薄实现 | 中：旧版 detail/历史/筛选更深，React 有列表和 modal。 | 延后补历史和筛选；P0 确认 API list path 已正确。 |
| `/purchase` | 购买订阅/余额、选择支付渠道、创建订单。 | `GET /payment/config`, `/payment/plans`, `/payment/channels`, `/payment/checkout-info`, `/payment/limits`, `POST /payment/orders`。 | 有。 | 薄实现 | 高：商业主流程有页面，但旧版 WeChat/Stripe/支付限制/恢复处理复杂；React 需端到端验证。 | 必须立刻做支付 happy path 契约测试和 WeChat/Stripe route spot-check。 |
| `/orders` | 我的订单、取消、退款申请。 | `GET /payment/orders/my`, `GET /payment/orders/:id`, `POST /payment/orders/:id/cancel`, `POST /payment/orders/:id/refund-request`。 | 有。 | 薄实现 | 中：基础列表/操作有；退款 provider eligibility 可能未完整接入。 | P1 补退款细节。 |
| `/image-studio` | 旧图片创作 / Miku iframe 入口。 | 外部 `https://ai.mikuapi.org/auth/xlab/callback...`、`https://iframe.mikuapi.org/...`；依赖 `/oauth/consent`。 | 有。 | 接近完整 | 中：iframe URL 常量已迁移；需核对登录态 code/state 自动注入行为。 | 与 OAuth consent 一起测试。 |
| `/payment/qrcode` | 二维码支付页。 | `GET /payment/orders/:id`, verify/resolve public order。 | 有。 | 薄实现 | 高：支付流程关键路径，需端到端验证。 | 必须立刻测试订单创建 -> qrcode -> result。 |
| `/payment/result` | 支付结果恢复页，可 public 访问。 | `POST /payment/public/orders/verify`, `POST /payment/public/orders/resolve`, `GET /payment/orders/:id`。 | 有。 | 薄实现 | 高：旧 backend mode 白名单允许该页；React guard 是否等价需确认。 | 必须确认未被 RequireAuth 包裹且 public resume 可用。 |
| `/payment/stripe` | Stripe 支付页。 | `GET /payment/orders/:id` 或 Stripe client secret 相关。 | 有。 | 薄实现 | 中-高：页面存在，但需核对 Stripe SDK/redirect 语义是否与旧版一致。 | 支付闭环测试覆盖。 |
| `/payment/stripe-popup` | Stripe popup/嵌入支付页。 | direct fetch `/api/v1/payment/orders/:id` 等。 | 有。 | 薄实现 | 中：旧版特殊 popup 行为可能只部分迁移。 | 延后但需 smoke test。 |
| `/custom/:id` | 自定义菜单 iframe，按 public/admin settings 找配置并注入 user/token/theme/locale。 | `GET /settings/public`；admin 可读 admin settings；前端 `buildEmbeddedUrl`。 | 有：`ParityPlaceholder`。 | 占位 | 高：入口存在但核心 iframe 功能缺失；自定义菜单会退化。 | 必须实现最小 iframe：读 public settings、找 item、校验 URL、open in new tab。 |

### 3.3 Admin Core / Ops / Commercial

| xlabapi 路由 | 功能简述 | 用到的 API 端点 | frontend-v2 是否有 | 实现深度 | 风险 | 建议 |
|---|---|---|---|---|---|---|
| `/admin` -> `/admin/dashboard` | 管理入口兼容。 | 同 dashboard。 | 有重定向。 | 重定向 | 低。 | 保留。 |
| `/admin/dashboard` | 管理仪表盘。 | 旧版 `GET /admin/dashboard/stats`, `/admin/dashboard/realtime`, `/trend`, `/models`, `/groups`, `/user-breakdown`, `/snapshot-v2`, batch usage。 | 有，但 `frontend-v2/src/api/admin.ts` 当前调用 `GET /admin/dashboard`。 | 缺失/错配 | 高：基线明确要求 `/admin/dashboard/stats`，当前 client 仍是错路径。页面会直接打错 API。 | 必须立刻改 `adminAPI.getDashboard()` 到 `/admin/dashboard/stats`，并补最小测试。 |
| `/admin/ops` | 运维监控总览、错误、系统日志、并发、告警。 | 大量 `/admin/ops/*`：dashboard overview/snapshot、traffic、concurrency、errors、alert-rules/events、runtime/logging、system-logs。 | 有。 | 薄实现 | 中-高：React client 只覆盖子集，页面 284 行明显少于旧 Vue ops 组件群。 | P1 深化；P0 保 overview/errors/system logs 关键读操作。 |
| `/admin/users` | 用户管理。 | `GET/PUT/DELETE /admin/users`, user API keys/balance/allowed groups 等子接口。 | 有。 | 薄实现 | 中-高：React 基础列表/编辑，旧版 modal 子功能多。 | P1 补 user detail modals；P0 确认不会丢关键封禁/余额操作。 |
| `/admin/groups` | 分组管理、倍率、RPM override。 | `/admin/groups`, `/admin/groups/all`, status, rate multipliers, RPM override。 | 有。 | 薄实现 | 中：旧版 group 语义复杂，是倍率双轨风险点。 | 依赖 T003 语义审计，先不扩散。 |
| `/admin/channels` -> `/admin/channels/pricing` | 渠道入口兼容。 | 同 pricing。 | 有重定向。 | 重定向 | 低。 | 保留。 |
| `/admin/channels/pricing` | 渠道管理与模型价格。 | `GET/POST/PUT/DELETE /admin/channels`, `GET /admin/channels/model-pricing`。 | 有。 | 薄实现 | 中：React client 覆盖 CRUD+pricing，但旧版 channel form/account pricing rule 更深。 | P1 补编辑细节和 pricing rules。 |
| `/admin/channels/monitor` | 管理端渠道监控。 | `/admin/channel-monitors`, `/:id/run`, `/:id/history`，template APIs。 | 有。 | 薄实现 | 中：React client 无 monitor template API，旧版有模板管理。 | P1 补 template manager；P0 保 list/run/history。 |
| `/admin/accounts` | 上游账号管理、测试、批量编辑、重鉴权、CRS sync。 | `/admin/accounts`, `/:id/test`, `/:id/schedulable`, `/:id/refresh`, clear error/rate-limit，CRS/import 等。 | 有。 | 薄实现 | 中-高：基线 API 已修 schedulable/refresh；旧版 modal 功能非常多，React 可能只覆盖主 CRUD。 | P1 拆分账号功能闭合；P0 验证 schedulable/refresh。 |
| `/admin/subscriptions` | 用户订阅管理。 | `/admin/subscriptions`, assign, extend, reset quota, `DELETE /admin/subscriptions/:id`。 | 有。 | 接近完整 | 中：基础订阅管理可用；产品订阅另有缺口。 | 延后。 |
| `/admin/subscription-products` -> `/admin/subscriptions` | 旧产品订阅列表兼容。 | 同 `/admin/subscriptions` 或 product subscriptions。 | 有重定向。 | 重定向 | 中：旧路径名语义是产品，但现在跳用户订阅，可能隐藏产品管理缺口。 | 保留兼容，但必须补 `/admin/subscription-product-config`。 |
| `/admin/subscription-product-config` | 订阅产品配置、绑定 group、给用户分配产品订阅。 | `/admin/subscription-products`, `/admin/product-subscriptions`, bindings, assign/update/revoke。 | 有：`ParityPlaceholder`。 | 占位 | 高：旧 Vue 是完整产品配置；当前无法管理产品订阅。 | 必须立刻从占位升级为最小 CRUD/绑定/assign 页面，或至少提供 admin API client。 |
| `/admin/announcements` | 公告管理和 read status。 | `/admin/announcements`, `/:id/read-status`。 | 有。 | 接近完整 | 低-中。 | 延后。 |
| `/admin/proxies` | 代理管理、批量导入、质量检测、账号关联。 | `/admin/proxies`, `/all`, `/batch`, `/batch-delete`, `/:id/quality-check`, `/:id/accounts`, data import/export。 | 有。 | 薄实现 | 中：React 覆盖 list/test/detail，缺批量和 data import。 | P1 补批量导入导出。 |
| `/admin/redeem` | 兑换码管理。 | `/admin/redeem-codes`, batch generate/delete/expire。 | 有。 | 接近完整 | 低-中。 | 延后。 |
| `/admin/invites` -> `/admin/users` | 旧邀请管理兼容。 | 同 users。 | 有重定向。 | 重定向 | 低。 | 保留。 |
| `/admin/promo-codes` | 促销码管理。 | `/admin/promo-codes`, usage。 | 有。 | 接近完整 | 低-中。 | 延后。 |
| `/admin/settings` | 系统设置，OAuth、支付、SMTP、自定义菜单、功能开关。 | `/admin/settings`, `/admin/settings/test-smtp`, `/admin/settings/send-test-email`，以及 OAuth/payment/settings fields。 | 有。 | 薄实现 | 中-高：React settings 652 行但旧版 6000+ 行，可能缺 WeChat/OIDC/payment/custom menu 深层配置。 | P1 分块迁移；P0 确认 SMTP 修正和 custom menu 配置入口。 |
| `/admin/usage` | 管理端用量查询、stats、search users/keys、cleanup/export。 | `/admin/usage`, `/admin/usage/stats`, `/admin/usage/search-users`, `/admin/usage/search-api-keys`, cleanup tasks。 | 有。 | 薄实现 | 中：React 列表/filter 有，stats/search/cleanup 可能缺 UI。 | P1 补 stats cards 和 cleanup。 |
| `/admin/affiliates` -> `/admin/affiliates/invites` | 管理返利入口兼容。 | 同 invites。 | 有重定向。 | 重定向 | 低。 | 保留。 |
| `/admin/affiliates/invites` | 邀请记录。 | `GET /admin/affiliates/invites`。 | 有。 | 薄实现 | 低-中：列表可用，用户返利配置 API 未覆盖。 | P1 补 affiliate user config。 |
| `/admin/affiliates/rebates` | 返利记录。 | `GET /admin/affiliates/rebates`。 | 有。 | 薄实现 | 低-中。 | 延后。 |
| `/admin/affiliates/transfers` | 返利转账记录。 | `GET /admin/affiliates/transfers`。 | 有。 | 薄实现 | 低-中。 | 延后。 |
| `/admin/orders/dashboard` | 支付管理仪表盘。 | `GET /admin/payment/dashboard`。 | 有。 | 薄实现 | 中：列表/汇总有，需对齐旧 stats 字段。 | 支付 P0 测试覆盖。 |
| `/admin/orders` | 支付订单管理。 | `/admin/payment/orders`, `/:id/cancel`, `/:id/retry`, `/:id/refund`。 | 有。 | 薄实现 | 中：基础操作有，筛选/退款细节需核对。 | 支付 P0 测试覆盖。 |
| `/admin/orders/plans` | 支付套餐管理。 | `/admin/payment/plans`, `/admin/payment/providers`, `/admin/payment/channels`。 | 有。 | 薄实现 | 中：React client 有 plans/providers 但缺 admin payment channels CRUD。 | P1 补 payment channel/provider 配置。 |
| `/admin/backup` | 备份、S3 配置、计划、恢复。 | `/admin/backups`, `/admin/backups/s3-config`, `/test`, `/schedule`, `/:id/download-url`, restore/delete。 | 有。 | 接近完整 | 低-中：基线路径已修。 | 延后。 |
| 404 | 未匹配路由。 | 无。 | 有：`*`。 | 完整 | 低。 | 无需动。 |

---

## 4. 必须立刻动

1. `/admin/dashboard` API 错路径：`frontend-v2/src/api/admin.ts` 当前仍调用 `GET /admin/dashboard`，与基线要求 `/admin/dashboard/stats` 不一致。这是明确 P0 契约 bug。
2. `/key-usage` 不能继续占位：旧功能是 public 输入 API key 调 `GET /v1/usage`，当前占位列 `/usage/dashboard/*` 属于登录态 dashboard API，语义错配。
3. `/setup` 不能继续占位：旧安装向导承担数据库/Redis 测试和 install，当前只恢复入口但不可安装。
4. WeChat/OIDC callbacks 不能复用 generic code/state copy 页：`/auth/wechat/callback`、`/auth/wechat/payment/callback`、`/auth/oidc/callback` 需要自动处理 token/pending/支付恢复，不是展示 code/state。
5. `/custom/:id` 不能继续占位：旧功能是自定义菜单 iframe，并注入 user/token/theme/locale；当前占位会让已配置菜单不可用。
6. `/admin/subscription-product-config` 不能继续占位：旧功能是订阅产品配置、group binding、assign product subscription；当前 admin 无法管理产品订阅。
7. 支付主链路需立刻做契约验证：`/purchase` -> `/payment/qrcode` / Stripe -> `/payment/result` -> `/orders`，路由和 thin clients 已存在，但实现明显薄，且 WeChat payment callback 仍是高风险。
8. 路由守卫 parity 需复核：旧 Vue 的 `requiresPayment`、simple mode、backend mode 白名单，在 React router 文件中未直接体现。若 `guards.tsx` 未实现，可能出现付费关闭仍可访问购买页、backend mode 错误跳转等行为差异。

## 5. 占位符清单

| 路由 | 当前占位方式 | 旧功能真实要求 | 建议优先级 |
|---|---|---|---|
| `/setup` | `ParityPlaceholder standalone` | 完整安装 wizard：status、DB test、Redis test、install。 | P0 |
| `/key-usage` | `ParityPlaceholder standalone`，且 API 标注为 `/usage/dashboard/*` | public API key 查询 `/v1/usage`，展示今日/总量/model stats。 | P0 |
| `/custom/:id` | `ParityPlaceholder` | 读取 public/admin custom menu item，构造 embedded URL，iframe 展示并支持新窗口打开。 | P0 |
| `/admin/subscription-product-config` | `ParityPlaceholder` | 产品订阅 CRUD、group binding、assign/list product subscriptions。 | P0 |
| `/auth/wechat/callback` | generic `OAuthCallbackPage` | WeChat 登录/注册/pending 自动处理。 | P0 |
| `/auth/wechat/payment/callback` | generic `OAuthCallbackPage` | WeChat 支付 OAuth 回调和支付恢复。 | P0 |
| `/auth/oidc/callback` | generic `OAuthCallbackPage` | OIDC 登录/注册/pending 自动处理。 | P0 |
| `/auth/callback` | generic `OAuthCallbackPage` | 取决于后端模式；若只用于 admin copy code 可保留，否则需自动 token exchange。 | P1/P0 待确认 |

## 6. 可延后

1. `/docs` 内容深度：路由已存在，非旧 Vue 核心功能。
2. `/available-channels` 的完整倍率解释、group rates fallback 和高级筛选：入口和主要 API 已有，可作为 P1。
3. `/monitor` 用户端历史、筛选和细节图表：入口和 list/detail 已有，可作为 P1。
4. `/admin/ops` 深层告警规则、运行时设置、WebSocket 实时 QPS、复杂图表：P0 先保证 overview/errors/system logs，完整迁移可拆 P1。
5. `/admin/proxies` 批量导入导出、质量检测细节、账号关联视图：入口已可用，可 P1。
6. `/admin/usage` stats cards、cleanup tasks、export：基础列表/filter 已有，可 P1。
7. `/admin/settings` 深层配置项视觉/交互完整迁移：先保证 SMTP/OAuth/payment/custom menu 关键字段，其他可 P1/P2。
8. `/admin/orders/*` 的 provider/channel 完整配置和退款细节：支付主链路先测通，管理细节可 P1。
9. `/profile` 的完整安全中心可拆分：但 OAuth binding、TOTP、通知邮箱建议排在 P0 后第一批 P1，不宜长期延后。

---

## 7. 审计结论

`origin/test/xlabapi:frontend-v2` 已经比 2026-05-11 基线前进一大步：多数 P0 路由和 thin API client 已存在，导航也覆盖了用户、admin、支付、affiliate、channel、backup 等入口。当前主要风险不是“是否有 route”，而是以下三点：

1. 少数入口仍是纯占位，且其中 `/setup`、`/key-usage`、`/custom/:id`、`/admin/subscription-product-config` 都是旧前端真实业务功能。
2. OAuth/支付回调存在“有路由但处理器错配”的风险，尤其 WeChat/OIDC 被 generic copy page 接住，不能视为 parity。
3. 至少一个明确 API path bug 仍存在：admin dashboard client 应为 `/admin/dashboard/stats`，当前是 `/admin/dashboard`。

建议下一步把“入口存在”从完成标准里剔除，改用三档验收：route present、API contract correct、old behavior minimally executable。按这个标准，T001 后续实施应优先处理上面 P0 清单。
