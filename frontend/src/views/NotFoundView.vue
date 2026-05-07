<template>
  <div
    class="relative flex min-h-screen items-center justify-center overflow-hidden px-4"
    style="background: var(--bg-0)"
  >
    <!-- 背景装饰 -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute -right-40 -top-40 h-96 w-96 rounded-full blur-3xl"
        style="background: rgba(125, 211, 252, 0.10)"
      ></div>
      <div
        class="absolute -bottom-40 -left-40 h-96 w-96 rounded-full blur-3xl"
        style="background: rgba(167, 139, 250, 0.08)"
      ></div>
      <div
        class="absolute inset-0 grid-bg"
        style="
          opacity: 0.4;
          mask-image: radial-gradient(
            ellipse 60% 50% at 50% 50%,
            #000 30%,
            transparent 80%
          );
          -webkit-mask-image: radial-gradient(
            ellipse 60% 50% at 50% 50%,
            #000 30%,
            transparent 80%
          );
        "
      ></div>
    </div>

    <div class="relative z-10 w-full max-w-md text-center">
      <!-- eyebrow -->
      <div class="eyebrow mb-4">{{ t('errors.pageNotFound') }}</div>

      <!-- 404 Display -->
      <div class="mb-8">
        <div class="relative inline-block">
          <span
            style="
              font-family: 'Inter', system-ui, sans-serif;
              font-size: clamp(140px, 18vw, 200px);
              font-weight: 500;
              line-height: 1;
              letter-spacing: -0.05em;
              background: linear-gradient(
                180deg,
                color-mix(in srgb, var(--text-1) 20%, transparent),
                transparent 90%
              );
              -webkit-background-clip: text;
              background-clip: text;
              -webkit-text-fill-color: transparent;
            "
            >404</span
          >
        </div>
      </div>

      <!-- Text Content -->
      <div class="mb-10">
        <h1 class="h-2 mb-3">
          {{ t('errors.pageNotFound') }}
        </h1>
        <p class="lead" style="font-size: 14px; max-width: 360px; margin: 0 auto">
          The page you are looking for doesn't exist or has been moved.
        </p>
      </div>

      <!-- Action Buttons -->
      <div class="flex flex-col justify-center gap-3 sm:flex-row">
        <button @click="goBack" class="btn btn-secondary">
          <Icon name="arrowLeft" size="md" />
          Go Back
        </button>
        <router-link to="/dashboard" class="btn btn-primary">
          <Icon name="home" size="md" />
          Go to Dashboard
        </router-link>
      </div>

      <!-- Help Link -->
      <p class="mt-10 mono" style="font-size: 12px; color: var(--text-4)">
        Need help?
        <a
          href="#"
          style="
            color: var(--accent);
            margin-left: 4px;
            transition: color 150ms ease;
          "
          onmouseover="this.style.color='var(--accent-hover)'"
          onmouseout="this.style.color='var(--accent)'"
        >
          Contact support →
        </a>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const router = useRouter()

function goBack(): void {
  router.back()
}
</script>
