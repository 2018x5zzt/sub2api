<template>
  <div class="min-h-screen" style="background: var(--bg-0)">
    <!-- 背景装饰：极淡的网状光晕 + 网格 -->
    <div
      class="pointer-events-none fixed inset-0 bg-mesh-gradient opacity-60"
    ></div>
    <div
      class="pointer-events-none fixed inset-0 grid-bg"
      style="
        opacity: 0.35;
        mask-image: radial-gradient(
          ellipse 80% 60% at 50% 0%,
          #000 30%,
          transparent 80%
        );
        -webkit-mask-image: radial-gradient(
          ellipse 80% 60% at 50% 0%,
          #000 30%,
          transparent 80%
        );
      "
    ></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main class="p-4 md:p-6 lg:p-8">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
