export type Space = { id: string; name: string; locationKind: 'local' | 'remote'; workspaceRef: string; health: string; active: boolean; order: number }

export function normalizeSpace(raw: unknown): Space {
  const value = raw && typeof raw === 'object' ? raw as Record<string, unknown> : {}
  const locationKind = value.location_kind === 'remote' ? 'remote' : 'local'
  const workspaceRef = locationKind === 'remote' ? 'Remote workspace unavailable' : String(value.workspace_ref || '')
  return { id: String(value.id || ''), name: String(value.name || value.title || ''), locationKind, workspaceRef, health: String(value.health || value.status || 'unknown'), active: Boolean(value.active), order: Number.isFinite(Number(value.order)) ? Number(value.order) : 0 }
}

export function normalizeSpaces(raw: unknown): Space[] {
  if (!Array.isArray(raw)) return []
  const seen = new Set<string>()
  return raw.map(normalizeSpace).filter(space => space.id && space.name && !seen.has(space.id) && seen.add(space.id)).sort((a, b) => a.order - b.order || a.name.localeCompare(b.name))
}
