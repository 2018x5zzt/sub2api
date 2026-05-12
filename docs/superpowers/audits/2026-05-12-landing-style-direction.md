# Landing Style Direction — 基于用户截图 + 新 logo 的方向反推

- 审计目标：基于用户三张实际参考截图 + 新 logo（橙黑多面体几何），反推 frontend-v2 Landing 的可行风格方向；列出与 Plato 现有 token 的 gap、Hero 字号决策、装饰强度清单、Console 边界；**不写代码，不立刻替换 logo**，只给影响分析与取舍。
- 输入源：
  - 截图 1（暖橙深色 SaaS Landing）：`state/blobs/7f66dde7…_8a2126107dd1e264.png`
  - 截图 2（紫粉调大字 + 几何 block）：`state/blobs/12ae069b…_2df20e5eb6200e81.png`
  - 截图 3（浅底米白色 + 粗几何拼贴）：`state/blobs/45daee7a…_03e8c7f47157c6a3.png`
  - 新 logo（橙黑多面体几何）：`state/blobs/be89b9f1…_image.png`
  - 现有 token：`frontend-v2/src/index.css` + `frontend-v2/tailwind.config.ts`
  - 现有 Landing：`frontend-v2/src/pages/Landing.tsx`
  - 风格 gap 审计：`docs/superpowers/audits/2026-05-12-frontend-v2-style-gap.md`（方案 A 已选）
- 红线：橙 #FF5722 锁定；Console 不动；仅读代码不写代码；**不建议立刻替换 logo**；看结构语言而非配色抄袭；新 logo 替换范围仅做影响分析。

---

## 1. 三张截图的风格特征对比表

| 维度 | 截图 1（暖橙深色 SaaS） | 截图 2（紫粉大字） | 截图 3（浅底几何拼贴） |
|---|---|---|---|
| **Type scale（Hero 主字号目测）** | 72–88 px，字重 semi-bold 500–600，字距 ≈ -0.025em；副标题 14–16 px 灰；eyebrow mono 11 px | 88–120 px，字重 medium 400–500，字距 -0.03em；大字占比极高；副文本极少 | 约 100–130 px，字重 light/regular 300–400，极紧字距 ≈ -0.035em；大字 + 大量留白 |
| **容器与间距节奏** | 容器 ≈ 1200–1280 px；section padding 上下 60–80 px；cards 间 gap 24 px；**紧凑、工业** | 容器 ≈ 1200 px；section padding 80–120 px；几何 block 之间留白大；**慢节奏、艺术感** | 容器 ≈ 1280–1440 px；section padding 100–140 px；大量留白；**美术馆式展示** |
| **装饰语言** | 极克制：规则网格细线 + 小 chip/badge + 小橙色标记块；**无光晕、无 plasma** | 多色几何 block（方块、斜切拼贴、弧形）；深紫渐变背景作衬底；**装饰密度中等偏高** | 粗线条对角分割 + 大色块几何 + 有机拼贴；**装饰密度最高**；视觉优先级 > 信息密度 |
| **信息密度** | 高：同屏 hero + sub + CTA + 3–4 个 stat 数据点 | 低：一屏只放 hero 大字 + 一个 CTA | 极低：一屏只放 hero + 1 个视觉图形；纯艺术展示 |
| **与 sub2api 视觉同源性** | **最高**（暗 + 橙 + 网格 + 克制装饰 + 工业感） | 中（同样暗色 + 大字）但色系对立（紫粉 vs 橙） | 低（浅底色 + 高饱和彩色，与 Plato 暗黑橙完全不同轴） |
| **可直接借鉴层** | type scale ≈ 匹配 Plato；装饰密度匹配；容器/节奏直接参考 | 只借鉴"大字主导 + 几何 block 装饰"的思路，**不借鉴色板** | 只借鉴"粗几何切面的装饰语汇"，**不借鉴色板和信息密度** |

---

## 2. 与 Plato 当前 token 的 gap

以截图 1 作为**主参考**（视觉同源最高），对照 `tailwind.config.ts` + `index.css`：

### 2.1 Token 层：完全够用 ✅

| 维度 | Plato 现有 token | 截图 1 实际需求 | 是否需要新增 |
|---|---|---|---|
| 主色 | `--orange=#ff5722` + `.orange` | 暖橙小块高亮 | **不需要** |
| 背景层 | `--bg-0..3` | 纯黑 + 细微深层 | **不需要** |
| 边线/网格 | `--line-1..4` + `.grid-bg` utility | 规则细网格 | **不需要**（已有 `.grid-bg`） |
| 文本阶梯 | `--text-1..4` | #fff + rgba(.55/.7/.5) | **不需要** |
| mono caption | `.eyebrow`（mono 11px 0.18em） | mono chip/badge | **不需要** |
| 圆角/阴影 | `--r-sm..xl` / `shadow-card/elev/glow` | 6/8/12 圆角 | **不需要** |

**结论**：`token 100% 可用`，不需要扩。

### 2.2 写死值层：大量需重构（这是 gap B，source = 风格审计 §2.2 已列）

Landing.tsx 内联了 80+ 处写死值（背景色、边线色、字号、圆角、间距），应全部换成 `className + token`。

### 2.3 字号层：**这才是真正的 gap**

| Landing.tsx 实际 | Plato token | 截图 1 目测 | 是否需要新增 |
|---|---|---|---|
| Hero h1 `fontSize: 84` | `text-display-2xl=72` ❌ 不匹配 | 72–88 可达 | **讨论**（见 §3） |
| Stats 大数字 `56 / 64` | `display-xl=56` ✅ | 52–60 | 勉强够用 |
| CaseStudy h2 `32` | `display-md=40` / 无 32 | 28–36 | 可用 text-3xl(30)/4xl(36) 近似 |
| SecurityBand h2 `44` | `display-md=40` | 40–48 | 可用 40 或扩 44 |
| CtaFooter h2 `40` | `display-md=40` ✅ | 40–44 | 匹配 |

**核心决策点**：Hero 字号 84 是否属于"合理的 marketing 强调"还是"滥用的写死值"？见下一节。

---

## 3. Hero 大字号决策（核心问题）

### 路径 A：Hero 字号降到现有 token 网格内（最稳）

- Hero h1 → `text-display-2xl = 72px`
- OperationsHub h2 → `text-display-xl = 56px` 或 `text-display-lg = 48px`
- SecurityBand h2 → `text-display-md = 40px`（从 44 降 4px）
- CaseStudy / CtaFooter → `display-md = 40`

**利**：
- token 0 改动，与 Console 共享同一套 token 基础网格（`display-md=40` 就是 Console `PageHeader` 理论上可用的顶级标题，虽现在 Console 只用到 30px = `text-3xl`）
- 修改成本最低，未来维护一致
- 对用户反馈的"字号不要写死"最直接回应

**弊**：
- Hero 从 84 降到 72，视觉体量**略有损失**（约 14% 视觉面积）
- 截图 1 的 Hero 体量大约是 72–88，72 属于该区间下沿，**不会太小，但也不够冲击**
- 截图 2/3 的 Hero 120 更不可能在此路径达成

### 路径 B：新增"Marketing 强调层"token（推荐，对齐用户原意）

在 tailwind.config.ts 中新增一层 marketing-only 字号，**仅 Landing 使用，不污染 Console**：

```
hero-2xl: 96px  // 截图 2/3 级别极端冲击
hero-xl:  84px  // 截图 1 级别常规 SaaS Hero
hero-lg:  64px  // section 大 eyebrow
```

Console 端依然严格限定在现有 `display-2xl..md (72/56/48/40)` + `text-3xl..xs`。

**利**：
- Console 信息密度保留，不被"营销化大字"污染
- Landing 拥有合法的"营销字号"，不再是写死值，而是有 token 支撑
- 三张截图的 Hero 字号都可表达：截图 1 用 hero-xl(84)，截图 2/3 用 hero-2xl(96) 近似
- 解决了用户"降字号损失冲击力"与"扩 token 引入散乱"的死线 —— **只在 Landing 可见的命名空间里扩 token**
- 名称 prefix `hero-*` 形成天然隔离，组件库 review 时一眼看出"只能用于 Landing"

**弊**：
- token 需要扩 3 个值（成本可控）
- 需要规约：`hero-*` 系列严禁出现在 Console / ui 组件库中（可通过 eslint 或 code review 约束）
- Landing.tsx 里 `fontSize:84` 换成 `className="text-hero-xl"`，是替换工作量但不是字号决策工作量

### §3 推荐

**推荐路径 B**。理由：
1. 用户明确说"不想降字号损失冲击力，也不想扩 token 污染基础网格"；B 路径用命名空间隔离解决了这个死线。
2. 截图 1 是 Plato 视觉同源最高的参考，其 Hero 84 恰好落在 `hero-xl`，这说明 84 不是"写死的异常值"而是"marketing 层合法字号"。
3. 与方案 A（Landing 重构对齐 Console）**完全不冲突**——A 是消费方式的重构，B 是字号命名空间的补充，两者合并后 Landing 既干净又有表现力。
4. 即使日后 Landing 风格演进到截图 2/3 的极端冲击方向（`hero-2xl=96`），token 也预留好了接口，不需要二次扩表。
5. 命名前缀 `hero-*` 让组件库约束可执行：review 时一眼识别；未来 Console 若误用，审计脚本可扫描。

**落地时机**：S3 实施方案 A 时同步引入 `hero-*` 系列，不单开 task。

---

## 4. 装饰强度建议（Landing 各 section 逐一）

用户已答：**Landing 保留少量 Plasma/halftone 作品牌签名，但降密度**。以下给出 Landing.tsx 每个 section 的"保留/削减/移除"清单。

### 4.1 原则

- **保留**：只在 Hero 主视觉区、CaseStudy 大图区（强装饰合法位）保留 plasma
- **削减**：OperationsHub 原有 2 个 plasma → 减到 1 个；SecurityBand 取消 plasma
- **移除**：dot-bg 全局背景可视化密度 → 降到 2/3 甚至更低（目前 `radial-gradient size 32px` 偏密）
- **替换**：部分散点装饰可考虑用 **logo 几何家族**（见 §4.3）替代 plasma，保持品牌一致性

### 4.2 Landing.tsx 逐 section 清单

| Section (Landing.tsx 行号) | 当前装饰 | 建议（降密度） |
|---|---|---|
| `NavD` (78-170) | 内联菱形 logo + pill nav 背景 rgba(255,255,255,0.04) | **保留**，但 logo 替换为 `bus/Wordmark` 或新 logo（见 §5 取舍） |
| `HeroD` (172-334) | PlasmaBlob + HalftoneOverlay opacity=0.85 + dot-bg + SectionFrame | **保留** plasma + halftone（品牌签名），但降 halftone opacity 从 0.85 → 0.6；dot-bg 间距 32px → 48px（降密度） |
| `BackedRow` (336-382) | 纯文本 logo 列表 + SectionFrame | **保留**，极克制 |
| `OperationsHub` (384-804) | PlasmaBlob #1（右侧 540px 高装饰图）+ PlasmaBlob #2（可选，目前只有一个）+ mock console 卡 + HalftoneOverlay 0.7 + capabilities 列表 | **削减**：保留 1 个 plasma 作为中央视觉锚点；HalftoneOverlay opacity 0.7 → 0.5；mock console 卡保留（这是 Landing 的"内容锚"不是装饰）|
| `SecurityBand` (806-873) | 纯文本 + 2 个圆形徽章 | **保留**，无需 plasma |
| `CaseStudy` (875-1052) | 纯彩色渐变 bg + HalftoneOverlay 0.4 + 3 张渐变卡 | **保留**，渐变是 case 卡的"身份标识"，不是装饰污染 |
| `CtaFooter` (1054-1223) | SectionFrame | **保留**，克制 |

**dot-bg 全局**：目前 `.dot-bg` utility 定义在 `index.css` 中（`radial-gradient rgba(255,255,255,0.06) size 32px`）。建议：
- Hero 保留 32px（强装饰区）
- 其他 section 的 dot-bg 若使用，size 提到 48–56px，opacity 保持 0.06

### 4.3 Logo 几何家族作 Landing 品牌签名（候选替代方案）

新 logo 为**橙黑多面体几何**，色板与 Plato 100% 同源，装饰语言是"硬边直线 + 斜切 + 菱形"，与 PlasmaBlob 的"柔软光晕"在美学系统上属于不同方言。可考虑用 logo 衍生几何元素替代部分 plasma 散点装饰。

#### 适合位置（密度低、品牌感强）

- **Hero accent**（Landing.tsx:213-240 的 badge chip 旁）：在 "Plato Plus" chip 附近放一个小型多面体 mark（28–40px），作 hero 视觉辅助，替换 plasma 右下的一处辅助光晕
- **Section dividers**（各 section 之间）：用一个 16–24px 的迷你多面体作分段锚点，替换 hr-fade 或补充在 SectionFrame 十字 marker 的中心
- **Footer mark**（Landing.tsx:1148-1166 CtaFooter 当前菱形 logo）：**已经是菱形**，若替换成新 logo 是最自然的位置
- **CTA corner mark**（CtaFooter 右上或左下）：40–48px 的新 logo 作 CTA 区视觉锚

#### 不适合位置（避免过度堆叠）

- **OperationsHub mock console 卡**：已有橙色 marker 小方块 + 渐变 plasma，再叠 logo 几何会视觉打架
- **CaseStudy 渐变卡**：渐变卡自己是 case 身份标识，叠 logo 几何会破坏渐变氛围
- **Hero 大标题区**：logo 是"辅助视觉"，不是"标题伴生物"，放在 h1 旁会喧宾夺主
- **全局 dot-bg 替代**：logo 几何作为**点状**装饰会太锐利，反而不如 dot-bg 自然

### 4.4 降密度的量化建议

| 装饰元素 | 当前密度 | 建议密度 | 衡量指标 |
|---|---|---|---|
| PlasmaBlob 数量（全页） | 3 次实例 | 1–2 次实例 | 数量 |
| HalftoneOverlay opacity | 0.85 / 0.7 / 0.4 / 0.3 | 0.6 / 0.5 / 0.3 / 0.2 | 每 section -0.1 到 -0.2 |
| dot-bg 间距 | 32px（密） | 非 Hero 48–56px | 间距 +50% |
| SectionFrame 十字 marker | 所有 section | 保留 | 不动（视觉签名） |
| logo 几何元素（新增） | 0 | 3–5 处小型辅助 | 总数受控 |

---

## 5. 与 Console 的边界

### 5.1 原子组件应统一（Landing ↔ Console 共享）

- `bus/Eyebrow`：Landing 私版应替换为 bus 版（已在风格审计 §6 方案 A 提及）
- `bus/PillBtn`：Landing 私版应替换为 bus 版；同时扩展 variant/size 以覆盖 Landing 所有按钮形态
- `bus/Wordmark` + `bus/BusMark`：作为唯一品牌 logo 原语（见 5.3 取舍）
- `.btn / .input / .card / .badge / .eyebrow / .container-bus / .container-console`：index.css 中已定义，两侧都应消费

### 5.2 Landing 私有装饰保留（品牌签名，Console 端不复用）

- `bus/PlasmaBlob` + `HalftoneOverlay`：Landing-only 的品牌签名，Console 不复用
- `bus/SectionFrame`：Landing-only 的 hairline + 十字 marker，Console 不复用
- `bus/Sparkline`：Landing mock console 卡内的装饰，Console 如需类似图表应单独走标准图表库（recharts 等），不共用

### 5.3 Logo 三轨合一（用户已拍板 = 方案 A：全站统一新 logo）

#### 现状三轨（合一前）

| 位置 | 当前 logo | 行号定位 | 问题 |
|---|---|---|---|
| Landing NavD | 内联菱形 SVG（写死 28px，1.5px stroke，两层嵌套） | `Landing.tsx:111-122` | 写死值，不复用 `bus/Wordmark` |
| Landing CtaFooter | 同款菱形 SVG | `Landing.tsx:1148-1166` | 重复写死 |
| AuthLayout | `siteLogo`（`publicSettings.site_logo`）图片 + 圆角方块容器 + siteName 文字 | `AuthLayout.tsx:14-23` | 图片源 + 文字标 |
| ConsoleLayout | `siteLogo` 图片 + siteName 文字 | `ConsoleLayout.tsx:123-127` | 同上 |
| bus/Wordmark | BusMark（菱形 SVG） + 文字组件，标准化 | `bus/Wordmark.tsx` | 已标准化但**未被任何页面消费** |

#### 目标架构

- **唯一 logo 入口**：`bus/Wordmark`（含 `BusMark` 作为内部几何 atom）
- **三处落点**全部消费 `bus/Wordmark`，不允许私有菱形或自行嵌入图片
- **内部资产**：`BusMark` 原来绘制两层嵌套菱形的 DOM 结构，替换为**新 logo 资产**（svg/png，见下文）；保留 `size` prop 做尺寸适配；颜色锁 `#FF5722` + `#000`
- **`publicSettings.site_logo`**：后续若用户需要白牌，可作为 `Wordmark` 的可选 override（给 `logoSrc` prop），默认为 undefined 即走新 logo，非阻塞

#### 三处替换动作清单（S3 落地用，不写代码）

**1) `Landing.tsx` NavD + CtaFooter**
- 行号：`111-122`（NavD 顶部 logo）+ `1148-1166`（CtaFooter 底部 logo）
- 动作：删除内联菱形 SVG（4 个 absolute 定位 div）；改用 `<Wordmark size="lg" name="XLABAPI" />` 调用
- Prop 建议：
  - NavD 顶部：`size="lg"`（BusMark 26–28px + 14.5px 文字）
  - CtaFooter 底部：`size="md"`（BusMark 22–24px + 14px 文字）
- 注意：Landing 两处目前 `letterSpacing` + `fontWeight` 写死，`bus/Wordmark` 已封装；替换后统一，无需再调

**2) `AuthLayout.tsx` 顶部 header logo**
- 行号：`14-23`
- 动作：删除 `h-9 w-9 rounded-lg bg-bg-1 border border-line-2` 容器 + `<img src={siteLogo}>` + `<span>{siteName}</span>` 组合；改用 `<Wordmark size="md" name={siteName} />`（保留 siteName 作为文字 prop，支持租户定制文字）
- Prop 建议：`size="md"`（BusMark 24–28px）
- 注意：Auth 页登录前用户未登录，`publicSettings` 可能未加载完毕；新 logo 作为默认不依赖后端加载，ls UX 更流畅

**3) `ConsoleLayout.tsx` sidebar 顶部 logo**
- 行号：`123-127`
- 动作：删除 `h-8 w-8 rounded-lg bg-bg-3 border border-line-2` 容器 + `<img src={siteLogo}>` + `<span>{siteName}</span>` 组合；改用 `<Wordmark size="sm" name={siteName} />`
- Prop 建议：`size="sm"`（BusMark 18–22px，侧栏较窄）
- 注意：Console sidebar 目前宽度 `w-60=240px`，减去左右 padding 后 logo 区实际宽 ≈ 200px；新 logo 在 18–22px 下的可读性需要验证（多面体几何在小尺寸下细节可能丢失，**建议 svg 而非 png**）

#### 资产格式建议（待用户提供）

| 格式 | 利 | 弊 | 推荐度 |
|---|---|---|---|
| **SVG（矢量）** | 无限缩放不失真；可通过 CSS 或 JS 换色；文件体积小；打包后可内联；不依赖 base64 | 如果几何复杂（多面体渐变），SVG 文件可能比 PNG 大；需要清理 Figma/AI 导出的冗余元数据 | **强烈推荐** |
| **PNG（光栅）** | 绘制效果与设计稿 100% 一致；浏览器兼容最好 | 多尺寸需要 @1x/@2x/@3x 多套；换色需重新导出；sidebar 18–22px 下可能模糊 | 备选（仅当 SVG 复杂度过高时） |

**建议用户提供**：
1. 优先：一份干净的 SVG（清理所有 Figma 元数据，保留 `viewBox`、颜色用 `currentColor` 或 `#FF5722` + `#000000`）
2. 次选：三套 PNG（96px / 48px / 24px，对应 Landing nav / Auth / Console sidebar 尺寸）
3. 附加：一份 `favicon.ico`（16/32/48px 合并，浏览器标签用）

#### 尺寸适配与色板保留

| 落点 | 尺寸（w×h） | 色板 | 备注 |
|---|---|---|---|
| Landing NavD 顶部 | 28–32 px | `#FF5722` + `#000` | h1 附近，需要保持视觉重量 |
| Landing CtaFooter 底部 | 22–26 px | `#FF5722` + `#000` | 与 footer 文字同高 |
| AuthLayout header | 26–30 px | `#FF5722` + `#000` | login 居中页，视觉锚点 |
| ConsoleLayout sidebar | 18–22 px | `#FF5722` + `#000` | **小尺寸核心挑战**，建议 svg 保证锐利 |
| favicon（浏览器标签） | 16 / 32 px | `#FF5722` + `#000` | 极小尺寸，几何需简化版本 |

**色板红线**：橙 `#FF5722` 已用户锁定；黑色基底 `#000`；**不允许**引入第三色（除非作为 hover/active 高亮的 `--orange-hover=#ff693a`）。

#### S3 落地拆分（承接风格审计 §7 "第一刀 破壳"）

在方案 A 的 4 刀架构中，logo 三轨合一**归到第一刀同步完成**：
- 第一刀原计划：删 Landing 私有 `OrangeMark / Eyebrow / PillBtn / NavD-Logo` 共 4 件，统一引 `bus/Wordmark` + `bus/Eyebrow` + `bus/PillBtn`
- **追加动作**：
  - 1a. 替换 `bus/BusMark` 内部资产为新 logo svg（用户提供后）
  - 1b. 替换 `AuthLayout.tsx:14-23` 为 `<Wordmark size="md" />`
  - 1c. 替换 `ConsoleLayout.tsx:123-127` 为 `<Wordmark size="sm" />`
  - 1d. favicon 同步更新（`public/favicon.ico` 或 `index.html` `<link rel="icon">`）
- **工作量增幅**：约 +0.5 天（主要是 svg 清理 + 多尺寸视觉验证），总第一刀工作量从 1.5–2 天 → 2–2.5 天
- **关键风险**：Console sidebar 18–22px 下的 logo 锐利度——这是 S3 必须视觉验证的点

---

## 6. 落地路线（S3 细化）

承接风格审计 §7（方案 A 的 4 刀），结合本次方向补充每刀的"输入参考点"。**不是工作量表，不是执行计划**，只是方向澄清供 S3 规划时细化。

### 第一刀（破壳）：删 Landing 私有定义，引 bus/* 原子
- **输入参考**：§5.1 原子组件统一清单
- **涉及 Landing 段**：NavD、Eyebrow、PillBtn、CtaFooter logo 段
- **可能同步动作**：引入新 logo 相关决策（§5.3 A/B/C 三选一）

### 第二刀（容器）：抽 `<Section padding frame>` 薄壳
- **输入参考**：截图 1 section 节奏 60–80 px；截图 2 的 80–120 px
- **Section 节奏预设**：sm=48, md=64, lg=80, xl=120（可讨论）

### 第三刀（token 化 + 字号决策）：内联 style 全部换 className + 字号落到 token
- **输入参考**：§3 推荐路径 B，扩 `hero-2xl/xl/lg`；其余字号严格落到现有 `display-*` / `text-*`
- **涉及 Landing 段**：全部 7 个 section
- **装饰强度**：§4.2 清单落地

### 第四刀（响应式）：加 sm/md/lg 断点
- **输入参考**：截图 1 的一屏紧凑布局可作为 `lg+` 基准；`md` 重排；`sm` 简化到单列
- **关键断点**：Hero stats 四列 → 二列 → 单列；OperationsHub 并列 1fr 280px → 垂直堆叠

---

## 7. 红线复核

- [x] 未触碰任何 .tsx / .ts / .css 代码文件
- [x] 未改动 `--orange=#ff5722`
- [x] 未改动 Console 端任何视觉
- [x] 未提议立即替换 logo；仅做 A/B/C 影响分析
- [x] 未抄截图配色；只反推结构语言（type scale / 间距 / 装饰强度 / 字号分层）
- [x] §3 Hero 字号给了明确推荐（路径 B），用户可一眼决策
- [x] §4.2 装饰削减清单具体到 Landing.tsx section
- [x] §5.3 logo 三轨 A/B/C 未给推荐（按领班要求），供用户拍板

---

## 附录 A：三张截图可直接借鉴的最小集

- **截图 1**：type scale（72–88 Hero）、dot-bg 网格密度、mock console 卡的"内容锚 + 克制装饰"；Plato 最贴近同源参考
- **截图 2**：Hero 大字占屏比例策略（一屏只放一件事）、section 间留白节奏；**色板不借鉴**
- **截图 3**：粗几何拼贴可作 CaseStudy 或 Testimonial 的艺术化升级参考；**色板/信息密度不借鉴**

## 附录 B：新 logo 几何家族可衍生的装饰元素（思路，不作设计交付）

- 单体多面体（原 logo 尺寸）
- 简化菱形单元（保留 logo 一个切面）
- 等距多面体阵列（3–5 个成排）
- 半透明多面体水印（opacity 0.1–0.2，作 section 背景辅助）
- 轮廓线版本（只保留硬边线，填充透明，作极克制 accent）

