<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="flex flex-col items-center gap-3">
          <div class="h-10 w-10 animate-spin rounded-full border-[3px] border-primary-200 border-t-primary-600 dark:border-dark-600 dark:border-t-primary-400" />
          <span class="text-sm text-gray-400 dark:text-gray-500">{{ t('common.loading', 'Loading...') }}</span>
        </div>
      </div>

      <template v-if="!loading">
        <!-- Balance Fallback Card -->
        <div class="overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-start gap-3">
              <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30">
                <Icon name="creditCard" size="sm" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div class="min-w-0">
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('userSubscriptions.balanceFallback.title', '订阅消耗完时，自动消耗余额') }}
                </h2>
                <p class="mt-0.5 text-xs leading-relaxed text-gray-500 dark:text-dark-400">
                  {{ t('userSubscriptions.balanceFallback.description', '开启后，产品订阅额度耗尽且分组存在余额兜底映射时，会在你设置的上限内自动改用余额。') }}
                </p>
              </div>
            </div>
            <label class="inline-flex shrink-0 cursor-pointer items-center gap-2.5">
              <div class="relative">
                <input
                  v-model="fallbackEnabled"
                  type="checkbox"
                  class="peer sr-only"
                  :disabled="savingFallback"
                  @change="saveBalanceFallbackSettings"
                />
                <div class="h-6 w-11 rounded-full bg-gray-200 transition-colors after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:bg-white after:shadow-sm after:transition-all peer-checked:bg-emerald-500 peer-checked:after:translate-x-full peer-disabled:opacity-50 dark:bg-dark-600 dark:after:bg-dark-300 dark:peer-checked:bg-emerald-600" />
              </div>
              <span class="text-xs font-medium" :class="fallbackEnabled ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400 dark:text-gray-500'">
                {{ fallbackEnabled ? t('common.enabled', '已开启') : t('common.disabled', '已关闭') }}
              </span>
            </label>
          </div>
          <div v-if="fallbackEnabled" class="border-t border-gray-100 bg-gray-50/50 px-5 py-4 dark:border-dark-700 dark:bg-dark-800/50">
            <div class="flex flex-wrap items-end gap-4">
              <label class="block w-56">
                <span class="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('userSubscriptions.balanceFallback.group', 'Balance group') }}
                </span>
                <select
                  v-model.number="fallbackGroupId"
                  class="input"
                  :disabled="savingFallback"
                  @change="saveBalanceFallbackSettings"
                >
                  <option :value="null">{{ t('userSubscriptions.balanceFallback.selectGroup', 'Select balance group') }}</option>
                  <option
                    v-for="option in fallbackGroupOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </option>
                </select>
              </label>
              <label class="block w-48">
                <span class="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('userSubscriptions.balanceFallback.limit', '余额兜底上限') }}
                </span>
                <div class="relative">
                  <span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-gray-400">$</span>
                  <input
                    v-model.number="fallbackLimit"
                    type="number"
                    min="0"
                    step="0.01"
                    class="input pl-7"
                    :disabled="savingFallback"
                    @blur="saveBalanceFallbackSettings"
                  />
                </div>
              </label>
              <span v-if="fallbackLimit > 0" class="pb-2 text-xs tabular-nums text-gray-500 dark:text-dark-400">
                {{ t('userSubscriptions.balanceFallback.usage', { used: fallbackUsed.toFixed(2), remaining: fallbackRemaining.toFixed(2) }) }}
              </span>
              <span v-else class="pb-2 text-xs text-gray-400 dark:text-gray-500">
                {{ t('userSubscriptions.balanceFallback.setLimitHint', '请设置一个大于 0 的上限以启用余额兜底') }}
              </span>
            </div>
            <p class="mt-3 text-xs leading-relaxed text-amber-700 dark:text-amber-300">
              {{ t('userSubscriptions.balanceFallback.negativeBalanceHint', 'If your balance becomes negative, future requests will be blocked until you recharge.') }}
            </p>
          </div>
        </div>

        <!-- API Key Reminder -->
        <div
          v-if="subscriptionProducts.length > 0 || visibleSubscriptions.length > 0"
          class="flex flex-col items-start gap-3 rounded-2xl border border-sky-200/80 bg-gradient-to-r from-sky-50 to-blue-50 p-4 sm:flex-row sm:items-center sm:justify-between dark:border-sky-900/50 dark:from-sky-950/40 dark:to-blue-950/30"
        >
          <div class="flex items-start gap-3">
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-sky-100 dark:bg-sky-900/40">
              <Icon name="key" size="sm" class="text-sky-600 dark:text-sky-300" />
            </div>
            <div>
              <h2 class="text-sm font-semibold text-sky-900 dark:text-sky-100">
                {{ t('userSubscriptions.keyReminder.title', '激活后建议新建分组专用 API Key') }}
              </h2>
              <p class="mt-0.5 text-xs text-sky-700/80 dark:text-sky-300/70">
                {{ t('userSubscriptions.keyReminder.description', '这样可以避免继续误用旧 key 的余额或限额，并按分组隔离你的订阅用量。') }}
              </p>
            </div>
          </div>
          <button
            class="inline-flex shrink-0 items-center gap-1.5 rounded-lg bg-sky-600 px-3.5 py-2 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-sky-700"
            @click="router.push('/keys')"
          >
            <Icon name="plus" size="sm" />
            {{ t('userSubscriptions.keyReminder.action', '去生成 API Key') }}
          </button>
        </div>

        <!-- Empty State -->
        <div v-if="subscriptionProducts.length === 0 && visibleSubscriptions.length === 0" class="rounded-2xl border border-dashed border-gray-300 bg-white py-16 text-center dark:border-dark-600 dark:bg-dark-800">
          <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-700">
            <Icon name="creditCard" size="xl" class="text-gray-400" />
          </div>
          <h3 class="mb-1 text-base font-semibold text-gray-900 dark:text-white">
            {{ t('userSubscriptions.noActiveSubscriptions') }}
          </h3>
          <p class="text-sm text-gray-500 dark:text-dark-400">
            {{ t('userSubscriptions.noActiveSubscriptionsDesc') }}
          </p>
        </div>

        <!-- Product Subscription Cards -->
        <div v-if="subscriptionProducts.length > 0 || visibleSubscriptions.length > 0" class="grid gap-5 lg:grid-cols-2">

          <!-- Product Cards -->
          <div
            v-for="product in subscriptionProducts"
            :key="`product-${product.subscription_id}`"
            class="group overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-sm transition-shadow hover:shadow-md dark:border-dark-700 dark:bg-dark-800"
          >
            <!-- Card Header -->
            <div class="relative border-b border-gray-100 dark:border-dark-700">
              <div class="absolute inset-x-0 top-0 h-1" :class="productGradientClass(product)" />
              <div class="flex items-center justify-between px-5 pb-4 pt-5">
                <div class="flex items-center gap-3 min-w-0">
                  <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl" :class="productIconBgClass(product)">
                    <PlatformIcon :platform="productPlatform(product) as any" size="md" class="opacity-80" />
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center gap-2.5">
                      <h3 class="truncate text-base font-bold text-gray-900 dark:text-white">
                        {{ product.name }}
                      </h3>
                      <span class="sub-status-badge" :class="product.status === 'active' ? 'sub-status-active' : product.status === 'expired' ? 'sub-status-expired' : 'sub-status-revoked'">
                        {{ t(`userSubscriptions.status.${product.status}`) }}
                      </span>
                    </div>
                    <p v-if="product.description" class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                      {{ product.description }}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <!-- Card Body -->
            <div class="space-y-5 px-5 py-4">
              <!-- Expiry -->
              <div class="flex items-center justify-between">
                <span class="text-xs font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('userSubscriptions.expires') }}</span>
                <span v-if="product.expires_at" class="text-sm font-medium" :class="getExpirationClass(product.expires_at)">
                  {{ formatExpirationDate(product.expires_at) }}
                </span>
                <span v-else class="text-sm text-gray-500 dark:text-gray-400">{{ t('userSubscriptions.noExpiration') }}</span>
              </div>

              <!-- Usage Section -->
              <div v-if="product.daily_limit_usd || product.weekly_limit_usd || product.monthly_limit_usd" class="space-y-3">
                <!-- Daily -->
                <div class="sub-usage-block">
                  <div class="flex items-center justify-between">
                    <span class="sub-usage-label">{{ t('userSubscriptions.daily') }}</span>
                    <div class="flex items-center gap-2">
                      <span class="sub-usage-value">${{ (product.daily_usage_usd || 0).toFixed(2) }} / ${{ getProductDailyDisplayLimit(product).toFixed(2) }}</span>
                      <span v-if="hasProductDailyCarryover(product)" class="sub-carryover-tag">{{ t('productSubscription.carryover', 'Carry') }} ${{ (product.daily_carryover_in_usd || 0).toFixed(2) }}</span>
                    </div>
                  </div>
                  <div class="sub-bar-track">
                    <div class="sub-bar-fill" :class="getProgressBarClass(product.daily_usage_usd, getProductDailyDisplayLimit(product))" :style="{ width: getProgressWidth(product.daily_usage_usd, getProductDailyDisplayLimit(product)) }" />
                  </div>
                  <p v-if="hasProductDailyCarryover(product)" class="text-[11px] text-gray-400 dark:text-gray-500">
                    {{ formatProductDailyQuotaBreakdown(product) }}
                  </p>
                </div>
                <!-- Weekly -->
                <div class="sub-usage-block">
                  <div class="flex items-center justify-between">
                    <span class="sub-usage-label">{{ t('userSubscriptions.weekly') }}</span>
                    <span class="sub-usage-value">${{ (product.weekly_usage_usd || 0).toFixed(2) }} / ${{ inferLimit(product.weekly_limit_usd, product.daily_limit_usd, 7).toFixed(2) }}</span>
                  </div>
                  <div class="sub-bar-track">
                    <div class="sub-bar-fill" :class="getProgressBarClass(product.weekly_usage_usd, inferLimit(product.weekly_limit_usd, product.daily_limit_usd, 7))" :style="{ width: getProgressWidth(product.weekly_usage_usd, inferLimit(product.weekly_limit_usd, product.daily_limit_usd, 7)) }" />
                  </div>
                </div>
                <!-- Monthly -->
                <div class="sub-usage-block">
                  <div class="flex items-center justify-between">
                    <span class="sub-usage-label">{{ t('userSubscriptions.monthly') }}</span>
                    <span class="sub-usage-value">${{ (product.monthly_usage_usd || 0).toFixed(2) }} / ${{ inferLimit(product.monthly_limit_usd, product.daily_limit_usd, 30).toFixed(2) }}</span>
                  </div>
                  <div class="sub-bar-track">
                    <div class="sub-bar-fill" :class="getProgressBarClass(product.monthly_usage_usd, inferLimit(product.monthly_limit_usd, product.daily_limit_usd, 30))" :style="{ width: getProgressWidth(product.monthly_usage_usd, inferLimit(product.monthly_limit_usd, product.daily_limit_usd, 30)) }" />
                  </div>
                </div>
              </div>

              <!-- Unlimited -->
              <div v-if="!product.daily_limit_usd && !product.weekly_limit_usd && !product.monthly_limit_usd" class="flex items-center gap-3 rounded-xl bg-gradient-to-r from-emerald-50 to-teal-50 px-4 py-5 dark:from-emerald-900/20 dark:to-teal-900/20">
                <span class="text-3xl text-emerald-500 dark:text-emerald-400">∞</span>
                <div>
                  <p class="text-sm font-semibold text-emerald-700 dark:text-emerald-300">{{ t('userSubscriptions.unlimited') }}</p>
                  <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">{{ t('userSubscriptions.unlimitedDesc') }}</p>
                </div>
              </div>

              <!-- Groups -->
              <div v-if="product.groups.length" class="border-t border-gray-100 pt-4 dark:border-dark-700">
                <p class="mb-2 text-[11px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('userSubscriptions.visibleGroups') }}</p>
                <div class="flex flex-wrap gap-1.5">
                  <span
                    v-for="group in product.groups"
                    :key="group.group_id"
                    class="inline-flex items-center gap-1.5 rounded-lg bg-gray-50 px-2.5 py-1 text-xs text-gray-600 ring-1 ring-inset ring-gray-200/80 dark:bg-dark-700 dark:text-gray-300 dark:ring-dark-600"
                  >
                    <PlatformIcon :platform="group.group_platform as any" size="xs" />
                    <span class="truncate">{{ group.group_name }}</span>
                    <span class="font-semibold text-gray-400 dark:text-gray-500">{{ t('userSubscriptions.groupMultiplier', { multiplier: group.debit_multiplier }) }}</span>
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Legacy Group Subscription Cards -->
          <div
            v-for="subscription in visibleSubscriptions"
            :key="subscription.id"
            class="group overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-sm transition-shadow hover:shadow-md dark:border-dark-700 dark:bg-dark-800"
          >
            <div class="relative border-b border-gray-100 dark:border-dark-700">
              <div :class="['absolute inset-x-0 top-0 h-1', platformGradientClass(subscription.group?.platform || '')]" />
              <div class="flex items-center justify-between px-5 pb-4 pt-5">
                <div class="flex items-center gap-3 min-w-0">
                  <div :class="['h-2 w-2 shrink-0 rounded-full', platformAccentDotClass(subscription.group?.platform || '')]" />
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <h3 class="truncate text-base font-bold text-gray-900 dark:text-white">
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
                <div class="flex shrink-0 items-center gap-2">
                  <span class="sub-status-badge" :class="subscription.status === 'active' ? 'sub-status-active' : subscription.status === 'expired' ? 'sub-status-expired' : 'sub-status-revoked'">
                    {{ t(`userSubscriptions.status.${subscription.status}`) }}
                  </span>
                  <button
                    v-if="subscription.status === 'active'"
                    :class="['rounded-lg px-3 py-1.5 text-xs font-semibold text-white shadow-sm transition-colors', platformButtonClass(subscription.group?.platform || '')]"
                    @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(subscription.group_id) } })"
                  >
                    {{ t('payment.renewNow') }}
                  </button>
                </div>
              </div>
            </div>

            <div class="space-y-5 px-5 py-4">
              <!-- Expiry -->
              <div class="flex items-center justify-between">
                <span class="text-xs font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('userSubscriptions.expires') }}</span>
                <span v-if="subscription.expires_at" class="text-sm font-medium" :class="getExpirationClass(subscription.expires_at)">
                  {{ formatExpirationDate(subscription.expires_at) }}
                </span>
                <span v-else class="text-sm text-gray-500 dark:text-gray-400">{{ t('userSubscriptions.noExpiration') }}</span>
              </div>

              <!-- Usage -->
              <div v-if="subscription.group?.daily_limit_usd || subscription.group?.weekly_limit_usd || subscription.group?.monthly_limit_usd" class="space-y-3">
                <div v-if="subscription.group?.daily_limit_usd" class="sub-usage-block">
                  <div class="flex items-center justify-between">
                    <span class="sub-usage-label">{{ t('userSubscriptions.daily') }}</span>
                    <span class="sub-usage-value">${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{ subscription.group.daily_limit_usd.toFixed(2) }}</span>
                  </div>
                  <div class="sub-bar-track">
                    <div class="sub-bar-fill" :class="getProgressBarClass(subscription.daily_usage_usd, subscription.group.daily_limit_usd)" :style="{ width: getProgressWidth(subscription.daily_usage_usd, subscription.group.daily_limit_usd) }" />
                  </div>
                  <p v-if="subscription.daily_window_start" class="text-[11px] text-gray-400 dark:text-gray-500">
                    {{ t('userSubscriptions.resetIn', { time: formatResetTime(subscription.daily_window_start, 24) }) }}
                  </p>
                </div>

                <div v-if="subscription.group?.weekly_limit_usd" class="sub-usage-block">
                  <div class="flex items-center justify-between">
                    <span class="sub-usage-label">{{ t('userSubscriptions.weekly') }}</span>
                    <span class="sub-usage-value">${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{ subscription.group.weekly_limit_usd.toFixed(2) }}</span>
                  </div>
                  <div class="sub-bar-track">
                    <div class="sub-bar-fill" :class="getProgressBarClass(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd)" :style="{ width: getProgressWidth(subscription.weekly_usage_usd, subscription.group.weekly_limit_usd) }" />
                  </div>
                  <p v-if="subscription.weekly_window_start" class="text-[11px] text-gray-400 dark:text-gray-500">
                    {{ t('userSubscriptions.resetIn', { time: formatResetTime(subscription.weekly_window_start, 168) }) }}
                  </p>
                </div>

                <div v-if="subscription.group?.monthly_limit_usd" class="sub-usage-block">
                  <div class="flex items-center justify-between">
                    <span class="sub-usage-label">{{ t('userSubscriptions.monthly') }}</span>
                    <span class="sub-usage-value">${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{ subscription.group.monthly_limit_usd.toFixed(2) }}</span>
                  </div>
                  <div class="sub-bar-track">
                    <div class="sub-bar-fill" :class="getProgressBarClass(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd)" :style="{ width: getProgressWidth(subscription.monthly_usage_usd, subscription.group.monthly_limit_usd) }" />
                  </div>
                  <p v-if="subscription.monthly_window_start" class="text-[11px] text-gray-400 dark:text-gray-500">
                    {{ t('userSubscriptions.resetIn', { time: formatResetTime(subscription.monthly_window_start, 720) }) }}
                  </p>
                </div>
              </div>

              <!-- Unlimited -->
              <div v-if="!subscription.group?.daily_limit_usd && !subscription.group?.weekly_limit_usd && !subscription.group?.monthly_limit_usd" class="flex items-center gap-3 rounded-xl bg-gradient-to-r from-emerald-50 to-teal-50 px-4 py-5 dark:from-emerald-900/20 dark:to-teal-900/20">
                <span class="text-3xl text-emerald-500 dark:text-emerald-400">∞</span>
                <div>
                  <p class="text-sm font-semibold text-emerald-700 dark:text-emerald-300">{{ t('userSubscriptions.unlimited') }}</p>
                  <p class="text-xs text-emerald-600/70 dark:text-emerald-400/70">{{ t('userSubscriptions.unlimitedDesc') }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
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
import userGroupsAPI from '@/api/groups'
import { updateProfile } from '@/api/user'
import type { ActiveSubscriptionProduct, Group, UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { formatDateOnly } from '@/utils/format'
import { platformBadgeClass, platformButtonClass, platformLabel, platformGradientClass } from '@/utils/platformColors'

function platformAccentDotClass(p: string): string {
  switch (p) {
    case 'anthropic': return 'bg-orange-500'
    case 'openai': return 'bg-emerald-500'
    case 'antigravity': return 'bg-purple-500'
    case 'gemini': return 'bg-blue-500'
    default: return 'bg-gray-400'
  }
}

function productPlatform(product: ActiveSubscriptionProduct): string {
  if (product.groups.length > 0 && product.groups[0].group_platform) {
    return product.groups[0].group_platform
  }
  return ''
}

function productGradientClass(product: ActiveSubscriptionProduct): string {
  const p = productPlatform(product)
  const map: Record<string, string> = {
    anthropic: 'bg-gradient-to-r from-amber-400 to-orange-500',
    openai: 'bg-gradient-to-r from-emerald-400 to-teal-500',
    gemini: 'bg-gradient-to-r from-sky-400 to-blue-500',
    antigravity: 'bg-gradient-to-r from-violet-400 to-purple-500',
  }
  return map[p] || 'bg-gradient-to-r from-slate-300 to-slate-400 dark:from-slate-600 dark:to-slate-500'
}

function productIconBgClass(product: ActiveSubscriptionProduct): string {
  const p = productPlatform(product)
  const map: Record<string, string> = {
    anthropic: 'bg-orange-100 text-orange-600 dark:bg-orange-900/30 dark:text-orange-400',
    openai: 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400',
    gemini: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
    antigravity: 'bg-violet-100 text-violet-600 dark:bg-violet-900/30 dark:text-violet-400',
  }
  return map[p] || 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
}

function inferLimit(limit: number, dailyLimit: number, multiplier: number): number {
  if (limit > 0) return limit
  return dailyLimit > 0 ? dailyLimit * multiplier : 0
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()

const subscriptions = ref<UserSubscription[]>([])
const subscriptionProducts = ref<ActiveSubscriptionProduct[]>([])
const selectableGroups = ref<Group[]>([])
const loading = ref(true)
const savingFallback = ref(false)
const fallbackEnabled = ref(false)
const fallbackLimit = ref(0)
const fallbackGroupId = ref<number | null>(null)

const fallbackUsed = computed(() => authStore.user?.subscription_balance_fallback_used_usd || 0)
const fallbackRemaining = computed(() => Math.max((fallbackLimit.value || 0) - fallbackUsed.value, 0))
const fallbackGroupOptions = computed(() => {
  return selectableGroups.value
    .filter((group) => group.status === 'active' && group.subscription_type !== 'subscription')
    .map((group) => ({
      value: group.id,
      label: group.name || `Group #${group.id}`
    }))
})

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
    const [legacySubscriptions, products, profile, groups] = await Promise.all([
      subscriptionsAPI.getActiveSubscriptions(),
      subscriptionProductsAPI.getActive(),
      authStore.refreshUser(),
      userGroupsAPI.getAvailable()
    ])
    subscriptions.value = legacySubscriptions
    subscriptionProducts.value = products
    selectableGroups.value = groups
    fallbackEnabled.value = Boolean(profile.subscription_balance_fallback_enabled)
    fallbackLimit.value = profile.subscription_balance_fallback_limit_usd || 0
    fallbackGroupId.value = profile.subscription_balance_fallback_group_id || null
  } catch (error) {
    console.error('Failed to load subscriptions:', error)
    appStore.showError(t('userSubscriptions.failedToLoad'))
  } finally {
    loading.value = false
  }
}

async function saveBalanceFallbackSettings() {
  const hasPositiveLimit = (fallbackLimit.value || 0) > 0
  const hasFallbackGroup = !!fallbackGroupId.value

  if (fallbackEnabled.value && !hasFallbackGroup) {
    appStore.showError(t('userSubscriptions.balanceFallback.groupRequired', 'Please select a balance group'))
    fallbackEnabled.value = false
    return
  }
  if (fallbackEnabled.value && !hasPositiveLimit) {
    appStore.showError(t('userSubscriptions.balanceFallback.limitRequired', 'Please set a positive fallback limit'))
    fallbackEnabled.value = false
    return
  }
  savingFallback.value = true
  try {
    if (!fallbackEnabled.value && fallbackLimit.value < 0) {
      fallbackLimit.value = 0
    }
    const updated = await updateProfile({
      subscription_balance_fallback_enabled: fallbackEnabled.value,
      subscription_balance_fallback_limit_usd: Math.max(fallbackLimit.value || 0, 0),
      subscription_balance_fallback_group_id: fallbackEnabled.value ? fallbackGroupId.value : null
    })
    authStore.user = updated
    fallbackEnabled.value = Boolean(updated.subscription_balance_fallback_enabled)
    fallbackLimit.value = updated.subscription_balance_fallback_limit_usd || 0
    fallbackGroupId.value = updated.subscription_balance_fallback_group_id || null
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    console.error('Failed to save subscription balance fallback:', error)
    appStore.showError(t('common.error'))
    fallbackEnabled.value = Boolean(authStore.user?.subscription_balance_fallback_enabled)
    fallbackLimit.value = authStore.user?.subscription_balance_fallback_limit_usd || 0
    fallbackGroupId.value = authStore.user?.subscription_balance_fallback_group_id || null
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

<style scoped>
.sub-status-badge {
  @apply rounded-full px-2.5 py-0.5 text-[11px] font-semibold;
}
.sub-status-active {
  @apply bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400;
}
.sub-status-expired {
  @apply bg-gray-100 text-gray-500 dark:bg-dark-600 dark:text-gray-400;
}
.sub-status-revoked {
  @apply bg-red-100 text-red-700 dark:bg-red-500/15 dark:text-red-400;
}
.sub-usage-block {
  @apply space-y-1.5;
}
.sub-usage-label {
  @apply text-xs font-medium text-gray-500 dark:text-gray-400;
}
.sub-usage-value {
  @apply text-xs font-semibold tabular-nums text-gray-700 dark:text-gray-200;
}
.sub-bar-track {
  @apply h-2 w-full overflow-hidden rounded-full bg-gray-100 dark:bg-dark-600;
}
.sub-bar-fill {
  @apply h-full rounded-full transition-all duration-500 ease-out;
}
.sub-carryover-tag {
  @apply rounded bg-amber-50 px-1.5 py-0.5 text-[10px] tabular-nums text-amber-600 dark:bg-amber-500/10 dark:text-amber-400;
}
</style>
