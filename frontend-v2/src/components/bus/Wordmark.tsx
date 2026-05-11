import { cn } from '@/lib/cn'

/** Rotated nested-square mark used as the logo throughout the design. */
export function BusMark({ size = 28, className }: { size?: number; className?: string }) {
  const inset = Math.round(size * 0.21)
  return (
    <div
      className={cn('relative shrink-0', className)}
      style={{ width: size, height: size }}
      aria-hidden
    >
      <div
        className="absolute inset-0 border-[1.5px] border-white"
        style={{ transform: 'rotate(45deg)', borderRadius: 2 }}
      />
      <div
        className="absolute border-[1.5px] border-white"
        style={{ inset, borderRadius: 2 }}
      />
    </div>
  )
}

export function Wordmark({
  small,
  name,
  className
}: {
  small?: boolean
  name?: string
  className?: string
}) {
  return (
    <div className={cn('inline-flex items-center gap-3', className)}>
      <BusMark size={small ? 22 : 26} />
      <span
        className="text-ink-1 tracking-tight"
        style={{
          fontFamily: 'var(--font-display, Inter)',
          fontWeight: 600,
          fontSize: small ? 14 : 14.5,
          letterSpacing: '-0.01em',
          textTransform: 'uppercase'
        }}
      >
        {name || 'XLABAPI'}
      </span>
    </div>
  )
}
