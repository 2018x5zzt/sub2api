import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import SettingsView from '../SettingsView.vue'

const {
  getSettings,
  updateSettings,
  getAdminApiKey,
  getOverloadCooldownSettings,
  getStreamTimeoutSettings,
  getRectifierSettings,
  getBetaPolicySettings,
  getAllGroups,
  listProducts,
  showError,
  showSuccess,
  fetchPublicSettings,
  fetchAdminSettings
} = vi.hoisted(() => ({
  getSettings: vi.fn(),
  updateSettings: vi.fn(),
  getAdminApiKey: vi.fn(),
  getOverloadCooldownSettings: vi.fn(),
  getStreamTimeoutSettings: vi.fn(),
  getRectifierSettings: vi.fn(),
  getBetaPolicySettings: vi.fn(),
  getAllGroups: vi.fn(),
  listProducts: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  fetchPublicSettings: vi.fn(),
  fetchAdminSettings: vi.fn()
}))

vi.mock('@/api', () => ({
  adminAPI: {
    settings: {
      getSettings,
      updateSettings,
      getAdminApiKey,
      getOverloadCooldownSettings,
      getStreamTimeoutSettings,
      getRectifierSettings,
      getBetaPolicySettings
    },
    groups: {
      getAll: getAllGroups
    },
    subscriptionProducts: {
      listProducts
    }
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
    fetchPublicSettings
  })
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    fetch: fetchAdminSettings
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, values?: Record<string, unknown>) => {
        if (values?.groupId) {
          return `${key}:${String(values.groupId)}`
        }
        return key
      }
    })
  }
})

const createSystemSettings = () =>
  ({
    registration_enabled: true,
    email_verify_enabled: false,
    registration_email_suffix_whitelist: [],
    promo_code_enabled: false,
    password_reset_enabled: false,
    frontend_url: '',
    totp_enabled: false,
    totp_encryption_key_configured: false,
    default_balance: 0,
    default_concurrency: 1,
    default_subscriptions: [],
    default_subscription_products: [],
    enterprise_visible_groups: [
      {
        enterprise_name: 'acme',
        visible_group_ids: [2, 9]
      }
    ],
    site_name: 'Sub2API',
    site_logo: '',
    site_subtitle: 'Subscription to API Conversion Platform',
    api_base_url: '',
    contact_info: '',
    doc_url: '',
    home_content: '',
    hide_ccs_import_button: false,
    purchase_subscription_enabled: false,
    purchase_subscription_url: '',
    sora_client_enabled: false,
    backend_mode_enabled: false,
    custom_menu_items: [],
    custom_endpoints: [],
    smtp_host: '',
    smtp_port: 587,
    smtp_username: '',
    smtp_password_configured: false,
    smtp_from_email: '',
    smtp_from_name: '',
    smtp_use_tls: true,
    turnstile_enabled: false,
    turnstile_site_key: '',
    turnstile_secret_key_configured: false,
    linuxdo_connect_enabled: false,
    linuxdo_connect_client_id: '',
    linuxdo_connect_client_secret_configured: false,
    linuxdo_connect_redirect_url: '',
    enable_model_fallback: false,
    fallback_model_anthropic: 'claude-3-5-sonnet-20241022',
    fallback_model_openai: 'gpt-4o',
    fallback_model_gemini: 'gemini-2.5-pro',
    fallback_model_antigravity: 'gemini-2.5-pro',
    enable_identity_patch: true,
    identity_patch_prompt: '',
    ops_monitoring_enabled: true,
    ops_realtime_monitoring_enabled: true,
    ops_query_mode_default: 'auto',
    ops_metrics_interval_seconds: 60,
    min_claude_code_version: '',
    max_claude_code_version: '',
    allow_ungrouped_key_scheduling: false,
    enable_fingerprint_unification: true,
    enable_metadata_passthrough: false
  }) as any

describe('admin SettingsView enterprise visible groups', () => {
  beforeEach(() => {
    getSettings.mockReset()
    updateSettings.mockReset()
    getAdminApiKey.mockReset()
    getOverloadCooldownSettings.mockReset()
    getStreamTimeoutSettings.mockReset()
    getRectifierSettings.mockReset()
    getBetaPolicySettings.mockReset()
    getAllGroups.mockReset()
    listProducts.mockReset()
    showError.mockReset()
    showSuccess.mockReset()
    fetchPublicSettings.mockReset()
    fetchAdminSettings.mockReset()

    getSettings.mockResolvedValue(createSystemSettings())
    updateSettings.mockImplementation(async (payload: any) => ({
      ...createSystemSettings(),
      ...payload
    }))
    getAdminApiKey.mockResolvedValue({ exists: false, masked_key: '' })
    getOverloadCooldownSettings.mockResolvedValue({ enabled: true, cooldown_minutes: 10 })
    getStreamTimeoutSettings.mockResolvedValue({
      enabled: true,
      action: 'temp_unsched',
      temp_unsched_minutes: 5,
      threshold_count: 3,
      threshold_window_minutes: 10
    })
    getRectifierSettings.mockResolvedValue({
      enabled: true,
      thinking_signature_enabled: true,
      thinking_budget_enabled: true,
      apikey_signature_enabled: false,
      apikey_signature_patterns: []
    })
    getBetaPolicySettings.mockResolvedValue({ rules: [] })
    getAllGroups.mockResolvedValue([
      {
        id: 2,
        name: 'openai-team',
        description: 'OpenAI team group',
        platform: 'openai',
        subscription_type: 'standard',
        status: 'active',
        rate_multiplier: 1,
        sort_order: 1
      },
      {
        id: 5,
        name: 'gemini-team',
        description: 'Gemini team group',
        platform: 'gemini',
        subscription_type: 'standard',
        status: 'active',
        rate_multiplier: 1,
        sort_order: 2
      },
      {
        id: 9,
        name: 'anthropic-team',
        description: 'Anthropic team group',
        platform: 'anthropic',
        subscription_type: 'subscription',
        status: 'active',
        rate_multiplier: 1,
        sort_order: 3
      }
    ])
    listProducts.mockResolvedValue([
      {
        id: 101,
        code: 'gpt_monthly',
        name: 'GPT 月卡',
        description: 'GPT monthly shared subscription',
        status: 'active',
        default_validity_days: 30,
        daily_limit_usd: 0,
        weekly_limit_usd: 0,
        monthly_limit_usd: 100,
        sort_order: 10,
        created_at: '2026-04-25T00:00:00Z',
        updated_at: '2026-04-25T00:00:00Z'
      }
    ])
    fetchPublicSettings.mockResolvedValue(undefined)
    fetchAdminSettings.mockResolvedValue(undefined)
  })

  it('loads and saves normalized enterprise visible group rules', async () => {
    const wrapper = mount(SettingsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          Select: true,
          GroupBadge: true,
          GroupOptionItem: true,
          Toggle: true,
          ImageUpload: true,
          BackupSettings: true,
          DataManagementSettings: true,
          GroupSelector: true,
          Teleport: true
        }
      }
    })

    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    expect(setupState.form.enterprise_visible_groups).toEqual([
      {
        enterprise_name: 'acme',
        visible_group_ids: [2, 9]
      }
    ])

    setupState.form.enterprise_visible_groups = [
      {
        enterprise_name: '  Acme ',
        visible_group_ids: [9, 2, 9]
      },
      {
        enterprise_name: 'Bustest',
        visible_group_ids: [5, 5]
      },
      {
        enterprise_name: '   ',
        visible_group_ids: [2]
      }
    ]

    await setupState.saveSettings()

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        enterprise_visible_groups: [
          {
            enterprise_name: 'acme',
            visible_group_ids: [2, 9]
          },
          {
            enterprise_name: 'bustest',
            visible_group_ids: [5]
          }
        ]
      })
    )
    expect(showSuccess).toHaveBeenCalledWith('admin.settings.settingsSaved')
  })

  it('persists default product subscriptions from settings', async () => {
    getSettings.mockResolvedValue({
      ...createSystemSettings(),
      default_subscription_products: [{ product_id: 101, validity_days: 30 }]
    })

    const wrapper = mount(SettingsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          Icon: true,
          Select: true,
          GroupBadge: true,
          GroupOptionItem: true,
          Toggle: true,
          ImageUpload: true,
          BackupSettings: true,
          DataManagementSettings: true,
          GroupSelector: true,
          Teleport: true
        }
      }
    })

    await flushPromises()

    const setupState = (wrapper.vm as any).$?.setupState
    expect(setupState.form.default_subscription_products).toEqual([
      { product_id: 101, validity_days: 30 }
    ])

    setupState.form.default_subscription_products = [
      { product_id: 101, validity_days: 30 },
      { product_id: 0, validity_days: 30 },
      { product_id: 101, validity_days: 7 }
    ]

    await setupState.saveSettings()

    expect(updateSettings).toHaveBeenCalledWith(
      expect.objectContaining({
        default_subscription_products: [{ product_id: 101, validity_days: 30 }]
      })
    )
  })
})
