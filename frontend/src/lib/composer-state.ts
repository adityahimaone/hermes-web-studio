import { validModelSelection, type ModelCatalogItem } from './model-catalog'

export function resolveComposerModel(profile: { model?: string; provider?: string }, catalog: ModelCatalogItem[], status: 'loading' | 'ready' | 'unavailable' | 'error') {
  const model = profile.model || 'default'
  const provider = profile.provider || ''
  const stale = model !== 'default' && (status !== 'ready' || !validModelSelection(catalog, model, provider))
  return { model, provider, stale }
}
