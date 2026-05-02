import { beforeEach, describe, expect, it, vi } from 'vitest'

const { post } = vi.hoisted(() => ({
  post: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    post,
  },
}))

import { redeemAPI } from '@/api/admin/redeem'

describe('admin redeem api', () => {
  beforeEach(() => {
    post.mockReset()
    post.mockResolvedValue({ data: [] })
  })

  it('generates subscription card codes with product_id for product subscription stock', async () => {
    await redeemAPI.generate(20, 'subscription', 1, {
      productId: 88,
      validityDays: 30,
    })

    expect(post).toHaveBeenCalledWith('/admin/redeem-codes/generate', {
      count: 20,
      type: 'subscription',
      value: 1,
      product_id: 88,
      validity_days: 30,
    })
  })
})
