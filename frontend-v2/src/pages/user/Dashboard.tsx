import { useMemo, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import {
  Activity,
  ArrowRight,
  BadgeCheck,
  BarChart3,
  Clock3,
  Copy,
  CreditCard,
  KeyRound,
  LayoutDashboard,
  ReceiptText,
  UserCircle,
  WalletCards,
  Zap,
} from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { toast } from '@/components/ui/Toast'
import { usageAPI } from '@/api/usage'
import { subscriptionsAPI } from '@/api/subscriptions'
import { subscriptionProductsAPI } from '@/api/subscriptionProducts'
import { useAuthStore } from '@/stores/auth'
import { cn } from '@/lib/cn'
import type {
  ActiveSubscriptionProduct,
  ModelStat,
  TrendDataPoint,
  UsageLog,
  UserSubscription,
} from '@/types'

interface SubscriptionOverviewData {
  subscriptions: UserSubscription[]
  products: ActiveSubscriptionProduct[]
  subscriptionsFailed: boolean
  productsFailed: boolean
}

interface SnapshotMetric {
  label: string
  value: string
  detail: string
  progress: number
  tone: string
}

interface QuickLinkProps {
  to: string
  icon: ReactNode
  title: string
  description: string
  primary?: boolean
}

const modelToneClasses = ['bg-orange', 'bg-signal-ok', 'bg-signal-info', 'bg-signal-warn']

function formatMoney(value?: number | string | null, currency = 'US$') {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount)) return `${currency}0.00`
  return `${currency}${amount.toFixed(2)}`
}

function formatPreciseMoney(value?: number | string | null) {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount)) return 'US$0.0000'
  return `US$${amount.toFixed(4)}`
}

function formatCount(value?: number | string | null) {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount)) return '0'
  return amount.toLocaleString()
}

function formatCompact(value?: number | string | null) {
  const amount = Number(value ?? 0)
  if (!Number.isFinite(amount)) return '0'
  if (amount >= 1_000_000) return `${(amount / 1_000_000).toFixed(1)}M`
  if (amount >= 1_000) return `${(amount / 1_000).toFixed(1)}K`
  return amount.toLocaleString()
}

function formatDate(value?: string | null) {
  if (!value) return '无'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '无'
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

function formatTime(value?: string | null) {
  if (!value) return '--:--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--:--'
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatDuration(value?: number | string | null) {
  const ms = Number(value ?? 0)
  if (!Number.isFinite(ms) || ms <= 0) return '0ms'
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${Math.round(ms)}ms`
}

function clampPercent(value: number) {
  if (!Number.isFinite(value)) return 0
  return Math.min(100, Math.max(0, value))
}

function getProgress(current?: number | string | null, total?: number | string | null) {
  const currentNumber = Number(current ?? 0)
  const totalNumber = Number(total ?? 0)
  if (!Number.isFinite(currentNumber) || !Number.isFinite(totalNumber) || totalNumber <= 0) {
    return currentNumber > 0 ? 8 : 0
  }
  return clampPercent((currentNumber / totalNumber) * 100)
}

function getLastSevenDayRange() {
  const end = new Date()
  const start = new Date()
  start.setDate(end.getDate() - 6)
  return {
    startDate: start.toISOString().slice(0, 10),
    endDate: end.toISOString().slice(0, 10),
  }
}

function getApiBaseUrl(apiBaseUrl?: string | null) {
  if (apiBaseUrl) return apiBaseUrl
  if (typeof window !== 'undefined') return window.location.origin
  return ''
}

function getProductLimitCount(product: ActiveSubscriptionProduct) {
  return [product.daily_limit_usd, product.weekly_limit_usd, product.monthly_limit_usd].filter(
    (value) => Number(value ?? 0) > 0,
  ).length
}

function getLegacyLimitCount(subscription: UserSubscription) {
  const group = subscription.group
  return [group?.daily_limit_usd, group?.weekly_limit_usd, group?.monthly_limit_usd].filter(
    (value) => Number(value ?? 0) > 0,
  ).length
}

function daysRemaining(value?: string | null) {
  if (!value) return null
  const target = new Date(value)
  if (Number.isNaN(target.getTime())) return null
  const diff = target.getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / 86_400_000))
}

function getTrendLabel(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(5)
  return `${date.getMonth() + 1}/${date.getDate()}`
}

function QuickLink({ to, icon, title, description, primary }: QuickLinkProps) {
  return (
    <Link
      to={to}
      className={cn(
        'group flex min-h-[96px] items-center gap-4 rounded-md border p-4 transition-all',
        primary
          ? 'border-ink-1 bg-ink-1 text-bg-1 hover:bg-black'
          : 'border-line-2 bg-bg-1 text-ink-1 hover:border-line-3 hover:bg-bg-3',
      )}
    >
      <span
        className={cn(
          'flex h-11 w-11 shrink-0 items-center justify-center rounded-md',
          primary ? 'bg-bg-1/10 text-bg-1' : 'bg-orange-soft text-orange',
        )}
      >
        {icon}
      </span>
      <span className="min-w-0">
        <span className="block text-sm font-semibold">{title}</span>
        <span className={cn('mt-1 block text-xs', primary ? 'text-bg-3' : 'text-ink-3')}>
          {description}
        </span>
      </span>
      <ArrowRight
        className={cn(
          'ml-auto h-4 w-4 shrink-0 transition-transform group-hover:translate-x-0.5',
          primary ? 'text-bg-3' : 'text-ink-3',
        )}
      />
    </Link>
  )
}

function SnapshotRow({ metric }: { metric: SnapshotMetric }) {
  return (
    <div className="border-t border-line-2 py-4 first:border-t-0 first:pt-0 last:pb-0">
      <div className="flex items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold text-ink-1">{metric.label}</p>
          <p className="mt-1 text-xs text-ink-3">{metric.detail}</p>
        </div>
        <p className="text-2xl font-semibold tracking-normal text-ink-1">{metric.value}</p>
      </div>
      <div className="mt-3 h-2 overflow-hidden rounded-full bg-bg-3">
        <div
          className={cn('h-full rounded-full transition-all', metric.tone)}
          style={{ width: `${clampPercent(metric.progress)}%` }}
        />
      </div>
    </div>
  )
}

function TinyStat({
  label,
  value,
  detail,
}: {
  label: string
  value: string
  detail: string
}) {
  return (
    <div className="rounded-md border border-line-2 bg-bg-1 p-4">
      <p className="text-xs font-semibold text-ink-3">{label}</p>
      <p className="mt-3 text-2xl font-semibold tracking-normal text-ink-1">{value}</p>
      <p className="mt-1 text-xs text-ink-3">{detail}</p>
    </div>
  )
}

function MiniKpi({
  icon,
  label,
  value,
  detail,
}: {
  icon: ReactNode
  label: string
  value: string
  detail: string
}) {
  return (
    <div className="rounded-lg border border-line-2 bg-bg-1 p-5 shadow-card">
      <span className="flex h-10 w-10 items-center justify-center rounded-md bg-orange-soft text-orange">
        {icon}
      </span>
      <p className="mt-4 text-xs font-semibold text-ink-3">{label}</p>
      <p className="mt-2 text-2xl font-semibold tracking-normal text-ink-1">{value}</p>
      <p className="mt-1 text-xs text-ink-3">{detail}</p>
    </div>
  )
}

function TrendBars({ data }: { data: TrendDataPoint[] }) {
  const normalized = data.slice(-7)
  const max = Math.max(...normalized.map((item) => Number(item.requests ?? 0)), 1)
  const hasData = normalized.some((item) => Number(item.requests ?? 0) > 0)

  if (normalized.length === 0) {
    return (
      <div className="flex h-40 items-center justify-center rounded-md border border-dashed border-line-2 text-sm text-ink-3">
        暂无 7 日趋势
      </div>
    )
  }

  return (
    <div className="flex h-44 items-end gap-2 rounded-md border border-line-2 bg-bg-1 px-4 py-4">
      {normalized.map((item) => {
        const requests = Number(item.requests ?? 0)
        const height = hasData ? Math.max(10, (requests / max) * 128) : 10
        return (
          <div key={item.date} className="flex min-w-0 flex-1 flex-col items-center gap-2">
            <div className="flex h-32 w-full items-end">
              <div
                className={cn(
                  'w-full rounded-t-md transition-all',
                  requests > 0 ? 'bg-orange' : 'bg-bg-3',
                )}
                style={{ height }}
                title={`${getTrendLabel(item.date)}: ${formatCount(requests)} 次`}
              />
            </div>
            <span className="text-[11px] text-ink-3">{getTrendLabel(item.date)}</span>
          </div>
        )
      })}
    </div>
  )
}

function ModelDistribution({
  models,
  recentUsage,
}: {
  models: ModelStat[]
  recentUsage: UsageLog[]
}) {
  const rows = useMemo(() => {
    if (models.length > 0) {
      const max = Math.max(...models.map((item) => Number(item.requests ?? 0)), 1)
      return models.slice(0, 5).map((model, index) => ({
        name: model.model || '默认模型',
        meta: `${formatCount(model.requests)} 次 / ${formatCompact(model.total_tokens)} tokens`,
        progress: (Number(model.requests ?? 0) / max) * 100,
        tone: modelToneClasses[index % modelToneClasses.length],
      }))
    }

    const buckets = new Map<string, number>()
    for (const item of recentUsage) {
      const model = item.model || '默认模型'
      buckets.set(model, (buckets.get(model) ?? 0) + 1)
    }
    const max = Math.max(...Array.from(buckets.values()), 1)
    return Array.from(buckets.entries())
      .slice(0, 5)
      .map(([name, count], index) => ({
        name,
        meta: `${formatCount(count)} 次`,
        progress: (count / max) * 100,
        tone: modelToneClasses[index % modelToneClasses.length],
      }))
  }, [models, recentUsage])

  if (rows.length === 0) {
    return (
      <div className="flex h-44 items-center justify-center rounded-md border border-dashed border-line-2 text-sm text-ink-3">
        暂无模型使用数据
      </div>
    )
  }

  return (
    <div className="space-y-4 rounded-md border border-line-2 bg-bg-1 p-4">
      {rows.map((row) => (
        <div key={row.name} className="space-y-2">
          <div className="flex items-center justify-between gap-3 text-sm">
            <span className="min-w-0 truncate font-medium text-ink-1">{row.name}</span>
            <span className="shrink-0 text-xs text-ink-3">{row.meta}</span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-bg-3">
            <div
              className={cn('h-full rounded-full', row.tone)}
              style={{ width: `${clampPercent(row.progress)}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  )
}

function UserDashboard() {
  const user = useAuthStore((state) => state.user)
  const publicSettings = useAuthStore((state) => state.publicSettings)
  const siteName = publicSettings?.site_name || 'XlabAPI'
  const paymentEnabled = publicSettings?.payment_enabled !== false
  const rechargeRoute = paymentEnabled ? '/purchase' : '/redeem'
  const apiEndpoint = getApiBaseUrl(publicSettings?.api_base_url)

  const dateRange = useMemo(() => getLastSevenDayRange(), [])

  const dashboardQuery = useQuery({
    queryKey: ['user-dashboard'],
    queryFn: usageAPI.getUserDashboard,
    refetchInterval: 60_000,
  })

  const subscriptionsQuery = useQuery<SubscriptionOverviewData>({
    queryKey: ['dashboard-subscription-overview'],
    queryFn: async () => {
      const [subscriptionsResult, productsResult] = await Promise.allSettled([
        subscriptionsAPI.getActiveSubscriptions(),
        subscriptionProductsAPI.getActive(),
      ])

      return {
        subscriptions: subscriptionsResult.status === 'fulfilled' ? subscriptionsResult.value : [],
        products: productsResult.status === 'fulfilled' ? productsResult.value : [],
        subscriptionsFailed: subscriptionsResult.status === 'rejected',
        productsFailed: productsResult.status === 'rejected',
      }
    },
    refetchInterval: 60_000,
  })

  const recentUsageQuery = useQuery({
    queryKey: ['dashboard-recent-usage'],
    queryFn: () => usageAPI.listUsage({ page: 1, page_size: 5 }),
    refetchInterval: 60_000,
  })

  const trendQuery = useQuery({
    queryKey: ['dashboard-trend', dateRange.startDate, dateRange.endDate],
    queryFn: () =>
      usageAPI.getUserTrend({
        start_date: dateRange.startDate,
        end_date: dateRange.endDate,
        granularity: 'day',
      }),
    refetchInterval: 60_000,
  })

  const modelQuery = useQuery({
    queryKey: ['dashboard-models', dateRange.startDate, dateRange.endDate],
    queryFn: () =>
      usageAPI.getUserModelStats({
        start_date: dateRange.startDate,
        end_date: dateRange.endDate,
      }),
    refetchInterval: 60_000,
  })

  const stats = dashboardQuery.data
  const subscriptions = subscriptionsQuery.data?.subscriptions ?? []
  const products = subscriptionsQuery.data?.products ?? []
  const recentUsage = recentUsageQuery.data?.items ?? []
  const trendData = trendQuery.data?.trend ?? []
  const modelStats = modelQuery.data?.models ?? []
  const balance = Number(user?.balance ?? 0)

  const productGroupIds = useMemo(
    () =>
      new Set(
        products
          .flatMap((product) => product.groups.map((group) => group.group_id))
          .filter(Boolean),
      ),
    [products],
  )

  const legacyOnlySubscriptions = useMemo(
    () =>
      subscriptions.filter(
        (subscription) =>
          !subscription.group_id || !productGroupIds.has(subscription.group_id),
      ),
    [productGroupIds, subscriptions],
  )

  const activeSubscriptionCount = products.length + legacyOnlySubscriptions.length
  const primarySubscriptionName =
    products[0]?.name || legacyOnlySubscriptions[0]?.group?.name || '暂无有效订阅'

  const nextExpiry = useMemo(() => {
    const values = [
      ...products.map((product) => product.expires_at),
      ...legacyOnlySubscriptions.map((subscription) => subscription.expires_at),
    ].filter(Boolean) as string[]

    const sorted = values
      .map((value) => new Date(value))
      .filter((date) => !Number.isNaN(date.getTime()))
      .sort((left, right) => left.getTime() - right.getTime())

    return sorted[0]?.toISOString() ?? null
  }, [legacyOnlySubscriptions, products])

  const quotaWindowCount = useMemo(() => {
    const productCount = products.reduce(
      (sum, product) => sum + getProductLimitCount(product),
      0,
    )
    const legacyCount = legacyOnlySubscriptions.reduce(
      (sum, subscription) => sum + getLegacyLimitCount(subscription),
      0,
    )
    return productCount + legacyCount
  }, [legacyOnlySubscriptions, products])

  const subscriptionUsage = useMemo(() => {
    const productUsage = products.reduce((sum, item) => sum + Number(item.monthly_usage_usd ?? 0), 0)
    const productLimit = products.reduce((sum, item) => sum + Number(item.monthly_limit_usd ?? 0), 0)
    const legacyUsage = legacyOnlySubscriptions.reduce(
      (sum, item) => sum + Number(item.monthly_usage_usd ?? 0),
      0,
    )
    const legacyLimit = legacyOnlySubscriptions.reduce(
      (sum, item) => sum + Number(item.group?.monthly_limit_usd ?? 0),
      0,
    )

    return {
      usage: productUsage + legacyUsage,
      limit: productLimit + legacyLimit,
    }
  }, [legacyOnlySubscriptions, products])

  const subscriptionRoute =
    activeSubscriptionCount > 0 ? '/subscriptions' : paymentEnabled ? '/purchase' : '/subscriptions'

  const snapshotMetrics: SnapshotMetric[] = [
    {
      label: '余额',
      detail: '账户可用额度',
      value: formatMoney(balance),
      progress: balance > 0 ? Math.min(100, Math.max(4, balance * 10)) : 0,
      tone: 'bg-signal-ok',
    },
    {
      label: 'API 密钥',
      detail: `${formatCount(stats?.active_api_keys)} 个活跃 / 共 ${formatCount(stats?.total_api_keys)} 个`,
      value: `${formatCount(stats?.active_api_keys)} / ${formatCount(stats?.total_api_keys)}`,
      progress: getProgress(stats?.active_api_keys, stats?.total_api_keys),
      tone: 'bg-orange',
    },
    {
      label: '今日消费',
      detail: `${formatCount(stats?.today_requests)} 次请求`,
      value: formatPreciseMoney(stats?.today_actual_cost ?? stats?.today_cost),
      progress: getProgress(
        stats?.today_actual_cost ?? stats?.today_cost,
        Math.max(Number(stats?.total_actual_cost ?? stats?.total_cost ?? 0), 1),
      ),
      tone: 'bg-ink-3',
    },
  ]

  const syncState = dashboardQuery.isLoading
    ? { tone: 'warning' as const, label: '同步中' }
    : dashboardQuery.isError
      ? { tone: 'danger' as const, label: '数据异常' }
      : { tone: 'success' as const, label: '已同步' }

  const copyText = async (value: string, label: string) => {
    if (!value) return
    try {
      await navigator.clipboard.writeText(value)
      toast.success(`${label}已复制`)
    } catch {
      toast.error(`${label}复制失败`)
    }
  }

  const expiryDays = daysRemaining(nextExpiry)
  const subscriptionUsageProgress = getProgress(subscriptionUsage.usage, subscriptionUsage.limit)

  return (
    <div className="mx-auto max-w-[1500px] space-y-5">
      <section className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div className="flex items-center gap-3">
          <span className="grid h-11 w-11 place-items-center rounded-xl border border-line-2 bg-bg-1 shadow-card">
            <LayoutDashboard className="h-5 w-5 text-orange" />
          </span>
          <div>
            <h1 className="text-3xl font-semibold tracking-normal text-ink-1">概览</h1>
            <p className="mt-1 text-sm text-ink-3">欢迎使用 {siteName}</p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <span className="inline-flex h-10 items-center gap-2 rounded-full border border-line-2 bg-bg-1 px-4 text-sm font-semibold text-ink-1 shadow-card">
            <WalletCards className="h-4 w-4 text-orange" />
            {formatMoney(balance)}
          </span>
          <Badge tone={syncState.tone} dot>
            {syncState.label}
          </Badge>
        </div>
      </section>

      <section className="grid gap-5 xl:grid-cols-[minmax(0,1fr)_minmax(420px,1fr)]">
        <Card className="p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-sm font-semibold text-ink-3">运行快照</p>
              <h2 className="mt-2 text-3xl font-semibold tracking-normal text-ink-1">余额</h2>
            </div>
            <div className="flex flex-wrap items-center justify-end gap-2">
              <Badge tone="accent">快捷操作</Badge>
              <Link to={rechargeRoute} className="btn btn-accent">
                <CreditCard className="h-4 w-4" />
                充值
              </Link>
            </div>
          </div>

          <div className="mt-6">
            {snapshotMetrics.map((metric) => (
              <SnapshotRow key={metric.label} metric={metric} />
            ))}
          </div>

          <div className="mt-5 grid gap-3 sm:grid-cols-3">
            <TinyStat label="今日请求" value={formatCount(stats?.today_requests)} detail={`累计 ${formatCount(stats?.total_requests)} 次`} />
            <TinyStat label="今日 Token" value={formatCompact(stats?.today_tokens)} detail={`累计 ${formatCompact(stats?.total_tokens)}`} />
            <TinyStat label="平均响应" value={formatDuration(stats?.average_duration_ms)} detail="请求平均耗时" />
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-sm font-semibold text-ink-3">我的订阅</p>
              <h2 className="mt-2 text-3xl font-semibold tracking-normal text-ink-1">我的订阅</h2>
            </div>
            <Badge tone={activeSubscriptionCount > 0 ? 'success' : 'neutral'} dot>
              {activeSubscriptionCount > 0 ? '有效订阅' : '暂无有效订阅'}
            </Badge>
          </div>

          <div className="mt-6 rounded-xl border border-line-2 bg-bg-1 p-5">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div className="flex min-w-0 gap-4">
                <span className="grid h-12 w-12 shrink-0 place-items-center rounded-xl bg-orange-soft text-orange">
                  <BadgeCheck className="h-5 w-5" />
                </span>
                <div className="min-w-0">
                  <h3 className="truncate text-2xl font-semibold tracking-normal text-ink-1">
                  {primarySubscriptionName}
                  </h3>
                  <p className="mt-2 text-sm leading-6 text-ink-3">
                    {activeSubscriptionCount > 0
                      ? '当前订阅正在生效，可继续使用对应分组和产品额度。'
                      : '当前没有有效订阅，可通过充值、兑换码或管理员分配继续使用。'}
                  </p>
                </div>
              </div>
              <Link to={subscriptionRoute} className="btn btn-ghost shrink-0">
                <ReceiptText className="h-4 w-4" />
                {activeSubscriptionCount > 0 ? '查看订阅' : '订阅'}
              </Link>
            </div>

            <div className="mt-5 h-2 overflow-hidden rounded-full bg-bg-3">
              <div
                className="h-full rounded-full bg-orange"
                style={{
                  width: `${subscriptionUsageProgress}%`,
                }}
              />
            </div>
            <div className="mt-2 flex items-center justify-between text-xs text-ink-3">
              <span>本月已用 {formatMoney(subscriptionUsage.usage)}</span>
              <span>
                {subscriptionUsage.limit > 0 ? formatMoney(subscriptionUsage.limit) : '未配置上限'}
              </span>
            </div>
          </div>

          <div className="mt-4 grid gap-3 sm:grid-cols-3">
            <TinyStat
              label="生效计划"
              value={formatCount(activeSubscriptionCount)}
              detail={activeSubscriptionCount > 0 ? '正在生效' : '暂无'}
            />
            <TinyStat
              label="到期时间"
              value={formatDate(nextExpiry)}
              detail={expiryDays === null ? '暂无订阅' : `剩余 ${expiryDays} 天`}
            />
            <TinyStat
              label="额度窗口"
              value={formatCount(quotaWindowCount)}
              detail={quotaWindowCount > 0 ? '已配置' : '暂无'}
            />
          </div>

          {(subscriptionsQuery.data?.subscriptionsFailed ||
            subscriptionsQuery.data?.productsFailed) && (
            <div className="mt-4 rounded-md border border-signal-warn/30 bg-signal-warn/10 px-4 py-3 text-sm text-signal-warn">
            订阅数据暂时未完全返回，请稍后刷新。
            </div>
          )}
        </Card>
      </section>

      <section>
        <Card className="p-6">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <p className="text-sm font-semibold text-ink-3">快捷操作</p>
              <h2 className="mt-2 text-3xl font-semibold tracking-normal text-ink-1">开发者入口</h2>
              <p className="mt-2 text-sm text-ink-3">
                复制账号信息、确认 API 端点，并快速进入密钥、用量和资料页面。
              </p>
            </div>
            <Badge tone="accent">API</Badge>
          </div>

          <div className="mt-6 grid overflow-hidden rounded-xl border border-line-2 md:grid-cols-3">
            <div className="border-b border-line-2 p-4 md:border-r md:border-b-0">
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-xs font-semibold text-ink-3">邮箱</p>
                  <p className="mt-2 truncate font-medium text-ink-1">{user?.email || '-'}</p>
                </div>
                <button
                  type="button"
                  className="btn btn-ghost btn-icon shrink-0"
                  onClick={() => copyText(user?.email || '', '邮箱')}
                  aria-label="复制邮箱"
                >
                  <Copy className="h-4 w-4" />
                </button>
              </div>
            </div>
            <div className="border-b border-line-2 p-4 md:border-r md:border-b-0">
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <p className="text-xs font-semibold text-ink-3">API 端点</p>
                  <p className="mt-2 break-all font-medium text-ink-1">{apiEndpoint}</p>
                </div>
                <button
                  type="button"
                  className="btn btn-ghost btn-icon shrink-0"
                  onClick={() => copyText(apiEndpoint, 'API 端点')}
                  aria-label="复制 API 端点"
                >
                  <Copy className="h-4 w-4" />
                </button>
              </div>
            </div>
            <div className="p-4">
              <p className="text-xs font-semibold text-ink-3">注册时间</p>
              <p className="mt-2 font-medium text-ink-1">{formatDate(user?.created_at)}</p>
              <p className="mt-1 text-xs text-ink-3">
                账户状态：{user?.status === 'disabled' ? '已禁用' : '正常'}
              </p>
            </div>
          </div>

          <div className="mt-5 grid gap-3 lg:grid-cols-3">
            <QuickLink
              to="/keys"
              icon={<KeyRound className="h-5 w-5" />}
              title="创建 API 密钥"
              description="生成新的接入凭证"
              primary
            />
            <QuickLink
              to="/usage"
              icon={<BarChart3 className="h-5 w-5" />}
              title="查看使用记录"
              description="请求、消耗和模型"
            />
            <QuickLink
              to="/profile"
              icon={<UserCircle className="h-5 w-5" />}
              title="查看资料"
              description="账号信息和安全设置"
            />
          </div>
        </Card>
      </section>

      <div className="grid gap-4 md:grid-cols-3">
        <MiniKpi
          icon={<Zap className="h-5 w-5" />}
          label="RPM"
          value={formatCount(stats?.rpm)}
          detail="近窗口每分钟请求"
        />
        <MiniKpi
          icon={<Activity className="h-5 w-5" />}
          label="TPM"
          value={formatCompact(stats?.tpm)}
          detail="近窗口每分钟 Token"
        />
        <MiniKpi
          icon={<WalletCards className="h-5 w-5" />}
          label="并发额度"
          value={formatCount(user?.concurrency)}
          detail="当前账户可用并发"
        />
      </div>

      <section className="grid gap-5 xl:grid-cols-[minmax(0,1.1fr)_minmax(360px,0.9fr)]">
        <Card className="p-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <p className="text-sm font-semibold text-ink-3">运行快照</p>
              <h2 className="mt-2 text-2xl font-semibold tracking-normal text-ink-1">
                查看最近请求、消费和模型使用活动。
              </h2>
            </div>
            <Link to="/usage" className="btn btn-ghost">
              查看全部
              <ArrowRight className="h-4 w-4" />
            </Link>
          </div>

          <div className="mt-6 overflow-hidden rounded-md border border-line-2 bg-bg-1">
            <div className="hidden grid-cols-[88px_minmax(0,1fr)_104px_96px] gap-3 border-b border-line-2 px-4 py-3 text-xs font-semibold text-ink-3 md:grid">
              <span>时间</span>
              <span>模型</span>
              <span>消耗</span>
              <span>延迟</span>
            </div>
            <div className="divide-y divide-line-2">
              {recentUsage.length > 0 ? (
                recentUsage.map((item) => (
                  <div
                    key={item.id}
                    className="grid gap-3 px-4 py-4 text-sm md:grid-cols-[88px_minmax(0,1fr)_104px_96px]"
                  >
                    <span className="text-ink-3">{formatTime(item.created_at)}</span>
                    <span className="min-w-0 truncate font-medium text-ink-1">
                      {item.model || '默认模型'}
                    </span>
                    <span className="text-ink-3">
                      {formatPreciseMoney(item.actual_cost ?? item.total_cost)}
                    </span>
                    <span className="text-ink-3">{formatDuration(item.duration_ms)}</span>
                  </div>
                ))
              ) : (
                <div className="flex min-h-48 items-center justify-center text-sm text-ink-3">
                  暂无最近活动
                </div>
              )}
            </div>
          </div>
        </Card>

        <div className="grid gap-5">
          <Card className="p-6">
            <div className="mb-5 flex items-center justify-between gap-4">
              <div>
                <p className="text-sm font-semibold text-ink-3">7 日趋势</p>
                <h3 className="mt-2 text-xl font-semibold tracking-normal text-ink-1">
                  请求量
                </h3>
              </div>
              <Badge tone={trendQuery.isError ? 'danger' : 'neutral'}>
                {trendQuery.isLoading ? '加载中' : '近 7 天'}
              </Badge>
            </div>
            <TrendBars data={trendData} />
          </Card>

          <Card className="p-6">
            <div className="mb-5 flex items-center justify-between gap-4">
              <div>
                <p className="text-sm font-semibold text-ink-3">模型分布</p>
                <h3 className="mt-2 text-xl font-semibold tracking-normal text-ink-1">
                  模型使用
                </h3>
              </div>
              <Badge tone={modelQuery.isError ? 'danger' : 'neutral'}>
                {modelQuery.isLoading ? '加载中' : '实时'}
              </Badge>
            </div>
            <ModelDistribution models={modelStats} recentUsage={recentUsage} />
          </Card>
        </div>
      </section>

      <div className="flex flex-wrap items-center gap-3 rounded-xl border border-line-2 bg-bg-1 px-5 py-4 text-sm text-ink-3 shadow-card">
        <Clock3 className="h-4 w-4" />
        <span>注册时间：{formatDate(user?.created_at)}</span>
        <span className="hidden text-line-4 sm:inline">/</span>
        <span>账户状态：{user?.status === 'disabled' ? '已禁用' : '正常'}</span>
        <span className="hidden text-line-4 sm:inline">/</span>
        <span>并发额度：{formatCount(user?.concurrency)}</span>
      </div>
    </div>
  )
}

export default UserDashboard
