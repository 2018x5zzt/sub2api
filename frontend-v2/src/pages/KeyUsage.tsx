import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  BarChart3,
  BookOpen,
  CalendarDays,
  Clock3,
  Eye,
  EyeOff,
  KeyRound,
  Search,
  ShieldCheck,
  Wallet
} from 'lucide-react'
import { LocaleSwitcher } from '@/components/layout/LocaleSwitcher'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { useAuthStore } from '@/stores/auth'

type DateRangeKey = 'today' | '7d' | '30d' | 'custom'

interface RateLimitUsage {
  window: string
  limit: number
  used: number
  remaining?: number
  reset_at?: string | null
}

interface UsageSummary {
  requests?: number
  input_tokens?: number
  output_tokens?: number
  cache_creation_tokens?: number
  cache_read_tokens?: number
  total_tokens?: number
  cost?: number
  actual_cost?: number
}

interface KeyUsageResponse {
  mode?: 'quota_limited' | 'unrestricted' | string
  isValid?: boolean
  status?: string
  planName?: string
  balance?: number
  remaining?: number
  expires_at?: string | null
  days_until_expiry?: number | null
  quota?: {
    limit: number
    used: number
    remaining: number
  }
  rate_limits?: RateLimitUsage[]
  subscription?: {
    daily_usage_usd?: number
    weekly_usage_usd?: number
    monthly_usage_usd?: number
    daily_limit_usd?: number
    weekly_limit_usd?: number
    monthly_limit_usd?: number
    expires_at?: string | null
  }
  usage?: {
    today?: UsageSummary
    total?: UsageSummary
    average_duration_ms?: number
    rpm?: number
    tpm?: number
  }
  model_stats?: Array<{
    model?: string
    requests?: number
    input_tokens?: number
    output_tokens?: number
    cache_creation_tokens?: number
    cache_read_tokens?: number
    total_tokens?: number
    cost?: number
    actual_cost?: number
  }>
}

interface RingItem {
  title: string
  amount: string
  pct: number
  icon: 'clock' | 'calendar' | 'wallet'
  resetAt?: string | null
  balanceOnly?: boolean
}

interface DetailRow {
  label: string
  value: string
  tone?: 'success' | 'warning' | 'danger' | 'neutral'
  icon: 'shield' | 'calendar' | 'wallet' | 'check'
}

const RANGES: DateRangeKey[] = ['today', '7d', '30d', 'custom']
const RING_COLORS = ['#ff5722', '#2f8f5e', '#a8761a', '#3d3d3a']

function formatUSD(value: number | null | undefined, digits = 2): string {
  if (value == null || Number.isNaN(Number(value))) return '-'
  return `$${Number(value).toFixed(digits)}`
}

function formatNumber(value: number | null | undefined): string {
  if (value == null || Number.isNaN(Number(value))) return '-'
  return Number(value).toLocaleString()
}

function formatDate(value: string | null | undefined, locale: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return date.toLocaleDateString(locale === 'zh' ? 'zh-CN' : 'en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

function rangeParams(range: DateRangeKey, customStart: string, customEnd: string): URLSearchParams {
  const params = new URLSearchParams()
  const end = new Date()
  const toDate = (date: Date) => date.toISOString().split('T')[0]

  if (range === 'custom') {
    if (customStart && customEnd) {
      params.set('start_date', customStart)
      params.set('end_date', customEnd)
    }
    return params
  }

  const start = new Date(end)
  if (range === '7d') start.setDate(start.getDate() - 7)
  if (range === '30d') start.setDate(start.getDate() - 30)

  params.set('start_date', toDate(range === 'today' ? end : start))
  params.set('end_date', toDate(end))
  return params
}

function getUsageTone(pct: number): 'success' | 'warning' | 'danger' {
  if (pct > 90) return 'danger'
  if (pct > 70) return 'warning'
  return 'success'
}

function toneClass(tone: DetailRow['tone']) {
  if (tone === 'danger') return 'text-red-700'
  if (tone === 'warning') return 'text-amber-700'
  if (tone === 'success') return 'text-green-700'
  return 'text-ink-1'
}

function resetLabel(resetAt: string | null | undefined, t: (key: string) => string): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - Date.now()
  if (diff <= 0) return t('keyUsage.resetNow')
  const days = Math.floor(diff / 86_400_000)
  const hours = Math.floor((diff % 86_400_000) / 3_600_000)
  const mins = Math.floor((diff % 3_600_000) / 60_000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

function expirySuffix(days: number | null | undefined, t: (key: string, options?: Record<string, unknown>) => string): string {
  if (days == null) return ''
  if (days > 0) return ` ${t('keyUsage.daysLeft', { days })}`
  if (days === 0) return ` ${t('keyUsage.todayExpires')}`
  return ''
}

function IconBadge({ icon, tone = 'neutral' }: { icon: DetailRow['icon']; tone?: DetailRow['tone'] }) {
  const cls =
    tone === 'danger'
      ? 'bg-red-50 text-red-700'
      : tone === 'warning'
        ? 'bg-amber-50 text-amber-700'
        : tone === 'success'
          ? 'bg-green-50 text-green-700'
          : 'bg-bg-2 text-ink-2'
  const Icon = icon === 'calendar' ? CalendarDays : icon === 'wallet' ? Wallet : icon === 'check' ? ShieldCheck : ShieldCheck
  return (
    <span className={`inline-flex h-8 w-8 items-center justify-center rounded-lg ${cls}`}>
      <Icon className="h-4 w-4" />
    </span>
  )
}

function ProgressRing({ ring, index }: { ring: RingItem; index: number }) {
  const color = RING_COLORS[index % RING_COLORS.length]
  const radius = 54
  const circumference = 2 * Math.PI * radius
  const pct = Math.max(0, Math.min(ring.pct, 100))
  const offset = ring.balanceOnly ? 0 : circumference - (pct / 100) * circumference
  const Icon = ring.icon === 'clock' ? Clock3 : ring.icon === 'calendar' ? CalendarDays : Wallet

  return (
    <Card className="p-6 card-hover">
      <div className="flex items-center justify-between gap-3">
        <div className="text-xs uppercase tracking-[0.16em] text-ink-3 font-mono">{ring.title}</div>
        <Icon className="h-4 w-4 text-ink-3" />
      </div>
      <div className="mt-6 flex justify-center">
        <div className="relative h-36 w-36">
          <svg className="h-36 w-36 -rotate-90" viewBox="0 0 128 128">
            <circle cx="64" cy="64" r={radius} stroke="var(--bg-3)" strokeWidth="10" fill="none" />
            <circle
              cx="64"
              cy="64"
              r={radius}
              stroke={color}
              strokeWidth="10"
              strokeLinecap="round"
              fill="none"
              strokeDasharray={circumference}
              strokeDashoffset={offset}
              style={{ transition: 'stroke-dashoffset 700ms cubic-bezier(0.4, 0, 0.2, 1)' }}
            />
          </svg>
          <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
            {ring.balanceOnly ? (
              <div className="font-display text-2xl text-ink-1">{ring.amount}</div>
            ) : (
              <>
                <div className="font-display text-3xl text-ink-1">{pct}%</div>
                <div className="mt-1 text-[11px] uppercase tracking-[0.14em] text-ink-3">{ring.amount}</div>
              </>
            )}
            {ring.resetAt && (
              <div className="mt-1 text-xs text-ink-3">reset {ring.resetAt}</div>
            )}
          </div>
        </div>
      </div>
    </Card>
  )
}

export default function KeyUsagePage() {
  const { t, i18n } = useTranslation()
  const publicSettings = useAuthStore((s) => s.publicSettings)
  const [apiKey, setApiKey] = useState('')
  const [keyVisible, setKeyVisible] = useState(false)
  const [range, setRange] = useState<DateRangeKey>('today')
  const [customStart, setCustomStart] = useState('')
  const [customEnd, setCustomEnd] = useState('')
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<KeyUsageResponse | null>(null)
  const [hasQueried, setHasQueried] = useState(false)

  const siteName = publicSettings?.site_name || 'XlabAPI'
  const siteLogo = publicSettings?.site_logo || '/logo.png'
  const docUrl = publicSettings?.doc_url || '/docs'

  const queryKey = async (nextRange = range) => {
    const key = apiKey.trim()
    if (!key) {
      toast.warning(t('keyUsage.enterApiKey') as string)
      return
    }

    const params = rangeParams(nextRange, customStart, customEnd)
    const qs = params.toString()
    setLoading(true)
    setHasQueried(true)
    try {
      const res = await fetch(`/v1/usage${qs ? `?${qs}` : ''}`, {
        headers: { Authorization: `Bearer ${key}` }
      })
      const body = await res.json().catch(() => null)
      if (!res.ok) {
        throw new Error(body?.error?.message || body?.message || `${t('keyUsage.queryFailed')} (${res.status})`)
      }
      setData(body as KeyUsageResponse)
      toast.success(t('keyUsage.querySuccess') as string)
    } catch (error) {
      setData(null)
      toast.error((error as Error)?.message || (t('keyUsage.queryFailedRetry') as string))
    } finally {
      setLoading(false)
    }
  }

  const statusInfo = useMemo(() => {
    if (!data) return null
    if (data.mode === 'quota_limited') {
      const statusText =
        data.status === 'active'
          ? 'Active'
          : data.status === 'quota_exhausted'
            ? 'Quota Exhausted'
            : data.status === 'expired'
              ? 'Expired'
              : data.status || 'Unknown'
      return {
        label: t('keyUsage.quotaMode') as string,
        statusText,
        active: data.isValid !== false && data.status === 'active'
      }
    }
    return {
      label: data.planName || (t('keyUsage.walletBalance') as string),
      statusText: 'Active',
      active: data.isValid !== false
    }
  }, [data, t])

  const ringItems = useMemo<RingItem[]>(() => {
    if (!data) return []
    const items: RingItem[] = []
    if (data.mode === 'quota_limited') {
      if (data.quota) {
        const pct = data.quota.limit > 0 ? Math.round((data.quota.used / data.quota.limit) * 100) : 0
        items.push({
          title: t('keyUsage.totalQuota') as string,
          amount: `${formatUSD(data.quota.used)} / ${formatUSD(data.quota.limit)}`,
          pct,
          icon: 'wallet'
        })
      }
      for (const rateLimit of data.rate_limits ?? []) {
        const windowLabels: Record<string, string> = {
          '5h': t('keyUsage.limit5h') as string,
          '1d': t('keyUsage.limitDaily') as string,
          '7d': t('keyUsage.limit7d') as string
        }
        const pct = rateLimit.limit > 0 ? Math.round((rateLimit.used / rateLimit.limit) * 100) : 0
        items.push({
          title: windowLabels[rateLimit.window] || rateLimit.window,
          amount: `${formatUSD(rateLimit.used)} / ${formatUSD(rateLimit.limit)}`,
          pct,
          icon: rateLimit.window === '5h' ? 'clock' : 'calendar',
          resetAt: resetLabel(rateLimit.reset_at, t as (key: string) => string)
        })
      }
      return items
    }

    const sub = data.subscription
    if (sub) {
      const limits = [
        [t('keyUsage.limitDaily') as string, sub.daily_usage_usd, sub.daily_limit_usd],
        [t('keyUsage.limitWeekly') as string, sub.weekly_usage_usd, sub.weekly_limit_usd],
        [t('keyUsage.limitMonthly') as string, sub.monthly_usage_usd, sub.monthly_limit_usd]
      ] as const
      for (const [label, used, limit] of limits) {
        if (limit && limit > 0) {
          items.push({
            title: label,
            amount: `${formatUSD(used)} / ${formatUSD(limit)}`,
            pct: Math.round(((used ?? 0) / limit) * 100),
            icon: 'calendar'
          })
        }
      }
    } else if (data.balance != null) {
      items.push({
        title: t('keyUsage.walletBalance') as string,
        amount: formatUSD(data.balance),
        pct: 0,
        icon: 'wallet',
        balanceOnly: true
      })
    }
    return items
  }, [data, t])

  const detailRows = useMemo<DetailRow[]>(() => {
    if (!data) return []
    const rows: DetailRow[] = []
    if (data.mode === 'quota_limited') {
      if (data.quota) {
        const pct = data.quota.limit > 0 ? (data.quota.used / data.quota.limit) * 100 : 0
        rows.push({
          icon: 'shield',
          label: t('keyUsage.remainingQuota') as string,
          value: formatUSD(data.quota.remaining),
          tone: data.quota.remaining <= 0 ? 'danger' : pct > 70 ? 'warning' : 'success'
        })
      }
      if (data.expires_at) {
        const days = data.days_until_expiry
        rows.push({
          icon: 'calendar',
          label: t('keyUsage.expiresAt') as string,
          value: `${formatDate(data.expires_at, i18n.language)}${expirySuffix(days, t as (key: string, options?: Record<string, unknown>) => string)}`
        })
      }
      for (const rateLimit of data.rate_limits ?? []) {
        const pct = rateLimit.limit > 0 ? (rateLimit.used / rateLimit.limit) * 100 : 0
        const reset = resetLabel(rateLimit.reset_at, t as (key: string) => string)
        rows.push({
          icon: 'wallet',
          label: `${t('keyUsage.usedQuota')} (${rateLimit.window.toUpperCase()})`,
          value: `${formatUSD(rateLimit.used)} / ${formatUSD(rateLimit.limit)}${reset ? ` / reset ${reset}` : ''}`,
          tone: getUsageTone(pct)
        })
      }
      return rows
    }

    rows.push({
      icon: 'check',
      label: t('keyUsage.subscriptionType') as string,
      value: data.planName || (t('keyUsage.walletBalance') as string)
    })
    const sub = data.subscription
    if (sub) {
      const limits = [
        ['D', sub.daily_usage_usd, sub.daily_limit_usd],
        ['W', sub.weekly_usage_usd, sub.weekly_limit_usd],
        ['M', sub.monthly_usage_usd, sub.monthly_limit_usd]
      ] as const
      for (const [label, used, limit] of limits) {
        if (limit && limit > 0) {
          const pct = ((used ?? 0) / limit) * 100
          rows.push({
            icon: 'wallet',
            label: `${t('keyUsage.usedQuota')} (${label})`,
            value: `${formatUSD(used)} / ${formatUSD(limit)}`,
            tone: getUsageTone(pct)
          })
        }
      }
      if (sub.expires_at) {
        rows.push({
          icon: 'calendar',
          label: t('keyUsage.subscriptionExpires') as string,
          value: formatDate(sub.expires_at, i18n.language)
        })
      }
    }
    rows.push({
      icon: 'shield',
      label: t('keyUsage.remainingQuota') as string,
      value: data.remaining != null ? formatUSD(data.remaining) : '-',
      tone: data.remaining == null ? 'neutral' : data.remaining <= 0 ? 'danger' : data.remaining < 10 ? 'warning' : 'success'
    })
    return rows
  }, [data, i18n.language, t])

  const usageCells = useMemo(() => {
    const usage = data?.usage
    if (!usage) return []
    const today = usage.today ?? {}
    const total = usage.total ?? {}
    return [
      [t('keyUsage.todayRequests'), formatNumber(today.requests)],
      [t('keyUsage.todayInputTokens'), formatNumber(today.input_tokens)],
      [t('keyUsage.todayOutputTokens'), formatNumber(today.output_tokens)],
      [t('keyUsage.todayTokens'), formatNumber(today.total_tokens)],
      [t('keyUsage.todayCacheCreation'), formatNumber(today.cache_creation_tokens)],
      [t('keyUsage.todayCacheRead'), formatNumber(today.cache_read_tokens)],
      [t('keyUsage.todayCost'), formatUSD(today.actual_cost, 4)],
      [t('keyUsage.rpmTpm'), `${usage.rpm ?? 0} / ${usage.tpm ?? 0}`],
      [t('keyUsage.totalRequests'), formatNumber(total.requests)],
      [t('keyUsage.totalInputTokens'), formatNumber(total.input_tokens)],
      [t('keyUsage.totalOutputTokens'), formatNumber(total.output_tokens)],
      [t('keyUsage.totalTokensLabel'), formatNumber(total.total_tokens)],
      [t('keyUsage.totalCacheCreation'), formatNumber(total.cache_creation_tokens)],
      [t('keyUsage.totalCacheRead'), formatNumber(total.cache_read_tokens)],
      [t('keyUsage.totalCost'), formatUSD(total.actual_cost, 4)],
      [t('keyUsage.avgDuration'), usage.average_duration_ms ? `${Math.round(usage.average_duration_ms)} ms` : '-']
    ]
  }, [data, t])

  const changeRange = (nextRange: DateRangeKey) => {
    setRange(nextRange)
    if (nextRange !== 'custom' && apiKey.trim() && hasQueried) {
      void queryKey(nextRange)
    }
  }

  return (
    <div className="min-h-screen bg-bg-0 text-ink-1">
      <header className="relative z-10 border-b border-line-2 bg-white/50 backdrop-blur">
        <nav className="container-bus flex h-20 items-center justify-between">
          <Link to="/" className="flex items-center gap-3">
            <img src={siteLogo} alt={siteName} className="h-9 w-9 rounded-lg border border-line-2 bg-bg-1 object-contain" />
            <span className="font-display text-lg font-semibold tracking-tight">{siteName}</span>
          </Link>
          <div className="flex items-center gap-2">
            <LocaleSwitcher compact />
            <a href={docUrl} target="_blank" rel="noreferrer" className="btn btn-ghost btn-icon" title={t('home.viewDocs') as string}>
              <BookOpen className="h-4 w-4" />
            </a>
          </div>
        </nav>
      </header>

      <main className="container-bus py-12">
        <section className="mx-auto max-w-3xl text-center">
          <div className="eyebrow justify-center">PUBLIC KEY INSPECTOR</div>
          <h1 className="mt-4 font-display text-4xl tracking-tight sm:text-5xl">{t('keyUsage.title')}</h1>
          <p className="mx-auto mt-4 max-w-xl text-ink-3">{t('keyUsage.subtitle')}</p>
        </section>

        <Card className="mx-auto mt-10 max-w-3xl p-4 sm:p-5">
          <div className="grid gap-3 sm:grid-cols-[1fr_auto]">
            <Input
              value={apiKey}
              type={keyVisible ? 'text' : 'password'}
              placeholder={t('keyUsage.placeholder') as string}
              onChange={(event) => setApiKey(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') void queryKey()
              }}
              leftIcon={<KeyRound className="h-4 w-4" />}
              rightAdornment={
                <button
                  type="button"
                  className="rounded-full p-1 text-ink-3 hover:text-ink-1"
                  onClick={() => setKeyVisible((visible) => !visible)}
                  aria-label={keyVisible ? 'Hide API key' : 'Show API key'}
                >
                  {keyVisible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              }
              className="h-11"
            />
            <Button className="h-11" variant="accent" loading={loading} onClick={() => void queryKey()}>
              {!loading && <Search className="h-4 w-4" />}
              {loading ? t('keyUsage.querying') : t('keyUsage.query')}
            </Button>
          </div>
          <p className="mt-3 text-center text-xs text-ink-3">{t('keyUsage.privacyNote')}</p>

          {hasQueried && (
            <div className="mt-5 flex flex-wrap items-center justify-center gap-2 border-t border-line-2 pt-4">
              <span className="text-xs uppercase tracking-[0.14em] text-ink-3 font-mono">{t('keyUsage.dateRange')}</span>
              {RANGES.map((key) => (
                <button
                  key={key}
                  type="button"
                  className={`pill-nav-item ${range === key ? 'pill-nav-item-active' : ''}`}
                  onClick={() => changeRange(key)}
                >
                  {t(`keyUsage.dateRange${key === 'today' ? 'Today' : key === '7d' ? '7d' : key === '30d' ? '30d' : 'Custom'}`)}
                </button>
              ))}
              {range === 'custom' && (
                <div className="flex flex-wrap items-center justify-center gap-2">
                  <input className="input h-9 w-36" type="date" value={customStart} onChange={(event) => setCustomStart(event.target.value)} />
                  <span className="text-ink-3">-</span>
                  <input className="input h-9 w-36" type="date" value={customEnd} onChange={(event) => setCustomEnd(event.target.value)} />
                  <Button size="sm" variant="accent" onClick={() => void queryKey()}>{t('keyUsage.apply')}</Button>
                </div>
              )}
            </div>
          )}
        </Card>

        {loading && (
          <div className="mx-auto mt-8 max-w-5xl space-y-4">
            <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
              {Array.from({ length: 3 }).map((_, index) => (
                <Card key={index} className="p-6">
                  <Skeleton className="h-4 w-24" />
                  <Skeleton className="mx-auto mt-6 h-36 w-36 rounded-full" />
                </Card>
              ))}
            </div>
            <Card className="p-6">
              <Skeleton className="h-5 w-32" />
              <Skeleton className="mt-5 h-24" />
            </Card>
          </div>
        )}

        {!loading && data && (
          <section className="mx-auto mt-8 max-w-5xl space-y-5">
            {statusInfo && (
              <div className="flex justify-center">
                <Badge tone={statusInfo.active ? 'success' : 'danger'} dot className="h-9 px-4">
                  {statusInfo.label} / {statusInfo.statusText}
                </Badge>
              </div>
            )}

            {ringItems.length > 0 && (
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
                {ringItems.map((ring, index) => (
                  <ProgressRing key={`${ring.title}-${index}`} ring={ring} index={index} />
                ))}
              </div>
            )}

            {detailRows.length > 0 && (
              <Card className="overflow-hidden">
                <div className="border-b border-line-2 px-6 py-4">
                  <div className="flex items-center gap-2 text-xs uppercase tracking-[0.16em] text-ink-3 font-mono">
                    <ShieldCheck className="h-4 w-4" />
                    {t('keyUsage.detailInfo')}
                  </div>
                </div>
                <div className="divide-y divide-line-1">
                  {detailRows.map((row) => (
                    <div key={`${row.label}-${row.value}`} className="flex items-center justify-between gap-4 px-6 py-4">
                      <div className="flex items-center gap-3">
                        <IconBadge icon={row.icon} tone={row.tone} />
                        <span className="text-sm text-ink-2">{row.label}</span>
                      </div>
                      <span className={`text-right font-mono text-sm font-semibold ${toneClass(row.tone)}`}>{row.value}</span>
                    </div>
                  ))}
                </div>
              </Card>
            )}

            {usageCells.length > 0 && (
              <Card className="overflow-hidden">
                <div className="border-b border-line-2 px-6 py-4">
                  <div className="flex items-center gap-2 text-xs uppercase tracking-[0.16em] text-ink-3 font-mono">
                    <BarChart3 className="h-4 w-4" />
                    {t('keyUsage.tokenStats')}
                  </div>
                </div>
                <div className="grid grid-cols-2 divide-x divide-y divide-line-1 sm:grid-cols-4">
                  {usageCells.map(([label, value]) => (
                    <div key={`${label}-${value}`} className="bg-white/40 px-5 py-4">
                      <div className="text-xs text-ink-3">{label}</div>
                      <div className="mt-1 font-mono text-sm font-semibold text-ink-1">{value}</div>
                    </div>
                  ))}
                </div>
              </Card>
            )}

            {(data.model_stats ?? []).length > 0 && (
              <Card className="overflow-hidden">
                <div className="border-b border-line-2 px-6 py-4">
                  <div className="text-xs uppercase tracking-[0.16em] text-ink-3 font-mono">{t('keyUsage.modelStats')}</div>
                </div>
                <Table>
                  <THead>
                    <TR>
                      <TH>{t('keyUsage.model')}</TH>
                      <TH className="text-right">{t('keyUsage.requests')}</TH>
                      <TH className="text-right">{t('keyUsage.inputTokens')}</TH>
                      <TH className="text-right">{t('keyUsage.outputTokens')}</TH>
                      <TH className="text-right">{t('keyUsage.cacheCreationTokens')}</TH>
                      <TH className="text-right">{t('keyUsage.cacheReadTokens')}</TH>
                      <TH className="text-right">{t('keyUsage.totalTokens')}</TH>
                      <TH className="text-right">{t('keyUsage.cost')}</TH>
                    </TR>
                  </THead>
                  <TBody>
                    {(data.model_stats ?? []).map((model, index) => (
                      <TR key={`${model.model || 'model'}-${index}`}>
                        <TD className="whitespace-nowrap font-mono text-xs">{model.model || '-'}</TD>
                        <TD className="text-right font-mono">{formatNumber(model.requests)}</TD>
                        <TD className="text-right font-mono">{formatNumber(model.input_tokens)}</TD>
                        <TD className="text-right font-mono">{formatNumber(model.output_tokens)}</TD>
                        <TD className="text-right font-mono">{formatNumber(model.cache_creation_tokens)}</TD>
                        <TD className="text-right font-mono">{formatNumber(model.cache_read_tokens)}</TD>
                        <TD className="text-right font-mono">{formatNumber(model.total_tokens)}</TD>
                        <TD className="text-right font-mono">{formatUSD(model.actual_cost ?? model.cost, 4)}</TD>
                      </TR>
                    ))}
                  </TBody>
                </Table>
              </Card>
            )}
          </section>
        )}
      </main>

      <footer className="border-t border-line-2 px-6 py-8 text-center text-sm text-ink-3">
        &copy; {new Date().getFullYear()} {siteName}. {t('home.footer.allRightsReserved')}
      </footer>
    </div>
  )
}
