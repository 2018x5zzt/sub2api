import { describe, expect, it } from 'vitest'
import { buildClientTabs, buildUsageFiles, ensureApiV1Base, normalizeApiRoot } from '../keyUsageConfig'

const t = (key: string) => key

describe('key usage config helpers', () => {
  it('normalizes roots and API version suffixes', () => {
    expect(normalizeApiRoot('https://api.example.com/v1/')).toBe('https://api.example.com')
    expect(ensureApiV1Base('https://api.example.com')).toBe('https://api.example.com/v1')
    expect(ensureApiV1Base('https://api.example.com/v1')).toBe('https://api.example.com/v1')
  })

  it('builds OpenAI client tabs and only includes Claude Code when messages dispatch is allowed', () => {
    expect(buildClientTabs('openai', false, t).map((tab) => tab.id)).toEqual(['codex', 'codex-ws', 'opencode'])
    expect(buildClientTabs('openai', true, t).map((tab) => tab.id)).toEqual(['codex', 'codex-ws', 'claude', 'opencode'])
  })

  it('builds Codex config files for OpenAI keys', () => {
    const files = buildUsageFiles({
      platform: 'openai',
      clientTab: 'codex',
      shellTab: 'unix',
      baseUrl: 'https://api.example.com',
      apiKey: 'sk-openai-key',
      t
    })

    expect(files).toHaveLength(2)
    expect(files[0].path).toBe('~/.codex/config.toml')
    expect(files[0].content).toContain('wire_api = "responses"')
    expect(files[0].content).toContain('base_url = "https://api.example.com"')
    expect(files[1].path).toBe('~/.codex/auth.json')
    expect(files[1].content).toContain('sk-openai-key')
  })

  it('preserves configured v1 suffix for non-OpenCode client configs', () => {
    const files = buildUsageFiles({
      platform: 'openai',
      clientTab: 'codex',
      shellTab: 'unix',
      baseUrl: 'https://api.example.com/v1',
      apiKey: 'sk-openai-key',
      t
    })

    expect(files[0].content).toContain('base_url = "https://api.example.com/v1"')
  })

  it('builds Antigravity Gemini CLI files with antigravity base url', () => {
    const files = buildUsageFiles({
      platform: 'antigravity',
      clientTab: 'gemini',
      shellTab: 'unix',
      baseUrl: 'https://api.example.com',
      apiKey: 'sk-ag-key',
      t
    })

    expect(files).toHaveLength(1)
    expect(files[0].content).toContain('GOOGLE_GEMINI_BASE_URL="https://api.example.com/antigravity"')
    expect(files[0].content).toContain('GEMINI_API_KEY="sk-ag-key"')
  })
})
