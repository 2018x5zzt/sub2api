import type { Config } from 'tailwindcss'

const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Keep legacy class names as aliases while moving the site to a Claude-light palette.
        bg: {
          0: '#FAF9F5',
          1: '#FFFFFF',
          2: '#F6F2EA',
          3: '#F0EEE6',
          4: '#FFFFFF'
        },
        line: {
          1: '#ECEAE0',
          2: '#E8E6DC',
          3: '#D8D4C6',
          4: '#C4BFAE'
        },
        ink: {
          0: '#FFFFFF',
          1: '#1F1E1D',
          2: '#3D3D3A',
          3: '#8A8780',
          4: '#B7B3A8'
        },
        orange: {
          DEFAULT: '#C0360B',
          display: '#FF5722',
          hover: '#E84713',
          soft: '#FFE5DD',
          softer: '#FFF2EE',
          line: '#FFB7A0'
        },
        signal: {
          ok: '#2F8F5E',
          warn: '#A8761A',
          err: '#B3261E',
          info: '#5F5A9F'
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
        display: ['"Source Serif 4"', '"Source Serif Pro"', 'Newsreader', 'Georgia', 'serif'],
        mono: ['"JetBrains Mono"', '"SF Mono"', 'ui-monospace', 'Menlo', 'monospace'],
        serif: ['"Source Serif 4"', '"Source Serif Pro"', 'Newsreader', 'Georgia', 'serif']
      },
      fontSize: {
        'display-2xl': ['72px', { lineHeight: '1.02', letterSpacing: '-0.035em', fontWeight: '500' }],
        'display-xl': ['56px', { lineHeight: '1.05', letterSpacing: '-0.030em', fontWeight: '500' }],
        'display-lg': ['48px', { lineHeight: '1.06', letterSpacing: '-0.025em', fontWeight: '500' }],
        'display-md': ['40px', { lineHeight: '1.10', letterSpacing: '-0.025em', fontWeight: '500' }],
        'display-sm': ['32px', { lineHeight: '1.15', letterSpacing: '-0.020em', fontWeight: '500' }],
        eyebrow: ['11px', { lineHeight: '1.4', letterSpacing: '0.18em' }],
        caption: ['12.5px', { lineHeight: '1.5' }]
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
        soft: '0 1px 2px rgba(31,30,29,0.04)',
        card: '0 1px 2px rgba(31,30,29,0.04), 0 4px 16px rgba(31,30,29,0.06)',
        elevated: '0 2px 4px rgba(31,30,29,0.06), 0 12px 32px rgba(31,30,29,0.08)',
        glow: '0 0 0 3px rgba(255,87,34,0.20)'
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
