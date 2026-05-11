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

function asPlans(data: unknown): UnknownRecord[] {
  if (Array.isArray(data)) return data.filter((item): item is UnknownRecord => !!item && typeof item === 'object')
  if (!data || typeof data !== 'object') return []
  const record = data as UnknownRecord
  for (const key of ['items', 'plans', 'data']) {
    const value = record[key]
    if (Array.isArray(value)) return value.filter((item): item is UnknownRecord => !!item && typeof item === 'object')
  }
  return []
}

function text(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function money(value: unknown) {
  const number = Number(value || 0)
  return `$${(Number.isFinite(number) ? number : 0).toFixed(2)}`
}

function limit(value: unknown) {
  const number = Number(value || 0)
  if (!Number.isFinite(number) || number <= 0) return 'Unlimited'
  return money(number)
}

function enabled(plan: UnknownRecord) {
  if (typeof plan.enabled === 'boolean') return plan.enabled
  if (typeof plan.is_active === 'boolean') return plan.is_active
  if (typeof plan.active === 'boolean') return plan.active
  if (typeof plan.status === 'string') return !['disabled', 'inactive', 'archived'].includes(plan.status.toLowerCase())
  return true
}

export default function AdminPaymentPlansPage() {
  const query = useQuery({
    queryKey: ['admin-payment-plans'],
    queryFn: () => adminPaymentAPI.getPaymentPlans()
  })

  const plans = useMemo(() => asPlans(query.data), [query.data])

  return (
    <>
      <PageHeader
        title="Payment plans"
        description="View configured subscription and recharge plans."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

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
              <TR>
                <TH>Plan</TH>
                <TH>Platform</TH>
                <TH className="text-right">Price</TH>
                <TH className="text-right">Daily</TH>
                <TH className="text-right">Weekly</TH>
                <TH className="text-right">Monthly</TH>
                <TH>Status</TH>
              </TR>
            </THead>
            <TBody>
              {plans.map((plan, index) => (
                <TR key={text(plan.id) === '-' ? index : text(plan.id)}>
                  <TD>
                    <div className="font-medium text-ink-1">{text(plan.name)}</div>
                    {plan.description ? (
                      <div className="mt-0.5 max-w-md truncate text-xs text-ink-3">{text(plan.description)}</div>
                    ) : null}
                  </TD>
                  <TD>{text(plan.group_platform ?? plan.platform ?? plan.group)}</TD>
                  <TD className="text-right font-mono">{money(plan.price ?? plan.amount)}</TD>
                  <TD className="text-right font-mono">{limit(plan.daily_limit_usd ?? plan.daily_limit)}</TD>
                  <TD className="text-right font-mono">{limit(plan.weekly_limit_usd ?? plan.weekly_limit)}</TD>
                  <TD className="text-right font-mono">{limit(plan.monthly_limit_usd ?? plan.monthly_limit)}</TD>
                  <TD>
                    {enabled(plan) ? (
                      <Badge tone="success" dot>
                        active
                      </Badge>
                    ) : (
                      <Badge dot>disabled</Badge>
                    )}
                  </TD>
                </TR>
              ))}
              {plans.length === 0 && (
                <TR>
                  <TD colSpan={7} className="py-8 text-center text-ink-3">
                    No data
                  </TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}
      </Card>
    </>
  )
}
