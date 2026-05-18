import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { KeyRound } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { Skeleton } from '@/components/ui/Skeleton'
import { toast } from '@/components/ui/Toast'
import { adminAPI } from '@/api/admin'
import { adminKeysAPI } from '@/api/admin/keys'
import { adminGroupsAPI } from '@/api/admin/groups'
import type { AdminUser, ApiKey, Group } from '@/types'

function statusTone(s: ApiKey['status']) {
  switch (s) {
    case 'active': return 'success' as const
    case 'inactive': return 'neutral' as const
    case 'quota_exhausted': return 'warning' as const
    case 'expired': return 'danger' as const
    default: return 'neutral' as const
  }
}

function toGroupText(group: Group | undefined, t: (key: string) => string) {
  if (!group) return t('admin.users.none')
  return group.name
}

function maskKey(value: string) {
  if (!value) return ''
  if (value.length <= 14) return value
  return `${value.slice(0, 10)}...${value.slice(-4)}`
}

function selectClass() {
  return 'input appearance-none cursor-pointer bg-bg-4'
}

function sortByEmail(items: AdminUser[]) {
  return [...items].sort((a, b) => a.email.localeCompare(b.email))
}

interface KeyEditDialogProps {
  open: boolean
  user: AdminUser | null
  onClose: () => void
}

function KeyEditDialog({ open, user, onClose }: KeyEditDialogProps) {
  const { t } = useTranslation()
  const qc = useQueryClient()

  const [updatingId, setUpdatingId] = useState<number | null>(null)

  const keysQuery = useQuery({
    queryKey: ['admin-user-api-keys', user?.id],
    enabled: open && !!user,
    queryFn: () => adminKeysAPI.listUserApiKeys(user!.id)
  })

  const groupsQuery = useQuery({
    queryKey: ['admin-groups-all-for-user-api-keys'],
    enabled: open,
    queryFn: () => adminGroupsAPI.listAllGroups()
  })

  const groupsById = useMemo(() => {
    const map = new Map<number, Group>()
    for (const g of groupsQuery.data ?? []) {
      map.set(g.id, g)
    }
    return map
  }, [groupsQuery.data])

  const updateMut = useMutation({
    mutationFn: ({ id, groupId }: { id: number; groupId: number | null }) =>
      adminKeysAPI.updateApiKeyGroup(id, { group_id: groupId }),
    onSuccess: (result, variables) => {
      qc.setQueryData(['admin-user-api-keys', user?.id], (prev: any) => {
        if (!prev || !Array.isArray(prev.items)) return prev
        const nextItems = prev.items.map((item: ApiKey) =>
          item.id === variables.id ? result.api_key : item
        )
        return { ...prev, items: nextItems }
      })

      if (result.auto_granted_group_access && result.granted_group_name) {
        toast.success(
          t('admin.users.groupChangedWithGrant', { group: result.granted_group_name }) as string
        )
      } else {
        toast.success(t('admin.users.groupChangedSuccess') as string)
      }
    },
    onError: (error: { message?: string }) => {
      toast.error(error?.message || (t('admin.users.groupChangeFailed') as string))
    },
    onSettled: () => {
      setUpdatingId(null)
    }
  })

  const keys = keysQuery.data?.items ?? []
  const loading = keysQuery.isLoading || groupsQuery.isLoading

  function resolveGroup(k: ApiKey) {
    if (k.group) return k.group
    if (k.group_id == null) return undefined
    return groupsById.get(k.group_id)
  }

  function handleGroupChange(keyId: number, value: string) {
    const groupId = value === '' ? null : Number(value)
    setUpdatingId(keyId)
    updateMut.mutate({ id: keyId, groupId })
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={t('admin.users.userApiKeys')}
      footer={<Button variant="secondary" onClick={onClose}>{t('common.close')}</Button>}
    >
      {!user ? null : (
        <div className="space-y-4">
          <div className="rounded-xl border border-line-2 bg-bg-2 p-3">
            <div className="text-sm font-medium text-ink-1">{user.email}</div>
            <div className="text-xs text-ink-3">{user.username || '-'}</div>
          </div>

          {loading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
            </div>
          ) : keys.length === 0 ? (
            <div className="py-8 text-center text-sm text-ink-3">{t('admin.users.noApiKeys')}</div>
          ) : (
            <Table>
              <THead>
                <TR>
                  <TH>{t('keys.nameLabel')}</TH>
                  <TH>{t('keys.apiKey')}</TH>
                  <TH>{t('admin.users.group')}</TH>
                  <TH>{t('keys.statusLabel')}</TH>
                </TR>
              </THead>
              <TBody>
                {keys.map((k) => {
                  const resolved = resolveGroup(k)
                  const disabled = updatingId === k.id || updateMut.isPending
                  return (
                    <TR key={k.id}>
                      <TD className="font-medium text-ink-1">{k.name}</TD>
                      <TD className="font-mono text-xs text-ink-2" title={k.key}>{maskKey(k.key)}</TD>
                      <TD>
                        <label className="sr-only" htmlFor={`group-${k.id}`}>{t('admin.users.group')}</label>
                        <select
                          id={`group-${k.id}`}
                          aria-label={t('admin.users.group') as string}
                          className={selectClass()}
                          value={k.group_id ?? ''}
                          disabled={disabled}
                          onChange={(e) => handleGroupChange(k.id, e.target.value)}
                        >
                          <option value="">{t('admin.users.none')}</option>
                          {(groupsQuery.data ?? []).map((group) => (
                            <option key={group.id} value={group.id}>{group.name}</option>
                          ))}
                        </select>
                        <div className="mt-1 text-xs text-ink-3">{toGroupText(resolved, t)}</div>
                      </TD>
                      <TD>
                        <Badge tone={statusTone(k.status)}>{k.status}</Badge>
                      </TD>
                    </TR>
                  )
                })}
              </TBody>
            </Table>
          )}
        </div>
      )}
    </Modal>
  )
}

export default function AdminKeysPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null)
  const [modalOpen, setModalOpen] = useState(false)

  const usersQuery = useQuery({
    queryKey: ['admin-users-for-api-keys', page, search],
    queryFn: () => adminAPI.listUsers(page, 20, search ? { search } : undefined)
  })

  const users = useMemo(() => sortByEmail(usersQuery.data?.items ?? []), [usersQuery.data])

  function openModal(user: AdminUser) {
    setSelectedUser(user)
    setModalOpen(true)
  }

  return (
    <>
      <PageHeader
        title={t('nav.apiKeys')}
        description={t('admin.users.description') as string}
      />

      <Card className="p-4">
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input
            name="search"
            placeholder={t('admin.users.searchUsers') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
        </div>

        {usersQuery.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>{t('v2Common.id')}</TH>
                <TH>{t('common.email')}</TH>
                <TH>{t('common.name')}</TH>
                <TH>{t('admin.users.columns.status')}</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {users.map((user) => (
                <TR key={user.id}>
                  <TD className="font-mono text-xs text-ink-3">{user.id}</TD>
                  <TD className="font-medium text-ink-1">{user.email}</TD>
                  <TD className="text-ink-2">{user.username || '-'}</TD>
                  <TD>
                    <Badge tone={user.status === 'active' ? 'success' : 'danger'}>
                      {user.status === 'active' ? t('common.active') : t('admin.users.disabled')}
                    </Badge>
                  </TD>
                  <TD className="text-right">
                    <Button size="sm" variant="secondary" onClick={() => openModal(user)}>
                      <KeyRound className="h-3.5 w-3.5" />
                      {t('admin.users.apiKeys')}
                    </Button>
                  </TD>
                </TR>
              ))}

              {users.length === 0 && (
                <TR>
                  <TD colSpan={5} className="py-8 text-center text-ink-3">{t('common.noData')}</TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        {usersQuery.data && usersQuery.data.pages > 1 && (
          <div className="mt-4 flex items-center justify-between border-t border-line-2 pt-3 text-sm">
            <div className="text-ink-3">{t('common.total')}: {usersQuery.data.total}</div>
            <div className="flex gap-2">
              <Button variant="secondary" size="sm" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                {t('common.back')}
              </Button>
              <Button variant="secondary" size="sm" disabled={page >= usersQuery.data.pages} onClick={() => setPage((p) => p + 1)}>
                {t('common.next')}
              </Button>
            </div>
          </div>
        )}
      </Card>

      <KeyEditDialog
        open={modalOpen}
        user={selectedUser}
        onClose={() => {
          setModalOpen(false)
          setSelectedUser(null)
        }}
      />
    </>
  )
}
