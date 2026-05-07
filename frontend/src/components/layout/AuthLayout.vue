<template>
  <div
    class="relative flex min-h-screen items-center justify-center overflow-hidden p-4"
    style="background: var(--bg-0)"
  >
    <!-- 背景：网状光晕 -->
    <div
      class="pointer-events-none absolute inset-0 bg-mesh-gradient"
      style="opacity: 0.7"
    ></div>

    <!-- 装饰光斑（cyan + violet glow） -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute -right-40 -top-40 h-96 w-96 rounded-full blur-3xl"
        style="background: rgba(125, 211, 252, 0.18)"
      ></div>
      <div
        class="absolute -bottom-40 -left-40 h-96 w-96 rounded-full blur-3xl"
        style="background: rgba(167, 139, 250, 0.14)"
      ></div>

      <!-- 网格背景 -->
      <div
        class="absolute inset-0 grid-bg"
        style="
          opacity: 0.5;
          mask-image: radial-gradient(
            ellipse 70% 60% at 50% 30%,
            #000 30%,
            transparent 80%
          );
          -webkit-mask-image: radial-gradient(
            ellipse 70% 60% at 50% 30%,
            #000 30%,
            transparent 80%
          );
        "
      ></div>
    </div>

    <!-- Content Container -->
    <div class="relative z-10 w-full max-w-md">
      <!-- Logo/Brand -->
      <div class="mb-8 text-center">
        <template v-if="settingsLoaded">
          <div
            class="mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-[14px] shadow-glow"
            style="background: var(--bg-2); border: 1px solid var(--line-2)"
          >
            <img
              :src="siteLogo || '/logo.png'"
              alt="Logo"
              class="h-full w-full object-contain"
            />
          </div>
          <h1
            class="text-gradient mb-2"
            style="
              font-family: 'Inter', system-ui, sans-serif;
              font-size: 32px;
              font-weight: 500;
              letter-spacing: -0.02em;
            "
          >
            {{ siteName }}
          </h1>
          <p style="color: var(--text-3); font-size: 13px">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- Card Container -->
      <div class="card-glass" style="padding: 32px; border-radius: 14px">
        <slot />
      </div>

      <!-- Footer Links -->
      <div
        class="mt-6 text-center"
        style="font-size: 13px; color: var(--text-3)"
      >
        <slot name="footer" />
      </div>

      <!-- Copyright -->
      <div
        class="mt-8 text-center mono"
        style="font-size: 11px; color: var(--text-4)"
      >
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

