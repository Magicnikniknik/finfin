import { Card, CardTitle } from './ui/Card'
import { Button } from './ui/Button'
import { Badge } from './ui/Badge'

export function DemoSession({ presentationMode, onLoadDefaults, onReset, onHappyDemo, onErrorDemo, onTogglePresentation }) {
  return (
    <Card className="mb-4 p-5">
      <div className="mb-4 flex items-center justify-between gap-3">
        <CardTitle className="mb-0 border-0 pb-0">Demo Session</CardTitle>
        <Badge variant={presentationMode ? 'ok' : 'neutral'}>
          {presentationMode ? 'presentation on' : 'presentation off'}
        </Badge>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button variant="success"   onClick={onLoadDefaults}>Load Defaults</Button>
        <Button variant="danger"    onClick={onReset}>Reset</Button>
        <Button variant="primary"   onClick={onHappyDemo}>Happy Path</Button>
        <Button variant="danger"    onClick={onErrorDemo}>Error Path</Button>
        <Button variant="secondary" onClick={onTogglePresentation}>
          {presentationMode ? 'Show Debug' : 'Hide Debug'}
        </Button>
      </div>
    </Card>
  )
}
