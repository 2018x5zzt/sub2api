import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { adminPaymentAPI } from '@/api/admin/payment'

type UnknownRecord = Record<string, unknown>

function asRecord(value: unknown): UnknownRecord {
  return value && typeof value === 'object' && !Array.isArray(value) ? (value as UnknownRecord) : {}
}

function asArray(value: unknown): UnknownRecord[] {
  if (!Array.isArray(value)) return []
  return value.filter((item): item is UnknownRecord => !!item && typeof item === 'object' && !Array.isArray(item))
}

function pickNumber(source: UnknownRecord, keys: string[]) {
  for (const key of keys) {
    const value = source[key]
    const number = typeof value === 'number' ? value : typeof value === 'string' ? Number(value) : NaN
    if (Number.isFinite(number)) return number
  }
  return 0
}

function pickString(source: UnknownRecord, keys: string[]) {
  for (const key of keys) {
    const value = source[key]
    if (typeof value === 'string' && value.trim()) return value
    if (typeof value === 'number') return String(value)
  }
  return ''
}

function money(value: unknown) {
  const number = Number(value || 0)
  return `$${(Number.isFinite(number) ? number : 0).toFixed(2)}`
}

function numberText(value: unknown) {
  const number = Number(value || 0)
  return (Number.isFinite(number) ? number : 0).toLocaleString()
}

function formatDate(value: unknown) {
  if (!value) return '-'
  const date = new Date(String(value))
  return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString()
}

function statusTone(status: string) {
  const normalized = status.toUpperCase()
  if (['COMPLETED', 'PAID', 'SUCCESS', 'SUCCEEDED', 'ACTIVE'].includes(normalized)) return 'success' as const
  if (['PENDING', 'PROCESSING', 'RECHARGING'].includes(normalized)) return 'warning' as const
  if (['FAILED', 'CANCELLED', 'CANCELED', 'EXPIRED', 'REFUNDED'].includes(normalized)) return 'danger' as const
  return 'neutral' as const
}

export default function AdminPaymentDashboardPage() {
  const query = useQuery({
    queryKey: ['admin-payment-dashboard'],
    queryFn: () => adminPaymentAPI.getPaymentDashboard()
  })

  const dashboard = asRecord(query.data)
  const stats = useMemo(
    () => [
      {
        label: 'Total revenue',
        value: money(pickNumber(dashboard, ['total_revenue', 'total_amount', 'revenue', 'paid_amount'])),
        hint: 'All paid payment volume'
      },
      {
        label: 'Today revenue',
        value: money(pickNumber(dashboard, ['today_revenue', 'today_amount', 'daily_revenue'])),
        hint: 'Paid volume today'
      },
      {
        label: 'Orders',
        value: numberText(pickNumber(dashboard, ['total_orders', 'orders_count', 'order_count', 'orders'])),
        hint: 'All known orders'
      },
      {
        label: 'Pending',
        value: numberText(pickNumber(dashboard, ['pending_orders', 'pending_count', 'pending'])),
        hint: 'Orders awaiting completion'
      }
    ],
    [dashboard]
  )

  const recentOrders = asArray(
    dashboard.recent_orders ?? dashboard.latest_orders ?? dashboard.orders ?? dashboard.recentOrders
  )
  const providerStats = asArray(
    dashboard.provider_stats ?? dashboard.providers ?? dashboard.provider_summary ?? dashboard.providerStats
  )

  return (
    <>
      <PageHeader
        title="Payment dashboard"
        description="Monitor payment totals, recent orders, and provider activity."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      {query.isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <Card key={index} className="p-5 space-y-3">
              <Skeleton className="h-4 w-28" />
              <Skeleton className="h-9 w-36" />
              <Skeleton className="h-3 w-40" />
            </Card>
          ))}
        </div>
      ) : (
        <div className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat) => (
              <Card key={stat.label} className="p-5">
                <div className="text-sm text-ink-3">{stat.label}</div>
                <div className="mt-2 font-display text-3xl text-ink-1">{stat.value}</div>
                <div className="mt-1 text-xs text-ink-3">{stat.hint}</div>
              </Card>
            ))}
          </div>

          {query.isError && (
            <Card className="p-5 text-sm text-signal-err">
              Failed to load payment dashboard.
            </Card>
          )}

          <div className="grid gap-4 xl:grid-cols-[1fr_380px]">
            <Card className="p-4">
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-base font-medium text-ink-1">Recent orders</h2>
                <Badge>{recentOrders.length}</Badge>
              </div>
              <Table>
                <THead>
                  <TR>
                    <TH>Order</TH>
                    <TH>User</TH>
                    <TH>Type</TH>
                    <TH className="text-right">Amount</TH>
                    <TH>Status</TH>
                    <TH>Created</TH>
                  </TR>
                </THead>
                <TBody>
                  {recentOrders.map((order, index) => {
                    const status = pickString(order, ['status', 'state']) || '-'
                    return (
                      <TR key={pickString(order, ['id', 'out_trade_no', 'trade_no']) || index}>
                        <TD className="font-mono text-xs">
                          {pickString(order, ['out_trade_no', 'trade_no', 'order_no', 'id']) || '-'}
                        </TD>
                        <TD className="text-ink-2 text-xs">
                          {pickString(order, ['user_email', 'email', 'username', 'user_id']) || '-'}
                        </TD>
                        <TD>{pickString(order, ['order_type', 'type', 'payment_type']) || '-'}</TD>
                        <TD className="text-right font-mono">
                          {money(pickNumber(order, ['pay_amount', 'amount', 'total_amount', 'price']))}
                        </TD>
                        <TD>
                          <Badge tone={statusTone(status)}>{status}</Badge>
                        </TD>
                        <TD className="text-xs text-ink-3">{formatDate(order.created_at ?? order.createdAt)}</TD>
                      </TR>
                    )
                  })}
                  {recentOrders.length === 0 && (
                    <TR>
                      <TD colSpan={6} className="py-8 text-center text-ink-3">
                        No data
                      </TD>
                    </TR>
                  )}
                </TBody>
              </Table>
            </Card>

            <Card className="p-4">
              <div className="mb-3 flex items-center justify-between">
                <h2 className="text-base font-medium text-ink-1">Providers</h2>
                <Badge>{providerStats.length}</Badge>
              </div>
              <Table>
                <THead>
                  <TR>
                    <TH>Provider</TH>
                    <TH className="text-right">Orders</TH>
                    <TH className="text-right">Amount</TH>
                  </TR>
                </THead>
                <TBody>
                  {providerStats.map((provider, index) => (
                    <TR key={pickString(provider, ['provider', 'name', 'payment_type']) || index}>
                      <TD>{pickString(provider, ['provider', 'name', 'payment_type']) || '-'}</TD>
                      <TD className="text-right font-mono">
                        {numberText(pickNumber(provider, ['orders', 'order_count', 'count']))}
                      </TD>
                      <TD className="text-right font-mono">
                        {money(pickNumber(provider, ['amount', 'total_amount', 'revenue']))}
                      </TD>
                    </TR>
                  ))}
                  {providerStats.length === 0 && (
                    <TR>
                      <TD colSpan={3} className="py-8 text-center text-ink-3">
                        No data
                      </TD>
                    </TR>
                  )}
                </TBody>
              </Table>
            </Card>
          </div>
        </div>
      )}
    </>
  )
}
