import { describe, expect, it } from 'vitest'
import { sessionActionVisibilityClass } from './session-rail'

describe('session action visibility', () => {
  it('keeps overflow trigger visible for touch while retaining hover and focus styling', () => {
    const classes = sessionActionVisibilityClass()

    expect(classes).not.toContain('opacity-0')
    expect(classes).toContain('group-hover:opacity-100')
    expect(classes).toContain('group-focus-within:opacity-100')
  })
})
