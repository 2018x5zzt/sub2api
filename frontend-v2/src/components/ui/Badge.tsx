import type { HTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

type Tone = 'neutral' | 'success' | 'warning' | 'danger' | 'accent'

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: Tone
  dot?: boolean
}

const toneClass: Record<Tone, string> = {
  neutral: '',
  success: 'badge-success',
  warning: 'badge-warning',
  danger: 'badge-danger',
  accent: 'badge-accent'
}

export function Badge({ tone = 'neutral', dot, className, children, ...rest }: BadgeProps) {
  return (
    <span className={cn('badge', toneClass[tone], className)} {...rest}>
      {dot && (
        <span
          className="inline-block rounded-full"
          style={{
            width: 6,
            height: 6,
            background: 'currentColor'
          }}
        />
      )}
      {children}
    </span>
  )
}
