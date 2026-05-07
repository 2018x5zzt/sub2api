<template>
  <AppLayout>
    <div class="image-studio-page">
      <div class="image-studio-toolbar">
        <div class="image-studio-title">
          <p class="image-studio-kicker">创作图片</p>
          <h1>{{ activeMode === 'new' ? '新版 Miku 生图' : '旧版生图工具' }}</h1>
        </div>

        <div class="image-studio-actions">
          <a
            :href="activeUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm"
          >
            <Icon name="externalLink" size="sm" class="mr-1.5" :stroke-width="2" />
            新窗口打开
          </a>
          <button
            v-if="activeMode === 'new'"
            type="button"
            class="btn btn-secondary btn-sm"
            data-testid="legacy-image-studio"
            @click="activeMode = 'legacy'"
          >
            回到旧版
          </button>
          <button
            v-else
            type="button"
            class="btn btn-primary btn-sm"
            data-testid="new-image-studio"
            @click="activeMode = 'new'"
          >
            返回新版
          </button>
        </div>
      </div>

      <div class="image-studio-frame-shell">
        <iframe
          :src="activeUrl"
          class="image-studio-frame"
          allowfullscreen
        ></iframe>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { xlabOAuthAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

const MIKU_CALLBACK_URL = 'https://ai.mikuapi.org/auth/xlab/callback?layout=horizontal'
const MIKU_CLIENT_ID = 'miku-app'
const MIKU_LOGIN_URL = 'https://ai.mikuapi.org/login?xlab_auto=1'
const LEGACY_IMAGE_STUDIO_URL = 'https://iframe.mikuapi.org/?codexCli=true'

const activeMode = ref<'new' | 'legacy'>('new')
const mikuOAuthUrl = ref(MIKU_LOGIN_URL)

const activeUrl = computed(() => (
  activeMode.value === 'new' ? mikuOAuthUrl.value : LEGACY_IMAGE_STUDIO_URL
))

async function refreshMikuOAuthUrl(): Promise<void> {
  try {
    const result = await xlabOAuthAPI.authorize({
      client_id: MIKU_CLIENT_ID,
      redirect_uri: MIKU_CALLBACK_URL,
      response_type: 'code',
      scope: 'profile email',
      state: 'image-studio',
    })
    mikuOAuthUrl.value = result.redirect_uri
  } catch {
    mikuOAuthUrl.value = MIKU_LOGIN_URL
  }
}

watch(activeMode, (mode) => {
  if (mode === 'new') {
    void refreshMikuOAuthUrl()
  }
})

void refreshMikuOAuthUrl()
</script>

<style scoped>
.image-studio-page {
  @apply flex min-h-0 flex-col gap-3;
  height: calc(100vh - 64px - 4rem);
}

.image-studio-toolbar {
  @apply flex flex-col gap-3 rounded-xl border border-gray-200 bg-white px-4 py-3 shadow-sm;
  @apply dark:border-dark-700 dark:bg-dark-800;
}

.image-studio-title {
  @apply min-w-0;
}

.image-studio-kicker {
  @apply text-xs font-medium text-gray-500 dark:text-dark-400;
}

.image-studio-title h1 {
  @apply mt-0.5 text-lg font-semibold text-gray-900 dark:text-white;
}

.image-studio-actions {
  @apply flex flex-wrap items-center gap-2;
}

.image-studio-frame-shell {
  @apply min-h-0 flex-1 overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm;
  @apply dark:border-dark-700 dark:bg-dark-900;
}

.image-studio-frame {
  display: block;
  width: 100%;
  height: 100%;
  border: 0;
  background: transparent;
}

@media (min-width: 768px) {
  .image-studio-toolbar {
    @apply flex-row items-center justify-between;
  }
}
</style>
