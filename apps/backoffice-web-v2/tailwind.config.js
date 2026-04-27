/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'Consolas', 'monospace'],
      },
      borderRadius: {
        card:  '20px',
        input: '12px',
        btn:   '12px',
        chip:  '999px',
        inner: '12px',
      },
      colors: {
        surface: {
          body:     '#0b0b0d',
          panel:    '#131316',
          elevated: '#1c1b20',
          muted:    '#242229',
          hover:    '#2c2a33',
        },
        line: {
          subtle: 'rgba(255,255,255,0.07)',
          strong: 'rgba(255,255,255,0.13)',
        },
        ink: {
          primary:   '#F0EEE8',
          secondary: 'rgba(240,238,232,0.60)',
          tertiary:  'rgba(240,238,232,0.34)',
          muted:     'rgba(240,238,232,0.22)',
        },
        accent: {
          DEFAULT: '#C4A96A',
          hover:   '#D8BE88',
          soft:    'rgba(196,169,106,0.12)',
        },
        ok: {
          DEFAULT: '#6A9B7E',
          soft:    'rgba(106,155,126,0.14)',
        },
        warn: {
          DEFAULT: '#B8924E',
          soft:    'rgba(184,146,78,0.14)',
        },
        danger: {
          DEFAULT: '#B86B64',
          soft:    'rgba(184,107,100,0.14)',
        },
        neutral: {
          DEFAULT: '#74788A',
          soft:    'rgba(116,120,138,0.14)',
        },
      },
      boxShadow: {
        card:  '0 24px 64px rgba(0,0,0,0.36)',
        soft:  '0 8px 24px rgba(0,0,0,0.24)',
        inset: 'inset 0 1px 0 rgba(255,255,255,0.08)',
      },
      animation: {
        'pulse-dot': 'pulse-dot 2.4s ease-in-out infinite',
        shimmer:     'shimmer 1.4s linear infinite',
      },
      keyframes: {
        'pulse-dot': {
          '0%,100%': { opacity: '0.9' },
          '50%':     { opacity: '0.35' },
        },
        shimmer: {
          from: { transform: 'translateX(-100%)' },
          to:   { transform: 'translateX(100%)' },
        },
      },
    },
  },
  plugins: [],
}
