<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <div
        v-if="!loading"
        class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <Icon name="creditCard" size="sm" class="text-emerald-600 dark:text-emerald-400" />
              <h2 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('userSubscriptions.balanceFallback.title', '订阅消耗完时，自动消耗余额') }}
              </h2>
            </div>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
              {{
                t(
                  'userSubscriptions.balanceFallback.description',
                  '开启后，产品订阅额度耗尽且分组存在余额兜底映射时，会在你设置的上限内自动改用余额。'
                )
              }}
            </p>
          </div>
          <label class="inline-flex shrink-0 items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input
              v-model="fallbackEnabled"
              type="checkbox"
              class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              :disabled="savingFallback"
              @change="saveBalanceFallbackSettings"
            />
            <span>{{ fallbackEnabled ? t('common.enabled', '已开启') : t('common.disabled', '已关闭') }}</span>
          </label>
        </div>

        <div v-if="fallbackEnabled" class="mt-4 grid gap-3 sm:grid-cols-[minmax(0,220px)_1fr] sm:items-end">
          <label class="block">
            <span class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-300">
              {{ t('userSubscriptions.balanceFallback.limit', '余额兜底上限') }}
            </span>
            <input
              v-model.number="fallbackLimit"
              type="number"
              min="0"
              step="0.01"
              class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-900 dark:text-white"
              :disabled="savingFallback"
              @blur="saveBalanceFallbackSettings"
            />
          </label>
          <div class="text-xs text-gray-500 dark:text-dark-400">
            {{
              t('userSubscriptions.balanceFallback.usage', {
                used: fallbackUsed.toFixed(2),
                remaining: fallbackRemaining.toFixed(2)
              })
            }}
          </div>
        </div>
      </div>

      <div
        v-if="!loading && (subscriptionProducts.length > 0 || visibleSubscriptions.length > 0)"
        class="rounded-lg border border-sky-200 bg-sky-50 p-4 text-sky-900 dark:border-sky-900/60 dark:bg-sky-950/30 dark:text-sky-100"
      >
        <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div class="min-w-0">
            <div class="flex items-center gap-2">
              <Icon name="key" size="sm" class="text-sky-600 dark:text-sky-300" />
              <h2 class="text-sm font-semibold">
                {{ t('userSubscriptions.keyReminder.title', '激活后建议新建分组专用 API Key') }}
              </h2>
            </div>
            <p class="mt-1 text-xs text-sky-700/90 dark:text-sky-200/80">
              {{
                t(
                  'userSubscriptions.keyReminder.description',
                  '这样可以避免继续误用旧 key 的余额或限额，并按分组隔离你的订阅用量。'
                )
              }}
            </p>
          </div>
          <button
            class="inline-flex shrink-0 items-center justify-center gap-2 rounded-md bg-sky-600 px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-sky-700"
            @click="router.push('/keys')"
          >
            <Icon name="plus" size="sm" />
            {{ t('userSubscriptions.keyReminder.action', '去生成 API Key') }}
          </button>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="!loading && subscriptionProducts.length === 0 && visibleSubscriptions.length === 0" class="card p-12 text-center">
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
      <div v-if="!loading && (subscriptionProducts.length > 0 || visibleSubscriptions.length > 0)" class="grid gap-6 lg:grid-cols-2">
        <div
          v-for="product in subscriptionProducts"
          :key="`product-${product.subscription_id}`"
          class="overflow-hidden rounded-2xl border border-emerald-200 bg-white dark:border-emerald-900/50 dark:bg-dark-800"
        >
          <div
            class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <h3 class="truncate font-semibold text-gray-900 dark:text-white">
                  {{ product.name }}
                </h3>
                <span
                  class="rounded-md border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[11px] font-medium text-emerald-700 dark:border-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-300"
                >
                  {{ product.code }}
                </span>
              </div>
              <p v-if="product.description" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                {{ product.description }}
              </p>
            </div>
            <span
              :class="[
                'shrink-0 rounded-full px-2 py-0.5 text-xs font-medium',
                product.status === 'active'
                  ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                  : product.status === 'expired'
                    ? 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
                    : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
              ]"
            >
              {{ t(`userSubscriptions.status.${product.status}`) }}
            </span>
          </div>

          <div class="space-y-4 p-4">
            <div v-if="product.expires_at" class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{ t('userSubscriptions.expires') }}</span>
              <span :class="getExpirationClass(product.expires_at)">
                {{ formatExpirationDate(product.expires_at) }}
              </span>
            </div>
            <div v-else class="flex items-center justify-between text-sm">
              <span class="text-gray-500 dark:text-dark-400">{{ t('userSubscriptions.expires') }}</span>
              <span class="text-gray-700 dark:text-gray-300">{{ t('userSubscriptions.noExpiration') }}</span>
            </div>

            <div v-if="product.daily_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (product.daily_usage_usd || 0).toFixed(2) }} / ${{
                    getProductDailyDisplayLimit(product).toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="getProgressBarClass(product.daily_usage_usd, getProductDailyDisplayLimit(product))"
                  :style="{ width: getProgressWidth(product.daily_usage_usd, getProductDailyDisplayLimit(product)) }"
                ></div>
              </div>
              <p
                v-if="hasProductDailyCarryover(product)"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{ formatProductDailyQuotaBreakdown(product) }}
              </p>
            </div>

            <div v-if="product.weekly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (product.weekly_usage_usd || 0).toFixed(2) }} / ${{ product.weekly_limit_usd.toFixed(2) }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="getProgressBarClass(product.weekly_usage_usd, product.weekly_limit_usd)"
                  :style="{ width: getProgressWidth(product.weekly_usage_usd, product.weekly_limit_usd) }"
                ></div>
              </div>
            </div>

            <div v-if="product.monthly_limit_usd" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (product.monthly_usage_usd || 0).toFixed(2) }} / ${{ product.monthly_limit_usd.toFixed(2) }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
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

            <div v-if="product.groups.length" class="space-y-2 border-t border-gray-100 pt-4 dark:border-dark-700">
              <p class="text-xs font-medium text-gray-500 dark:text-dark-400">
                {{ t('userSubscriptions.visibleGroups') }}
              </p>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="group in product.groups"
                  :key="group.group_id"
                  class="inline-flex max-w-full items-center gap-1 rounded-md border border-gray-200 px-2 py-1 text-xs text-gray-700 dark:border-dark-600 dark:text-gray-300"
                >
                  <span class="truncate">{{ group.group_name }}</span>
                  <span class="shrink-0 text-gray-400">
                    {{ t('userSubscriptions.groupMultiplier', { multiplier: group.debit_multiplier }) }}
                  </span>
                </span>
              </div>
            </div>
          </div>
        </div>

        <div
          v-for="subscription in visibleSubscriptions"
          :key="subscription.id"
          class="overflow-hidden rounded-2xl border bg-white dark:bg-dark-800"
          :class="platformBorderClass(subscription.group?.platform || '')"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex items-center gap-3">
              <div :class="['h-1.5 w-1.5 shrink-0 rounded-full', platformAccentDotClass(subscription.group?.platform || '')]" />
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold text-gray-900 dark:text-white">
                    {{ subscription.group?.name || `Group #${subscription.group_id}` }}
                  </h3>
                  <span :class="['rounded-md border px-2 py-0.5 text-[11px] font-medium', platformBadgeClass(subscription.group?.platform || '')]">
                    {{ platformLabel(subscription.group?.platform || '') }}
                  </span>
                </div>
                <p v-if="subscription.group?.description" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                  {{ subscription.group.description }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <span
                :class="[
                  'rounded-full px-2 py-0.5 text-xs font-medium',
                  subscription.status === 'active'
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
                    : subscription.status === 'expired'
                      ? 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
                ]"
              >
                {{ t(`userSubscriptions.status.${subscription.status}`) }}
              </span>
              <button
                v-if="subscription.status === 'active'"
                :class="['rounded-lg px-3 py-1.5 text-xs font-semibold text-white transition-colors', platformButtonClass(subscription.group?.platform || '')]"
                @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
              >
                {{ t('payment.renewNow') }}
              </button>
            </div>
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
                    subscription.group.daily_limit_usd.toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.daily_usage_usd,
                      subscription.group.daily_limit_usd
                    )
                  }"
                ></div>
              </div>
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
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import subscriptionsAPI from '@/api/subscriptions'
import subscriptionProductsAPI from '@/api/subscriptionProducts'
import { updateProfile } from '@/api/user'
import type { ActiveSubscriptionProduct, UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateOnly } from '@/utils/format'
import { platformBorderClass, platformBadgeClass, platformButtonClass, platformLabel } from '@/utils/platformColors'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'bg-orange-500'
    case 'openai': return 'bg-emerald-500'
    case 'antigravity': return 'bg-purple-500'
    case 'gemini': return 'bg-blue-500'
    default: return 'bg-gray-400'
  }
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const subscriptions = ref<UserSubscription[]>([])
const subscriptionProducts = ref<ActiveSubscriptionProduct[]>([])
const loading = ref(true)
const savingFallback = ref(false)
const fallbackEnabled = ref(false)
const fallbackLimit = ref(0)

const fallbackUsed = computed(() => authStore.user?.subscription_balance_fallback_used_usd || 0)
const fallbackRemaining = computed(() => Math.max((fallbackLimit.value || 0) - fallbackUsed.value, 0))

const productGroupIDs = computed(() => {
  const ids = new Set<number>()
  for (const product of subscriptionProducts.value) {
    for (const group of product.groups || []) {
      ids.add(group.group_id)
    }
  }
  return ids
})

const visibleSubscriptions = computed(() =>
  subscriptions.value.filter((subscription) => !productGroupIDs.value.has(subscription.group_id))
)

async function loadSubscriptions() {
  try {
    loading.value = true
    const [legacySubscriptions, products, profile] = await Promise.all([
      subscriptionsAPI.getActiveSubscriptions(),
      subscriptionProductsAPI.getActive(),
      authStore.refreshUser()
    ])
    subscriptions.value = legacySubscriptions
    subscriptionProducts.value = products
    fallbackEnabled.value = Boolean(profile.subscription_balance_fallback_enabled)
    fallbackLimit.value = profile.subscription_balance_fallback_limit_usd || 0
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

async function saveBalanceFallbackSettings() {
  savingFallback.value = true
  try {
    if (!fallbackEnabled.value && fallbackLimit.value < 0) {
      fallbackLimit.value = 0
    }
    const updated = await updateProfile({
      subscription_balance_fallback_enabled: fallbackEnabled.value,
      subscription_balance_fallback_limit_usd: Math.max(fallbackLimit.value || 0, 0)
    })
    authStore.user = updated
    fallbackEnabled.value = Boolean(updated.subscription_balance_fallback_enabled)
    fallbackLimit.value = updated.subscription_balance_fallback_limit_usd || 0
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    console.error('Failed to save subscription balance fallback:', error)
    appStore.showError(t('common.error'))
    fallbackEnabled.value = Boolean(authStore.user?.subscription_balance_fallback_enabled)
    fallbackLimit.value = authStore.user?.subscription_balance_fallback_limit_usd || 0
  } finally {
    savingFallback.value = false
  }
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

function getProductDailyDisplayLimit(product: ActiveSubscriptionProduct): number {
  return (product.daily_limit_usd || 0) + (product.daily_carryover_in_usd || 0)
}

function hasProductDailyCarryover(product: ActiveSubscriptionProduct): boolean {
  return (product.daily_carryover_in_usd || 0) > 0
}

function formatProductDailyQuotaBreakdown(product: ActiveSubscriptionProduct): string {
  return t('userSubscriptions.dailyQuotaBreakdown', {
    carryover: (product.daily_carryover_in_usd || 0).toFixed(2),
    today: (product.daily_limit_usd || 0).toFixed(2),
    total: getProductDailyDisplayLimit(product).toFixed(2)
  })
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
    return `${dateStr} (${t('common.today')})`
  }
  if (days === 1) {
    return `${dateStr} (${t('common.tomorrow')})`
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
