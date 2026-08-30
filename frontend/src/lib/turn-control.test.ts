import { describe, expect, it } from 'vitest'
import { clearPendingTurns, normalizeTurnMode, planTurn, type PendingTurn } from './turn-control'

const existing: PendingTurn[] = [{ content: 'first', attachmentNames: [] }]
const next: PendingTurn = { content: 'replacement', attachmentNames: ['notes.txt'] }

describe('turn control', () => {
  it('defaults unknown modes to queue and preserves pending work', () => {
    expect(normalizeTurnMode('unknown')).toBe('queue')
    expect(planTurn('queue', existing, next)).toEqual([...existing, next])
  })

  it('interrupts with only the replacement turn', () => {
    expect(planTurn('interrupt', existing, next)).toEqual([next])
  })

  it('steers by replacing an identical pending intent', () => {
    expect(planTurn('steer', [{ content: 'replacement', attachmentNames: [] }], next)).toEqual([next])
  })

  it('clears pending work on every lifecycle exit', () => {
    expect(clearPendingTurns()).toEqual([])
  })
})
