import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Badge } from '@/components/ui/Badge'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminUsageAPI } from '@/api/admin/usage'
import type { UsageQueryParams, UsageRequestType } from '@/types'

const REQ_TYPES: Array<UsageRequestType | ''> = ['', 'sync', 'stream', 'ws_v2']

export default function AdminUsagePage() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)

  const [userIdInput, setUserIdInput] = useState('')
  const [apiKeyIdInput, setApiKeyIdInput] = useState('')
  const [model, setModel] = useState('')
  const [reqType, setReqType] = useState<'' | UsageRequestType>('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  const [filters, setFilters] = useState<UsageQueryParams>({})

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['admin-usage', page, filters],
    queryFn: () => adminUsageAPI.listAdminUsage({ ...filters, page, page_size: 25 })
  })

  function applyFilters() {
    setPage(1)
    const next: UsageQueryParams = {}
    const userId = Number(userIdInput)
    const apiKeyId = Number(apiKeyIdInput)
    if (Number.isFinite(userId) && userId > 0) next.user_id = userId
    if (Number.isFinite(apiKeyId) && apiKeyId > 0) next.api_key_id = apiKeyId
    if (model.trim()) next.model = model.trim()
    if (reqType) next.request_type = reqType
    if (startDate) next.start_date = startDate
    if (endDate) next.end_date = endDate
    setFilters(next)
  }

  function reset() {
    setUserIdInput('')
    setApiKeyIdInput('')
    setModel('')
    setReqType('')
    setStartDate('')
    setEndDate('')
    setPage(1)
    setFilters({})
  }

  const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

  return (
    <>
      <PageHeader title="Usage" description="All requests across users — searchable log." />

      <Card className="p-4 mb-4">
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mb-3">
          <Input
            name="user_id"
            label="User ID"
            placeholder="—"
            value={userIdInput}
            onChange={(e) => setUserIdInput(e.target.value)}
          />
          <Input
            name="api_key_id"
            label="API Key ID"
            placeholder="—"
            value={apiKeyIdInput}
            onChange={(e) => setApiKeyIdInput(e.target.value)}
          />
          <Input
            name="model"
            label="Model"
            placeholder="claude-sonnet-…"
            value={model}
            onChange={(e) => setModel(e.target.value)}
          />
          <div>
            <label className="input-label">Type</label>
            <select
              value={reqType}
              onChange={(e) => setReqType(e.target.value as '' | UsageRequestType)}
              className={selectClass}
            >
              {REQ_TYPES.map((r) => (
                <option key={r || 'all'} value={r} className="bg-bg-4">
                  {r || 'all'}
                </option>
              ))}
            </select>
          </div>
          <Input
            name="start_date"
            type="date"
            label="From"
            value={startDate}
            onChange={(e) => setStartDate(e.target.value)}
          />
          <Input
            name="end_date"
            type="date"
            label="To"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
          />
        </div>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="sm" onClick={reset}>
            {t('common.reset')}
          </Button>
          <Button variant="accent" size="sm" onClick={applyFilters} loading={isFetching}>
            {t('common.filter')}
          </Button>
        </div>
      </Card>

      <Card className="p-4">
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-10" />
            ))}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Time</TH>
                <TH>User</TH>
                <TH>Key</TH>
                <TH>{t('dashboard.model')}</TH>
                <TH className="text-right">{t('dashboard.tokens')}</TH>
                <TH className="text-right">{t('dashboard.actual')}</TH>
                <TH>Type</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((log) => {
                const totalTokens =
                  log.input_tokens + log.output_tokens + log.cache_creation_tokens + log.cache_read_tokens
                return (
                  <TR key={log.id}>
                    <TD className="text-ink-3 text-xs font-mono whitespace-nowrap">
                      {new Date(log.created_at).toLocaleString()}
                    </TD>
                    <TD className="text-ink-2 text-xs">
                      {log.user?.email || (log.user_id ? `#${log.user_id}` : '—')}
                    </TD>
                    <TD className="text-ink-3 text-xs font-mono">
                      {log.api_key?.name || (log.api_key_id ? `#${log.api_key_id}` : '—')}
                    </TD>
                    <TD className="font-mono text-xs">{log.model}</TD>
                    <TD className="text-right font-mono">{totalTokens.toLocaleString()}</TD>
                    <TD className="text-right font-mono">${log.actual_cost.toFixed(6)}</TD>
                    <TD>
                      {log.stream ? (
                        <Badge tone="accent">stream</Badge>
                      ) : (
                        <Badge>{log.request_type || 'sync'}</Badge>
                      )}
                    </TD>
                  </TR>
                )
              })}
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
              {t('common.total')}: {data.total.toLocaleString()}
            </div>
            <div className="flex gap-2 items-center">
              <span className="text-ink-3 text-xs font-mono">
                {page} / {data.pages}
              </span>
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
