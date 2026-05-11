import { useState, useRef, useEffect } from 'react'
import { Globe } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { availableLocales, setLocale, getLocale, type LocaleCode } from '@/i18n'
import { cn } from '@/lib/cn'

export function LocaleSwitcher({ compact }: { compact?: boolean }) {
  const { i18n, t } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const current = getLocale()

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    window.addEventListener('mousedown', onClick)
    return () => window.removeEventListener('mousedown', onClick)
  }, [])

  return (
    <div className="relative" ref={ref}>
      <button
        className={cn('btn btn-ghost', compact ? 'btn-icon' : '')}
        onClick={() => setOpen((v) => !v)}
        aria-label={t('accessibility.switchLanguage') as string}
      >
        <Globe className="h-4 w-4" />
        {!compact && (
          <span className="text-sm">
            {availableLocales.find((l) => l.code === current)?.flag}{' '}
            {availableLocales.find((l) => l.code === current)?.name}
          </span>
        )}
      </button>
      {open && (
        <div className="absolute right-0 mt-2 w-40 card shadow-elevated overflow-hidden z-30">
          {availableLocales.map((l) => (
            <button
              key={l.code}
              onClick={async () => {
                await setLocale(l.code as LocaleCode)
                i18n.changeLanguage(l.code)
                setOpen(false)
              }}
              className={cn(
                'flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-bg-3 transition-colors',
                l.code === current && 'bg-bg-3 font-medium'
              )}
            >
              <span>{l.flag}</span>
              <span>{l.name}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
