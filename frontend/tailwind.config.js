/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主色调 —— Cyan / Sky 青色（参考 busapi --accent #7dd3fc）
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc', // busapi --accent
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
          950: '#082f49'
        },
        // 辅助色 —— Violet 紫色（参考 busapi --accent-2 #a78bfa）
        accent: {
          50: '#f5f3ff',
          100: '#ede9fe',
          200: '#ddd6fe',
          300: '#c4b5fd',
          400: '#a78bfa', // busapi --accent-2
          500: '#8b5cf6',
          600: '#7c3aed',
          700: '#6d28d9',
          800: '#5b21b6',
          900: '#4c1d95',
          950: '#2e1065'
        },
        // 深色模式色阶 —— 参考 busapi 的近黑灰阶
        // 50–500: 文本色（亮 → 暗）
        // 600–950: 背景/边框（边框 → 最深底色）
        dark: {
          50: '#f5f6f8', // text-1 最亮文本
          100: '#e1e3e8',
          200: '#c8cbd2', // text-2 主体文本
          300: '#a8acb5',
          400: '#8a8f99', // text-3 次要文本
          500: '#5b606b', // text-4 暗文本/图标
          600: '#2a2f38', // 强边框（hover）
          700: '#1d222c', // bg-4 输入框 / 中等边框
          800: '#11141a', // bg-2 弹窗 / 下拉
          850: '#0e1117',
          900: '#0c0e12', // bg-1 卡片 / 侧边栏
          950: '#07080a'  // bg-0 页面底色
        }
      },
      fontFamily: {
        sans: [
          'Inter',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'sans-serif'
        ],
        display: ['Inter', 'PingFang SC', 'system-ui', 'sans-serif'],
        mono: [
          'JetBrains Mono',
          'SF Mono',
          'ui-monospace',
          'SFMono-Regular',
          'Menlo',
          'Monaco',
          'Consolas',
          'monospace'
        ]
      },
      letterSpacing: {
        tightest: '-0.035em',
        'tight-2': '-0.02em',
        'tight-1': '-0.01em',
        eyebrow: '0.18em'
      },
      boxShadow: {
        // 暗色优先的阴影
        glass: '0 8px 32px rgba(0, 0, 0, 0.45)',
        'glass-sm': '0 4px 16px rgba(0, 0, 0, 0.35)',
        // 青色辉光（accent glow）
        glow: '0 0 0 1px rgba(125, 211, 252, 0.30), 0 8px 32px rgba(125, 211, 252, 0.18)',
        'glow-lg': '0 0 0 1px rgba(125, 211, 252, 0.30), 0 20px 60px rgba(125, 211, 252, 0.25)',
        // 卡片阴影：更克制
        card: '0 1px 2px rgba(0, 0, 0, 0.4)',
        'card-hover': '0 8px 24px rgba(0, 0, 0, 0.5)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.06)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        // 主色：青色溶解
        'gradient-primary':
          'linear-gradient(135deg, #7dd3fc 0%, #38bdf8 100%)',
        // 暗色面板：bg-2 → bg-1
        'gradient-dark': 'linear-gradient(180deg, #11141a 0%, #0c0e12 100%)',
        // 玻璃卡：略带高光
        'gradient-glass':
          'linear-gradient(180deg, rgba(20, 24, 32, 0.9), rgba(8, 10, 14, 0.9))',
        // 网状光晕（hero 背景）
        'mesh-gradient':
          'radial-gradient(at 30% 20%, rgba(125, 211, 252, 0.10) 0px, transparent 50%), radial-gradient(at 80% 0%, rgba(167, 139, 250, 0.08) 0px, transparent 55%), radial-gradient(at 0% 80%, rgba(125, 211, 252, 0.06) 0px, transparent 50%)',
        // 网格背景（busapi grid-bg）
        'grid-pattern':
          'linear-gradient(rgba(255, 255, 255, 0.06) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, 0.06) 1px, transparent 1px)',
        // 圆点背景
        'dot-pattern':
          'radial-gradient(rgba(255, 255, 255, 0.06) 1px, transparent 1px)'
      },
      backgroundSize: {
        grid: '56px 56px',
        dot: '24px 24px'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate',
        marquee: 'marquee 30s linear infinite'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': {
            boxShadow:
              '0 0 0 1px rgba(125, 211, 252, 0.3), 0 0 20px rgba(125, 211, 252, 0.18)'
          },
          '100%': {
            boxShadow:
              '0 0 0 1px rgba(125, 211, 252, 0.4), 0 0 30px rgba(125, 211, 252, 0.30)'
          }
        },
        marquee: {
          from: { transform: 'translate3d(0, 0, 0)' },
          to: { transform: 'translate3d(-50%, 0, 0)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
