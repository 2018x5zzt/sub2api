const AFFILIATE_REFERRAL_CODE_KEY = 'affiliate_referral_code'
const AFFILIATE_REFERRAL_TTL_MS = 30 * 24 * 60 * 60 * 1000

type StoredAffiliateReferralCode = {
  code: string
  expiresAt: number
}

function normalizeAffiliateCode(value?: unknown): string {
  const raw = Array.isArray(value) ? value[0] : value
  return typeof raw === 'string' ? raw.trim() : ''
}

export function storeAffiliateReferralCode(value?: unknown, now = Date.now()): void {
  if (typeof window === 'undefined') return
  const code = normalizeAffiliateCode(value)
  if (!code) return
  try {
    const payload: StoredAffiliateReferralCode = {
      code,
      expiresAt: now + AFFILIATE_REFERRAL_TTL_MS
    }
    window.localStorage.setItem(AFFILIATE_REFERRAL_CODE_KEY, JSON.stringify(payload))
  } catch {
    // Ignore browser storage exceptions.
  }
}

export function loadAffiliateReferralCode(now = Date.now()): string {
  if (typeof window === 'undefined') return ''
  try {
    const raw = window.localStorage.getItem(AFFILIATE_REFERRAL_CODE_KEY)
    if (!raw) return ''
    const parsed = JSON.parse(raw) as Partial<StoredAffiliateReferralCode>
    const code = normalizeAffiliateCode(parsed.code)
    const expiresAt = Number(parsed.expiresAt) || 0
    if (!code || expiresAt <= now) {
      clearAffiliateReferralCode()
      return ''
    }
    return code
  } catch {
    clearAffiliateReferralCode()
    return ''
  }
}

export function clearAffiliateReferralCode(): void {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.removeItem(AFFILIATE_REFERRAL_CODE_KEY)
  } catch {
    // Ignore browser storage exceptions.
  }
}

export function resolveAffiliateReferralCode(...values: unknown[]): string {
  for (const value of values) {
    const code = normalizeAffiliateCode(value)
    if (code) {
      storeAffiliateReferralCode(code)
      return code
    }
  }
  return loadAffiliateReferralCode()
}

