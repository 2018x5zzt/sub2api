import { describe, expect, it, vi } from 'vitest'

vi.mock('../../xlabClient', () => ({
  xlabClient: {
    get: vi.fn(async (path: string) => ({ data: path }))
  }
}))

import { xlabClient } from '../../xlabClient'
import { getActive, getProgress, getSummary } from '../subscriptionProducts'

describe('xlab subscriptionProducts API', () => {
  it('uses xlab endpoints for product subscription reads', async () => {
    await expect(getActive()).resolves.toBe('/subscription-products/active')
    await expect(getSummary()).resolves.toBe('/subscription-products/summary')
    await expect(getProgress()).resolves.toBe('/subscription-products/progress')
    expect(xlabClient.get).toHaveBeenCalledTimes(3)
  })
})
