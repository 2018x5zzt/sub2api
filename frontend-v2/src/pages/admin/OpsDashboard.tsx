import { useMemo, useState, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Activity, RefreshCw, Server, ShieldAlert, TimerReset } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { getOpsSnapshot, listOpsSystemLogs } from '@/api/admin/ops'

type AnyRecord = Record<string, unknown>

function asObject(value: unknown): AnyRecord {
  return value && typeof value === 'object' && !Array.isArray(value) ? (value as AnyRecord) : {}
}

function asArray(value: unknown): unknown[] {
  if (Array.isArray(value)) return value
  if (Array.isArray(asObject(value).items)) return asObject(value).items as unknown[]
  if (Array.isArray(asObject(value).rows)) return asObject(value).rows as unknown[]
  if (Array.isArray(asObject(value).logs)) return asObject(value).logs as unknown[]
  return []
}

function asText(value: unknown, fallback = '-') {
  if (value === null || value === undefined) return fallback
  if (typeof value === 'string' && value.trim() === '') return fallback
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value)
    } catch {
      return fallback
    }
  }
  return String(value)
}

function formatNumber(value: unknown) {
  const n = Number(value)
  return Number.isFinite(n) ? n.toLocaleString() : asText(value)
}

function rawEntries(value: unknown) {
  return Object.entries(asObject(value))
}

function StatCard({
  label,
  value,
  sub,
  icon
}: {
  label: string
  value: string
  sub?: string
  icon: ReactNode
}) {
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="text-[11px] uppercase tracking-[0.14em] text-ink-3">{label}</div>
        <div className="text-orange">{icon}</div>
      </div>
      <div className="mt-2 text-2xl font-display text-ink-1">{value}</div>
      {sub && <div className="mt-1 text-xs text-ink-3">{sub}</div>}
    </Card>
  )
}

export default function OpsDashboard() {
  const [systemLogFilter, setSystemLogFilter] = useState('')
  const [selectedLog, setSelectedLog] = useState<AnyRecord | null>(null)

  const snapshotQuery = useQuery({
    queryKey: ['admin-ops-snapshot'],
    queryFn: () => getOpsSnapshot(),
    refetchInterval: 60_000
  })

  const logsQuery = useQuery({
    queryKey: ['admin-ops-system-logs', systemLogFilter],
    queryFn: () =>
      listOpsSystemLogs({
        search: systemLogFilter || undefined,
        q: systemLogFilter || undefined
      }),
    refetchInterval: 30_000
  })

  const snapshot = asObject(snapshotQuery.data)
  const logs = useMemo(() => asArray(logsQuery.data).map((item) => asObject(item)), [logsQuery.data])

  const summary = {
    activeAccounts: snapshot.active_accounts ?? snapshot.activeAccounts,
    totalChannels: snapshot.total_channels ?? snapshot.totalChannels,
    healthyChannels: snapshot.healthy_channels ?? snapshot.healthyChannels,
    failedRequests: snapshot.failed_requests ?? snapshot.failedRequests,
    warnings: snapshot.warnings ?? snapshot.warning_count ?? snapshot.warningCount
  }

  return (
    <>
      <PageHeader
        title="Ops dashboard"
        description="Snapshot summary and system logs."
        actions={
          <Button
            variant="ghost"
            onClick={() => {
              snapshotQuery.refetch()
              logsQuery.refetch()
            }}
            disabled={snapshotQuery.isFetching || logsQuery.isFetching}
          >
            <RefreshCw className={`h-4 w-4 ${snapshotQuery.isFetching || logsQuery.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-5">
        <StatCard
          label="Active accounts"
          value={formatNumber(summary.activeAccounts)}
          sub="Operational account pool"
          icon={<Server className="h-4 w-4" />}
        />
        <StatCard
          label="Channels"
          value={formatNumber(summary.totalChannels)}
          sub={`${formatNumber(summary.healthyChannels)} healthy`}
          icon={<Activity className="h-4 w-4" />}
        />
        <StatCard
          label="Warnings"
          value={formatNumber(summary.warnings)}
          sub="Observed in snapshot"
          icon={<ShieldAlert className="h-4 w-4" />}
        />
        <StatCard
          label="Failed requests"
          value={formatNumber(summary.failedRequests)}
          sub="Recent request error count"
          icon={<TimerReset className="h-4 w-4" />}
        />
        <StatCard
          label="Snapshot"
          value={snapshotQuery.isLoading ? 'Loading' : snapshotQuery.isError ? 'Error' : 'Ready'}
          sub="Refreshed on demand"
          icon={<RefreshCw className="h-4 w-4" />}
        />
      </div>

      <div className="grid grid-cols-1 gap-4 mt-4 xl:grid-cols-3">
        <Card className="p-5 xl:col-span-2">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 className="text-sm font-medium text-ink-1">Snapshot details</h2>
              <p className="mt-1 text-xs text-ink-3">Raw fields from the current ops snapshot.</p>
            </div>
            <Badge tone="accent">summary</Badge>
          </div>

          {snapshotQuery.isLoading ? (
            <div className="mt-4 space-y-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-10" />
              ))}
            </div>
          ) : (
            <Table className="mt-4">
              <THead>
                <TR>
                  <TH>Field</TH>
                  <TH>Value</TH>
                </TR>
              </THead>
              <TBody>
                {rawEntries(snapshot).slice(0, 24).map(([key, value]) => (
                  <TR key={key}>
                    <TD className="font-mono text-xs text-ink-3">{key}</TD>
                    <TD className="font-mono text-xs text-ink-2 break-all">{asText(value)}</TD>
                  </TR>
                ))}
                {rawEntries(snapshot).length === 0 && (
                  <TR>
                    <TD colSpan={2} className="py-8 text-center text-ink-3">
                      No snapshot data.
                    </TD>
                  </TR>
                )}
              </TBody>
            </Table>
          )}
        </Card>

        <Card className="p-5">
          <div className="flex items-center justify-between gap-3">
            <div>
              <h2 className="text-sm font-medium text-ink-1">System logs</h2>
              <p className="mt-1 text-xs text-ink-3">Recent admin-facing system log rows.</p>
            </div>
            <Badge>{logs.length}</Badge>
          </div>

          <div className="mt-4">
            <label className="input-label">Filter</label>
            <input
              value={systemLogFilter}
              onChange={(e) => setSystemLogFilter(e.target.value)}
              placeholder="Search logs"
              className="input"
            />
          </div>

          <div className="mt-4 space-y-2">
            {logsQuery.isLoading ? (
              Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-12" />)
            ) : logs.length === 0 ? (
              <div className="rounded-lg border border-line-1 bg-bg-2 px-4 py-8 text-center text-sm text-ink-3">
                No system logs.
              </div>
            ) : (
              logs.slice(0, 8).map((log, idx) => (
                <button
                  key={`${asText(log.id, String(idx))}-${idx}`}
                  type="button"
                  onClick={() => setSelectedLog(log)}
                  className="block w-full text-left"
                >
                  <Card className="p-4 hover:border-orange/40 transition-colors">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="text-sm font-medium text-ink-1 truncate">
                          {asText(log.message || log.msg || log.title || `Log #${idx + 1}`)}
                        </div>
                        <div className="mt-1 text-xs text-ink-3 truncate">
                          {asText(log.level || log.severity || log.type || 'info')}
                        </div>
                      </div>
                      <Badge tone={String(log.level || log.severity || '').toLowerCase().includes('error') ? 'danger' : 'neutral'}>
                        {asText(log.level || log.severity || 'info')}
                      </Badge>
                    </div>
                    <div className="mt-3 text-xs text-ink-3 line-clamp-2 break-all">
                      {asText(log.detail || log.error || log.payload || log.context)}
                    </div>
                  </Card>
                </button>
              ))
            )}
          </div>
        </Card>
      </div>

      <Modal
        open={!!selectedLog}
        onClose={() => setSelectedLog(null)}
        title={selectedLog ? asText(selectedLog.message || selectedLog.msg || selectedLog.title || 'System log detail') : 'System log detail'}
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
            {selectedLog &&
              Object.entries(selectedLog).map(([key, value]) => (
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
