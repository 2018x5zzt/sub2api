import { useMemo, useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Power,
  AlertCircle,
  RefreshCw,
  Pencil,
  Trash2,
  TestTube,
  KeyRound
} from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminAccountsAPI } from '@/api/admin/accounts'
import { toast } from '@/components/ui/Toast'
import type {
  Account,
  AccountPlatform,
  AccountType,
  CreateAccountRequest,
  UpdateAccountRequest
} from '@/types'

const PLATFORMS: AccountPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'sora']
const TYPES: AccountType[] = ['oauth', 'setup-token', 'apikey', 'upstream', 'bedrock']

const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

function statusBadge(a: Account) {
  if (!a.schedulable) return <Badge tone="warning" dot>paused</Badge>
  if (a.status === 'error') return <Badge tone="danger" dot>error</Badge>
  if (a.status === 'inactive') return <Badge dot>inactive</Badge>
  if (a.rate_limit_reset_at) return <Badge tone="warning" dot>rate-limited</Badge>
  return <Badge tone="success" dot>active</Badge>
}

function relativeTime(iso: string | null): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  return `${d}d ago`
}

interface FormState {
  name: string
  notes: string
  platform: AccountPlatform
  type: AccountType
  credentials_json: string
  proxy_id: string
  concurrency: string
  priority: string
  rate_multiplier: string
  group_ids: string
  expires_at: string
  auto_pause_on_expired: boolean
  status: 'active' | 'inactive' | 'error'
  schedulable: boolean
}

const empty: FormState = {
  name: '',
  notes: '',
  platform: 'anthropic',
  type: 'apikey',
  credentials_json: '{\n  \n}',
  proxy_id: '',
  concurrency: '5',
  priority: '0',
  rate_multiplier: '',
  group_ids: '',
  expires_at: '',
  auto_pause_on_expired: false,
  status: 'active',
  schedulable: true
}

function fromAccount(a: Account): FormState {
  const credsObj = (a.credentials || {}) as Record<string, unknown>
  // Mask anything that looks like a long secret with •••, so admins editing other
  // fields don't accidentally clobber the key. Anything they paste fresh wins.
  const masked: Record<string, unknown> = {}
  for (const [k, v] of Object.entries(credsObj)) {
    if (typeof v === 'string' && v.length > 12) masked[k] = '••• (unchanged)'
    else masked[k] = v
  }
  return {
    name: a.name,
    notes: a.notes ?? '',
    platform: a.platform,
    type: a.type,
    credentials_json: JSON.stringify(masked, null, 2),
    proxy_id: a.proxy_id != null ? String(a.proxy_id) : '',
    concurrency: String(a.concurrency ?? 5),
    priority: String(a.priority ?? 0),
    rate_multiplier: a.rate_multiplier != null ? String(a.rate_multiplier) : '',
    group_ids: (a.group_ids ?? []).join(','),
    expires_at: a.expires_at ? new Date(a.expires_at * 1000).toISOString().slice(0, 16) : '',
    auto_pause_on_expired: a.auto_pause_on_expired,
    status: a.status,
    schedulable: a.schedulable
  }
}

interface ParsedForm {
  payload: CreateAccountRequest | UpdateAccountRequest
  error?: string
}

function parseForm(f: FormState, original?: Account): ParsedForm {
  let creds: Record<string, unknown>
  try {
    creds = JSON.parse(f.credentials_json || '{}')
  } catch (e) {
    return {
      payload: { name: f.name } as CreateAccountRequest,
      error: 'Credentials must be valid JSON: ' + (e as Error).message
    }
  }

  // Strip masked placeholders so we don't overwrite stored secrets with the
  // placeholder string. Caller will merge with original.credentials on update.
  for (const [k, v] of Object.entries(creds)) {
    if (typeof v === 'string' && v.startsWith('••• (unchanged')) {
      delete creds[k]
    }
  }

  // On update, fields the form left as ••• keep their existing values from the
  // server. Merge by overlaying the parsed (potentially-cleared) creds onto the
  // original ones.
  const finalCreds = original
    ? { ...((original.credentials as Record<string, unknown>) || {}), ...creds }
    : creds

  const groupIds = f.group_ids
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
    .map(Number)
    .filter((n) => Number.isFinite(n) && n > 0)

  const proxyId = f.proxy_id.trim() ? Number(f.proxy_id) : null
  const expiresEpoch = f.expires_at
    ? Math.floor(new Date(f.expires_at).getTime() / 1000)
    : null

  const num = (s: string, fallback: number) => {
    const n = Number(s)
    return Number.isFinite(n) ? n : fallback
  }

  const rateMult = f.rate_multiplier.trim() === '' ? undefined : Number(f.rate_multiplier)

  const base: CreateAccountRequest = {
    name: f.name.trim(),
    notes: f.notes.trim() || null,
    platform: f.platform,
    type: f.type,
    credentials: finalCreds,
    proxy_id: proxyId,
    concurrency: num(f.concurrency, 5),
    priority: num(f.priority, 0),
    rate_multiplier: Number.isFinite(rateMult as number) ? (rateMult as number) : undefined,
    group_ids: groupIds.length > 0 ? groupIds : undefined,
    expires_at: expiresEpoch,
    auto_pause_on_expired: f.auto_pause_on_expired
  }

  if (original) {
    const updatePayload: UpdateAccountRequest = {
      ...base,
      status: f.status,
      schedulable: f.schedulable
    }
    return { payload: updatePayload }
  }
  return { payload: base }
}

const CRED_HINTS: Record<string, string> = {
  'anthropic.apikey': '{\n  "api_key": "sk-ant-…"\n}',
  'anthropic.oauth':
    '{\n  "access_token": "…",\n  "refresh_token": "…",\n  "expires_at": "ISO-8601"\n}',
  'anthropic.setup-token': '{\n  "setup_token": "sk-ant-setup-…"\n}',
  'openai.apikey': '{\n  "api_key": "sk-…"\n}',
  'openai.oauth': '{\n  "access_token": "…",\n  "refresh_token": "…"\n}',
  'gemini.apikey': '{\n  "api_key": "AIza…"\n}',
  'gemini.oauth':
    '{\n  "access_token": "…",\n  "refresh_token": "…",\n  "oauth_type": "code_assist|google_one|ai_studio",\n  "tier_id": "google_ai_pro",\n  "project_id": "…"\n}',
  'antigravity.oauth': '{\n  "access_token": "…",\n  "refresh_token": "…"\n}',
  'sora.apikey': '{\n  "session_token": "…"\n}'
}

function credsHint(platform: AccountPlatform, type: AccountType): string | undefined {
  return CRED_HINTS[`${platform}.${type}`]
}

export default function AdminAccountsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [platformFilter, setPlatformFilter] = useState<AccountPlatform | ''>('')
  const [statusFilter, setStatusFilter] = useState<string>('')

  const [form, setForm] = useState<FormState>(empty)
  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<Account | null>(null)
  const [formError, setFormError] = useState<string | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-accounts', page, search, platformFilter, statusFilter],
    queryFn: () =>
      adminAccountsAPI.listAccounts(page, 20, {
        search: search || undefined,
        platform: (platformFilter || undefined) as AccountPlatform | undefined,
        status: statusFilter || undefined
      })
  })

  const createMut = useMutation({
    mutationFn: (payload: CreateAccountRequest) => adminAccountsAPI.createAccount(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      close()
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const updateMut = useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: UpdateAccountRequest }) =>
      adminAccountsAPI.updateAccount(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      close()
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => adminAccountsAPI.deleteAccount(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const toggleSched = useMutation({
    mutationFn: ({ id, schedulable }: { id: number; schedulable: boolean }) =>
      adminAccountsAPI.setAccountSchedulable(id, schedulable),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['admin-accounts'] }),
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const clearError = useMutation({
    mutationFn: (id: number) => adminAccountsAPI.clearAccountError(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const clearRate = useMutation({
    mutationFn: (id: number) => adminAccountsAPI.clearAccountRateLimit(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const testMut = useMutation({
    mutationFn: (id: number) => adminAccountsAPI.testAccount(id),
    onSuccess: (r) => {
      if (r.success) {
        toast.success(
          `OK${r.latency_ms ? ` · ${r.latency_ms}ms` : ''}${
            r.models?.length ? ` · ${r.models.length} models` : ''
          }`
        )
      } else {
        toast.error(r.message || 'Test failed')
      }
    },
    onError: (e: { message?: string }) => toast.error(e?.message || 'Test failed')
  })

  const refreshMut = useMutation({
    mutationFn: (id: number) => adminAccountsAPI.refreshAccountCredentials(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-accounts'] })
      toast.success('Credentials refreshed')
    },
    onError: (e: { message?: string }) => toast.error(e?.message || 'Refresh failed')
  })

  function openCreate() {
    setForm(empty)
    setFormError(null)
    setCreating(true)
  }

  function openEdit(a: Account) {
    setForm(fromAccount(a))
    setFormError(null)
    setEditing(a)
  }

  function close() {
    setCreating(false)
    setEditing(null)
    setFormError(null)
  }

  function onPlatformChange(p: AccountPlatform) {
    setForm((s) => {
      // When the platform changes during create, refresh the credentials hint so
      // the textarea isn't left with the wrong shape.
      if (creating) {
        const hint = credsHint(p, s.type)
        return { ...s, platform: p, credentials_json: hint || '{\n  \n}' }
      }
      return { ...s, platform: p }
    })
  }

  function onTypeChange(typ: AccountType) {
    setForm((s) => {
      if (creating) {
        const hint = credsHint(s.platform, typ)
        return { ...s, type: typ, credentials_json: hint || '{\n  \n}' }
      }
      return { ...s, type: typ }
    })
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    setFormError(null)
    if (!form.name.trim()) {
      setFormError('Name is required')
      return
    }
    const parsed = parseForm(form, editing || undefined)
    if (parsed.error) {
      setFormError(parsed.error)
      return
    }
    if (editing) {
      updateMut.mutate({ id: editing.id, payload: parsed.payload as UpdateAccountRequest })
    } else {
      createMut.mutate(parsed.payload as CreateAccountRequest)
    }
  }

  const credsPlaceholder = useMemo(
    () => credsHint(form.platform, form.type) || '{}',
    [form.platform, form.type]
  )

  return (
    <>
      <PageHeader
        title={t('nav.accounts')}
        description="Upstream provider accounts. Credentials are stored as platform-specific JSON; the form keeps long secrets masked when editing so you can change other fields without re-pasting them."
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
            placeholder={t('common.searchPlaceholder') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
          <select
            value={platformFilter}
            onChange={(e) => setPlatformFilter(e.target.value as AccountPlatform | '')}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">All platforms</option>
            {PLATFORMS.map((p) => (
              <option key={p} value={p} className="capitalize bg-bg-4">{p}</option>
            ))}
          </select>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">All status</option>
            <option value="active" className="bg-bg-4">active</option>
            <option value="inactive" className="bg-bg-4">inactive</option>
            <option value="error" className="bg-bg-4">error</option>
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
                <TH>{t('common.name')}</TH>
                <TH>Platform / Type</TH>
                <TH>{t('common.status')}</TH>
                <TH className="text-right">Concurrency</TH>
                <TH className="text-right">Priority</TH>
                <TH>Last used</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((a) => (
                <TR key={a.id}>
                  <TD>
                    <div className="text-ink-1 font-medium">{a.name}</div>
                    {a.notes && (
                      <div className="text-xs text-ink-3 mt-0.5 truncate max-w-md">{a.notes}</div>
                    )}
                    {a.error_message && (
                      <div
                        className="text-xs text-signal-err mt-0.5 truncate max-w-md"
                        title={a.error_message}
                      >
                        <AlertCircle className="inline h-3 w-3 mr-1" />
                        {a.error_message}
                      </div>
                    )}
                  </TD>
                  <TD>
                    <div className="flex items-center gap-1.5">
                      <Badge tone="accent">{a.platform}</Badge>
                      <span className="text-xs text-ink-3 font-mono">{a.type}</span>
                    </div>
                  </TD>
                  <TD>{statusBadge(a)}</TD>
                  <TD className="text-right font-mono text-sm">
                    {a.current_concurrency ?? 0}
                    <span className="text-ink-3"> / {a.concurrency}</span>
                  </TD>
                  <TD className="text-right font-mono text-sm">{a.priority}</TD>
                  <TD className="text-ink-3 text-xs font-mono">{relativeTime(a.last_used_at)}</TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <button
                        title="Test"
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => testMut.mutate(a.id)}
                        disabled={testMut.isPending}
                      >
                        <TestTube className="h-3.5 w-3.5" />
                      </button>
                      {(a.type === 'oauth' || a.type === 'setup-token') && (
                        <button
                          title="Refresh credentials"
                          className="btn btn-ghost btn-icon btn-sm"
                          onClick={() => refreshMut.mutate(a.id)}
                          disabled={refreshMut.isPending}
                        >
                          <KeyRound className="h-3.5 w-3.5" />
                        </button>
                      )}
                      <button
                        title={a.schedulable ? 'Pause' : 'Resume'}
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => toggleSched.mutate({ id: a.id, schedulable: !a.schedulable })}
                      >
                        <Power className="h-3.5 w-3.5" />
                      </button>
                      {a.error_message && (
                        <button
                          title="Clear error"
                          className="btn btn-ghost btn-icon btn-sm text-signal-warn"
                          onClick={() => clearError.mutate(a.id)}
                        >
                          <AlertCircle className="h-3.5 w-3.5" />
                        </button>
                      )}
                      {a.rate_limit_reset_at && (
                        <button
                          title="Clear rate limit"
                          className="btn btn-ghost btn-icon btn-sm"
                          onClick={() => clearRate.mutate(a.id)}
                        >
                          <RefreshCw className="h-3.5 w-3.5" />
                        </button>
                      )}
                      <button
                        title={t('common.edit') as string}
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => openEdit(a)}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title={t('common.delete') as string}
                        className="btn btn-ghost btn-icon btn-sm text-signal-err"
                        onClick={() => {
                          if (confirm(`Delete account "${a.name}"?`)) {
                            deleteMut.mutate(a.id)
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
            <div className="text-ink-3">{t('common.total')}: {data.total}</div>
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
        title={editing ? `Edit ${editing.name}` : 'Add account'}
        size="lg"
        footer={
          <>
            <Button variant="ghost" onClick={close}>
              {t('common.cancel')}
            </Button>
            <Button
              type="submit"
              form="account-form"
              variant="accent"
              loading={createMut.isPending || updateMut.isPending}
            >
              {t('common.save')}
            </Button>
          </>
        }
      >
        <form id="account-form" onSubmit={onSubmit} className="space-y-4">
          {formError && (
            <div className="rounded-lg border border-signal-err/30 bg-signal-err/5 px-3 py-2 text-sm text-signal-err">
              {formError}
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <Input
              name="name"
              label={t('common.name') as string}
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              autoFocus
              required
            />
            <Input
              name="notes"
              label="Notes"
              hint="Visible only to admins"
              value={form.notes}
              onChange={(e) => setForm({ ...form, notes: e.target.value })}
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="input-label">Platform</label>
              <select
                value={form.platform}
                onChange={(e) => onPlatformChange(e.target.value as AccountPlatform)}
                className={selectClass}
                disabled={!!editing}
              >
                {PLATFORMS.map((p) => (
                  <option key={p} value={p} className="capitalize bg-bg-4">{p}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="input-label">Type</label>
              <select
                value={form.type}
                onChange={(e) => onTypeChange(e.target.value as AccountType)}
                className={selectClass}
              >
                {TYPES.map((typ) => (
                  <option key={typ} value={typ} className="capitalize bg-bg-4">{typ}</option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label className="input-label">Credentials (JSON)</label>
            <textarea
              value={form.credentials_json}
              onChange={(e) => setForm({ ...form, credentials_json: e.target.value })}
              className="input font-mono text-xs whitespace-pre-wrap"
              style={{ height: 'auto', minHeight: 160 }}
              placeholder={credsPlaceholder}
              spellCheck={false}
            />
            <p className="text-xs text-ink-3 mt-1">
              {editing
                ? 'Existing secrets show as ••• (unchanged) — replace them with a fresh value to rotate.'
                : credsHint(form.platform, form.type)
                ? `Suggested shape for ${form.platform}/${form.type} (replace placeholders with real values).`
                : 'Provide platform-specific keys as JSON.'}
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <Input
              name="proxy_id"
              type="number"
              min="0"
              label="Proxy ID"
              hint="Blank = no proxy"
              value={form.proxy_id}
              onChange={(e) => setForm({ ...form, proxy_id: e.target.value })}
            />
            <Input
              name="concurrency"
              type="number"
              min="0"
              label="Concurrency"
              value={form.concurrency}
              onChange={(e) => setForm({ ...form, concurrency: e.target.value })}
            />
            <Input
              name="priority"
              type="number"
              label="Priority"
              hint="Higher = preferred"
              value={form.priority}
              onChange={(e) => setForm({ ...form, priority: e.target.value })}
            />
            <Input
              name="rate_multiplier"
              type="number"
              step="0.01"
              min="0"
              label="Rate ×"
              hint="Blank = use group"
              value={form.rate_multiplier}
              onChange={(e) => setForm({ ...form, rate_multiplier: e.target.value })}
            />
          </div>

          <Input
            name="group_ids"
            label="Group IDs"
            hint="Comma-separated, e.g. 1,2,5"
            value={form.group_ids}
            onChange={(e) => setForm({ ...form, group_ids: e.target.value })}
            className="font-mono"
          />

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">
            <Input
              name="expires_at"
              type="datetime-local"
              label="Expires at"
              hint="Account will reject requests after this point"
              value={form.expires_at}
              onChange={(e) => setForm({ ...form, expires_at: e.target.value })}
            />
            <label className="inline-flex items-center gap-2 cursor-pointer pb-3">
              <input
                type="checkbox"
                checked={form.auto_pause_on_expired}
                onChange={(e) =>
                  setForm({ ...form, auto_pause_on_expired: e.target.checked })
                }
                className="w-4 h-4 accent-orange"
              />
              <span className="text-sm text-ink-2">Auto-pause when expired</span>
            </label>
          </div>

          {editing && (
            <div className="grid grid-cols-2 gap-3 pt-3 border-t border-line-1">
              <div>
                <label className="input-label">Status</label>
                <select
                  value={form.status}
                  onChange={(e) =>
                    setForm({ ...form, status: e.target.value as 'active' | 'inactive' | 'error' })
                  }
                  className={selectClass}
                >
                  <option value="active" className="bg-bg-4">active</option>
                  <option value="inactive" className="bg-bg-4">inactive</option>
                  <option value="error" className="bg-bg-4">error</option>
                </select>
              </div>
              <label className="inline-flex items-center gap-2 cursor-pointer pb-3 self-end">
                <input
                  type="checkbox"
                  checked={form.schedulable}
                  onChange={(e) => setForm({ ...form, schedulable: e.target.checked })}
                  className="w-4 h-4 accent-orange"
                />
                <span className="text-sm text-ink-2">Schedulable (eligible for routing)</span>
              </label>
            </div>
          )}
        </form>
      </Modal>
    </>
  )
}
