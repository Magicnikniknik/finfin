import { motion } from 'framer-motion'
import { AnimatedNumber } from './ui/AnimatedNumber'

const META = [
  { key: 'activeQuotes',   label: 'Active quotes',   color: '#0A84FF' },
  { key: 'reservedOrders', label: 'Reserved',         color: '#64D2FF' },
  { key: 'completedToday', label: 'Completed',        color: '#32D74B' },
  { key: 'cancelledToday', label: 'Cancelled',        color: '#FFD60A' },
  { key: 'grossMargin',    label: 'Est. margin',      color: '#BF5AF2', dec: 2 },
  { key: 'staleAlerts',    label: 'Stale alerts',     color: '#FF453A' },
]

function KpiCard({ meta, value, i }) {
  return (
    <motion.article
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, delay: 0.06 * i, ease: [0.25, 1, 0.5, 1] }}
      className="group relative rounded-card glass p-4 transition-[border-color] duration-200 hover:border-white/[0.13]"
    >
      {/* Thin left accent line — subtle, not a stripe */}
      <div
        className="absolute inset-y-3 left-0 w-[2px] rounded-full opacity-50"
        style={{ background: meta.color }}
      />
      <p className="mb-2.5 pl-3 text-[0.62rem] font-semibold uppercase tracking-[0.09em] text-white/30">
        {meta.label}
      </p>
      <p className="pl-3 text-[1.6rem] font-semibold leading-none tracking-[-0.03em] text-white/90 tabular-nums">
        <AnimatedNumber value={typeof value === 'number' ? value : 0} decimals={meta.dec ?? 0} />
      </p>
    </motion.article>
  )
}

export function KpiBar({ metrics }) {
  return (
    <div className="mb-4 grid grid-cols-6 gap-2 max-[1100px]:grid-cols-3 max-[600px]:grid-cols-2">
      {META.map((m, i) => (
        <KpiCard key={m.key} meta={m} value={metrics[m.key] ?? 0} i={i} />
      ))}
    </div>
  )
}
