/**
 * 薄壳软不变量测试 (S1–S3)
 *
 * 读取 router/index.ts 原始文本做字符串断言，无需 import router。
 * 失败时 CI 产生 warning annotation，需人工确认后方可合并。
 *
 * 运行方式：pnpm vitest run src/xlab/__tests__/shell_invariants.spec.ts
 */
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const ROUTER_PATH = resolve(__dirname, '../../router/index.ts')
const routerContent = readFileSync(ROUTER_PATH, 'utf-8')

describe('Shell Invariants (Soft) — S1–S3', () => {
  // S1: 生图路由必须存在并指向 ImageStudioView
  // 丢失后果：生图入口消失，用户无法访问 canvas.xlabapi.com 生图工作台
  // 恢复方式：在 router/index.ts 用户路由区补回 /image-studio → ImageStudioView 路由
  it('S1: /image-studio 路由指向 ImageStudioView', () => {
    expect(
      routerContent,
      'S1 FAIL: router/index.ts 缺少 /image-studio 路由\n恢复方式：在用户路由区补回 { path: "/image-studio", component: () => import("@/views/user/ImageStudioView.vue") }',
    ).toContain('image-studio')

    expect(
      routerContent,
      'S1 FAIL: router/index.ts 中 /image-studio 未指向 ImageStudioView\n恢复方式：确认 component 指向 @/views/user/ImageStudioView.vue',
    ).toContain('ImageStudioView')
  })

  // S2: /oauth/consent 路由必须存在并指向 XlabOAuthConsentView
  // 丢失后果：Miku OAuth 授权确认页404，SSO 登录流程中断
  // 恢复方式：在 router/index.ts 中补回 { path: "/oauth/consent", name: "XlabOAuthConsent", ... }
  it('S2: /oauth/consent 路由指向 XlabOAuthConsent', () => {
    expect(
      routerContent,
      'S2 FAIL: router/index.ts 缺少 /oauth/consent 路由\n恢复方式：补回 { path: "/oauth/consent", name: "XlabOAuthConsent", component: XlabOAuthConsentView }',
    ).toContain('/oauth/consent')

    expect(
      routerContent,
      'S2 FAIL: /oauth/consent 路由未命名 XlabOAuthConsent\n恢复方式：确认 name: "XlabOAuthConsent" 字段存在',
    ).toContain('XlabOAuthConsent')
  })

  // S3: 订阅/兑换/模型中心路由必须存在（后端定制强制配套）
  // 丢失后果：前后端契约断裂，product-subscription 后端有 API 但前端无对应页面
  // 恢复方式：确认对应 View 文件及路由均在，参考 xlabapi-snapshot-0.1.137
  it('S3: /subscriptions、/redeem、/model-hub 路由均存在', () => {
    const required: Array<[string, string]> = [
      ['/subscriptions', '订阅'],
      ['/redeem', '兑换码'],
      ['/models', '模型中心'],
    ]
    for (const [path, label] of required) {
      expect(
        routerContent,
        `S3 FAIL: router/index.ts 缺少 ${path}（${label}）路由\n恢复方式：确认 ${label} View 文件及路由注册均在`,
      ).toContain(path)
    }
  })
})
