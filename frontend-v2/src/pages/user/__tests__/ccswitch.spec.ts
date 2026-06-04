import { describe, expect, it } from 'vitest'
import { buildCcsImportUrl, getCcsPlatformDefaultClient, normalizeApiBaseUrl } from '../ccswitch'

describe('ccswitch helpers', () => {
  it('normalizes public api base url with window origin fallback', () => {
    expect(normalizeApiBaseUrl('https://api.example.com/')).toBe('https://api.example.com')
    expect(normalizeApiBaseUrl('', 'https://fallback.example.com')).toBe('https://fallback.example.com')
  })

  it('builds an OpenAI Codex provider deeplink', () => {
    const url = buildCcsImportUrl({
      apiKey: 'sk-openai-key',
      platform: 'openai',
      baseUrl: 'https://api.example.com/',
      siteName: 'XlabAPI'
    })

    expect(url.startsWith('ccswitch://v1/import?')).toBe(true)
    const params = new URLSearchParams(url.replace('ccswitch://v1/import?', ''))
    expect(params.get('resource')).toBe('provider')
    expect(params.get('app')).toBe('codex')
    expect(params.get('name')).toBe('XlabAPI')
    expect(params.get('homepage')).toBe('https://api.example.com')
    expect(params.get('endpoint')).toBe('https://api.example.com')
    expect(params.get('apiKey')).toBe('sk-openai-key')
    expect(params.get('configFormat')).toBe('json')
    expect(params.get('usageEnabled')).toBe('true')
    expect(params.get('usageAutoInterval')).toBe('30')
    expect(params.get('usageScript')).toBeTruthy()
  })

  it('builds an Antigravity Gemini deeplink with antigravity endpoint suffix', () => {
    const url = buildCcsImportUrl({
      apiKey: 'sk-antigravity-key',
      platform: 'antigravity',
      clientType: 'gemini',
      baseUrl: 'https://api.example.com',
      siteName: '  '
    })

    const params = new URLSearchParams(url.replace('ccswitch://v1/import?', ''))
    expect(params.get('app')).toBe('gemini')
    expect(params.get('name')).toBe('sub2api')
    expect(params.get('homepage')).toBe('https://api.example.com')
    expect(params.get('endpoint')).toBe('https://api.example.com/antigravity')
    expect(params.get('apiKey')).toBe('sk-antigravity-key')
  })

  it('resolves default direct-import clients by platform', () => {
    expect(getCcsPlatformDefaultClient('openai')).toBe('claude')
    expect(getCcsPlatformDefaultClient('gemini')).toBe('gemini')
    expect(getCcsPlatformDefaultClient('anthropic')).toBe('claude')
    expect(getCcsPlatformDefaultClient(null)).toBe('claude')
  })
})
