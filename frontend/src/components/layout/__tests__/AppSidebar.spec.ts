import { describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'
import AppSidebar from '../AppSidebar.vue'

const mocks = vi.hoisted(() => ({
  route: { path: '/dashboard' },
  authStore: {
    isAuthenticated: true,
    isAdmin: false,
    isSimpleMode: false
  },
  appStore: {
    sidebarCollapsed: false,
    mobileOpen: false,
    siteName: 'Sub2API',
    siteLogo: '',
    siteVersion: 'test',
    publicSettingsLoaded: true,
    cachedPublicSettings: {
      available_channels_enabled: true,
      affiliate_enabled: false,
      sora_client_enabled: false,
      purchase_subscription_enabled: false,
      custom_menu_items: []
    },
    backendModeEnabled: false,
    toggleSidebar: vi.fn(),
    setMobileOpen: vi.fn()
  },
  adminSettingsStore: {
    opsMonitoringEnabled: false,
    customMenuItems: [],
    fetch: vi.fn()
  },
  onboardingStore: {
    isCurrentStep: vi.fn(() => false),
    nextStep: vi.fn()
  }
}))

vi.mock('vue-router', () => ({
  useRoute: () => mocks.route
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      locale: { value: 'en' },
      setLocaleMessage: vi.fn()
    }
  }),
  useI18n: () => ({
    t: (key: string, fallback?: string) => {
      const labels: Record<string, string> = {
        'nav.dashboard': 'Dashboard',
        'nav.apiKeys': 'API Keys',
        'nav.modelHub': 'Model Hub',
        'nav.availableChannels': 'Available Channels',
        'nav.usage': 'Usage',
        'nav.mySubscriptions': 'My Subscriptions',
        'nav.redeem': 'Redeem',
        'nav.profile': 'Profile',
        'nav.lightMode': 'Light Mode',
        'nav.darkMode': 'Dark Mode',
        'nav.collapse': 'Collapse',
        'nav.expand': 'Expand'
      }
      return labels[key] ?? fallback ?? key
    }
  })
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => mocks.authStore,
  useAppStore: () => mocks.appStore,
  useAdminSettingsStore: () => mocks.adminSettingsStore,
  useOnboardingStore: () => mocks.onboardingStore
}))

vi.mock('@/utils/sanitize', () => ({
  sanitizeSvg: (svg: string) => svg
}))

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(() => ({
    matches: false,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn()
  }))
})

describe('AppSidebar', () => {
  it('shows Model Hub as the user model entry and hides Available Channels from navigation', () => {
    const wrapper = mount(AppSidebar, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
          VersionBadge: true
        }
      }
    })

    expect(wrapper.text()).toContain('Model Hub')
    expect(wrapper.text()).not.toContain('Available Channels')
    expect(wrapper.findComponent(RouterLinkStub).exists()).toBe(true)
    expect(
      wrapper
        .findAllComponents(RouterLinkStub)
        .some((link) => link.props('to') === '/models')
    ).toBe(true)
    expect(
      wrapper
        .findAllComponents(RouterLinkStub)
        .some((link) => link.props('to') === '/available-channels')
    ).toBe(false)
  })
})
