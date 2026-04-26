import { Card, CardTitle } from './ui/Card'
import { Input } from './ui/Input'

export function ConfigSection({ config, setConfig }) {
  const set = (key) => (e) => setConfig((c) => ({ ...c, [key]: e.target.value }))

  return (
    <Card className="mb-4 p-5">
      <CardTitle>Connection</CardTitle>
      <div className="grid grid-cols-2 gap-3 max-[640px]:grid-cols-1">
        <Input label="API base URL" value={config.baseUrl}  onChange={set('baseUrl')}  />
        <Input label="Tenant ID"    value={config.tenantId} onChange={set('tenantId')} placeholder="11111111-1111-1111-1111-111111111111" />
        <Input label="Client ref"   value={config.clientRef} onChange={set('clientRef')} placeholder="client_demo_console" />
        <Input label="cashier"      value={config.cashierId} onChange={set('cashierId')} />
      </div>
    </Card>
  )
}
