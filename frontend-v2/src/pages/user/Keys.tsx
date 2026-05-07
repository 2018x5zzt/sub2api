import { useState } from 'react'
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
import { toast } from '@/components/ui/Toast'
import type { ApiKey } from '@/types'

function maskKey(k: string) {
  if (!k) return ''
  if (k.length <= 12) return k
  return `${k.slice(0, 8)}...${k.slice(-4)}`
}

function copy(text: string) {
  navigator.clipboard.writeText(text).then(
    () => toast.success('Copied'),
    () => toast.error('Copy failed')
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

export default function KeysPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [createOpen, setCreateOpen] = useState(false)
  const [createName, setCreateName] = useState('')
  const [creating, setCreating] = useState(false)

  const { data, isLoading } = useQuery({
    queryKey: ['user-keys', page, search],
    queryFn: () => keysAPI.listKeys(page, 20, search ? { search } : undefined)
  })

  const deleteMut = useMutation({
    mutationFn: (id: number) => keysAPI.deleteKey(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['user-keys'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  async function onCreate() {
    if (!createName.trim()) {
      toast.warning('Name is required')
      return
    }
    setCreating(true)
    try {
      const k = await keysAPI.createKey({ name: createName.trim() })
      qc.invalidateQueries({ queryKey: ['user-keys'] })
      toast.success(t('common.success') as string)
      setCreateOpen(false)
      setCreateName('')
      // Show the newly-created key in a follow-up modal would be ideal — for v2 phase 1 we just toast it
      setTimeout(() => copy(k.key), 200)
    } catch (e) {
      toast.error((e as { message?: string })?.message || (t('common.error') as string))
    } finally {
      setCreating(false)
    }
  }

  return (
    <>
      <PageHeader
        title={t('keys.title')}
        description={t('keys.description') as string}
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4" />
            {t('keys.createKey')}
          </Button>
        }
      />

      <Card className="p-4">
        <div className="flex flex-wrap gap-3 items-center mb-4">
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
                      className="font-mono text-xs text-ink-2 hover:text-orange inline-flex items-center gap-1.5"
                      title={k.key}
                    >
                      {maskKey(k.key)}
                      <Copy className="h-3 w-3 opacity-60" />
                    </button>
                  </TD>
                  <TD>
                    <Badge tone={statusTone(k.status)}>{k.status}</Badge>
                  </TD>
                  <TD className="text-ink-2 text-xs">
                    {new Date(k.created_at).toLocaleDateString()}
                  </TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <button
                        title={t('keys.editKey') as string}
                        className="btn btn-ghost btn-icon btn-sm"
                        onClick={() => toast.info('Edit not yet migrated to v2')}
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
                  <TD colSpan={5} className="text-center text-ink-3 py-8">
                    {t('common.noData')}
                  </TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        {data && data.pages > 1 && (
          <div className="flex items-center justify-between mt-4 pt-3 border-t border-line-2 text-sm">
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
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        title={t('keys.createKey')}
        footer={
          <>
            <Button variant="secondary" onClick={() => setCreateOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button onClick={onCreate} loading={creating}>
              {t('common.create')}
            </Button>
          </>
        }
      >
        <Input
          name="name"
          autoFocus
          label={t('keys.nameLabel') as string}
          placeholder={t('keys.namePlaceholder') as string}
          value={createName}
          onChange={(e) => setCreateName(e.target.value)}
        />
      </Modal>
    </>
  )
}
