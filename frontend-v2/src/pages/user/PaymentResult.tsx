import { Link, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, Clock, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { Badge } from '@/components/ui/Badge'
import { paymentAPI } from '@/api/payment'

function statusKind(status: string) {
  if (['COMPLETED', 'PAID', 'RECHARGING', 'success'].includes(status)) return 'success'
  if (['PENDING', 'CREATED', 'WAITING', 'PROCESSING'].includes(status)) return 'pending'
  return 'failed'
}

function money(v: unknown) {
  return `$${Number(v || 0).toFixed(4)}`
}

export default function PaymentResultPage() {
  const [params] = useSearchParams()
  const orderId = params.get('order_id')
  const resumeToken = params.get('resume_token')
  const outTradeNo = params.get('out_trade_no')
  const queryStatus = params.get('status') || ''

  const query = useQuery({
    queryKey: ['payment-result', orderId, resumeToken, outTradeNo],
    queryFn: async () => {
      if (resumeToken) return paymentAPI.resolvePublicOrder({ resume_token: resumeToken })
      if (orderId) return paymentAPI.getOrder(orderId)
      if (outTradeNo) return paymentAPI.verifyPublicOrder({ out_trade_no: outTradeNo })
      return null
    },
    enabled: !!(orderId || resumeToken || outTradeNo)
  })

  const order: any = (query.data as any)?.order || query.data
  const status = order?.status || queryStatus || 'FAILED'
  const kind = statusKind(status)
  const Icon = kind === 'success' ? CheckCircle2 : kind === 'pending' ? Clock : XCircle

  return (
    <main className="min-h-screen bg-bg-0 p-6 text-ink-1 flex items-center justify-center">
      <Card className="w-full max-w-lg p-6 text-center">
        {query.isLoading ? (
          <div className="space-y-4">
            <Skeleton className="mx-auto h-14 w-14 rounded-full" />
            <Skeleton className="mx-auto h-6 w-48" />
            <Skeleton className="h-28 w-full" />
          </div>
        ) : (
          <>
            <div className={`mx-auto h-14 w-14 rounded-full flex items-center justify-center ${
              kind === 'success' ? 'bg-signal-ok/10 text-signal-ok' : kind === 'pending' ? 'bg-signal-warn/10 text-signal-warn' : 'bg-signal-err/10 text-signal-err'
            }`}>
              <Icon className="h-7 w-7" />
            </div>
            <h1 className="mt-4 font-display text-3xl text-ink-1">
              {kind === 'success' ? 'Payment completed' : kind === 'pending' ? 'Payment processing' : 'Payment not completed'}
            </h1>
            <p className="mt-2 text-sm text-ink-3">
              {kind === 'pending' ? 'The order is still being processed. Check orders for the latest state.' : 'You can return to the console from here.'}
            </p>

            {order && (
              <div className="mt-5 rounded-lg border border-line-2 bg-bg-2 p-4 text-left space-y-2">
                <div className="flex justify-between gap-3 text-sm"><span className="text-ink-3">Order</span><code className="font-mono">{order.out_trade_no || order.id}</code></div>
                <div className="flex justify-between gap-3 text-sm"><span className="text-ink-3">Amount</span><code className="font-mono">{money(order.pay_amount ?? order.amount)}</code></div>
                <div className="flex justify-between gap-3 text-sm"><span className="text-ink-3">Method</span><span>{order.payment_type || '-'}</span></div>
                <div className="flex justify-between gap-3 text-sm"><span className="text-ink-3">Status</span><Badge>{status}</Badge></div>
              </div>
            )}

            {!order && outTradeNo && (
              <div className="mt-5 rounded-lg border border-line-2 bg-bg-2 p-4 text-left text-sm text-ink-2">
                Legacy return: <code className="font-mono">{outTradeNo}</code>
              </div>
            )}

            <div className="mt-6 flex flex-col sm:flex-row gap-2 justify-center">
              <Link to="/purchase" className="btn btn-ghost">Back to purchase</Link>
              <Link to="/orders" className="btn btn-primary">View orders</Link>
            </div>
          </>
        )}
      </Card>
    </main>
  )
}
