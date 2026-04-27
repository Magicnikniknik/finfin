import { useState, useEffect } from 'react'
import { useAppState } from './hooks/useAppState'

import { Header }           from './components/Header'
import { KpiBar }           from './components/KpiBar'
import { ConfigSection }    from './components/ConfigSection'
import { DemoSession }      from './components/DemoSession'
import { ScenarioConsole }  from './components/ScenarioConsole'
import { QuoteCalculator }  from './components/QuoteCalculator'
import { QuoteSummary }     from './components/QuoteSummary'
import { ReserveWorkspace } from './components/ReserveWorkspace'
import { OrderOperations }  from './components/OrderOperations'
import { OrderSummary }     from './components/OrderSummary'
import { StateTimeline }    from './components/StateTimeline'
import { AuditPanel }       from './components/AuditPanel'
import { DebugPanel }       from './components/DebugPanel'

function useUptime() {
  const [secs, setSecs] = useState(0)
  useEffect(() => {
    const t = setInterval(() => setSecs(s => s + 1), 1000)
    return () => clearInterval(t)
  }, [])
  const h = String(Math.floor(secs / 3600)).padStart(2, '0')
  const m = String(Math.floor((secs % 3600) / 60)).padStart(2, '0')
  const s = String(secs % 60).padStart(2, '0')
  return `${h}:${m}:${s}`
}

export default function App() {
  const s = useAppState()
  const uptime = useUptime()
  const [showConnection, setShowConnection] = useState(false)

  return (
    <div className="min-h-screen px-5 py-5 max-w-[1600px] mx-auto flex flex-col gap-3">

      <Header />

      <KpiBar metrics={s.appState.metrics} />

      {/* 3-column workspace */}
      <div className="grid grid-cols-3 gap-3 max-[1200px]:grid-cols-1">

        {/* Left — controls */}
        <div className="flex flex-col gap-3">
          {/* Connection — collapsed by default */}
          <button
            onClick={() => setShowConnection(v => !v)}
            className="flex items-center justify-between rounded-btn border border-line-subtle bg-surface-panel px-3.5 py-2 text-[0.78rem] text-ink-tertiary transition-colors hover:border-line-strong hover:text-ink-secondary"
          >
            <span className="flex items-center gap-2">
              <span className="text-ink-muted">⚙</span> Connection
            </span>
            <span className="font-mono text-[0.65rem]">{showConnection ? '↑' : '↓'}</span>
          </button>
          {showConnection && (
            <ConfigSection config={s.config} setConfig={s.setConfig} />
          )}

          <DemoSession
            presentationMode={s.appState.presentationMode}
            onLoadDefaults={s.loadSandboxDefaults}
            onReset={s.resetDemoState}
            onHappyDemo={s.runHappyDemo}
            onErrorDemo={s.runErrorDemo}
            onTogglePresentation={s.togglePresentationMode}
          />
          <ScenarioConsole
            currentId={s.appState.currentScenarioId}
            onSelect={s.selectScenario}
            onLoad={() => s.selectScenario(s.appState.currentScenarioId)}
            onRunQuote={s.runQuote}
            onRunReserve={s.runReserve}
            onRunFull={s.runFullFlow}
          />
        </div>

        {/* Center — quote */}
        <div className="flex flex-col gap-3">
          <QuoteCalculator
            form={s.quoteForm}
            setForm={s.setQuoteForm}
            onSubmit={s.runQuote}
            onCopyId={s.copyQuoteId}
            onUseInReserve={s.useQuoteInReserve}
            currentScenarioId={s.appState.currentScenarioId}
          />
          <QuoteSummary data={s.quoteSummaryData} />
        </div>

        {/* Right — execution: reserve → complete/cancel → result */}
        <div className="flex flex-col gap-3">
          <ReserveWorkspace
            form={s.reserveForm}
            setForm={s.setReserveForm}
            onSubmit={s.runReserve}
          />
          <OrderOperations
            completeForm={s.completeForm}
            setCompleteForm={s.setCompleteForm}
            cancelForm={s.cancelForm}
            setCancelForm={s.setCancelForm}
            onComplete={s.runComplete}
            onCancel={s.runCancel}
          />
          <OrderSummary data={s.orderSummaryData} />
        </div>
      </div>

      {/* Timeline — full width, always visible */}
      <StateTimeline steps={s.timeline} />

      {/* Audit — full width, collapsible */}
      <AuditPanel rows={s.auditRows} />

      {/* Debug — dev only */}
      {!s.appState.presentationMode && <DebugPanel response={s.response} />}

      {/* Footer */}
      <footer className="mt-1 border-t border-line-subtle pt-3 flex justify-between items-center">
        <div className="flex gap-5">
          <span className="text-[0.65rem] font-semibold uppercase tracking-widest text-ink-muted">
            Environment: <span className="text-ink-tertiary">Local</span>
          </span>
          <span className="text-[0.65rem] font-semibold uppercase tracking-widest text-ink-muted">
            Uptime: <span className="font-mono text-ink-tertiary">{uptime}</span>
          </span>
        </div>
        <span className="text-[0.65rem] text-ink-muted">V 1.0.4-BETA</span>
      </footer>

    </div>
  )
}
