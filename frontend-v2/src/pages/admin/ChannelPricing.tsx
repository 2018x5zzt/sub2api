import { useMemo, useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Plus, RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { adminChannelsAPI } from '@/api/admin/channels'
import { adminGroupsAPI } from '@/api/admin/groups'
import type { AdminGroup } from '@/types'

type Row = Record<string, any>
type BillingModelSource = 'channel_mapped' | 'requested' | 'upstream'

interface CreateChannelFormState {
  name: string
  description: string
  billingModelSource: BillingModelSource
  restrictModels: boolean
  groupIds: number[]
}

const DEFAULT_CREATE_FORM: CreateChannelFormState = {
  name: '',
  description: '',
  billingModelSource: 'channel_mapped',
  restrictModels: false,
  groupIds: []
}

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
  const qc = useQueryClient()
  const [search, setSearch] = useState('')
  const [selected, setSelected] = useState<Row | null>(null)
  const [creating, setCreating] = useState(false)
  const [createForm, setCreateForm] = useState<CreateChannelFormState>(DEFAULT_CREATE_FORM)

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

  const groupsQuery = useQuery({
    queryKey: ['admin-groups-all-for-channel-create'],
    queryFn: () => adminGroupsAPI.listAllGroups()
  })

  const createMutation = useMutation({
    mutationFn: (payload: Record<string, unknown>) => adminChannelsAPI.createAdminChannel(payload),
    onSuccess: async () => {
      toast.success('Channel created')
      setCreating(false)
      setCreateForm(DEFAULT_CREATE_FORM)
      await Promise.all([
        qc.invalidateQueries({ queryKey: ['admin-channel-pricing'] }),
        qc.invalidateQueries({ queryKey: ['admin-channel-pricing-models'] })
      ])
    },
    onError: (error: { message?: string }) => toast.error(error?.message || 'Failed to create channel')
  })

  const groups = useMemo(() => {
    const list = Array.isArray(groupsQuery.data) ? groupsQuery.data : []
    return list.slice().sort((a: AdminGroup, b: AdminGroup) => a.name.localeCompare(b.name))
  }, [groupsQuery.data])

  function resetCreateForm() {
    setCreateForm(DEFAULT_CREATE_FORM)
  }

  function openCreateModal() {
    resetCreateForm()
    setCreating(true)
  }

  function closeCreateModal() {
    if (createMutation.isPending) return
    setCreating(false)
  }

  function toggleGroup(groupId: number) {
    setCreateForm((prev) => {
      if (prev.groupIds.includes(groupId)) {
        return { ...prev, groupIds: prev.groupIds.filter((id) => id !== groupId) }
      }
      return { ...prev, groupIds: [...prev.groupIds, groupId] }
    })
  }

  function submitCreateForm(e: FormEvent) {
    e.preventDefault()

    if (!createForm.name.trim()) {
      toast.warning('Channel name is required')
      return
    }
    if (createForm.groupIds.length === 0) {
      toast.warning('Select at least one group')
      return
    }

    createMutation.mutate({
      name: createForm.name.trim(),
      description: createForm.description.trim(),
      group_ids: createForm.groupIds,
      billing_model_source: createForm.billingModelSource,
      restrict_models: createForm.restrictModels,
      model_pricing: []
    })
  }

  return (
    <>
      <PageHeader
        title="Channel pricing"
        description="Review channel definitions and model pricing."
        actions={
          <div className="flex items-center gap-2">
            <Button variant="accent" onClick={openCreateModal}>
              <Plus className="h-4 w-4" />
              Create channel
            </Button>
            <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
              <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
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

      <Modal
        open={creating}
        onClose={closeCreateModal}
        title="Create channel"
        size="lg"
        footer={
          <>
            <Button variant="ghost" onClick={closeCreateModal} disabled={createMutation.isPending}>Cancel</Button>
            <Button form="create-channel-form" type="submit" loading={createMutation.isPending}>Create</Button>
          </>
        }
      >
        <form id="create-channel-form" className="space-y-4" onSubmit={submitCreateForm}>
          <Input
            name="channel-name"
            label="Name"
            value={createForm.name}
            onChange={(e) => setCreateForm((prev) => ({ ...prev, name: e.target.value }))}
            placeholder="Enter channel name"
            autoFocus
            required
          />
          <Input
            name="channel-description"
            label="Description"
            value={createForm.description}
            onChange={(e) => setCreateForm((prev) => ({ ...prev, description: e.target.value }))}
            placeholder="Optional description"
          />
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3 items-end">
            <div>
              <label className="input-label">Billing model</label>
              <select
                className="input appearance-none cursor-pointer bg-bg-4"
                value={createForm.billingModelSource}
                onChange={(e) =>
                  setCreateForm((prev) => ({
                    ...prev,
                    billingModelSource: e.target.value as BillingModelSource
                  }))
                }
              >
                <option value="channel_mapped" className="bg-bg-4">Bill by channel-mapped model</option>
                <option value="requested" className="bg-bg-4">Bill by requested model</option>
                <option value="upstream" className="bg-bg-4">Bill by final upstream model</option>
              </select>
            </div>
            <label className="inline-flex items-center gap-2 cursor-pointer pb-2">
              <input
                type="checkbox"
                checked={createForm.restrictModels}
                onChange={(e) =>
                  setCreateForm((prev) => ({ ...prev, restrictModels: e.target.checked }))
                }
                className="w-4 h-4 accent-orange"
              />
              <span className="text-sm text-ink-2">Restrict models</span>
            </label>
          </div>

          <div>
            <div className="flex items-center justify-between gap-3">
              <label className="input-label mb-0">Associated groups</label>
              <span className="text-xs text-ink-3">{createForm.groupIds.length} selected</span>
            </div>

            {groupsQuery.isLoading ? (
              <div className="space-y-2 mt-2">
                {Array.from({ length: 4 }).map((_, i) => (
                  <Skeleton key={i} className="h-10" />
                ))}
              </div>
            ) : groups.length === 0 ? (
              <p className="text-sm text-ink-3 mt-2">No groups available. Create a group first.</p>
            ) : (
              <div className="mt-2 max-h-56 overflow-y-auto rounded-xl border border-line-2 divide-y divide-line-2">
                {groups.map((group) => {
                  const checked = createForm.groupIds.includes(group.id)
                  return (
                    <label key={group.id} className="flex items-center justify-between gap-3 px-3 py-2 cursor-pointer hover:bg-bg-3/60">
                      <div className="min-w-0">
                        <div className="text-sm font-medium text-ink-1 truncate">{group.name}</div>
                        <div className="text-xs text-ink-3">
                          {group.platform} · {group.status}
                        </div>
                      </div>
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => toggleGroup(group.id)}
                        className="w-4 h-4 accent-orange shrink-0"
                      />
                    </label>
                  )
                })}
              </div>
            )}
          </div>
        </form>
      </Modal>
    </>
  )
}
