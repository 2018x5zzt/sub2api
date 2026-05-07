<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden"
    style="background: var(--bg-0)"
  >
    <!-- Background Decorations -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- mesh gradient（cyan / violet glow） -->
      <div
        class="absolute inset-0 bg-mesh-gradient"
        style="opacity: 0.7"
      ></div>
      <!-- 网格 -->
      <div
        class="absolute inset-0 grid-bg"
        style="
          opacity: 0.4;
          mask-image: radial-gradient(
            ellipse 80% 60% at 50% 25%,
            #000 30%,
            transparent 80%
          );
          -webkit-mask-image: radial-gradient(
            ellipse 80% 60% at 50% 25%,
            #000 30%,
            transparent 80%
          );
        "
      ></div>
    </div>

    <!-- Header -->
    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div
            class="h-9 w-9 overflow-hidden flex items-center justify-center"
            style="
              border-radius: 8px;
              background: var(--bg-2);
              border: 1px solid var(--line-2);
            "
          >
            <img
              :src="siteLogo || '/logo.png'"
              alt="Logo"
              class="h-full w-full object-contain"
            />
          </div>
          <span
            style="
              font-family: 'Inter', system-ui, sans-serif;
              font-size: 15px;
              font-weight: 500;
              letter-spacing: -0.01em;
              color: var(--text-1);
            "
            >{{ siteName }}</span
          >
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-2">
          <!-- Language Switcher -->
          <LocaleSwitcher />

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-ghost btn-icon"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="btn btn-ghost btn-icon"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="btn btn-secondary btn-sm"
            style="padding-left: 4px; padding-right: 12px"
          >
            <span
              class="flex h-6 w-6 items-center justify-center"
              style="
                border-radius: 5px;
                background: linear-gradient(
                  135deg,
                  var(--accent),
                  var(--accent-2)
                );
                color: #052330;
                font-size: 10.5px;
                font-weight: 600;
              "
            >
              {{ userInitial }}
            </span>
            <span style="margin-left: 6px">{{ t('home.dashboard') }}</span>
            <svg
              class="h-3 w-3"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
              style="color: var(--text-3); margin-left: 6px"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link v-else to="/login" class="btn btn-primary btn-sm">
            {{ t('home.login') }} →
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1 px-6 py-12 md:py-16">
      <div class="mx-auto max-w-6xl">
        <!-- Hero Section -->
        <div
          class="mb-16 flex flex-col items-center justify-between gap-12 lg:flex-row lg:gap-16"
        >
          <!-- Left: Text Content -->
          <div class="flex-1 text-center lg:text-left">
            <div
              class="badge mb-6"
              style="
                background: color-mix(
                  in srgb,
                  var(--accent-3) 12%,
                  transparent
                );
                border-color: color-mix(
                  in srgb,
                  var(--accent-3) 32%,
                  transparent
                );
                color: var(--accent-3);
              "
            >
              <span
                style="
                  width: 6px;
                  height: 6px;
                  border-radius: 999px;
                  background: var(--accent-3);
                  box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-3) 24%, transparent);
                "
              ></span>
              <span class="mono" style="letter-spacing: 0.02em">{{
                t('home.tags.realtimeBilling')
              }}</span>
            </div>

            <h1 class="h-display mb-5">
              {{ siteName }}
            </h1>
            <p class="lead mb-8" style="max-width: 540px">
              {{ siteSubtitle }}
            </p>

            <!-- CTA -->
            <div class="flex items-center gap-3 justify-center lg:justify-start">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary btn-lg"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" :stroke-width="2" />
              </router-link>
              <a
                v-if="docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary btn-lg"
              >
                {{ t('home.docs') }}
              </a>
            </div>

            <p
              class="mono mt-6"
              style="font-size: 12px; color: var(--text-4)"
            >
              # 兼容 OpenAI / Anthropic / Gemini SDK
            </p>
          </div>

          <!-- Right: Terminal Animation -->
          <div class="flex flex-1 justify-center lg:justify-end">
            <div class="terminal-container">
              <div class="terminal-window">
                <!-- Window header -->
                <div class="terminal-header">
                  <div class="terminal-buttons">
                    <span class="btn-close"></span>
                    <span class="btn-minimize"></span>
                    <span class="btn-maximize"></span>
                  </div>
                  <span class="terminal-title">terminal</span>
                </div>
                <!-- Terminal content -->
                <div class="terminal-body">
                  <div class="code-line line-1">
                    <span class="code-prompt">$</span>
                    <span class="code-cmd">curl</span>
                    <span class="code-flag">-X POST</span>
                    <span class="code-url">/v1/messages</span>
                  </div>
                  <div class="code-line line-2">
                    <span class="code-comment"># Routing to upstream...</span>
                  </div>
                  <div class="code-line line-3">
                    <span class="code-success">200 OK</span>
                    <span class="code-response">{ "content": "Hello!" }</span>
                  </div>
                  <div class="code-line line-4">
                    <span class="code-prompt">$</span>
                    <span class="cursor"></span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Feature Tags -->
        <div class="mb-16 flex flex-wrap items-center justify-center gap-3">
          <div class="badge" style="height: 30px; padding: 0 14px">
            <Icon
              name="swap"
              size="sm"
              style="color: var(--accent)"
            />
            <span style="color: var(--text-2)">{{
              t('home.tags.subscriptionToApi')
            }}</span>
          </div>
          <div class="badge" style="height: 30px; padding: 0 14px">
            <Icon
              name="shield"
              size="sm"
              style="color: var(--accent)"
            />
            <span style="color: var(--text-2)">{{
              t('home.tags.stickySession')
            }}</span>
          </div>
          <div class="badge" style="height: 30px; padding: 0 14px">
            <Icon
              name="chart"
              size="sm"
              style="color: var(--accent)"
            />
            <span style="color: var(--text-2)">{{
              t('home.tags.realtimeBilling')
            }}</span>
          </div>
        </div>

        <!-- Features Grid -->
        <div class="mb-16 grid gap-4 md:grid-cols-3">
          <!-- Feature 1: Unified Gateway -->
          <div class="card card-hover" style="padding: 22px">
            <div
              class="mb-4 flex h-11 w-11 items-center justify-center"
              style="
                border-radius: 10px;
                background: color-mix(in srgb, var(--accent) 14%, transparent);
                color: var(--accent);
              "
            >
              <Icon name="server" size="lg" />
            </div>
            <h3 class="h-3 mb-2">
              {{ t('home.features.unifiedGateway') }}
            </h3>
            <p style="font-size: 13.5px; line-height: 1.6; color: var(--text-3)">
              {{ t('home.features.unifiedGatewayDesc') }}
            </p>
          </div>

          <!-- Feature 2: Account Pool -->
          <div class="card card-hover" style="padding: 22px">
            <div
              class="mb-4 flex h-11 w-11 items-center justify-center"
              style="
                border-radius: 10px;
                background: color-mix(in srgb, var(--accent-2) 14%, transparent);
                color: var(--accent-2);
              "
            >
              <svg
                class="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
            <h3 class="h-3 mb-2">
              {{ t('home.features.multiAccount') }}
            </h3>
            <p style="font-size: 13.5px; line-height: 1.6; color: var(--text-3)">
              {{ t('home.features.multiAccountDesc') }}
            </p>
          </div>

          <!-- Feature 3: Billing & Quota -->
          <div class="card card-hover" style="padding: 22px">
            <div
              class="mb-4 flex h-11 w-11 items-center justify-center"
              style="
                border-radius: 10px;
                background: color-mix(in srgb, var(--accent-3) 14%, transparent);
                color: var(--accent-3);
              "
            >
              <svg
                class="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                />
              </svg>
            </div>
            <h3 class="h-3 mb-2">
              {{ t('home.features.balanceQuota') }}
            </h3>
            <p style="font-size: 13.5px; line-height: 1.6; color: var(--text-3)">
              {{ t('home.features.balanceQuotaDesc') }}
            </p>
          </div>
        </div>

        <!-- Supported Providers -->
        <div class="mb-8 text-center">
          <div class="eyebrow mb-3">{{ t('home.providers.title') }}</div>
          <p style="font-size: 13.5px; color: var(--text-3)">
            {{ t('home.providers.description') }}
          </p>
        </div>

        <div class="mb-12 flex flex-wrap items-center justify-center gap-3">
          <!-- Claude -->
          <div class="provider-tag">
            <div
              class="provider-icon"
              style="
                background: linear-gradient(135deg, #d97757, #c4602e);
              "
            >
              <span>C</span>
            </div>
            <span class="provider-name">{{ t('home.providers.claude') }}</span>
            <span class="provider-status">{{
              t('home.providers.supported')
            }}</span>
          </div>
          <!-- GPT -->
          <div class="provider-tag">
            <div
              class="provider-icon"
              style="background: linear-gradient(135deg, #10a37f, #0d8f6f)"
            >
              <span>G</span>
            </div>
            <span class="provider-name">GPT</span>
            <span class="provider-status">{{
              t('home.providers.supported')
            }}</span>
          </div>
          <!-- Gemini -->
          <div class="provider-tag">
            <div
              class="provider-icon"
              style="background: linear-gradient(135deg, #4285f4, #1a73e8)"
            >
              <span>G</span>
            </div>
            <span class="provider-name">{{ t('home.providers.gemini') }}</span>
            <span class="provider-status">{{
              t('home.providers.supported')
            }}</span>
          </div>
          <!-- Antigravity -->
          <div class="provider-tag">
            <div
              class="provider-icon"
              style="background: linear-gradient(135deg, #f43f5e, #db2777)"
            >
              <span>A</span>
            </div>
            <span class="provider-name">{{
              t('home.providers.antigravity')
            }}</span>
            <span class="provider-status">{{
              t('home.providers.supported')
            }}</span>
          </div>
          <!-- More -->
          <div class="provider-tag" style="opacity: 0.5">
            <div
              class="provider-icon"
              style="background: var(--bg-3); color: var(--text-3)"
            >
              <span>+</span>
            </div>
            <span class="provider-name">{{ t('home.providers.more') }}</span>
            <span
              class="provider-status"
              style="background: var(--bg-3); color: var(--text-3)"
              >{{ t('home.providers.soon') }}</span
            >
          </div>
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer
      class="relative z-10 px-6 py-8"
      style="border-top: 1px solid var(--line-1)"
    >
      <div
        class="mx-auto flex max-w-6xl flex-col items-center justify-center gap-4 text-center sm:flex-row sm:text-left"
      >
        <p
          class="mono"
          style="font-size: 12px; color: var(--text-4)"
        >
          &copy; {{ currentYear }} {{ siteName }}.
          {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-4">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="footer-link"
          >
            {{ t('home.docs') }}
          </a>
          <a
            :href="githubUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="footer-link"
          >
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))

const githubUrl = 'https://github.com/Wei-Shaw/sub2api'

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* Provider tag */
.provider-tag {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  background: var(--bg-1);
  border: 1px solid var(--line-1);
  border-radius: 10px;
  transition: all 150ms ease;
}
.provider-tag:hover {
  border-color: var(--line-2);
  background: var(--bg-2);
}
.provider-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  color: #ffffff;
}
.provider-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-1);
  letter-spacing: -0.005em;
}
.provider-status {
  font-family: 'JetBrains Mono', monospace;
  font-size: 9.5px;
  font-weight: 500;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 2px 6px;
  border-radius: 4px;
  background: color-mix(in srgb, var(--accent) 14%, transparent);
  color: var(--accent);
}

/* Footer link */
.footer-link {
  font-size: 13px;
  color: var(--text-3);
  transition: color 150ms ease;
}
.footer-link:hover {
  color: var(--text-1);
}

/* Terminal Container */
.terminal-container {
  position: relative;
  display: inline-block;
}

/* Terminal Window */
.terminal-window {
  width: 420px;
  background: linear-gradient(180deg, #11141a 0%, #07080a 100%);
  border-radius: 14px;
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.55),
    0 0 0 1px rgba(255, 255, 255, 0.06),
    inset 0 1px 0 rgba(255, 255, 255, 0.05);
  overflow: hidden;
  transform: perspective(1000px) rotateX(2deg) rotateY(-2deg);
  transition: transform 0.3s ease;
}

.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-4px);
  box-shadow:
    0 35px 70px -12px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(125, 211, 252, 0.18),
    0 0 60px rgba(125, 211, 252, 0.10),
    inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

/* Terminal Header */
.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: rgba(15, 18, 24, 0.6);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 11px;
  height: 11px;
  border-radius: 50%;
}

.btn-close {
  background: #ff5f56;
}
.btn-minimize {
  background: #ffbd2e;
}
.btn-maximize {
  background: #27c93f;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 11.5px;
  font-family: 'JetBrains Mono', ui-monospace, monospace;
  color: var(--text-4);
  margin-right: 52px;
}

/* Terminal Body */
.terminal-body {
  padding: 20px 24px;
  font-family: 'JetBrains Mono', ui-monospace, 'Fira Code', monospace;
  font-size: 13px;
  line-height: 2;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.3s;
}
.line-2 {
  animation-delay: 1s;
}
.line-3 {
  animation-delay: 1.8s;
}
.line-4 {
  animation-delay: 2.5s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #34d399;
  font-weight: 600;
}
.code-cmd {
  color: #7dd3fc;
}
.code-flag {
  color: #a78bfa;
}
.code-url {
  color: #c8cbd2;
}
.code-comment {
  color: #5b606b;
  font-style: italic;
}
.code-success {
  color: #34d399;
  background: rgba(52, 211, 153, 0.12);
  padding: 1px 7px;
  border-radius: 4px;
  font-weight: 600;
  font-size: 12px;
}
.code-response {
  color: #fbbf24;
}

/* Blinking Cursor */
.cursor {
  display: inline-block;
  width: 7px;
  height: 14px;
  background: #34d399;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}
</style>
