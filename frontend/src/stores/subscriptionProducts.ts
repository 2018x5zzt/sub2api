/**
 * Subscription Product Store
 * Global state for user-facing shared subscription products.
 */

import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import subscriptionProductsAPI from '@/api/subscriptionProducts'
import type { ActiveSubscriptionProduct } from '@/types'

const CACHE_TTL_MS = 60_000

let requestGeneration = 0

export const useSubscriptionProductStore = defineStore('subscriptionProducts', () => {
  const items = ref<ActiveSubscriptionProduct[]>([])
  const loading = ref(false)
  const loaded = ref(false)
  const lastFetchedAt = ref<number | null>(null)

  let activePromise: Promise<ActiveSubscriptionProduct[]> | null = null
  let pollerInterval: ReturnType<typeof setInterval> | null = null

  const hasActiveProducts = computed(() => items.value.length > 0)

  async function fetchActive(force = false): Promise<ActiveSubscriptionProduct[]> {
    const now = Date.now()

    if (
      !force &&
      loaded.value &&
      lastFetchedAt.value &&
      now - lastFetchedAt.value < CACHE_TTL_MS
    ) {
      return items.value
    }

    if (activePromise && !force) {
      return activePromise
    }

    const currentGeneration = ++requestGeneration
    loading.value = true

    const requestPromise = subscriptionProductsAPI
      .getActive()
      .then((data) => {
        if (currentGeneration === requestGeneration) {
          items.value = data
          loaded.value = true
          lastFetchedAt.value = Date.now()
        }
        return data
      })
      .catch((error) => {
        console.error('Failed to fetch subscription products:', error)
        throw error
      })
      .finally(() => {
        if (activePromise === requestPromise) {
          loading.value = false
          activePromise = null
        }
      })

    activePromise = requestPromise
    return activePromise
  }

  function startPolling() {
    if (pollerInterval) return
    pollerInterval = setInterval(() => {
      fetchActive(true).catch((error) => {
        console.error('Subscription product polling failed:', error)
      })
    }, 5 * 60 * 1000)
  }

  function stopPolling() {
    if (pollerInterval) {
      clearInterval(pollerInterval)
      pollerInterval = null
    }
  }

  function clear() {
    requestGeneration++
    activePromise = null
    items.value = []
    loaded.value = false
    lastFetchedAt.value = null
    stopPolling()
  }

  function invalidateCache() {
    lastFetchedAt.value = null
  }

  return {
    items,
    loading,
    hasActiveProducts,
    fetchActive,
    startPolling,
    stopPolling,
    clear,
    invalidateCache
  }
})
