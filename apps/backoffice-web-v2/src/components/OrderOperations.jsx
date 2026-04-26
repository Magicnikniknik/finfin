import { Card, CardTitle } from './ui/Card'
import { Input } from './ui/Input'
import { Button } from './ui/Button'

function FormSection({ title, children }) {
  return (
    <div className="flex flex-col gap-3">
      <p className="text-[0.8rem] font-medium text-white/55">{title}</p>
      {children}
    </div>
  )
}

export function OrderOperations({ completeForm, setCompleteForm, cancelForm, setCancelForm, onComplete, onCancel }) {
  const setC = (key) => (e) => setCompleteForm((f) => ({ ...f, [key]: e.target.value }))
  const setX = (key) => (e) => setCancelForm((f) => ({ ...f, [key]: e.target.value }))

  return (
    <Card className="p-5">
      <CardTitle>Order Operations</CardTitle>
      <div className="grid grid-cols-2 gap-5 max-[640px]:grid-cols-1">
        <form onSubmit={(e) => { e.preventDefault(); onComplete() }}>
          <FormSection title="Complete order">
            <Input label="Idempotency key"   value={completeForm.idempotency_key}  onChange={setC('idempotency_key')} required />
            <Input label="Order ID"          value={completeForm.order_id}         onChange={setC('order_id')}         required />
            <Input label="Expected version"  value={completeForm.expected_version} onChange={setC('expected_version')} type="number" min="0" required />
            <Input label="Cashier"           value={completeForm.cashier_id}       onChange={setC('cashier_id')}       required />
            <Button type="submit" variant="success" className="mt-1">Complete</Button>
          </FormSection>
        </form>

        <form onSubmit={(e) => { e.preventDefault(); onCancel() }}>
          <FormSection title="Cancel order">
            <Input label="Idempotency key"   value={cancelForm.idempotency_key}  onChange={setX('idempotency_key')} required />
            <Input label="Order ID"          value={cancelForm.order_id}         onChange={setX('order_id')}         required />
            <Input label="Expected version"  value={cancelForm.expected_version} onChange={setX('expected_version')} type="number" min="0" required />
            <Input label="Reason"            value={cancelForm.reason}           onChange={setX('reason')}           required />
            <Button type="submit" variant="danger" className="mt-1">Cancel</Button>
          </FormSection>
        </form>
      </div>
    </Card>
  )
}
