# frontend-v2 系统性整改总规划 v1

- 规划目标：把 2026-05-12 的功能/风格/语义/测试审计合成一份用户拍板入口，后续实施任务只从本文拆分，不再直接从单份审计散点开工。
- 输入来源：`2026-05-12-frontend-v2-parity-matrix.md`、`2026-05-12-frontend-v2-style-gap.md`、`2026-05-12-placeholder-inventory.md`、`2026-05-12-semantic-dual-tracks.md`、`2026-05-12-test-coverage-inventory.md`，并兼容 `origin/test/xlabapi:docs/superpowers/plans/2026-05-11-frontend-v2-xlabapi-parity.md`。
- 规划性质：只合成，不新增实现决策；凡审计未覆盖或输入不完整的点，进入“开放问题”。
- 当前约束：不修改既有 `docs/superpowers/plans/*`；本文等用户审过后才进入实施。

---

## 1. 目标与红线

### 1.1 三优先级

| 优先级 | 目标 | 验收口径 |
|---|---|---|
| P0 功能不丢 | xlabapi 线上旧 Vue 前端已有的功能入口、路由、API 契约在 frontend-v2 中必须存在且可执行；不允许“有 route 但行为错配”。 | 旧路由不 404；关键页面能完成最小业务闭环；frontend-v2 调用的 method/path/status/envelope/字段名与后端契约一致；外部 `/api/v1` 和 `/v1/*` HTTP 契约不变。 |
| P1 风格统一 | Landing 与 Console/Dashboard 使用同一套 frontend-v2 token 和组件原语，消除“原型内联样式 vs 控制台标准组件”的断层。 | Landing 不再复制 `Eyebrow/PillBtn/Wordmark` 等私有实现；颜色、字号、容器、圆角、边线尽量走现有 token/class；Console 行为和信息密度不被破坏。 |
| P2 语义去双轨 | 订阅、邀请码、倍率三组高混淆概念建立稳定术语边界，能改名的先改名，必须重构的有测试护栏后再动。 | 文档/代码/DTO/UI 文案不再裸用歧义词；产品订阅、注册准入码、affiliate code、倍率口径互不串线；关键运行时行为有隔离测试保护。 |

### 1.2 红线清单

| 红线 | 说明 |
|---|---|
| 对外 HTTP 契约不许动 | `/api/v1` xlabapi API、OpenAI/Anthropic/Gemini-compatible `/v1/*` 与 root alias 的 method/path/status/header/cookie/body shape 不允许因 frontend-v2 整改漂移。 |
| 不写高耦合代码 | 页面迁移、API client、测试 fixture、语义重命名都必须保持模块边界；不能为了快速过页面把后端业务语义塞进前端特殊判断。 |
| 设计 -> 文档 -> 测试 -> 开发 | 本文是设计/规划入口；实施前先补契约文档和最小测试门禁，再改 API path、路由、页面、样式、语义。 |
| 不动生产部署 | 继承既有 parity design 约束：不编辑 `/root/sub2api-deploy`，不重启 production `sub2api`，不使用 production 端口 `8081` 验证。 |
| test image 隔离 | 部署验证只面向 `/root/test_xlab` 和 `sub2api-test-xlab`，继续使用 test image / override 约束，除非用户另行拍板。 |
| 旧 Vue 不在本轮退役 | 老 Vue frontend 只作为 parity 对照，不在本规划内决定退役时间。 |

### 1.3 与 2026-05-11 parity plan 的关系

| 既有 task | 本规划处理方式 |
|---|---|
| Task 1 P0 API Contract Corrections | 保留为 S1 的核心输入；新增 T001 发现的 `/admin/dashboard` path 错仍归此 task；需要先被契约测试门禁保护。 |
| Task 2 Route And Navigation Entry Parity | 已有多数路由和导航入口，S2 不再重复“加 route”，而是处理“有 route 但占位/行为错配”的剩余项。 |
| Task 3 Minimal Missing API Clients | 多数 thin clients 已出现，S2/S4 只补仍缺的 product subscription、OAuth pending/profile binding 等客户端面。 |
| Task 4 Test Deployment | 仍作为每个实施阶段的 test image 验证出口；本文增加阶段化进入/退出条件和回滚方式。 |

---

## 2. 三轴路线图

### 2.1 轴 A：功能对齐

范围来自 T001 功能/路由/API 矩阵和 placeholder inventory。当前结论是：`origin/test/xlabapi:frontend-v2` 已经不再是“大量入口 404”，多数 route 和 thin client 已存在；剩余 P0 风险集中在“有入口但不可执行 / 行为错配 / API path 错”。

| 子项 | 范围 | 依赖 | 工作量估计 | 风险 |
|---|---|---|---|---|
| API path P0 修正 | 继承 2026-05-11 plan Task 1：usage/admin usage/admin dashboard/backup/SMTP/accounts/subscriptions/groups；T001 明确当前 `frontend-v2/src/api/admin.ts` 仍调 `/admin/dashboard`，应为 `/admin/dashboard/stats`。 | T004/T008 最小 HTTP contract；T006 API-CONTRACT.md。 | 小；以 API client 修正和 build 为主。 | 高：路径错会直接导致 admin dashboard 失败。 |
| 占位符消除 | `/setup`、`/key-usage`、`/custom/:id`、`/admin/subscription-product-config`。T005 占位审计确认全部集中在 `frontend-v2/src/router/index.tsx`，其中 `/key-usage` 当前 endpoints 标注还错指登录态 dashboard API。 | 旧 Vue 页面行为；后端 setup/key usage/product subscription/custom menu 契约；T004 contract batch。 | 中；四个页面拆开实施，不应一次大改。 | 高：这些入口不是装饰，分别影响首次部署、公开 key 查询、自定义菜单、订阅产品管理。 |
| OAuth callback 行为修复 | `/auth/wechat/callback`、`/auth/wechat/payment/callback`、`/auth/oidc/callback` 当前复用 generic `OAuthCallbackPage`，旧行为需要自动 token/pending/支付恢复处理。 | Auth/OAuth pending contract；旧 Vue callback 页面；后端 auth service tests。 | 中-高；需要对 WeChat/OIDC/payment 三条流分别建测试。 | 高：有 route 但行为错配，用户会停在 code/state copy 页，登录或支付恢复断裂。 |
| 支付主链路验证 | `/purchase` -> `/payment/qrcode`/Stripe -> `/payment/result` -> `/orders`；admin payment dashboard/orders/plans。 | Payment contract batch；test image；支付 provider stub 或可控测试配置。 | 中；先测通 happy path，再补边界。 | 高：provider-specific payload、resume token、public verify/resolve 容易漂移。 |
| 路由守卫 parity | 旧 Vue 的 `requiresPayment`、simple mode、backend mode 白名单；React router 文件未直接体现。 | `frontend-v2/src/router/guards.tsx` 和 auth/store 审核；contract/smoke。 | 小-中；先验证，再决定是否改。 | 中-高：付费关闭仍可访问购买页、backend mode 错误跳转会造成线上行为差异。 |
| P1 功能闭合 | Profile 安全中心、available channels 倍率展示、usage stats/filter/export、admin ops/proxies/settings 深层功能、admin users/accounts modals。 | P0 完成；语义术语表；前端测试框架。 | 大；应按页面域拆任务。 | 中：长期不补会形成“入口有但功能缩水”。 |

### 2.2 轴 B：风格统一

采用 T002 §6 推荐方案 A：Landing 重构对齐 Console。规划只确定路线，不在本文给具体代码实现。

| 子项 | 范围 | 依赖 | 工作量估计 | 风险 |
|---|---|---|---|---|
| Landing 原子层去重复 | 删除 Landing 私有 `OrangeMark`、`Eyebrow`、`PillBtn`、内联菱形 logo，统一使用 `bus/Eyebrow`、`bus/PillBtn`、`bus/Wordmark`。 | P0 功能稳定后；frontend-v2 build；视觉回归。 | T002 估计 1 名前端 1.5-2 天总量的一部分。 | 低：纯样式/组件消费方式调整，但需避免 Link/button 嵌套问题。 |
| Landing token 化 | 颜色、边线、字号、圆角、容器从内联 style 改为现有 token/class；字号 84/64/52 等非 token 值需要用户决定降到 token 网格还是扩 token。 | 用户对视觉取舍拍板；不改 Console 行为。 | 中。 | 中：视觉差异会明显，必须通过截图/浏览器审查。 |
| Landing 容器/Section 响应式 | 用 `.container-bus`，抽薄 Section 枚举，修移动端塞车；校准 `SectionFrame` halfWidth。 | 现有 bus 原语；移动端 QA。 | 中。 | 中：布局重排风险高于颜色替换。 |
| Console 保持稳定 | 不把 Landing 的展会感装饰搬进 Console，不改 dashboard/admin 信息密度。 | 红线复核。 | 无独立实现量。 | 低：避免方案 B 的大范围回归。 |
| S0 小样式 bug | `card-hover` 在 Dashboard 使用但全局未定义；S0 可补定义。 | 用户确认 S0 修。 | 小。 | 低：但属代码改动，需用户批准后执行。 |

T002 明确不推荐方案 B（Console 升级对齐 Landing）和方案 C（抽第三层 theme），原因是回归面大或抽象过早。除非用户改方向，否则本文后续阶段只按方案 A 排期。

### 2.3 轴 C：语义去双轨

范围来自 T003：订阅、邀请码、倍率三组。本文只按审计归纳“改名能解决”与“需真重构”，不新增新的业务规则。

| 语义组 | 当前边界 | 改名/文档能解决 | 需真重构或强测试护栏 | 依赖 | 风险 |
|---|---|---|---|---|---|
| 订阅：旧分组订阅 vs 新产品订阅 | `UserSubscription` 是旧分组订阅；`SubscriptionProduct/UserProductSubscription/SubscriptionProductBinding` 是新产品订阅；产品解析失败不得隐式回退旧订阅。 | 术语统一为 `legacy group subscription`、`product subscription`、`product binding group`、`product settlement`；注释说明 `UserSubscription` 兼容投影不是 legacy 落库。 | 保护 `api_key_auth.go` 的 `productSubscriptionChecked` 语义；产品订阅解析失败不回退旧订阅；产品兑换码必须有 `product_id`；旧分组订阅仍可按 legacy 流程工作。 | T004/T008 subscription contract；产品订阅既有设计文档。 | 高：误重构会重新打开额度绕过或错误 fallback。 |
| 邀请码：registration invitation / invite_code / affiliate code | `redeem_codes.type=invitation` 是注册准入；`users.invite_code` 是传统邀请关系；`user_affiliates.aff_code` 是 affiliate 返利码。 | UI/文档避免裸写“邀请码”：注册准入码、邀请关系码、返利推广码分开；admin affiliate 的“专属邀请码”应明确是 affiliate code。 | 传统 `InviteService` 与 `AffiliateService` 是否收敛需要用户拍板；若不收敛，必须补两套奖励是否会同时触发的冲突测试。 | 注册/affiliate/redeem contract；用户拍板奖励模型。 | 中-高：同名 code 可能导致错误入口复用或返利重复。 |
| 倍率：固定/动态/预期 | 实际三态是固定计费倍率、动态预算/账号选择倍率、展示用动态预期倍率；`fixed_multiplier`/`expected_multiplier` 字面在后端未命中。 | 建立术语表：`user_billing_multiplier`、`account_cost_multiplier`、`account_group_billing_multiplier`、`product_debit_multiplier`、`dynamic_budget_multiplier`、`dynamic_display_matched_multiplier`。 | 保护 `resolveBillingMultiplierForUsage`：fixed 用户专属倍率、dynamic account-group 倍率、账号成本倍率、产品 debit 倍率互不串线；不要新增 DB 字段 `expected_multiplier`。 | group/usage/channel contract；T003 术语表。 | 高：倍率串线会直接造成扣费错误。 |

轴 C 的实施顺序应是：先文档和测试术语护栏，再做命名/DTO/UI 文案，最后才碰运行时重构。订阅和倍率属于“需要真重构护栏”的高风险项；邀请码首先可通过术语拆分和 UI 文案降低歧义，但传统 invite 与 affiliate 是否收敛不是本文决策范围。

---

## 3. 测试驱动入口

### 3.1 重构前必须先落地的最小契约测试集

当前可直接引用的测试审计输入来自 T004 `docs/superpowers/audits/2026-05-12-test-coverage-inventory.md`。T008 作为正式测试设计任务目前已在 control plane 里变为 active，但仓库中尚未看到单独落盘的 `CONTRACT-TEST-PLAN.md`，因此本规划先把 T004 §4.1 作为临时测试入口，等 T008 产物落地后再替换引用。

| 批次 | 作用 | 临时引用来源 | 验收门禁 |
|---|---|---|---|
| 1 | auth P0：login/register/refresh/logout/2FA/OAuth pending | T004 §4.1 批次 1 | 任何一条 auth contract 变更都必须先过；否则不进入 S1。 |
| 2 | keys + groups + usage | T004 §4.1 批次 2-3 | `/usage`、`/usage/stats`、`/groups/available`、`/groups/rates` 先稳定。 |
| 3 | subscriptions + redeem + affiliate/invite | T004 §4.1 批次 4-5 | 用户商业闭环可被 smoke；不通过不进入 S2。 |
| 4 | payment flow | T004 §4.1 批次 6 | 订单创建/verify/public resolve 可用，才能动 `/purchase` 和支付页。 |
| 5 | admin P0 + gateway snapshots | T004 §4.1 批次 7-8 | admin/dashboard/settings/users/groups/accounts/redeem/payment 基线稳定后，再碰 ops 深层页。 |

### 3.2 阶段门禁原则

| 阶段 | 进入前必须满足 | 允许开的工作 | 阶段通过条件 |
|---|---|---|---|
| S0 | 不要求新业务接口完成，但 `card-hover`、`docs` 规则、死代码是否清理、测试框架落点必须确认；至少有可执行的最小测试计划入口。 | 只做基建/治理/测试框架/文档约定，不碰业务 flow。 | repo 基础治理完成，后续阶段能稳定跑 `build` 和最小测试命令。 |
| S1 | T004 批次 1-2 门禁通过，且 `admin dashboard` path 错修正已被测试保护。 | P0 API path 修正、最小 contract 测试第一批、后台最危险的 endpoint 纠偏。 | auth/keys/usage/groups/contracts 稳定，frontend-v2 不再因 path 错/字段漂移断主登录态和 dashboard。 |
| S2 | S1 通过；T004 批次 3-4 至少通过 smoke；占位路由定位已确认。 | `/setup`、`/key-usage`、`/custom/:id`、`/admin/subscription-product-config`、WeChat/OIDC callback 行为修复。 | 四个占位/错配入口至少能执行最小旧功能，支付/登录回调不再停在 generic copy 页。 |
| S3 | S2 通过；P0 route/API 已稳；用户确认走方案 A。 | Landing 风格统一到 Console。 | Landing 不再复制私有原语，UI token/class 统一，视觉回归通过。 |
| S4 | S3 通过；语义术语表确认；T003 的区分点已写入文档/测试。 | 订阅/邀请码/倍率去双轨。 | 文档、DTO、UI 文案、运行时逻辑在三组术语上不再串线。 |

### 3.3 现阶段的测试缺口结论

- frontend-v2 目前没有 Vitest/RTL/MSW 测试体系，任何页面或 API client 重构都缺前端自动回归。
- 后端已有大量 service/repository 测试，但 P0 HTTP contract 仍需补 golden/snapshot，否则 frontend-v2 只会继续依赖人工 spot-check。
- 本规划把 `T004` 的 30-40 个 golden contract 方向视作第一批必须先补的“测试地基”；`T008` 一旦落盘，应替换这里的临时引用并作为唯一正式测试设计入口。

---

## 4. 分阶段实施

> 分派说明：以下 actor 分派是后续拆任务建议，不代表已经开始实施。每阶段都必须在用户审过本文后，由领班重新拆具体任务卡。

### S0 基建

| 项 | 内容 |
|---|---|
| 进入条件 | 用户批准本文 v1 的整体方向；确认本轮基线是 `test/xlabapi` 还是 `main`；确认是否允许修改 `.gitignore:129` 的 `docs/*`。 |
| 工作内容 | 文档治理：处理 `docs/*` ignore 导致审计/plan 文件未被 git 跟踪的问题（文档整理者）。测试入口：按 T004/T008 确认 contract test plan 路径和第一批命令（测试者）。前端小修：确认是否补 `.card-hover` 定义（前端开发实现者）。死代码：确认是否删除 `frontend-v2/src/pages/admin/Placeholder.tsx`（前端开发实现者）。契约文档：T006 产出 `docs/superpowers/contracts/API-CONTRACT.md`（后端开发实现者）。 |
| 退出条件 | 新规划、审计、contract 文档能被 git 正常追踪；已有 `build/typecheck` 和最小 contract 测试命令可在后续阶段复用；`card-hover` 和死代码处理方式有用户明确决定。 |
| 回滚方式 | 文档治理改动用单独 commit；若 ignore 调整误伤，`git revert <docs-governance-commit>`；若前端小修引发视觉问题，`git revert <s0-frontend-cleanup-commit>`。 |

### S1 P0 API path 修 + 契约测试第一批

| 项 | 内容 |
|---|---|
| 进入条件 | S0 通过；T004/T008 第一批 auth/keys/groups/usage contract 测试可执行；T006 API-CONTRACT.md 至少覆盖 S1 endpoint。 |
| 工作内容 | 修正 2026-05-11 parity plan Task 1 中的 API path：user usage、admin usage、admin dashboard `/admin/dashboard/stats`、backup `/admin/backups`、SMTP、accounts schedulable/refresh、admin subscriptions revoke、groups available/rates（前端开发实现者）。补/固化对应后端 contract golden（测试者 + 后端开发实现者）。执行 frontend-v2 build/typecheck（前端开发实现者）。 |
| 退出条件 | `/admin/dashboard` 页面实际调用 `/admin/dashboard/stats`；user/admin usage、backup、SMTP、accounts、subscriptions、groups 的 method/path 与 API-CONTRACT.md 一致；第一批 contract 测试和 frontend-v2 build 通过。 |
| 回滚方式 | API client path 修正集中在一个或少量 commit；如 test image 验证失败，`git revert <s1-api-path-commit>` 并切回上一 test image；不改生产。 |

### S2 路由/占位符补齐

| 项 | 内容 |
|---|---|
| 进入条件 | S1 通过；payment/subscriptions/redeem/affiliate/auth OAuth 的 contract 或 smoke 至少覆盖相关最小流；四个 ParityPlaceholder 的真实功能边界已被用户接受。 |
| 工作内容 | 替换 `/setup` 占位为最小安装向导：status、DB test、Redis test、install（前端开发实现者 + 后端开发实现者核 API）。替换 `/key-usage` 占位为 public API key usage 查询，调用旧语义 `/v1/usage`（前端开发实现者 + 测试者）。替换 `/custom/:id` 占位为 custom menu iframe，读取 public/admin settings 并注入 user/token/theme/locale（前端开发实现者）。替换 `/admin/subscription-product-config` 占位为最小产品订阅配置页或先补 API client + 可执行管理页（前端开发实现者 + 后端开发实现者）。修复 `/auth/wechat/callback`、`/auth/wechat/payment/callback`、`/auth/oidc/callback` 行为错配，不再停留 generic code/state copy 页（前端开发实现者 + 后端开发实现者 + 测试者）。 |
| 退出条件 | `/setup`、`/key-usage`、`/custom/:id`、`/admin/subscription-product-config` 不再渲染 `ParityPlaceholder`；WeChat/OIDC/payment callback 至少能完成旧 Vue 对应的最小自动处理；代表路由在 test image 返回 200 且关键交互能执行；payment happy path smoke 通过。 |
| 回滚方式 | 每个占位替换独立 commit；单一路由失败时 revert 对应 commit，不回滚全部 S2；如 callback 修复影响登录，优先 `git revert <oauth-callback-commit>` 并切回上一 test image。 |

### S3 风格统一

| 项 | 内容 |
|---|---|
| 进入条件 | S2 通过；用户确认 T002 方案 A；P0 功能不再依赖 Landing 内联原型结构；已有视觉回归检查方式。 |
| 工作内容 | Landing 原子层统一：替换私有 `Eyebrow/PillBtn/OrangeMark/NavD logo` 为 bus 原语（前端开发实现者）。Landing token 化：颜色/边线/字号/圆角/容器改用现有 token/class（前端开发实现者）。Landing Section/响应式：使用 `.container-bus`、薄 Section 枚举、修移动端布局；不改 Console/Dashboard 行为（前端开发实现者）。审查视觉差异并记录截图或浏览器检查结果（审查者/测试者）。 |
| 退出条件 | Landing 与 Console 共用 token 消费方式；`Landing.tsx` 私有重复原语消除；desktop/mobile 打开正常；Console/Dashboard 无行为和信息密度回归；frontend-v2 build 通过。 |
| 回滚方式 | 风格统一独立 commit；如视觉回归不可接受，`git revert <s3-landing-style-commit>`；不影响 S1/S2 功能修复。 |

### S4 语义去双轨

| 项 | 内容 |
|---|---|
| 进入条件 | S3 通过；用户拍板邀请码/affiliate 是否收敛、倍率术语是否采用 T003 建议；订阅和倍率隔离测试已准备。 |
| 工作内容 | 订阅术语：文档/注释/DTO/UI 统一 legacy group subscription vs product subscription；保护产品解析不回退 legacy（后端开发实现者 + 测试者）。邀请码术语：注册准入码、邀请关系码、affiliate 推广码分层；若用户决定收敛传统 invite 与 affiliate，再另拆重构任务（后端开发实现者 + 前端开发实现者）。倍率术语：建立 `user_billing_multiplier`、`account_cost_multiplier`、`account_group_billing_multiplier`、`product_debit_multiplier`、`dynamic_budget_multiplier`、`dynamic_display_matched_multiplier` 映射；保护 `resolveBillingMultiplierForUsage`（后端开发实现者 + 测试者）。 |
| 退出条件 | 三组术语在文档/API/UI 中有清晰边界；关键隔离测试通过：产品订阅失败不回退、产品 redeem 必须 product_id、legacy 订阅仍可用、fixed/dynamic/account/product multiplier 不串线；无对外契约漂移。 |
| 回滚方式 | 先文档/命名，后行为；命名 commit 可独立 revert；任何运行时重构必须单独 commit，若 contract 或 billing 测试失败，`git revert <s4-runtime-commit>` 并保留文档术语改动是否回滚由用户决定。 |

---

## 5. 开放问题

以下问题不在审计里形成确定答案，需要用户或领班拍板后才能进入实施。

| 问题 | 为什么需要拍板 | 可选方向 |
|---|---|---|
| `.gitignore:129` 的 `docs/*` 规则是否改 | 当前 `docs/superpowers/audits/` 和本文计划文件在工作树中显示为 untracked，`docs/*` 会影响后续文档产物进入 repo。 | 改 ignore 规则并显式纳入 `docs/superpowers/**`；或继续作为本地协作产物，不入 repo。 |
| `API-CONTRACT.md` 进入 repo 的路径 | T006 目标是 `docs/superpowers/contracts/API-CONTRACT.md`，但 docs ignore 规则和文档治理尚未完成。 | 使用 T006 指定路径；或由文档治理任务指定 contracts 目录。 |
| `frontend-v2/src/pages/admin/Placeholder.tsx` 是否 S0 就删 | Placeholder inventory 判定其 0 引用、与 `ParityPlaceholder` 重复，但删除仍是代码改动。 | S0 删除；或等 S2 占位替换一起清理。 |
| `.card-hover` 类是否 S0 就补定义 | T002 发现 Dashboard 使用但全局 CSS 未定义，属于小 bug，但仍需改 CSS。 | S0 小修；或留到 S3 风格统一处理。 |
| test 分支 vs main 分支分工 | 当前工作树为 `main`，审计目标和输入多来自 `origin/test/xlabapi`；实施基线需明确。 | 以 `test/xlabapi` 为开发基线；或先把必要内容同步回 main 再实施。 |
| 部署目标是否继承既有约束 | 2026-05-11 design 约束只动 `/root/test_xlab` 和 `sub2api-test-xlab`，不动 production。 | 默认继承；若用户要求其他环境，需重写部署门禁。 |
| T008 契约测试设计何时替换本文临时引用 | T008 控制面任务已存在，但当前未看到落盘产物；本文只能临时引用 T004 §4.1。 | 等 T008 `CONTRACT-TEST-PLAN.md` 完成后更新本文；或 T009 v1 先提交给用户，后续 v2 合入。 |
| 邀请增长模型是否收敛 | T003 明确 traditional invite 与 affiliate 都有奖励语义，是否收敛不是审计结论。 | 保持双轨并强化术语/测试；或另立产品决策合并奖励模型。 |
| Landing 非 token 字号如何处理 | T002 指出 84/64/52 等写死值不在 token 网格内。 | 降到现有 token；或扩 token，但这会新增设计决策。 |

---

## 6. 不在本规划内的事

| 排除项 | 原因 |
|---|---|
| 老 Vue `frontend/` 的退役时间表 | 本轮只做 frontend-v2 对齐和整改；旧前端仍是 parity baseline，不决定下线。 |
| 后端架构级重构 | 除订阅/邀请码/倍率三组语义边界外，不扩大到整体后端架构重写。 |
| 新功能开发 | 本轮目标是功能不丢、风格统一、语义去双轨；不新增线上没有的新能力。 |
| 生产部署改造 | 不改 `/root/sub2api-deploy`，不重启 production，不调整 production 端口。 |
| 全量 Playwright e2e 或覆盖率 KPI | T004 明确当前重点是 P0 HTTP contract，不追覆盖率数字和大量 UI snapshot。 |

---

## 7. v1 结论

1. 先保 P0：功能对齐不是“路由存在”，而是 route present、API contract correct、old behavior minimally executable 三项同时成立。
2. 再做 P1：风格统一按 T002 方案 A，只让 Landing 收敛到现有 Console token/组件消费方式，不反向改 Console。
3. 最后做 P2：语义去双轨先立术语和测试，再做命名或运行时重构，避免订阅/邀请码/倍率串线。
4. 所有实施前先补测试地基：T004/T008 的 contract suite 和 T006 的 API-CONTRACT.md 是 S1/S2 的硬前置。
5. 本文不是实施授权；用户审过开放问题和阶段边界后，才拆后续 actor 任务。
