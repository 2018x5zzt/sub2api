# sub2api 上游重集成 migration 重排方案（2026-04-21）

## 1. 文档目的

这份文档用于解决一个非常具体的问题：

- 当我们从最新 `upstream/main` 重新切 replay 分支时
- 当前 `xlabapi` 上的本地 migration 不能原样照搬
- 必须先决定哪些 migration 保留、哪些改用 upstream 版本、哪些延后按语义补 delta

本方案只定义 migration 策略，不执行真实改号。

## 2. 基线结论

文档生成时已核对：

- `upstream/main = 78f691d2`
- `xlabapi = 390cd00a`

`upstream/main` 当前 migration 已经走到：

- `107_add_account_cost_to_dashboard_tables.sql`

但上游本身已经不是严格单线编号，已经存在重复前缀，例如：

- `095_channel_features.sql`
- `095_subscription_plans.sql`
- `101_add_account_stats_pricing.sql`
- `101_add_balance_notify_fields.sql`
- `101_add_channel_features_config.sql`
- `101_add_payment_mode.sql`

因此这次重集成不能再依赖“紧跟上游最后一个编号 +1”这种脆弱策略。

## 3. 编号策略

### 3.1 总原则

对本地保留的 replay-only migration，统一放入一个新的本地区间：

- `120_*` 起步

这样做的原因：

- 避开 upstream 当前已存在的 `001-107`
- 避开 upstream 已经出现的重复前缀污染
- 允许后续在 `128+` 继续补 replay 阶段发现的 local delta migration

### 3.2 三种处理方式

后续每个本地 migration 只允许落入三类之一：

1. **保留并改号**
   - upstream 没有等价能力
   - 本地语义必须保留

2. **丢弃本地文件，直接采用 upstream 版本**
   - upstream 已有同构 migration
   - 本地文件只是编号不同或注释差异

3. **不直接重放，等代码回放后再补 delta migration**
   - upstream 已有大部分能力
   - 本地 migration 与 upstream 高度重叠
   - 是否还需要额外 schema 变化，要看代码 port 后的真实差异

## 4. 建议保留并改号的 migration

这些 migration 在当前 `upstream/main` 中没有直接等价物，应作为 replay-only migration 保留。

### 4.1 新编号映射

- `077_add_promo_code_scene_and_success_message.sql`
  - 建议改为 `120_add_promo_code_scene_and_success_message.sql`
- `078_add_account_group_billing_multiplier.sql`
  - 建议改为 `121_add_account_group_billing_multiplier.sql`
- `079_add_benefit_red_packet_fields.sql`
  - 建议改为 `122_add_benefit_red_packet_fields.sql`
- `080_add_group_health_snapshots.sql`
  - 建议改为 `123_add_group_health_snapshots.sql`
- `081_add_dynamic_group_budget_multiplier.sql`
  - 建议改为 `124_add_dynamic_group_budget_multiplier.sql`
- `081_add_invite_growth_foundation.sql`
  - 建议改为 `125_add_invite_growth_foundation.sql`
- `082_add_invite_admin_ops.sql`
  - 建议改为 `126_add_invite_admin_ops.sql`
- `086_rotate_invite_codes_to_letters.sql`
  - 建议改为 `127_rotate_invite_codes_to_letters.sql`

### 4.2 保留原因

#### `120_add_promo_code_scene_and_success_message.sql`

保留原因：

- `promo_codes.scene`
- `promo_codes.success_message`

这两项直接支撑：

- 福利码 / 注册码场景区分
- 福利码兑换成功弹窗文案

上游当前没有等价字段。

#### `121_add_account_group_billing_multiplier.sql`

保留原因：

- 为 `account_groups` 增加绑定级 `billing_multiplier`

这项直接支撑：

- 动态倍率
- 分组扣费链路
- 本地 account-group 级别的价格控制

上游当前没有等价 migration。

#### `122_add_benefit_red_packet_fields.sql`

保留原因：

- `promo_codes.random_bonus_pool_amount`
- `promo_codes.random_bonus_remaining`
- `promo_codes.leaderboard_enabled`
- `promo_code_usages.fixed_bonus_amount`
- `promo_code_usages.random_bonus_amount`

这项直接支撑红包池和幸运排行榜。

#### `123_add_group_health_snapshots.sql`

保留原因：

- 新建 `group_health_snapshots`

这项支撑：

- group health snapshot 持久化
- 企业可见分组和 pool 状态观察

上游当前没有同名或同构表。

#### `124_add_dynamic_group_budget_multiplier.sql`

保留原因：

- `groups.pricing_mode`
- `groups.default_budget_multiplier`
- `api_keys.budget_multiplier`

这项直接支撑动态预算倍率。

#### `125_add_invite_growth_foundation.sql`

保留原因：

- `users.invite_code`
- `users.invited_by_user_id`
- `users.invite_bound_at`
- `redeem_codes.source_type`
- `invite_reward_records`

这项是邀请返利系统的数据基础。

#### `126_add_invite_admin_ops.sql`

保留原因：

- `invite_relationship_events`
- `invite_admin_actions`
- `invite_reward_records.admin_action_id`

这项是邀请后台管理与审计能力的数据基础。

#### `127_rotate_invite_codes_to_letters.sql`

保留原因：

- `invite_code_aliases`
- 字母邀请码轮换

这项是邀请体系后续兼容层的一部分，不属于一次性数据修补。

## 5. 建议直接丢弃、改用 upstream 版本的 migration

这些 migration 不应该在 replay 分支里继续保留为本地新文件。
应直接采用 `upstream/main` 已存在版本。

### 5.1 完全同构或实质同构

- 本地 `083_add_tls_fingerprint_profiles.sql`
  - 改用 upstream `080_create_tls_fingerprint_profiles.sql`
- 本地 `084_add_usage_log_requested_model.sql`
  - 改用 upstream `077_add_usage_log_requested_model.sql`
- 本地 `085_add_usage_log_requested_model_index_notx.sql`
  - 改用 upstream `078_add_usage_log_requested_model_index_notx.sql`
- 本地 `087_create_channels.sql`
  - 改用 upstream `081_create_channels.sql`
- 本地 `088_refactor_channel_pricing.sql`
  - 改用 upstream `082_refactor_channel_pricing.sql`
- 本地 `089_channel_model_mapping.sql`
  - 改用 upstream `083_channel_model_mapping.sql`
- 本地 `095_channel_billing_model_source.sql`
  - 改用 upstream `084_channel_billing_model_source.sql`
- 本地 `097_channel_restrict_and_per_request_price.sql`
  - 改用 upstream `085_channel_restrict_and_per_request_price.sql`
- 本地 `108_add_usage_log_billing_mode.sql`
  - 改用 upstream `087_usage_log_billing_mode.sql`

### 5.2 应视为 upstream 已吸收的 migration

- 本地 `107_add_account_cost_to_dashboard_tables.sql`
  - 默认改用 upstream `107_add_account_cost_to_dashboard_tables.sql`

说明：

- 本地与 upstream 的 SQL 结构一致，主要差异在注释语义
- replay 时不值得再生成第二个本地版本

## 6. 需要“代码回放后再判定”的 migration

这些功能面已经大面积被 upstream 吸收，但不排除本地代码还有少量 schema delta。
因此不建议现在直接拍板“新增本地 migration 文件”，而是先 port 代码，再看是否真的缺字段。

### 6.1 channel 相关后续 delta

默认先采用 upstream 现有链路：

- `081_create_channels.sql`
- `082_refactor_channel_pricing.sql`
- `083_channel_model_mapping.sql`
- `084_channel_billing_model_source.sql`
- `085_channel_restrict_and_per_request_price.sql`
- `087_usage_log_billing_mode.sql`
- `107_add_account_cost_to_dashboard_tables.sql`

只有在 port 完本地 channel 代码后，才允许新增 `128+` 的 delta migration。

例如：

- `128_add_channel_local_delta.sql`
- `129_add_usage_account_cost_local_delta.sql`

但默认不应提前创建。

### 6.2 TLS profile 后续 delta

默认先采用 upstream `080_create_tls_fingerprint_profiles.sql`。

只有在本地 TLS profile 代码与上游表结构不匹配时，才补：

- `130_add_tls_profile_local_delta.sql`

## 7. 实施顺序

真正执行 migration 重排时，建议按下面顺序：

1. 从 replay worktree 中删除不再采用的本地 migration 文件
2. 先保留 upstream 原生 migration 不动
3. 再把本地必须保留的 migration 统一改到 `120-127`
4. 修改任何引用旧文件名的测试、schema 校验、文档
5. 跑一次 migration/schema 级验证
6. 只有在代码回放后仍缺 schema 时，才新增 `128+` delta migration

## 8. 变更清单

### 8.1 建议在 replay 分支中新增的 migration 文件名

- `120_add_promo_code_scene_and_success_message.sql`
- `121_add_account_group_billing_multiplier.sql`
- `122_add_benefit_red_packet_fields.sql`
- `123_add_group_health_snapshots.sql`
- `124_add_dynamic_group_budget_multiplier.sql`
- `125_add_invite_growth_foundation.sql`
- `126_add_invite_admin_ops.sql`
- `127_rotate_invite_codes_to_letters.sql`

### 8.2 建议不再保留为本地 replay 文件的 migration

- `083_add_tls_fingerprint_profiles.sql`
- `084_add_usage_log_requested_model.sql`
- `085_add_usage_log_requested_model_index_notx.sql`
- `087_create_channels.sql`
- `088_refactor_channel_pricing.sql`
- `089_channel_model_mapping.sql`
- `095_channel_billing_model_source.sql`
- `097_channel_restrict_and_per_request_price.sql`
- `107_add_account_cost_to_dashboard_tables.sql`
- `108_add_usage_log_billing_mode.sql`

## 9. 风险与注意事项

### 9.1 不要把“文件名不同”误判成“语义不同”

本次核对已经证明，多条本地 migration 只是编号与 upstream 不同，内容却同构。

如果不先做这个识别，后续非常容易：

- 重复建表
- 重复加列
- 在 schema 校验里制造噪音

### 9.2 不要在当前阶段为 channel 预写大量新 migration

channel 这条线已经大量吸收进 upstream。

更稳的做法是：

- 先复用 upstream migration
- 等代码 port 完后，再补最小 delta

### 9.3 预留 `128+` 是为了未来 delta，不是为了现在凑数

只有“上游已有基础，但本地 replay 后仍缺 schema”的情况，才应该用 `128+`。

## 10. 当前建议结论

如果下一步要真正开始 replay，migration 层面应该执行的是：

- 保留并改号：`120-127`
- 直接采用 upstream：requested_model、tls profile、channels、billing_mode、account_cost
- channel / usage / tls 余量：代码 port 后再决定是否补 `128+` delta

这样做的好处是：

- 编号空间清晰
- 与 upstream 的语义边界清晰
- 后续每加一个本地 migration，都知道它为什么存在
