import { describe, expect, it } from 'vitest'
import { resolveComposerModel } from './composer-state'

describe('composer model state', () => {
  it('preserves stale profile model as unavailable instead of defaulting', () => {
    expect(resolveComposerModel({ model: 'legacy-model', provider: 'legacy' }, [], 'ready')).toEqual({ model: 'legacy-model', provider: 'legacy', stale: true })
  })

  it('uses default when profile has no stale model', () => {
    expect(resolveComposerModel({ model: 'default', provider: '' }, [], 'ready')).toEqual({ model: 'default', provider: '', stale: false })
  })
})
