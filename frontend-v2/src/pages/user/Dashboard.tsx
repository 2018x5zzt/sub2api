import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { KeyRound, BarChart3, Gift, Activity, Clock, Coins } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { usageAPI } from '@/api/usage'
import { useAuthStore } from '@/stores/auth'

function StatCard({ label, value, sub, icon }: { label: string; value: string; sub?: string; icon?: React.ReactNode }) {
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between">
        <div className="text-xs uppercase tracking-wider text-ink-3">{label}</div>
        {icon && <div className="text-orange">{icon}</div>}
      </div>
      <div className="mt-2 font-display text-3xl text-ink-1">{value}</div>
      {sub && <div className="mt-1 text-xs text-ink-3">{sub}</div>}
    </Card>
  )
}

export default function UserDashboard() {
  const { t } = useTranslation()
  const user = useAuthStore((s) => s.user)
  const { data, isLoading } = useQuery({
    queryKey: ['user-dashboard'],
    queryFn: () => usageAPI.getUserDashboard(),
    refetchInterval: 60_000
  })

  return (
    <>
      <PageHeader
        title={t('dashboard.title')}
        description={t('dashboard.welcomeMessage') as string}
      />

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="p-5">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-8 w-32 mt-3" />
            </Card>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            label={t('dashboard.balance')}
            value={`$${user?.balance?.toFixed(4) ?? '0.0000'}`}
            icon={<Coins className="h-4 w-4" />}
          />
          <StatCard
            label={t('dashboard.apiKeys')}
            value={`${data?.active_api_keys ?? 0} / ${data?.total_api_keys ?? 0}`}
            sub={`${t('common.active')} / ${t('common.total')}`}
            icon={<KeyRound className="h-4 w-4" />}
          />
          <StatCard
            label={t('dashboard.todayRequests')}
            value={(data?.today_requests ?? 0).toLocaleString()}
            sub={`RPM ${(data?.rpm ?? 0).toFixed(1)} · TPM ${(data?.tpm ?? 0).toFixed(0)}`}
            icon={<Activity className="h-4 w-4" />}
          />
          <StatCard
            label={t('dashboard.todayCost')}
            value={`$${(data?.today_actual_cost ?? 0).toFixed(4)}`}
            sub={`${t('dashboard.standard')}: $${(data?.today_cost ?? 0).toFixed(4)}`}
            icon={<BarChart3 className="h-4 w-4" />}
          />
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mt-4">
        <Card className="p-5 lg:col-span-2">
          <div className="text-xs uppercase tracking-wider text-ink-3">
            {t('dashboard.totalTokens')}
          </div>
          <div className="mt-3 grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.input')}</div>
              <div className="font-mono mt-0.5">{(data?.total_input_tokens ?? 0).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.output')}</div>
              <div className="font-mono mt-0.5">{(data?.total_output_tokens ?? 0).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.cache')} ({t('dashboard.input')})</div>
              <div className="font-mono mt-0.5">{(data?.total_cache_creation_tokens ?? 0).toLocaleString()}</div>
            </div>
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.cache')} ({t('dashboard.output')})</div>
              <div className="font-mono mt-0.5">{(data?.total_cache_read_tokens ?? 0).toLocaleString()}</div>
            </div>
          </div>
        </Card>
        <Card className="p-5">
          <div className="text-xs uppercase tracking-wider text-ink-3 flex items-center gap-2">
            <Clock className="h-3.5 w-3.5" /> {t('dashboard.performance')}
          </div>
          <div className="mt-3">
            <div className="text-xs text-ink-3">{t('dashboard.avgResponse')}</div>
            <div className="font-mono mt-0.5 text-lg">
              {data?.average_duration_ms ? `${data.average_duration_ms} ms` : '—'}
            </div>
          </div>
        </Card>
      </div>

      <Card className="p-5 mt-4">
        <div className="text-xs uppercase tracking-wider text-ink-3 mb-3">
          {t('dashboard.quickActions')}
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <Link to="/keys" className="card-hover card p-4 group">
            <div className="flex items-center gap-3">
              <div className="h-9 w-9 rounded-lg bg-orange-soft flex items-center justify-center text-orange">
                <KeyRound className="h-4 w-4" />
              </div>
              <div>
                <div className="font-medium text-ink-1">{t('dashboard.createApiKey')}</div>
                <div className="text-xs text-ink-3 mt-0.5">{t('dashboard.generateNewKey')}</div>
              </div>
            </div>
          </Link>
          <Link to="/usage" className="card-hover card p-4 group">
            <div className="flex items-center gap-3">
              <div className="h-9 w-9 rounded-lg bg-orange-soft flex items-center justify-center text-orange">
                <BarChart3 className="h-4 w-4" />
              </div>
              <div>
                <div className="font-medium text-ink-1">{t('dashboard.viewUsage')}</div>
                <div className="text-xs text-ink-3 mt-0.5">{t('dashboard.checkDetailedLogs')}</div>
              </div>
            </div>
          </Link>
          <Link to="/redeem" className="card-hover card p-4 group">
            <div className="flex items-center gap-3">
              <div className="h-9 w-9 rounded-lg bg-orange-soft flex items-center justify-center text-orange">
                <Gift className="h-4 w-4" />
              </div>
              <div>
                <div className="font-medium text-ink-1">{t('dashboard.redeemCode')}</div>
                <div className="text-xs text-ink-3 mt-0.5">{t('dashboard.addBalanceWithCode')}</div>
              </div>
            </div>
          </Link>
        </div>
      </Card>
    </>
  )
}
