<template>
  <AppLayout>
    <div class="flex h-full flex-col">
      <!-- Toolbar -->
      <div class="flex shrink-0 items-center justify-between border-b border-gray-200 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-900">
        <h1 class="text-base font-semibold text-gray-900 dark:text-white">
          {{ t('imageStudio.pageTitle') }}
        </h1>
        <a
          v-if="canvasUrl"
          :href="canvasUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="btn btn-secondary btn-sm flex items-center gap-1.5"
        >
          <Icon name="externalLink" size="sm" />
          {{ t('imageStudio.openInNewWindow') }}
        </a>
      </div>

      <!-- Loading state -->
      <div v-if="loadState === 'loading'" class="flex flex-1 items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent" />
      </div>

      <!-- Error state -->
      <div
        v-else-if="loadState === 'error'"
        class="flex flex-1 flex-col items-center justify-center gap-4 text-center"
      >
        <p class="text-sm text-red-500 dark:text-red-400">{{ t('imageStudio.loadError') }}</p>
        <button class="btn btn-secondary btn-sm" @click="loadKey">
          {{ t('common.retry') }}
        </button>
      </div>

      <!-- No enabled key -->
      <div
        v-else-if="loadState === 'no-key'"
        class="flex flex-1 flex-col items-center justify-center gap-3 text-center"
      >
        <p class="text-sm text-gray-600 dark:text-dark-300">
          {{ t('imageStudio.noEnabledKey') }}
        </p>
        <router-link to="/keys" class="btn btn-primary btn-sm">
          {{ t('imageStudio.goToKeys') }}
        </router-link>
      </div>

      <!-- Canvas iframe -->
      <iframe
        v-else-if="loadState === 'ready' && canvasUrl"
        :src="canvasUrl"
        class="image-studio-frame flex-1 border-0"
        sandbox="allow-scripts allow-same-origin allow-forms allow-popups allow-downloads"
        referrerpolicy="no-referrer"
        :title="t('imageStudio.pageTitle')"
        allow="clipboard-write"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI } from '@/api/keys'
import type { ApiKey } from '@/types'

const CANVAS_BASE_URL = 'https://canvas.xlabapi.com'
const API_GATEWAY_URL = 'https://api.xlabapi.com'

type LoadState = 'loading' | 'error' | 'no-key' | 'ready'

const { t } = useI18n()

const loadState = ref<LoadState>('loading')
const canvasUrl = ref<string | null>(null)

function buildCanvasUrl(apiKey: string): string {
  const url = new URL(CANVAS_BASE_URL)
  url.searchParams.set('apiKey', apiKey)
  url.searchParams.set('baseUrl', API_GATEWAY_URL)
  return url.toString()
}

async function loadKey(): Promise<void> {
  loadState.value = 'loading'
  canvasUrl.value = null

  try {
    const result = await keysAPI.list(1, 20)
    const enabledKey = result.items.find((k: ApiKey) => k.status === 'active') ?? null

    if (!enabledKey) {
      loadState.value = 'no-key'
      return
    }

    canvasUrl.value = buildCanvasUrl(enabledKey.key)
    loadState.value = 'ready'
  } catch {
    loadState.value = 'error'
  }
}

onMounted(loadKey)
</script>

<style scoped>
.image-studio-frame {
  width: 100%;
  min-height: 0;
}
</style>
