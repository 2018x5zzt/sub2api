import { describe, expect, it } from 'vitest'
import { routes } from '@/router'

describe('admin subscription management routes', () => {
  it('serves the product subscription user overview at the canonical subscriptions path', () => {
    const canonicalRoute = routes.find((route) => route.path === '/admin/subscriptions')
    const compatibilityRoute = routes.find((route) => route.path === '/admin/subscription-products')

    expect(canonicalRoute).toBeDefined()
    expect(canonicalRoute?.name).toBe('AdminSubscriptions')
    expect(canonicalRoute?.redirect).toBeUndefined()
    expect(canonicalRoute?.meta?.titleKey).toBe('admin.subscriptionProducts.title')

    expect(compatibilityRoute).toBeDefined()
    expect(compatibilityRoute?.redirect).toBe('/admin/subscriptions')
  })
})
