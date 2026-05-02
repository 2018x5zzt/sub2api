# 产品订阅模块系统设计摘要

**日期：** 2026-05-02

## 目标

本设计把 Upstream 的“旧版订阅”和 xlabapi 的“新版产品订阅”分开治理。旧版订阅是一个兑换码激活一个分组；新版产品订阅是一个产品激活多个真实分组，并共享同一份产品额度池。

本轮新增能力：

- 产品订阅按产品族 `product_family` 做同族自动切换，不跨未来其他产品类型切换。
- 用户可以默认关闭地开启“订阅消耗完时，自动消耗余额”，并设置累计余额兜底上限。
- 分组增加 `balance_fallback_group_id`，订阅分组额度耗尽后只能切到明确映射的余额分组。
- 新增余额模式 `Team/Plus 混合余额号池`，倍率为 1。

完整中文设计记录保留在本地 `docs/PRODUCT_SUBSCRIPTIONS_CN.md`；该路径受 `.gitignore` 影响不会默认进入 git diff。

## 产品订阅规则

新版产品由 `subscription_products`、`subscription_product_groups` 和 `user_product_subscriptions` 组成。API Key 仍绑定真实 `groups.id`，请求运行时通过用户、真实分组和产品绑定解析产品订阅上下文。

产品用量写入同一个 `user_product_subscriptions` 记录。不同产品分组使用 `subscription_product_groups.debit_multiplier` 折算产品额度消耗。

昨日结转字段：

- `daily_carryover_in_usd`：今天带入的昨日可结转总额。
- `daily_carryover_remaining_usd`：今天剩余的昨日结转。
- `daily_usage_usd`：今天总产品消耗，包含结转和今日额度。

前端每日进度条按 `daily_limit_usd + daily_carryover_in_usd` 作为总可用额度，并展示今日额度与昨日结转组成。

## 产品选择与自动切换

旧逻辑按 `expires_at DESC` 取一个产品订阅，用户同时购买多个同分组产品时会选错产品。

新规则：

1. 查询用户在真实分组下所有活跃产品订阅。
2. 先按 `product_family`、`sort_order`、`starts_at`、`id` 排序。
3. 只在第一个产品族内寻找仍有剩余额度的产品。
4. 第一个同族产品额度耗尽后，自动切到同族下一个产品。
5. 遇到不同 `product_family` 立即停止，不跨族消耗。

## 余额兜底

用户字段：

- `subscription_balance_fallback_enabled`
- `subscription_balance_fallback_limit_usd`
- `subscription_balance_fallback_used_usd`

兜底只在用户开启、上限大于 0、已用未达上限、用户有余额、订阅分组配置了 `balance_fallback_group_id` 时生效。运行时会把本次请求的 API Key 分组临时切到映射的标准余额分组，因此 usage log 和实际扣费都是余额模式，不再写产品订阅用量。

扣费事务必须同时扣余额并累加 `subscription_balance_fallback_used_usd`。如果本次费用会超过兜底上限，事务失败，不允许静默透支兜底额度。

## 兼容边界

- 不批量迁移现有 API Key。
- 不批量迁移账号池绑定。
- 旧兑换码如果有 `product_id`，激活新版产品；只有旧版订阅字段时继续走旧版订阅。
- 商业售卖必须使用明确 `product_id`，不能用 `group_id` 猜产品。

## 验证

后端覆盖：

- 同族产品额度耗尽后切换到下一个同族产品。
- 不跨不同产品族自动切换。
- 余额兜底扣费事务同时扣余额和累加已用额度，超上限失败。

前端覆盖：

- 订阅页仍能展示产品分组和昨日结转。
- 订阅页挂载 Pinia/auth store 后测试通过。
- 类型检查通过。

