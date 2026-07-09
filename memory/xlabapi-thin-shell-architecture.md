---
name: xlabapi-thin-shell-architecture
description: 2026-07-07 grilling 会话确定的 xlabapi 薄壳架构设计决策全记录
metadata:
  type: project
---

## 核心决策（全部锁定）

**薄壳策略**：定制收敛到新增文件/独立容器，不散落在上游文件里。拒绝双后端套壳（product-subscription 和 Miku OAuth 在内核深处，proxy 隔离不了）。见 ADR 0001。

**隔离机制**：CI 断言守卫（C）+ 目录约定（A）。9 条不变量分两级，硬阻断 H1-H6，软警告 S1-S3。见 ADR 0002。

**生图站**：infinite-canvas 单模式，取代 ai.mikuapi.org（Miku SSO）和 iframe.mikuapi.org（魔改 playground）两个外部依赖。见 ADR 0003。

## 关键参数

- 生产 API 网关：`api.xlabapi.com`（支持最长 2min 超时）
- canvas 子域名：`canvas.xlabapi.com`
- iframe 注入 URL：`https://canvas.xlabapi.com/?apiKey=<user xlab key>&baseUrl=https://api.xlabapi.com`
- infinite-canvas 部署：同一台 baota 服务器，独立 docker compose service，Caddy 反代

## 当前待实现项（按优先级）

1. **恢复 Forced Frontend Set 缺失页**：`oauth/ConsentView.vue` 丢失，需从 snapshot 恢复并补路由
2. **CI 断言守卫**：H1-H6（Go unit test + Vitest）、S1-S3（独立 CI job）
3. **ImageStudioView 重写**：单模式，注入用户 key + api.xlabapi.com，替换旧双模式
4. **infinite-canvas 容器化**：deploy/docker-compose.yml 加 service，Caddy 加 canvas.xlabapi.com 反代
5. **前端目录约定**：xlab 专属新增文件移入 `frontend/src/xlab/`

**Why**：历史上每次 roll-up 后定制静默丢失导致生产宕机，薄壳+CI守卫从根本上消除这个模式。

**How to apply**：实现时先做 CI 断言（先有守卫再动接缝），再恢复 ConsentView，最后做 infinite-canvas 集成。[[xlabapi-upstream-sync-gotchas]]
