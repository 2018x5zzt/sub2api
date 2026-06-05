# Upstream 选择性移植设计

## 背景

`xlabapi` 是当前线上主分支，已经包含企业 BFF、联盟、支付、OpenAI 网关兼容、frontend-v2 等定制改动。本地已成功刷新 `upstream/main` 到 `f1aa5896`（`Merge pull request #2993 from ghostg00/fix/openai-5h-used-percent-direct`）。

`xlabapi` 与 `upstream/main` 从 `48912014` 分叉后差异很大。直接合并 upstream 会引入大量 schema、支付、OAuth、内容审核、渠道监控、模型同步和 gateway 重构，同时 upstream 还删除了部分 `enterprisebff` 文件。这些变化与线上定制冲突风险高，不适合作为单次全量合并。

本轮目标是选择性移植 upstream 中高价值、低到中风险的修复，并保留 `xlabapi` 的线上定制能力。

## 范围

### 第一批：低风险高价值修复

优先移植以下修复，若 cherry-pick 冲突较小则保留原提交语义；若冲突较大则手工移植并补测试。

1. `0f8e2d09 fix(security): 屏蔽 admin 账号接口返回的敏感凭证字段`
   - 防止 admin 账号接口返回敏感 credentials。
   - 保留 `xlabapi` 现有账号字段和 enterprise 权限逻辑。

2. `cf2d5067 fix(security): add JWT auth + visibility check to pages API`
   - 给自定义页面 API 增加 JWT 鉴权和可见性检查。
   - 需要确认公开页面仍能按预期访问。

3. `18790386 fix(deploy): 移除数据库与 Redis 宿主机端口映射`
   - 从生产 `deploy/docker-compose.yml` 移除 PostgreSQL/Redis 宿主机端口暴露。
   - 不修改 local/dev compose 的调试端口。

4. `4d51e53d fix(redeem): 修复批量复制兑换码兼容性`
   - 旧版前端兑换码批量复制兼容性修复。

5. `360f8dec fix: 修复管理后台分组页可用账号数显示错误`
   - 修复旧版前端管理后台分组可用账号数展示。

6. `26ca73a4 fix: hide model scopes for non-antigravity plans`
   - 非 Antigravity 套餐隐藏模型 scope UI。
   - 保留现有订阅/产品展示逻辑。

7. `e46d2c21 fix: avoid ops deep link initialization error`
   - 修复 Ops dashboard deep link 初始化错误。

8. `b0c77233 fix(admin/settings): make tab shell readable in dark mode`
   - 修复旧版设置页 tab shell 深色模式可读性。

9. `44679221 fix: add autocomplete="one-time-code" for TOTP autofill support`
   - 改善 TOTP 输入自动填充体验。

10. `65493df9 fix(ccswitch): add codex model to import deeplink`
    - 旧版前端 CC-Switch 导入 deeplink 增加 Codex 模型信息。
    - 新版 frontend-v2 已有本地实现，本项主要服务旧版 8084 入口。

### 第二批：必要 gateway 修复

优先移植以下三个 gateway/稳定性修复；每个独立处理，避免一次性引入 upstream 大规模 gateway 重构。

1. `679c0865 fix(openai): handle versioned compatible base URLs`
   - 兼容带 `/v1` 等版本路径的 OpenAI compatible base URL。
   - 与 `xlabapi` 的 OpenAI endpoint、image、chat completions 路径生成逻辑合并。

2. `a6117429 fix(gateway): detach upstream context unconditionally for image generation`
   - 图片生成上游请求不应被下游连接取消直接影响。
   - 需要与当前 `xlabapi` 的 image diagnostics / failover 逻辑兼容。

3. `33ac8eb2 fix openai http2 response header timeout`
   - 增加 HTTP/2 response header timeout 配置，改善部分上游卡住时的稳定性。
   - 需要移植配置、默认值、deploy env 示例和 targeted tests。

## 暂不纳入本轮

以下 upstream 功能价值可能很高，但本轮不直接移植，因为它们涉及产品策略、数据库 schema、支付/OAuth 或 gateway 大规模重构：

- Airwallex / 支付系统大改。
- GitHub / Google / DingTalk OAuth 登录大改。
- 内容审核 / risk-control。
- 用户 x 平台 USD 配额。
- 账号 5h/7d 用量阈值自动暂停。
- Channel monitor 大改和模板协议管理。
- 上游模型同步管理端功能。
- 邮件模板编辑器。
- Bedrock Claude Code compatibility。
- Image billing size normalization（涉及 usage log schema 与大量展示）。
- ent schema 大批迁移和删除 enterprise-bff 相关代码。
- Codex Responses bridge 大规模重构与 OOM/WS oversized request 改造。

这些内容需要单独设计、单独迁移、单独验证。

## 集成方式

1. 从 `xlabapi` 创建隔离 worktree。
2. 先尝试 cherry-pick 第一批低风险提交；遇到冲突时手工移植最小必要 diff。
3. 第一批通过测试后提交一个聚合 commit，避免与后续 gateway 修复混在一起。
4. 对第二批三个 gateway 修复逐个移植、逐个测试；每个修复单独提交。
5. 不接受 upstream 对 `enterprisebff` 的删除，也不接受会回退 `xlabapi` frontend-v2 / affiliate / subscription 定制的改动。
6. 全部通过后合并回 `xlabapi`，再 push `origin xlabapi`。

## 验证策略

每个批次至少运行：

- 相关 Go targeted tests：账号 DTO 脱敏、pages API、OpenAI endpoint URL、OpenAI images、HTTP upstream/profile。
- `go test` 的最小相关包集合：`./internal/handler/...`、`./internal/service/...`、`./internal/repository/...` 中受影响包。
- 前端相关测试：旧版 frontend 受影响组件/工具测试；若改到 frontend-v2，则跑对应 Vitest。
- `frontend-v2` 的 `npm run typecheck` 与 `npm run build`，确保不会破坏当前线上新版 UI。
- 合并回 `xlabapi` 后重复关键验证。

如果某个 upstream commit 冲突过大或需要引入大量 schema/产品逻辑依赖，则暂停该项并记录为后续专项，不在本轮强行合并。

## 推送与上线

本轮完成后目标是：

1. 在本地 `xlabapi` 上产生清晰提交。
2. `git push origin xlabapi`。
3. 若网络可达，运行 `deploy.sh` 上线并检查容器状态和 `/health`。

如果网络再次不可达，保留本地提交并输出可手工执行的 push/deploy 命令。
