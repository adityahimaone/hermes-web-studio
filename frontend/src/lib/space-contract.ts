export type Space = { id: string; name: string; locationKind: 'local' | 'remote'; workspaceRef: string; health: string; active: boolean; order: number }

const stringValue = (value: unknown) => typeof value === 'string' ? value : ''

export function normalizeSpace(raw: unknown): Space {
  const value = raw && typeof raw === 'object' ? raw as Record<string, unknown> : {}
  const locationKind = value.location_kind === 'remote' ? 'remote' : value.location_kind === 'local' ? 'local' : 'local'
  return { id: stringValue(value.id), name: stringValue(value.name) || stringValue(value.title), locationKind, workspaceRef: locationKind === 'remote' ? 'Remote workspace unavailable' : stringValue(value.workspace_ref), health: stringValue(value.health) || stringValue(value.status) || 'unknown', active: value.active === true, order: typeof value.order === 'number' && Number.isFinite(value.order) ? value.order : 0 }
}

export function normalizeSpaces(raw: unknown): Space[] {
  if (!Array.isArray(raw)) return []
  const seen = new Set<string>()
  return raw.map(normalizeSpace).filter(space => space.id && space.name && !seen.has(space.id) && seen.add(space.id)).sort((a, b) => a.order - b.order || a.name.localeCompare(b.name))
}
