import { cn } from '../../lib/utils'

const labelCls = 'text-xs text-white/40 font-normal'

export function Input({ className, label, ...props }) {
  return (
    <label className="flex flex-col gap-1.5">
      {label && <span className={labelCls}>{label}</span>}
      <input
        className={cn(
          'w-full rounded-input border border-white/[0.08] bg-white/[0.05]',
          'px-3 py-[9px] text-[0.875rem] text-white/85 placeholder-white/20',
          'transition-[border-color,box-shadow] duration-150',
          'focus:border-[#0A84FF]/50 focus:outline-none focus:ring-2 focus:ring-[#0A84FF]/15',
          'read-only:cursor-default read-only:text-white/25',
          className,
        )}
        {...props}
      />
    </label>
  )
}

export function Select({ className, label, children, ...props }) {
  return (
    <label className="flex flex-col gap-1.5">
      {label && <span className={labelCls}>{label}</span>}
      <select
        className={cn(
          'w-full cursor-pointer appearance-none rounded-input border border-white/[0.08] bg-white/[0.05]',
          'px-3 py-[9px] text-[0.875rem] text-white/85',
          'transition-[border-color,box-shadow] duration-150',
          'focus:border-[#0A84FF]/50 focus:outline-none focus:ring-2 focus:ring-[#0A84FF]/15',
          className,
        )}
        {...props}
      >
        {children}
      </select>
    </label>
  )
}

export function SummaryRow({ label, value }) {
  return (
    <div className="flex flex-col gap-1 rounded-input bg-white/[0.03] px-3 py-2.5">
      <span className={labelCls}>{label}</span>
      <span className="break-all text-[0.875rem] text-white/80 tabular-nums">{value || '—'}</span>
    </div>
  )
}
