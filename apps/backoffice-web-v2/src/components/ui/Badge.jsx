import { motion, AnimatePresence } from 'framer-motion'
import { cn } from '../../lib/utils'

// Apple system colors (dark mode)
const V = {
  ok:      { dot: 'bg-[#32D74B]', ring: 'border-[#32D74B]/20 bg-[#32D74B]/10 text-[#32D74B]', pulse: true  },
  warn:    { dot: 'bg-[#FFD60A]', ring: 'border-[#FFD60A]/20 bg-[#FFD60A]/10 text-[#FFD60A]', pulse: false },
  err:     { dot: 'bg-[#FF453A]', ring: 'border-[#FF453A]/20 bg-[#FF453A]/10 text-[#FF453A]', pulse: false },
  neutral: { dot: 'bg-white/20',  ring: 'border-white/[0.08] bg-white/[0.05] text-white/40',   pulse: false },
}

export function Badge({ variant = 'neutral', children, className }) {
  const v = V[variant] ?? V.neutral
  return (
    <AnimatePresence mode="wait">
      <motion.span
        key={variant + String(children)}
        initial={{ opacity: 0, scale: 0.88 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.88 }}
        transition={{ type: 'spring', stiffness: 400, damping: 28 }}
        className={cn(
          'inline-flex items-center gap-1.5 rounded-chip border px-2.5 py-[3px]',
          'text-[0.7rem] font-medium tracking-wide',
          v.ring, className,
        )}
      >
        <span className={cn('h-[5px] w-[5px] shrink-0 rounded-full', v.dot, v.pulse && 'animate-pulse-dot')} />
        {children}
      </motion.span>
    </AnimatePresence>
  )
}
