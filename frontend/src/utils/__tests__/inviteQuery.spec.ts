import { describe, expect, it } from 'vitest'
import { getInviteCodeFromQuery } from '@/utils/inviteQuery'

describe('getInviteCodeFromQuery', () => {
  it('preserves invite code case from string query params', () => {
    expect(getInviteCodeFromQuery('AbCdEfGh')).toBe('AbCdEfGh')
  })

  it('trims surrounding whitespace from string query params', () => {
    expect(getInviteCodeFromQuery('  AbCdEfGh  ')).toBe('AbCdEfGh')
  })

  it('returns an empty string for non-string query params', () => {
    expect(getInviteCodeFromQuery(['hello123'])).toBe('')
    expect(getInviteCodeFromQuery(undefined)).toBe('')
  })
})
