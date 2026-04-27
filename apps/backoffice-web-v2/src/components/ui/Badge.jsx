import { motion, AnimatePresence } from 'framer-motion'
import { cn } from '../../lib/utils'

const V = {
  ok:      { dot: 'bg-ok',      ring: 'border-ok/20 bg-ok/[0.10] text-ok',          pulse: true  },
  warn:    { dot: 'bg-warn',    ring: 'border-warn/20 bg-warn/[0.10] text-warn',      pulse: false },
  err:     { dot: 'bg-danger',  ring: 'border-danger/20 bg-danger/[0.10] text-danger', pulse: false },
  neutral: { dot: 'bg-neutral', ring: 'border-line-subtle bg-surface-elevated text-ink-tertiary', pulse: false },
}

export function Badge({ variant = 'neutral', children, className }) {
  const v = V[variant] ?? V.neutral
  return (
    <AnimatePresence mode="wait">
      <motion.span
        key={variant + String(children)}
        initial={{ opacity: 0, scale: 0.90 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.90 }}
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
