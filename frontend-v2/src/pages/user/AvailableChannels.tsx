import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { RefreshCw, Search } from 'lucide-react'
import { PageHeader } from '@/components/layout/ConsoleLayout'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { channelsAPI, type AvailableChannelEntry } from '@/api/channels'

function priceText(v: unknown) {
  if (v == null || v === '') return null
  const n = Number(v)
  if (!Number.isFinite(n)) return String(v)
  if (n === 0) return '0'
  return n < 0.01 ? `$${n.toFixed(6)}` : `$${n.toFixed(4)}`
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
            (section.groups ?? []).some((g) => g.name.toLowerCase().includes(q)) ||
            (section.supported_models ?? []).some((m) => m.name.toLowerCase().includes(q))
          )
        })
        return platforms.length > 0 ? { ...entry, platforms } : null
      })
      .filter((x): x is AvailableChannelEntry => !!x)
  }, [query.data, search])

  return (
    <>
      <PageHeader
        title={t('nav.availableChannels')}
        description="Browse enabled channels, groups, models, and visible pricing."
        actions={
          <Button variant="ghost" onClick={() => query.refetch()} disabled={query.isFetching}>
            <RefreshCw className={`h-4 w-4 ${query.isFetching ? 'animate-spin' : ''}`} />
            {t('common.refresh')}
          </Button>
        }
      />

      <Card className="p-4 mb-4">
        <Input
          name="search"
          placeholder="Search channel, group, platform, or model"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          leftIcon={<Search className="h-4 w-4" />}
        />
      </Card>

      {query.isLoading ? (
        <div className="grid gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <Card key={i} className="p-5 space-y-3">
              <Skeleton className="h-5 w-44" />
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-20 w-full" />
            </Card>
          ))}
        </div>
      ) : rows.length === 0 ? (
        <Card className="p-12 text-center text-ink-3">{t('common.noData')}</Card>
      ) : (
        <div className="space-y-4">
          {rows.map((entry) => (
            <Card key={entry.name} className="p-5">
              <div className="mb-4">
                <h2 className="text-base font-medium text-ink-1">{entry.name}</h2>
                {entry.description && <p className="mt-1 text-sm text-ink-3">{entry.description}</p>}
              </div>
              <div className="space-y-4">
                {(entry.platforms ?? []).map((section) => (
                  <div key={section.platform} className="rounded-lg border border-line-2 bg-bg-2 p-4">
                    <div className="flex flex-wrap items-center gap-2 mb-3">
                      <Badge tone="accent">{section.platform}</Badge>
                      {(section.groups ?? []).map((g) => (
                        <Badge key={g.id}>{g.name}</Badge>
                      ))}
                    </div>
                    {(section.supported_models ?? []).length > 0 ? (
                      <div className="grid gap-2 lg:grid-cols-2">
                        {section.supported_models.map((m) => (
                          <div key={`${section.platform}-${m.name}`} className="rounded-md border border-line-1 bg-bg-1 p-3">
                            <div className="font-mono text-sm text-ink-1 truncate">{m.name}</div>
                            {m.pricing && (
                              <div className="mt-2 flex flex-wrap gap-2 text-xs text-ink-3">
                                {Object.entries(m.pricing).map(([k, v]) => {
                                  const txt = priceText(v)
                                  return txt ? <span key={k} className="font-mono">{k}: {txt}</span> : null
                                })}
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-sm text-ink-3">{t('common.noData')}</div>
                    )}
                  </div>
                ))}
              </div>
            </Card>
          ))}
        </div>
      )}
    </>
  )
}
