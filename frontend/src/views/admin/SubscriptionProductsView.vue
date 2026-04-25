<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_360px]">
        <section class="card overflow-hidden">
          <div class="flex items-center justify-between border-b border-gray-100 p-4 dark:border-dark-700">
            <div>
              <h2 class="font-semibold text-gray-900 dark:text-white">
                {{ t('admin.subscriptionProducts.title') }}
              </h2>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('admin.subscriptionProducts.description') }}
              </p>
            </div>
            <button class="btn btn-secondary" :disabled="loading" @click="loadProducts">
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>

          <div v-if="products.length === 0 && !loading" class="p-8 text-center text-sm text-gray-500">
            {{ t('admin.subscriptionProducts.empty') }}
          </div>

          <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
            <button
              v-for="product in products"
              :key="product.id"
              type="button"
              class="grid w-full gap-3 p-4 text-left transition-colors hover:bg-gray-50 md:grid-cols-[minmax(0,1fr)_120px_160px] dark:hover:bg-dark-800"
              @click="selectProduct(product)"
            >
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <span class="truncate font-medium text-gray-900 dark:text-white">
                    {{ product.name }}
                  </span>
                  <span
                    :class="[
                      'badge',
                      product.status === 'active'
                        ? 'badge-success'
                        : product.status === 'draft'
                          ? 'badge-warning'
                          : 'badge-gray'
                    ]"
                  >
                    {{ product.status }}
                  </span>
                </div>
                <p class="mt-1 truncate text-sm text-gray-500 dark:text-dark-400">
                  {{ product.code }} · {{ product.description }}
                </p>
              </div>
              <div class="text-sm text-gray-600 dark:text-dark-300">
                ${{ product.monthly_limit_usd.toFixed(2) }}
              </div>
              <div class="text-sm text-gray-500 dark:text-dark-400">
                {{ product.default_validity_days }} {{ t('admin.subscriptionProducts.days') }}
              </div>
            </button>
          </div>
        </section>

        <aside class="space-y-6">
          <form class="card space-y-4 p-4" @submit.prevent="createProduct">
            <h3 class="font-semibold text-gray-900 dark:text-white">
              {{ t('admin.subscriptionProducts.create') }}
            </h3>
            <div>
              <label class="input-label">{{ t('admin.subscriptionProducts.code') }}</label>
              <input v-model.trim="productForm.code" class="input" required />
            </div>
            <div>
              <label class="input-label">{{ t('admin.subscriptionProducts.name') }}</label>
              <input v-model.trim="productForm.name" class="input" required />
            </div>
            <div>
              <label class="input-label">{{ t('admin.subscriptionProducts.descriptionLabel') }}</label>
              <textarea v-model.trim="productForm.description" class="input" rows="2" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="input-label">{{ t('admin.subscriptionProducts.status') }}</label>
                <Select v-model="productForm.status" :options="statusOptions" />
              </div>
              <div>
                <label class="input-label">{{ t('admin.subscriptionProducts.validityDays') }}</label>
                <input v-model.number="productForm.default_validity_days" class="input" type="number" min="1" />
              </div>
            </div>
            <div>
              <label class="input-label">{{ t('admin.subscriptionProducts.monthlyLimit') }}</label>
              <input v-model.number="productForm.monthly_limit_usd" class="input" type="number" min="0" step="0.01" />
            </div>
            <button class="btn btn-primary w-full" type="submit" :disabled="savingProduct">
              {{ t('admin.subscriptionProducts.save') }}
            </button>
          </form>

          <section class="card space-y-4 p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.subscriptionProducts.bindings') }}
                </h3>
                <p class="text-xs text-gray-500 dark:text-dark-400">
                  {{ selectedProduct?.name || t('admin.subscriptionProducts.selectProduct') }}
                </p>
              </div>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="!selectedProduct || availableGroupOptions.length === 0"
                @click="addBinding"
              >
                {{ t('admin.subscriptionProducts.addBinding') }}
              </button>
            </div>

            <div v-if="!selectedProduct" class="rounded border border-dashed border-gray-300 p-3 text-sm text-gray-500">
              {{ t('admin.subscriptionProducts.selectProductHint') }}
            </div>

            <div v-else class="space-y-3">
              <div
                v-for="(binding, index) in bindingDrafts"
                :key="index"
                class="space-y-2 rounded border border-gray-200 p-3 dark:border-dark-600"
              >
                <Select v-model="binding.group_id" :options="availableGroupOptions" />
                <div class="grid grid-cols-2 gap-2">
                  <input
                    v-model.number="binding.debit_multiplier"
                    class="input"
                    type="number"
                    min="0"
                    step="0.1"
                  />
                  <Select v-model="binding.status" :options="bindingStatusOptions" />
                </div>
                <button type="button" class="text-xs text-red-500" @click="removeBinding(index)">
                  {{ t('common.remove') }}
                </button>
              </div>
              <button class="btn btn-primary w-full" type="button" :disabled="savingBindings" @click="saveBindings">
                {{ t('admin.subscriptionProducts.saveBindings') }}
              </button>
            </div>
          </section>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import type {
  AdminGroup,
  AdminSubscriptionProduct,
  CreateSubscriptionProductRequest,
  ProductGroupBindingInput
} from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const products = ref<AdminSubscriptionProduct[]>([])
const groups = ref<AdminGroup[]>([])
const loading = ref(false)
const savingProduct = ref(false)
const savingBindings = ref(false)
const selectedProduct = ref<AdminSubscriptionProduct | null>(null)
const bindingDrafts = ref<ProductGroupBindingInput[]>([])

const productForm = reactive<CreateSubscriptionProductRequest>({
  code: '',
  name: '',
  description: '',
  status: 'draft',
  default_validity_days: 30,
  daily_limit_usd: 0,
  weekly_limit_usd: 0,
  monthly_limit_usd: 0,
  sort_order: 0
})

const statusOptions = computed(() => [
  { value: 'draft', label: t('admin.subscriptionProducts.statusDraft') },
  { value: 'active', label: t('admin.subscriptionProducts.statusActive') },
  { value: 'disabled', label: t('admin.subscriptionProducts.statusDisabled') }
])

const bindingStatusOptions = computed(() => [
  { value: 'active', label: t('common.active') },
  { value: 'inactive', label: t('common.inactive') }
])

const availableGroupOptions = computed(() =>
  groups.value
    .filter((group) => group.status === 'active')
    .map((group) => ({
      value: group.id,
      label: group.name
    }))
)

async function loadProducts() {
  loading.value = true
  try {
    products.value = await adminAPI.subscriptionProducts.listProducts()
  } catch (error: any) {
    appStore.showError(error.message || t('admin.subscriptionProducts.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch (error) {
    console.error('Failed to load groups for subscription products:', error)
  }
}

function resetProductForm() {
  Object.assign(productForm, {
    code: '',
    name: '',
    description: '',
    status: 'draft',
    default_validity_days: 30,
    daily_limit_usd: 0,
    weekly_limit_usd: 0,
    monthly_limit_usd: 0,
    sort_order: 0
  })
}

async function createProduct() {
  savingProduct.value = true
  try {
    await adminAPI.subscriptionProducts.createProduct({ ...productForm })
    appStore.showSuccess(t('admin.subscriptionProducts.created'))
    resetProductForm()
    await loadProducts()
  } catch (error: any) {
    appStore.showError(error.message || t('admin.subscriptionProducts.createFailed'))
  } finally {
    savingProduct.value = false
  }
}

function selectProduct(product: AdminSubscriptionProduct) {
  selectedProduct.value = product
  bindingDrafts.value = []
}

function addBinding() {
  const existing = new Set(bindingDrafts.value.map((binding) => binding.group_id))
  const candidate = groups.value.find((group) => group.status === 'active' && !existing.has(group.id))
  if (!candidate) return
  bindingDrafts.value.push({
    group_id: candidate.id,
    debit_multiplier: 1,
    status: 'active',
    sort_order: (bindingDrafts.value.length + 1) * 10
  })
}

function removeBinding(index: number) {
  bindingDrafts.value.splice(index, 1)
}

async function saveBindings() {
  if (!selectedProduct.value) return
  const bindings = bindingDrafts.value
    .filter((binding) => binding.group_id > 0 && binding.debit_multiplier >= 0)
    .map((binding, index) => ({
      group_id: binding.group_id,
      debit_multiplier: binding.debit_multiplier,
      status: binding.status || 'active',
      sort_order: binding.sort_order || (index + 1) * 10
    }))

  savingBindings.value = true
  try {
    await adminAPI.subscriptionProducts.syncBindings(selectedProduct.value.id, bindings)
    appStore.showSuccess(t('admin.subscriptionProducts.bindingsSaved'))
  } catch (error: any) {
    appStore.showError(error.message || t('admin.subscriptionProducts.bindingsSaveFailed'))
  } finally {
    savingBindings.value = false
  }
}

onMounted(() => {
  loadProducts()
  loadGroups()
})
</script>
