<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-start">
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="flex rounded-md border border-gray-200 bg-gray-50 p-1 dark:border-dark-700 dark:bg-dark-800">
              <button
                type="button"
                class="tab-button"
                :class="activeTab === 'subscriptions' ? 'tab-button-active' : 'tab-button-inactive'"
                @click="activeTab = 'subscriptions'"
              >
                {{ t('admin.subscriptionProducts.userSubscriptionsTab', 'User Subscriptions') }}
              </button>
              <button
                type="button"
                class="tab-button"
                :class="activeTab === 'products' ? 'tab-button-active' : 'tab-button-inactive'"
                @click="activeTab = 'products'"
              >
                {{ t('admin.subscriptionProducts.productConfigTab', 'Product Config') }}
              </button>
            </div>

            <div class="relative w-full sm:w-72">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-if="activeTab === 'subscriptions'"
                v-model="subscriptionSearchQuery"
                type="text"
                :placeholder="t('admin.subscriptionProducts.searchSubscriptionsPlaceholder', 'Search users or products')"
                class="input pl-10"
                @input="debounceSubscriptionSearch"
              />
              <input
                v-else
                v-model="productSearchQuery"
                type="text"
                :placeholder="t('admin.subscriptionProducts.searchPlaceholder', 'Search products')"
                class="input pl-10"
              />
            </div>

            <Select
              v-if="activeTab === 'subscriptions'"
              v-model="subscriptionStatusFilter"
              :options="statusFilterOptions"
              :placeholder="t('admin.subscriptionProducts.allStatus', 'All Status')"
              class="w-40"
            />
            <Select
              v-else
              v-model="productStatusFilter"
              :options="productStatusFilterOptions"
              :placeholder="t('admin.subscriptionProducts.allStatus', 'All Status')"
              class="w-40"
            />
            <Select
              v-if="activeTab === 'subscriptions'"
              v-model="subscriptionProductFilter"
              :options="productFilterOptions"
              :placeholder="t('admin.subscriptionProducts.allProducts', 'All Products')"
              class="w-56"
            />
          </div>

          <div class="flex w-full flex-shrink-0 flex-wrap items-center justify-end gap-3 lg:w-auto">
            <button
              @click="refreshActiveTab"
              :disabled="activeLoading"
              class="btn btn-secondary"
              :title="t('common.refresh', 'Refresh')"
            >
              <Icon name="refresh" size="md" :class="activeLoading ? 'animate-spin' : ''" />
            </button>
            <button v-if="activeTab === 'products'" @click="openProductDialog(null)" class="btn btn-primary">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.subscriptionProducts.createProduct', 'Create Product') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          v-if="activeTab === 'subscriptions'"
          :columns="userSubscriptionColumns"
          :data="userSubscriptions"
          :loading="userSubscriptionsLoading"
        >
          <template #cell-user="{ row }">
            <div class="min-w-[220px]">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.user_email }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ row.user_username || '-' }} #{{ row.user_id }}
              </div>
            </div>
          </template>

          <template #cell-product="{ row }">
            <div class="min-w-[220px]">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.product_name }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ row.product_code }}</div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span class="status-badge" :class="statusBadgeClass(value)">
              {{ statusLabel(value) }}
            </span>
          </template>

          <template #cell-daily_usage="{ row }">
            <div class="space-y-1 text-xs text-gray-600 dark:text-gray-300">
              <div>{{ formatUsageUSD(row.daily_usage_usd) }} / {{ formatLimitUSD(row.daily_limit_usd) }}</div>
              <div>{{ t('admin.subscriptionProducts.weekly', 'Weekly') }}: {{ formatUsageUSD(row.weekly_usage_usd) }}</div>
              <div>{{ t('admin.subscriptionProducts.monthly', 'Monthly') }}: {{ formatUsageUSD(row.monthly_usage_usd) }}</div>
            </div>
          </template>

          <template #cell-carryover="{ row }">
            <div class="space-y-1 text-xs text-gray-600 dark:text-gray-300">
              <div>{{ t('admin.subscriptionProducts.carryoverIn', 'In') }}: {{ formatUsageUSD(row.daily_carryover_in_usd) }}</div>
              <div>{{ t('admin.subscriptionProducts.carryoverUsed', 'Used') }}: {{ formatUsageUSD(row.carryover_used_usd) }}</div>
              <div>{{ t('admin.subscriptionProducts.carryoverRemaining', 'Remaining') }}: {{ formatUsageUSD(row.daily_carryover_remaining_usd) }}</div>
            </div>
          </template>

          <template #cell-fresh_daily_usage="{ value }">
            {{ formatUsageUSD(value) }}
          </template>

          <template #cell-period="{ row }">
            <div class="space-y-1 text-xs text-gray-600 dark:text-gray-300">
              <div>{{ formatDateOnly(row.starts_at) }}</div>
              <div>{{ formatDateOnly(row.expires_at) }}</div>
            </div>
          </template>

          <template #cell-notes="{ value }">
            <span class="block max-w-[220px] truncate text-gray-600 dark:text-gray-400">
              {{ value || '-' }}
            </span>
          </template>
        </DataTable>

        <DataTable v-else :columns="productColumns" :data="pagedProducts" :loading="loading">
          <template #cell-name="{ row }">
            <div class="min-w-[220px]">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ row.code }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.subscriptionProducts.family', 'Family') }}: {{ row.product_family || 'default' }}
              </div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span class="status-badge" :class="statusBadgeClass(value)">
              {{ statusLabel(value) }}
            </span>
          </template>

          <template #cell-limits="{ row }">
            <div class="min-w-[210px] space-y-1 text-xs text-gray-600 dark:text-gray-300">
              <div>{{ t('admin.subscriptionProducts.daily', 'Daily') }}: {{ formatLimitUSD(row.daily_limit_usd) }}</div>
              <div>{{ t('admin.subscriptionProducts.weekly', 'Weekly') }}: {{ formatLimitUSD(row.weekly_limit_usd) }}</div>
              <div>{{ t('admin.subscriptionProducts.monthly', 'Monthly') }}: {{ formatLimitUSD(row.monthly_limit_usd) }}</div>
            </div>
          </template>

          <template #cell-description="{ value }">
            <span class="block max-w-[280px] truncate text-gray-600 dark:text-gray-400">
              {{ value || '-' }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center justify-end gap-2">
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :title="t('admin.subscriptionProducts.editProduct', 'Edit Product')"
                @click="openProductDialog(row)"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :title="t('admin.subscriptionProducts.bindGroups', 'Bind Groups')"
                @click="openBindingsDialog(row)"
              >
                <Icon name="link" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :title="t('admin.subscriptionProducts.assignUser', 'Assign User')"
                @click="openAssignDialog(row)"
              >
                <Icon name="userPlus" size="sm" />
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :title="t('admin.subscriptionProducts.viewSubscriptions', 'View Subscriptions')"
                @click="openSubscriptionsDialog(row)"
              >
                <Icon name="users" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.subscriptionProducts.emptyTitle', 'No Product Subscriptions')"
              :description="t('admin.subscriptionProducts.emptyDescription', 'Create a product to share quota across multiple groups.')"
              :action-text="t('admin.subscriptionProducts.createProduct', 'Create Product')"
              @action="openProductDialog(null)"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="activeTab === 'subscriptions' && subscriptionPagination.total > 0"
          :page="subscriptionPagination.page"
          :total="subscriptionPagination.total"
          :page-size="subscriptionPagination.page_size"
          @update:page="handleSubscriptionPageChange"
          @update:pageSize="handleSubscriptionPageSizeChange"
        />
        <Pagination
          v-else-if="activeTab === 'products' && filteredProducts.length > 0"
          :page="pagination.page"
          :total="filteredProducts.length"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="showProductDialog"
      :title="editingProduct ? t('admin.subscriptionProducts.editProduct', 'Edit Product') : t('admin.subscriptionProducts.createProduct', 'Create Product')"
      width="wide"
      @close="closeProductDialog"
    >
      <form id="product-form" class="grid gap-4 md:grid-cols-2" @submit.prevent="submitProduct">
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.code', 'Code') }}</label>
          <input v-model.trim="productForm.code" type="text" required maxlength="64" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.name', 'Name') }}</label>
          <input v-model.trim="productForm.name" type="text" required maxlength="255" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.status', 'Status') }}</label>
          <Select v-model="productForm.status" :options="statusEditOptions" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.productFamily', 'Product Family') }}</label>
          <input v-model.trim="productForm.product_family" type="text" maxlength="64" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.validityDays', 'Default Validity Days') }}</label>
          <input v-model.number="productForm.default_validity_days" type="number" min="1" max="36500" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.dailyLimit', 'Daily Limit USD') }}</label>
          <input v-model.number="productForm.daily_limit_usd" type="number" min="0" step="0.0001" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.weeklyLimit', 'Weekly Limit USD') }}</label>
          <input v-model.number="productForm.weekly_limit_usd" type="number" min="0" step="0.0001" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.monthlyLimit', 'Monthly Limit USD') }}</label>
          <input v-model.number="productForm.monthly_limit_usd" type="number" min="0" step="0.0001" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.sortOrder', 'Sort Order') }}</label>
          <input v-model.number="productForm.sort_order" type="number" class="input" />
        </div>
        <div class="md:col-span-2">
          <label class="input-label">{{ t('admin.subscriptionProducts.form.description', 'Description') }}</label>
          <textarea v-model.trim="productForm.description" rows="3" class="input" />
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeProductDialog">{{ t('common.cancel', 'Cancel') }}</button>
        <button type="submit" form="product-form" class="btn btn-primary" :disabled="submitting">
          {{ submitting ? t('common.saving', 'Saving...') : t('common.save', 'Save') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showBindingsDialog"
      :title="formatProductDialogTitle(t('admin.subscriptionProducts.bindGroups', 'Bind Groups'))"
      width="extra-wide"
      @close="closeBindingsDialog"
    >
      <div class="space-y-4">
        <div class="flex justify-end">
          <button type="button" class="btn btn-secondary btn-sm" @click="addBindingRow">
            <Icon name="plus" size="sm" class="mr-1.5" />
            {{ t('admin.subscriptionProducts.addBinding', 'Add Binding') }}
          </button>
        </div>
        <div class="space-y-3">
          <div
            v-for="(binding, index) in bindingForm"
            :key="binding.local_id"
            class="grid gap-3 rounded-lg border border-gray-200 p-3 dark:border-dark-700 md:grid-cols-[1fr_140px_140px_120px_auto]"
          >
            <Select
              v-model="binding.group_id"
              :options="groupOptions"
              searchable
              :placeholder="t('admin.subscriptionProducts.selectGroup', 'Select group')"
            />
            <input v-model.number="binding.debit_multiplier" type="number" min="0.0001" step="0.0001" class="input" />
            <Select v-model="binding.status" :options="bindingStatusOptions" />
            <input v-model.number="binding.sort_order" type="number" class="input" />
            <button type="button" class="btn btn-secondary btn-sm" @click="removeBindingRow(index)">
              <Icon name="trash" size="sm" />
            </button>
          </div>
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeBindingsDialog">{{ t('common.cancel', 'Cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="submitting" @click="submitBindings">
          {{ submitting ? t('common.saving', 'Saving...') : t('common.save', 'Save') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showAssignDialog"
      :title="t('admin.subscriptionProducts.assignUser', 'Assign User')"
      width="normal"
      @close="closeAssignDialog"
    >
      <form id="assign-product-form" class="space-y-4" @submit.prevent="submitAssign">
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.user', 'User') }}</label>
          <div class="relative">
            <input
              v-model="userKeyword"
              type="text"
              class="input"
              :placeholder="t('admin.users.searchUsers', 'Search users')"
              @input="debounceSearchUsers"
              @focus="showUserDropdown = true"
            />
            <div
              v-if="showUserDropdown && (userResults.length > 0 || userKeyword)"
              class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-lg border border-gray-200 bg-white shadow-lg dark:border-gray-700 dark:bg-gray-800"
            >
              <div v-if="userSearchLoading" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                {{ t('common.loading', 'Loading...') }}
              </div>
              <button
                v-for="user in userResults"
                :key="user.id"
                type="button"
                class="w-full px-4 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
                @click="selectUser(user)"
              >
                <span class="font-medium text-gray-900 dark:text-white">{{ user.email }}</span>
                <span class="ml-2 text-gray-500 dark:text-gray-400">#{{ user.id }}</span>
              </button>
              <div v-if="!userSearchLoading && userResults.length === 0 && userKeyword" class="px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                {{ t('common.noOptionsFound', 'No options found') }}
              </div>
            </div>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.validityDays', 'Validity Days') }}</label>
          <input v-model.number="assignForm.validity_days" type="number" min="1" max="36500" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.subscriptionProducts.form.notes', 'Notes') }}</label>
          <textarea v-model.trim="assignForm.notes" rows="3" class="input" />
        </div>
      </form>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="closeAssignDialog">{{ t('common.cancel', 'Cancel') }}</button>
        <button type="submit" form="assign-product-form" class="btn btn-primary" :disabled="submitting">
          {{ submitting ? t('common.processing', 'Processing...') : t('admin.subscriptionProducts.assignUser', 'Assign User') }}
        </button>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="showSubscriptionsDialog"
      :title="formatProductDialogTitle(t('admin.subscriptionProducts.viewSubscriptions', 'View Subscriptions'))"
      width="extra-wide"
      @close="closeSubscriptionsDialog"
    >
      <DataTable :columns="productSubscriptionColumns" :data="productSubscriptions" :loading="subscriptionsLoading">
        <template #cell-user_id="{ value }">#{{ value }}</template>
        <template #cell-expires_at="{ value }">{{ formatDateOnly(value) }}</template>
        <template #cell-usage="{ row }">
          <div class="space-y-1 text-xs text-gray-600 dark:text-gray-300">
            <div>{{ t('admin.subscriptionProducts.daily', 'Daily') }}: {{ formatUsageUSD(row.daily_usage_usd) }}</div>
            <div>{{ t('admin.subscriptionProducts.monthly', 'Monthly') }}: {{ formatUsageUSD(row.monthly_usage_usd) }}</div>
          </div>
        </template>
      </DataTable>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type {
  AdminProductSubscriptionListItem,
  AdminGroup,
  AdminSubscriptionProduct,
  AdminSubscriptionProductBinding,
  AdminUserProductSubscription
} from '@/types'
import type { SimpleUser } from '@/api/admin/usage'
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
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

type ProductForm = {
  code: string
  name: string
  description: string
  status: string
  product_family: string
  default_validity_days: number
  daily_limit_usd: number
  weekly_limit_usd: number
  monthly_limit_usd: number
  sort_order: number
}

type BindingFormRow = {
  local_id: string
  group_id: number | null
  debit_multiplier: number
  status: string
  sort_order: number
}

const products = ref<AdminSubscriptionProduct[]>([])
const groups = ref<AdminGroup[]>([])
const loading = ref(false)
const userSubscriptionsLoading = ref(false)
const submitting = ref(false)
const activeTab = ref<'subscriptions' | 'products'>('subscriptions')
const productSearchQuery = ref('')
const productStatusFilter = ref<string | null>(null)
const subscriptionSearchQuery = ref('')
const subscriptionStatusFilter = ref<string | null>(null)
const subscriptionProductFilter = ref<number | null>(null)
const editingProduct = ref<AdminSubscriptionProduct | null>(null)
const selectedProduct = ref<AdminSubscriptionProduct | null>(null)
const showProductDialog = ref(false)
const showBindingsDialog = ref(false)
const showAssignDialog = ref(false)
const showSubscriptionsDialog = ref(false)
const bindingForm = ref<BindingFormRow[]>([])
const currentBindings = ref<AdminSubscriptionProductBinding[]>([])
const productSubscriptions = ref<AdminUserProductSubscription[]>([])
const subscriptionsLoading = ref(false)
const userKeyword = ref('')
const selectedUser = ref<SimpleUser | null>(null)
const userResults = ref<SimpleUser[]>([])
const showUserDropdown = ref(false)
const userSearchLoading = ref(false)
let userSearchTimeout: ReturnType<typeof setTimeout> | null = null
let subscriptionSearchTimeout: ReturnType<typeof setTimeout> | null = null

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
})

const subscriptionPagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
})

const userSubscriptions = ref<AdminProductSubscriptionListItem[]>([])

const activeLoading = computed(() =>
  activeTab.value === 'subscriptions' ? userSubscriptionsLoading.value : loading.value
)

const productForm = reactive<ProductForm>({
  code: '',
  name: '',
  description: '',
  status: 'active',
  product_family: 'default',
  default_validity_days: 30,
  daily_limit_usd: 0,
  weekly_limit_usd: 0,
  monthly_limit_usd: 0,
  sort_order: 0,
})

const assignForm = reactive({
  validity_days: 30,
  notes: '',
})

const userSubscriptionColumns: Column[] = [
  { key: 'user', label: t('admin.subscriptionProducts.columns.user', 'User') },
  { key: 'product', label: t('admin.subscriptionProducts.columns.product', 'Product') },
  { key: 'status', label: t('admin.subscriptionProducts.columns.status', 'Status') },
  { key: 'daily_usage', label: t('admin.subscriptionProducts.columns.dailyUsage', 'Daily Usage') },
  { key: 'carryover', label: t('admin.subscriptionProducts.columns.carryover', 'Carryover') },
  { key: 'fresh_daily_usage', label: t('admin.subscriptionProducts.columns.freshDailyUsage', 'Fresh Today') },
  { key: 'period', label: t('admin.subscriptionProducts.columns.period', 'Period') },
  { key: 'notes', label: t('admin.subscriptionProducts.columns.notes', 'Notes') },
]

const productColumns: Column[] = [
  { key: 'name', label: t('admin.subscriptionProducts.columns.product', 'Product') },
  { key: 'status', label: t('admin.subscriptionProducts.columns.status', 'Status') },
  { key: 'limits', label: t('admin.subscriptionProducts.columns.limits', 'Limits') },
  { key: 'default_validity_days', label: t('admin.subscriptionProducts.columns.defaultValidity', 'Default Validity Days') },
  { key: 'description', label: t('admin.subscriptionProducts.columns.description', 'Description') },
  { key: 'actions', label: t('common.actions', 'Actions'), class: 'text-right' },
]

const productSubscriptionColumns: Column[] = [
  { key: 'id', label: 'ID' },
  { key: 'user_id', label: t('admin.subscriptionProducts.columns.user', 'User') },
  { key: 'status', label: t('admin.subscriptionProducts.columns.status', 'Status') },
  { key: 'expires_at', label: t('admin.subscriptionProducts.columns.expiresAt', 'Expires At') },
  { key: 'usage', label: t('admin.subscriptionProducts.columns.usage', 'Usage') },
  { key: 'notes', label: t('admin.subscriptionProducts.columns.notes', 'Notes') },
]

const statusFilterOptions = [
  { value: null, label: t('admin.subscriptionProducts.allStatus', 'All Status') },
  { value: 'active', label: t('admin.subscriptionProducts.status.active', 'Active') },
  { value: 'expired', label: t('admin.subscriptionProducts.status.expired', 'Expired') },
  { value: 'revoked', label: t('admin.subscriptionProducts.status.revoked', 'Revoked') },
]

const productStatusFilterOptions = [
  { value: null, label: t('admin.subscriptionProducts.allStatus', 'All Status') },
  { value: 'active', label: t('admin.subscriptionProducts.status.active', 'Active') },
  { value: 'draft', label: t('admin.subscriptionProducts.status.draft', 'Draft') },
  { value: 'disabled', label: t('admin.subscriptionProducts.status.disabled', 'Disabled') },
]

const statusEditOptions = productStatusFilterOptions.filter((item) => item.value !== null)

const bindingStatusOptions = [
  { value: 'active', label: t('admin.subscriptionProducts.status.active', 'Active') },
  { value: 'inactive', label: t('admin.subscriptionProducts.status.inactive', 'Inactive') },
]

const groupOptions = computed(() =>
  groups.value.map((group) => ({
    value: group.id,
    label: `${group.name} #${group.id}`,
  }))
)

const productFilterOptions = computed(() => [
  { value: null, label: t('admin.subscriptionProducts.allProducts', 'All Products') },
  ...products.value.map((product) => ({
    value: product.id,
    label: `${product.name} (${product.code})`,
  })),
])

const filteredProducts = computed(() => {
  const keyword = productSearchQuery.value.trim().toLowerCase()
  return products.value.filter((product) => {
    if (productStatusFilter.value && product.status !== productStatusFilter.value) return false
    if (!keyword) return true
    return [product.name, product.code, product.description, product.product_family].some((value) =>
      String(value || '').toLowerCase().includes(keyword)
    )
  })
})

const pagedProducts = computed(() => {
  const start = (pagination.page - 1) * pagination.page_size
  return filteredProducts.value.slice(start, start + pagination.page_size)
})

watch([productSearchQuery, productStatusFilter], () => {
  pagination.page = 1
})

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
  loading.value = true
  try {
    products.value = await adminAPI.subscriptionProducts.list()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.loadError', 'Failed to load products')))
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.groups.loadError', 'Failed to load groups')))
  }
}

function openProductDialog(product: AdminSubscriptionProduct | null) {
  editingProduct.value = product
  Object.assign(productForm, product
    ? {
        code: product.code,
        name: product.name,
        description: product.description || '',
        status: product.status || 'active',
        product_family: product.product_family || 'default',
        default_validity_days: product.default_validity_days || 30,
        daily_limit_usd: product.daily_limit_usd || 0,
        weekly_limit_usd: product.weekly_limit_usd || 0,
        monthly_limit_usd: product.monthly_limit_usd || 0,
        sort_order: product.sort_order || 0,
      }
    : {
        code: '',
        name: '',
        description: '',
        status: 'active',
        product_family: 'default',
        default_validity_days: 30,
        daily_limit_usd: 0,
        weekly_limit_usd: 0,
        monthly_limit_usd: 0,
        sort_order: 0,
      })
  showProductDialog.value = true
}

function closeProductDialog() {
  showProductDialog.value = false
  editingProduct.value = null
}

async function submitProduct() {
  submitting.value = true
  try {
    if (editingProduct.value) {
      await adminAPI.subscriptionProducts.update(editingProduct.value.id, { ...productForm })
      appStore.showSuccess(t('admin.subscriptionProducts.updated', 'Product updated'))
    } else {
      await adminAPI.subscriptionProducts.create({ ...productForm })
      appStore.showSuccess(t('admin.subscriptionProducts.created', 'Product created'))
    }
    closeProductDialog()
    await loadProducts()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.saveError', 'Failed to save product')))
  } finally {
    submitting.value = false
  }
}

async function openBindingsDialog(product: AdminSubscriptionProduct) {
  selectedProduct.value = product
  currentBindings.value = []
  bindingForm.value = []
  showBindingsDialog.value = true
  try {
    currentBindings.value = await adminAPI.subscriptionProducts.listBindings(product.id)
    bindingForm.value = currentBindings.value.map((binding) => ({
      local_id: `${binding.group_id}-${binding.sort_order}`,
      group_id: binding.group_id,
      debit_multiplier: binding.debit_multiplier || 1,
      status: binding.status || 'active',
      sort_order: binding.sort_order || 0,
    }))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.bindingsLoadError', 'Failed to load bindings')))
  }
}

function closeBindingsDialog() {
  showBindingsDialog.value = false
  selectedProduct.value = null
  bindingForm.value = []
}

function addBindingRow() {
  bindingForm.value.push({
    local_id: `${Date.now()}-${Math.random()}`,
    group_id: null,
    debit_multiplier: 1,
    status: 'active',
    sort_order: bindingForm.value.length + 1,
  })
}

function removeBindingRow(index: number) {
  bindingForm.value.splice(index, 1)
}

async function submitBindings() {
  if (!selectedProduct.value) return
  const seen = new Set<number>()
  const bindings = bindingForm.value
    .filter((binding) => binding.group_id)
    .map((binding) => {
      const groupID = Number(binding.group_id)
      if (seen.has(groupID)) {
        throw new Error(t('admin.subscriptionProducts.duplicateGroup', 'Duplicate group binding'))
      }
      seen.add(groupID)
      return {
        group_id: groupID,
        debit_multiplier: Number(binding.debit_multiplier) || 1,
        status: binding.status || 'active',
        sort_order: Number(binding.sort_order) || 0,
      }
    })
  submitting.value = true
  try {
    await adminAPI.subscriptionProducts.syncBindings(selectedProduct.value.id, bindings)
    appStore.showSuccess(t('admin.subscriptionProducts.bindingsSaved', 'Bindings saved'))
    closeBindingsDialog()
  } catch (error) {
    appStore.showError(error instanceof Error ? error.message : extractApiErrorMessage(error, t('admin.subscriptionProducts.bindingsSaveError', 'Failed to save bindings')))
  } finally {
    submitting.value = false
  }
}

function openAssignDialog(product: AdminSubscriptionProduct) {
  selectedProduct.value = product
  selectedUser.value = null
  userKeyword.value = ''
  userResults.value = []
  assignForm.validity_days = product.default_validity_days || 30
  assignForm.notes = ''
  showAssignDialog.value = true
}

function closeAssignDialog() {
  showAssignDialog.value = false
  selectedProduct.value = null
  showUserDropdown.value = false
}

function debounceSearchUsers() {
  if (userSearchTimeout) clearTimeout(userSearchTimeout)
  userSearchTimeout = setTimeout(searchUsers, 300)
}

function debounceSubscriptionSearch() {
  if (subscriptionSearchTimeout) clearTimeout(subscriptionSearchTimeout)
  subscriptionSearchTimeout = setTimeout(() => {
    subscriptionPagination.page = 1
    void loadUserSubscriptions()
  }, 300)
}

async function searchUsers() {
  const keyword = userKeyword.value.trim()
  if (!keyword) {
    userResults.value = []
    return
  }
  userSearchLoading.value = true
  try {
    userResults.value = await adminAPI.usage.searchUsers(keyword)
  } catch {
    userResults.value = []
  } finally {
    userSearchLoading.value = false
  }
}

function selectUser(user: SimpleUser) {
  selectedUser.value = user
  userKeyword.value = `${user.email} (#${user.id})`
  showUserDropdown.value = false
}

async function submitAssign() {
  if (!selectedProduct.value || !selectedUser.value) {
    appStore.showError(t('admin.subscriptionProducts.selectUserRequired', 'Please select a user'))
    return
  }
  submitting.value = true
  try {
    await adminAPI.subscriptionProducts.assign(selectedProduct.value.id, {
      user_id: selectedUser.value.id,
      validity_days: assignForm.validity_days || selectedProduct.value.default_validity_days || 30,
      notes: assignForm.notes,
    })
    appStore.showSuccess(t('admin.subscriptionProducts.assigned', 'Product subscription assigned'))
    closeAssignDialog()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.assignError', 'Failed to assign product subscription')))
  } finally {
    submitting.value = false
  }
}

async function openSubscriptionsDialog(product: AdminSubscriptionProduct) {
  selectedProduct.value = product
  productSubscriptions.value = []
  showSubscriptionsDialog.value = true
  subscriptionsLoading.value = true
  try {
    productSubscriptions.value = await adminAPI.subscriptionProducts.listSubscriptions(product.id)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.subscriptionProducts.subscriptionsLoadError', 'Failed to load subscriptions')))
  } finally {
    subscriptionsLoading.value = false
  }
}

function closeSubscriptionsDialog() {
  showSubscriptionsDialog.value = false
  selectedProduct.value = null
  productSubscriptions.value = []
}

function handlePageChange(page: number) {
  pagination.page = page
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  setPersistedPageSize(pageSize)
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

function refreshActiveTab() {
  if (activeTab.value === 'subscriptions') {
    void loadUserSubscriptions()
    return
  }
  void loadProducts()
}

function formatUsageUSD(value: number | null | undefined): string {
  return `$${Number(value || 0).toFixed(2)}`
}

function formatLimitUSD(value: number | null | undefined): string {
  const amount = Number(value || 0)
  return amount > 0 ? `$${amount.toFixed(2)}` : t('admin.subscriptionProducts.unlimited', 'Unlimited')
}

function statusLabel(status: string): string {
  return t(`admin.subscriptionProducts.status.${status}`, status)
}

function statusBadgeClass(status: string): string {
  if (status === 'active') return 'status-active'
  if (status === 'disabled' || status === 'inactive') return 'status-disabled'
  return 'status-draft'
}

function formatProductDialogTitle(prefix: string): string {
  return selectedProduct.value?.name ? `${prefix}: ${selectedProduct.value.name}` : prefix
}

onMounted(async () => {
  await Promise.all([loadUserSubscriptions(), loadProducts(), loadGroups()])
})
</script>

<style scoped>
.status-badge {
  @apply inline-flex rounded-full px-2 py-0.5 text-xs font-medium;
}

.status-active {
  @apply bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300;
}

.status-disabled {
  @apply bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300;
}

.status-draft {
  @apply bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300;
}
</style>
