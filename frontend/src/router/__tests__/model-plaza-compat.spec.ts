import { describe, expect, it, vi } from 'vitest'

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isAuthenticated: true,
    isAdmin: false,
    isSimpleMode: false,
    checkAuth: vi.fn()
  })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    siteName: 'Test Site',
    cachedPublicSettings: {},
    publicSettingsLoaded: true,
    backendModeEnabled: false,
    initializePublicSettings: vi.fn()
  })
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    opsMonitoringEnabled: false,
    customMenuItems: [],
    fetch: vi.fn()
  })
}))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    startNavigation: vi.fn(),
    endNavigation: vi.fn()
  })
}))

vi.mock('@/composables/useRoutePrefetch', () => ({
  useRoutePrefetch: () => ({
    triggerPrefetch: vi.fn()
  })
}))

describe('legacy model plaza routes', () => {
  it.each(['/models', '/model-hub', '/model-plaza'])(
    'redirects %s to available channels',
    async (path) => {
      const router = (await import('../index')).default
      const route = router.getRoutes().find((item) => item.path === path)

      expect(route).toBeTruthy()
      expect(route?.redirect).toBe('/available-channels')
    }
  )
})
