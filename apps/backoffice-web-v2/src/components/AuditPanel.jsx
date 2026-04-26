import { motion, AnimatePresence } from 'framer-motion'
import { Card, CardTitle } from './ui/Card'

export function AuditPanel({ rows }) {
  return (
    <Card className="p-5">
      <CardTitle>Why / Audit</CardTitle>
      <div className="flex flex-col gap-1.5">
        <AnimatePresence mode="sync">
          {rows.map((row) => (
            <motion.div
              key={row.label}
              initial={{ opacity: 0, x: -8 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.2, ease: [0.25, 1, 0.5, 1] }}
              className="grid grid-cols-[minmax(90px,0.9fr)_2fr] items-center gap-3 rounded-lg border border-white/[0.05] bg-[rgba(9,14,26,0.5)] px-3 py-2"
            >
              <span className="text-[0.65rem] font-semibold uppercase tracking-widest text-slate-600">{row.label}</span>
              <span className="break-all text-sm text-slate-400">{row.value || '—'}</span>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>
    </Card>
  )
}
