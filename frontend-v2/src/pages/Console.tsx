/**
 * Console v4 — Plato (BusAPI dark + warm orange).
 * Direct port of /tmp/bus2api-handoff/bus2api/project/components/console-v4.jsx,
 * keeping the original mock data verbatim.
 */
import { useState, type ReactNode } from 'react'
import { SectionFrame } from '@/components/bus/SectionFrame'

const ORANGE = '#ff5722'
const LINE = 'rgba(255,255,255,0.10)'
const LINE_2 = 'rgba(255,255,255,0.06)'

function OrangeMark({ size = 8, mr = 10 }: { size?: number; mr?: number }) {
  return (
    <span
      style={{
        display: 'inline-block',
        width: size,
        height: size,
        background: ORANGE,
        marginRight: mr,
        verticalAlign: 'middle',
        flexShrink: 0
      }}
    />
  )
}

function Eyebrow({ children }: { children: ReactNode }) {
  return (
    <div
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        fontSize: 11,
        letterSpacing: '0.18em',
        color: 'rgba(255,255,255,0.5)',
        textTransform: 'uppercase',
        fontFamily: 'JetBrains Mono, monospace'
      }}
    >
      <OrangeMark />
      {children}
    </div>
  )
}

function ProjectSwitcher() {
  const [open, setOpen] = useState(false)
  const [proj, setProj] = useState<'production' | 'staging' | 'dev'>('production')
  const envs: Array<[typeof proj, string, string]> = [
    ['production', '生产环境', '5,221 req/min'],
    ['staging', '预发布', '102 req/min'],
    ['dev', '开发', '12 req/min']
  ]
  return (
    <div style={{ position: 'relative' }}>
      <button
        onClick={() => setOpen(!open)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '6px 14px 6px 8px',
          borderRadius: 999,
          border: `1px solid ${LINE}`,
          background: 'transparent',
          cursor: 'pointer',
          color: '#fff',
          fontFamily: 'inherit'
        }}
      >
        <span
          style={{
            width: 18,
            height: 18,
            borderRadius: 4,
            background: ORANGE,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: 10,
            fontWeight: 700
          }}
        >
          P
        </span>
        <span style={{ fontSize: 13 }}>Plato Inc.</span>
        <span
          style={{
            fontSize: 10.5,
            padding: '2px 6px',
            background: 'rgba(255,87,34,0.15)',
            color: ORANGE,
            borderRadius: 4,
            fontFamily: 'JetBrains Mono, monospace',
            letterSpacing: '0.02em'
          }}
        >
          {proj.toUpperCase()}
        </span>
        <span style={{ opacity: 0.5, fontSize: 10 }}>▾</span>
      </button>
      {open && (
        <div
          onClick={() => setOpen(false)}
          style={{ position: 'fixed', inset: 0, zIndex: 1 }}
        />
      )}
      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            left: 0,
            marginTop: 8,
            width: 280,
            background: '#000',
            border: `1px solid ${LINE}`,
            borderRadius: 8,
            zIndex: 2,
            padding: 6
          }}
        >
          <div
            style={{
              padding: '8px 10px',
              fontSize: 10.5,
              color: 'rgba(255,255,255,0.4)',
              fontFamily: 'JetBrains Mono, monospace',
              letterSpacing: '0.14em'
            }}
          >
            ENVIRONMENTS
          </div>
          {envs.map(([k, label, sub]) => (
            <div
              key={k}
              onClick={() => {
                setProj(k)
                setOpen(false)
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                padding: '10px 10px',
                cursor: 'pointer',
                borderRadius: 6,
                background: proj === k ? 'rgba(255,87,34,0.08)' : 'transparent'
              }}
            >
              <span
                style={{
                  width: 6,
                  height: 6,
                  borderRadius: 999,
                  background:
                    k === 'production'
                      ? '#34d399'
                      : k === 'staging'
                      ? '#fbbf24'
                      : 'rgba(255,255,255,0.3)'
                }}
              />
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13, color: '#fff' }}>{label}</div>
                <div
                  style={{
                    fontSize: 11,
                    color: 'rgba(255,255,255,0.45)',
                    fontFamily: 'JetBrains Mono, monospace'
                  }}
                >
                  {sub}
                </div>
              </div>
              {proj === k && <span style={{ color: ORANGE, fontSize: 12 }}>●</span>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function NavBar() {
  const [active, setActive] = useState('Overview')
  const tabs = ['Overview', 'Logs', 'Models', 'Routing', 'Keys', 'Billing', 'Docs']
  return (
    <div
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 50,
        background: 'rgba(0,0,0,0.86)',
        backdropFilter: 'blur(12px)',
        borderBottom: `1px solid ${LINE_2}`
      }}
    >
      <div
        style={{
          maxWidth: 1440,
          margin: '0 auto',
          padding: '18px 40px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 24
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <div style={{ width: 26, height: 26, position: 'relative' }}>
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
                inset: 6,
                border: '1.5px solid #fff',
                borderRadius: 2
              }}
            />
          </div>
          <span style={{ fontSize: 14.5, fontWeight: 600, letterSpacing: '-0.01em' }}>BUSAPI</span>
          <div style={{ width: 1, height: 18, background: LINE, margin: '0 6px' }} />
          <ProjectSwitcher />
        </div>

        <div
          style={{
            display: 'flex',
            gap: 2,
            padding: 4,
            borderRadius: 999,
            background: 'rgba(255,255,255,0.03)',
            border: `1px solid ${LINE}`
          }}
        >
          {tabs.map((t) => (
            <button
              key={t}
              onClick={() => setActive(t)}
              style={{
                padding: '8px 16px',
                fontSize: 13,
                color: active === t ? '#fff' : 'rgba(255,255,255,0.6)',
                background: active === t ? 'rgba(255,255,255,0.06)' : 'transparent',
                border: 'none',
                borderRadius: 999,
                cursor: 'pointer',
                fontFamily: 'inherit',
                letterSpacing: '-0.005em'
              }}
            >
              {t}
            </button>
          ))}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <button
            style={{
              width: 36,
              height: 36,
              borderRadius: 999,
              border: `1px solid ${LINE}`,
              background: 'transparent',
              color: '#fff',
              cursor: 'pointer',
              fontSize: 13
            }}
          >
            ?
          </button>
          <button
            style={{
              width: 36,
              height: 36,
              borderRadius: 999,
              border: `1px solid ${LINE}`,
              background: 'transparent',
              color: '#fff',
              cursor: 'pointer',
              position: 'relative'
            }}
          >
            <span style={{ fontSize: 13 }}>🔔</span>
            <span
              style={{
                position: 'absolute',
                top: 7,
                right: 8,
                width: 7,
                height: 7,
                borderRadius: 999,
                background: ORANGE,
                boxShadow: '0 0 0 2px #000'
              }}
            />
          </button>
          <div
            style={{
              width: 36,
              height: 36,
              borderRadius: 999,
              background: 'linear-gradient(135deg, #ff5722, #ff9966)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 12,
              fontWeight: 600
            }}
          >
            YL
          </div>
        </div>
      </div>
    </div>
  )
}

function PageHeader() {
  const [range, setRange] = useState('24h')
  return (
    <section style={{ position: 'relative', padding: '56px 0 36px', overflow: 'hidden' }}>
      <SectionFrame showTop={false} halfWidth={720} />
      <div
        style={{
          position: 'relative',
          maxWidth: 1440,
          margin: '0 auto',
          padding: '0 40px',
          zIndex: 5
        }}
      >
        <Eyebrow>Console / Overview</Eyebrow>
        <div
          style={{
            display: 'flex',
            alignItems: 'flex-end',
            justifyContent: 'space-between',
            marginTop: 16,
            gap: 32
          }}
        >
          <div>
            <h1
              style={{
                fontSize: 48,
                fontWeight: 500,
                lineHeight: 1.06,
                letterSpacing: '-0.025em',
                margin: 0,
                color: '#fff'
              }}
            >
              系统状态{' '}
              <span
                style={{
                  fontFamily: 'Georgia, serif',
                  fontStyle: 'italic',
                  color: 'rgba(255,255,255,0.55)',
                  fontWeight: 400
                }}
              >
                at a glance
              </span>
            </h1>
            <p
              style={{
                fontSize: 14.5,
                color: 'rgba(255,255,255,0.55)',
                margin: '12px 0 0',
                maxWidth: 560,
                lineHeight: 1.55
              }}
            >
              过去 24 小时 5,221,489 次调用，平均延迟 312ms。1 个上游节点降级（OpenAI us-east-1），自动 Failover 已启用。
            </p>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              style={{
                display: 'flex',
                padding: 3,
                gap: 2,
                borderRadius: 999,
                border: `1px solid ${LINE}`,
                background: 'rgba(255,255,255,0.02)'
              }}
            >
              {['1h', '24h', '7d', '30d'].map((r) => (
                <button
                  key={r}
                  onClick={() => setRange(r)}
                  style={{
                    padding: '6px 14px',
                    fontSize: 12.5,
                    border: 'none',
                    cursor: 'pointer',
                    background: range === r ? '#fff' : 'transparent',
                    color: range === r ? '#000' : 'rgba(255,255,255,0.65)',
                    borderRadius: 999,
                    fontFamily: 'JetBrains Mono, monospace',
                    letterSpacing: '-0.01em',
                    fontWeight: 500
                  }}
                >
                  {r}
                </button>
              ))}
            </div>
            <button
              style={{
                height: 36,
                padding: '0 16px',
                borderRadius: 999,
                background: 'transparent',
                border: `1px solid ${LINE}`,
                color: '#fff',
                fontSize: 13,
                cursor: 'pointer',
                display: 'inline-flex',
                alignItems: 'center',
                gap: 6
              }}
            >
              <span style={{ fontSize: 11 }}>↗</span> Export
            </button>
            <button
              style={{
                height: 36,
                padding: '0 18px',
                borderRadius: 999,
                background: ORANGE,
                border: 'none',
                color: '#fff',
                fontSize: 13,
                fontWeight: 500,
                cursor: 'pointer'
              }}
            >
              + 创建密钥
            </button>
          </div>
        </div>
      </div>
    </section>
  )
}

function Sparkline({
  data,
  color = ORANGE,
  height = 28
}: {
  data: number[]
  color?: string
  height?: number
}) {
  const w = 100
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const pts = data
    .map(
      (v, i) =>
        `${(i / (data.length - 1)) * w},${height - ((v - min) / range) * (height - 4) - 2}`
    )
    .join(' ')
  return (
    <svg
      width="100%"
      height={height}
      viewBox={`0 0 ${w} ${height}`}
      preserveAspectRatio="none"
      style={{ display: 'block' }}
    >
      <polyline
        fill="none"
        stroke={color}
        strokeWidth="1.2"
        points={pts}
        vectorEffect="non-scaling-stroke"
      />
      <polyline
        fill={color}
        fillOpacity="0.08"
        stroke="none"
        points={`0,${height} ${pts} ${w},${height}`}
      />
    </svg>
  )
}

interface KPIProps {
  label: string
  value: string
  unit?: string
  delta: string
  deltaTone?: 'pos' | 'neg' | 'neu'
  spark: number[]
  sparkColor?: string
}

function KPI({ label, value, unit, delta, deltaTone = 'pos', spark, sparkColor }: KPIProps) {
  return (
    <div style={{ padding: '24px 28px', position: 'relative' }}>
      <div
        style={{
          fontSize: 11,
          letterSpacing: '0.16em',
          textTransform: 'uppercase',
          color: 'rgba(255,255,255,0.45)',
          fontFamily: 'JetBrains Mono, monospace',
          marginBottom: 14
        }}
      >
        {label}
      </div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 12 }}>
        <span
          style={{
            fontSize: 42,
            fontWeight: 500,
            letterSpacing: '-0.025em',
            color: '#fff',
            lineHeight: 1
          }}
        >
          {value}
        </span>
        {unit && (
          <span style={{ fontSize: 16, color: 'rgba(255,255,255,0.4)', fontWeight: 400 }}>
            {unit}
          </span>
        )}
      </div>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 12
        }}
      >
        <span
          style={{
            fontSize: 12,
            color:
              deltaTone === 'pos'
                ? '#34d399'
                : deltaTone === 'neg'
                ? '#f87171'
                : 'rgba(255,255,255,0.5)',
            fontFamily: 'JetBrains Mono, monospace'
          }}
        >
          {deltaTone === 'pos' ? '↑' : deltaTone === 'neg' ? '↓' : '→'} {delta}
        </span>
        <div style={{ flex: 1, maxWidth: 110 }}>
          <Sparkline data={spark} color={sparkColor || ORANGE} />
        </div>
      </div>
    </div>
  )
}

function KPIBand() {
  const kpis: KPIProps[] = [
    {
      label: '总调用量 (24h)',
      value: '5.22',
      unit: 'M',
      delta: '+12.4% vs 昨日',
      spark: [22, 28, 24, 31, 38, 35, 42, 48, 45, 52, 58, 55]
    },
    {
      label: '总成本 (24h)',
      value: '$1,284',
      unit: '.50',
      delta: '+8.1% vs 昨日',
      deltaTone: 'neg',
      spark: [40, 42, 38, 45, 48, 52, 50, 58, 62, 60, 65, 68]
    },
    {
      label: '平均延迟 (P50)',
      value: '312',
      unit: 'ms',
      delta: '−18ms vs 昨日',
      spark: [380, 360, 350, 340, 335, 320, 318, 322, 315, 310, 308, 312]
    },
    {
      label: '错误率',
      value: '0.08',
      unit: '%',
      delta: '−0.03% vs 昨日',
      spark: [0.15, 0.14, 0.12, 0.13, 0.11, 0.1, 0.09, 0.1, 0.08, 0.07, 0.08, 0.08]
    }
  ]
  return (
    <section style={{ position: 'relative', overflow: 'hidden' }}>
      <SectionFrame showTop={false} showBottom={false} halfWidth={720} />
      <div
        style={{
          position: 'relative',
          maxWidth: 1440,
          margin: '0 auto',
          padding: '0 40px',
          zIndex: 5
        }}
      >
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(4, 1fr)',
            borderTop: `1px solid ${LINE}`,
            borderBottom: `1px solid ${LINE}`
          }}
        >
          {kpis.map((k, i) => (
            <div key={k.label} style={{ borderLeft: i === 0 ? 'none' : `1px solid ${LINE}` }}>
              <KPI {...k} />
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}

function SeriesChart() {
  const [metric, setMetric] = useState<'requests' | 'errors' | 'latency'>('requests')
  const [hover, setHover] = useState<number | null>(null)

  const reqs = [
    820, 920, 880, 760, 640, 540, 480, 520, 680, 920, 1180, 1340, 1480, 1520, 1620, 1700, 1780,
    1820, 1750, 1640, 1520, 1380, 1180, 980
  ]
  const errs = reqs.map((r) => Math.round(r * (0.005 + Math.random() * 0.012)))
  const data =
    metric === 'requests'
      ? reqs
      : metric === 'errors'
      ? errs
      : reqs.map((_, i) => 280 + Math.sin(i / 3) * 30 + Math.random() * 24)
  const ylabel = metric === 'requests' ? 'requests' : metric === 'errors' ? 'errors' : 'ms'

  const w = 1280
  const h = 280
  const padL = 56
  const padR = 24
  const padT = 24
  const padB = 36
  const chartW = w - padL - padR
  const chartH = h - padT - padB
  const max = Math.max(...data) * 1.15
  const bars = data.map((v, i) => {
    const x = padL + (i / data.length) * chartW
    const bw = (chartW / data.length) * 0.72
    const bh = (v / max) * chartH
    return { x, y: padT + chartH - bh, bw, bh, v }
  })
  const ticks = 4
  const gridY = Array.from({ length: ticks + 1 }, (_, i) => padT + (chartH / ticks) * i)
  const valForY = (y: number) => Math.round((1 - (y - padT) / chartH) * max)

  const buttons: Array<['requests' | 'errors' | 'latency', string]> = [
    ['requests', '调用量'],
    ['errors', '错误'],
    ['latency', '延迟']
  ]

  return (
    <section style={{ position: 'relative', padding: '36px 0 64px', overflow: 'hidden' }}>
      <SectionFrame showTop={false} showBottom={false} halfWidth={720} />
      <div
        style={{
          position: 'relative',
          maxWidth: 1440,
          margin: '0 auto',
          padding: '0 40px',
          zIndex: 5
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 24
          }}
        >
          <div>
            <Eyebrow>调用量 / 24 hours</Eyebrow>
            <div
              style={{
                fontSize: 22,
                color: '#fff',
                marginTop: 12,
                letterSpacing: '-0.015em'
              }}
            >
              <span
                style={{
                  fontFamily: 'Georgia, serif',
                  fontStyle: 'italic',
                  color: 'rgba(255,255,255,0.6)',
                  fontWeight: 400
                }}
              >
                Hourly
              </span>{' '}
              traffic distribution
            </div>
          </div>
          <div
            style={{
              display: 'flex',
              padding: 3,
              gap: 2,
              borderRadius: 999,
              border: `1px solid ${LINE}`,
              background: 'rgba(255,255,255,0.02)'
            }}
          >
            {buttons.map(([k, label]) => (
              <button
                key={k}
                onClick={() => setMetric(k)}
                style={{
                  padding: '6px 14px',
                  fontSize: 12.5,
                  border: 'none',
                  cursor: 'pointer',
                  background: metric === k ? '#fff' : 'transparent',
                  color: metric === k ? '#000' : 'rgba(255,255,255,0.65)',
                  borderRadius: 999,
                  fontFamily: 'inherit'
                }}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        <div style={{ position: 'relative' }}>
          <svg width="100%" viewBox={`0 0 ${w} ${h}`} style={{ display: 'block' }}>
            {gridY.map((y, i) => (
              <g key={i}>
                <line x1={padL} x2={w - padR} y1={y} y2={y} stroke={LINE_2} strokeWidth="1" />
                <text
                  x={padL - 12}
                  y={y + 4}
                  fontSize="10"
                  fill="rgba(255,255,255,0.4)"
                  textAnchor="end"
                  fontFamily="JetBrains Mono, monospace"
                >
                  {valForY(y).toLocaleString()}
                </text>
              </g>
            ))}
            {bars.map((b, i) => (
              <g
                key={i}
                onMouseEnter={() => setHover(i)}
                onMouseLeave={() => setHover(null)}
                style={{ cursor: 'crosshair' }}
              >
                <rect x={b.x - 4} y={padT} width={b.bw + 8} height={chartH} fill="transparent" />
                <rect
                  x={b.x}
                  y={b.y}
                  width={b.bw}
                  height={b.bh}
                  fill={hover === i ? ORANGE : 'rgba(255,87,34,0.55)'}
                />
                {hover === i && <circle cx={b.x + b.bw / 2} cy={b.y - 6} r="2" fill="#fff" />}
              </g>
            ))}
            {[0, 6, 12, 18, 23].map((i) => {
              const b = bars[i]
              return (
                <text
                  key={i}
                  x={b.x + b.bw / 2}
                  y={h - 12}
                  fontSize="10"
                  fill="rgba(255,255,255,0.45)"
                  textAnchor="middle"
                  fontFamily="JetBrains Mono, monospace"
                >
                  {String(i).padStart(2, '0')}:00
                </text>
              )
            })}
            {hover != null && (() => {
              const b = bars[hover]
              const tx = b.x + b.bw / 2
              return (
                <g>
                  <line
                    x1={tx}
                    x2={tx}
                    y1={padT}
                    y2={padT + chartH}
                    stroke="rgba(255,255,255,0.25)"
                    strokeDasharray="2 2"
                  />
                  <g transform={`translate(${tx}, ${b.y - 18})`}>
                    <rect
                      x="-44"
                      y="-22"
                      width="88"
                      height="22"
                      rx="4"
                      fill="#000"
                      stroke={ORANGE}
                      strokeWidth="0.8"
                    />
                    <text
                      x="0"
                      y="-7"
                      fontSize="11"
                      fill="#fff"
                      textAnchor="middle"
                      fontFamily="JetBrains Mono, monospace"
                    >
                      {b.v.toLocaleString()} {ylabel}
                    </text>
                  </g>
                </g>
              )
            })()}
          </svg>
        </div>
      </div>
    </section>
  )
}

function ModelHealth() {
  const models = [
    {
      name: 'Claude Sonnet 4.5',
      vendor: 'Anthropic',
      status: 'healthy',
      latency: 284,
      share: 32.4,
      color: '#d97757',
      requests: '1.69M'
    },
    {
      name: 'GPT-5',
      vendor: 'OpenAI',
      status: 'degraded',
      latency: 612,
      share: 24.8,
      color: '#10a37f',
      requests: '1.29M',
      alert: 'us-east-1 延迟升高'
    },
    {
      name: 'Gemini 2.5 Pro',
      vendor: 'Google',
      status: 'healthy',
      latency: 198,
      share: 18.2,
      color: '#4285f4',
      requests: '951K'
    },
    {
      name: 'DeepSeek V4',
      vendor: 'DeepSeek',
      status: 'healthy',
      latency: 156,
      share: 14.6,
      color: '#1a73e8',
      requests: '763K'
    },
    {
      name: 'Qwen3 Max',
      vendor: 'Alibaba',
      status: 'healthy',
      latency: 232,
      share: 6.8,
      color: '#615ced',
      requests: '355K'
    },
    {
      name: 'Llama 3.1 405B',
      vendor: 'Meta',
      status: 'healthy',
      latency: 318,
      share: 3.2,
      color: '#0866ff',
      requests: '167K'
    }
  ] as const

  const events: Array<{ t: string; tone: 'warn' | 'ok' | 'info'; msg: string; sub: string }> = [
    { t: '2m', tone: 'warn', msg: 'gpt-5/us-east-1 → claude-4.5/us-west-2', sub: '失败率 4.2% 触发降级' },
    { t: '14m', tone: 'ok', msg: 'gpt-5/us-east-1 已恢复', sub: '上游延迟回落至 230ms' },
    { t: '38m', tone: 'warn', msg: 'gemini-2.5/asia → gemini-2.5/us-central', sub: '区域配额占满' },
    { t: '1h', tone: 'info', msg: '新增 Failover 路由：deepseek-v4', sub: 'priority=3 retry=2' },
    { t: '2h', tone: 'ok', msg: '所有上游健康度 100%', sub: '全部恢复正常' }
  ]

  return (
    <section style={{ position: 'relative', padding: '64px 0', overflow: 'hidden' }}>
      <SectionFrame halfWidth={720} />
      <div
        style={{
          position: 'relative',
          maxWidth: 1440,
          margin: '0 auto',
          padding: '0 40px',
          zIndex: 5
        }}
      >
        <div style={{ display: 'grid', gridTemplateColumns: '1.4fr 1fr', gap: 0 }}>
          <div style={{ borderRight: `1px solid ${LINE}`, paddingRight: 48 }}>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 28
              }}
            >
              <div>
                <Eyebrow>Top Models / by traffic</Eyebrow>
                <div
                  style={{
                    fontSize: 22,
                    color: '#fff',
                    marginTop: 12,
                    letterSpacing: '-0.015em'
                  }}
                >
                  <span
                    style={{
                      fontFamily: 'Georgia, serif',
                      fontStyle: 'italic',
                      color: 'rgba(255,255,255,0.6)',
                      fontWeight: 400
                    }}
                  >
                    Model
                  </span>{' '}
                  health & distribution
                </div>
              </div>
              <a href="#" style={{ fontSize: 12.5, color: ORANGE, textDecoration: 'none' }}>
                查看全部 →
              </a>
            </div>

            <div>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: '2fr 80px 80px 100px 80px',
                  gap: 16,
                  padding: '8px 0',
                  borderBottom: `1px solid ${LINE_2}`,
                  fontSize: 10.5,
                  letterSpacing: '0.14em',
                  color: 'rgba(255,255,255,0.4)',
                  textTransform: 'uppercase',
                  fontFamily: 'JetBrains Mono, monospace'
                }}
              >
                <div>Model</div>
                <div style={{ textAlign: 'right' }}>Status</div>
                <div style={{ textAlign: 'right' }}>Latency</div>
                <div style={{ textAlign: 'right' }}>Requests</div>
                <div style={{ textAlign: 'right' }}>Share</div>
              </div>
              {models.map((m, i) => (
                <div
                  key={m.name}
                  style={{
                    display: 'grid',
                    gridTemplateColumns: '2fr 80px 80px 100px 80px',
                    gap: 16,
                    padding: '14px 0',
                    borderBottom: i === models.length - 1 ? 'none' : `1px solid ${LINE_2}`,
                    alignItems: 'center',
                    position: 'relative'
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                    <span
                      style={{
                        width: 20,
                        height: 20,
                        borderRadius: 4,
                        background: m.color,
                        opacity: 0.9,
                        display: 'inline-flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: 10,
                        fontWeight: 700,
                        color: '#fff'
                      }}
                    >
                      {m.name[0]}
                    </span>
                    <div>
                      <div style={{ fontSize: 13.5, color: '#fff', fontWeight: 500 }}>{m.name}</div>
                      <div
                        style={{
                          fontSize: 11,
                          color: 'rgba(255,255,255,0.45)',
                          marginTop: 2
                        }}
                      >
                        {m.vendor}
                        {'alert' in m && m.alert ? (
                          <span style={{ color: '#fbbf24', marginLeft: 8 }}>· {m.alert}</span>
                        ) : null}
                      </div>
                    </div>
                  </div>
                  <div
                    style={{
                      textAlign: 'right',
                      fontSize: 12,
                      fontFamily: 'JetBrains Mono, monospace'
                    }}
                  >
                    <span
                      style={{
                        display: 'inline-flex',
                        alignItems: 'center',
                        gap: 6,
                        color: m.status === 'healthy' ? '#34d399' : '#fbbf24'
                      }}
                    >
                      <span
                        style={{ width: 6, height: 6, borderRadius: 999, background: 'currentColor' }}
                      />
                      {m.status}
                    </span>
                  </div>
                  <div
                    style={{
                      textAlign: 'right',
                      fontSize: 13,
                      color: '#fff',
                      fontFamily: 'JetBrains Mono, monospace'
                    }}
                  >
                    {m.latency}
                    <span style={{ color: 'rgba(255,255,255,0.4)', fontSize: 11 }}>ms</span>
                  </div>
                  <div
                    style={{
                      textAlign: 'right',
                      fontSize: 13,
                      color: 'rgba(255,255,255,0.85)',
                      fontFamily: 'JetBrains Mono, monospace'
                    }}
                  >
                    {m.requests}
                  </div>
                  <div
                    style={{
                      textAlign: 'right',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'flex-end',
                      gap: 8
                    }}
                  >
                    <div
                      style={{
                        width: 36,
                        height: 4,
                        background: 'rgba(255,255,255,0.06)',
                        position: 'relative'
                      }}
                    >
                      <div
                        style={{
                          position: 'absolute',
                          left: 0,
                          top: 0,
                          bottom: 0,
                          width: `${m.share * 2}%`,
                          background: ORANGE,
                          opacity: 0.7
                        }}
                      />
                    </div>
                    <span
                      style={{
                        fontSize: 12,
                        color: 'rgba(255,255,255,0.85)',
                        fontFamily: 'JetBrains Mono, monospace',
                        minWidth: 38,
                        textAlign: 'right'
                      }}
                    >
                      {m.share}%
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div style={{ paddingLeft: 48 }}>
            <Eyebrow>Failover / live</Eyebrow>
            <div
              style={{
                fontSize: 22,
                color: '#fff',
                marginTop: 12,
                letterSpacing: '-0.015em',
                marginBottom: 28
              }}
            >
              <span
                style={{
                  fontFamily: 'Georgia, serif',
                  fontStyle: 'italic',
                  color: 'rgba(255,255,255,0.6)',
                  fontWeight: 400
                }}
              >
                Routing
              </span>{' '}
              events
            </div>

            <div
              style={{
                display: 'flex',
                gap: 20,
                marginBottom: 28,
                padding: '20px 0',
                borderTop: `1px solid ${LINE_2}`,
                borderBottom: `1px solid ${LINE_2}`
              }}
            >
              <div style={{ flex: 1 }}>
                <div
                  style={{
                    fontSize: 32,
                    fontWeight: 500,
                    letterSpacing: '-0.025em',
                    color: '#fff'
                  }}
                >
                  5/6
                </div>
                <div
                  style={{
                    fontSize: 11,
                    color: 'rgba(255,255,255,0.45)',
                    marginTop: 4,
                    fontFamily: 'JetBrains Mono, monospace'
                  }}
                >
                  healthy upstreams
                </div>
              </div>
              <div style={{ flex: 1 }}>
                <div
                  style={{
                    fontSize: 32,
                    fontWeight: 500,
                    letterSpacing: '-0.025em',
                    color: '#fff'
                  }}
                >
                  847
                </div>
                <div
                  style={{
                    fontSize: 11,
                    color: 'rgba(255,255,255,0.45)',
                    marginTop: 4,
                    fontFamily: 'JetBrains Mono, monospace'
                  }}
                >
                  failovers (24h)
                </div>
              </div>
            </div>

            <div
              style={{
                fontSize: 11,
                letterSpacing: '0.16em',
                textTransform: 'uppercase',
                color: 'rgba(255,255,255,0.5)',
                fontFamily: 'JetBrains Mono, monospace',
                marginBottom: 14
              }}
            >
              Recent events
            </div>
            <div>
              {events.map((e, i) => (
                <div
                  key={i}
                  style={{
                    display: 'flex',
                    gap: 14,
                    padding: '12px 0',
                    borderTop: i === 0 ? 'none' : `1px solid ${LINE_2}`
                  }}
                >
                  <span
                    style={{
                      fontSize: 11,
                      color: 'rgba(255,255,255,0.4)',
                      fontFamily: 'JetBrains Mono, monospace',
                      width: 32,
                      flexShrink: 0,
                      paddingTop: 2
                    }}
                  >
                    {e.t}
                  </span>
                  <span
                    style={{
                      width: 6,
                      height: 6,
                      borderRadius: 999,
                      marginTop: 7,
                      flexShrink: 0,
                      background: e.tone === 'warn' ? '#fbbf24' : e.tone === 'ok' ? '#34d399' : ORANGE
                    }}
                  />
                  <div style={{ flex: 1 }}>
                    <div
                      style={{
                        fontSize: 13,
                        color: '#fff',
                        fontFamily: 'JetBrains Mono, monospace',
                        letterSpacing: '-0.005em'
                      }}
                    >
                      {e.msg}
                    </div>
                    <div
                      style={{
                        fontSize: 11.5,
                        color: 'rgba(255,255,255,0.5)',
                        marginTop: 3
                      }}
                    >
                      {e.sub}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

function RecentRequests() {
  const rows = [
    { id: 'req_4f8a2c91', model: 'claude-sonnet-4.5', status: 200, latency: 287, tokens: '1,284', cost: '0.0192', time: '14:32:18' },
    { id: 'req_4f8a2c8e', model: 'gpt-5', status: 200, latency: 412, tokens: '892', cost: '0.0089', time: '14:32:17' },
    { id: 'req_4f8a2c8d', model: 'gemini-2.5-pro', status: 200, latency: 198, tokens: '2,104', cost: '0.0042', time: '14:32:17' },
    { id: 'req_4f8a2c8c', model: 'gpt-5', status: 429, latency: 1820, tokens: '0', cost: '0.0000', time: '14:32:16', err: 'rate_limit' },
    { id: 'req_4f8a2c8b', model: 'deepseek-v4', status: 200, latency: 156, tokens: '648', cost: '0.0006', time: '14:32:16' },
    { id: 'req_4f8a2c8a', model: 'claude-sonnet-4.5', status: 200, latency: 312, tokens: '4,210', cost: '0.0631', time: '14:32:15' },
    { id: 'req_4f8a2c89', model: 'gpt-5', status: 500, latency: 5024, tokens: '0', cost: '0.0000', time: '14:32:14', err: 'upstream_timeout' },
    { id: 'req_4f8a2c88', model: 'qwen3-max', status: 200, latency: 232, tokens: '892', cost: '0.0027', time: '14:32:13' },
    { id: 'req_4f8a2c87', model: 'claude-sonnet-4.5', status: 200, latency: 268, tokens: '1,648', cost: '0.0247', time: '14:32:12' }
  ]
  return (
    <section style={{ position: 'relative', padding: '64px 0 96px', overflow: 'hidden' }}>
      <SectionFrame showBottom={false} halfWidth={720} />
      <div
        style={{
          position: 'relative',
          maxWidth: 1440,
          margin: '0 auto',
          padding: '0 40px',
          zIndex: 5
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'flex-end',
            justifyContent: 'space-between',
            marginBottom: 28
          }}
        >
          <div>
            <Eyebrow>Recent / live tail</Eyebrow>
            <div
              style={{
                fontSize: 22,
                color: '#fff',
                marginTop: 12,
                letterSpacing: '-0.015em'
              }}
            >
              <span
                style={{
                  fontFamily: 'Georgia, serif',
                  fontStyle: 'italic',
                  color: 'rgba(255,255,255,0.6)',
                  fontWeight: 400
                }}
              >
                Last
              </span>{' '}
              100 requests
            </div>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: 8,
                fontSize: 12,
                color: 'rgba(255,255,255,0.6)',
                fontFamily: 'JetBrains Mono, monospace'
              }}
            >
              <span
                style={{
                  width: 7,
                  height: 7,
                  borderRadius: 999,
                  background: '#34d399',
                  boxShadow: '0 0 0 3px rgba(52,211,153,0.18)',
                  animation: 'pulseDot 1.6s ease-in-out infinite'
                }}
              />
              streaming
            </span>
            <button
              style={{
                height: 32,
                padding: '0 14px',
                borderRadius: 999,
                background: 'transparent',
                border: `1px solid ${LINE}`,
                color: '#fff',
                fontSize: 12.5,
                cursor: 'pointer'
              }}
            >
              Pause
            </button>
            <a href="#" style={{ fontSize: 12.5, color: ORANGE, textDecoration: 'none' }}>
              所有日志 →
            </a>
          </div>
        </div>

        <div
          style={{
            border: `1px solid ${LINE}`,
            fontFamily: 'JetBrains Mono, monospace',
            fontSize: 12.5
          }}
        >
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '110px 1fr 200px 90px 90px 110px 100px',
              padding: '12px 20px',
              background: 'rgba(255,255,255,0.02)',
              borderBottom: `1px solid ${LINE}`,
              fontSize: 10.5,
              letterSpacing: '0.14em',
              textTransform: 'uppercase',
              color: 'rgba(255,255,255,0.45)'
            }}
          >
            <div>Time</div>
            <div>Request ID</div>
            <div>Model</div>
            <div style={{ textAlign: 'right' }}>Status</div>
            <div style={{ textAlign: 'right' }}>Latency</div>
            <div style={{ textAlign: 'right' }}>Tokens</div>
            <div style={{ textAlign: 'right' }}>Cost</div>
          </div>
          {rows.map((r, i) => {
            const ok = r.status < 300
            return (
              <div
                key={r.id}
                style={{
                  display: 'grid',
                  gridTemplateColumns: '110px 1fr 200px 90px 90px 110px 100px',
                  padding: '11px 20px',
                  borderBottom: i === rows.length - 1 ? 'none' : `1px solid ${LINE_2}`,
                  color: 'rgba(255,255,255,0.85)',
                  cursor: 'pointer',
                  transition: 'background 0.1s'
                }}
                onMouseEnter={(e) => (e.currentTarget.style.background = 'rgba(255,87,34,0.04)')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                <div style={{ color: 'rgba(255,255,255,0.5)' }}>{r.time}</div>
                <div style={{ color: 'rgba(255,255,255,0.85)' }}>{r.id}</div>
                <div style={{ color: '#fff' }}>
                  {r.model}
                  {r.err && <span style={{ color: '#f87171', marginLeft: 10 }}>· {r.err}</span>}
                </div>
                <div style={{ textAlign: 'right' }}>
                  <span
                    style={{
                      display: 'inline-block',
                      padding: '2px 8px',
                      borderRadius: 3,
                      fontSize: 11.5,
                      color: ok ? '#34d399' : '#f87171',
                      background: ok ? 'rgba(52,211,153,0.10)' : 'rgba(248,113,113,0.12)'
                    }}
                  >
                    {r.status}
                  </span>
                </div>
                <div
                  style={{
                    textAlign: 'right',
                    color: r.latency > 1000 ? '#fbbf24' : 'rgba(255,255,255,0.85)'
                  }}
                >
                  {r.latency}ms
                </div>
                <div style={{ textAlign: 'right', color: 'rgba(255,255,255,0.7)' }}>{r.tokens}</div>
                <div style={{ textAlign: 'right', color: 'rgba(255,255,255,0.85)' }}>${r.cost}</div>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}

export default function Console() {
  return (
    <div
      style={{
        background: '#000',
        color: '#fff',
        minHeight: '100vh',
        fontFamily: 'Inter, "PingFang SC", system-ui, sans-serif'
      }}
    >
      <NavBar />
      <PageHeader />
      <KPIBand />
      <SeriesChart />
      <ModelHealth />
      <RecentRequests />
    </div>
  )
}
