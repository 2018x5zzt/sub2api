<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="displayedSubscriptions.length === 0 && subscriptionProducts.length === 0" class="card p-12 text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
        >
          <Icon name="creditCard" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('userSubscriptions.noActiveSubscriptions') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
        </p>
      </div>

      <!-- Subscriptions Grid -->
      <div v-else class="grid gap-6 lg:grid-cols-2">
        <div
          v-for="product in subscriptionProducts"
          :key="`product-${product.product_id}`"
          class="card overflow-hidden"
        >
          <div
            class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex min-w-0 items-center gap-3">
              <div
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-purple-100 dark:bg-purple-900/30"
              >
                <Icon name="creditCard" size="md" class="text-purple-600 dark:text-purple-400" />
              </div>
              <div class="min-w-0">
                <h3 class="truncate font-semibold text-gray-900 dark:text-white">
                  {{ product.name }}
                </h3>
                <p class="truncate text-xs text-gray-500 dark:text-dark-400">
                  {{ product.description }}
                </p>
              </div>
            </div>
            <span
              :class="[
                'badge',
                product.status === 'active'
                  ? 'badge-success'
                  : product.status === 'expired'
                    ? 'badge-warning'
                    : 'badge-danger'
              ]"
            >
              {{ t(`userSubscriptions.status.${product.status}`) }}
            </span>
          </div>

          <div class="space-y-4 p-4">
            <div v-if="product.expires_at" class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span :class="getExpirationClass(product.expires_at)">
                {{ formatExpirationDate(product.expires_at) }}
              </span>
            </div>
            <div v-else class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{
                t('userSubscriptions.noExpiration')
              }}</span>
            </div>

            <div class="flex flex-wrap gap-2">
              <span
                v-for="group in product.groups"
                :key="group.group_id"
                class="rounded-lg bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-dark-300"
              >
                {{ group.group_name }} · {{ group.debit_multiplier }}x
              </span>
            </div>

            <div v-if="product.daily_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (product.daily_usage_usd || 0).toFixed(2) }} / ${{
                    getProductDailyDisplayLimit(product)?.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="getProgressBarClass(product.daily_usage_usd, getProductDailyDisplayLimit(product))"
                  :style="{
                    width: getProgressWidth(product.daily_usage_usd, getProductDailyDisplayLimit(product))
                  }"
                ></div>
              </div>
              <p
                v-if="hasProductDailyCarryover(product)"
                class="text-xs text-amber-600 dark:text-amber-400"
              >
                {{ formatProductDailyQuotaBreakdown(product) }}
              </p>
            </div>

            <div v-if="shouldShowProductPeriodUsage(product.weekly_limit_usd, product.weekly_usage_usd)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.last7DaysUsage') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  {{ formatPeriodUsage(product.weekly_usage_usd, product.weekly_limit_usd) }}
                </span>
              </div>
              <div
                v-if="hasPositiveLimit(product.weekly_limit_usd)"
                class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600"
              >
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="getProgressBarClass(product.weekly_usage_usd, product.weekly_limit_usd)"
                  :style="{ width: getProgressWidth(product.weekly_usage_usd, product.weekly_limit_usd) }"
                ></div>
              </div>
            </div>

            <div v-if="shouldShowProductPeriodUsage(product.monthly_limit_usd, product.monthly_usage_usd)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.last30DaysUsage') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  {{ formatPeriodUsage(product.monthly_usage_usd, product.monthly_limit_usd) }}
                </span>
              </div>
              <div
                v-if="hasPositiveLimit(product.monthly_limit_usd)"
                class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600"
              >
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="getProgressBarClass(product.monthly_usage_usd, product.monthly_limit_usd)"
                  :style="{ width: getProgressWidth(product.monthly_usage_usd, product.monthly_limit_usd) }"
                ></div>
              </div>
            </div>

            <div
              v-if="!product.daily_limit_usd && !product.weekly_limit_usd && !product.monthly_limit_usd"
              class="flex items-center justify-center rounded-xl bg-gradient-to-r from-emerald-50 to-teal-50 py-6 dark:from-emerald-900/20 dark:to-teal-900/20"
            >
              <div class="flex items-center gap-3">
                <span class="text-4xl text-emerald-600 dark:text-emerald-400">∞</span>
                <div>
                  <p class="text-sm font-medium text-emerald-700 dark:text-emerald-300">
                    {{ t('userSubscriptions.unlimited') }}
                  </p>
                  <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">
                    {{ t('userSubscriptions.unlimitedDesc') }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div
          v-for="subscription in displayedSubscriptions"
          :key="`subscription-${subscription.id}`"
          class="card overflow-hidden"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-xl bg-purple-100 dark:bg-purple-900/30"
              >
                <Icon name="creditCard" size="md" class="text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">
                  {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                </h3>
                <p class="text-xs text-gray-500 dark:text-dark-400">
                  {{ subscription.group?.description || '' }}
                </p>
              </div>
            </div>
            <span
              :class="[
                'badge',
                subscription.status === 'active'
                  ? 'badge-success'
                  : subscription.status === 'expired'
                    ? 'badge-warning'
                    : 'badge-danger'
              ]"
            >
              {{ t(`userSubscriptions.status.${subscription.status}`) }}
            </span>
          </div>

          <!-- Usage Progress -->
          <div class="space-y-4 p-4">
            <!-- Expiration Info -->
            <div v-if="subscription.expires_at" class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span :class="getExpirationClass(subscription.expires_at)">
                {{ formatExpirationDate(subscription.expires_at) }}
              </span>
            </div>
            <div v-else class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{
                t('userSubscriptions.expires')
              }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{
                t('userSubscriptions.noExpiration')
              }}</span>
            </div>

            <!-- Daily Usage -->
            <div v-if="subscription.group?.daily_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{
                    getSubscriptionDailyDisplayLimit(subscription)?.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.daily_usage_usd,
                      getSubscriptionDailyDisplayLimit(subscription)
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.daily_usage_usd,
                      getSubscriptionDailyDisplayLimit(subscription)
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="hasSubscriptionDailyCarryover(subscription)"
                class="text-xs text-amber-600 dark:text-amber-400"
              >
                {{ formatSubscriptionDailyQuotaBreakdown(subscription) }}
              </p>
              <p
                v-if="subscription.daily_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.daily_window_start, 24)
                  })
                }}
              </p>
            </div>

            <!-- Weekly Usage -->
            <div v-if="subscription.group?.weekly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.weekly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.weekly_usage_usd,
                      subscription.group.weekly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.weekly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.weekly_window_start, 168)
                  })
                }}
              </p>
            </div>

            <!-- Monthly Usage -->
            <div v-if="subscription.group?.monthly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{
                    subscription.group.monthly_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.monthly_usage_usd,
                      subscription.group.monthly_limit_usd
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.monthly_window_start"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{
                  t('userSubscriptions.resetIn', {
                    time: formatResetTime(subscription.monthly_window_start, 720)
                  })
                }}
              </p>
            </div>

            <!-- No limits configured - Unlimited badge -->
            <div
              v-if="
                !subscription.group?.daily_limit_usd &&
                !subscription.group?.weekly_limit_usd &&
                !subscription.group?.monthly_limit_usd
              "
              class="flex items-center justify-center rounded-xl bg-gradient-to-r from-emerald-50 to-teal-50 py-6 dark:from-emerald-900/20 dark:to-teal-900/20"
            >
              <div class="flex items-center gap-3">
                <span class="text-4xl text-emerald-600 dark:text-emerald-400">∞</span>
                <div>
                  <p class="text-sm font-medium text-emerald-700 dark:text-emerald-300">
                    {{ t('userSubscriptions.unlimited') }}
                  </p>
                  <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">
                    {{ t('userSubscriptions.unlimitedDesc') }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useSubscriptionProductStore } from '@/stores/subscriptionProducts'
import subscriptionsAPI from '@/api/subscriptions'
import type { ActiveSubscriptionProduct, UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateOnly } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const subscriptionProductStore = useSubscriptionProductStore()

const subscriptions = ref<UserSubscription[]>([])
const loading = ref(true)
const subscriptionProducts = computed(() => subscriptionProductStore.items)
const productCoveredGroupIds = computed(() => {
  const groupIds = new Set<number>()
  for (const product of subscriptionProducts.value) {
    for (const group of product.groups || []) {
      groupIds.add(group.group_id)
    }
  }
  return groupIds
})
const displayedSubscriptions = computed(() =>
  subscriptions.value.filter((subscription) => !productCoveredGroupIds.value.has(subscription.group_id))
)

async function loadSubscriptions() {
  try {
    loading.value = true
    const [legacySubscriptions] = await Promise.all([
      subscriptionsAPI.getMySubscriptions(),
      subscriptionProductStore.fetchActive(true)
    ])
    subscriptions.value = legacySubscriptions
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function getProductDailyDisplayLimit(product: ActiveSubscriptionProduct): number | null {
  if (product.daily_effective_limit_usd && product.daily_effective_limit_usd > 0) {
    return product.daily_effective_limit_usd
  }
  if (!product.daily_limit_usd) return product.daily_limit_usd
  return product.daily_limit_usd + (product.daily_carryover_in_usd || 0)
}

function getProductDailyAvailableWithCarryover(product: ActiveSubscriptionProduct): number | null {
  return getProductDailyDisplayLimit(product)
}

function hasProductDailyCarryover(product: ActiveSubscriptionProduct): boolean {
  return (product.daily_carryover_in_usd || 0) > 0
}

function formatProductDailyQuotaBreakdown(product: ActiveSubscriptionProduct): string {
  const carryover = (product.daily_carryover_in_usd || 0).toFixed(2)
  const today = (product.daily_limit_usd || 0).toFixed(2)
  const total = (getProductDailyAvailableWithCarryover(product) || 0).toFixed(2)
  return t('userSubscriptions.dailyQuotaBreakdown', { carryover, today, total })
}

function getSubscriptionDailyDisplayLimit(subscription: UserSubscription): number | null | undefined {
  if (subscription.daily_effective_limit_usd && subscription.daily_effective_limit_usd > 0) {
    return subscription.daily_effective_limit_usd
  }
  if (!subscription.group?.daily_limit_usd) return subscription.group?.daily_limit_usd
  return subscription.group.daily_limit_usd + (subscription.daily_carryover_in_usd || 0)
}

function hasSubscriptionDailyCarryover(subscription: UserSubscription): boolean {
  return (subscription.daily_carryover_in_usd || 0) > 0
}

function formatSubscriptionDailyQuotaBreakdown(subscription: UserSubscription): string {
  const carryover = (subscription.daily_carryover_in_usd || 0).toFixed(2)
  const today = (subscription.group?.daily_limit_usd || 0).toFixed(2)
  const total = (getSubscriptionDailyDisplayLimit(subscription) || 0).toFixed(2)
  return t('userSubscriptions.dailyQuotaBreakdown', { carryover, today, total })
}

function hasPositiveLimit(limit: number | null | undefined): boolean {
  return !!limit && limit > 0
}

function shouldShowProductPeriodUsage(
  limit: number | null | undefined,
  used: number | undefined
): boolean {
  return limit !== null && limit !== undefined || (used || 0) > 0
}

function formatPeriodUsage(used: number | undefined, limit: number | null | undefined): string {
  const usedValue = `$${(used || 0).toFixed(2)}`
  if (hasPositiveLimit(limit)) {
    return `${usedValue} / $${limit?.toFixed(2)}`
  }
  return `${usedValue} / ${t('userSubscriptions.unlimited')}`
}

function getProgressWidth(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return '0%'
  const percentage = Math.min(((used || 0) / limit) * 100, 100)
  return `${percentage}%`
}

function getProgressBarClass(used: number | undefined, limit: number | null | undefined): string {
  if (!limit || limit === 0) return 'bg-gray-400'
  const percentage = ((used || 0) / limit) * 100
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-orange-500'
  return 'bg-green-500'
}

function formatExpirationDate(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days < 0) {
    return t('userSubscriptions.status.expired')
  }

  const dateStr = formatDateOnly(expires)

  if (days === 0) {
    return `${dateStr} (Today)`
  }
  if (days === 1) {
    return `${dateStr} (Tomorrow)`
  }

  return t('userSubscriptions.daysRemaining', { days }) + ` (${dateStr})`
}

function getExpirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-700 dark:text-gray-300'
}

function formatResetTime(windowStart: string | null, windowHours: number): string {
  if (!windowStart) return t('userSubscriptions.windowNotActive')

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const now = new Date()
  const diff = end.getTime() - now.getTime()

  if (diff <= 0) return t('userSubscriptions.windowNotActive')

  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))

  if (hours > 24) {
    const days = Math.floor(hours / 24)
    const remainingHours = hours % 24
    return `${days}d ${remainingHours}h`
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  return `${minutes}m`
}

onMounted(() => {
  loadSubscriptions()
})
</script>
