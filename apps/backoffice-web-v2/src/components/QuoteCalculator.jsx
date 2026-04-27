import { Card, CardTitle } from './ui/Card'
import { Input, Select } from './ui/Input'
import { Button } from './ui/Button'

export function QuoteCalculator({ form, setForm, onSubmit, onCopyId, onUseInReserve, currentScenarioId }) {
  const set = (key) => (e) => setForm((f) => ({ ...f, [key]: e.target.value }))

  return (
    <Card className="p-5">
      <CardTitle info="Request a price quote from the pricing engine. Give/Get currencies define trade direction. Amount is in the Give currency. The quote has a short TTL — reserve quickly after receiving it. Use 'Use in Reserve' to carry the Quote ID forward.">Quote Calculator</CardTitle>
      <form onSubmit={(e) => { e.preventDefault(); onSubmit() }} className="grid grid-cols-2 gap-4 max-[640px]:grid-cols-1">
        <Input label="Office ID"       value={form.office_id}         onChange={set('office_id')}         required />
        <Select label="Input mode"     value={form.input_mode}        onChange={set('input_mode')}>
          <option value="GIVE">Give</option>
          <option value="GET">Get</option>
        </Select>
        <Input label="Give currency"   value={form.give_currency_id}  onChange={set('give_currency_id')}  required />
        <Input label="Get currency"    value={form.get_currency_id}   onChange={set('get_currency_id')}   required />
        <Input label="Amount"          value={form.amount}            onChange={set('amount')}            required />
        <Input label="Scenario"        value={currentScenarioId}      readOnly />
        <div className="col-span-2 flex flex-wrap gap-2 max-[640px]:col-span-1">
          <Button type="submit">Calculate Quote</Button>
          <Button type="button" variant="secondary" onClick={onCopyId}>Copy ID</Button>
          <Button type="button" variant="secondary" onClick={onUseInReserve}>Use in Reserve</Button>
        </div>
      </form>
    </Card>
  )
}
