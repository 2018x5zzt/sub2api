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

describe('model hub route', () => {
  it('serves the model hub at /models instead of redirecting to available channels', async () => {
    const router = (await import('../index')).default
    const route = router.getRoutes().find((item) => item.path === '/models')

    expect(route).toBeTruthy()
    expect(route?.redirect).toBeUndefined()
    expect(route?.name).toBe('ModelHub')
    expect(route?.meta.titleKey).toBe('modelHub.title')
    expect(route?.meta.descriptionKey).toBe('modelHub.description')
  })
})
