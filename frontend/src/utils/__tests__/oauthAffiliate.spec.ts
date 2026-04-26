import { beforeEach, describe, expect, it } from 'vitest'
import {
  clearAffiliateReferralCode,
  loadAffiliateReferralCode,
  normalizeOAuthAffiliateCode,
  pickOAuthAffiliateCode,
  resolveAffiliateReferralCode,
  storeAffiliateReferralCode
} from '@/utils/oauthAffiliate'

describe('oauthAffiliate', () => {
  beforeEach(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  it('normalizes string and array query values without changing code case', () => {
    expect(normalizeOAuthAffiliateCode('  Aff-Code_2026  ')).toBe('Aff-Code_2026')
    expect(normalizeOAuthAffiliateCode(['FIRST', 'SECOND'])).toBe('FIRST')
    expect(normalizeOAuthAffiliateCode(undefined)).toBe('')
  })

  it('prefers aff values before legacy invite values', () => {
    expect(pickOAuthAffiliateCode('', 'AFF2026', 'INVITE2025')).toBe('AFF2026')
    expect(pickOAuthAffiliateCode('', '', 'INVITE2025')).toBe('INVITE2025')
  })

  it('stores referral codes and ignores expired stored codes', () => {
    storeAffiliateReferralCode('AFF2026', 1000)
    expect(loadAffiliateReferralCode(1000)).toBe('AFF2026')
    expect(loadAffiliateReferralCode(1000 + 31 * 24 * 60 * 60 * 1000)).toBe('')
  })

  it('resolves explicit codes before stored codes for compatibility redirects', () => {
    storeAffiliateReferralCode('STORED')
    expect(resolveAffiliateReferralCode('', 'CURRENT')).toBe('CURRENT')
    expect(loadAffiliateReferralCode()).toBe('CURRENT')
    clearAffiliateReferralCode()
    expect(resolveAffiliateReferralCode('', '')).toBe('')
  })
})
