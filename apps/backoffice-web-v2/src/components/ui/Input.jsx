import { cn } from '../../lib/utils'

const labelCls = 'text-xs text-ink-tertiary font-normal'

export function Input({ className, label, ...props }) {
  return (
    <label className="flex flex-col gap-1.5">
      {label && <span className={labelCls}>{label}</span>}
      <input
        className={cn(
          'w-full rounded-input border border-line-subtle bg-surface-elevated',
          'px-3 py-[9px] text-[0.875rem] text-ink-primary placeholder-ink-muted',
          'transition-[border-color,box-shadow] duration-150',
          'focus:border-accent/40 focus:outline-none focus:ring-2 focus:ring-accent/10',
          'read-only:cursor-default read-only:text-ink-muted',
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
          'w-full cursor-pointer appearance-none rounded-input border border-line-subtle bg-surface-elevated',
          'px-3 py-[9px] text-[0.875rem] text-ink-primary',
          'transition-[border-color,box-shadow] duration-150',
          'focus:border-accent/40 focus:outline-none focus:ring-2 focus:ring-accent/10',
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
    <div className="flex flex-col gap-1 rounded-inner glass-inner px-3 py-2.5">
      <span className={labelCls}>{label}</span>
      <span className="break-all text-[0.875rem] text-ink-primary tabular-nums">{value || '—'}</span>
    </div>
  )
}
