import { useEffect, useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Save,
  Database,
  CloudCog,
  Download,
  Trash2,
  RotateCcw,
  CalendarClock,
  TestTube
} from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import {
  adminBackupAPI,
  type BackupRecord,
  type BackupS3Config,
  type BackupScheduleConfig
} from '@/api/admin/backup'
import { toast } from '@/components/ui/Toast'

function formatBytes(n: number): string {
  if (!n) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = n
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(v >= 100 ? 0 : 1)} ${units[i]}`
}

function statusTone(s: BackupRecord['status']) {
  switch (s) {
    case 'completed':
      return 'success' as const
    case 'failed':
      return 'danger' as const
    case 'in_progress':
      return 'warning' as const
    default:
      return 'neutral' as const
  }
}

function S3ConfigSection() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['admin-backup-s3'],
    queryFn: () => adminBackupAPI.getS3Config()
  })

  const [draft, setDraft] = useState<BackupS3Config | null>(null)
  const [secret, setSecret] = useState('')

  useEffect(() => {
    if (data) setDraft({ ...data, secret_access_key: undefined })
  }, [data])

  const saveMut = useMutation({
    mutationFn: () => {
      if (!draft) throw new Error('not loaded')
      return adminBackupAPI.updateS3Config({
        ...draft,
        secret_access_key: secret || undefined
      })
    },
    onSuccess: (fresh) => {
      qc.setQueryData(['admin-backup-s3'], fresh)
      setSecret('')
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const testMut = useMutation({
    mutationFn: () => {
      if (!draft) throw new Error('not loaded')
      return adminBackupAPI.testS3Connection({
        ...draft,
        secret_access_key: secret || undefined
      })
    },
    onSuccess: (r) => {
      if (r.success) toast.success(r.message)
      else toast.error(r.message + (r.details ? ` — ${r.details}` : ''))
    },
    onError: (e: { message?: string }) => toast.error(e?.message || 'S3 test failed')
  })

  if (isLoading || !draft) {
    return (
      <Card className="p-6 mb-4">
        <Skeleton className="h-4 w-32 mb-4" />
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-10 mb-2" />
        ))}
      </Card>
    )
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    saveMut.mutate()
  }

  return (
    <Card className="mb-4">
      <div className="px-6 py-4 border-b border-line-1 flex items-center gap-2">
        <CloudCog className="h-4 w-4 text-orange" />
        <h2 className="text-base font-medium text-ink-1">S3 storage</h2>
      </div>
      <form onSubmit={onSubmit} className="p-6 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <Input
            name="endpoint"
            label="Endpoint"
            placeholder="https://s3.us-west-2.amazonaws.com"
            value={draft.endpoint}
            onChange={(e) => setDraft({ ...draft, endpoint: e.target.value })}
            className="font-mono text-xs"
          />
          <Input
            name="region"
            label="Region"
            placeholder="us-west-2"
            value={draft.region}
            onChange={(e) => setDraft({ ...draft, region: e.target.value })}
            className="font-mono text-xs"
          />
          <Input
            name="bucket"
            label="Bucket"
            value={draft.bucket}
            onChange={(e) => setDraft({ ...draft, bucket: e.target.value })}
            className="font-mono text-xs"
          />
          <Input
            name="prefix"
            label="Prefix"
            placeholder="sub2api/backups/"
            value={draft.prefix}
            onChange={(e) => setDraft({ ...draft, prefix: e.target.value })}
            className="font-mono text-xs"
          />
          <Input
            name="access_key_id"
            label="Access key ID"
            value={draft.access_key_id}
            onChange={(e) => setDraft({ ...draft, access_key_id: e.target.value })}
            className="font-mono text-xs"
          />
          <div>
            <label className="input-label">Secret access key</label>
            <Input
              name="secret_access_key"
              type="password"
              placeholder={
                draft.secret_access_key_configured
                  ? 'Leave blank to keep existing'
                  : 'Enter secret'
              }
              value={secret}
              onChange={(e) => setSecret(e.target.value)}
              className="font-mono text-xs"
            />
            <p className="text-xs mt-1 text-ink-3">
              {draft.secret_access_key_configured ? (
                <span className="text-signal-ok">configured</span>
              ) : (
                <span>not set</span>
              )}
            </p>
          </div>
        </div>
        <label className="inline-flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={draft.use_path_style}
            onChange={(e) => setDraft({ ...draft, use_path_style: e.target.checked })}
            className="w-4 h-4 accent-orange"
          />
          <span className="text-sm text-ink-2">Use path-style addressing</span>
        </label>
        <div className="flex items-center justify-end gap-2 pt-2 border-t border-line-1">
          <Button
            type="button"
            variant="ghost"
            onClick={() => testMut.mutate()}
            loading={testMut.isPending}
          >
            <TestTube className="h-3.5 w-3.5" />
            Test connection
          </Button>
          <Button type="submit" variant="accent" loading={saveMut.isPending}>
            <Save className="h-3.5 w-3.5" />
            {t('common.save')}
          </Button>
        </div>
      </form>
    </Card>
  )
}

function ScheduleSection() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['admin-backup-schedule'],
    queryFn: () => adminBackupAPI.getSchedule()
  })

  const [draft, setDraft] = useState<BackupScheduleConfig | null>(null)

  useEffect(() => {
    if (data) setDraft(data)
  }, [data])

  const saveMut = useMutation({
    mutationFn: () => {
      if (!draft) throw new Error('not loaded')
      return adminBackupAPI.updateSchedule(draft)
    },
    onSuccess: (fresh) => {
      qc.setQueryData(['admin-backup-schedule'], fresh)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  if (isLoading || !draft) {
    return (
      <Card className="p-6 mb-4">
        <Skeleton className="h-4 w-32 mb-4" />
        {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-10 mb-2" />)}
      </Card>
    )
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    saveMut.mutate()
  }

  return (
    <Card className="mb-4">
      <div className="px-6 py-4 border-b border-line-1 flex items-center gap-2">
        <CalendarClock className="h-4 w-4 text-orange" />
        <h2 className="text-base font-medium text-ink-1">Schedule</h2>
      </div>
      <form onSubmit={onSubmit} className="p-6 space-y-4">
        <label className="inline-flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={draft.enabled}
            onChange={(e) => setDraft({ ...draft, enabled: e.target.checked })}
            className="w-4 h-4 accent-orange"
          />
          <span className="text-sm text-ink-2">Enable scheduled backups</span>
        </label>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <Input
            name="cron_expression"
            label="Cron expression"
            placeholder="0 3 * * *"
            hint="Standard 5-field cron"
            value={draft.cron_expression}
            onChange={(e) => setDraft({ ...draft, cron_expression: e.target.value })}
            className="font-mono"
            disabled={!draft.enabled}
          />
          <Input
            name="retention_days"
            type="number"
            min="0"
            label="Retention (days)"
            hint="0 = keep forever"
            value={String(draft.retention_days)}
            onChange={(e) =>
              setDraft({ ...draft, retention_days: Number(e.target.value) || 0 })
            }
          />
        </div>
        <div className="flex justify-end pt-2 border-t border-line-1">
          <Button type="submit" variant="accent" loading={saveMut.isPending}>
            <Save className="h-3.5 w-3.5" />
            {t('common.save')}
          </Button>
        </div>
      </form>
    </Card>
  )
}

function BackupsList() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['admin-backups'],
    queryFn: () => adminBackupAPI.listBackups(),
    refetchInterval: (q) => {
      const items = (q.state.data as { items: BackupRecord[] } | undefined)?.items ?? []
      return items.some((b) => b.status === 'in_progress') ? 4000 : false
    }
  })

  const [createOpen, setCreateOpen] = useState(false)
  const [createPassword, setCreatePassword] = useState('')
  const [createDescription, setCreateDescription] = useState('')

  const [restoring, setRestoring] = useState<BackupRecord | null>(null)
  const [restorePassword, setRestorePassword] = useState('')

  const createMut = useMutation({
    mutationFn: () =>
      adminBackupAPI.createBackup({
        password: createPassword || undefined,
        description: createDescription || undefined
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-backups'] })
      setCreateOpen(false)
      setCreatePassword('')
      setCreateDescription('')
      toast.success('Backup started')
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const deleteMut = useMutation({
    mutationFn: (id: string) => adminBackupAPI.deleteBackup(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-backups'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const restoreMut = useMutation({
    mutationFn: () => {
      if (!restoring) throw new Error('No backup selected')
      return adminBackupAPI.restoreBackup(restoring.id, restorePassword)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-backups'] })
      setRestoring(null)
      setRestorePassword('')
      toast.success('Restore started')
    },
    onError: (e: { message?: string }) =>
      toast.error(e?.message || 'Restore failed (wrong password?)')
  })

  async function download(id: string) {
    try {
      const { url } = await adminBackupAPI.getBackupDownloadURL(id)
      window.open(url, '_blank', 'noopener')
    } catch (e) {
      toast.error((e as { message?: string })?.message || 'Could not get download URL')
    }
  }

  return (
    <>
      <Card>
        <div className="px-6 py-4 border-b border-line-1 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Database className="h-4 w-4 text-orange" />
            <h2 className="text-base font-medium text-ink-1">Backups</h2>
          </div>
          <Button variant="accent" size="sm" onClick={() => setCreateOpen(true)}>
            Create now
          </Button>
        </div>

        {isLoading ? (
          <div className="p-4 space-y-2">
            {Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Created</TH>
                <TH>Status</TH>
                <TH>Storage</TH>
                <TH className="text-right">Size</TH>
                <TH>Encrypted</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((b) => (
                <TR key={b.id}>
                  <TD className="text-xs font-mono text-ink-2">
                    {new Date(b.created_at).toLocaleString()}
                    <div className="text-ink-4 mt-0.5 truncate max-w-[260px]">{b.id}</div>
                    {b.error_message && (
                      <div className="text-signal-err mt-1 text-[11px] truncate max-w-md">
                        {b.error_message}
                      </div>
                    )}
                  </TD>
                  <TD>
                    <Badge tone={statusTone(b.status)} dot={b.status === 'in_progress'}>
                      {b.status}
                    </Badge>
                  </TD>
                  <TD>
                    <Badge>{b.storage}</Badge>
                  </TD>
                  <TD className="text-right font-mono">{formatBytes(b.size_bytes)}</TD>
                  <TD>
                    {b.encrypted ? (
                      <Badge tone="accent">encrypted</Badge>
                    ) : (
                      <Badge>plaintext</Badge>
                    )}
                  </TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <button
                        title="Download"
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => download(b.id)}
                        disabled={b.status !== 'completed'}
                      >
                        <Download className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title="Restore"
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => {
                          setRestorePassword('')
                          setRestoring(b)
                        }}
                        disabled={b.status !== 'completed'}
                      >
                        <RotateCcw className="h-3.5 w-3.5" />
                      </button>
                      <button
                        title={t('common.delete') as string}
                        className="btn btn-ghost btn-icon btn-sm text-signal-err"
                        onClick={() => {
                          if (confirm('Delete this backup?')) deleteMut.mutate(b.id)
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
      </Card>

      <Modal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title="Create backup"
        footer={
          <>
            <Button variant="ghost" onClick={() => setCreateOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="accent"
              onClick={() => createMut.mutate()}
              loading={createMut.isPending}
            >
              Create
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input
            name="description"
            label="Description"
            placeholder="optional"
            value={createDescription}
            onChange={(e) => setCreateDescription(e.target.value)}
          />
          <Input
            name="password"
            type="password"
            label="Encryption password"
            hint="Leave blank to skip encryption"
            value={createPassword}
            onChange={(e) => setCreatePassword(e.target.value)}
          />
        </div>
      </Modal>

      <Modal
        open={!!restoring}
        onClose={() => setRestoring(null)}
        title="Restore backup"
        footer={
          <>
            <Button variant="ghost" onClick={() => setRestoring(null)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="danger"
              onClick={() => restoreMut.mutate()}
              loading={restoreMut.isPending}
            >
              Restore
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <div className="rounded-lg border border-signal-err/30 bg-signal-err/5 p-3 text-sm text-signal-err">
            <strong>Destructive operation.</strong> Restoring will overwrite current data with the
            backup contents.
          </div>
          {restoring?.encrypted && (
            <Input
              name="restore_password"
              type="password"
              label="Encryption password"
              hint="Required — same password used to create the backup"
              value={restorePassword}
              onChange={(e) => setRestorePassword(e.target.value)}
              autoFocus
              required
            />
          )}
        </div>
      </Modal>
    </>
  )
}

export default function AdminBackupPage() {
  return (
    <>
      <PageHeader
        title="Backup"
        description="Configure S3, schedule periodic snapshots, or trigger backups manually."
      />
      <S3ConfigSection />
      <ScheduleSection />
      <BackupsList />
    </>
  )
}
