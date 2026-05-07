import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { Skeleton } from '@/components/ui/Skeleton'
import { Badge } from '@/components/ui/Badge'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { adminAPI } from '@/api/admin'

export default function AdminUsersPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-users', page, search],
    queryFn: () => adminAPI.listUsers(page, 20, search ? { search } : undefined)
  })

  return (
    <>
      <PageHeader title={t('nav.users')} description="Manage all platform users" />
      <Card className="p-4">
        <div className="flex flex-wrap gap-3 items-center mb-4">
          <Input
            name="search"
            placeholder={t('common.searchPlaceholder') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>ID</TH>
                <TH>{t('common.email')}</TH>
                <TH>{t('common.name')}</TH>
                <TH>Role</TH>
                <TH>{t('common.status')}</TH>
                <TH className="text-right">{t('common.balance')}</TH>
                <TH>{t('keys.created')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((u) => (
                <TR key={u.id}>
                  <TD className="font-mono text-xs text-ink-3">{u.id}</TD>
                  <TD className="font-medium text-ink-1">{u.email}</TD>
                  <TD className="text-ink-2">{u.username}</TD>
                  <TD>
                    {u.role === 'admin' ? <Badge tone="accent">admin</Badge> : <Badge>user</Badge>}
                  </TD>
                  <TD>
                    {u.status === 'active' ? <Badge tone="success">{t('common.active')}</Badge> : <Badge tone="danger">{t('common.disabled')}</Badge>}
                  </TD>
                  <TD className="text-right font-mono">${u.balance.toFixed(4)}</TD>
                  <TD className="text-ink-3 text-xs">
                    {new Date(u.created_at).toLocaleDateString()}
                  </TD>
                </TR>
              ))}
              {(data?.items ?? []).length === 0 && (
                <TR>
                  <TD colSpan={7} className="text-center text-ink-3 py-8">{t('common.noData')}</TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}

        {data && data.pages > 1 && (
          <div className="flex items-center justify-between mt-4 pt-3 border-t border-line-2 text-sm">
            <div className="text-ink-3">{t('common.total')}: {data.total}</div>
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
    </>
  )
}
