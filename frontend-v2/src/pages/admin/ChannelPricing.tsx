import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { adminChannelsAPI } from '@/api/admin/channels'

type Row = Record<string, any>

function asArray(data: any): Row[] {
  if (Array.isArray(data)) return data
  for (const key of ['items', 'rows', 'data', 'channels']) {
    if (Array.isArray(data?.[key])) return data[key]
  }
  return []
}

function text(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function money(value: unknown) {
  const n = Number(value || 0)
  return Number.isFinite(n) ? `$${n.toFixed(4)}` : text(value)
}

export default function ChannelPricingPage() {
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Row | null>(null)

  const query = useQuery({
    queryKey: ['admin-channel-pricing', search],
    queryFn: () => adminChannelsAPI.listAdminChannels({ search: search || undefined, q: search || undefined })
  })

  const items = useMemo(() => asArray(query.data), [query.data])

  const pricingQuery = useQuery({
    queryKey: ['admin-channel-pricing-models'],
    queryFn: () => adminChannelsAPI.listChannelPricing(),
    enabled: !!query.data
  })

  return (
    <>
      <PageHeader
        title="Channel pricing"
        description="Review channel definitions and model pricing."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      <Card className="p-4 mb-4">
        <Input name="search" placeholder="Search channel or platform" value={search} onChange={(e) => setSearch(e.target.value)} />
      </Card>

      <Card className="p-4">
        {query.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Channel</TH>
                <TH>Platform</TH>
                <TH>Status</TH>
                <TH className="text-right">Models</TH>
                <TH className="text-right">Pricing</TH>
              </TR>
            </THead>
            <TBody>
              {items.map((row, index) => {
                const id = row.id ?? index
                const models = Array.isArray(row.supported_models) ? row.supported_models : Array.isArray(row.models) ? row.models : []
                const status = text(row.status || row.enabled || row.state)
                const pricing = row.pricing || row.model_pricing || {}
                return (
                  <TR key={String(id)}>
                    <TD>
                      <button type="button" className="text-left" onClick={() => setSelected(row)}>
                        <div className="font-medium text-ink-1">{text(row.name || row.channel_name || `Channel #${id}`)}</div>
                        <div className="mt-1 text-xs text-ink-3 break-all">{text(row.description || row.note || row.provider)}</div>
                      </button>
                    </TD>
                    <TD className="text-sm text-ink-2">{text(row.platform || row.provider || row.group_platform)}</TD>
                    <TD>
                      <Badge tone={String(status).toLowerCase().includes('enable') ? 'success' : 'neutral'}>{status}</Badge>
                    </TD>
                    <TD className="text-right font-mono text-ink-2">{models.length}</TD>
                    <TD className="text-right font-mono text-ink-2">
                      {Object.keys(pricing).length > 0 ? Object.values(pricing).slice(0, 3).map(money).join(' / ') : '-'}
                    </TD>
                  </TR>
                )
              })}
              {items.length === 0 && (
                <TR>
                  <TD colSpan={5} className="py-8 text-center text-ink-3">No channels found.</TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}
      </Card>

      <Card className="p-4 mt-4">
        <div className="mb-3 flex items-center justify-between">
          <div>
            <h2 className="text-base font-medium text-ink-1">Model pricing</h2>
            <p className="text-xs text-ink-3">Defensive view of the model pricing matrix.</p>
          </div>
          <Badge>{Array.isArray((pricingQuery.data as any)?.items) ? (pricingQuery.data as any).items.length : 0}</Badge>
        </div>
        {pricingQuery.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} className="h-10" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR><TH>Field</TH><TH>Value</TH></TR>
            </THead>
            <TBody>
              {Object.entries((pricingQuery.data as any) || {}).map(([key, value]) => (
                <TR key={key}>
                  <TD className="font-mono text-xs text-ink-3">{key}</TD>
                  <TD className="font-mono text-xs break-all text-ink-2">{typeof value === 'object' ? JSON.stringify(value) : text(value)}</TD>
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      <Modal open={!!selected} onClose={() => setSelected(null)} title="Channel detail" size="lg">
        <Table>
          <THead>
            <TR><TH>Field</TH><TH>Value</TH></TR>
          </THead>
          <TBody>
            {selected && Object.entries(selected).map(([key, value]) => (
              <TR key={key}>
                <TD className="font-mono text-xs text-ink-3">{key}</TD>
                <TD className="font-mono text-xs break-all text-ink-2">{typeof value === 'object' ? JSON.stringify(value) : text(value)}</TD>
              </TR>
            ))}
          </TBody>
        </Table>
      </Modal>
    </>
  )
}
