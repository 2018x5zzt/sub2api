import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'

const requiredKeys = [
  'eyebrow',
  'title',
  'description',
  'searchLabel',
  'searchPlaceholder',
  'platformFilterLabel',
  'groupFilterLabel',
  'allPlatforms',
  'allGroups',
  'groupsLabel',
  'uniqueModelsLabel',
  'visibleModelsLabel',
  'platformsLabel',
  'sourceDefault',
  'sourceMapping',
  'sourceMixed',
  'pricingComputedWithRate',
  'rateShort',
  'inputPriceShort',
  'outputPriceShort',
  'defaultPriceShort',
  'perMillionTokens',
  'perRequest',
  'perImage',
  'pricingUnavailable',
  'copyVisible',
  'copyGroup',
  'copiedModel',
  'copiedGroup',
  'copiedVisible',
  'modelCount',
  'inputPrice',
  'outputPrice',
  'clearFilters',
  'emptyTitle',
  'emptyDescription',
  'noModelsInGroup',
  'loadFailedTitle',
  'loadFailedDescription',
] as const

describe('model hub locale keys', () => {
  it('contains zh labels used by ModelHubView', () => {
    for (const key of requiredKeys) {
      expect(zh.modelHub[key], `zh.modelHub.${key}`).toEqual(expect.any(String))
      expect(zh.modelHub[key], `zh.modelHub.${key}`).not.toBe('')
    }
    expect(zh.modelHub.title).toBe('模型广场')
    expect(zh.modelHub.copyVisible).toBe('复制当前结果')
  })

  it('contains en labels used by ModelHubView', () => {
    for (const key of requiredKeys) {
      expect(en.modelHub[key], `en.modelHub.${key}`).toEqual(expect.any(String))
      expect(en.modelHub[key], `en.modelHub.${key}`).not.toBe('')
    }
    expect(en.modelHub.title).toBe('Model Hub')
    expect(en.modelHub.copyVisible).toBe('Copy visible results')
  })
})
