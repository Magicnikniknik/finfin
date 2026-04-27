import { AnimatedNumber } from './ui/AnimatedNumber'
import { InfoTooltip } from './ui/Tooltip'

const META = [
  { key: 'activeQuotes',   label: 'Active quotes', color: '#B89B5E', info: 'Quotes currently live and not yet reserved or expired. High count may indicate clients are comparing rates but not converting.' },
  { key: 'reservedOrders', label: 'Reserved',       color: '#8AABB0', info: 'Orders in RESERVED state — funds are earmarked, awaiting Complete or Cancel. Ageing reserves may indicate a stuck flow.' },
  { key: 'completedToday', label: 'Completed',       color: '#5F8F72', info: 'Orders successfully completed today. The primary success metric — this is revenue-generating activity.' },
  { key: 'cancelledToday', label: 'Cancelled',       color: '#6E7683', info: 'Orders cancelled today. Some cancellations are normal; a spike relative to completed may signal a pricing or liquidity issue.' },
  { key: 'grossMargin',    label: 'Est. margin',    color: '#B08A4A', dec: 2, info: 'Estimated gross margin on completed orders today. Calculated from the spread between base rate and client rate minus fees.' },
  { key: 'staleAlerts',    label: 'Stale alerts',   color: '#B46A63', info: 'Quotes or reserves that have exceeded their TTL without being acted on. Should trend toward zero in a healthy system.' },
]

function KpiCard({ meta, value }) {
  return (
    <article className="group relative rounded-card glass p-4 transition-shadow duration-200 hover:shadow-soft">
      <div
        className="absolute inset-y-3 left-0 w-[2px] rounded-full opacity-40"
        style={{ background: meta.color }}
      />
      <div className="mb-2.5 flex items-center justify-between pl-3 pr-1">
        <p className="text-[0.62rem] font-semibold uppercase tracking-[0.09em] text-ink-muted">
          {meta.label}
        </p>
        <span className="opacity-0 transition-opacity duration-150 group-hover:opacity-100">
          <InfoTooltip text={meta.info} position="below" />
        </span>
      </div>
      <p className="pl-3 text-[1.6rem] font-semibold leading-none tracking-[-0.03em] text-ink-primary tabular-nums">
        <AnimatedNumber value={typeof value === 'number' ? value : 0} decimals={meta.dec ?? 0} />
      </p>
    </article>
  )
}

export function KpiBar({ metrics }) {
  return (
    <div className="grid grid-cols-6 gap-2 max-[1100px]:grid-cols-3 max-[600px]:grid-cols-2">
      {META.map((m) => (
        <KpiCard key={m.key} meta={m} value={metrics[m.key] ?? 0} />
      ))}
    </div>
  )
}
