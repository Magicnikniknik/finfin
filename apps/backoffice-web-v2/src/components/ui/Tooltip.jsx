import { cn } from '../../lib/utils'

export function InfoTooltip({ text, position = 'above' }) {
  return (
    <div className="group relative inline-flex shrink-0">
      <button
        type="button"
        tabIndex={-1}
        className={cn(
          'flex h-[16px] w-[16px] items-center justify-center rounded-full',
          'border border-line-subtle text-[0.58rem] font-semibold text-ink-muted',
          'transition-colors duration-150',
          'hover:border-line-strong hover:text-ink-tertiary',
        )}
      >
        ?
      </button>

      {/* Tooltip bubble */}
      <div className={cn(
        'pointer-events-none absolute z-50 w-56',
        'rounded-card glass-elevated px-3 py-2.5',
        'text-[0.75rem] leading-[1.5] text-ink-secondary',
        'opacity-0 transition-opacity duration-150 group-hover:opacity-100',
        'left-1/2 -translate-x-1/2',
        position === 'above'
          ? 'bottom-full mb-2'
          : 'top-full mt-2',
      )}>
        {text}
        {/* Arrow */}
        <span className={cn(
          'absolute left-1/2 -translate-x-1/2 border-4 border-transparent',
          position === 'above'
            ? 'top-full border-t-[#1c1b20]'
            : 'bottom-full border-b-[#1c1b20]',
        )} />
      </div>
    </div>
  )
}
