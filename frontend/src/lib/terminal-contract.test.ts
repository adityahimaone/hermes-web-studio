import { describe, expect, it } from 'vitest'
import { terminalCapability } from './terminal-contract'

describe('terminal capability contract', () => {
  it('accepts an explicit unavailable capability without terminal output', () => {
    expect(terminalCapability({ id: 'terminal', available: false, state: 'unavailable', reason: 'sandbox_required', message: 'Contained runtime required.' })).toMatchObject({ available: false, state: 'unavailable' })
  })

  it('rejects an optimistic or incomplete capability response', () => {
    expect(() => terminalCapability({ id: 'terminal', available: true })).toThrow('unsafe state')
    expect(() => terminalCapability({ id: 'terminal', available: false, state: 'unavailable' })).toThrow('incomplete')
  })
})
