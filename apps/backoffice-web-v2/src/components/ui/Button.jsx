import { motion } from 'framer-motion'
import { cn } from '../../lib/utils'

// primary  — Sign In style: neutral elevated dark, the default action weight
// accent   — champagne tint: for the single most important CTA on a page
// secondary — quiet surface ghost
// danger    — muted clay red for destructive actions
// ghost     — near-invisible for utility/toggle actions
const S = {
  primary:   'bg-surface-elevated border border-line-strong shadow-inset text-ink-primary hover:bg-surface-muted hover:border-line-strong',
  accent:    'bg-accent/90 text-surface-body border border-accent/0 hover:bg-accent',
  secondary: 'bg-surface-panel border border-line-subtle text-ink-secondary hover:border-line-strong hover:text-ink-primary',
  danger:    'bg-danger/[0.08] text-danger border border-danger/[0.14] hover:bg-danger/[0.14]',
  ghost:     'bg-transparent border border-transparent text-ink-tertiary hover:text-ink-secondary hover:bg-surface-elevated',
}

export function Button({ children, variant = 'secondary', className, disabled, type = 'button', onClick }) {
  return (
    <motion.button
      type={type}
      whileTap={{ scale: disabled ? 1 : 0.97 }}
      transition={{ type: 'spring', stiffness: 500, damping: 30 }}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'inline-flex shrink-0 cursor-pointer select-none items-center justify-center rounded-btn',
        'px-3.5 py-[7px] text-[0.82rem] font-medium tracking-[-0.01em]',
        'transition-colors duration-150',
        'disabled:pointer-events-none disabled:opacity-40',
        S[variant] ?? S.secondary,
        className,
      )}
    >
      {children}
    </motion.button>
  )
}
