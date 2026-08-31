import { describe, expect, it } from 'vitest'
import { profileMeta, profileStatusLabel, type Profile } from './profile-contract'

const profile: Profile = { id: 'default', name: 'Default', model: 'codex', provider: 'gateway', health: 'gateway' }

describe('profile presentation contract', () => {
  it('keeps compact profile metadata ordered and readable', () => {
    expect(profileMeta(profile)).toBe('codex · gateway')
    expect(profileMeta({ ...profile, provider: undefined })).toBe('codex')
  })

  it('marks only the active row as active', () => {
    expect(profileStatusLabel(profile, 'default')).toBe('Active')
    expect(profileStatusLabel(profile, 'other')).toBe('gateway')
  })
})
