import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react'

type TurnstileTheme = 'light' | 'dark' | 'auto'
type TurnstileSize = 'normal' | 'compact' | 'flexible'

interface TurnstileRenderOptions {
  sitekey: string
  callback: (token: string) => void
  'expired-callback'?: () => void
  'error-callback'?: () => void
  theme?: TurnstileTheme
  size?: TurnstileSize
}

interface TurnstileAPI {
  render: (container: HTMLElement, options: TurnstileRenderOptions) => string
  reset: (widgetId?: string) => void
  remove: (widgetId?: string) => void
}

declare global {
  interface Window {
    turnstile?: TurnstileAPI
    onTurnstileLoad?: () => void
  }
}

export interface TurnstileWidgetHandle {
  reset: () => void
}

interface TurnstileWidgetProps {
  siteKey: string
  theme?: TurnstileTheme
  size?: TurnstileSize
  onVerify: (token: string) => void
  onExpire?: () => void
  onError?: () => void
}

function loadTurnstileScript(): Promise<void> {
  return new Promise((resolve, reject) => {
    if (window.turnstile) {
      resolve()
      return
    }

    const existingScript = document.querySelector('script[src*="turnstile"]')
    if (existingScript) {
      window.onTurnstileLoad = () => resolve()
      return
    }

    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?onload=onTurnstileLoad'
    script.async = true
    script.defer = true
    window.onTurnstileLoad = () => resolve()
    script.onerror = () => reject(new Error('Failed to load Turnstile script'))
    document.head.appendChild(script)
  })
}

export const TurnstileWidget = forwardRef<TurnstileWidgetHandle, TurnstileWidgetProps>(
  function TurnstileWidget(
    { siteKey, theme = 'auto', size = 'flexible', onVerify, onExpire, onError },
    ref
  ) {
    const containerRef = useRef<HTMLDivElement | null>(null)
    const widgetIdRef = useRef<string | null>(null)
    const [scriptLoaded, setScriptLoaded] = useState(false)

    useImperativeHandle(ref, () => ({
      reset: () => {
        if (window.turnstile && widgetIdRef.current) {
          window.turnstile.reset(widgetIdRef.current)
        }
      }
    }))

    useEffect(() => {
      if (!siteKey) return
      let cancelled = false

      loadTurnstileScript()
        .then(() => {
          if (!cancelled) setScriptLoaded(true)
        })
        .catch(() => {
          if (!cancelled) onError?.()
        })

      return () => {
        cancelled = true
      }
    }, [onError, siteKey])

    useEffect(() => {
      if (!scriptLoaded || !window.turnstile || !containerRef.current || !siteKey) return

      if (widgetIdRef.current) {
        try {
          window.turnstile.remove(widgetIdRef.current)
        } catch {
          // Ignore stale widget cleanup errors.
        }
        widgetIdRef.current = null
      }

      containerRef.current.innerHTML = ''
      widgetIdRef.current = window.turnstile.render(containerRef.current, {
        sitekey: siteKey,
        callback: onVerify,
        'expired-callback': onExpire,
        'error-callback': onError,
        theme,
        size
      })

      return () => {
        if (window.turnstile && widgetIdRef.current) {
          try {
            window.turnstile.remove(widgetIdRef.current)
          } catch {
            // Ignore cleanup errors from an already-removed widget.
          }
          widgetIdRef.current = null
        }
      }
    }, [onError, onExpire, onVerify, scriptLoaded, siteKey, size, theme])

    if (!siteKey) return null
    return (
      <div className="w-full">
        <div ref={containerRef} className="min-h-[65px] w-full [&_iframe]:!w-full" />
      </div>
    )
  }
)

