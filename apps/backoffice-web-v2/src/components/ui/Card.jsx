import { useRef, useCallback } from 'react'
import { motion } from 'framer-motion'
import { cn } from '../../lib/utils'

/**
 * Apple Liquid Glass card.
 * - True black bg, 5% white surface
 * - 14px radius (Apple card standard)
 * - Spotlight follows cursor via CSS var
 * - Concentric: inner elements should use r=9px
 */
export function Card({ children, className, ...props }) {
  const ref = useRef(null)

  const onMouseMove = useCallback((e) => {
    const r = ref.current?.getBoundingClientRect()
    if (!r) return
    ref.current.style.setProperty('--mx', `${e.clientX - r.left}px`)
    ref.current.style.setProperty('--my', `${e.clientY - r.top}px`)
  }, [])

  return (
    <motion.div
      ref={ref}
      onMouseMove={onMouseMove}
      className={cn(
        'relative overflow-hidden rounded-card glass',
        'transition-[border-color] duration-200',
        'hover:border-white/[0.13]',
        className,
      )}
      style={{ '--mx': '50%', '--my': '50%' }}
      {...props}
    >
      {/* Spotlight — no glow, just subtle warmth */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 rounded-card opacity-0 transition-opacity duration-500 hover:opacity-100"
        style={{ background: 'radial-gradient(240px circle at var(--mx) var(--my), rgba(255,255,255,0.03), transparent 70%)' }}
      />
      <div className="relative z-10">{children}</div>
    </motion.div>
  )
}

export function CardTitle({ children, className }) {
  return (
    <h2 className={cn(
      'mb-4 text-[0.82rem] font-medium text-white/50',
      className,
    )}>
      {children}
    </h2>
  )
}
