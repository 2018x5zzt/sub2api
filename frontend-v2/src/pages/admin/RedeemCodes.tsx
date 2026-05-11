import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2, Copy, Ban } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminRedeemAPI } from '@/api/admin/redeem'
import { toast } from '@/components/ui/Toast'
import type { GenerateRedeemCodesRequest, RedeemCode, RedeemCodeType } from '@/types'

const TYPES: RedeemCodeType[] = ['balance', 'concurrency', 'subscription', 'invitation']

const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

interface GenForm {
  type: RedeemCodeType
  value: string
  count: string
  group_id: string
  validity_days: string
}

const emptyForm: GenForm = {
  type: 'balance',
  value: '10',
  count: '1',
  group_id: '',
  validity_days: ''
}

function statusTone(s: RedeemCode['status']) {
  switch (s) {
    case 'active':
    case 'unused':
      return 'success' as const
    case 'used':
      return 'neutral' as const
    case 'expired':
      return 'warning' as const
    default:
      return 'neutral' as const
  }
}

export default function AdminRedeemCodesPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [type, setType] = useState<RedeemCodeType | ''>('')
  const [status, setStatus] = useState('')

  const [genOpen, setGenOpen] = useState(false)
  const [form, setForm] = useState<GenForm>(emptyForm)
  const [generated, setGenerated] = useState<RedeemCode[] | null>(null)

  const [selected, setSelected] = useState<Set<number>>(new Set())

  const { data, isLoading } = useQuery({
    queryKey: ['admin-redeem-codes', page, search, type, status],
    queryFn: () =>
      adminRedeemAPI.listRedeemCodes(page, 25, {
        search: search || undefined,
        type: (type || undefined) as RedeemCodeType | undefined,
        status: status || undefined
      })
  })

  const genMut = useMutation({
    mutationFn: (payload: GenerateRedeemCodesRequest) => adminRedeemAPI.generateRedeemCodes(payload),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['admin-redeem-codes'] })
      setGenerated(res.codes)
      toast.success(t('v2Common.generatedCount', { count: res.count }) as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const expireMut = useMutation({
    mutationFn: (id: number) => adminRedeemAPI.expireRedeemCode(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-redeem-codes'] }),
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => adminRedeemAPI.deleteRedeemCode(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-redeem-codes'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const batchDelMut = useMutation({
    mutationFn: (ids: number[]) => adminRedeemAPI.batchDeleteRedeemCodes(ids),
    onSuccess: (r) => {
      qc.invalidateQueries({ queryKey: ['admin-redeem-codes'] })
      setSelected(new Set())
      toast.success(t('v2Common.deletedCount', { count: r.deleted }) as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  function copyCode(code: string) {
    navigator.clipboard.writeText(code).then(
      () => toast.success(t('common.copiedToClipboard') as string),
      () => toast.error(t('common.copyFailed') as string)
    )
  }

  function copyAllGenerated() {
    if (!generated) return
    navigator.clipboard.writeText(generated.map((c) => c.code).join('\n'))
    toast.success(t('v2Common.copiedCount', { count: generated.length }) as string)
  }

  function toggle(id: number) {
    setSelected((s) => {
      const next = new Set(s)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  function toggleAll() {
    if (!data) return
    if (selected.size === data.items.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(data.items.map((c) => c.id)))
    }
  }

  function onGenSubmit(e: FormEvent) {
    e.preventDefault()
    const value = Number(form.value)
    const count = Number(form.count)
    if (!Number.isFinite(value) || value <= 0) {
      toast.warning(t('v2Common.valuePositive') as string)
      return
    }
    if (!Number.isFinite(count) || count < 1 || count > 1000) {
      toast.warning(t('v2Common.countRange') as string)
      return
    }
    const payload: GenerateRedeemCodesRequest = {
      type: form.type,
      value,
      count
    }
    if (form.type === 'subscription') {
      const gid = Number(form.group_id)
      if (!Number.isFinite(gid) || gid <= 0) {
        toast.warning(t('v2Common.subscriptionGroupRequired') as string)
        return
      }
      payload.group_id = gid
      const days = Number(form.validity_days)
      if (Number.isFinite(days) && days > 0) payload.validity_days = days
    }
    genMut.mutate(payload)
  }

  function closeGen() {
    setGenOpen(false)
    setGenerated(null)
    setForm(emptyForm)
  }

  return (
    <>
      <PageHeader
        title={t('v2Admin.redeemCodes.title')}
        description={t('v2Admin.redeemCodes.description') as string}
        actions={
          <>
            {selected.size > 0 && (
              <Button
                variant="ghost"
                onClick={() => {
                  if (confirm(t('v2Common.deleteCountConfirm', { count: selected.size }) as string)) {
                    batchDelMut.mutate(Array.from(selected))
                  }
                }}
                loading={batchDelMut.isPending}
                className="text-signal-err"
              >
                <Trash2 className="h-3.5 w-3.5" />
                {t('v2Common.deleteSelected', { count: selected.size })}
              </Button>
            )}
            <Button variant="accent" onClick={() => setGenOpen(true)}>
              <Plus className="h-3.5 w-3.5" />
              {t('admin.redeem.generate')}
            </Button>
          </>
        }
      />

      <Card className="p-4">
        <div className="flex flex-wrap gap-3 items-center mb-4">
          <Input
            name="search"
            placeholder={t('v2Common.codeSearch') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
          <select
            value={type}
            onChange={(e) => setType(e.target.value as RedeemCodeType | '')}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">
              {t('v2Common.allTypes')}
            </option>
            {TYPES.map((tp) => (
              <option key={tp} value={tp} className="capitalize bg-bg-4">
                {tp}
              </option>
            ))}
          </select>
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value)}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">
              {t('v2Common.allStatus')}
            </option>
            <option value="unused" className="bg-bg-4">{t('admin.redeem.status.unused')}</option>
            <option value="used" className="bg-bg-4">{t('admin.redeem.status.used')}</option>
            <option value="expired" className="bg-bg-4">{t('admin.redeem.status.expired')}</option>
          </select>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-10" />
            ))}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH style={{ width: 32 }}>
                  <input
                    type="checkbox"
                    className="accent-orange"
                    checked={
                      !!data && data.items.length > 0 && selected.size === data.items.length
                    }
                    onChange={toggleAll}
                  />
                </TH>
                <TH>{t('admin.redeem.columns.code')}</TH>
                <TH>{t('admin.redeem.columns.type')}</TH>
                <TH className="text-right">{t('admin.redeem.columns.value')}</TH>
                <TH>{t('common.status')}</TH>
                <TH>{t('admin.redeem.columns.usedBy')}</TH>
                <TH>{t('admin.redeem.columns.usedAt')}</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((c) => (
                <TR key={c.id}>
                  <TD>
                    <input
                      type="checkbox"
                      className="accent-orange"
                      checked={selected.has(c.id)}
                      onChange={() => toggle(c.id)}
                    />
                  </TD>
                  <TD>
                    <button
                      onClick={() => copyCode(c.code)}
                      className="font-mono text-xs text-ink-1 hover:text-orange inline-flex items-center gap-1.5"
                    >
                      {c.code}
                      <Copy className="h-3 w-3 opacity-60" />
                    </button>
                  </TD>
                  <TD>
                    <Badge tone={c.type === 'subscription' ? 'accent' : 'neutral'}>{c.type}</Badge>
                  </TD>
                  <TD className="text-right font-mono">
                    {c.type === 'subscription' ? (
                      <span>
                        {c.group?.name || `#${c.group_id}`}
                        {c.validity_days ? (
                          <span className="text-ink-3"> · {c.validity_days}d</span>
                        ) : null}
                      </span>
                    ) : c.type === 'balance' ? (
                      `$${c.value.toFixed(2)}`
                    ) : (
                      c.value
                    )}
                  </TD>
                  <TD>
                    <Badge tone={statusTone(c.status)} dot={c.status === 'unused' || c.status === 'active'}>
                      {c.status}
                    </Badge>
                  </TD>
                  <TD className="text-ink-3 text-xs">
                    {c.user?.email || (c.used_by ? `#${c.used_by}` : '—')}
                  </TD>
                  <TD className="text-ink-3 text-xs font-mono">
                    {c.used_at ? new Date(c.used_at).toLocaleString() : '—'}
                  </TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      {(c.status === 'unused' || c.status === 'active') && (
                        <button
                          title={t('admin.redeem.status.expired') as string}
                          className="btn btn-ghost btn-icon btn-sm"
                          onClick={() => expireMut.mutate(c.id)}
                        >
                          <Ban className="h-3.5 w-3.5" />
                        </button>
                      )}
                      <button
                        title={t('common.delete') as string}
                        className="btn btn-ghost btn-icon btn-sm text-signal-err"
                        onClick={() => {
                          if (confirm(t('v2Admin.redeemCodes.deleteConfirm') as string)) deleteMut.mutate(c.id)
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
        open={genOpen}
        onClose={closeGen}
        title={generated ? t('v2Admin.redeemCodes.generatedCount', { count: generated.length }) : t('v2Admin.redeemCodes.generateTitle')}
        size="lg"
        footer={
          generated ? (
            <>
              <Button variant="ghost" onClick={closeGen}>
                {t('common.close')}
              </Button>
              <Button variant="accent" onClick={copyAllGenerated}>
                <Copy className="h-3.5 w-3.5" />
                {t('v2Admin.redeemCodes.copyAll')}
              </Button>
            </>
          ) : (
            <>
              <Button variant="ghost" onClick={closeGen}>
                {t('common.cancel')}
              </Button>
              <Button
                variant="accent"
                type="submit"
                form="redeem-gen-form"
                loading={genMut.isPending}
              >
                {t('admin.redeem.generate')}
              </Button>
            </>
          )
        }
      >
        {generated ? (
          <div>
            <div className="bg-bg-3 border border-line-2 rounded-md p-3 max-h-80 overflow-y-auto">
              <pre className="font-mono text-xs text-ink-1 whitespace-pre-wrap break-all">
                {generated.map((c) => c.code).join('\n')}
              </pre>
            </div>
            <p className="mt-3 text-xs text-ink-3">
              {t('admin.redeem.codesCreated', { count: generated.length })}
            </p>
          </div>
        ) : (
          <form onSubmit={onGenSubmit} id="redeem-gen-form" className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="input-label">{t('admin.redeem.codeType')}</label>
                <select
                  value={form.type}
                  onChange={(e) => setForm({ ...form, type: e.target.value as RedeemCodeType })}
                  className={selectClass}
                >
                  {TYPES.map((tp) => (
                    <option key={tp} value={tp} className="capitalize bg-bg-4">
                      {tp}
                    </option>
                  ))}
                </select>
              </div>
              <Input
                name="count"
                type="number"
                min="1"
                max="1000"
                label={t('admin.redeem.count') as string}
                value={form.count}
                onChange={(e) => setForm({ ...form, count: e.target.value })}
                required
              />
            </div>
            <Input
              name="value"
              type="number"
              step="0.01"
              min="0"
              label={
                form.type === 'balance'
                  ? t('admin.redeem.amount')
                  : form.type === 'concurrency'
                  ? t('admin.redeem.concurrency')
                  : form.type === 'subscription'
                  ? t('admin.redeem.subscription')
                  : t('admin.redeem.value')
              }
              value={form.value}
              onChange={(e) => setForm({ ...form, value: e.target.value })}
              required
            />
            {form.type === 'subscription' && (
              <div className="grid grid-cols-2 gap-3">
                <Input
                  name="group_id"
                  type="number"
                  min="1"
                  label={t('v2Common.groupId') as string}
                  hint={t('admin.redeem.selectGroup') as string}
                  value={form.group_id}
                  onChange={(e) => setForm({ ...form, group_id: e.target.value })}
                  required
                />
                <Input
                  name="validity_days"
                  type="number"
                  min="0"
                  label={t('admin.redeem.validityDays') as string}
                  hint={t('v2Common.unlimitedValidityHint') as string}
                  value={form.validity_days}
                  onChange={(e) => setForm({ ...form, validity_days: e.target.value })}
                />
              </div>
            )}
          </form>
        )}
      </Modal>
    </>
  )
}
