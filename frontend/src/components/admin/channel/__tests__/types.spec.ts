import { describe, expect, it } from 'vitest'

import { apiPricingToFormEntry, formPricingToAPI } from '../types'

describe('channel pricing transforms', () => {
  it('maps legacy image_output_price into the editable image default price', () => {
    const entry = apiPricingToFormEntry({
      platform: 'openai',
      models: ['gpt-image-1'],
      billing_mode: 'image',
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      image_output_price: 0.75,
      per_request_price: null,
      intervals: [],
    })

    expect(entry.per_request_price).toBe(0.75)
  })

  it('serializes image billing defaults without token unit conversion', () => {
    const pricing = formPricingToAPI('openai', {
      models: ['gpt-image-1'],
      billing_mode: 'image',
      input_price: null,
      output_price: null,
      cache_write_price: null,
      cache_read_price: null,
      image_output_price: null,
      per_request_price: 0.75,
      intervals: [],
    })

    expect(pricing.per_request_price).toBe(0.75)
    expect(pricing.image_output_price).toBe(0.75)
  })
})
