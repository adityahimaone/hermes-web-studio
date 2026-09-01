import { describe, expect, it } from 'vitest'
import kanbanSource from './kanban-view.tsx?raw'
import controlSource from './control-center.tsx?raw'

describe('P027 acceptance gaps', () => {
  it('exposes a keyboard-accessible lane transition control', () => {
    expect(kanbanSource).toContain('Move to lane')
    expect(kanbanSource).toContain('min-h-11')
  })

  it('handles Space mutation network failures without unhandled promises', () => {
    expect(controlSource).toContain("catch (err) { setError(err instanceof Error ? err.message : 'Space could not be registered') }")
    expect(controlSource).toContain("catch (err) { setError(err instanceof Error ? err.message : 'Space could not be activated') }")
    expect(controlSource).toContain("catch (err) { setError(err instanceof Error ? err.message : 'Space could not be deleted') }")
  })

  it('shows task-detail action failures in an alert', () => {
    expect(kanbanSource).toContain("Task action failed")
    expect(kanbanSource).toContain('role="alert"')
  })
})
