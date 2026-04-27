<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <template v-else-if="detail">
        <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
          <div class="card p-5">
            <p class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-dark-400">
              <Icon name="dollar" size="sm" class="text-primary-500" />
              {{ t('affiliate.stats.rebateRate') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-primary-600 dark:text-primary-400">
              {{ formattedRebateRate }}<span class="ml-0.5 text-base font-medium">%</span>
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ t('affiliate.stats.rebateRateHint') }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.effectiveInvitees') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ effectiveInviteeCount.toLocaleString() }}
            </p>
            <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
              {{ t('affiliate.stats.effectiveInviteesHint') }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.invitedUsers') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ detail.aff_count.toLocaleString() }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.availableQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
              {{ formatCurrency(detail.aff_quota) }}
            </p>
          </div>
          <div class="card p-5">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.stats.totalQuota') }}</p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCurrency(detail.aff_history_quota) }}
            </p>
            <p v-if="detail.aff_frozen_quota > 0" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
              {{ t('affiliate.stats.frozenQuota') }}: {{ formatCurrency(detail.aff_frozen_quota) }}
            </p>
          </div>
        </div>

        <div class="card overflow-hidden">
          <div class="border-b border-gray-100 p-6 dark:border-dark-800">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <p class="text-sm font-medium text-primary-600 dark:text-primary-400">{{ t('affiliate.title') }}</p>
                <h2 class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ t('affiliate.ladder.slogan') }}</h2>
                <p class="mt-2 max-w-3xl text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.ladder.description') }}</p>
              </div>
              <div class="flex flex-wrap gap-2">
                <span class="inline-flex items-center rounded-full bg-emerald-50 px-3 py-1 text-sm font-medium text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300">
                  {{ t('affiliate.ladder.maxRate') }}
                </span>
                <span class="inline-flex items-center rounded-full bg-sky-50 px-3 py-1 text-sm font-medium text-sky-700 dark:bg-sky-900/20 dark:text-sky-300">
                  {{ t('affiliate.ladder.balancePayout') }}
                </span>
              </div>
            </div>

            <div class="mt-5">
              <div class="flex items-center justify-between text-xs text-gray-500 dark:text-dark-400">
                <span>{{ t('affiliate.ladder.currentTier', { tier: currentTierName }) }}</span>
                <span v-if="nextTier">{{ t('affiliate.ladder.nextTier', { count: nextTier.min }) }}</span>
                <span v-else>{{ t('affiliate.ladder.topTier') }}</span>
              </div>
              <div class="mt-2 h-2 rounded-full bg-gray-100 dark:bg-dark-800">
                <div class="h-2 rounded-full bg-primary-500 transition-all" :style="{ width: `${tierProgressPercent}%` }"></div>
              </div>
            </div>
          </div>

          <div class="grid gap-0 lg:grid-cols-[1.4fr_1fr]">
            <div class="border-b border-gray-100 p-6 dark:border-dark-800 lg:border-b-0 lg:border-r">
              <div class="flex items-center justify-between gap-3">
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.ladder.tierTitle') }}</h3>
                <span class="text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.ladder.effectiveInviteeLabel') }}</span>
              </div>
              <div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
                <div
                  v-for="tier in rebateTiers"
                  :key="tier.name"
                  class="rounded-lg border p-3"
                  :class="tier.name === currentTierName
                    ? 'border-primary-300 bg-primary-50 text-primary-900 dark:border-primary-700 dark:bg-primary-900/20 dark:text-primary-100'
                    : 'border-gray-200 bg-white text-gray-800 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-200'"
                >
                  <div class="flex items-center justify-between gap-2">
                    <p class="text-sm font-semibold">{{ tier.name }}</p>
                    <p class="text-base font-semibold">{{ tier.rate }}%</p>
                  </div>
                  <p class="mt-1 text-xs opacity-75">{{ tier.range }}</p>
                </div>
              </div>
            </div>

            <div class="p-6">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.ladder.factorTitle') }}</h3>
              <div class="mt-4 space-y-2">
                <div v-for="factor in skuFactors" :key="factor.name" class="flex items-center justify-between rounded-lg border border-gray-200 px-3 py-2 text-sm dark:border-dark-700">
                  <span class="font-medium text-gray-700 dark:text-gray-200">{{ factor.name }}</span>
                  <span class="font-semibold text-gray-900 dark:text-white">{{ factor.value }}</span>
                </div>
              </div>
              <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-dark-400">{{ t('affiliate.ladder.factorHint') }}</p>
            </div>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.concurrency.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.concurrency.description') }}</p>
            </div>
            <div class="grid min-w-full gap-3 sm:grid-cols-3 md:min-w-[520px]">
              <div v-for="item in concurrencyRewards" :key="item.label" class="rounded-lg border border-gray-200 px-3 py-2 dark:border-dark-700">
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ item.label }}</p>
                <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ item.value }}</p>
              </div>
            </div>
          </div>
          <p class="mt-4 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.concurrency.smallAccountHint') }}</p>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.title') }}</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.description') }}</p>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.yourCode') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm font-semibold text-gray-900 dark:text-white">{{ detail.aff_code }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyCode') }}</span>
                </button>
              </div>
            </div>

            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.inviteLink') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">{{ inviteLink }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyLink') }}</span>
                </button>
              </div>
            </div>
          </div>

          <div class="mt-5 rounded-xl border border-primary-200 bg-primary-50 p-4 dark:border-primary-900/40 dark:bg-primary-900/20">
            <p class="text-sm font-medium text-primary-800 dark:text-primary-200">{{ t('affiliate.tips.title') }}</p>
            <ul class="mt-2 space-y-1 text-sm text-primary-700 dark:text-primary-300">
              <li>1. {{ t('affiliate.tips.line1') }}</li>
              <li>2. {{ t('affiliate.tips.line2', { rate: `${formattedRebateRate}%` }) }}</li>
              <li>3. {{ t('affiliate.tips.line3') }}</li>
              <li v-if="detail.aff_frozen_quota > 0">4. {{ t('affiliate.tips.line4') }}</li>
            </ul>
          </div>
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.transfer.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.transfer.description') }}</p>
            </div>
            <button class="btn btn-primary" :disabled="transferring || detail.aff_quota <= 0" @click="transferQuota">
              <Icon v-if="transferring" name="refresh" size="sm" class="animate-spin" />
              <Icon v-else name="dollar" size="sm" />
              <span>{{ transferring ? t('affiliate.transfer.transferring') : t('affiliate.transfer.button') }}</span>
            </button>
          </div>
          <p v-if="detail.aff_quota <= 0" class="mt-3 text-sm text-amber-600 dark:text-amber-400">
            {{ t('affiliate.transfer.empty') }}
          </p>
        </div>

        <div class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.email') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.username') }}</th>
                  <th class="px-3 py-2 text-right font-medium">{{ t('affiliate.invitees.columns.rebate') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in detail.invitees"
                  :key="item.user_id"
                  class="border-b border-gray-100 last:border-b-0 dark:border-dark-800"
                >
                  <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                  <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.total_rebate) }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type { UserAffiliateDetail } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const transferring = ref(false)
const detail = ref<UserAffiliateDetail | null>(null)

const rebateTiers = computed(() => [
  { name: t('affiliate.tiers.bronze'), min: 1, max: 2, rate: 5, range: t('affiliate.tiers.range', { range: '1-2' }) },
  { name: t('affiliate.tiers.silver'), min: 3, max: 9, rate: 8, range: t('affiliate.tiers.range', { range: '3-9' }) },
  { name: t('affiliate.tiers.gold'), min: 10, max: 29, rate: 12, range: t('affiliate.tiers.range', { range: '10-29' }) },
  { name: t('affiliate.tiers.platinum'), min: 30, max: 49, rate: 15, range: t('affiliate.tiers.range', { range: '30-49' }) },
  { name: t('affiliate.tiers.diamond'), min: 50, max: null, rate: 20, range: t('affiliate.tiers.rangeAbove', { count: 50 }) }
])

const skuFactors = computed(() => [
  { name: t('affiliate.ladder.balanceDaily'), value: '100%' },
  { name: t('affiliate.ladder.weekly'), value: '60%' },
  { name: t('affiliate.ladder.monthly'), value: '30%' }
])

const concurrencyRewards = computed(() => [
  { label: t('affiliate.concurrency.defaultLabel'), value: t('affiliate.concurrency.defaultValue') },
  { label: t('affiliate.concurrency.oneInviteLabel'), value: t('affiliate.concurrency.oneInviteValue') },
  { label: t('affiliate.concurrency.fiveInvitesLabel'), value: t('affiliate.concurrency.fiveInvitesValue') }
])

const inviteLink = computed(() => {
  if (!detail.value) return ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(detail.value.aff_code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(detail.value.aff_code)}`
})

const effectiveInviteeCount = computed(() => detail.value?.effective_invitee_count ?? 0)

const currentTier = computed(() => {
  const count = effectiveInviteeCount.value
  return [...rebateTiers.value].reverse().find((tier) => count >= tier.min) ?? null
})

const currentTierName = computed(() => currentTier.value?.name ?? t('affiliate.tiers.base'))

const nextTier = computed(() => {
  const count = effectiveInviteeCount.value
  return rebateTiers.value.find((tier) => count < tier.min) ?? null
})

const tierProgressPercent = computed(() => {
  const count = effectiveInviteeCount.value
  if (!nextTier.value) return 100
  const previousMin = currentTier.value?.min ?? 0
  const span = Math.max(nextTier.value.min - previousMin, 1)
  return Math.min(100, Math.max(0, ((count - previousMin) / span) * 100))
})

const formattedRebateRate = computed(() => {
  const value = detail.value?.effective_rebate_rate_percent ?? 0
  const rounded = Math.round(value * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})

function extractMessage(error: unknown, fallback: string): string {
  const err = error as { message?: string; response?: { data?: { detail?: string; message?: string } } }
  return err.response?.data?.detail || err.response?.data?.message || err.message || fallback
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) loading.value = true
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) loading.value = false
  }
}

async function copyCode(): Promise<void> {
  if (!detail.value?.aff_code) return
  await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied'))
}

async function copyInviteLink(): Promise<void> {
  if (!inviteLink.value) return
  await copyToClipboard(inviteLink.value, t('affiliate.linkCopied'))
}

async function transferQuota(): Promise<void> {
  if (!detail.value || detail.value.aff_quota <= 0 || transferring.value) return
  transferring.value = true
  try {
    const response = await userAPI.transferAffiliateQuota()
    const transferred = response.transferred ?? response.transferred_quota ?? 0
    appStore.showSuccess(t('affiliate.transfer.success', { amount: formatCurrency(transferred) }))
    await Promise.all([
      loadAffiliateDetail(true),
      authStore.refreshUser().catch(() => undefined)
    ])
  } catch (error) {
    appStore.showError(extractMessage(error, t('affiliate.transferFailed')))
  } finally {
    transferring.value = false
  }
}

onMounted(() => {
  void loadAffiliateDetail()
})
</script>
