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
  { name: 'Claude Messages', endpoint: '/v1/messages', color: ORANGE },
  { name: 'OpenAI Responses', endpoint: '/v1/responses', color: GREEN },
  { name: 'Gemini Native', endpoint: '/v1beta/models/*', color: SKY },
  { name: 'Antigravity', endpoint: '/antigravity/v1', color: VIOLET }
] as const

const heroControlRows = [
  { labelKey: 'landing.demo.controls.keys', route: '/api/v1/keys', color: ORANGE },
  { labelKey: 'landing.demo.controls.usage', route: '/api/v1/usage/dashboard/*', color: GREEN },
  { labelKey: 'landing.demo.controls.payment', route: '/api/v1/payment/*', color: SKY },
  { labelKey: 'landing.demo.controls.admin', route: '/api/v1/admin/*', color: VIOLET }
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

const topologyRows = [
  ['landing.operations.demo.chain.group', 'priority / rpm / tpm'],
  ['landing.operations.demo.chain.account', 'concurrency / schedulable'],
  ['landing.operations.demo.chain.fallback', 'fallback_group_id'],
  ['landing.operations.apiRows.ops', '/api/v1/admin/ops/*']
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

function EndpointDeck() {
  const { t } = useTranslation()

  return (
    <div
      className="landing-endpoint-deck"
      style={{
        borderRadius: 26,
        background: 'rgba(18,15,13,0.92)',
        boxShadow: '0 42px 110px rgba(20,10,0,0.38)',
        color: '#ffffff',
        overflow: 'hidden',
        fontFamily: UI_FONT
      }}
    >
      <div style={{ padding: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 20 }}>
        <div>
          <div style={{ fontSize: 24, fontWeight: 700 }}>{t('landing.demo.gatewayTitle')}</div>
          <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(255,255,255,0.68)' }}>{t('landing.demo.fromRoutes')}</div>
        </div>
        <div style={{ padding: '7px 11px', borderRadius: 999, background: 'rgba(34,197,94,0.15)', color: '#86efac', fontSize: 12 }}>
          {t('landing.operations.systemActive')}
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 0.82fr' }} className="landing-deck-body">
        <div style={{ padding: '8px 24px 24px', display: 'flex', flexDirection: 'column', gap: 16 }}>
          {heroEndpointRows.map((row) => (
            <div key={row.name}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 16, fontSize: 12, color: 'rgba(255,255,255,0.72)', marginBottom: 8 }}>
                <span>{row.name}</span>
                <span style={{ fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace' }}>{row.endpoint}</span>
              </div>
              <div style={{ height: 8, borderRadius: 999, background: 'rgba(255,255,255,0.10)', overflow: 'hidden' }}>
                <div style={{ width: '100%', height: '100%', borderRadius: 999, background: row.color }} />
              </div>
            </div>
          ))}
        </div>

        <div
          style={{
            padding: 20,
            background: 'rgba(255,255,255,0.06)',
            display: 'flex',
            flexDirection: 'column',
            gap: 10
          }}
        >
          <div style={{ fontSize: 11.5, color: 'rgba(255,255,255,0.66)', marginBottom: 2 }}>
            {t('landing.demo.controlTitle')}
          </div>
          {heroControlRows.map((row) => (
            <div key={row.labelKey} style={{ display: 'grid', gridTemplateColumns: '16px 1fr', gap: 10, alignItems: 'center' }}>
              <span style={{ width: 8, height: 8, borderRadius: 999, background: row.color }} />
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: 12.5, fontWeight: 600 }}>{t(row.labelKey)}</div>
                <div
                  style={{
                    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
                    fontSize: 10,
                    color: 'rgba(255,255,255,0.62)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap'
                  }}
                >
                  {row.route}
                </div>
              </div>
            </div>
          ))}
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
          <EndpointDeck />
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

        <div className="landing-atlas-grid" style={{ display: 'grid', gridTemplateColumns: '1.05fr 0.95fr', gap: 18, alignItems: 'stretch' }}>
          <div
            className="landing-ops-console"
            style={{
              minHeight: 640,
              borderRadius: 28,
              background: INK,
              color: '#ffffff',
              padding: 28,
              display: 'grid',
              gridTemplateRows: 'auto 1fr auto',
              gap: 26,
              boxShadow: '0 34px 90px rgba(31,22,12,0.22)',
              fontFamily: UI_FONT
            }}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 20, alignItems: 'flex-start' }}>
              <div>
                <div style={{ fontSize: 28, fontWeight: 700 }}>{t('landing.operations.consoleTitle')}</div>
                <div style={{ marginTop: 8, fontSize: 12.5, color: 'rgba(255,255,255,0.66)' }}>{t('landing.operations.apiTitle')}</div>
              </div>
              <span style={{ padding: '7px 11px', borderRadius: 999, background: 'rgba(255,255,255,0.10)', fontSize: 12, color: 'rgba(255,255,255,0.72)' }}>
                {t('landing.operations.deploy')}
              </span>
            </div>

            <div className="landing-topology-grid" style={{ display: 'grid', gridTemplateColumns: '0.8fr 1.2fr', gap: 18, alignItems: 'stretch' }}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {topologyRows.map(([labelKey, route], index) => (
                  <div
                    key={labelKey}
                    style={{
                      padding: 16,
                      borderRadius: 16,
                      background: index === 0 ? 'rgba(255,87,34,0.18)' : 'rgba(255,255,255,0.075)'
                    }}
                  >
                    <div style={{ fontSize: 13, fontWeight: 600 }}>{t(labelKey)}</div>
                    <div style={{ marginTop: 8, fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace', fontSize: 10.5, color: 'rgba(255,255,255,0.64)' }}>
                      {route}
                    </div>
                  </div>
                ))}
              </div>

              <div style={{ borderRadius: 20, background: 'rgba(255,255,255,0.06)', padding: 20, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                {heroEndpointRows.map((row, index) => (
                  <div key={row.name}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: 14, fontSize: 12, color: 'rgba(255,255,255,0.72)', marginBottom: 8 }}>
                      <span>{row.name}</span>
                      <span>{index === 0 ? t('landing.operations.metrics.available') : row.endpoint}</span>
                    </div>
                    <div style={{ height: 10, borderRadius: 999, background: 'rgba(255,255,255,0.10)', overflow: 'hidden' }}>
                      <div style={{ height: '100%', width: `${100 - index * 13}%`, borderRadius: 999, background: row.color }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="landing-console-strip" style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
              {[
                ['landing.operations.metrics.usageDashboard', '/api/v1/usage/dashboard/*'],
                ['landing.operations.metrics.opsMonitor', '/api/v1/admin/ops/*'],
                ['landing.operations.metrics.retryTrace', '/api/v1/admin/ops/errors']
              ].map(([labelKey, route]) => (
                <div key={labelKey} style={{ padding: 14, borderRadius: 14, background: 'rgba(255,255,255,0.075)' }}>
                  <div style={{ fontSize: 12.5, fontWeight: 600 }}>{t(labelKey)}</div>
                  <div style={{ marginTop: 8, fontSize: 10, color: 'rgba(255,255,255,0.62)', fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {route}
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
                  background: index === 3 ? 'linear-gradient(145deg, #271915, #11100f)' : 'rgba(255,255,255,0.70)',
                  color: index === 3 ? '#ffffff' : INK,
                  boxShadow: '0 18px 54px rgba(31,25,18,0.07)',
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
