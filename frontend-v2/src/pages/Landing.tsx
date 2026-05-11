/**
 * Landing v4 — Plato (Xlabapi dark + warm orange)
 * Adapted from the frontend-v2 landing prototype.
 */
import type { CSSProperties, ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { SectionFrame } from '@/components/bus/SectionFrame'
import { LocaleSwitcher } from '@/components/layout/LocaleSwitcher'
import { PlasmaBlob, HalftoneOverlay } from '@/components/bus/PlasmaBlob'

const ORANGE = '#ff5722'

function OrangeMark({ size = 8 }: { size?: number }) {
  return (
    <span
      style={{
        display: 'inline-block',
        width: size,
        height: size,
        background: ORANGE,
        marginRight: 10,
        verticalAlign: 'middle',
        flexShrink: 0
      }}
    />
  )
}

function Eyebrow({ children, mb = 18 }: { children: ReactNode; mb?: number }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', marginBottom: mb }}>
      <OrangeMark />
      <span
        className="font-mono"
        style={{
          fontSize: 11.5,
          letterSpacing: '0.18em',
          color: 'rgba(255,255,255,0.7)',
          textTransform: 'uppercase'
        }}
      >
        {children}
      </span>
    </div>
  )
}

function PillBtn({
  children,
  primary,
  ghost
}: {
  children: ReactNode
  primary?: boolean
  ghost?: boolean
}) {
  return (
    <button
      style={{
        height: 44,
        padding: '0 22px',
        borderRadius: 999,
        fontSize: 14,
        fontWeight: 500,
        cursor: 'pointer',
        border: ghost ? '1px solid rgba(255,255,255,0.16)' : 'none',
        background: primary ? ORANGE : ghost ? 'transparent' : '#fff',
        color: primary ? '#fff' : ghost ? '#fff' : '#000',
        letterSpacing: '-0.01em'
      }}
    >
      {children}
    </button>
  )
}

function NavD() {
  const { t } = useTranslation()
  const navItems = [
    ['#case-studies', t('landing.nav.caseStudies')],
    ['#solutions', t('landing.nav.solutions')],
    ['#enterprise', t('landing.nav.enterprise')],
    ['#company', t('landing.nav.company')]
  ] as const

  return (
    <nav
      style={{
        position: 'absolute',
        top: 24,
        left: 0,
        right: 0,
        zIndex: 10,
        padding: '0 40px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between'
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 12,
          fontSize: 15,
          fontWeight: 600,
          letterSpacing: '-0.01em'
        }}
      >
        <div style={{ width: 28, height: 28, position: 'relative' }}>
          <div
            style={{
              position: 'absolute',
              inset: 0,
              border: '1.5px solid #fff',
              transform: 'rotate(45deg)',
              borderRadius: 2
            }}
          />
          <div style={{ position: 'absolute', inset: 6, border: '1.5px solid #fff', borderRadius: 2 }} />
        </div>
        XLABAPI
      </div>
      <div
        style={{
          display: 'flex',
          gap: 4,
          padding: 4,
          borderRadius: 999,
          background: 'rgba(255,255,255,0.04)',
          border: '1px solid rgba(255,255,255,0.08)'
        }}
      >
        {navItems.map(([href, label]) => (
          <a
            key={href}
            href={href}
            style={{
              padding: '8px 16px',
              fontSize: 13.5,
              color: 'rgba(255,255,255,0.7)',
              textDecoration: 'none',
              borderRadius: 999
            }}
          >
            {label}
          </a>
        ))}
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <a
          href="/docs"
          style={{
            padding: '8px 14px',
            fontSize: 13.5,
            color: 'rgba(255,255,255,0.7)',
            textDecoration: 'none'
          }}
        >
          {t('nav.docs')}
        </a>
        <LocaleSwitcher compact />
        <Link to="/login" style={{ textDecoration: 'none' }}>
          <PillBtn primary>{t('landing.nav.signIn')}</PillBtn>
        </Link>
      </div>
    </nav>
  )
}

function HeroD() {
  const { t } = useTranslation()
  return (
    <section
      style={{
        position: 'relative',
        minHeight: 880,
        paddingTop: 180,
        paddingBottom: 120,
        overflow: 'hidden',
        background: '#000'
      }}
    >
      <div
        style={{
          position: 'absolute',
          top: 60,
          left: '50%',
          transform: 'translateX(-50%)',
          width: 1280,
          height: 800,
          pointerEvents: 'none'
        }}
      >
        <div style={{ position: 'absolute', right: -40, top: 0, width: 760, height: 760 }}>
          <PlasmaBlob />
          <HalftoneOverlay opacity={0.85} />
        </div>
      </div>
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: 'radial-gradient(rgba(255,255,255,0.03) 1px, transparent 1px)',
          backgroundSize: '32px 32px',
          pointerEvents: 'none'
        }}
      />
      <SectionFrame showTop={false} />

      <div style={{ position: 'relative', maxWidth: 1280, margin: '0 auto', padding: '0 40px' }}>
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 10,
            padding: '6px 14px 6px 6px',
            borderRadius: 999,
            background: 'rgba(255,255,255,0.04)',
            border: '1px solid rgba(255,255,255,0.1)',
            marginBottom: 28
          }}
        >
          <span
            style={{
              padding: '2px 10px',
              borderRadius: 999,
              background: ORANGE,
              fontSize: 11,
              fontWeight: 600,
              color: '#fff'
            }}
          >
            {t('landing.hero.badge')}
          </span>
          <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.7)' }}>
            {t('landing.hero.release')}
          </span>
        </div>

        <h1
          style={{
            fontSize: 84,
            fontWeight: 500,
            lineHeight: 1.04,
            letterSpacing: '-0.035em',
            margin: '0 0 28px',
            color: '#fff',
            maxWidth: 760
          }}
        >
          {t('landing.hero.titleLine1')}
          <br />
          {t('landing.hero.titleLine2Prefix')}{' '}
          <em
            style={{ fontStyle: 'italic', fontFamily: 'Georgia, serif', fontWeight: 400 }}
          >
            {t('landing.hero.titleEmphasis')}
          </em>{' '}
          {t('landing.hero.titleLine2Suffix')}
        </h1>
        <p
          style={{
            fontSize: 16,
            lineHeight: 1.6,
            color: 'rgba(255,255,255,0.55)',
            maxWidth: 540,
            margin: '0 0 36px'
          }}
        >
          {t('landing.hero.description')}
        </p>
        <div style={{ display: 'flex', gap: 12, marginBottom: 100 }}>
          <Link to="/register" style={{ textDecoration: 'none' }}>
            <PillBtn primary>{t('landing.hero.primaryCta')}</PillBtn>
          </Link>
          <a href="/docs" style={{ textDecoration: 'none' }}>
            <PillBtn>{t('landing.hero.secondaryCta')}</PillBtn>
          </a>
        </div>

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr 1fr 1fr',
            gap: 60,
            alignItems: 'flex-end',
            paddingTop: 80,
            borderTop: '1px solid rgba(255,255,255,0.08)'
          }}
        >
          <div
            style={{
              fontSize: 14,
              color: 'rgba(255,255,255,0.6)',
              lineHeight: 1.5,
              maxWidth: 280
            }}
          >
            {t('landing.hero.summaryLine1')}<br />
            <strong style={{ color: '#fff' }}>{t('landing.hero.summaryStrong')}</strong><br />
            {t('landing.hero.summaryLine2')}
          </div>
          {(
            [
              ['99', '%', 'Uptime\nSLA 保证'],
              ['5.5', 'x', 'Faster Failover\n上游切换速度'],
              ['10B', '+', 'Tokens Processed\n每月调用量']
            ] as const
          ).map(([n, u, l]) => (
            <div key={l}>
              <div style={{ display: 'flex', alignItems: 'baseline', color: '#fff' }}>
                <span style={{ fontSize: 56, fontWeight: 500, letterSpacing: '-0.03em', lineHeight: 1 }}>{n}</span>
                <span style={{ fontSize: 32, color: ORANGE, marginLeft: 2, fontWeight: 500 }}>{u}</span>
              </div>
              <div
                style={{
                  fontSize: 11.5,
                  color: 'rgba(255,255,255,0.5)',
                  marginTop: 8,
                  whiteSpace: 'pre-line',
                  lineHeight: 1.4
                }}
              >
                {l}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

function BackedRow() {
  const { t } = useTranslation()
  return (
    <section style={{ position: 'relative', background: '#000', padding: '40px 0' }}>
      <SectionFrame />
      <div
        style={{
          position: 'relative',
          maxWidth: 1280,
          margin: '0 auto',
          padding: '0 40px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          gap: 40
        }}
      >
        <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.45)', lineHeight: 1.5 }}>
          {t('landing.backed.by')}
          <br />
          {t('landing.backed.teams')}
        </div>
        <div
          style={{
            display: 'flex',
            gap: 56,
            color: 'rgba(255,255,255,0.55)',
            fontSize: 16,
            fontWeight: 600
          }}
        >
          {['OpenAI', 'ANTHROPIC', 'Meta', 'DeepMind', 'Google', 'Mistral'].map((b) => (
            <span
              key={b}
              style={{
                fontFamily: ['OpenAI', 'Meta', 'Google'].includes(b) ? 'Georgia, serif' : 'inherit',
                letterSpacing: b === 'ANTHROPIC' ? '0.1em' : 0
              }}
            >
              {b}
            </span>
          ))}
        </div>
      </div>
    </section>
  )
}

function OperationsHub() {
  const { t } = useTranslation()
  const featureCards = [
    [t('landing.operations.features.routing.title'), t('landing.operations.features.routing.description')],
    [t('landing.operations.features.unified.title'), t('landing.operations.features.unified.description')],
    [t('landing.operations.features.reliability.title'), t('landing.operations.features.reliability.description')]
  ] as const
  const operationStats = [
    ['60K', '+', t('landing.operations.stats.workflows')],
    ['18B', '+', t('landing.operations.stats.tokens')],
    ['100', '+', t('landing.operations.stats.models')]
  ] as const
  const capabilities = [
    t('landing.operations.capabilities.smartRouting'),
    t('landing.operations.capabilities.calculationEngine'),
    t('landing.operations.capabilities.predictiveFailover'),
    t('landing.operations.capabilities.integration'),
    t('landing.operations.capabilities.analytics'),
    t('landing.operations.capabilities.deployment')
  ]

  return (
    <section style={{ position: 'relative', background: '#000', padding: '100px 0' }}>
      <SectionFrame />
      <div
        style={{
          maxWidth: 1280,
          margin: '0 auto',
          padding: '0 40px',
          position: 'relative',
          zIndex: 5
        }}
      >
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr 1fr 1fr',
            gap: 48,
            paddingBottom: 64,
            borderBottom: '1px solid rgba(255,255,255,0.06)'
          }}
        >
          {(
            featureCards
          ).map(([t, d]) => (
            <div key={t}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 16 }}>
                <span style={{ width: 12, height: 12, background: ORANGE, borderRadius: 2 }} />
                <span style={{ fontSize: 14, fontWeight: 600, color: '#fff' }}>{t}</span>
              </div>
              <p
                style={{
                  fontSize: 13,
                  lineHeight: 1.6,
                  color: 'rgba(255,255,255,0.5)',
                  margin: 0,
                  maxWidth: 240
                }}
              >
                {d}
              </p>
            </div>
          ))}
          <div>
            <Eyebrow mb={16}>{t('landing.operations.eyebrow')}</Eyebrow>
            <h2
              style={{
                fontSize: 38,
                fontWeight: 500,
                lineHeight: 1.08,
                letterSpacing: '-0.025em',
                margin: 0,
                color: '#fff'
              }}
            >
              {t('landing.operations.titleLine1')}
              <br />
              {t('landing.operations.titleLine2')}
            </h2>
          </div>
        </div>

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 280px',
            gap: 48,
            padding: '48px 0'
          }}
        >
          <div
            style={{
              position: 'relative',
              height: 540,
              borderRadius: 12,
              overflow: 'hidden',
              background: '#000'
            }}
          >
            <PlasmaBlob />
            <HalftoneOverlay opacity={0.7} />

            <div
              style={{
                position: 'absolute',
                left: '12%',
                top: '12%',
                right: '8%',
                bottom: '12%',
                background: 'rgba(20,8,4,0.72)',
                backdropFilter: 'blur(20px)',
                border: '1px solid rgba(255,255,255,0.08)',
                borderRadius: 14,
                padding: 24,
                color: '#fff',
                display: 'flex',
                flexDirection: 'column',
                boxShadow: '0 30px 80px rgba(0,0,0,0.5)'
              }}
            >
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  paddingBottom: 14,
                  borderBottom: '1px solid rgba(255,255,255,0.08)'
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
                  <span
                    style={{ width: 14, height: 14, borderRadius: 3, background: ORANGE, opacity: 0.9 }}
                  />
                  {t('landing.operations.consoleTitle')}
                </div>
                <div style={{ display: 'flex', gap: 10, fontSize: 11.5 }}>
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 6,
                      padding: '3px 10px',
                      borderRadius: 999,
                      background: 'rgba(34,197,94,0.15)',
                      border: '1px solid rgba(34,197,94,0.3)',
                      color: '#4ade80'
                    }}
                  >
                    <span
                      style={{ width: 6, height: 6, borderRadius: 999, background: '#22c55e' }}
                    />
                    {t('landing.operations.systemActive')}
                  </span>
                  <span
                    style={{
                      padding: '3px 10px',
                      borderRadius: 999,
                      border: '1px solid rgba(255,255,255,0.15)'
                    }}
                  >
                    {t('landing.operations.deploy')} ▾
                  </span>
                </div>
              </div>
              <div style={{ paddingTop: 18, paddingBottom: 14 }}>
                <div style={{ fontSize: 22, fontWeight: 500, marginBottom: 4 }}>
                  {t('landing.operations.consoleTitle')}
                </div>
                <div style={{ fontSize: 11.5, color: 'rgba(255,255,255,0.5)' }}>
                  {t('landing.operations.consoleSubtitle')}
                </div>
              </div>

              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: '1fr 1fr',
                  gap: 28,
                  flex: 1
                }}
              >
                <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                    {(
                      [
                        [t('landing.operations.strategy'), t('landing.operations.costOptimized')],
                        [t('landing.operations.tier'), t('landing.operations.production')]
                      ] as const
                    ).map(([k, v]) => (
                      <div key={k}>
                        <div
                          style={{
                            fontSize: 10.5,
                            color: 'rgba(255,255,255,0.5)',
                            marginBottom: 6
                          }}
                        >
                          {k}
                        </div>
                        <div
                          style={{
                            height: 32,
                            borderRadius: 6,
                            border: '1px solid rgba(255,255,255,0.12)',
                            display: 'flex',
                            alignItems: 'center',
                            padding: '0 10px',
                            fontSize: 11.5,
                            color: 'rgba(255,255,255,0.6)'
                          }}
                        >
                          {v}
                        </div>
                      </div>
                    ))}
                  </div>
                  <div>
                    <div
                      style={{ fontSize: 10.5, color: 'rgba(255,255,255,0.5)', marginBottom: 6 }}
                    >
                      {t('landing.operations.fallbackChain')}
                    </div>
                    <div
                      style={{
                        height: 32,
                        borderRadius: 6,
                        border: '1px solid rgba(255,255,255,0.12)'
                      }}
                    />
                  </div>
                  <div>
                    <div
                      style={{ fontSize: 10.5, color: 'rgba(255,255,255,0.5)', marginBottom: 6 }}
                    >
                      {t('landing.operations.customRoutingRules')}
                    </div>
                    <div
                      style={{
                        height: 96,
                        borderRadius: 6,
                        border: '1px solid rgba(255,255,255,0.12)'
                      }}
                    />
                  </div>
                </div>
                <div
                  style={{ borderLeft: '1px solid rgba(255,255,255,0.08)', paddingLeft: 22 }}
                >
                  <div style={{ fontSize: 13, fontWeight: 500 }}>
                    {t('landing.operations.routingIntelligence')}
                  </div>
                  <div
                    style={{
                      fontSize: 10.5,
                      color: 'rgba(255,255,255,0.5)',
                      marginTop: 2,
                      marginBottom: 16
                    }}
                  >
                    {t('landing.operations.insights')}
                  </div>
                  <div style={{ fontSize: 10.5, color: 'rgba(255,255,255,0.5)' }}>
                    {t('landing.operations.healthScore')}
                  </div>
                  <div
                    style={{ display: 'flex', alignItems: 'baseline', gap: 4, marginTop: 4 }}
                  >
                    <span
                      style={{ fontSize: 42, fontWeight: 500, letterSpacing: '-0.02em' }}
                    >
                      96%
                    </span>
                    <span style={{ color: ORANGE, fontSize: 18 }}>↗</span>
                  </div>
                  <div style={{ fontSize: 10.5, color: ORANGE, marginTop: 4 }}>
                    {t('landing.operations.improvement')}
                  </div>
                  <div
                    style={{
                      marginTop: 18,
                      display: 'flex',
                      flexDirection: 'column',
                      gap: 12
                    }}
                  >
                    {(
                      [
                        [t('landing.operations.metrics.accuracy'), '#22c55e', '2.8%'],
                        [t('landing.operations.metrics.completeness'), '#a78bfa', '3.9%'],
                        [t('landing.operations.metrics.latency'), '#fb923c', '1.8%']
                      ] as const
                    ).map(([l, c, v]) => (
                      <div key={l}>
                        <div
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            fontSize: 10.5,
                            color: 'rgba(255,255,255,0.6)',
                            marginBottom: 4
                          }}
                        >
                          <span>{l}</span>
                          <span>{v}</span>
                        </div>
                        <div style={{ display: 'flex', gap: 1.5, height: 8 }}>
                          {Array.from({ length: 32 }).map((_, i) => (
                            <div
                              key={i}
                              style={{
                                flex: 1,
                                background: ((i * 7 + l.length) % 10) > 5 ? c : 'rgba(255,255,255,0.06)',
                                borderRadius: 1
                              }}
                            />
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  paddingTop: 16,
                  borderTop: '1px solid rgba(255,255,255,0.08)',
                  marginTop: 16
                }}
              >
                <span style={{ fontSize: 11.5, color: 'rgba(255,255,255,0.4)' }}>
                  {t('common.cancel')}
                </span>
                <div style={{ display: 'flex', gap: 8 }}>
                  <span
                    style={{
                      padding: '7px 16px',
                      borderRadius: 6,
                      background: '#fff',
                      color: '#000',
                      fontSize: 11.5,
                      fontWeight: 500
                    }}
                  >
                    {t('landing.operations.createRouting')}
                  </span>
                  <span
                    style={{
                      padding: '7px 14px',
                      borderRadius: 6,
                      border: '1px solid rgba(255,255,255,0.15)',
                      fontSize: 11.5
                    }}
                  >
                    {t('landing.operations.syncHub')} →
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 36 }}>
            {operationStats.map(([n, u, l]) => (
              <div key={l}>
                <div style={{ display: 'flex', alignItems: 'baseline', color: '#fff' }}>
                  <span
                    style={{ fontSize: 52, fontWeight: 500, letterSpacing: '-0.03em', lineHeight: 1 }}
                  >
                    {n}
                  </span>
                  <span style={{ fontSize: 32, color: ORANGE, marginLeft: 2 }}>{u}</span>
                </div>
                <div
                  style={{
                    fontSize: 11.5,
                    color: 'rgba(255,255,255,0.55)',
                    marginTop: 10,
                    whiteSpace: 'pre-line',
                    lineHeight: 1.45
                  }}
                >
                  {l}
                </div>
              </div>
            ))}
            <div
              style={{
                paddingTop: 24,
                marginTop: 12,
                borderTop: '1px solid rgba(255,255,255,0.08)'
              }}
            >
              <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.45)', marginBottom: 14 }}>
                {t('landing.operations.capabilities.title')}
              </div>
              {capabilities.map((c, i) => (
                <div
                  key={c}
                  style={{
                    padding: '8px 12px',
                    fontSize: 12.5,
                    color: i === 1 ? '#fff' : 'rgba(255,255,255,0.55)',
                    background: i === 1 ? 'rgba(255,87,34,0.12)' : 'transparent',
                    border: i === 1 ? '1px solid rgba(255,87,34,0.3)' : 'none',
                    borderRadius: 6,
                    marginBottom: 1
                  }}
                >
                  {c}
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

function SecurityBand() {
  const { t } = useTranslation()
  return (
    <section style={{ position: 'relative', background: '#000', padding: '80px 0' }}>
      <SectionFrame />
      <div
        style={{
          position: 'relative',
          maxWidth: 1280,
          margin: '0 auto',
          padding: '0 40px',
          display: 'grid',
          gridTemplateColumns: '1.2fr 1fr 0.8fr',
          gap: 48,
          alignItems: 'center',
          zIndex: 5
        }}
      >
        <div>
          <Eyebrow>{t('landing.security.eyebrow')}</Eyebrow>
          <h2
            style={{
              fontSize: 44,
              fontWeight: 500,
              lineHeight: 1.05,
              letterSpacing: '-0.025em',
              margin: 0,
              color: '#fff'
            }}
          >
            {t('landing.security.titleLine1')}
            <br />
            {t('landing.security.titleLine2')}
          </h2>
        </div>
        <p
          style={{ fontSize: 14, lineHeight: 1.65, color: 'rgba(255,255,255,0.55)', margin: 0 }}
        >
          {t('landing.security.description')}
        </p>
        <div style={{ display: 'flex', gap: 16, justifyContent: 'flex-end' }}>
          {['SOC 2 TYPE II', 'ISO 27001'].map((b) => (
            <div
              key={b}
              style={{
                width: 110,
                height: 110,
                borderRadius: 999,
                border: '1px solid rgba(255,255,255,0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 11,
                fontWeight: 600,
                letterSpacing: '0.06em',
                color: 'rgba(255,255,255,0.7)',
                textAlign: 'center',
                lineHeight: 1.3
              }}
            >
              {b}
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
      gradient: 'linear-gradient(135deg, #1a1f3a, #2a3550 40%, #f97316 90%)',
      t: t('landing.caseStudy.cards.clarity.title'),
      sub: t('landing.caseStudy.cards.clarity.sub')
    },
    {
      gradient: 'linear-gradient(160deg, #0f172a, #1e293b 30%, #f97316 90%)',
      t: t('landing.caseStudy.cards.performance.title'),
      sub: t('landing.caseStudy.cards.performance.sub')
    },
    {
      gradient: 'linear-gradient(135deg, #c2780f, #d97757 50%, #fbbf24 100%)',
      t: t('landing.caseStudy.cards.consistency.title'),
      sub: t('landing.caseStudy.cards.consistency.sub')
    }
  ]

  return (
    <section style={{ position: 'relative', background: '#000', padding: '120px 0' }}>
      <SectionFrame />
      <div
        style={{ position: 'relative', maxWidth: 1280, margin: '0 auto', padding: '0 40px', zIndex: 5 }}
      >
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1.4fr',
            gap: 48,
            paddingBottom: 80
          }}
        >
          <div>
            <h2
              style={{
                fontSize: 32,
                fontWeight: 500,
                lineHeight: 1.2,
                letterSpacing: '-0.02em',
                margin: '0 0 28px',
                color: '#fff'
            }}
          >
              {t('landing.caseStudy.titleLine1')}
              <br />
              {t('landing.caseStudy.titleLine2')}
            </h2>
            <PillBtn primary>{t('common.more')}</PillBtn>

            <div style={{ marginTop: 80 }}>
              <div
                style={{
                  fontSize: 22,
                  fontWeight: 700,
                  letterSpacing: '0.02em',
                  color: ORANGE,
                  marginBottom: 28
                }}
              >
                NOVAGRID
              </div>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  marginBottom: 18
                }}
              >
                <div
                  style={{
                    width: 32,
                    height: 32,
                    borderRadius: 999,
                    background: 'linear-gradient(135deg, #d97757, #6c4434)'
                  }}
                />
                <div>
                  <div style={{ fontSize: 13.5, fontWeight: 500, color: '#fff' }}>Daniel Reyes</div>
                  <div style={{ fontSize: 11.5, color: 'rgba(255,255,255,0.5)' }}>
                    {t('landing.caseStudy.personRole')}
                  </div>
                </div>
              </div>
              <p
                style={{
                  fontSize: 14,
                  lineHeight: 1.65,
                  color: 'rgba(255,255,255,0.6)',
                  margin: 0,
                  maxWidth: 420
                }}
              >
                {t('landing.caseStudy.quote')}
              </p>
            </div>
          </div>

          <div
            style={{
              position: 'relative',
              minHeight: 460,
              borderRadius: 12,
              overflow: 'hidden',
              background: '#000'
            }}
          >
            <div
              style={{
                position: 'absolute',
                inset: 0,
                background: `
                  linear-gradient(135deg, transparent 0%, rgba(255,87,34,0.4) 30%, rgba(255,140,60,0.7) 60%, rgba(255,180,80,0.5) 100%),
                  repeating-linear-gradient(45deg, rgba(255,87,34,0.2) 0px, rgba(255,87,34,0.2) 80px, rgba(150,40,10,0.4) 80px, rgba(150,40,10,0.4) 160px)
                `
              }}
            />
            <HalftoneOverlay opacity={0.4} />
            <div style={{ position: 'absolute', left: 32, bottom: 32, color: '#fff' }}>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 4 }}>
                <span
                  style={{ fontSize: 64, fontWeight: 500, letterSpacing: '-0.03em', lineHeight: 1 }}
                >
                  99.20
                </span>
                <span style={{ fontSize: 28, color: 'rgba(255,255,255,0.7)' }}>%</span>
              </div>
              <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.7)', marginTop: 6 }}>
                {t('landing.caseStudy.reliabilityRate')}
              </div>
            </div>
          </div>
        </div>

        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1fr 1fr 1fr',
            gap: 24,
            paddingTop: 56,
            borderTop: '1px solid rgba(255,255,255,0.08)'
          }}
        >
          {cards.map((c, i) => (
            <div key={i}>
              <div
                style={{
                  height: 200,
                  borderRadius: 8,
                  background: c.gradient,
                  position: 'relative',
                  overflow: 'hidden',
                  marginBottom: 18
                } as CSSProperties}
              >
                <HalftoneOverlay opacity={0.3} />
              </div>
              <div
                style={{
                  fontSize: 14.5,
                  fontWeight: 500,
                  color: '#fff',
                  lineHeight: 1.45,
                  marginBottom: 8
                }}
              >
                {c.t}
              </div>
              <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.45)' }}>{c.sub}</div>
            </div>
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
      [t('landing.footer.routing'), t('landing.footer.analytics'), t('landing.footer.auditLogs'), t('landing.footer.pricing')]
    ],
    [
      t('landing.footer.platform'),
      [t('landing.footer.models'), t('landing.footer.playground'), t('landing.footer.enterprise'), t('landing.footer.status')]
    ],
    [
      t('landing.footer.resources'),
      [t('landing.footer.documentation'), t('landing.footer.changelog'), t('landing.footer.guides'), t('landing.footer.helpCenter')]
    ],
    [
      t('landing.footer.company'),
      [t('landing.footer.about'), t('landing.footer.customers'), t('landing.footer.careers'), t('landing.footer.contact')]
    ]
  ] as const

  return (
    <section style={{ position: 'relative', background: '#000', padding: '80px 0 0' }}>
      <SectionFrame showBottom={false} />
      <div
        style={{ position: 'relative', maxWidth: 1280, margin: '0 auto', padding: '0 40px', zIndex: 5 }}
      >
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: '1.1fr 1fr 0.6fr',
            gap: 48,
            alignItems: 'flex-start',
            paddingBottom: 80
          }}
        >
          <div>
            <Eyebrow>{t('landing.cta.eyebrow')}</Eyebrow>
            <h2
              style={{
                fontSize: 40,
                fontWeight: 500,
                lineHeight: 1.1,
                letterSpacing: '-0.025em',
                margin: 0,
                color: '#fff'
            }}
          >
              {t('landing.cta.titleLine1')}{' '}
              <em style={{ fontStyle: 'italic', fontFamily: 'Georgia, serif' }}>{t('landing.cta.titleEmphasis1')}</em>
              <br />
              <em style={{ fontStyle: 'italic', fontFamily: 'Georgia, serif' }}>{t('landing.cta.titleEmphasis2')}</em>{' '}
              {t('landing.cta.titleLine2')}{' '}
              <em style={{ fontStyle: 'italic', fontFamily: 'Georgia, serif' }}>{t('landing.cta.titleEmphasis3')}</em>?
            </h2>
          </div>
          <p
            style={{
              fontSize: 14,
              lineHeight: 1.65,
              color: 'rgba(255,255,255,0.55)',
              margin: 0,
              paddingTop: 32
            }}
          >
            {t('landing.cta.description')}
          </p>
          <div style={{ paddingTop: 32, display: 'flex', justifyContent: 'flex-end' }}>
            <PillBtn primary>{t('landing.footer.contact')} →</PillBtn>
          </div>
        </div>

        <div
          style={{
            paddingTop: 48,
            paddingBottom: 32,
            borderTop: '1px solid rgba(255,255,255,0.08)',
            display: 'grid',
            gridTemplateColumns: '1.5fr 1fr 1fr 1fr 1fr',
            gap: 32
          }}
        >
          <div>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                marginBottom: 16,
                color: '#fff',
                fontSize: 14,
                fontWeight: 600
              }}
            >
              <div style={{ width: 24, height: 24, position: 'relative' }}>
                <div
                  style={{
                    position: 'absolute',
                    inset: 0,
                    border: '1.5px solid #fff',
                    transform: 'rotate(45deg)',
                    borderRadius: 2
                  }}
                />
                <div
                  style={{
                    position: 'absolute',
                    inset: 5,
                    border: '1.5px solid #fff',
                    borderRadius: 2
                  }}
                />
              </div>
              XLABAPI
            </div>
            <p
              style={{
                fontSize: 12.5,
                color: 'rgba(255,255,255,0.5)',
                lineHeight: 1.65,
                maxWidth: 280,
                margin: 0
              }}
            >
              {t('landing.footer.description')}
            </p>
          </div>
          {footerGroups.map(([title, items]) => (
            <div key={title}>
              <div
                style={{ fontSize: 12, color: '#fff', fontWeight: 600, marginBottom: 14 }}
              >
                {title}
              </div>
              {items.map((it) => (
                <a
                  key={it}
                  href="#"
                  style={{
                    display: 'block',
                    fontSize: 12.5,
                    color: 'rgba(255,255,255,0.5)',
                    textDecoration: 'none',
                    padding: '5px 0'
                  }}
                >
                  {it}
                </a>
              ))}
            </div>
          ))}
        </div>
        <div
          style={{
            paddingTop: 24,
            paddingBottom: 24,
            borderTop: '1px solid rgba(255,255,255,0.06)',
            display: 'flex',
            justifyContent: 'space-between',
            fontSize: 11.5,
            color: 'rgba(255,255,255,0.4)'
          }}
        >
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
      style={{
        background: '#000',
        color: '#fff',
        minHeight: '100vh',
        fontFamily: 'Inter, "PingFang SC", system-ui, sans-serif'
      }}
    >
      <NavD />
      <HeroD />
      <BackedRow />
      <OperationsHub />
      <SecurityBand />
      <CaseStudy />
      <CtaFooter />
    </div>
  )
}
