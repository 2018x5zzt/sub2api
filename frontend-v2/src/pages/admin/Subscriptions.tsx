import { useState, type FormEvent } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, CalendarPlus, Ban } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { adminSubscriptionsAPI } from '@/api/admin/subscriptions'
import { toast } from '@/components/ui/Toast'
import type { UserSubscription } from '@/types'

const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

function statusTone(s: UserSubscription['status']) {
  switch (s) {
    case 'active':
      return 'success' as const
    case 'expired':
      return 'warning' as const
    case 'revoked':
      return 'danger' as const
    default:
      return 'neutral' as const
  }
}

function expiryLabel(expiresAt: string | null): { label: string; tone: string } {
  if (!expiresAt) return { label: '∞', tone: 'text-ink-3' }
  const expires = new Date(expiresAt)
  const days = Math.ceil((expires.getTime() - Date.now()) / 86_400_000)
  const dateStr = expires.toLocaleDateString()
  if (days < 0) return { label: `Expired · ${dateStr}`, tone: 'text-signal-err' }
  if (days <= 3) return { label: `${days}d · ${dateStr}`, tone: 'text-signal-err' }
  if (days <= 7) return { label: `${days}d · ${dateStr}`, tone: 'text-signal-warn' }
  return { label: `${days}d · ${dateStr}`, tone: 'text-ink-2' }
}

interface AssignForm {
  user_id: string
  group_id: string
  validity_days: string
}

const emptyAssign: AssignForm = { user_id: '', group_id: '', validity_days: '' }

export default function AdminSubscriptionsPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<'' | 'active' | 'expired' | 'revoked'>('')
  const [groupIdFilter, setGroupIdFilter] = useState('')

  const [assigning, setAssigning] = useState(false)
  const [assignForm, setAssignForm] = useState<AssignForm>(emptyAssign)

  const [extending, setExtending] = useState<UserSubscription | null>(null)
  const [extendDays, setExtendDays] = useState('')

  const { data, isLoading } = useQuery({
    queryKey: ['admin-subscriptions', page, search, status, groupIdFilter],
    queryFn: () =>
      adminSubscriptionsAPI.listAdminSubscriptions(page, 25, {
        search: search || undefined,
        status: (status || undefined) as 'active' | 'expired' | 'revoked' | undefined,
        group_id: groupIdFilter ? Number(groupIdFilter) : undefined
      })
  })

  const assignMut = useMutation({
    mutationFn: () => {
      const userId = Number(assignForm.user_id)
      const groupId = Number(assignForm.group_id)
      if (!Number.isFinite(userId) || userId <= 0) throw new Error('User ID required')
      if (!Number.isFinite(groupId) || groupId <= 0) throw new Error('Group ID required')
      const days = Number(assignForm.validity_days)
      return adminSubscriptionsAPI.assignSubscription({
        user_id: userId,
        group_id: groupId,
        validity_days: Number.isFinite(days) && days > 0 ? days : undefined
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-subscriptions'] })
      setAssigning(false)
      setAssignForm(emptyAssign)
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const extendMut = useMutation({
    mutationFn: () => {
      if (!extending) throw new Error('No subscription selected')
      const days = Number(extendDays)
      if (!Number.isFinite(days) || days === 0) throw new Error('Days must be non-zero')
      return adminSubscriptionsAPI.extendSubscription(extending.id, { days })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-subscriptions'] })
      setExtending(null)
      setExtendDays('')
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const revokeMut = useMutation({
    mutationFn: (id: number) => adminSubscriptionsAPI.revokeSubscription(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-subscriptions'] })
      toast.success(t('common.success') as string)
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  function ProgressBar({
    used,
    limit,
    label
  }: {
    used: number | undefined
    limit: number | null | undefined
    label: string
  }) {
    if (!limit) return null
    const u = used ?? 0
    const pct = Math.min((u / limit) * 100, 100)
    const tone = pct >= 90 ? 'bg-signal-err' : pct >= 70 ? 'bg-signal-warn' : 'bg-signal-ok'
    return (
      <div className="text-xs">
        <div className="flex justify-between text-ink-3 font-mono">
          <span>{label}</span>
          <span>
            ${u.toFixed(2)}<span className="text-ink-4"> / ${limit.toFixed(0)}</span>
          </span>
        </div>
        <div className="h-1 bg-bg-3 rounded-full overflow-hidden mt-1">
          <div className={`h-full ${tone}`} style={{ width: `${pct}%` }} />
        </div>
      </div>
    )
  }

  function onAssignSubmit(e: FormEvent) {
    e.preventDefault()
    assignMut.mutate()
  }

  function onExtendSubmit(e: FormEvent) {
    e.preventDefault()
    extendMut.mutate()
  }

  return (
    <>
      <PageHeader
        title={t('v2Admin.subscriptions.title')}
        description={t('v2Admin.subscriptions.description') as string}
        actions={
          <Button variant="accent" onClick={() => setAssigning(true)}>
            <Plus className="h-3.5 w-3.5" />
            {t('v2Admin.subscriptions.assign')}
          </Button>
        }
      />

      <Card className="p-4">
        <div className="flex flex-wrap gap-3 items-center mb-4">
          <Input
            name="search"
            placeholder={t('v2Admin.subscriptions.searchPlaceholder') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="max-w-xs"
          />
          <Input
            name="group_id"
            placeholder={t('v2Common.groupId') as string}
            value={groupIdFilter}
            onChange={(e) => setGroupIdFilter(e.target.value)}
            className="max-w-[120px]"
          />
          <select
            value={status}
            onChange={(e) => setStatus(e.target.value as '' | 'active' | 'expired' | 'revoked')}
            className={selectClass + ' max-w-[160px]'}
          >
            <option value="" className="bg-bg-4">{t('v2Common.allStatus')}</option>
            <option value="active" className="bg-bg-4">{t('admin.subscriptions.status.active')}</option>
            <option value="expired" className="bg-bg-4">{t('admin.subscriptions.status.expired')}</option>
            <option value="revoked" className="bg-bg-4">{t('admin.subscriptions.status.revoked')}</option>
          </select>
        </div>

        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="h-12" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>{t('v2Common.user')}</TH>
                <TH>{t('v2Common.group')}</TH>
                <TH>{t('common.status')}</TH>
                <TH>{t('v2Common.usageProgress')}</TH>
                <TH>{t('v2Common.expiry')}</TH>
                <TH className="text-right">{t('common.actions')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((s) => {
                const exp = expiryLabel(s.expires_at)
                return (
                  <TR key={s.id}>
                    <TD className="text-xs text-ink-2">
                      {s.user?.email || `#${s.user_id}`}
                    </TD>
                    <TD>
                      <div className="text-sm text-ink-1">
                        {s.group?.name || `#${s.group_id}`}
                      </div>
                      {s.group?.platform && (
                        <Badge tone="accent" className="mt-1 text-[10px]">
                          {s.group.platform}
                        </Badge>
                      )}
                    </TD>
                    <TD>
                      <Badge tone={statusTone(s.status)} dot={s.status === 'active'}>
                        {t(`admin.subscriptions.status.${s.status}`)}
                      </Badge>
                    </TD>
                    <TD className="min-w-[200px]">
                      <div className="space-y-1.5">
                        <ProgressBar
                          used={s.daily_usage_usd}
                          limit={s.group?.daily_limit_usd}
                          label={t('v2Common.day') as string}
                        />
                        <ProgressBar
                          used={s.weekly_usage_usd}
                          limit={s.group?.weekly_limit_usd}
                          label={t('v2Common.week') as string}
                        />
                        <ProgressBar
                          used={s.monthly_usage_usd}
                          limit={s.group?.monthly_limit_usd}
                          label={t('v2Common.month') as string}
                        />
                        {!s.group?.daily_limit_usd &&
                          !s.group?.weekly_limit_usd &&
                          !s.group?.monthly_limit_usd && (
                            <span className="text-xs text-ink-4 font-mono">{t('v2Common.noLimits')}</span>
                          )}
                      </div>
                    </TD>
                    <TD className={`text-xs font-mono ${exp.tone}`}>{exp.label}</TD>
                    <TD className="text-right">
                      <div className="inline-flex gap-1">
                        <button
                          title={t('v2Admin.subscriptions.extend') as string}
                          className="btn btn-ghost btn-icon btn-sm"
                          onClick={() => {
                            setExtendDays('')
                            setExtending(s)
                          }}
                          disabled={s.status === 'revoked'}
                        >
                          <CalendarPlus className="h-3.5 w-3.5" />
                        </button>
                        <button
                          title={t('v2Admin.subscriptions.revoke') as string}
                          className="btn btn-ghost btn-icon btn-sm text-signal-err"
                          onClick={() => {
                            if (
                              confirm(
                                t('v2Admin.subscriptions.revokeConfirm', { id: s.id, group: s.group?.name || s.group_id }) as string
                              )
                            ) {
                              revokeMut.mutate(s.id)
                            }
                          }}
                          disabled={s.status === 'revoked'}
                        >
                          <Ban className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    </TD>
                  </TR>
                )
              })}
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
        open={assigning}
        onClose={() => setAssigning(false)}
        title={t('v2Admin.subscriptions.assignTitle')}
        footer={
          <>
            <Button variant="ghost" onClick={() => setAssigning(false)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="accent"
              type="submit"
              form="assign-form"
              loading={assignMut.isPending}
            >
              {t('v2Admin.subscriptions.assign')}
            </Button>
          </>
        }
      >
        <form id="assign-form" onSubmit={onAssignSubmit} className="space-y-4">
          <Input
            name="user_id"
            type="number"
            min="1"
            label={t('v2Common.userId') as string}
            value={assignForm.user_id}
            onChange={(e) => setAssignForm({ ...assignForm, user_id: e.target.value })}
            required
            autoFocus
          />
          <Input
            name="group_id"
            type="number"
            min="1"
            label={t('v2Common.groupId') as string}
            value={assignForm.group_id}
            onChange={(e) => setAssignForm({ ...assignForm, group_id: e.target.value })}
            required
          />
          <Input
            name="validity_days"
            type="number"
            min="0"
            label={t('admin.subscriptions.form.validityDays') as string}
            hint={t('v2Common.unlimitedValidityHint') as string}
            value={assignForm.validity_days}
            onChange={(e) => setAssignForm({ ...assignForm, validity_days: e.target.value })}
          />
        </form>
      </Modal>

      <Modal
        open={!!extending}
        onClose={() => setExtending(null)}
        title={extending ? t('v2Admin.subscriptions.extendTitle', { id: extending.id }) : ''}
        footer={
          <>
            <Button variant="ghost" onClick={() => setExtending(null)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="accent"
              type="submit"
              form="extend-form"
              loading={extendMut.isPending}
            >
              {t('v2Admin.subscriptions.extend')}
            </Button>
          </>
        }
      >
        <form id="extend-form" onSubmit={onExtendSubmit} className="space-y-4">
          {extending && (
            <div className="rounded-lg border border-line-2 bg-bg-2 p-3 text-sm text-ink-2">
              <div>{extending.user?.email || `User #${extending.user_id}`}</div>
              <div className="text-xs text-ink-3 font-mono mt-1">
                {extending.group?.name || `Group #${extending.group_id}`} · {t('v2Common.currentExpiry')}:{' '}
                {extending.expires_at ? new Date(extending.expires_at).toLocaleDateString() : t('v2Common.noExpirationSymbol')}
              </div>
            </div>
          )}
          <Input
            name="days"
            type="number"
            label={t('admin.subscriptions.form.adjustDays') as string}
            hint={t('v2Common.negativeShortensHint') as string}
            value={extendDays}
            onChange={(e) => setExtendDays(e.target.value)}
            required
            autoFocus
          />
        </form>
      </Modal>
    </>
  )
}
