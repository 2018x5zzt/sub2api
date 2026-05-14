import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LocaleSwitcher } from '@/components/layout/LocaleSwitcher'

const ORANGE = '#ff5722'
const GREEN = '#22c55e'
const SKY = '#38bdf8'
const VIOLET = '#a78bfa'
const INK = '#191715'
const PAPER = '#f4f0e7'
const UI_FONT = '"PingFang SC", "Microsoft YaHei", "Noto Sans SC", "Inter", system-ui, sans-serif'

const heroEndpointRows = [
  { name: 'Claude', endpoint: '消息模型', color: ORANGE },
  { name: 'OpenAI', endpoint: '响应模型', color: GREEN },
  { name: 'Gemini', endpoint: '原生模型', color: SKY },
  { name: 'Antigravity', endpoint: '代码智能体', color: VIOLET }
] as const

const modelLaunchpadRows = [
  ['Claude', '高上下文推理', '稳定接入', ORANGE],
  ['OpenAI', '响应式接口', '快速切换', GREEN],
  ['Gemini', '原生模型族', '多模态支持', SKY],
  ['Antigravity', '代码智能体', '研发场景', VIOLET]
] as const

const quickStartRows = [
  ['密钥接入', 'API Key'],
  ['用量追踪', '用量'],
  ['余额计费', '计费']
] as const

const heroSummaryStats = [
  ['OAuth', '+ API Key', 'landing.hero.stats.accounts', 'landing.hero.statDetails.accounts'],
  ['Token', '', 'landing.hero.stats.billing', 'landing.hero.statDetails.billing'],
  ['RPM', '+ TPM', 'landing.hero.stats.limits', 'landing.hero.statDetails.limits'],
  ['Ops', '', 'landing.hero.stats.observability', 'landing.hero.statDetails.observability']
] as const

const operationsModules = [
  ['landing.operations.modules.accounts.title', 'landing.operations.modules.accounts.description', '/api/v1/admin/accounts', ORANGE],
  ['landing.operations.modules.groups.title', 'landing.operations.modules.groups.description', '/api/v1/admin/groups', GREEN],
  ['landing.operations.modules.commerce.title', 'landing.operations.modules.commerce.description', '/api/v1/payment/*', SKY],
  ['landing.operations.modules.maintenance.title', 'landing.operations.modules.maintenance.description', '/api/v1/admin/backups', VIOLET]
] as const

const enterpriseStackRows = [
  ['订阅转 API', '把上游订阅、账号池和用户 Key 统一成可分发的 API 能力。', '订阅 / API', ORANGE],
  ['统一身份', '支持 OIDC、GitHub、JWT 与管理员后台权限控制。', '登录 / 权限', GREEN],
  ['分组调度', '按模型倍率、RPM/TPM、账号状态和兜底分组控制请求。', '策略 / 限流', SKY],
  ['运维治理', '用量、错误、备份、日志和告警集中在一个运营面板。', '监控 / 备份', VIOLET]
] as const

const operationsSignalRows = [
  ['模型接入', 'Claude / OpenAI / Gemini'],
  ['商业化', '套餐 / 订单 / 兑换码'],
  ['管理后台', '账号 / 分组 / 运维']
] as const

function LogoMark({ dark = false }: { dark?: boolean }) {
  const color = dark ? '#ffffff' : INK

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12, color, fontSize: 15, fontWeight: 700 }}>
      <div style={{ width: 28, height: 28, position: 'relative' }}>
        <div
          style={{
            position: 'absolute',
            inset: 0,
            border: `1.5px solid ${color}`,
            transform: 'rotate(45deg)',
            borderRadius: 3
          }}
        />
        <div style={{ position: 'absolute', inset: 7, border: `1.5px solid ${color}`, borderRadius: 2 }} />
      </div>
      XLABAPI
    </div>
  )
}

function Eyebrow({ children, dark }: { children: ReactNode; dark?: boolean }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 18 }}>
      <span style={{ width: 8, height: 8, background: ORANGE, flexShrink: 0 }} />
      <span
        className="font-mono"
        style={{
          fontSize: 11.5,
          letterSpacing: 0,
          color: dark ? 'rgba(255,255,255,0.62)' : 'var(--text-3)',
          textTransform: 'uppercase'
        }}
      >
        {children}
      </span>
    </div>
  )
}

function PillBtn({ children, primary, dark }: { children: ReactNode; primary?: boolean; dark?: boolean }) {
  return (
    <span
      className="landing-pill"
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: 44,
        padding: '0 22px',
        borderRadius: 999,
        fontSize: 14,
        fontWeight: 600,
        cursor: 'pointer',
        background: primary ? ORANGE : dark ? 'rgba(255,255,255,0.10)' : INK,
        color: primary || dark ? '#ffffff' : PAPER,
        letterSpacing: 0,
        whiteSpace: 'nowrap'
      }}
    >
      {children}
    </span>
  )
}

function NavD() {
  const { t } = useTranslation()
  const navItems = [
    ['#solutions', t('landing.nav.solutions')],
    ['#enterprise', t('landing.nav.enterprise')],
    ['#case-studies', t('landing.nav.caseStudies')],
    ['#company', t('landing.nav.company')]
  ] as const

  return (
    <nav
      className="landing-nav"
      aria-label={t('landing.nav.primary')}
      style={{
        position: 'absolute',
        top: 24,
        left: 0,
        right: 0,
        zIndex: 30,
        padding: '0 40px'
      }}
    >
      <div
        style={{
          maxWidth: 1280,
          margin: '0 auto',
          display: 'grid',
          gridTemplateColumns: '1fr auto 1fr',
          alignItems: 'center',
          gap: 20
        }}
      >
        <LogoMark />
        <div
          className="landing-nav-menu"
          style={{
            display: 'flex',
            gap: 4,
            padding: 5,
            borderRadius: 999,
            background: 'rgba(255,255,255,0.92)',
            boxShadow: '0 10px 30px rgba(35,25,10,0.08)'
          }}
        >
          {navItems.map(([href, label]) => (
            <a
              key={href}
              href={href}
              style={{
                padding: '8px 16px',
                fontSize: 13,
                color: 'var(--text-2)',
                textDecoration: 'none',
                borderRadius: 999
              }}
            >
              {label}
            </a>
          ))}
        </div>
        <div
          className="landing-nav-actions"
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            gap: 8,
            justifySelf: 'end',
            padding: 5,
            borderRadius: 999,
            background: 'rgba(255,255,255,0.94)',
            boxShadow: '0 10px 30px rgba(35,25,10,0.10)'
          }}
        >
          <a
            className="landing-doc-link"
            href="/docs"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: 44,
              padding: '0 16px',
              borderRadius: 999,
              fontSize: 13,
              fontWeight: 600,
              color: INK,
              textDecoration: 'none',
              background: '#ffffff'
            }}
          >
            {t('nav.docs')}
          </a>
          <LocaleSwitcher compact />
          <Link to="/login" style={{ textDecoration: 'none' }}>
            <PillBtn primary>{t('landing.nav.signIn')}</PillBtn>
          </Link>
        </div>
      </div>
    </nav>
  )
}

function ModelLaunchpad() {
  const { t } = useTranslation()

  return (
    <div
      className="landing-model-launchpad"
      style={{
        position: 'relative',
        width: '100%',
        borderRadius: 30,
        padding: 22,
        overflow: 'hidden',
        background: 'rgba(255,255,255,0.88)',
        border: '1px solid rgba(31,23,12,0.10)',
        boxShadow: '0 28px 90px rgba(31,22,12,0.12)',
        fontFamily: UI_FONT
      }}
    >
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: 'linear-gradient(rgba(31,23,12,0.035) 1px, transparent 1px), linear-gradient(90deg, rgba(31,23,12,0.035) 1px, transparent 1px)',
          backgroundSize: '34px 34px',
          pointerEvents: 'none'
        }}
      />
      <div style={{ position: 'relative' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 18, alignItems: 'flex-start', marginBottom: 22 }}>
          <div style={{ maxWidth: 420 }}>
            <div style={{ fontSize: 13, color: 'var(--text-3)', marginBottom: 8 }}>模型启动台</div>
            <div style={{ fontSize: 30, lineHeight: 1.08, fontWeight: 700, color: INK }}>
              一个入口，调度所有主流模型
            </div>
          </div>
          <div style={{ padding: '8px 11px', borderRadius: 999, background: INK, color: '#ffffff', fontSize: 12, whiteSpace: 'nowrap' }}>
            API
          </div>
        </div>

        <div className="landing-quick-start-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10, marginBottom: 14 }}>
          {quickStartRows.map(([label, value], index) => (
            <div
              key={label}
              style={{
                minHeight: 94,
                borderRadius: 18,
                background: 'rgba(255,255,255,0.74)',
                border: '1px solid rgba(31,23,12,0.06)',
                padding: 14,
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between'
              }}
            >
              <span style={{ width: 8, height: 8, borderRadius: 999, background: [ORANGE, GREEN, SKY][index] }} />
              <div>
                <div style={{ fontSize: 14, fontWeight: 700, color: INK }}>{value}</div>
                <div style={{ marginTop: 4, fontSize: 11.5, color: 'var(--text-3)' }}>{label}</div>
              </div>
            </div>
          ))}
        </div>

        <div className="landing-model-board" style={{ display: 'grid', gridTemplateColumns: '1fr 0.72fr', gap: 12 }}>
          <div style={{ borderRadius: 22, background: 'rgba(255,255,255,0.78)', border: '1px solid rgba(31,23,12,0.06)', padding: 18 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 14, marginBottom: 14 }}>
            <div style={{ fontSize: 12, color: 'var(--text-3)' }}>模型能力</div>
            <div style={{ fontSize: 11, color: GREEN, fontWeight: 700 }}>能力启用</div>
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {modelLaunchpadRows.map(([name, summary, detail, color]) => (
              <div key={name} className="landing-route-row" style={{ display: 'grid', gridTemplateColumns: '14px 1fr auto', gap: 10, alignItems: 'center' }}>
                <span style={{ width: 8, height: 8, borderRadius: 999, background: color as string }} />
                <span style={{ minWidth: 0 }}>
                  <span style={{ display: 'block', fontSize: 12.5, color: 'var(--text-2)' }}>{name}</span>
                  <span style={{ display: 'block', marginTop: 2, fontSize: 10.5, color: 'var(--text-3)' }}>{summary}</span>
                </span>
                <span style={{ fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace', fontSize: 10.5, color: 'var(--text-3)' }}>
                  {detail}
                </span>
              </div>
            ))}
          </div>
        </div>

          <div style={{ borderRadius: 22, background: 'rgba(255,87,34,0.10)', border: '1px solid rgba(255,87,34,0.18)', padding: 18, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
            <div>
              <div style={{ fontSize: 12, color: 'var(--text-3)', marginBottom: 8 }}>可接入模型</div>
              <div style={{ fontSize: 38, lineHeight: 1, fontWeight: 700, color: INK }}>100+</div>
            </div>
            <div style={{ marginTop: 24, fontSize: 12.5, lineHeight: 1.55, color: 'var(--text-2)' }}>
              按团队策略统一分发，减少多平台密钥和账号管理成本。
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function HeroD() {
  const { t } = useTranslation()

  return (
    <section
      className="landing-hero landing-hero-advanced"
      style={{
        position: 'relative',
        minHeight: 920,
        padding: '172px 0 70px',
        overflow: 'hidden',
        background: PAPER
      }}
    >
      <div
        className="landing-hero-photo"
        style={{
          position: 'absolute',
          inset: '0 0 0 auto',
          width: '58vw',
          backgroundImage:
            'linear-gradient(90deg, rgba(244,240,231,0.92), rgba(244,240,231,0.15) 34%, rgba(13,10,8,0.42)), url("/landing/case-datacenter-aisle.jpg")',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
          clipPath: 'polygon(16% 0, 100% 0, 100% 100%, 0 100%)'
        }}
      />
      <div
        style={{
          position: 'absolute',
          left: 0,
          right: 0,
          bottom: 0,
          height: 210,
          background: 'linear-gradient(0deg, rgba(244,240,231,1), rgba(244,240,231,0))'
        }}
      />

      <div className="landing-hero-grid" style={{ position: 'relative', maxWidth: 1280, margin: '0 auto', padding: '0 40px', display: 'grid', gridTemplateColumns: '0.92fr 1.08fr', gap: 54, alignItems: 'center' }}>
        <div>
          <Eyebrow>{t('landing.hero.release')}</Eyebrow>
          <h1
            className="landing-hero-title"
            style={{
              fontSize: 86,
              lineHeight: 0.98,
              fontWeight: 700,
              letterSpacing: 0,
              margin: '0 0 28px',
              color: INK,
              maxWidth: 700
            }}
          >
            {t('landing.hero.titleLine1')}
            <br />
            {t('landing.hero.titleLine2Prefix')}{' '}
            <em style={{ fontFamily: '"Source Serif 4", "Source Serif Pro", Newsreader, Georgia, serif', fontStyle: 'italic', fontWeight: 400 }}>
              {t('landing.hero.titleEmphasis')}
            </em>
            <br />
            {t('landing.hero.titleLine2Suffix')}
          </h1>
          <p style={{ fontSize: 16, lineHeight: 1.72, color: 'var(--text-3)', maxWidth: 540, margin: '0 0 34px' }}>
            {t('landing.hero.description')}
          </p>
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', marginBottom: 46 }}>
            <Link to="/register" style={{ textDecoration: 'none' }}>
              <PillBtn primary>{t('landing.hero.primaryCta')}</PillBtn>
            </Link>
            <a href="/docs" style={{ textDecoration: 'none' }}>
              <PillBtn>{t('landing.hero.secondaryCta')}</PillBtn>
            </a>
          </div>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 10 }}>
            {heroEndpointRows.map((row) => (
              <span
                key={row.name}
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 8,
                  padding: '8px 11px',
                  borderRadius: 999,
                  background: 'rgba(255,255,255,0.72)',
                  boxShadow: '0 8px 24px rgba(31,25,18,0.05)',
                  fontSize: 11.5,
                  color: 'var(--text-2)'
                }}
              >
                <span style={{ width: 7, height: 7, borderRadius: 999, background: row.color }} />
                {row.endpoint}
              </span>
            ))}
          </div>
        </div>

        <div className="landing-hero-stage" style={{ minHeight: 620, display: 'flex', alignItems: 'center' }}>
          <ModelLaunchpad />
        </div>
      </div>

      <div
        className="landing-capability-rail"
        style={{
          position: 'relative',
          maxWidth: 1280,
          margin: '72px auto 0',
          padding: '0 40px',
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: 12
        }}
      >
        {heroSummaryStats.map(([n, u, labelKey, detailKey], index) => (
          <div
            key={labelKey}
            style={{
              minHeight: 184,
              padding: 20,
              borderRadius: 18,
              background: index === 0 ? INK : 'rgba(255,255,255,0.68)',
              color: index === 0 ? '#ffffff' : INK,
              boxShadow: index === 0 ? '0 24px 70px rgba(20,15,8,0.18)' : '0 16px 44px rgba(30,23,12,0.06)',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between'
            }}
          >
            <div style={{ display: 'flex', alignItems: 'baseline' }}>
              <span style={{ fontSize: 44, lineHeight: 1, fontWeight: 600 }}>{n}</span>
              <span style={{ fontSize: 24, color: ORANGE, marginLeft: 3, fontWeight: 600 }}>{u}</span>
            </div>
            <div>
              <div style={{ fontSize: 12, lineHeight: 1.45, whiteSpace: 'pre-line', color: index === 0 ? 'rgba(255,255,255,0.64)' : 'var(--text-3)' }}>
                {t(labelKey)}
              </div>
              <div style={{ marginTop: 12, fontSize: 12.5, lineHeight: 1.55, color: index === 0 ? 'rgba(255,255,255,0.80)' : 'var(--text-2)' }}>
                {t(detailKey)}
              </div>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}

function OperationsHub() {
  const { t } = useTranslation()

  return (
    <section id="solutions" style={{ background: PAPER, padding: '96px 0 116px', scrollMarginTop: 96 }}>
      <div className="landing-section-container" style={{ maxWidth: 1280, margin: '0 auto', padding: '0 40px' }}>
        <div className="landing-section-head" style={{ display: 'grid', gridTemplateColumns: '0.9fr 1.1fr', gap: 60, alignItems: 'end', marginBottom: 46 }}>
          <div>
            <Eyebrow>{t('landing.operations.eyebrow')}</Eyebrow>
            <h2 style={{ fontSize: 58, lineHeight: 1.02, letterSpacing: 0, fontWeight: 700, margin: 0, color: INK }}>
              {t('landing.operations.titleLine1')}
              <br />
              {t('landing.operations.titleLine2')}
            </h2>
          </div>
          <p style={{ margin: 0, maxWidth: 620, fontSize: 15, lineHeight: 1.8, color: 'var(--text-3)' }}>
            {t('landing.operations.consoleSubtitle')}
          </p>
        </div>

        <div className="landing-atlas-grid" style={{ display: 'grid', gridTemplateColumns: '0.98fr 1.02fr', gap: 18, alignItems: 'stretch' }}>
          <div
            className="landing-enterprise-stack"
            style={{
              minHeight: 640,
              borderRadius: 30,
              background: INK,
              color: '#ffffff',
              border: '1px solid rgba(31,23,12,0.08)',
              padding: 28,
              boxShadow: '0 26px 80px rgba(31,22,12,0.10)',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'space-between',
              gap: 28,
              fontFamily: UI_FONT
            }}
          >
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 18, alignItems: 'flex-start', marginBottom: 26 }}>
                <div>
                  <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.58)', marginBottom: 10 }}>企业级中转</div>
                  <div style={{ fontSize: 30, lineHeight: 1.08, color: '#ffffff', fontWeight: 700 }}>从订阅、账号到 API 分发</div>
                </div>
                <span style={{ padding: '8px 12px', borderRadius: 999, background: 'rgba(255,255,255,0.10)', color: 'rgba(255,255,255,0.82)', fontSize: 12 }}>
                  Dragon style
                </span>
              </div>

              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {enterpriseStackRows.map(([title, detail, token, color], index) => (
                  <div
                    key={title}
                    className="landing-ops-flow-row"
                    style={{
                      display: 'grid',
                      gridTemplateColumns: '38px 1fr auto',
                      gap: 14,
                      alignItems: 'center',
                      padding: 16,
                      borderRadius: 20,
                      background: index === 0 ? 'rgba(255,87,34,0.18)' : 'rgba(255,255,255,0.075)',
                      border: '1px solid rgba(255,255,255,0.08)'
                    }}
                  >
                    <div style={{ width: 38, height: 38, borderRadius: 14, background: color as string, color: '#ffffff', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 12, fontWeight: 700 }}>
                      0{index + 1}
                    </div>
                    <div>
                      <div style={{ fontSize: 15, fontWeight: 700, color: '#ffffff', marginBottom: 5 }}>{title}</div>
                      <div style={{ fontSize: 12.5, lineHeight: 1.55, color: 'rgba(255,255,255,0.62)' }}>{detail}</div>
                    </div>
                    <div style={{ fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace', fontSize: 10.5, color: 'rgba(255,255,255,0.50)', whiteSpace: 'nowrap' }}>
                      {token}
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="landing-console-strip" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
              {operationsSignalRows.map(([label, value], index) => (
                <div key={label} style={{ padding: 14, borderRadius: 16, background: 'rgba(255,255,255,0.075)', color: '#ffffff', minHeight: 96 }}>
                  <span style={{ display: 'block', width: 8, height: 8, borderRadius: 999, background: [ORANGE, GREEN, SKY][index], marginBottom: 12 }} />
                  <div style={{ fontSize: 12.5, fontWeight: 700 }}>{label}</div>
                  <div style={{ marginTop: 8, fontSize: 10.5, color: 'rgba(255,255,255,0.58)', fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {value}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="landing-module-grid" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 18 }}>
            {operationsModules.map(([titleKey, descriptionKey, route, color], index) => (
              <div
                key={titleKey}
                style={{
                  borderRadius: 24,
                  minHeight: 300,
                  padding: 22,
                  background: index === 3 ? INK : 'rgba(255,255,255,0.72)',
                  color: index === 3 ? '#ffffff' : INK,
                  boxShadow: '0 18px 54px rgba(31,25,18,0.07)',
                  border: index === 3 ? '1px solid rgba(255,255,255,0.08)' : '1px solid rgba(31,23,12,0.06)',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between'
                }}
              >
                <div>
                  <span style={{ display: 'inline-block', width: 10, height: 10, background: color as string, marginBottom: 20 }} />
                  <h3 style={{ fontSize: 23, lineHeight: 1.15, margin: '0 0 12px', fontWeight: 700 }}>{t(titleKey)}</h3>
                  <p style={{ margin: 0, fontSize: 13.5, lineHeight: 1.7, color: index === 3 ? 'rgba(255,255,255,0.68)' : 'var(--text-3)' }}>
                    {t(descriptionKey)}
                  </p>
                </div>
                <div
                  style={{
                    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
                    fontSize: 10.5,
                    color: index === 3 ? 'rgba(255,255,255,0.62)' : 'var(--text-3)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap'
                  }}
                >
                  {route}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}

function SecurityBand() {
  const { t } = useTranslation()
  const securityItems = [
    ['landing.security.items.identity.title', 'landing.security.items.identity.description'],
    ['landing.security.items.backup.title', 'landing.security.items.backup.description'],
    ['landing.security.items.audit.title', 'landing.security.items.audit.description'],
    ['landing.security.items.alerts.title', 'landing.security.items.alerts.description']
  ] as const

  return (
    <section id="enterprise" style={{ background: '#151311', color: '#ffffff', padding: '110px 0', scrollMarginTop: 96 }}>
      <div className="landing-security-stack" style={{ maxWidth: 1280, margin: '0 auto', padding: '0 40px', display: 'grid', gridTemplateColumns: '0.84fr 1.16fr', gap: 52, alignItems: 'center' }}>
        <div>
          <Eyebrow dark>{t('landing.security.eyebrow')}</Eyebrow>
          <h2 style={{ fontSize: 62, lineHeight: 1.02, letterSpacing: 0, fontWeight: 700, margin: '0 0 26px' }}>
            {t('landing.security.titleLine1')}
            <br />
            {t('landing.security.titleLine2')}
          </h2>
          <p style={{ margin: 0, maxWidth: 500, fontSize: 15, lineHeight: 1.8, color: 'rgba(255,255,255,0.62)' }}>
            {t('landing.security.description')}
          </p>
        </div>
        <div
          style={{
            minHeight: 500,
            borderRadius: 30,
            backgroundImage:
              'linear-gradient(90deg, rgba(10,8,6,0.84), rgba(10,8,6,0.46)), url("/landing/case-server-racks.jpg")',
            backgroundSize: 'cover',
            backgroundPosition: 'center 44%',
            padding: 26,
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gap: 14,
            alignContent: 'end'
          }}
          className="landing-security-items"
        >
          {securityItems.map(([titleKey, descriptionKey], index) => (
            <div
              key={titleKey}
              style={{
                minHeight: 142,
                padding: 18,
                borderRadius: 18,
                background: index === 0 ? 'rgba(255,87,34,0.22)' : 'rgba(255,255,255,0.10)',
                backgroundClip: 'padding-box'
              }}
            >
              <div style={{ fontSize: 14, fontWeight: 700, marginBottom: 9 }}>{t(titleKey)}</div>
              <div style={{ fontSize: 12.5, lineHeight: 1.65, color: 'rgba(255,255,255,0.66)' }}>{t(descriptionKey)}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

function CaseStudy() {
  const { t } = useTranslation()
  const cards = [
    {
      image: '/landing/case-server-closeup.jpg',
      title: t('landing.caseStudy.cards.clarity.metric'),
      body: t('landing.caseStudy.cards.clarity.title'),
      route: '/api/v1/keys'
    },
    {
      image: '/landing/case-ops-room.jpg',
      title: t('landing.caseStudy.cards.performance.metric'),
      body: t('landing.caseStudy.cards.performance.title'),
      route: '/api/v1/admin/ops/*'
    },
    {
      image: '/landing/case-datacenter-aisle.jpg',
      title: t('landing.caseStudy.cards.consistency.metric'),
      body: t('landing.caseStudy.cards.consistency.title'),
      route: '/api/v1/payment/*'
    }
  ]

  return (
    <section id="case-studies" style={{ background: PAPER, padding: '116px 0 104px', scrollMarginTop: 96 }}>
      <div className="landing-section-container" style={{ maxWidth: 1280, margin: '0 auto', padding: '0 40px' }}>
        <div className="landing-blueprint-head" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 44, alignItems: 'end', marginBottom: 38 }}>
          <h2 style={{ fontSize: 58, lineHeight: 1.04, letterSpacing: 0, fontWeight: 700, margin: 0, color: INK }}>
            {t('landing.caseStudy.titleLine1')}
            <br />
            {t('landing.caseStudy.titleLine2')}
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            {[
              ['landing.caseStudy.highlights.gateway', '/v1/messages'],
              ['landing.caseStudy.highlights.responses', '/v1/responses'],
              ['landing.caseStudy.highlights.gemini', '/v1beta/models'],
              ['landing.caseStudy.highlights.ops', '/api/v1/admin/ops']
            ].map(([labelKey, route]) => (
              <div key={labelKey} style={{ padding: 14, borderRadius: 16, background: 'rgba(255,255,255,0.70)' }}>
                <div style={{ fontSize: 12.5, fontWeight: 700, color: INK, marginBottom: 7 }}>{t(labelKey)}</div>
                <div style={{ fontSize: 10, fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace', color: 'var(--text-3)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {route}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="landing-blueprint-grid" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 18 }}>
          {cards.map((card, index) => (
            <article
              key={card.title}
              style={{
                minHeight: 520,
                borderRadius: 28,
                overflow: 'hidden',
                backgroundImage: `linear-gradient(0deg, rgba(10,8,6,0.84), rgba(10,8,6,0.18) 54%, rgba(10,8,6,0.10)), url("${card.image}")`,
                backgroundSize: 'cover',
                backgroundPosition: index === 1 ? 'center' : 'center 48%',
                color: '#ffffff',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'flex-end',
                padding: 24
              }}
            >
              <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.56)', marginBottom: 12 }}>{card.route}</div>
              <h3 style={{ fontSize: 28, lineHeight: 1.1, margin: '0 0 12px', fontWeight: 700 }}>{card.title}</h3>
              <p style={{ margin: 0, fontSize: 14, lineHeight: 1.68, color: 'rgba(255,255,255,0.72)' }}>{card.body}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  )
}

function CtaFooter() {
  const { t } = useTranslation()
  const footerGroups = [
    [
      t('landing.footer.product'),
      [
        [t('landing.footer.routing'), '#solutions'],
        [t('landing.footer.analytics'), '/dashboard'],
        [t('landing.footer.auditLogs'), '/admin/ops'],
        [t('landing.footer.pricing'), '/models']
      ]
    ],
    [
      t('landing.footer.platform'),
      [
        [t('landing.footer.models'), '/models'],
        [t('landing.footer.playground'), '/keys'],
        [t('landing.footer.enterprise'), '#enterprise'],
        [t('landing.footer.status'), '/monitor']
      ]
    ],
    [
      t('landing.footer.resources'),
      [
        [t('landing.footer.documentation'), '/docs'],
        [t('landing.footer.changelog'), '/docs'],
        [t('landing.footer.guides'), '/docs'],
        [t('landing.footer.helpCenter'), '/docs']
      ]
    ],
    [
      t('landing.footer.company'),
      [
        [t('landing.footer.about'), '#company'],
        [t('landing.footer.customers'), '#case-studies'],
        [t('landing.footer.careers'), 'mailto:hello@xlabapi.com'],
        [t('landing.footer.contact'), 'mailto:hello@xlabapi.com']
      ]
    ]
  ] as const

  return (
    <section id="company" style={{ background: '#151311', color: '#ffffff', padding: '92px 0 0', scrollMarginTop: 96 }}>
      <div className="landing-section-container" style={{ maxWidth: 1280, margin: '0 auto', padding: '0 40px' }}>
        <div className="landing-footer-advanced" style={{ display: 'grid', gridTemplateColumns: '1fr 0.78fr', gap: 60, alignItems: 'start', paddingBottom: 78 }}>
          <div>
            <Eyebrow dark>{t('landing.cta.eyebrow')}</Eyebrow>
            <h2 style={{ fontSize: 60, lineHeight: 1.02, letterSpacing: 0, fontWeight: 700, margin: '0 0 24px' }}>
              {t('landing.cta.titleLine1')}{' '}
              <em style={{ fontFamily: '"Source Serif 4", "Source Serif Pro", Newsreader, Georgia, serif', fontStyle: 'italic', fontWeight: 400 }}>
                {t('landing.cta.titleEmphasis1')}
              </em>
              <br />
              {t('landing.cta.titleEmphasis2')} {t('landing.cta.titleLine2')}{' '}
              <em style={{ fontFamily: '"Source Serif 4", "Source Serif Pro", Newsreader, Georgia, serif', fontStyle: 'italic', fontWeight: 400 }}>
                {t('landing.cta.titleEmphasis3')}
              </em>
              ?
            </h2>
            <p style={{ maxWidth: 580, margin: 0, fontSize: 15, lineHeight: 1.8, color: 'rgba(255,255,255,0.62)' }}>
              {t('landing.cta.description')}
            </p>
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', paddingTop: 46 }}>
            <a href="mailto:hello@xlabapi.com" style={{ textDecoration: 'none' }}>
              <PillBtn primary>{t('landing.footer.contact')} →</PillBtn>
            </a>
          </div>
        </div>

        <div className="landing-footer-links" style={{ display: 'grid', gridTemplateColumns: '1.4fr 1fr 1fr 1fr 1fr', gap: 32, padding: '32px 0' }}>
          <div>
            <LogoMark dark />
            <p style={{ maxWidth: 280, fontSize: 12.5, lineHeight: 1.7, color: 'rgba(255,255,255,0.62)', margin: '18px 0 0' }}>
              {t('landing.footer.description')}
            </p>
          </div>
          {footerGroups.map(([title, items]) => (
            <div key={title}>
              <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.84)', fontWeight: 700, marginBottom: 14 }}>{title}</div>
              {items.map(([it, href]) => (
                <a key={it} href={href} style={{ display: 'block', fontSize: 12.5, color: 'rgba(255,255,255,0.58)', textDecoration: 'none', padding: '5px 0' }}>
                  {it}
                </a>
              ))}
            </div>
          ))}
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 20, padding: '24px 0', fontSize: 11.5, color: 'rgba(255,255,255,0.38)' }}>
          <span>{t('landing.footer.copyright')}</span>
          <span>{t('landing.footer.legal')}</span>
        </div>
      </div>
    </section>
  )
}

export default function Landing() {
  return (
    <div
      className="landing-page landing-page-advanced"
      style={{
        background: PAPER,
        color: INK,
        minHeight: '100vh',
        fontFamily: UI_FONT
      }}
    >
      <NavD />
      <HeroD />
      <OperationsHub />
      <SecurityBand />
      <CaseStudy />
      <CtaFooter />
    </div>
  )
}
