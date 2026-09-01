import { describe, expect, it } from 'vitest'
import { createElement } from 'react'
import { readOptions, selectTriggerClassName } from './select'

describe('select primitive', () => {
  it('keeps default fields consistent with reference controls', () => {
    const classes = selectTriggerClassName({})

    expect(classes).toContain('h-9')
    expect(classes).toContain('rounded-xl')
    expect(classes).toContain('border-border/70')
    expect(classes).toContain('bg-card')
    expect(classes).toContain('focus-visible:ring-2')
  })

  it('supports compact controls and invalid state without page-level styling', () => {
    const classes = selectTriggerClassName({ size: 'compact', invalid: true })

    expect(classes).toContain('h-7')
    expect(classes).toContain('rounded-lg')
    expect(classes).toContain('border-destructive/60')
    expect(classes).not.toContain('min-h-9')
  })

  it('allows legacy page h-7 overrides to shrink the trigger', () => {
    const classes = selectTriggerClassName({ className: 'h-7 rounded-full' })

    expect(classes).not.toContain('min-h-9')
    expect(classes).toContain('h-7')
  })

  it('flattens optgroup options with provider labels', () => {
    const children = createElement('optgroup', { label: 'OpenAI' }, createElement('option', { value: 'gpt-5' }, 'GPT-5'))
    expect(readOptions(children)).toEqual([{ value: 'gpt-5', label: 'GPT-5', disabled: false, group: 'OpenAI' }])
  })
})
