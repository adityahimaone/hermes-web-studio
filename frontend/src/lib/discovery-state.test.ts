import { describe, expect, it } from 'vitest'
import { groupSkills, initialDiscoveryState, reduceDiscoveryState } from './discovery-state'

describe('skill grouping', () => {
  it('groups nested skills by category and root skills under General', () => {
    expect(groupSkills([
      { name: 'zeta', path: 'zeta/SKILL.md' },
      { name: 'reminders', path: 'apple/apple-reminders/SKILL.md' },
      { name: 'research', path: 'research/arxiv/SKILL.md' },
      { name: 'notes', path: 'apple/apple-notes/SKILL.md' },
    ])).toEqual([
      { key: '(general)', items: [{ name: 'zeta', path: 'zeta/SKILL.md' }] },
      { key: 'apple', items: [{ name: 'notes', path: 'apple/apple-notes/SKILL.md' }, { name: 'reminders', path: 'apple/apple-reminders/SKILL.md' }] },
      { key: 'research', items: [{ name: 'research', path: 'research/arxiv/SKILL.md' }] },
    ])
  })
})

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
