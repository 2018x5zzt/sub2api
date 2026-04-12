<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="space-y-4">
          <div class="flex justify-end">
            <button type="button" class="btn btn-secondary" :disabled="isRefreshing" @click="refreshAll">
              {{ t('common.refresh') }}
            </button>
          </div>

          <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
            <div v-for="card in statCards" :key="card.key" class="card p-5">
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ card.label }}</p>
              <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">{{ card.value }}</p>
            </div>
          </div>
        </div>
      </template>

      <template #table>
        <div class="space-y-6">
          <section class="rounded-2xl border border-amber-200 bg-amber-50 p-5 dark:border-amber-900/60 dark:bg-amber-950/20">
            <h2 class="text-base font-semibold text-amber-800 dark:text-amber-300">
              {{ t('admin.invites.riskPanelTitle') }}
            </h2>
            <p class="mt-2 text-sm leading-6 text-amber-700 dark:text-amber-200">
              {{ t('admin.invites.riskPanelBody') }}
            </p>
          </section>

          <section class="grid gap-6 xl:grid-cols-3">
            <article class="card p-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.invites.rebind.title') }}
              </h3>
              <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
                {{ t('admin.invites.rebind.description') }}
              </p>

              <div class="mt-4 grid gap-3">
                <input
                  v-model.trim="rebindForm.invitee_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.rebind.inviteeUserId')"
                />
                <input
                  v-model.trim="rebindForm.new_inviter_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.rebind.newInviterUserId')"
                />
                <textarea
                  v-model.trim="rebindForm.reason"
                  class="input min-h-24"
                  :placeholder="t('admin.invites.rebind.reason')"
                />
                <input
                  v-model.trim="rebindForm.confirm_text"
                  type="text"
                  class="input"
                  placeholder="REBIND"
                />
                <p class="text-xs text-gray-500 dark:text-dark-400">
                  {{ t('admin.invites.rebind.confirmHelp') }}
                </p>
                <button
                  type="button"
                  class="btn btn-primary"
                  :disabled="!canSubmitRebind || loading.rebind"
                  @click="submitRebind"
                >
                  {{ t('admin.invites.rebind.submit') }}
                </button>
              </div>
            </article>

            <article class="card p-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.invites.manualGrant.title') }}
              </h3>
              <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
                {{ t('admin.invites.manualGrant.description') }}
              </p>

              <div class="mt-4 grid gap-3 md:grid-cols-2">
                <input
                  v-model.trim="manualGrantForm.target_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.manualGrant.targetUserId')"
                />
                <input
                  v-model.trim="manualGrantForm.reward_target_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.manualGrant.rewardTargetUserId')"
                />
                <input
                  v-model.trim="manualGrantForm.inviter_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.manualGrant.inviterUserId')"
                />
                <input
                  v-model.trim="manualGrantForm.invitee_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.manualGrant.inviteeUserId')"
                />
                <select v-model="manualGrantForm.reward_role" class="input">
                  <option value="inviter">{{ t('admin.invites.roles.inviter') }}</option>
                  <option value="invitee">{{ t('admin.invites.roles.invitee') }}</option>
                </select>
                <input
                  v-model.trim="manualGrantForm.reward_amount"
                  type="number"
                  min="0.01"
                  step="0.01"
                  class="input"
                  :placeholder="t('admin.invites.manualGrant.rewardAmount')"
                />
                <input
                  v-model.trim="manualGrantForm.notes"
                  type="text"
                  class="input md:col-span-2"
                  :placeholder="t('admin.invites.manualGrant.notes')"
                />
                <textarea
                  v-model.trim="manualGrantForm.reason"
                  class="input min-h-24 md:col-span-2"
                  :placeholder="t('admin.invites.manualGrant.reason')"
                />
                <input
                  v-model.trim="manualGrantForm.confirm_text"
                  type="text"
                  class="input md:col-span-2"
                  placeholder="GRANT"
                />
              </div>

              <p class="mt-3 text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.invites.manualGrant.confirmHelp') }}
              </p>

              <button
                type="button"
                class="btn btn-primary mt-4"
                :disabled="!canSubmitManualGrant || loading.manualGrant"
                @click="submitManualGrant"
              >
                {{ t('admin.invites.manualGrant.submit') }}
              </button>
            </article>

            <article class="card p-6">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.invites.recompute.title') }}
              </h3>
              <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
                {{ t('admin.invites.recompute.description') }}
              </p>

              <div class="mt-4 grid gap-3 md:grid-cols-2">
                <input
                  v-model.trim="recomputeForm.invitee_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.recompute.inviteeUserId')"
                />
                <input
                  v-model.trim="recomputeForm.inviter_user_id"
                  type="number"
                  min="1"
                  class="input"
                  :placeholder="t('admin.invites.recompute.inviterUserId')"
                />
                <div>
                  <label class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">
                    {{ t('admin.invites.recompute.startAt') }}
                  </label>
                  <input v-model="recomputeForm.start_at" type="datetime-local" class="input" />
                </div>
                <div>
                  <label class="mb-1 block text-xs font-medium text-gray-500 dark:text-dark-400">
                    {{ t('admin.invites.recompute.endAt') }}
                  </label>
                  <input v-model="recomputeForm.end_at" type="datetime-local" class="input" />
                </div>
              </div>

              <textarea
                v-model.trim="recomputeForm.reason"
                class="input mt-3 min-h-24"
                :placeholder="t('admin.invites.recompute.reason')"
              />
              <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.invites.recompute.scopeHint') }}
              </p>

              <div class="mt-4 flex flex-wrap gap-3">
                <button
                  type="button"
                  class="btn btn-secondary"
                  :disabled="!canPreviewRecompute || loading.recomputePreview"
                  @click="runRecomputePreview"
                >
                  {{ t('admin.invites.recompute.preview') }}
                </button>
                <input
                  v-model.trim="recomputeForm.confirm_text"
                  type="text"
                  class="input min-w-44 flex-1"
                  placeholder="RECOMPUTE"
                />
                <button
                  data-test="execute-recompute"
                  type="button"
                  class="btn btn-primary"
                  :disabled="!canExecuteRecompute || loading.recomputeExecute"
                  @click="submitRecompute"
                >
                  {{ t('admin.invites.recompute.execute') }}
                </button>
              </div>

              <div class="mt-4 rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/40">
                <template v-if="recomputePreview">
                  <div class="flex flex-wrap gap-4 text-sm text-gray-700 dark:text-dark-300">
                    <span>{{ t('admin.invites.recompute.qualifyingEventCount') }}: {{ recomputePreview.qualifying_event_count }}</span>
                    <span>{{ t('admin.invites.recompute.currentLedgerTotal') }}: {{ formatMoney(previewCurrentTotal) }}</span>
                    <span>{{ t('admin.invites.recompute.expectedLedgerTotal') }}: {{ formatMoney(previewExpectedTotal) }}</span>
                    <span>{{ t('admin.invites.recompute.netDelta') }}: {{ formatMoney(previewDeltaTotal) }}</span>
                  </div>

                  <p
                    v-if="recomputePreviewStale"
                    class="mt-3 text-sm text-amber-700 dark:text-amber-300"
                  >
                    {{ t('admin.invites.recompute.previewStale') }}
                  </p>

                  <div v-if="recomputePreview.deltas.length > 0" class="mt-4 space-y-2">
                    <div
                      v-for="delta in recomputePreview.deltas"
                      :key="`${delta.reward_target_user_id}-${delta.reward_role}`"
                      class="rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-300"
                    >
                      #{{ delta.reward_target_user_id }} · {{ formatRewardRole(delta.reward_role) }} ·
                      {{ formatMoney(delta.delta_amount) }}
                    </div>
                  </div>
                </template>
                <p v-else class="text-sm text-gray-500 dark:text-dark-400">
                  {{ t('admin.invites.recompute.noPreview') }}
                </p>
              </div>
            </article>
          </section>

          <section class="card p-6">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.invites.tables.relationships.title') }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('admin.invites.tables.relationships.description') }}
                </p>
              </div>
              <div class="flex flex-wrap gap-3">
                <input
                  v-model.trim="relationshipFilters.search"
                  type="text"
                  class="input min-w-64"
                  :placeholder="t('admin.invites.tables.relationships.search')"
                  @keydown.enter="resetRelationshipPageAndLoad"
                />
                <button type="button" class="btn btn-secondary" @click="resetRelationshipPageAndLoad">
                  {{ t('common.refresh') }}
                </button>
              </div>
            </div>

            <div class="mt-5">
              <DataTable :columns="relationshipColumns" :data="relationships" :loading="loading.relationships" />
            </div>

            <div class="mt-5">
              <Pagination
                v-if="relationshipPagination.total > 0"
                :page="relationshipPagination.page"
                :total="relationshipPagination.total"
                :page-size="relationshipPagination.page_size"
                @update:page="handleRelationshipPageChange"
                @update:pageSize="handleRelationshipPageSizeChange"
              />
            </div>
          </section>

          <section class="card p-6">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.invites.tables.rewards.title') }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('admin.invites.tables.rewards.description') }}
                </p>
              </div>
              <div class="flex flex-wrap gap-3">
                <input
                  v-model.trim="rewardFilters.search"
                  type="text"
                  class="input min-w-64"
                  :placeholder="t('admin.invites.tables.rewards.search')"
                  @keydown.enter="resetRewardPageAndLoad"
                />
                <select v-model="rewardFilters.reward_type" class="input min-w-44" @change="resetRewardPageAndLoad">
                  <option value="">{{ t('admin.invites.filters.all') }}</option>
                  <option value="base_invite_reward">{{ t('admin.invites.filters.baseReward') }}</option>
                  <option value="manual_invite_grant">{{ t('admin.invites.filters.manualGrant') }}</option>
                  <option value="recompute_delta">{{ t('admin.invites.filters.recomputeDelta') }}</option>
                </select>
                <button type="button" class="btn btn-secondary" @click="resetRewardPageAndLoad">
                  {{ t('common.refresh') }}
                </button>
              </div>
            </div>

            <div class="mt-5">
              <DataTable :columns="rewardColumns" :data="rewards" :loading="loading.rewards" />
            </div>

            <div class="mt-5">
              <Pagination
                v-if="rewardPagination.total > 0"
                :page="rewardPagination.page"
                :total="rewardPagination.total"
                :page-size="rewardPagination.page_size"
                @update:page="handleRewardPageChange"
                @update:pageSize="handleRewardPageSizeChange"
              />
            </div>
          </section>

          <section class="card p-6">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.invites.tables.actions.title') }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('admin.invites.tables.actions.description') }}
                </p>
              </div>
              <div class="flex flex-wrap gap-3">
                <select v-model="actionFilters.action_type" class="input min-w-52" @change="resetActionPageAndLoad">
                  <option value="">{{ t('admin.invites.filters.all') }}</option>
                  <option value="rebind_inviter">{{ t('admin.invites.actionTypes.rebind_inviter') }}</option>
                  <option value="manual_reward_grant">{{ t('admin.invites.actionTypes.manual_reward_grant') }}</option>
                  <option value="recompute_rewards">{{ t('admin.invites.actionTypes.recompute_rewards') }}</option>
                </select>
                <button type="button" class="btn btn-secondary" @click="resetActionPageAndLoad">
                  {{ t('common.refresh') }}
                </button>
              </div>
            </div>

            <div class="mt-5">
              <DataTable :columns="actionColumns" :data="actions" :loading="loading.actions" />
            </div>

            <div class="mt-5">
              <Pagination
                v-if="actionPagination.total > 0"
                :page="actionPagination.page"
                :total="actionPagination.total"
                :page-size="actionPagination.page_size"
                @update:page="handleActionPageChange"
                @update:pageSize="handleActionPageSizeChange"
              />
            </div>
          </section>
        </div>
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import type { Column } from '@/components/common/types'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import adminInvitesAPI from '@/api/admin/invites'
import { useAppStore } from '@/stores/app'
import type {
  AdminInviteAction,
  AdminInviteRelationshipRow,
  AdminInviteRecomputePreview,
  AdminInviteRewardRow,
  AdminInviteStats
} from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loading = reactive({
  stats: false,
  relationships: false,
  rewards: false,
  actions: false,
  rebind: false,
  manualGrant: false,
  recomputePreview: false,
  recomputeExecute: false
})

const stats = ref<AdminInviteStats | null>(null)
const relationships = ref<AdminInviteRelationshipRow[]>([])
const rewards = ref<AdminInviteRewardRow[]>([])
const actions = ref<AdminInviteAction[]>([])

const relationshipPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const rewardPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const actionPagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const relationshipFilters = reactive({
  search: ''
})

const rewardFilters = reactive({
  search: '',
  reward_type: ''
})

const actionFilters = reactive({
  action_type: ''
})

const rebindForm = reactive({
  invitee_user_id: '',
  new_inviter_user_id: '',
  reason: '',
  confirm_text: ''
})

const manualGrantForm = reactive({
  target_user_id: '',
  inviter_user_id: '',
  invitee_user_id: '',
  reward_target_user_id: '',
  reward_role: 'inviter' as 'inviter' | 'invitee',
  reward_amount: '',
  notes: '',
  reason: '',
  confirm_text: ''
})

const recomputeForm = reactive({
  invitee_user_id: '',
  inviter_user_id: '',
  start_at: '',
  end_at: '',
  reason: '',
  confirm_text: ''
})

const recomputePreview = ref<AdminInviteRecomputePreview | null>(null)
const recomputePreviewSignature = ref('')

const isRefreshing = computed(
  () => loading.stats || loading.relationships || loading.rewards || loading.actions
)

const statCards = computed(() => [
  {
    key: 'total_invited_users',
    label: t('admin.invites.summary.totalInvitedUsers'),
    value: String(stats.value?.total_invited_users ?? 0)
  },
  {
    key: 'qualified_reward_users_total',
    label: t('admin.invites.summary.qualifiedRewardUsers'),
    value: String(stats.value?.qualified_reward_users_total ?? 0)
  },
  {
    key: 'base_rewards_total',
    label: t('admin.invites.summary.baseRewardsTotal'),
    value: formatMoney(stats.value?.base_rewards_total)
  },
  {
    key: 'manual_grants_total',
    label: t('admin.invites.summary.manualGrantsTotal'),
    value: formatMoney(stats.value?.manual_grants_total)
  },
  {
    key: 'recompute_adjustments_total',
    label: t('admin.invites.summary.recomputeAdjustmentsTotal'),
    value: formatMoney(stats.value?.recompute_adjustments_total)
  }
])

const relationshipColumns = computed<Column[]>(() => [
  {
    key: 'invitee_email',
    label: t('admin.invites.tables.relationships.columns.invitee')
  },
  {
    key: 'invite_code',
    label: t('admin.invites.tables.relationships.columns.inviteCode')
  },
  {
    key: 'current_inviter_email',
    label: t('admin.invites.tables.relationships.columns.currentInviter'),
    formatter: (_value: string, row: AdminInviteRelationshipRow) =>
      row.current_inviter_email || formatOptionalUserID(row.current_inviter_user_id)
  },
  {
    key: 'invite_bound_at',
    label: t('admin.invites.tables.relationships.columns.boundAt'),
    formatter: (value: string | null) => formatDateTime(value)
  },
  {
    key: 'last_event_type',
    label: t('admin.invites.tables.relationships.columns.lastEvent'),
    formatter: (value: string) => formatEventType(value)
  },
  {
    key: 'last_event_at',
    label: t('admin.invites.tables.relationships.columns.lastEventAt'),
    formatter: (value: string | null) => formatDateTime(value)
  }
])

const rewardColumns = computed<Column[]>(() => [
  {
    key: 'reward_target_email',
    label: t('admin.invites.tables.rewards.columns.targetUser'),
    formatter: (_value: string, row: AdminInviteRewardRow) =>
      formatUserCell(row.reward_target_email, row.reward_target_user_id)
  },
  {
    key: 'inviter_email',
    label: t('admin.invites.tables.rewards.columns.inviter'),
    formatter: (_value: string, row: AdminInviteRewardRow) =>
      formatUserCell(row.inviter_email, row.inviter_user_id)
  },
  {
    key: 'invitee_email',
    label: t('admin.invites.tables.rewards.columns.invitee'),
    formatter: (_value: string, row: AdminInviteRewardRow) =>
      formatUserCell(row.invitee_email, row.invitee_user_id)
  },
  {
    key: 'reward_role',
    label: t('admin.invites.tables.rewards.columns.rewardRole'),
    formatter: (value: string) => formatRewardRole(value)
  },
  {
    key: 'reward_type',
    label: t('admin.invites.tables.rewards.columns.rewardType'),
    formatter: (value: string) => formatRewardType(value)
  },
  {
    key: 'reward_amount',
    label: t('admin.invites.tables.rewards.columns.amount'),
    formatter: (value: number) => formatMoney(value)
  },
  {
    key: 'created_at',
    label: t('admin.invites.tables.rewards.columns.createdAt'),
    formatter: (value: string) => formatDateTime(value)
  },
  {
    key: 'trigger_redeem_code_id',
    label: t('admin.invites.tables.rewards.columns.trigger'),
    formatter: (_value: number | undefined, row: AdminInviteRewardRow) => {
      if (row.trigger_redeem_code_id) {
        return `redeem#${row.trigger_redeem_code_id}`
      }
      if (row.admin_action_id) {
        return `action#${row.admin_action_id}`
      }
      return '--'
    }
  }
])

const actionColumns = computed<Column[]>(() => [
  {
    key: 'id',
    label: t('admin.invites.tables.actions.columns.id')
  },
  {
    key: 'action_type',
    label: t('admin.invites.tables.actions.columns.actionType'),
    formatter: (value: string) => formatActionType(value)
  },
  {
    key: 'operator_user_id',
    label: t('admin.invites.tables.actions.columns.operatorUserId')
  },
  {
    key: 'target_user_id',
    label: t('admin.invites.tables.actions.columns.targetUserId')
  },
  {
    key: 'reason',
    label: t('admin.invites.tables.actions.columns.reason'),
    class: 'max-w-[280px]'
  },
  {
    key: 'created_at',
    label: t('admin.invites.tables.actions.columns.createdAt'),
    formatter: (value: string) => formatDateTime(value)
  }
])

const canSubmitRebind = computed(() =>
  Boolean(
    parsePositiveInt(rebindForm.invitee_user_id) &&
      parsePositiveInt(rebindForm.new_inviter_user_id) &&
      rebindForm.reason.trim() &&
      rebindForm.confirm_text === 'REBIND'
  )
)

const canSubmitManualGrant = computed(() =>
  Boolean(
    parsePositiveInt(manualGrantForm.target_user_id) &&
      parsePositiveInt(manualGrantForm.inviter_user_id) &&
      parsePositiveInt(manualGrantForm.invitee_user_id) &&
      parsePositiveInt(manualGrantForm.reward_target_user_id) &&
      parsePositiveFloat(manualGrantForm.reward_amount) &&
      manualGrantForm.reason.trim() &&
      manualGrantForm.confirm_text === 'GRANT'
  )
)

const hasRecomputeScope = computed(
  () =>
    Boolean(
      parsePositiveInt(recomputeForm.invitee_user_id) ||
        parsePositiveInt(recomputeForm.inviter_user_id) ||
        normalizeDateTime(recomputeForm.start_at) ||
        normalizeDateTime(recomputeForm.end_at)
    )
)

const canPreviewRecompute = computed(
  () => Boolean(recomputeForm.reason.trim()) && hasRecomputeScope.value
)

const recomputeSignature = computed(() => JSON.stringify(buildRecomputePreviewPayload()))

const recomputePreviewStale = computed(
  () =>
    Boolean(recomputePreview.value) && recomputePreviewSignature.value !== recomputeSignature.value
)

const canExecuteRecompute = computed(
  () =>
    Boolean(recomputePreview.value) &&
    !recomputePreviewStale.value &&
    recomputeForm.confirm_text === 'RECOMPUTE'
)

const previewCurrentTotal = computed(() =>
  (recomputePreview.value?.deltas ?? []).reduce((sum, item) => sum + item.current_amount, 0)
)

const previewExpectedTotal = computed(() =>
  (recomputePreview.value?.deltas ?? []).reduce((sum, item) => sum + item.expected_amount, 0)
)

const previewDeltaTotal = computed(() =>
  (recomputePreview.value?.deltas ?? []).reduce((sum, item) => sum + item.delta_amount, 0)
)

function formatMoney(value?: number): string {
  return `$${(value ?? 0).toFixed(2)}`
}

function formatDateTime(value?: string | null): string {
  if (!value) return '--'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString()
}

function parsePositiveInt(raw: string): number | undefined {
  if (!raw.trim()) return undefined
  const value = Number(raw)
  if (!Number.isInteger(value) || value <= 0) return undefined
  return value
}

function parsePositiveFloat(raw: string): number | undefined {
  if (!raw.trim()) return undefined
  const value = Number(raw)
  if (!Number.isFinite(value) || value <= 0) return undefined
  return value
}

function normalizeDateTime(value: string): string | undefined {
  if (!value.trim()) return undefined
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return undefined
  return parsed.toISOString()
}

function formatOptionalUserID(value: number | null): string {
  return value ? `#${value}` : '--'
}

function formatUserCell(email: string, userID: number): string {
  return email ? `${email} (#${userID})` : `#${userID}`
}

function formatRewardRole(value: string): string {
  return t(`admin.invites.roles.${value}`)
}

function formatRewardType(value: string): string {
  return t(`admin.invites.rewardTypes.${value}`)
}

function formatActionType(value: string): string {
  return t(`admin.invites.actionTypes.${value}`)
}

function formatEventType(value: string): string {
  return t(`admin.invites.eventTypes.${value}`)
}

function extractErrorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'object' && error !== null && 'message' in error) {
    const message = (error as { message?: string }).message
    if (message) return message
  }
  return fallback
}

async function loadStats(): Promise<void> {
  loading.stats = true
  try {
    stats.value = await adminInvitesAPI.getStats()
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('errors.networkError')))
  } finally {
    loading.stats = false
  }
}

async function loadRelationships(): Promise<void> {
  loading.relationships = true
  try {
    const page = await adminInvitesAPI.listRelationships(
      relationshipPagination.page,
      relationshipPagination.page_size,
      {
        search: relationshipFilters.search.trim() || undefined
      }
    )
    relationships.value = page.items
    relationshipPagination.total = page.total
    relationshipPagination.pages = page.pages
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('errors.networkError')))
  } finally {
    loading.relationships = false
  }
}

async function loadRewards(): Promise<void> {
  loading.rewards = true
  try {
    const page = await adminInvitesAPI.listRewards(rewardPagination.page, rewardPagination.page_size, {
      search: rewardFilters.search.trim() || undefined,
      reward_type: rewardFilters.reward_type || undefined
    })
    rewards.value = page.items
    rewardPagination.total = page.total
    rewardPagination.pages = page.pages
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('errors.networkError')))
  } finally {
    loading.rewards = false
  }
}

async function loadActions(): Promise<void> {
  loading.actions = true
  try {
    const page = await adminInvitesAPI.listActions(actionPagination.page, actionPagination.page_size, {
      action_type: actionFilters.action_type || undefined
    })
    actions.value = page.items
    actionPagination.total = page.total
    actionPagination.pages = page.pages
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('errors.networkError')))
  } finally {
    loading.actions = false
  }
}

async function refreshAll(): Promise<void> {
  await Promise.all([loadStats(), loadRelationships(), loadRewards(), loadActions()])
}

function resetRelationshipPageAndLoad(): void {
  relationshipPagination.page = 1
  void loadRelationships()
}

function resetRewardPageAndLoad(): void {
  rewardPagination.page = 1
  void loadRewards()
}

function resetActionPageAndLoad(): void {
  actionPagination.page = 1
  void loadActions()
}

function handleRelationshipPageChange(page: number): void {
  relationshipPagination.page = page
  void loadRelationships()
}

function handleRelationshipPageSizeChange(pageSize: number): void {
  relationshipPagination.page = 1
  relationshipPagination.page_size = pageSize
  void loadRelationships()
}

function handleRewardPageChange(page: number): void {
  rewardPagination.page = page
  void loadRewards()
}

function handleRewardPageSizeChange(pageSize: number): void {
  rewardPagination.page = 1
  rewardPagination.page_size = pageSize
  void loadRewards()
}

function handleActionPageChange(page: number): void {
  actionPagination.page = page
  void loadActions()
}

function handleActionPageSizeChange(pageSize: number): void {
  actionPagination.page = 1
  actionPagination.page_size = pageSize
  void loadActions()
}

function resetRebindForm(): void {
  rebindForm.invitee_user_id = ''
  rebindForm.new_inviter_user_id = ''
  rebindForm.reason = ''
  rebindForm.confirm_text = ''
}

function resetManualGrantForm(): void {
  manualGrantForm.target_user_id = ''
  manualGrantForm.inviter_user_id = ''
  manualGrantForm.invitee_user_id = ''
  manualGrantForm.reward_target_user_id = ''
  manualGrantForm.reward_role = 'inviter'
  manualGrantForm.reward_amount = ''
  manualGrantForm.notes = ''
  manualGrantForm.reason = ''
  manualGrantForm.confirm_text = ''
}

function resetRecomputePreview(): void {
  recomputePreview.value = null
  recomputePreviewSignature.value = ''
}

function buildRecomputePreviewPayload() {
  return {
    reason: recomputeForm.reason.trim(),
    invitee_user_id: parsePositiveInt(recomputeForm.invitee_user_id),
    inviter_user_id: parsePositiveInt(recomputeForm.inviter_user_id),
    start_at: normalizeDateTime(recomputeForm.start_at),
    end_at: normalizeDateTime(recomputeForm.end_at)
  }
}

async function submitRebind(): Promise<void> {
  const inviteeUserID = parsePositiveInt(rebindForm.invitee_user_id)
  const newInviterUserID = parsePositiveInt(rebindForm.new_inviter_user_id)
  if (!inviteeUserID || !newInviterUserID || !rebindForm.reason.trim()) {
    return
  }

  loading.rebind = true
  try {
    await adminInvitesAPI.rebindInviter({
      invitee_user_id: inviteeUserID,
      new_inviter_user_id: newInviterUserID,
      reason: rebindForm.reason.trim()
    })
    appStore.showSuccess(t('admin.invites.rebind.success'))
    resetRebindForm()
    await refreshAll()
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('admin.invites.rebind.failure')))
  } finally {
    loading.rebind = false
  }
}

async function submitManualGrant(): Promise<void> {
  const targetUserID = parsePositiveInt(manualGrantForm.target_user_id)
  const inviterUserID = parsePositiveInt(manualGrantForm.inviter_user_id)
  const inviteeUserID = parsePositiveInt(manualGrantForm.invitee_user_id)
  const rewardTargetUserID = parsePositiveInt(manualGrantForm.reward_target_user_id)
  const rewardAmount = parsePositiveFloat(manualGrantForm.reward_amount)

  if (!targetUserID || !inviterUserID || !inviteeUserID || !rewardTargetUserID || !rewardAmount) {
    return
  }

  loading.manualGrant = true
  try {
    await adminInvitesAPI.createManualGrant({
      target_user_id: targetUserID,
      reason: manualGrantForm.reason.trim(),
      lines: [
        {
          inviter_user_id: inviterUserID,
          invitee_user_id: inviteeUserID,
          reward_target_user_id: rewardTargetUserID,
          reward_role: manualGrantForm.reward_role,
          reward_amount: rewardAmount,
          notes: manualGrantForm.notes.trim() || undefined
        }
      ]
    })
    appStore.showSuccess(t('admin.invites.manualGrant.success'))
    resetManualGrantForm()
    await refreshAll()
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('admin.invites.manualGrant.failure')))
  } finally {
    loading.manualGrant = false
  }
}

async function runRecomputePreview(): Promise<void> {
  if (!canPreviewRecompute.value) {
    return
  }

  loading.recomputePreview = true
  try {
    recomputePreview.value = await adminInvitesAPI.previewRecompute(buildRecomputePreviewPayload())
    recomputePreviewSignature.value = recomputeSignature.value
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('errors.networkError')))
    resetRecomputePreview()
  } finally {
    loading.recomputePreview = false
  }
}

async function submitRecompute(): Promise<void> {
  if (!recomputePreview.value || recomputePreviewStale.value) {
    return
  }

  loading.recomputeExecute = true
  try {
    await adminInvitesAPI.executeRecompute({
      ...buildRecomputePreviewPayload(),
      scope_hash: recomputePreview.value.scope_hash
    })
    appStore.showSuccess(t('admin.invites.recompute.success'))
    resetRecomputePreview()
    recomputeForm.confirm_text = ''
    await refreshAll()
  } catch (error) {
    appStore.showError(extractErrorMessage(error, t('admin.invites.recompute.failure')))
  } finally {
    loading.recomputeExecute = false
  }
}

onMounted(() => {
  void refreshAll()
})
</script>
