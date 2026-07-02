<template>
  <div class="xlab-oauth-consent">
    <div class="xlab-oauth-consent__panel">
      <div v-if="!errorMessage" class="xlab-oauth-consent__spinner" aria-hidden="true"></div>
      <h1>{{ errorMessage ? 'Xlab 授权失败' : '正在连接 Miku' }}</h1>
      <p>{{ errorMessage || '正在使用你的 XlabAPI 登录状态完成授权。' }}</p>
      <RouterLink v-if="errorMessage" to="/dashboard" class="btn btn-secondary btn-sm">
        返回控制台
      </RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { xlabOAuthAPI } from '@/api'

const route = useRoute()
const errorMessage = ref('')

function queryValue(key: string): string {
  const value = route.query[key]
  if (Array.isArray(value)) {
    return value[0] || ''
  }
  return typeof value === 'string' ? value : ''
}

onMounted(async () => {
  try {
    const result = await xlabOAuthAPI.authorize({
      client_id: queryValue('client_id'),
      redirect_uri: queryValue('redirect_uri'),
      response_type: queryValue('response_type'),
      scope: queryValue('scope'),
      state: queryValue('state'),
    })
    window.location.replace(result.redirect_uri)
  } catch (error: unknown) {
    const err = error as { message?: string }
    errorMessage.value = err.message || '无法完成 Xlab 授权'
  }
})
</script>

<style scoped>
.xlab-oauth-consent {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background: #f7f9fc;
}

.xlab-oauth-consent__panel {
  width: min(420px, 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
  padding: 32px;
  text-align: center;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.08);
}

.xlab-oauth-consent__panel h1 {
  margin: 0;
  color: #111827;
  font-size: 20px;
  font-weight: 700;
}

.xlab-oauth-consent__panel p {
  margin: 0;
  color: #4b5563;
  font-size: 14px;
  line-height: 1.6;
}

.xlab-oauth-consent__spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #dbeafe;
  border-top-color: #2563eb;
  border-radius: 999px;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

:global(.dark) .xlab-oauth-consent {
  background: #0f172a;
}

:global(.dark) .xlab-oauth-consent__panel {
  border-color: #334155;
  background: #111827;
  box-shadow: none;
}

:global(.dark) .xlab-oauth-consent__panel h1 {
  color: #f9fafb;
}

:global(.dark) .xlab-oauth-consent__panel p {
  color: #cbd5e1;
}
</style>
