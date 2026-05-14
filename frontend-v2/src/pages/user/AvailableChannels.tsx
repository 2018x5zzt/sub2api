import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { Boxes, Layers, RefreshCw, Search, Server, SlidersHorizontal } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import {
  channelsAPI,
  type AvailableChannelEntry,
  type AvailableChannelModel,
  type AvailableChannelPricing,
  type AvailableChannelSection
} from '@/api/channels'
import { cn } from '@/lib/cn'

const PER_MILLION_TOKENS = 1_000_000

function formatCount(value: number) {
  return value.toLocaleString('zh-CN')
}

function formatMoney(value: number) {
  if (!Number.isFinite(value)) return '-'
  if (value === 0) return '$0'
  if (Math.abs(value) >= 1) return `$${value.toFixed(2)}`
  return `$${value.toFixed(4)}`
}

function formatRate(value?: number | null) {
  const rate = Number(value ?? 1)
  if (!Number.isFinite(rate)) return '1x'
  return `${rate.toFixed(2).replace(/\.?0+$/, '')}x`
}

function platformName(value: string) {
  const map: Record<string, string> = {
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    gemini: 'Gemini',
    antigravity: 'Antigravity'
  }
  return map[value] || value
}

function billingModeLabel(value?: string | null) {
  if (!value) return '按量计费'
  const map: Record<string, string> = {
    token: '按 Token',
    request: '按请求',
    image: '图片计费',
    fixed: '固定价格'
  }
  return map[value] || value
}

function perMillion(price?: number | null) {
  if (price === null || price === undefined) return null
  return `${formatMoney(Number(price) * PER_MILLION_TOKENS)} / 1M`
}

function priceItems(pricing?: AvailableChannelPricing | null) {
  if (!pricing) return []
  const items = [
    { label: '输入', value: perMillion(pricing.input_price), tone: 'success' },
    { label: '输出', value: perMillion(pricing.output_price), tone: 'warning' },
    { label: '缓存写入', value: perMillion(pricing.cache_write_price), tone: 'neutral' },
    { label: '缓存读取', value: perMillion(pricing.cache_read_price), tone: 'neutral' },
    { label: '图片输出', value: perMillion(pricing.image_output_price), tone: 'accent' },
    {
      label: '每次请求',
      value: pricing.per_request_price === null || pricing.per_request_price === undefined
        ? null
        : formatMoney(Number(pricing.per_request_price)),
      tone: 'neutral'
    }
  ] as const

  return items.filter((item) => item.value !== null)
}

function totalModels(entry: AvailableChannelEntry) {
  return (entry.platforms ?? []).reduce((sum, section) => sum + (section.supported_models ?? []).length, 0)
}

function totalGroups(entry: AvailableChannelEntry) {
  return (entry.platforms ?? []).reduce((sum, section) => sum + (section.groups ?? []).length, 0)
}

function uniquePlatforms(entries: AvailableChannelEntry[]) {
  return new Set(entries.flatMap((entry) => (entry.platforms ?? []).map((section) => section.platform))).size
}

function SummaryStat({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="rounded-xl border border-line-2 bg-bg-1 p-4">
      <div className="flex items-center gap-2 text-xs font-semibold text-ink-3">
        <span className="grid h-7 w-7 place-items-center rounded-lg bg-orange-soft text-orange">{icon}</span>
        {label}
      </div>
      <div className="mt-3 text-2xl font-semibold tracking-normal text-ink-1">{value}</div>
    </div>
  )
}

function ModelCard({ model }: { model: AvailableChannelModel }) {
  const items = priceItems(model.pricing)
  return (
    <div className="rounded-xl border border-line-1 bg-bg-1 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="truncate font-mono text-sm font-semibold text-ink-1">{model.name}</div>
          <div className="mt-1 text-xs text-ink-3">{billingModeLabel(model.pricing?.billing_mode)}</div>
        </div>
        <Badge tone="neutral">{platformName(model.platform)}</Badge>
      </div>

      {items.length > 0 ? (
        <div className="mt-4 grid gap-2 sm:grid-cols-2">
          {items.map((item) => (
            <div key={item.label} className="rounded-lg bg-bg-2 px-3 py-2">
              <div className="text-[11px] text-ink-3">{item.label}</div>
              <div className="mt-1 font-mono text-xs text-ink-1">{item.value}</div>
            </div>
          ))}
        </div>
      ) : (
        <div className="mt-4 rounded-lg border border-dashed border-line-2 bg-bg-2 px-3 py-4 text-center text-sm text-ink-3">
          暂无可见价格
        </div>
      )}
    </div>
  )
}

function SectionBlock({ section }: { section: AvailableChannelSection }) {
  return (
    <div className="rounded-xl border border-line-2 bg-bg-2 p-4">
      <div className="flex flex-col gap-3 border-b border-line-2 pb-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-wrap items-center gap-2">
          <Badge tone="accent">{platformName(section.platform)}</Badge>
          {(section.groups ?? []).slice(0, 6).map((group) => (
            <Badge key={group.id}>
              {group.name}
              {group.rate_multiplier ? ` · ${formatRate(group.rate_multiplier)}` : ''}
            </Badge>
          ))}
          {(section.groups ?? []).length > 6 && <Badge>+{(section.groups ?? []).length - 6} 个分组</Badge>}
        </div>
        <div className="text-xs text-ink-3">
          {formatCount((section.supported_models ?? []).length)} 个模型 · {formatCount((section.groups ?? []).length)} 个分组
        </div>
      </div>

      {(section.supported_models ?? []).length > 0 ? (
        <div className="mt-4 grid gap-3 xl:grid-cols-2 2xl:grid-cols-3">
          {section.supported_models.map((model) => (
            <ModelCard key={`${section.platform}-${model.name}`} model={model} />
          ))}
        </div>
      ) : (
        <div className="mt-4 rounded-lg border border-dashed border-line-2 bg-bg-1 p-8 text-center text-sm text-ink-3">
          暂无模型
        </div>
      )}
    </div>
  )
}

function ChannelCard({ entry }: { entry: AvailableChannelEntry }) {
  const platforms = entry.platforms ?? []
  return (
    <Card className="overflow-hidden">
      <div className="grid gap-4 border-b border-line-1 p-5 lg:grid-cols-[1fr_auto] lg:items-start">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="truncate text-lg font-semibold text-ink-1">{entry.name}</h2>
            <Badge tone="success" dot>可用</Badge>
          </div>
          {entry.description && <p className="mt-2 text-sm leading-6 text-ink-3">{entry.description}</p>}
        </div>
        <div className="grid grid-cols-3 gap-2 lg:min-w-[280px]">
          <div className="rounded-lg bg-bg-2 px-3 py-2">
            <div className="text-[11px] text-ink-3">平台</div>
            <div className="mt-1 font-semibold text-ink-1">{formatCount(platforms.length)}</div>
          </div>
          <div className="rounded-lg bg-bg-2 px-3 py-2">
            <div className="text-[11px] text-ink-3">分组</div>
            <div className="mt-1 font-semibold text-ink-1">{formatCount(totalGroups(entry))}</div>
          </div>
          <div className="rounded-lg bg-bg-2 px-3 py-2">
            <div className="text-[11px] text-ink-3">模型</div>
            <div className="mt-1 font-semibold text-ink-1">{formatCount(totalModels(entry))}</div>
          </div>
        </div>
      </div>

      <div className="space-y-4 p-5">
        {platforms.map((section) => (
          <SectionBlock key={section.platform} section={section} />
        ))}
      </div>
    </Card>
  )
}

export default function AvailableChannelsPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')

  const query = useQuery({
    queryKey: ['available-channels'],
    queryFn: () => channelsAPI.listAvailableChannels()
  })

  const rows = useMemo(() => {
    const q = search.trim().toLowerCase()
    const data = query.data ?? []
    if (!q) return data
    return data
      .map((entry) => {
        const entryHit = `${entry.name} ${entry.description}`.toLowerCase().includes(q)
        if (entryHit) return entry
        const platforms = (entry.platforms ?? []).filter((section) => {
          return (
            section.platform.toLowerCase().includes(q) ||
            (section.groups ?? []).some((group) => group.name.toLowerCase().includes(q)) ||
            (section.supported_models ?? []).some((model) => model.name.toLowerCase().includes(q))
          )
        })
        return platforms.length > 0 ? { ...entry, platforms } : null
      })
      .filter((item): item is AvailableChannelEntry => !!item)
  }, [query.data, search])

  const allRows = query.data ?? []
  const totalChannelCount = allRows.length
  const totalGroupCount = allRows.reduce((sum, entry) => sum + totalGroups(entry), 0)
  const totalModelCount = allRows.reduce((sum, entry) => sum + totalModels(entry), 0)
  const visibleModelCount = rows.reduce((sum, entry) => sum + totalModels(entry), 0)

  return (
    <>
      <PageHeader
        title={t('nav.availableChannels')}
        description="查看当前账号可见的渠道、平台分组、模型与公开价格。"
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={cn('h-4 w-4', query.isFetching && 'animate-spin')} />
            {t('common.refresh')}
          </Button>
        }
      />

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <Card className="p-4">
          <Input
            name="search"
            label="搜索"
            placeholder="搜索渠道、分组、平台或模型"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            leftIcon={<Search className="h-4 w-4" />}
          />
        </Card>
        <Card className="p-4">
          <div className="flex items-center gap-3">
            <span className="grid h-10 w-10 place-items-center rounded-xl bg-orange-soft text-orange">
              <SlidersHorizontal className="h-5 w-5" />
            </span>
            <div>
              <div className="text-sm font-semibold text-ink-1">筛选结果</div>
              <div className="text-xs text-ink-3">
                当前展示 {formatCount(rows.length)} 个渠道，{formatCount(visibleModelCount)} 个模型
              </div>
            </div>
          </div>
        </Card>
      </div>

      <div className="mt-4 grid gap-3 md:grid-cols-4">
        <SummaryStat icon={<Server className="h-4 w-4" />} label="可用渠道" value={formatCount(totalChannelCount)} />
        <SummaryStat icon={<Layers className="h-4 w-4" />} label="平台" value={formatCount(uniquePlatforms(allRows))} />
        <SummaryStat icon={<Boxes className="h-4 w-4" />} label="分组" value={formatCount(totalGroupCount)} />
        <SummaryStat icon={<Search className="h-4 w-4" />} label="可见模型" value={formatCount(totalModelCount)} />
      </div>

      {query.isLoading ? (
        <div className="mt-4 grid gap-4">
          {Array.from({ length: 3 }).map((_, index) => (
            <Card key={index} className="space-y-4 p-5">
              <Skeleton className="h-5 w-44" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-32 w-full" />
            </Card>
          ))}
        </div>
      ) : rows.length === 0 ? (
        <Card className="mt-4 p-12 text-center">
          <div className="text-base font-semibold text-ink-1">没有匹配的渠道</div>
          <p className="mt-1 text-sm text-ink-3">请尝试换一个关键词，或确认管理员已启用可用渠道展示。</p>
        </Card>
      ) : (
        <div className="mt-4 space-y-4">
          {rows.map((entry) => (
            <ChannelCard key={entry.name} entry={entry} />
          ))}
        </div>
      )}
    </>
  )
}
