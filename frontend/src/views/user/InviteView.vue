<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6">
      <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div class="card p-6">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('invite.myCode') }}</p>
          <p class="mt-3 text-2xl font-semibold tracking-[0.18em] text-gray-900 dark:text-white">
            {{ summary?.invite_code || '--' }}
          </p>
        </div>

        <div class="card p-6">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('invite.invitedUsers') }}</p>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">
            {{ summary?.invited_users_total ?? 0 }}
          </p>
        </div>

        <div class="card p-6">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('invite.totalRecharge') }}</p>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">
            {{ formatMoney(summary?.invitees_recharge_total) }}
          </p>
        </div>

        <div class="card p-6">
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('invite.totalRewards') }}</p>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">
            {{ formatMoney(summary?.base_rewards_total) }}
          </p>
        </div>
      </section>

      <section class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div class="space-y-2">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('invite.link') }}</p>
            <p class="text-sm text-gray-600 dark:text-dark-300">{{ t('invite.description') }}</p>
          </div>
          <button type="button" class="btn btn-primary" @click="copyLink">
            {{ t('invite.copyLink') }}
          </button>
        </div>

        <div class="mt-4 rounded-2xl border border-dashed border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/40">
          <input
            :value="summary?.invite_link || ''"
            class="input w-full bg-white dark:bg-dark-800"
            readonly
          />
        </div>
      </section>

      <section class="card p-6">
        <div class="flex items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('invite.rewardHistory') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ totalRewardsCountText }}
            </p>
          </div>
        </div>

        <div v-if="rewards.length > 0" class="mt-5 space-y-3">
          <div
            v-for="record in rewards"
            :key="`${record.created_at}-${record.reward_role}-${record.reward_amount}`"
            class="flex flex-col gap-3 rounded-2xl border border-gray-100 p-4 dark:border-dark-800 sm:flex-row sm:items-center sm:justify-between"
          >
            <div>
              <p class="font-medium text-gray-900 dark:text-white">
                {{ t(`invite.roles.${record.reward_role}`) }}
              </p>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ formatDate(record.created_at) }}
              </p>
            </div>
            <div class="text-right">
              <p class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ formatMoney(record.reward_amount) }}
              </p>
              <p class="mt-1 text-xs uppercase tracking-[0.16em] text-gray-400 dark:text-dark-500">
                {{ record.reward_type }}
              </p>
            </div>
          </div>
        </div>

        <div
          v-else
          class="mt-5 rounded-2xl border border-dashed border-gray-200 px-6 py-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400"
        >
          {{ t('invite.emptyRewards') }}
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { inviteAPI } from '@/api/invite'
import type { InviteRewardRecord, InviteSummary } from '@/types'

const { t } = useI18n()

const summary = ref<InviteSummary | null>(null)
const rewards = ref<InviteRewardRecord[]>([])
const rewardsTotal = ref(0)

const totalRewardsCountText = computed(() =>
  t('invite.totalRewardsCount', { total: rewardsTotal.value })
)

function formatMoney(value?: number): string {
  return `$${(value ?? 0).toFixed(2)}`
}

function formatDate(value: string): string {
  if (!value) return '--'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString()
}

async function load(): Promise<void> {
  summary.value = await inviteAPI.getSummary()
  const page = await inviteAPI.listRewards()
  rewards.value = page.items
  rewardsTotal.value = page.total
}

async function copyLink(): Promise<void> {
  const inviteLink = summary.value?.invite_link
  if (!inviteLink || !navigator?.clipboard?.writeText) return
  await navigator.clipboard.writeText(inviteLink)
}

onMounted(() => {
  void load()
})
</script>
