import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Copy, DollarSign, HandCoins, Users } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { affiliateAPI } from '@/api/affiliate'
import { useAuthStore } from '@/stores/auth'

interface Invitee {
  user_id: number
  email?: string
  username?: string
  total_rebate?: number
  created_at?: string
}

interface AffiliateDetail {
  aff_code?: string
  aff_count?: number
  aff_quota?: number
  aff_frozen_quota?: number
  aff_history_quota?: number
  effective_rebate_rate_percent?: number
  invitees?: Invitee[]
}

function money(v: unknown) {
  const n = Number(v || 0)
  return `$${n.toFixed(4)}`
}

function copyText(text: string, copied: string, failed: string) {
  navigator.clipboard.writeText(text).then(
    () => toast.success(copied),
    () => toast.error(failed)
  )
}

export default function AffiliatePage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const refreshUser = useAuthStore((s) => s.refreshUser)

  const { data, isLoading } = useQuery({
    queryKey: ['affiliate-detail'],
    queryFn: () => affiliateAPI.getAffiliateSummary() as Promise<AffiliateDetail>
  })

  const inviteLink = useMemo(() => {
    const code = data?.aff_code || ''
    if (!code) return ''
    const path = `/register?aff=${encodeURIComponent(code)}`
    return typeof window === 'undefined' ? path : `${window.location.origin}${path}`
  }, [data?.aff_code])

  const transferMut = useMutation({
    mutationFn: () => affiliateAPI.transferAffiliateQuota({}),
    onSuccess: async (res: any) => {
      toast.success(t('userAffiliate.transferred', { amount: money(res?.transferred_quota) }) as string)
      await Promise.all([
        qc.invalidateQueries({ queryKey: ['affiliate-detail'] }),
        refreshUser().catch(() => null)
      ])
    },
    onError: (e: { message?: string }) => toast.error(e?.message || (t('common.error') as string))
  })

  const rate = Number(data?.effective_rebate_rate_percent || 0)
  const invitees = data?.invitees ?? []

  return (
    <>
      <PageHeader
        title={t('nav.affiliate')}
        description={t('userAffiliate.description') as string}
      />

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i} className="p-5 space-y-3">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-8 w-28" />
            </Card>
          ))}
        </div>
      ) : (
        <div className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card className="p-5">
              <div className="text-sm text-ink-3 flex items-center gap-2">
                <HandCoins className="h-4 w-4 text-orange" />
                {t('userAffiliate.rebateRate')}
              </div>
              <div className="mt-2 font-display text-3xl text-orange">{rate.toFixed(2).replace(/\.00$/, '')}%</div>
            </Card>
            <Card className="p-5">
              <div className="text-sm text-ink-3 flex items-center gap-2">
                <Users className="h-4 w-4 text-ink-3" />
                {t('userAffiliate.invitedUsers')}
              </div>
              <div className="mt-2 font-display text-3xl text-ink-1">{Number(data?.aff_count || 0).toLocaleString()}</div>
            </Card>
            <Card className="p-5">
              <div className="text-sm text-ink-3 flex items-center gap-2">
                <DollarSign className="h-4 w-4 text-signal-ok" />
                {t('userAffiliate.availableQuota')}
              </div>
              <div className="mt-2 font-display text-3xl text-signal-ok">{money(data?.aff_quota)}</div>
            </Card>
            <Card className="p-5">
              <div className="text-sm text-ink-3">{t('userAffiliate.totalRebateQuota')}</div>
              <div className="mt-2 font-display text-3xl text-ink-1">{money(data?.aff_history_quota)}</div>
              {Number(data?.aff_frozen_quota || 0) > 0 && (
                <div className="mt-1 text-xs text-signal-warn">{t('userAffiliate.frozen', { amount: money(data?.aff_frozen_quota) })}</div>
              )}
            </Card>
          </div>

          <Card className="p-5 space-y-4">
            <div>
              <h2 className="text-base font-medium text-ink-1">{t('userAffiliate.inviteLink')}</h2>
              <p className="text-sm text-ink-3 mt-1">{t('userAffiliate.inviteLinkDescription')}</p>
            </div>
            <div className="grid gap-3 md:grid-cols-2">
              <div className="rounded-lg border border-line-2 bg-bg-2 p-3">
                <div className="text-xs text-ink-3 mb-2">{t('userAffiliate.code')}</div>
                <div className="flex items-center gap-2">
                  <code className="flex-1 truncate font-mono text-sm text-ink-1">{data?.aff_code || '-'}</code>
                  <Button size="sm" variant="ghost" onClick={() => data?.aff_code && copyText(data.aff_code, t('common.copied') as string, t('userAffiliate.copyFailed') as string)}>
                    <Copy className="h-3.5 w-3.5" />
                    {t('common.copy')}
                  </Button>
                </div>
              </div>
              <div className="rounded-lg border border-line-2 bg-bg-2 p-3">
                <div className="text-xs text-ink-3 mb-2">{t('userAffiliate.link')}</div>
                <div className="flex items-center gap-2">
                  <code className="flex-1 truncate font-mono text-sm text-ink-1">{inviteLink || '-'}</code>
                  <Button size="sm" variant="ghost" onClick={() => inviteLink && copyText(inviteLink, t('common.copied') as string, t('userAffiliate.copyFailed') as string)}>
                    <Copy className="h-3.5 w-3.5" />
                    {t('common.copy')}
                  </Button>
                </div>
              </div>
            </div>
          </Card>

          <Card className="p-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-base font-medium text-ink-1">{t('userAffiliate.transferQuota')}</h2>
              <p className="text-sm text-ink-3 mt-1">{t('userAffiliate.transferDescription')}</p>
            </div>
            <Button
              variant="accent"
              loading={transferMut.isPending}
              disabled={Number(data?.aff_quota || 0) <= 0}
              onClick={() => transferMut.mutate()}
            >
              <HandCoins className="h-4 w-4" />
              {t('userAffiliate.transfer')}
            </Button>
          </Card>

          <Card className="p-4">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-base font-medium text-ink-1">{t('userAffiliate.invitees')}</h2>
              <Badge>{invitees.length}</Badge>
            </div>
            <Table>
              <THead>
                <TR>
                  <TH>{t('userAffiliate.email')}</TH>
                  <TH>{t('userAffiliate.username')}</TH>
                  <TH className="text-right">{t('userAffiliate.rebate')}</TH>
                  <TH>{t('userAffiliate.joined')}</TH>
                </TR>
              </THead>
              <TBody>
                {invitees.map((item) => (
                  <TR key={item.user_id}>
                    <TD>{item.email || '-'}</TD>
                    <TD className="text-ink-2">{item.username || '-'}</TD>
                    <TD className="text-right font-mono text-signal-ok">{money(item.total_rebate)}</TD>
                    <TD className="text-xs text-ink-3">{item.created_at ? new Date(item.created_at).toLocaleString() : '-'}</TD>
                  </TR>
                ))}
                {invitees.length === 0 && (
                  <TR>
                    <TD colSpan={4} className="py-8 text-center text-ink-3">
                      {t('common.noData')}
                    </TD>
                  </TR>
                )}
              </TBody>
            </Table>
          </Card>
        </div>
      )}
    </>
  )
}
