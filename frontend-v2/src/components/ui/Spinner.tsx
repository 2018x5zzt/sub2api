import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/cn'

export function Spinner({ className }: { className?: string }) {
  return <Loader2 className={cn('h-4 w-4 animate-spin text-ink-3', className)} />
}

export function FullPageSpinner() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-bg-0">
      <Spinner className="h-8 w-8 text-orange" />
    </div>
  )
}
