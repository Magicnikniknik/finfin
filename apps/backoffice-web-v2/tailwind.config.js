/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'Consolas', 'monospace'],
      },
      borderRadius: {
        card:   '14px',
        input:  '9px',
        btn:    '10px',
        chip:   '6px',
      },
      colors: {
        // Apple system colors for dark mode
        accent: {
          DEFAULT: '#0A84FF',
          hover:   '#409CFF',
        },
        ok:     '#32D74B',
        warn:   '#FFD60A',
        danger: '#FF453A',
      },
      animation: {
        'pulse-dot': 'pulse-dot 2.2s ease-in-out infinite',
        'shimmer':   'shimmer 1.2s linear infinite',
      },
      keyframes: {
        'pulse-dot': { '0%,100%': { opacity: '1' }, '50%': { opacity: '0.3' } },
        'shimmer':   { from: { transform: 'translateX(-100%)' }, to: { transform: 'translateX(100%)' } },
      },
    },
  },
  plugins: [],
}



