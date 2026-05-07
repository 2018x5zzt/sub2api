import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Gift, Coins, Zap, BadgeCheck, AlertCircle, CheckCircle2 } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { useAuthStore } from '@/stores/auth'
import { redeemAPI, type RedeemHistoryItem, type RedeemResult } from '@/api/redeem'
import { toast } from '@/components/ui/Toast'

const SUBSCRIPTION_TYPES = new Set(['subscription'])
const BALANCE_TYPES = new Set(['balance', 'admin_balance', 'invitation'])

function isBalanceType(type: string) {
  return BALANCE_TYPES.has(type) || type.endsWith('_balance')
}

function isSubscriptionType(type: string) {
  return SUBSCRIPTION_TYPES.has(type)
}

function HistoryRow({ item }: { item: RedeemHistoryItem }) {
  const { t } = useTranslation()
  const isBalance = isBalanceType(item.type)
  const isSub = isSubscriptionType(item.type)
  const positive = item.value >= 0

  const Icon = isBalance ? Coins : isSub ? BadgeCheck : Zap
  const tone = isSub
    ? 'text-signal-info bg-signal-info/10'
    : isBalance
    ? positive
      ? 'text-signal-ok bg-signal-ok/10'
      : 'text-signal-err bg-signal-err/10'
    : positive
    ? 'text-signal-info bg-signal-info/10'
    : 'text-signal-warn bg-signal-warn/10'

  const title = isSub
    ? `${t('redeem.subscriptionAssigned')}${item.group?.name ? ` · ${item.group.name}` : ''}`
    : `${item.code}`

  return (
    <div className="flex items-center justify-between rounded-xl bg-bg-2 p-4 border border-line-1">
      <div className="flex items-center gap-3 min-w-0">
        <div className={`flex h-10 w-10 items-center justify-center rounded-xl shrink-0 ${tone}`}>
          <Icon className="h-5 w-5" />
        </div>
        <div className="min-w-0">
          <p className="text-sm font-medium text-ink-1 truncate font-mono">{title}</p>
          <p className="text-xs text-ink-3 mt-0.5">{new Date(item.used_at || item.created_at).toLocaleString()}</p>
        </div>
      </div>
      <div className="text-right shrink-0 ml-3">
        {isSub ? (
          item.validity_days ? (
            <p className="text-sm text-ink-1 font-mono">{t('redeem.subscriptionDays', { days: item.validity_days })}</p>
          ) : null
        ) : isBalance ? (
          <p className={`text-sm font-mono ${positive ? 'text-signal-ok' : 'text-signal-err'}`}>
            {positive ? '+' : ''}${item.value.toFixed(2)}
          </p>
        ) : (
          <p className={`text-sm font-mono ${positive ? 'text-signal-info' : 'text-signal-warn'}`}>
            {positive ? '+' : ''}{item.value} {t('redeem.requests')}
          </p>
        )}
      </div>
    </div>
  )
}

export default function RedeemPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const user = useAuthStore((s) => s.user)
  const refreshUser = useAuthStore((s) => s.refreshUser)

  const [code, setCode] = useState('')
  const [result, setResult] = useState<RedeemResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const { data: history, isLoading: loadingHistory } = useQuery({
    queryKey: ['redeem-history'],
    queryFn: () => redeemAPI.getHistory()
  })

  const redeemMut = useMutation({
    mutationFn: (c: string) => redeemAPI.redeem(c),
    onSuccess: (r) => {
      setResult(r)
      setError(null)
      setCode('')
      qc.invalidateQueries({ queryKey: ['redeem-history'] })
      refreshUser()
      toast.success(t('redeem.redeemSuccess') as string)
    },
    onError: (e: { message?: string }) => {
      setResult(null)
      setError(e?.message || (t('redeem.redeemFailed') as string))
    }
  })

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!code.trim()) return
    redeemMut.mutate(code.trim())
  }

  return (
    <>
      <PageHeader title={t('redeem.title')} description={t('redeem.description') as string} />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-4">
        <Card className="lg:col-span-1 p-6 relative overflow-hidden">
          <div
            className="absolute inset-0 opacity-30 pointer-events-none"
            style={{
              background:
                'radial-gradient(ellipse 50% 60% at 80% 0%, rgba(255,87,34,0.35), transparent 60%)'
            }}
          />
          <div className="relative">
            <div className="text-eyebrow tracking-wider uppercase text-ink-3 font-mono">
              {t('redeem.currentBalance')}
            </div>
            <div className="mt-3 font-display text-4xl text-ink-1">
              ${user?.balance?.toFixed(2) ?? '0.00'}
            </div>
            <div className="mt-2 text-xs text-ink-3 font-mono">
              {t('redeem.concurrency')}: {user?.concurrency ?? 0} {t('redeem.requests')}
            </div>
          </div>
        </Card>

        <Card className="lg:col-span-2 p-6">
          <form onSubmit={onSubmit} className="space-y-4">
            <Input
              name="code"
              label={t('redeem.redeemCodeLabel') as string}
              placeholder={t('redeem.redeemCodePlaceholder') as string}
              hint={t('redeem.redeemCodeHint') as string}
              leftIcon={<Gift className="h-4 w-4" />}
              value={code}
              onChange={(e) => setCode(e.target.value)}
              disabled={redeemMut.isPending}
              autoFocus
              required
            />
            <Button
              type="submit"
              loading={redeemMut.isPending}
              disabled={!code.trim()}
              size="lg"
              variant="accent"
              className="w-full"
            >
              {redeemMut.isPending ? t('redeem.redeeming') : t('redeem.redeemButton')}
            </Button>
          </form>
        </Card>
      </div>

      {result && (
        <Card className="p-5 mb-4 border-signal-ok/30 bg-signal-ok/5">
          <div className="flex items-start gap-3">
            <CheckCircle2 className="h-5 w-5 text-signal-ok shrink-0 mt-0.5" />
            <div className="text-sm">
              <p className="font-medium text-ink-1">{t('redeem.redeemSuccess')}</p>
              <p className="text-ink-2 mt-1">{result.message}</p>
              <div className="mt-2 space-y-1 text-ink-2">
                {result.type === 'balance' && (
                  <p>
                    {t('redeem.added')}: <span className="font-mono">${result.value.toFixed(2)}</span>
                  </p>
                )}
                {result.type === 'balance' && result.fixed_value !== undefined && (
                  <p className="text-xs text-ink-3">
                    {t('redeem.fixedAmount')}: ${result.fixed_value.toFixed(2)} · {t('redeem.luckyAmount')}: $
                    {(result.random_value ?? 0).toFixed(2)}
                  </p>
                )}
                {result.type === 'concurrency' && (
                  <p>
                    {t('redeem.added')}: <span className="font-mono">{result.value}</span>{' '}
                    {t('redeem.concurrentRequests')}
                  </p>
                )}
                {result.type === 'subscription' && (
                  <p>
                    {t('redeem.subscriptionAssigned')}
                    {result.group_name ? ` — ${result.group_name}` : ''}
                    {result.validity_days
                      ? ` (${t('redeem.subscriptionDays', { days: result.validity_days })})`
                      : ''}
                  </p>
                )}
                {result.new_balance !== undefined && (
                  <p>
                    {t('redeem.newBalance')}:{' '}
                    <span className="font-medium font-mono">${result.new_balance.toFixed(2)}</span>
                  </p>
                )}
                {result.new_concurrency !== undefined && (
                  <p>
                    {t('redeem.newConcurrency')}:{' '}
                    <span className="font-medium font-mono">{result.new_concurrency}</span>
                  </p>
                )}
              </div>
            </div>
          </div>
        </Card>
      )}

      {error && (
        <Card className="p-5 mb-4 border-signal-err/30 bg-signal-err/5">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-signal-err shrink-0 mt-0.5" />
            <div className="text-sm">
              <p className="font-medium text-ink-1">{t('redeem.redeemFailed')}</p>
              <p className="text-ink-2 mt-1">{error}</p>
            </div>
          </div>
        </Card>
      )}

      <Card className="p-0">
        <div className="px-6 py-4 border-b border-line-1">
          <h2 className="text-base font-medium text-ink-1">{t('redeem.recentActivity')}</h2>
        </div>
        <div className="p-4 space-y-3">
          {loadingHistory ? (
            Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-16" />)
          ) : history && history.length > 0 ? (
            history.map((item) => <HistoryRow key={item.id} item={item} />)
          ) : (
            <div className="text-center text-ink-3 py-12 text-sm">{t('common.noData')}</div>
          )}
        </div>
      </Card>
    </>
  )
}
