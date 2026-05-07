import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, Users } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminPromoAPI } from '@/api/admin/promo'
import { toast } from '@/components/ui/Toast'
import type {
  CreatePromoCodeRequest,
  PromoCode,
  PromoCodeScene,
  UpdatePromoCodeRequest
} from '@/types'

const SCENES: PromoCodeScene[] = ['register', 'benefit']
const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

interface FormState {
  code: string
  scene: PromoCodeScene
  bonus_amount: string
  random_bonus_pool_amount: string
  max_uses: string
  leaderboard_enabled: boolean
  expires_at: string
  success_message: string
  notes: string
}

const empty: FormState = {
  code: '',
  scene: 'register',
  bonus_amount: '0',
  random_bonus_pool_amount: '0',
  max_uses: '0',
  leaderboard_enabled: false,
  expires_at: '',
  success_message: '',
  notes: ''
}

function fromPromo(p: PromoCode): FormState {
  return {
    code: p.code,
    scene: p.scene,
    bonus_amount: String(p.bonus_amount),
    random_bonus_pool_amount: String(p.random_bonus_pool_amount),
    max_uses: String(p.max_uses),
    leaderboard_enabled: p.leaderboard_enabled,
    expires_at: p.expires_at ? p.expires_at.slice(0, 16) : '',
    success_message: p.success_message ?? '',
    notes: p.notes ?? ''
  }
}

function toPayload(f: FormState): CreatePromoCodeRequest {
  const epoch = (s: string) => {
    if (!s) return null
    const t = new Date(s).getTime()
    return Number.isFinite(t) ? Math.floor(t / 1000) : null
  }
  return {
    code: f.code.trim() || undefined,
    scene: f.scene,
    bonus_amount: Number(f.bonus_amount) || 0,
    random_bonus_pool_amount: Number(f.random_bonus_pool_amount) || 0,
    max_uses: Number(f.max_uses) || 0,
    leaderboard_enabled: f.leaderboard_enabled,
    expires_at: epoch(f.expires_at),
    success_message: f.success_message,
    notes: f.notes
  }
}

function toUpdatePayload(f: FormState): UpdatePromoCodeRequest {
  // Same fields except `scene` is fixed at creation
  const payload: UpdatePromoCodeRequest = toPayload(f)
  delete (payload as Partial<CreatePromoCodeRequest>).scene
  return payload
}

export default function AdminPromoCodesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [scene, setScene] = useState<PromoCodeScene | ''>('')
  const [status, setStatus] = useState<string>('')

  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<PromoCode | null>(null)
  const [form, setForm] = useState<FormState>(empty)
  const [usagesFor, setUsagesFor] = useState<PromoCode | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-promo-codes', page, search, scene, status],
    queryFn: () =>
      adminPromoAPI.listPromoCodes(page, 25, {
        search: search || undefined,
        scene: (scene || undefined) as PromoCodeScene | undefined,
        status: status || undefined
      })
  })

  const usagesQuery = useQuery({
    queryKey: ['admin-promo-usages', usagesFor?.id],
    queryFn: () => (usagesFor ? adminPromoAPI.getPromoUsages(usagesFor.id, 1, 50) : null),
    enabled: !!usagesFor
  })

  const createMut = useMutation({
    mutationFn: (payload: CreatePromoCodeRequest) => adminPromoAPI.createPromoCode(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-promo-codes'] })
      setCreating(false)
      setForm(empty)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const updateMut = useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdatePromoCodeRequest }) =>
      adminPromoAPI.updatePromoCode(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-promo-codes'] })
      setEditing(null)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => adminPromoAPI.deletePromoCode(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-promo-codes'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  function openCreate() {
    setForm(empty)
    setCreating(true)
  }

  function openEdit(p: PromoCode) {
    setForm(fromPromo(p))
    setEditing(p)
  }

  function close() {
    setCreating(false)
    setEditing(null)
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (editing) {
      updateMut.mutate({ id: editing.id, payload: toUpdatePayload(form) })
    } else {
      createMut.mutate(toPayload(form))
    }
  }

  return (
    <>
      <PageHeader
        title="Promo Codes"
        description="Award bonus balance on registration or via a benefit code with optional random pool."
        actions={
          <Button variant="accent" onClick={openCreate}>
            <Plus className="h-3.5 w-3.5" />
            {t('common.create')}
          </Button>
        }
      />

      <Card className="p-4">
        <div className="flex flex-wrap gap-3 items-center mb-4">
          <Input
            name="search"
            placeholder="Code search…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
          <select
            value={scene}
            onChange={(e) => setScene(e.target.value as PromoCodeScene | '')}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">All scenes</option>
            {SCENES.map((s) => (
              <option key={s} value={s} className="capitalize bg-bg-4">{s}</option>
            ))}
          </select>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">All status</option>
            <option value="active" className="bg-bg-4">active</option>
            <option value="disabled" className="bg-bg-4">disabled</option>
          </select>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Code</TH>
                <TH>Scene</TH>
                <TH className="text-right">Fixed bonus</TH>
                <TH className="text-right">Random pool</TH>
                <TH className="text-right">Used / Max</TH>
                <TH>{t('common.status')}</TH>
                <TH>Expires</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((p) => (
                <TR key={p.id}>
                  <TD className="font-mono text-xs text-ink-1">{p.code}</TD>
                  <TD>
                    <Badge tone={p.scene === 'benefit' ? 'accent' : 'neutral'}>{p.scene}</Badge>
                    {p.leaderboard_enabled && (
                      <Badge className="ml-1.5 text-[10px]">leaderboard</Badge>
                    )}
                  </TD>
                  <TD className="text-right font-mono">${p.bonus_amount.toFixed(2)}</TD>
                  <TD className="text-right font-mono">
                    {p.random_bonus_pool_amount > 0 ? (
                      <>
                        ${p.random_bonus_remaining.toFixed(2)}
                        <span className="text-ink-3"> / ${p.random_bonus_pool_amount.toFixed(2)}</span>
                      </>
                    ) : (
                      '—'
                    )}
                  </TD>
                  <TD className="text-right font-mono">
                    {p.used_count}
                    <span className="text-ink-3"> / {p.max_uses === 0 ? '∞' : p.max_uses}</span>
                  </TD>
                  <TD>
                    {p.status === 'active' ? (
                      <Badge tone="success" dot>active</Badge>
                    ) : (
                      <Badge dot>disabled</Badge>
                    )}
                  </TD>
                  <TD className="text-ink-3 text-xs font-mono">
                    {p.expires_at ? new Date(p.expires_at).toLocaleString() : '—'}
                  </TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <button
                        title="Usages"
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => setUsagesFor(p)}
                      >
                        <Users className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title={t('common.edit') as string}
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => openEdit(p)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title={t('common.delete') as string}
                        className="btn btn-ghost btn-icon btn-sm text-signal-err"
                        onClick={() => {
                          if (confirm(`Delete promo code "${p.code}"?`)) {
                            deleteMut.mutate(p.id)
                          }
                        }}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </TD>
                </TR>
              ))}
              {(data?.items ?? []).length === 0 && (
                <TR>
                  <TD colSpan={8} className="text-center text-ink-3 py-8">
                    {t('common.noData')}
                  </TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        {data && data.pages > 1 && (
          <div className="flex items-center justify-between mt-4 pt-3 border-t border-line-1 text-sm">
            <div className="text-ink-3">
              {t('common.total')}: {data.total}
            </div>
            <div className="flex gap-2">
              <Button variant="ghost" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                {t('common.back')}
              </Button>
              <Button variant="ghost" size="sm" disabled={page >= data.pages} onClick={() => setPage((p) => p + 1)}>
                {t('common.next')}
              </Button>
            </div>
          </div>
        )}
      </Card>

      <Modal
        open={creating || !!editing}
        onClose={close}
        title={editing ? `Edit ${editing.code}` : 'Create promo code'}
        size="lg"
        footer={
          <>
            <Button variant="ghost" onClick={close}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="accent"
              type="submit"
              form="promo-form"
              loading={createMut.isPending || updateMut.isPending}
            >
              {t('common.save')}
            </Button>
          </>
        }
      >
        <form id="promo-form" onSubmit={onSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <Input
              name="code"
              label="Code"
              placeholder="leave blank to auto-generate"
              value={form.code}
              onChange={(e) => setForm({ ...form, code: e.target.value })}
              disabled={!!editing}
              className="font-mono"
            />
            <div>
              <label className="input-label">Scene</label>
              <select
                value={form.scene}
                onChange={(e) => setForm({ ...form, scene: e.target.value as PromoCodeScene })}
                className={selectClass}
                disabled={!!editing}
              >
                {SCENES.map((s) => (
                  <option key={s} value={s} className="capitalize bg-bg-4">{s}</option>
                ))}
              </select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Input
              name="bonus_amount"
              type="number"
              step="0.01"
              min="0"
              label="Fixed bonus ($)"
              value={form.bonus_amount}
              onChange={(e) => setForm({ ...form, bonus_amount: e.target.value })}
            />
            <Input
              name="random_bonus_pool_amount"
              type="number"
              step="0.01"
              min="0"
              label="Random pool ($)"
              hint="Used as a draw pool; remainder shrinks each redemption."
              value={form.random_bonus_pool_amount}
              onChange={(e) => setForm({ ...form, random_bonus_pool_amount: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Input
              name="max_uses"
              type="number"
              min="0"
              label="Max uses"
              hint="0 = unlimited"
              value={form.max_uses}
              onChange={(e) => setForm({ ...form, max_uses: e.target.value })}
            />
            <Input
              name="expires_at"
              type="datetime-local"
              label="Expires at"
              value={form.expires_at}
              onChange={(e) => setForm({ ...form, expires_at: e.target.value })}
            />
          </div>
          <label className="inline-flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={form.leaderboard_enabled}
              onChange={(e) =>
                setForm({ ...form, leaderboard_enabled: e.target.checked })
              }
              className="w-4 h-4 accent-orange"
            />
            <span className="text-sm text-ink-2">Enable benefit leaderboard</span>
          </label>
          <Input
            name="success_message"
            label="Success message"
            placeholder="Shown after a user redeems"
            value={form.success_message}
            onChange={(e) => setForm({ ...form, success_message: e.target.value })}
          />
          <Input
            name="notes"
            label="Notes (admin-only)"
            value={form.notes}
            onChange={(e) => setForm({ ...form, notes: e.target.value })}
          />
        </form>
      </Modal>

      <Modal
        open={!!usagesFor}
        onClose={() => setUsagesFor(null)}
        title={usagesFor ? `Usages — ${usagesFor.code}` : ''}
        size="lg"
        footer={
          <Button variant="ghost" onClick={() => setUsagesFor(null)}>
            {t('common.close')}
          </Button>
        }
      >
        {usagesQuery.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-8" />
            ))}
          </div>
        ) : !usagesQuery.data || usagesQuery.data.items.length === 0 ? (
          <div className="text-center text-ink-3 py-8">{t('common.noData')}</div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>User</TH>
                <TH className="text-right">Fixed</TH>
                <TH className="text-right">Random</TH>
                <TH className="text-right">Total</TH>
                <TH>At</TH>
              </TR>
            </THead>
            <TBody>
              {usagesQuery.data.items.map((u) => (
                <TR key={u.id}>
                  <TD className="text-ink-2 text-xs">
                    {u.user?.email || `#${u.user_id}`}
                  </TD>
                  <TD className="text-right font-mono">${u.fixed_bonus_amount.toFixed(2)}</TD>
                  <TD className="text-right font-mono">${u.random_bonus_amount.toFixed(2)}</TD>
                  <TD className="text-right font-mono text-ink-1">
                    ${u.bonus_amount.toFixed(2)}
                  </TD>
                  <TD className="text-ink-3 text-xs font-mono">
                    {new Date(u.used_at).toLocaleString()}
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Modal>
    </>
  )
}
