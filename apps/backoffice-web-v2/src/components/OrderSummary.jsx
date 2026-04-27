import { Card, CardTitle } from './ui/Card'
import { Badge } from './ui/Badge'
import { SummaryRow } from './ui/Input'
import { mapStatusVariant } from '../lib/utils'

export function OrderSummary({ data }) {
  const variant = data ? mapStatusVariant(data.status) : 'neutral'

  return (
    <Card className="p-5">
      <div className="mb-4 flex items-center justify-between gap-2">
        <CardTitle className="mb-0 border-0 pb-0" info="Result of the last order action. Badge shows final state: COMPLETED means the trade executed, CANCELLED means the reservation was released. Held asset and amount show what was locked during the reserve phase.">Order Summary</CardTitle>
        <Badge variant={variant}>{data?.status ?? 'no order'}</Badge>
      </div>
      <div className="grid grid-cols-2 gap-2 max-[640px]:grid-cols-1">
        <SummaryRow label="Order ID"     value={data?.order_id} />
        <SummaryRow label="Order ref"    value={data?.order_ref} />
        <SummaryRow label="status"       value={data?.status} />
        <SummaryRow label="Version"      value={data?.version} />
        <SummaryRow label="Expires"   value={data?.expires_at} />
        <SummaryRow label="Quote status" value={data?.quote_status} />
        <SummaryRow label="Held asset"   value={data?.held_currency} />
        <SummaryRow label="Held amount"  value={data?.held_amount} />
      </div>
    </Card>
  )
}
