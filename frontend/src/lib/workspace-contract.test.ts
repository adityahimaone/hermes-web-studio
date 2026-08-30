import { describe, expect, it } from 'vitest'
import { breadcrumbs } from './workspace-contract'

describe('workspace contract', () => {
  it('builds safe relative breadcrumbs', () => {
    expect(breadcrumbs('src/components')).toEqual([
      { label: 'Workspace', path: '.' },
      { label: 'src', path: 'src' },
      { label: 'components', path: 'src/components' },
    ])
  })
})
