import { describe, expect, it } from 'vitest'
import { normalizeSpace, normalizeSpaces } from './space-contract'

describe('space contract', () => {
  it('keeps remote spaces explicit and never derives local paths', () => {
    expect(normalizeSpace({ id: 'remote-1', name: 'Mac', location_kind: 'remote', workspace_ref: 'mac-project', health: 'unavailable' })).toEqual({ id: 'remote-1', name: 'Mac', locationKind: 'remote', workspaceRef: 'Remote workspace unavailable', health: 'unavailable', active: false, order: 0 })
  })
  it('rejects malformed entries and duplicate IDs', () => {
    expect(normalizeSpaces({ spaces: [{ id: 'x' }] })).toEqual([])
    expect(normalizeSpaces([{ id: 'x', name: 'one' }, { id: 'x', name: 'two' }, { id: '', name: 'bad' }])).toHaveLength(1)
  })
})
