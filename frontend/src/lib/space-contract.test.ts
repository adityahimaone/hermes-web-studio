import { describe, expect, it } from 'vitest'
import { normalizeSpace } from './space-contract'

describe('space contract', () => {
  it('keeps remote spaces explicit and never derives local paths', () => {
    expect(normalizeSpace({ id: 'remote-1', name: 'Mac', location_kind: 'remote', workspace_ref: 'mac-project', health: 'unavailable' })).toEqual({ id: 'remote-1', name: 'Mac', locationKind: 'remote', workspaceRef: 'mac-project', health: 'unavailable', active: false, order: 0 })
  })
})
