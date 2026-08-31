import { describe, expect, it } from 'vitest'
import { filterAriaPressed, sessionActionVisibilityClass } from './session-rail'

describe('session filter accessibility', () => {
  it.each([
    ['CLI selected', 'cli', 'cli', true],
    ['CLI unselected', 'all', 'cli', false],
    ['All selected', 'all', 'all', true],
    ['All unselected', 'webui', 'all', false],
    ['Unassigned selected', 'unassigned', 'unassigned', true],
    ['Unassigned unselected', 'all', 'unassigned', false],
    ['dynamic tag selected', 'work', 'work', true],
    ['dynamic tag unselected', 'all', 'work', false],
  ])('%s', (_label, selected, value, expected) => {
    expect(filterAriaPressed(selected, value)).toBe(expected)
  })
})

describe('session action visibility', () => {
  it('keeps overflow trigger visible for touch while retaining hover and focus styling', () => {
    const classes = sessionActionVisibilityClass()

    expect(classes).not.toContain('opacity-0')
    expect(classes).toContain('group-hover:opacity-100')
    expect(classes).toContain('group-focus-within:opacity-100')
  })
})
