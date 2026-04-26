<template>
  <div v-if="hasActiveSubscriptions" class="relative" ref="containerRef">
    <!-- Mini Progress Display -->
    <button
      @click="toggleTooltip"
      class="flex cursor-pointer items-center gap-2 rounded-xl bg-purple-50 px-3 py-1.5 transition-colors hover:bg-purple-100 dark:bg-purple-900/20 dark:hover:bg-purple-900/30"
      :title="t('subscriptionProgress.viewDetails')"
    >
      <Icon name="creditCard" size="sm" class="text-purple-600 dark:text-purple-400" />
      <div class="flex items-center gap-1.5">
        <!-- Combined progress indicator -->
        <div class="flex items-center gap-0.5">
          <div
            v-for="(indicator, index) in displayIndicators.slice(0, 3)"
            :key="index"
            class="h-2 w-2 rounded-full"
            :class="getProgressIndicatorClass(indicator)"
          ></div>
        </div>
        <span class="text-xs font-medium text-purple-700 dark:text-purple-300">
          {{ activeCount }}
        </span>
      </div>
    </button>

    <!-- Hover/Click Tooltip -->
    <transition name="dropdown">
      <div
        v-if="tooltipOpen"
        class="absolute right-0 z-50 mt-2 w-[340px] overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xl dark:border-dark-700 dark:bg-dark-800"
      >
        <div class="border-b border-gray-100 p-3 dark:border-dark-700">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('subscriptionProgress.title') }}
          </h3>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
            {{ t('subscriptionProgress.activeCount', { count: activeCount }) }}
          </p>
        </div>

        <div class="max-h-64 overflow-y-auto">
          <div
            v-for="product in displayProducts"
            :key="`product-${product.product_id}`"
            class="border-b border-gray-50 p-3 last:border-b-0 dark:border-dark-700/50"
          >
            <div class="mb-2 flex items-center justify-between">
              <div class="min-w-0">
                <span class="block truncate text-sm font-medium text-gray-900 dark:text-white">
                  {{ product.name }}
                </span>
                <div class="mt-1 flex flex-wrap gap-1">
                  <span
                    v-for="group in product.groups"
                    :key="group.group_id"
                    class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] text-gray-600 dark:bg-dark-700 dark:text-dark-300"
                  >
                    {{ group.group_name }} · {{ group.debit_multiplier }}x
                  </span>
                </div>
              </div>
              <span
                v-if="product.expires_at"
                class="ml-3 shrink-0 text-xs"
                :class="getDaysRemainingClass(product.expires_at)"
              >
                {{ formatDaysRemaining(product.expires_at) }}
              </span>
            </div>

            <div class="space-y-1.5">
              <div
                v-if="isProductUnlimited(product)"
                class="flex items-center gap-2 rounded-lg bg-gradient-to-r from-emerald-50 to-teal-50 px-2.5 py-1.5 dark:from-emerald-900/20 dark:to-teal-900/20"
              >
                <span class="text-lg text-emerald-600 dark:text-emerald-400">∞</span>
                <span class="text-xs font-medium text-emerald-700 dark:text-emerald-300">
                  {{ t('subscriptionProgress.unlimited') }}
                </span>
              </div>

              <template v-else>
                <div v-if="product.daily_limit_usd" class="flex items-center gap-2">
                  <span class="w-8 flex-shrink-0 text-[10px] text-gray-500">{{
                    t('subscriptionProgress.daily')
                  }}</span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="
                        getProgressBarClass(product.daily_usage_usd, getProductDailyDisplayLimit(product))
                      "
                      :style="{
                        width: getProgressWidth(
                          product.daily_usage_usd,
                          getProductDailyDisplayLimit(product)
                        )
                      }"
                    ></div>
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-gray-500">
                    {{ formatUsage(product.daily_usage_usd, getProductDailyDisplayLimit(product)) }}
                  </span>
                </div>
                <div
                  v-if="hasProductDailyCarryover(product)"
                  class="ml-10 rounded-md bg-amber-50 px-2 py-1 text-[10px] text-amber-700 dark:bg-amber-900/20 dark:text-amber-300"
                >
                  <div>{{ formatProductDailyCarryoverMessage(product) }}</div>
                  <div>{{ t('subscriptionProgress.carryoverRule') }}</div>
                </div>

                <div
                  v-if="shouldShowProductPeriodUsage(product.weekly_limit_usd, product.weekly_usage_usd)"
                  class="flex items-center gap-2"
                >
                  <span class="w-8 flex-shrink-0 text-[10px] text-gray-500">{{
                    t('subscriptionProgress.last7DaysShort')
                  }}</span>
                  <div
                    v-if="hasPositiveLimit(product.weekly_limit_usd)"
                    class="h-1.5 min-w-0 flex-1 rounded-full bg-gray-200 dark:bg-dark-600"
                  >
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="getProgressBarClass(product.weekly_usage_usd, product.weekly_limit_usd)"
                      :style="{
                        width: getProgressWidth(product.weekly_usage_usd, product.weekly_limit_usd)
                      }"
                    ></div>
                  </div>
                  <div v-else class="min-w-0 flex-1"></div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-gray-500">
                    {{ formatUsage(product.weekly_usage_usd, product.weekly_limit_usd) }}
                  </span>
                </div>

                <div
                  v-if="shouldShowProductPeriodUsage(product.monthly_limit_usd, product.monthly_usage_usd)"
                  class="flex items-center gap-2"
                >
                  <span class="w-8 flex-shrink-0 text-[10px] text-gray-500">{{
                    t('subscriptionProgress.last30DaysShort')
                  }}</span>
                  <div
                    v-if="hasPositiveLimit(product.monthly_limit_usd)"
                    class="h-1.5 min-w-0 flex-1 rounded-full bg-gray-200 dark:bg-dark-600"
                  >
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="getProgressBarClass(product.monthly_usage_usd, product.monthly_limit_usd)"
                      :style="{
                        width: getProgressWidth(product.monthly_usage_usd, product.monthly_limit_usd)
                      }"
                    ></div>
                  </div>
                  <div v-else class="min-w-0 flex-1"></div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-gray-500">
                    {{ formatUsage(product.monthly_usage_usd, product.monthly_limit_usd) }}
                  </span>
                </div>
              </template>
            </div>
          </div>

          <div
            v-for="subscription in displaySubscriptions"
            :key="`subscription-${subscription.id}`"
            class="border-b border-gray-50 p-3 last:border-b-0 dark:border-dark-700/50"
          >
            <div class="mb-2 flex items-center justify-between">
              <span class="text-sm font-medium text-gray-900 dark:text-white">
                {{ subscription.group?.name || `Group #${subscription.group_id}` }}
              </span>
              <span
                v-if="subscription.expires_at"
                class="text-xs"
                :class="getDaysRemainingClass(subscription.expires_at)"
              >
                {{ formatDaysRemaining(subscription.expires_at) }}
              </span>
            </div>

            <!-- Progress bars or Unlimited badge -->
            <div class="space-y-1.5">
              <!-- Unlimited subscription badge -->
              <div
                v-if="isUnlimited(subscription)"
                class="flex items-center gap-2 rounded-lg bg-gradient-to-r from-emerald-50 to-teal-50 px-2.5 py-1.5 dark:from-emerald-900/20 dark:to-teal-900/20"
              >
                <span class="text-lg text-emerald-600 dark:text-emerald-400">∞</span>
                <span class="text-xs font-medium text-emerald-700 dark:text-emerald-300">
                  {{ t('subscriptionProgress.unlimited') }}
                </span>
              </div>

              <!-- Progress bars for limited subscriptions -->
              <template v-else>
                <div v-if="subscription.group?.daily_limit_usd" class="flex items-center gap-2">
                  <span class="w-8 flex-shrink-0 text-[10px] text-gray-500">{{
                    t('subscriptionProgress.daily')
                  }}</span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.daily_usage_usd,
                          getDailyDisplayLimit(subscription)
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.daily_usage_usd,
                          getDailyDisplayLimit(subscription)
                        )
                      }"
                    ></div>
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-gray-500">
                    {{
                      formatUsage(subscription.daily_usage_usd, getDailyDisplayLimit(subscription))
                    }}
                  </span>
                </div>
                <div
                  v-if="hasDailyCarryover(subscription)"
                  class="ml-10 rounded-md bg-amber-50 px-2 py-1 text-[10px] text-amber-700 dark:bg-amber-900/20 dark:text-amber-300"
                >
                  <div>{{ formatDailyCarryoverMessage(subscription) }}</div>
                  <div>{{ t('subscriptionProgress.carryoverRule') }}</div>
                </div>

                <div v-if="subscription.group?.weekly_limit_usd" class="flex items-center gap-2">
                  <span class="w-8 flex-shrink-0 text-[10px] text-gray-500">{{
                    t('subscriptionProgress.weekly')
                  }}</span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.weekly_usage_usd,
                          subscription.group?.weekly_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.weekly_usage_usd,
                          subscription.group?.weekly_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-gray-500">
                    {{
                      formatUsage(subscription.weekly_usage_usd, subscription.group?.weekly_limit_usd)
                    }}
                  </span>
                </div>

                <div v-if="subscription.group?.monthly_limit_usd" class="flex items-center gap-2">
                  <span class="w-8 flex-shrink-0 text-[10px] text-gray-500">{{
                    t('subscriptionProgress.monthly')
                  }}</span>
                  <div class="h-1.5 min-w-0 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
                    <div
                      class="h-1.5 rounded-full transition-all"
                      :class="
                        getProgressBarClass(
                          subscription.monthly_usage_usd,
                          subscription.group?.monthly_limit_usd
                        )
                      "
                      :style="{
                        width: getProgressWidth(
                          subscription.monthly_usage_usd,
                          subscription.group?.monthly_limit_usd
                        )
                      }"
                    ></div>
                  </div>
                  <span class="w-24 flex-shrink-0 text-right text-[10px] text-gray-500">
                    {{
                      formatUsage(
                        subscription.monthly_usage_usd,
                        subscription.group?.monthly_limit_usd
                      )
                    }}
                  </span>
                </div>
              </template>
            </div>
          </div>
        </div>

        <div class="border-t border-gray-100 p-2 dark:border-dark-700">
          <router-link
            to="/subscriptions"
            @click="closeTooltip"
            class="block w-full py-1 text-center text-xs text-primary-600 hover:underline dark:text-primary-400"
          >
            {{ t('subscriptionProgress.viewAll') }}
          </router-link>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useSubscriptionProductStore, useSubscriptionStore } from '@/stores'
import type { ActiveSubscriptionProduct, UserSubscription } from '@/types'

const { t } = useI18n()

const subscriptionStore = useSubscriptionStore()
const subscriptionProductStore = useSubscriptionProductStore()

const containerRef = ref<HTMLElement | null>(null)
const tooltipOpen = ref(false)

// Use store data instead of local state
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)
const activeProducts = computed(() => subscriptionProductStore.items)
const activeCount = computed(() => displaySubscriptions.value.length + displayProducts.value.length)
const hasActiveSubscriptions = computed(() => activeCount.value > 0)

interface ProgressIndicator {
  unlimited: boolean
  maxPercentage: number
}

const displayIndicators = computed<ProgressIndicator[]>(() => [
  ...displayProducts.value.map((product) => ({
    unlimited: isProductUnlimited(product),
    maxPercentage: getMaxProductUsagePercentage(product)
  })),
  ...displaySubscriptions.value.map((subscription) => ({
    unlimited: isUnlimited(subscription),
    maxPercentage: getMaxUsagePercentage(subscription)
  }))
])

const displayProducts = computed(() => {
  return [...activeProducts.value].sort((a, b) => {
    const aMax = getMaxProductUsagePercentage(a)
    const bMax = getMaxProductUsagePercentage(b)
    return bMax - aMax
  })
})

const productCoveredGroupIds = computed(() => {
  const groupIds = new Set<number>()
  for (const product of activeProducts.value) {
    for (const group of product.groups || []) {
      groupIds.add(group.group_id)
    }
  }
  return groupIds
})

const displaySubscriptions = computed(() => {
  // Sort by most usage (highest percentage first)
  return activeSubscriptions.value
    .filter((subscription) => !productCoveredGroupIds.value.has(subscription.group_id))
    .sort((a, b) => {
      const aMax = getMaxUsagePercentage(a)
      const bMax = getMaxUsagePercentage(b)
      return bMax - aMax
    })
})

function getMaxUsagePercentage(sub: UserSubscription): number {
  const percentages: number[] = []
  if (sub.group?.daily_limit_usd) {
    const dailyLimit = getDailyDisplayLimit(sub)
    if (dailyLimit) {
      percentages.push(((sub.daily_usage_usd || 0) / dailyLimit) * 100)
    }
  }
  if (sub.group?.weekly_limit_usd) {
    percentages.push(((sub.weekly_usage_usd || 0) / sub.group.weekly_limit_usd) * 100)
  }
  if (sub.group?.monthly_limit_usd) {
    percentages.push(((sub.monthly_usage_usd || 0) / sub.group.monthly_limit_usd) * 100)
  }
  return percentages.length > 0 ? Math.max(...percentages) : 0
}

function getMaxProductUsagePercentage(product: ActiveSubscriptionProduct): number {
  const percentages: number[] = []
  const dailyLimit = getProductDailyDisplayLimit(product)
  if (dailyLimit) {
    percentages.push(((product.daily_usage_usd || 0) / dailyLimit) * 100)
  }
  if (product.weekly_limit_usd) {
    percentages.push(((product.weekly_usage_usd || 0) / product.weekly_limit_usd) * 100)
  }
  if (product.monthly_limit_usd) {
    percentages.push(((product.monthly_usage_usd || 0) / product.monthly_limit_usd) * 100)
  }
  return percentages.length > 0 ? Math.max(...percentages) : 0
}

function getDailyDisplayLimit(sub: UserSubscription): number | null | undefined {
  if (sub.daily_effective_limit_usd && sub.daily_effective_limit_usd > 0) {
    return sub.daily_effective_limit_usd
  }
  return sub.group?.daily_limit_usd
}

function hasDailyCarryover(sub: UserSubscription): boolean {
  return (sub.daily_carryover_in_usd || 0) > 0
}

function formatDailyCarryoverMessage(sub: UserSubscription): string {
  const total = (getDailyDisplayLimit(sub) || 0).toFixed(2)
  const carryover = (sub.daily_carryover_in_usd || 0).toFixed(2)
  return t('subscriptionProgress.todayAvailable', { total, carryover })
}

function getProductDailyDisplayLimit(product: ActiveSubscriptionProduct): number | null {
  return product.daily_limit_usd
}

function getProductDailyAvailableWithCarryover(product: ActiveSubscriptionProduct): number | null {
  if (!product.daily_limit_usd) return product.daily_limit_usd
  return product.daily_limit_usd + (product.daily_carryover_in_usd || 0)
}

function hasProductDailyCarryover(product: ActiveSubscriptionProduct): boolean {
  return (product.daily_carryover_in_usd || 0) > 0
}

function formatProductDailyCarryoverMessage(product: ActiveSubscriptionProduct): string {
  const total = (getProductDailyAvailableWithCarryover(product) || 0).toFixed(2)
  const carryover = (product.daily_carryover_in_usd || 0).toFixed(2)
  return t('subscriptionProgress.todayAvailable', { total, carryover })
}

function isProductUnlimited(product: ActiveSubscriptionProduct): boolean {
  return !product.daily_limit_usd && !product.weekly_limit_usd && !product.monthly_limit_usd
}

function isUnlimited(sub: UserSubscription): boolean {
  return (
    !sub.group?.daily_limit_usd &&
    !sub.group?.weekly_limit_usd &&
    !sub.group?.monthly_limit_usd
  )
}

function getProgressIndicatorClass(indicator: ProgressIndicator): string {
  if (indicator.unlimited) {
    return 'bg-emerald-500'
  }
  if (indicator.maxPercentage >= 90) return 'bg-red-500'
  if (indicator.maxPercentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function formatUsage(used: number | undefined, limit: number | null | undefined): string {
  const usedValue = (used || 0).toFixed(2)
  if (hasPositiveLimit(limit)) {
    return `$${usedValue}/$${limit?.toFixed(2)}`
  }
  return `$${usedValue}/∞`
}

function hasPositiveLimit(limit: number | null | undefined): boolean {
  return !!limit && limit > 0
}

function shouldShowProductPeriodUsage(
  limit: number | null | undefined,
  used: number | undefined
): boolean {
  return (limit !== null && limit !== undefined) || (used || 0) > 0
}

function formatDaysRemaining(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  if (diff < 0) return t('subscriptionProgress.expired')
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return t('subscriptionProgress.expiresToday')
  if (days === 1) return t('subscriptionProgress.expiresTomorrow')
  return t('subscriptionProgress.daysRemaining', { days })
}

function getDaysRemainingClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-500 dark:text-dark-400'
}

function toggleTooltip() {
  tooltipOpen.value = !tooltipOpen.value
}

function closeTooltip() {
  tooltipOpen.value = false
}

function handleClickOutside(event: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    closeTooltip()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  // Trigger initial fetch if not already loaded
  // The actual data loading is handled by App.vue globally
  subscriptionStore.fetchActiveSubscriptions().catch((error) => {
    console.error('Failed to load subscriptions in SubscriptionProgressMini:', error)
  })
  subscriptionProductStore.fetchActive().catch((error) => {
    console.error('Failed to load subscription products in SubscriptionProgressMini:', error)
  })
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
</style>
