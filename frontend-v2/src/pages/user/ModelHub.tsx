import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Check, Copy, RefreshCw, Search } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import {
  channelsAPI,
  type AvailableChannelEntry,
  type AvailableChannelGroup,
  type AvailableChannelPricing
} from '@/api/channels'
import { toast } from '@/components/ui/Toast'
import { cn } from '@/lib/cn'

const PER_MILLION_TOKENS = 1_000_000
const ALL = 'all'

interface CatalogModel {
  id: string
  pricing?: AvailableChannelPricing | null
  channels: string[]
}

interface GroupCatalog {
  group: AvailableChannelGroup
  effectiveRate: number
  userRate: number | null
  channels: string[]
  models: CatalogModel[]
}

function formatRate(rate: number | null | undefined): string {
  const value = Number(rate ?? 1)
  if (!Number.isFinite(value)) return '1'
  return value.toFixed(2).replace(/\.?0+$/, '')
}

function formatMoney(value: number): string {
  if (!Number.isFinite(value)) return '-'
  if (Math.abs(value) >= 1) return `$${value.toFixed(2)}`
  if (value === 0) return '$0'
  return `$${value.toFixed(4)}`
}

function formatPerMillion(price: number | null | undefined, rate: number): string | null {
  if (price === null || price === undefined) return null
  return `${formatMoney(price * PER_MILLION_TOKENS * rate)} / 1M`
}

function formatPerRequest(price: number | null | undefined, rate: number): string | null {
  if (price === null || price === undefined) return null
  return `${formatMoney(price * rate)} / 次`
}

function platformLabel(platform: string, t: ReturnType<typeof useTranslation>['t']) {
  const key = `admin.groups.platforms.${platform}`
  const label = t(key)
  return label === key ? platform : label
}

function collectCatalogs(channels: AvailableChannelEntry[], rates: Record<number, number>): GroupCatalog[] {
  const byGroup = new Map<number, GroupCatalog>()
  const seenModels = new Map<number, Set<string>>()

  for (const channel of channels) {
    for (const section of channel.platforms || []) {
      for (const group of section.groups || []) {
        let catalog = byGroup.get(group.id)
        if (!catalog) {
          const userRate = rates[group.id] ?? null
          catalog = {
            group,
            effectiveRate: userRate ?? group.rate_multiplier ?? 1,
            userRate,
            channels: [],
            models: []
          }
          byGroup.set(group.id, catalog)
          seenModels.set(group.id, new Set<string>())
        }
        if (!catalog.channels.includes(channel.name)) {
          catalog.channels.push(channel.name)
        }

        const seen = seenModels.get(group.id)!
        for (const model of section.supported_models || []) {
          if (!model?.name || seen.has(model.name)) continue
          seen.add(model.name)
          catalog.models.push({
            id: model.name,
            pricing: model.pricing,
            channels: [channel.name]
          })
        }
      }
    }
  }

  return Array.from(byGroup.values())
    .map((catalog) => ({
      ...catalog,
      models: catalog.models.sort((a, b) => a.id.localeCompare(b.id))
    }))
    .sort((a, b) => a.group.name.localeCompare(b.group.name))
}

function pricingBadges(model: CatalogModel, rate: number) {
  const pricing = model.pricing
  const badges: Array<{ key: string; text: string; tone: 'neutral' | 'success' | 'warning' | 'accent' }> = []

  const input = formatPerMillion(pricing?.input_price, rate)
  const output = formatPerMillion(pricing?.output_price, rate)
  const image = formatPerMillion(pricing?.image_output_price, rate)
  const request = formatPerRequest(pricing?.per_request_price, rate)

  if (input) badges.push({ key: 'input', text: `输入 ${input}`, tone: 'success' })
  if (output) badges.push({ key: 'output', text: `输出 ${output}`, tone: 'warning' })
  if (image) badges.push({ key: 'image', text: `图片 ${image}`, tone: 'accent' })
  if (request) badges.push({ key: 'request', text: `请求 ${request}`, tone: 'neutral' })

  for (const [index, interval] of (pricing?.intervals || []).entries()) {
    const perReq = formatPerRequest(interval.per_request_price, rate)
    if (perReq) {
      badges.push({
        key: `request-${index}`,
        text: `${interval.tier_label || '阶梯'} ${perReq}`,
        tone: 'neutral'
      })
    }
  }

  return badges
}

export default function ModelHubPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [platform, setPlatform] = useState<string>(ALL)
  const [activeGroupId, setActiveGroupId] = useState<number | typeof ALL>(ALL)
  const [copiedKey, setCopiedKey] = useState<string | null>(null)

  const channelsQuery = useQuery({
    queryKey: ['available-channels'],
    queryFn: () => channelsAPI.listAvailableChannels()
  })
  const ratesQuery = useQuery({
    queryKey: ['user-group-rates'],
    queryFn: () => channelsAPI.getUserGroupRates(),
    retry: false
  })

  const catalogs = useMemo(
    () => collectCatalogs(channelsQuery.data ?? [], (ratesQuery.data ?? {}) as Record<number, number>),
    [channelsQuery.data, ratesQuery.data]
  )

  const platformOptions = useMemo(
    () => Array.from(new Set(catalogs.map((catalog) => catalog.group.platform).filter(Boolean))),
    [catalogs]
  )

  const filteredCatalogs = useMemo(() => {
    const q = search.trim().toLowerCase()
    return catalogs
      .filter((catalog) => platform === ALL || catalog.group.platform === platform)
      .filter((catalog) => activeGroupId === ALL || catalog.group.id === activeGroupId)
      .map((catalog) => {
        if (!q) return catalog
        const groupMatch =
          catalog.group.name.toLowerCase().includes(q) ||
          catalog.group.platform.toLowerCase().includes(q) ||
          catalog.channels.some((channel) => channel.toLowerCase().includes(q))
        const models = groupMatch ? catalog.models : catalog.models.filter((model) => model.id.toLowerCase().includes(q))
        return models.length > 0 ? { ...catalog, models } : null
      })
      .filter((catalog): catalog is GroupCatalog => catalog !== null)
  }, [activeGroupId, catalogs, platform, search])

  const visibleModelIds = useMemo(
    () => Array.from(new Set(filteredCatalogs.flatMap((catalog) => catalog.models.map((model) => model.id)))),
    [filteredCatalogs]
  )
  const allModelIds = useMemo(
    () => Array.from(new Set(catalogs.flatMap((catalog) => catalog.models.map((model) => model.id)))),
    [catalogs]
  )

  const loading = channelsQuery.isLoading
  const hasFilters = search.trim().length > 0 || platform !== ALL || activeGroupId !== ALL

  function markCopied(key: string) {
    setCopiedKey(key)
    window.setTimeout(() => setCopiedKey((current) => (current === key ? null : current)), 1600)
  }

  function copyText(text: string, key: string) {
    if (!text) return
    navigator.clipboard.writeText(text).then(
      () => {
        markCopied(key)
        toast.success(t('common.copiedToClipboard') as string)
      },
      () => toast.error(t('common.copyFailed') as string)
    )
  }

  function copyVisible() {
    copyText(visibleModelIds.join('\n'), 'visible')
  }

  function copyGroup(catalog: GroupCatalog) {
    copyText(catalog.models.map((model) => model.id).join('\n'), `group:${catalog.group.id}`)
  }

  const Stat = ({ label, value }: { label: string; value: number }) => (
    <div className="rounded-xl border border-line-1 bg-bg-1 p-4">
      <div className="font-mono text-[11px] uppercase tracking-[0.14em] text-ink-3">{label}</div>
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
                channelsQuery.refetch()
                ratesQuery.refetch()
              }}
              disabled={loading || channelsQuery.isFetching}
            >
              <RefreshCw className={cn('h-3.5 w-3.5', (loading || channelsQuery.isFetching) && 'animate-spin')} />
              {t('common.refresh')}
            </Button>
            <Button type="button" variant="accent" onClick={copyVisible} disabled={!visibleModelIds.length}>
              {copiedKey === 'visible' ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
              {t('modelHub.copyVisible')}
            </Button>
          </>
        }
      />

      <div className="grid grid-cols-2 gap-3 md:grid-cols-4">
        <Stat label={t('modelHub.groupsLabel') as string} value={catalogs.length} />
        <Stat label={t('modelHub.uniqueModelsLabel') as string} value={allModelIds.length} />
        <Stat label={t('modelHub.visibleModelsLabel') as string} value={visibleModelIds.length} />
        <Stat label={t('modelHub.platformsLabel') as string} value={platformOptions.length} />
      </div>

      <div className="mt-4 grid gap-4 lg:grid-cols-[280px,minmax(0,1fr)]">
        <Card className="p-4 lg:sticky lg:top-4 lg:self-start">
          <div className="space-y-5">
            <Input
              name="search"
              label={t('modelHub.searchLabel') as string}
              placeholder={t('modelHub.searchPlaceholder') as string}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              leftIcon={<Search className="h-4 w-4" />}
            />

            <div>
              <div className="input-label">{t('modelHub.platformFilterLabel')}</div>
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  onClick={() => setPlatform(ALL)}
                  className={cn('rounded-full border px-3 py-1.5 text-xs transition-colors', platform === ALL ? 'border-transparent bg-orange text-white' : 'border-line-2 text-ink-2 hover:border-line-3')}
                >
                  {t('modelHub.allPlatforms')}
                </button>
                {platformOptions.map((item) => (
                  <button
                    key={item}
                    type="button"
                    onClick={() => setPlatform(item)}
                    className={cn('rounded-full border px-3 py-1.5 text-xs transition-colors', platform === item ? 'border-transparent bg-orange text-white' : 'border-line-2 text-ink-2 hover:border-line-3')}
                  >
                    {platformLabel(item, t)}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <div className="mb-2 flex items-center justify-between">
                <div className="input-label mb-0">{t('modelHub.groupFilterLabel')}</div>
                {hasFilters && (
                  <button
                    type="button"
                    className="text-xs font-medium text-orange hover:text-orange-hover"
                    onClick={() => {
                      setSearch('')
                      setPlatform(ALL)
                      setActiveGroupId(ALL)
                    }}
                  >
                    {t('modelHub.clearFilters')}
                  </button>
                )}
              </div>

              <div className="space-y-2">
                <button
                  type="button"
                  onClick={() => setActiveGroupId(ALL)}
                  className={cn(
                    'w-full rounded-xl border px-3 py-3 text-left transition-colors',
                    activeGroupId === ALL ? 'border-ink-1 bg-ink-1 text-bg-0' : 'border-line-2 bg-bg-1 text-ink-2 hover:bg-bg-2'
                  )}
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-sm font-semibold">{t('modelHub.allGroups')}</span>
                    <span className="font-mono text-xs">{catalogs.length}</span>
                  </div>
                  <div className={cn('mt-1 text-xs', activeGroupId === ALL ? 'text-bg-0/70' : 'text-ink-3')}>
                    {t('modelHub.modelCount', { count: visibleModelIds.length })}
                  </div>
                </button>

                {catalogs.map((catalog) => (
                  <button
                    key={catalog.group.id}
                    type="button"
                    onClick={() => setActiveGroupId(catalog.group.id)}
                    className={cn(
                      'w-full rounded-xl border px-3 py-3 text-left transition-colors',
                      activeGroupId === catalog.group.id ? 'border-orange bg-orange-soft' : 'border-line-2 bg-bg-1 hover:bg-bg-2'
                    )}
                  >
                    <div className="flex items-center justify-between gap-3">
                      <span className="truncate text-sm font-semibold text-ink-1">{catalog.group.name}</span>
                      <span className="font-mono text-xs text-ink-3">{catalog.models.length}</span>
                    </div>
                    <div className="mt-1 truncate text-xs text-ink-3">{platformLabel(catalog.group.platform, t)}</div>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </Card>

        <section className="space-y-4">
          {loading ? (
            Array.from({ length: 3 }).map((_, index) => (
              <Card key={index} className="p-5 space-y-4">
                <Skeleton className="h-5 w-48" />
                <Skeleton className="h-20 w-full" />
              </Card>
            ))
          ) : channelsQuery.isError ? (
            <Card className="p-6">
              <div className="text-sm font-semibold text-ink-1">{t('modelHub.loadFailedTitle')}</div>
              <p className="mt-1 text-sm text-ink-3">{t('modelHub.loadFailedDescription')}</p>
            </Card>
          ) : filteredCatalogs.length === 0 ? (
            <Card className="p-12 text-center">
              <div className="text-base font-semibold text-ink-1">{t('modelHub.emptyTitle')}</div>
              <p className="mt-1 text-sm text-ink-3">{t('modelHub.emptyDescription')}</p>
            </Card>
          ) : (
            filteredCatalogs.map((catalog) => (
              <Card key={catalog.group.id} className="overflow-hidden">
                <div className="border-b border-line-1 p-5">
                  <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <h2 className="truncate text-base font-semibold text-ink-1">{catalog.group.name}</h2>
                        <Badge tone="accent">{platformLabel(catalog.group.platform, t)}</Badge>
                        {catalog.group.subscription_type === 'subscription' && <Badge tone="success">订阅</Badge>}
                        <Badge>倍率 {formatRate(catalog.effectiveRate)}x</Badge>
                      </div>
                      <div className="mt-2 flex flex-wrap gap-2 text-xs text-ink-3">
                        {catalog.channels.map((channel) => (
                          <span key={channel} className="rounded-full bg-bg-2 px-2 py-1">{channel}</span>
                        ))}
                      </div>
                    </div>

                    <Button type="button" variant="ghost" size="sm" disabled={!catalog.models.length} onClick={() => copyGroup(catalog)}>
                      {copiedKey === `group:${catalog.group.id}` ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
                      {t('modelHub.copyGroup')}
                    </Button>
                  </div>
                </div>

                <div className="grid gap-3 p-5 md:grid-cols-2 2xl:grid-cols-3">
                  {catalog.models.length === 0 ? (
                    <div className="rounded-xl border border-dashed border-line-2 bg-bg-2 p-8 text-center text-sm text-ink-3 md:col-span-2 2xl:col-span-3">
                      {t('modelHub.noModelsInGroup')}
                    </div>
                  ) : (
                    catalog.models.map((model) => {
                      const badges = pricingBadges(model, catalog.effectiveRate)
                      return (
                        <button
                          key={`${catalog.group.id}-${model.id}`}
                          type="button"
                          onClick={() => copyText(model.id, `model:${catalog.group.id}:${model.id}`)}
                          className="group rounded-xl border border-line-1 bg-bg-1 p-4 text-left transition-all hover:-translate-y-0.5 hover:border-line-3 hover:bg-bg-2"
                        >
                          <div className="flex items-start justify-between gap-3">
                            <div className="min-w-0">
                              <div className="truncate font-mono text-sm font-semibold text-ink-1">{model.id}</div>
                              <div className="mt-1 text-xs text-ink-3">{badges.length > 0 ? '可见价格已按倍率换算' : '暂无价格配置'}</div>
                            </div>
                            <Copy className="mt-0.5 h-4 w-4 shrink-0 text-ink-4 transition-colors group-hover:text-orange" />
                          </div>
                          {badges.length > 0 && (
                            <div className="mt-3 flex flex-wrap gap-1.5">
                              {badges.map((badge) => (
                                <Badge key={badge.key} tone={badge.tone} className="h-auto min-h-[22px] py-0.5">
                                  {badge.text}
                                </Badge>
                              ))}
                            </div>
                          )}
                        </button>
                      )
                    })
                  )}
                </div>
              </Card>
            ))
          )}
        </section>
      </div>
    </>
  )
}
