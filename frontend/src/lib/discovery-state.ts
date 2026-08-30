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
