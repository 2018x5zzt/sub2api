import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import PricingEntryCard from '../PricingEntryCard.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (_key: string, fallback?: string) => fallback ?? _key,
    }),
  }
})

describe('PricingEntryCard', () => {
  it('uses per_request_price as the editable default value for image billing mode', async () => {
    const wrapper = mount(PricingEntryCard, {
      props: {
        entry: {
          models: ['gpt-image-1'],
          billing_mode: 'image',
          input_price: null,
          output_price: null,
          cache_write_price: null,
          cache_read_price: null,
          image_output_price: null,
          per_request_price: 0.8,
          intervals: [],
        },
      },
      global: {
        stubs: {
          Icon: { template: '<span />' },
          Select: { template: '<span />' },
          IntervalRow: { template: '<span />' },
          ModelTagInput: { template: '<span />' },
        },
      },
    })

    await wrapper.find('.cursor-pointer').trigger('click')

    const priceInput = wrapper.find('input[type="number"]')
    expect((priceInput.element as HTMLInputElement).value).toBe('0.8')

    await priceInput.setValue('1.2')

    const events = wrapper.emitted('update')
    expect(events).toBeTruthy()

    const lastEntry = events!.at(-1)![0] as Record<string, unknown>
    expect(lastEntry.per_request_price).toBe('1.2')
    expect(lastEntry.image_output_price).toBeNull()
  })
})
