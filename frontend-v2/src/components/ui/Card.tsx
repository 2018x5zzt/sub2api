import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  flat?: boolean
}

export function Card({ className, flat, ...rest }: CardProps) {
  return <div className={cn(flat ? 'card-flat' : 'card', className)} {...rest} />
}

export function CardHeader({
  title,
  description,
  action,
  className
}: {
  title: ReactNode
  description?: ReactNode
  action?: ReactNode
  className?: string
}) {
  return (
    <div className={cn('flex items-start justify-between gap-4 px-6 pt-5 pb-3', className)}>
      <div className="min-w-0">
        <h3 className="text-base font-medium text-ink-1">{title}</h3>
        {description && <p className="text-sm text-ink-3 mt-0.5">{description}</p>}
      </div>
      {action}
    </div>
  )
}

export function CardBody({ className, ...rest }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('px-6 pb-5', className)} {...rest} />
}
