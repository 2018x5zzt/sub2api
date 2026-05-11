import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Table, THead, TBody, TR, TH, TD } from '@/components/ui/Table'
import { Skeleton } from '@/components/ui/Skeleton'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { usageAPI } from '@/api/usage'

export default function UsagePage() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['user-usage', page],
    queryFn: () => usageAPI.listUsage({ page, page_size: 20 })
  })

  return (
    <>
      <PageHeader title={t('usage.title')} description={t('usage.description') as string} />
      <Card className="p-4">
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>{t('dashboard.model')}</TH>
                <TH className="text-right">{t('dashboard.tokens')}</TH>
                <TH className="text-right">{t('dashboard.actual')}</TH>
                <TH>{t('common.status')}</TH>
                <TH>{t('usage.time')}</TH>
              </TR>
            </THead>
            <TBody>
              {(data?.items ?? []).map((log) => (
                <TR key={log.id}>
                  <TD className="font-mono text-xs">{log.model}</TD>
                  <TD className="text-right font-mono">
                    {(log.input_tokens + log.output_tokens + log.cache_creation_tokens + log.cache_read_tokens).toLocaleString()}
                  </TD>
                  <TD className="text-right font-mono">${log.actual_cost.toFixed(6)}</TD>
                  <TD>
                    {log.stream ? <Badge tone="accent">{t('usage.stream')}</Badge> : <Badge>{t('usage.sync')}</Badge>}
                  </TD>
                  <TD className="text-ink-3 text-xs">
                    {new Date(log.created_at).toLocaleString()}
                  </TD>
                </TR>
              ))}
              {(data?.items ?? []).length === 0 && (
                <TR>
                  <TD colSpan={5} className="text-center text-ink-3 py-8">{t('common.noData')}</TD>
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
    </>
  )
}
