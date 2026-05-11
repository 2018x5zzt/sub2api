import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Activity, RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { channelMonitorAPI } from '@/api/channelMonitor'

function asArray(data: any): any[] {
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.items)) return data.items
  if (Array.isArray(data?.monitors)) return data.monitors
  return []
}

function statusTone(status: string | undefined) {
  const s = String(status || '').toLowerCase()
  if (['ok', 'success', 'healthy', 'operational', 'available'].includes(s)) return 'success' as const
  if (['failed', 'error', 'down', 'unavailable'].includes(s)) return 'danger' as const
  if (['degraded', 'warning', 'partial'].includes(s)) return 'warning' as const
  return 'neutral' as const
}

export default function ChannelStatusPage() {
  const { t } = useTranslation()
  const [selected, setSelected] = useState<any | null>(null)

  const query = useQuery({
    queryKey: ['channel-monitors'],
    queryFn: () => channelMonitorAPI.listChannelMonitors()
  })

  const monitors = asArray(query.data)
  const degraded = useMemo(
    () => monitors.some((m) => statusTone(m.status || m.latest_status || m.availability_status) !== 'success'),
    [monitors]
  )

  return (
    <>
      <PageHeader
        title={t('nav.channelStatus')}
        description="Monitor public channel availability and recent status."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            {t('common.refresh')}
          </Button>
        }
      />

      <Card className="p-5 mb-4 flex items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="h-9 w-9 rounded-lg bg-orange-soft text-orange flex items-center justify-center">
            <Activity className="h-4 w-4" />
          </div>
          <div>
            <div className="text-sm font-medium text-ink-1">Overall status</div>
            <div className="text-xs text-ink-3">{monitors.length} monitors visible</div>
          </div>
        </div>
        <Badge tone={degraded ? 'warning' : 'success'} dot>{degraded ? 'degraded' : 'operational'}</Badge>
      </Card>

      {query.isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i} className="p-5 space-y-3">
              <Skeleton className="h-5 w-40" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-10 w-full" />
            </Card>
          ))}
        </div>
      ) : monitors.length === 0 ? (
        <Card className="p-12 text-center text-ink-3">{t('common.noData')}</Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {monitors.map((m) => {
            const status = m.status || m.latest_status || m.availability_status || 'unknown'
            return (
              <button key={m.id} className="text-left" onClick={() => setSelected(m)}>
                <Card className="p-5 h-full hover:border-orange/40 transition-colors">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <h2 className="font-medium text-ink-1 truncate">{m.name || `Monitor #${m.id}`}</h2>
                      <p className="mt-1 text-xs text-ink-3 truncate">{m.provider || m.platform || '-'}</p>
                    </div>
                    <Badge tone={statusTone(status)} dot>{status}</Badge>
                  </div>
                  <div className="mt-4 grid grid-cols-2 gap-3 text-xs">
                    <div className="rounded-md border border-line-1 bg-bg-2 p-3">
                      <div className="text-ink-3">Latency</div>
                      <div className="mt-1 font-mono text-ink-1">{m.latency_ms ?? m.avg_latency_ms ?? '-'} ms</div>
                    </div>
                    <div className="rounded-md border border-line-1 bg-bg-2 p-3">
                      <div className="text-ink-3">Availability</div>
                      <div className="mt-1 font-mono text-ink-1">{m.availability_7d ?? m.availability ?? '-'}%</div>
                    </div>
                  </div>
                </Card>
              </button>
            )
          })}
        </div>
      )}

      <Modal open={!!selected} onClose={() => setSelected(null)} title={selected?.name || 'Monitor detail'}>
        <Table>
          <THead>
            <TR>
              <TH>Field</TH>
              <TH>Value</TH>
            </TR>
          </THead>
          <TBody>
            {selected &&
              Object.entries(selected).slice(0, 24).map(([k, v]) => (
                <TR key={k}>
                  <TD className="font-mono text-xs text-ink-3">{k}</TD>
                  <TD className="font-mono text-xs text-ink-2 break-all">{typeof v === 'object' ? JSON.stringify(v) : String(v ?? '-')}</TD>
                </TR>
              ))}
          </TBody>
        </Table>
      </Modal>
    </>
  )
}
