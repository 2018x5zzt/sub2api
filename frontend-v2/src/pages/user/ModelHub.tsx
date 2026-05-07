import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQueries, useQuery } from '@tanstack/react-query'
import { Search, Copy, RefreshCw } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Skeleton } from '@/components/ui/Skeleton'
import { modelsAPI } from '@/api/models'
import { toast } from '@/components/ui/Toast'
import type { Group, GroupModelCatalog, GroupPlatform, SupportedModel } from '@/types'

interface FlatRow {
  group: Group
  model: SupportedModel
}

function formatPrice(price?: number): string {
  if (price == null) return '—'
  if (price === 0) return 'Free'
  if (price < 0.01) return `$${price.toFixed(6)}`
  return `$${price.toFixed(4)}`
}

const ALL: 'all' = 'all'

export default function ModelHubPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [platform, setPlatform] = useState<GroupPlatform | typeof ALL>(ALL)

  const groupsQuery = useQuery({
    queryKey: ['user-groups'],
    queryFn: () => modelsAPI.getUserGroups()
  })

  const groups = groupsQuery.data ?? []

  const catalogQueries = useQueries({
    queries: groups.map((g) => ({
      queryKey: ['group-catalog', g.id],
      queryFn: () => modelsAPI.getGroupModelCatalog(g.id),
      staleTime: 60_000
    }))
  })

  const catalogs: GroupModelCatalog[] = catalogQueries
    .map((q) => q.data)
    .filter((c): c is GroupModelCatalog => !!c)

  const loading = groupsQuery.isLoading || catalogQueries.some((q) => q.isLoading)

  const platformOptions = useMemo(() => {
    const set = new Set<GroupPlatform>()
    groups.forEach((g) => set.add(g.platform))
    return Array.from(set)
  }, [groups])

  const flatRows: FlatRow[] = useMemo(() => {
    const rows: FlatRow[] = []
    for (const c of catalogs) {
      if (platform !== ALL && c.group.platform !== platform) continue
      for (const m of c.models) {
        rows.push({ group: c.group, model: m })
      }
    }
    return rows
  }, [catalogs, platform])

  const visibleRows = useMemo(() => {
    if (!search.trim()) return flatRows
    const q = search.trim().toLowerCase()
    return flatRows.filter(
      ({ model, group }) =>
        model.id.toLowerCase().includes(q) ||
        (model.display_name || '').toLowerCase().includes(q) ||
        group.name.toLowerCase().includes(q)
    )
  }, [flatRows, search])

  const uniqueModelIds = useMemo(
    () => Array.from(new Set(catalogs.flatMap((c) => c.models.map((m) => m.id)))),
    [catalogs]
  )
  const visibleIds = useMemo(
    () => Array.from(new Set(visibleRows.map((r) => r.model.id))),
    [visibleRows]
  )

  function copyVisible() {
    if (!visibleIds.length) return
    navigator.clipboard.writeText(visibleIds.join('\n')).then(
      () => toast.success(t('common.copiedToClipboard') as string),
      () => toast.error(t('common.copyFailed') as string)
    )
  }

  const Stat = ({ label, value }: { label: string; value: string | number }) => (
    <div className="border border-line-1 rounded-xl p-4 bg-bg-1">
      <div className="text-eyebrow uppercase tracking-wider text-ink-3 font-mono">{label}</div>
      <div className="mt-2 font-display text-2xl text-ink-1">{value}</div>
    </div>
  )

  return (
    <>
      <PageHeader
        title={t('modelHub.title')}
        description={t('modelHub.description') as string}
        actions={
          <>
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                groupsQuery.refetch()
                catalogQueries.forEach((q) => q.refetch())
              }}
              disabled={loading}
            >
              <RefreshCw className={`h-3.5 w-3.5 ${loading ? 'animate-spin' : ''}`} />
              {t('common.refresh')}
            </Button>
            <Button type="button" variant="accent" onClick={copyVisible} disabled={!visibleIds.length}>
              <Copy className="h-3.5 w-3.5" />
              {t('modelHub.copyVisible')}
            </Button>
          </>
        }
      />

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
        <Stat label={t('modelHub.groupsLabel')} value={catalogs.length} />
        <Stat label={t('modelHub.uniqueModelsLabel')} value={uniqueModelIds.length} />
        <Stat label={t('modelHub.visibleModelsLabel')} value={visibleIds.length} />
        <Stat label={t('modelHub.platformsLabel')} value={platformOptions.length} />
      </div>

      <Card className="p-4 mb-4">
        <div className="flex flex-col md:flex-row gap-3 md:items-end">
          <div className="flex-1">
            <Input
              name="search"
              placeholder={t('modelHub.searchPlaceholder') as string}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              leftIcon={<Search className="h-4 w-4" />}
            />
          </div>
          <div className="flex flex-wrap gap-2">
            <button
              onClick={() => setPlatform(ALL)}
              className={`px-3 py-1.5 rounded-full text-xs border transition-colors ${
                platform === ALL
                  ? 'bg-orange text-white border-transparent'
                  : 'border-line-2 text-ink-2 hover:border-line-3'
              }`}
            >
              {t('common.all')}
            </button>
            {platformOptions.map((p) => (
              <button
                key={p}
                onClick={() => setPlatform(p)}
                className={`px-3 py-1.5 rounded-full text-xs border transition-colors capitalize ${
                  platform === p
                    ? 'bg-orange text-white border-transparent'
                    : 'border-line-2 text-ink-2 hover:border-line-3'
                }`}
              >
                {p}
              </button>
            ))}
          </div>
        </div>
      </Card>

      <Card className="p-0 overflow-hidden">
        {loading && catalogs.length === 0 ? (
          <div className="p-4 space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-12" />
            ))}
          </div>
        ) : visibleRows.length === 0 ? (
          <div className="text-center text-ink-3 py-12 text-sm">{t('common.noData')}</div>
        ) : (
          <div className="overflow-x-auto">
            <div
              className="grid px-5 py-3 border-b border-line-2 bg-bg-2 text-eyebrow uppercase tracking-wider text-ink-3 font-mono"
              style={{ gridTemplateColumns: '1fr 1.5fr 120px 140px 140px' }}
            >
              <div>Group</div>
              <div>Model</div>
              <div>Platform</div>
              <div className="text-right">Input / Mtok</div>
              <div className="text-right">Output / Mtok</div>
            </div>
            {visibleRows.map(({ group, model }) => (
              <div
                key={`${group.id}-${model.id}`}
                className="grid px-5 py-3 border-b border-line-1 hover:bg-bg-3 transition-colors items-center"
                style={{ gridTemplateColumns: '1fr 1.5fr 120px 140px 140px' }}
              >
                <div className="min-w-0 truncate text-ink-2 text-sm">
                  {group.name}
                  {group.rate_multiplier && group.rate_multiplier !== 1 && (
                    <span className="ml-2 font-mono text-xs text-orange">×{group.rate_multiplier}</span>
                  )}
                </div>
                <div className="min-w-0">
                  <div className="text-sm text-ink-1 font-mono truncate">{model.id}</div>
                  {model.display_name && model.display_name !== model.id && (
                    <div className="text-xs text-ink-3 truncate">{model.display_name}</div>
                  )}
                </div>
                <div>
                  <Badge>{group.platform}</Badge>
                </div>
                <div className="text-right font-mono text-sm text-ink-2">
                  {formatPrice(model.input_price_per_mtoken)}
                </div>
                <div className="text-right font-mono text-sm text-ink-2">
                  {formatPrice(model.output_price_per_mtoken)}
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </>
  )
}
