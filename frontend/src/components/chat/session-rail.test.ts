import { describe, expect, it } from 'vitest'
import { filterAriaPressed, findSessionButton, focusSessionButton, nextSessionId, sessionActionVisibilityClass, sessionRowAriaCurrent } from './session-rail'

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

describe('session keyboard navigation', () => {
  it('wraps ArrowDown and ArrowUp selection across visible session ids', () => {
    expect(nextSessionId(['one', 'two', 'three'], 'one', 'next')).toBe('two')
    expect(nextSessionId(['one', 'two', 'three'], 'one', 'previous')).toBe('three')
    expect(nextSessionId(['one', 'two', 'three'], 'three', 'next')).toBe('one')
  })

  it('focuses exact target button after selection rerender', () => {
    const root = document.createElement('div')
    const target = document.createElement('button')
    target.dataset.sessionId = 'beta'
    root.append(target)
    document.body.append(root)
    focusSessionButton(root, 'beta')
    expect(document.activeElement).toBe(target)
  })

  it('finds session button by exact data attribute without CSS selector escaping', () => {
    const root = document.createElement('div')
    const target = document.createElement('button')
    target.dataset.sessionId = 'quote" ]: hostile'
    root.append(target)
    expect(findSessionButton(root, 'quote" ]: hostile')).toBe(target)
    expect(findSessionButton(root, 'quote" ]: missing')).toBeUndefined()
  })
})

describe('active session accessibility', () => {
  it('marks only active session row as current', () => {
    expect(sessionRowAriaCurrent('active', 'active')).toBe('page')
    expect(sessionRowAriaCurrent('other', 'active')).toBeUndefined()
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
