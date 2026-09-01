export type Space = { id: string; name: string; locationKind: 'local' | 'remote'; workspaceRef: string; health: string; active: boolean; order: number }

export function normalizeSpace(raw: unknown): Space {
  const value = raw && typeof raw === 'object' ? raw as Record<string, unknown> : {}
  const locationKind = value.location_kind === 'remote' ? 'remote' : 'local'
  return { id: String(value.id || ''), name: String(value.name || value.title || ''), locationKind, workspaceRef: String(value.workspace_ref || ''), health: String(value.health || value.status || 'unknown'), active: Boolean(value.active), order: Number(value.order) || 0 }
}

export function normalizeSpaces(raw: unknown): Space[] {
  return Array.isArray(raw) ? raw.map(normalizeSpace).sort((a, b) => a.order - b.order || a.name.localeCompare(b.name)) : []
}
