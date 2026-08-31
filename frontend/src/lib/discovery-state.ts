export type DiscoveryState = {
  preview: 'idle' | 'loading' | 'ready' | 'error'
  content: string
  previewError: string
}

export const initialDiscoveryState: DiscoveryState = {
  preview: 'idle',
  content: '',
  previewError: '',
}

export type DiscoveryItem = { name: string; path: string; description?: string }

export type DiscoveryGroup = { key: string; items: DiscoveryItem[] }

/** Keep long discovery lists scannable without changing the server contract. */
export function groupSkills(items: DiscoveryItem[]): DiscoveryGroup[] {
  const groups = new Map<string, DiscoveryItem[]>()
  for (const item of [...items].sort((a, b) => a.name.localeCompare(b.name))) {
    const parts = item.path.split('/').filter(Boolean)
    const key = parts.length > 2 ? parts[0] : '(general)'
    groups.set(key, [...(groups.get(key) || []), item])
  }
  return [...groups.entries()].sort(([a], [b]) => a.localeCompare(b)).map(([key, groupedItems]) => ({ key, items: groupedItems }))
}

export type DiscoveryAction =
  | { type: 'preview-start' }
  | { type: 'preview-success'; content: string }
  | { type: 'preview-error'; message: string }
  | { type: 'reset' }

export function reduceDiscoveryState(state: DiscoveryState, action: DiscoveryAction): DiscoveryState {
  switch (action.type) {
    case 'preview-start':
      return { preview: 'loading', content: '', previewError: '' }
    case 'preview-success':
      return { preview: 'ready', content: action.content, previewError: '' }
    case 'preview-error':
      return { preview: 'error', content: '', previewError: action.message }
    case 'reset':
      return initialDiscoveryState
  }
}
