import { cn } from '../../lib/utils'
import { InfoTooltip } from './Tooltip'

export function Card({ children, className, ...props }) {
  return (
    <div
      className={cn(
        'relative rounded-card glass',
        'transition-shadow duration-200',
        'hover:shadow-card',
        className,
      )}
      {...props}
    >
      {children}
    </div>
  )
}

export function CardTitle({ children, className, info, infoPosition }) {
  return (
    <h2 className={cn(
      'mb-4 flex items-center gap-1.5 text-[0.82rem] font-medium text-ink-tertiary',
      className,
    )}>
      {children}
      {info && <InfoTooltip text={info} position={infoPosition} />}
    </h2>
  )
}
