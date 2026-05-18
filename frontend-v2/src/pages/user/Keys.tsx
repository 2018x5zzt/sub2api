import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Copy, Trash2, Pencil } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { Skeleton } from '@/components/ui/Skeleton'
import { keysAPI } from '@/api/keys'
import { modelsAPI } from '@/api/models'
import { toast } from '@/components/ui/Toast'
import i18n from '@/i18n'
import type { ApiKey, CreateApiKeyRequest, Group, UpdateApiKeyRequest } from '@/types'

interface KeyFormState {
  name: string
  groupId: number | null
  budgetMultiplier: string
  status: 'active' | 'inactive'
  useCustomKey: boolean
  customKey: string
  enableIpRestriction: boolean
  ipWhitelist: string
  ipBlacklist: string
  quota: string
  enableRateLimit: boolean
  rateLimit5h: string
  rateLimit1d: string
  rateLimit7d: string
  enableExpiration: boolean
  expirationDate: string
}

function newFormState(): KeyFormState {
  return {
    name: '',
    groupId: null,
    budgetMultiplier: '',
    status: 'active',
    useCustomKey: false,
    customKey: '',
    enableIpRestriction: false,
    ipWhitelist: '',
    ipBlacklist: '',
    quota: '',
    enableRateLimit: false,
    rateLimit5h: '',
    rateLimit1d: '',
    rateLimit7d: '',
    enableExpiration: false,
    expirationDate: ''
  }
}

function maskKey(k: string) {
  if (!k) return ''
  if (k.length <= 12) return k
  return `${k.slice(0, 8)}...${k.slice(-4)}`
}

function copy(text: string) {
  navigator.clipboard.writeText(text).then(
    () => toast.success(i18n.t('common.copiedToClipboard')),
    () => toast.error(i18n.t('common.copyFailed'))
  )
}

function statusTone(s: ApiKey['status']) {
  switch (s) {
    case 'active': return 'success' as const
    case 'inactive': return 'neutral' as const
    case 'quota_exhausted': return 'warning' as const
    case 'expired': return 'danger' as const
    default: return 'neutral' as const
  }
}

function findGroup(groups: Group[], groupId: number | null) {
  if (groupId == null) return null
  return groups.find((group) => group.id === groupId) || null
}

function defaultBudgetMultiplier(group: Group | null) {
  return group?.default_budget_multiplier ?? 8
}

function isDynamicGroup(group: Group | null) {
  return group?.pricing_mode === 'dynamic'
}

function parseBudgetMultiplier(raw: string): number | null {
  const value = Number(raw)
  if (!Number.isFinite(value)) return null
  if (value < 3 || value > 50) return null
  return Number(value.toFixed(2))
}

function parseNonNegativeNumber(raw: string): number | null {
  const value = Number(raw)
  if (!Number.isFinite(value)) return null
  if (value < 0) return null
  return Number(value.toFixed(4))
}

function parseLimitOrZero(raw: string): number | null {
  if (!raw.trim()) return 0
  return parseNonNegativeNumber(raw)
}

function parseIpList(raw: string): string[] {
  return raw
    .split('\n')
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

function toLocalDateTimeValue(iso: string | null): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function toCreateExpiresInDays(localDateTime: string): number | null {
  const target = new Date(localDateTime)
  if (Number.isNaN(target.getTime())) return null
  const now = new Date()
  const diffDays = Math.ceil((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  return diffDays > 0 ? diffDays : 1
}

function toGroupText(group: Group | undefined, t: (key: string) => string) {
  if (!group) return t('keys.noGroup')
  return group.name
}

function selectClass() {
  return 'input appearance-none cursor-pointer bg-bg-4'
}

export default function KeysPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const [formOpen, setFormOpen] = useState(false)
  const [editingKey, setEditingKey] = useState<ApiKey | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [form, setForm] = useState<KeyFormState>(newFormState())

  const { data, isLoading } = useQuery({
    queryKey: ['user-keys', page, search],
    queryFn: () => keysAPI.listKeys(page, 20, search ? { search } : undefined)
  })

  const { data: groups = [] } = useQuery({
    queryKey: ['user-visible-groups-for-keys'],
    queryFn: () => modelsAPI.getUserGroups()
  })

  const groupById = useMemo(() => {
    const map = new Map<number, Group>()
    for (const group of groups) {
      map.set(group.id, group)
    }
    return map
  }, [groups])

  const selectedFormGroup = useMemo(() => findGroup(groups, form.groupId), [groups, form.groupId])

  const deleteMut = useMutation({
    mutationFn: (id: number) => keysAPI.deleteKey(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['user-keys'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  function closeForm() {
    setFormOpen(false)
    setEditingKey(null)
    setSubmitting(false)
    setForm(newFormState())
  }

  function openCreateForm() {
    setEditingKey(null)
    setForm(newFormState())
    setFormOpen(true)
  }

  function openEditForm(key: ApiKey) {
    const keyGroup = key.group_id != null ? groupById.get(key.group_id) ?? null : null
    const nextBudget = isDynamicGroup(keyGroup)
      ? String(key.budget_multiplier ?? defaultBudgetMultiplier(keyGroup ?? null))
      : ''

    setEditingKey(key)
    setForm({
      name: key.name,
      groupId: key.group_id,
      budgetMultiplier: nextBudget,
      status: key.status === 'active' ? 'active' : 'inactive',
      useCustomKey: false,
      customKey: '',
      enableIpRestriction: key.ip_whitelist.length > 0 || key.ip_blacklist.length > 0,
      ipWhitelist: key.ip_whitelist.join('\n'),
      ipBlacklist: key.ip_blacklist.join('\n'),
      quota: key.quota > 0 ? String(key.quota) : '',
      enableRateLimit: key.rate_limit_5h > 0 || key.rate_limit_1d > 0 || key.rate_limit_7d > 0,
      rateLimit5h: key.rate_limit_5h > 0 ? String(key.rate_limit_5h) : '',
      rateLimit1d: key.rate_limit_1d > 0 ? String(key.rate_limit_1d) : '',
      rateLimit7d: key.rate_limit_7d > 0 ? String(key.rate_limit_7d) : '',
      enableExpiration: key.expires_at != null,
      expirationDate: toLocalDateTimeValue(key.expires_at)
    })
    setFormOpen(true)
  }

  function handleGroupChange(value: string) {
    const groupId = value ? Number(value) : null
    const nextGroup = findGroup(groups, groupId)
    setForm((prev) => {
      if (isDynamicGroup(nextGroup)) {
        return {
          ...prev,
          groupId,
          budgetMultiplier: String(defaultBudgetMultiplier(nextGroup))
        }
      }
      return { ...prev, groupId, budgetMultiplier: '' }
    })
  }

  function validateCustomKey(raw: string): string | null {
    if (!raw.trim()) return t('keys.customKeyRequired') as string
    if (raw.length < 16) return t('keys.customKeyTooShort') as string
    if (!/^[a-zA-Z0-9_-]+$/.test(raw)) return t('keys.customKeyInvalidChars') as string
    return null
  }

  async function handleResetQuota() {
    if (!editingKey) return
    const ok = confirm(
      t('keys.resetQuotaConfirmMessage', {
        name: editingKey.name,
        used: (editingKey.quota_used ?? 0).toFixed(4)
      }) as string
    )
    if (!ok) return

    try {
      await keysAPI.updateKey(editingKey.id, { reset_quota: true })
      toast.success(t('keys.quotaResetSuccess') as string)
      setEditingKey((prev) => (prev ? { ...prev, quota_used: 0 } : prev))
      await qc.invalidateQueries({ queryKey: ['user-keys'] })
    } catch (e) {
      toast.error((e as { message?: string })?.message || (t('keys.failedToResetQuota') as string))
    }
  }

  async function handleResetRateLimitUsage() {
    if (!editingKey) return
    const ok = confirm(
      t('keys.resetRateLimitConfirmMessage', {
        name: editingKey.name
      }) as string
    )
    if (!ok) return

    try {
      await keysAPI.updateKey(editingKey.id, { reset_rate_limit_usage: true })
      toast.success(t('keys.rateLimitResetSuccess') as string)
      setEditingKey((prev) => {
        if (!prev) return prev
        return {
          ...prev,
          usage_5h: 0,
          usage_1d: 0,
          usage_7d: 0,
          window_5h_start: null,
          window_1d_start: null,
          window_7d_start: null,
          reset_5h_at: null,
          reset_1d_at: null,
          reset_7d_at: null
        }
      })
      await qc.invalidateQueries({ queryKey: ['user-keys'] })
    } catch (e) {
      toast.error((e as { message?: string })?.message || (t('keys.failedToResetRateLimit') as string))
    }
  }

  async function onSubmit() {
    const name = form.name.trim()
    if (!name) {
      toast.warning('Name is required')
      return
    }

    if (form.groupId == null) {
      toast.warning(t('keys.groupRequired') as string)
      return
    }

    const dynamic = isDynamicGroup(selectedFormGroup)
    const budget = parseBudgetMultiplier(form.budgetMultiplier)
    if (dynamic && budget == null) {
      toast.warning(t('keys.budgetMultiplierRequired') as string)
      return
    }

    if (!editingKey && form.useCustomKey) {
      const customKeyError = validateCustomKey(form.customKey.trim())
      if (customKeyError) {
        toast.warning(customKeyError)
        return
      }
    }

    const ipWhitelist = form.enableIpRestriction ? parseIpList(form.ipWhitelist) : []
    const ipBlacklist = form.enableIpRestriction ? parseIpList(form.ipBlacklist) : []

    const quota = parseLimitOrZero(form.quota)
    if (quota == null) {
      toast.warning(t('keys.quotaAmountHint') as string)
      return
    }

    const rateLimit5h = form.enableRateLimit ? parseLimitOrZero(form.rateLimit5h) : 0
    const rateLimit1d = form.enableRateLimit ? parseLimitOrZero(form.rateLimit1d) : 0
    const rateLimit7d = form.enableRateLimit ? parseLimitOrZero(form.rateLimit7d) : 0
    if (rateLimit5h == null || rateLimit1d == null || rateLimit7d == null) {
      toast.warning(t('keys.rateLimitHint') as string)
      return
    }

    let editExpiresAt = ''
    let createExpiresInDays: number | undefined
    if (form.enableExpiration) {
      if (!form.expirationDate) {
        toast.warning(t('keys.expirationDate') as string)
        return
      }
      if (editingKey) {
        const expiresAt = new Date(form.expirationDate)
        if (Number.isNaN(expiresAt.getTime())) {
          toast.warning(t('keys.expirationDate') as string)
          return
        }
        editExpiresAt = expiresAt.toISOString()
      } else {
        const expiresInDays = toCreateExpiresInDays(form.expirationDate)
        if (expiresInDays == null) {
          toast.warning(t('keys.expirationDate') as string)
          return
        }
        createExpiresInDays = expiresInDays
      }
    }

    const payloadBase = {
      name,
      group_id: form.groupId,
      ip_whitelist: ipWhitelist,
      ip_blacklist: ipBlacklist,
      quota,
      rate_limit_5h: rateLimit5h,
      rate_limit_1d: rateLimit1d,
      rate_limit_7d: rateLimit7d
    }

    setSubmitting(true)
    try {
      if (editingKey) {
        const payload: UpdateApiKeyRequest = {
          ...payloadBase,
          status: form.status,
          expires_at: editExpiresAt
        }
        if (dynamic) payload.budget_multiplier = budget
        await keysAPI.updateKey(editingKey.id, payload)
        toast.success(t('keys.keyUpdatedSuccess') as string)
      } else {
        const payload: CreateApiKeyRequest = {
          ...payloadBase
        }
        if (dynamic) payload.budget_multiplier = budget
        if (form.useCustomKey) payload.custom_key = form.customKey.trim()
        if (createExpiresInDays != null) payload.expires_in_days = createExpiresInDays
        const created = await keysAPI.createKey(payload)
        toast.success(t('keys.keyCreatedSuccess') as string)
        setTimeout(() => copy(created.key), 200)
      }

      await qc.invalidateQueries({ queryKey: ['user-keys'] })
      closeForm()
    } catch (e) {
      toast.error((e as { message?: string })?.message || (t('common.error') as string))
      setSubmitting(false)
    }
  }

  return (
    <>
      <PageHeader
        title={t('keys.title')}
        description={t('keys.description') as string}
        actions={
          <Button onClick={openCreateForm}>
            <Plus className="h-4 w-4" />
            {t('keys.createKey')}
          </Button>
        }
      />

      <Card className="p-4">
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input
            name="search"
            placeholder={t('keys.searchPlaceholder') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>{t('keys.nameLabel')}</TH>
                <TH>{t('keys.apiKey')}</TH>
                <TH>{t('keys.group')}</TH>
                <TH>{t('keys.statusLabel')}</TH>
                <TH>{t('keys.created')}</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((k) => (
                <TR key={k.id}>
                  <TD className="font-medium text-ink-1">{k.name}</TD>
                  <TD>
                    <button
                      onClick={() => copy(k.key)}
                      className="inline-flex items-center gap-1.5 font-mono text-xs text-ink-2 hover:text-orange"
                      title={k.key}
                    >
                      {maskKey(k.key)}
                      <Copy className="h-3 w-3 opacity-60" />
                    </button>
                  </TD>
                  <TD className="text-sm text-ink-2">{toGroupText(k.group || (k.group_id != null ? groupById.get(k.group_id) : undefined), t)}</TD>
                  <TD>
                    <Badge tone={statusTone(k.status)}>{k.status}</Badge>
                  </TD>
                  <TD className="text-xs text-ink-2">
                    {new Date(k.created_at).toLocaleDateString()}
                  </TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <button
                        title={t('keys.editKey') as string}
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => openEditForm(k)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title={t('keys.deleteKey') as string}
                        className="btn btn-ghost btn-icon btn-sm text-signal-err"
                        onClick={() => {
                          if (confirm(t('keys.deleteConfirmMessage', { name: k.name }) as string)) {
                            deleteMut.mutate(k.id)
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
                  <TD colSpan={6} className="py-8 text-center text-ink-3">
                    {t('common.noData')}
                  </TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        {data && data.pages > 1 && (
          <div className="mt-4 flex items-center justify-between border-t border-line-2 pt-3 text-sm">
            <div className="text-ink-3">
              {t('common.total')}: {data.total}
            </div>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                {t('common.back')}
              </Button>
              <Button variant="secondary" size="sm" disabled={page >= data.pages} onClick={() => setPage((p) => p + 1)}>
                {t('common.next')}
              </Button>
            </div>
          </div>
        )}
      </Card>

      <Modal
        open={formOpen}
        onClose={closeForm}
        title={editingKey ? t('keys.editKey') : t('keys.createKey')}
        footer={
          <>
            <Button variant="secondary" onClick={closeForm}>
              {t('common.cancel')}
            </Button>
            <Button onClick={onSubmit} loading={submitting}>
              {editingKey ? t('common.update') : t('common.create')}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            name="name"
            autoFocus
            label={t('keys.nameLabel') as string}
            placeholder={t('keys.namePlaceholder') as string}
            value={form.name}
            onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))}
          />

          <div>
            <label htmlFor="key-group" className="input-label">
              {t('keys.groupLabel')}
            </label>
            <select
              id="key-group"
              className={selectClass()}
              value={form.groupId ?? ''}
              onChange={(e) => handleGroupChange(e.target.value)}
            >
              <option value="">{t('keys.selectGroup')}</option>
              {groups.map((group) => (
                <option key={group.id} value={group.id}>
                  {group.name}
                </option>
              ))}
            </select>
          </div>

          {isDynamicGroup(selectedFormGroup) && (
            <Input
              name="budget_multiplier"
              label={t('keys.budgetMultiplierLabel') as string}
              type="number"
              step="0.1"
              min={3}
              max={50}
              value={form.budgetMultiplier}
              onChange={(e) => setForm((prev) => ({ ...prev, budgetMultiplier: e.target.value }))}
              hint={t('keys.budgetMultiplierHint', {
                min: 3,
                max: 50,
                default: defaultBudgetMultiplier(selectedFormGroup)
              }) as string}
            />
          )}

          {!editingKey && (
            <div className="rounded-xl border border-line-2 bg-bg-2 p-3">
              <div className="flex items-center justify-between">
                <label htmlFor="use-custom-key" className="input-label mb-0">
                  {t('keys.customKeyLabel')}
                </label>
                <input
                  id="use-custom-key"
                  type="checkbox"
                  checked={form.useCustomKey}
                  onChange={(e) => setForm((prev) => ({ ...prev, useCustomKey: e.target.checked }))}
                />
              </div>
              {form.useCustomKey && (
                <Input
                  name="custom_key"
                  label={t('keys.customKeyLabel') as string}
                  value={form.customKey}
                  onChange={(e) => setForm((prev) => ({ ...prev, customKey: e.target.value }))}
                  hint={t('keys.customKeyHint') as string}
                />
              )}
            </div>
          )}

          {editingKey && (
            <div>
              <label htmlFor="key-status" className="input-label">
                {t('keys.statusLabel')}
              </label>
              <select
                id="key-status"
                className={selectClass()}
                value={form.status}
                onChange={(e) => setForm((prev) => ({ ...prev, status: e.target.value as 'active' | 'inactive' }))}
              >
                <option value="active">{t('common.active')}</option>
                <option value="inactive">{t('common.inactive')}</option>
              </select>
            </div>
          )}

          <div className="rounded-xl border border-line-2 bg-bg-2 p-3 space-y-3">
            <div className="flex items-center justify-between">
              <label htmlFor="enable-ip-restriction" className="input-label mb-0">
                {t('keys.ipRestriction')}
              </label>
              <input
                id="enable-ip-restriction"
                type="checkbox"
                checked={form.enableIpRestriction}
                onChange={(e) => setForm((prev) => ({ ...prev, enableIpRestriction: e.target.checked }))}
              />
            </div>
            {form.enableIpRestriction && (
              <div className="space-y-3">
                <div>
                  <label htmlFor="ip-whitelist" className="input-label">
                    {t('keys.ipWhitelist')}
                  </label>
                  <textarea
                    id="ip-whitelist"
                    className="input min-h-20"
                    value={form.ipWhitelist}
                    onChange={(e) => setForm((prev) => ({ ...prev, ipWhitelist: e.target.value }))}
                    placeholder={t('keys.ipWhitelistPlaceholder') as string}
                  />
                  <p className="mt-1 text-xs text-ink-3">{t('keys.ipWhitelistHint')}</p>
                </div>
                <div>
                  <label htmlFor="ip-blacklist" className="input-label">
                    {t('keys.ipBlacklist')}
                  </label>
                  <textarea
                    id="ip-blacklist"
                    className="input min-h-20"
                    value={form.ipBlacklist}
                    onChange={(e) => setForm((prev) => ({ ...prev, ipBlacklist: e.target.value }))}
                    placeholder={t('keys.ipBlacklistPlaceholder') as string}
                  />
                  <p className="mt-1 text-xs text-ink-3">{t('keys.ipBlacklistHint')}</p>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-2 rounded-xl border border-line-2 bg-bg-2 p-3">
            <Input
              name="quota"
              label={t('keys.quotaAmount') as string}
              type="number"
              min={0}
              step="0.01"
              value={form.quota}
              onChange={(e) => setForm((prev) => ({ ...prev, quota: e.target.value }))}
              placeholder={t('keys.quotaAmountPlaceholder') as string}
              hint={t('keys.quotaAmountHint') as string}
            />

            {editingKey && editingKey.quota > 0 && (
              <div className="flex items-center justify-between rounded-lg border border-line-2 bg-bg-1 px-3 py-2">
                <span className="text-sm text-ink-2">
                  {t('keys.quotaUsed')}: ${(editingKey.quota_used ?? 0).toFixed(4)} / ${editingKey.quota.toFixed(2)}
                </span>
                <Button variant="secondary" size="sm" onClick={handleResetQuota}>
                  {t('keys.reset')}
                </Button>
              </div>
            )}
          </div>

          <div className="rounded-xl border border-line-2 bg-bg-2 p-3 space-y-3">
            <div className="flex items-center justify-between">
              <label htmlFor="enable-rate-limit" className="input-label mb-0">
                {t('keys.rateLimitSection')}
              </label>
              <input
                id="enable-rate-limit"
                type="checkbox"
                checked={form.enableRateLimit}
                onChange={(e) => setForm((prev) => ({ ...prev, enableRateLimit: e.target.checked }))}
              />
            </div>

            {form.enableRateLimit && (
              <div className="space-y-3">
                <Input
                  name="rate_limit_5h"
                  label={t('keys.rateLimit5h') as string}
                  type="number"
                  min={0}
                  step="0.01"
                  value={form.rateLimit5h}
                  onChange={(e) => setForm((prev) => ({ ...prev, rateLimit5h: e.target.value }))}
                />
                <Input
                  name="rate_limit_1d"
                  label={t('keys.rateLimit1d') as string}
                  type="number"
                  min={0}
                  step="0.01"
                  value={form.rateLimit1d}
                  onChange={(e) => setForm((prev) => ({ ...prev, rateLimit1d: e.target.value }))}
                />
                <Input
                  name="rate_limit_7d"
                  label={t('keys.rateLimit7d') as string}
                  type="number"
                  min={0}
                  step="0.01"
                  value={form.rateLimit7d}
                  onChange={(e) => setForm((prev) => ({ ...prev, rateLimit7d: e.target.value }))}
                  hint={t('keys.rateLimitHint') as string}
                />
              </div>
            )}

            {editingKey && (editingKey.rate_limit_5h > 0 || editingKey.rate_limit_1d > 0 || editingKey.rate_limit_7d > 0) && (
              <div className="rounded-lg border border-line-2 bg-bg-1 p-3 text-sm text-ink-2 space-y-1">
                <div>
                  5h: ${(editingKey.usage_5h ?? 0).toFixed(4)} / ${editingKey.rate_limit_5h.toFixed(2)}
                </div>
                <div>
                  1d: ${(editingKey.usage_1d ?? 0).toFixed(4)} / ${editingKey.rate_limit_1d.toFixed(2)}
                </div>
                <div>
                  7d: ${(editingKey.usage_7d ?? 0).toFixed(4)} / ${editingKey.rate_limit_7d.toFixed(2)}
                </div>
                <div className="pt-1">
                  <Button variant="secondary" size="sm" onClick={handleResetRateLimitUsage}>
                    {t('keys.resetRateLimitUsage')}
                  </Button>
                </div>
              </div>
            )}
          </div>

          <div className="rounded-xl border border-line-2 bg-bg-2 p-3 space-y-3">
            <div className="flex items-center justify-between">
              <label htmlFor="enable-expiration" className="input-label mb-0">
                {t('keys.expiration')}
              </label>
              <input
                id="enable-expiration"
                type="checkbox"
                checked={form.enableExpiration}
                onChange={(e) => setForm((prev) => ({ ...prev, enableExpiration: e.target.checked }))}
              />
            </div>

            {form.enableExpiration && (
              <div className="space-y-3">
                <div className="flex flex-wrap gap-2">
                  {[7, 30, 90].map((days) => (
                    <Button
                      key={days}
                      size="sm"
                      variant="secondary"
                      onClick={() => {
                        const target = new Date()
                        target.setDate(target.getDate() + days)
                        setForm((prev) => ({
                          ...prev,
                          expirationDate: toLocalDateTimeValue(target.toISOString())
                        }))
                      }}
                    >
                      {editingKey ? t('keys.extendDays', { days }) : t('keys.expiresInDays', { days })}
                    </Button>
                  ))}
                </div>
                <Input
                  name="expiration_date"
                  label={t('keys.expirationDate') as string}
                  type="datetime-local"
                  value={form.expirationDate}
                  onChange={(e) => setForm((prev) => ({ ...prev, expirationDate: e.target.value }))}
                  hint={t('keys.expirationDateHint') as string}
                />
                {editingKey?.expires_at && (
                  <p className="text-xs text-ink-3">
                    {t('keys.currentExpiration')}: {new Date(editingKey.expires_at).toLocaleString()}
                  </p>
                )}
              </div>
            )}
          </div>
        </div>
      </Modal>
    </>
  )
}
