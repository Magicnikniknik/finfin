import { motion } from 'framer-motion'

export function Header() {
  return (
    <motion.header
      initial={{ opacity: 0, y: -14 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: [0.25, 1, 0.5, 1] }}
      className="mb-4 flex items-center justify-between gap-4 px-1 py-2"
    >
      <div>
        <h1 className="text-[1.35rem] font-semibold tracking-[-0.035em] text-ink-primary">
          Backoffice
        </h1>
        <p className="mt-0.5 text-[0.8rem] text-ink-tertiary">
          Scenario operations console · pricing, reserve, order lifecycle
        </p>
      </div>

      <div className="flex items-center gap-2 rounded-chip border border-line-subtle bg-surface-panel px-3 py-1.5">
        <span className="h-[6px] w-[6px] rounded-full bg-ok animate-pulse-dot" />
        <span className="text-[0.7rem] font-medium tracking-[0.08em] text-ink-tertiary uppercase">
          Sandbox
        </span>
      </div>
    </motion.header>
  )
}
