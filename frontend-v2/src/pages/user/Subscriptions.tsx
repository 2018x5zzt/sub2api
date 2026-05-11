import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { BadgeCheck, CalendarClock } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { subscriptionsAPI } from '@/api/subscriptions'
import type { UserSubscription } from '@/types'

function money(value: unknown, precision = 4) {
  const n = Number(value || 0)
  return `$${(Number.isFinite(n) ? n : 0).toFixed(precision)}`
}

function dateText(value: string | null | undefined) {
  if (!value) return ''
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleDateString()
}

function expiryLabel(sub: UserSubscription, t: ReturnType<typeof useTranslation>['t']) {
  if (!sub.expires_at) return t('userSubscriptions.noExpiration') as string
  const date = dateText(sub.expires_at)
  if (sub.status === 'expired' || new Date(sub.expires_at).getTime() < Date.now()) {
    return t('v2Common.expiredOn', { date }) as string
  }
  return t('userSubscriptions.expiresOn', { date }) as string
}

function statusTone(status: UserSubscription['status']) {
  if (status === 'active') return 'success' as const
  if (status === 'expired') return 'warning' as const
  if (status === 'revoked') return 'danger' as const
  return 'neutral' as const
}

function hasLimit(limit: number | null | undefined) {
  return typeof limit === 'number' && limit > 0
}

function UsageLine({
  label,
  used,
  limit
}: {
  label: string
  used: number | undefined
  limit: number | null | undefined
}) {
  const limited = hasLimit(limit)
  const safeUsed = Number(used || 0)
  const pct = limited ? Math.min((safeUsed / Number(limit)) * 100, 100) : 0
  const tone = limited && pct >= 90 ? 'bg-signal-err' : limited && pct >= 70 ? 'bg-signal-warn' : 'bg-orange'

  return (
    <div className="rounded-lg border border-line-1 bg-bg-2 p-3">
      <div className="flex items-baseline justify-between gap-3">
        <span className="text-sm font-medium text-ink-1">{label}</span>
        <span className="font-mono text-sm text-ink-1">
          {money(safeUsed)}
          {limited ? <span className="text-ink-3"> / {money(limit, 2)}</span> : null}
        </span>
      </div>
      {limited && (
        <div className="mt-2 h-1.5 overflow-hidden rounded-full bg-bg-3">
          <div className={`h-full ${tone}`} style={{ width: `${pct}%` }} />
        </div>
      )}
    </div>
  )
}

function subscriptionTitle(sub: UserSubscription) {
  const name = sub.group?.name || `Group #${sub.group_id}`
  return name.startsWith('【订阅】') ? name : `【订阅】${name}`
}

function SubscriptionCard({ sub }: { sub: UserSubscription }) {
  const { t } = useTranslation()
  const noLimits = !hasLimit(sub.group?.daily_limit_usd) && !hasLimit(sub.group?.weekly_limit_usd) && !hasLimit(sub.group?.monthly_limit_usd)

  return (
    <Card className="p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <BadgeCheck className="h-4 w-4 shrink-0 text-orange" />
            <h2 className="truncate text-base font-medium text-ink-1">{subscriptionTitle(sub)}</h2>
          </div>
          <div className="mt-2 flex items-center gap-1.5 text-xs text-ink-3">
            <CalendarClock className="h-3.5 w-3.5" />
            <span className="font-mono">{expiryLabel(sub, t)}</span>
          </div>
        </div>
        <Badge tone={statusTone(sub.status)}>{t(`userSubscriptions.status.${sub.status}`)}</Badge>
      </div>

      <div className="mt-4 grid gap-3">
        <UsageLine label={t('userSubscriptions.daily') as string} used={sub.daily_usage_usd} limit={sub.group?.daily_limit_usd} />
        <UsageLine label={t('userSubscriptions.weekly') as string} used={sub.weekly_usage_usd} limit={sub.group?.weekly_limit_usd} />
        <UsageLine label={t('userSubscriptions.monthly') as string} used={sub.monthly_usage_usd} limit={sub.group?.monthly_limit_usd} />
      </div>

      {noLimits && (
        <div className="mt-3 rounded-lg border border-line-1 bg-bg-2 px-3 py-2 text-sm text-ink-3">
          {t('userSubscriptions.unlimitedDesc')}
        </div>
      )}
    </Card>
  )
}

export default function SubscriptionsPage() {
  const { t } = useTranslation()
  const query = useQuery({
    queryKey: ['my-subscriptions'],
    queryFn: () => subscriptionsAPI.getMySubscriptions()
  })

  const subscriptions = query.data ?? []

  return (
    <>
      <PageHeader
        title={t('userSubscriptions.title')}
        description={t('userSubscriptions.description') as string}
      />

      {query.isLoading ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="p-5 space-y-4">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
            </Card>
          ))}
        </div>
      ) : subscriptions.length > 0 ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {subscriptions.map((sub) => (
            <SubscriptionCard key={sub.id} sub={sub} />
          ))}
        </div>
      ) : (
        <Card className="p-12 text-center">
          <div className="text-base font-medium text-ink-1">{t('userSubscriptions.noActiveSubscriptions')}</div>
          <p className="mt-1 text-sm text-ink-3">{t('userSubscriptions.noActiveSubscriptionsDesc')}</p>
        </Card>
      )}
    </>
  )
}
