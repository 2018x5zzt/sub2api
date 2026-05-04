import { beforeEach, describe, expect, it, vi } from 'vitest'

const { post, put } = vi.hoisted(() => ({
  post: vi.fn(),
  put: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: {
    post,
    put,
  },
}))

import { keysAPI } from '@/api/keys'
import { apiKeysAPI } from '@/api/admin/apiKeys'

describe('keys api', () => {
  beforeEach(() => {
    post.mockReset()
    put.mockReset()
    post.mockResolvedValue({ data: {} })
    put.mockResolvedValue({ data: {} })
  })

  it('does not send subscription product family when creating a user key', async () => {
    await keysAPI.create('gpt key', 21)

    expect(post).toHaveBeenCalledWith('/keys', {
      name: 'gpt key',
      group_id: 21,
    })
  })

  it('does not send subscription product family on user key update', async () => {
    await keysAPI.update(42, {
      group_id: 9,
    })

    expect(put).toHaveBeenCalledWith('/keys/42', {
      group_id: 9,
    })
  })

  it('allows admin group updates to reset rate limit usage', async () => {
    await apiKeysAPI.updateApiKeyGroup(99, {
      group_id: 7,
      reset_rate_limit_usage: true,
    })

    expect(put).toHaveBeenCalledWith('/admin/api-keys/99', {
      group_id: 7,
      reset_rate_limit_usage: true,
    })
  })

  it('keeps admin group updates backward compatible with a bare group id', async () => {
    await apiKeysAPI.updateApiKeyGroup(77, 9)

    expect(put).toHaveBeenCalledWith('/admin/api-keys/77', {
      group_id: 9,
    })
  })
})
