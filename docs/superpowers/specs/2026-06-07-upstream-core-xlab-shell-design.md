# Upstream Core + Xlab Shell 架构设计

## 背景

`xlabapi` 目前是线上主分支，但它已经从 upstream `sub2api` 演化成一个大 fork：既包含 upstream core 网关能力，也包含 xlab 自己的产品订阅、支付、联盟、frontend-v2、企业 BFF、OpenAI 兼容和部署逻辑。随着 upstream 在 `v0.1.122` 之后快速迭代，直接合并或直接替换为 upstream 最新容器会持续产生高风险冲突，尤其是数据库 schema、订阅/支付、quota、gateway 和前端接口。

新的长期目标是把 `sub2api` 重新收敛为尽量贴近 upstream 的 core，把 xlab 的业务能力从 core fork 中拆出来，形成独立的 xlab backend 和统一的 frontend-v2 shell。这样 upstream core 可以按版本持续升级，xlab 业务功能通过外置服务和适配层演进，不再和 upstream core 强耦合。

当前已经完成但尚未推送的相关本地提交包括 API Key 使用密钥/CCSwitch、upstream 选择性移植设计与版本段审计。已有的选择性移植 worktree `/root/.config/superpowers/worktrees/sub2api-src/upstream-selective-xlabapi-20260605` 只作为参考/可摘取补丁来源，不作为新的主线合并路径。

## 目标

1. **Core 回归 upstream**：`sub2api` core 尽量不再承载 xlab 专有业务逻辑，只保留必要的兼容配置和少量适配点。
2. **业务外置**：产品订阅、支付、联盟、兑换码商品化、运营策略放入独立 xlab backend。
3. **前端套壳**：frontend-v2 作为统一 UI shell，同时调用 upstream core API 和 xlab backend API。
4. **可持续升级**：未来按 upstream release/tag 升级 core，而不是长期维护大规模 fork。
5. **保护线上用户**：已有产品订阅用户、支付订单、联盟记录和 API Key 授权必须有明确迁移和回滚方案。

## 非目标

- 本设计不要求立即删除现有 `xlabapi` core 中的所有定制代码。
- 本设计不直接执行 upstream 全量 merge 或替换线上容器。
- 本设计不在本轮实现 Airwallex、内容审核、用户 x 平台 quota、channel monitor 等 upstream 大功能。
- 本设计不改变当前线上数据库，迁移会在后续实施计划中分阶段执行。

## 目标架构

```text
                         ┌─────────────────────────┐
                         │       frontend-v2        │
                         │   unified xlab shell     │
                         └───────────┬─────────────┘
                                     │
                    ┌────────────────┴────────────────┐
                    │                                 │
             /api/v1│                                 │/xapi/v1
                    ▼                                 ▼
        ┌──────────────────────┐          ┌──────────────────────┐
        │ upstream sub2api core │          │     xlab backend      │
        │ gateway/accounts/keys │◀────────▶│ products/payments/etc │
        └──────────┬───────────┘          └──────────┬───────────┘
                   │                                  │
                   ▼                                  ▼
        ┌──────────────────────┐          ┌──────────────────────┐
        │     core database     │          │      xlab database    │
        │ upstream migrations   │          │ xlab business schema  │
        └──────────────────────┘          └──────────────────────┘
```

### Upstream core 职责

- AI gateway：OpenAI/Gemini/Anthropic/Antigravity/compatible upstream 转发。
- 上游账号、分组、API Key、基础用户、基础用量日志。
- upstream 自带的 admin/user 基础功能。
- upstream 自己的数据库 schema 和 migrations。

Core 不再新增以下 xlab 专有业务：产品订阅、支付订单、联盟结算、商品化兑换码、企业 BFF、xlab 专属前端页面。

### Xlab backend 职责

- 产品订阅和商品订阅。
- 支付订单、支付回调、订单履约。
- 联盟/邀请/返利/结算。
- 兑换码商品化、批量发放和权益映射。
- xlab 运营后台 API。
- 将业务权益同步或映射到 core 能理解的用户、分组、API Key、余额、quota、rate-limit 等原语。

### Frontend-v2 shell 职责

- 保持 xlab 的新版 UI 和交互风格。
- 使用 `src/api/core/*` 调 upstream core。
- 使用 `src/api/xlab/*` 调 xlab backend。
- 对用户隐藏双后端细节，页面仍表现为一个系统。

## API 与路由边界

建议保留清晰路由前缀：

- `/api/v1/**`：upstream core API。
- `/xapi/v1/**`：xlab backend API。
- `/admin/**`、`/dashboard/**` 等前端路由继续由 frontend-v2 控制。

反向代理或部署入口可以按路径转发：

```text
/api/v1      -> sub2api core
/v1          -> sub2api core gateway
/xapi/v1     -> xlab backend
/*           -> frontend-v2 static shell
```

如果短期仍使用 core embed 前端，则可先让 core 服务 frontend-v2，但中长期建议 frontend-v2 独立构建和部署，避免 core 容器升级时带走 xlab UI。

## 认证设计

### 初期推荐方案：core 作为认证源

1. 用户登录仍走 upstream core。
2. frontend-v2 保存 core JWT。
3. 调 xlab backend 时携带同一个 JWT。
4. xlab backend 校验 JWT：
   - 本地验证 core JWT 签名；或
   - 调 core `/api/v1/user/me` / current-user API 验证 token；或
   - 通过共享 JWKS/secret 验证。
5. xlab backend 内部保存 `core_user_id` 作为用户外键。

优点是迁移成本最低，不需要立刻重写登录/OAuth/2FA。缺点是 xlab backend 依赖 core auth 可用性。

### 中长期可选方案：xlab 作为统一身份层

后续如果产品订阅和支付完全外置，可以考虑 xlab backend 做统一身份层，再通过服务账号调用 core。但这会显著增加迁移复杂度，不建议第一阶段做。

## 数据归属

### Core database

由 upstream migrations 管理，包含：

- users / auth identities（初期仍由 core auth 管理）
- groups / accounts / API keys
- core usage logs
- upstream 自带 settings
- upstream 自带支付/订阅表（如果 upstream core 需要），但 xlab 业务不依赖其语义

### Xlab database

由 xlab backend migrations 管理，包含：

- subscription products
- product subscriptions
- payment orders / payment callbacks / provider snapshots
- affiliate invitations / ledgers / reward records
- redeem product grants
- xlab-specific billing and settlement records
- core mapping 表，例如 `core_user_id`、`core_group_id`、`core_api_key_id`

### 关键原则

- 新的 xlab 业务表不再进入 core migrations。
- core migrations 可以随 upstream 更新。
- xlab backend 通过稳定 API 或同步任务把权益投影到 core。

## 产品订阅与 API 调用授权

产品订阅是本架构的核心风险点。短期推荐采用 **权益同步模式**，避免改 core gateway 热路径。

### 权益同步模式

xlab backend 维护产品订阅真相，并把结果同步到 core：

- 产品订阅激活后，为用户授予 core group 可见性或绑定默认 group。
- 根据产品配置更新 core API Key quota / rate limit / budget multiplier。
- 到期、取消或额度耗尽时，撤销或调整 core 权益。
- usage log 仍由 core 记录，xlab backend 周期性拉取或监听后做二次结算。

优点：core 尽量不改，upstream 升级更容易。缺点：需要保证同步一致性和补偿任务。

### 不推荐的短期方案：gateway 前置拦截

在 core gateway 前再包一层 xlab proxy 可以做到实时产品订阅校验，但会引入热路径延迟、故障面和更多协议兼容问题。除非权益同步无法满足产品需求，否则不作为第一阶段方案。

## 迁移现有线上数据

需要单独实施迁移工具，流程如下：

1. 从现有 `xlabapi` core DB 导出产品订阅相关表。
2. 写入 xlab backend DB。
3. 生成 `core_user_id`、`core_group_id`、`core_api_key_id` 映射。
4. 对每个 active product subscription 计算权益投影。
5. 同步到 core 的用户/group/api key/quota/rate-limit。
6. 对账：
   - active subscription 数量一致。
   - 到期时间一致。
   - 已用额度和剩余额度一致。
   - payment order 与 fulfillment 状态一致。
   - redeem code grant 映射一致。
7. 迁移后保留只读旧表或备份，直到至少一个完整账期后再决定清理。

## Upstream core 升级策略

升级 core 时按 release tag 逐段推进：

1. 建立 core upgrade worktree。
2. 合并一个 upstream tag range。
3. 不接纳 xlab 业务逻辑回流进 core。
4. 跑 core 测试：gateway、accounts、groups、API keys、usage、migrations。
5. 跑 xlab adapter 测试：frontend-v2 调 core API 是否兼容。
6. 在 staging 使用复制数据库跑 migration。
7. 验证 xlab backend 权益同步和产品订阅链路。
8. 通过后再上线。

## 当前 selective worktree 的处理

`upstream-selective-xlabapi-20260605` 中已有部分低风险补丁，可作为后续迁移参考：

- credential redaction 可考虑保留到 core。
- custom page visibility 可视 core 是否仍提供 pages API 决定。
- legacy UI/deploy/TOTP/CCSwitch 对未来独立 frontend-v2 价值有限，可按需丢弃或手工迁移。
- versioned compatible base URL 是 core gateway 修复，可在逐版本升级中自然获得或手工保留。

短期不要把该 worktree 直接合入生产 `xlabapi`，避免和新的逐版本/拆分架构路线混杂。

## 分阶段路线

### Phase 0：冻结 core 新业务

- 不再把新的产品订阅、支付、联盟、兑换码业务写入 core。
- 只允许安全修复、gateway 稳定性修复、升级适配进入 core。

### Phase 1：定义 xlab backend 最小边界

- 新建 xlab backend 项目或服务目录。
- 实现 core JWT 校验。
- 实现 product subscription 只读 API。
- frontend-v2 增加 `/xapi/v1` API client。

### Phase 2：迁移产品订阅只读视图

- 把现有产品订阅数据同步到 xlab DB。
- frontend-v2 订阅页面从 xlab backend 读取。
- core 仍负责实际 API key 授权。

### Phase 3：迁移支付与履约

- 新订单进入 xlab backend。
- xlab backend 完成支付回调和权益同步。
- core 不再承载新的产品订单履约。

### Phase 4：迁移联盟与兑换码商品化

- affiliate / redeem product grants 移到 xlab backend。
- core 只保留基础 redeem 或不再使用 upstream redeem。

### Phase 5：core 逐版本收敛 upstream

- 从 `v0.1.122` 开始按 tag 升级。
- 每个版本段在 staging 验证。
- 对 upstream 支付/订阅功能只保留 core 需要部分，不让其替代 xlab 业务真相。

## 风险与缓解

| 风险 | 缓解 |
|---|---|
| 迁移后用户权益不同步 | 使用同步任务、幂等投影、差异对账和补偿任务。 |
| core migration 影响 xlab 业务 | 拆库；xlab 业务表不再由 core migrations 管理。 |
| frontend-v2 API 适配成本高 | 引入 `core` / `xlab` API client 分层和 adapter tests。 |
| 支付回调迁移导致订单丢失 | 双写/灰度期；旧 core 回调保持只读/兼容转发。 |
| 回滚困难 | core 与 xlab backend 独立部署；迁移前备份 DB；每阶段设置回滚点。 |

## 下一步

1. 为 Phase 1 写实施计划：xlab backend 最小身份校验 + 产品订阅只读 API + frontend-v2 adapter。
2. 为现有产品订阅表做字段盘点和迁移映射。
3. 建立产品订阅回归测试：active subscription、quota decrement、payment fulfillment、redeem grant。
4. 暂停直接合并 upstream 最新容器，直到 xlab 业务从 core 中拆出关键路径。
