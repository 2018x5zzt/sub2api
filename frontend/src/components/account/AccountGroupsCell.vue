<template>
  <div v-if="entries.length > 0" class="relative max-w-56">
    <!-- 分组容器：固定最大宽度，最多显示2行 -->
    <div class="flex flex-wrap gap-1 max-h-14 overflow-hidden">
      <div
        v-for="entry in displayEntries"
        :key="entry.key"
        class="inline-flex items-center gap-1"
      >
        <GroupBadge
          :name="entry.group.name"
          :platform="entry.group.platform"
          :subscription-type="entry.group.subscription_type"
          :rate-multiplier="entry.group.rate_multiplier"
          :show-rate="false"
          class="max-w-24"
        />
        <span
          v-if="showBindingMultiplier && entry.bindingMultiplier !== null"
          class="inline-flex items-center rounded-md bg-blue-50 px-1.5 py-0.5 text-[10px] font-semibold text-blue-700 dark:bg-blue-900/20 dark:text-blue-300"
          :title="t('admin.accounts.groupBindingMultiplierValue', { rate: formatMultiplier(entry.bindingMultiplier) })"
        >
          {{ t('admin.accounts.groupBindingMultiplierShort') }} {{ formatMultiplier(entry.bindingMultiplier) }}x
        </span>
      </div>
      <!-- 更多数量徽章 -->
      <button
        v-if="hiddenCount > 0"
        ref="moreButtonRef"
        @click.stop="showPopover = !showPopover"
        class="inline-flex items-center gap-0.5 rounded-md px-1.5 py-0.5 text-xs font-medium bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-600 dark:text-gray-300 dark:hover:bg-dark-500 transition-colors cursor-pointer whitespace-nowrap"
      >
        <span>+{{ hiddenCount }}</span>
      </button>
    </div>

    <!-- Popover 显示完整列表 -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div
          v-if="showPopover"
          ref="popoverRef"
          class="fixed z-50 min-w-48 max-w-96 rounded-lg border border-gray-200 bg-white p-3 shadow-lg dark:border-dark-600 dark:bg-dark-800"
          :style="popoverStyle"
        >
          <div class="mb-2 flex items-center justify-between">
            <span class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('admin.accounts.groupCountTotal', { count: entries.length }) }}
            </span>
            <button
              @click="showPopover = false"
              class="rounded p-0.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-300"
            >
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="flex flex-wrap gap-1.5 max-h-64 overflow-y-auto">
            <div
              v-for="entry in entries"
              :key="entry.key"
              class="inline-flex items-center gap-1"
            >
              <GroupBadge
                :name="entry.group.name"
                :platform="entry.group.platform"
                :subscription-type="entry.group.subscription_type"
                :rate-multiplier="entry.group.rate_multiplier"
                :show-rate="false"
              />
              <span
                v-if="showBindingMultiplier && entry.bindingMultiplier !== null"
                class="inline-flex items-center rounded-md bg-blue-50 px-1.5 py-0.5 text-[10px] font-semibold text-blue-700 dark:bg-blue-900/20 dark:text-blue-300"
                :title="t('admin.accounts.groupBindingMultiplierValue', { rate: formatMultiplier(entry.bindingMultiplier) })"
              >
                {{ t('admin.accounts.groupBindingMultiplierShort') }} {{ formatMultiplier(entry.bindingMultiplier) }}x
              </span>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 点击外部关闭 popover -->
    <div
      v-if="showPopover"
      class="fixed inset-0 z-40"
      @click="showPopover = false"
    />
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import GroupBadge from '@/components/common/GroupBadge.vue'
import type { AccountGroup, Group } from '@/types'

interface GroupEntry {
  key: string
  group: Group
  bindingMultiplier: number | null
}

interface Props {
  groups: Group[] | null | undefined
  accountGroups?: AccountGroup[] | null | undefined
  maxDisplay?: number
  showBindingMultiplier?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  maxDisplay: 4,
  showBindingMultiplier: false
})

const { t } = useI18n()

const moreButtonRef = ref<HTMLElement | null>(null)
const popoverRef = ref<HTMLElement | null>(null)
const showPopover = ref(false)

const formatMultiplier = (value: number) => value.toFixed(2)

const entries = computed<GroupEntry[]>(() => {
  if (props.accountGroups && props.accountGroups.length > 0) {
    const mapped: GroupEntry[] = []
    for (const [index, accountGroup] of props.accountGroups.entries()) {
      const group = accountGroup.group ?? props.groups?.find(item => item.id === accountGroup.group_id)
      if (!group) continue
      mapped.push({
        key: `ag-${accountGroup.account_id}-${accountGroup.group_id}-${index}`,
        group,
        bindingMultiplier: accountGroup.billing_multiplier ?? 1
      })
    }
    return mapped
  }

  return (props.groups ?? []).map((group, index) => ({
    key: `g-${group.id}-${index}`,
    group,
    bindingMultiplier: null
  }))
})

// 显示的分组（最多显示 maxDisplay 个）
const displayEntries = computed(() => {
  if (entries.value.length <= props.maxDisplay) {
    return entries.value
  }
  // 留一个位置给 +N 按钮
  return entries.value.slice(0, props.maxDisplay - 1)
})

// 隐藏的数量
const hiddenCount = computed(() => {
  if (entries.value.length <= props.maxDisplay) return 0
  return entries.value.length - (props.maxDisplay - 1)
})

// Popover 位置样式
const popoverStyle = computed(() => {
  if (!moreButtonRef.value) return {}
  const rect = moreButtonRef.value.getBoundingClientRect()
  const viewportHeight = window.innerHeight
  const viewportWidth = window.innerWidth

  let top = rect.bottom + 8
  let left = rect.left

  // 如果下方空间不足，显示在上方
  if (top + 280 > viewportHeight) {
    top = Math.max(8, rect.top - 280)
  }

  // 如果右侧空间不足，向左偏移
  if (left + 384 > viewportWidth) {
    left = Math.max(8, viewportWidth - 392)
  }

  return {
    top: `${top}px`,
    left: `${left}px`
  }
})

// 关闭 popover 的键盘事件
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    showPopover.value = false
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>
