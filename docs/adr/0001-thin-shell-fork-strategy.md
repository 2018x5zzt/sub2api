# ADR 0001 — 薄壳 fork 策略（拒绝双后端套壳）

**状态**：已采纳  
**日期**：2026-07-07

## 背景

xlabapi 是 sub2api 的 fork，在其上叠加了若干定制（Miku OAuth SSO、product-subscription 计费、迁移 checksum 兼容规则等）。随着上游不断迭代，有两种维护路径可选：

1. **厚壳**：把大量前端 / 后端定制散落在上游文件里，每次 Roll-up 人工 merge 冲突。
2. **双后端套壳**：跑一份未改动的上游 backend，前面起一个 xlab 代理层拦截定制逻辑，其余透传。
3. **薄壳**：定制收敛到上游不碰的**新增文件 / 独立容器**；接缝（路由注册、provider 注册等）的改动由 CI 断言守卫盯着。

## 决策

采用**薄壳**策略（路径 3），明确拒绝双后端套壳（路径 2）。

## 拒绝双后端的理由

product-subscription 结算和 Miku OAuth 两个核心定制**不在 HTTP 边界上，而在内核深处**（`RecordUsage` 计费路径、session/auth 存储）。代理层要拦截它们，就必须重放内核的内部流程，最终被迫连 DB schema 和结算逻辑一起分叉。结果：双后端不但没减少同步成本，反而多养一个必须跟着内核漂移的服务，净负担增加。

## 权衡

| | 薄壳（采纳）| 双后端（拒绝）|
|---|---|---|
| Roll-up 冲突 | 少（定制在新增文件）| 多（proxy 层随内核漂移）|
| 运维复杂度 | 低（一个 compose stack）| 高（两套 backend）|
| 定制可控性 | 高（接缝明确，CI 守卫）| 低（边界模糊，容易漏）|
| 迁移成本 | 现有 delta 直接沿用 | 需重写 proxy 层 |

## 结果

- 前端：基本使用上游原版，只加 Forced Frontend Set 和 Image Studio 入口（新增文件）。
- 后端：保留 Backend Deltas，收敛到专用文件（`*_xlab.go`、`migrations_runner.go` 兼容规则段）。
- CI：用 Shell Invariants 断言守卫盯住所有接缝，任一缺失硬阻断部署。
