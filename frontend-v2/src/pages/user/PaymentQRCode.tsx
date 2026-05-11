import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { QrCode, XCircle } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { paymentAPI } from '@/api/payment'
import { toast } from '@/components/ui/Toast'

function secondsLeft(expiresAt: string | null) {
  if (!expiresAt) return 30 * 60
  return Math.max(0, Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000))
}

export default function PaymentQRCodePage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const orderId = params.get('order_id') || ''
  const qr = params.get('qr') || ''
  const payUrl = params.get('pay_url') || ''
  const paymentType = params.get('payment_type') || ''
  const expiresAt = params.get('expires_at')
  const [left, setLeft] = useState(() => secondsLeft(expiresAt))

  const query = useQuery({
    queryKey: ['payment-qrcode-order', orderId],
    queryFn: () => paymentAPI.getOrder(orderId),
    enabled: !!orderId,
    refetchInterval: (q) => {
      const status = (q.state.data as any)?.status
      return ['COMPLETED', 'PAID', 'FAILED', 'EXPIRED', 'CANCELLED'].includes(status) ? false : 3000
    }
  })

  useEffect(() => {
    const timer = window.setInterval(() => setLeft(secondsLeft(expiresAt)), 1000)
    return () => window.clearInterval(timer)
  }, [expiresAt])

  useEffect(() => {
    const status = (query.data as any)?.status
    if (['COMPLETED', 'PAID'].includes(status)) {
      navigate(`/payment/result?order_id=${encodeURIComponent(orderId)}&status=success`, { replace: true })
    }
  }, [navigate, orderId, query.data])

  const cancelMut = useMutation({
    mutationFn: () => paymentAPI.cancelOrder(orderId),
    onSuccess: () => navigate('/purchase'),
    onError: (e: { message?: string }) => toast.error(e?.message || 'Cancel failed')
  })

  const expired = left <= 0 || ['EXPIRED', 'CANCELLED', 'FAILED'].includes(String((query.data as any)?.status || ''))
  const mmss = useMemo(() => {
    const m = Math.floor(left / 60)
    const s = left % 60
    return `${m}:${String(s).padStart(2, '0')}`
  }, [left])

  return (
    <main className="min-h-screen bg-bg-0 p-6 text-ink-1 flex items-center justify-center">
      <Card className="w-full max-w-md p-6 text-center">
        <div className="mx-auto h-12 w-12 rounded-lg bg-orange-soft text-orange flex items-center justify-center">
          <QrCode className="h-6 w-6" />
        </div>
        <h1 className="mt-4 font-display text-2xl">Scan to pay</h1>
        <p className="mt-1 text-sm text-ink-3">{paymentType || 'payment'} order {orderId || '-'}</p>

        <div className="mt-5 rounded-lg border border-line-2 bg-bg-2 p-4">
          {qr ? (
            <img src={qr} alt="Payment QR" className="mx-auto aspect-square w-56 rounded-md bg-white p-2 object-contain" />
          ) : payUrl ? (
            <a className="btn btn-primary" href={payUrl} target="_blank" rel="noreferrer">Open payment window</a>
          ) : (
            <div className="py-12 text-sm text-ink-3">No QR payload available</div>
          )}
        </div>

        <div className="mt-4 flex items-center justify-center gap-2">
          {expired ? <Badge tone="danger">expired</Badge> : <Badge tone="warning">waiting</Badge>}
          <code className="font-mono text-sm text-ink-2">{mmss}</code>
        </div>

        <div className="mt-6 flex gap-2 justify-center">
          <Link to="/purchase" className="btn btn-ghost">Back</Link>
          {!expired && orderId && (
            <Button variant="danger" loading={cancelMut.isPending} onClick={() => cancelMut.mutate()}>
              <XCircle className="h-4 w-4" />
              Cancel
            </Button>
          )}
        </div>
      </Card>
    </main>
  )
}
