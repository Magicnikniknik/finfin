import { motion, AnimatePresence } from 'framer-motion'
import { Card, CardTitle } from './ui/Card'
import { cn } from '../lib/utils'

const DOT = {
  ok:      'bg-[#32D74B]',
  warn:    'bg-[#FFD60A]',
  err:     'bg-[#FF453A]',
  neutral: 'bg-white/20',
}
const ROW = {
  ok:      'bg-[#32D74B]/[0.06] text-[#32D74B]/80',
  warn:    'bg-[#FFD60A]/[0.06] text-[#FFD60A]/80',
  err:     'bg-[#FF453A]/[0.06] text-[#FF453A]/80',
  neutral: 'bg-white/[0.03] text-white/35',
}

export function StateTimeline({ steps }) {
  return (
    <Card className="p-5">
      <CardTitle>State Timeline</CardTitle>
      <ol className="flex flex-col gap-1.5">
        <AnimatePresence mode="popLayout">
          {steps.map((step) => (
            <motion.li
              key={step.key}
              layout
              initial={{ opacity: 0, x: -8 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.22, ease: [0.25, 1, 0.5, 1] }}
              className={cn(
                'flex items-center gap-2.5 rounded-input px-3 py-2 text-[0.8rem]',
                'transition-colors duration-300',
                ROW[step.variant] ?? ROW.neutral,
              )}
            >
              <span className={cn('h-[6px] w-[6px] shrink-0 rounded-full', DOT[step.variant] ?? DOT.neutral)} />
              {step.label}
            </motion.li>
          ))}
        </AnimatePresence>
      </ol>
    </Card>
  )
}
