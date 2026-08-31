export type Profile = {
  id: string
  name: string
  model: string
  provider?: string
  health: string
}

export function profileStatusLabel(profile: Profile, active: string): string {
  return active === profile.id ? 'Active' : profile.health || 'Unknown'
}

export function profileMeta(profile: Profile): string {
  return [profile.model, profile.provider].filter(Boolean).join(' · ') || 'No model configured'
}
