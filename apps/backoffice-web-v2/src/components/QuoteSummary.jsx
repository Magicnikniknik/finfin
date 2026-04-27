import { Card, CardTitle } from './ui/Card'
import { Badge } from './ui/Badge'
import { SummaryRow } from './ui/Input'
import { mapStatusVariant } from '../lib/utils'

export function QuoteSummary({ data }) {
  const variant = data ? mapStatusVariant(data.status) : 'neutral'

  return (
    <Card className="p-5">
      <div className="mb-4 flex items-center justify-between gap-2">
        <CardTitle className="mb-0 border-0 pb-0" info="Pricing engine response. Key fields to watch: base_rate (raw market price), client_rate (after margin), fee (operator charge). Badge shows quote status — green means still valid and reservable.">Quote Summary</CardTitle>
        <Badge variant={variant}>{data?.status ?? 'no quote'}</Badge>
      </div>
      <div className="grid grid-cols-2 gap-2 max-[640px]:grid-cols-1">
        <SummaryRow label="Quote ID"     value={data?.quote_id} />
        <SummaryRow label="office"       value={data?.office} />
        <SummaryRow label="source"       value={data?.source_name} />
        <SummaryRow label="base rate"    value={data?.base_rate} />
        <SummaryRow label="client rate"  value={data?.fixed_rate} />
        <SummaryRow label="fee"          value={data?.fee} />
        <SummaryRow label="applied rule" value={data?.rule} />
        <SummaryRow label="expires_at"   value={data?.expires_at} />
        <SummaryRow label="give"         value={data?.give} />
        <SummaryRow label="get"          value={data?.get} />
      </div>
    </Card>
  )
}
