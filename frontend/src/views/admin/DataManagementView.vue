<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
              {{ t('admin.dataManagement.title') }}
            </h1>
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.dataManagement.description') }}
            </p>
          </div>
          <button class="btn btn-secondary" :disabled="loading" @click="reload">
            {{ loading ? t('common.loading') : t('admin.dataManagement.actions.refresh') }}
          </button>
        </div>
      </section>

      <section class="card p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.dataManagement.agent.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.dataManagement.agent.description') }}
            </p>
          </div>
          <span
            class="inline-flex w-fit items-center rounded px-2.5 py-1 text-xs font-medium"
            :class="agentHealth?.enabled ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'"
          >
            {{ agentHealth?.enabled ? t('common.enabled') : t('common.disabled') }}
          </span>
        </div>

        <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <InfoTile :label="t('admin.dataManagement.agent.socketPath')" :value="agentHealth?.socket_path || '-'" />
          <InfoTile :label="t('admin.dataManagement.agent.status')" :value="agentHealth?.agent?.status || '-'" />
          <InfoTile :label="t('admin.dataManagement.agent.version')" :value="agentHealth?.agent?.version || '-'" />
          <InfoTile :label="t('admin.dataManagement.agent.uptime')" :value="formatUptime(agentHealth?.agent?.uptime_seconds)" />
        </div>

        <div
          v-if="agentHealth && !agentHealth.enabled"
          class="mt-5 rounded border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-900/40 dark:bg-amber-900/10 dark:text-amber-200"
        >
          <div class="font-medium">{{ t('admin.dataManagement.agent.reasonLabel') }}</div>
          <div class="mt-1">{{ reasonText(agentHealth.reason) }}</div>
          <div class="mt-2 text-xs opacity-80">{{ t('admin.dataManagement.actions.disabledHint') }}</div>
        </div>
      </section>

      <section class="grid gap-6 xl:grid-cols-2">
        <div class="card p-6">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.dataManagement.sections.config.title') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.dataManagement.sections.config.description') }}
              </p>
            </div>
          </div>

          <div v-if="config" class="grid gap-3 text-sm">
            <InfoRow :label="t('admin.dataManagement.form.sourceMode')" :value="config.source_mode" />
            <InfoRow :label="t('admin.dataManagement.form.backupRoot')" :value="config.backup_root || '-'" />
            <InfoRow :label="t('admin.dataManagement.form.retentionDays')" :value="String(config.retention_days)" />
            <InfoRow :label="t('admin.dataManagement.form.keepLast')" :value="String(config.keep_last)" />
            <InfoRow :label="t('admin.dataManagement.form.activePostgresProfile')" :value="config.active_postgres_profile_id || '-'" />
            <InfoRow :label="t('admin.dataManagement.form.activeRedisProfile')" :value="config.active_redis_profile_id || '-'" />
            <InfoRow :label="t('admin.dataManagement.form.activeS3Profile')" :value="config.active_s3_profile_id || '-'" />
          </div>
          <div v-else class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
            {{ loading ? t('common.loading') : t('common.notAvailable') }}
          </div>
        </div>

        <div class="card p-6">
          <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.dataManagement.sections.backup.title') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.dataManagement.sections.backup.description') }}
              </p>
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-3">
            <button
              v-for="type in backupTypes"
              :key="type"
              class="btn btn-secondary"
              :disabled="creatingBackup || !agentHealth?.enabled"
              @click="createBackup(type)"
            >
              {{ type }}
            </button>
          </div>
        </div>
      </section>

      <section class="card p-6">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.dataManagement.sections.s3.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.dataManagement.sections.s3.description') }}
            </p>
          </div>
          <button class="btn btn-secondary btn-sm" :disabled="loadingS3Profiles" @click="loadS3Profiles">
            {{ loadingS3Profiles ? t('common.loading') : t('admin.dataManagement.actions.reloadProfiles') }}
          </button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full min-w-[760px] text-sm">
            <thead>
              <tr class="border-b border-gray-200 text-left text-xs uppercase text-gray-500 dark:border-dark-700 dark:text-gray-400">
                <th class="py-2 pr-4">{{ t('admin.dataManagement.s3Profiles.columns.profile') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.s3Profiles.columns.active') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.s3Profiles.columns.storage') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.s3Profiles.columns.updatedAt') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="profile in s3Profiles" :key="profile.profile_id" class="border-b border-gray-100 dark:border-dark-800">
                <td class="py-3 pr-4">
                  <div class="font-mono text-xs text-gray-900 dark:text-white">{{ profile.profile_id }}</div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ profile.name }}</div>
                </td>
                <td class="py-3 pr-4">
                  <span
                    class="rounded px-2 py-0.5 text-xs"
                    :class="profile.is_active ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-gray-300'"
                  >
                    {{ profile.is_active ? t('common.enabled') : t('common.disabled') }}
                  </span>
                </td>
                <td class="py-3 pr-4 text-xs text-gray-600 dark:text-gray-300">
                  <div>{{ profile.s3.bucket || '-' }}</div>
                  <div class="mt-1 text-gray-500 dark:text-gray-400">{{ profile.s3.region || '-' }} {{ profile.s3.prefix || '' }}</div>
                </td>
                <td class="py-3 pr-4 text-xs text-gray-500 dark:text-gray-400">{{ formatDate(profile.updated_at) }}</td>
              </tr>
              <tr v-if="!loadingS3Profiles && s3Profiles.length === 0">
                <td colspan="4" class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.dataManagement.s3Profiles.empty') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="card p-6">
        <div class="mb-4 flex items-center justify-between gap-3">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.dataManagement.sections.history.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.dataManagement.sections.history.description') }}
            </p>
          </div>
          <button class="btn btn-secondary btn-sm" :disabled="loadingJobs" @click="loadJobs">
            {{ loadingJobs ? t('common.loading') : t('admin.dataManagement.actions.refreshJobs') }}
          </button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full min-w-[920px] text-sm">
            <thead>
              <tr class="border-b border-gray-200 text-left text-xs uppercase text-gray-500 dark:border-dark-700 dark:text-gray-400">
                <th class="py-2 pr-4">{{ t('admin.dataManagement.history.columns.jobID') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.history.columns.type') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.history.columns.status') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.history.columns.finishedAt') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.history.columns.artifact') }}</th>
                <th class="py-2 pr-4">{{ t('admin.dataManagement.history.columns.error') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="job in jobs" :key="job.job_id" class="border-b border-gray-100 align-top dark:border-dark-800">
                <td class="py-3 pr-4 font-mono text-xs">{{ job.job_id }}</td>
                <td class="py-3 pr-4">{{ job.backup_type }}</td>
                <td class="py-3 pr-4">{{ statusText(job.status) }}</td>
                <td class="py-3 pr-4 text-xs text-gray-500 dark:text-gray-400">{{ formatDate(job.finished_at || job.started_at) }}</td>
                <td class="py-3 pr-4 text-xs text-gray-600 dark:text-gray-300">{{ formatArtifact(job) }}</td>
                <td class="py-3 pr-4 text-xs text-red-600 dark:text-red-300">{{ job.error_message || '-' }}</td>
              </tr>
              <tr v-if="!loadingJobs && jobs.length === 0">
                <td colspan="6" class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.dataManagement.history.empty') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineComponent, h, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { adminAPI } from '@/api'
import type {
  BackupAgentHealth,
  BackupJob,
  BackupType,
  DataManagementConfig,
  DataManagementS3Profile,
} from '@/api/admin/dataManagement'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const loadingS3Profiles = ref(false)
const loadingJobs = ref(false)
const creatingBackup = ref(false)

const agentHealth = ref<BackupAgentHealth | null>(null)
const config = ref<DataManagementConfig | null>(null)
const s3Profiles = ref<DataManagementS3Profile[]>([])
const jobs = ref<BackupJob[]>([])
const backupTypes: BackupType[] = ['postgres', 'redis', 'full']

const InfoTile = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'rounded border border-gray-200 p-3 dark:border-dark-700' }, [
      h('div', { class: 'text-xs text-gray-500 dark:text-gray-400' }, props.label),
      h('div', { class: 'mt-1 break-all text-sm font-medium text-gray-900 dark:text-white' }, props.value),
    ])
  },
})

const InfoRow = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'flex items-center justify-between gap-4 border-b border-gray-100 py-2 dark:border-dark-800' }, [
      h('span', { class: 'text-gray-500 dark:text-gray-400' }, props.label),
      h('span', { class: 'text-right font-medium text-gray-900 dark:text-white' }, props.value),
    ])
  },
})

async function reload() {
  loading.value = true
  try {
    await Promise.all([loadAgentHealth(), loadConfig(), loadS3Profiles(), loadJobs()])
  } finally {
    loading.value = false
  }
}

async function loadAgentHealth() {
  try {
    agentHealth.value = await adminAPI.dataManagement.getAgentHealth()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function loadConfig() {
  try {
    config.value = await adminAPI.dataManagement.getConfig()
  } catch (error) {
    config.value = null
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function loadS3Profiles() {
  loadingS3Profiles.value = true
  try {
    const response = await adminAPI.dataManagement.listS3Profiles()
    s3Profiles.value = response.items || []
  } catch (error) {
    s3Profiles.value = []
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    loadingS3Profiles.value = false
  }
}

async function loadJobs() {
  loadingJobs.value = true
  try {
    const response = await adminAPI.dataManagement.listBackupJobs({ page_size: 20 })
    jobs.value = response.items || []
  } catch (error) {
    jobs.value = []
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    loadingJobs.value = false
  }
}

async function createBackup(backupType: BackupType) {
  creatingBackup.value = true
  try {
    const response = await adminAPI.dataManagement.createBackupJob({
      backup_type: backupType,
      upload_to_s3: true,
    })
    appStore.showSuccess(t('admin.dataManagement.actions.jobCreated', {
      jobID: response.job_id,
      status: response.status,
    }))
    await loadJobs()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    creatingBackup.value = false
  }
}

function reasonText(reason?: string): string {
  if (!reason) return '-'
  const key = `admin.dataManagement.agent.reason.${reason}`
  const translated = t(key)
  return translated === key ? reason : translated
}

function statusText(status: string): string {
  const key = `admin.dataManagement.history.status.${status}`
  const translated = t(key)
  return translated === key ? status : translated
}

function formatDate(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function formatUptime(seconds?: number): string {
  if (!seconds || seconds <= 0) return '-'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

function formatArtifact(job: BackupJob): string {
  if (job.s3?.bucket && job.s3?.key) return `${job.s3.bucket}/${job.s3.key}`
  if (job.artifact?.local_path) return job.artifact.local_path
  return '-'
}

onMounted(() => {
  void reload()
})
</script>
