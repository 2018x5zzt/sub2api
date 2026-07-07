# Context: xlabapi Fork Glossary

术语表。只记录领域词汇的精确定义，不放实现细节。

## Core / Kernel（内核）
上游 `Wei-Shaw/sub2api`（`upstream` remote）。xlabapi 逐版本从它 roll-up 追平。内核文件默认**不由 xlabapi 主动改写**。

## Roll-up（滚动追平）
在 `rollup/codex-images-3377` 分支上逐版本吸收上游 tag（`chore(kernel): roll up to vX`），再 merge 进 `xlabapi`。push `xlabapi` 即上生产。

## Shell / 套壳（薄壳）
xlabapi 叠加在 Core 之上的定制层。目标形态是"薄壳"：定制收敛到 Core 不碰的**新增文件 / 独立容器**，使 Roll-up 时零冲突、定制不丢。散落在上游文件里的定制是要消除的反模式。

## Backend Deltas（后端定制，保留）
Miku OAuth SSO provider、product-subscription 计费系统、迁移 checksum 兼容规则、CI/deploy workflow、Dockerfile 的 legal 文档拷贝。

## Forced Frontend Set（后端强制配套前端页）
因保留后端定制而**必须存在**的前端页：Subscriptions、Redeem、SubscriptionProducts、ModelHub（现存）、OAuth ConsentView（现丢失，需恢复）。

## Image Studio / 生图站
通过 `ImageStudioView.vue` 的 `<iframe>` 嵌入的独立容器服务。唯一模式：infinite-canvas，由前端注入当前用户的 xlab API key 和网关 URL（`?apiKey=…&baseUrl=…`）。原 ai.mikuapi.org / iframe.mikuapi.org 两个外部依赖均已下掉。

## Shell Invariants（薄壳不变量）
CI 断言守卫的检查清单。分两级：
- **硬不变量**（缺失即阻断部署）：Miku OAuth provider 注册、embed bypass 路由、迁移 checksum 兼容规则、product-subscription 结算链路、Dockerfile legal 拷贝、Dockerfile pnpm 版本锁定。
- **软不变量**（缺失 CI 警告+人工确认）：生图路由、OAuth ConsentView 路由、Subscriptions/Redeem/ModelHub 路由。
