<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-72">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="subscriptionSearchQuery"
                type="text"
                :placeholder="t('admin.subscriptionProducts.searchSubscriptionsPlaceholder', 'Search users or products')"
                class="input pl-10"
                @input="debounceSubscriptionSearch"
              />
            </div>

            <Select
              v-model="subscriptionStatusFilter"
              :options="statusFilterOptions"
              :placeholder="t('admin.subscriptionProducts.allStatus', 'All Status')"
              class="w-40"
            />
            <Select
              v-model="subscriptionProductFilter"
              :options="productFilterOptions"
              :placeholder="t('admin.subscriptionProducts.allProducts', 'All Products')"
              class="w-56"
            />
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="loadUserSubscriptions"
              :disabled="userSubscriptionsLoading"
              class="btn btn-secondary"
              :title="t('common.refresh', 'Refresh')"
            >
              <Icon name="refresh" size="md" :class="userSubscriptionsLoading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="userSubscriptionColumns"
          :data="userSubscriptions"
          :loading="userSubscriptionsLoading"
        >
          <template #cell-user="{ row }">
            <div class="flex min-w-[200px] items-center gap-3">
              <span
                class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white"
                :class="avatarColor(row.user_email)"
              >
                {{ avatarInitial(row.user_email) }}
              </span>
              <div class="min-w-0">
                <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ row.user_email }}</div>
                <div class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">
                  {{ row.user_username || '-' }} #{{ row.user_id }}
                </div>
              </div>
            </div>
          </template>

          <template #cell-product="{ row }">
            <span class="inline-flex items-center gap-1.5 rounded-md bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 ring-1 ring-inset ring-emerald-600/10 dark:bg-emerald-500/10 dark:text-emerald-400 dark:ring-emerald-500/20">
              {{ row.product_name }}
            </span>
          </template>

          <template #cell-usage="{ row }">
            <div class="min-w-[260px] space-y-2">
              <div class="usage-row">
                <span class="usage-label">{{ t('admin.subscriptionProducts.daily', 'Daily') }}</span>
                <div class="usage-bar-track">
                  <div class="usage-bar-fill bg-emerald-500" :style="{ width: usagePercent(row.daily_usage_usd, row.daily_limit_usd) + '%' }" />
                </div>
                <span class="usage-text">{{ formatUsageUSD(row.daily_usage_usd) }} / {{ formatLimitUSD(row.daily_limit_usd) }}</span>
                <span v-if="Number(row.daily_carryover_in_usd || 0) > 0" class="usage-tag-inline">{{ t('admin.subscriptionProducts.carryoverIn', 'Carry') }} {{ formatUsageUSD(row.daily_carryover_in_usd) }}</span>
              </div>
              <div class="usage-row">
                <span class="usage-label">{{ t('admin.subscriptionProducts.weekly', 'Weekly') }}</span>
                <div class="usage-bar-track">
                  <div class="usage-bar-fill bg-sky-500" :style="{ width: usagePercent(row.weekly_usage_usd, inferLimit(row.weekly_limit_usd, row.daily_limit_usd, 7)) + '%' }" />
                </div>
                <span class="usage-text">{{ formatUsageUSD(row.weekly_usage_usd) }} / {{ formatLimitUSD(inferLimit(row.weekly_limit_usd, row.daily_limit_usd, 7)) }}</span>
              </div>
              <div class="usage-row">
                <span class="usage-label">{{ t('admin.subscriptionProducts.monthly', 'Monthly') }}</span>
                <div class="usage-bar-track">
                  <div class="usage-bar-fill bg-violet-500" :style="{ width: usagePercent(row.monthly_usage_usd, inferLimit(row.monthly_limit_usd, row.daily_limit_usd, 30)) + '%' }" />
                </div>
                <span class="usage-text">{{ formatUsageUSD(row.monthly_usage_usd) }} / {{ formatLimitUSD(inferLimit(row.monthly_limit_usd, row.daily_limit_usd, 30)) }}</span>
              </div>
            </div>
          </template>

          <template #cell-period="{ row }">
            <div class="min-w-[100px] text-xs">
              <div class="font-medium text-gray-700 dark:text-gray-300">{{ formatDateOnly(row.expires_at) }}</div>
              <div class="mt-0.5 text-gray-400 dark:text-gray-500">{{ remainingDays(row.expires_at) }}</div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span class="status-badge" :class="statusBadgeClass(value)">
              <span class="status-dot" :class="statusDotClass(value)" />
              {{ statusLabel(value) }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-4">
              <button
                type="button"
                class="action-btn"
                :title="t('admin.subscriptions.adjustSubscription', 'Adjust')"
                @click="openAdjustDialog(row)"
              >
                <Icon name="calendar" size="md" />
                <span>{{ t('admin.subscriptions.adjustSubscription', 'Adjust') }}</span>
              </button>
              <button
                type="button"
                class="action-btn"
                :title="t('admin.subscriptionProducts.resetQuota', 'Reset Quota')"
                @click="openResetQuotaDialog(row)"
              >
                <Icon name="refresh" size="md" />
                <span>{{ t('admin.subscriptionProducts.resetQuota', 'Reset Quota') }}</span>
              </button>
              <button
                type="button"
                class="action-btn text-red-500 hover:text-red-600 dark:text-red-400 dark:hover:text-red-300"
                :title="t('admin.subscriptions.revokeSubscription', 'Revoke')"
                @click="openRevokeConfirm(row)"
              >
                <Icon name="ban" size="md" />
                <span>{{ t('admin.subscriptions.revokeSubscription', 'Revoke') }}</span>
              </button>
            </div>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="subscriptionPagination.total > 0"
          :page="subscriptionPagination.page"
          :total="subscriptionPagination.total"
          :page-size="subscriptionPagination.page_size"
          @update:page="handleSubscriptionPageChange"
          @update:pageSize="handleSubscriptionPageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="showAdjustDialog"
      :title="t('admin.subscriptions.adjustSubscription', 'Adjust Subscription')"
      width="normal"
      @close="closeAdjustDialog"
    >
      <form id="adjust-form" class="space-y-4" @submit.prevent="submitAdjust">
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.dailyLimit', 'Daily Limit USD') }}</label>
          <input v-model.number="adjustForm.daily_limit_usd" type="number" min="0" step="0.0001" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.columns.expiresAt', 'Expires At') }}</label>
          <input v-model="adjustForm.expires_at" type="date" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.notes', 'Notes') }}</label>
          <textarea v-model.trim="adjustForm.notes" rows="2" class="input" />
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeAdjustDialog">{{ t('common.cancel', 'Cancel') }}</button>
        <button type="submit" form="adjust-form" class="btn btn-primary" :disabled="actionLoading">
          {{ actionLoading ? t('common.saving', 'Saving...') : t('common.save', 'Save') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showResetQuotaDialog"
      :title="t('admin.subscriptionProducts.resetQuota', 'Reset Quota')"
      width="normal"
      @close="closeResetQuotaDialog"
    >
      <div class="space-y-3">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('admin.subscriptionProducts.resetQuotaDesc', 'Select which usage windows to reset to zero.') }}
        </p>
        <label class="flex items-center gap-2">
          <input v-model="resetQuotaForm.daily" type="checkbox" class="checkbox" />
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.subscriptionProducts.daily', 'Daily') }}</span>
        </label>
        <label class="flex items-center gap-2">
          <input v-model="resetQuotaForm.weekly" type="checkbox" class="checkbox" />
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.subscriptionProducts.weekly', 'Weekly') }}</span>
        </label>
        <label class="flex items-center gap-2">
          <input v-model="resetQuotaForm.monthly" type="checkbox" class="checkbox" />
          <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.subscriptionProducts.monthly', 'Monthly') }}</span>
        </label>
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeResetQuotaDialog">{{ t('common.cancel', 'Cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="actionLoading" @click="submitResetQuota">
          {{ actionLoading ? t('common.processing', 'Processing...') : t('admin.subscriptionProducts.resetQuota', 'Reset Quota') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showRevokeConfirm"
      :title="t('admin.subscriptions.revokeSubscription', 'Revoke Subscription')"
      width="normal"
      @close="closeRevokeConfirm"
    >
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ t('admin.subscriptionProducts.revokeConfirm', 'Are you sure you want to revoke this subscription? This action cannot be undone.') }}
      </p>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeRevokeConfirm">{{ t('common.cancel', 'Cancel') }}</button>
        <button type="button" class="btn bg-red-600 text-white hover:bg-red-700" :disabled="actionLoading" @click="submitRevoke">
          {{ actionLoading ? t('common.processing', 'Processing...') : t('admin.subscriptions.revokeSubscription', 'Revoke') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type {
  AdminProductSubscriptionListItem,
  AdminSubscriptionProduct,
} from '@/types'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateOnly } from '@/utils/format'
import { getPersistedPageSize, setPersistedPageSize } from '@/composables/usePersistedPageSize'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const products = ref<AdminSubscriptionProduct[]>([])
const userSubscriptionsLoading = ref(false)
const subscriptionSearchQuery = ref('')
const subscriptionStatusFilter = ref<string | null>(null)
const subscriptionProductFilter = ref<number | null>(null)
let subscriptionSearchTimeout: ReturnType<typeof setTimeout> | null = null

const subscriptionPagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})

const userSubscriptions = ref<AdminProductSubscriptionListItem[]>([])

const userSubscriptionColumns: Column[] = [
  { key: 'user', label: t('admin.subscriptionProducts.columns.user', 'User') },
  { key: 'product', label: t('admin.subscriptionProducts.columns.product', 'Product') },
  { key: 'usage', label: t('admin.subscriptionProducts.columns.dailyUsage', 'Usage') },
  { key: 'period', label: t('admin.subscriptionProducts.columns.expiresAt', 'Expires') },
  { key: 'status', label: t('admin.subscriptionProducts.columns.status', 'Status') },
  { key: 'actions', label: t('common.actions', 'Actions') },
]

const statusFilterOptions = [
  { value: null, label: t('admin.subscriptionProducts.allStatus', 'All Status') },
  { value: 'active', label: t('admin.subscriptionProducts.status.active', 'Active') },
  { value: 'expired', label: t('admin.subscriptionProducts.status.expired', 'Expired') },
  { value: 'revoked', label: t('admin.subscriptionProducts.status.revoked', 'Revoked') },
]

const productFilterOptions = computed(() => [
  { value: null, label: t('admin.subscriptionProducts.allProducts', 'All Products') },
  ...products.value.map((product) => ({
    value: product.id,
    label: `${product.name} (${product.code})`,
  })),
])

const AVATAR_COLORS = [
  'bg-emerald-500',
  'bg-sky-500',
  'bg-violet-500',
  'bg-amber-500',
  'bg-rose-500',
  'bg-teal-500',
  'bg-indigo-500',
  'bg-pink-500',
]

function avatarInitial(email: string): string {
  return (email?.[0] ?? '?').toUpperCase()
}

function avatarColor(email: string): string {
  let hash = 0
  for (let i = 0; i < (email?.length ?? 0); i++) {
    hash = email.charCodeAt(i) + ((hash << 5) - hash)
  }
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length]
}

function inferLimit(limit: number | null | undefined, dailyLimit: number | null | undefined, multiplier: number): number {
  const l = Number(limit || 0)
  if (l > 0) return l
  const d = Number(dailyLimit || 0)
  return d > 0 ? d * multiplier : 0
}

function usagePercent(usage: number | null | undefined, limit: number | null | undefined): number {
  const u = Number(usage || 0)
  const l = Number(limit || 0)
  if (l <= 0) return u > 0 ? 100 : 0
  return Math.min(100, (u / l) * 100)
}

watch([subscriptionStatusFilter, subscriptionProductFilter], () => {
  subscriptionPagination.page = 1
  void loadUserSubscriptions()
})

async function loadUserSubscriptions() {
  userSubscriptionsLoading.value = true
  try {
    const params: {
      page: number
      page_size: number
      search?: string
      product_id?: number
      status?: string
    } = {
      page: subscriptionPagination.page,
      page_size: subscriptionPagination.page_size,
    }
    const keyword = subscriptionSearchQuery.value.trim()
    if (keyword) params.search = keyword
    if (subscriptionProductFilter.value) params.product_id = subscriptionProductFilter.value
    if (subscriptionStatusFilter.value) params.status = subscriptionStatusFilter.value
    const response = await adminAPI.subscriptionProducts.listUserSubscriptions(params)
    userSubscriptions.value = response.items
    subscriptionPagination.total = response.total
    subscriptionPagination.page = response.page
    subscriptionPagination.page_size = response.page_size
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.subscriptionsLoadError', 'Failed to load subscriptions')))
  } finally {
    userSubscriptionsLoading.value = false
  }
}

async function loadProducts() {
  try {
    products.value = await adminAPI.subscriptionProducts.list()
  } catch {
    // product list is only used for the filter dropdown; swallow errors
  }
}

function debounceSubscriptionSearch() {
  if (subscriptionSearchTimeout) clearTimeout(subscriptionSearchTimeout)
  subscriptionSearchTimeout = setTimeout(() => {
    subscriptionPagination.page = 1
    void loadUserSubscriptions()
  }, 300)
}

function handleSubscriptionPageChange(page: number) {
  subscriptionPagination.page = page
  void loadUserSubscriptions()
}

function handleSubscriptionPageSizeChange(pageSize: number) {
  subscriptionPagination.page_size = pageSize
  subscriptionPagination.page = 1
  setPersistedPageSize(pageSize)
  void loadUserSubscriptions()
}

function formatUsageUSD(value: number | null | undefined): string {
  return `$${Number(value || 0).toFixed(2)}`
}

function formatLimitUSD(value: number | null | undefined): string {
  const amount = Number(value || 0)
  return amount > 0 ? `$${amount.toFixed(2)}` : t('admin.subscriptionProducts.unlimited', 'Unlimited')
}

function remainingDays(expiresAt: string): string {
  if (!expiresAt) return '-'
  const now = new Date()
  const expires = new Date(expiresAt)
  const diffMs = expires.getTime() - now.getTime()
  const days = Math.ceil(diffMs / (1000 * 60 * 60 * 24))
  if (days > 0) return t('admin.subscriptionProducts.daysRemaining', { n: days })
  if (days === 0) return t('admin.subscriptionProducts.expiresToday', 'Expires today')
  return t('admin.subscriptionProducts.daysExpired', { n: Math.abs(days) })
}

function statusLabel(status: string): string {
  return t(`admin.subscriptionProducts.status.${status}`, status)
}

function statusBadgeClass(status: string): string {
  if (status === 'active') return 'status-active'
  if (status === 'expired') return 'status-expired'
  if (status === 'disabled' || status === 'inactive') return 'status-disabled'
  return 'status-revoked'
}

function statusDotClass(status: string): string {
  if (status === 'active') return 'bg-emerald-500'
  if (status === 'expired') return 'bg-amber-500'
  if (status === 'disabled' || status === 'inactive') return 'bg-gray-400'
  return 'bg-red-500'
}

const actionLoading = ref(false)
const selectedSubscription = ref<AdminProductSubscriptionListItem | null>(null)

const showAdjustDialog = ref(false)
const adjustForm = reactive({
  daily_limit_usd: 0,
  expires_at: '',
  notes: '',
})

const showResetQuotaDialog = ref(false)
const resetQuotaForm = reactive({
  daily: true,
  weekly: true,
  monthly: true,
})

const showRevokeConfirm = ref(false)

function openAdjustDialog(row: AdminProductSubscriptionListItem) {
  selectedSubscription.value = row
  adjustForm.daily_limit_usd = row.daily_limit_usd || 0
  adjustForm.expires_at = row.expires_at ? row.expires_at.slice(0, 10) : ''
  adjustForm.notes = row.notes || ''
  showAdjustDialog.value = true
}

function closeAdjustDialog() {
  showAdjustDialog.value = false
  selectedSubscription.value = null
}

async function submitAdjust() {
  if (!selectedSubscription.value) return
  actionLoading.value = true
  try {
    await adminAPI.subscriptionProducts.adjustSubscription(selectedSubscription.value.id, {
      daily_limit_usd: adjustForm.daily_limit_usd,
      expires_at: adjustForm.expires_at || undefined,
      notes: adjustForm.notes || undefined,
    })
    appStore.showSuccess(t('admin.subscriptionProducts.adjusted', 'Subscription adjusted'))
    closeAdjustDialog()
    void loadUserSubscriptions()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.adjustError', 'Failed to adjust subscription')))
  } finally {
    actionLoading.value = false
  }
}

function openResetQuotaDialog(row: AdminProductSubscriptionListItem) {
  selectedSubscription.value = row
  resetQuotaForm.daily = true
  resetQuotaForm.weekly = true
  resetQuotaForm.monthly = true
  showResetQuotaDialog.value = true
}

function closeResetQuotaDialog() {
  showResetQuotaDialog.value = false
  selectedSubscription.value = null
}

async function submitResetQuota() {
  if (!selectedSubscription.value) return
  actionLoading.value = true
  try {
    await adminAPI.subscriptionProducts.resetSubscriptionQuota(selectedSubscription.value.id, {
      daily: resetQuotaForm.daily,
      weekly: resetQuotaForm.weekly,
      monthly: resetQuotaForm.monthly,
    })
    appStore.showSuccess(t('admin.subscriptionProducts.quotaReset', 'Quota reset successfully'))
    closeResetQuotaDialog()
    void loadUserSubscriptions()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.resetQuotaError', 'Failed to reset quota')))
  } finally {
    actionLoading.value = false
  }
}

function openRevokeConfirm(row: AdminProductSubscriptionListItem) {
  selectedSubscription.value = row
  showRevokeConfirm.value = true
}

function closeRevokeConfirm() {
  showRevokeConfirm.value = false
  selectedSubscription.value = null
}

async function submitRevoke() {
  if (!selectedSubscription.value) return
  actionLoading.value = true
  try {
    await adminAPI.subscriptionProducts.revokeSubscription(selectedSubscription.value.id)
    appStore.showSuccess(t('admin.subscriptionProducts.revoked', 'Subscription revoked'))
    closeRevokeConfirm()
    void loadUserSubscriptions()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.revokeError', 'Failed to revoke subscription')))
  } finally {
    actionLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([loadUserSubscriptions(), loadProducts()])
})
</script>

<style scoped>
.status-badge {
  @apply inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium;
}

.status-dot {
  @apply h-1.5 w-1.5 rounded-full;
}

.status-active {
  @apply bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-600/20 dark:bg-emerald-500/10 dark:text-emerald-400 dark:ring-emerald-500/20;
}

.status-expired {
  @apply bg-amber-50 text-amber-700 ring-1 ring-inset ring-amber-600/20 dark:bg-amber-500/10 dark:text-amber-400 dark:ring-amber-500/20;
}

.status-disabled {
  @apply bg-gray-50 text-gray-600 ring-1 ring-inset ring-gray-500/10 dark:bg-gray-500/10 dark:text-gray-400 dark:ring-gray-500/20;
}

.status-revoked {
  @apply bg-red-50 text-red-700 ring-1 ring-inset ring-red-600/10 dark:bg-red-500/10 dark:text-red-400 dark:ring-red-500/20;
}

.usage-row {
  @apply flex items-center gap-2;
}

.usage-label {
  @apply w-10 shrink-0 text-[11px] text-gray-400 dark:text-gray-500;
}

.usage-bar-track {
  @apply h-1.5 w-20 shrink-0 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-600;
}

.usage-bar-fill {
  @apply h-full rounded-full transition-all duration-300;
}

.usage-text {
  @apply shrink-0 text-[11px] tabular-nums text-gray-600 dark:text-gray-300;
}

.usage-tag-inline {
  @apply ml-1 shrink-0 rounded bg-amber-50 px-1.5 py-0.5 text-[10px] tabular-nums text-amber-600 dark:bg-amber-500/10 dark:text-amber-400;
}

.action-btn {
  @apply flex flex-col items-center gap-1 text-gray-400 transition-colors hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300;
}

.action-btn span {
  @apply text-[11px] leading-none;
}
</style>
