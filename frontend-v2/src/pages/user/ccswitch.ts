import type { GroupPlatform } from '@/types'

export type CcsClientType = 'claude' | 'gemini'

export interface BuildCcsImportUrlOptions {
  apiKey: string
  platform?: GroupPlatform | null
  clientType?: CcsClientType
  baseUrl?: string | null
  siteName?: string | null
  fallbackOrigin?: string
}

export function normalizeApiBaseUrl(baseUrl?: string | null, fallbackOrigin = window.location.origin) {
  const raw = (baseUrl || fallbackOrigin || '').trim()
  return raw.replace(/\/+$/, '')
}

function providerName(siteName?: string | null) {
  const trimmed = (siteName || '').trim()
  return trimmed || 'sub2api'
}

export function getCcsPlatformDefaultClient(platform?: GroupPlatform | null): CcsClientType {
  return platform === 'gemini' ? 'gemini' : 'claude'
}

function resolveCcsTarget(platform: GroupPlatform | null | undefined, clientType: CcsClientType | undefined, baseUrl: string) {
  if (platform === 'antigravity') {
    return {
      app: clientType === 'gemini' ? 'gemini' : 'claude',
      endpoint: `${baseUrl}/antigravity`
    }
  }

  switch (platform) {
    case 'openai':
      return { app: 'codex', endpoint: baseUrl }
    case 'gemini':
      return { app: 'gemini', endpoint: baseUrl }
    default:
      return { app: 'claude', endpoint: baseUrl }
  }
}

export function buildCcsUsageScript() {
  return `({
    request: {
      url: "{{baseUrl}}/v1/usage",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`
}

function encodeBase64(value: string) {
  if (typeof btoa === 'function') return btoa(value)
  return Buffer.from(value, 'utf8').toString('base64')
}

export function buildCcsImportUrl(options: BuildCcsImportUrlOptions) {
  const baseUrl = normalizeApiBaseUrl(options.baseUrl, options.fallbackOrigin)
  const target = resolveCcsTarget(options.platform, options.clientType, baseUrl)
  const params = new URLSearchParams({
    resource: 'provider',
    app: target.app,
    name: providerName(options.siteName),
    homepage: baseUrl,
    endpoint: target.endpoint,
    apiKey: options.apiKey,
    configFormat: 'json',
    usageEnabled: 'true',
    usageScript: encodeBase64(buildCcsUsageScript()),
    usageAutoInterval: '30'
  })

  return `ccswitch://v1/import?${params.toString()}`
}
