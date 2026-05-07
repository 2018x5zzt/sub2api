import type { Config } from 'tailwindcss'

const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // BusAPI Plato — dark enterprise tech
        bg: {
          0: '#000000',  // page bg
          1: '#0a0a0a',  // surface
          2: '#0f0f10',  // elevated
          3: '#15161a'   // hover / row
        },
        line: {
          1: 'rgba(255,255,255,0.06)',
          2: 'rgba(255,255,255,0.10)',
          3: 'rgba(255,255,255,0.13)',
          4: 'rgba(255,255,255,0.16)'
        },
        ink: {
          0: '#ffffff',
          1: '#f5f6f8',
          2: '#c8cbd2',
          3: '#8a8f99',
          4: '#5b606b'
        },
        // Single accent: warm orange used across all CTAs/highlights
        orange: {
          DEFAULT: '#ff5722',
          hover: '#ff693a',
          soft: 'rgba(255,87,34,0.12)',
          softer: 'rgba(255,87,34,0.06)',
          line: 'rgba(255,87,34,0.30)'
        },
        // Status hues sampled from console-v4
        signal: {
          ok: '#34d399',
          warn: '#fbbf24',
          err: '#f87171',
          info: '#a78bfa'
        }
      },
      fontFamily: {
        sans: [
          'Inter',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'system-ui',
          'sans-serif'
        ],
        display: ['Inter', 'PingFang SC', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"SF Mono"', 'ui-monospace', 'Menlo', 'monospace'],
        serif: ['Georgia', '"Times New Roman"', 'serif']
      },
      fontSize: {
        // Display headings
        'display-2xl': ['72px', { lineHeight: '1.02', letterSpacing: '-0.035em', fontWeight: '500' }],
        'display-xl':  ['56px', { lineHeight: '1.05', letterSpacing: '-0.030em', fontWeight: '500' }],
        'display-lg':  ['48px', { lineHeight: '1.06', letterSpacing: '-0.025em', fontWeight: '500' }],
        'display-md':  ['40px', { lineHeight: '1.10', letterSpacing: '-0.025em', fontWeight: '500' }],
        // Eyebrow / mono caption
        eyebrow: ['11px', { lineHeight: '1.4', letterSpacing: '0.18em' }]
      },
      borderRadius: {
        none: '0',
        xs: '2px',
        sm: '4px',
        DEFAULT: '6px',
        md: '8px',
        lg: '10px',
        xl: '14px',
        '2xl': '20px',
        full: '9999px'
      },
      boxShadow: {
        soft: '0 1px 2px rgba(0,0,0,0.4)',
        card: '0 8px 24px rgba(0,0,0,0.4)',
        elevated: '0 30px 80px rgba(0,0,0,0.5)',
        glow: '0 0 0 1px rgba(255,87,34,0.30), 0 8px 32px rgba(255,87,34,0.18)'
      },
      keyframes: {
        pulseDot: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.4' }
        }
      },
      animation: {
        pulseDot: 'pulseDot 1.6s ease-in-out infinite'
      }
    }
  },
  plugins: []
}

export default config
