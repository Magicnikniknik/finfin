import { useEffect, useRef } from 'react'
import { useMotionValue, useSpring, useTransform, motion } from 'framer-motion'

// 21st.dev NumberTicker pattern — smooth spring counter
export function AnimatedNumber({ value, decimals = 0, className }) {
  const motionVal = useMotionValue(0)
  const spring    = useSpring(motionVal, { stiffness: 75, damping: 18, restDelta: 0.01 })
  const display   = useTransform(spring, (n) =>
    decimals > 0 ? n.toFixed(decimals) : Math.round(n).toString()
  )

  const prev = useRef(0)
  useEffect(() => {
    const num = typeof value === 'number' ? value : parseFloat(value)
    if (!Number.isNaN(num)) { motionVal.set(num); prev.current = num }
  }, [value, motionVal])

  return <motion.span className={className}>{display}</motion.span>
}
