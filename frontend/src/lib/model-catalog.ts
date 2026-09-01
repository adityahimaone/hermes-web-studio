export type ModelCatalogItem = { id: string; label: string; provider: string; aliases: string[]; capabilities: string[]; available: boolean }
export type ModelCatalogGroup = { provider: string; models: ModelCatalogItem[] }

export function searchModelCatalog(models: ModelCatalogItem[], query: string) {
  const needle = query.trim().toLowerCase()
  return needle ? models.filter(model => [model.id, model.label, model.provider, ...model.aliases].join(' ').toLowerCase().includes(needle)) : models
}

export function groupModelCatalog(models: ModelCatalogItem[]): ModelCatalogGroup[] {
  const groups = new Map<string, ModelCatalogItem[]>()
  for (const model of models) {
    if (!model.available) continue
    const provider = model.provider || 'unknown'
    groups.set(provider, [...(groups.get(provider) || []), model])
  }
  return [...groups].sort(([a], [b]) => a.localeCompare(b)).map(([provider, grouped]) => ({ provider, models: grouped.sort((a, b) => a.label.localeCompare(b.label)) }))
}
