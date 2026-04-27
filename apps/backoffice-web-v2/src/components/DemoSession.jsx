import { Card, CardTitle } from './ui/Card'
import { Button } from './ui/Button'
import { Badge } from './ui/Badge'

export function DemoSession({ presentationMode, onLoadDefaults, onReset, onHappyDemo, onErrorDemo, onTogglePresentation }) {
  return (
    <Card className="p-4">
      <div className="mb-4 flex items-center justify-between gap-3">
        <CardTitle className="mb-0 border-0 pb-0" info="Orchestrated demo flows. Happy Path runs the ideal lifecycle: quote → reserve → complete. Error Path triggers failure scenarios. Presentation mode hides raw debug output for clean demos.">Demo Session</CardTitle>
        <Badge variant={presentationMode ? 'ok' : 'neutral'}>
          {presentationMode ? 'presentation on' : 'presentation off'}
        </Badge>
      </div>
      <div className="flex items-center gap-2">
        <div className="flex flex-1 gap-2 overflow-x-auto [&::-webkit-scrollbar]:hidden">
          <Button variant="secondary" onClick={onLoadDefaults}>Load Defaults</Button>
          <Button variant="primary"   onClick={onHappyDemo}>Happy Path</Button>
          <Button variant="secondary" onClick={onErrorDemo}>Error Path</Button>
          <Button variant="ghost"     onClick={onTogglePresentation}>
            {presentationMode ? 'Show Debug' : 'Hide Debug'}
          </Button>
        </div>
        <div className="h-5 w-px shrink-0 bg-line-subtle" />
        <Button variant="danger" onClick={onReset}>Reset</Button>
      </div>
    </Card>
  )
}
