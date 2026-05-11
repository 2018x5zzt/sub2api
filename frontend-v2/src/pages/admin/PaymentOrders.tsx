import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Eye, RefreshCw, RotateCcw, XCircle } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { adminPaymentAPI } from '@/api/admin/payment'

const STATUSES = ['', 'PENDING', 'COMPLETED', 'FAILED', 'REFUNDED', 'CANCELLED', 'EXPIRED']
const PAGE_SIZE = 25

type Row = Record<string, any>

function asItems(data: any): Row[] {
  if (Array.isArray(data)) return data
  for (const key of ['items', 'orders', 'data', 'records']) {
    if (Array.isArray(data?.[key])) return data[key]
  }
  return []
}

function pageCount(data: any, items: Row[]) {
  const pages = Number(data?.pages ?? data?.total_pages ?? data?.page_count)
  if (Number.isFinite(pages) && pages > 0) return pages
  const total = Number(data?.total ?? data?.count)
  if (Number.isFinite(total) && total > 0) return Math.max(1, Math.ceil(total / PAGE_SIZE))
  return items.length === PAGE_SIZE ? 2 : 1
}

function totalCount(data: any, items: Row[]) {
  const total = Number(data?.total ?? data?.count)
  return Number.isFinite(total) ? total : items.length
}

function statusTone(status: string) {
  const s = status.toUpperCase()
  if (['COMPLETED', 'PAID', 'SUCCESS', 'RECHARGING'].includes(s)) return 'success' as const
  if (['PENDING', 'CREATED', 'WAITING', 'PROCESSING'].includes(s)) return 'warning' as const
  if (['FAILED', 'CANCELLED', 'CANCELED', 'EXPIRED', 'REFUNDED'].includes(s)) return 'danger' as const
  return 'neutral' as const
}

function money(value: unknown) {
  const n = Number(value || 0)
  return `$${(Number.isFinite(n) ? n : 0).toFixed(2)}`
}

function dateText(value: unknown) {
  if (!value) return '-'
  const d = new Date(String(value))
  return Number.isNaN(d.getTime()) ? String(value) : d.toLocaleString()
}

function text(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function orderId(row: Row) {
  return row.id ?? row.order_id ?? row.out_trade_no ?? row.trade_no
}

export default function AdminPaymentOrdersPage() {
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [searchInput, setSearchInput] = useState('')
  const [statusInput, setStatusInput] = useState('')
  const [fromInput, setFromInput] = useState('')
  const [toInput, setToInput] = useState('')
  const [filters, setFilters] = useState<Record<string, string>>({})
  const [selected, setSelected] = useState<Row | null>(null)

  const query = useQuery({
    queryKey: ['admin-payment-orders', page, filters],
    queryFn: () =>
      adminPaymentAPI.getPaymentOrders({
        page,
        page_size: PAGE_SIZE,
        keyword: filters.search || undefined,
        search: filters.search || undefined,
        status: filters.status || undefined,
        start_date: filters.from || undefined,
        end_date: filters.to || undefined
      })
  })

  const rows = useMemo(() => asItems(query.data), [query.data])
  const pages = pageCount(query.data, rows)
  const total = totalCount(query.data, rows)

  const cancelMut = useMutation({
    mutationFn: (id: number | string) => adminPaymentAPI.cancelPaymentOrder(id),
    onSuccess: () => {
      toast.success('Order cancelled')
      qc.invalidateQueries({ queryKey: ['admin-payment-orders'] })
    },
    onError: (e: { message?: string }) => toast.error(e?.message || 'Cancel failed')
  })

  const retryMut = useMutation({
    mutationFn: (id: number | string) => adminPaymentAPI.retryPaymentOrder(id),
    onSuccess: () => {
      toast.success('Retry submitted')
      qc.invalidateQueries({ queryKey: ['admin-payment-orders'] })
    },
    onError: (e: { message?: string }) => toast.error(e?.message || 'Retry failed')
  })

  function applyFilters() {
    setPage(1)
    setFilters({
      search: searchInput.trim(),
      status: statusInput,
      from: fromInput,
      to: toInput
    })
  }

  function resetFilters() {
    setSearchInput('')
    setStatusInput('')
    setFromInput('')
    setToInput('')
    setFilters({})
    setPage(1)
  }

  return (
    <>
      <PageHeader
        title="Payment orders"
        description="Review, filter, cancel, and retry payment orders."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      <Card className="p-4 mb-4">
        <div className="grid grid-cols-2 gap-3 md:grid-cols-[1fr_160px_160px_160px_auto]">
          <Input name="search" label="Search" placeholder="Order, user, email" value={searchInput} onChange={(e) => setSearchInput(e.target.value)} />
          <div>
            <label className="input-label">Status</label>
            <select className="input bg-bg-4" value={statusInput} onChange={(e) => setStatusInput(e.target.value)}>
              {STATUSES.map((status) => (
                <option key={status || 'all'} value={status} className="bg-bg-4">
                  {status || 'All'}
                </option>
              ))}
            </select>
          </div>
          <Input name="from" type="date" label="From" value={fromInput} onChange={(e) => setFromInput(e.target.value)} />
          <Input name="to" type="date" label="To" value={toInput} onChange={(e) => setToInput(e.target.value)} />
          <div className="flex items-end justify-end gap-2">
            <Button variant="ghost" onClick={resetFilters}>Reset</Button>
            <Button variant="accent" onClick={applyFilters} loading={query.isFetching}>Filter</Button>
          </div>
        </div>
      </Card>

      <Card className="p-4">
        {query.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Order</TH>
                <TH>User</TH>
                <TH>Type</TH>
                <TH>Method</TH>
                <TH className="text-right">Amount</TH>
                <TH>Status</TH>
                <TH>Created</TH>
                <TH className="text-right">Actions</TH>
              </TR>
            </THead>
            <TBody>
              {rows.map((row, index) => {
                const id = orderId(row)
                const status = text(row.status ?? row.state)
                const canCancel = ['PENDING', 'CREATED', 'WAITING', 'PROCESSING'].includes(status.toUpperCase())
                const canRetry = ['FAILED'].includes(status.toUpperCase())
                return (
                  <TR key={String(id ?? index)}>
                    <TD className="font-mono text-xs">{text(row.out_trade_no ?? row.trade_no ?? row.order_no ?? row.id)}</TD>
                    <TD className="text-xs text-ink-2">{text(row.user_email ?? row.email ?? row.username ?? row.user_id)}</TD>
                    <TD>{text(row.order_type ?? row.type)}</TD>
                    <TD>{text(row.payment_type ?? row.provider ?? row.method)}</TD>
                    <TD className="text-right font-mono">{money(row.pay_amount ?? row.amount ?? row.total_amount)}</TD>
                    <TD><Badge tone={statusTone(status)}>{status}</Badge></TD>
                    <TD className="text-xs text-ink-3">{dateText(row.created_at ?? row.createdAt)}</TD>
                    <TD>
                      <div className="flex justify-end gap-1">
                        <Button size="icon" variant="ghost" onClick={() => setSelected(row)} title="Details">
                          <Eye className="h-3.5 w-3.5" />
                        </Button>
                        {canRetry && id && (
                          <Button size="icon" variant="ghost" loading={retryMut.isPending} onClick={() => retryMut.mutate(id)} title="Retry">
                            <RotateCcw className="h-3.5 w-3.5" />
                          </Button>
                        )}
                        {canCancel && id && (
                          <Button size="icon" variant="danger" loading={cancelMut.isPending} onClick={() => cancelMut.mutate(id)} title="Cancel">
                            <XCircle className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </div>
                    </TD>
                  </TR>
                )
              })}
              {rows.length === 0 && (
                <TR>
                  <TD colSpan={8} className="py-8 text-center text-ink-3">No data</TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        <div className="mt-4 flex items-center justify-between border-t border-line-1 pt-3 text-sm">
          <div className="text-ink-3">Total: {total.toLocaleString()}</div>
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-ink-3">{page} / {pages}</span>
            <Button variant="ghost" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Back</Button>
            <Button variant="ghost" size="sm" disabled={page >= pages} onClick={() => setPage((p) => p + 1)}>Next</Button>
          </div>
        </div>
      </Card>

      <Modal open={!!selected} onClose={() => setSelected(null)} title="Order details" size="lg">
        <Table>
          <THead>
            <TR><TH>Field</TH><TH>Value</TH></TR>
          </THead>
          <TBody>
            {selected && Object.entries(selected).map(([key, value]) => (
              <TR key={key}>
                <TD className="font-mono text-xs text-ink-3">{key}</TD>
                <TD className="font-mono text-xs break-all text-ink-2">{typeof value === 'object' ? JSON.stringify(value) : text(value)}</TD>
              </TR>
            ))}
          </TBody>
        </Table>
      </Modal>
    </>
  )
}
