<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="relative overflow-hidden rounded-[28px] border border-slate-200 bg-[radial-gradient(circle_at_top_left,_rgba(14,165,233,0.16),_transparent_28%),linear-gradient(135deg,_#f8fafc_0%,_#ecfeff_48%,_#fefce8_100%)] p-6 shadow-sm dark:border-dark-700 dark:bg-[radial-gradient(circle_at_top_left,_rgba(56,189,248,0.12),_transparent_30%),linear-gradient(135deg,_rgba(15,23,42,0.96)_0%,_rgba(17,24,39,0.98)_45%,_rgba(23,37,84,0.94)_100%)]">
        <div class="absolute right-0 top-0 h-40 w-40 -translate-y-8 translate-x-10 rounded-full bg-amber-200/40 blur-3xl dark:bg-amber-400/10"></div>
        <div class="absolute bottom-0 left-0 h-32 w-32 -translate-x-8 translate-y-8 rounded-full bg-sky-200/40 blur-3xl dark:bg-sky-400/10"></div>
        <div class="relative space-y-5">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div class="space-y-2">
              <div class="inline-flex items-center gap-2 rounded-full border border-white/70 bg-white/80 px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em] text-sky-700 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5 dark:text-sky-300">
                {{ t('modelHub.eyebrow') }}
              </div>
              <div>
                <h1 class="text-2xl font-semibold text-slate-900 dark:text-white">
                  {{ t('modelHub.title') }}
                </h1>
                <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600 dark:text-slate-300">
                  {{ t('modelHub.description') }}
                </p>
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-3">
              <button
                class="btn btn-secondary"
                :disabled="loading"
                @click="loadCatalogs"
              >
                <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" class="mr-2" />
                {{ t('common.refresh') }}
              </button>
              <button
                class="btn btn-primary"
                :disabled="visibleModelIds.length === 0"
                @click="copyVisibleModels"
              >
                <Icon :name="copiedKey === 'visible' ? 'check' : 'clipboard'" size="md" class="mr-2" />
                {{ t('modelHub.copyVisible') }}
              </button>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <div class="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5">
              <div class="text-xs uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
                {{ t('modelHub.groupsLabel') }}
              </div>
              <div class="mt-2 text-2xl font-semibold text-slate-900 dark:text-white">
                {{ catalogs.length }}
              </div>
            </div>
            <div class="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5">
              <div class="text-xs uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
                {{ t('modelHub.uniqueModelsLabel') }}
              </div>
              <div class="mt-2 text-2xl font-semibold text-slate-900 dark:text-white">
                {{ allModelIds.length }}
              </div>
            </div>
            <div class="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5">
              <div class="text-xs uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
                {{ t('modelHub.visibleModelsLabel') }}
              </div>
              <div class="mt-2 text-2xl font-semibold text-slate-900 dark:text-white">
                {{ visibleModelIds.length }}
              </div>
            </div>
            <div class="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm backdrop-blur dark:border-white/10 dark:bg-white/5">
              <div class="text-xs uppercase tracking-[0.18em] text-slate-500 dark:text-slate-400">
                {{ t('modelHub.platformsLabel') }}
              </div>
              <div class="mt-2 text-2xl font-semibold text-slate-900 dark:text-white">
                {{ platformOptions.length }}
              </div>
            </div>
          </div>
        </div>
      </section>

      <div class="grid gap-6 xl:grid-cols-[300px,minmax(0,1fr)]">
        <aside class="card p-4">
          <div class="space-y-5">
            <div>
              <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('modelHub.searchLabel') }}
              </label>
              <div class="relative">
                <Icon
                  name="search"
                  size="md"
                  class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
                />
                <input
                  v-model="searchQuery"
                  type="text"
                  class="input pl-10"
                  :placeholder="t('modelHub.searchPlaceholder')"
                />
              </div>
            </div>

            <div>
              <div class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('modelHub.platformFilterLabel') }}
              </div>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="option in platformFilterOptions"
                  :key="option.value"
                  type="button"
                  :class="[
                    'rounded-full border px-3 py-1.5 text-xs font-medium transition-colors',
                    platformFilter === option.value
                      ? 'border-sky-500 bg-sky-500 text-white shadow-sm'
                      : 'border-gray-200 bg-white text-gray-600 hover:border-sky-300 hover:text-sky-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:border-sky-500/60 dark:hover:text-sky-300'
                  ]"
                  @click="platformFilter = option.value"
                >
                  {{ option.label }}
                </button>
              </div>
            </div>

            <div>
              <div class="mb-2 flex items-center justify-between text-sm font-medium text-gray-700 dark:text-gray-300">
                <span>{{ t('modelHub.groupFilterLabel') }}</span>
                <button
                  v-if="hasActiveFilters"
                  type="button"
                  class="text-xs font-medium text-sky-600 transition-colors hover:text-sky-700 dark:text-sky-400 dark:hover:text-sky-300"
                  @click="clearFilters"
                >
                  {{ t('modelHub.clearFilters') }}
                </button>
              </div>

              <div class="space-y-2">
                <button
                  type="button"
                  :class="[
                    'flex w-full items-center justify-between rounded-2xl border px-3 py-3 text-left transition-colors',
                    activeGroupId === 'all'
                      ? 'border-slate-900 bg-slate-900 text-white shadow-sm dark:border-slate-100 dark:bg-slate-100 dark:text-slate-900'
                      : 'border-gray-200 bg-white hover:border-slate-300 hover:bg-slate-50 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500 dark:hover:bg-dark-700'
                  ]"
                  @click="activeGroupId = 'all'"
                >
                  <div>
                    <div class="text-sm font-semibold">{{ t('modelHub.allGroups') }}</div>
                    <div
                      class="mt-1 text-xs"
                      :class="activeGroupId === 'all' ? 'text-white/80 dark:text-slate-600' : 'text-gray-500 dark:text-gray-400'"
                    >
                      {{ t('modelHub.modelCount', { count: visibleModelIds.length }) }}
                    </div>
                  </div>
                  <span
                    class="rounded-full px-2 py-1 text-xs font-semibold"
                    :class="activeGroupId === 'all' ? 'bg-white/15 text-white dark:bg-slate-900/10 dark:text-slate-700' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'"
                  >
                    {{ platformFilteredCatalogs.length }}
                  </span>
                </button>

                <button
                  v-for="catalog in platformFilteredCatalogs"
                  :key="catalog.group.id"
                  type="button"
                  :class="[
                    'flex w-full items-center justify-between rounded-2xl border px-3 py-3 text-left transition-colors',
                    activeGroupId === catalog.group.id
                      ? 'border-sky-500 bg-sky-50 shadow-sm dark:border-sky-500/60 dark:bg-sky-500/10'
                      : 'border-gray-200 bg-white hover:border-slate-300 hover:bg-slate-50 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500 dark:hover:bg-dark-700'
                  ]"
                  @click="activeGroupId = catalog.group.id"
                >
                  <div class="min-w-0 pr-3">
                    <div class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                      {{ catalog.group.name }}
                    </div>
                    <div class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">
                      {{ getPlatformLabel(catalog.group.platform) }}
                    </div>
                  </div>
                  <span class="rounded-full bg-gray-100 px-2 py-1 text-xs font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                    {{ catalog.models.length }}
                  </span>
                </button>
              </div>
            </div>
          </div>
        </aside>

        <section class="space-y-4">
          <div v-if="loading" class="card flex items-center justify-center py-20">
            <LoadingSpinner />
          </div>

          <div v-else-if="errorMessage" class="card space-y-4 p-6">
            <div class="flex items-start gap-3 rounded-2xl border border-red-200 bg-red-50 p-4 text-red-700 dark:border-red-900/40 dark:bg-red-900/10 dark:text-red-300">
              <Icon name="exclamationTriangle" size="md" class="mt-0.5 shrink-0" />
              <div>
                <div class="text-sm font-semibold">{{ t('modelHub.loadFailedTitle') }}</div>
                <div class="mt-1 text-sm">{{ errorMessage }}</div>
              </div>
            </div>
            <div>
              <button class="btn btn-primary" @click="loadCatalogs">
                {{ t('common.refresh') }}
              </button>
            </div>
          </div>

          <EmptyState
            v-else-if="visibleCatalogs.length === 0"
            :title="t('modelHub.emptyTitle')"
            :description="t('modelHub.emptyDescription')"
          >
            <template #action>
              <button class="btn btn-secondary" @click="clearFilters">
                {{ t('modelHub.clearFilters') }}
              </button>
            </template>
          </EmptyState>

          <div v-else class="space-y-4">
            <article
              v-for="catalog in visibleCatalogs"
              :key="catalog.group.id"
              class="card overflow-hidden"
            >
              <div class="border-b border-gray-100 p-5 dark:border-dark-700">
                <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                  <div class="min-w-0 space-y-3">
                    <div class="flex flex-wrap items-center gap-3">
                      <GroupBadge
                        :name="catalog.group.name"
                        :platform="catalog.group.platform"
                        :subscription-type="catalog.group.subscription_type"
                        :rate-multiplier="catalog.group.rate_multiplier"
                        :user-rate-multiplier="catalog.user_rate_multiplier ?? null"
                      />
                      <span class="rounded-full bg-gray-100 px-2.5 py-1 text-xs font-semibold text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                        {{ getSourceLabel(catalog.source) }}
                      </span>
                      <span class="rounded-full bg-sky-50 px-2.5 py-1 text-xs font-semibold text-sky-700 dark:bg-sky-500/10 dark:text-sky-300">
                        {{ t('modelHub.modelCount', { count: catalog.models.length }) }}
                      </span>
                    </div>
                    <p
                      v-if="catalog.group.description"
                      class="text-sm leading-6 text-gray-600 dark:text-gray-300"
                    >
                      {{ catalog.group.description }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-gray-400">
                      {{ t('modelHub.pricingComputedWithRate', { rate: formatRateMultiplier(catalog.effective_rate_multiplier) }) }}
                    </p>
                  </div>

                  <div class="flex flex-wrap items-center gap-3">
                    <button
                      class="btn btn-secondary"
                      :disabled="catalog.models.length === 0"
                      @click="copyGroupModels(catalog)"
                    >
                      <Icon
                        :name="copiedKey === `group:${catalog.group.id}` ? 'check' : 'clipboard'"
                        size="md"
                        class="mr-2"
                      />
                      {{ t('modelHub.copyGroup') }}
                    </button>
                  </div>
                </div>
              </div>

              <div class="p-5">
                <div
                  v-if="catalog.models.length === 0"
                  class="rounded-2xl border border-dashed border-gray-200 bg-gray-50 px-4 py-8 text-center text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-400"
                >
                  {{ t('modelHub.noModelsInGroup') }}
                </div>

                <div v-else class="grid gap-3 md:grid-cols-2 2xl:grid-cols-3">
                  <button
                    v-for="model in catalog.models"
                    :key="`${catalog.group.id}-${model.id}`"
                    type="button"
                    class="group flex items-start justify-between rounded-2xl border border-gray-200 bg-white px-4 py-3 text-left transition-all hover:-translate-y-0.5 hover:border-sky-300 hover:shadow-sm dark:border-dark-600 dark:bg-dark-800 dark:hover:border-sky-500/60"
                    @click="copyModel(model.id)"
                  >
                    <div class="flex min-w-0 items-start gap-3">
                      <div class="mt-0.5 flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-slate-100 text-slate-700 dark:bg-dark-700 dark:text-slate-200">
                        <ModelIcon :model="model.id" size="20px" />
                      </div>
                      <div class="min-w-0">
                        <div class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                          {{ model.display_name }}
                        </div>
                        <code class="mt-1 block truncate text-xs text-gray-500 dark:text-gray-400">
                          {{ model.id }}
                        </code>
                        <div class="mt-2 flex flex-wrap gap-1.5">
                          <span
                            v-for="badge in getPricingBadges(model, catalog.effective_rate_multiplier)"
                            :key="badge.key"
                            :class="pricingBadgeClass(badge.tone)"
                          >
                            {{ badge.text }}
                          </span>
                        </div>
                        <div
                          v-if="!hasDisplayedPricing(model)"
                          class="mt-2 text-[11px] text-gray-400 dark:text-gray-500"
                        >
                          {{ t('modelHub.pricingUnavailable') }}
                        </div>
                      </div>
                    </div>

                    <div
                      class="ml-3 mt-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-transparent text-gray-400 transition-colors group-hover:border-sky-200 group-hover:text-sky-600 dark:group-hover:border-sky-500/30 dark:group-hover:text-sky-300"
                    >
                      <Icon
                        :name="copiedKey === `model:${model.id}` ? 'check' : 'clipboard'"
                        size="sm"
                      />
                    </div>
                  </button>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { userGroupsAPI } from '@/api'
import userChannelsAPI, {
  type UserAvailableChannel,
  type UserAvailableGroup,
  type UserPricingInterval,
  type UserSupportedModel,
  type UserSupportedModelPricing,
} from '@/api/channels'
import type { GroupPlatform, SubscriptionType } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency } from '@/utils/format'

type PlatformFilter = GroupPlatform | 'all'
type GroupFilter = number | 'all'

type CatalogSource = 'default' | 'mapping' | 'mixed'

interface SupportedModel {
  id: string
  display_name: string
  pricing: UserSupportedModelPricing | null
}

interface GroupModelCatalog {
  group: {
    id: number
    name: string
    description: string | null
    platform: GroupPlatform
    subscription_type: SubscriptionType
    rate_multiplier: number
  }
  source: CatalogSource
  user_rate_multiplier: number | null
  effective_rate_multiplier: number
  models: SupportedModel[]
}

interface PlatformOption {
  value: PlatformFilter
  label: string
}

interface PricingBadge {
  key: string
  text: string
  tone: 'rate' | 'input' | 'output' | 'request' | 'interval'
}

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const catalogs = ref<GroupModelCatalog[]>([])
const loading = ref(false)
const errorMessage = ref('')
const searchQuery = ref('')
const platformFilter = ref<PlatformFilter>('all')
const activeGroupId = ref<GroupFilter>('all')
const copiedKey = ref<string | null>(null)

const platformOptions = computed<GroupPlatform[]>(() => {
  const seen = new Set<GroupPlatform>()
  const options: GroupPlatform[] = []
  for (const catalog of catalogs.value) {
    if (!seen.has(catalog.group.platform)) {
      seen.add(catalog.group.platform)
      options.push(catalog.group.platform)
    }
  }
  return options
})

const platformFilterOptions = computed<PlatformOption[]>(() => [
  { value: 'all', label: t('modelHub.allPlatforms') },
  ...platformOptions.value.map((platform) => ({
    value: platform,
    label: getPlatformLabel(platform)
  }))
])

const platformFilteredCatalogs = computed(() => {
  if (platformFilter.value === 'all') {
    return catalogs.value
  }
  return catalogs.value.filter((catalog) => catalog.group.platform === platformFilter.value)
})

watch(
  platformFilteredCatalogs,
  (items) => {
    if (activeGroupId.value === 'all') {
      return
    }
    const exists = items.some((catalog) => catalog.group.id === activeGroupId.value)
    if (!exists) {
      activeGroupId.value = 'all'
    }
  },
  { immediate: true }
)

const visibleCatalogs = computed<GroupModelCatalog[]>(() => {
  const query = searchQuery.value.trim().toLowerCase()
  const scopedCatalogs =
    activeGroupId.value === 'all'
      ? platformFilteredCatalogs.value
      : platformFilteredCatalogs.value.filter((catalog) => catalog.group.id === activeGroupId.value)

  if (!query) {
    return scopedCatalogs
  }

  return scopedCatalogs
    .map((catalog) => {
      const groupMatches =
        catalog.group.name.toLowerCase().includes(query) ||
        (catalog.group.description || '').toLowerCase().includes(query) ||
        getPlatformLabel(catalog.group.platform).toLowerCase().includes(query)

      const models = groupMatches
        ? catalog.models
        : catalog.models.filter((model) => {
            const displayName = model.display_name || model.id
            return (
              model.id.toLowerCase().includes(query) ||
              displayName.toLowerCase().includes(query)
            )
          })

      if (!groupMatches && models.length === 0) {
        return null
      }

      return {
        ...catalog,
        models
      }
    })
    .filter((catalog): catalog is GroupModelCatalog => catalog !== null)
})

const allModelIds = computed(() => collectUniqueModelIds(catalogs.value))
const visibleModelIds = computed(() => collectUniqueModelIds(visibleCatalogs.value))

const hasActiveFilters = computed(() => {
  return (
    searchQuery.value.trim().length > 0 ||
    platformFilter.value !== 'all' ||
    activeGroupId.value !== 'all'
  )
})

function getPlatformLabel(platform: GroupPlatform): string {
  return t(`admin.groups.platforms.${platform}`)
}

function getSourceLabel(source: CatalogSource): string {
  if (source === 'mapping') {
    return t('modelHub.sourceMapping')
  }
  if (source === 'mixed') {
    return t('modelHub.sourceMixed')
  }
  return t('modelHub.sourceDefault')
}

function formatRateMultiplier(rate: number): string {
  if (!Number.isFinite(rate)) {
    return '1'
  }
  return rate.toFixed(2).replace(/\.?0+$/, '')
}

function formatPerMillionPrice(price?: number | null): string {
  if (price === null || price === undefined) {
    return t('modelHub.pricingUnavailable')
  }
  return `${formatCurrency(price)} ${t('modelHub.perMillionTokens')}`
}

function formatPerRequestPrice(price?: number | null, billingMode?: string | null): string {
  if (price === null || price === undefined) {
    return t('modelHub.pricingUnavailable')
  }
  const unitKey = billingMode === 'image' ? 'modelHub.perImage' : 'modelHub.perRequest'
  return `${formatCurrency(price)} ${t(unitKey)}`
}

function formatCompactTokenCount(tokens?: number | null): string {
  if (tokens === null || tokens === undefined || !Number.isFinite(tokens)) {
    return '0'
  }
  if (tokens >= 1_000_000) {
    return `${(tokens / 1_000_000).toFixed(tokens % 1_000_000 === 0 ? 0 : 1)}M`
  }
  if (tokens >= 1_000) {
    return `${(tokens / 1_000).toFixed(tokens % 1_000 === 0 ? 0 : 1)}K`
  }
  return `${tokens}`
}

function formatTokenRange(minTokens?: number, maxTokens?: number | null): string {
  const lower = formatCompactTokenCount(minTokens ?? 0)
  if (maxTokens === null || maxTokens === undefined) {
    return `${lower}+`
  }
  return `${lower}-${formatCompactTokenCount(maxTokens)}`
}

function formatTokenIntervalBadge(interval: UserPricingInterval): string | null {
  const parts: string[] = []
  if (interval.input_price !== undefined && interval.input_price !== null) {
    parts.push(`${t('modelHub.inputPriceShort')} ${formatPerMillionPrice(interval.input_price)}`)
  }
  if (interval.output_price !== undefined && interval.output_price !== null) {
    parts.push(`${t('modelHub.outputPriceShort')} ${formatPerMillionPrice(interval.output_price)}`)
  }
  if (parts.length === 0) {
    return null
  }
  return `${formatTokenRange(interval.min_tokens, interval.max_tokens)} ${parts.join(' · ')}`
}

function formatRequestTierLabel(tier: UserPricingInterval): string {
  if (tier.tier_label) {
    return tier.tier_label
  }
  return formatTokenRange(tier.min_tokens ?? 0, tier.max_tokens ?? null)
}

function getPricingBadges(model: SupportedModel, rate: number): PricingBadge[] {
  const badges: PricingBadge[] = [
    {
      key: `rate:${model.id}`,
      text: `${t('modelHub.rateShort')} ${formatRateMultiplier(rate)}x`,
      tone: 'rate'
    }
  ]

  const pricing = model.pricing
  if (!pricing) {
    return badges
  }

  if (pricing.input_price !== undefined && pricing.input_price !== null) {
    badges.push({
      key: `input:${model.id}`,
      text: `${t('modelHub.inputPriceShort')} ${formatPerMillionPrice(pricing.input_price)}`,
      tone: 'input'
    })
  }
  if (pricing.output_price !== undefined && pricing.output_price !== null) {
    badges.push({
      key: `output:${model.id}`,
      text: `${t('modelHub.outputPriceShort')} ${formatPerMillionPrice(pricing.output_price)}`,
      tone: 'output'
    })
  }
  if (pricing.per_request_price !== undefined && pricing.per_request_price !== null) {
    badges.push({
      key: `request-default:${model.id}`,
      text: `${t('modelHub.defaultPriceShort')} ${formatPerRequestPrice(pricing.per_request_price, pricing.billing_mode)}`,
      tone: 'request'
    })
  }
  for (const [index, tier] of (pricing.intervals || []).entries()) {
    if (tier.per_request_price === undefined || tier.per_request_price === null) {
      continue
    }
    badges.push({
      key: `request-tier:${model.id}:${index}`,
      text: `${formatRequestTierLabel(tier)} ${formatPerRequestPrice(tier.per_request_price, pricing.billing_mode)}`,
      tone: 'request'
    })
  }
  for (const [index, interval] of (pricing.intervals || []).entries()) {
    const text = formatTokenIntervalBadge(interval)
    if (!text) {
      continue
    }
    badges.push({
      key: `token-interval:${model.id}:${index}`,
      text,
      tone: 'interval'
    })
  }

  return badges
}

function pricingBadgeClass(tone: PricingBadge['tone']): string {
  const base = 'inline-flex items-center rounded-full px-2.5 py-1 text-[11px] font-semibold'
  if (tone === 'rate') {
    return `${base} bg-slate-100 text-slate-700 dark:bg-dark-700 dark:text-slate-200`
  }
  if (tone === 'input') {
    return `${base} bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300`
  }
  if (tone === 'output') {
    return `${base} bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300`
  }
  if (tone === 'request') {
    return `${base} bg-sky-50 text-sky-700 dark:bg-sky-500/10 dark:text-sky-300`
  }
  return `${base} bg-violet-50 text-violet-700 dark:bg-violet-500/10 dark:text-violet-300`
}

function hasDisplayedPricing(model: SupportedModel): boolean {
  return Boolean(
      model.pricing &&
    (
      model.pricing.input_price !== undefined ||
      model.pricing.output_price !== undefined ||
      model.pricing.per_request_price !== undefined ||
      (model.pricing.intervals || []).some((interval) =>
        interval.input_price !== undefined ||
        interval.output_price !== undefined ||
        interval.per_request_price !== undefined
      )
    )
  )
}

function collectUniqueModelIds(groupCatalogs: GroupModelCatalog[]): string[] {
  const seen = new Set<string>()
  const ids: string[] = []
  for (const catalog of groupCatalogs) {
    for (const model of catalog.models) {
      if (seen.has(model.id)) {
        continue
      }
      seen.add(model.id)
      ids.push(model.id)
    }
  }
  return ids
}

function markCopied(key: string) {
  copiedKey.value = key
  window.setTimeout(() => {
    if (copiedKey.value === key) {
      copiedKey.value = null
    }
  }, 2000)
}

async function copyModel(modelId: string) {
  const copied = await copyToClipboard(modelId, t('modelHub.copiedModel'))
  if (copied) {
    markCopied(`model:${modelId}`)
  }
}

async function copyGroupModels(catalog: GroupModelCatalog) {
  const copied = await copyToClipboard(
    catalog.models.map((model) => model.id).join('\n'),
    t('modelHub.copiedGroup')
  )
  if (copied) {
    markCopied(`group:${catalog.group.id}`)
  }
}

async function copyVisibleModels() {
  const copied = await copyToClipboard(
    visibleModelIds.value.join('\n'),
    t('modelHub.copiedVisible')
  )
  if (copied) {
    markCopied('visible')
  }
}

function clearFilters() {
  searchQuery.value = ''
  platformFilter.value = 'all'
  activeGroupId.value = 'all'
}

async function loadCatalogs() {
  loading.value = true
  errorMessage.value = ''
  try {
    const [channels, rates] = await Promise.all([
      userChannelsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates().catch(() => ({} as Record<number, number>)),
    ])
    catalogs.value = buildCatalogs(channels, rates)
  } catch (error) {
    console.error('Failed to load model catalogs:', error)
    errorMessage.value = t('modelHub.loadFailedDescription')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCatalogs()
})

function buildCatalogs(channels: UserAvailableChannel[], rates: Record<number, number>): GroupModelCatalog[] {
  const byGroup = new Map<number, GroupModelCatalog>()
  const modelKeysByGroup = new Map<number, Set<string>>()

  for (const channel of channels) {
    for (const section of channel.platforms) {
      for (const group of section.groups) {
        const catalog = ensureCatalog(byGroup, modelKeysByGroup, group, rates)
        for (const model of section.supported_models) {
          appendModel(catalog, modelKeysByGroup.get(group.id)!, model)
        }
      }
    }
  }

  return Array.from(byGroup.values()).sort((a, b) => a.group.name.localeCompare(b.group.name))
}

function ensureCatalog(
  byGroup: Map<number, GroupModelCatalog>,
  modelKeysByGroup: Map<number, Set<string>>,
  group: UserAvailableGroup,
  rates: Record<number, number>,
): GroupModelCatalog {
  const existing = byGroup.get(group.id)
  if (existing) {
    return existing
  }

  const userRate = rates[group.id] ?? null
  const rate = userRate ?? group.rate_multiplier
  const catalog: GroupModelCatalog = {
    group: {
      id: group.id,
      name: group.name,
      description: null,
      platform: group.platform as GroupPlatform,
      subscription_type: (group.subscription_type || 'standard') as SubscriptionType,
      rate_multiplier: group.rate_multiplier,
    },
    source: 'default',
    user_rate_multiplier: userRate,
    effective_rate_multiplier: rate,
    models: [],
  }
  byGroup.set(group.id, catalog)
  modelKeysByGroup.set(group.id, new Set<string>())
  return catalog
}

function appendModel(catalog: GroupModelCatalog, seen: Set<string>, model: UserSupportedModel) {
  if (seen.has(model.name)) {
    return
  }
  seen.add(model.name)
  catalog.models.push({
    id: model.name,
    display_name: model.name,
    pricing: model.pricing,
  })
}
</script>
