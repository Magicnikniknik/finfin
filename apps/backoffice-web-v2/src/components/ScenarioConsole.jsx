import { motion } from 'framer-motion'
import { DEMO_SCENARIOS } from '../lib/scenarios'
import { Card, CardTitle } from './ui/Card'
import { Button } from './ui/Button'
import { SummaryRow } from './ui/Input'

export function ScenarioConsole({ currentId, onSelect, onLoad, onRunQuote, onRunReserve, onRunFull }) {
  const scenario = DEMO_SCENARIOS.find((s) => s.id === currentId) ?? DEMO_SCENARIOS[0]

  return (
    <Card className="p-4">
      <CardTitle info="Select a pre-built scenario — it loads realistic parameters into the quote form. Run Full Flow executes the entire lifecycle automatically. Watch the State Timeline below to see each step as it happens.">Scenario Console</CardTitle>

      {/* Chips */}
      <div className="mb-4 flex flex-wrap gap-1.5">
        {DEMO_SCENARIOS.map((s) => {
          const active = s.id === currentId
          return (
            <motion.button
              key={s.id}
              whileTap={{ scale: 0.95 }}
              transition={{ type: 'spring', stiffness: 500, damping: 30 }}
              onClick={() => onSelect(s.id)}
              className={[
                'cursor-pointer rounded-chip px-3 py-1 text-[0.75rem] font-medium transition-all duration-150',
                active
                  ? 'bg-accent/12 text-accent ring-1 ring-accent/25'
                  : 'bg-surface-elevated text-ink-tertiary hover:bg-surface-muted hover:text-ink-secondary',
              ].join(' ')}
            >
              {s.title}
            </motion.button>
          )
        })}
      </div>

      {/* Info row */}
      <div className="mb-4 grid grid-cols-3 gap-2 max-[640px]:grid-cols-1">
        <SummaryRow label="Scenario"      value={scenario.title} />
        <SummaryRow label="What it shows" value={scenario.description} />
        <SummaryRow label="Expected"      value={scenario.expected} />
      </div>

      <div className="flex flex-wrap gap-2">
        <Button variant="secondary" onClick={onLoad}>Load</Button>
        <Button variant="secondary" onClick={onRunQuote}>Run Quote</Button>
        <Button variant="secondary" onClick={onRunReserve}>Run Reserve</Button>
        <Button variant="primary"   onClick={onRunFull}>Run Full Flow</Button>
      </div>
    </Card>
  )
}
