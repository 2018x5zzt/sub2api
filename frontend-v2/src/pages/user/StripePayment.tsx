import { useEffect, useMemo } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CreditCard, ExternalLink, Loader2 } from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { paymentAPI } from '@/api/payment'

function terminal(status: string) {
  return ['COMPLETED', 'PAID', 'FAILED', 'EXPIRED', 'CANCELLED', 'REFUNDED'].includes(status.toUpperCase())
}

function success(status: string) {
  return ['COMPLETED', 'PAID', 'RECHARGING'].includes(status.toUpperCase())
}

function money(value: unknown) {
  const n = Number(value || 0)
  return `$${(Number.isFinite(n) ? n : 0).toFixed(2)}`
}

export default function StripePaymentPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const orderId = params.get('order_id') || ''
  const clientSecret = params.get('client_secret') || ''
  const method = params.get('method') || params.get('payment_type') || 'stripe'
  const payUrl = params.get('pay_url') || params.get('redirect_url') || ''

  const orderQuery = useQuery({
    queryKey: ['stripe-payment-order', orderId],
    queryFn: () => paymentAPI.getOrder(orderId),
    enabled: !!orderId,
    refetchInterval: (query) => {
      const status = String((query.state.data as any)?.status || '')
      return terminal(status) ? false : 3000
    }
  })

  const order: any = orderQuery.data
  const status = String(order?.status || params.get('status') || 'PENDING')
  const canOpen = !!payUrl
  const resultUrl = useMemo(() => {
    const query = orderId ? `?order_id=${encodeURIComponent(orderId)}&status=${success(status) ? 'success' : 'pending'}` : ''
    return `/payment/result${query}`
  }, [orderId, status])

  useEffect(() => {
    if (success(status) && orderId) {
      navigate(`/payment/result?order_id=${encodeURIComponent(orderId)}&status=success`, { replace: true })
    }
  }, [navigate, orderId, status])

  return (
    <main className="min-h-screen bg-bg-0 p-6 text-ink-1 flex items-center justify-center">
      <Card className="w-full max-w-lg p-6">
        <div className="text-center">
          <div className="mx-auto h-12 w-12 rounded-lg bg-orange-soft text-orange flex items-center justify-center">
            <CreditCard className="h-6 w-6" />
          </div>
          <h1 className="mt-4 font-display text-2xl">Stripe payment</h1>
          <p className="mt-1 text-sm text-ink-3">Complete the payment in the provider window, then return here for status.</p>
        </div>

        {orderQuery.isLoading ? (
          <div className="mt-5 space-y-3">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : (
          <div className="mt-5 rounded-lg border border-line-2 bg-bg-2 p-4 space-y-2 text-sm">
            <div className="flex justify-between gap-3"><span className="text-ink-3">Order</span><code className="font-mono">{order?.out_trade_no || orderId || '-'}</code></div>
            <div className="flex justify-between gap-3"><span className="text-ink-3">Method</span><span>{method}</span></div>
            <div className="flex justify-between gap-3"><span className="text-ink-3">Amount</span><code className="font-mono">{money(order?.pay_amount ?? order?.amount)}</code></div>
            <div className="flex justify-between gap-3"><span className="text-ink-3">Status</span><Badge>{status}</Badge></div>
            {clientSecret && <div className="pt-2 text-xs text-ink-3 break-all">Client secret received. Stripe Elements is not bundled in frontend-v2 yet, so this compatibility page uses provider redirects and order polling.</div>}
          </div>
        )}

        <div className="mt-6 grid gap-2 sm:grid-cols-2">
          {canOpen ? (
            <a className="btn btn-primary" href={payUrl} target="_blank" rel="noreferrer">
              <ExternalLink className="h-4 w-4" />
              Open provider
            </a>
          ) : (
            <Button disabled>
              <Loader2 className="h-4 w-4" />
              Waiting for provider
            </Button>
          )}
          <Link to={resultUrl} className="btn btn-ghost">Check result</Link>
          <Link to="/purchase" className="btn btn-ghost">Back to purchase</Link>
          <Link to="/orders" className="btn btn-ghost">View orders</Link>
        </div>
      </Card>
    </main>
  )
}
