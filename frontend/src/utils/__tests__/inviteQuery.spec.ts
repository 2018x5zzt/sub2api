import { describe, expect, it } from 'vitest'
import { getInviteCodeFromQuery } from '@/utils/inviteQuery'

describe('getInviteCodeFromQuery', () => {
  it('returns an uppercased invite code from string query params', () => {
    expect(getInviteCodeFromQuery('hello123')).toBe('HELLO123')
  })

  it('returns an empty string for non-string query params', () => {
    expect(getInviteCodeFromQuery(['hello123'])).toBe('')
    expect(getInviteCodeFromQuery(undefined)).toBe('')
  })
})
