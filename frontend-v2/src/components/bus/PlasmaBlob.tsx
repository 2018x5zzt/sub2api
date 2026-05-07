import type { CSSProperties } from 'react'

/** Warm orange plasma glow used in hero / case-study artwork. */
export function PlasmaBlob({ style }: { style?: CSSProperties }) {
  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        ...style,
        background: `
          radial-gradient(ellipse 45% 35% at 62% 55%, rgba(255,140,60,0.85), rgba(255,87,34,0.55) 25%, rgba(180,40,10,0.3) 45%, transparent 65%),
          radial-gradient(ellipse 30% 25% at 70% 45%, rgba(255,200,120,0.6), transparent 50%),
          radial-gradient(ellipse 25% 20% at 50% 65%, rgba(255,87,34,0.4), transparent 55%),
          radial-gradient(circle at 75% 60%, rgba(255,170,80,0.3), transparent 40%)
        `
      }}
    />
  )
}

/** Multiplied dot-pattern overlay that sits on top of the plasma blob. */
export function HalftoneOverlay({ opacity = 0.6 }: { opacity?: number }) {
  return (
    <svg
      className="absolute inset-0 w-full h-full pointer-events-none"
      style={{ opacity, mixBlendMode: 'multiply' }}
      aria-hidden
    >
      <defs>
        <pattern id="halftone-d" x="0" y="0" width="3" height="3" patternUnits="userSpaceOnUse">
          <circle cx="1.5" cy="1.5" r="0.7" fill="#000" />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#halftone-d)" />
    </svg>
  )
}
