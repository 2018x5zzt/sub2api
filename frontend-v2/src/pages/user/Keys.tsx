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
  const [form, setForm] = useState<KeyFormState>({ name: '', groupId: null, budgetMultiplier: '' })

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
    setForm({ name: '', groupId: null, budgetMultiplier: '' })
  }

  function openCreateForm() {
    setEditingKey(null)
    setForm({ name: '', groupId: null, budgetMultiplier: '' })
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
      budgetMultiplier: nextBudget
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

  async function onSubmit() {
    if (!form.name.trim()) {
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

    const payloadBase = {
      name: form.name.trim(),
      group_id: form.groupId
    }

    const payloadWithBudget = dynamic
      ? { ...payloadBase, budget_multiplier: budget }
      : payloadBase

    setSubmitting(true)
    try {
      if (editingKey) {
        await keysAPI.updateKey(editingKey.id, payloadWithBudget as UpdateApiKeyRequest)
        toast.success(t('keys.keyUpdatedSuccess') as string)
      } else {
        const created = await keysAPI.createKey(payloadWithBudget as CreateApiKeyRequest)
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
        </div>
      </Modal>
    </>
  )
}
