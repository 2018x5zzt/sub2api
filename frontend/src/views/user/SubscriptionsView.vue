<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <template v-else>
      <!-- Balance Fallback Card -->
      <div class="card overflow-hidden p-0">
        <div class="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex items-start gap-3">
            <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-emerald-100 dark:bg-emerald-900/30">
              <Icon name="creditCard" size="sm" class="text-emerald-600 dark:text-emerald-400" />
            </div>
            <div class="min-w-0">
              <h2 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('userSubscriptions.balanceFallback.title') }}
              </h2>
              <p class="mt-0.5 text-xs leading-relaxed text-gray-500 dark:text-dark-400">
                {{ t('userSubscriptions.balanceFallback.description') }}
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
              />
              <div class="h-6 w-11 rounded-full bg-gray-200 transition-colors after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:bg-white after:shadow-sm after:transition-all peer-checked:bg-emerald-500 peer-checked:after:translate-x-full peer-disabled:opacity-50 dark:bg-dark-600 dark:after:bg-dark-300 dark:peer-checked:bg-emerald-600" />
            </div>
            <span class="text-xs font-medium" :class="fallbackEnabled ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400 dark:text-gray-500'">
              {{ fallbackEnabled ? t('common.enabled') : t('common.disabled') }}
            </span>
          </label>
        </div>
        <div v-if="fallbackEnabled" class="border-t border-gray-100 bg-gray-50/50 px-5 py-4 dark:border-dark-700 dark:bg-dark-800/50">
          <div class="flex flex-wrap items-end gap-4">
            <label class="block w-56">
              <span class="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('userSubscriptions.balanceFallback.group') }}
              </span>
              <select v-model.number="fallbackGroupId" class="input" :disabled="savingFallback">
                <option :value="null">{{ t('userSubscriptions.balanceFallback.selectGroup') }}</option>
                <option v-for="option in fallbackGroupOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>
            <label class="block w-48">
              <span class="mb-1.5 block text-[11px] font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('userSubscriptions.balanceFallback.limit') }}
              </span>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-gray-400">$</span>
                <input v-model.number="fallbackLimit" type="number" min="0" step="0.01" class="input pl-7" :disabled="savingFallback" />
              </div>
            </label>
            <span v-if="fallbackLimit > 0" class="pb-2 text-xs tabular-nums text-gray-500 dark:text-dark-400">
              {{ t('userSubscriptions.balanceFallback.usage', { used: fallbackUsed.toFixed(2), remaining: fallbackRemaining.toFixed(2) }) }}
            </span>
            <span v-else class="pb-2 text-xs text-gray-400 dark:text-gray-500">
              {{ t('userSubscriptions.balanceFallback.setLimitHint') }}
            </span>
          </div>
          <p class="mt-3 text-xs leading-relaxed text-amber-700 dark:text-amber-300">
            {{ t('userSubscriptions.balanceFallback.negativeBalanceHint') }}
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2 border-t border-gray-100 px-5 py-4 dark:border-dark-700">
          <button type="button" class="btn btn-primary btn-sm" :disabled="savingFallback" @click="saveBalanceFallbackSettings">
            {{ savingFallback ? t('common.saving') : t('common.save') }}
          </button>
          <button type="button" class="btn btn-secondary btn-sm" :disabled="savingFallback" @click="resetBalanceFallbackForm">
            {{ t('common.cancel') }}
          </button>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="subscriptions.length === 0" class="card p-12 text-center">
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
          v-for="subscription in subscriptions"
          :key="subscription.subscription_id"
          class="overflow-hidden rounded-2xl border bg-white dark:bg-dark-800"
          :class="platformBorderClass(productPlatform(subscription))"
        >
          <!-- Header -->
          <div
            class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700"
          >
            <div class="flex items-center gap-3">
              <div :class="['h-1.5 w-1.5 shrink-0 rounded-full', platformAccentDotClass(productPlatform(subscription))]" />
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="font-semibold text-gray-900 dark:text-white">
                    {{ subscription.name }}
                  </h3>
                  <span :class="['rounded-md border px-2 py-0.5 text-[11px] font-medium', platformBadgeClass(productPlatform(subscription))]">
                    {{ platformLabel(productPlatform(subscription)) }}
                  </span>
                </div>
                <p v-if="subscription.description" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">
                  {{ subscription.description }}
                </p>
                <div v-if="subscription.groups.length > 0" class="mt-2 flex flex-wrap gap-1.5">
                  <span
                    v-for="group in subscription.groups"
                    :key="group.group_id"
                    :class="['rounded-md border px-2 py-0.5 text-[11px] font-medium', platformBadgeClass(group.group_platform || '')]"
                  >
                    {{ group.group_name || `Group #${group.group_id}` }}
                  </span>
                </div>
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
                v-if="subscription.status === 'active' && primaryGroupID(subscription)"
                :class="['rounded-lg px-3 py-1.5 text-xs font-semibold text-white transition-colors', platformButtonClass(productPlatform(subscription))]"
                @click="router.push({ path: '/purchase', query: { tab: 'subscription', group: String(primaryGroupID(subscription)) } })"
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
            <div v-if="dailyEffectiveLimit(subscription)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.daily') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.daily_usage_usd || 0).toFixed(2) }} / ${{
                    dailyEffectiveLimit(subscription).toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.daily_usage_usd,
                      dailyEffectiveLimit(subscription)
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.daily_usage_usd,
                      dailyEffectiveLimit(subscription)
                    )
                  }"
                ></div>
              </div>
              <p
                v-if="subscription.daily_carryover_in_usd > 0"
                class="text-xs text-gray-500 dark:text-dark-400"
              >
                {{ t('userSubscriptions.balanceFallback.usage', { used: (subscription.daily_carryover_in_usd - subscription.daily_carryover_remaining_usd).toFixed(2), remaining: subscription.daily_carryover_remaining_usd.toFixed(2) }) }}
              </p>
            </div>

            <!-- Weekly Usage -->
            <div v-if="weeklyEffectiveLimit(subscription)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.weekly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.weekly_usage_usd || 0).toFixed(2) }} / ${{
                    weeklyEffectiveLimit(subscription).toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.weekly_usage_usd,
                      weeklyEffectiveLimit(subscription)
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.weekly_usage_usd,
                      weeklyEffectiveLimit(subscription)
                    )
                  }"
                ></div>
              </div>
            </div>

            <!-- Monthly Usage -->
            <div v-if="monthlyEffectiveLimit(subscription)" class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('userSubscriptions.monthly') }}
                </span>
                <span class="text-sm text-gray-500 dark:text-dark-400">
                  ${{ (subscription.monthly_usage_usd || 0).toFixed(2) }} / ${{
                    monthlyEffectiveLimit(subscription).toFixed(2)
                  }}
                </span>
              </div>
              <div class="relative h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="absolute inset-y-0 left-0 rounded-full transition-all duration-300"
                  :class="
                    getProgressBarClass(
                      subscription.monthly_usage_usd,
                      monthlyEffectiveLimit(subscription)
                    )
                  "
                  :style="{
                    width: getProgressWidth(
                      subscription.monthly_usage_usd,
                      monthlyEffectiveLimit(subscription)
                    )
                  }"
                ></div>
              </div>
            </div>

            <!-- No limits configured - Unlimited badge -->
            <div
              v-if="
                !dailyEffectiveLimit(subscription) &&
                !weeklyEffectiveLimit(subscription) &&
                !monthlyEffectiveLimit(subscription)
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
import { subscriptionProductsAPI } from '@/api/subscriptionProducts'
import userGroupsAPI from '@/api/groups'
import { updateProfile } from '@/api/user'
import type { ActiveSubscriptionProduct, Group } from '@/types'
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

const subscriptions = ref<ActiveSubscriptionProduct[]>([])
const loading = ref(true)
const selectableGroups = ref<Group[]>([])
const savingFallback = ref(false)
const fallbackEnabled = ref(false)
const fallbackLimit = ref(0)
const fallbackGroupId = ref<number | null>(null)

const fallbackUsed = computed(() => authStore.user?.subscription_balance_fallback_used_usd || 0)
const fallbackRemaining = computed(() => Math.max((fallbackLimit.value || 0) - fallbackUsed.value, 0))
const fallbackGroupOptions = computed(() =>
  selectableGroups.value
    .filter((group) => group.status === 'active' && group.subscription_type === 'standard')
    .map((group) => ({ value: group.id, label: group.name || `Group #${group.id}` }))
)

async function loadSubscriptions() {
  try {
    loading.value = true
    const [subs, profile, groups] = await Promise.all([
      subscriptionProductsAPI.getActive(),
      authStore.refreshUser(),
      userGroupsAPI.getAvailable()
    ])
    subscriptions.value = subs
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

function resetBalanceFallbackForm() {
  fallbackEnabled.value = Boolean(authStore.user?.subscription_balance_fallback_enabled)
  fallbackLimit.value = authStore.user?.subscription_balance_fallback_limit_usd || 0
  fallbackGroupId.value = authStore.user?.subscription_balance_fallback_group_id || null
}

async function saveBalanceFallbackSettings() {
  const hasPositiveLimit = (fallbackLimit.value || 0) > 0
  const hasSelectableFallbackGroup = fallbackGroupOptions.value.some(
    (option) => option.value === fallbackGroupId.value
  )

  if (fallbackEnabled.value && !hasSelectableFallbackGroup) {
    appStore.showError(t('userSubscriptions.balanceFallback.groupRequired'))
    return
  }
  if (fallbackEnabled.value && !hasPositiveLimit) {
    appStore.showError(t('userSubscriptions.balanceFallback.limitRequired'))
    return
  }
  savingFallback.value = true
  try {
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
    resetBalanceFallbackForm()
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

function productPlatform(product: ActiveSubscriptionProduct): string {
  return product.groups.find((group) => group.group_platform)?.group_platform || ''
}

function primaryGroupID(product: ActiveSubscriptionProduct): number | null {
  return product.groups[0]?.group_id ?? null
}

function dailyEffectiveLimit(product: ActiveSubscriptionProduct): number {
  return (product.daily_limit_usd || 0) + (product.daily_carryover_in_usd || 0)
}

function weeklyEffectiveLimit(product: ActiveSubscriptionProduct): number {
  return product.weekly_limit_usd || (product.daily_limit_usd ? product.daily_limit_usd * 7 : 0)
}

function monthlyEffectiveLimit(product: ActiveSubscriptionProduct): number {
  return product.monthly_limit_usd || (product.daily_limit_usd ? product.daily_limit_usd * 30 : 0)
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

onMounted(() => {
  loadSubscriptions()
})
</script>
