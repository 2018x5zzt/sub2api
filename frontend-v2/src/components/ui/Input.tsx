import { forwardRef, type InputHTMLAttributes, type ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
  hint?: string
  leftIcon?: ReactNode
  rightAdornment?: ReactNode
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { label, error, hint, leftIcon, rightAdornment, className, id, ...rest },
  ref
) {
  const inputId = id || rest.name
  return (
    <div>
      {label && (
        <label htmlFor={inputId} className="input-label">
          {label}
        </label>
      )}
      <div className="relative">
        {leftIcon && (
          <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5 text-ink-3">
            {leftIcon}
          </div>
        )}
        <input
          ref={ref}
          id={inputId}
          className={cn(
            'input',
            leftIcon && 'pl-10',
            rightAdornment && 'pr-10',
            error && 'input-error',
            className
          )}
          {...rest}
        />
        {rightAdornment && (
          <div className="absolute inset-y-0 right-0 flex items-center pr-2.5">{rightAdornment}</div>
        )}
      </div>
      {error ? (
        <p className="input-error-text">{error}</p>
      ) : hint ? (
        <p className="text-xs mt-1 text-ink-3">{hint}</p>
      ) : null}
    </div>
  )
})
