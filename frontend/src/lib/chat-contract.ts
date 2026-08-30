export type ChatEventType =
  | 'token'
  | 'reasoning'
  | 'tool'
  | 'tool_complete'
  | 'done'
  | 'cancel'
  | 'apperror'

export interface ChatEvent {
  type: ChatEventType
  data: Record<string, unknown>
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  status?: 'streaming' | 'complete' | 'error' | 'cancelled'
}

export interface ToolActivity {
  id: string
  name: string
  status: 'running' | 'complete' | 'error'
}

export interface ChatState {
  answer: string
  reasoning: string
  tools: ToolActivity[]
  status: 'idle' | 'streaming' | 'complete' | 'error' | 'cancelled'
  error?: string
}

export const initialChatState: ChatState = {
  answer: '',
  reasoning: '',
  tools: [],
  status: 'idle',
}

function eventText(data: Record<string, unknown>, key: string) {
  return typeof data[key] === 'string' ? data[key] : ''
}

export function reduceChatEvent(state: ChatState, event: ChatEvent): ChatState {
  switch (event.type) {
    case 'token':
      return { ...state, status: 'streaming', answer: state.answer + eventText(event.data, 'text') }
    case 'reasoning':
      return { ...state, status: 'streaming', reasoning: state.reasoning + eventText(event.data, 'text') }
    case 'tool': {
      const id = eventText(event.data, 'tid') || `${eventText(event.data, 'name')}-${state.tools.length}`
      return {
        ...state,
        status: 'streaming',
        tools: [...state.tools, { id, name: eventText(event.data, 'name') || 'Hermes tool', status: 'running' }],
      }
    }
    case 'tool_complete': {
      const id = eventText(event.data, 'tid')
      const name = eventText(event.data, 'name')
      let matched = false
      const tools = state.tools.map((tool) => {
        if (!matched && ((id && tool.id === id) || (!id && tool.name === name && tool.status === 'running'))) {
          matched = true
          return { ...tool, status: event.data.is_error ? 'error' : 'complete' } as ToolActivity
        }
        return tool
      })
      return { ...state, tools }
    }
    case 'done':
      return {
        ...state,
        answer: state.answer || eventText(event.data, 'answer'),
        status: 'complete',
      }
    case 'cancel':
      return { ...state, status: 'cancelled' }
    case 'apperror':
      return { ...state, status: 'error', error: eventText(event.data, 'message') || 'Hermes could not complete the request.' }
  }
}

