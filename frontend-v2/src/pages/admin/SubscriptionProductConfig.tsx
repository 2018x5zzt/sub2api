import { useMemo, useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Eye, Link2, Plus, RefreshCw, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Modal } from '@/components/ui/Modal'
import { Skeleton } from '@/components/ui/Skeleton'
import { Table, TBody, TD, TH, THead, TR } from '@/components/ui/Table'
import { toast } from '@/components/ui/Toast'
import { adminGroupsAPI } from '@/api/admin/groups'
import {
  adminSubscriptionProductsAPI,
  type AdminSubscriptionProduct,
  type AdminSubscriptionProductBinding,
  type AdminUserProductSubscription,
  type SubscriptionProductPayload,
  type SyncSubscriptionProductBindingRequest
} from '@/api/admin/subscriptionProducts'

type ProductForm = Required<Pick<SubscriptionProductPayload, 'code' | 'name' | 'description' | 'status' | 'product_family'>> & {
  default_validity_days: string
  daily_limit_usd: string
  weekly_limit_usd: string
  monthly_limit_usd: string
  sort_order: string
}

type BindingFormRow = {
  local_id: string
  group_id: string
  debit_multiplier: string
  status: string
  sort_order: string
}

const emptyProductForm: ProductForm = {
  code: '',
  name: '',
  description: '',
  status: 'active',
  product_family: 'gpt',
  default_validity_days: '30',
  daily_limit_usd: '0',
  weekly_limit_usd: '0',
  monthly_limit_usd: '0',
  sort_order: '0'
}

const selectClass = 'input appearance-none cursor-pointer bg-bg-4'

function money(value: number | null | undefined) {
  const n = Number(value || 0)
  return n > 0 ? `$${n.toFixed(2)}` : 'Unlimited'
}

function usage(value: number | null | undefined) {
  return `$${Number(value || 0).toFixed(2)}`
}

function statusTone(status: string) {
  if (status === 'active') return 'success' as const
  if (status === 'disabled' || status === 'inactive' || status === 'revoked') return 'danger' as const
  if (status === 'expired' || status === 'draft') return 'warning' as const
  return 'neutral' as const
}

function toProductForm(product: AdminSubscriptionProduct | null): ProductForm {
  if (!product) return { ...emptyProductForm }
  return {
    code: product.code || '',
    name: product.name || '',
    description: product.description || '',
    status: product.status || 'active',
    product_family: product.product_family || 'gpt',
    default_validity_days: String(product.default_validity_days ?? 30),
    daily_limit_usd: String(product.daily_limit_usd ?? 0),
    weekly_limit_usd: String(product.weekly_limit_usd ?? 0),
    monthly_limit_usd: String(product.monthly_limit_usd ?? 0),
    sort_order: String(product.sort_order ?? 0)
  }
}

function formPayload(form: ProductForm): SubscriptionProductPayload {
  return {
    code: form.code.trim(),
    name: form.name.trim(),
    description: form.description.trim(),
    status: form.status,
    product_family: form.product_family.trim() || 'gpt',
    default_validity_days: Number(form.default_validity_days) || 30,
    daily_limit_usd: Number(form.daily_limit_usd) || 0,
    weekly_limit_usd: Number(form.weekly_limit_usd) || 0,
    monthly_limit_usd: Number(form.monthly_limit_usd) || 0,
    sort_order: Number(form.sort_order) || 0
  }
}

export default function AdminSubscriptionProductConfigPage() {
  const { t } = useTranslation()
  const qc = useQueryClient()
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('')
  const [editing, setEditing] = useState<AdminSubscriptionProduct | null>(null)
  const [form, setForm] = useState<ProductForm>(emptyProductForm)
  const [productDialogOpen, setProductDialogOpen] = useState(false)
  const [bindingProduct, setBindingProduct] = useState<AdminSubscriptionProduct | null>(null)
  const [bindingRows, setBindingRows] = useState<BindingFormRow[]>([])
  const [subscriptionProduct, setSubscriptionProduct] = useState<AdminSubscriptionProduct | null>(null)

  const productsQuery = useQuery({
    queryKey: ['admin-subscription-products'],
    queryFn: adminSubscriptionProductsAPI.listProducts
  })

  const groupsQuery = useQuery({
    queryKey: ['admin-groups-all-for-product-bindings'],
    queryFn: () => adminGroupsAPI.listAllGroups()
  })

  const bindingsQuery = useQuery({
    queryKey: ['admin-subscription-product-bindings', bindingProduct?.id],
    enabled: !!bindingProduct,
    queryFn: () => adminSubscriptionProductsAPI.listBindings(bindingProduct!.id)
  })

  const subscriptionsQuery = useQuery({
    queryKey: ['admin-subscription-product-users', subscriptionProduct?.id],
    enabled: !!subscriptionProduct,
    queryFn: () => adminSubscriptionProductsAPI.listProductSubscriptions(subscriptionProduct!.id)
  })

  const filteredProducts = useMemo(() => {
    const q = search.trim().toLowerCase()
    return (productsQuery.data ?? []).filter((product) => {
      if (status && product.status !== status) return false
      if (!q) return true
      return [product.name, product.code, product.description, product.product_family].some((value) =>
        String(value || '').toLowerCase().includes(q)
      )
    })
  }, [productsQuery.data, search, status])

  const saveProduct = useMutation({
    mutationFn: () => {
      const payload = formPayload(form)
      if (!payload.code || !payload.name) throw new Error('Code and name are required')
      return editing
        ? adminSubscriptionProductsAPI.updateProduct(editing.id, payload)
        : adminSubscriptionProductsAPI.createProduct(payload)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-subscription-products'] })
      setProductDialogOpen(false)
      setEditing(null)
      toast.success(t('common.success') as string)
    },
    onError: (error: { message?: string }) => toast.error(error?.message || (t('common.error') as string))
  })

  const saveBindings = useMutation({
    mutationFn: () => {
      if (!bindingProduct) throw new Error('No product selected')
      const seen = new Set<number>()
      const bindings: SyncSubscriptionProductBindingRequest[] = bindingRows
        .filter((row) => row.group_id)
        .map((row) => {
          const groupId = Number(row.group_id)
          if (seen.has(groupId)) throw new Error('Duplicate group binding')
          seen.add(groupId)
          return {
            group_id: groupId,
            debit_multiplier: Number(row.debit_multiplier) || 1,
            status: row.status || 'active',
            sort_order: Number(row.sort_order) || 0
          }
        })
      return adminSubscriptionProductsAPI.syncBindings(bindingProduct.id, bindings)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-subscription-product-bindings', bindingProduct?.id] })
      setBindingProduct(null)
      setBindingRows([])
      toast.success(t('common.success') as string)
    },
    onError: (error: { message?: string }) => toast.error(error?.message || (t('common.error') as string))
  })

  function openProduct(product: AdminSubscriptionProduct | null) {
    setEditing(product)
    setForm(toProductForm(product))
    setProductDialogOpen(true)
  }

  function openBindings(product: AdminSubscriptionProduct) {
    setBindingProduct(product)
    setBindingRows([])
    adminSubscriptionProductsAPI.listBindings(product.id)
      .then((bindings) => setBindingRows(bindings.map(bindingToRow)))
      .catch((error: { message?: string }) => toast.error(error?.message || (t('common.error') as string)))
  }

  function addBindingRow() {
    setBindingRows((rows) => [
      ...rows,
      {
        local_id: `${Date.now()}-${Math.random()}`,
        group_id: '',
        debit_multiplier: '1',
        status: 'active',
        sort_order: String(rows.length + 1)
      }
    ])
  }

  function updateBindingRow(localId: string, patch: Partial<BindingFormRow>) {
    setBindingRows((rows) => rows.map((row) => row.local_id === localId ? { ...row, ...patch } : row))
  }

  function removeBindingRow(localId: string) {
    setBindingRows((rows) => rows.filter((row) => row.local_id !== localId))
  }

  function bindingToRow(binding: AdminSubscriptionProductBinding): BindingFormRow {
    return {
      local_id: `${binding.group_id}-${binding.sort_order ?? 0}`,
      group_id: String(binding.group_id),
      debit_multiplier: String(binding.debit_multiplier ?? 1),
      status: binding.status || 'active',
      sort_order: String(binding.sort_order ?? 0)
    }
  }

  function submitProduct(event: FormEvent) {
    event.preventDefault()
    saveProduct.mutate()
  }

  function submitBindings(event: FormEvent) {
    event.preventDefault()
    saveBindings.mutate()
  }

  return (
    <>
      <PageHeader
        title={t('nav.subscriptionProductConfig')}
        description="Configure product-level subscription quotas and bind products to groups."
        actions={
          <Button variant="accent" onClick={() => openProduct(null)}>
            <Plus className="h-4 w-4" />
            Create Product
          </Button>
        }
      />

      <Card className="p-4">
        <div className="mb-4 flex flex-wrap items-center gap-3">
          <Input
            name="search"
            placeholder="Search products"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            leftIcon={<Search className="h-4 w-4" />}
            className="max-w-xs"
          />
          <select value={status} onChange={(event) => setStatus(event.target.value)} className={`${selectClass} max-w-[160px]`}>
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="draft">Draft</option>
            <option value="disabled">Disabled</option>
          </select>
          <Button variant="secondary" onClick={() => productsQuery.refetch()} loading={productsQuery.isFetching}>
            <RefreshCw className="h-4 w-4" />
            Refresh
          </Button>
        </div>

        {productsQuery.isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, index) => <Skeleton key={index} className="h-12" />)}
          </div>
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>Product</TH>
                <TH>Status</TH>
                <TH>Limits</TH>
                <TH className="text-right">Validity</TH>
                <TH className="text-right">Actions</TH>
              </TR>
            </THead>
            <TBody>
              {filteredProducts.map((product) => (
                <TR key={product.id}>
                  <TD>
                    <div className="font-medium text-ink-1">{product.name}</div>
                    <div className="mt-1 font-mono text-xs text-ink-3">{product.code}</div>
                    {product.description && <div className="mt-1 max-w-md truncate text-xs text-ink-3">{product.description}</div>}
                  </TD>
                  <TD><Badge tone={statusTone(product.status)}>{product.status}</Badge></TD>
                  <TD className="text-xs text-ink-2">
                    <div>Daily: {money(product.daily_limit_usd)}</div>
                    <div>Weekly: {money(product.weekly_limit_usd)}</div>
                    <div>Monthly: {money(product.monthly_limit_usd)}</div>
                  </TD>
                  <TD className="text-right font-mono text-sm">{product.default_validity_days ?? 30}d</TD>
                  <TD className="text-right">
                    <div className="inline-flex gap-1">
                      <Button size="sm" variant="secondary" onClick={() => openProduct(product)}>Edit</Button>
                      <button className="btn btn-ghost btn-icon btn-sm" title="Bind Groups" onClick={() => openBindings(product)}>
                        <Link2 className="h-3.5 w-3.5" />
                      </button>
                      <button className="btn btn-ghost btn-icon btn-sm" title="View Subscriptions" onClick={() => setSubscriptionProduct(product)}>
                        <Eye className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </TD>
                </TR>
              ))}
              {filteredProducts.length === 0 && (
                <TR><TD colSpan={5} className="py-10 text-center text-ink-3">No products</TD></TR>
              )}
            </TBody>
          </Table>
        )}
      </Card>

      <Modal
        open={productDialogOpen}
        onClose={() => setProductDialogOpen(false)}
        title={editing ? 'Edit Product' : 'Create Product'}
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => setProductDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button variant="accent" loading={saveProduct.isPending} onClick={() => saveProduct.mutate()}>{t('common.save')}</Button>
          </>
        }
      >
        <form id="subscription-product-form" className="grid gap-4 sm:grid-cols-2" onSubmit={submitProduct}>
          <Input label="Code" required value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} />
          <Input label="Name" required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <div>
            <label className="input-label">Status</label>
            <select className={selectClass} value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
              <option value="active">Active</option>
              <option value="draft">Draft</option>
              <option value="disabled">Disabled</option>
            </select>
          </div>
          <Input label="Default Validity Days" type="number" min="1" value={form.default_validity_days} onChange={(e) => setForm({ ...form, default_validity_days: e.target.value })} />
          <Input label="Daily Limit USD" type="number" min="0" step="0.0001" value={form.daily_limit_usd} onChange={(e) => setForm({ ...form, daily_limit_usd: e.target.value })} />
          <Input label="Weekly Limit USD" type="number" min="0" step="0.0001" value={form.weekly_limit_usd} onChange={(e) => setForm({ ...form, weekly_limit_usd: e.target.value })} />
          <Input label="Monthly Limit USD" type="number" min="0" step="0.0001" value={form.monthly_limit_usd} onChange={(e) => setForm({ ...form, monthly_limit_usd: e.target.value })} />
          <Input label="Sort Order" type="number" value={form.sort_order} onChange={(e) => setForm({ ...form, sort_order: e.target.value })} />
          <div className="sm:col-span-2">
            <label className="input-label">Description</label>
            <textarea className="input min-h-24 py-2" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </div>
        </form>
      </Modal>

      <Modal
        open={!!bindingProduct}
        onClose={() => setBindingProduct(null)}
        title={bindingProduct ? `Bind Groups: ${bindingProduct.name}` : 'Bind Groups'}
        size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => setBindingProduct(null)}>{t('common.cancel')}</Button>
            <Button variant="accent" loading={saveBindings.isPending} onClick={() => saveBindings.mutate()}>{t('common.save')}</Button>
          </>
        }
      >
        <form className="space-y-3" onSubmit={submitBindings}>
          <div className="flex justify-end">
            <Button type="button" size="sm" variant="secondary" onClick={addBindingRow}>
              <Plus className="h-3.5 w-3.5" />
              Add Binding
            </Button>
          </div>
          {bindingsQuery.isFetching && bindingRows.length === 0 ? (
            <Skeleton className="h-28" />
          ) : bindingRows.length === 0 ? (
            <div className="rounded-xl border border-line-2 bg-bg-2 p-6 text-center text-sm text-ink-3">No bindings</div>
          ) : (
            bindingRows.map((row) => (
              <div key={row.local_id} className="grid gap-3 rounded-xl border border-line-2 p-3 md:grid-cols-[1fr_120px_120px_90px_auto]">
                <select className={selectClass} value={row.group_id} onChange={(event) => updateBindingRow(row.local_id, { group_id: event.target.value })}>
                  <option value="">Select group</option>
                  {(groupsQuery.data ?? []).map((group) => (
                    <option key={group.id} value={group.id}>{group.name} #{group.id}</option>
                  ))}
                </select>
                <Input aria-label="Debit multiplier" type="number" min="0.0001" step="0.0001" value={row.debit_multiplier} onChange={(e) => updateBindingRow(row.local_id, { debit_multiplier: e.target.value })} />
                <select className={selectClass} value={row.status} onChange={(event) => updateBindingRow(row.local_id, { status: event.target.value })}>
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
                <Input aria-label="Sort order" type="number" value={row.sort_order} onChange={(e) => updateBindingRow(row.local_id, { sort_order: e.target.value })} />
                <Button type="button" size="sm" variant="danger" onClick={() => removeBindingRow(row.local_id)}>Remove</Button>
              </div>
            ))
          )}
        </form>
      </Modal>

      <Modal open={!!subscriptionProduct} onClose={() => setSubscriptionProduct(null)} title={subscriptionProduct ? `Subscriptions: ${subscriptionProduct.name}` : 'Subscriptions'} size="lg">
        {subscriptionsQuery.isFetching ? (
          <Skeleton className="h-40" />
        ) : (
          <Table>
            <THead>
              <TR>
                <TH>ID</TH>
                <TH>User</TH>
                <TH>Status</TH>
                <TH>Expires</TH>
                <TH>Usage</TH>
              </TR>
            </THead>
            <TBody>
              {(subscriptionsQuery.data ?? []).map((item) => <SubscriptionRow key={item.id} item={item} />)}
              {(subscriptionsQuery.data ?? []).length === 0 && <TR><TD colSpan={5} className="py-8 text-center text-ink-3">No subscriptions</TD></TR>}
            </TBody>
          </Table>
        )}
      </Modal>
    </>
  )
}

function SubscriptionRow({ item }: { item: AdminUserProductSubscription }) {
  return (
    <TR>
      <TD className="font-mono text-xs">#{item.id}</TD>
      <TD className="font-mono text-xs">#{item.user_id}</TD>
      <TD><Badge tone={statusTone(item.status)}>{item.status}</Badge></TD>
      <TD className="font-mono text-xs">{item.expires_at ? new Date(item.expires_at).toLocaleDateString() : '-'}</TD>
      <TD className="text-xs text-ink-2">
        <div>Daily: {usage(item.daily_usage_usd)}</div>
        <div>Weekly: {usage(item.weekly_usage_usd)}</div>
        <div>Monthly: {usage(item.monthly_usage_usd)}</div>
      </TD>
    </TR>
  )
}
