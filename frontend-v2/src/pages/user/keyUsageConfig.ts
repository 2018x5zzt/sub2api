import type { GroupPlatform } from '@/types'

export type ClientTabId = 'claude' | 'gemini' | 'codex' | 'codex-ws' | 'opencode'
export type ShellTabId = 'unix' | 'cmd' | 'powershell' | 'windows'
export type Translate = (key: string) => string

export interface TabConfig {
  id: string
  label: string
}

export interface UsageFileConfig {
  path: string
  content: string
  hint?: string
}

export interface BuildUsageFilesOptions {
  platform?: GroupPlatform | null
  clientTab: ClientTabId
  shellTab: ShellTabId
  baseUrl: string
  apiKey: string
  t: Translate
}

export function normalizeApiRoot(baseUrl: string) {
  return (baseUrl || window.location.origin).replace(/\/+$/, '').replace(/\/v1\/?$/, '')
}

export function ensureApiV1Base(value: string) {
  const trimmed = value.replace(/\/+$/, '')
  return trimmed.endsWith('/v1') ? trimmed : `${trimmed}/v1`
}

function ensureGeminiV1BetaBase(value: string) {
  const trimmed = value.replace(/\/+$/, '')
  return trimmed.endsWith('/v1beta') ? trimmed : `${trimmed}/v1beta`
}

export function defaultClientTab(platform?: GroupPlatform | null): ClientTabId {
  switch (platform) {
    case 'openai':
      return 'codex'
    case 'gemini':
      return 'gemini'
    case 'antigravity':
      return 'claude'
    default:
      return 'claude'
  }
}

export function buildClientTabs(platform: GroupPlatform | null | undefined, allowMessagesDispatch: boolean | undefined, t: Translate): TabConfig[] {
  switch (platform) {
    case 'openai': {
      const tabs: TabConfig[] = [
        { id: 'codex', label: t('keys.useKeyModal.cliTabs.codexCli') },
        { id: 'codex-ws', label: t('keys.useKeyModal.cliTabs.codexCliWs') }
      ]
      if (allowMessagesDispatch) tabs.push({ id: 'claude', label: t('keys.useKeyModal.cliTabs.claudeCode') })
      tabs.push({ id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode') })
      return tabs
    }
    case 'gemini':
      return [
        { id: 'gemini', label: t('keys.useKeyModal.cliTabs.geminiCli') },
        { id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode') }
      ]
    case 'antigravity':
      return [
        { id: 'claude', label: t('keys.useKeyModal.cliTabs.claudeCode') },
        { id: 'gemini', label: t('keys.useKeyModal.cliTabs.geminiCli') },
        { id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode') }
      ]
    default:
      return [
        { id: 'claude', label: t('keys.useKeyModal.cliTabs.claudeCode') },
        { id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode') }
      ]
  }
}

export function buildShellTabs(clientTab: ClientTabId): TabConfig[] {
  if (clientTab === 'opencode') return []
  if (clientTab === 'codex' || clientTab === 'codex-ws') {
    return [
      { id: 'unix', label: 'macOS / Linux' },
      { id: 'windows', label: 'Windows' }
    ]
  }
  return [
    { id: 'unix', label: 'macOS / Linux' },
    { id: 'cmd', label: 'Windows CMD' },
    { id: 'powershell', label: 'PowerShell' }
  ]
}

export function platformDescription(platform: GroupPlatform | null | undefined, clientTab: ClientTabId, t: Translate) {
  if (platform === 'openai' && clientTab !== 'claude') return t('keys.useKeyModal.openai.description')
  if (platform === 'gemini') return t('keys.useKeyModal.gemini.description')
  if (platform === 'antigravity') return t('keys.useKeyModal.antigravity.description')
  return t('keys.useKeyModal.description')
}

export function platformNote(platform: GroupPlatform | null | undefined, clientTab: ClientTabId, shellTab: ShellTabId, t: Translate) {
  if (clientTab === 'opencode') return ''
  if (platform === 'openai' && clientTab !== 'claude') {
    return shellTab === 'windows' ? t('keys.useKeyModal.openai.noteWindows') : t('keys.useKeyModal.openai.note')
  }
  if (platform === 'gemini') return t('keys.useKeyModal.gemini.note')
  if (platform === 'antigravity') {
    return clientTab === 'gemini' ? t('keys.useKeyModal.antigravity.geminiNote') : t('keys.useKeyModal.antigravity.claudeNote')
  }
  return t('keys.useKeyModal.note')
}

function buildAnthropicFiles(baseUrl: string, apiKey: string, shellTab: ShellTabId): UsageFileConfig[] {
  const shellContent = shellTab === 'cmd'
    ? `set ANTHROPIC_BASE_URL=${baseUrl}\nset ANTHROPIC_AUTH_TOKEN=${apiKey}\nset CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`
    : shellTab === 'powershell'
      ? `$env:ANTHROPIC_BASE_URL="${baseUrl}"\n$env:ANTHROPIC_AUTH_TOKEN="${apiKey}"\n$env:CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`
      : `export ANTHROPIC_BASE_URL="${baseUrl}"\nexport ANTHROPIC_AUTH_TOKEN="${apiKey}"\nexport CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1`

  return [
    { path: shellTab === 'cmd' ? 'Command Prompt' : shellTab === 'powershell' ? 'PowerShell' : 'Terminal', content: shellContent },
    {
      path: shellTab === 'unix' ? '~/.claude/settings.json' : '%userprofile%\\.claude\\settings.json',
      hint: 'VSCode Claude Code',
      content: `{
  "env": {
    "ANTHROPIC_BASE_URL": "${baseUrl}",
    "ANTHROPIC_AUTH_TOKEN": "${apiKey}",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"
  }
}`
    }
  ]
}

function buildGeminiCliFile(baseUrl: string, apiKey: string, shellTab: ShellTabId, t: Translate): UsageFileConfig {
  const model = 'gemini-2.0-flash'
  const modelComment = t('keys.useKeyModal.gemini.modelComment')
  if (shellTab === 'cmd') {
    return {
      path: 'Command Prompt',
      content: `set GOOGLE_GEMINI_BASE_URL=${baseUrl}\nset GEMINI_API_KEY=${apiKey}\nset GEMINI_MODEL=${model}\nREM ${modelComment}`
    }
  }
  if (shellTab === 'powershell') {
    return {
      path: 'PowerShell',
      content: `$env:GOOGLE_GEMINI_BASE_URL="${baseUrl}"\n$env:GEMINI_API_KEY="${apiKey}"\n$env:GEMINI_MODEL="${model}"  # ${modelComment}`
    }
  }
  return {
    path: 'Terminal',
    content: `export GOOGLE_GEMINI_BASE_URL="${baseUrl}"\nexport GEMINI_API_KEY="${apiKey}"\nexport GEMINI_MODEL="${model}"  # ${modelComment}`
  }
}

function buildOpenAIFiles(baseUrl: string, apiKey: string, shellTab: ShellTabId, webSocket: boolean, t: Translate): UsageFileConfig[] {
  const configDir = shellTab === 'windows' ? '%userprofile%\\.codex' : '~/.codex'
  const wsLines = webSocket ? '\nsupports_websockets = true\nrequires_openai_auth = true\n\n[features]\nresponses_websockets_v2 = true' : '\nrequires_openai_auth = true'
  return [
    {
      path: `${configDir}/config.toml`,
      hint: t('keys.useKeyModal.openai.configTomlHint'),
      content: `model_provider = "OpenAI"
model = "gpt-5.4"
review_model = "gpt-5.4"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${baseUrl}"
wire_api = "responses"${wsLines}`
    },
    {
      path: `${configDir}/auth.json`,
      content: `{
  "OPENAI_API_KEY": "${apiKey}"
}`
    }
  ]
}

function buildOpenCodeConfig(platform: string, baseUrl: string, apiKey: string, t: Translate, pathLabel = 'opencode.json'): UsageFileConfig {
  const modelId = platform.includes('gemini') ? 'gemini-2.5-pro' : platform.includes('anthropic') ? 'claude-sonnet-4-5' : 'gpt-5.4'
  return {
    path: pathLabel,
    hint: t('keys.useKeyModal.opencode.hint'),
    content: JSON.stringify({
      provider: {
        [platform]: {
          options: {
            baseURL: baseUrl,
            apiKey
          }
        }
      },
      model: modelId
    }, null, 2)
  }
}

export function buildUsageFiles(options: BuildUsageFilesOptions): UsageFileConfig[] {
  const root = normalizeApiRoot(options.baseUrl)
  const directBase = (options.baseUrl || window.location.origin).replace(/\/+$/, '')
  const apiBase = ensureApiV1Base(root)
  const antigravityBase = ensureApiV1Base(`${root}/antigravity`)
  const geminiBase = ensureGeminiV1BetaBase(root)
  const antigravityGeminiBase = ensureGeminiV1BetaBase(`${root}/antigravity`)

  if (options.clientTab === 'opencode') {
    switch (options.platform) {
      case 'anthropic':
        return [buildOpenCodeConfig('anthropic', apiBase, options.apiKey, options.t)]
      case 'openai':
        return [buildOpenCodeConfig('openai', apiBase, options.apiKey, options.t)]
      case 'gemini':
        return [buildOpenCodeConfig('gemini', geminiBase, options.apiKey, options.t)]
      case 'antigravity':
        return [
          buildOpenCodeConfig('antigravity-claude', antigravityBase, options.apiKey, options.t, 'opencode.json (Claude)'),
          buildOpenCodeConfig('antigravity-gemini', antigravityGeminiBase, options.apiKey, options.t, 'opencode.json (Gemini)')
        ]
      default:
        return [buildOpenCodeConfig('anthropic', apiBase, options.apiKey, options.t)]
    }
  }

  switch (options.platform) {
    case 'openai':
      if (options.clientTab === 'claude') return buildAnthropicFiles(directBase, options.apiKey, options.shellTab)
      return buildOpenAIFiles(directBase, options.apiKey, options.shellTab, options.clientTab === 'codex-ws', options.t)
    case 'gemini':
      return [buildGeminiCliFile(directBase, options.apiKey, options.shellTab, options.t)]
    case 'antigravity':
      if (options.clientTab === 'gemini') return [buildGeminiCliFile(`${directBase}/antigravity`, options.apiKey, options.shellTab, options.t)]
      return buildAnthropicFiles(`${directBase}/antigravity`, options.apiKey, options.shellTab)
    default:
      return buildAnthropicFiles(directBase, options.apiKey, options.shellTab)
  }
}
