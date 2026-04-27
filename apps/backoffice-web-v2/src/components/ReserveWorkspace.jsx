import { Card, CardTitle } from './ui/Card'
import { Input, Select } from './ui/Input'
import { Button } from './ui/Button'

function MoneyGroup({ title, children }) {
  return (
    <div className="rounded-input border border-white/[0.07] bg-white/[0.02] p-4 space-y-3">
      <p className="text-[0.8rem] font-medium text-white/40">{title}</p>
      {children}
    </div>
  )
}

export function ReserveWorkspace({ form, setForm, onSubmit }) {
  const set = (key) => (e) => setForm((f) => ({ ...f, [key]: e.target.value }))

  return (
    <Card className="p-5">
      <CardTitle info="Lock in the quoted price by creating a reserve order. Requires a valid Quote ID from the summary above. After reserving, the order is RESERVED — funds are earmarked. Proceed to Order Operations to complete or cancel.">Reserve</CardTitle>
      <form onSubmit={(e) => { e.preventDefault(); onSubmit() }} className="space-y-4">
        <div className="grid grid-cols-2 gap-4 max-[640px]:grid-cols-1">
          <Input label="Idempotency key" value={form.idempotency_key} onChange={set('idempotency_key')} required />
          <Select label="Side"           value={form.side}            onChange={set('side')}>
            <option value="BUY">Buy</option>
            <option value="SELL">Sell</option>
          </Select>
          <Input label="Office ID"  value={form.office_id} onChange={set('office_id')} required />
          <Input label="Quote ID"   value={form.quote_id}  onChange={set('quote_id')}  required />
        </div>

        <div className="grid grid-cols-2 gap-3 max-[640px]:grid-cols-1">
          <MoneyGroup title="Give">
            <Input label="Amount"   value={form.give_amount}  onChange={set('give_amount')}  required />
            <Input label="Currency" value={form.give_code}    onChange={set('give_code')}    required />
            <Input label="Network"  value={form.give_network} onChange={set('give_network')} required />
          </MoneyGroup>
          <MoneyGroup title="Get">
            <Input label="Amount"   value={form.get_amount}   onChange={set('get_amount')}   required />
            <Input label="Currency" value={form.get_code}     onChange={set('get_code')}     required />
            <Input label="Network"  value={form.get_network}  onChange={set('get_network')}  required />
          </MoneyGroup>
        </div>

        <Button type="submit">Reserve Order</Button>
      </form>
    </Card>
  )
}
