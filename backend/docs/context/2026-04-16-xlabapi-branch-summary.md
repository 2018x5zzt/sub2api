# xlabapi 分支变更总结（2026-04-16）

## 1. 文档目的

这份文档用于在后续对话中快速恢复 `sub2api` 主源码仓库当前集成分支的上下文，避免每次重新从提交历史反推。

本文档总结的对象是：

- 仓库：`/root/sub2api-src`
- 当前分支：`xlabapi`
- 对比基线：`main`
- 对比范围：`main..xlabapi`

不包含的范围：

- `/root/sub2api-deploy` 里的部署快照、回滚 compose 文件等独立部署仓库变更
- 其他试验分支或归档分支的未并入内容

## 2. 当前分支快照

- `HEAD`: `b0993dda` (`fix(openai): remap legacy chatgpt oauth codex models`)
- 工作区状态：干净
- 相对 `main` 的提交增量：`269` 个提交
- 代码规模变化：`461 files changed, 68682 insertions(+), 6806 deletions(-)`
- 当前描述版本：`v0.1.106-146-gb0993dda`

这说明 `xlabapi` 不是一个小修小补分支，而是一个已经吸收多轮功能、修复、迁移和发布准备工作的集成分支。

## 3. 这条分支在做什么

`xlabapi` 的核心目标不是单点功能开发，而是把一批已经通过验证的能力整合成一个可部署的主线候选版本，同时保留邀请码、注册、兑换、OpenAI/OAuth 网关兼容、企业 BFF、运维观测等已经存在的行为面。

仓库内现有设计文档显示，这条分支曾按固定顺序吸收下列能力分支：

1. `api-key-quota-help-url-plus-oauth-itemrefs-20260412-084117`
2. `dynamic-group-budget-multiplier`
3. `enterprise-visible-groups`
4. `openai-oauth-store-false-itemrefs`
5. `smtp-outlook-starttls-20260412-151123`
6. `upgrade-v0.1.106-merge`

同时它保留并继续扩展了本地已有的邀请体系、注册路径、兑换逻辑和运维修复。

## 4. 相对 main 的主要新增能力与改动

### 4.1 邀请体系、注册链路和兑换体验

这是 `xlabapi` 最重的一组业务改造，已经不是单点 patch，而是一套完整的邀请增长体系。

已落地内容包括：

- 正式建立 invite growth foundation，新增邀请关系、奖励记录、管理动作等完整数据模型
- 退役旧的邀请码流程，逐步收敛到新的邀请体系
- 注册流程移除 promo code 主入口，改为与 invite 语义对齐
- 用户侧新增邀请中心页，支持查看邀请信息与奖励说明
- 管理后台新增邀请管理页和相关 API
- 新增邀请后台 rollout verification SQL，用于核验上线后的邀请绑定、奖励结算和数据一致性
- 邀请码轮换改成字母别名兼容模式，降低历史码切换带来的兼容问题
- 邀请奖励文案、基线比例、结算原子性和重复奖励预检查都做了收敛
- 幸运红包榜单功能加入，并且后续被限制到兑换结果弹窗场景中显示，避免入口泛滥
- 兑换页和前端错误展示做了结构化保留，避免重复兑换或权益耗尽时无法继续查看榜单

可以把这部分理解为：

- 注册和邀请已经不再是边缘功能，而是 `xlabapi` 的核心业务面之一
- 管理侧、用户侧、数据库侧和 SQL 验证侧都已经补齐

### 4.2 OpenAI、Anthropic 与多协议网关兼容层

这部分是第二大主线，重点在于把多上游、多协议兼容做稳，并把新旧接口行为统一。

已落地内容包括：

- 新增 `Responses <-> Anthropic` 双向格式转换
- 网关显式新增 `/v1/responses` 与 `/v1/chat/completions` 的平台化分流
- `GatewayService` 新增 `ForwardAsResponses` 和 `ForwardAsChatCompletions`
- 补齐 OpenAI、Anthropic、Gemini 兼容过程中的模型、reasoning、usage 字段保留逻辑
- 修复 OpenAI 到 Anthropic 转换路径中 system prompt 被静默丢弃的问题
- 修复 Anthropic 空响应、capacity-limit 400、thinking 空 200、terminal stream 缺失等异常场景
- 修复 replay/buffered response 文本重复、delta-only buffered response 聚合等流式转发问题
- 增加 OpenAI `429` silent failover，以及部分上游异常归一化
- 修复 passthrough 模式下的 429 限流持久化与连接隔离问题
- 追加 legacy chatgpt oauth codex 模型 remap
- 增加 `gpt-5.4-mini`、`gpt-5.4-nano` 支持与定价，并修复 `gpt-5.4-xhigh` 兼容映射
- 保留 requested model，并将其贯穿 usage log、账单和后台 usage 展示，避免只看到映射后的上游模型

这条主线的结果是：

- 网关更像一个真正的多协议、多模型兼容层
- 后台看到的 usage / error / model 数据与用户原始请求更一致
- OpenAI OAuth 和兼容路由的边缘失败情况被压低

### 4.3 OpenAI OAuth、账号调度、隐私和 TLS 指纹能力

这部分是账号基础设施层面的增强，和实际可用性直接相关。

已落地内容包括：

- 新增 OpenAI OAuth 手动输入 `Mobile RT` 入口
- `Mobile RT` 流程补全 `plan_type`、精确匹配账号，并在刷新时自动设置隐私
- 为 OpenAI OAuth 账号新增前端手动设置隐私按钮
- 创建和批量创建 OpenAI OAuth 账号时异步设置隐私
- 刷新令牌失败时仍尝试设置隐私模式
- antigravity 路径支持自动设置隐私和后台手动重试
- 从 `LoadCodeAssist` 复用 `TierInfo` 提取 `plan_type`
- 新增 TLS 指纹 Profile 数据库管理能力与对应后台管理接口
- 支持 OpenAI OAuth WS mode 的批量编辑
- 支持 OpenAI passthrough 开关的批量编辑
- 新增 `user:file_upload` OAuth scope
- Anthropic `oauth/setup-token` 账号支持自定义转发 URL
- 账号调度和生命周期相关代码补了多个稳定性修复，比如 stop 重复 close、防临时 unsched 状态残留等

这部分意味着：

- OAuth 账号的运维手段更完整
- 账号创建、刷新、隐私设置、WS 转发、TLS 指纹控制都进入了后台可操作范围

### 4.4 企业 BFF、分组可见性和预算倍率

`xlabapi` 已经不仅仅面向原始 Web 控制台，还新增了企业前台所需的一层兼容 BFF。

已落地内容包括：

- 新增 `enterprise-bff` 可执行入口和独立部署 compose
- 企业登录、企业用户请求代理、key 写入授权、pool status 可见性过滤等后端逻辑
- 企业分组可见性从“全部可见”收紧为“按配置映射可见”
- 企业用户和管理员对 key 写入时的 group 绑定做授权限制
- 保留 public groups 在企业可见性里的兼容处理
- 新增 group health snapshot 持久化和后台定时采集
- 新增 dynamic group budget multiplier / dynamic pricing 逻辑
- 前端新增 group budget multiplier 控件和相关工具函数
- 明确禁止 enterprise key 的 group rebinding

可以把它理解为：

- 企业前台开始和原始 core UI 解耦
- 企业用户只看到自己应该看到的组和池信息
- group 维度预算和健康度开始具备独立控制面

### 4.5 运维观测、后台可用性和管理员体验

这条分支对 `ops` 和后台表单做了很多密集修复，虽然不少是 `fix`，但整体上已经构成了一个运维面升级。

已落地内容包括：

- ops error log 新增 `endpoint`、`model`、`request_type`、`upstream_url` 等上下文字段
- 后台 ops 表格和详情面板新增这些字段展示
- ops 错误展示优先使用 `upstream_model`
- 修复 runtime log 控件在前端溢出的问题
- 修复后台设置页静默保存失败问题
- 修复 SMTP 配置被覆盖、刷新不稳定的问题
- 新增 Outlook `STARTTLS` SMTP 提交支持
- exhausted API key quota 时附加 help URL
- 自定义 endpoint 的配置与展示打通到前端
- Key 使用页、Endpoint popover、Model Hub 等前端页面补了说明与展示
- README 增加日文版入口，并新增完整 `README_JA.md`

从运维视角看，这部分的价值在于：

- 出错时能更快定位“哪个 endpoint、哪个模型、哪类请求、打到哪个上游”
- 管理后台不再只是 CRUD，而是更适合日常运维

### 4.6 模型、计费和用量语义

计费和 usage 这部分做了若干关键矫正，避免用户请求模型和上游计费模型混淆。

已落地内容包括：

- usage log 新增 `requested_model` 字段及索引
- 账单记录、usage 查询、DTO 映射、前端 usage 表格统一保留 requested model
- 修复 billing 始终使用映射后上游模型的问题，改为使用用户原始请求模型计费
- Model Hub 新增 effective model pricing
- quota reset 后累计用量显示陈旧的问题已修复

结果是：

- 计费解释性增强
- 用户、管理员和运维看到的“请求模型”含义更一致

## 5. 关键数据库和迁移变化

相对 `main`，这条分支引入了多组重要 schema / migration 变化。

新增或显著扩展的核心表结构包括：

- `group_health_snapshot`
- `invite_admin_action`
- `invite_relationship_event`
- `invite_reward_record`
- `tls_fingerprint_profile`
- `usage_log.requested_model`

关键迁移文件包括：

- `079_ops_error_logs_add_endpoint_fields.sql`
- `080_add_group_health_snapshots.sql`
- `080_create_tls_fingerprint_profiles.sql`
- `081_add_dynamic_group_budget_multiplier.sql`
- `081_add_invite_growth_foundation.sql`
- `082_add_invite_admin_ops.sql`
- `083_add_tls_fingerprint_profiles.sql`
- `084_add_usage_log_requested_model.sql`
- `085_add_usage_log_requested_model_index_notx.sql`
- `086_rotate_invite_codes_to_letters.sql`

注意：

- 这里存在同号迁移文件分别用于不同功能的情况，后续如果做迁移清理或重建，需要特别小心顺序与执行环境

## 6. 代码热区

从 `main...xlabapi` 的目录分布看，改动最集中的区域是：

- `backend/internal/service/`：约 `30.8%`
- `backend/ent/`：约 `9.5%`
- `backend/internal/repository/`：约 `6.9%`
- `backend/internal/handler/`：约 `5.4%`
- `backend/internal/enterprisebff/`：约 `3.6%`
- `frontend` 多个视图、API、组件和测试文件：合计占比较高

这说明：

- 大部分变化都不是 UI 表层 patch，而是 service、repository、schema、handler 的行为级修改
- 后续任何回归排查，都应优先看服务层和 ent schema，而不是只盯前端

## 7. 仓库里已经存在的关键设计文档

如果下次对话需要继续恢复上下文，优先读这些文档：

- `/root/sub2api-src/backend/docs/context/2026-04-16-xlabapi-branch-summary.md`
- `/root/sub2api-src/docs/superpowers/specs/2026-04-13-xlabapi-final-release-integration-design.md`
- `/root/sub2api-src/docs/superpowers/plans/2026-04-13-xlabapi-final-release-integration.md`
- `/root/sub2api-src/docs/superpowers/specs/2026-04-11-invite-growth-design.md`
- `/root/sub2api-src/docs/superpowers/specs/2026-04-11-invite-foundation-completion-design.md`
- `/root/sub2api-src/docs/superpowers/specs/2026-04-11-dynamic-group-budget-multiplier-design.md`
- `/root/sub2api-src/docs/superpowers/specs/2026-04-11-enterprise-visible-groups-design.md`
- `/root/sub2api-src/docs/superpowers/specs/2026-04-11-openai-oauth-expired-recovery-design.md`
- `/root/sub2api-src/docs/enterprise-bff.md`

## 8. 下次对话的推荐恢复提示词

如果需要快速接续，可以直接基于这份文档给出类似提示：

`先读 /root/sub2api-src/backend/docs/context/2026-04-16-xlabapi-branch-summary.md，并把它作为当前 sub2api 主源码仓库上下文。仓库是 /root/sub2api-src，分支是 xlabapi。后续讨论默认以 main..xlabapi 的增量为背景。`

## 9. 恢复上下文时建议先确认的命令

后续重新开对话后，建议先执行：

```bash
git -C /root/sub2api-src status --short --branch
git -C /root/sub2api-src log --oneline --decorate -n 20
git -C /root/sub2api-src diff --stat main...xlabapi
```

目的很简单：

- 确认当前是否仍在 `xlabapi`
- 确认这份文档写完后有没有新的提交
- 确认是否出现了新的大范围改动需要补充到文档

## 10. 一句话结论

`xlabapi` 已经演化成 `sub2api` 的一个重度集成分支：它把邀请增长体系、OpenAI/OAuth 账号基础设施、多协议网关兼容、企业 BFF、动态分组预算、运维观测增强和一整套数据库迁移整合到了一起，远超“单个功能分支”的复杂度。
