import { Card, CardTitle } from './ui/Card'
import { Badge } from './ui/Badge'
import { mapStatusVariant } from '../lib/utils'

export function DebugPanel({ response }) {
  return (
    <Card className="p-5">
      <div className="mb-3 flex items-center justify-between gap-2">
        <CardTitle className="mb-0 border-0 pb-0">Raw Response</CardTitle>
        <Badge variant={mapStatusVariant(response.status)}>{response.text}</Badge>
      </div>
      <div className="relative overflow-hidden rounded-input">
        <pre className={cn(
          'min-h-[220px] overflow-auto rounded-input p-4',
          'border border-white/[0.06] bg-black/60',
          'font-mono text-[0.75rem] leading-relaxed text-white/35',
          response.loading && 'opacity-50',
        )}>
          {JSON.stringify(response.json, null, 2)}
        </pre>
        {response.loading && (
          <div className="pointer-events-none absolute inset-0 overflow-hidden rounded-input">
            <div className="absolute inset-y-0 w-1/3 animate-shimmer bg-gradient-to-r from-transparent via-white/[0.04] to-transparent" />
          </div>
        )}
      </div>
    </Card>
  )
}

function cn(...c) { return c.filter(Boolean).join(' ') }
