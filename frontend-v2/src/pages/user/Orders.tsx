import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { RefreshCw, RotateCcw, XCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { paymentAPI } from '@/api/payment'

const STATUSES = ['', 'PENDING', 'COMPLETED', 'FAILED', 'REFUNDED', 'CANCELLED']

function items(data: any): any[] {
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.items)) return data.items
  if (Array.isArray(data?.orders)) return data.orders
  return []
}

function pages(data: any) {
  return Number(data?.pages || 1)
}

function total(data: any) {
  return Number(data?.total || items(data).length)
}

function money(v: unknown) {
  return `$${Number(v || 0).toFixed(4)}`
}

function statusTone(status: string) {
  if (['COMPLETED', 'PAID'].includes(status)) return 'success' as const
  if (['PENDING', 'RECHARGING', 'REFUND_REQUESTED'].includes(status)) return 'warning' as const
  if (['FAILED', 'EXPIRED', 'CANCELLED', 'REFUND_FAILED'].includes(status)) return 'danger' as const
  return 'neutral' as const
}

export default function OrdersPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [refundOrder, setRefundOrder] = useState<any | null>(null)
  const [refundReason, setRefundReason] = useState('')

  const query = useQuery({
    queryKey: ['my-payment-orders', page, status],
    queryFn: () => paymentAPI.getMyOrders({ page, page_size: 20, status: status || undefined })
  })

  const eligibleQuery = useQuery({
    queryKey: ['refund-eligible-providers'],
    queryFn: () => paymentAPI.getRefundEligibleProviders(),
    retry: false
  })

  const cancelMut = useMutation({
    mutationFn: (id: number | string) => paymentAPI.cancelOrder(id),
    onSuccess: () => {
      toast.success(t('common.success') as string)
      qc.invalidateQueries({ queryKey: ['my-payment-orders'] })
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const refundMut = useMutation({
    mutationFn: ({ id, reason }: { id: number | string; reason: string }) =>
      paymentAPI.requestOrderRefund(id, { reason }),
    onSuccess: () => {
      toast.success(t('common.success') as string)
      setRefundOrder(null)
      setRefundReason('')
      qc.invalidateQueries({ queryKey: ['my-payment-orders'] })
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const rows = items(query.data)
  const eligible = new Set(
    Array.isArray(eligibleQuery.data)
      ? eligibleQuery.data.map(String)
      : Array.isArray((eligibleQuery.data as any)?.providers)
        ? (eligibleQuery.data as any).providers.map(String)
        : []
  )

  return (
    <>
      <PageHeader
        title={t('nav.myOrders')}
        description="Review payment orders, cancel pending orders, and request refunds when available."
        actions={
          <Link to="/purchase" className="btn btn-accent">
            Purchase
          </Link>
        }
      />

      <Card className="p-4 mb-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div className="w-full sm:max-w-xs">
            <label className="input-label">Status</label>
            <select className="input bg-bg-4" value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }}>
              {STATUSES.map((s) => (
                <option key={s || 'all'} value={s} className="bg-bg-4">
                  {s || t('common.all')}
                </option>
              ))}
            </select>
          </div>
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            {t('common.refresh')}
          </Button>
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
                <TH>Type</TH>
                <TH>Method</TH>
                <TH className="text-right">Amount</TH>
                <TH>Status</TH>
                <TH>Created</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {rows.map((o) => {
                const canRefund =
                  o.status === 'COMPLETED' &&
                  o.provider_instance_id &&
                  (eligible.size === 0 ? false : eligible.has(String(o.provider_instance_id)))
                return (
                  <TR key={o.id}>
                    <TD className="font-mono text-xs">{o.out_trade_no || `#${o.id}`}</TD>
                    <TD>{o.order_type || '-'}</TD>
                    <TD>{o.payment_type || '-'}</TD>
                    <TD className="text-right font-mono">{money(o.pay_amount ?? o.amount)}</TD>
                    <TD><Badge tone={statusTone(o.status)}>{o.status}</Badge></TD>
                    <TD className="text-xs text-ink-3">{o.created_at ? new Date(o.created_at).toLocaleString() : '-'}</TD>
                    <TD className="text-right">
                      <div className="inline-flex gap-1">
                        {o.status === 'PENDING' && (
                          <button
                            className="btn btn-ghost btn-icon btn-sm text-signal-err"
                            title="Cancel"
                            onClick={() => confirm('Cancel this order?') && cancelMut.mutate(o.id)}
                          >
                            <XCircle className="h-3.5 w-3.5" />
                          </button>
                        )}
                        {canRefund && (
                          <button
                            className="btn btn-ghost btn-icon btn-sm"
                            title="Refund"
                            onClick={() => setRefundOrder(o)}
                          >
                            <RotateCcw className="h-3.5 w-3.5" />
                          </button>
                        )}
                      </div>
                    </TD>
                  </TR>
                )
              })}
              {rows.length === 0 && (
                <TR>
                  <TD colSpan={7} className="text-center text-ink-3 py-8">{t('common.noData')}</TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        {pages(query.data) > 1 && (
          <div className="flex items-center justify-between mt-4 pt-3 border-t border-line-2 text-sm">
            <div className="text-ink-3">{t('common.total')}: {total(query.data)}</div>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>{t('common.back')}</Button>
              <Button variant="secondary" size="sm" disabled={page >= pages(query.data)} onClick={() => setPage((p) => p + 1)}>{t('common.next')}</Button>
            </div>
          </div>
        )}
      </Card>

      <Modal
        open={!!refundOrder}
        onClose={() => setRefundOrder(null)}
        title="Request refund"
        footer={
          <>
            <Button variant="secondary" onClick={() => setRefundOrder(null)}>{t('common.cancel')}</Button>
            <Button
              variant="danger"
              loading={refundMut.isPending}
              disabled={!refundReason.trim()}
              onClick={() => refundOrder && refundMut.mutate({ id: refundOrder.id, reason: refundReason.trim() })}
            >
              Submit
            </Button>
          </>
        }
      >
        <Input
          name="refund_reason"
          label="Reason"
          value={refundReason}
          onChange={(e) => setRefundReason(e.target.value)}
          placeholder="Explain why this order should be refunded"
        />
      </Modal>
    </>
  )
}
