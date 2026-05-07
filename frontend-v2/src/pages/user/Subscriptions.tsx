import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { CalendarClock, BadgeCheck } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { subscriptionsAPI } from '@/api/subscriptions'
import type { UserSubscription } from '@/types'

function pct(used: number, limit: number | null | undefined): number {
  if (!limit || limit <= 0) return 0
  return Math.min(((used || 0) / limit) * 100, 100)
}

function progressTone(used: number, limit: number | null | undefined): string {
  if (!limit || limit <= 0) return 'bg-line-3'
  const p = (used / limit) * 100
  if (p >= 90) return 'bg-signal-err'
  if (p >= 70) return 'bg-signal-warn'
  return 'bg-signal-ok'
}

function ProgressRow({
  label,
  used,
  limit
}: {
  label: string
  used: number | undefined
  limit: number | null | undefined
}) {
  const u = used ?? 0
  const p = pct(u, limit)
  return (
    <div>
      <div className="flex items-baseline justify-between text-xs text-ink-3 mb-1">
        <span className="font-mono uppercase tracking-wider">{label}</span>
        <span className="font-mono text-ink-2">
          ${u.toFixed(4)}
          {limit ? <span className="text-ink-3"> / ${limit.toFixed(2)}</span> : null}
        </span>
      </div>
      <div className="h-1.5 bg-bg-3 rounded-full overflow-hidden">
        <div
          className={`h-full transition-all ${progressTone(u, limit)}`}
          style={{ width: `${p}%` }}
        />
      </div>
    </div>
  )
}

function expiryInfo(expiresAt: string | null) {
  if (!expiresAt) return { label: '∞', tone: 'text-ink-3' }
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - Date.now()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  const dateStr = expires.toLocaleDateString()
  if (days < 0) return { label: `Expired · ${dateStr}`, tone: 'text-signal-err' }
  if (days === 0) return { label: `${dateStr} (today)`, tone: 'text-signal-warn' }
  if (days === 1) return { label: `${dateStr} (tomorrow)`, tone: 'text-signal-warn' }
  if (days <= 3) return { label: `${days}d · ${dateStr}`, tone: 'text-signal-err' }
  if (days <= 7) return { label: `${days}d · ${dateStr}`, tone: 'text-signal-warn' }
  return { label: `${days}d · ${dateStr}`, tone: 'text-ink-2' }
}

function statusTone(s: UserSubscription['status']) {
  switch (s) {
    case 'active':
      return 'success' as const
    case 'expired':
      return 'warning' as const
    case 'revoked':
      return 'danger' as const
    default:
      return 'neutral' as const
  }
}

function SubscriptionCard({ sub }: { sub: UserSubscription }) {
  const expiry = expiryInfo(sub.expires_at)
  const groupName = sub.group?.name || `Group #${sub.group_id}`
  return (
    <Card className="p-5 space-y-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2 mb-1.5">
            <BadgeCheck className="h-4 w-4 text-orange shrink-0" />
            <h3 className="font-medium text-ink-1 truncate">{groupName}</h3>
          </div>
          <div className="flex items-center gap-1.5 text-xs">
            <CalendarClock className="h-3.5 w-3.5 text-ink-3" />
            <span className={`font-mono ${expiry.tone}`}>{expiry.label}</span>
          </div>
        </div>
        <Badge tone={statusTone(sub.status)}>{sub.status}</Badge>
      </div>

      <div className="space-y-3 pt-1">
        {sub.group?.daily_limit_usd != null && (
          <ProgressRow label="Daily" used={sub.daily_usage_usd} limit={sub.group.daily_limit_usd} />
        )}
        {sub.group?.weekly_limit_usd != null && (
          <ProgressRow label="Weekly" used={sub.weekly_usage_usd} limit={sub.group.weekly_limit_usd} />
        )}
        {sub.group?.monthly_limit_usd != null && (
          <ProgressRow label="Monthly" used={sub.monthly_usage_usd} limit={sub.group.monthly_limit_usd} />
        )}
        {!sub.group?.daily_limit_usd &&
          !sub.group?.weekly_limit_usd &&
          !sub.group?.monthly_limit_usd && (
            <div className="text-xs text-ink-3 font-mono">No usage limits configured</div>
          )}
      </div>
    </Card>
  )
}

export default function SubscriptionsPage() {
  const { t } = useTranslation()

  const { data, isLoading } = useQuery({
    queryKey: ['my-subscriptions'],
    queryFn: () => subscriptionsAPI.getMySubscriptions()
  })

  return (
    <>
      <PageHeader
        title={t('userSubscriptions.title')}
        description={t('userSubscriptions.subtitle') as string}
      />

      {isLoading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {Array.from({ length: 2 }).map((_, i) => (
            <Card key={i} className="p-5 space-y-4">
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-3 w-full" />
              <Skeleton className="h-3 w-full" />
              <Skeleton className="h-3 w-3/4" />
            </Card>
          ))}
        </div>
      ) : data && data.length > 0 ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {data.map((sub) => (
            <SubscriptionCard key={sub.id} sub={sub} />
          ))}
        </div>
      ) : (
        <Card className="p-12 text-center text-ink-3">
          {t('userSubscriptions.noSubscriptions')}
        </Card>
      )}
    </>
  )
}
