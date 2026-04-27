import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Card, CardTitle } from './ui/Card'

export function AuditPanel({ rows }) {
  const [open, setOpen] = useState(false)

  return (
    <Card className="p-4">
      <button
        onClick={() => setOpen(o => !o)}
        className="flex w-full items-center justify-between"
      >
        <CardTitle className="mb-0" info="Pricing decision audit trail. Shows exactly how the rate was calculated: source rate, margin applied, rule matched, adjustments. Use this for compliance review, explaining rates to clients, or debugging unexpected prices.">Why / Audit</CardTitle>
        <span className="text-[0.7rem] text-ink-muted transition-colors hover:text-ink-tertiary">
          {open ? '↑ collapse' : '↓ expand'}
        </span>
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.2, ease: [0.25, 1, 0.5, 1] }}
            className="overflow-hidden"
          >
            <div className="mt-3 flex flex-col gap-1.5">
              {rows.map((row) => (
                <div
                  key={row.label}
                  className="grid grid-cols-[minmax(100px,0.8fr)_2fr] items-start gap-3 rounded-inner glass-inner px-3 py-2"
                >
                  <span className="text-[0.65rem] font-semibold uppercase tracking-widest text-ink-muted">{row.label}</span>
                  <span className="break-all text-[0.82rem] text-ink-secondary">{row.value || '—'}</span>
                </div>
              ))}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </Card>
  )
}
