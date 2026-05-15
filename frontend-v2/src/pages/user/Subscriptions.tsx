import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { BadgeCheck, CreditCard, KeyRound, Plus, ShieldCheck } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { subscriptionsAPI } from '@/api/subscriptions'
import { subscriptionProductsAPI } from '@/api/subscriptionProducts'
import { modelsAPI } from '@/api/models'
import { userAPI } from '@/api/user'
import { useAuthStore } from '@/stores/auth'
import { toast } from '@/components/ui/Toast'
import { cn } from '@/lib/cn'
import type { ActiveSubscriptionProduct, Group, User, UserSubscription } from '@/types'

interface SubscriptionPageData {
  subscriptions: UserSubscription[]
  products: ActiveSubscriptionProduct[]
  profile: User | null
  groups: Group[]
  subscriptionsFailed: boolean
  productsFailed: boolean
  profileFailed: boolean
  groupsFailed: boolean
}

function isAuthFailure(error: unknown) {
  const maybe = error as { status?: number; response?: { status?: number } } | null
  return maybe?.status === 401 || maybe?.response?.status === 401
}

function money(value: unknown, precision = 2) {
  const n = Number(value || 0)
  return `$${(Number.isFinite(n) ? n : 0).toFixed(precision)}`
}

function hasLimit(limit: number | null | undefined) {
  return typeof limit === 'number' && limit > 0
}

function statusTone(status: string) {
  if (status === 'active') return 'success' as const
  if (status === 'expired') return 'warning' as const
  if (status === 'revoked') return 'danger' as const
  return 'neutral' as const
}

function statusLabel(status: string, t: ReturnType<typeof useTranslation>['t']) {
  const key = `userSubscriptions.status.${status}`
  const label = t(key)
  return label === key ? status : label
}

function platformLabel(platform: string | undefined, t: ReturnType<typeof useTranslation>['t']) {
  if (!platform) return t('common.unknown') as string
  const key = `admin.groups.platforms.${platform}`
  const label = t(key)
  return label === key ? platform : label
}

function platformDotClass(platform: string | undefined) {
  switch (platform) {
    case 'anthropic':
      return 'bg-orange'
    case 'openai':
      return 'bg-signal-ok'
    case 'gemini':
      return 'bg-sky-400'
    case 'antigravity':
      return 'bg-violet-500'
    case 'sora':
      return 'bg-fuchsia-500'
    default:
      return 'bg-ink-4'
  }
}

function progressClass(used: number | undefined, limit: number | null | undefined) {
  if (!hasLimit(limit)) return 'bg-ink-4'
  const pct = (Number(used || 0) / Number(limit)) * 100
  if (pct >= 90) return 'bg-signal-err'
  if (pct >= 70) return 'bg-signal-warn'
  return 'bg-signal-ok'
}

function progressWidth(used: number | undefined, limit: number | null | undefined) {
  if (!hasLimit(limit)) return '0%'
  return `${Math.min((Number(used || 0) / Number(limit)) * 100, 100)}%`
}

function formatDateOnly(value: string | null | undefined) {
  if (!value) return ''
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleDateString()
}

function expirationText(value: string | null | undefined, t: ReturnType<typeof useTranslation>['t']) {
  if (!value) return t('userSubscriptions.noExpiration') as string
  const expires = new Date(value)
  if (Number.isNaN(expires.getTime())) return String(value)

  const now = new Date()
  const days = Math.ceil((expires.getTime() - now.getTime()) / 86_400_000)
  if (days < 0) return t('userSubscriptions.status.expired') as string
  const date = formatDateOnly(value)
  if (days === 0) return `${date} (${t('common.today')})`
  if (days === 1) return `${date} (${t('common.tomorrow')})`
  return `${t('userSubscriptions.daysRemaining', { days })} (${date})`
}

function expirationClass(value: string | null | undefined) {
  if (!value) return 'text-ink-3'
  const days = Math.ceil((new Date(value).getTime() - Date.now()) / 86_400_000)
  if (days <= 0) return 'text-signal-err font-medium'
  if (days <= 3) return 'text-signal-err'
  if (days <= 7) return 'text-signal-warn'
  return 'text-ink-2'
}

function resetText(windowStart: string | null | undefined, windowHours: number, t: ReturnType<typeof useTranslation>['t']) {
  if (!windowStart) return t('userSubscriptions.windowNotActive') as string
  const start = new Date(windowStart)
  const diff = start.getTime() + windowHours * 3_600_000 - Date.now()
  if (Number.isNaN(start.getTime()) || diff <= 0) return t('userSubscriptions.windowNotActive') as string

  const hours = Math.floor(diff / 3_600_000)
  const minutes = Math.floor((diff % 3_600_000) / 60_000)
  if (hours > 24) {
    const days = Math.floor(hours / 24)
    return `${days}d ${hours % 24}h`
  }
  return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`
}

function UsageLine({
  label,
  used,
  limit,
  hint
}: {
  label: string
  used: number | undefined
  limit: number | null | undefined
  hint?: string
}) {
  const limited = hasLimit(limit)
  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between gap-3">
        <span className="text-xs font-medium text-ink-3">{label}</span>
        <span className="font-mono text-xs font-semibold text-ink-2">
          {money(used)}
          {limited && <span className="text-ink-3"> / {money(limit)}</span>}
        </span>
      </div>
      {limited && (
        <div className="h-2 overflow-hidden rounded-full bg-bg-3">
          <div className={cn('h-full rounded-full transition-all', progressClass(used, limit))} style={{ width: progressWidth(used, limit) }} />
        </div>
      )}
      {hint && <p className="text-[11px] text-ink-3">{hint}</p>}
    </div>
  )
}

function ProductCard({ product }: { product: ActiveSubscriptionProduct }) {
  const { t } = useTranslation()
  const primaryPlatform = product.groups[0]?.group_platform
  const dailyBaseLimit = Number(product.daily_limit_usd || 0)
  const dailyCarryover = Number(product.daily_carryover_in_usd || 0)
  const dailyLimit = dailyBaseLimit + dailyCarryover
  const hasDailyQuota = dailyLimit > 0
  const hasWeeklyQuota = hasLimit(product.weekly_limit_usd)
  const hasMonthlyQuota = hasLimit(product.monthly_limit_usd)
  const hasQuota = hasDailyQuota || hasWeeklyQuota || hasMonthlyQuota

  return (
    <Card className="overflow-hidden">
      <div className="border-b border-line-1 p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="flex min-w-0 items-start gap-3">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-orange-soft text-orange">
              <CreditCard className="h-4 w-4" />
            </div>
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-2">
                <h2 className="truncate text-base font-semibold text-ink-1">{product.name}</h2>
                <Badge tone={statusTone(product.status)}>{statusLabel(product.status, t)}</Badge>
              </div>
              {product.description && <p className="mt-1 text-xs leading-5 text-ink-3">{product.description}</p>}
            </div>
          </div>
          <span className={cn('mt-1 h-2 w-2 shrink-0 rounded-full', platformDotClass(primaryPlatform))} />
        </div>
      </div>

      <div className="space-y-5 p-5">
        <div className="flex items-center justify-between gap-3">
          <span className="font-mono text-[11px] uppercase tracking-[0.14em] text-ink-3">{t('userSubscriptions.expires')}</span>
          <span className={cn('text-sm', expirationClass(product.expires_at))}>{expirationText(product.expires_at, t)}</span>
        </div>

        {hasQuota ? (
          <div className="space-y-3">
            {hasDailyQuota && (
              <UsageLine
                label={t('userSubscriptions.daily') as string}
                used={product.daily_usage_usd}
                limit={dailyLimit}
                hint={dailyCarryover > 0 ? `结转 ${money(dailyCarryover)} + 今日 ${money(dailyBaseLimit)} = 可用 ${money(dailyLimit)}` : undefined}
              />
            )}
            {hasWeeklyQuota && <UsageLine label={t('userSubscriptions.weekly') as string} used={product.weekly_usage_usd} limit={product.weekly_limit_usd} />}
            {hasMonthlyQuota && <UsageLine label={t('userSubscriptions.monthly') as string} used={product.monthly_usage_usd} limit={product.monthly_limit_usd} />}
          </div>
        ) : (
          <div className="flex items-center gap-3 rounded-xl bg-bg-2 px-4 py-5">
            <span className="font-display text-3xl text-signal-ok">∞</span>
            <div>
              <p className="text-sm font-semibold text-signal-ok">{t('userSubscriptions.unlimited')}</p>
              <p className="text-xs text-ink-3">{t('userSubscriptions.unlimitedDesc')}</p>
            </div>
          </div>
        )}

        {product.groups.length > 0 && (
          <div className="border-t border-line-1 pt-4">
            <p className="mb-2 font-mono text-[11px] uppercase tracking-[0.14em] text-ink-3">可用分组</p>
            <div className="flex flex-wrap gap-1.5">
              {product.groups.map((group) => (
                <span
                  key={group.group_id}
                  className="inline-flex items-center gap-1.5 rounded-lg bg-bg-2 px-2.5 py-1 text-xs text-ink-2 ring-1 ring-inset ring-line-1"
                >
                  <span className={cn('h-1.5 w-1.5 rounded-full', platformDotClass(group.group_platform))} />
                  <span className="max-w-48 truncate">{group.group_name}</span>
                  <span className="font-mono text-ink-3">{Number(group.debit_multiplier || 1).toFixed(2).replace(/\.?0+$/, '')}x</span>
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </Card>
  )
}

function LegacySubscriptionCard({ subscription }: { subscription: UserSubscription }) {
  const { t } = useTranslation()
  const group = subscription.group
  const noLimits = !hasLimit(group?.daily_limit_usd) && !hasLimit(group?.weekly_limit_usd) && !hasLimit(group?.monthly_limit_usd)

  return (
    <Card className="overflow-hidden">
      <div className="border-b border-line-1 p-5">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <span className={cn('h-2 w-2 rounded-full', platformDotClass(group?.platform))} />
              <h2 className="truncate text-base font-semibold text-ink-1">{group?.name || `Group #${subscription.group_id}`}</h2>
              <Badge>{platformLabel(group?.platform, t)}</Badge>
              <Badge tone={statusTone(subscription.status)}>{statusLabel(subscription.status, t)}</Badge>
            </div>
            {group?.description && <p className="mt-1 text-xs text-ink-3">{group.description}</p>}
          </div>
        </div>
      </div>

      <div className="space-y-5 p-5">
        <div className="flex items-center justify-between gap-3">
          <span className="font-mono text-[11px] uppercase tracking-[0.14em] text-ink-3">{t('userSubscriptions.expires')}</span>
          <span className={cn('text-sm', expirationClass(subscription.expires_at))}>{expirationText(subscription.expires_at, t)}</span>
        </div>

        {noLimits ? (
          <div className="flex items-center gap-3 rounded-xl bg-bg-2 px-4 py-5">
            <span className="font-display text-3xl text-signal-ok">∞</span>
            <div>
              <p className="text-sm font-semibold text-signal-ok">{t('userSubscriptions.unlimited')}</p>
              <p className="text-xs text-ink-3">{t('userSubscriptions.unlimitedDesc')}</p>
            </div>
          </div>
        ) : (
          <div className="space-y-3">
            {hasLimit(group?.daily_limit_usd) && (
              <UsageLine
                label={t('userSubscriptions.daily') as string}
                used={subscription.daily_usage_usd}
                limit={group?.daily_limit_usd}
                hint={t('userSubscriptions.resetIn', { time: resetText(subscription.daily_window_start, 24, t) }) as string}
              />
            )}
            {hasLimit(group?.weekly_limit_usd) && (
              <UsageLine
                label={t('userSubscriptions.weekly') as string}
                used={subscription.weekly_usage_usd}
                limit={group?.weekly_limit_usd}
                hint={t('userSubscriptions.resetIn', { time: resetText(subscription.weekly_window_start, 168, t) }) as string}
              />
            )}
            {hasLimit(group?.monthly_limit_usd) && (
              <UsageLine
                label={t('userSubscriptions.monthly') as string}
                used={subscription.monthly_usage_usd}
                limit={group?.monthly_limit_usd}
                hint={t('userSubscriptions.resetIn', { time: resetText(subscription.monthly_window_start, 720, t) }) as string}
              />
            )}
          </div>
        )}
      </div>
    </Card>
  )
}

export default function SubscriptionsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const authUser = useAuthStore((s) => s.user)
  const refreshUser = useAuthStore((s) => s.refreshUser)

  const [fallbackEnabled, setFallbackEnabled] = useState(false)
  const [fallbackLimit, setFallbackLimit] = useState(0)
  const [fallbackGroupId, setFallbackGroupId] = useState<number | null>(null)

  const query = useQuery<SubscriptionPageData>({
    queryKey: ['user-subscriptions-page'],
    queryFn: async () => {
      const [subscriptionsResult, productsResult, profileResult, groupsResult] = await Promise.allSettled([
        subscriptionsAPI.getActiveSubscriptions(),
        subscriptionProductsAPI.getActive(),
        refreshUser(),
        modelsAPI.getUserGroups()
      ])

      const authFailure = [subscriptionsResult, productsResult, profileResult, groupsResult].find(
        (result) => result.status === 'rejected' && isAuthFailure(result.reason)
      )
      if (authFailure?.status === 'rejected') throw authFailure.reason

      return {
        subscriptions: subscriptionsResult.status === 'fulfilled' ? subscriptionsResult.value : [],
        products: productsResult.status === 'fulfilled' ? productsResult.value : [],
        profile: profileResult.status === 'fulfilled' ? profileResult.value : null,
        groups: groupsResult.status === 'fulfilled' ? groupsResult.value : [],
        subscriptionsFailed: subscriptionsResult.status === 'rejected',
        productsFailed: productsResult.status === 'rejected',
        profileFailed: profileResult.status === 'rejected',
        groupsFailed: groupsResult.status === 'rejected'
      }
    }
  })

  const profile: User | null = query.data?.profile ?? authUser
  const subscriptions = query.data?.subscriptions ?? []
  const products = query.data?.products ?? []
  const groups = query.data?.groups ?? []
  const subscriptionDataFailed = Boolean(query.data?.subscriptionsFailed || query.data?.productsFailed)
  const allSubscriptionDataFailed = Boolean(query.data?.subscriptionsFailed && query.data?.productsFailed)
  const auxiliaryDataFailed = Boolean(query.data?.profileFailed || query.data?.groupsFailed)

  useEffect(() => {
    if (!profile) return
    setFallbackEnabled(Boolean(profile.subscription_balance_fallback_enabled))
    setFallbackLimit(profile.subscription_balance_fallback_limit_usd || 0)
    setFallbackGroupId(profile.subscription_balance_fallback_group_id || null)
  }, [
    profile?.id,
    profile?.subscription_balance_fallback_enabled,
    profile?.subscription_balance_fallback_group_id,
    profile?.subscription_balance_fallback_limit_usd
  ])

  const fallbackUsed = profile?.subscription_balance_fallback_used_usd || 0
  const fallbackRemaining = Math.max(fallbackLimit - fallbackUsed, 0)

  const fallbackGroupOptions = useMemo(
    () =>
      groups
        .filter((group: Group) => group.status === 'active' && group.subscription_type !== 'subscription')
        .map((group: Group) => ({ value: group.id, label: group.name || `Group #${group.id}` })),
    [groups]
  )

  const productGroupIds = useMemo(() => {
    const ids = new Set<number>()
    for (const product of products) {
      for (const group of product.groups || []) ids.add(group.group_id)
    }
    return ids
  }, [products])

  const visibleSubscriptions = useMemo(
    () => subscriptions.filter((subscription) => !productGroupIds.has(subscription.group_id)),
    [productGroupIds, subscriptions]
  )
  const hasActiveSubscriptions = products.length > 0 || visibleSubscriptions.length > 0

  const saveMutation = useMutation({
    mutationFn: () => {
      const groupSelectable = fallbackGroupOptions.some((option) => option.value === fallbackGroupId)
      if (fallbackEnabled && query.data?.groupsFailed) {
        throw new Error('余额兜底分组尚未加载成功，请刷新后再保存')
      }
      if (fallbackEnabled && !fallbackGroupId) {
        throw new Error('请选择余额兜底分组')
      }
      if (fallbackEnabled && !groupSelectable) {
        throw new Error('请选择可用的余额兜底分组')
      }
      if (fallbackEnabled && fallbackLimit <= 0) {
        throw new Error('请设置大于 0 的余额兜底上限')
      }
      return userAPI.updateProfile({
        subscription_balance_fallback_enabled: fallbackEnabled,
        subscription_balance_fallback_limit_usd: Math.max(fallbackLimit || 0, 0),
        subscription_balance_fallback_group_id: fallbackEnabled ? fallbackGroupId : null
      })
    },
    onSuccess: async (updated) => {
      useAuthStore.setState({ user: updated })
      await refreshUser()
      toast.success(t('common.saved') as string)
      query.refetch()
    },
    onError: (error: { message?: string }) => {
      toast.error(error.message || (t('common.error') as string))
    }
  })

  function resetFallbackForm() {
    setFallbackEnabled(Boolean(profile?.subscription_balance_fallback_enabled))
    setFallbackLimit(profile?.subscription_balance_fallback_limit_usd || 0)
    setFallbackGroupId(profile?.subscription_balance_fallback_group_id || null)
  }

  return (
    <>
      <PageHeader
        title={t('userSubscriptions.title')}
        description={t('userSubscriptions.description') as string}
      />

      {query.isLoading ? (
        <div className="mx-auto grid max-w-5xl gap-4 lg:grid-cols-2">
          {Array.from({ length: 4 }).map((_, index) => (
            <Card key={index} className="space-y-4 p-5">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-14 w-full" />
              <Skeleton className="h-14 w-full" />
            </Card>
          ))}
        </div>
      ) : (
        <div className="mx-auto max-w-5xl space-y-5">
          <Card className="overflow-hidden">
            <div className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-start gap-3">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-emerald-100 text-signal-ok">
                  <ShieldCheck className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <h2 className="text-sm font-semibold text-ink-1">订阅额度耗尽后自动使用余额</h2>
                  <p className="mt-0.5 text-xs leading-relaxed text-ink-3">
                    开启后，产品订阅额度耗尽且已选择余额分组时，会在你设置的上限内自动改用余额。
                  </p>
                </div>
              </div>

              <label className="inline-flex shrink-0 cursor-pointer items-center gap-2.5">
                <span className="relative inline-flex h-6 w-11 items-center">
                  <input
                    type="checkbox"
                    className="peer sr-only"
                    checked={fallbackEnabled}
                    disabled={saveMutation.isPending}
                    onChange={(event) => setFallbackEnabled(event.target.checked)}
                  />
                  <span className="absolute inset-0 rounded-full bg-bg-3 transition-colors peer-checked:bg-signal-ok peer-disabled:opacity-50" />
                  <span className="absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow-sm transition-transform peer-checked:translate-x-5" />
                </span>
                <span className={cn('text-xs font-medium', fallbackEnabled ? 'text-signal-ok' : 'text-ink-3')}>
                  {fallbackEnabled ? t('common.enabled') : t('common.disabled')}
                </span>
              </label>
            </div>

            <div className="border-t border-line-1 bg-bg-2/70 p-5">
              {fallbackEnabled && (
                <div className="mb-4 grid gap-4 sm:grid-cols-[minmax(0,1fr)_180px_auto] sm:items-end">
                  <div>
                    <label className="input-label" htmlFor="fallback_group">余额分组</label>
                    <select
                      id="fallback_group"
                      className="input"
                      value={fallbackGroupId ?? ''}
                      disabled={saveMutation.isPending}
                      onChange={(event) => setFallbackGroupId(event.target.value ? Number(event.target.value) : null)}
                    >
                      <option value="">请选择余额分组</option>
                      {fallbackGroupOptions.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </div>
                  <Input
                    name="fallback_limit"
                    type="number"
                    min="0"
                    step="0.01"
                    label="余额兜底上限"
                    value={fallbackLimit}
                    disabled={saveMutation.isPending}
                    onChange={(event) => setFallbackLimit(Number(event.target.value || 0))}
                    leftIcon={<span className="font-mono text-xs">$</span>}
                  />
                  <div className="pb-2 font-mono text-xs text-ink-3">
                    已用 {money(fallbackUsed)} / 剩余 {money(fallbackRemaining)}
                  </div>
                </div>
              )}

              <p className="mb-4 text-xs leading-relaxed text-signal-warn">
                如果余额被扣为负数，后续请求会被阻止，直到重新充值。
              </p>
              <div className="flex flex-wrap gap-2">
                <Button type="button" size="sm" loading={saveMutation.isPending} onClick={() => saveMutation.mutate()}>
                  {t('common.save')}
                </Button>
                <Button type="button" size="sm" variant="ghost" disabled={saveMutation.isPending} onClick={resetFallbackForm}>
                  {t('common.cancel')}
                </Button>
              </div>
            </div>
          </Card>

          {hasActiveSubscriptions && (
            <div className="flex flex-col items-start justify-between gap-3 rounded-xl border border-line-2 bg-bg-1 p-4 sm:flex-row sm:items-center">
              <div className="flex items-start gap-3">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-orange-soft text-orange">
                  <KeyRound className="h-4 w-4" />
                </div>
                <div>
                  <h2 className="text-sm font-semibold text-ink-1">激活后建议新建分组专用 API Key</h2>
                  <p className="mt-0.5 text-xs text-ink-3">这样可以避免继续误用旧 key 的余额或限额，并按分组隔离你的订阅用量。</p>
                </div>
              </div>
              <Button type="button" size="sm" variant="accent" onClick={() => navigate('/keys')}>
                <Plus className="h-3.5 w-3.5" />
                去生成 API Key
              </Button>
            </div>
          )}

          {query.isError && (
            <Card className="flex flex-col gap-3 p-5 text-sm text-signal-err sm:flex-row sm:items-center sm:justify-between">
              <span>{t('userSubscriptions.failedToLoad')}</span>
              <Button type="button" size="sm" variant="ghost" onClick={() => query.refetch()}>
                {t('common.retry')}
              </Button>
            </Card>
          )}

          {!query.isError && subscriptionDataFailed && (
            <Card className="flex flex-col gap-3 p-5 text-sm text-signal-warn sm:flex-row sm:items-center sm:justify-between">
              <span>
                {allSubscriptionDataFailed
                  ? '订阅数据加载失败，暂时无法判断当前用户是否有有效订阅。'
                  : '部分订阅数据加载失败，当前只显示已成功返回的订阅数据。'}
              </span>
              <Button type="button" size="sm" variant="ghost" onClick={() => query.refetch()}>
                {t('common.retry')}
              </Button>
            </Card>
          )}

          {!query.isError && auxiliaryDataFailed && !subscriptionDataFailed && (
            <Card className="p-4 text-sm text-signal-warn">
              订阅已加载，部分账户或分组设置同步失败；余额兜底设置可能需要刷新后再修改。
            </Card>
          )}

          {!query.isError && !allSubscriptionDataFailed && !hasActiveSubscriptions ? (
            <Card className="p-12 text-center">
              <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-xl bg-bg-2 text-ink-3">
                <BadgeCheck className="h-6 w-6" />
              </div>
              <div className="text-base font-semibold text-ink-1">{t('userSubscriptions.noActiveSubscriptions')}</div>
              <p className="mt-1 text-sm text-ink-3">{t('userSubscriptions.noActiveSubscriptionsDesc')}</p>
            </Card>
          ) : null}

          {!query.isError && !allSubscriptionDataFailed && hasActiveSubscriptions ? (
            <div className="grid gap-5 lg:grid-cols-2">
              {products.map((product) => (
                <ProductCard key={`product-${product.subscription_id}`} product={product} />
              ))}
              {visibleSubscriptions.map((subscription) => (
                <LegacySubscriptionCard key={subscription.id} subscription={subscription} />
              ))}
            </div>
          ) : null}
        </div>
      )}
    </>
  )
}
