import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

describe('admin subscription product locale keys', () => {
  it('contains zh labels for the subscription products admin page', () => {
    expect(zh.admin.subscriptionProducts.title).toBe('产品订阅')
    expect(zh.admin.subscriptionProducts.description).toBe('管理共享订阅产品、分组绑定和用户产品订阅')
    expect(zh.admin.subscriptionProducts.createProduct).toBe('创建产品')
    expect(zh.admin.subscriptionProducts.columns.defaultValidity).toBe('有效天数')
  })

  it('contains en labels for the subscription products admin page', () => {
    expect(en.admin.subscriptionProducts.title).toBe('Product Subscriptions')
    expect(en.admin.subscriptionProducts.description).toBe('Manage shared subscription products, group bindings, and user product subscriptions')
    expect(en.admin.subscriptionProducts.createProduct).toBe('Create Product')
    expect(en.admin.subscriptionProducts.columns.defaultValidity).toBe('Validity Days')
  })
})
