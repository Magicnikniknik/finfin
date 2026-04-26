import { motion } from 'framer-motion'
import { cn } from '../../lib/utils'

/**
 * Apple-style buttons.
 * - Flat fill, NO gradients, NO shimmer sweep
 * - 10px radius (btn)
 * - hover = lighter fill via opacity
 * - active = scale 0.97 via Framer
 */
const S = {
  primary:   'bg-[#0A84FF] text-white hover:bg-[#409CFF] active:bg-[#0A84FF]',
  secondary: 'bg-white/[0.07] text-white/75 border border-white/[0.09] hover:bg-white/[0.11] hover:text-white/90',
  danger:    'bg-[#FF453A]/10 text-[#FF453A] border border-[#FF453A]/15 hover:bg-[#FF453A]/18',
  success:   'bg-[#32D74B]/10 text-[#32D74B] border border-[#32D74B]/15 hover:bg-[#32D74B]/18',
}

export function Button({ children, variant = 'primary', className, disabled, type = 'button', onClick }) {
  return (
    <motion.button
      type={type}
      whileTap={{ scale: disabled ? 1 : 0.97 }}
      transition={{ type: 'spring', stiffness: 500, damping: 30 }}
      disabled={disabled}
      onClick={onClick}
      className={cn(
        'inline-flex cursor-pointer select-none items-center justify-center rounded-btn',
        'px-3.5 py-[7px] text-[0.82rem] font-medium tracking-[-0.01em]',
        'transition-colors duration-150',
        'disabled:pointer-events-none disabled:opacity-40',
        S[variant] ?? S.primary,
        className,
      )}
    >
      {children}
    </motion.button>
  )
}
