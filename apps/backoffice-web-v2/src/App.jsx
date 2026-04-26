import { useAppState } from './hooks/useAppState'
import { getScenarioById } from './lib/scenarios'

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

export default function App() {
  const s = useAppState()

  return (
    <div className="min-h-screen px-6 py-6 pb-14 max-w-[1600px] mx-auto">
      <Header />

      <KpiBar metrics={s.appState.metrics} />

      <ConfigSection config={s.config} setConfig={s.setConfig} />

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

      {/* Main workspace */}
      <div className="grid grid-cols-[2fr_1fr] gap-4 max-[1100px]:grid-cols-1">
        {/* Left column */}
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
          <ReserveWorkspace
            form={s.reserveForm}
            setForm={s.setReserveForm}
            onSubmit={s.runReserve}
          />
        </div>

        {/* Right column */}
        <div className="flex flex-col gap-3">
          <OrderOperations
            completeForm={s.completeForm}
            setCompleteForm={s.setCompleteForm}
            cancelForm={s.cancelForm}
            setCancelForm={s.setCancelForm}
            onComplete={s.runComplete}
            onCancel={s.runCancel}
          />
          <OrderSummary data={s.orderSummaryData} />
          <StateTimeline steps={s.timeline} />
          <AuditPanel rows={s.auditRows} />
          {!s.appState.presentationMode && <DebugPanel response={s.response} />}
        </div>
      </div>

      {/* Footer walkthrough */}
      <footer className="mt-4 rounded-card glass px-6 py-4">
        <p className="mb-3 text-xs text-white/30">Walkthrough</p>
        <ol className="flex flex-wrap gap-2">
          {['Load scenario and run quote.', 'Reserve from quote.', 'Complete or cancel order.', 'Review timeline + audit.'].map((step, i) => (
            <li key={i} className="flex items-center gap-2 rounded-chip border border-white/[0.06] bg-white/[0.03] px-3 py-1.5 text-xs text-white/35">
              <span className="flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-[#0A84FF]/20 text-[0.6rem] font-semibold text-[#0A84FF]">
                {i + 1}
              </span>
              {step}
            </li>
          ))}
        </ol>
      </footer>
    </div>
  )
}

