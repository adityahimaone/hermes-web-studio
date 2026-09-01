import { describe, expect, it } from 'vitest'
import { resolveComposerModel } from './composer-state'

describe('composer model state', () => {
  it('preserves stale profile model as unavailable instead of defaulting', () => {
    expect(resolveComposerModel({ model: 'legacy-model', provider: 'legacy' }, [], 'ready')).toEqual({ model: 'legacy-model', provider: 'legacy', stale: true })
  })

  it('uses default when profile has no stale model', () => {
    expect(resolveComposerModel({ model: 'default', provider: '' }, [], 'ready')).toEqual({ model: 'default', provider: '', stale: false })
  })

  it.each(['unavailable', 'error', 'loading'] as const)('blocks stale profile model while catalog is %s', (status) => {
    expect(resolveComposerModel({ model: 'legacy-model', provider: 'legacy' }, [], status)).toEqual({ model: 'legacy-model', provider: 'legacy', stale: true })
  })

  it('accepts valid profile model once ready catalog confirms it', () => {
    expect(resolveComposerModel({ model: 'ready-model', provider: 'gateway' }, [{ id: 'ready-model', label: 'Ready model', provider: 'gateway', aliases: [], capabilities: [], available: true }], 'ready')).toEqual({ model: 'ready-model', provider: 'gateway', stale: false })
  })

  it('keeps empty ready catalog distinct from unavailable catalog', () => {
    expect(resolveComposerModel({ model: 'legacy-model', provider: 'legacy' }, [], 'ready').stale).toBe(true)
    expect(resolveComposerModel({ model: 'legacy-model', provider: 'legacy' }, [], 'unavailable').stale).toBe(true)
  })
})
