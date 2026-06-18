import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('../../xlabClient', () => ({
  xlabClient: {
    get: vi.fn(async (path: string, config?: unknown) => ({ data: { path, config } }))
  }
}))

import { xlabClient } from '../../xlabClient'
import { getMyOrders, getOrder, xlabPaymentAPI } from '../payment'

describe('xlab payment API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('uses the xlab endpoint and params for my payment orders', async () => {
    const params = { page: 2, page_size: 20, status: 'PENDING' }

    await expect(getMyOrders(params)).resolves.toEqual({
      path: '/payment/orders/my',
      config: { params }
    })

    expect(xlabClient.get).toHaveBeenCalledWith('/payment/orders/my', { params })
  })

  it('uses the xlab endpoint for a payment order read', async () => {
    await expect(getOrder(123)).resolves.toEqual({
      path: '/payment/orders/123',
      config: undefined
    })

    expect(xlabClient.get).toHaveBeenCalledWith('/payment/orders/123')
  })

  it('exports the xlab payment API facade', () => {
    expect(xlabPaymentAPI).toEqual({ getMyOrders, getOrder })
  })
})
