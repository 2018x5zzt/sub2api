# API Key Use Key and CC-Switch Import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the legacy `Use Key` and `Import to CCS` user API key actions in `frontend-v2` while keeping the new React table/modal style.

**Architecture:** Keep behavior in small, testable units: pure helpers generate client configuration and CC-Switch deeplinks, focused React modals render those helpers, and `Keys.tsx` only wires row actions and public settings. Existing API key CRUD, usage stats, and backend APIs remain unchanged.

**Tech Stack:** React 18, TypeScript, TanStack Query, Zustand auth store, existing `frontend-v2` UI primitives, Vitest, Testing Library.

---

## File Structure

- Create `frontend-v2/src/pages/user/ccswitch.ts`
  - Pure functions for resolving public settings fallbacks, platform/client mapping, usage script encoding, and `ccswitch://v1/import` URL generation.
- Create `frontend-v2/src/pages/user/keyUsageConfig.ts`
  - Pure functions for generating tabs, notes, and code-file configs for Claude Code, Codex CLI, Gemini CLI, and OpenCode.
- Create `frontend-v2/src/pages/user/CcsClientSelectModal.tsx`
  - Small React modal for Antigravity Claude/Gemini import selection.
- Create `frontend-v2/src/pages/user/UseKeyModal.tsx`
  - React modal for platform-specific key usage instructions and copy buttons.
- Modify `frontend-v2/src/pages/user/Keys.tsx`
  - Add row action buttons, modal state, public settings reads, and CCS import handler.
- Modify `frontend-v2/src/pages/user/__tests__/Keys.spec.tsx`
  - Add page-level tests for visible actions, hide flag, deeplink behavior, Antigravity client selection, and Use Key modal content.
- Create `frontend-v2/src/pages/user/__tests__/ccswitch.spec.ts`
  - Unit tests for platform/client deeplink generation.
- Create `frontend-v2/src/pages/user/__tests__/keyUsageConfig.spec.ts`
  - Unit tests for config generation branches.

Do not modify backend files or legacy `frontend` files. The repository currently has unrelated backend changes in the working tree; leave them untouched.

---

### Task 1: Add failing CC-Switch helper tests

**Files:**
- Create: `frontend-v2/src/pages/user/__tests__/ccswitch.spec.ts`
- Later implementation target: `frontend-v2/src/pages/user/ccswitch.ts`

- [ ] **Step 1: Write the failing tests**

Create `frontend-v2/src/pages/user/__tests__/ccswitch.spec.ts`:

```ts
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
```

- [ ] **Step 2: Run the tests to verify RED**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/ccswitch.spec.ts
```

Expected: FAIL because `../ccswitch` does not exist.

---

### Task 2: Implement CC-Switch helper

**Files:**
- Create: `frontend-v2/src/pages/user/ccswitch.ts`
- Test: `frontend-v2/src/pages/user/__tests__/ccswitch.spec.ts`

- [ ] **Step 1: Write minimal implementation**

Create `frontend-v2/src/pages/user/ccswitch.ts`:

```ts
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
```

- [ ] **Step 2: Run tests to verify GREEN**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/ccswitch.spec.ts
```

Expected: PASS.

---

### Task 3: Add failing key usage config tests

**Files:**
- Create: `frontend-v2/src/pages/user/__tests__/keyUsageConfig.spec.ts`
- Later implementation target: `frontend-v2/src/pages/user/keyUsageConfig.ts`

- [ ] **Step 1: Write failing tests**

Create `frontend-v2/src/pages/user/__tests__/keyUsageConfig.spec.ts`:

```ts
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
```

- [ ] **Step 2: Run the tests to verify RED**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/keyUsageConfig.spec.ts
```

Expected: FAIL because `../keyUsageConfig` does not exist.

---

### Task 4: Implement key usage config helper

**Files:**
- Create: `frontend-v2/src/pages/user/keyUsageConfig.ts`
- Test: `frontend-v2/src/pages/user/__tests__/keyUsageConfig.spec.ts`

- [ ] **Step 1: Write implementation**

Create `frontend-v2/src/pages/user/keyUsageConfig.ts` with these exported types and functions. Keep the OpenCode model catalog concise; the UI only needs a useful config example and the tests focus on routing/platform correctness.

```ts
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

function buildOpenCodeConfig(platform: string, baseUrl: string, apiKey: string, pathLabel = 'opencode.json'): UsageFileConfig {
  const modelId = platform.includes('gemini') ? 'gemini-2.5-pro' : platform.includes('anthropic') ? 'claude-sonnet-4-5' : 'gpt-5.4'
  return {
    path: pathLabel,
    hint: 'keys.useKeyModal.opencode.hint',
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
  const apiBase = ensureApiV1Base(root)
  const antigravityBase = ensureApiV1Base(`${root}/antigravity`)
  const geminiBase = ensureGeminiV1BetaBase(root)
  const antigravityGeminiBase = ensureGeminiV1BetaBase(`${root}/antigravity`)

  if (options.clientTab === 'opencode') {
    switch (options.platform) {
      case 'anthropic':
        return [buildOpenCodeConfig('anthropic', apiBase, options.apiKey)]
      case 'openai':
        return [buildOpenCodeConfig('openai', apiBase, options.apiKey)]
      case 'gemini':
        return [buildOpenCodeConfig('gemini', geminiBase, options.apiKey)]
      case 'antigravity':
        return [
          buildOpenCodeConfig('antigravity-claude', antigravityBase, options.apiKey, 'opencode.json (Claude)'),
          buildOpenCodeConfig('antigravity-gemini', antigravityGeminiBase, options.apiKey, 'opencode.json (Gemini)')
        ]
      default:
        return [buildOpenCodeConfig('anthropic', apiBase, options.apiKey)]
    }
  }

  switch (options.platform) {
    case 'openai':
      if (options.clientTab === 'claude') return buildAnthropicFiles(root, options.apiKey, options.shellTab)
      return buildOpenAIFiles(root, options.apiKey, options.shellTab, options.clientTab === 'codex-ws', options.t)
    case 'gemini':
      return [buildGeminiCliFile(root, options.apiKey, options.shellTab, options.t)]
    case 'antigravity':
      if (options.clientTab === 'gemini') return [buildGeminiCliFile(`${root}/antigravity`, options.apiKey, options.shellTab, options.t)]
      return buildAnthropicFiles(`${root}/antigravity`, options.apiKey, options.shellTab)
    default:
      return buildAnthropicFiles(root, options.apiKey, options.shellTab)
  }
}
```

- [ ] **Step 2: Run tests to verify GREEN**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/keyUsageConfig.spec.ts
```

Expected: PASS.

---

### Task 5: Add failing page integration tests

**Files:**
- Modify: `frontend-v2/src/pages/user/__tests__/Keys.spec.tsx`
- Later implementation targets: `frontend-v2/src/pages/user/Keys.tsx`, `UseKeyModal.tsx`, `CcsClientSelectModal.tsx`

- [ ] **Step 1: Update the auth store mock**

In `Keys.spec.tsx`, replace the current `vi.mock('@/stores/auth', ...)` block with a mutable store mock:

```tsx
const publicSettings = vi.hoisted(() => ({
  current: {
    registration_enabled: true,
    email_verify_enabled: false,
    registration_email_suffix_whitelist: [],
    promo_code_enabled: false,
    password_reset_enabled: true,
    invitation_code_enabled: false,
    turnstile_enabled: false,
    turnstile_site_key: '',
    site_name: 'XlabAPI',
    site_logo: '',
    site_subtitle: '',
    api_base_url: 'https://api.example.com',
    contact_info: '',
    doc_url: '',
    home_content: '',
    hide_ccs_import_button: false,
    payment_enabled: false,
    channel_monitor_enabled: false,
    available_channels_enabled: false,
    affiliate_enabled: false,
    purchase_subscription_enabled: false,
    purchase_subscription_url: '',
    custom_menu_items: [],
    linuxdo_oauth_enabled: false,
    sora_client_enabled: false,
    backend_mode_enabled: false,
    version: ''
  },
  loadPublicSettings: vi.fn()
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: (selector?: any) => {
    const state = {
      user: { email: 'user@example.com', balance: 0 },
      runMode: 'standard',
      publicSettings: publicSettings.current,
      loadPublicSettings: publicSettings.loadPublicSettings,
      isAdmin: () => false,
      logout: vi.fn()
    }
    return typeof selector === 'function' ? selector(state) : state
  }
}))
```

Add this reset inside `beforeEach`:

```ts
publicSettings.current.hide_ccs_import_button = false
publicSettings.current.api_base_url = 'https://api.example.com'
publicSettings.current.site_name = 'XlabAPI'
publicSettings.loadPublicSettings.mockResolvedValue(undefined)
vi.spyOn(window, 'open').mockImplementation(() => null)
vi.spyOn(document, 'hasFocus').mockReturnValue(false)
```

Add this cleanup inside `beforeEach` after `vi.clearAllMocks()` if not already present:

```ts
vi.restoreAllMocks()
```

Then call `vi.clearAllMocks()` after the spies are created so the render assertions are clean:

```ts
vi.clearAllMocks()
```

- [ ] **Step 2: Add row action and hide flag tests**

Append these tests inside `describe('KeysPage', ...)`:

```tsx
it('renders use key and CC-Switch row actions', async () => {
  renderPage()

  await screen.findByText('fixed key')
  expect(screen.getByRole('button', { name: 'keys.useKey' })).toBeTruthy()
  expect(screen.getByRole('button', { name: 'keys.importToCcSwitch' })).toBeTruthy()
})

it('hides CC-Switch import action when public setting disables it', async () => {
  publicSettings.current.hide_ccs_import_button = true
  renderPage()

  await screen.findByText('fixed key')
  expect(screen.getByRole('button', { name: 'keys.useKey' })).toBeTruthy()
  expect(screen.queryByRole('button', { name: 'keys.importToCcSwitch' })).toBeNull()
})
```

- [ ] **Step 3: Add direct CCS deeplink test**

Append:

```tsx
it('imports an OpenAI key to CC-Switch with the legacy deeplink payload', async () => {
  const user = userEvent.setup()
  renderPage()

  await screen.findByText('fixed key')
  await user.click(screen.getByRole('button', { name: 'keys.importToCcSwitch' }))

  expect(window.open).toHaveBeenCalledTimes(1)
  const [url, target] = (window.open as any).mock.calls[0] as [string, string]
  expect(target).toBe('_self')
  expect(url.startsWith('ccswitch://v1/import?')).toBe(true)
  const params = new URLSearchParams(url.replace('ccswitch://v1/import?', ''))
  expect(params.get('app')).toBe('codex')
  expect(params.get('name')).toBe('XlabAPI')
  expect(params.get('homepage')).toBe('https://api.example.com')
  expect(params.get('endpoint')).toBe('https://api.example.com')
  expect(params.get('apiKey')).toBe('sk-test-key')
  expect(params.get('usageEnabled')).toBe('true')
})
```

- [ ] **Step 4: Add Antigravity client selection test**

Append:

```tsx
it('asks for Antigravity CC-Switch client type before importing', async () => {
  const user = userEvent.setup()
  const antigravityGroup = { ...fixedGroup, id: 30, name: 'AG Pool', platform: 'antigravity' }
  listKeys.mockResolvedValueOnce({
    items: [{ ...keyItem, id: 101, name: 'ag key', key: 'sk-ag-key', group_id: 30, group: antigravityGroup }],
    total: 1,
    pages: 1
  })
  getUserGroups.mockResolvedValueOnce([fixedGroup, dynamicGroup, antigravityGroup])

  renderPage()

  await screen.findByText('ag key')
  await user.click(screen.getByRole('button', { name: 'keys.importToCcSwitch' }))

  const dialog = await screen.findByRole('dialog')
  expect(within(dialog).getByText('keys.ccsClientSelect.title')).toBeTruthy()
  await user.click(within(dialog).getByRole('button', { name: /keys\.ccsClientSelect\.geminiCli/i }))

  const [url] = (window.open as any).mock.calls[0] as [string]
  const params = new URLSearchParams(url.replace('ccswitch://v1/import?', ''))
  expect(params.get('app')).toBe('gemini')
  expect(params.get('endpoint')).toBe('https://api.example.com/antigravity')
  expect(params.get('apiKey')).toBe('sk-ag-key')
})
```

- [ ] **Step 5: Add Use Key modal test**

Append:

```tsx
it('opens use key modal with OpenAI Codex configuration', async () => {
  const user = userEvent.setup()
  renderPage()

  await screen.findByText('fixed key')
  await user.click(screen.getByRole('button', { name: 'keys.useKey' }))

  const dialog = await screen.findByRole('dialog')
  expect(within(dialog).getByText('keys.useKeyModal.title')).toBeTruthy()
  expect(within(dialog).getByRole('button', { name: 'keys.useKeyModal.cliTabs.codexCli' })).toBeTruthy()
  expect(within(dialog).getByRole('button', { name: 'keys.useKeyModal.cliTabs.codexCliWs' })).toBeTruthy()
  expect(within(dialog).getByRole('button', { name: 'keys.useKeyModal.cliTabs.opencode' })).toBeTruthy()
  expect(within(dialog).getByText('~/.codex/config.toml')).toBeTruthy()
  expect(within(dialog).getByText(/base_url = "https:\/\/api\.example\.com"/)).toBeTruthy()
})
```

- [ ] **Step 6: Run integration tests to verify RED**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/Keys.spec.tsx
```

Expected: FAIL because row actions and new modals are not implemented yet.

---

### Task 6: Add React modals

**Files:**
- Create: `frontend-v2/src/pages/user/CcsClientSelectModal.tsx`
- Create: `frontend-v2/src/pages/user/UseKeyModal.tsx`
- Tests: `frontend-v2/src/pages/user/__tests__/Keys.spec.tsx`

- [ ] **Step 1: Create `CcsClientSelectModal.tsx`**

```tsx
import { useTranslation } from 'react-i18next'
import { Bot, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import type { CcsClientType } from './ccswitch'

interface CcsClientSelectModalProps {
  open: boolean
  onClose: () => void
  onSelect: (clientType: CcsClientType) => void
}

export function CcsClientSelectModal({ open, onClose, onSelect }: CcsClientSelectModalProps) {
  const { t } = useTranslation()

  return (
    <Modal open={open} onClose={onClose} title={t('keys.ccsClientSelect.title')}>
      <div className="space-y-4">
        <p className="text-sm text-ink-2">{t('keys.ccsClientSelect.description')}</p>
        <div className="grid gap-3 sm:grid-cols-2">
          <Button variant="secondary" className="h-auto justify-start p-4 text-left" onClick={() => onSelect('claude')}>
            <Bot className="h-5 w-5 text-orange" />
            <span>
              <span className="block font-medium text-ink-1">{t('keys.ccsClientSelect.claudeCode')}</span>
              <span className="block text-xs text-ink-3">{t('keys.ccsClientSelect.claudeCodeDesc')}</span>
            </span>
          </Button>
          <Button variant="secondary" className="h-auto justify-start p-4 text-left" onClick={() => onSelect('gemini')}>
            <Sparkles className="h-5 w-5 text-orange" />
            <span>
              <span className="block font-medium text-ink-1">{t('keys.ccsClientSelect.geminiCli')}</span>
              <span className="block text-xs text-ink-3">{t('keys.ccsClientSelect.geminiCliDesc')}</span>
            </span>
          </Button>
        </div>
      </div>
    </Modal>
  )
}
```

- [ ] **Step 2: Create `UseKeyModal.tsx`**

```tsx
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Check, Clipboard, Info, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { toast } from '@/components/ui/Toast'
import type { GroupPlatform } from '@/types'
import {
  buildClientTabs,
  buildShellTabs,
  buildUsageFiles,
  defaultClientTab,
  platformDescription,
  platformNote,
  type ClientTabId,
  type ShellTabId
} from './keyUsageConfig'

interface UseKeyModalProps {
  open: boolean
  apiKey: string
  baseUrl: string
  platform?: GroupPlatform | null
  allowMessagesDispatch?: boolean
  onClose: () => void
}

export function UseKeyModal({ open, apiKey, baseUrl, platform, allowMessagesDispatch, onClose }: UseKeyModalProps) {
  const { t } = useTranslation()
  const [clientTab, setClientTab] = useState<ClientTabId>(defaultClientTab(platform))
  const [shellTab, setShellTab] = useState<ShellTabId>('unix')
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)

  useEffect(() => {
    if (!open) return
    setClientTab(defaultClientTab(platform))
    setShellTab('unix')
    setCopiedIndex(null)
  }, [open, platform])

  useEffect(() => {
    setShellTab('unix')
  }, [clientTab])

  const clientTabs = useMemo(
    () => buildClientTabs(platform, allowMessagesDispatch, t as (key: string) => string),
    [allowMessagesDispatch, platform, t]
  )
  const shellTabs = useMemo(() => buildShellTabs(clientTab), [clientTab])
  const files = useMemo(
    () => buildUsageFiles({ platform, clientTab, shellTab, baseUrl, apiKey, t: t as (key: string) => string }),
    [apiKey, baseUrl, clientTab, platform, shellTab, t]
  )
  const note = platformNote(platform, clientTab, shellTab, t as (key: string) => string)

  async function copyContent(content: string, index: number) {
    try {
      await navigator.clipboard.writeText(content)
      setCopiedIndex(index)
      toast.success(t('common.copiedToClipboard') as string)
      window.setTimeout(() => setCopiedIndex(null), 1500)
    } catch {
      toast.error(t('common.copyFailed') as string)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={t('keys.useKeyModal.title')} size="lg">
      <div className="space-y-4">
        {!platform ? (
          <div className="flex items-start gap-3 rounded-xl border border-yellow-500/30 bg-yellow-500/10 p-4">
            <AlertTriangle className="mt-0.5 h-5 w-5 flex-shrink-0 text-yellow-500" />
            <div>
              <p className="text-sm font-medium text-ink-1">{t('keys.useKeyModal.noGroupTitle')}</p>
              <p className="mt-1 text-sm text-ink-2">{t('keys.useKeyModal.noGroupDescription')}</p>
            </div>
          </div>
        ) : (
          <>
            <p className="text-sm text-ink-2">{platformDescription(platform, clientTab, t as (key: string) => string)}</p>

            {clientTabs.length > 0 && (
              <div className="flex flex-wrap gap-2 border-b border-line-2 pb-2">
                {clientTabs.map((tab) => (
                  <button
                    key={tab.id}
                    type="button"
                    className={tab.id === clientTab ? 'btn btn-primary btn-sm' : 'btn btn-ghost btn-sm'}
                    onClick={() => setClientTab(tab.id as ClientTabId)}
                  >
                    <Terminal className="h-4 w-4" />
                    {tab.label}
                  </button>
                ))}
              </div>
            )}

            {shellTabs.length > 0 && (
              <div className="flex flex-wrap gap-2 border-b border-line-2 pb-2">
                {shellTabs.map((tab) => (
                  <button
                    key={tab.id}
                    type="button"
                    className={tab.id === shellTab ? 'btn btn-accent btn-sm' : 'btn btn-ghost btn-sm'}
                    onClick={() => setShellTab(tab.id as ShellTabId)}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>
            )}

            <div className="space-y-4">
              {files.map((file, index) => (
                <div key={`${file.path}-${index}`}>
                  {file.hint && <p className="mb-1.5 text-xs text-yellow-500">{file.hint}</p>}
                  <div className="overflow-hidden rounded-xl border border-line-2 bg-slate-950">
                    <div className="flex items-center justify-between border-b border-slate-800 bg-slate-900 px-4 py-2">
                      <span className="font-mono text-xs text-slate-300">{file.path}</span>
                      <button
                        type="button"
                        className="btn btn-ghost btn-sm text-slate-200"
                        onClick={() => copyContent(file.content, index)}
                      >
                        {copiedIndex === index ? <Check className="h-3.5 w-3.5" /> : <Clipboard className="h-3.5 w-3.5" />}
                        {copiedIndex === index ? t('keys.useKeyModal.copied') : t('keys.useKeyModal.copy')}
                      </button>
                    </div>
                    <pre className="max-h-80 overflow-auto p-4 text-sm text-slate-100"><code>{file.content}</code></pre>
                  </div>
                </div>
              ))}
            </div>

            {note && (
              <div className="flex items-start gap-3 rounded-xl border border-blue-500/30 bg-blue-500/10 p-3">
                <Info className="mt-0.5 h-4 w-4 flex-shrink-0 text-blue-400" />
                <p className="text-sm text-ink-2">{note}</p>
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  )
}
```

- [ ] **Step 3: Run page tests to confirm they still fail at integration wiring**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/Keys.spec.tsx
```

Expected: Still FAIL because `Keys.tsx` has not wired the row actions yet, but TypeScript import errors for the new modal files should be gone.

---

### Task 7: Wire row actions in `Keys.tsx`

**Files:**
- Modify: `frontend-v2/src/pages/user/Keys.tsx`
- Tests: `frontend-v2/src/pages/user/__tests__/Keys.spec.tsx`

- [ ] **Step 1: Update imports**

In `Keys.tsx`, change the icon import:

```ts
import { Plus, Copy, Trash2, Pencil, KeyRound, Download } from 'lucide-react'
```

Add imports:

```ts
import { useAuthStore } from '@/stores/auth'
import { CcsClientSelectModal } from './CcsClientSelectModal'
import { UseKeyModal } from './UseKeyModal'
import { buildCcsImportUrl, type CcsClientType } from './ccswitch'
```

- [ ] **Step 2: Add state and public settings selectors**

Inside `KeysPage`, after `const qc = useQueryClient()`, add:

```ts
const publicSettings = useAuthStore((s) => s.publicSettings)
const loadPublicSettings = useAuthStore((s) => s.loadPublicSettings)
```

After existing `useState` declarations, add:

```ts
const [useKeyTarget, setUseKeyTarget] = useState<ApiKey | null>(null)
const [pendingCcsKey, setPendingCcsKey] = useState<ApiKey | null>(null)
```

- [ ] **Step 3: Load public settings when missing**

After the usage-stats `useEffect`, add:

```ts
useEffect(() => {
  if (!publicSettings) {
    loadPublicSettings().catch(() => {})
  }
}, [loadPublicSettings, publicSettings])
```

- [ ] **Step 4: Add helper functions**

Before `onSubmit`, add:

```ts
function apiBaseUrl() {
  return (publicSettings?.api_base_url || window.location.origin).replace(/\/+$/, '')
}

function executeCcsImport(row: ApiKey, clientType?: CcsClientType) {
  const deeplink = buildCcsImportUrl({
    apiKey: row.key,
    platform: row.group?.platform || 'anthropic',
    clientType,
    baseUrl: publicSettings?.api_base_url || window.location.origin,
    siteName: publicSettings?.site_name || 'sub2api'
  })

  try {
    window.open(deeplink, '_self')
    window.setTimeout(() => {
      if (document.hasFocus()) {
        toast.error(t('keys.ccSwitchNotInstalled') as string)
      }
    }, 100)
  } catch {
    toast.error(t('keys.ccSwitchNotInstalled') as string)
  }
}

function importToCcswitch(row: ApiKey) {
  if (row.group?.platform === 'antigravity') {
    setPendingCcsKey(row)
    return
  }
  executeCcsImport(row, row.group?.platform === 'gemini' ? 'gemini' : 'claude')
}

function handleCcsClientSelect(clientType: CcsClientType) {
  if (pendingCcsKey) executeCcsImport(pendingCcsKey, clientType)
  setPendingCcsKey(null)
}
```

- [ ] **Step 5: Add buttons to the row action column**

Replace the existing action `<div className="inline-flex gap-1">` content with:

```tsx
<button
  type="button"
  className="btn btn-ghost btn-sm"
  onClick={() => setUseKeyTarget(k)}
>
  <KeyRound className="h-3.5 w-3.5" />
  {t('keys.useKey')}
</button>
{!publicSettings?.hide_ccs_import_button && (
  <button
    type="button"
    className="btn btn-ghost btn-sm"
    onClick={() => importToCcswitch(k)}
  >
    <Download className="h-3.5 w-3.5" />
    {t('keys.importToCcSwitch')}
  </button>
)}
<button
  title={t('keys.editKey') as string}
  className="btn btn-ghost btn-icon btn-sm"
  onClick={() => openEditForm(k)}
>
  <Pencil className="h-3.5 w-3.5" />
</button>
<button
  title={t('keys.deleteKey') as string}
  className="btn btn-ghost btn-icon btn-sm text-signal-err"
  onClick={() => {
    if (confirm(t('keys.deleteConfirmMessage', { name: k.name }) as string)) {
      deleteMut.mutate(k.id)
    }
  }}
>
  <Trash2 className="h-3.5 w-3.5" />
</button>
```

Also change the action wrapper class to allow wrapping:

```tsx
<div className="inline-flex flex-wrap justify-end gap-1">
```

- [ ] **Step 6: Render the modals**

Before the existing create/edit `<Modal ...>` block or after it within the fragment, add:

```tsx
<UseKeyModal
  open={!!useKeyTarget}
  apiKey={useKeyTarget?.key || ''}
  baseUrl={apiBaseUrl()}
  platform={useKeyTarget?.group?.platform || null}
  allowMessagesDispatch={useKeyTarget?.group?.allow_messages_dispatch}
  onClose={() => setUseKeyTarget(null)}
/>

<CcsClientSelectModal
  open={!!pendingCcsKey}
  onClose={() => setPendingCcsKey(null)}
  onSelect={handleCcsClientSelect}
/>
```

- [ ] **Step 7: Run page tests to verify GREEN**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/Keys.spec.tsx
```

Expected: PASS. If the accessible name for the Gemini button includes both title and description, adjust the test matcher to `name: /keys\.ccsClientSelect\.geminiCli/`.

---

### Task 8: Typecheck and full frontend verification

**Files:**
- No new code unless verification finds issues.

- [ ] **Step 1: Run focused tests together**

Run from `frontend-v2`:

```bash
npm exec vitest run src/pages/user/__tests__/ccswitch.spec.ts src/pages/user/__tests__/keyUsageConfig.spec.ts src/pages/user/__tests__/Keys.spec.tsx
```

Expected: PASS.

- [ ] **Step 2: Run typecheck**

Run from `frontend-v2`:

```bash
npm run typecheck
```

Expected: PASS. Fix any TypeScript errors only in the files touched by this plan.

- [ ] **Step 3: Run production build if typecheck passes**

Run from `frontend-v2`:

```bash
npm run build
```

Expected: PASS. If build fails because of pre-existing unrelated issues outside touched files, record the failure and do not change unrelated files without user approval.

- [ ] **Step 4: Read lints for touched files**

Use the IDE lint reader for:

```text
/root/sub2api-src/frontend-v2/src/pages/user/Keys.tsx
/root/sub2api-src/frontend-v2/src/pages/user/UseKeyModal.tsx
/root/sub2api-src/frontend-v2/src/pages/user/CcsClientSelectModal.tsx
/root/sub2api-src/frontend-v2/src/pages/user/ccswitch.ts
/root/sub2api-src/frontend-v2/src/pages/user/keyUsageConfig.ts
```

Expected: no introduced diagnostics.

---

## Self-Review

- Spec coverage: The plan covers direct row buttons, public settings, Use Key modal, CC-Switch direct import, Antigravity client selection, tests, and verification.
- Placeholder scan: No TBD/TODO/later placeholders remain. The only conditional instruction is limited to correcting an accessible-name matcher if Testing Library reports a different button name.
- Type consistency: `CcsClientType`, `ClientTabId`, `ShellTabId`, `GroupPlatform`, and file paths are defined before use and reused consistently across tasks.
