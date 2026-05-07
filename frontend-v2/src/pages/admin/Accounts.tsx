import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Power, AlertCircle, RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminAccountsAPI } from '@/api/admin/accounts'
import { toast } from '@/components/ui/Toast'
import type { Account, AccountPlatform } from '@/types'

const PLATFORMS: AccountPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'sora']

function statusBadge(a: Account) {
  if (!a.schedulable) return <Badge tone="warning" dot>paused</Badge>
  if (a.status === 'error') return <Badge tone="danger" dot>error</Badge>
  if (a.status === 'inactive') return <Badge dot>inactive</Badge>
  if (a.rate_limit_reset_at) return <Badge tone="warning" dot>rate-limited</Badge>
  return <Badge tone="success" dot>active</Badge>
}

function relativeTime(iso: string | null): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  return `${d}d ago`
}

export default function AdminAccountsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [platform, setPlatform] = useState<AccountPlatform | ''>('')
  const [status, setStatus] = useState<string>('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin-accounts', page, search, platform, status],
    queryFn: () =>
      adminAccountsAPI.listAccounts(page, 20, {
        search: search || undefined,
        platform: (platform || undefined) as AccountPlatform | undefined,
        status: status || undefined
      })
  })

  const toggleSched = useMutation({
    mutationFn: ({ id, schedulable }: { id: number; schedulable: boolean }) =>
      adminAccountsAPI.setAccountSchedulable(id, schedulable),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const clearError = useMutation({
    mutationFn: (id: number) => adminAccountsAPI.clearAccountError(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const clearRate = useMutation({
    mutationFn: (id: number) => adminAccountsAPI.clearAccountRateLimit(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const selectClass = 'input appearance-none cursor-pointer bg-bg-4 max-w-[160px]'

  return (
    <>
      <PageHeader
        title={t('nav.accounts')}
        description="Upstream provider accounts. Full CRUD/credential management deferred to Phase 3 — this view is read-only with quick actions."
      />

      <Card className="p-4">
        <div className="flex flex-wrap gap-3 items-center mb-4">
          <Input
            name="search"
            placeholder={t('common.searchPlaceholder') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
          <select
            value={platform}
            onChange={(e) => setPlatform(e.target.value as AccountPlatform | '')}
            className={selectClass}
          >
            <option value="" className="bg-bg-4">
              All platforms
            </option>
            {PLATFORMS.map((p) => (
              <option key={p} value={p} className="capitalize bg-bg-4">
                {p}
              </option>
            ))}
          </select>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={selectClass}
          >
            <option value="" className="bg-bg-4">
              All status
            </option>
            <option value="active" className="bg-bg-4">active</option>
            <option value="inactive" className="bg-bg-4">inactive</option>
            <option value="error" className="bg-bg-4">error</option>
          </select>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-10" />
            ))}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>{t('common.name')}</TH>
                <TH>Platform / Type</TH>
                <TH>{t('common.status')}</TH>
                <TH className="text-right">Concurrency</TH>
                <TH className="text-right">Priority</TH>
                <TH>Last used</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((a) => (
                <TR key={a.id}>
                  <TD>
                    <div className="text-ink-1 font-medium">{a.name}</div>
                    {a.error_message && (
                      <div
                        className="text-xs text-signal-err mt-0.5 truncate max-w-md"
                        title={a.error_message}
                      >
                        <AlertCircle className="inline h-3 w-3 mr-1" />
                        {a.error_message}
                      </div>
                    )}
                  </TD>
                  <TD>
                    <div className="flex items-center gap-1.5">
                      <Badge tone="accent">{a.platform}</Badge>
                      <span className="text-xs text-ink-3 font-mono">{a.type}</span>
                    </div>
                  </TD>
                  <TD>{statusBadge(a)}</TD>
                  <TD className="text-right font-mono text-sm">
                    {a.current_concurrency ?? 0}
                    <span className="text-ink-3"> / {a.concurrency}</span>
                  </TD>
                  <TD className="text-right font-mono text-sm">{a.priority}</TD>
                  <TD className="text-ink-3 text-xs font-mono">{relativeTime(a.last_used_at)}</TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <button
                        title={a.schedulable ? 'Pause' : 'Resume'}
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => toggleSched.mutate({ id: a.id, schedulable: !a.schedulable })}
                      >
                        <Power className="h-3.5 w-3.5" />
                      </button>
                      {a.error_message && (
                        <button
                          title="Clear error"
                          className="btn btn-ghost btn-icon btn-sm text-signal-warn"
                          onClick={() => clearError.mutate(a.id)}
                        >
                          <AlertCircle className="h-3.5 w-3.5" />
                        </button>
                      )}
                      {a.rate_limit_reset_at && (
                        <button
                          title="Clear rate limit"
                          className="btn btn-ghost btn-icon btn-sm"
                          onClick={() => clearRate.mutate(a.id)}
                        >
                          <RefreshCw className="h-3.5 w-3.5" />
                        </button>
                      )}
                    </div>
                  </TD>
                </TR>
              ))}
              {(data?.items ?? []).length === 0 && (
                <TR>
                  <TD colSpan={7} className="text-center text-ink-3 py-8">
                    {t('common.noData')}
                  </TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        {data && data.pages > 1 && (
          <div className="flex items-center justify-between mt-4 pt-3 border-t border-line-1 text-sm">
            <div className="text-ink-3">
              {t('common.total')}: {data.total}
            </div>
            <div className="flex gap-2">
              <Button variant="ghost" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                {t('common.back')}
              </Button>
              <Button variant="ghost" size="sm" disabled={page >= data.pages} onClick={() => setPage((p) => p + 1)}>
                {t('common.next')}
              </Button>
            </div>
          </div>
        )}
      </Card>
    </>
  )
}
