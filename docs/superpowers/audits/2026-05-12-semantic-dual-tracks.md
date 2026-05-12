# Semantic Dual Tracks Audit — Backend

- 审计目标：清点 `origin/test/xlabapi` 后端里三组容易混淆的语义双轨：订阅、邀请码、倍率。只盘点，不重构。
- 审计范围：`origin/test/xlabapi:backend/`，并参考既有产品订阅计划/设计文档。
- 交付方式：按任务要求三节独立取证、独立落盘；本文件只记录可由 `git show` / `git grep` 复核的代码位置。

---

## 1. 订阅双轨：旧分组订阅 vs 新产品订阅

### 概念边界

| 轨道 | 核心模型 | 额度来源 | 绑定对象 | 运行时语义 | 主要入口 |
|---|---|---|---|---|---|
| 旧分组订阅 | `UserSubscription` | `groups.daily/weekly/monthly_limit_usd` | 用户 + 单个 `group_id` | 一个订阅分组一份独立额度，API Key 命中该分组后走 legacy subscription | `SubscriptionService` / `UserSubscriptionRepository` / admin group subscriptions |
| 新产品订阅 | `SubscriptionProduct` + `UserProductSubscription` + `SubscriptionProductBinding` | `subscription_products.daily/weekly/monthly_limit_usd` | 用户 + 产品实例；产品绑定多个真实订阅分组 | 多个真实订阅分组共享同一个 `user_product_subscriptions` 额度池，请求按绑定倍率折算产品扣费 | `SubscriptionProductService` / `subscriptionProductRepository` / product admin APIs / redeem product assignment |

既有中文设计文档已把边界写清楚：旧版是“一个兑换码激活一个分组限额”，新版是“一个产品激活多个真实订阅分组，共享同一份产品额度”。对应文档位置：`origin/test/xlabapi:docs/PRODUCT_SUBSCRIPTIONS_CN.md:7`、`origin/test/xlabapi:docs/PRODUCT_SUBSCRIPTIONS_CN.md:21`、`origin/test/xlabapi:docs/PRODUCT_SUBSCRIPTIONS_CN.md:29`。

更关键的闭环规则也已写入设计：产品订阅解析失败时不能隐式回退到旧版 `user_subscriptions`，API Key 仍绑定真实 `groups.id`，必要时用 `api_keys.subscription_product_family` 消除多产品族歧义。对应文档位置：`origin/test/xlabapi:docs/PRODUCT_SUBSCRIPTIONS_CN.md:27`、`origin/test/xlabapi:docs/PRODUCT_SUBSCRIPTIONS_CN.md:37`、`origin/test/xlabapi:docs/PRODUCT_SUBSCRIPTIONS_CN.md:89`、`origin/test/xlabapi:docs/superpowers/specs/2026-05-04-product-subscription-closure-design.md:13`。

### 代码占位（file:line）

| 位置 | 语义 | 观察 |
|---|---|---|
| `origin/test/xlabapi:backend/internal/service/user_subscription.go:5` | 旧 `UserSubscription` | 字段只有 `UserID`、`GroupID`、窗口用量、状态、分配信息；限额检查直接读 `Group`。 |
| `origin/test/xlabapi:backend/internal/service/user_subscription.go:98` | 旧分组限额检查 | `CheckDailyLimit/Weekly/Monthly` 使用 `group.Has*Limit()` 与 `group.*LimitUSD`。 |
| `origin/test/xlabapi:backend/internal/service/subscription_service.go:42` | 旧订阅服务 | `SubscriptionService` 聚合 `groupRepo` + `userSubRepo`，负责 legacy subscription。 |
| `origin/test/xlabapi:backend/internal/service/subscription_service.go:625` | 旧 admin 分组订阅列表 | `ListGroupSubscriptions(ctx, groupID, ...)` 明确按分组列旧订阅。 |
| `origin/test/xlabapi:backend/internal/handler/admin/subscription_handler.go:322` | 旧 admin API | 管理端按 `groupID` 调用 `ListGroupSubscriptions`。 |
| `origin/test/xlabapi:backend/internal/service/subscription_product.go:21` | 新产品定义 | `SubscriptionProduct` 带 `Code/Name/ProductFamily/DefaultValidityDays/DailyLimitUSD/WeeklyLimitUSD/MonthlyLimitUSD`。 |
| `origin/test/xlabapi:backend/internal/service/subscription_product.go:77` | 产品绑定真实分组 | `SubscriptionProductBinding` 同时含产品字段、真实 `GroupID`、`GroupSubscription`、`DebitMultiplier`。 |
| `origin/test/xlabapi:backend/internal/service/subscription_product.go:115` | 用户产品订阅实例 | `UserProductSubscription` 保存产品实例、共享窗口用量、昨日结转。 |
| `origin/test/xlabapi:backend/internal/service/subscription_product.go:269` | 新运行时上下文 | `ProductSettlementContext{Binding, Subscription}` 是网关/计费传递产品扣费语义的对象。 |
| `origin/test/xlabapi:backend/internal/service/subscription_product.go:296` | 新 repository port | `GetActiveProductSubscriptionByUserAndGroupID`、`ListActiveProductsByUserID`、`AssignOrExtendProductSubscription` 等均是产品维度。 |
| `origin/test/xlabapi:backend/internal/repository/subscription_product_repo.go:35` | 产品解析 | 以用户 + 真实分组 + 可选产品族解析活跃产品订阅。 |
| `origin/test/xlabapi:backend/internal/repository/subscription_product_repo.go:80` | 产品解析 SQL | 查询 `user_product_subscriptions ups` 连接产品/绑定/分组。 |
| `origin/test/xlabapi:backend/internal/repository/subscription_product_repo.go:124` | 绑定结果扫描 | 扫入 `binding.GroupSubscription`，说明产品轨仍需要读取真实分组的订阅类型。 |
| `origin/test/xlabapi:backend/internal/service/subscription_product_service.go:26` | 产品服务入口 | `GetActiveProductSubscriptionForFamily` 包装 repository 解析并返回 `ProductSettlementContext`。 |
| `origin/test/xlabapi:backend/internal/service/subscription_product_service.go:226` | 产品额度校验 | `CheckProductLimits` 使用产品定义和 `UserProductSubscription` 用量，而不是 group 限额。 |
| `origin/test/xlabapi:backend/internal/server/middleware/api_key_auth.go:147` | 网关双轨分流 | API Key 鉴权先尝试产品订阅，记录 `productSubscriptionChecked`。 |
| `origin/test/xlabapi:backend/internal/server/middleware/api_key_auth.go:180` | legacy fallback guard | 只有产品服务没有检查时才走 legacy `subscriptionService`，避免产品解析失败后隐式回退旧订阅。 |
| `origin/test/xlabapi:backend/internal/server/middleware/api_key_auth.go:230` | 产品额度不足兜底 | 产品额度不足时按显式余额兜底逻辑处理，不转旧订阅。 |
| `origin/test/xlabapi:backend/internal/service/product_settlement.go:24` | 产品 usage log 标记 | `applyProductSettlementUsageLog` 写 `ProductID/ProductSubscriptionID/ProductDebitCost`。 |
| `origin/test/xlabapi:backend/internal/service/gateway_service.go:8486` | 网关 usage log 应用 | 订阅分组且存在 `productSettlement` 时写产品订阅扣费字段。 |
| `origin/test/xlabapi:backend/internal/repository/usage_billing_repo.go:114` | 产品扣费落库 | 当 `ProductDebitCost > 0 && ProductSubscriptionID != nil` 时增量扣 `user_product_subscriptions`。 |
| `origin/test/xlabapi:backend/internal/service/product_aware_subscription_assigner.go:29` | 旧 assigner 桥接新产品 | 分配订阅时如果能解析产品则调用 `AssignOrExtendProductSubscription`。 |
| `origin/test/xlabapi:backend/internal/service/product_aware_subscription_assigner.go:75` | 产品转旧 DTO | `userSubscriptionFromProductAssignment` 将产品分配结果投影成 `UserSubscription` 形态，属于兼容桥，不代表同一模型。 |
| `origin/test/xlabapi:backend/internal/service/redeem_service.go:217` | redeem 兼容入口 | 兑换时先 `normalizeLegacyGroupSubscriptionCode`，说明历史卡密仍可能携带旧 group 语义。 |
| `origin/test/xlabapi:backend/internal/service/redeem_service.go:443` | 产品卡密兑换 | `assignProductSubscriptionFromRedeem` 要求 product assigner 并分配产品订阅。 |

### DB 字段

| 表/字段 | 轨道 | 说明 |
|---|---|---|
| `user_subscriptions.user_id` / `group_id` | 旧 | 用户与单个订阅分组绑定。Go 结构对应 `UserSubscription.UserID/GroupID`：`origin/test/xlabapi:backend/internal/service/user_subscription.go:5`。 |
| `user_subscriptions.daily_usage_usd` / `weekly_usage_usd` / `monthly_usage_usd` | 旧 | 旧分组订阅自己的用量窗口。 |
| `groups.subscription_type` | 两轨交界 | 旧轨通过 group 判断是否订阅分组；新产品绑定也只能绑定真实订阅分组。产品绑定结构保留 `GroupSubscription`：`origin/test/xlabapi:backend/internal/service/subscription_product.go:92`。 |
| `subscription_products.code/name/product_family/default_validity_days/daily_limit_usd/weekly_limit_usd/monthly_limit_usd/sort_order/status` | 新 | 产品定义，见迁移 `origin/test/xlabapi:backend/migrations/140_restore_shared_subscription_products.sql:1` 与结构 `origin/test/xlabapi:backend/internal/service/subscription_product.go:21`。 |
| `subscription_product_groups.product_id/group_id/debit_multiplier/status/sort_order` | 新 | 产品到真实订阅分组的绑定；倍率是产品额度扣减倍率。见 `origin/test/xlabapi:backend/migrations/140_restore_shared_subscription_products.sql:32`。 |
| `user_product_subscriptions.user_id/product_id/starts_at/expires_at/status` | 新 | 用户产品订阅实例。见 `origin/test/xlabapi:backend/migrations/140_restore_shared_subscription_products.sql:56`。 |
| `user_product_subscriptions.daily_window_start/weekly_window_start/monthly_window_start` | 新 | 产品维度额度窗口。 |
| `user_product_subscriptions.daily_usage_usd/weekly_usage_usd/monthly_usage_usd` | 新 | 多个真实分组共享的产品用量池。 |
| `user_product_subscriptions.daily_carryover_in_usd/daily_carryover_remaining_usd` | 新 | 昨日额度结转，只在产品订阅模型中出现。 |
| `usage_logs.product_id/product_subscription_id/product_debit_cost` | 新 | usage log 对产品订阅计费的审计字段，迁移位置 `origin/test/xlabapi:backend/migrations/140_restore_shared_subscription_products.sql:94`。 |
| `api_keys.subscription_product_family` | 新歧义消除 | API Key 对同一真实订阅分组的产品族选择字段，迁移位置 `origin/test/xlabapi:backend/migrations/145_product_subscription_explicit_fallback_family.sql:8`。 |
| `product_subscription_migration_sources.product_subscription_id/legacy_user_subscription_id` | 迁移桥 | 记录旧 `user_subscriptions` 收敛到新 `user_product_subscriptions` 的来源，迁移位置 `origin/test/xlabapi:backend/migrations/141_converge_legacy_group_subscriptions_to_products.sql:29`。 |

### 是否真歧义

结论：语义上已基本拆开，命名和兼容桥仍制造阅读歧义。

- 真边界清楚：`UserSubscription` 是旧分组订阅；`UserProductSubscription` 是新产品订阅。新产品解析/扣费链路有独立 service、repo、context、usage log 字段。
- 真歧义点 1：产品绑定结构里有 `GroupSubscription` 字段，名称像“分组订阅实体”，实际是绑定分组的 `subscription_type`/订阅属性快照。读代码时容易把它误认为旧 `UserSubscription`。
- 真歧义点 2：`product_aware_subscription_assigner.go` 会把产品分配结果投影成 `UserSubscription` 返回，用于兼容旧接口；这个 DTO 形状会掩盖“已分配的是产品订阅”这个事实。
- 真歧义点 3：redeem/service 中仍有 `normalizeLegacyGroupSubscriptionCode` 与 `assignProductSubscriptionFromRedeem` 并存；同一“订阅兑换码”入口根据 `product_id`/`group_id` 分流，需要靠规则而不是类型名识别。
- 非歧义但高风险边界：`api_key_auth.go` 的 `productSubscriptionChecked` 是防止产品解析失败回退旧订阅的关键布尔；若未来重构时移除或误用，会重新打开隐式 fallback。

### 建议路线

1. 保持运行时规则不变：订阅分组请求优先产品订阅；一旦产品订阅服务执行过解析，不允许再因产品 not found/ambiguous/limit failed 隐式回退 legacy `user_subscriptions`。
2. 文档和代码命名统一：把审计/后续设计里的术语固定为 `legacy group subscription`、`product subscription`、`product binding group`、`product settlement`，避免单独说“subscription”指代不明。
3. 收窄兼容桥说明：在 `product_aware_subscription_assigner.go` 和 redeem 分流处补文档/注释时，应明确“返回 `UserSubscription` 只是旧接口兼容投影，不是 legacy subscription 落库”。本轮只读，不改。
4. 后续重构优先增加类型隔离测试：产品订阅解析失败不回退旧订阅、产品兑换码必须有 `product_id`、旧分组订阅仍可按 legacy 流程工作，这三类测试应作为改名/重构护栏。

---

## 2. 邀请码双轨：redeem invitation vs 分级邀请码 / affiliate

### 概念边界

| 轨道 | 核心模型 | 主要用途 | 生命周期 | 是否发放权益 |
|---|---|---|---|---|
| 注册准入邀请码 | `redeem_codes` 中 `type = invitation` | 注册开关打开时作为一次性注册门票 | 管理员生成，注册时校验 unused，注册后标记 used | 不直接返利；只允许注册通过 |
| 传统邀请增长 | `users.invite_code` + `users.invited_by_user_id` + `invite_reward_records` | 绑定邀请关系，记录基础邀请奖励 | 用户持有自己的 `invite_code`，新用户注册/后台可绑定 inviter | `InviteService.ApplyBaseRechargeRewards` 对商业余额卡密充值做 3% 基础奖励 |
| 新 affiliate / 分级返利 | `user_affiliates.aff_code` + `user_affiliate_ledger` + affiliate settings tiers | 邀请返利、专属邀请码、分级返佣、冻结/转余额 | 每个用户 lazy ensure affiliate profile；注册时可绑定 `affiliateCode`；管理员可改专属码/比例 | 充值完成后按全局/专属/分级比例累计返利额度，可冻结、可转余额 |

这里的“邀请码”在代码里至少有三层含义：`invitationCode` 是注册门票，`invite_code` 是用户邀请关系码，`aff_code` 是 affiliate 返利码。它们都可能被中文叫“邀请码”，但业务语义不同。

### 代码占位（file:line）

| 位置 | 语义 | 观察 |
|---|---|---|
| `origin/test/xlabapi:backend/internal/service/redeem_code.go:10` | 兑换码实体 | `RedeemCode` 通用模型含 `Type/Value/Status/SourceType/UsedBy`，还含订阅相关 `GroupID/ProductID/ValidityDays`。 |
| `origin/test/xlabapi:backend/internal/service/domain_constants.go:55` | redeem 类型常量 | `RedeemTypeBalance/Concurrency/Subscription/Invitation` 共用一套 redeem 类型空间，另有 `RedeemTypeAffiliateBalance` 作为余额历史展示类型。 |
| `origin/test/xlabapi:backend/internal/service/redeem_service.go:145` | 批量生成 redeem code | `RedeemTypeInvitation` 不要求 value，生成随机卡密；这仍属于 redeem/card code，不是 affiliate 码。 |
| `origin/test/xlabapi:backend/internal/service/redeem_service.go:299` | redeem 使用入口 | `Redeem` 只处理 balance/concurrency/subscription；`invitation` 不在这里兑换，而是在注册流程校验/标记。 |
| `origin/test/xlabapi:backend/internal/service/auth_service.go:151` | 邮箱注册 invitation gate | 开启 `InvitationCodeEnabled` 后，要求 `invitationCode`，从 `redeemRepo.GetByCode` 查，必须 `Type == invitation` 且 unused。 |
| `origin/test/xlabapi:backend/internal/service/auth_service.go:245` | invitation 使用标记 | 注册成功后 `redeemRepo.Use(ctx, invitationRedeemCode.ID, user.ID)`，失败只记日志。 |
| `origin/test/xlabapi:backend/internal/service/auth_service.go:598` | OAuth 注册 invitation gate | OAuth 首次登录注册同样校验 `RedeemTypeInvitation`。 |
| `origin/test/xlabapi:backend/internal/service/auth_oauth_email_flow.go:56` | OAuth 邮箱流 invitation 校验 | 独立函数 `validateOAuthRegistrationInvitation` 复用 redeem invitation 语义。 |
| `origin/test/xlabapi:backend/internal/service/invite_service.go:18` | 传统邀请码生成 | `inviteCodeAlphabet` 生成 8 位大小写字母码，写入 `users.invite_code`。 |
| `origin/test/xlabapi:backend/internal/service/invite_service.go:136` | 传统邀请关系解析 | `ResolveInviterByCode` 通过 `userRepo.GetByInviteCode` 查邀请人。 |
| `origin/test/xlabapi:backend/internal/service/invite_service.go:165` | 传统邀请摘要 | `InviteSummary` 暴露 `InviteCode/InviteLink/InvitedUsersTotal/BaseRewardsTotal`。 |
| `origin/test/xlabapi:backend/internal/service/invite_service.go:210` | 传统基础奖励 | 只对 `RedeemTypeBalance` 且 `SourceType == commercial` 的余额卡密应用邀请基础奖励。 |
| `origin/test/xlabapi:backend/internal/service/invite.go:30` | 传统邀请奖励记录 | `InviteRewardRecord` 记录 inviter/invitee、触发 redeem code、奖励角色/类型/金额。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:60` | affiliate 用户摘要 | `AffiliateSummary` 暴露 `AffCode/AffCodeCustom/AffRebateRatePercent/InviterID/AffQuota/AffFrozenQuota`。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:98` | affiliate repository port | 包含 ensure profile、按 code 查找、绑定 inviter、累计返利、冻结解冻、转余额、管理专属码/比例、记录查询。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:233` | affiliate profile 初始化 | `EnsureUserAffiliate` 为用户创建/读取 `user_affiliates`。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:276` | affiliate 绑定 | `BindInviterByCode` 使用 `aff_code` 绑定 inviter；开关关闭时注册阶段静默忽略。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:325` | affiliate 返利累计 | `AccrueInviteRebateForOrder` 对充值金额按专属/分级/全局比例计算 rebate 并写 ledger/quota。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:412` | 分级倍率解析 | 没有用户专属比例时，按有效邀请人数匹配 `AffiliateRebateTiers`。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:478` | 返利转余额 | `TransferAffiliateQuota` 将可用返利额度转入用户余额。 |
| `origin/test/xlabapi:backend/internal/service/affiliate_service.go:579` | 管理员改专属码 | `AdminUpdateUserAffCode` 改写 `aff_code`，并标记 custom。 |
| `origin/test/xlabapi:backend/internal/repository/affiliate_repo.go:66` | affiliate ensure repo | `EnsureUserAffiliate` 缺 profile 时插入 `user_affiliates`。 |
| `origin/test/xlabapi:backend/internal/repository/affiliate_repo.go:74` | affiliate code 查询 | `GetAffiliateByCode` 按 `aff_code` 查邀请人。 |
| `origin/test/xlabapi:backend/internal/repository/affiliate_repo.go:117` | affiliate ledger/quota 累计 | `AccrueQuota` 更新 `aff_quota/aff_frozen_quota/aff_history_quota` 并写 ledger。 |
| `origin/test/xlabapi:backend/internal/repository/affiliate_repo.go:282` | affiliate 转余额落库 | `TransferQuotaToBalance` 原子扣 `aff_quota`、加 `users.balance`、写 transfer ledger。 |
| `origin/test/xlabapi:backend/internal/repository/affiliate_repo.go:1029` | aff_code 管理 | `UpdateUserAffCode` 改 `aff_code` 且 `aff_code_custom = true`。 |
| `origin/test/xlabapi:backend/internal/handler/admin/affiliate_handler.go:54` | admin affiliate API DTO | 管理接口使用 JSON 字段 `aff_code`、`aff_rebate_rate_percent`。 |
| `origin/test/xlabapi:backend/internal/handler/user_handler.go:180` | 用户 affiliate 页面 | 用户详情接口从 `AffiliateService.GetAffiliateDetail` 读取 affiliate 信息。 |

### DB 字段

| 表/字段 | 轨道 | 说明 |
|---|---|---|
| `redeem_codes.code/type/value/status/used_by/used_at` | 注册准入 + 通用兑换码 | 基础卡密表，初始迁移位置 `origin/test/xlabapi:backend/migrations/001_init.sql:118`。 |
| `redeem_codes.group_id/validity_days` | redeem 订阅兼容 | 旧订阅卡密字段，迁移位置 `origin/test/xlabapi:backend/migrations/005_schema_parity.sql:27`。 |
| `redeem_codes.source_type` | redeem 来源 | migration 139 为 redeem code 增加来源类型，商业充值/系统赠送等会影响邀请奖励触发，位置 `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:73`。 |
| `redeem_codes.product_id` | 产品订阅 redeem | 产品订阅卡密必须绑定产品，位置 `origin/test/xlabapi:backend/migrations/142_add_redeem_code_product_id.sql:6`。 |
| `users.invite_code` | 传统邀请增长 | 用户自己的邀请关系码，migration 139 添加并规范成 8 位字母，位置 `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:1`。 |
| `users.invited_by_user_id` / `invite_bound_at` | 传统邀请增长 | 记录用户被谁邀请以及绑定时间，位置 `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:1`。 |
| `invite_code_aliases.alias_code/user_id/source` | 传统邀请兼容 | 保存旧格式 invite_code 别名，避免迁移后旧码失效，位置 `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:10`。 |
| `invite_relationship_events` | 传统邀请审计 | 记录注册绑定、后台改绑等关系变更，位置 `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:97`。 |
| `invite_reward_records.trigger_redeem_code_id/trigger_redeem_code_value` | 传统邀请奖励 | 传统邀请奖励可以由商业余额 redeem code 触发，位置 `origin/test/xlabapi:backend/migrations/139_restore_invite_growth_tables.sql:138`。 |
| `user_affiliates.user_id/aff_code/inviter_id` | affiliate | affiliate 专属码和邀请人绑定，位置 `origin/test/xlabapi:backend/migrations/130_add_user_affiliates.sql:1`。 |
| `user_affiliates.aff_count/aff_quota/aff_history_quota` | affiliate | 邀请人数、可提现/可转余额返利、历史返利总额。 |
| `user_affiliates.aff_rebate_rate_percent/aff_code_custom` | affiliate 管理 | 专属返利比例和管理员自定义码标记，位置 `origin/test/xlabapi:backend/migrations/132_affiliate_custom_settings.sql:5`。 |
| `user_affiliates.aff_frozen_quota` | affiliate 冻结 | 返利冻结期内额度，位置 `origin/test/xlabapi:backend/migrations/133_affiliate_rebate_freeze.sql:2`。 |
| `user_affiliate_ledger.user_id/action/amount/source_user_id/frozen_until/source_order_id` | affiliate ledger | 返利累计/转余额/冻结/订单来源审计，位置 `origin/test/xlabapi:backend/migrations/131_affiliate_rebate_hardening.sql:13`、`origin/test/xlabapi:backend/migrations/134_affiliate_ledger_audit_snapshots.sql:4`。 |
| settings `invitation_code_enabled` | 注册准入 | 控制是否要求 redeem invitation code，DTO 位置 `origin/test/xlabapi:backend/internal/handler/dto/settings.go:33`。 |
| settings `affiliate_enabled` / `affiliate_rebate_*` / `affiliate_rebate_tiers` | affiliate | 控制 affiliate 开关、全局比例、冻结期、有效期、单人上限、分级比例，代码位置 `origin/test/xlabapi:backend/internal/service/setting_service.go:1544` 附近。 |

### 是否真歧义

结论：是，且是命名层面的真歧义，不是单纯代码重复。

- `invitationCode`、`invite_code`、`aff_code` 都能被中文 UI/文档叫“邀请码”，但一个是注册准入卡密，一个是用户邀请关系码，一个是 affiliate 返利码。
- `redeem_codes.type=invitation` 不通过 `RedeemService.Redeem` 发放权益，而在 `AuthService.Register*` 里校验并 `Use`。如果只搜索“redeem”，容易误以为所有 redeem code 都走同一兑换流程。
- 传统 `InviteService` 和新 `AffiliateService` 都涉及邀请奖励：前者基于 `users.invited_by_user_id` + `invite_reward_records` 给商业余额卡密充值发基础奖励；后者基于 `user_affiliates.inviter_id` + ledger/quota 做返利、冻结、分级、转余额。两套奖励模型并存，概念边界需要在规划里显式确认。
- affiliate 代码中 `AdminUpdateUserAffCode` 注释称“邀请码（专属邀请码）”，但字段是 `aff_code`；这会让“邀请码”在 admin 语境下更偏 affiliate，而在注册语境下更偏 redeem invitation。

### 建议路线

1. 术语分层固定：`registration invitation code` 只指 `redeem_codes.type=invitation`；`invite_code` 只指用户邀请关系码；`affiliate code` 只指 `user_affiliates.aff_code`。
2. 后续 UI/接口文案避免裸写“邀请码”：注册页用“注册邀请码/准入码”，affiliate 页面用“返利邀请码/推广码”，传统 invite 页面用“邀请关系码”。
3. 规划时确认传统 `InviteService` 与 `AffiliateService` 是否需要收敛；如果不收敛，应明确两套奖励是否会同时触发、触发源分别是什么，并补冲突测试。
4. 不建议把 `redeem` 和 `affiliate` 复用同一个 `code` 字段或同一个兑换入口。它们的生命周期、幂等、风控和收益结算完全不同。

---

## 3. 倍率三态：固定 / 动态 / 预期

### 概念边界

| 语义 | 代码里的主要名字 | 数据来源 | 用途 | 是否真实扣费 |
|---|---|---|---|---|
| 固定倍率 | `groups.rate_multiplier`、`user_group_rate_multipliers.rate_multiplier`、`usage_logs.rate_multiplier` | 分组默认倍率，可被用户专属分组倍率覆盖 | 固定定价分组的用户侧最终扣费倍率 | 是，进入 `CostBreakdown.ActualCost` 与 `usage_logs.rate_multiplier` |
| 账号/绑定倍率 | `accounts.rate_multiplier`、`account_groups.billing_multiplier`、`usage_logs.account_rate_multiplier` | 账号自身或账号-分组绑定 | 账号侧成本统计；动态分组下 `account_groups.billing_multiplier` 也作为用户侧实际扣费倍率 | 部分是：`account_groups.billing_multiplier` 会影响用户侧；`accounts.rate_multiplier` 主要影响账号成本口径 |
| 产品订阅扣减倍率 | `subscription_product_groups.debit_multiplier`、`usage_logs.group_debit_multiplier/product_debit_cost` | 产品绑定真实分组 | 把本次用户侧费用折算成产品额度池消耗 | 是，但扣的是产品额度，不是余额扣费倍率 |
| 动态预算倍率 | `api_keys.budget_multiplier`、`groups.default_budget_multiplier`、`DefaultBudgetMultiplier` | API Key 或动态分组默认值 | 选择动态账号时的预算上限/目标，不直接等于本次扣费倍率 | 不是直接扣费；用于筛选和排序账号 |
| 展示/预期倍率 | `dynamic_multiplier_min/max`、`dynamic_budget_multiplier`、`dynamic_budget_matched_multiplier` | 可用渠道视图从账号绑定倍率推导 | 给前端展示动态分组可用区间、预算和当前可匹配倍率 | 不是落账字段，只是摘要/预期 |

任务里提到的 `fixed_multiplier` / `expected_multiplier` 在 `origin/test/xlabapi:backend/` 没有直接命中。当前代码的实际三态是：固定计费倍率、动态预算/账号选择倍率、展示用动态预期倍率。

### 代码占位（file:line）

| 位置 | 语义 | 观察 |
|---|---|---|
| `origin/test/xlabapi:backend/internal/service/group.go:17` | 固定分组默认倍率 | `Group.RateMultiplier` 是固定分组默认计费倍率。 |
| `origin/test/xlabapi:backend/internal/service/group.go:22` | 动态定价模式 | `PricingMode` + `DefaultBudgetMultiplier` 标识 fixed/dynamic 和动态默认预算。 |
| `origin/test/xlabapi:backend/internal/service/group.go:88` | 动态分组判断 | `IsDynamicPricing()` 只看 `pricing_mode == dynamic`。 |
| `origin/test/xlabapi:backend/internal/service/user_group_rate.go:5` | 用户专属固定倍率 | `UserGroupRateEntry.RateMultiplier` 可空，NULL 表示未设置。 |
| `origin/test/xlabapi:backend/internal/service/user_group_rate_resolver.go:44` | 固定倍率解析 | 以 `groupDefaultMultiplier` 为默认，若用户专属倍率存在则覆盖。 |
| `origin/test/xlabapi:backend/internal/service/account.go:30` | 账号倍率 | `Account.RateMultiplier` 是账号计费倍率，nil 兼容旧缓存按 1.0。 |
| `origin/test/xlabapi:backend/internal/service/account.go:80` | 账号倍率归一 | `BillingRateMultiplier()` nil/负数兜底为 1.0，允许 0 表示账号计费为 0。 |
| `origin/test/xlabapi:backend/internal/service/account_group.go:5` | 账号-分组绑定倍率 | `AccountGroup.BillingMultiplier` 是账号在特定分组下的用户侧扣费乘数。 |
| `origin/test/xlabapi:backend/internal/service/account_group.go:21` | 绑定倍率归一 | 未配置/非法时按 1.0。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:13` | 动态定价常量 | 定义 `fixed/dynamic`、预算倍率范围 3-50、默认预算 8。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:39` | 动态预算状态 | `dynamicPricingBudgetState` 记录预算倍率、7 日窗口、当前平均倍率、下一次标准成本估计。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:50` | 真实倍率解析结构 | `billingMultiplierResolution` 拆成基础倍率、调整倍率、最终倍率、定价模式。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:110` | 预算倍率解析 | 优先 API Key `BudgetMultiplier`，再 group 默认，最后系统默认 8。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:123` | 最终扣费倍率解析 | fixed：`group/user group rate * account group billing_multiplier`；dynamic：基础倍率固定 1，再乘账号分组绑定倍率。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:177` | 预算窗口状态构建 | 基于 API Key 7 日 usage stats 计算当前平均倍率与下一请求估算标准成本。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:247` | 动态账号偏好排序 | 预算内优先选更高 multiplier；都超预算时选较低 multiplier。 |
| `origin/test/xlabapi:backend/internal/service/dynamic_pricing.go:288` | 动态预算准入 | 单账号倍率超预算时，若 7 日平均仍低于预算可放行，否则 block。 |
| `origin/test/xlabapi:backend/internal/service/api_key_service.go:466` | 创建 API Key 动态预算默认 | 动态分组创建 key 时，未传预算则用 group default 或系统 default。 |
| `origin/test/xlabapi:backend/internal/service/api_key_service.go:692` | 更新到动态分组要求预算 | 切换到动态分组时 `BudgetMultiplier` 必填，否则 `ErrAPIKeyBudgetRequired`。 |
| `origin/test/xlabapi:backend/internal/service/openai_gateway_service.go:1531` | 动态账号选择接入 | 账号选择时调用 `compareDynamicPricingAccountPreference`。 |
| `origin/test/xlabapi:backend/internal/service/openai_gateway_service.go:1581` | 动态预算准入接入 | 调度前调用 `isAccountWithinDynamicBudget`。 |
| `origin/test/xlabapi:backend/internal/service/openai_gateway_service.go:1614` | 预算状态注入 | OpenAI 账号选择前把 dynamic budget state 放入 context。 |
| `origin/test/xlabapi:backend/internal/service/openai_gateway_service.go:1730` | 动态预算阻断 | 候选账号不满足预算时标记 `dynamicBudgetBlocked`，最终可能返回 `ErrDynamicPricingBudgetExceeded`。 |
| `origin/test/xlabapi:backend/internal/service/openai_gateway_service.go:5292` | OpenAI usage 落账倍率 | 记录 usage 时调用 `resolveBillingMultiplierForUsage`，动态分组用账号分组倍率作为用户侧倍率。 |
| `origin/test/xlabapi:backend/internal/service/gateway_service.go:8423` | 通用 gateway usage 落账倍率 | 同样使用 `resolveBillingMultiplierForUsage`。 |
| `origin/test/xlabapi:backend/internal/service/openai_gateway_record_usage_test.go:283` | 动态倍率测试 | 验证动态分组下 `account_groups.billing_multiplier` 写入 `usageLog.RateMultiplier` 并影响费用。 |
| `origin/test/xlabapi:backend/internal/service/channel_available.go:125` | 展示预期倍率摘要 | `applyDynamicGroupSummary` 推导动态分组的最小/最大/预算/预算内匹配倍率。 |
| `origin/test/xlabapi:backend/internal/handler/available_channel_handler.go:63` | 前端可见动态字段 | 用户可见 DTO 暴露 `default_budget_multiplier`、`dynamic_multiplier_min/max`、`dynamic_budget_multiplier`、`dynamic_budget_matched_multiplier`。 |
| `origin/test/xlabapi:backend/internal/service/product_settlement.go:24` | 产品扣减倍率 | 产品订阅 usage log 写 `GroupDebitMultiplier` 和 `ProductDebitCost = totalCost * debit_multiplier`。 |
| `origin/test/xlabapi:backend/internal/service/usage_log.go:148` | usage log 倍率字段 | `RateMultiplier` 是用户侧最终倍率；`AccountRateMultiplier` 是账号倍率快照；`GroupDebitMultiplier` 是产品额度扣减倍率。 |
| `origin/test/xlabapi:backend/internal/service/account_stats_pricing.go:8` | 账号成本预计算 | `account_stats_cost` 为账号成本基础价，nil 时 dashboard 用 `total_cost * account_rate_multiplier`。 |
| `origin/test/xlabapi:backend/internal/repository/dashboard_aggregation_repo.go:334` | 账号成本聚合 | 账号成本表达式为 `COALESCE(account_stats_cost, total_cost) * COALESCE(account_rate_multiplier, 1)`。 |

### DB 字段

| 表/字段 | 语义 | 说明 |
|---|---|---|
| `groups.rate_multiplier` | 固定分组默认倍率 | 初始 schema 字段，位置 `origin/test/xlabapi:backend/migrations/001_init.sql:27`。 |
| `usage_logs.rate_multiplier` | 用户侧最终倍率快照 | 旧迁移添加，位置 `origin/test/xlabapi:backend/migrations/003_subscription.sql:59`；fixed 和 dynamic 最终都落这里。 |
| `accounts.rate_multiplier` | 账号倍率 | 账号侧成本倍率，位置 `origin/test/xlabapi:backend/migrations/037_add_account_rate_multiplier.sql:11`。 |
| `usage_logs.account_rate_multiplier` | 账号倍率快照 | 每条 usage log 的账号倍率快照，位置 `origin/test/xlabapi:backend/migrations/037_add_account_rate_multiplier.sql:14`。 |
| `user_group_rate_multipliers.rate_multiplier` | 用户专属固定分组倍率 | 覆盖 group 默认倍率，位置 `origin/test/xlabapi:backend/migrations/047_add_user_group_rate_multipliers.sql:3`；migration 127 放宽为 NULL。 |
| `user_group_rate_multipliers.rpm_override` | 同表非倍率配置 | 与倍率同表但语义是 RPM override，位置 `origin/test/xlabapi:backend/migrations/127_add_user_group_rpm_override.sql:9`。 |
| `subscription_product_groups.debit_multiplier` | 产品扣减倍率 | 产品绑定真实分组时的额度扣减倍数，位置 `origin/test/xlabapi:backend/migrations/140_restore_shared_subscription_products.sql:34`。 |
| `usage_logs.group_debit_multiplier/product_debit_cost` | 产品扣减快照 | 产品订阅计费审计字段，位置 `origin/test/xlabapi:backend/migrations/140_restore_shared_subscription_products.sql:96`。 |
| `account_groups.billing_multiplier` | 账号-分组绑定倍率 | 动态定价分组用于账号侧/用户侧倍率，位置 `origin/test/xlabapi:backend/migrations/148_add_account_group_billing_multiplier.sql:5`。 |
| `groups.pricing_mode` | fixed/dynamic 模式 | 位置 `origin/test/xlabapi:backend/migrations/149_add_group_dynamic_pricing_fields.sql:5`。 |
| `groups.default_budget_multiplier` | 动态默认预算倍率 | 动态分组默认 API Key 预算，位置 `origin/test/xlabapi:backend/migrations/149_add_group_dynamic_pricing_fields.sql:7`。 |
| `api_keys.budget_multiplier` | API Key 预算倍率 | ent schema 定义在 `origin/test/xlabapi:backend/ent/schema/api_key.go:51`；用于动态分组预算，不是直接落账倍率。 |
| `usage_logs.account_stats_cost` | 账号统计基础成本 | 与 `account_rate_multiplier` 组合用于账号成本聚合；service 注释位置 `origin/test/xlabapi:backend/internal/service/account_stats_pricing.go:8`。 |

### 是否真歧义

结论：是，但歧义主要来自同名 `multiplier` 横跨四条账务口径。

- `rate_multiplier` 在 fixed 分组中是用户侧扣费倍率；在 dynamic 分组中代码刻意把 base 设为 1，再用 `account_groups.billing_multiplier` 作为用户侧最终倍率。若只看字段名，会误以为 `groups.rate_multiplier` 对所有分组都直接生效。
- `budget_multiplier` 是动态预算/准入/排序目标，不是实际账单倍率；但前端可见字段里又有 `dynamic_budget_multiplier` 和 `dynamic_budget_matched_multiplier`，容易被当成当前请求的预估扣费倍率。
- `accounts.rate_multiplier` 与 `account_groups.billing_multiplier` 都像“账号倍率”，但前者主要用于账号成本统计快照，后者在 dynamic 分组下直接决定用户侧扣费倍率。
- `subscription_product_groups.debit_multiplier` 是产品额度消耗倍率，不是余额扣费倍率；它基于已经计算出的 `totalCost/actualCost` 再折算产品额度。
- `expected_multiplier` / `fixed_multiplier` 字面没有后端命中，说明文档/口头语和代码命名未对齐。后续需求如果继续使用这两个词，需要先映射到现有字段，避免新增第五套倍率名。

### 建议路线

1. 建立倍率术语表并强制在后续设计里使用：`user_billing_multiplier`、`account_cost_multiplier`、`account_group_billing_multiplier`、`product_debit_multiplier`、`dynamic_budget_multiplier`、`dynamic_display_matched_multiplier`。
2. UI 展示不要把 `budget_multiplier` 标成“当前倍率”；应标为“预算倍率/目标上限”。实际扣费倍率应来自 selected account 的 `billing_multiplier` 或最终 usage log。
3. 后续重构时优先保护 `resolveBillingMultiplierForUsage`：这是 fixed/dynamic 真正汇合点，任何改名都要用测试证明 fixed 用户专属倍率、dynamic account-group 倍率、账号成本倍率、产品 debit 倍率互不串线。
4. 如果要引入“预期倍率”概念，建议只作为 DTO 名称并明确“展示估计，不落账”；不要落 DB 字段 `expected_multiplier`，否则会和 usage log 的真实倍率快照冲突。

---

## 附录：执行过的只读检索

- `git show origin/test/xlabapi:docs/superpowers/plans/2026-05-02-product-subscription-restoration.md`
- `git show origin/test/xlabapi:docs/superpowers/plans/2026-05-04-product-subscription-closure.md`
- `git show origin/test/xlabapi:docs/superpowers/specs/2026-05-02-product-subscription-restoration-design.md`
- `git show origin/test/xlabapi:docs/superpowers/specs/2026-05-04-product-subscription-closure-design.md`
- `git show origin/test/xlabapi:docs/PRODUCT_SUBSCRIPTIONS_CN.md`
- `git grep -n -i "productsubscription\|groupsubscription\|product_subscription" origin/test/xlabapi -- backend/`
- `git grep -n -i "redeem\|affiliate\|invite_code\|invitation" origin/test/xlabapi -- backend/`
- `git grep -n -i "multiplier\|group_multiplier\|dynamic_group_budget\|fixed_multiplier\|expected_multiplier" origin/test/xlabapi -- backend/`
