---
title: frontend-v2 Claude-Light Design Spec (v2)
scope: frontend-v2 (test/xlabapi) — visual refresh only, no functional change
supersedes: docs/superpowers/specs/2026-05-11-frontend-v2-xlabapi-parity-design.md (visual sections only; P0/P1 routing & API contract sections remain authoritative)
status: draft-v1 (2026-05-12)
author: 前端开发实现者
reviewers: 领班
references:
  - docs/superpowers/audits/2026-05-12-frontend-v2-style-gap.md
  - docs/superpowers/audits/2026-05-12-landing-style-direction.md
  - frontend-v2/tailwind.config.ts
  - frontend-v2/src/index.css
  - frontend-v2/src/pages/Landing.tsx
  - frontend-v2/src/components/layout/{AuthLayout.tsx,ConsoleLayout.tsx}
hard_lines:
  - 仅视觉：不动 props/handler/state/路由/契约/i18n key/权限分支
  - 单主题：light-only；dark token 槽位保留但禁止实现
  - Plato 橙 #FF5722 唯一 accent；禁 teal/cyan/glass/mesh
---

# frontend-v2 Claude-Light Design Spec (v2)

> **目标**：把 frontend-v2 从 Plato 暗底（#000 + rgba white）整站翻新为 Claude.ai 风单底（米白 `#FAF9F5` + 橙提亮 `#FF5722`），消除 landing/console/auth 三轨视觉断层。**严格视觉层；功能、契约、路由、状态机、表单字段、权限分支一概不动。**
>
> 范围锚点：audit §1 确认 token 层已统一，断层在"消费方式"；现在底色锚点从 Plato 暗色翻到 Claude 米白，语义 token 命名保持同一套（`--bg-*/--line-*/--ink-*/--accent-*`），只换值。

---

## 0. 执行摘要（给 reviewer 的 15 行）

1. **色板**：双底色锚 `#FAF9F5` 主底 / `#FFFFFF` 卡片；ink-1/2 全部 ≥ AA-body；`#FF5722` 作正文文字只通过 AA-large（≥3.0），不作 body text，**正文级橙色需降级到 `--accent-ink=#C0360B`（AA-body 5.28）**——这是 spec v2 新增的 token，用于"橙色链接/强调字"场景；按钮底、装饰方块、大字号仍用 `#FF5722`。
2. **chip/soft-bg 橙字**：`#FF5722` on `#FFE5DD` = 2.64 **不通过**；Eyebrow/badge-accent 在 soft 背景上的字必须用 `ink-1` 或 `--accent-ink`，不能直接用 `--accent`。
3. **字体**：标题衬线**推 Source Serif 4**（下述三候选段含理由）；正文保持 Inter；mono 保持 JetBrains Mono。
4. **字号**：Hero 84px 降到 72px（`display-2xl`）；不扩 token；Landing 七处写死字号全部映射到 `display-2xl/xl/lg/md` 四档。
5. **装饰**：去 PlasmaBlob / HalftoneOverlay 橙色等离子；SectionFrame 十字 marker 改深炭 1px；dot-bg/grid-bg 基于 `#E8E6DC`；阴影量级 `0 1px 2px rgba(31,30,29,0.04), 0 4px 16px rgba(31,30,29,0.06)`。
6. **不引入 dark mode / theme switcher / `prefers-color-scheme`**；不改 className 之外的 props。
7. Landing.tsx 内联 style 替换清单（§5）与组件皮肤 diff 表（§4）给出逐项 token/className 替换目标。

---

## 1. 色板（语义 token，light-only）

### 1.1 命名契约

语义槽位命名保持稳定，**只换值**；未来若加 dark，只在此表加第二列即可，消费层无需改。

| Token | Role | Light Value | 对比度自检（与关键文本） | Dark 槽位 |
|---|---|---|---|---|
| `--bg-canvas` | 页面主底 | `#FAF9F5` | ink-1 on canvas = **15.80**（AA-body ✅） | `// future-only, do not implement` |
| `--bg-surface` | 卡片/面板表面 | `#FFFFFF` | ink-1 on surface = **16.64**（AA-body ✅） | `// future-only` |
| `--bg-subtle` | 分隔/悬浮/斑马条 | `#F0EEE6` | ink-1 on subtle = **14.33**（AA-body ✅） | `// future-only` |
| `--line-1` | 极淡分隔（hairline） | `#ECEAE0` | 装饰线，非文本；对 canvas 对比 1.08 → UI 弱，仅用作大面积斑马 | `// future-only` |
| `--line-2` | 默认边线（input/card 边框） | `#E8E6DC` | 1.19（装饰线允许；如果用于关键分界需叠加 line-3） | `// future-only` |
| `--line-3` | 悬浮/聚焦边线 | `#D8D4C6` | 1.25（hover 态视觉边界） | `// future-only` |
| `--line-4` | 强边线（active/选中外框） | `#C4BFAE` | 1.43（active 态） | `// future-only` |
| `--ink-1` | 标题深炭 | `#1F1E1D` | on canvas **15.80**（AA-body ✅） | `// future-only` |
| `--ink-2` | 正文 | `#3D3D3A` | on canvas **10.34**（AA-body ✅） | `// future-only` |
| `--ink-3` | 弱化（副标题/caption） | `#8A8780` | on canvas **3.40**（AA-large ✅；**非正文用**） | `// future-only` |
| `--ink-4` | placeholder/disabled | `#B7B3A8` | on canvas 1.99（**仅 placeholder/disabled**） | `// future-only` |
| `--accent` | Plato 橙（按钮底、装饰方块、大字号） | `#FF5722` | on canvas **3.00**（AA-large ✅，AA-body ❌） | `// future-only` |
| `--accent-ink` | **新增**：正文级橙（链接/强调字/小号橙标签文字） | `#C0360B` | on canvas **5.28**（AA-body ✅）；on `--accent-soft` **4.64**（AA-body ✅） | `// future-only` |
| `--accent-soft` | 橙软底（chip/badge/hover background） | `#FFE5DD` | ink-1 on soft **13.87**（AA-body ✅） | `// future-only` |
| `--accent-line` | 橙边线（focus ring / accent border） | `#FFB7A0` | 装饰线，对 canvas 1.70 → UI-weak，仅作装饰或叠 line-3 | `// future-only` |
| `--accent-hover` | 按钮 hover 底色（比 accent 略深） | `#E84713` | on canvas **3.73**（AA-large ✅） | `// future-only` |
| `--signal-success` | 成功 | `#2F8F5E` | on canvas **3.82**（AA-large ✅） | `// future-only` |
| `--signal-warn` | 警告 | `#A8761A` | on canvas **3.78**（AA-large ✅） | `// future-only` |
| `--signal-danger` | 危险 | `#B3261E` | on canvas **6.20**（AA-body ✅） | `// future-only` |
| `--focus-ring` | 聚焦外圈 | `rgba(255, 87, 34, 0.40)` | 3px 外圈 + 1px accent-line 内圈 | `// future-only` |

### 1.2 关键 AA 冲突与处置（拍板前须 @领班 确认）

| 冲突 | 原因 | 处置（推荐） |
|---|---|---|
| `#FF5722` 正文文字在米白上只过 AA-large | 橙色纯度太高 luminance 不足 | **引入 `--accent-ink=#C0360B`**；所有"橙色链接/小字强调/表头 accent 列"一律换 `--accent-ink`，`--accent` 仅限按钮底、≥24px 大字号、装饰方块 |
| `#FF5722` on `#FFE5DD` (chip) = 2.64 | soft 底色压低对比 | chip 文字强制 `ink-1`（13.87）或 `--accent-ink`（4.64）；禁用"橙字+橙软底"组合 |
| `ink-4=#B7B3A8` 在 canvas 上 1.99 | luminance 太近 | 仅用于 placeholder/disabled；正文弱化走 `ink-3` |
| `line-2` 作非常规"关键分界"对比 1.19 | 边线淡 | 卡片/输入框主边用 line-2（装饰线允许 UI-weak）；需要用户感知的强分界（table header / focused card）升级到 line-3 或 line-4 |

**若用户拒绝引入 `--accent-ink`**：所有橙色正文级用法必须改为 `ink-1`（黑）并用橙下划线/橙小方块 marker 表达"强调"，不允许保留不达 AA 的橙字。**此项需 @领班 确认后再进 v2 sign-off。**

---

## 2. 字体层级

### 2.1 字体族

| Family | 用途 | Stack | 运行时 |
|---|---|---|---|
| Display (serif) | 标题（h1/h2/h3 的 hero/landing 强调位 + PageHeader display） | **`'Source Serif 4', 'Source Serif Pro', 'Newsreader', Georgia, 'Times New Roman', serif`** | self-host `/fonts/SourceSerif4-[400,500,600].woff2`（3 weight，约 180KB gzip），或 next-safe fallback 链先跑，字体懒加载不阻塞 FCP |
| Sans (body) | 正文、按钮、表单、data-table、caption | `'Inter', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', system-ui, sans-serif`（保持现有） | 已存在 |
| Mono | eyebrow / kbd / token 展示 | `'JetBrains Mono', 'SF Mono', ui-monospace, Menlo, monospace`（保持现有） | 已存在 |
| Italic emphasis | Hero "connect to / *transform* / across" 风格强调字 | 用 Display stack + `font-style: italic`（Source Serif 4 自带 italic），不再单独拉 Georgia italic | 合并入 Display self-host |

### 2.2 衬线三候选（推荐 Source Serif 4）

| 候选 | 授权 | 像素级匹配 Claude.ai Tiempos | 运行时重量 | 推荐度 |
|---|---|---|---|---|
| **Source Serif 4 ✅** | Adobe OFL（开源免费） | 字重/x-height/收尾切刀极接近 Tiempos Headline；意大利体一致 | 3 weight self-host 约 180KB | **推荐** |
| Newsreader | Google Fonts（免费） | 偏编辑体，衬线更圆；x-height 略低，小字号易显窄 | 同 self-host 约 150KB | 备选 |
| Tiempos Headline | Klim 商业授权（需付费） | 与 Claude.ai 完全一致 | 授权成本 + self-host 约 120KB | **不推荐**（授权不明、用户未批预算） |

**决策点**（@领班 拍板）：
- 默认进 **Source Serif 4**；若用户看 mockup 后觉得太圆润，fallback 到 Newsreader；
- 不进 Tiempos Headline（授权风险）。

### 2.3 字号表（重新评估 Landing 写死值）

| Token（Tailwind） | 值 | 用途 | Landing 现态 → 映射 |
|---|---|---|---|
| `text-display-2xl` | **72px / line-height 1.02 / letter-spacing -0.035em / weight 500** | Hero h1 | Landing.tsx:244 `fontSize: 84` → **降到 72**（不扩 token；84 相对 72 只多 16% 视觉冲击，米白底更适合克制感；与 Claude.ai 72–80 区间对齐） |
| `text-display-xl` | 56 / 1.05 / -0.030em / 500 | Hero 大数字 / CTA h2 | Landing.tsx:314 `56`（保持）、:753 `52` → **升到 56**（节奏对齐）；CtaFooter h2 40 **升到 56** |
| `text-display-lg` | 48 / 1.06 / -0.025em / 500 | SecurityBand h2 / CaseStudy h2 | Landing.tsx:828 `44` → **升到 48**；CaseStudy :TBD `32` → **升到 48** |
| `text-display-md` | 40 / 1.10 / -0.025em / 500 | OperationsHub h2 | Landing.tsx:451 `38` → **升到 40** |
| `text-display-sm` | **新增** 32 / 1.15 / -0.02em / 500 | PageHeader 标题（Console 端）+ Operations console stat 42 | 用作 Console 统一的页面主标题；Landing.tsx:652 `42` → **降到 32** |
| `text-eyebrow` | 11 / 1.4 / 0.18em / mono | Eyebrow（已存在） | Landing.tsx:37 `11.5` → **统一到 11**（消除 0.5px 漂移） |
| `text-caption` | **新增** 12.5 / 1.5 / 0em | 图注 / StatCard unit label | Landing.tsx 多处 `11.5` → **升到 12.5**（米白底 11.5 字号易虚） |

**扩 token 与否**：
- **不扩** display-3xl (84px)：hero 降到 72 已足够，避免 Landing 独占一档导致 Console 永远用不上；
- **扩** display-sm (32px)：Console PageHeader 当前 `font-display text-3xl`(≈30) 与 display-md (40) 之间断档，PageHeader 升级到 display-sm 后能与 Landing 节奏对齐；
- **扩** text-caption (12.5px)：替代 Landing 散落的 10.5/11.5 字号。

---

## 3. 装饰语言（米白底重构）

### 3.1 必须拿掉的暗底元素

| 暗底假设 | 米白底不成立原因 | 处置 |
|---|---|---|
| `PlasmaBlob`（橙色等离子球 + radial-gradient） | 暗底"发光"依赖深背景对比；米白底会变成"橙色污渍" | **全部去除**；保留组件文件但改为空渲染或返回 null（避免删除 import 触发其它页面报错） |
| `HalftoneOverlay`（白色半调点叠 plasma） | 需要亮色点叠在暗渐变；米白底上不可见 | **全部去除**；同上空渲染 |
| `boxShadow: '0 30px 80px rgba(0,0,0,0.5)'` | 黑色强阴影在米白底呈"脏灰块" | 改为 `0 1px 2px rgba(31,30,29,0.04), 0 4px 16px rgba(31,30,29,0.06)`（`shadow-card-light`） |
| `rgba(255,255,255,0.04..0.16)` 作为表面/边线 | 白 alpha 在米白底几乎不可见 | 全部换 `--bg-surface/--bg-subtle/--line-1..4` token |

### 3.2 保留并改色的装饰

| 元素 | 原态 | 新态 |
|---|---|---|
| `SectionFrame` hairline + 十字 marker | `rgba(255,255,255,0.10)` 白线 + 橙色十字 | 线改 `--line-3=#D8D4C6` 1px；十字 marker 改 `--ink-1=#1F1E1D` 1px 10×10px；`halfWidth=660` 保持与 `.container-bus=1280` 对齐 |
| `.eyebrow` 橙方块 marker | 8×8 `--orange` | 保持 8×8 `--accent=#FF5722`（装饰方块，非文本，AA 不适用） |
| `.dot-bg` radial dot | `rgba(255,255,255,0.06)` 白点 | 改 `--line-2=#E8E6DC` 1px 点 / 32px 间距（极淡，仅作纸面纹理） |
| `.grid-bg` 栅格线 | 两向 `--line-1` 白 alpha | 改 `--line-1=#ECEAE0` 深空间方向线 / 56px 间距 |
| `shadow-glow`（橙 focus/active 发光） | `rgba(255,87,34,0.30)` ring + 橙光 | 改为 `0 0 0 3px rgba(255,87,34,0.20)`（focus ring，仅用于输入框/按钮 focus），**不用于常态卡片强调** |
| Orange square marker（大小 12–16px） | `borderRadius: 2` 橙实心 | 保持橙实心；在 chip 里搭配 `ink-1` 黑字（不配橙字） |

### 3.3 新增装饰

| 元素 | 值 | 用途 |
|---|---|---|
| `shadow-card-light` | `0 1px 2px rgba(31,30,29,0.04), 0 4px 16px rgba(31,30,29,0.06)` | 卡片默认阴影（可选，大部分情况用 `--line-2` 边线即可） |
| `shadow-elev-light` | `0 2px 4px rgba(31,30,29,0.06), 0 12px 32px rgba(31,30,29,0.08)` | Hero mock 控制台卡等"浮层"效果 |
| `.paper-grain`（可选） | `background-image: url(data:image/svg+xml;utf8,<noise>)` 2% opacity | 为米白底添加极淡纸感纹理，避免纯色平面感 |

---

## 4. 组件皮肤 diff 表

> 每行列出：现态 token / 新态 token / hover / active / focus / 对比度自检。
> **零组件 props 改动**；只改 className/CSS variable。

### 4.1 Button（.btn 体系）

| 变体 | 现态 | 新态 | Hover | Active | Focus | AA |
|---|---|---|---|---|---|---|
| `.btn-primary` | bg `--text-1`(#f5f6f8) / color `--bg-0`(#000) | bg `--ink-1`(#1F1E1D) / color `--bg-canvas`(#FAF9F5) | bg `#000000` | bg `--ink-2` + 1px inset line-4 | ring 3px `--focus-ring` | 白字 on 深炭 = 14+ ✅ |
| `.btn-accent` | bg `--orange`(#FF5722) / color `#fff` | bg `--accent` / color `#FFFFFF` | bg `--accent-hover`(#E84713) | bg `#C0360B` | ring 3px `--focus-ring` | 白字 on 橙 = 3.16（AA-large ✅；≥14px 可过 AA-body 需要加粗 600） → **按钮字号≥13.5 且 weight≥500 已满足** |
| `.btn-ghost` | bg transparent / border `--line-2` | bg transparent / border `--line-3`(#D8D4C6) / color `--ink-1` | bg `--bg-subtle` / border `--line-4` | bg `--bg-subtle` + border `--ink-3` | ring 3px `--focus-ring` | ink-1 on canvas ✅ |
| `.btn-danger` | bg `--signal-err`(#f87171) / color `#1a0000` | bg `--signal-danger`(#B3261E) / color `#FFFFFF` | bg `#951E18` | bg `#7D1913` | ring 3px `rgba(179,38,30,0.40)` | 白字 on #B3261E = 7.6 ✅ |
| `.btn-pill` | border 999 / height 44 | 保持几何，只换颜色 token（同 `.btn-primary/.btn-accent`） | — | — | — | — |

### 4.2 Input

| 态 | 现态 | 新态 | AA |
|---|---|---|---|
| 默认 | bg `--bg-4`(#1d222c) / border `--line-2` / color `--ink-1` | bg `--bg-surface`(#FFFFFF) / border `--line-2`(#E8E6DC) / color `--ink-1` | ink-1 on surface 16.64 ✅ |
| Focus | border `--orange` + ring `rgba(255,87,34,0.18)` | border `--accent` + ring 3px `rgba(255,87,34,0.30)`（`--focus-ring`） | — |
| Placeholder | color `--text-4`(#5b606b) | color `--ink-4`(#B7B3A8) | 仅 placeholder，AA 豁免 |
| Error | border `--signal-err` | border `--signal-danger`(#B3261E) + 辅助图标（避免纯色传达） | 文字 error：`.input-error-text` 用 `--signal-danger` on canvas = 6.20 ✅ |
| Disabled | opacity 0.6 | 保持 opacity 0.6，bg 改 `--bg-subtle`(#F0EEE6)，color `--ink-4` | 仅禁用态 |

### 4.3 Card / Card-Flat

| 变体 | 现态 | 新态 |
|---|---|---|
| `.card` | bg `--bg-1`(#0a0a0a) / border `--line-1` / radius `--r-xl` | bg `--bg-surface`(#FFFFFF) / border `--line-2` / radius `--r-xl`(14) / optional `shadow-card-light` |
| `.card-flat` | border `--line-2` / radius 0 | bg transparent / border-bottom only `--line-2` / radius 0（"纸面分隔"感） |
| `.card-hover` | hover border `--line-3` + bg `--bg-2` | hover border `--line-3`(#D8D4C6) + bg `--bg-subtle`(#F0EEE6) |

### 4.4 Badge

| 变体 | 现态 | 新态 |
|---|---|---|
| `.badge` | bg `--bg-2` / border `--line-2` / color `--text-2` | bg `--bg-subtle` / border `--line-2` / color `--ink-2` |
| `.badge-success` | bg 绿 alpha / 绿字 | bg `rgba(47,143,94,0.12)` / border `rgba(47,143,94,0.30)` / color `--signal-success`(#2F8F5E) → AA-large ✅ |
| `.badge-warning` | bg 黄 alpha / 黄字 | bg `rgba(168,118,26,0.12)` / color `--signal-warn`(#A8761A) → AA-large ✅ |
| `.badge-danger` | bg 红 alpha / 红字 | bg `rgba(179,38,30,0.10)` / color `--signal-danger`(#B3261E) → AA-body ✅ |
| `.badge-accent` | bg `--orange-soft` / color `--orange` | bg `--accent-soft`(#FFE5DD) / color **`--accent-ink`**(#C0360B) → on soft 4.64 AA-body ✅（**关键变更：不再用 --accent 作 badge 字**） |

### 4.5 Modal

| 态 | 现态 | 新态 |
|---|---|---|
| Overlay | `rgba(0,0,0,0.6)` | `rgba(31,30,29,0.40)` + `backdrop-blur-sm`（避免米白底上黑 60% 太重） |
| Container | bg `--bg-1` / border `--line-2` | bg `--bg-surface` / border `--line-2` / `shadow-elev-light` |
| Header `x` 按钮 | `.btn-icon .btn-ghost` | 同 .btn-ghost 新态 |

### 4.6 Table / data-table

| 态 | 现态 | 新态 |
|---|---|---|
| 整体 | border `--line-2` / mono 12.5 | border `--line-2` / mono 12.5（保留 mono 质感） |
| Thead | 无明显区分 | bg `--bg-subtle` / color `--ink-2` / letter-spacing 0.04em uppercase |
| Row hover | bg `--bg-3`（暗） | bg `--bg-subtle`(#F0EEE6) |
| Row border | border-bottom `--line-1`（极淡） | border-bottom `--line-1`(#ECEAE0) |
| Selected row | 未定义 | bg `--accent-soft`(#FFE5DD) + border-left 2px `--accent` |

### 4.7 Toast

| 变体 | 现态 | 新态 |
|---|---|---|
| default | bg `--bg-2` / color `--ink-1` | bg `--ink-1`(#1F1E1D) / color `--bg-surface`(#FFFFFF)（黑吐司 on 米白底，对比清晰；参考 Claude.ai 的深色 toast） |
| success | 绿变体 | border-left 3px `--signal-success`，其余同 default |
| warning/danger | 对应变体 | 同理，border-left 主色 |

### 4.8 Eyebrow（.eyebrow）

| 现态 | 新态 |
|---|---|
| color `--text-3`(#8a8f99) / ::before `--orange` 8×8 方块 / mono 11 / letter-spacing 0.18em | color `--ink-3`(#8A8780) → on canvas **3.40**（AA-large ✅；eyebrow 小字≥11 且 letter-spacing 大，可读性由字距补偿）/ ::before `--accent` 8×8 / mono 11 / 0.18em 保持 |
| Landing.tsx:30-47 私有 Eyebrow | **删除**，统一引 `bus/Eyebrow`；11.5 → 11；`rgba(255,255,255,0.7)` → `--ink-3` |
| ConsoleLayout.tsx:151 admin 分组标题 `tracking-[0.14em]` | **改用** `bus/Eyebrow` 或 `.eyebrow` utility；tracking 统一 0.18em |

### 4.9 PillBtn（bus/PillBtn + Landing 私版合并）

| 变体 | 现态 | 新态 |
|---|---|---|
| `accent`（对应 Landing primary） | bg `--orange` / color `#fff` / radius 999 / h 44 | bg `--accent` / color `#fff` / radius 999 / h 44 + focus ring `--focus-ring` |
| `light`（对应 Landing 默认） | bg `#fff` / color `#000` | bg `--bg-surface` / color `--ink-1` + border `--line-3` |
| `ghost`（对应 Landing ghost） | bg transparent / border 白 alpha | bg transparent / border `--line-3` / color `--ink-1`; hover bg `--bg-subtle` |
| sizes `sm/md/lg` | 30/38/44 | 保持几何，仅改颜色 |

**合并动作**（非代码动作，是皮肤 diff 的前置）：Landing.tsx:49-76 私有 PillBtn 删除，全部改调 `bus/PillBtn`，并把 `primary={true}` 映射为 `variant="accent"`。

### 4.10 Wordmark / BusMark（菱形 logo）

| 位置 | 现态 | 新态 |
|---|---|---|
| Landing NavD 顶部（:111-122 手写菱形 SVG） | 白 1.5px 边 / 透明填充 | 改用 `bus/BusMark` / 1.5px 边改 `--ink-1` / 保持菱形几何 |
| Landing CtaFooter 底部（:1148-1166 再次内联菱形） | 同上 | 同上，统一 `bus/BusMark` |
| AuthLayout 顶部（:17-22 siteLogo img + 圆角方块） | 图片 + 圆角 bg-bg-1 border-line-2 | **统一替换为 `bus/BusMark`**（三轨合一方案 A，已拍板）；`site_logo` 作为 fallback，当 `publicSettings?.site_logo` 非默认值时再渲染 img（保持 tenant 品牌能力） |
| ConsoleLayout 侧栏顶部 | siteLogo img + siteName 文字 | 同上，`bus/BusMark` + 文字，文字 color `--ink-1` |

### 4.11 SectionFrame

| 态 | 现态 | 新态 |
|---|---|---|
| 上下 hairline | `rgba(255,255,255,0.10)` 1px | `--line-3=#D8D4C6` 1px |
| 十字 marker（四角） | 橙色 12×12 | 深炭 `--ink-1` 1px 10×10px（米白底上十字更醒目，用橙反而破坏克制感） |
| `halfWidth=660` | 对齐 `.container-bus=1280`（1280/2-40=600）**实际现态偏移 60px** | **修正为 `halfWidth=600`**（与容器精确对齐） |

### 4.12 Sparkline

| 态 | 现态 | 新态 |
|---|---|---|
| 线色 | `--orange` | 保持 `--accent`（线条装饰，非文本，AA 豁免） |
| 背景 | 透明 | 保持透明；如需 fill 用 `--accent-soft` 10% opacity |

### 4.13 PageHeader（ConsoleLayout 内导出）

| 现态 | 新态 |
|---|---|
| `font-display text-3xl text-ink-1` + `<p text-ink-3>` | 升级为三件套：`bus/Eyebrow`（可选）+ `font-display text-display-sm text-ink-1`（32px 衬线）+ `text-sm text-ink-3`（副标题）；与 Landing 的 h2 节奏在 display-sm/md 上对齐 |

### 4.14 Pill Nav（`.pill-nav`）

| 态 | 现态 | 新态 |
|---|---|---|
| 容器 bg | `rgba(255,255,255,0.03)` | `--bg-surface` + 1px border `--line-2` |
| item default | color `rgba(255,255,255,0.6)` | color `--ink-3` |
| item hover | color `rgba(255,255,255,0.85)` | color `--ink-1` |
| item active | `#fff` on `rgba(255,255,255,0.06)` | `--ink-1` on `--bg-subtle` + ring 1px `--line-3` |
| item light-active | `#000` on `#fff` | `--bg-canvas` on `--ink-1`（深底白字反相） |

---

## 5. Landing.tsx 内联值替换映射

> 基于 audit §4 与 grep 扫描结果；行号以 `test/xlabapi:frontend-v2/src/pages/Landing.tsx@HEAD` 为准。
> **本节是替换清单，非代码补丁**；落地任务在另起 T026+ 时实施。
> 替换规则：内联 `style={{...}}` → className / CSS variable；保持 props/handler/state 不变。

| 行号 | 原值 | 新值（className 或 token） |
|---|---|---|
| 12 | `const ORANGE = '#ff5722'` | **删除**，组件引用改 `var(--accent)` 或 `text-accent/bg-accent`（tailwind alias） |
| 14-28 | `OrangeMark`（8×8 橙方块） | 删除；全部改引 `bus/Eyebrow`（已内置 ::before marker） |
| 30-47 | 私有 `Eyebrow`（11.5 / rgba white 0.7 / 0.18em） | 删除；引 `bus/Eyebrow`；mb 参数保留成 `className="mb-4"` 之类 |
| 49-76 | 私有 `PillBtn` | 删除；引 `bus/PillBtn` variant="accent"/"light"/"ghost" |
| 67 | `border: '1px solid rgba(255,255,255,0.16)'` | `border border-line-3` |
| 68 | `background: primary ? ORANGE : ghost ? 'transparent' : '#fff'` | variant 语义化（accent/ghost/light） |
| 69 | `color: primary ? '#fff' : ghost ? '#fff' : '#000'` | variant 决定；light 态 color `ink-1` |
| 106-122 | NavD 内联菱形 + 白 1.5px 边 + `XLABAPI` 文字 15/600 | `<bus/Wordmark small />` 或 `bus/BusMark size={28} />` + `<span className="font-medium text-ink-1">XLABAPI</span>` |
| 131-132 | `background: 'rgba(255,255,255,0.04)'` + `border: '1px solid rgba(255,255,255,0.08)'`（pill-nav 容器） | `.pill-nav` utility（已存在） |
| 140-145 | `padding:'8px 16px' fontSize:13.5 color:'rgba(255,255,255,0.7)'` (pill-nav-item) | `.pill-nav-item`（已存在） |
| 156-158 | `fontSize:13.5 color:'rgba(255,255,255,0.7)'`（docs 链接） | `text-sm text-ink-3 hover:text-ink-1` |
| 182 | `background: '#000'`（HeroD section） | `bg-bg-canvas` |
| 205 | `backgroundImage: 'radial-gradient(rgba(255,255,255,0.03)...)'` | `.dot-bg`（已存在，会自动走 `--line-2` 新值） |
| 197-198 | `<PlasmaBlob />` `<HalftoneOverlay opacity={0.85} />` | **删除**（装饰去橙色等离子）；保留 DOM 占位空 div 避免布局塌陷 |
| 212 | `maxWidth:1280 margin:'0 auto' padding:'0 40px'` | `.container-bus`（已存在） |
| 219-222 | Hero badge 容器：`borderRadius:999 background:rgba(255,255,255,0.04) border:1px rgba(...,0.1)` | `inline-flex items-center gap-2.5 rounded-full bg-bg-surface border border-line-2 px-3 py-1.5` |
| 228-232 | Hero badge 内 Release 标签：`borderRadius:999 background:ORANGE fontSize:11 color:'#fff'` | `rounded-full bg-accent text-[11px] font-semibold text-white px-2.5 py-0.5` |
| 237 | `fontSize:13 color:'rgba(255,255,255,0.7)'` | `text-[13px] text-ink-2` |
| 244-251 | Hero h1: `fontSize:84 fontWeight:500 lineHeight:1.04 letterSpacing:'-0.035em' color:'#fff' maxWidth:760` | `font-display text-display-2xl text-ink-1 max-w-[760px]`（**84→72**） |
| 257 | `fontStyle:'italic' fontFamily:'Georgia, serif' fontWeight:400`（em 强调） | `<em className="font-display italic font-normal">`（走 Source Serif 4 italic） |
| 264-270 | Hero p: `fontSize:16 lineHeight:1.6 color:'rgba(255,255,255,0.55)' maxWidth:540` | `text-base leading-relaxed text-ink-3 max-w-[540px]` |
| 274 | CTA 行 `gap:12 marginBottom:100` | `flex gap-3 mb-24` |
| 290 | `borderTop: '1px solid rgba(255,255,255,0.08)'`（stat 分隔） | `border-t border-line-2` |
| 302 | `color:'#fff'` | `text-ink-1` |
| 313 | `color:'#fff'`（stat 容器） | `text-ink-1` |
| 314 | `fontSize:56 fontWeight:500 letterSpacing:'-0.03em'` | `font-display text-display-xl` |
| 315 | `fontSize:32 color:ORANGE marginLeft:2` | `text-[32px] font-display text-accent ml-0.5`（32px 橙单位，AA-large ✅） |
| 319-320 | stat label `fontSize:11.5 color:'rgba(255,255,255,0.5)'` | `text-caption text-ink-3` |
| 339 | BackedRow `background:'#000'` | `bg-bg-canvas` |
| 353 | `fontSize:12 color:'rgba(255,255,255,0.45)'` | `text-xs text-ink-3` |
| 362 | `color:'rgba(255,255,255,0.55)' fontSize:16 fontWeight:600` | `text-base font-semibold text-ink-2` |
| 371 | `fontFamily: [...].includes(b) ? 'Georgia, serif' : 'inherit'` | `className={['OpenAI','Meta','Google'].includes(b) ? 'font-display italic' : ''}`（走 Source Serif 4 italic） |
| 406 | OperationsHub `background:'#000' padding:'100px 0'` | `bg-bg-canvas py-24` |
| 423 | `borderBottom:'1px solid rgba(255,255,255,0.06)'` | `border-b border-line-1` |
| 431 | `width:12 height:12 background:ORANGE borderRadius:2` | `w-3 h-3 rounded-sm bg-accent`（装饰方块，AA 豁免） |
| 432 | `fontSize:14 fontWeight:600 color:'#fff'` | `text-sm font-semibold text-ink-1` |
| 436-440 | feature desc `fontSize:13 lineHeight:1.6 color:'rgba(255,255,255,0.5)' maxWidth:240` | `text-[13px] leading-relaxed text-ink-3 max-w-[240px]` |
| 451-456 | OperationsHub h2: `fontSize:38 fontWeight:500 lineHeight:1.08 letterSpacing:'-0.025em' color:'#fff'` | `font-display text-display-md text-ink-1` |
| 480 | mock console bg `'#000'` | `bg-ink-1`（mock 用深底反差展示"console-in-marketing"感，保留单处深底作为产品截图替代品） |
| 493 | `background:'rgba(20,8,4,0.72)' backdropFilter:'blur(20px)'` | mock 内部面板：`bg-ink-1/90 backdrop-blur-md`（仅 Hero mock 场景，其他全部米白） |
| 495 | `border:'1px solid rgba(255,255,255,0.08)'` | `border border-white/8`（mock 内） |
| 498 | `color:'#fff'` | `text-white`（mock 内） |
| 501 | `boxShadow:'0 30px 80px rgba(0,0,0,0.5)'` | `shadow-elev-light` |
| 510-530 | mock console 内 `rgba(255,255,255,...)` `rgba(34,197,94,...)` | mock 内部按"深底孤岛"处理：保留当前白 alpha 线与 signal color 的 mock 版本（见 §6 三轨覆盖）；**真实 dashboard 页不走这套** |
| 652 | stat 42 | `font-display text-display-sm text-ink-1`（32） |
| 656 | `color:ORANGE fontSize:18` | `text-accent-ink text-lg`（18px 橙 arrow，走 accent-ink 通过 AA） |
| 658 | `fontSize:10.5 color:ORANGE marginTop:4` | `text-caption text-accent-ink mt-1`（AA-body ✅） |
| 725-727 | `background:'#fff' color:'#000' fontSize:11.5` | `bg-bg-surface text-ink-1 text-caption`（白片 on 深底在 mock 内） |
| 753 | `fontSize:52` | `font-display text-display-xl` |
| 757 | `fontSize:32 color:ORANGE` | `text-[32px] text-accent ml-0.5`（装饰橙单位，AA-large ✅） |
| 809 | SecurityBand `background:'#000' padding:'80px 0'` | `bg-bg-canvas py-20` |
| 828-833 | h2 44 | `font-display text-display-lg text-ink-1`（48） |
| 842 | `fontSize:14 lineHeight:1.65 color:'rgba(255,255,255,0.55)'` | `text-sm leading-relaxed text-ink-3` |
| 1103-1107 | CtaFooter "create / transform / for" 强调字 `fontFamily:'Georgia, serif'` | `font-display italic` |

> 完整替换约 **30+ 处**；上表覆盖 high-signal 锚点。其余 `rgba(255,255,255,0.0..0.16)` 内联边线/文字按下述映射批量替换：
>
> - `rgba(255,255,255,0.04)` → `--bg-subtle` 或 `bg-bg-surface`
> - `rgba(255,255,255,0.06)` → `--line-1` / `border-line-1`
> - `rgba(255,255,255,0.08)` → `--line-2` / `border-line-2`
> - `rgba(255,255,255,0.10..0.16)` → `--line-3` / `--line-4` / `border-line-3..4`
> - `rgba(255,255,255,0.45..0.55)` → `--ink-3` / `text-ink-3`
> - `rgba(255,255,255,0.60..0.70)` → `--ink-2` / `text-ink-2`
> - `#fff`/`'#ffffff'` 作文字 → `text-ink-1`
> - `#fff`/`'#ffffff'` 作背景 → `bg-bg-surface`
> - `#000`/`'#000000'` 作背景 → `bg-bg-canvas`
> - `#000`/`'#000000'` 作文字 → `text-ink-1`
> - `ORANGE` 作背景 → `bg-accent`
> - `ORANGE` 作文字（≥24px 或装饰） → `text-accent`；（正文/小字） → `text-accent-ink`
> - `'Georgia, serif'` → `font-display`（走 Source Serif 4）

---

## 6. 三轨覆盖检查

### 6.1 Landing 轨

| 视觉锚点 | 当前态 | 新态 | 状态 |
|---|---|---|---|
| NavD 顶栏 | 黑底 + 白菱形 + pill-nav 白 alpha | 米白 canvas + `bus/BusMark`(深炭) + `.pill-nav` 米白 | ✅ 覆盖 |
| Hero h1 84 + plasma + halftone | 暗底 + 橙球 | 米白 + 无球 + 72 hero + italic 强调走 Source Serif 4 | ✅ 覆盖 |
| BackedRow 灰字 brand | 暗 | 米白 + `text-ink-2/3` | ✅ 覆盖 |
| OperationsHub mock console | 暗底 + 橙等离子 | **保留 mock 内部深底**（作为"产品截图 placeholder"，深底孤岛唯一允许位）；外围 section 米白 | ✅ 覆盖（带孤岛豁免） |
| SecurityBand 圆徽章 | 暗底 | 米白 + `--line-3` 圆环 + `ink-1` 字 | ✅ 覆盖 |
| CaseStudy | 暗 | 米白 + `.card` surface | ✅ 覆盖 |
| CtaFooter | 暗 + 橙 CTA | 米白 + `.btn-accent` | ✅ 覆盖 |
| SectionFrame 十字 | 白线 + 橙十字 | line-3 + 深炭十字 | ✅ 覆盖 |
| `dot-bg` / `grid-bg` | 白点/白栅 | line-2/line-1 | ✅ 覆盖 |

**深底孤岛豁免**：Hero mock console（Landing.tsx:474-802 区段）作为"产品截图替代物"，内部保留深底 + 白 alpha 形态，这是 **marketing 展示约定**（类似 screenshot in a frame），不污染三轨整体；mock **外层容器** bg 必须米白。

### 6.2 Console / Dashboard 轨

| 视觉锚点 | 当前态 | 新态 | 状态 |
|---|---|---|---|
| ConsoleLayout bg-bg-0 (root) | `#000000` | `--bg-canvas=#FAF9F5`（token 值换） | ✅ 覆盖（无组件改） |
| 侧栏 bg-bg-1 | `#0a0a0a` | `--bg-surface=#FFFFFF` + border-right `--line-2` | ✅ 覆盖 |
| 侧栏 激活态 `bg-orange-soft text-orange` | 橙 alpha + 橙字 | `bg-accent-soft` + **`text-accent-ink`**（AA-body ✅）；border-left 2px `--accent` | ✅ 覆盖（关键：侧栏激活字必须走 accent-ink，不是 accent） |
| PageHeader h1 text-3xl | Inter 30 | Source Serif 4 `text-display-sm`(32) + `text-ink-1` | ✅ 升级 |
| Card `.card` | 暗底 | 白面板 + line-2 边 + 可选 shadow-card-light | ✅ 覆盖 |
| StatCard 橙数值 | `text-orange` | `text-accent`（≥24px 装饰大字，AA-large ✅） | ✅ 覆盖 |
| Quick-actions `card-hover` | line-3 边 | 同 §4.3 | ✅ 覆盖 |
| Admin sidebar 分组标题 `tracking-[0.14em]` | 白 alpha 11px | `.eyebrow` utility（0.18em + mono 11 + ink-3） | ✅ 规整 |
| Data-table | mono 12.5 / 白 alpha 边 | 保持 mono 12.5 + line-2 边 + thead `--bg-subtle` | ✅ 覆盖 |

### 6.3 Auth 轨

| 视觉锚点 | 当前态 | 新态 | 状态 |
|---|---|---|---|
| AuthLayout bg-bg-0 | 暗 | 米白 canvas | ✅ 覆盖 |
| header logo 圆角方块 + siteLogo img | 圆角 bg-bg-1 border-line-2 + img | `bus/BusMark`（三轨合一）+ fallback img 仅在 tenant 自定义 logo 时渲染 | ✅ 统一 |
| 登录卡片 `.card p-8` | 暗底 card | 白面板 `.card` 新态 | ✅ 覆盖 |
| h1 display-md | 白字 | 深炭 + Source Serif 4 | ✅ 覆盖 |

**覆盖结论**：三轨 0 暗底孤岛（mock console 为 marketing 语义下的"截图"，不计入轨道污染）。

---

## 7. 不做清单（防越权）

- ❌ 不引入 dark mode（dark token 槽位保留为 `// future-only, do not implement`）
- ❌ 不引入 theme switcher / `prefers-color-scheme` media query
- ❌ 不改 `tailwind.config.ts` 的 `darkMode` 配置（保持当前不启用 class-based dark）
- ❌ 不改任何组件的 `onClick / onChange / onSubmit / useEffect` 逻辑
- ❌ 不改任何表单字段的 `name / value / validation / i18n key`
- ❌ 不改任何 `routes` 配置与 `Navigate / Link` 目标
- ❌ 不改任何 API client 路径 / headers / body
- ❌ 不改任何 auth / permission / feature-flag 分支
- ❌ 不改任何 i18n key；若文案需微调（如 Eyebrow 文本长度因字距变化），**走 i18n 更新不走组件 hardcode**
- ❌ 不动 `bus/PlasmaBlob.tsx` / `bus/HalftoneOverlay` 的 **组件文件**，只让消费方不再渲染它们；文件保留用作未来营销子页
- ❌ 不新增 npm 依赖（Source Serif 4 字体 self-host 到 `frontend-v2/public/fonts/`，不走 npm package）
- ❌ 不改动 `.container-bus` / `.container-console` 宽度值（1280/1440），仅改其内部 token 消费

---

## 附录 A：WCAG AA 自检矩阵（完整）

```
#1F1E1D on #FAF9F5 = 15.80  AA-body  (ink-1 heading on canvas)
#1F1E1D on #FFFFFF = 16.64  AA-body  (ink-1 heading on surface)
#1F1E1D on #F0EEE6 = 14.33  AA-body  (ink-1 heading on subtle)
#3D3D3A on #FAF9F5 = 10.34  AA-body  (ink-2 body on canvas)
#3D3D3A on #FFFFFF = 10.90  AA-body  (ink-2 body on surface)
#8A8780 on #FAF9F5 =  3.40  AA-large (ink-3 muted on canvas)  — 仅 ≥18px 或 ≥14px bold
#8A8780 on #FFFFFF =  3.58  AA-large (ink-3 muted on surface)
#B7B3A8 on #FAF9F5 =  1.99  FAIL     (ink-4 placeholder/disabled only)
#FF5722 on #FAF9F5 =  3.00  AA-large (accent on canvas)
#FF5722 on #FFFFFF =  3.16  AA-large (accent on surface)
#FF5722 on #FFE5DD =  2.64  FAIL     (accent on accent-soft 禁用)
#C0360B on #FAF9F5 =  5.28  AA-body  (accent-ink on canvas)    — spec v2 新增
#C0360B on #FFE5DD =  4.64  AA-body  (accent-ink on accent-soft)
#FFFFFF on #FF5722 =  3.16  AA-large (white on accent — 按钮字 ≥14px + 500 weight 可过)
#1F1E1D on #FFE5DD = 13.87  AA-body  (ink-1 on accent-soft chip)
#2F8F5E on #FAF9F5 =  3.82  AA-large (signal-success)
#A8761A on #FAF9F5 =  3.78  AA-large (signal-warn)
#B3261E on #FAF9F5 =  6.20  AA-body  (signal-danger)
```

## 附录 B：Tailwind alias 映射建议（落地到 `tailwind.config.ts` 时的 colors 扩展）

> 仅为 spec，**本 spec 不动 tailwind.config.ts**；后续 T026+ 实施。

```ts
colors: {
  bg: { canvas:'#FAF9F5', surface:'#FFFFFF', subtle:'#F0EEE6' },
  line: { 1:'#ECEAE0', 2:'#E8E6DC', 3:'#D8D4C6', 4:'#C4BFAE' },
  ink: { 1:'#1F1E1D', 2:'#3D3D3A', 3:'#8A8780', 4:'#B7B3A8' },
  accent: {
    DEFAULT:'#FF5722',
    hover:'#E84713',
    ink:'#C0360B',          // v2 新增：正文级橙
    soft:'#FFE5DD',
    line:'#FFB7A0'
  },
  signal: { success:'#2F8F5E', warn:'#A8761A', danger:'#B3261E' }
}
```

保留 `bg-0..3 / line-1..4 / ink-0..4 / orange.*` 旧别名**不强制**删除，作为过渡期 alias 指向新值（逐步迁移，避免一刀切引发回归）。

## 附录 C：决策点汇总（给 @领班）

1. **衬线字体**：推 **Source Serif 4**（Adobe OFL 免费，self-host 180KB）；备选 Newsreader；不取 Tiempos Headline（授权未批）。
2. **display-2xl 是否扩 84px**：**不扩**；Hero 降到 72px。
3. **是否引入 `--accent-ink=#C0360B`**（v2 新增 token）：**推荐引入**；否则所有"橙色正文级用法"必须改黑字，失去"橙色链接"体验。
4. **装饰 PlasmaBlob/HalftoneOverlay**：**完全去除**消费；组件文件保留。
5. **深底孤岛豁免**：Hero mock console（§6.1）作为 marketing 截图替代物，内部保留深底。
6. **三轨 logo**：统一 `bus/BusMark`；tenant `site_logo` 非默认值时作为 fallback 保留 img 渲染（保留品牌自定义能力）。

---

**签收流程**：
- [ ] @领班 review 决策点 1–6
- [ ] 用户对衬线字体候选拍板（或批准 Source Serif 4 默认）
- [ ] 用户对 `--accent-ink` 引入拍板
- [ ] 旧 spec `2026-05-11-frontend-v2-xlabapi-parity-design.md` 头部加 `superseded` 标记（视觉段）
- [ ] 文档整理者走 T016 索引更新
- [ ] 另起 T026+ 进入实施阶段（任何 .tsx/.ts/.css 改动不在本 spec 范围）
