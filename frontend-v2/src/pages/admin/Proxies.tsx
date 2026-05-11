import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { RefreshCw, Search, ShieldCheck } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { adminProxiesAPI } from '@/api/admin/proxies'

type Row = Record<string, any>

function asArray(data: any): Row[] {
  if (Array.isArray(data)) return data
  for (const key of ['items', 'rows', 'data', 'proxies']) {
    if (Array.isArray(data?.[key])) return data[key]
  }
  return []
}

function text(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function statusTone(status: string) {
  const s = status.toLowerCase()
  if (['active', 'ok', 'success', 'healthy'].includes(s)) return 'success' as const
  if (['testing', 'checking', 'pending'].includes(s)) return 'warning' as const
  if (['error', 'failed', 'disabled'].includes(s)) return 'danger' as const
  return 'neutral' as const
}

export default function ProxiesPage() {
  const qc = useQueryClient()
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Row | null>(null)

  const query = useQuery({
    queryKey: ['admin-proxies', search],
    queryFn: () => adminProxiesAPI.listAdminProxies({ search: search || undefined, q: search || undefined })
  })

  const rows = useMemo(() => asArray(query.data), [query.data])

  const testMut = useMutation({
    mutationFn: (id: number | string) => adminProxiesAPI.testAdminProxy(id),
    onSuccess: async () => {
      toast.success('Proxy test submitted')
      await qc.invalidateQueries({ queryKey: ['admin-proxies'] })
    },
    onError: (e: { message?: string }) => toast.error(e?.message || 'Proxy test failed')
  })

  return (
    <>
      <PageHeader
        title="Proxies"
        description="Review proxy records and run connection tests."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        }
      />

      <Card className="p-4 mb-4">
        <div className="relative">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-ink-3" />
          <Input name="search" placeholder="Search proxies" value={search} onChange={(e) => setSearch(e.target.value)} className="pl-9" />
        </div>
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
                <TH>Proxy</TH>
                <TH>Type</TH>
                <TH>Status</TH>
                <TH>Region</TH>
                <TH className="text-right">Actions</TH>
              </TR>
            </THead>
            <TBody>
              {rows.map((row, index) => {
                const id = row.id ?? index
                const status = text(row.status || row.state || row.enabled)
                return (
                  <TR key={String(id)}>
                    <TD>
                      <button type="button" className="text-left" onClick={() => setSelected(row)}>
                        <div className="font-medium text-ink-1">{text(row.name || row.host || `Proxy #${id}`)}</div>
                        <div className="mt-1 text-xs text-ink-3 break-all">{text(row.url || row.endpoint || row.address)}</div>
                      </button>
                    </TD>
                    <TD className="text-sm text-ink-2">{text(row.proxy_type || row.type || row.protocol)}</TD>
                    <TD><Badge tone={statusTone(status)}>{status}</Badge></TD>
                    <TD className="text-xs text-ink-3">{text(row.region || row.country || row.location)}</TD>
                    <TD>
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => testMut.mutate(id)}
                          loading={testMut.isPending && testMut.variables === id}
                        >
                          <ShieldCheck className="h-3.5 w-3.5" />
                          Test
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => setSelected(row)}>
                          Detail
                        </Button>
                      </div>
                    </TD>
                  </TR>
                )
              })}
              {rows.length === 0 && (
                <TR>
                  <TD colSpan={5} className="py-8 text-center text-ink-3">No proxies found.</TD>
                </TR>
              )}
            </TBody>
          </Table>
        )}
      </Card>

      <Modal open={!!selected} onClose={() => setSelected(null)} title="Proxy detail" size="lg">
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
