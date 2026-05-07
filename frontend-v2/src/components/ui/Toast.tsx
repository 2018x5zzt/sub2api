import { create } from 'zustand'
import { CheckCircle2, XCircle, Info, AlertTriangle, X } from 'lucide-react'
import { cn } from '@/lib/cn'

type Tone = 'success' | 'error' | 'info' | 'warning'

interface ToastItem {
  id: string
  tone: Tone
  message: string
  title?: string
  duration?: number
}

interface ToastState {
  toasts: ToastItem[]
  push: (t: Omit<ToastItem, 'id'>) => void
  dismiss: (id: string) => void
}

export const useToastStore = create<ToastState>((set, get) => ({
  toasts: [],
  push: (t) => {
    const id = Math.random().toString(36).slice(2)
    const toast: ToastItem = { duration: 4000, ...t, id }
    set((s) => ({ toasts: [...s.toasts, toast] }))
    if (toast.duration && toast.duration > 0) {
      setTimeout(() => get().dismiss(id), toast.duration)
    }
  },
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) }))
}))

export const toast = {
  success: (message: string, title?: string) =>
    useToastStore.getState().push({ tone: 'success', message, title }),
  error: (message: string, title?: string) =>
    useToastStore.getState().push({ tone: 'error', message, title }),
  info: (message: string, title?: string) =>
    useToastStore.getState().push({ tone: 'info', message, title }),
  warning: (message: string, title?: string) =>
    useToastStore.getState().push({ tone: 'warning', message, title })
}

const iconMap = {
  success: <CheckCircle2 className="h-5 w-5 text-signal-ok" />,
  error: <XCircle className="h-5 w-5 text-signal-err" />,
  info: <Info className="h-5 w-5 text-signal-info" />,
  warning: <AlertTriangle className="h-5 w-5 text-signal-warn" />
} as const

export function ToastViewport() {
  const toasts = useToastStore((s) => s.toasts)
  const dismiss = useToastStore((s) => s.dismiss)
  return (
    <div className="pointer-events-none fixed top-4 right-4 z-[100] flex w-full max-w-sm flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={cn(
            'pointer-events-auto card shadow-elevated p-3 flex items-start gap-3'
          )}
        >
          <div className="mt-0.5">{iconMap[t.tone]}</div>
          <div className="flex-1 min-w-0">
            {t.title && <div className="text-sm font-medium text-ink-1">{t.title}</div>}
            <div className="text-sm text-ink-2 break-words">{t.message}</div>
          </div>
          <button onClick={() => dismiss(t.id)} className="text-ink-3 hover:text-ink-1">
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
    </div>
  )
}
