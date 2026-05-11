import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Users, KeyRound, Server, Activity } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { adminAPI } from '@/api/admin'

function Stat({ label, value, sub, icon }: { label: string; value: string; sub?: string; icon: React.ReactNode }) {
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between">
        <div className="text-xs uppercase tracking-wider text-ink-3">{label}</div>
        <div className="text-orange">{icon}</div>
      </div>
      <div className="mt-2 font-display text-3xl text-ink-1">{value}</div>
      {sub && <div className="mt-1 text-xs text-ink-3">{sub}</div>}
    </Card>
  )
}

export default function AdminDashboard() {
  const { t } = useTranslation()
  const { data, isLoading } = useQuery({
    queryKey: ['admin-dashboard'],
    queryFn: () => adminAPI.getDashboard(),
    refetchInterval: 60_000
  })

  if (isLoading || !data) {
    return (
      <>
        <PageHeader title={t('dashboard.title')} description={t('admin.dashboard.description') as string} />
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="p-5">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-8 w-24 mt-3" />
            </Card>
          ))}
        </div>
      </>
    )
  }

  return (
    <>
      <PageHeader title={t('dashboard.title')} description={t('admin.dashboard.description') as string} />
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Stat
          label={t('nav.users')}
          value={data.total_users.toLocaleString()}
          sub={t('admin.dashboard.usersSub', {
            today: data.today_new_users,
            active: data.active_users
          }) as string}
          icon={<Users className="h-4 w-4" />}
        />
        <Stat
          label={t('nav.apiKeys')}
          value={`${data.active_api_keys} / ${data.total_api_keys}`}
          sub={t('common.active') + ' / ' + t('common.total')}
          icon={<KeyRound className="h-4 w-4" />}
        />
        <Stat
          label={t('nav.accounts')}
          value={data.total_accounts.toLocaleString()}
          sub={t('admin.dashboard.accountsSub', {
            normal: data.normal_accounts,
            error: data.error_accounts,
            rate: data.ratelimit_accounts
          }) as string}
          icon={<Server className="h-4 w-4" />}
        />
        <Stat
          label={t('dashboard.todayRequests')}
          value={data.today_requests.toLocaleString()}
          sub={`RPM ${data.rpm.toFixed(1)} · TPM ${data.tpm.toFixed(0)}`}
          icon={<Activity className="h-4 w-4" />}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 mt-4">
        <Card className="p-5 lg:col-span-2">
          <div className="text-xs uppercase tracking-wider text-ink-3">
            {t('dashboard.totalTokens')}
          </div>
          <div className="mt-3 grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.input')}</div>
              <div className="font-mono mt-0.5">{data.total_input_tokens.toLocaleString()}</div>
            </div>
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.output')}</div>
              <div className="font-mono mt-0.5">{data.total_output_tokens.toLocaleString()}</div>
            </div>
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.cache')} ({t('dashboard.input')})</div>
              <div className="font-mono mt-0.5">{data.total_cache_creation_tokens.toLocaleString()}</div>
            </div>
            <div>
              <div className="text-xs text-ink-3">{t('dashboard.cache')} ({t('dashboard.output')})</div>
              <div className="font-mono mt-0.5">{data.total_cache_read_tokens.toLocaleString()}</div>
            </div>
          </div>
        </Card>
        <Card className="p-5">
          <div className="text-xs uppercase tracking-wider text-ink-3">
            {t('dashboard.todayCost')}
          </div>
          <div className="mt-2 font-display text-3xl text-ink-1">
            ${data.today_actual_cost.toFixed(4)}
          </div>
          <div className="mt-1 text-xs text-ink-3">
            {t('dashboard.standard')}: ${data.today_cost.toFixed(4)}
          </div>
          <div className="mt-3 text-xs text-ink-3">
            {t('common.total')}: ${data.total_actual_cost.toFixed(4)}
          </div>
        </Card>
      </div>
    </>
  )
}
