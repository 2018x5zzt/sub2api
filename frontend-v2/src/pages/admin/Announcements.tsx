import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminAnnouncementsAPI } from '@/api/admin/announcements'
import { toast } from '@/components/ui/Toast'
import type {
  Announcement,
  AnnouncementNotifyMode,
  AnnouncementStatus,
  CreateAnnouncementRequest
} from '@/types'

const STATUSES: AnnouncementStatus[] = ['draft', 'active', 'archived']
const MODES: AnnouncementNotifyMode[] = ['silent', 'popup']

interface FormState {
  title: string
  content: string
  status: AnnouncementStatus
  notify_mode: AnnouncementNotifyMode
  starts_at: string // datetime-local
  ends_at: string
}

const empty: FormState = {
  title: '',
  content: '',
  status: 'draft',
  notify_mode: 'silent',
  starts_at: '',
  ends_at: ''
}

function fromAnnouncement(a: Announcement): FormState {
  const toLocal = (iso?: string) => (iso ? iso.slice(0, 16) : '')
  return {
    title: a.title,
    content: a.content,
    status: a.status,
    notify_mode: a.notify_mode,
    starts_at: toLocal(a.starts_at),
    ends_at: toLocal(a.ends_at)
  }
}

function toPayload(f: FormState): CreateAnnouncementRequest {
  const epoch = (s: string) => {
    if (!s) return undefined
    const t = new Date(s).getTime()
    return Number.isFinite(t) ? Math.floor(t / 1000) : undefined
  }
  return {
    title: f.title.trim(),
    content: f.content,
    status: f.status,
    notify_mode: f.notify_mode,
    targeting: {}, // all users; targeting builder deferred to Phase 3
    starts_at: epoch(f.starts_at),
    ends_at: epoch(f.ends_at)
  }
}

function statusTone(s: AnnouncementStatus) {
  switch (s) {
    case 'active':
      return 'success' as const
    case 'archived':
      return 'neutral' as const
    case 'draft':
      return 'warning' as const
    default:
      return 'neutral' as const
  }
}

const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

export default function AdminAnnouncementsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState<AnnouncementStatus | ''>('')

  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<Announcement | null>(null)
  const [form, setForm] = useState<FormState>(empty)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-announcements', page, search, statusFilter],
    queryFn: () =>
      adminAnnouncementsAPI.listAnnouncements(page, 20, {
        search: search || undefined,
        status: (statusFilter || undefined) as AnnouncementStatus | undefined
      })
  })

  const createMut = useMutation({
    mutationFn: (payload: CreateAnnouncementRequest) => adminAnnouncementsAPI.createAnnouncement(payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-announcements'] })
      setCreating(false)
      setForm(empty)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const updateMut = useMutation({
    mutationFn: ({ id, payload }: { id: number; payload: CreateAnnouncementRequest }) =>
      adminAnnouncementsAPI.updateAnnouncement(id, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-announcements'] })
      setEditing(null)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => adminAnnouncementsAPI.deleteAnnouncement(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-announcements'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  function openCreate() {
    setForm(empty)
    setCreating(true)
  }

  function openEdit(a: Announcement) {
    setForm(fromAnnouncement(a))
    setEditing(a)
  }

  function close() {
    setCreating(false)
    setEditing(null)
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!form.title.trim()) {
      toast.warning(t('v2Common.fieldRequired', { field: t('v2Common.title') }) as string)
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
    <form onSubmit={onSubmit} id="announcement-form" className="space-y-4">
      <Input
        name="title"
        label={t('v2Common.title') as string}
        value={form.title}
        onChange={(e) => setForm({ ...form, title: e.target.value })}
        autoFocus
        required
      />
      <div>
        <label className="input-label">{t('v2Common.content')}</label>
        <textarea
          rows={8}
          value={form.content}
          onChange={(e) => setForm({ ...form, content: e.target.value })}
          className="input font-mono text-xs whitespace-pre-wrap"
          style={{ height: 'auto', minHeight: 160 }}
          placeholder={t('admin.announcements.form.content') as string}
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div>
          <label className="input-label">{t('common.status')}</label>
          <select
            value={form.status}
            onChange={(e) => setForm({ ...form, status: e.target.value as AnnouncementStatus })}
            className={selectClass}
          >
            {STATUSES.map((s) => (
              <option key={s} value={s} className="capitalize bg-bg-4">
                {s}
              </option>
            ))}
          </select>
        </div>
        <div>
          <label className="input-label">{t('admin.announcements.form.notifyMode')}</label>
          <select
            value={form.notify_mode}
            onChange={(e) =>
              setForm({ ...form, notify_mode: e.target.value as AnnouncementNotifyMode })
            }
            className={selectClass}
          >
            {MODES.map((m) => (
              <option key={m} value={m} className="capitalize bg-bg-4">
                {m}
              </option>
            ))}
          </select>
          <p className="text-xs text-ink-3 mt-1">{t('admin.announcements.form.notifyModeHint')}</p>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Input
          name="starts_at"
          type="datetime-local"
          label={t('admin.announcements.form.startsAt') as string}
          value={form.starts_at}
          onChange={(e) => setForm({ ...form, starts_at: e.target.value })}
        />
        <Input
          name="ends_at"
          type="datetime-local"
          label={t('admin.announcements.form.endsAt') as string}
          value={form.ends_at}
          onChange={(e) => setForm({ ...form, ends_at: e.target.value })}
        />
      </div>
      <p className="text-xs text-ink-3">
        {t('admin.announcements.form.targetingAll')}
      </p>
    </form>
  )

  return (
    <>
      <PageHeader
        title={t('nav.announcements')}
        description={t('v2Admin.announcements.description') as string}
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
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as AnnouncementStatus | '')}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">
              {t('v2Common.allStatus')}
            </option>
            {STATUSES.map((s) => (
              <option key={s} value={s} className="capitalize bg-bg-4">
                {s}
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
                <TH>{t('v2Common.title')}</TH>
                <TH>{t('common.status')}</TH>
                <TH>{t('v2Common.mode')}</TH>
                <TH>{t('v2Common.starts')}</TH>
                <TH>{t('v2Common.ends')}</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((a) => (
                <TR key={a.id}>
                  <TD>
                    <div className="text-ink-1 font-medium">{a.title}</div>
                    <div className="text-xs text-ink-3 truncate max-w-md mt-0.5">
                      {a.content.slice(0, 120)}
                      {a.content.length > 120 ? '…' : ''}
                    </div>
                  </TD>
                  <TD>
                    <Badge tone={statusTone(a.status)} dot={a.status === 'active'}>
                      {a.status}
                    </Badge>
                  </TD>
                  <TD>{a.notify_mode === 'popup' ? <Badge tone="accent">popup</Badge> : <Badge>silent</Badge>}</TD>
                  <TD className="text-ink-3 text-xs font-mono">
                    {a.starts_at ? new Date(a.starts_at).toLocaleString() : '—'}
                  </TD>
                  <TD className="text-ink-3 text-xs font-mono">
                    {a.ends_at ? new Date(a.ends_at).toLocaleString() : '—'}
                  </TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
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
                          if (confirm(t('v2Admin.announcements.deleteConfirm', { title: a.title }) as string)) {
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
                  <TD colSpan={6} className="text-center text-ink-3 py-8">
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
        title={editing ? t('common.edit') : t('common.create')}
        size="lg"
        footer={
          <>
            <Button variant="ghost" onClick={close}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="accent"
              type="submit"
              form="announcement-form"
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
