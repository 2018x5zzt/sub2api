import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('admin navigation locale labels', () => {
  it('defines subscription navigation labels in both locales', () => {
    expect(en.nav.subscriptionManagement).toBe('Subscription Management')
    expect(en.nav.subscriptionProductConfig).toBe('Subscription Products')
    expect(zh.nav.subscriptionManagement).toBe('订阅管理')
    expect(zh.nav.subscriptionProductConfig).toBe('订阅产品')
  })
})
