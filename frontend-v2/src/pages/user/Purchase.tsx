import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { CreditCard, ExternalLink, ShoppingCart } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { toast } from '@/components/ui/Toast'
import { paymentAPI, type PaymentPlan, type PaymentChannel } from '@/api/payment'
import { useAuthStore } from '@/stores/auth'

const PRESETS = [10, 20, 50, 100, 200, 500, 1000]

function money(v: unknown) {
  return `$${Number(v || 0).toFixed(2)}`
}

function asPlans(data: any): PaymentPlan[] {
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.plans)) return data.plans
  return []
}

function asChannels(data: any): PaymentChannel[] {
  if (Array.isArray(data)) return data
  if (Array.isArray(data?.channels)) return data.channels
  if (Array.isArray(data?.methods)) return data.methods
  if (data?.methods && typeof data.methods === 'object') {
    return Object.keys(data.methods).map((key, i) => ({ id: i + 1, name: key, provider: key, enabled: true }))
  }
  return []
}

export default function PurchasePage() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const [mode, setMode] = useState<'balance' | 'subscription'>('balance')
  const [amount, setAmount] = useState('50')
  const [selectedPlan, setSelectedPlan] = useState<number | null>(null)
  const [paymentType, setPaymentType] = useState('')

  const checkout = useQuery({
    queryKey: ['payment-checkout-info'],
    queryFn: () => paymentAPI.getPaymentCheckoutInfo()
  })
  const plansQuery = useQuery({
    queryKey: ['payment-plans'],
    queryFn: () => paymentAPI.getPaymentPlans()
  })
  const channelsQuery = useQuery({
    queryKey: ['payment-channels'],
    queryFn: () => paymentAPI.getPaymentChannels()
  })

  const plans = useMemo(() => asPlans(checkout.data).length ? asPlans(checkout.data) : asPlans(plansQuery.data), [checkout.data, plansQuery.data])
  const channels = useMemo(() => {
    const fromCheckout = asChannels(checkout.data)
    return fromCheckout.length ? fromCheckout : asChannels(channelsQuery.data)
  }, [checkout.data, channelsQuery.data])
  const enabledChannels = channels.filter((c) => c.enabled !== false)
  const effectivePaymentType = paymentType || enabledChannels[0]?.provider || enabledChannels[0]?.name || ''
  const plan = plans.find((p) => p.id === selectedPlan)
  const numericAmount = mode === 'subscription' ? Number(plan?.price || 0) : Number(amount)

  const createMut = useMutation({
    mutationFn: () => {
      if (!effectivePaymentType) throw new Error('Select a payment method')
      if (!numericAmount || numericAmount <= 0) throw new Error('Enter a valid amount')
      return paymentAPI.createPaymentOrder({
        amount: numericAmount,
        payment_type: effectivePaymentType,
        order_type: mode,
        plan_id: mode === 'subscription' ? selectedPlan : undefined,
        return_url: typeof window === 'undefined' ? '/payment/result' : `${window.location.origin}/payment/result`,
        is_mobile: typeof window !== 'undefined' ? /Mobile|Android|iPhone/i.test(window.navigator.userAgent) : false,
        payment_source: 'hosted_redirect'
      })
    },
    onSuccess: (order: any) => {
      toast.success('Order created')
      if (order?.pay_url) {
        window.location.href = order.pay_url
        return
      }
      if (order?.qr_code) {
        navigate(`/payment/qrcode?order_id=${encodeURIComponent(order.id)}&qr=${encodeURIComponent(order.qr_code)}&payment_type=${encodeURIComponent(effectivePaymentType)}&expires_at=${encodeURIComponent(order.expires_at || '')}`)
        return
      }
      navigate(`/orders`)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || 'Failed to create order')
  })

  const loading = checkout.isLoading || plansQuery.isLoading || channelsQuery.isLoading

  return (
    <>
      <PageHeader
        title="Recharge / Subscription"
        description="Create a balance recharge or subscription order using the configured payment methods."
        actions={<Link to="/orders" className="btn btn-ghost">My orders</Link>}
      />

      {loading ? (
        <div className="grid gap-4 lg:grid-cols-[1fr_360px]">
          <Card className="p-5 space-y-3"><Skeleton className="h-8 w-40" /><Skeleton className="h-40 w-full" /></Card>
          <Card className="p-5 space-y-3"><Skeleton className="h-8 w-32" /><Skeleton className="h-24 w-full" /></Card>
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-[1fr_360px]">
          <div className="space-y-4">
            <Card className="p-4">
              <div className="inline-flex rounded-lg border border-line-2 bg-bg-2 p-1">
                {(['balance', 'subscription'] as const).map((m) => (
                  <button
                    key={m}
                    className={`px-3 py-1.5 rounded-md text-sm ${mode === m ? 'bg-orange text-white' : 'text-ink-2 hover:bg-bg-3'}`}
                    onClick={() => setMode(m)}
                  >
                    {m === 'balance' ? 'Balance recharge' : 'Subscription'}
                  </button>
                ))}
              </div>
            </Card>

            {mode === 'balance' ? (
              <Card className="p-5 space-y-4">
                <div>
                  <h2 className="text-base font-medium text-ink-1">Recharge amount</h2>
                  <p className="text-sm text-ink-3 mt-1">Current balance: <span className="font-mono text-ink-2">{money(user?.balance)}</span></p>
                </div>
                <Input name="amount" type="number" min="0" step="0.01" label="Amount" value={amount} onChange={(e) => setAmount(e.target.value)} leftIcon={<CreditCard className="h-4 w-4" />} />
                <div className="flex flex-wrap gap-2">
                  {PRESETS.map((p) => (
                    <button key={p} className={`px-3 py-1.5 rounded-md border text-sm ${Number(amount) === p ? 'bg-orange text-white border-orange' : 'border-line-2 text-ink-2 hover:bg-bg-3'}`} onClick={() => setAmount(String(p))}>
                      {money(p)}
                    </button>
                  ))}
                </div>
              </Card>
            ) : (
              <div className="grid gap-3 md:grid-cols-2">
                {plans.map((p) => (
                  <button key={p.id} className="text-left" onClick={() => setSelectedPlan(p.id)}>
                    <Card className={`p-5 h-full transition-colors ${selectedPlan === p.id ? 'border-orange' : 'hover:border-orange/40'}`}>
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <h2 className="font-medium text-ink-1">{p.name}</h2>
                          {p.description && <p className="text-sm text-ink-3 mt-1">{p.description}</p>}
                        </div>
                        <Badge tone="accent">{p.group_platform || 'plan'}</Badge>
                      </div>
                      <div className="mt-4 font-display text-3xl text-ink-1">{money(p.price)}</div>
                      <div className="mt-2 text-xs text-ink-3">{p.daily_limit_usd ? `Daily ${money(p.daily_limit_usd)}` : 'Unlimited daily quota'}</div>
                    </Card>
                  </button>
                ))}
                {plans.length === 0 && <Card className="p-12 text-center text-ink-3 md:col-span-2">No plans available</Card>}
              </div>
            )}
          </div>

          <Card className="p-5 h-fit space-y-4">
            <div>
              <h2 className="text-base font-medium text-ink-1">Checkout</h2>
              <p className="text-sm text-ink-3 mt-1">Choose a payment method and create an order.</p>
            </div>
            <div>
              <label className="input-label">Payment method</label>
              <select className="input bg-bg-4" value={effectivePaymentType} onChange={(e) => setPaymentType(e.target.value)}>
                {enabledChannels.map((c) => (
                  <option key={`${c.id}-${c.provider}`} value={c.provider || c.name} className="bg-bg-4">
                    {c.name || c.provider}
                  </option>
                ))}
              </select>
            </div>
            <div className="rounded-lg border border-line-2 bg-bg-2 p-4 space-y-2 text-sm">
              <div className="flex justify-between"><span className="text-ink-3">Order type</span><span>{mode}</span></div>
              <div className="flex justify-between"><span className="text-ink-3">Amount</span><code className="font-mono">{money(numericAmount)}</code></div>
              {mode === 'subscription' && <div className="flex justify-between"><span className="text-ink-3">Plan</span><span>{plan?.name || '-'}</span></div>}
            </div>
            <Button
              className="w-full"
              variant="accent"
              loading={createMut.isPending}
              disabled={!effectivePaymentType || numericAmount <= 0 || (mode === 'subscription' && !selectedPlan)}
              onClick={() => createMut.mutate()}
            >
              <ShoppingCart className="h-4 w-4" />
              Create order
            </Button>
            <div className="text-xs text-ink-3 flex items-start gap-2">
              <ExternalLink className="h-3.5 w-3.5 mt-0.5 shrink-0" />
              Some providers may redirect to an external payment window.
            </div>
          </Card>
        </div>
      )}
    </>
  )
}
