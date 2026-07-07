# ADR 0002 — CI 断言守卫作为薄壳隔离机制

**状态**：已采纳  
**日期**：2026-07-07

## 背景

薄壳策略下，定制与内核的"接缝"（路由注册、OAuth provider 注册、Dockerfile 特殊指令等）仍必须改上游文件。历史记录显示这些接缝在每次 Roll-up 后**静默丢失**——Miku OAuth 丢过两次、迁移兼容规则丢过、Dockerfile legal 拷贝丢过——每次都直接导致生产宕机。

## 决策

引入 **CI 断言守卫**：将所有薄壳接缝编码为可执行的不变量检查，在每次 push 时运行。任一不变量失败，CI 红、部署被硬阻断。不依赖人工核对清单。

目录约定（`frontend/src/xlab/`、后端 `*_xlab.go`）作为辅助，使定制一眼可辨、断言好写。

## 不变量清单

### 硬不变量（失败即阻断部署）
| # | 检查项 | 历史宕机记录 |
|---|---|---|
| H1 | Miku OAuth provider 已注册（provider registry）| 丢过 2 次 |
| H2 | embed_on.go 对 `/oauth/token`、`/oauth/userinfo`、`/oauth/consent` 的 bypass 存在 | 丢过 1 次 |
| H3 | 迁移 checksum 兼容规则存在（047/063/107/140 等关键条目）| 丢过 2 次 |
| H4 | product-subscription 结算链路在（`productSettlement` 传入 `RecordUsage`）| 丢过 1 次 |
| H5 | Dockerfile 有 `COPY docs/legal/` | 丢过 1 次 |
| H6 | Dockerfile pnpm 版本锁定为 9 | 丢过 1 次 |

### 软不变量（失败 CI 警告 + 必须人工确认后方可合并）
| # | 检查项 |
|---|---|
| S1 | 生图路由 `/image-studio` 存在且指向 `ImageStudioView` |
| S2 | `/oauth/consent` 路由存在且指向 `ConsentView` |
| S3 | `/subscriptions`、`/redeem`、`/model-hub` 路由存在 |

## 实现方式

- 硬不变量：Go 单元测试（`//go:build unit`）+ 前端 Vitest 测试，统一在现有 CI job 里跑。
- 软不变量：独立 CI job，失败只产生 warning annotation，不阻断 merge（但 deploy job 依赖人工 approve）。

## 权衡

| | CI 断言守卫（采纳）| 可重放 apply 脚本 | 人工核对清单 |
|---|---|---|---|
| 防止静默丢失 | ✅ 构建期暴露 | ✅ 但脚本本身可能漂移 | ❌ 依赖纪律 |
| 维护成本 | 低（测试与代码同步）| 高（需跟上游结构变动）| 低但不可靠 |
| 上游文件变动适应性 | 高（测试直接检查符号/文件）| 低（patch 偏移即失效）| 高（但静默）|

## 结果

静默丢失 → 生产宕机 的模式被彻底消除。Roll-up 后若有接缝丢失，CI 第一时间响亮报错，永不上生产。
