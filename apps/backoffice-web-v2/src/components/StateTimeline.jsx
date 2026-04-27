import { motion, AnimatePresence } from 'framer-motion'
import { Card, CardTitle } from './ui/Card'
import { cn } from '../lib/utils'

const DOT = {
  ok:      'bg-ok',
  warn:    'bg-warn',
  err:     'bg-danger',
  neutral: 'bg-neutral',
}
const CHIP = {
  ok:      'bg-ok/[0.08] text-ok border-ok/20',
  warn:    'bg-warn/[0.08] text-warn border-warn/20',
  err:     'bg-danger/[0.08] text-danger border-danger/20',
  neutral: 'bg-surface-elevated text-ink-tertiary border-line-subtle',
}

export function StateTimeline({ steps }) {
  return (
    <Card className="p-4">
      <CardTitle info="Live order lifecycle — read left to right. Each chip is a state the order passed through. Green = success, amber = warning, red = error. This is the key view for demonstrating system correctness to investors and auditors.">State Timeline</CardTitle>
      {steps.length === 0 ? (
        <p className="text-[0.78rem] text-ink-muted">No events yet. Run a scenario to see the order lifecycle.</p>
      ) : (
        <div className="flex flex-wrap items-center gap-1.5">
          <AnimatePresence mode="popLayout">
            {steps.map((step, i) => (
              <motion.div
                key={step.key}
                layout
                initial={{ opacity: 0, scale: 0.85 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.2, ease: [0.25, 1, 0.5, 1] }}
                className="flex items-center gap-1.5"
              >
                <span className={cn(
                  'flex items-center gap-1.5 rounded-chip border px-2.5 py-1 text-[0.75rem] font-medium',
                  CHIP[step.variant] ?? CHIP.neutral,
                )}>
                  <span className={cn('h-[5px] w-[5px] shrink-0 rounded-full', DOT[step.variant] ?? DOT.neutral)} />
                  {step.label}
                </span>
                {i < steps.length - 1 && (
                  <span className="text-[0.65rem] text-ink-muted">→</span>
                )}
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      )}
    </Card>
  )
}
