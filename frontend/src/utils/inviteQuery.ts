export function getInviteCodeFromQuery(raw: unknown): string {
  if (typeof raw !== 'string') return ''
  return raw.trim()
}
