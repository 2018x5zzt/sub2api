import { describe, expect, it } from 'vitest'

import { isLikelyOpenAICacheCreationMetricUnavailable } from '../cacheMetrics'

describe('isLikelyOpenAICacheCreationMetricUnavailable', () => {
  it('returns true when OpenAI responses rows have cache reads but zero cache creation', () => {
    expect(isLikelyOpenAICacheCreationMetricUnavailable({
      cache_creation_tokens: 0,
      cache_read_tokens: 128,
      upstream_endpoint: '/v1/responses'
    })).toBe(true)
  })

  it('returns false when cache creation is already reported', () => {
    expect(isLikelyOpenAICacheCreationMetricUnavailable({
      cache_creation_tokens: 32,
      cache_read_tokens: 128,
      upstream_endpoint: '/v1/responses'
    })).toBe(false)
  })

  it('returns false for non-OpenAI-looking rows', () => {
    expect(isLikelyOpenAICacheCreationMetricUnavailable({
      cache_creation_tokens: 0,
      cache_read_tokens: 128,
      upstream_endpoint: '/v1/messages',
      model: 'claude-sonnet-4-20250514'
    })).toBe(false)
  })
})
