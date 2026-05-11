import { useMemo, useState } from 'react'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { RefreshCw, Play, Search } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { listAdminChannelMonitors, runAdminChannelMonitor } from '@/api/admin/channelMonitor'

type AnyRecord = Record<string, unknown>

function asObject(value: unknown): AnyRecord {
  return value && typeof value === 'object' && !Array.isArray(value) ? (value as AnyRecord) : {}
}

function asArray(value: unknown): AnyRecord[] {
  if (Array.isArray(value)) return value.map((item) => asObject(item))
  const obj = asObject(value)
  if (Array.isArray(obj.items)) return obj.items.map((item) => asObject(item))
  if (Array.isArray(obj.monitors)) return obj.monitors.map((item) => asObject(item))
  if (Array.isArray(obj.rows)) return obj.rows.map((item) => asObject(item))
  return []
}

function asText(value: unknown, fallback = '-') {
  if (value === null || value === undefined) return fallback
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value)
    } catch {
      return fallback
    }
  }
  const text = String(value)
  return text.trim() ? text : fallback
}

function statusTone(status: unknown) {
  const s = String(status || '').toLowerCase()
  if (['ok', 'success', 'healthy', 'operational', 'available', 'up'].includes(s)) return 'success' as const
  if (['warning', 'degraded', 'partial', 'unstable'].includes(s)) return 'warning' as const
  if (['error', 'failed', 'down', 'unavailable', 'disabled'].includes(s)) return 'danger' as const
  return 'neutral' as const
}

function rowTitle(row: AnyRecord, fallback: string) {
  return asText(row.name || row.title || row.monitor_name || row.channel_name || fallback)
}

export default function ChannelMonitorPage() {
  const qc = useQueryClient()
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<AnyRecord | null>(null)

  const query = useQuery({
    queryKey: ['admin-channel-monitors', search],
    queryFn: () =>
      listAdminChannelMonitors({
        search: search || undefined,
        q: search || undefined
      })
  })

  const items = useMemo(() => asArray(query.data), [query.data])

  const runMutation = useMutation({
    mutationFn: (id: string | number) => runAdminChannelMonitor(id),
    onSuccess: async () => {
      toast.success('Monitor run started')
      await qc.invalidateQueries({ queryKey: ['admin-channel-monitors'] })
    },
    onError: (error: { message?: string }) => toast.error(error?.message || 'Failed to run monitor')
  })

  return (
    <>
      <PageHeader
        title="Channel monitors"
        description="Read and trigger existing admin channel monitors."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      <Card className="p-4 mb-4">
        <div className="flex flex-wrap items-end gap-3">
          <div className="min-w-[260px] flex-1">
            <label className="input-label">Filter</label>
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-ink-3" />
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search monitors"
                className="input pl-9"
              />
            </div>
          </div>
          <div className="text-xs text-ink-3">
            {items.length} monitors
          </div>
        </div>
      </Card>

      {query.isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <Card key={i} className="p-5 space-y-3">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-10 w-full" />
            </Card>
          ))}
        </div>
      ) : (
        <Card className="p-0 overflow-hidden">
          <Table>
            <THead>
              <TR>
                <TH>Monitor</TH>
                <TH>Status</TH>
                <TH>Schedule</TH>
                <TH>Last run</TH>
                <TH className="text-right">Actions</TH>
              </TR>
            </THead>
            <TBody>
              {items.map((row, index) => {
                const id = String(row.id ?? row.monitor_id ?? row.key ?? index)
                const status = row.status || row.latest_status || row.availability_status || row.state || 'unknown'
                const schedule = row.cron || row.schedule || row.interval || row.frequency || '-'
                const lastRun = row.last_run_at || row.last_checked_at || row.updated_at || row.created_at || '-'
                return (
                  <TR key={String(id)}>
                    <TD>
                      <button type="button" onClick={() => setSelected(row)} className="text-left">
                        <div className="font-medium text-ink-1">{rowTitle(row, `Monitor #${id}`)}</div>
                        <div className="mt-1 text-xs text-ink-3 break-all">
                          {asText(row.description || row.note || row.channel || row.provider || row.platform)}
                        </div>
                      </button>
                    </TD>
                    <TD>
                      <Badge tone={statusTone(status)} dot>{asText(status)}</Badge>
                    </TD>
                    <TD className="text-sm text-ink-2">{asText(schedule)}</TD>
                    <TD className="text-xs text-ink-3">{asText(lastRun)}</TD>
                    <TD className="text-right">
                      <div className="inline-flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => runMutation.mutate(id)}
                          loading={runMutation.isPending && runMutation.variables === id}
                        >
                          <Play className="h-3.5 w-3.5" />
                          Run
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => setSelected(row)}>
                          Detail
                        </Button>
                      </div>
                    </TD>
                  </TR>
                )
              })}
              {items.length === 0 && (
                <TR>
                  <TD colSpan={5} className="py-8 text-center text-ink-3">
                    No monitors found.
                  </TD>
                </TR>
              )}
            </TBody>
          </Table>
        </Card>
      )}

      <Modal
        open={!!selected}
        onClose={() => setSelected(null)}
        title={selected ? rowTitle(selected, 'Monitor detail') : 'Monitor detail'}
        size="lg"
      >
        <Table>
          <THead>
            <TR>
              <TH>Field</TH>
              <TH>Value</TH>
            </TR>
          </THead>
          <TBody>
            {selected &&
              Object.entries(selected).map(([key, value]) => (
                <TR key={key}>
                  <TD className="font-mono text-xs text-ink-3">{key}</TD>
                  <TD className="font-mono text-xs text-ink-2 break-all">{asText(value)}</TD>
                </TR>
              ))}
          </TBody>
        </Table>
      </Modal>
    </>
  )
}
