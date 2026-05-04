<template>
  <div v-if="summary" class="relative" ref="containerRef">
    <!-- Mini Display Button -->
    <button
      @click="togglePopover"
      @mouseenter="handleMouseEnter"
      @mouseleave="handleMouseLeave"
      class="flex cursor-pointer items-center gap-2 rounded-xl bg-purple-50 px-3 py-1.5 transition-colors hover:bg-purple-100 dark:bg-purple-900/20 dark:hover:bg-purple-900/30"
      :title="t('productSubscription.viewDetails')"
    >
      <Icon name="creditCard" size="sm" class="text-purple-600 dark:text-purple-400" />
      <span v-if="summary.active_count > 0" class="h-2 w-2 rounded-full bg-emerald-500"></span>
      <span class="text-xs font-semibold text-purple-700 dark:text-purple-300">
        {{ summary.active_count }}
      </span>
    </button>

    <!-- Popover -->
    <transition name="dropdown">
      <div
        v-if="popoverOpen"
        class="absolute right-0 z-50 mt-2 w-[380px] overflow-hidden rounded-2xl border border-gray-200/80 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-800"
        @mouseenter="handleMouseEnter"
        @mouseleave="handleMouseLeave"
      >
        <!-- Header -->
        <div class="px-5 pt-5 pb-3">
          <h3 class="text-base font-bold text-gray-900 dark:text-white">
            {{ t('productSubscription.mySubscriptions') }}
          </h3>
          <p class="mt-0.5 text-xs text-gray-400 dark:text-dark-400">
            {{ t('productSubscription.activeCount', { count: summary.active_count }) }}
          </p>
        </div>

        <!-- Products -->
        <div class="max-h-[360px] overflow-y-auto px-5">
          <div
            v-for="(product, idx) in summary.products"
            :key="product.subscription_id"
            class="py-4"
            :class="{ 'border-t border-gray-100 dark:border-dark-700/50': idx > 0 }"
          >
            <!-- Product name + days -->
            <div class="mb-3 flex items-center justify-between">
              <span class="inline-flex items-center gap-1.5 rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 ring-1 ring-inset ring-emerald-600/10 dark:bg-emerald-500/10 dark:text-emerald-400 dark:ring-emerald-500/20">
                <PlatformIcon :platform="getProductPlatform(product) as any" size="xs" />
                {{ product.name }}
              </span>
              <span
                v-if="product.expires_at"
                class="text-xs font-medium"
                :class="getDaysClass(product.expires_at)"
              >
                {{ formatDaysRemaining(product.expires_at) }}
              </span>
            </div>

            <!-- Usage rows -->
            <div class="space-y-2.5">
              <!-- Daily -->
              <div class="flex items-center gap-3">
                <span class="w-8 shrink-0 text-xs text-gray-400 dark:text-gray-500">{{ t('productSubscription.daily') }}</span>
                <div class="h-2 min-w-0 flex-1 rounded-full bg-gray-100 dark:bg-dark-600">
                  <div
                    class="h-2 rounded-full transition-all duration-300"
                    :class="barClass(product.daily_usage_usd, dailyEffectiveLimit(product))"
                    :style="{ width: barWidth(product.daily_usage_usd, dailyEffectiveLimit(product)) }"
                  />
                </div>
                <span class="w-[90px] shrink-0 text-right text-xs tabular-nums text-gray-500 dark:text-gray-400">
                  {{ fmtUsage(product.daily_usage_usd, dailyEffectiveLimit(product)) }}
                </span>
              </div>
              <!-- Weekly -->
              <div class="flex items-center gap-3">
                <span class="w-8 shrink-0 text-xs text-gray-400 dark:text-gray-500">{{ t('productSubscription.weekly') }}</span>
                <div class="h-2 min-w-0 flex-1 rounded-full bg-gray-100 dark:bg-dark-600">
                  <div
                    class="h-2 rounded-full transition-all duration-300"
                    :class="barClass(product.weekly_usage_usd, effectiveLimit(product.weekly_limit_usd, product.daily_limit_usd, 7))"
                    :style="{ width: barWidth(product.weekly_usage_usd, effectiveLimit(product.weekly_limit_usd, product.daily_limit_usd, 7)) }"
                  />
                </div>
                <span class="w-[90px] shrink-0 text-right text-xs tabular-nums text-gray-500 dark:text-gray-400">
                  {{ fmtUsage(product.weekly_usage_usd, effectiveLimit(product.weekly_limit_usd, product.daily_limit_usd, 7)) }}
                </span>
              </div>
              <!-- Monthly -->
              <div class="flex items-center gap-3">
                <span class="w-8 shrink-0 text-xs text-gray-400 dark:text-gray-500">{{ t('productSubscription.monthly') }}</span>
                <div class="h-2 min-w-0 flex-1 rounded-full bg-gray-100 dark:bg-dark-600">
                  <div
                    class="h-2 rounded-full transition-all duration-300"
                    :class="barClass(product.monthly_usage_usd, effectiveLimit(product.monthly_limit_usd, product.daily_limit_usd, 30))"
                    :style="{ width: barWidth(product.monthly_usage_usd, effectiveLimit(product.monthly_limit_usd, product.daily_limit_usd, 30)) }"
                  />
                </div>
                <span class="w-[90px] shrink-0 text-right text-xs tabular-nums text-gray-500 dark:text-gray-400">
                  {{ fmtUsage(product.monthly_usage_usd, effectiveLimit(product.monthly_limit_usd, product.daily_limit_usd, 30)) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer link -->
        <div class="border-t border-gray-100 px-5 py-3 text-center dark:border-dark-700">
          <router-link
            to="/subscriptions"
            class="text-sm font-medium text-emerald-600 hover:text-emerald-700 dark:text-emerald-400 dark:hover:text-emerald-300"
            @click="popoverOpen = false"
          >
            {{ t('productSubscription.viewAll') }}
          </router-link>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { subscriptionProductsAPI } from '@/api/subscriptionProducts'
import type { SubscriptionProductSummary, ActiveSubscriptionProduct } from '@/types'

const { t } = useI18n()

const containerRef = ref<HTMLElement | null>(null)
const popoverOpen = ref(false)
const summary = ref<SubscriptionProductSummary | null>(null)
let hoverTimer: ReturnType<typeof setTimeout> | null = null

function getProductPlatform(product: ActiveSubscriptionProduct): string {
  if (product.groups && product.groups.length > 0 && product.groups[0].group_platform) {
    return product.groups[0].group_platform
  }
  return ''
}

async function loadSummary() {
  try {
    const data = await subscriptionProductsAPI.getProgress()
    summary.value = data
  } catch (e) {
    console.error('[ProductSubscriptionMini] Failed to load:', e)
  }
}

function dailyEffectiveLimit(product: ActiveSubscriptionProduct): number {
  return (product.daily_limit_usd || 0) + (product.daily_carryover_in_usd || 0)
}

function effectiveLimit(limit: number, dailyLimit: number, multiplier: number): number {
  if (limit > 0) return limit
  return dailyLimit > 0 ? dailyLimit * multiplier : 0
}

function barWidth(used: number | undefined, limit: number): string {
  const u = Number(used || 0)
  if (limit <= 0) return u > 0 ? '100%' : '0%'
  return `${Math.min((u / limit) * 100, 100)}%`
}

function barClass(used: number | undefined, limit: number): string {
  if (limit <= 0) return 'bg-gray-400'
  const pct = ((used || 0) / limit) * 100
  if (pct >= 90) return 'bg-red-500'
  if (pct >= 70) return 'bg-orange-500'
  return 'bg-emerald-500'
}

function fmtUsage(used: number | undefined, limit: number): string {
  const u = `$${(used || 0).toFixed(2)}`
  const l = limit > 0 ? `$${limit.toFixed(2)}` : '∞'
  return `${u}/${l}`
}

function formatDaysRemaining(expiresAt: string): string {
  const diff = new Date(expiresAt).getTime() - Date.now()
  if (diff < 0) return t('productSubscription.expired')
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return t('productSubscription.expiresToday')
  return t('productSubscription.daysRemaining', { days })
}

function getDaysClass(expiresAt: string): string {
  const days = Math.ceil((new Date(expiresAt).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-500 dark:text-dark-400'
}

function togglePopover() {
  popoverOpen.value = !popoverOpen.value
}

function handleMouseEnter() {
  if (hoverTimer) { clearTimeout(hoverTimer); hoverTimer = null }
  popoverOpen.value = true
}

function handleMouseLeave() {
  hoverTimer = setTimeout(() => { popoverOpen.value = false }, 200)
}

function handleClickOutside(event: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(event.target as Node)) {
    popoverOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  loadSummary()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  if (hoverTimer) clearTimeout(hoverTimer)
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
