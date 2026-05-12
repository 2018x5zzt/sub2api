# frontend-v2 Parity Placeholder Inventory

- 审计目标：列出 `frontend-v2`（origin/test/xlabapi）中所有 `ParityPlaceholder` 调用点，按 `file:line` + props 全量盘点，作为 P1（占位符减少）整改的输入数据。
- 审计范围：`origin/test/xlabapi:frontend-v2/`
- 关联组件：`src/pages/ParityPlaceholder.tsx`（默认导出，唯一占位符渲染组件）
- 旁路检索：`PlaceholderPage`（`src/pages/admin/Placeholder.tsx`）— **见 §4 死代码**

---

## 1. 调用点清单（总计 4 处）

所有调用点都集中在路由表 `src/router/index.tsx`，没有任何业务页面把 `ParityPlaceholder` 当通用占位组件来用。

| # | path | 路由层 | file:line | title | legacyPath | endpoints | actions | standalone |
|---|---|---|---|---|---|---|---|---|
| 1 | `/setup` | Public（顶层路由） | `src/router/index.tsx:58` | `Setup Wizard` | `/setup` | `GET /setup/status` | `[{ label: 'Login', to: '/login' }]` | `true` |
| 2 | `/key-usage` | Public（顶层路由） | `src/router/index.tsx:59` | `API Key Usage` | `/key-usage` | `GET /usage/dashboard/stats`, `GET /usage/dashboard/models` | `[{ label: 'Console', to: '/dashboard' }]` | `true` |
| 3 | `/custom/:id` | RequireAuth + ConsoleLayout | `src/router/index.tsx:94` | `Custom Page` | `/custom/:id` | `GET /settings/public` | — | `false`（默认） |
| 4 | `/admin/subscription-product-config` | RequireAdmin + ConsoleLayout(admin) | `src/router/index.tsx:113` | `Subscription Product Config` | `/admin/subscription-product-config` | `GET /admin/subscription-products`, `GET /admin/product-subscriptions` | — | `false`（默认） |

---

## 2. 调用点逐行 props 详情

### 1) `/setup` — Setup Wizard（standalone）
```
src/router/index.tsx:58
{ path: '/setup', element: <ParityPlaceholder
    standalone
    title="Setup Wizard"
    legacyPath="/setup"
    endpoints={['GET /setup/status']}
    actions={[{ label: 'Login', to: '/login' }]}
/> }
```
- **守卫**：无（公开页面）
- **standalone=true** → 渲染独立的 `<main className="min-h-screen bg-bg-0 p-6 ... sm:p-10">`，不走 ConsoleLayout
- **目标实现**：MIGRATION_TODO.md 列为 `Setup wizard / SetupWizardView`（first-run DB/Redis/admin config）
- **后端口**：`GET /setup/status`（仅 1 个端点声明，老前端 setup 流程实际走更多端点，需在迁移时核对）

### 2) `/key-usage` — Public per-key usage（standalone）
```
src/router/index.tsx:59
{ path: '/key-usage', element: <ParityPlaceholder
    standalone
    title="API Key Usage"
    legacyPath="/key-usage"
    endpoints={['GET /usage/dashboard/stats', 'GET /usage/dashboard/models']}
    actions={[{ label: 'Console', to: '/dashboard' }]}
/> }
```
- **守卫**：无（公开页面，老前端原本是按 key 公开访问的统计页）
- **standalone=true** → 同上
- **目标实现**：MIGRATION_TODO.md 列为 `KeyUsageView (public per-key usage page)`
- **关键风险**：这条路由是公共展示页，老前端有完整图表 + 模型分组；恢复时不能丢失公开访问语义

### 3) `/custom/:id` — Custom Page（embedded in ConsoleLayout）
```
src/router/index.tsx:94
{ path: '/custom/:id', element: <ParityPlaceholder
    title="Custom Page"
    legacyPath="/custom/:id"
    endpoints={['GET /settings/public']}
/> }
```
- **守卫**：`RequireAuth` + 嵌入 ConsoleLayout
- **standalone=false** → 走 ConsoleLayout 主区，渲染 PageHeader + 单 Card
- **目标实现**：MIGRATION_TODO.md 列为 `CustomPageView (admin-defined custom menu items)`
- **数据源**：管理员通过 `GET /settings/public` 配置 `customMenu`，按 `:id` 渲染对应自定义页面（说明：当前 endpoints 只声明了 settings 入口，真正的 customMenu 项渲染端点需迁移时补全）

### 4) `/admin/subscription-product-config` — Subscription Product Config（admin）
```
src/router/index.tsx:113
{ path: '/admin/subscription-product-config', element: <ParityPlaceholder
    title="Subscription Product Config"
    legacyPath="/admin/subscription-product-config"
    endpoints={['GET /admin/subscription-products', 'GET /admin/product-subscriptions']}
/> }
```
- **守卫**：`RequireAdmin` + 嵌入 ConsoleLayout(admin)
- **standalone=false**
- **目标实现**：在 `MIGRATION_TODO.md` 中**未单独列项**——这是 admin 订阅子系统的"产品配置"分页（区别于已迁移的 `/admin/subscriptions` 列表页）
- **关联导航**：ConsoleLayout.tsx:71 adminNav 中有 `{ to: '/admin/subscription-product-config', labelKey: 'nav.subscriptionProductConfig', Icon: CreditCard }`——侧栏入口已存在，所以 placeholder 不能直接删，必须有页面或重定向
- **同侧路由**：第 112 行 `/admin/subscription-products` 走 `<Navigate to="/admin/subscriptions" replace />`，存在两个相邻 path 但语义不同（products = 订阅产品定义，subscriptions = 订阅订单/授权）

---

## 3. ParityPlaceholder 组件契约（供整改时核对，不动代码）

`src/pages/ParityPlaceholder.tsx`：

```ts
interface ParityPlaceholderProps {
  title: string                                                 // 必填，PageHeader 标题
  description?: string                                          // 选填，未提供则取 i18n parity.description
  legacyPath?: string                                           // 选填，渲染 "Legacy path: <code>{legacyPath}</code>"
  endpoints?: string[]                                          // 选填，渲染 "Expected API surface" + chip 列表
  actions?: Array<{ label: string; to: string }>                // 选填，PageHeader 右侧 ghost 按钮
  standalone?: boolean                                          // 选填，true → 自带 <main>+max-w-4xl 容器；false → 嵌入父 layout
}
```

i18n 资源（`src/i18n/locales/en.ts:385-391`，对应 `zh.ts:385-391`）：
- `parity.description` — 默认描述
- `parity.entryRestored` — 卡片副标题
- `parity.entryRestoredDescription` — 卡片正文
- `parity.legacyPath` — 旧版路径前缀
- `parity.expectedApiSurface` — API 面板小标题

---

## 4. 旁路死代码：`admin/Placeholder.tsx`（PlaceholderPage）

- 文件路径：`src/pages/admin/Placeholder.tsx`
- 导出：`export function PlaceholderPage({ title, description })`
- 引用次数：**0 次**（`git grep PlaceholderPage` 仅命中定义本身）
- 内容：渲染 PageHeader + Card + 硬编码英文 `This admin section is not yet migrated to v2 — see MIGRATION_TODO.md.`
- **判定**：与 ParityPlaceholder 功能高度重复且无人调用，属可清理的死代码。建议在 P1 整改时随手删除，避免后续误用产生第三套占位 UI。

---

## 5. 占位符分布概览

按守卫分组：

```
Public (无守卫, standalone)            2 项  →  /setup, /key-usage
RequireAuth + Console                  1 项  →  /custom/:id
RequireAdmin + Console                 1 项  →  /admin/subscription-product-config
─────────────────────────────────────────────
合计                                    4 项
```

按 endpoints 数量：

```
1 端点                                  2 项  →  /setup, /custom/:id
2 端点                                  2 项  →  /key-usage, /admin/subscription-product-config
```

按是否带 actions：

```
带 actions（"Login" / "Console" 跳转）  2 项  →  /setup, /key-usage（standalone 公共页）
不带 actions                            2 项  →  /custom/:id, /admin/subscription-product-config
```

---

## 6. 与 P1（占位符减少）的关系

- 这 4 个 placeholder 是**显式声明的待迁移页面**，每一项消除都需要：
  1. 实现真页面（按 MIGRATION_TODO.md 节奏）
  2. 替换 `src/router/index.tsx` 对应行的 `<ParityPlaceholder ... />` 为新组件
  3. 校对 endpoints 列表是否与真实 API 一致（当前 endpoints 是手写的"预期 API 面"，**不是**真调用，迁移时应交叉验证后端契约）
- 优先级建议（仅供参考，不替规划做决策）：
  - **P1.a 高**：`/key-usage`（公共页面，对外展示，丢失影响最大）
  - **P1.a 高**：`/setup`（首次部署流程，缺失会卡新部署）
  - **P1.b 中**：`/admin/subscription-product-config`（admin 入口已挂在侧栏，未迁移则点击侧栏只看到 placeholder，体验断层）
  - **P1.c 低**：`/custom/:id`（动态路由，仅在管理员配置 customMenu 时触发，使用率低）

---

## 7. 红线复核

- [x] 仅审计现有代码，未修改任何文件。
- [x] 未提议组件库重写、未提议改 ParityPlaceholder 自身契约。
- [x] 输出严格限定在 `docs/superpowers/audits/2026-05-12-placeholder-inventory.md`。

---

## 附录 A：完整路由对比（占位 vs 已实现，仅 router/index.tsx 范围）

| 状态 | 数量 | 备注 |
|---|---|---|
| 已实现页面（具体组件） | 39 项 | 含 admin/user/auth/payment 全部已迁移页面 |
| Navigate 重定向 | 6 项 | `/home → /` `/admin → /admin/dashboard` `/admin/subscription-products → /admin/subscriptions` `/admin/channels → /admin/channels/pricing` `/admin/invites → /admin/users` `/admin/affiliates → /admin/affiliates/invites` `/invite → /affiliate` |
| ParityPlaceholder 占位 | 4 项 | 本文清单 |
| `*` 兜底 NotFound | 1 项 | `<NotFoundPage />` |

总路由数 ≈ 50（router/index.tsx 共 137 行）。占位符占比 4/50 = **8%**，绝对值小但全部位于关键路径（公开 / 部署 / admin 入口 / 自定义页面）。

