import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { adminAffiliateAPI } from '@/api/admin/affiliate'

export type AffiliateRecordType = 'invites' | 'rebates' | 'transfers'

interface AffiliateRecordsProps {
  type: AffiliateRecordType
}

type UnknownRecord = Record<string, unknown>
type Filters = {
  search?: string
  start_date?: string
  end_date?: string
}

const PAGE_SIZE = 25

const TITLES: Record<AffiliateRecordType, { title: string; description: string }> = {
  invites: {
    title: 'Affiliate invites',
    description: 'Review invite relationships and invited users.'
  },
  rebates: {
    title: 'Affiliate rebates',
    description: 'Review rebate records generated from affiliate activity.'
  },
  transfers: {
    title: 'Affiliate transfers',
    description: 'Review affiliate quota transfers into user balances.'
  }
}

function asItems(data: unknown): UnknownRecord[] {
  if (Array.isArray(data)) return data.filter((item): item is UnknownRecord => !!item && typeof item === 'object')
  if (!data || typeof data !== 'object') return []
  const record = data as UnknownRecord
  for (const key of ['items', 'records', 'data', 'invites', 'rebates', 'transfers']) {
    const value = record[key]
    if (Array.isArray(value)) return value.filter((item): item is UnknownRecord => !!item && typeof item === 'object')
  }
  return []
}

function pageCount(data: unknown, items: UnknownRecord[]) {
  if (!data || typeof data !== 'object' || Array.isArray(data)) return 1
  const record = data as UnknownRecord
  const direct = Number(record.pages ?? record.total_pages ?? record.page_count)
  if (Number.isFinite(direct) && direct > 0) return direct
  const total = Number(record.total ?? record.count)
  if (Number.isFinite(total) && total > 0) return Math.max(1, Math.ceil(total / PAGE_SIZE))
  return items.length === PAGE_SIZE ? 2 : 1
}

function totalCount(data: unknown, items: UnknownRecord[]) {
  if (!data || typeof data !== 'object' || Array.isArray(data)) return items.length
  const total = Number((data as UnknownRecord).total ?? (data as UnknownRecord).count)
  return Number.isFinite(total) ? total : items.length
}

function text(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function pick(record: UnknownRecord, keys: string[]) {
  for (const key of keys) {
    if (record[key] !== null && record[key] !== undefined && record[key] !== '') return record[key]
  }
  return undefined
}

function money(value: unknown) {
  const number = Number(value || 0)
  return `$${(Number.isFinite(number) ? number : 0).toFixed(4)}`
}

function formatDate(value: unknown) {
  if (!value) return '-'
  const date = new Date(String(value))
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}

function statusTone(status: string) {
  const normalized = status.toLowerCase()
  if (['completed', 'success', 'paid', 'settled'].includes(normalized)) return 'success' as const
  if (['pending', 'processing', 'frozen'].includes(normalized)) return 'warning' as const
  if (['failed', 'cancelled', 'canceled', 'rejected'].includes(normalized)) return 'danger' as const
  return 'neutral' as const
}

function loadRecords(type: AffiliateRecordType, params: Record<string, unknown>) {
  if (type === 'invites') return adminAffiliateAPI.listAffiliateInvites(params)
  if (type === 'rebates') return adminAffiliateAPI.listAffiliateRebates(params)
  return adminAffiliateAPI.listAffiliateTransfers(params)
}

export default function AffiliateRecords({ type }: AffiliateRecordsProps) {
  const [page, setPage] = useState(1)
  const [searchInput, setSearchInput] = useState('')
  const [startDateInput, setStartDateInput] = useState('')
  const [endDateInput, setEndDateInput] = useState('')
  const [filters, setFilters] = useState<Filters>({})

  const title = TITLES[type]
  const query = useQuery({
    queryKey: ['admin-affiliate-records', type, page, filters],
    queryFn: () =>
      loadRecords(type, {
        ...filters,
        page,
        page_size: PAGE_SIZE
      })
  })

  const rows = useMemo(() => asItems(query.data), [query.data])
  const pages = pageCount(query.data, rows)
  const total = totalCount(query.data, rows)

  function applyFilters() {
    setPage(1)
    setFilters({
      search: searchInput.trim() || undefined,
      start_date: startDateInput || undefined,
      end_date: endDateInput || undefined
    })
  }

  function resetFilters() {
    setSearchInput('')
    setStartDateInput('')
    setEndDateInput('')
    setFilters({})
    setPage(1)
  }

  return (
    <>
      <PageHeader
        title={title.title}
        description={title.description}
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      <Card className="p-4 mb-4">
        <div className="grid grid-cols-2 gap-3 md:grid-cols-[1fr_180px_180px_auto]">
          <Input
            name="search"
            label="Search"
            placeholder="User, email, order, invite code"
            value={searchInput}
            onChange={(event) => setSearchInput(event.target.value)}
          />
          <Input
            name="start_date"
            type="date"
            label="From"
            value={startDateInput}
            onChange={(event) => setStartDateInput(event.target.value)}
          />
          <Input
            name="end_date"
            type="date"
            label="To"
            value={endDateInput}
            onChange={(event) => setEndDateInput(event.target.value)}
          />
          <div className="flex items-end justify-end gap-2">
            <Button variant="ghost" onClick={resetFilters}>
              Reset
            </Button>
            <Button variant="accent" onClick={applyFilters} loading={query.isFetching}>
              Filter
            </Button>
          </div>
        </div>
      </Card>

      <Card className="p-4">
        {query.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, index) => (
              <Skeleton key={index} className="h-10" />
            ))}
          </div>
        ) : (
          <Table>
            <THead>
              {type === 'invites' && (
                <TR>
                  <TH>Inviter</TH>
                  <TH>Invitee</TH>
                  <TH>Code</TH>
                  <TH className="text-right">Total rebate</TH>
                  <TH>Created</TH>
                </TR>
              )}
              {type === 'rebates' && (
                <TR>
                  <TH>Inviter</TH>
                  <TH>Invitee</TH>
                  <TH>Order</TH>
                  <TH className="text-right">Amount</TH>
                  <TH>Status</TH>
                  <TH>Created</TH>
                </TR>
              )}
              {type === 'transfers' && (
                <TR>
                  <TH>User</TH>
                  <TH className="text-right">Amount</TH>
                  <TH>Status</TH>
                  <TH>Created</TH>
                  <TH>Processed</TH>
                </TR>
              )}
            </THead>
            <TBody>
              {rows.map((row, index) => {
                const key = text(pick(row, ['id', 'record_id', 'order_id'])) === '-' ? index : text(pick(row, ['id', 'record_id', 'order_id']))

                if (type === 'invites') {
                  return (
                    <TR key={key}>
                      <TD>{text(pick(row, ['inviter_email', 'inviter_username', 'inviter_id', 'user_id']))}</TD>
                      <TD>{text(pick(row, ['invitee_email', 'invitee_username', 'invitee_id', 'invited_user_id']))}</TD>
                      <TD className="font-mono text-xs">{text(pick(row, ['aff_code', 'invite_code', 'code']))}</TD>
                      <TD className="text-right font-mono text-signal-ok">
                        {money(pick(row, ['total_rebate', 'rebate_amount', 'amount']))}
                      </TD>
                      <TD className="text-xs text-ink-3">{formatDate(pick(row, ['created_at', 'createdAt']))}</TD>
                    </TR>
                  )
                }

                if (type === 'rebates') {
                  const status = text(pick(row, ['status', 'state']))
                  return (
                    <TR key={key}>
                      <TD>{text(pick(row, ['inviter_email', 'inviter_username', 'inviter_id', 'user_id']))}</TD>
                      <TD>{text(pick(row, ['invitee_email', 'invitee_username', 'invitee_id', 'invited_user_id']))}</TD>
                      <TD className="font-mono text-xs">
                        {text(pick(row, ['order_no', 'out_trade_no', 'order_id', 'payment_order_id']))}
                      </TD>
                      <TD className="text-right font-mono text-signal-ok">
                        {money(pick(row, ['rebate_amount', 'amount', 'quota', 'commission']))}
                      </TD>
                      <TD>
                        <Badge tone={statusTone(status)}>{status}</Badge>
                      </TD>
                      <TD className="text-xs text-ink-3">{formatDate(pick(row, ['created_at', 'createdAt']))}</TD>
                    </TR>
                  )
                }

                const status = text(pick(row, ['status', 'state']))
                return (
                  <TR key={key}>
                    <TD>{text(pick(row, ['user_email', 'username', 'user_id']))}</TD>
                    <TD className="text-right font-mono text-signal-ok">
                      {money(pick(row, ['amount', 'quota', 'transfer_amount']))}
                    </TD>
                    <TD>
                      <Badge tone={statusTone(status)}>{status}</Badge>
                    </TD>
                    <TD className="text-xs text-ink-3">{formatDate(pick(row, ['created_at', 'createdAt']))}</TD>
                    <TD className="text-xs text-ink-3">
                      {formatDate(pick(row, ['processed_at', 'completed_at', 'updated_at']))}
                    </TD>
                  </TR>
                )
              })}
              {rows.length === 0 && (
                <TR>
                  <TD colSpan={type === 'rebates' ? 6 : 5} className="py-8 text-center text-ink-3">
                    No data
                  </TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        <div className="mt-4 flex items-center justify-between border-t border-line-1 pt-3 text-sm">
          <div className="text-ink-3">Total: {total.toLocaleString()}</div>
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-ink-3">
              {page} / {pages}
            </span>
            <Button variant="ghost" size="sm" disabled={page <= 1} onClick={() => setPage((current) => current - 1)}>
              Back
            </Button>
            <Button
              variant="ghost"
              size="sm"
              disabled={page >= pages}
              onClick={() => setPage((current) => current + 1)}
            >
              Next
            </Button>
          </div>
        </div>
      </Card>
    </>
  )
}
