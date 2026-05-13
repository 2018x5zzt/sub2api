interface Props {
  showTop?: boolean
  showBottom?: boolean
  /** Half-width of the content column (sets where vertical gutter lines sit). */
  halfWidth?: number
}

/**
 * Per-section grid frame: vertical hairlines outside the content column, optional
 * top/bottom rules. Lives inside a relatively positioned section so it never
 * crosses content and respects section overflow.
 */
export function SectionFrame({ showTop = true, showBottom = true, halfWidth = 600 }: Props) {
  const leftX = `max(20px, calc(50% - ${halfWidth}px))`
  const rightX = `max(20px, calc(50% - ${halfWidth}px))`
  return (
    <div className="absolute inset-0 z-[3] pointer-events-none">
      <div className="absolute" style={{ top: 0, bottom: 0, left: leftX, width: 1, background: 'var(--line-3)' }} />
      <div className="absolute" style={{ top: 0, bottom: 0, right: rightX, width: 1, background: 'var(--line-3)' }} />
      {showTop && (
        <div className="absolute" style={{ top: 0, left: 0, right: 0, height: 1, background: 'var(--line-3)' }} />
      )}
      {showBottom && (
        <div className="absolute" style={{ bottom: 0, left: 0, right: 0, height: 1, background: 'var(--line-3)' }} />
      )}
    </div>
  )
}
