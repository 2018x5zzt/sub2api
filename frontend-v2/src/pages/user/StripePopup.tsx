import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, ExternalLink, Loader2, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { paymentAPI } from '@/api/payment'

function done(status: string) {
  return ['COMPLETED', 'PAID', 'FAILED', 'EXPIRED', 'CANCELLED'].includes(status.toUpperCase())
}

export default function StripePopupPage() {
  const [params] = useSearchParams()
  const orderId = params.get('order_id') || ''
  const payUrl = params.get('pay_url') || params.get('redirect_url') || ''
  const method = params.get('method') || 'stripe'
  const [opened, setOpened] = useState(false)

  const orderQuery = useQuery({
    queryKey: ['stripe-popup-order', orderId],
    queryFn: () => paymentAPI.getOrder(orderId),
    enabled: !!orderId,
    refetchInterval: (query) => done(String((query.state.data as any)?.status || '')) ? false : 3000
  })

  const status = String((orderQuery.data as any)?.status || 'PENDING')
  const paid = ['COMPLETED', 'PAID'].includes(status.toUpperCase())
  const failed = ['FAILED', 'EXPIRED', 'CANCELLED'].includes(status.toUpperCase())

  useEffect(() => {
    if (payUrl && !opened) {
      setOpened(true)
      window.location.href = payUrl
    }
  }, [opened, payUrl])

  return (
    <main className="min-h-screen bg-bg-0 p-4 text-ink-1 flex items-center justify-center">
      <Card className="w-full max-w-sm p-6 text-center">
        <div className="mx-auto h-12 w-12 rounded-lg bg-orange-soft text-orange flex items-center justify-center">
          {paid ? <CheckCircle2 className="h-6 w-6 text-signal-ok" /> : failed ? <XCircle className="h-6 w-6 text-signal-err" /> : <Loader2 className="h-6 w-6 animate-spin" />}
        </div>
        <h1 className="mt-4 font-display text-xl">Payment window</h1>
        <p className="mt-1 text-sm text-ink-3">{method} order {orderId || '-'}</p>
        <div className="mt-4"><Badge>{status}</Badge></div>
        <div className="mt-6 grid gap-2">
          {payUrl && (
            <a href={payUrl} className="btn btn-primary" target="_blank" rel="noreferrer">
              <ExternalLink className="h-4 w-4" />
              Open payment
            </a>
          )}
          <Button variant="ghost" onClick={() => window.close()}>Close</Button>
        </div>
      </Card>
    </main>
  )
}
