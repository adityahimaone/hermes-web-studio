export type ModelCatalogItem = { id: string; label: string; provider: string; aliases: string[]; capabilities: string[]; available: boolean }
export type ModelCatalogGroup = { provider: string; models: ModelCatalogItem[] }

export function normalizeModelCatalog(models: ModelCatalogItem[]): ModelCatalogItem[] {
  const seen = new Set<string>()
  return models.flatMap(model => {
    const id = clean(model.id, 256); const provider = clean(model.provider, 128); const key = modelKey(id, provider); if (!id || seen.has(key)) return []
    seen.add(key)
    const aliases = [...new Set(model.aliases.map(alias => clean(alias, 128)).filter(Boolean))]
    return [{ ...model, id, label: clean(model.label, 256), provider, aliases }]
  })
}

function clean(value: string, limit: number) {
  const sanitized = value.replace(/[\u0000-\u001f\u007f-\u009f]/g, '').trim()
  let result = ''
  for (const character of sanitized) {
    if (new TextEncoder().encode(result + character).length > limit) break
    result += character
  }
  return result
}

export function utf8Length(value: string, limit: number) {
  return clean(value, limit)
}

export function modelKey(id: string, provider = '') {
  return JSON.stringify([clean(provider, 128), clean(id, 256)])
}

export function encodeModelSelectionValue(provider: string, id: string) {
  return encodeURIComponent(JSON.stringify([provider, id]))
}

export function decodeModelSelectionValue(value: string) {
  try {
    const decoded: unknown = JSON.parse(decodeURIComponent(value))
    if (Array.isArray(decoded) && decoded.length === 2 && typeof decoded[0] === 'string' && typeof decoded[1] === 'string') {
      return { provider: decoded[0], id: decoded[1] }
    }
  } catch {
    // Invalid option values fail closed.
  }
  return undefined
}

export function modelSelectionValue(model: Pick<ModelCatalogItem, 'provider' | 'id'>) {
  return encodeModelSelectionValue(model.provider, model.id)
}

export function decodeModelSelection(value: string, models: ModelCatalogItem[]) {
  const selection = decodeModelSelectionValue(value)
  return selection ? findCatalogModel(models, selection.id, selection.provider) : undefined
}

export function findCatalogModel(models: ModelCatalogItem[], id: string, provider = '') {
  const normalizedId = clean(id, 256)
  const normalizedProvider = clean(provider, 128)
  return normalizeModelCatalog(models).find(model => model.available && model.id === normalizedId && (!normalizedProvider || model.provider === normalizedProvider))
}

export function validModelSelection(models: ModelCatalogItem[], id: string, provider = '') {
  return id === 'default' || Boolean(findCatalogModel(models, id, provider))
}


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
