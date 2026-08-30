import { describe, expect, it } from 'vitest'
import { initialDiscoveryState, reduceDiscoveryState } from './discovery-state'

describe('discovery preview state', () => {
  it('does not treat an empty successful preview as loading', () => {
    const state = reduceDiscoveryState(
      reduceDiscoveryState(initialDiscoveryState, { type: 'preview-start' }),
      { type: 'preview-success', content: '' },
    )

    expect(state).toEqual({ preview: 'ready', content: '', previewError: '' })
  })

  it('keeps preview failures separate from the discovery list', () => {
    const state = reduceDiscoveryState(initialDiscoveryState, { type: 'preview-error', message: 'Preview unavailable' })

    expect(state.preview).toBe('error')
    expect(state.previewError).toBe('Preview unavailable')
  })
})
