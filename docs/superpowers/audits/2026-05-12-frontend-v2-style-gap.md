# frontend-v2 Visual Style Gap Audit — Landing ↔ Console/Dashboard

- 审计目标：定位 `frontend-v2`（test/xlabapi 分支）首页 Landing 与登录后 Console / Dashboard 之间的视觉风格断层，输出**只盘点不动代码**的清单与统一方案建议。
- 审计范围：`origin/test/xlabapi:frontend-v2/`，重点文件 `src/pages/Landing.tsx`、`src/pages/Console.tsx`、`src/components/bus/*`、`src/components/layout/ConsoleLayout.tsx`、`src/components/layout/AuthLayout.tsx`、`src/pages/user/Dashboard.tsx`、`src/pages/admin/Dashboard.tsx`、`src/pages/user/Keys.tsx`、`tailwind.config.ts`、`src/index.css`。
- 参考对照：`origin/xlabapi:frontend/` 老 Vue 前端（teal 青色系 + glass shadow + mesh-gradient），与 frontend-v2 设计语言**根本不同源**，仅用于理解用户提到的"console 完全参考 xlabapi 布局"——这里"参考"指**抽屉式侧栏 + 内容主区**结构，**不**指视觉色板。本审计聚焦 frontend-v2 内部 Landing ↔ Console 断层。
- 审计基线：tailwind.config.ts 与 src/index.css 已经定义好同一套 Plato 暗色 + 橙色 token（`bg-0..bg-3 / line-1..line-4 / ink-0..ink-4 / orange / signal-* / display-2xl..display-md / r-sm..r-xl / shadow-card/elev/glow`），并提供 `.btn / .input / .card / .eyebrow / .badge / .container-bus / .container-console / .pill-nav` 等组件原语。**断层不在 token 层，断层在"消费方式"层**。

---

## 1. 设计 Token 对比表（同一份 token，消费方式各做各的）

| 维度 | 全局 Token / 工具类（已存在） | Landing 实际用法 | Console / Dashboard 实际用法 | 一致？ |
|---|---|---|---|---|
| 主背景 | `--bg-0=#000000` / `bg-bg-0` | 内联 `background: '#000'`（Landing.tsx:182、339、406、809、896、1076、1230） | `bg-bg-0`（ConsoleLayout.tsx:118、AuthLayout.tsx:14） | 值一致，**消费方式不一致** |
| 表面层 | `--bg-1=#0a0a0a` / `bg-bg-1` `--bg-2=#0f0f10` `--bg-3=#15161a` `--bg-4=#1d222c` | Landing 几乎不用 bg-1/2/3，直接用半透明白叠在 #000 上（如 `rgba(255,255,255,0.04)`，多见于 NavD 165、Hero badge 220、卡片底色） | Card / 侧栏 / Header 全部走 bg-1（ConsoleLayout 122、195、Card→.card→`background: var(--bg-1)`） | **不一致** |
| 边线 | `--line-1..4` / `border-line-1..4` | 内联 `rgba(255,255,255,0.08)` `rgba(255,255,255,0.10)` `rgba(255,255,255,0.13)` `rgba(255,255,255,0.16)` 共 30+ 处（NavD/Hero/SectionFrame/CaseStudy/CtaFooter） | `border-line-2` 等 utility class（ConsoleLayout.tsx 多处） | 值同源，**消费方式不一致** |
| 文字主色 | `--text-1=#f5f6f8` / `text-ink-1` `--text-2..4` | 内联 `'#fff'` `'rgba(255,255,255,0.55..0.7)'`（Hero、Operations、Footer 大量） | `text-ink-1 / text-ink-2 / text-ink-3` | **不一致**：Landing 的"主色"是真·#fff，比 Console 的 #f5f6f8 略偏白 |
| 强调橙 | `--orange=#ff5722` `orange` 工具类 / `--orange-soft / --orange-line` | Landing.tsx:12 顶部 `const ORANGE = '#ff5722'` 内联，并在 30+ 处直接拼字符串 | `text-orange` `bg-orange-soft` `border-orange-line` utility class（Dashboard StatCard、ConsoleLayout 激活态） | 值一致，**消费方式不一致** |
| 字体族 | `font-sans=Inter+PingFang SC+...`（tailwind） & `font-display=Inter` `font-mono='JetBrains Mono'` & `font-serif=Georgia`（强调 italic） | Landing.tsx:1232 行内 `fontFamily: 'Inter, "PingFang SC", system-ui, sans-serif'`；强调字内联 `Georgia, serif`（Hero 257、CtaFooter 1103/1105/1107） | body 全局继承（index.css body 字体族）；强调字 `font-display` `font-mono` utility class | 等价但 Landing 写死 |
| 显示字号 | `display-2xl=72/1.02/-0.035em` `display-xl=56` `display-lg=48` `display-md=40` | Hero h1: `fontSize: 84` `fontWeight: 500` `lineHeight: 1.04` `letterSpacing: '-0.035em'`（Landing.tsx:243-251）— **未对齐 token**；OperationsHub h2: 38；SecurityBand h2: 44；CaseStudy 32；CtaFooter h2: 40；统计大数字 56 / 52 / 64 / 42 | Dashboard `<h1 className="font-display text-3xl">`（≈30，PageHeader）；StatCard 数值 `font-display text-3xl`；Card 标题 `text-base font-medium`；caption `text-xs uppercase tracking-wider` | **断层显著**：Landing 用 84/64/56/52/44/40/38/32 一组写死值，Console 永远只用 30 / 16 / 13.5 / 12 / 11 |
| Eyebrow | `.eyebrow` 类（mono、11px、letter-spacing 0.18em、橙方块 8x8 marker） + `bus/Eyebrow.tsx`（套 .eyebrow） | Landing.tsx:30-47 自己又写了一份 `function Eyebrow`，内联 11.5/0.18em/橙方块 + 自己写的 `OrangeMark`；不调用 `.eyebrow` 也不调用 `bus/Eyebrow.tsx` | Console 端目前**几乎不用 eyebrow**：仅 ConsoleLayout.tsx:151 admin 分组标题用了 `text-[11px] font-mono uppercase tracking-[0.14em]`（**0.14em ≠ token 的 0.18em**） | **三套 eyebrow 并存**：bus/Eyebrow（标准）、Landing 内嵌、Console admin 标签自创 |
| 圆角 | `r-sm=4 / r-md=6 / r-lg=10 / r-xl=14`；rounded-* utility | Landing：pill `borderRadius: 999` 主导；卡片 8/12/14 写死；input mock 6 写死 | Card .card → `--r-xl=14`；btn → 6；btn-pill → 999；侧栏图标块 `rounded-lg=10`；StatCard quick-actions `rounded-lg` | 值范围一致，**Landing 写死数字 vs Console 走 token**——一旦 token 改 14→16，Landing 不会跟随 |
| 阴影 | `shadow-soft / card / elevated / glow`（含 shadow-glow 橙光） | Hero mock 卡片 `boxShadow: '0 30px 80px rgba(0,0,0,0.5)'`（Landing.tsx:501 = `shadow-elevated` 但写死值） | Card 默认 .card 不带 shadow；只在 hover 卡片用 `card-hover` | **数值同源，Landing 不用 shadow-glow 橙光（CTA 没有 token-level emphasis）** |
| 动效 | `pulseDot`（1.6s pulse） + transition 默认 0.15s | Landing：几乎没动效；唯一是 SVG halftone + plasma static 渐变 | btn / input / nav-link `transition-all duration-150` 一致（index.css 76、`transition-colors` ConsoleLayout） | **断层**：Landing 静态电影感，Console 微交互；中间没有桥（没有 hover 提升、没有微动） |
| 间距节奏 | tailwind 默认 + 自定义 `padding: '0 40px'` 容器 | Landing 内联 `padding: '0 40px'` + `paddingTop/Bottom: 80/100/120/180` 直接写死 | Console `p-5 sm:p-8`（ConsoleLayout main）；Card `p-5`；StatCard 间 `gap-4` | **节奏不同源**：Landing 偏"展会大屏"（120/80/60），Console 偏"工作台密度"（20/16） |

**要点**：token 层是 100% 同源的（Plato 暗色 + 橙色），断层在 **消费方式**——Landing 走"内联 style + 写死数值"路线，Console/Dashboard 走"tailwind class + .btn/.card/.input/.eyebrow 全局原语"路线。**修这层不必动 token，只需要把 Landing 的内联 style 替换成 token 类**。

---

## 2. 组件库分歧

### 2.1 已有的标准库（应作为唯一来源）

- `src/components/ui/`：`Button.tsx`、`Input.tsx`、`Card.tsx`、`Badge.tsx`、`Modal.tsx`、`Skeleton.tsx`、`Spinner.tsx`、`Table.tsx`、`Toast.tsx`
- `src/components/bus/`：`Eyebrow.tsx`（标准 .eyebrow 包装）、`PillBtn.tsx` & `PillLink`（pill-shaped CTA，含 accent/light/ghost × sm/md/lg + loading）、`Wordmark.tsx` & `BusMark`（菱形 logo）、`SectionFrame.tsx`（hairline 边框 + 十字 marker，**这是 Landing 视觉签名**）、`PlasmaBlob.tsx` & `HalftoneOverlay`（橙色等离子装饰）、`Sparkline.tsx`（橙色 sparkline）

### 2.2 Landing 自带的"重复实现"

| 在 `Landing.tsx` 内私有定义 | 已有等价标准组件 | 行号 | 影响 |
|---|---|---|---|
| `OrangeMark` | `bus/Eyebrow` 内置 marker（.eyebrow::before） | 14-28 | 三处重复 marker |
| 私有 `Eyebrow` | `bus/Eyebrow.tsx` | 30-47 | 字号 11.5 vs 标准 11；color rgba(255,255,255,0.7) vs token --text-3=#8a8f99；前者更亮 |
| 私有 `PillBtn` | `bus/PillBtn.tsx`（accent/light/ghost × sm/md/lg） | 49-76 | Landing 私版只支持 primary/ghost 两态，且把 `<button>` 嵌在 `<Link>` 内（HTML 嵌套警告） |
| `NavD` 顶部 logo（写死菱形 SVG） | `bus/Wordmark.tsx` & `BusMark` | 111-122 | 同一个菱形 logo 写两遍，size/letter-spacing 略有差 |
| `CtaFooter` 底部 logo（再次内联菱形） | 同上 | 1148-1166 | 同一份代码再 copy 一份 |
| 内联 nav pill 容器 | `.pill-nav` + `.pill-nav-item`（index.css） | 125-150 | 全局 .pill-nav 已有 light-active 变体，但 Landing 没用 |
| Hero badge `padding: '6px 14px 6px 6px'...` | 没有标准 chip，但与 `Badge tone="accent"` 形态一致 | 213-240 | 应抽 `<Chip>` 或扩展 Badge |
| 私有"backed by"行内 brand list | 无标准等价 | 336-381 | 一次性，可保留 |

### 2.3 Console 端的"标准库消费"

- ConsoleLayout.tsx 几乎全部走 tailwind class + .btn 全局类（btn-ghost btn-icon），sidebar 激活态用 `bg-orange-soft text-orange`。
- Dashboard.tsx 全部 `<Card className="p-5">` + `font-display text-3xl text-ink-1`。
- Keys.tsx：`<Button>` `<Input>` `<Modal>` `<Badge>` `<Table>` 全部来自 `ui/*`，零内联 style。
- 但是 Console 端**完全不消费 bus/* 装饰原语**——没有 `<Eyebrow>` 引导、没有 PillBtn 大号 CTA、没有 SectionFrame hairline、没有 PlasmaBlob 任何装饰，导致 Console 视觉密度高、装饰为 0。

### 2.4 PageHeader 的孤儿地位

- `PageHeader` 定义在 `ConsoleLayout.tsx` 末尾（导出但耦合在布局文件里），用了 `font-display text-3xl`。
- Landing 的"section header"是用内联 h2 38–44 px 写死的，**没有抽 SectionHeader/PageHeader 共用**。

---

## 3. 布局栅格断层

| 维度 | Landing | Console | 全局 token | 一致？ |
|---|---|---|---|---|
| 容器最大宽 | `maxWidth: 1280` 内联（NavD/Hero/Operations/Security/CaseStudy/CtaFooter 全部 1280） | ConsoleLayout main 没设 max-width，**自适应填满侧栏右侧** | `.container-bus=1280` `.container-console=1440` | **不一致**：Console 从未走 .container-console；Landing 从未走 .container-bus |
| 横向 padding | 内联 `padding: '0 40px'` | `p-5 sm:p-8`（20→32 px，移动端更小） | 标准容器 `padding: 0 40px` | Landing 写死，Console 用响应式（更合理） |
| 断点 | 无（单一桌面布局，min-width 写死如 `gridTemplateColumns: '1.2fr 1fr 0.8fr'`） | `grid-cols-1 sm:grid-cols-2 lg:grid-cols-4`（响应式） | tailwind 默认 sm/md/lg/xl/2xl | **断层**：Landing 移动端会塞车 |
| Section 节奏 | section padding 上下 80/100/120 px，内部 `paddingTop: 80` `paddingBottom: 80` 多次 | 主区 padding `p-8` (32 px)，Card 间 `gap-4` (16 px)，PageHeader `mb-6` (24 px) | 无统一 section 节奏 token | 完全两个量级 |
| 装饰元素 | SectionFrame（hairline + 十字 marker）+ PlasmaBlob + HalftoneOverlay + dot-bg radial-gradient | 无 | dot-bg / grid-bg utility 已存在 | Console 端 0 装饰 |
| Header 高度 | NavD 自由布局 absolute top:24 | header h-16 (64 px) 固定 | 无 | 不同范式（marketing vs app） |
| 侧栏 | 无 | `w-60` (240 px) bg-bg-1 + border-line-2 | 无 | 不冲突 |

---

## 4. 关键页面位置定位（file:line 而非截图）

### Landing 视觉锚点（test/xlabapi:frontend-v2/）

- 顶部菱形 nav + pill nav + Sign-in：`src/pages/Landing.tsx:78-170` (`NavD`)
- Hero 84 px 大标题 + 等离子球 + 4 列统计：`src/pages/Landing.tsx:172-334` (`HeroD`)
- "Backed by" 灰色 brand 行：`src/pages/Landing.tsx:336-382` (`BackedRow`)
- Operations Hub mock 控制台卡（最像 dashboard 的展示位）：`src/pages/Landing.tsx:384-804` (`OperationsHub`)
- Security 圆形徽章：`src/pages/Landing.tsx:806-873` (`SecurityBand`)
- 客户证言：`src/pages/Landing.tsx:875-1052` (`CaseStudy`)
- 大 CTA + footer：`src/pages/Landing.tsx:1054-1223` (`CtaFooter`)
- 装饰：`src/components/bus/SectionFrame.tsx:1-50`、`src/components/bus/PlasmaBlob.tsx:1-43`

### Console 视觉锚点

- 侧栏 logo + 主导航：`src/components/layout/ConsoleLayout.tsx:118-180`
- Sidebar 激活态：`bg-orange-soft text-orange font-medium`，行号 `ConsoleLayout.tsx:135-141`、`174-179`
- Header（mobile menu + LocaleSwitcher）：`ConsoleLayout.tsx:194-205`
- Main padding：`ConsoleLayout.tsx:206`（`p-5 sm:p-8`）
- PageHeader（仅有的"装饰位"，display 字 + 描述）：`ConsoleLayout.tsx:217-228`
- StatCard 模式（Console 视觉签名）：`src/pages/user/Dashboard.tsx:13-24`
- Quick action 卡：`src/pages/user/Dashboard.tsx:107-149`（用 `card-hover` 但全局 css 中没有 `.card-hover` 定义，**找不到该类**——潜在 bug，等审计 A/C 一并验证）

### Console 与 Landing 的"接缝"

- Login 页：`src/components/layout/AuthLayout.tsx:14-44`，bg-bg-0 + 顶 logo + Card 居中——风格更靠近 Console；
- 但 Login 页的 logo 用的是 `siteLogo` 图片 + 圆角方块，**不用 bus/Wordmark 菱形**，与 Landing NavD 的菱形 logo 视觉断层；
- 用户从 Landing 点 "Sign in" → AuthLayout → 登录后 → ConsoleLayout，**Logo 系统切换 2 次**：菱形 → 图片 → 图片+siteName 文字（侧栏）。

---

## 5. 断层根因总结

1. **路线分裂**：Landing 是"prototype 直接落地"产物（文件头注释 `Landing v4 — Plato (Xlabapi dark + warm orange) Adapted from the frontend-v2 landing prototype.`），保留了原型的内联 style 写法；Console 是"标准化 React + tailwind"产物。两者**共享 token，不共享消费方式**。
2. **bus/* 与 ui/* 单向流**：bus/* 设计语言（pill / eyebrow / wordmark / sparkline / plasma / sectionframe）几乎只服务 Landing；ui/* 几乎只服务 Console。**双方没有打通**。
3. **PageHeader 缺失上层装饰**：Console 永远只有 `<h1>+<p>` 两行，缺一个像 Landing 的 `Eyebrow + display heading + accent rule` 三件套，**视觉权重差距巨大**。
4. **栅格容器不同源**：`.container-bus=1280` 与 `.container-console=1440` 已在 index.css 定义但**两侧都没用**（Landing 内联 1280；Console 不限宽）。
5. **Logo 三轨**：Landing 菱形（BusMark）、Auth/Console siteLogo 图片+文字、Login 圆角方块——同一品牌三套表达。

---

## 6. 统一方案建议（3 选 1）

### 方案 A：Landing 重构对齐到 Console（去内联 style）  ← **推荐**

**核心动作**（不写代码，只列工作量）：

1. 把 `Landing.tsx` 内私有的 `Eyebrow / PillBtn / OrangeMark / NavD logo` 替换为 `bus/Eyebrow` `bus/PillBtn` `bus/Wordmark`。删 4 处重复定义，约 -80 行。
2. 把所有内联 `style={{...}}` 中的颜色/字号/间距值替换为 `className`：`background:'#000'` → `bg-bg-0`；`rgba(255,255,255,0.10)` → `border-line-2`；`color: ORANGE` → `text-orange`；`fontFamily: 'Georgia, serif'` 强调字 → 抽 `<em className="em-serif">` 走 index.css `.em-serif`。
3. 写死字号 84/56/52/44/42/40/38 → 用 token：`text-display-2xl`(72) / `text-display-xl`(56) / `text-display-lg`(48) / `text-display-md`(40)；不在 token 里的中间值（如 84、64、52）**讨论是否扩 token** 还是降到 token 网格。
4. 容器 `maxWidth:1280; padding:'0 40px'` → 全部走 `.container-bus`。
5. Section 节奏 80/100/120 → 抽 `<Section padding="lg|xl|2xl">` 一层薄壳，落到固定枚举。
6. 引入响应式断点：`gridTemplateColumns: '1.2fr 1fr 0.8fr'` → tailwind grid + sm/md/lg。Landing 现在移动端基本会塞车，必须修。

**为什么选这个**：
- token 层 100% 复用现有 tailwind.config.ts 与 index.css，**零设计决策成本**。
- `bus/*` 标准件（PillBtn / Eyebrow / SectionFrame / Wordmark）已经按 token 写好，只是 Landing 没用——重构本质是"删重复"。
- 不动 Console 端任何视觉，**完全满足红线**（不动 dashboard 行为、不动样式）。
- 对 Console 的"装饰升级"（方案 C 的一部分）可以下一轮单独做，互不阻塞。

**工作量估计**：
- Landing.tsx：1244 行 → 预计删到 800–900 行（高密度重构，但纯前端、纯样式映射，**功能 0 风险**）。
- 需要扩展 bus/PillBtn 的 size/variant 到覆盖 Landing 当前所有按钮形态（约 +20 行）。
- 需要新增 `bus/Section.tsx` 薄壳（约 30 行）。
- 总工作量：1 名前端 1.5–2 天可完成，含视觉回归 review。
- 风险：**SectionFrame 的 halfWidth 写死 660，需配合容器 1280 校准**；不能简单替换。

### 方案 B：Console 升级对齐到 Landing（往大装饰收）

**核心动作**：
1. ConsoleLayout 主区改用 `.container-console=1440`，加上 SectionFrame hairline + 十字 marker，作为 Console 的视觉签名。
2. PageHeader 升级为"Eyebrow + display 标题 + 描述"三件套，配合 Landing 的橙方块 marker 节奏。
3. StatCard 加 `shadow-glow` 橙光强调主指标。
4. 各 admin 页加微 PlasmaBlob 装饰（弱化版 opacity 0.2）。

**为什么不推荐**：
- 触动近 30 个 Console/admin 页的样式，**违反"只动样式不动行为"红线难以保证**——一旦 PageHeader 容器尺寸变化，下游 Card grid 都会重排，回归面太大。
- Landing 的"展会感"（halftone、plasma、80/120 px section padding）在 dashboard 工作场景反而**降低信息密度**——dashboard 用户每天进来不是看艺术品，是看数据。
- 工作量估计 4–6 天 + 全量视觉回归。

### 方案 C：抽公共主题层（中间路线）

**核心动作**：
1. 新增 `src/components/theme/` 目录，把 Landing 与 Console 共用的 atoms 提到这一层：`Section`、`Eyebrow`、`PageTitle`（display 三件套）、`PillBtn`（合并 bus/PillBtn 与 Landing 私版）、`StatBig`（Landing 大数字 + 单位橙色那个组件）。
2. Landing 与 Console 都改吃 theme/* 而非各自的 bus/* / ui/*。
3. 保留 ui/*（Card/Input/Modal/Table 等"工具组件"，dashboard 专用）和 bus/*（Plasma/SectionFrame/Sparkline 等"marketing 装饰"，Landing 专用）。

**为什么不首选**：
- 抽象一层 = 两边都要改 + 多一层认知负担。
- 当前 bus/Eyebrow + bus/PillBtn 已经是"通用 atom"，不需要再抽一层。
- 在没有第三方页面（如 marketing 子页 / pricing 页）的情况下，**两层架构（bus + ui）已足够**，第三层（theme）属 YAGNI。
- 工作量估计 3–4 天。

---

## 7. 推荐路径（如方案 A 通过）

不动代码，仅作落地顺序提案，待总规划批准后另起任务执行：

1. **第一刀（破壳）**：把 Landing.tsx 私有的 4 个组件（OrangeMark / Eyebrow / PillBtn / NavD-Logo）删掉，统一引 bus/Wordmark + bus/Eyebrow + bus/PillBtn。这一刀让 Landing 与 Console 在"原子层"统一。
2. **第二刀（容器）**：抽 `<Section padding="lg|xl|2xl" frame>` 薄壳，把 Landing 7 个 section 全部走它。SectionFrame 的 halfWidth 跟随 Section 容器自动算。
3. **第三刀（token 化）**：内联 `style={{}}` 全部换成 className，颜色/字号/圆角全部走 token；保留只有 Landing 才有的渐变/halftone/plasma 内联（这些**应该**留内联，因为它们是装饰，不是设计 token）。
4. **第四刀（响应式）**：Landing 加 sm/md/lg 断点。
5. **不需要的**：Console 端不动；token 不动；ui/* 不动。

---

## 8. 红线复核

- [x] 不修改任何 Console / Dashboard / admin 页面的功能行为——本审计未提议改 Console 任何代码，仅 Landing。
- [x] 不提议"重写组件库"——bus/* 与 ui/* 完整保留，方案 A 只是把 Landing 私版定义合并到 bus/*，删的是重复，不是基建。
- [x] token 层 0 改动——tailwind.config.ts 与 index.css 不动。
- [x] 仅产出文档，未触碰任何 .tsx / .ts / .css 文件。

---

## 附录 A：测试 / xlabapi 分支 frontend-v2 全局 CSS 关键原语索引

- `.btn` `.btn-primary` `.btn-accent` `.btn-ghost` `.btn-danger` `.btn-sm` `.btn-lg` `.btn-icon` `.btn-pill`：index.css :69-122
- `.input` `.input-error` `.input-error-text` `.input-label`：index.css :125-156
- `.card` `.card-flat`：index.css :159-167
- `.badge` `.badge-success` `.badge-warning` `.badge-danger` `.badge-accent`：index.css :170-194
- `.eyebrow`（含橙方块 ::before）：index.css :197-211
- `.em-serif`：index.css :214-218
- `.container-bus`(1280) `.container-console`(1440)：index.css :221-230
- `.pill-nav` `.pill-nav-item` `.pill-nav-item-active` `.pill-nav-item-light-active`：index.css :233-263
- `.hr-fade`：index.css :266-276
- `.skeleton`：index.css :279-282
- `.data-table`：index.css :285-289
- `.dot-bg` `.grid-bg`：index.css :293-302

## 附录 B：bus/* 组件接口速查

- `bus/Eyebrow`：`{ children, className? }` → 渲染 `.eyebrow`
- `bus/PillBtn`：`{ variant?: 'accent'|'light'|'ghost', size?: 'sm'|'md'|'lg', loading?, ...buttonProps }` + `PillLink`（同 props，渲染 `<a>`）
- `bus/Wordmark`：`{ small?, name?, className? }`；`BusMark`：`{ size?, className? }`
- `bus/SectionFrame`：`{ showTop?, showBottom?, halfWidth=660 }`
- `bus/PlasmaBlob`：`{ style? }`；`HalftoneOverlay`：`{ opacity? }`
- `bus/Sparkline`：`{ data: number[], color?, height? }`

## 附录 C：与老前端 frontend/（origin/xlabapi）对照（仅供理解用户语义，**不**作为视觉对齐目标）

- 老前端是青色 teal-500 (`#14b8a6`) + glass shadow + mesh-gradient + dark mode toggle (`darkMode: 'class'`)，与 frontend-v2 Plato 暗色 + 橙色**完全不同源**。
- 用户说"console 完全参考 xlabapi 布局"应理解为**结构参考**（侧栏 + 主区 + 顶 header），不是色板参考。色板必须留在 Plato 暗色 + 橙色。
- 因此本审计聚焦 frontend-v2 内部 Landing ↔ Console，老前端不作为 baseline。

