import { forwardRef, type AnchorHTMLAttributes, type ButtonHTMLAttributes, type ReactNode } from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/cn'

type Variant = 'accent' | 'light' | 'ghost'
type Size = 'sm' | 'md' | 'lg'

const variant: Record<Variant, string> = {
  accent: 'bg-orange-display text-white hover:bg-orange-hover border-transparent',
  light: 'bg-ink-1 text-bg-0 hover:bg-black border-transparent',
  ghost: 'bg-transparent text-ink-1 border-line-3 hover:bg-bg-3'
}

const sizeClass: Record<Size, string> = {
  sm: 'h-9 px-4 text-xs',
  md: 'h-11 px-5 text-sm',
  lg: 'h-12 px-6 text-[15px]'
}

interface PillProps {
  variant?: Variant
  size?: Size
  loading?: boolean
  children?: ReactNode
}

type PillButtonProps = PillProps & ButtonHTMLAttributes<HTMLButtonElement>
type PillLinkProps = PillProps & AnchorHTMLAttributes<HTMLAnchorElement> & { as: 'a' }

export const PillBtn = forwardRef<HTMLButtonElement, PillButtonProps>(function PillBtn(
  { variant: v = 'accent', size = 'md', className, loading, disabled, children, ...rest },
  ref
) {
  return (
    <button
      ref={ref}
      disabled={disabled || loading}
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-full font-medium transition-all border tracking-tight',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        variant[v],
        sizeClass[size],
        className
      )}
      {...rest}
    >
      {loading && <Loader2 className="h-4 w-4 animate-spin" />}
      {children}
    </button>
  )
})

export function PillLink({
  variant: v = 'accent',
  size = 'md',
  className,
  children,
  ...rest
}: PillLinkProps) {
  return (
    <a
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-full font-medium transition-all border tracking-tight no-underline',
        variant[v],
        sizeClass[size],
        className
      )}
      {...rest}
    >
      {children}
    </a>
  )
}
