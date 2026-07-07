# ADR 0003 — infinite-canvas 取代全部生图外部依赖

**状态**：已采纳  
**日期**：2026-07-07

## 背景

原 ImageStudioView 有两个模式：
- **新版**：iframe 嵌入 `ai.mikuapi.org`，走 Miku OAuth SSO 免登。
- **旧版**：iframe 嵌入 `iframe.mikuapi.org`（魔改的 gpt-image-playground）。

两者均依赖**不受 xlabapi 控制的外部站点**。`ai.mikuapi.org` 曾多次因 OAuth 流程变动导致生图功能中断；魔改 gpt-image-playground 有持续维护负担。

## 决策

用 **infinite-canvas**（`github.com/basketikun/infinite-canvas`）取代上述两个外部依赖，作为**唯一**生图入口。

### 集成方案

1. infinite-canvas 以**独立 Docker 容器**部署在同一台 baota 服务器，Caddy 反代到 `canvas.xlabapi.com` 子域名。
2. ImageStudioView 变为**单模式**：iframe src = `https://canvas.xlabapi.com/?apiKey=<当前用户 xlab key>&baseUrl=<xlab 网关>`（infinite-canvas 原生支持的自动配置 URL 约定）。
3. 前端从已登录用户会话取 API key，**无需新增后端端点**。
4. 暂不签发专用 image-scoped 子 key（可未来按需加）。

## 权衡

| | infinite-canvas（采纳）| 保留 ai.mikuapi.org 新版 | 保留魔改 gpt-image-playground |
|---|---|---|---|
| 外部依赖 | 无（自托管）| 依赖 mikuapi.org 可用性 | 依赖 iframe.mikuapi.org |
| 鉴权 | 注入用户 xlab key，天然按用户隔离 | Miku OAuth SSO（多次出故障）| 无鉴权 |
| 维护 | 跟 infinite-canvas 上游，可 fork | 跟 mikuapi 外部站变动 | 需持续魔改 |
| 计费 | 直连 xlab 网关，复用 product-subscription | 不经 xlab 网关 | 不经 xlab 网关 |
| 改源义务 | AGPL-3.0（不改源则仅需保留声明）| 无 | 有（已魔改）|

## 约束

- infinite-canvas 是 AGPL-3.0。**只要不修改其源码**，以网络服务形式部署仅需保留原作者信息和前端页面标识，无需公开 xlabapi 自身代码。若将来需要魔改，需评估 AGPL 开源义务。
- README 警告"不建议直接公网多人共用"——此风险由**每用户注入独立 key** 的方案化解：服务本身无共享凭据。

## 结果

生图功能完全自主可控，消除对 mikuapi.org 外部站的运行时依赖，计费天然走 xlab product-subscription 体系。
