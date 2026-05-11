import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

/** Mono uppercase caption with an orange square marker. */
export function Eyebrow({ children, className }: { children: ReactNode; className?: string }) {
  return <span className={cn('eyebrow', className)}>{children}</span>
}
