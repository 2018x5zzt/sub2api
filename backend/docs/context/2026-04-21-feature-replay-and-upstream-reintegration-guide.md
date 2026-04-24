# sub2api 特性回放与上游重集成指南（2026-04-21）

## 1. 文档目的

这份文档用于回答一个具体工程问题：

- 以 `upstream/main` 最新提交为新基线
- 不直接复用当前脏工作区
- 按可控顺序把 `xlabapi` 上已经开发完成的本地能力重新融合进去

它不是发布公告，也不是一次性 merge 记录。

它是后续执行“从上游主线重新切出集成分支，并按功能包回放我们自己的修改”时的作战手册。

## 2. 适用仓库与基线

文档生成时的实际基线：

- 仓库：`/root/sub2api-src`
- 本地集成分支：`xlabapi`
- 本地主分支：`main`
- 上游远程：`upstream = git@github.com:Wei-Shaw/sub2api`
- 本次文档写作前已执行：`git fetch upstream`

刷新后的关键提交：

- `upstream/main = 78f691d2`
- `main = 7e5b25c5`
- `xlabapi = 390cd00a`
- `merge-base(upstream/main, xlabapi) = 6a2cf09e`

当前分叉规模：

- `upstream/main` 相对共同基线多 `412` 个提交
- `xlabapi` 相对共同基线多 `166` 个提交

这说明后续工作不应该被理解成“跟上游小步同步一下”。
它本质上是一次新的集成项目。

## 3. 当前工作区状态

文档生成时，`/root/sub2api-src` 不是干净工作区，仍有未提交修改：

- `backend/internal/domain/constants_test.go`
- `backend/internal/pkg/antigravity/claude_types_test.go`
- `backend/internal/pkg/antigravity/request_transformer_test.go`
- `backend/internal/pkg/claude/constants_test.go`
- `backend/internal/service/antigravity_model_mapping_test.go`
- `backend/internal/service/billing_service_test.go`
- `backend/internal/service/pricing_service_test.go`
- `docs/superpowers/specs/2026-04-11-dynamic-group-budget-multiplier-design.md`

因此后续重集成执行规则必须是：

- 不在当前 `xlabapi` 脏工作区上直接开工
- 必须从 `upstream/main` 新建独立 worktree 或新分支
- 只把已确认的功能包按顺序回放

## 4. 总体判断

### 4.1 可行性结论

- 写清楚“功能包 + 依赖顺序 + 核心文件 + 验收入口”的文档，可行性高
- 基于最新 `upstream/main` 重建一条新集成分支，可行性中高
- 直接把当前 `xlabapi` 整体 merge 到 `upstream/main`，可行性中低

### 4.2 根本原因

难点不在 git 命令本身，而在于这些功能已经跨越多个层次：

- `ent schema / migration`
- `repository`
- `service`
- `handler / route`
- `frontend page / settings / i18n`

也就是说，这不是一组可以随手启停的“插件式 patch”。

## 5. 回放原则

后续重集成必须遵守下面几条原则：

1. 不直接运行 `git merge xlabapi` 或 `git merge upstream/main` 试图一次性解决全部问题。
2. 把本地修改拆成“功能包”而不是“分支名列表”。
3. 先铺运行时和兼容性底座，再回放业务功能。
4. 每个功能包都要带上：
   - 来源分支
   - 关键提交
   - 关键文件
   - 依赖前置
   - 验收点
5. 迁移文件编号在真正执行前必须统一重排，不允许把当前编号直接照搬进新基线。
6. 企业 BFF、channel pricing、usage account cost 这类大功能可以延后，但不能在文档里被省略。

## 6. 功能包总表

下面不是按提交时间排序，而是按后续重集成时更合理的“功能包”排序。

### 6.1 运行时兼容与 OpenAI/OAuth 底座

这是最先回放的一组，因为它决定网关、OAuth、流式转发、错误处理、上游兼容是否稳定。

来源分支与关键提交：

- `upgrade-v0.1.106-merge`
  - `08108fdf feat(openai): add 429 silent failover`
- `api-key-quota-help-url-plus-oauth-itemrefs-20260412-084117`
  - `23770f48 fix(openai): extract oauth system input_text prompts`
- `openai-oauth-store-false-itemrefs`
  - `8cbb4b1d fix(openai): drop store-false native oauth item refs`
- `reasoning-audit-20260415`
  - `04d21ea7 fix(openai): surface missing terminal stream failures`
- `fix/gpt54-xhigh-relay`
  - `1ffd9b71 fix(gateway): avoid empty 200 on responses/chat streams`
- `gpt54-xhigh-relay-fix-20260420`
  - `7cdaf9a0 fix(openai): normalize gpt-5.4 reasoning relay aliases`
- `integrate/xlabapi-upstream-v01114-core-20260420`
  - `210fe224 merge: integrate upstream v0.1.114 core fixes into xlabapi`

核心文件热区：

- `backend/internal/service/openai_gateway_service.go`
- `backend/internal/service/openai_gateway_messages.go`
- `backend/internal/service/gateway_service.go`
- `backend/internal/service/openai_codex_transform.go`
- `backend/internal/service/ratelimit_service.go`
- `backend/internal/service/scheduler_snapshot_service.go`
- `backend/internal/repository/scheduler_cache.go`

回放目标：

- 保住 `Responses / ChatCompletions` 兼容层
- 保住 OAuth passthrough / item refs / relay 逻辑
- 补齐上游 runtime 修复
- 降低直接回放业务功能时的底层噪音

### 6.2 福利码 / 红包 / 排行榜

这是当前 `main` 已经承载的一组功能，但在真正重集成时，仍应作为独立功能包处理，不要默认它已经自然存在于最新 upstream/main。

关键提交：

- `89001d11 release: snapshot deployed group billing multiplier build`
- `4d8e8ec6 feat(benefit): add lucky red packet leaderboard`
- `7e5b25c5 feat(benefit): add lucky red packet leaderboard`
- `023bca90 fix(redeem): limit lucky leaderboard entry to benefit dialog`
- `a3e91b74 fix(frontend): allow leaderboard access after repeat redeem`
- `8c41916c fix(frontend): keep leaderboard access after benefit exhaustion`
- `28ee6d42 fix(benefit): return full leaderboard entries`

核心文件：

- `backend/migrations/077_add_promo_code_scene_and_success_message.sql`
- `backend/migrations/079_add_benefit_red_packet_fields.sql`
- `backend/internal/service/promo_service.go`
- `backend/internal/handler/redeem_handler.go`
- `frontend/src/views/admin/PromoCodesView.vue`
- `frontend/src/views/user/RedeemView.vue`

功能点：

- benefit 场景 promo code
- 随机红包池
- `leaderboard_enabled`
- 兑换后查看幸运排行榜
- 排行榜显示范围被收敛到兑换结果弹窗
- 重复兑换或权益耗尽后仍允许查看榜单
- 需要用户名才能进入榜单

### 6.3 邀请增长 / 人拉人 / 返利系统

这是 `xlabapi` 最重的一条业务线，不能再按“邀请码小修小补”理解。

关键提交：

- `6be674c0 feat: formalize invite foundation`
- `a5a94ef7 feat: retire legacy invitation code flow`
- `3e58cf6b fix: make invite reward settlement atomic`
- `1a555fdc feat: tune baseline invite reward to three percent`
- `a668d3b3 feat: soften invite center reward copy`
- `28d9a9ff feat(invite): rotate codes with alias compatibility`
- `83092b3d ops(invite): add legacy admin bind backfill`
- `402a4d23 fix: tighten invite duplicate settlement handling`

核心文件：

- `backend/migrations/081_add_invite_growth_foundation.sql`
- `backend/migrations/082_add_invite_admin_ops.sql`
- `backend/migrations/086_rotate_invite_codes_to_letters.sql`
- `backend/internal/service/invite.go`
- `backend/internal/service/invite_service.go`
- `backend/internal/service/admin_invite.go`
- `backend/internal/service/admin_service_invite.go`
- `backend/internal/handler/invite_handler.go`
- `backend/internal/handler/admin/invite_handler.go`
- `backend/internal/service/auth_service.go`
- `frontend/src/views/auth/RegisterView.vue`
- `frontend/src/views/user/InviteView.vue`
- `frontend/src/views/admin/InvitesView.vue`

功能点：

- 永久用户邀请码
- 邀请链接注册绑定
- 用户邀请中心
- 管理后台邀请关系、手工补发、重算
- 邀请关系事件与奖励台账
- 旧版一次性邀请码流程退役
- 返利结算原子化
- 邀请人 `3%`、被邀请人 `3%` 的双边返利规则

### 6.4 动态倍率 / 调度算法 / 计价约束

这是“预算倍率约束路由，但不改最终账单公式”的一组设计与实现。

来源分支与关键提交：

- `dynamic-group-budget-multiplier`
  - `ed344ee2 feat(backend): add dynamic group budget pricing`
  - `f37b4ee0 feat(frontend): add budget multiplier controls`
  - `a34a5c3a fix: fail over on anthropic capacity-limit 400`
- `dynamic-pricing-tiered-selection-20260420`
  - `31787590 fix(dynamic): prefer target-fit accounts in pricing groups`

核心文件：

- `docs/superpowers/specs/2026-04-11-dynamic-group-budget-multiplier-design.md`
- `backend/migrations/081_add_dynamic_group_budget_multiplier.sql`
- `backend/internal/service/dynamic_pricing.go`
- `backend/internal/service/api_key_service.go`
- `backend/internal/handler/api_key_handler.go`
- `backend/internal/handler/admin/group_handler.go`
- `frontend/src/components/account/CreateAccountModal.vue`
- `frontend/src/components/account/EditAccountModal.vue`

功能点：

- group 新增 `fixed / dynamic` pricing mode
- group 默认 `default_budget_multiplier`
- API key 独立 `budget_multiplier`
- 最近 7 天标准成本加权平均倍率预算
- 超预算直接拒绝或过滤候选账号
- 最终扣费仍按 group multiplier 与 account-group multiplier 结算
- 候选账号优先选择更贴合预算的目标账号

### 6.5 企业可见分组 / enterprise-bff / group health

这条线对应的是企业前台隔离和企业侧 key 写入授权。

关键提交：

- `482477ca feat(enterprise-bff): add enterprise-aware company login`
- `7ce45e8b feat(enterprise-bff): filter visible pool status`
- `16c347e8 feat(bff): enforce enterprise group authorization for key writes`
- `89cba682 feat(admin): configure enterprise visible groups in settings`
- `5302da0a fix: restrict enterprise groups to configured mappings`
- `0b7a51b2 fix: restrict enterprise groups to configured mappings`
- `bf37a2a0 feat(core): collect group health history in background`
- `4c5d75b4 feat(core): persist group health snapshots`

核心文件：

- `backend/cmd/enterprise-bff/main.go`
- `backend/internal/enterprisebff/`
- `backend/internal/service/enterprise_visible_groups_setting.go`
- `backend/internal/service/setting_service.go`
- `frontend/src/views/admin/SettingsView.vue`
- `backend/migrations/080_add_group_health_snapshots.sql`

功能点：

- 企业登录与 BFF 转发
- 企业用户可见分组映射
- 企业 key 写入 group 授权
- pool status 过滤
- group health snapshot 采集与持久化

### 6.6 Channel pricing / usage account cost / 管理后台渠道面

这条线非常大，执行时建议单独视作一个独立阶段，而不是与前面业务功能混在一起。

关键提交：

- `aa4dfa30 feat(integrate): absorb channel pricing and usage account cost`
- `b6ee1f06 feat(channel): apply mapping and restrictions across gateway paths`

核心文件：

- `backend/internal/handler/admin/channel_handler.go`
- `backend/internal/repository/channel_repo.go`
- `backend/internal/repository/channel_repo_pricing.go`
- `backend/internal/service/channel_service.go`
- `backend/internal/service/model_pricing_resolver.go`
- `frontend/src/views/admin/ChannelsView.vue`
- `backend/migrations/087_create_channels.sql`
- `backend/migrations/088_refactor_channel_pricing.sql`
- `backend/migrations/089_channel_model_mapping.sql`
- `backend/migrations/095_channel_billing_model_source.sql`
- `backend/migrations/097_channel_restrict_and_per_request_price.sql`
- `backend/migrations/107_add_account_cost_to_dashboard_tables.sql`
- `backend/migrations/108_add_usage_log_billing_mode.sql`

功能点：

- 后台 channel 管理
- channel model mapping
- channel pricing
- usage account cost 展示
- gateway path 上的 mapping / restriction 落地

## 7. 推荐回放顺序

如果目标是“基于最新 `upstream/main` 重建一条新集成分支，并尽可能早得到可运行状态”，建议顺序如下。

### 阶段 0：冻结基线

- 从最新 `upstream/main` 新建 worktree
- 不复用当前 `xlabapi`
- 记录当前 `xlabapi` 功能包清单

### 阶段 1：先回放运行时兼容底座

先处理 6.1。

原因：

- 这是所有后续业务功能的共同底座
- 不先补 runtime/relay/OAuth 兼容，后面业务回放会把失败信号污染在一起

### 阶段 2：回放福利码 / 红包 / 排行榜

先处理 6.2。

原因：

- 依赖面相对可控
- 已经在 `main` 上跑过
- 能最快得到一个完整的用户侧业务功能面

### 阶段 3：回放邀请返利系统

再处理 6.3。

原因：

- 这是最大的业务改动
- 依赖 auth/register/redeem/admin/user 多处表面
- 应该建立在阶段 1 和 2 已稳定的前提上

### 阶段 4：回放动态倍率与预算约束

再处理 6.4。

原因：

- 它会动到 API key、group、gateway、billing
- 与邀请返利基本无直接耦合，但和计费、调度深度耦合

### 阶段 5：回放 enterprise-visible-groups 与 enterprise-bff

再处理 6.5。

原因：

- 面向企业侧，和普通用户流程解耦度更高
- 可以在基础业务面稳定后单独拉起验证

### 阶段 6：最后回放 channel pricing / usage account cost

最后处理 6.6。

原因：

- 这是最大的技术面 patch 之一
- 冲突热区多，最容易拖慢整体进度
- 放在最后更容易判断哪些冲突来自它本身，哪些来自前面阶段

## 8. 绝对不能忽略的风险

### 8.1 migration 编号冲突

当前本地树内部已经存在并行编号：

- `079_add_benefit_red_packet_fields.sql`
- `079_ops_error_logs_add_endpoint_fields.sql`
- `080_add_group_health_snapshots.sql`
- `080_create_tls_fingerprint_profiles.sql`
- `081_add_dynamic_group_budget_multiplier.sql`
- `081_add_invite_growth_foundation.sql`

这说明：

- 当前编号体系已经不是单线发展
- 真正回放到新基线前，必须统一迁移编号和顺序
- 这件事不能等到部署前再处理

### 8.2 共享冲突热区

这些目录在 `upstream/main...xlabapi` 差异里占比最高：

- `backend/internal/service/`
- `backend/ent/`
- `backend/internal/repository/`
- `backend/internal/handler/`
- `frontend/src/views/admin/`
- `frontend/src/views/user/`

实际执行时，优先把这些区域视为“高冲突区”。

### 8.3 不能把分支名当作功能边界

例如：

- 邀请返利并没有一直保留成一条单独头分支
- `xlabapi` 上很多能力是通过多个 merge 和后续修补拼起来的

所以真正执行时，应该按本文件的功能包划分，而不是简单说：

- merge 这个分支
- 再 merge 那个分支

## 9. 推荐执行方式

建议未来真正执行时采用下面的方式，而不是直接在本仓库里动：

1. `git fetch upstream origin`
2. 从 `upstream/main` 新建独立 worktree
3. 建一条新的 replay/integration 分支
4. 先回放“运行时兼容底座”
5. 每完成一个功能包，就立即做一次定向验证
6. 每个阶段单独提交，保持可回退

不建议的做法：

- 直接把 `xlabapi` rebase 到 `upstream/main`
- 直接 merge 整条 `xlabapi`
- 在当前脏工作区边修边 merge

## 10. 后续文档拆分建议

这份文档是总索引。真正执行前，建议再拆出三份子文档：

1. `功能包回放清单`
   - 每个功能包的提交范围、文件范围、测试范围

2. `migration 重排方案`
   - 新编号、执行顺序、兼容策略

3. `upstream 重集成执行计划`
   - 具体 worktree、分支名、执行顺序、验证门禁

## 11. 当前建议结论

基于 2026-04-21 的仓库状态，最合理的下一步不是直接开始 merge，而是：

- 先以这份文档为总索引
- 再补一份“功能包回放清单”
- 然后从 `upstream/main` 新开干净集成 worktree
- 按第 7 节顺序逐包融合

这样做的目标不是“最快”，而是“最不容易把所有风险混在一起”。
