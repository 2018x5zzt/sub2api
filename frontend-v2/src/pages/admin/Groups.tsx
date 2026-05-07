import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2, Power } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminGroupsAPI } from '@/api/admin/groups'
import { toast } from '@/components/ui/Toast'
import type { AdminGroup, CreateGroupRequest, GroupPlatform, SubscriptionType } from '@/types'

const PLATFORMS: GroupPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'sora']
const SUB_TYPES: SubscriptionType[] = ['standard', 'subscription']

interface FormState {
  name: string
  description: string
  platform: GroupPlatform
  rate_multiplier: string
  is_exclusive: boolean
  subscription_type: SubscriptionType
  daily_limit_usd: string
  weekly_limit_usd: string
  monthly_limit_usd: string
}

const empty: FormState = {
  name: '',
  description: '',
  platform: 'anthropic',
  rate_multiplier: '1',
  is_exclusive: false,
  subscription_type: 'standard',
  daily_limit_usd: '',
  weekly_limit_usd: '',
  monthly_limit_usd: ''
}

function fromGroup(g: AdminGroup): FormState {
  return {
    name: g.name,
    description: g.description ?? '',
    platform: g.platform,
    rate_multiplier: String(g.rate_multiplier ?? 1),
    is_exclusive: g.is_exclusive,
    subscription_type: g.subscription_type,
    daily_limit_usd: g.daily_limit_usd != null ? String(g.daily_limit_usd) : '',
    weekly_limit_usd: g.weekly_limit_usd != null ? String(g.weekly_limit_usd) : '',
    monthly_limit_usd: g.monthly_limit_usd != null ? String(g.monthly_limit_usd) : ''
  }
}

function toPayload(f: FormState): CreateGroupRequest {
  const num = (s: string) => {
    const n = Number(s)
    return Number.isFinite(n) ? n : null
  }
  return {
    name: f.name.trim(),
    description: f.description.trim() || null,
    platform: f.platform,
    rate_multiplier: Number(f.rate_multiplier) || 1,
    is_exclusive: f.is_exclusive,
    subscription_type: f.subscription_type,
    daily_limit_usd: f.daily_limit_usd === '' ? null : num(f.daily_limit_usd),
    weekly_limit_usd: f.weekly_limit_usd === '' ? null : num(f.weekly_limit_usd),
    monthly_limit_usd: f.monthly_limit_usd === '' ? null : num(f.monthly_limit_usd)
  }
}

function selectClass() {
  return 'input appearance-none cursor-pointer bg-bg-4'
}

export default function AdminGroupsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [platform, setPlatform] = useState<GroupPlatform | ''>('')

  const [editing, setEditing] = useState<AdminGroup | null>(null)
  const [creating, setCreating] = useState(false)
  const [form, setForm] = useState<FormState>(empty)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-groups', page, search, platform],
    queryFn: () =>
      adminGroupsAPI.listGroups(page, 20, {
        search: search || undefined,
        platform: (platform || undefined) as GroupPlatform | undefined
      })
  })

  const createMut = useMutation({
    mutationFn: (payload: CreateGroupRequest) => adminGroupsAPI.createGroup(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-groups'] })
      setCreating(false)
      setForm(empty)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const updateMut = useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: CreateGroupRequest }) =>
      adminGroupsAPI.updateGroup(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-groups'] })
      setEditing(null)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const toggleMut = useMutation({
    mutationFn: ({ id, status }: { id: number; status: 'active' | 'inactive' }) =>
      adminGroupsAPI.toggleGroupStatus(id, status),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-groups'] })
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => adminGroupsAPI.deleteGroup(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-groups'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  function openCreate() {
    setForm(empty)
    setCreating(true)
  }

  function openEdit(g: AdminGroup) {
    setForm(fromGroup(g))
    setEditing(g)
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!form.name.trim()) {
      toast.warning(t('common.name') + ' required')
      return
    }
    const payload = toPayload(form)
    if (editing) {
      updateMut.mutate({ id: editing.id, payload })
    } else {
      createMut.mutate(payload)
    }
  }

  const FormBody = (
    <form onSubmit={onSubmit} id="group-form" className="space-y-4">
      <Input
        name="name"
        label={t('common.name') as string}
        value={form.name}
        onChange={(e) => setForm({ ...form, name: e.target.value })}
        autoFocus
        required
      />
      <Input
        name="description"
        label="Description"
        value={form.description}
        onChange={(e) => setForm({ ...form, description: e.target.value })}
      />
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="input-label">Platform</label>
          <select
            value={form.platform}
            onChange={(e) => setForm({ ...form, platform: e.target.value as GroupPlatform })}
            className={selectClass()}
            disabled={!!editing}
          >
            {PLATFORMS.map((p) => (
              <option key={p} value={p} className="capitalize bg-bg-4">
                {p}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="input-label">Subscription type</label>
          <select
            value={form.subscription_type}
            onChange={(e) =>
              setForm({ ...form, subscription_type: e.target.value as SubscriptionType })
            }
            className={selectClass()}
          >
            {SUB_TYPES.map((s) => (
              <option key={s} value={s} className="capitalize bg-bg-4">
                {s}
              </option>
            ))}
          </select>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Input
          name="rate_multiplier"
          type="number"
          step="0.01"
          min="0"
          label="Rate multiplier"
          value={form.rate_multiplier}
          onChange={(e) => setForm({ ...form, rate_multiplier: e.target.value })}
          required
        />
        <div className="flex items-end">
          <label className="inline-flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              checked={form.is_exclusive}
              onChange={(e) => setForm({ ...form, is_exclusive: e.target.checked })}
              className="w-4 h-4 accent-orange"
            />
            <span className="text-sm text-ink-2">Exclusive</span>
          </label>
        </div>
      </div>
      <div className="grid grid-cols-3 gap-3">
        <Input
          name="daily"
          type="number"
          step="0.01"
          min="0"
          label="Daily $"
          placeholder="—"
          value={form.daily_limit_usd}
          onChange={(e) => setForm({ ...form, daily_limit_usd: e.target.value })}
        />
        <Input
          name="weekly"
          type="number"
          step="0.01"
          min="0"
          label="Weekly $"
          placeholder="—"
          value={form.weekly_limit_usd}
          onChange={(e) => setForm({ ...form, weekly_limit_usd: e.target.value })}
        />
        <Input
          name="monthly"
          type="number"
          step="0.01"
          min="0"
          label="Monthly $"
          placeholder="—"
          value={form.monthly_limit_usd}
          onChange={(e) => setForm({ ...form, monthly_limit_usd: e.target.value })}
        />
      </div>
    </form>
  )

  return (
    <>
      <PageHeader
        title={t('nav.groups')}
        description="Manage user-facing model groups, billing multipliers, and limits"
        actions={
          <Button onClick={openCreate} variant="accent">
            <Plus className="h-3.5 w-3.5" />
            {t('common.create')}
          </Button>
        }
      />

      <Card className="p-4">
        <div className="flex flex-wrap gap-3 items-center mb-4">
          <Input
            name="search"
            placeholder={t('common.searchPlaceholder') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
          <select
            value={platform}
            onChange={(e) => setPlatform(e.target.value as GroupPlatform | '')}
            className={selectClass() + ' max-w-[180px]'}
          >
            <option value="" className="bg-bg-4">
              All platforms
            </option>
            {PLATFORMS.map((p) => (
              <option key={p} value={p} className="capitalize bg-bg-4">
                {p}
              </option>
            ))}
          </select>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-10" />
            ))}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>{t('common.name')}</TH>
                <TH>Platform</TH>
                <TH className="text-right">Rate</TH>
                <TH>Type</TH>
                <TH>{t('common.status')}</TH>
                <TH className="text-right">Limits ($d/w/m)</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((g) => (
                <TR key={g.id}>
                  <TD>
                    <div className="text-ink-1 font-medium">{g.name}</div>
                    {g.description && (
                      <div className="text-xs text-ink-3 truncate max-w-xs">{g.description}</div>
                    )}
                  </TD>
                  <TD>
                    <Badge tone="accent">{g.platform}</Badge>
                  </TD>
                  <TD className="text-right font-mono">×{g.rate_multiplier}</TD>
                  <TD>
                    <span className="text-xs text-ink-2 font-mono">{g.subscription_type}</span>
                    {g.is_exclusive && (
                      <Badge className="ml-1.5 text-[10px]">exclusive</Badge>
                    )}
                  </TD>
                  <TD>
                    {g.status === 'active' ? (
                      <Badge tone="success" dot>
                        {t('common.active')}
                      </Badge>
                    ) : (
                      <Badge dot>{t('common.inactive')}</Badge>
                    )}
                  </TD>
                  <TD className="text-right font-mono text-xs text-ink-2">
                    {g.daily_limit_usd ?? '—'} / {g.weekly_limit_usd ?? '—'} / {g.monthly_limit_usd ?? '—'}
                  </TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <button
                        title="Toggle active"
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() =>
                          toggleMut.mutate({
                            id: g.id,
                            status: g.status === 'active' ? 'inactive' : 'active'
                          })
                        }
                      >
                        <Power className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title={t('common.edit') as string}
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => openEdit(g)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title={t('common.delete') as string}
                        className="btn btn-ghost btn-icon btn-sm text-signal-err"
                        onClick={() => {
                          if (confirm(`Delete group "${g.name}"?`)) {
                            deleteMut.mutate(g.id)
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
                  <TD colSpan={7} className="text-center text-ink-3 py-8">
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
        onClose={() => {
          setCreating(false)
          setEditing(null)
        }}
        title={editing ? t('common.edit') : t('common.create')}
        size="lg"
        footer={
          <>
            <Button
              variant="ghost"
              onClick={() => {
                setCreating(false)
                setEditing(null)
              }}
            >
              {t('common.cancel')}
            </Button>
            <Button
              variant="accent"
              type="submit"
              form="group-form"
              loading={createMut.isPending || updateMut.isPending}
            >
              {t('common.save')}
            </Button>
          </>
        }
      >
        {FormBody}
      </Modal>
    </>
  )
}
