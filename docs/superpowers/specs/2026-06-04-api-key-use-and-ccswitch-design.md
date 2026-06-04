# API Key 使用密钥与 CC-Switch 导入设计

## 背景

`sub2api-src` 当前位于 `xlabapi` 分支，`frontend-v2` 的用户 API Key 页面已经有新版 React 表格、创建、编辑、删除和用量展示。旧版 Vue 前端已经实现了两个用户常用入口：

- “使用密钥”：按 API Key 所属分组平台展示 Claude Code、Codex、Gemini CLI、OpenCode 等客户端配置片段。
- “导入到 CCS”：生成 `ccswitch://v1/import` deeplink，并按平台选择 `app`、`endpoint`、usage 检测脚本等参数。

本次目标是在新版 `frontend-v2` 的用户 API Key 表格中恢复这两个功能，功能逻辑与旧版一致，UI 保持新版风格。

## 范围

### 包含

- 仅改新版前端用户 API Key 页面相关代码。
- 在每行 API Key 的操作区按方案 A 直接展示入口：`使用密钥`、`导入 CCS`、编辑、删除。
- 端到端保留旧版“使用密钥”的平台分支逻辑：
  - `anthropic`：Claude Code / OpenCode 配置。
  - `openai`：Codex CLI / Codex CLI WebSocket / OpenCode 配置；当分组允许时展示 Claude Code 配置。
  - `gemini`：Gemini CLI / OpenCode 配置。
  - `antigravity`：Claude Code、Gemini CLI、OpenCode 配置，并使用 `/antigravity` 后缀。
- 端到端保留旧版“导入 CCS”的 deeplink 逻辑：
  - `openai` -> `app=codex`，endpoint 为 API base URL。
  - `gemini` -> `app=gemini`，endpoint 为 API base URL。
  - `anthropic` -> `app=claude`，endpoint 为 API base URL。
  - `antigravity` -> 先让用户选 Claude/Gemini；endpoint 为 `${baseUrl}/antigravity`。
  - deeplink 包含 provider name、homepage、endpoint、apiKey、usageScript、usageAutoInterval 等旧版参数。
- 遵守公开设置：
  - `api_base_url` 为空时回退到 `window.location.origin`。
  - `site_name` 为空时回退到 `sub2api`。
  - `hide_ccs_import_button` 为真时隐藏“导入 CCS”。
- 新增/扩展 `frontend-v2` 现有测试，覆盖按钮展示、CCS deeplink、隐藏按钮、antigravity 客户端选择、使用密钥弹窗关键内容。

### 不包含

- 不修改旧版 `frontend`。
- 不修改后端接口和数据库。
- 不改变 API Key 创建、编辑、删除、用量统计逻辑。
- 不引入新的 UI 框架或大型依赖。

## UI 设计

采用已确认的方案 A：在表格右侧操作列直接展示按钮。

桌面端操作顺序：

1. `使用密钥`
2. `导入 CCS`（受 `hide_ccs_import_button` 控制）
3. 编辑图标
4. 删除图标

按钮使用现有 `btn btn-ghost btn-sm` / `Button` 风格，文字入口保持紧凑；编辑和删除维持现有图标按钮。窄屏时允许操作区换行，优先保持功能可见，不新增“更多”菜单。

“使用密钥”弹窗使用新版 `Modal`，尺寸使用 `lg`。内容结构与旧版一致但样式使用新版 token：

- 无分组或无平台时显示警告提示。
- 顶部展示平台说明。
- 平台客户端 tab：Claude Code、Codex CLI、Codex CLI WebSocket、Gemini CLI、OpenCode。
- OS/shell tab：macOS/Linux、Windows CMD、PowerShell 或 OpenAI 的 macOS/Linux、Windows。
- 配置文件块：深色代码块、路径标题、复制按钮。
- 平台 note：蓝色提示块。

Antigravity 的 CCS 导入客户端选择使用小型 `Modal`，列出 Claude Code 和 Gemini CLI 两个选项。

## 数据流与逻辑

`KeysPage` 继续通过 `keysAPI.listKeys` 获取 API Key，使用每行 `ApiKey` 的 `key`、`group.platform`、`group.allow_messages_dispatch` 判断展示内容。

公共设置从 `useAuthStore((s) => s.publicSettings)` 读取。若当前页面挂载时没有公共设置，可触发 `loadPublicSettings()`，避免用户直接进入 `/keys` 时缺少 `api_base_url`、`site_name`、`hide_ccs_import_button`。

建议拆分为聚焦工具与组件：

- `frontend-v2/src/pages/user/keyUsageConfig.ts`：纯函数生成“使用密钥”弹窗所需的 tab、代码文件、平台说明、note。该文件不依赖 React，便于测试。
- `frontend-v2/src/pages/user/UseKeyModal.tsx`：React 弹窗组件，只负责状态、渲染和复制。
- `frontend-v2/src/pages/user/ccswitch.ts`：纯函数生成 CCS deeplink 参数和 antigravity 客户端分支。
- `frontend-v2/src/pages/user/CcsClientSelectModal.tsx`：antigravity 客户端选择弹窗。
- `frontend-v2/src/pages/user/Keys.tsx`：接入行内按钮、弹窗 state、公共设置读取。

## 错误处理

- 复制配置失败时沿用现有 toast 错误提示。
- `window.open(deeplink, '_self')` 抛错时提示 `keys.ccSwitchNotInstalled`。
- 打开 deeplink 后沿用旧版 100ms focus 检测；若页面仍 focus，则提示未安装或未注册协议。
- API Key 无 group/platform 时，“使用密钥”弹窗显示旧版无分组提示；CCS 导入按旧版平台默认值 `anthropic` 处理。

## 测试策略

使用 `frontend-v2` 现有 Vitest + Testing Library 测试风格。

- `Keys.spec.tsx` 新增页面集成测试：
  - 表格行展示 `使用密钥`。
  - `hide_ccs_import_button=false` 时展示 `导入 CCS`。
  - `hide_ccs_import_button=true` 时隐藏 `导入 CCS`。
  - 点击 `导入 CCS` 对 openai key 调用 `window.open`，URL 包含 `ccswitch://v1/import`、`app=codex`、`endpoint`、`apiKey`。
  - 点击 antigravity key 的 `导入 CCS` 先打开客户端选择弹窗，选择 Gemini 后 URL endpoint 包含 `/antigravity` 且 app 为 `gemini`。
  - 点击 `使用密钥` 打开弹窗，并根据 openai key 展示 Codex/OpenCode 相关内容。
- `keyUsageConfig` 和 `ccswitch` 的核心分支可用轻量纯函数单元测试覆盖，降低 UI 测试复杂度。

## 验证

- 运行 `npm exec vitest run src/pages/user/__tests__/Keys.spec.tsx`。
- 运行 `npm run typecheck`。
- 若环境允许，运行 `npm run build`。

## 注意事项

当前仓库已有与本任务无关的后端未提交修改。本任务只应修改 `frontend-v2` 用户 API Key 相关文件和本设计文档，不触碰已有后端变更。
